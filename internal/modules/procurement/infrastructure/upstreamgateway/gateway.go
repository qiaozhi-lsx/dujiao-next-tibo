package upstreamgateway

import (
	"context"
	"errors"
	"sync"
	"time"

	procurementcontract "github.com/dujiao-next/internal/modules/procurement/contract"
	siteconnectiondomain "github.com/dujiao-next/internal/modules/siteconnection/domain"
	"github.com/dujiao-next/internal/upstream"
)

type Provider interface {
	GetByID(id uint) (*siteconnectiondomain.Connection, error)
	GetAdapter(connection *siteconnectiondomain.Connection) (upstream.Adapter, error)
}

// capabilityTTL 上游能力缓存时间：能力随上游升级才会变化，几分钟的滞后可接受。
const capabilityTTL = 5 * time.Minute

type capabilityEntry struct {
	caps      []string
	expiresAt time.Time
}

type Gateway struct {
	provider Provider
	mu       sync.Mutex
	caps     map[uint]capabilityEntry
}

var _ procurementcontract.ConnectionProvider = (*Gateway)(nil)

func New(provider Provider) *Gateway {
	if provider == nil {
		panic("procurement upstream gateway: provider is nil")
	}
	return &Gateway{provider: provider, caps: make(map[uint]capabilityEntry)}
}

func (g *Gateway) Open(connectionID uint) (procurementcontract.UpstreamConnection, error) {
	connection, err := g.provider.GetByID(connectionID)
	if err != nil || connection == nil {
		return nil, err
	}
	adapter, err := g.provider.GetAdapter(connection)
	if err != nil {
		return nil, err
	}
	return &session{connection: connection, adapter: adapter, gateway: g}, nil
}

func (g *Gateway) cachedCapabilities(connectionID uint) ([]string, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	entry, ok := g.caps[connectionID]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.caps, true
}

func (g *Gateway) storeCapabilities(connectionID uint, caps []string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.caps[connectionID] = capabilityEntry{caps: caps, expiresAt: time.Now().Add(capabilityTTL)}
}

type session struct {
	connection *siteconnectiondomain.Connection
	adapter    upstream.Adapter
	gateway    *Gateway
}

var _ procurementcontract.UpstreamConnection = (*session)(nil)

func (s *session) CallbackURL() string    { return s.connection.CallbackURL }
func (s *session) RetryMax() int          { return s.connection.RetryMax }
func (s *session) RetryIntervals() string { return s.connection.RetryIntervals }

func (s *session) CreateOrder(ctx context.Context, request procurementcontract.CreateOrderRequest) (*procurementcontract.CreateOrderResult, error) {
	result, err := s.adapter.CreateOrder(ctx, upstream.CreateUpstreamOrderReq{
		SKUID: request.SKUID, Quantity: request.Quantity, ManualFormData: request.ManualFormData,
		DownstreamOrderNo: request.DownstreamOrderNo, TraceID: request.TraceID, CallbackURL: request.CallbackURL,
	})
	if err != nil {
		return nil, err
	}
	return &procurementcontract.CreateOrderResult{
		OK: result.OK, OrderID: result.OrderID, OrderNo: result.OrderNo,
		Status: result.Status, Amount: result.Amount, Currency: result.Currency,
		ErrorCode: result.ErrorCode, ErrorMessage: result.ErrorMessage,
	}, nil
}

func (s *session) GetOrder(ctx context.Context, orderID uint) (*procurementcontract.UpstreamOrder, error) {
	detail, err := s.adapter.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	result := &procurementcontract.UpstreamOrder{
		OrderID: detail.OrderID, OrderNo: detail.OrderNo, Status: detail.Status,
		Amount: detail.Amount, RefundedAmount: detail.RefundedAmount,
		Currency: detail.Currency, RefundRecords: detail.RefundRecords,
	}
	if detail.Fulfillment != nil {
		result.Fulfillment = fromUpstreamFulfillment(detail.Fulfillment)
	}
	return result, nil
}

func (s *session) CancelOrder(ctx context.Context, orderID uint) error {
	return s.adapter.CancelOrder(ctx, orderID)
}

// Capabilities 通过（缓存的）ping 获取上游可选能力；ping 失败时返回空并短期缓存，避免每次轮询都打上游。
func (s *session) Capabilities(ctx context.Context) []string {
	if caps, ok := s.gateway.cachedCapabilities(s.connection.ID); ok {
		return caps
	}
	result, err := s.adapter.Ping(ctx)
	if err != nil || result == nil {
		s.gateway.storeCapabilities(s.connection.ID, nil)
		return nil
	}
	s.gateway.storeCapabilities(s.connection.ID, result.Capabilities)
	return result.Capabilities
}

func (s *session) CheckSession(ctx context.Context, skuID uint, sessionJSON string) (*procurementcontract.SessionCheckResult, error) {
	result, err := s.adapter.CheckSession(ctx, skuID, sessionJSON)
	if err != nil {
		return nil, mapCapabilityError(err)
	}
	out := &procurementcontract.SessionCheckResult{
		Valid: result.Valid, CanOpen: result.CanOpen, Message: result.Message,
	}
	if result.ReasonCode != nil {
		out.ReasonCode = *result.ReasonCode
	}
	if result.Account != nil {
		out.Account = &procurementcontract.SessionAccount{
			EmailMasked: result.Account.EmailMasked, PlanType: result.Account.PlanType, IsActive: result.Account.IsActive,
		}
		if result.Account.ExpiresAt != nil {
			out.Account.ExpiresAt = *result.Account.ExpiresAt
		}
	}
	return out, nil
}

func (s *session) GetOrderProgress(ctx context.Context, upstreamOrderID uint) (*procurementcontract.UpstreamProgress, error) {
	result, err := s.adapter.GetOrderProgress(ctx, upstreamOrderID)
	if err != nil {
		return nil, mapCapabilityError(err)
	}
	out := &procurementcontract.UpstreamProgress{
		Status: result.Status, Percent: result.Percent, Terminal: result.Terminal,
		Steps: make([]procurementcontract.ProgressStep, 0, len(result.Steps)),
	}
	for _, step := range result.Steps {
		out.Steps = append(out.Steps, procurementcontract.ProgressStep{Key: step.Key, Label: step.Label})
	}
	if result.Step != nil {
		out.Step = &procurementcontract.ProgressCurrentStep{
			Index: result.Step.Index, Total: result.Step.Total, Key: result.Step.Key, Label: result.Step.Label,
		}
	}
	if result.Outcome != nil {
		out.Outcome = *result.Outcome
	}
	if result.Message != nil {
		out.Message = *result.Message
	}
	if result.PollAfterMs != nil {
		out.PollAfterMs = *result.PollAfterMs
	}
	return out, nil
}

func mapCapabilityError(err error) error {
	if errors.Is(err, upstream.ErrCapabilityUnsupported) {
		return procurementcontract.ErrCapabilityUnsupported
	}
	return err
}

func fromUpstreamFulfillment(value *upstream.UpstreamFulfillment) *procurementcontract.Fulfillment {
	if value == nil {
		return nil
	}
	return &procurementcontract.Fulfillment{
		Type: value.Type, Status: value.Status, Payload: value.Payload,
		DeliveryData: value.DeliveryData, DeliveredAt: value.DeliveredAt,
	}
}
