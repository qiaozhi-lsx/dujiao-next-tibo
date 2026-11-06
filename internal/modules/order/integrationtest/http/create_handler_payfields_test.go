package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/constants"
	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
	orderhttp "github.com/dujiao-next/internal/modules/order/transport/http"
	paymentdomain "github.com/dujiao-next/internal/modules/payment/domain"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type stubOrderCreateService struct {
	order *orderdomain.Order
}

func (s stubOrderCreateService) CreateOrder(orderhttp.CreateOrderInput) (*orderdomain.Order, error) {
	return s.order, nil
}

func (s stubOrderCreateService) CreateGuestOrder(orderhttp.CreateGuestOrderInput) (*orderdomain.Order, error) {
	return s.order, nil
}

type stubOrderPaymentCreator struct {
	result *orderhttp.CreatePaymentResult
}

func (s stubOrderPaymentCreator) CreatePayment(orderhttp.CreatePaymentInput) (*orderhttp.CreatePaymentResult, error) {
	return s.result, nil
}

// TestRespondCreateAndPay_IncludesCurrencyAndFeePolicy 经 CreateOrderAndPay 走到 respondCreateAndPay，
// 断言支付记录的实付币种与手续费策略快照被放进响应。订单币种故意与支付币种不同，
// 确保 currency 取自支付记录（汇率换算后的实付币种）而不是订单。
func TestRespondCreateAndPay_IncludesCurrencyAndFeePolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	order := &orderdomain.Order{ID: 1, OrderNo: "DJ-CREATE-AND-PAY-1", Currency: "USD"}
	payment := &paymentdomain.Payment{
		ID:        9,
		OrderID:   order.ID,
		ChannelID: 7,
		Amount:    money.FromDecimal(decimal.RequireFromString("72.50")),
		FeeAmount: money.FromDecimal(decimal.RequireFromString("3.00")),
		FeePolicy: constants.PaymentFeePolicyMerchantAbsorbed,
		Currency:  "CNY",
	}
	h := orderhttp.NewCreateHandler(
		stubOrderCreateService{order: order},
		nil,
		nil,
		stubOrderPaymentCreator{result: &orderhttp.CreatePaymentResult{Payment: payment}},
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/orders/create-and-pay",
		strings.NewReader(`{"items":[{"product_id":1,"quantity":1}],"channel_id":7}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", uint(1))

	h.CreateOrderAndPay(c)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var body struct {
		StatusCode int                    `json:"status_code"`
		Data       map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, w.Body.String())
	}
	if body.StatusCode != response.CodeOK {
		t.Fatalf("status_code = %d, want %d; body=%s", body.StatusCode, response.CodeOK, w.Body.String())
	}
	for key, want := range map[string]string{
		"currency":       "CNY",
		"fee_policy":     "merchant_absorbed",
		"payable_amount": "72.50",
		"fee_amount":     "3.00",
	} {
		if got := body.Data[key]; got != want {
			t.Errorf("data.%s = %v, want %q", key, got, want)
		}
	}
}
