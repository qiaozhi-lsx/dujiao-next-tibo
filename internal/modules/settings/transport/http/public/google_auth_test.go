package publicconfighttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/dujiao-next/internal/shared/jsonmap"

	"github.com/gin-gonic/gin"
)

type googleAuthPublicStub map[string]interface{}

func (s googleAuthPublicStub) PublicConfig() map[string]interface{} {
	return map[string]interface{}(s)
}

func TestResolveGoogleAuthPublicConfig(t *testing.T) {
	tests := []struct {
		name     string
		source   GoogleAuthPublic
		fallback GoogleAuthFallback
		want     map[string]interface{}
	}{
		{
			name:     "fallback enabled",
			fallback: GoogleAuthFallback{Enabled: true, ClientID: " fallback-client "},
			want:     map[string]interface{}{"enabled": true, "client_id": "fallback-client"},
		},
		{
			name:     "service overrides fallback",
			source:   googleAuthPublicStub{"enabled": true, "client_id": " runtime-client ", "unexpected": "private"},
			fallback: GoogleAuthFallback{Enabled: false, ClientID: "fallback-client"},
			want:     map[string]interface{}{"enabled": true, "client_id": "runtime-client"},
		},
		{
			name:     "missing client id fails closed",
			source:   googleAuthPublicStub{"enabled": true, "client_id": ""},
			fallback: GoogleAuthFallback{Enabled: true, ClientID: "fallback-client"},
			want:     map[string]interface{}{"enabled": false, "client_id": ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := resolveGoogleAuthPublicConfig(test.source, test.fallback); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("public config = %#v, want %#v", got, test.want)
			}
		})
	}
}

// ---- GetConfig 缓存头 ----
//
// 以下测试与 Google 登录无关,但 internal/architecture 给本目录的 Go 文件预算是 3
// (handler.go / routes.go / 本文件),无法再新增测试文件,故放在这里。

// fakeConfigCache 以 JSON 往返模拟 Redis 缓存,记录命中/写入次数以证明两条成功路径都被覆盖。
type fakeConfigCache struct {
	entries map[string][]byte
	hits    int
	sets    int
}

func newFakeConfigCache() *fakeConfigCache {
	return &fakeConfigCache{entries: map[string][]byte{}}
}

func (f *fakeConfigCache) CacheKey(resellerID *uint) string {
	if resellerID == nil {
		return "public_config:main"
	}
	return fmt.Sprintf("public_config:%d", *resellerID)
}

func (f *fakeConfigCache) GetJSON(_ context.Context, key string, dest *map[string]interface{}) (bool, error) {
	raw, ok := f.entries[key]
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return false, err
	}
	f.hits++
	return true, nil
}

func (f *fakeConfigCache) SetJSON(_ context.Context, key string, value interface{}, _ time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f.entries[key] = raw
	f.sets++
	return nil
}

// stubPublicSettings 返回最小可用的公开配置;configErr 非空时 GetConfig 失败,用于覆盖错误路径。
type stubPublicSettings struct {
	configErr error
}

func (s stubPublicSettings) GetConfig(defaults map[string]interface{}) (map[string]interface{}, error) {
	if s.configErr != nil {
		return nil, s.configErr
	}
	data := make(map[string]interface{}, len(defaults))
	for key, value := range defaults {
		data[key] = value
	}
	return data, nil
}

