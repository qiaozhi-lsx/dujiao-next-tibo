package procurement_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mappingdomain "github.com/dujiao-next/internal/modules/catalog/mapping/domain"
	"github.com/dujiao-next/internal/modules/catalog/product/manualform"
	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
	procurementapp "github.com/dujiao-next/internal/modules/procurement/application"
	procurementcontract "github.com/dujiao-next/internal/modules/procurement/contract"
	siteconnectionapp "github.com/dujiao-next/internal/modules/siteconnection/application"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"gorm.io/gorm"

	"github.com/dujiao-next/internal/constants"
)

// ── CreateForOrder tests ──

func TestCreateForOrder_SkipsNonUpstreamItems(t *testing.T) {
	db := setupProcurementTestDB(t)

	order := createProcTestOrder(t, db, "PROC-SKIP-001", constants.OrderStatusPaid, constants.FulfillmentTypeAuto)

	connSvc := newTestSiteConnectionService(db, "test-key", t.TempDir())
	svc := newTestProcurementService(db, connSvc)

	if err := svc.CreateForOrder(order.ID); err != nil {
		t.Fatalf("CreateForOrder: %v", err)
	}

	// 验证没有创建采购单
	var count int64
	db.Model(&ProcurementOrder{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no procurement orders for auto fulfillment, got %d", count)
	}
}

func TestCreateForOrder_IdempotentSkipsDuplicate(t *testing.T) {
	db := setupProcurementTestDB(t)

	order := createProcTestOrder(t, db, "PROC-DUP-001", constants.OrderStatusPaid, constants.FulfillmentTypeUpstream)
	pm := &mappingdomain.Mapping{ConnectionID: 1, LocalProductID: 1, UpstreamProductID: 101, IsActive: true}
	db.Create(pm)

	connSvc := newTestSiteConnectionService(db, "test-key", t.TempDir())
	svc := newTestProcurementService(db, connSvc)

	// 第一次创建成功
	if err := svc.CreateForOrder(order.ID); err != nil {
		t.Fatalf("first CreateForOrder: %v", err)
	}

	// 第二次应该返回 ErrExists
	err := svc.CreateForOrder(order.ID)
	if err != ErrExists {
		t.Errorf("expected ErrExists on duplicate, got: %v", err)
	}
}

// ── 延后填写开通信息（fulfill after payment）──

type fulfillFixture struct {
	db      *gorm.DB
	svc     *procurementapp.Service
	order   *orderdomain.Order
	calls   *[]string
	checkOK *bool
	caps    *[]string
	server  *httptest.Server
	connID  uint
}

// newFulfillFixture 构造一个 paid 订单：单个上游交付项，带 session_json 必填表单快照、提交值为空。
func newFulfillFixture(t *testing.T, orderNo string, caps []string) *fulfillFixture {
	t.Helper()
	db := setupProcurementTestDB(t)
	order := createProcTestOrder(t, db, orderNo, constants.OrderStatusPaid, constants.FulfillmentTypeUpstream)
	schema := jsonmap.JSON{"fields": []interface{}{map[string]interface{}{
		"key": "session_json", "type": "textarea", "required": true,
		"label": map[string]interface{}{"zh-CN": "会话"},
	}}}
	if err := db.Model(&orderdomain.OrderItem{}).Where("order_id = ?", order.ID).
		Update("manual_form_schema_snapshot_json", schema).Error; err != nil {
		t.Fatalf("set schema snapshot: %v", err)
	}
	f := &fulfillFixture{db: db, order: order, calls: &[]string{}, checkOK: new(bool), caps: &caps}
	*f.checkOK = true
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*f.calls = append(*f.calls, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/upstream/ping":
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "site_name": "up", "capabilities": *f.caps})
		case "/api/v1/upstream/session/check":
			if *f.checkOK {
				_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "valid": true, "can_open": true, "reason_code": nil, "message": "ok", "account": nil})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok": true, "valid": true, "can_open": false, "reason_code": "already_subscribed", "message": "已订阅",
				"account": map[string]any{"email_masked": "a***@x.com", "plan_type": "plus", "is_active": true, "expires_at": nil},
			})
		case "/api/v1/upstream/orders/999/progress":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok": true, "order_id": 999, "status": "processing",
				"steps":   []map[string]string{{"key": "create", "label": "提交订单"}, {"key": "charging", "label": "扣款确认"}},
				"step":    map[string]any{"index": 2, "total": 2, "key": "charging", "label": "扣款确认"},
				"percent": 50, "terminal": false, "outcome": nil, "message": "处理中", "poll_after_ms": 1500,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(f.server.Close)

	connSvc := newTestSiteConnectionService(db, "test-key", t.TempDir())
	conn, err := connSvc.Create(siteconnectionapp.CreateInput{
		Name: "up", BaseURL: f.server.URL, ApiKey: "key", ApiSecret: "secret", Protocol: constants.ConnectionProtocolDujiaoNext,
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}
	f.connID = conn.ID
	pm := &mappingdomain.Mapping{ConnectionID: conn.ID, LocalProductID: 1, UpstreamProductID: 101, IsActive: true}
	db.Create(pm)
	db.Create(&mappingdomain.SKUMapping{ProductMappingID: pm.ID, LocalSKUID: 1, UpstreamSKUID: 201, UpstreamIsActive: true})
	f.svc = newTestProcurementService(db, connSvc)
	return f
}

