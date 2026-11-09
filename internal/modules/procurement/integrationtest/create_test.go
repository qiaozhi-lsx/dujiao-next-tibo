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
	mappinggormstore "github.com/dujiao-next/internal/modules/catalog/mapping/infrastructure/gormstore"
	"github.com/dujiao-next/internal/modules/catalog/product/manualform"
	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
	ordergormstore "github.com/dujiao-next/internal/modules/order/infrastructure/gormstore"
	procurementapp "github.com/dujiao-next/internal/modules/procurement/application"
	procurementfulfill "github.com/dujiao-next/internal/modules/procurement/application/fulfill"
	procurementcontract "github.com/dujiao-next/internal/modules/procurement/contract"
	procurementgormstore "github.com/dujiao-next/internal/modules/procurement/infrastructure/gormstore"
	procurementmapping "github.com/dujiao-next/internal/modules/procurement/infrastructure/mappingreader"
	procurementorder "github.com/dujiao-next/internal/modules/procurement/infrastructure/orderreader"
	procurementupstream "github.com/dujiao-next/internal/modules/procurement/infrastructure/upstreamgateway"
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
	fulfill *procurementfulfill.Service
	order   *orderdomain.Order
	calls   *[]string
	checkOK *bool
	// checkOverride 非空时接管 session/check 的响应（模拟不可用 / 报错 / 429）。
	checkOverride *func(w http.ResponseWriter)
	caps          *[]string
	server        *httptest.Server
	connID        uint
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
	f.checkOverride = new(func(w http.ResponseWriter))
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*f.calls = append(*f.calls, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/upstream/ping":
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "site_name": "up", "capabilities": *f.caps})
		case "/api/v1/upstream/session/check":
			if *f.checkOverride != nil {
				(*f.checkOverride)(w)
				return
			}
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
	f.fulfill = newTestFulfillService(db, connSvc, f.svc)
	return f
}

func newTestFulfillService(db *gorm.DB, connections *siteconnectionapp.Service, creator procurementfulfill.Creator) *procurementfulfill.Service {
	orders := ordergormstore.New(db, "test-guest-credential-secret-with-32-bytes")
	return procurementfulfill.New(procurementfulfill.Options{
		Repository:      procurementgormstore.New(db),
		Orders:          procurementorder.New(orders),
		ProductMappings: procurementmapping.NewProducts(mappinggormstore.NewMappingStore(db)),
		SKUMappings:     procurementmapping.NewSKUs(mappinggormstore.NewSKUMappingStore(db)),
		Connections:     procurementupstream.New(connections),
		Creator:         creator,
	})
}

type failingCreator struct{}

func (failingCreator) CreateForOrder(uint) error { return errors.New("boom") }

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
	if err := f.fulfill.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": `{"accessToken":"x"}`}); err != nil {
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
	err := f.fulfill.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": "abc"})
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
	if err := f.fulfill.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": "abc"}); err != nil {
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
	if err := f.fulfill.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{}); !errors.Is(err, manualform.ErrRequiredMissing) {
		t.Fatalf("expected required-missing error, got %v", err)
	}
	if err := f.fulfill.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": "abc"}); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	if err := f.fulfill.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": "def"}); !errors.Is(err, procurementcontract.ErrFulfillInfoAlreadySubmitted) {
		t.Fatalf("expected already-submitted conflict, got %v", err)
	}
	if f.submission()["session_json"] != "abc" {
		t.Fatalf("second submit must not overwrite: %#v", f.submission())
	}

	// 未支付订单不允许提交。
	g := newFulfillFixture(t, "FULFILL-UNPAID-001", nil)
	g.db.Model(&orderdomain.Order{}).Where("id = ?", g.order.ID).Update("status", constants.OrderStatusPendingPayment)
	if err := g.fulfill.SubmitFulfillInfo(context.Background(), g.order.ID, jsonmap.JSON{"session_json": "abc"}); !errors.Is(err, procurementcontract.ErrFulfillInfoNotAllowed) {
		t.Fatalf("expected not-allowed for unpaid order, got %v", err)
	}
}

