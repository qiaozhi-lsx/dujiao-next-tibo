package middleware

import "github.com/gin-gonic/gin"

// publicCacheControlValue 是写给 /api/v1/public/* 响应的 Cache-Control 头。
//
// - max-age=10  : 浏览器本地缓存 10s
// - s-maxage=10 : 共享缓存（Cloudflare 边缘节点）缓存 10s
// - public      : 允许任何缓存（含 CDN）缓存
//
// 10s 是甜蜜点:后台改商品/分类/配置后最多 10s 全球自动生效,
// 相比完全不缓存仍能消除 ~95% 的跨太平洋回源 RTT。
// 需要秒级一致性时走 URL 版本号或 CF Purge API,不要在这里拉长 TTL。
const publicCacheControlValue = "public, max-age=10, s-maxage=10"

// PublicCacheHeaders 给挂到它身上的路由组加上 Cache-Control 头,
// 使 Cloudflare / 浏览器可以在边缘/本地缓存公开数据 60s。
//
// 仅挂到 /api/v1/public/* 路由组,不要挂到 /user/* 或 /admin/* ——
// 带 Authorization 的鉴权路径误加 s-maxage 会导致用户数据被串缓存。
func PublicCacheHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", publicCacheControlValue)
		c.Next()
	}
}
