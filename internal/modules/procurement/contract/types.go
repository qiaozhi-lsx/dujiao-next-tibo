package contract

import (
	"context"
	"time"

	procurementdomain "github.com/dujiao-next/internal/modules/procurement/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

type ListFilter struct {
	ConnectionID    uint
	Status          string
	LocalOrderNo    string
	UpstreamOrderNo string
	CreatedFrom     *time.Time
	CreatedTo       *time.Time
	Page            int
	PageSize        int
}

type Fulfillment struct {
	Type         string
	Status       string
	Payload      string
	DeliveryData jsonmap.JSON
	DeliveredAt  *time.Time
}

type CreateOrderRequest struct {
	SKUID             uint
	Quantity          int
	ManualFormData    jsonmap.JSON
	DownstreamOrderNo string
	TraceID           string
	CallbackURL       string
}

type CreateOrderResult struct {
	OK           bool
	OrderID      uint
	OrderNo      string
	Status       string
	Amount       string
	Currency     string
	ErrorCode    string
	ErrorMessage string
}

type UpstreamOrder struct {
	OrderID        uint
	OrderNo        string
	Status         string
	Amount         string
	RefundedAmount string
	Currency       string
	Fulfillment    *Fulfillment
	RefundRecords  []jsonmap.JSON
}

// SessionAccount 会话预检返回的账号摘要。
type SessionAccount struct {
	EmailMasked string `json:"email_masked"`
	PlanType    string `json:"plan_type"`
	IsActive    bool   `json:"is_active"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

// SessionCheckResult 上游会话预检结果。
type SessionCheckResult struct {
	Valid      bool
	CanOpen    bool
	ReasonCode string
	Message    string
	Account    *SessionAccount
}

// SessionCheckRejectedError 预检判定不可开通时返回，携带给前端展示的详情。
type SessionCheckRejectedError struct {
	ReasonCode string          `json:"reason_code"`
	Message    string          `json:"message"`
	Account    *SessionAccount `json:"account"`
}

func (e *SessionCheckRejectedError) Error() string {
	return "session check rejected: " + e.ReasonCode
}

// ProgressStep 进度步骤定义。
type ProgressStep struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// ProgressCurrentStep 当前所处步骤。
type ProgressCurrentStep struct {
	Index int    `json:"index"`
	Total int    `json:"total"`
	Key   string `json:"key"`
	Label string `json:"label"`
}

// UpstreamProgress 上游订单开通进度。
type UpstreamProgress struct {
	Status      string
	Steps       []ProgressStep
	Step        *ProgressCurrentStep
	Percent     int
	Terminal    bool
	Outcome     string
	Message     string
	PollAfterMs int
}

// 开通进度阶段（FulfillProgress.Phase）。
const (
	FulfillPhaseAwaitingInfo = "awaiting_info"
	FulfillPhaseSubmitting   = "submitting"
	FulfillPhaseProcessing   = "processing"
	FulfillPhaseFailed       = "failed"
	FulfillPhaseCompleted    = "completed"
)

// FulfillProgress 是 fulfill-progress 接口的响应形状。
// Steps/Step/Percent/Terminal/Outcome/Message/PollAfterMs 仅在 processing 阶段或终态由上游进度填充。
// MessageKey 非空时表示 Message 需要由传输层按请求语言本地化。
type FulfillProgress struct {
	Phase       string               `json:"phase"`
	Steps       []ProgressStep       `json:"steps"`
	Step        *ProgressCurrentStep `json:"step"`
	Percent     int                  `json:"percent"`
	Terminal    bool                 `json:"terminal"`
	Outcome     string               `json:"outcome"`
	Message     string               `json:"message"`
	MessageKey  string               `json:"-"`
	PollAfterMs int                  `json:"poll_after_ms"`
}

type UpstreamConnection interface {
	CallbackURL() string
	RetryMax() int
	RetryIntervals() string
	CreateOrder(ctx context.Context, request CreateOrderRequest) (*CreateOrderResult, error)
	GetOrder(ctx context.Context, orderID uint) (*UpstreamOrder, error)
	CancelOrder(ctx context.Context, orderID uint) error
	// Capabilities 返回上游声明的可选能力（带缓存）；探测失败返回空。
	Capabilities(ctx context.Context) []string
	// CheckSession 会话预检；不支持返回 ErrCapabilityUnsupported。
	CheckSession(ctx context.Context, skuID uint, sessionJSON string) (*SessionCheckResult, error)
	// GetOrderProgress 查询上游开通进度；不支持返回 ErrCapabilityUnsupported。
	GetOrderProgress(ctx context.Context, upstreamOrderID uint) (*UpstreamProgress, error)
}

// UseCase 是支付、HTTP、上游回调与 Worker 共享的正式采购应用契约。
type UseCase interface {
	CreateForOrder(orderID uint) error
	SubmitToUpstream(procurementOrderID uint) error
	PollUpstreamStatus(procurementOrderID uint) error
	SyncAcceptedOrders()
	HandleUpstreamCallback(procurementOrderID uint, upstreamStatus string, fulfillment *Fulfillment) error
	GetByID(id uint) (*procurementdomain.Order, error)
	GetByLocalOrderNo(localOrderNo string) (*procurementdomain.Order, error)
	List(filter ListFilter) ([]procurementdomain.Order, int64, error)
	StatsByStatus(filter ListFilter) (map[string]int64, error)
	FillParentOrderNo(order *procurementdomain.Order)
	RetryManual(id uint) error
	CancelManual(id uint) error
	// SubmitFulfillInfo 支付后提交延后填写的开通信息（调用方已完成订单归属鉴权）。
	SubmitFulfillInfo(ctx context.Context, localOrderID uint, formData jsonmap.JSON) error
	// GetFulfillProgress 查询开通进度（调用方已完成订单归属鉴权）。
	GetFulfillProgress(ctx context.Context, localOrderID uint) (*FulfillProgress, error)
}
