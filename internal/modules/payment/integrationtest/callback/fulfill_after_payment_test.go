package paymentcallback_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/constants"
	mappingdomain "github.com/dujiao-next/internal/modules/catalog/mapping/domain"
	mappinggormstore "github.com/dujiao-next/internal/modules/catalog/mapping/infrastructure/gormstore"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	productgormstore "github.com/dujiao-next/internal/modules/catalog/product/store/gormstore"
	fulfillmentdomain "github.com/dujiao-next/internal/modules/fulfillment/domain"
	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
	ordergormstore "github.com/dujiao-next/internal/modules/order/infrastructure/gormstore"
	orderpresenter "github.com/dujiao-next/internal/modules/order/transport/presenter"
	paymentapp "github.com/dujiao-next/internal/modules/payment/application"
	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"
	paymentdomain "github.com/dujiao-next/internal/modules/payment/domain"
	okpayadapter "github.com/dujiao-next/internal/modules/payment/infrastructure/gateway/adapters/okpay"
	paymentprovider "github.com/dujiao-next/internal/modules/payment/infrastructure/gateway/provider"
	paymentgormstore "github.com/dujiao-next/internal/modules/payment/infrastructure/gormstore"
	procurementapp "github.com/dujiao-next/internal/modules/procurement/application"
	procurementfulfill "github.com/dujiao-next/internal/modules/procurement/application/fulfill"
	procurementcontract "github.com/dujiao-next/internal/modules/procurement/contract"
	procurementdomain "github.com/dujiao-next/internal/modules/procurement/domain"
	procurementgormstore "github.com/dujiao-next/internal/modules/procurement/infrastructure/gormstore"
	procurementmapping "github.com/dujiao-next/internal/modules/procurement/infrastructure/mappingreader"
	procurementorder "github.com/dujiao-next/internal/modules/procurement/infrastructure/orderreader"
	procurementupstream "github.com/dujiao-next/internal/modules/procurement/infrastructure/upstreamgateway"
	siteconnectionapp "github.com/dujiao-next/internal/modules/siteconnection/application"
	siteconnectiondomain "github.com/dujiao-next/internal/modules/siteconnection/domain"
	siteconnectiongormstore "github.com/dujiao-next/internal/modules/siteconnection/infrastructure/gormstore"
	"github.com/dujiao-next/internal/platform/database/gormdb"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// fakePaymentQueue 让 enqueueOrderPaidAsync 走完整的「已启用队列」分支（含创建采购单），任务本身不执行。
type fakePaymentQueue struct{}

func (fakePaymentQueue) Enabled() bool                                                { return true }
func (fakePaymentQueue) EnqueueTimeoutCancel(uint, time.Duration) error               { return nil }
func (fakePaymentQueue) EnqueueStatusEmail(uint, string) error                        { return nil }
func (fakePaymentQueue) EnqueueOrderAutoFulfill(uint) error                           { return nil }
func (fakePaymentQueue) EnqueueBotNotification(paymentcontract.BotNotification) error { return nil }
func (fakePaymentQueue) EnqueueWalletRechargeExpire(uint, time.Duration) error        { return nil }

type fulfillPaymentFixture struct {
	db       *gorm.DB
	payments *paymentapp.PaymentService
	orders   *ordergormstore.Store
	proc     *procurementapp.Service
	fulfill  *procurementfulfill.Service
	parent   *orderdomain.Order
	child    *orderdomain.Order
	payment  *paymentdomain.Payment
}

