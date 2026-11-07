package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newTestContext 构造一个已带上游 public 缓存头的测试上下文,
// 模拟请求经过 PublicCacheHeaders 中间件后进入 handler 的状态。
func newTestContext(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/config", nil)
	c.Header("Cache-Control", "public, max-age=10, s-maxage=10")
	return c, w
}

// TestError_SetsNoStoreCacheControl 验证业务错误(HTTP 200 + status_code != 0)
// 一律覆写为 Cache-Control: no-store,避免被 CDN/浏览器当成功响应缓存。
func TestError_SetsNoStoreCacheControl(t *testing.T) {
	c, w := newTestContext(t)

	Error(c, CodeInternal, "x")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want %q", got, "no-store")
	}
}

// TestErrorWithHTTPStatus_SetsNoStoreCacheControl 验证非 200 错误响应同样覆写为 no-store。
func TestErrorWithHTTPStatus_SetsNoStoreCacheControl(t *testing.T) {
	c, w := newTestContext(t)

	ErrorWithHTTPStatus(c, http.StatusServiceUnavailable, CodeInternal, "x")

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want %q", got, "no-store")
	}
}

// TestChannelError_SetsNoStoreCacheControl 验证渠道 API 错误响应与
// Error/ErrorWithHTTPStatus 口径一致,始终覆写为 Cache-Control: no-store,
// 避免被 CDN/浏览器当成功响应缓存并重放。
func TestChannelError_SetsNoStoreCacheControl(t *testing.T) {
	c, w := newTestContext(t)

	ChannelError(c, http.StatusBadRequest, CodeBadRequest, "x", "bad_request")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want %q", got, "no-store")
	}
}
