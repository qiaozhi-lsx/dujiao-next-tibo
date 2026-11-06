package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestValidateAdminPath(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantErr   bool
		errSubstr string
	}{
		{"valid default", "/admin", false, ""},
		{"valid custom", "/dj-mgmt-7x9k2", false, ""},
		{"valid with dashes", "/console-private", false, ""},

		{"empty", "", true, "must start with"},
		{"missing leading slash", "admin", true, "must start with"},
		{"only slash", "/", true, "cannot be"},
		{"trailing slash", "/admin/", true, "cannot end with"},
		{"trailing slash deep", "/foo/bar/", true, "cannot end with"},

		{"reserved api", "/api", true, "reserved"},
		{"reserved api prefix", "/api/v1", true, "reserved"},
		{"reserved uploads", "/uploads", true, "reserved"},
		{"reserved uploads prefix", "/uploads/x", true, "reserved"},
		{"reserved health", "/health", true, "reserved"},

		{"valid nested", "/ops/console", false, ""},
		{"valid underscores", "/dj_mgmt_7x9k2", false, ""},
		// 这几种在收紧校验之前就是合法的，不能因为一次安全加固把存量配置弄成起不来
		{"valid dot", "/ops.admin", false, ""},
		{"valid tilde", "/admin~private", false, ""},
		{"valid at sign", "/console@company", false, ""},

		// Gin 路由元字符：拼进 prefix + "/*filepath" 后会变成动态参数或 catch-all，
		// 前者把用户站路径吞成后台 SPA，后者直接让路由注册 panic
		{"gin param", "/:tenant", true, "invalid"},
		{"gin param nested", "/admin/:id", true, "invalid"},
		{"gin catch-all", "/*admin", true, "invalid"},
		{"gin catch-all nested", "/admin/*rest", true, "invalid"},

		{"double slash", "/admin//panel", true, "invalid"},
		{"space", "/my admin", true, "invalid"},
		{"tab", "/my\tadmin", true, "invalid"},
		{"backslash", "/admin\\panel", true, "invalid"},
		{"percent encoded", "/adm%2Fin", true, "invalid"},
		// "." / ".." 由字符集允许，但会被路径规范化吃掉，需要单独拒绝
		{"dot segment", "/../etc", true, "cannot contain"},
		{"single dot segment", "/./admin", true, "cannot contain"},
		{"trailing dot segment", "/admin/..", true, "cannot contain"},
		{"question mark", "/admin?x=1", true, "invalid"},
		{"hash", "/admin#top", true, "invalid"},
		{"control char", "/adm\x00in", true, "invalid"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateAdminPath(c.input)
			if c.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil")
				}
				if c.errSubstr != "" && !strings.Contains(err.Error(), c.errSubstr) {
					t.Fatalf("error %q does not contain %q", err.Error(), c.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func newAdminFS(indexHTML string) fstest.MapFS {
	return fstest.MapFS{
		"index.html":     &fstest.MapFile{Data: []byte(indexHTML)},
		"assets/app.js":  &fstest.MapFile{Data: []byte("console.log('app');")},
		"assets/app.css": &fstest.MapFile{Data: []byte("body{}")},
		"favicon.ico":    &fstest.MapFile{Data: []byte("\x00\x00")},
	}
}

func TestRegisterAdmin_PlaceholderReplacement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	html := `<!doctype html><html><head><base href="__DJ_ADMIN_BASE__/"><title>x</title></head><body></body></html>`
	if err := RegisterAdmin(r, "/dj-mgmt-7x9k2", newAdminFS(html)); err != nil {
		t.Fatalf("RegisterAdmin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/dj-mgmt-7x9k2/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), `<base href="/dj-mgmt-7x9k2/">`) {
		t.Fatalf("placeholder not replaced or missing leading slash; body = %s", body)
	}
	if strings.Contains(string(body), "__DJ_ADMIN_BASE__") {
		t.Fatalf("placeholder still present in body")
	}
}

// TestRegisterAdmin_BaseHrefIsAbsolute 防止 base href 退化成相对路径而触发
// 浏览器双前缀解析 bug（详见 handler.go:67 注释）。
func TestRegisterAdmin_BaseHrefIsAbsolute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name   string
		prefix string
		want   string
	}{
		{"default admin", "/admin", `<base href="/admin/">`},
		{"deep custom", "/dj-mgmt-7x9k2", `<base href="/dj-mgmt-7x9k2/">`},
		{"with multiple segments", "/console-private", `<base href="/console-private/">`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := gin.New()
			html := `<head><base href="__DJ_ADMIN_BASE__/"></head>`
			if err := RegisterAdmin(r, c.prefix, newAdminFS(html)); err != nil {
				t.Fatalf("RegisterAdmin: %v", err)
			}
			req := httptest.NewRequest(http.MethodGet, c.prefix+"/products/123/edit", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			body := w.Body.String()
			if !strings.Contains(body, c.want) {
				t.Fatalf("history-fallback body for %s missing %q\nbody = %s", c.prefix, c.want, body)
			}
		})
	}
}

func TestRegisterAdmin_ServesRealAsset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	if err := RegisterAdmin(r, "/admin", newAdminFS("<html></html>")); err != nil {
		t.Fatalf("RegisterAdmin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/assets/app.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	if string(body) != "console.log('app');" {
		t.Fatalf("body = %q", body)
	}
}

func TestRegisterAdmin_HistoryFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	html := `<!doctype html><html><head><base href="__DJ_ADMIN_BASE__/"></head><body>spa</body></html>`
	if err := RegisterAdmin(r, "/admin", newAdminFS(html)); err != nil {
		t.Fatalf("RegisterAdmin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/products/123/edit", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), "spa") {
		t.Fatalf("not index.html: %s", body)
	}
	ctype := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ctype, "text/html") {
		t.Fatalf("content-type = %q, want text/html*", ctype)
	}
}