func TestGetFulfillProgress_Phases(t *testing.T) {
	ctx := context.Background()
	f := newFulfillFixture(t, "FULFILL-PROG-001", []string{"order_progress"})

	p, err := f.fulfill.GetFulfillProgress(ctx, f.order.ID)
	if err != nil || p.Phase != procurementcontract.FulfillPhaseAwaitingInfo {
		t.Fatalf("expected awaiting_info, got %#v err=%v", p, err)
	}

	if err := f.fulfill.SubmitFulfillInfo(ctx, f.order.ID, jsonmap.JSON{"session_json": "abc"}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	p, _ = f.fulfill.GetFulfillProgress(ctx, f.order.ID)
	if p.Phase != procurementcontract.FulfillPhaseSubmitting {
		t.Fatalf("expected submitting, got %q", p.Phase)
	}

	// accepted + order_progress：透传上游步骤。
	f.db.Model(&ProcurementOrder{}).Where("local_order_id = ?", f.order.ID).
		Updates(map[string]any{"status": "accepted", "upstream_order_id": 999})
	p, _ = f.fulfill.GetFulfillProgress(ctx, f.order.ID)
	if p.Phase != procurementcontract.FulfillPhaseProcessing || len(p.Steps) != 2 || p.Step == nil || p.Step.Key != "charging" ||
		p.Percent != 50 || p.PollAfterMs != 1500 || p.Message != "处理中" {
		t.Fatalf("unexpected live progress: %#v", p)
	}

	// 无 order_progress 能力：降级为无步骤的 processing。
	g := newFulfillFixture(t, "FULFILL-PROG-002", nil)
	g.fulfill.SubmitFulfillInfo(ctx, g.order.ID, jsonmap.JSON{"session_json": "abc"})
	g.db.Model(&ProcurementOrder{}).Where("local_order_id = ?", g.order.ID).
		Updates(map[string]any{"status": "accepted", "upstream_order_id": 999})
	p, _ = g.fulfill.GetFulfillProgress(ctx, g.order.ID)
	if p.Phase != procurementcontract.FulfillPhaseProcessing || len(p.Steps) != 0 {
		t.Fatalf("expected degraded processing, got %#v", p)
	}

	// rejected → failed；fulfilled → completed。
	g.db.Model(&ProcurementOrder{}).Where("local_order_id = ?", g.order.ID).Update("status", "rejected")
	p, _ = g.fulfill.GetFulfillProgress(ctx, g.order.ID)
	if p.Phase != procurementcontract.FulfillPhaseFailed || !p.Terminal || p.MessageKey == "" {
		t.Fatalf("expected failed, got %#v", p)
	}
	g.db.Model(&ProcurementOrder{}).Where("local_order_id = ?", g.order.ID).Update("status", "fulfilled")
	p, _ = g.fulfill.GetFulfillProgress(ctx, g.order.ID)
	if p.Phase != procurementcontract.FulfillPhaseCompleted || !p.Terminal {
		t.Fatalf("expected completed, got %#v", p)
	}
}

// 客户已付款：预检不可用（check_unavailable）、上游报错或 429 时都不能拦住提交。
func TestSubmitFulfillInfo_PrecheckUnavailableDoesNotBlock(t *testing.T) {
	cases := map[string]func(w http.ResponseWriter){
		"check_unavailable": func(w http.ResponseWriter) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok": true, "valid": false, "can_open": false, "reason_code": "check_unavailable", "message": "暂时无法核验", "account": nil,
			})
		},
		"server_error": func(w http.ResponseWriter) { w.WriteHeader(http.StatusInternalServerError) },
		"rate_limited": func(w http.ResponseWriter) {
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error_code": "rate_limited"})
		},
	}
	i := 0
	for name, handler := range cases {
		i++
		t.Run(name, func(t *testing.T) {
			f := newFulfillFixture(t, "FULFILL-UNAVAIL-00"+string(rune('0'+i)), []string{"session_check"})
			*f.checkOverride = handler
			if err := f.fulfill.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": "abc"}); err != nil {
				t.Fatalf("precheck unavailable must not block: %v", err)
			}
			if !f.called("/session/check") {
				t.Fatal("expected session check to be attempted")
			}
			if f.submission()["session_json"] != "abc" || f.procCount() != 1 {
				t.Fatalf("submission must be saved and procurement created: %#v procs=%d", f.submission(), f.procCount())
			}
		})
	}
}

// CreateForOrder 失败（非 ErrExists）时撤销已保存的提交值，订单回到待填写，客户可重试。
func TestSubmitFulfillInfo_CreateFailureRollsBackSubmission(t *testing.T) {
	f := newFulfillFixture(t, "FULFILL-ROLLBACK-001", nil)
	failing := newTestFulfillService(f.db, newTestSiteConnectionService(f.db, "test-key", t.TempDir()), failingCreator{})
	if err := failing.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": "abc"}); err == nil {
		t.Fatal("expected the create failure to be returned")
	}
	if len(f.submission()) != 0 {
		t.Fatalf("submission must be cleared after a failed create: %#v", f.submission())
	}
	// 回到待填写：用正常的创建器重试成功。
	if err := f.fulfill.SubmitFulfillInfo(context.Background(), f.order.ID, jsonmap.JSON{"session_json": "abc"}); err != nil {
		t.Fatalf("retry should succeed: %v", err)
	}
	if f.procCount() != 1 {
		t.Fatalf("expected procurement order after retry, got %d", f.procCount())
	}
}

// 采购单被上游接收后又被拒/取消：用上游进度里的 message 作为原因展示。
func TestGetFulfillProgress_FailedUsesUpstreamMessage(t *testing.T) {
	f := newFulfillFixture(t, "FULFILL-REASON-001", []string{"order_progress"})
	ctx := context.Background()
	if err := f.fulfill.SubmitFulfillInfo(ctx, f.order.ID, jsonmap.JSON{"session_json": "abc"}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	f.db.Model(&ProcurementOrder{}).Where("local_order_id = ?", f.order.ID).
		Updates(map[string]any{"status": "rejected", "upstream_order_id": 999})
	p, _ := f.fulfill.GetFulfillProgress(ctx, f.order.ID)
	if p.Phase != procurementcontract.FulfillPhaseFailed || p.Message != "处理中" || p.MessageKey != "" {
		t.Fatalf("expected upstream message as the reason, got %#v", p)
	}
	// 没有上游订单号时回落到通用文案 key。
	f.db.Model(&ProcurementOrder{}).Where("local_order_id = ?", f.order.ID).Update("upstream_order_id", 0)
	p, _ = f.fulfill.GetFulfillProgress(ctx, f.order.ID)
	if p.Message != "" || p.MessageKey != procurementcontract.MessageKeyProcurementFailed {
		t.Fatalf("expected generic key, got %#v", p)
	}
}
