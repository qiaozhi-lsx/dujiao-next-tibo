package httpserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	resellercontract "github.com/dujiao-next/internal/modules/reseller/contract"
	"github.com/gin-gonic/gin"
)

func TestConfigureTrustedProxiesRejectsSpoofedForwardedIPFromUntrustedPeer(t *testing.T) {
	engine := gin.New()
	if err := configureTrustedProxies(engine, nil); err != nil {
		t.Fatal(err)
	}
	engine.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "203.0.113.9:1234"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	if resp.Body.String() != "203.0.113.9" {
		t.Fatalf("untrusted peer spoofed client IP: %q", resp.Body.String())
	}
}

func TestConfigureTrustedProxiesAcceptsForwardedIPFromConfiguredProxy(t *testing.T) {
	engine := gin.New()
	if err := configureTrustedProxies(engine, []string{"127.0.0.1/32"}); err != nil {
		t.Fatal(err)
	}
	engine.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	if resp.Body.String() != "1.2.3.4" {
		t.Fatalf("trusted proxy client IP=%q, want 1.2.3.4", resp.Body.String())
	}
}

func TestConfigureTrustedProxiesRejectsInvalidCIDR(t *testing.T) {
	if err := configureTrustedProxies(gin.New(), []string{"not-a-cidr"}); err == nil {
		t.Fatal("expected invalid trusted proxy configuration to fail")
	}
}

func TestConfigureTrustedProxiesRejectsTrustAllNetworks(t *testing.T) {
	for _, network := range []string{"0.0.0.0/0", "::/0"} {
		t.Run(network, func(t *testing.T) {
			if err := configureTrustedProxies(gin.New(), []string{network}); err == nil {
				t.Fatalf("expected %s to be rejected", network)
			}
		})
	}
}

// SPA 壳的租户判定必须与 API 侧 ResellerTenantMiddleware 一致：只有识别出的主站才注入
// 主站内联配置；分销站、未识别域名（API 会 404）、解析出错都不注入。
func TestSiteResolverFromTenant(t *testing.T) {
	cases := []struct {
		name     string
		tenant   resellercontract.TenantContext
		err      error
		wantMain bool
		wantHost string
	}{
		{"main", resellercontract.MainTenantContext("shop.example.test"), nil, true, "shop.example.test"},
		{"reseller", resellercontract.ResellerTenantContext("r.example.test", 7, 42, "r.example.test"), nil, false, "r.example.test"},
		{"unavailable", resellercontract.UnavailableTenantContext("unknown.example.test", "not_found"), nil, false, "unknown.example.test"},
		{"error", resellercontract.MainTenantContext("shop.example.test"), errors.New("resolve failed"), false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotMain, gotHost := siteResolverFromTenant(tc.tenant, tc.err)
			if gotMain != tc.wantMain || gotHost != tc.wantHost {
				t.Fatalf("siteResolverFromTenant() = (%v, %q), want (%v, %q)", gotMain, gotHost, tc.wantMain, tc.wantHost)
			}
		})
	}
}