func TestRegisterAdmin_NilFS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := RegisterAdmin(r, "/admin", nil); err == nil {
		t.Fatal("want error for nil fs.FS, got nil")
	}
}

func TestRegisterAdmin_MissingIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	fsys := fstest.MapFS{"assets/x.js": &fstest.MapFile{Data: []byte("x")}}
	if err := RegisterAdmin(r, "/admin", fsys); err == nil {
		t.Fatal("want error when index.html missing")
	}
}

func newUserFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":  &fstest.MapFile{Data: []byte("<html>user-spa</html>")},
		"assets/u.js": &fstest.MapFile{Data: []byte("console.log('u');")},
		"robots.txt":  &fstest.MapFile{Data: []byte("User-agent: *\n")},
	}
}

func TestRegisterUser_ServesIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := RegisterUser(r, newUserFS(), nil, nil); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), "user-spa") {
		t.Fatalf("body = %s", body)
	}
}

func TestRegisterUser_ServesAsset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := RegisterUser(r, newUserFS(), nil, nil); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/assets/u.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	if string(body) != "console.log('u');" {
		t.Fatalf("body = %q", body)
	}
}

func TestRegisterUser_HistoryFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := RegisterUser(r, newUserFS(), nil, nil); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/products/some-slug", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), "user-spa") {
		t.Fatalf("did not fallback to index.html: %s", body)
	}
}

// TestRegisterUser_ReservedPathsStay404 保证未命中的后端接口不会被 SPA 兜底成
// 200 HTML —— 否则 API 客户端会拿 index.html 去解析 JSON。
func TestRegisterUser_ReservedPathsStay404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := RegisterUser(r, newUserFS(), nil, nil); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	for _, p := range []string{
		"/api",
		"/api/v1/nonexistent",
		"/uploads/missing.png",
		"/health/subpath",
	} {
		t.Run(p, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, p, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", w.Code)
			}
			if strings.Contains(w.Body.String(), "user-spa") {
				t.Fatalf("reserved path fell back to SPA index: %s", w.Body.String())
			}
		})
	}
}

