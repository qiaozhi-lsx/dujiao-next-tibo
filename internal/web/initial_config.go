package web

import (
	"bytes"
	"encoding/json"
	"strings"
)

// renderInitialConfigScript 生成嵌入 HTML head 的 <script>window.__PUBLIC_CONFIG__ = {...};</script>
// 片段。payload 必须是可 JSON 序列化的 map，含敏感字段的调用方自己清洗。
//
// 依赖 encoding/json 默认开启的 HTML 转义：'<' '>' '&' 会被编为 \u003c \u003e \u0026，
// U+2028 / U+2029 也会被转义。输出仍是合法 JSON，浏览器把它当 JS 字面量或交给 JSON.parse
// 都能还原原值。这样一次性覆盖 HTML 词法分析器对 inline script 的全部闭合 / 双转义向量：
// </script（大小写不敏感，标签名后可跟空白，如 </SCRIPT >）、<!--（进入 escaped 状态）、
// 以及 <!-- 之后的 <script（进入 double-escaped 状态）。站长在"自定义脚本"里粘贴的 GTM
// 片段（典型含 <!-- Google Tag Manager --><script>）正是这种输入。
// 不要改回 SetEscapeHTML(false) + 手工替换 </script>：精确、小写的替换挡不住上述任何一种变体。
func renderInitialConfigScript(payload map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(payload); err != nil {
		return "", err
	}
	// json.Encoder.Encode 会追加换行符，去掉以便单行嵌入 <script>
	raw := strings.TrimRight(buf.String(), "\n")
	// data-cfasync="false" 阻止 Cloudflare Rocket Loader 把这段 inline script 异步化；
	// 否则 Rocket Loader 会推迟它到 Vue bundle 之后执行，fast-path 读到 undefined，相当于 T4 失效。
	return `<script data-cfasync="false">window.__PUBLIC_CONFIG__=` + raw + `;</script>`, nil
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
