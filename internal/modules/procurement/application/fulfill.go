package application

import (
	"context"
	"errors"
	"html"
	"strings"
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
	defaultSubmittingPollMs = 2000
	defaultProcessingPollMs = 3000
)

// 需要由传输层本地化的提示 key。
const (
	MessageKeyProcurementFailed = "error.fulfill_procurement_failed"
	MessageKeyOrderClosed       = "error.fulfill_order_closed"
)

type progressCacheEntry struct {
	progress  *procurementcontract.UpstreamProgress
	expiresAt time.Time
}

// schemaHasFields 判断表单 schema 快照是否包含至少一个字段。
func schemaHasFields(schema jsonmap.JSON) bool {
	if len(schema) == 0 {
		return false
	}
	switch fields := schema["fields"].(type) {
	case []interface{}:
		return len(fields) > 0
	case []jsonmap.JSON:
		return len(fields) > 0
	case []map[string]interface{}:
		return len(fields) > 0
	}
	return false
}

// isFormPending 订单项是上游交付、带表单快照但尚未提交表单值。
func isFormPending(item procurementdomain.LocalOrderItem) bool {
	return strings.TrimSpace(item.FulfillmentType) == constants.FulfillmentTypeUpstream &&
		schemaHasFields(item.ManualFormSchemaSnapshotJSON) &&
		len(item.ManualFormSubmissionJSON) == 0
}

func hasCapability(caps []string, name string) bool {
	for _, c := range caps {
		if c == name {
			return true
		}
	}
	return false
}

func findFormPendingItem(order *procurementdomain.LocalOrder) (procurementdomain.LocalOrderItem, bool) {
	for _, item := range order.Items {
		if isFormPending(item) {
			return item, true
		}
	}
	return procurementdomain.LocalOrderItem{}, false
}

// SubmitFulfillInfo 支付后提交开通信息：校验 → 上游预检（如支持）→ 保存 → 创建采购单。
// 调用方负责订单归属鉴权。预检阶段与上游的 HTTP 交互不在任何数据库事务内。
func (s *Service) SubmitFulfillInfo(ctx context.Context, localOrderID uint, formData jsonmap.JSON) error {
	order, err := s.orderRepo.GetByID(localOrderID)
	if err != nil {
		return err
	}
	if order == nil {
		return procurementcontract.ErrOrderNotFound
	}
	if err := s.ensureFulfillSubmittable(order); err != nil {
		return err
	}
	item, _ := findFormPendingItem(order)

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

	mu := &s.fulfillMu[localOrderID%uint(len(s.fulfillMu))]
	mu.Lock()
	defer mu.Unlock()

	// 加锁后重读，拦截并发的重复提交。
	order, err = s.orderRepo.GetByID(localOrderID)
	if err != nil {
		return err
	}
	if order == nil {
		return procurementcontract.ErrOrderNotFound
	}
	if err := s.ensureFulfillSubmittable(order); err != nil {
		return err
	}
	item, _ = findFormPendingItem(order)

	if err := s.orderRepo.SaveManualFormSubmission(item.ID, normalized); err != nil {
		return err
	}
	if err := s.CreateForOrder(localOrderID); err != nil {
		if errors.Is(err, procurementcontract.ErrExists) {
			return procurementcontract.ErrFulfillInfoAlreadySubmitted
		}
		return err
	}
	return nil
}

// ensureFulfillSubmittable 订单必须是 paid 且存在待填表单、尚无采购单。
func (s *Service) ensureFulfillSubmittable(order *procurementdomain.LocalOrder) error {
	if order.ParentID == nil && len(order.Children) > 0 {
		return procurementcontract.ErrFulfillInfoNotAllowed
	}
	existing, err := s.procRepo.GetByLocalOrderID(order.ID)
	if err != nil {
		return err
	}
	if existing != nil {
		return procurementcontract.ErrFulfillInfoAlreadySubmitted
	}
	if _, ok := findFormPendingItem(order); !ok {
		for _, item := range order.Items {
			if strings.TrimSpace(item.FulfillmentType) == constants.FulfillmentTypeUpstream &&
				schemaHasFields(item.ManualFormSchemaSnapshotJSON) {
				return procurementcontract.ErrFulfillInfoAlreadySubmitted
			}
		}
		return procurementcontract.ErrFulfillInfoNotAllowed
	}
	if order.Status != constants.OrderStatusPaid {
		return procurementcontract.ErrFulfillInfoNotAllowed
	}
	return nil
}