// TestRegisterUser_NonReservedLookalikeStillServesSPA 确认保留前缀判断是按路径段
// 匹配的，不会误伤 /apiary 这类只是恰好同前缀的前端路由。
func TestRegisterUser_NonReservedLookalikeStillServesSPA(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := RegisterUser(r, newUserFS(), nil, nil); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	for _, p := range []string{"/apiary", "/uploads-guide", "/healthy"} {
		t.Run(p, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, p, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", w.Code)
			}
			if !strings.Contains(w.Body.String(), "user-spa") {
				t.Fatalf("expected SPA fallback, got: %s", w.Body.String())
			}
		})
	}
}

func TestRegisterUser_DoesNotShadowExistingRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	if err := RegisterUser(r, newUserFS(), nil, nil); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if w.Body.String() != "pong" {
		t.Fatalf("body = %s, want pong", w.Body.String())
	}
}

// staticConfigSource 是测试用的 InitialConfigSource，固定返回预置的站点配置。
type staticConfigSource map[string]interface{}

func (s staticConfigSource) GetConfig(map[string]interface{}) (map[string]interface{}, error) {
	return s, nil
}

// newInjectableUserFS 返回带 </head> 的 user SPA 入口，让内联配置注入有落点。
func newInjectableUserFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`<!doctype html><html><head><title>u</title></head><body>user-spa</body></html>`)},
	}
}

// TestRegisterUser_SkipsInjectionForResellerHost 分销白标站的入口不能被注入主站的
// brand/seo/nav/scripts，否则首帧会闪出主站品牌、执行主站脚本；也不该为此去读 settings。
func TestRegisterUser_SkipsInjectionForResellerHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	settings := &countingConfigSource{data: map[string]interface{}{"brand": map[string]interface{}{"site_name": "Main Shop"}}}
	site := func(*http.Request) (bool, string) { return false, "r.example.test" }
	if err := RegisterUser(r, newInjectableUserFS(), settings, site); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if strings.Contains(body, "window.__PUBLIC_CONFIG__") {
		t.Fatalf("reseller host must not receive main-site inline config: %s", body)
	}
	if !strings.Contains(body, "user-spa") {
		t.Fatalf("expected raw SPA index, got: %s", body)
	}
	if got := settings.calls.Load(); got != 0 {
		t.Fatalf("GetConfig calls = %d, want 0 (non-main site must not read settings)", got)
	}
}

// TestRegisterUser_InjectsTenantForMainHost 主站入口的内联配置要带 tenant 字段，
// 与 /api/v1/public/config 同形，否则前端 isResellerTenant 首帧判错。
func TestRegisterUser_InjectsTenantForMainHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	settings := staticConfigSource{"brand": map[string]interface{}{"site_name": "Main Shop"}}
	site := func(*http.Request) (bool, string) { return true, "shop.example.test" }
	if err := RegisterUser(r, newInjectableUserFS(), settings, site); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "window.__PUBLIC_CONFIG__") {
		t.Fatalf("expected inline config on main host, got: %s", body)
	}
	if !strings.Contains(body, `"tenant":{"host":"shop.example.test","mode":"main"}`) {
		t.Fatalf("expected main tenant field in inline config, got: %s", body)
	}
}

// TestRegisterUser_NilSiteResolverOmitsTenant 没有租户解析器时视为主站照常注入，
// 但不编造 tenant 字段（保持旧调用方语义）。
func TestRegisterUser_NilSiteResolverOmitsTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	settings := staticConfigSource{"brand": map[string]interface{}{"site_name": "Main Shop"}}
	if err := RegisterUser(r, newInjectableUserFS(), settings, nil); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "window.__PUBLIC_CONFIG__") {
		t.Fatalf("expected inline config without site resolver, got: %s", body)
	}
	if strings.Contains(body, `"tenant"`) {
		t.Fatalf("nil site resolver must not emit tenant field, got: %s", body)
	}
}

// TestIsSafeTenantHost 内联 tenant.host 只允许 Host 字面量字符集；
// 其它字符（尤其 < > " 空格）一律视为不安全。
func TestIsSafeTenantHost(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		{"shop.example.test", true},
		{"127.0.0.1:8080", true},
		{"[::1]:8080", true},
		{"", true},
		{"</script ><img src=x>", false},
		{"a b.example", false},
		{`exa"mple`, false},
	}
	for _, tc := range cases {
		t.Run(tc.host, func(t *testing.T) {
			if got := isSafeTenantHost(tc.host); got != tc.want {
				t.Fatalf("isSafeTenantHost(%q) = %v, want %v", tc.host, got, tc.want)
			}
		})
	}
}

