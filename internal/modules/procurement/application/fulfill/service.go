// Package fulfill 实现「支付后补填开通信息 / 查询开通进度」：
// 校验表单 → 上游会话预检（可选能力）→ 保存提交值 → 创建采购单；以及汇总采购单与上游实时进度。
// 它与采购主服务共享契约端口，主服务只负责采购单生命周期。
package fulfill

import (
	"context"
	"errors"
	"html"
	"strings"
	"sync"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/modules/catalog/product/manualform"
	procurementcontract "github.com/dujiao-next/internal/modules/procurement/contract"
	procurementdomain "github.com/dujiao-next/internal/modules/procurement/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

const (
	capabilityOrderProgress = "order_progress"
	capabilitySessionCheck  = "session_check"

	sessionFormKey          = "session_json"
	progressCacheTTL        = 2 * time.Second
	upstreamProbeTimeout    = 8 * time.Second
	sessionCheckTimeout     = 20 * time.Second
	defaultSubmittingPollMs = 2000
	defaultProcessingPollMs = 3000
	// reasonCheckUnavailable 上游表示「暂时无法核验」，不能当作会话无效。
	reasonCheckUnavailable = "check_unavailable"
)

// Creator 为已支付订单创建采购单（由采购主服务实现）。
type Creator interface {
	CreateForOrder(orderID uint) error
}

type Options struct {
	Repository      procurementcontract.Repository
	Orders          procurementcontract.OrderRepository
	ProductMappings procurementcontract.ProductMappingReader
	SKUMappings     procurementcontract.SKUMappingReader
	Connections     procurementcontract.ConnectionProvider
	Creator         Creator
}

type Service struct {
	procRepo    procurementcontract.Repository
	orderRepo   procurementcontract.OrderRepository
	mappingRepo procurementcontract.ProductMappingReader
	skuMapRepo  procurementcontract.SKUMappingReader
	connections procurementcontract.ConnectionProvider
	creator     Creator

	// mu 按订单 ID 分片串行化「保存开通信息 + 创建采购单」，防止同一订单并发重复提交。
	mu [64]sync.Mutex
	// cache 缓存上游实时进度（procurement order ID → cacheEntry），合并多端轮询。
	cache sync.Map
}

var _ procurementcontract.FulfillUseCase = (*Service)(nil)

func New(options Options) *Service {
	return &Service{
		procRepo: options.Repository, orderRepo: options.Orders,
		mappingRepo: options.ProductMappings, skuMapRepo: options.SKUMappings,
		connections: options.Connections, creator: options.Creator,
	}
}

type cacheEntry struct {
	progress  *procurementcontract.UpstreamProgress
	expiresAt time.Time
}

func hasCapability(caps []string, name string) bool {
	for _, c := range caps {
		if c == name {
			return true
		}
	}
	return false
}

// resolveTarget 找到持有表单的订单：下单时是「父订单 + 每个商品一个子订单」，
// 表单在子订单的订单项上；无子订单（旧数据）时订单自身即目标。
// pendingOnly 为 true 时只认尚未提交表单的订单项。
func resolveTarget(order *procurementdomain.LocalOrder, pendingOnly bool) (*procurementdomain.LocalOrder, procurementdomain.LocalOrderItem, bool) {
	candidates := []*procurementdomain.LocalOrder{order}
	if order.ParentID == nil && len(order.Children) > 0 {
		candidates = candidates[:0]
		for i := range order.Children {
			candidates = append(candidates, &order.Children[i])
		}
	}
	for _, candidate := range candidates {
		for _, item := range candidate.Items {
			if (pendingOnly && item.IsFormPending()) || (!pendingOnly && item.HasFormSnapshot()) {
				return candidate, item, true
			}
		}
	}
	return nil, procurementdomain.LocalOrderItem{}, false
}

// SubmitFulfillInfo 支付后提交开通信息：校验 → 上游预检（如支持）→ 保存 → 创建采购单。
// 预检阶段与上游的 HTTP 交互不在任何数据库事务内；创建采购单失败时会清掉已保存的提交值，订单回到待填写状态。
func (s *Service) SubmitFulfillInfo(ctx context.Context, localOrderID uint, formData jsonmap.JSON) error {
	_, item, err := s.loadSubmittable(localOrderID)
	if err != nil {
		return err
	}

	_, normalized, err := manualform.ValidateAndNormalize(item.ManualFormSchemaSnapshotJSON, formData)
	if err != nil {
		return err
	}
	if len(normalized) == 0 {
		return manualform.ErrRequiredMissing
	}

	if err := s.precheckSession(ctx, item, normalized); err != nil {
		return err
	}

	mu := &s.mu[localOrderID%uint(len(s.mu))]
	mu.Lock()
	defer mu.Unlock()

	// 加锁后重读，拦截并发的重复提交。
	_, item, err = s.loadSubmittable(localOrderID)
	if err != nil {
		return err
	}
	if err := s.orderRepo.SaveManualFormSubmission(item.ID, normalized); err != nil {
		return err
	}
	// 与支付成功路径一致：以下单订单（父订单）ID 调用，由 CreateForOrder 遍历子订单。
	if err := s.creator.CreateForOrder(localOrderID); err != nil {
		if errors.Is(err, procurementcontract.ErrExists) {
			return procurementcontract.ErrFulfillInfoAlreadySubmitted
		}
		// 采购单没建成：撤销提交值，订单回到待填写，客户可重试。
		if rollbackErr := s.orderRepo.SaveManualFormSubmission(item.ID, jsonmap.JSON{}); rollbackErr != nil {
			logger.Errorw("procurement_fulfill_rollback_failed", "local_order_id", localOrderID, "error", rollbackErr)
		}
		return err
	}
	return nil
}

// loadSubmittable 返回持有待填表单的订单项；订单须为 paid 且尚无采购单。
func (s *Service) loadSubmittable(localOrderID uint) (*procurementdomain.LocalOrder, procurementdomain.LocalOrderItem, error) {
	var none procurementdomain.LocalOrderItem
	order, err := s.orderRepo.GetByID(localOrderID)
	if err != nil {
		return nil, none, err
	}
	if order == nil {
		return nil, none, procurementcontract.ErrOrderNotFound
	}
	if target, _, ok := resolveTarget(order, false); ok {
		existing, err := s.procRepo.GetByLocalOrderID(target.ID)
		if err != nil {
			return nil, none, err
		}
		if existing != nil {
			return nil, none, procurementcontract.ErrFulfillInfoAlreadySubmitted
		}
	}
	target, item, ok := resolveTarget(order, true)
	if !ok {
		if _, _, submitted := resolveTarget(order, false); submitted {
			return nil, none, procurementcontract.ErrFulfillInfoAlreadySubmitted
		}
		return nil, none, procurementcontract.ErrFulfillInfoNotAllowed
	}
	if target.Status != constants.OrderStatusPaid {
		return nil, none, procurementcontract.ErrFulfillInfoNotAllowed
	}
	return target, item, nil
}

// precheckSession 上游支持 session_check 时预检会话，只在上游明确判定「会话无效 / 已订阅」时拦截。
// 客户已付款：上游暂时核验不了（check_unavailable）、报错、超时或限流都放行，由采购提交阶段兜底。
func (s *Service) precheckSession(ctx context.Context, item procurementdomain.LocalOrderItem, submission jsonmap.JSON) error {
	sessionJSON, ok := submission[sessionFormKey].(string)
	if !ok || strings.TrimSpace(sessionJSON) == "" {
		return nil
	}
	// manualform 规范化会对文本做 HTML 转义（引号 → &#34;）；预检要还原成用户粘贴的原始 JSON，
	// 否则上游会把合法会话判为 invalid_session。入库值保持转义形态，与既有人工表单流程一致。
	sessionJSON = html.UnescapeString(sessionJSON)
	connectionID, found, err := s.mappingRepo.FindConnectionID(item.ProductID)
	if err != nil || !found {
		return nil
	}
	upstreamSKUID, found, err := s.skuMapRepo.FindUpstreamSKUID(item.SKUID)
	if err != nil || !found {
		return nil
	}
	connection, err := s.connections.Open(connectionID)
	if err != nil || connection == nil {
		return nil
	}
	capCtx, capCancel := context.WithTimeout(ctx, upstreamProbeTimeout)
	supported := hasCapability(connection.Capabilities(capCtx), capabilitySessionCheck)
	capCancel()
	if !supported {
		return nil
	}
	probeCtx, cancel := context.WithTimeout(ctx, sessionCheckTimeout)
	defer cancel()
	result, err := connection.CheckSession(probeCtx, upstreamSKUID, sessionJSON)
	if err != nil {
		if !errors.Is(err, procurementcontract.ErrCapabilityUnsupported) {
			// 不记录 session 内容，仅记录错误。
			logger.Warnw("procurement_session_check_degraded", "product_id", item.ProductID, "error", err)
		}
		return nil
	}
	if result == nil || result.CanOpen {
		return nil
	}
	if result.ReasonCode == reasonCheckUnavailable {
		logger.Warnw("procurement_session_check_unavailable", "product_id", item.ProductID)
		return nil
	}
	return &procurementcontract.SessionCheckRejectedError{
		ReasonCode: result.ReasonCode, Message: result.Message, Account: result.Account,
	}
}

// GetFulfillProgress 汇总采购单与上游进度，给前端一个统一的阶段视图。
func (s *Service) GetFulfillProgress(ctx context.Context, localOrderID uint) (*procurementcontract.FulfillProgress, error) {
	order, err := s.orderRepo.GetByID(localOrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, procurementcontract.ErrOrderNotFound
	}
	target, _, ok := resolveTarget(order, false)
	if !ok {
		return nil, procurementcontract.ErrFulfillInfoNotAllowed
	}
	proc, err := s.procRepo.GetByLocalOrderID(target.ID)
	if err != nil {
		return nil, err
	}

	switch target.Status {
	case constants.OrderStatusDelivered, constants.OrderStatusCompleted:
		return newProgress(procurementcontract.FulfillPhaseCompleted, true, "success", 0), nil
	case constants.OrderStatusCanceled, constants.OrderStatusRefunded:
		p := newProgress(procurementcontract.FulfillPhaseFailed, true, "cancelled", 0)
		p.MessageKey = procurementcontract.MessageKeyOrderClosed
		return p, nil
	}

	if proc == nil {
		if _, _, pending := resolveTarget(order, true); pending {
			return newProgress(procurementcontract.FulfillPhaseAwaitingInfo, false, "", 0), nil
		}
		return newProgress(procurementcontract.FulfillPhaseSubmitting, false, "", defaultSubmittingPollMs), nil
	}

	switch proc.Status {
	case "fulfilled":
		return newProgress(procurementcontract.FulfillPhaseCompleted, true, "success", 0), nil
	case "pending":
		return newProgress(procurementcontract.FulfillPhaseSubmitting, false, "", defaultSubmittingPollMs), nil
	case "failed":
		// failed 且已排定重试时仍在提交阶段；否则视为终态失败。
		if proc.NextRetryAt != nil {
			return newProgress(procurementcontract.FulfillPhaseSubmitting, false, "", defaultSubmittingPollMs), nil
		}
		return s.failedProgress(ctx, proc, "failed"), nil
	case "rejected":
		return s.failedProgress(ctx, proc, "rejected"), nil
	case "canceled":
		return s.failedProgress(ctx, proc, "cancelled"), nil
	case "accepted":
		return s.acceptedProgress(ctx, proc), nil
	}
	return newProgress(procurementcontract.FulfillPhaseProcessing, false, "", defaultProcessingPollMs), nil
}

func newProgress(phase string, terminal bool, outcome string, pollMs int) *procurementcontract.FulfillProgress {
	percent := 0
	if phase == procurementcontract.FulfillPhaseCompleted {
		percent = 100
	}
	return &procurementcontract.FulfillProgress{
		Phase: phase, Steps: []procurementcontract.ProgressStep{}, Percent: percent,
		Terminal: terminal, Outcome: outcome, PollAfterMs: pollMs,
	}
}

// failedProgress 终态失败。若采购单曾被上游接收（有上游订单号），尝试取一次上游进度，
// 用上游给出的原因（例如扣款被拒）展示给客户；取不到再回落到通用文案。
func (s *Service) failedProgress(ctx context.Context, proc *procurementdomain.Order, outcome string) *procurementcontract.FulfillProgress {
	p := newProgress(procurementcontract.FulfillPhaseFailed, true, outcome, 0)
	p.MessageKey = procurementcontract.MessageKeyProcurementFailed
	if proc.UpstreamOrderID != 0 {
		if upstreamProgress, ok := s.liveProgress(ctx, proc); ok && upstreamProgress.Message != "" {
			p.Message = upstreamProgress.Message
			p.MessageKey = ""
		}
	}
	return p
}

// liveProgress 取（带 2 秒缓存的）上游实时进度；上游不支持或出错返回 false。
func (s *Service) liveProgress(ctx context.Context, proc *procurementdomain.Order) (*procurementcontract.UpstreamProgress, bool) {
	if proc.UpstreamOrderID == 0 || s.connections == nil {
		return nil, false
	}
	if cached, ok := s.cache.Load(proc.ID); ok {
		if entry := cached.(cacheEntry); time.Now().Before(entry.expiresAt) {
			return entry.progress, true
		}
	}
	connection, err := s.connections.Open(proc.ConnectionID)
	if err != nil || connection == nil {
		return nil, false
	}
	probeCtx, cancel := context.WithTimeout(ctx, upstreamProbeTimeout)
	defer cancel()
	if !hasCapability(connection.Capabilities(probeCtx), capabilityOrderProgress) {
		return nil, false
	}
	upstreamProgress, err := connection.GetOrderProgress(probeCtx, proc.UpstreamOrderID)
	if err != nil || upstreamProgress == nil {
		if err != nil && !errors.Is(err, procurementcontract.ErrCapabilityUnsupported) {
			logger.Warnw("procurement_progress_degraded", "procurement_order_id", proc.ID, "error", err)
		}
		return nil, false
	}
	s.cache.Store(proc.ID, cacheEntry{progress: upstreamProgress, expiresAt: time.Now().Add(progressCacheTTL)})
	return upstreamProgress, true
}

// acceptedProgress 采购单已被上游接收：尽量取上游实时进度，不支持或失败时降级为无步骤的 processing。
func (s *Service) acceptedProgress(ctx context.Context, proc *procurementdomain.Order) *procurementcontract.FulfillProgress {
	upstreamProgress, ok := s.liveProgress(ctx, proc)
	if !ok {
		return newProgress(procurementcontract.FulfillPhaseProcessing, false, "", defaultProcessingPollMs)
	}
	p := &procurementcontract.FulfillProgress{
		Phase: procurementcontract.FulfillPhaseProcessing, Steps: upstreamProgress.Steps, Step: upstreamProgress.Step,
		Percent: upstreamProgress.Percent, Terminal: upstreamProgress.Terminal, Outcome: upstreamProgress.Outcome,
		Message: upstreamProgress.Message, PollAfterMs: upstreamProgress.PollAfterMs,
	}
	if p.Steps == nil {
		p.Steps = []procurementcontract.ProgressStep{}
	}
	if p.PollAfterMs <= 0 && !upstreamProgress.Terminal {
		p.PollAfterMs = defaultProcessingPollMs
	}
	if upstreamProgress.Terminal {
		p.PollAfterMs = 0
		if upstreamProgress.Outcome == "success" {
			p.Phase = procurementcontract.FulfillPhaseCompleted
			p.Percent = 100
		} else {
			p.Phase = procurementcontract.FulfillPhaseFailed
			if p.Message == "" {
				p.MessageKey = procurementcontract.MessageKeyProcurementFailed
			}
		}
	}
	return p
}
