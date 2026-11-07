package paymenthttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
	paymentdomain "github.com/dujiao-next/internal/modules/payment/domain"
	reseller "github.com/dujiao-next/internal/modules/reseller/contract"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

func TestRespondPaymentCreateError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		err  error
		code int
		msg  string
	}{
		{
			name: "gateway response invalid",
			err:  ErrPaymentGatewayResponseInvalid,
			code: response.CodeBadRequest,
			msg:  "支付网关响应异常",
		},
		{
			name: "recharge channel not allowed",
			err:  ErrPaymentChannelNotAllowedForRecharge,
			code: response.CodeBadRequest,
			msg:  "钱包充值不支持此支付渠道",
		},
		{
			name: "unknown error",
			err:  errors.New("boom"),
			code: response.CodeInternal,
			msg:  "创建支付失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, recorder := newResponseTestContext()

			respondPaymentCreateError(c, tt.err)

			assertErrorResponse(t, recorder, tt.code, tt.msg)
		})
	}
}

func TestRespondPaymentCaptureError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		err  error
		code int
		msg  string
	}{
		{
			name: "amount mismatch",
			err:  ErrPaymentAmountMismatch,
			code: response.CodeBadRequest,
			msg:  "支付金额不匹配",
		},
		{
			name: "unknown error",
			err:  errors.New("boom"),
			code: response.CodeInternal,
			msg:  "支付回调处理失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, recorder := newResponseTestContext()

			respondPaymentCaptureError(c, tt.err)

			assertErrorResponse(t, recorder, tt.code, tt.msg)
		})
	}
}

func newResponseTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	return c, recorder
}

func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantCode int, wantMsg string) {
	t.Helper()
	if recorder.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var body response.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.StatusCode != wantCode {
		t.Fatalf("status_code = %d, want %d; body=%s", body.StatusCode, wantCode, recorder.Body.String())
	}
	if body.Msg != wantMsg {
		t.Fatalf("msg = %q, want %q; body=%s", body.Msg, wantMsg, recorder.Body.String())
	}
}

// --- rate limit regression ---

// stubGuestOrders / stubUserOrders / stubPaymentWriter 仅用于让 NewWriteHandler 通过非空校验,
// 被测路由在限流命中前根本不会命中 handler 方法,所以这里全部返回零值即可。
type stubGuestOrders struct{}

func (stubGuestOrders) GetOrderByGuestOrderNoForTenant(reseller.TenantContext, string, string, string) (*orderdomain.Order, error) {
	return nil, nil
}
func (stubGuestOrders) GetOrderByGuestForTenant(reseller.TenantContext, uint, string, string) (*orderdomain.Order, error) {
	return nil, nil
}

type stubUserOrders struct{}

func (stubUserOrders) GetOrderByUserOrderNoForTenant(reseller.TenantContext, string, uint) (*orderdomain.Order, error) {
	return nil, nil
}
func (stubUserOrders) GetOrderByUserForTenant(reseller.TenantContext, uint, uint) (*orderdomain.Order, error) {
	return nil, nil
}

type stubPaymentWriter struct{}

func (stubPaymentWriter) CreatePayment(CreatePaymentInput) (*CreatePaymentResult, error) {
	return nil, nil
}
func (stubPaymentWriter) GetPayment(uint) (*paymentdomain.Payment, error)                  { return nil, nil }
func (stubPaymentWriter) CapturePayment(CapturePaymentInput) (*paymentdomain.Payment, error) {
	return nil, nil
}
func (stubPaymentWriter) SubmitTxHash(SubmitTxHashInput) (*paymentdomain.Payment, error) {
	return nil, nil
}

// countingRateLimit 模拟 middleware.RateLimitMiddleware 的"前 N 次放行,第 N+1 次 429"行为。
// 这里不直接 import app 层 middleware,避免 module transport 反向依赖;
// 真实中间件行为见 internal/app/httpserver/middleware/rate_limit_test.go。
func countingRateLimit(limit int) (gin.HandlerFunc, *int) {
	var calls int
	handler := func(c *gin.Context) {
		calls++
		if calls > limit {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
	return handler, &calls
}

func newStubWriteHandler() *WriteHandler {
	return NewWriteHandler(stubGuestOrders{}, stubUserOrders{}, stubPaymentWriter{})
}

// TestRegisterUserWriteRoutes_SubmitTxHashIsRateLimited 回归保护:
// /payments/:id/submit-tx-hash 必须经过独立的 txHashRateLimit 中间件,
// 前 10 次放行,第 11 次拿 429;同组其他写路由不受此限流影响。
func TestRegisterUserWriteRoutes_SubmitTxHashIsRateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rateLimit, calls := countingRateLimit(10)

	RegisterUserWriteRoutes(router, newStubWriteHandler(), rateLimit)

	for i := 1; i <= 10; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments/1/submit-tx-hash", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("request %d: unexpected 429 before limit (body=%s)", i, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/payments/1/submit-tx-hash", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("11th request: want 429, got %d (body=%s)", w.Code, w.Body.String())
	}

	// 其它 write 路由必须不走此限流
	before := *calls
	for _, path := range []string{"/payments", "/payments/1/capture"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
	if *calls != before {
		t.Fatalf("tx-hash rate limit leaked onto other write routes: counter moved from %d to %d", before, *calls)
	}
}

// TestRegisterGuestWriteRoutes_SubmitTxHashIsRateLimited 游客侧对应回归保护。
func TestRegisterGuestWriteRoutes_SubmitTxHashIsRateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rateLimit, calls := countingRateLimit(10)

	RegisterGuestWriteRoutes(router, newStubWriteHandler(), rateLimit)

	for i := 1; i <= 10; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments/1/submit-tx-hash", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("request %d: unexpected 429 before limit (body=%s)", i, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/payments/1/submit-tx-hash", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("11th request: want 429, got %d (body=%s)", w.Code, w.Body.String())
	}

	before := *calls
	for _, path := range []string{"/payments", "/payments/1/capture"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
	if *calls != before {
		t.Fatalf("tx-hash rate limit leaked onto other write routes: counter moved from %d to %d", before, *calls)
	}
}
