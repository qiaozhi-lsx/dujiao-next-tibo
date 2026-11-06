package web

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// initialConfigCache 是 SPA 壳内联配置 payload 的进程内 TTL 缓存。
//
// 入口 HTML 是每次刷新 / 深链接 / 404 兜底都会命中的请求，每次都去 settings 读库会在
// SQLite（MaxOpenConns=1）上和业务请求抢同一条连接。缓存只保存站点配置本体；
// 调用方拿到的 map 在请求间共享，必须视为只读，按请求变化的字段（tenant）要先复制再写。
type initialConfigCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	now     func() time.Time
	expires time.Time
	payload map[string]interface{}
}

// newInitialConfigCache 构造 ttl 内复用结果的缓存；now 为 nil 时用 time.Now，测试可注入假时钟。
func newInitialConfigCache(ttl time.Duration, now func() time.Time) *initialConfigCache {
	if now == nil {
		now = time.Now
	}
	return &initialConfigCache{ttl: ttl, now: now}
}

// get 未过期时返回缓存的 payload；否则调用 build，结果非空才缓存并把有效期设为 now()+ttl。
// 空结果（settings 出错 / 为空时的降级值）直接返回不缓存，下一次请求会再试。
//
// build 在锁内执行：过期瞬间的并发请求只会有一个真正去读 settings，其余等它写回后直接复用，
// 不会打成 N 次查库；代价是 build 期间其它入口请求要等，但它们本来也都需要这份新结果。
func (c *initialConfigCache) get(build func() map[string]interface{}) map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	if c.payload != nil && now.Before(c.expires) {
		return c.payload
	}
	payload := build()
	if len(payload) == 0 {
		return payload
	}
	c.payload = payload
	c.expires = now.Add(c.ttl)
	return payload
}

// renderInitialConfigScript 生成嵌入 HTML head 的 <script>window.__PUBLIC_CONFIG__ = {...};</script>
// 片段。payload 必须是可 JSON 序列化的 map，含敏感字段的调用方自己清洗。
//
// 关闭 json 默认的 HTML 转义（否则 '<' 会被编码成 <，看起来安全但不方便读也挡不住
// 其它注入向量），改为自己把 </script> 显式转义为 <\/script>——这才是 HTML 内嵌 JSON 的
// 标准做法：JSON 字符串字面量里的 "<\/" 和 "</" 等价，但 HTML parser 不会提前关闭 script。
func renderInitialConfigScript(payload map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(payload); err != nil {
		return "", err
	}
	// json.Encoder.Encode 会追加换行符，去掉以便单行嵌入 <script>
	raw := strings.TrimRight(buf.String(), "\n")
	safe := strings.ReplaceAll(raw, "</script>", `<\/script>`)
	// data-cfasync="false" 阻止 Cloudflare Rocket Loader 把这段 inline script 异步化；
	// 否则 Rocket Loader 会推迟它到 Vue bundle 之后执行，fast-path 读到 undefined，相当于 T4 失效。
	return `<script data-cfasync="false">window.__PUBLIC_CONFIG__=` + safe + `;</script>`, nil
}

// injectInitialConfigIntoHTML 把 initial config 脚本插入到 </head> 之前。
// 若 HTML 没有 </head>（例如 404 fallback 到非 SPA 页面），原样返回。
func injectInitialConfigIntoHTML(html string, payload map[string]interface{}) (string, error) {
	idx := strings.Index(html, "</head>")
	if idx < 0 {
		return html, nil
	}
	script, err := renderInitialConfigScript(payload)
	if err != nil {
		return "", err
	}
	return html[:idx] + script + html[idx:], nil
}
