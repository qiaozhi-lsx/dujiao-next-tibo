package middleware

import "github.com/gin-gonic/gin"

// publicCacheControlValue 是写给 /api/v1/public/* 响应的 Cache-Control 头。
//
// - max-age=60  : 浏览器本地缓存 60s
// - s-maxage=60 : 共享缓存（Cloudflare 边缘节点）缓存 60s
// - public      : 允许任何缓存（含 CDN）缓存
//
// 60s 是经验值:公开商品/分类/配置改动极少,边缘命中 60s 内的刷新延迟可接受,
// 但完全消除冷启动跨太平洋回源的 RTT。handler 若需要更严格的 no-cache
// 语义,可以在业务逻辑里覆盖此值。
const publicCacheControlValue = "public, max-age=60, s-maxage=60"

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
