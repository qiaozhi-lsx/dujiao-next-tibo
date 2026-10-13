package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestPublicCacheHeaders_SetsExpectedHeader 验证中间件给响应写入
// Cache-Control: public, max-age=60, s-maxage=60。
func TestPublicCacheHeaders_SetsExpectedHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	public := r.Group("/api/v1/public")
	public.Use(PublicCacheHeaders())
	public.GET("/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	got := w.Header().Get("Cache-Control")
	want := "public, max-age=60, s-maxage=60"
	if got != want {
		t.Errorf("Cache-Control = %q, want %q", got, want)
	}
}

// TestPublicCacheHeaders_DoesNotLeakToSiblingGroups 验证中间件只挂到公开路由组,
// 不会把 Cache-Control 泄漏到 /user/* 等鉴权路径。
func TestPublicCacheHeaders_DoesNotLeakToSiblingGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiV1 := r.Group("/api/v1")

	public := apiV1.Group("/public")
	public.Use(PublicCacheHeaders())
	public.GET("/config", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// 兄弟组:不挂中间件
	user := apiV1.Group("/user")
	user.GET("/profile", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// /user/profile 不应有 Cache-Control
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Cache-Control"); got != "" {
		t.Errorf("sibling group leaked Cache-Control = %q, want empty", got)
	}

	// /public/config 应该有 Cache-Control
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/public/config", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if got := w2.Header().Get("Cache-Control"); got == "" {
		t.Errorf("public group missing Cache-Control header")
	}
}

// TestPublicCacheHeaders_HandlerCanOverride 验证 handler 可以覆盖 Cache-Control
// (比如需要 no-store 的少数敏感公开接口),中间件只写默认值,不强制。
func TestPublicCacheHeaders_HandlerCanOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	public := r.Group("/api/v1/public")
	public.Use(PublicCacheHeaders())
	public.GET("/no-cache", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/no-cache", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("handler override failed, got %q", got)
	}
}
