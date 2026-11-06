package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// publicCacheControlValue 是写给 /api/v1/public/* 成功响应的 Cache-Control 头。
//
// - max-age=10  : 浏览器本地缓存 10s
// - s-maxage=10 : 共享缓存（Cloudflare 边缘节点）缓存 10s
// - public      : 允许任何缓存（含 CDN）缓存
//
// 10s 是甜蜜点:后台改商品/分类/配置后最多 10s 全球自动生效,
// 相比完全不缓存仍能消除 ~95% 的跨太平洋回源 RTT。
// 需要秒级一致性时走 URL 版本号或 CF Purge API,不要在这里拉长 TTL。
const publicCacheControlValue = "public, max-age=10, s-maxage=10"

// publicCacheVaryValue 声明公开接口响应可能随请求语言变化。
//
// 成功响应体当前不按请求语言渲染(多语言字段整包下发,由前端选语言);按 X-Lang /
// Accept-Language 本地化的只有错误体(见 i18n.ResolveLocale),而错误体已由 response.Error
// 写成 no-store。所以 Vary 在这里是兜底:为已 no-store 的错误体与未来可能按语言渲染的
// 公开字段兜住缓存键,不让 zh-CN 的响应被重放给 en-US 访客。
// 代价是尊重 Vary 的缓存(nginx / Varnish)会按 Accept-Language 碎片化;Cloudflare 忽略 Vary,不受影响。
// ?lang 查询参数本身就在缓存键里,无需声明。
const publicCacheVaryValue = "Accept-Language, X-Lang"

// PublicCacheHeaders 给挂到它身上的路由组加上 Cache-Control 头,
// 使 Cloudflare / 浏览器可以在边缘/本地缓存公开数据 10s。
//
// - 仅 GET / HEAD 请求写缓存头;组内的写操作(如 POST /affiliate/click)原样放行。
// - Vary 用 Add 追加,不覆盖 CORS 中间件已写入的 Vary: Origin。
// - 业务错误(HTTP 200 + status_code != 0)由 response.Error 覆写为 no-store,不会进入缓存。
//
// 仅挂到 /api/v1/public/* 路由组,不要挂到 /user/* 或 /admin/* ——
// 带 Authorization 的鉴权路径误加 s-maxage 会导致用户数据被串缓存。
func PublicCacheHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m := c.Request.Method; m != http.MethodGet && m != http.MethodHead {
			c.Next()
			return
		}
		c.Header("Cache-Control", publicCacheControlValue)
		c.Writer.Header().Add("Vary", publicCacheVaryValue)
		c.Next()
	}
}