// precheckSession 上游支持 session_check 时预检会话；不可开通返回 SessionCheckRejectedError。
// 上游不支持、不可达或限流时降级跳过（采购提交阶段上游仍会校验）。
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
	probeCtx, cancel := context.WithTimeout(ctx, upstreamProbeTimeout)
	defer cancel()
	if !hasCapability(connection.Capabilities(probeCtx), capabilitySessionCheck) {
		return nil
	}
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
	proc, err := s.procRepo.GetByLocalOrderID(localOrderID)
	if err != nil {
		return nil, err
	}

	switch order.Status {
	case constants.OrderStatusDelivered, constants.OrderStatusCompleted:
		return newFulfillProgress(procurementcontract.FulfillPhaseCompleted, true, "success", 0), nil
	case constants.OrderStatusCanceled, constants.OrderStatusRefunded:
		p := newFulfillProgress(procurementcontract.FulfillPhaseFailed, true, "cancelled", 0)
		p.MessageKey = MessageKeyOrderClosed
		return p, nil
	}

	if proc == nil {
		if _, pending := findFormPendingItem(order); pending {
			return newFulfillProgress(procurementcontract.FulfillPhaseAwaitingInfo, false, "", 0), nil
		}
		return newFulfillProgress(procurementcontract.FulfillPhaseSubmitting, false, "", defaultSubmittingPollMs), nil
	}

	switch proc.Status {
	case "fulfilled":
		return newFulfillProgress(procurementcontract.FulfillPhaseCompleted, true, "success", 0), nil
	case "pending":
		return newFulfillProgress(procurementcontract.FulfillPhaseSubmitting, false, "", defaultSubmittingPollMs), nil
	case "failed":
		// failed 且已排定重试时仍在提交阶段；否则视为终态失败。
		if proc.NextRetryAt != nil {
			return newFulfillProgress(procurementcontract.FulfillPhaseSubmitting, false, "", defaultSubmittingPollMs), nil
		}
		return failedProgress("failed"), nil
	case "rejected":
		return failedProgress("rejected"), nil
	case "canceled":
		return failedProgress("cancelled"), nil
	case "accepted":
		return s.acceptedProgress(ctx, proc), nil
	}
	return newFulfillProgress(procurementcontract.FulfillPhaseProcessing, false, "", defaultProcessingPollMs), nil
}

func newFulfillProgress(phase string, terminal bool, outcome string, pollMs int) *procurementcontract.FulfillProgress {
	percent := 0
	if phase == procurementcontract.FulfillPhaseCompleted {
		percent = 100
	}
	return &procurementcontract.FulfillProgress{
		Phase: phase, Steps: []procurementcontract.ProgressStep{}, Percent: percent,
		Terminal: terminal, Outcome: outcome, PollAfterMs: pollMs,
	}
}

func failedProgress(outcome string) *procurementcontract.FulfillProgress {
	p := newFulfillProgress(procurementcontract.FulfillPhaseFailed, true, outcome, 0)
	p.MessageKey = MessageKeyProcurementFailed
	return p
}

// acceptedProgress 采购单已被上游接收：尽量取上游实时进度，不支持或失败时降级为无步骤的 processing。
func (s *Service) acceptedProgress(ctx context.Context, proc *procurementdomain.Order) *procurementcontract.FulfillProgress {
	fallback := newFulfillProgress(procurementcontract.FulfillPhaseProcessing, false, "", defaultProcessingPollMs)
	if proc.UpstreamOrderID == 0 || s.connections == nil {
		return fallback
	}
	if cached, ok := s.progressCache.Load(proc.ID); ok {
		if entry := cached.(progressCacheEntry); time.Now().Before(entry.expiresAt) {
			return mapUpstreamProgress(entry.progress)
		}
	}
	connection, err := s.connections.Open(proc.ConnectionID)
	if err != nil || connection == nil {
		return fallback
	}
	probeCtx, cancel := context.WithTimeout(ctx, upstreamProbeTimeout)
	defer cancel()
	if !hasCapability(connection.Capabilities(probeCtx), capabilityOrderProgress) {
		return fallback
	}
	upstreamProgress, err := connection.GetOrderProgress(probeCtx, proc.UpstreamOrderID)
	if err != nil || upstreamProgress == nil {
		if err != nil && !errors.Is(err, procurementcontract.ErrCapabilityUnsupported) {
			logger.Warnw("procurement_progress_degraded", "procurement_order_id", proc.ID, "error", err)
		}
		return fallback
	}
	s.progressCache.Store(proc.ID, progressCacheEntry{progress: upstreamProgress, expiresAt: time.Now().Add(progressCacheTTL)})
	return mapUpstreamProgress(upstreamProgress)
}

func mapUpstreamProgress(up *procurementcontract.UpstreamProgress) *procurementcontract.FulfillProgress {
	p := &procurementcontract.FulfillProgress{
		Phase: procurementcontract.FulfillPhaseProcessing, Steps: up.Steps, Step: up.Step,
		Percent: up.Percent, Terminal: up.Terminal, Outcome: up.Outcome, Message: up.Message,
		PollAfterMs: up.PollAfterMs,
	}
	if p.Steps == nil {
		p.Steps = []procurementcontract.ProgressStep{}
	}
	if p.PollAfterMs <= 0 && !up.Terminal {
		p.PollAfterMs = defaultProcessingPollMs
	}
	if up.Terminal {
		p.PollAfterMs = 0
		if up.Outcome == "success" {
			p.Phase = procurementcontract.FulfillPhaseCompleted
			p.Percent = 100
		} else {
			p.Phase = procurementcontract.FulfillPhaseFailed
			if p.Message == "" {
				p.MessageKey = MessageKeyProcurementFailed
			}
		}
	}
	return p
}