func (f *fulfillFixture) procCount() int64 {
	var n int64
	f.db.Model(&ProcurementOrder{}).Where("local_order_id = ?", f.order.ID).Count(&n)
	return n
}

func (f *fulfillFixture) submission() jsonmap.JSON {
	var item orderdomain.OrderItem
	f.db.Where("order_id = ?", f.order.ID).First(&item)
	return item.ManualFormSubmissionJSON
}

func (f *fulfillFixture) called(path string) bool {
	for _, c := range *f.calls {
		if strings.HasSuffix(c, path) {
			return true
		}
	}
	return false
}

func TestCreateForOrder_SkipsFormPendingItem(t *testing.T) {
	f := newFulfillFixture(t, "FULFILL-SKIP-001", nil)
	if err := f.svc.CreateForOrder(f.order.ID); err != nil {
		t.Fatalf("CreateForOrder: %v", err)
	}
	if f.procCount() != 0 {
		t.Fatalf("expected no procurement order for form-pending item, got %d", f.procCount())
	}
	var o orderdomain.Order
	f.db.First(&o, f.order.ID)
	if o.Status != constants.OrderStatusPaid {
		t.Fatalf("order must stay paid, got %q", o.Status)
	}
}

func TestSubmitFulfillInfo_HappyPathWithPrecheck(t *testing.T) {
	f := newFulfillFixture(t, "FULFILL-OK-001", []string{"session_check", "order_progress"})
	if err := f.svc.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": `{"accessToken":"x"}`}); err != nil {
		t.Fatalf("SubmitFulfillInfo: %v", err)
	}
	if !f.called("/session/check") {
		t.Fatal("expected upstream session check to be called")
	}
	// manualform 规范化会 HTML 转义文本（与既有人工表单流程一致）。
	if f.submission()["session_json"] != `{&#34;accessToken&#34;:&#34;x&#34;}` {
		t.Fatalf("submission not saved: %#v", f.submission())
	}
	if f.procCount() != 1 {
		t.Fatalf("expected 1 procurement order, got %d", f.procCount())
	}
}

func TestSubmitFulfillInfo_PrecheckRejectSavesNothing(t *testing.T) {
	f := newFulfillFixture(t, "FULFILL-REJECT-001", []string{"session_check"})
	*f.checkOK = false
	err := f.svc.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": "abc"})
	var rejected *procurementcontract.SessionCheckRejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf("expected SessionCheckRejectedError, got %v", err)
	}
	if rejected.ReasonCode != "already_subscribed" || rejected.Account == nil || rejected.Account.EmailMasked != "a***@x.com" {
		t.Fatalf("unexpected rejection payload: %#v", rejected)
	}
	if len(f.submission()) != 0 || f.procCount() != 0 {
		t.Fatalf("nothing must be saved on rejection: submission=%#v procs=%d", f.submission(), f.procCount())
	}
}

func TestSubmitFulfillInfo_SkipsPrecheckWithoutCapability(t *testing.T) {
	f := newFulfillFixture(t, "FULFILL-NOCAP-001", nil)
	if err := f.svc.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": "abc"}); err != nil {
		t.Fatalf("SubmitFulfillInfo: %v", err)
	}
	if f.called("/session/check") {
		t.Fatal("session check must not be called without the capability")
	}
	if f.procCount() != 1 {
		t.Fatalf("expected procurement order, got %d", f.procCount())
	}
}