// newFulfillPaymentFixture 按真实下单形状建「父订单 + 每个商品一个子订单」（订单项在子订单上，父订单无订单项），
// 然后用真实的 PaymentService 回调处理支付成功。deferred=true 时订单项带表单 schema 快照、提交值为空（defer_manual_form）。
func newFulfillPaymentFixture(t *testing.T, deferred bool) *fulfillPaymentFixture {
	t.Helper()
	dsn := fmt.Sprintf("file:payment_fulfill_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(
		&productdomain.Product{}, &productdomain.ProductSKU{},
		&orderdomain.Order{}, &orderdomain.OrderItem{}, &orderdomain.OrderRefundRecord{},
		&fulfillmentdomain.Fulfillment{}, &paymentdomain.PaymentChannel{}, &paymentdomain.Payment{},
		&procurementdomain.Order{}, &siteconnectiondomain.Connection{},
		&mappingdomain.Mapping{}, &mappingdomain.SKUMapping{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	gormdb.DB = db

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/upstream/ping":
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "site_name": "up", "capabilities": []string{"order_progress"}})
		case "/api/v1/upstream/orders":
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "order_id": 999, "order_no": "UP-999", "status": "accepted", "amount": "50.00", "currency": "CNY"})
		case "/api/v1/upstream/orders/999/progress":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok": true, "order_id": 999, "status": "processing",
				"steps":   []map[string]string{{"key": "create", "label": "提交订单"}, {"key": "charging", "label": "扣款确认"}},
				"step":    map[string]any{"index": 2, "total": 2, "key": "charging", "label": "扣款确认"},
				"percent": 50, "terminal": false, "outcome": nil, "message": nil, "poll_after_ms": 2000,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	connSvc := siteconnectionapp.NewService(siteconnectiongormstore.New(db), "test-key", t.TempDir())
	conn, err := connSvc.Create(siteconnectionapp.CreateInput{
		Name: "up", BaseURL: server.URL, ApiKey: "key", ApiSecret: "secret", Protocol: constants.ConnectionProtocolDujiaoNext,
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}
	pm := &mappingdomain.Mapping{ConnectionID: conn.ID, LocalProductID: 1, UpstreamProductID: 101, IsActive: true}
	db.Create(pm)
	db.Create(&mappingdomain.SKUMapping{ProductMappingID: pm.ID, LocalSKUID: 1, UpstreamSKUID: 201, UpstreamIsActive: true})

	const guestSecret = "test-guest-credential-secret-with-32-bytes"
	orders := ordergormstore.New(db, guestSecret)
	now := time.Now().UTC().Truncate(time.Second)
	amount := money.FromDecimal(decimal.NewFromInt(88))
	zero := money.FromDecimal(decimal.Zero)
	parent := &orderdomain.Order{
		OrderNo: "DJFULFILLPARENT001", GuestEmail: "guest@example.com", Status: constants.OrderStatusPendingPayment, Currency: "CNY",
		OriginalAmount: amount, TotalAmount: amount, OnlinePaidAmount: amount, WalletPaidAmount: zero, RefundedAmount: zero,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(parent).Error; err != nil {
		t.Fatalf("create parent: %v", err)
	}
	child := &orderdomain.Order{
		OrderNo: "DJFULFILLPARENT001-1", ParentID: &parent.ID, GuestEmail: "guest@example.com", Status: constants.OrderStatusPendingPayment, Currency: "CNY",
		OriginalAmount: amount, TotalAmount: amount, OnlinePaidAmount: amount, WalletPaidAmount: zero, RefundedAmount: zero,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(child).Error; err != nil {
		t.Fatalf("create child: %v", err)
	}
	schema := jsonmap.JSON{"fields": []interface{}{map[string]interface{}{"key": "session_json", "type": "textarea", "required": true}}}
	item := &orderdomain.OrderItem{
		OrderID: child.ID, ProductID: 1, SKUID: 1, Quantity: 1, FulfillmentType: constants.FulfillmentTypeUpstream,
		TitleJSON: jsonmap.JSON{"zh-CN": "ChatGPT Plus"}, UnitPrice: amount, TotalPrice: amount,
		ManualFormSchemaSnapshotJSON: schema, ManualFormSubmissionJSON: jsonmap.JSON{},
	}
	if !deferred {
		item.ManualFormSubmissionJSON = jsonmap.JSON{"session_json": "abc"}
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("create item: %v", err)
	}

	channel := &paymentdomain.PaymentChannel{
		Name: "OKPAY", ProviderType: constants.PaymentProviderOkpay, ChannelType: constants.PaymentChannelTypeUsdt,
		InteractionMode: constants.PaymentInteractionQR, FeeRate: zero, IsActive: true, CreatedAt: now, UpdatedAt: now,
		ConfigJSON: jsonmap.JSON{"merchant_id": "shop-1", "merchant_token": "token-1", "return_url": "https://example.com/pay", "callback_url": "https://api.example.com/cb", "exchange_rate": "7"},
	}
	if err := db.Create(channel).Error; err != nil {
		t.Fatalf("create channel: %v", err)
	}
	payment := &paymentdomain.Payment{
		OrderID: parent.ID, ChannelID: channel.ID, ProviderType: channel.ProviderType, ChannelType: channel.ChannelType,
		InteractionMode: channel.InteractionMode, Amount: money.FromDecimal(decimal.NewFromInt(616)), FeeRate: zero, FeeAmount: zero,
		Currency: "USDT", Status: constants.PaymentStatusPending, ProviderRef: "OKPAY-ORDER-F1", GatewayOrderNo: "DJPF1",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(payment).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}

	registry := paymentprovider.NewRegistry()
	registry.Register(constants.PaymentProviderOkpay, "", okpayadapter.NewOkpayAdapter())
	payments := paymentapp.NewPaymentService(paymentapp.PaymentServiceOptions{
		OrderStore: orders, ProductRepo: productgormstore.NewProductStore(db), ProductSKURepo: productgormstore.NewSKUStore(db),
		PaymentStore: paymentgormstore.New(db, guestSecret), ChannelStore: paymentgormstore.NewChannelStore(db),
		Queue: fakePaymentQueue{}, ExpireMinutes: 15, PaymentProviderRegistry: registry,
	})
	procSvc := procurementapp.NewService(procurementapp.Options{
		Repository:      procurementgormstore.New(db),
		Orders:          procurementorder.New(orders),
		ProductMappings: procurementmapping.NewProducts(mappinggormstore.NewMappingStore(db)),
		SKUMappings:     procurementmapping.NewSKUs(mappinggormstore.NewSKUMappingStore(db)),
		Connections:     procurementupstream.New(connSvc),
		OrderLifecycle:  procurementgormstore.NewLifecycle(db, nil, nil, config.EmailConfig{}),
	})
	payments.SetProcurementService(procSvc)
	fulfillSvc := procurementfulfill.New(procurementfulfill.Options{
		Repository:      procurementgormstore.New(db),
		Orders:          procurementorder.New(orders),
		ProductMappings: procurementmapping.NewProducts(mappinggormstore.NewMappingStore(db)),
		SKUMappings:     procurementmapping.NewSKUs(mappinggormstore.NewSKUMappingStore(db)),
		Connections:     procurementupstream.New(connSvc),
		Creator:         procSvc,
	})
	return &fulfillPaymentFixture{db: db, payments: payments, orders: orders, proc: procSvc, fulfill: fulfillSvc, parent: parent, child: child, payment: payment}
}

// pay 走与支付网关回调相同的 PaymentService.HandleCallback。
func (f *fulfillPaymentFixture) pay(t *testing.T) {
	t.Helper()
	if _, err := f.payments.HandleCallback(paymentapp.PaymentCallbackInput{
		PaymentID: f.payment.ID, OrderNo: f.parent.OrderNo, ChannelID: f.payment.ChannelID,
		Status: constants.PaymentStatusSuccess, ProviderRef: "OKPAY-ORDER-F1",
		Amount: money.FromDecimal(decimal.NewFromInt(616)), Currency: "USDT",
	}); err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}
}

func (f *fulfillPaymentFixture) status(id uint) string {
	var o orderdomain.Order
	f.db.First(&o, id)
	return o.Status
}

func (f *fulfillPaymentFixture) procCount() int64 {
	var n int64
	f.db.Model(&procurementdomain.Order{}).Count(&n)
	return n
}

func TestDeferredFormOrderStaysPaidUntilFulfillInfoSubmitted(t *testing.T) {
	f := newFulfillPaymentFixture(t, true)
	f.pay(t)

	// 父、子订单都保持 paid，且没有采购单。
	if got := f.status(f.parent.ID); got != constants.OrderStatusPaid {
		t.Fatalf("parent must stay paid, got %q", got)
	}
	if got := f.status(f.child.ID); got != constants.OrderStatusPaid {
		t.Fatalf("child must stay paid, got %q", got)
	}
	if f.procCount() != 0 {
		t.Fatalf("no procurement order before the form is submitted, got %d", f.procCount())
	}

	// 展示层：订单项与父订单汇总都带 fulfill_info_pending。
	loaded, err := f.orders.GetByID(f.parent.ID)
	if err != nil || loaded == nil {
		t.Fatalf("reload parent: %v", err)
	}
	detail := orderpresenter.NewOrderDetail(loaded)
	if !detail.FulfillInfoPending || len(detail.Children) != 1 || len(detail.Children[0].Items) != 1 || !detail.Children[0].Items[0].FulfillInfoPending {
		t.Fatalf("expected fulfill_info_pending on parent and child item: %+v", detail)
	}
	if summary := orderpresenter.NewOrderSummary(loaded); !summary.FulfillInfoPending {
		t.Fatalf("expected fulfill_info_pending in list summary")
	}

	ctx := context.Background()
	progress, err := f.fulfill.GetFulfillProgress(ctx, f.parent.ID)
	if err != nil || progress.Phase != procurementcontract.FulfillPhaseAwaitingInfo {
		t.Fatalf("expected awaiting_info, got %+v err=%v", progress, err)
	}

	// 用父订单 ID 提交：表单落在子订单项上，并创建（子订单的）采购单。
	if err := f.fulfill.SubmitFulfillInfo(ctx, f.parent.ID, jsonmap.JSON{"session_json": "abc"}); err != nil {
		t.Fatalf("SubmitFulfillInfo: %v", err)
	}
	var item orderdomain.OrderItem
	f.db.Where("order_id = ?", f.child.ID).First(&item)
	if item.ManualFormSubmissionJSON["session_json"] != "abc" {
		t.Fatalf("submission must be saved on the child item: %#v", item.ManualFormSubmissionJSON)
	}
	var procOrder procurementdomain.Order
	if err := f.db.Where("local_order_id = ?", f.child.ID).First(&procOrder).Error; err != nil {
		t.Fatalf("procurement order for the child expected: %v", err)
	}
	progress, _ = f.fulfill.GetFulfillProgress(ctx, f.parent.ID)
	if progress.Phase != procurementcontract.FulfillPhaseSubmitting {
		t.Fatalf("expected submitting, got %q", progress.Phase)
	}
	if err := f.fulfill.SubmitFulfillInfo(ctx, f.parent.ID, jsonmap.JSON{"session_json": "def"}); err == nil {
		t.Fatal("second submit must conflict")
	}

	// worker 提交采购单：子订单进入 fulfilling，父订单随之同步，进度变为 processing 并透传上游步骤。
	if err := f.proc.SubmitToUpstream(procOrder.ID); err != nil {
		t.Fatalf("SubmitToUpstream: %v", err)
	}
	if got := f.status(f.child.ID); got != constants.OrderStatusFulfilling {
		t.Fatalf("child expected fulfilling, got %q", got)
	}
	if got := f.status(f.parent.ID); got != constants.OrderStatusFulfilling {
		t.Fatalf("parent expected fulfilling, got %q", got)
	}
	progress, _ = f.fulfill.GetFulfillProgress(ctx, f.parent.ID)
	if progress.Phase != procurementcontract.FulfillPhaseProcessing || len(progress.Steps) != 2 || progress.Step == nil || progress.Step.Key != "charging" {
		t.Fatalf("expected processing with upstream steps, got %+v", progress)
	}
}

// 结账时就填好表单的普通上游订单：行为不变，支付后立即 fulfilling 并创建采购单。
func TestFilledFormUpstreamOrderStillFulfillsImmediately(t *testing.T) {
	f := newFulfillPaymentFixture(t, false)
	f.pay(t)

	if got := f.status(f.child.ID); got != constants.OrderStatusFulfilling {
		t.Fatalf("child expected fulfilling, got %q", got)
	}
	if got := f.status(f.parent.ID); got != constants.OrderStatusFulfilling {
		t.Fatalf("parent expected fulfilling, got %q", got)
	}
	if f.procCount() != 1 {
		t.Fatalf("expected procurement order created immediately, got %d", f.procCount())
	}
	loaded, _ := f.orders.GetByID(f.parent.ID)
	if orderpresenter.NewOrderDetail(loaded).FulfillInfoPending {
		t.Fatal("filled order must not be fulfill_info_pending")
	}
}
