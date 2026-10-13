package web

import (
	"bytes"
	"encoding/json"
	"strings"
)

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
	return `<script>window.__PUBLIC_CONFIG__=` + safe + `;</script>`, nil
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
