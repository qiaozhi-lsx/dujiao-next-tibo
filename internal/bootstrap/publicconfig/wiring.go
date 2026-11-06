package publicconfigwiring

import (
	"github.com/dujiao-next/internal/app/container"
	publicconfigtransport "github.com/dujiao-next/internal/modules/settings/transport/http/public"
	"github.com/dujiao-next/internal/web"
)

// NewInitialConfigSource 返回 SPA 入口内联 window.__PUBLIC_CONFIG__ 的设置来源。
// 复用 /api/v1/public/config 的 settings 适配器，使入口同时拿到 web.InitialConfigFlagsSource
// 提供的 nav_config 与注册 / 邮箱验证开关；容器里没有 SettingService 时返回 nil，
// RegisterUser 据此退回不注入而不是在请求期间解引用空服务。
func NewInitialConfigSource(c *container.Container) web.InitialConfigSource {
	if c == nil || c.SettingService == nil {
		return nil
	}
	return publicConfigSettingsAdapter{settings: c.SettingService, cfg: c.Config}
}

func NewHandler(c *container.Container) *publicconfigtransport.Handler {
	var captcha publicconfigtransport.CaptchaPublic
	if c.CaptchaService != nil {
		captcha = publicConfigCaptchaAdapter{svc: c.CaptchaService}
	}
	var telegram publicconfigtransport.TelegramAuthPublic
	if c.TelegramAuthService != nil {
		telegram = publicConfigTelegramAdapter{svc: c.TelegramAuthService}
	}
	var google publicconfigtransport.GoogleAuthPublic
	if c.GoogleAuthService != nil {
		google = publicConfigGoogleAdapter{svc: c.GoogleAuthService}
	}
	var overlay publicconfigtransport.ResellerOverlay
	if c.ResellerSiteConfigService != nil {
		overlay = publicConfigResellerOverlayAdapter{svc: c.ResellerSiteConfigService}
	}
	fallback := publicconfigtransport.TelegramAuthFallback{}
	googleFallback := publicconfigtransport.GoogleAuthFallback{}
	if c.Config != nil {
		fallback = publicconfigtransport.TelegramAuthFallback{
			Enabled:     c.Config.TelegramAuth.Enabled,
			BotUsername: c.Config.TelegramAuth.BotUsername,
			MiniAppURL:  c.Config.TelegramAuth.MiniAppURL,
		}
		googleFallback = publicconfigtransport.GoogleAuthFallback{
			Enabled:  c.Config.GoogleAuth.Enabled,
			ClientID: c.Config.GoogleAuth.ClientID,
		}
	}
	return publicconfigtransport.NewHandler(
		publicConfigCacheAdapter{},
		publicConfigSettingsAdapter{settings: c.SettingService, cfg: c.Config},
		publicConfigPaymentAdapter{payments: c.PaymentService},
		captcha,
		telegram,
		fallback,
		google,
		googleFallback,
		overlay,
	)
}