func TestSubmitFulfillInfo_ValidationAndConflicts(t *testing.T) {
	f := newFulfillFixture(t, "FULFILL-DUP-001", nil)
	// 缺必填字段 → 表单校验错误，且未保存。
	if err := f.svc.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{}); !errors.Is(err, manualform.ErrRequiredMissing) {
		t.Fatalf("expected required-missing error, got %v", err)
	}
	if err := f.svc.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": "abc"}); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	if err := f.svc.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": "def"}); !errors.Is(err, procurementcontract.ErrFulfillInfoAlreadySubmitted) {
		t.Fatalf("expected already-submitted conflict, got %v", err)
	}
	if f.submission()["session_json"] != "abc" {
		t.Fatalf("second submit must not overwrite: %#v", f.submission())
	}

	// 未支付订单不允许提交。
	g := newFulfillFixture(t, "FULFILL-UNPAID-001", nil)
	g.db.Model(&orderdomain.Order{}).Where("id = ?", g.order.ID).Update("status", constants.OrderStatusPendingPayment)
	if err := g.svc.SubmitFulfillInfo(context.Background(), g.order.ID, jsonmap.JSON{"session_json": "abc"}); !errors.Is(err, procurementcontract.ErrFulfillInfoNotAllowed) {
		t.Fatalf("expected not-allowed for unpaid order, got %v", err)
	}
}

func TestGetFulfillProgress_Phases(t *testing.T) {
	ctx := context.Background()
	f := newFulfillFixture(t, "FULFILL-PROG-001", []string{"order_progress"})

	p, err := f.svc.GetFulfillProgress(ctx, f.order.ID)
	if err != nil || p.Phase != procurementcontract.FulfillPhaseAwaitingInfo {
		t.Fatalf("expected awaiting_info, got %#v err=%v", p, err)
	}

	if err := f.svc.SubmitFulfillInfo(ctx, f.order.ID, jsonmap.JSON{"session_json": "abc"}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	p, _ = f.svc.GetFulfillProgress(ctx, f.order.ID)
	if p.Phase != procurementcontract.FulfillPhaseSubmitting {
		t.Fatalf("expected submitting, got %q", p.Phase)
	}

	// accepted + order_progress：透传上游步骤。
	f.db.Model(&ProcurementOrder{}).Where("local_order_id = ?", f.order.ID).
		Updates(map[string]any{"status": "accepted", "upstream_order_id": 999})
	p, _ = f.svc.GetFulfillProgress(ctx, f.order.ID)
	if p.Phase != procurementcontract.FulfillPhaseProcessing || len(p.Steps) != 2 || p.Step == nil || p.Step.Key != "charging" ||
		p.Percent != 50 || p.PollAfterMs != 1500 || p.Message != "处理中" {
		t.Fatalf("unexpected live progress: %#v", p)
	}

	// 无 order_progress 能力：降级为无步骤的 processing。
	g := newFulfillFixture(t, "FULFILL-PROG-002", nil)
	g.svc.SubmitFulfillInfo(ctx, g.order.ID, jsonmap.JSON{"session_json": "abc"})
	g.db.Model(&ProcurementOrder{}).Where("local_order_id = ?", g.order.ID).
		Updates(map[string]any{"status": "accepted", "upstream_order_id": 999})
	p, _ = g.svc.GetFulfillProgress(ctx, g.order.ID)
	if p.Phase != procurementcontract.FulfillPhaseProcessing || len(p.Steps) != 0 {
		t.Fatalf("expected degraded processing, got %#v", p)
	}

	// rejected → failed；fulfilled → completed。
	g.db.Model(&ProcurementOrder{}).Where("local_order_id = ?", g.order.ID).Update("status", "rejected")
	p, _ = g.svc.GetFulfillProgress(ctx, g.order.ID)
	if p.Phase != procurementcontract.FulfillPhaseFailed || !p.Terminal || p.MessageKey == "" {
		t.Fatalf("expected failed, got %#v", p)
	}
	g.db.Model(&ProcurementOrder{}).Where("local_order_id = ?", g.order.ID).Update("status", "fulfilled")
	p, _ = g.svc.GetFulfillProgress(ctx, g.order.ID)
	if p.Phase != procurementcontract.FulfillPhaseCompleted || !p.Terminal {
		t.Fatalf("expected completed, got %#v", p)
	}
}
