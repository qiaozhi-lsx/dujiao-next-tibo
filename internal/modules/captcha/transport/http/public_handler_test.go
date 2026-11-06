package captchahttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	captchacontract "github.com/dujiao-next/internal/modules/captcha/contract"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

type fakeGenerator struct {
	challenge *captchacontract.ImageChallenge
	err       error
}

func (f fakeGenerator) GenerateImageChallenge() (*captchacontract.ImageChallenge, error) {
	return f.challenge, f.err
}

func TestGetImageCaptchaSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPublicHandler(fakeGenerator{challenge: &captchacontract.ImageChallenge{
		CaptchaID:   "id-1",
		ImageBase64: "data:image/png;base64,abc",
	}})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/captcha/image", nil)

	handler.GetImageCaptcha(c)

	var got response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.StatusCode != response.CodeOK {
		t.Fatalf("status want OK got %d body=%s", got.StatusCode, w.Body.String())
	}
	data, ok := got.Data.(map[string]interface{})
	if !ok || data["captcha_id"] != "id-1" {
		t.Fatalf("unexpected data %#v", got.Data)
	}
}

func TestGetImageCaptchaUnavailableWhenConfigInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPublicHandler(fakeGenerator{err: captchacontract.ErrConfigInvalid})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/captcha/image", nil)

	handler.GetImageCaptcha(c)

	var got response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.StatusCode != response.CodeBadRequest {
		t.Fatalf("status want 400 got %d", got.StatusCode)
	}
}

func TestGetImageCaptchaUnavailableWithoutGenerator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPublicHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/captcha/image", nil)

	handler.GetImageCaptcha(c)

	var got response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.StatusCode != response.CodeInternal {
		t.Fatalf("status want internal got %d", got.StatusCode)
	}
}

func TestGetImageCaptchaMapsUnknownErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPublicHandler(fakeGenerator{err: errors.New("boom")})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/captcha/image", nil)

	handler.GetImageCaptcha(c)

	var got response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.StatusCode != response.CodeInternal {
		t.Fatalf("status want internal got %d", got.StatusCode)
	}
}

// TestGetImageCaptcha_SetsNoStoreCacheControl 验证验证码图片接口在所有响应路径
// （成功、配置无效、未知错误、generator 缺失）都显式写 Cache-Control: no-store，
// 并覆盖上游 /api/v1/public 路由组中间件预先写入的 public 缓存头。
// 验证码是一次性挑战，错误分支同样是 HTTP 200 + 业务错误码，一旦被 CDN/浏览器
// 缓存 10s，用户刷新验证码会拿到同一个 captcha_id 或同一份错误响应。
func TestGetImageCaptcha_SetsNoStoreCacheControl(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name      string
		generator ImageChallengeGenerator
		wantCode  int
	}{
		{
			name: "success",
			generator: fakeGenerator{challenge: &captchacontract.ImageChallenge{
				CaptchaID:   "id-1",
				ImageBase64: "data:image/png;base64,abc",
			}},
			wantCode: response.CodeOK,
		},
		{
			name:      "config invalid",
			generator: fakeGenerator{err: captchacontract.ErrConfigInvalid},
			wantCode:  response.CodeBadRequest,
		},
		{
			name:      "unknown error",
			generator: fakeGenerator{err: errors.New("boom")},
			wantCode:  response.CodeInternal,
		},
		{
			name:      "nil generator",
			generator: nil,
			wantCode:  response.CodeInternal,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewPublicHandler(tc.generator)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/captcha/image", nil)
			// 模拟 /api/v1/public 路由组上的 PublicCacheHeaders() 在 c.Next() 之前写入的默认缓存头。
			c.Header("Cache-Control", "public, max-age=10, s-maxage=10")

			handler.GetImageCaptcha(c)

			var got response.Response
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got.StatusCode != tc.wantCode {
				t.Fatalf("status want %d got %d body=%s", tc.wantCode, got.StatusCode, w.Body.String())
			}
			if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
				t.Fatalf("Cache-Control = %q, want %q", cc, "no-store")
			}
			if vals := w.Header().Values("Cache-Control"); len(vals) != 1 {
				t.Fatalf("Cache-Control should be a single value, got %q", vals)
			}
		})
	}
}
