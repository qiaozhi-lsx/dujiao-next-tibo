package captchahttp

import (
	"errors"

	captchacontract "github.com/dujiao-next/internal/modules/captcha/contract"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

// ImageChallengeGenerator 生成图片验证码挑战。
type ImageChallengeGenerator interface {
	GenerateImageChallenge() (*captchacontract.ImageChallenge, error)
}

// PublicHandler 处理公开验证码请求。
type PublicHandler struct {
	generator ImageChallengeGenerator
}

func NewPublicHandler(generator ImageChallengeGenerator) *PublicHandler {
	return &PublicHandler{generator: generator}
}

// GetImageCaptcha 获取图片验证码挑战。
func (h *PublicHandler) GetImageCaptcha(c *gin.Context) {
	// 验证码是一次性挑战，且错误分支同样返回 HTTP 200 + 业务错误码；
	// 必须在任何响应路径之前覆盖 /api/v1/public 路由组写入的 public 缓存头，
	// 否则 CDN/浏览器会在 10s 内把同一个 captcha_id（或同一份错误）复用给刷新请求。
	c.Header("Cache-Control", "no-store")

	if h == nil || h.generator == nil {
		ginutil.RespondError(c, response.CodeInternal, "error.captcha_unavailable", captchacontract.ErrConfigInvalid)
		return
	}

	challenge, err := h.generator.GenerateImageChallenge()
	if err != nil {
		switch {
		case errors.Is(err, captchacontract.ErrConfigInvalid):
			ginutil.RespondError(c, response.CodeBadRequest, "error.captcha_unavailable", nil)
		default:
			ginutil.RespondError(c, response.CodeInternal, "error.captcha_generate_failed", err)
		}
		return
	}

	response.Success(c, gin.H{
		"captcha_id":   challenge.CaptchaID,
		"image_base64": challenge.ImageBase64,
	})
}