// TestRegisterUser_DropsUnsafeHostFromTenant tenant.host 是首个进入内联 <script> 的请求侧
// 数据（trusted_forwarded_host 开启时来自 X-Forwarded-Host），不能让它关闭脚本元素。
// 不安全的 host 置空，mode 仍为 main、注入照常。
func TestRegisterUser_DropsUnsafeHostFromTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	settings := staticConfigSource{"brand": map[string]interface{}{"site_name": "Main Shop"}}
	site := func(*http.Request) (bool, string) { return true, "</script ><img src=x onerror=alert(1)>" }
	if err := RegisterUser(r, newInjectableUserFS(), settings, site); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"tenant":{"host":"","mode":"main"}`) {
		t.Fatalf("unsafe host must be dropped to empty string, got: %s", body)
	}
	start := strings.Index(body, "window.__PUBLIC_CONFIG__=")
	end := strings.LastIndex(body, ";</script>")
	if start < 0 || end < 0 || end < start {
		t.Fatalf("inline config script not found in body: %s", body)
	}
	inline := body[start:end]
	if strings.Contains(inline, "</script") || strings.Contains(inline, "<img") {
		t.Fatalf("inline config leaked request-side markup: %s", inline)
	}
}

// countingConfigSource 是会统计 GetConfig 调用次数的 InitialConfigSource，
// 用来验证入口注入在 TTL 内复用 payload、不每次都读 settings。
type countingConfigSource struct {
	data  map[string]interface{}
	calls atomic.Int32
}

func (s *countingConfigSource) GetConfig(map[string]interface{}) (map[string]interface{}, error) {
	s.calls.Add(1)
	return s.data, nil
}

// TestRegisterUser_ReadsSettingsOncePerTTL 同一进程 5s 内连续的入口请求（首页、深层路由、
// 不存在的路径兜底）只读一次 settings；每个响应仍都带内联配置，证明不是靠跳过注入省掉的。
func TestRegisterUser_ReadsSettingsOncePerTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	settings := &countingConfigSource{data: map[string]interface{}{"brand": map[string]interface{}{"site_name": "Main Shop"}}}
	site := func(*http.Request) (bool, string) { return true, "shop.example.test" }
	if err := RegisterUser(r, newInjectableUserFS(), settings, site); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	for _, p := range []string{"/", "/products", "/not-exist"} {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("GET %s: status = %d", p, w.Code)
		}
		if !strings.Contains(w.Body.String(), "window.__PUBLIC_CONFIG__") {
			t.Fatalf("GET %s: expected inline config, got: %s", p, w.Body.String())
		}
	}
	if got := settings.calls.Load(); got != 1 {
		t.Fatalf("GetConfig calls = %d, want 1 (payload must be reused within TTL)", got)
	}
}

// TestRegisterUser_CachedPayloadKeepsPerRequestTenant 缓存命中的请求也要拿到自己的 tenant.host：
// 缓存复用的只是站点配置本体，tenant 必须在复制出的新 map 上按本次请求写入。
func TestRegisterUser_CachedPayloadKeepsPerRequestTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	settings := &countingConfigSource{data: map[string]interface{}{"brand": map[string]interface{}{"site_name": "Main Shop"}}}
	site := func(req *http.Request) (bool, string) { return true, req.Host }
	if err := RegisterUser(r, newInjectableUserFS(), settings, site); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	for _, host := range []string{"first.example.test", "second.example.test"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = host
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("host %s: status = %d", host, w.Code)
		}
		want := `"tenant":{"host":"` + host + `","mode":"main"}`
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("host %s: expected %s in inline config, got: %s", host, want, w.Body.String())
		}
	}
	if got := settings.calls.Load(); got != 1 {
		t.Fatalf("second request must be a cache hit, GetConfig calls = %d, want 1", got)
	}
}

// TestRegisterUser_ConcurrentRequestsShareCachedPayload 并发入口请求共享同一份缓存 payload：
// 每个响应仍是自己的 tenant.host，且 settings 只读一次（冷启动并发不打成 N 次查库）。
// 若 tenant 直接写进缓存的 map，这里会在 -race 下报并发写 map。
func TestRegisterUser_ConcurrentRequestsShareCachedPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	settings := &countingConfigSource{data: map[string]interface{}{"brand": map[string]interface{}{"site_name": "Main Shop"}}}
	site := func(req *http.Request) (bool, string) { return true, req.Host }
	if err := RegisterUser(r, newInjectableUserFS(), settings, site); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	const workers = 16
	bodies := make([]string, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = "shop" + strconv.Itoa(i) + ".example.test"
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			bodies[i] = w.Body.String()
		}(i)
	}
	wg.Wait()

	for i, body := range bodies {
		want := `"tenant":{"host":"shop` + strconv.Itoa(i) + `.example.test","mode":"main"}`
		if !strings.Contains(body, want) {
			t.Fatalf("worker %d: expected %s in inline config, got: %s", i, want, body)
		}
	}
	if got := settings.calls.Load(); got != 1 {
		t.Fatalf("GetConfig calls = %d, want 1 (concurrent misses must collapse into one read)", got)
	}
}

// SPA 入口被长期缓存会让升级后的浏览器继续加载上一版 chunk，与新后端形成契约错配，
// 因此入口必须回源校验；只有 assets/ 下带内容 hash 的产物才允许强缓存。
func TestCacheControlHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name string
		path string
		want string
	}{
		{"user index", "/", indexCacheControl},
		{"user history fallback", "/products/1", indexCacheControl},
		{"user hashed asset", "/assets/u.js", hashedAssetCacheControl},
		{"user unhashed file", "/robots.txt", indexCacheControl},
		{"admin index", "/admin/", indexCacheControl},
		{"admin history fallback", "/admin/orders/1", indexCacheControl},
		{"admin hashed asset", "/admin/assets/app.js", hashedAssetCacheControl},
	}

	r := gin.New()
	if err := RegisterAdmin(r, "/admin", newAdminFS("<html>admin-spa</html>")); err != nil {
		t.Fatalf("RegisterAdmin: %v", err)
	}
	if err := RegisterUser(r, newUserFS(), nil, nil); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d", w.Code)
			}
			if got := w.Header().Get("Cache-Control"); got != tc.want {
				t.Fatalf("Cache-Control = %q, want %q", got, tc.want)
			}
		})
	}
}

// fakeInitialConfigSource 是 InitialConfigSource 的最小测试替身，忽略 defaults 直接回放固定配置。
type fakeInitialConfigSource struct {
	data map[string]interface{}
}

func (f fakeInitialConfigSource) GetConfig(_ map[string]interface{}) (map[string]interface{}, error) {
	return f.data, nil
}

// TestBuildInitialConfigPayload_IncludesTemplateMode 保证 template_mode 随首屏内联配置下发。
// 路由里 /products、/categories/:slug 的组件解析依赖 appStore.config.template_mode，
// 该字段不在白名单时首路由会按 card 模式固化进 route record，list 模式站点首屏布局错位。
func TestBuildInitialConfigPayload_IncludesTemplateMode(t *testing.T) {
	src := fakeInitialConfigSource{data: map[string]interface{}{
		"template_mode":       "list",
		"storefront_template": "qiaozhi",
		"payment_channels":    []interface{}{map[string]interface{}{"code": "alipay"}},
	}}

	payload := buildInitialConfigPayload(context.Background(), src)

	if got, ok := payload["template_mode"]; !ok || got != "list" {
		t.Fatalf("payload[template_mode] = %v (present=%v), want \"list\"", got, ok)
	}
	if got, ok := payload["storefront_template"]; !ok || got != "qiaozhi" {
		t.Fatalf("payload[storefront_template] = %v (present=%v), want \"qiaozhi\"", got, ok)
	}
	if _, ok := payload["payment_channels"]; ok {
		t.Fatalf("payment_channels must not be inlined into first-screen payload, got %v", payload["payment_channels"])
	}
}
