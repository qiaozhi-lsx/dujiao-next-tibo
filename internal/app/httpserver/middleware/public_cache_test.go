package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

// TestPublicCacheHeaders_SetsExpectedHeader 验证中间件给响应写入
// Cache-Control: public, max-age=10, s-maxage=10。
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
	want := "public, max-age=10, s-maxage=10"
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

// TestPublicCacheHeaders_SkipsNonGET 验证组内的写操作(如 POST /affiliate/click)
// 不会被打上 public 缓存头:非 GET/HEAD 请求中间件什么都不写。
func TestPublicCacheHeaders_SkipsNonGET(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	public := r.Group("/api/v1/public")
	public.Use(PublicCacheHeaders())
	public.POST("/affiliate/click", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/affiliate/click", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Cache-Control"); got != "" {
		t.Errorf("POST got Cache-Control = %q, want empty", got)
	}
	if got := w.Header().Values("Vary"); len(got) != 0 {
		t.Errorf("POST got Vary = %q, want none", got)
	}
}

// TestPublicCacheHeaders_AddsVaryLanguage 验证 GET 响应带 Vary: Accept-Language, X-Lang
// (成功体当前不按语言渲染;Vary 为按语言本地化的错误体与未来可能按语言渲染的公开字段兜底,
// 见 publicCacheVaryValue 注释),且用 Add 追加、不覆盖 CORS 中间件已经写入的 Vary: Origin。
func TestPublicCacheHeaders_AddsVaryLanguage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// 模拟挂在外层的 CORS 中间件(middleware.go 用 Add 写 Vary: Origin)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Add("Vary", "Origin")
		c.Next()
	})
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
	got := w.Header().Values("Vary")
	for _, want := range []string{"Origin", "Accept-Language, X-Lang"} {
		found := false
		for _, v := range got {
			if v == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Vary = %q, want to contain %q", got, want)
		}
	}
}

// TestPublicCacheHeaders_ErrorResponseIsNoStore 验证业务错误(HTTP 200 + status_code != 0)
// 走 response.Error 时缓存头被覆写为 no-store,不会被 CDN 当成功响应缓存并重放给其他访客。
func TestPublicCacheHeaders_ErrorResponseIsNoStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	public := r.Group("/api/v1/public")
	public.Use(PublicCacheHeaders())
	public.GET("/boom", func(c *gin.Context) {
		response.Error(c, response.CodeInternal, "x")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/boom", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("error response Cache-Control = %q, want %q", got, "no-store")
	}
}

// TestPublicCacheHeaders_HandlerCanOverrideWithNoStore 验证挂在 /api/v1/public 组上的
// 中间件在 c.Next() 之前写默认头后，一次性挑战类接口（如 /captcha/image）在 handler 内
// c.Header("Cache-Control", "no-store") 能把最终响应头替换为唯一的 no-store（不是追加第二个值），
// 且该覆盖只作用于本次响应，同组其他路由仍拿到默认缓存头。
func TestPublicCacheHeaders_HandlerCanOverrideWithNoStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	public := r.Group("/api/v1/public")
	public.Use(PublicCacheHeaders())
	public.GET("/captcha/image", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, gin.H{"captcha_id": "id-1"})
	})
	public.GET("/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/captcha/image", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want %q", got, "no-store")
	}
	if vals := w.Header().Values("Cache-Control"); len(vals) != 1 {
		t.Errorf("Cache-Control should be replaced, not appended; got %q", vals)
	}

	// 覆盖只影响本次响应：同组未覆盖的路由仍带默认缓存头。
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/public/config", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if got := w2.Header().Get("Cache-Control"); got == "" || got == "no-store" {
		t.Errorf("sibling route Cache-Control = %q, want middleware default", got)
	}
}