func (stubPublicSettings) GetWalletRechargeChannelIDs() []uint { return nil }
func (stubPublicSettings) GetWalletOnlyPayment() bool          { return false }
func (stubPublicSettings) GetAffiliateSettingMap() (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (stubPublicSettings) GetSMTPEnabled() bool { return false }
func (stubPublicSettings) GetRegistrationEnabled(defaultValue bool) (bool, error) {
	return defaultValue, nil
}
func (stubPublicSettings) GetEmailVerificationEnabled(defaultValue bool) (bool, error) {
	return defaultValue, nil
}
func (stubPublicSettings) GetRegistrationEmailDomainPolicy() (bool, []string, error) {
	return false, nil, nil
}
func (stubPublicSettings) GetByKey(string) (interface{}, error)            { return nil, nil }
func (stubPublicSettings) GetActiveHomeAnnouncement() (jsonmap.JSON, bool) { return nil, false }

type stubPaymentChannels struct{}

func (stubPaymentChannels) GetOrderPaymentChannels() ([]map[string]interface{}, error) {
	return nil, nil
}

// publicGroupCacheControl 是 /api/v1/public 路由组中间件(internal/app/httpserver/middleware.PublicCacheHeaders)
// 写入的默认缓存头。模块 transport 不应反向依赖 internal/app,这里用内联中间件模拟。
const publicGroupCacheControl = "public, max-age=10, s-maxage=10"

func newPublicConfigRouter(t *testing.T, handler *Handler) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	public := engine.Group("/api/v1/public")
	public.Use(func(c *gin.Context) {
		c.Header("Cache-Control", publicGroupCacheControl)
		c.Next()
	})
	RegisterPublicRoutes(public, handler)
	return engine
}

func newPublicConfigHandler(cache ConfigCache, settings Settings) *Handler {
	return NewHandler(cache, settings, stubPaymentChannels{}, nil, nil, TelegramAuthFallback{}, nil, GoogleAuthFallback{}, nil)
}

type publicConfigBody struct {
	StatusCode int                    `json:"status_code"`
	Data       map[string]interface{} `json:"data"`
}

func getPublicConfig(t *testing.T, engine *gin.Engine) (*httptest.ResponseRecorder, publicConfigBody) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/config", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body publicConfigBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v; body=%s", err, rec.Body.String())
	}
	return rec, body
}

// TestPublicConfig_BrowserNoCacheEdgeCacheHeader 验证 /public/config 的两条成功路径
// (缓存未命中现算、缓存命中)都把路由组写的默认缓存头覆盖为
// "public, max-age=0, s-maxage=10":浏览器不缓存、CDN 边缘缓存 10s 并带 Age 头,
// 前端据此补偿响应体里每次现算的 server_time。
func TestPublicConfig_BrowserNoCacheEdgeCacheHeader(t *testing.T) {
	cache := newFakeConfigCache()
	engine := newPublicConfigRouter(t, newPublicConfigHandler(cache, stubPublicSettings{}))

	const want = "public, max-age=0, s-maxage=10"
	steps := []struct {
		name     string
		wantHits int
		wantSets int
	}{
		{name: "cache miss", wantHits: 0, wantSets: 1},
		{name: "cache hit", wantHits: 1, wantSets: 1},
	}
	for _, step := range steps {
		rec, body := getPublicConfig(t, engine)
		if cache.hits != step.wantHits || cache.sets != step.wantSets {
			t.Fatalf("%s: cache hits=%d sets=%d, want hits=%d sets=%d", step.name, cache.hits, cache.sets, step.wantHits, step.wantSets)
		}
		if got := rec.Header().Get("Cache-Control"); got != want {
			t.Errorf("%s: Cache-Control = %q, want %q", step.name, got, want)
		}
		if body.StatusCode != 0 {
			t.Errorf("%s: status_code = %d, want 0", step.name, body.StatusCode)
		}
		if serverTime, _ := body.Data["server_time"].(float64); serverTime <= 0 {
			t.Errorf("%s: server_time = %v, want a positive unix millisecond timestamp", step.name, body.Data["server_time"])
		}
	}
}

// TestPublicConfig_ErrorResponseStaysNoStore 验证错误路径不受成功路径缓存头影响:
// 配置读取失败时仍由 response.Error 写 no-store,错误响应不会进入浏览器/CDN 缓存。
func TestPublicConfig_ErrorResponseStaysNoStore(t *testing.T) {
	settings := stubPublicSettings{configErr: errors.New("settings unavailable")}
	engine := newPublicConfigRouter(t, newPublicConfigHandler(newFakeConfigCache(), settings))

	rec, body := getPublicConfig(t, engine)
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want %q", got, "no-store")
	}
	if body.StatusCode == 0 {
		t.Errorf("status_code = 0, want a business error code")
	}
}
