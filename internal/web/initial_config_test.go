package web

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestInjectInitialConfigEscapesScriptTag 验证注入的 JSON 不会被 </script> 污染破坏。
//
// HTML 词法分析器对 </script 的匹配大小写不敏感且允许标签名后跟空白，<!-- 后再出现 <script
// 会进入 double-escaped 状态；三类向量都不能以字面量形式出现在 inline script 的 JSON 体中。
func TestInjectInitialConfigEscapesScriptTag(t *testing.T) {
	const siteName = "</SCRIPT ><script>alert(1)</script><!--<script>"
	payload := map[string]interface{}{
		"storefront_template": "qiaozhi",
		"site_name":           siteName,
	}
	got, err := renderInitialConfigScript(payload)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	const prefix = "window.__PUBLIC_CONFIG__="
	start := strings.Index(got, prefix)
	if start < 0 {
		t.Fatalf("expected window.__PUBLIC_CONFIG__ assignment, got: %s", got)
	}
	start += len(prefix)
	// 包装器自身以 ;</script> 结尾，取最后一个作为 JSON 体的右边界
	end := strings.LastIndex(got, ";</script>")
	if end < start {
		t.Fatalf("expected trailing ;</script>, got: %s", got)
	}
	body := got[start:end]
	lower := strings.ToLower(body)
	for _, vector := range []string{"</script", "<script", "<!--"} {
		if strings.Contains(lower, vector) {
			t.Fatalf("JSON body must not contain %q (case-insensitive), got: %s", vector, body)
		}
	}
	// 转义后的 JSON 必须仍然合法，且能还原出原始字符串
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("JSON body must be valid JSON: %v, body: %s", err, body)
	}
	if m["site_name"] != siteName {
		t.Fatalf("site_name round-trip mismatch: got %q want %q", m["site_name"], siteName)
	}
	// 必须带 data-cfasync="false"，否则 Cloudflare Rocket Loader 会异步化这段 inline script
	if !strings.Contains(got, `data-cfasync="false"`) {
		t.Fatalf("expected data-cfasync=\"false\" to opt out of Cloudflare Rocket Loader, got: %s", got)
	}
}

// TestInjectIntoHTMLHead 验证 HTML head 中注入了脚本。
func TestInjectIntoHTMLHead(t *testing.T) {
	src := `<!doctype html><html><head><title>t</title></head><body><div id="app"></div></body></html>`
	payload := map[string]interface{}{"storefront_template": "qiaozhi"}
	out, err := injectInitialConfigIntoHTML(src, payload)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(out, "window.__PUBLIC_CONFIG__") {
		t.Fatalf("expected injected script, got: %s", out)
	}
	// 必须在 </head> 之前
	headEnd := strings.Index(out, "</head>")
	scriptPos := strings.Index(out, "window.__PUBLIC_CONFIG__")
	if headEnd == -1 || scriptPos == -1 || scriptPos > headEnd {
		t.Fatalf("script must appear before </head>: head=%d script=%d", headEnd, scriptPos)
	}
}

// TestInjectIntoHTMLHeadNoHead 当 HTML 没有 </head> 时不改原 HTML（防御）。
func TestInjectIntoHTMLHeadNoHead(t *testing.T) {
	src := `<!doctype html><html><body>x</body></html>`
	out, err := injectInitialConfigIntoHTML(src, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out != src {
		t.Fatalf("expected unchanged html when </head> missing, got: %s", out)
	}
}

// countingBuilder 返回一个每次调用都计数、并把当前调用序号写进 "build" 字段的 build 函数。
func countingBuilder(calls *int) func() map[string]interface{} {
	return func() map[string]interface{} {
		*calls++
		return map[string]interface{}{"storefront_template": "qiaozhi", "build": *calls}
	}
}

// TestInitialConfigCache_ReusesWithinTTL TTL 内第二次 get 不再调用 build，返回第一次的内容。
func TestInitialConfigCache_ReusesWithinTTL(t *testing.T) {
	t0 := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	cache := newInitialConfigCache(5*time.Second, func() time.Time { return t0 })
	calls := 0
	build := countingBuilder(&calls)

	first := cache.get(build)
	second := cache.get(build)

	if calls != 1 {
		t.Fatalf("build calls = %d, want 1", calls)
	}
	if first["storefront_template"] != "qiaozhi" || first["build"] != 1 {
		t.Fatalf("unexpected first payload: %v", first)
	}
	if second["storefront_template"] != "qiaozhi" || second["build"] != 1 {
		t.Fatalf("second get must return the cached payload, got: %v", second)
	}
}

// TestInitialConfigCache_RefreshesAfterTTL 时钟越过 TTL 后重新 build，并返回新内容。
func TestInitialConfigCache_RefreshesAfterTTL(t *testing.T) {
	t0 := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	now := t0
	cache := newInitialConfigCache(5*time.Second, func() time.Time { return now })
	calls := 0
	build := countingBuilder(&calls)

	cache.get(build)
	cache.get(build)
	now = t0.Add(6 * time.Second)
	third := cache.get(build)

	if calls != 2 {
		t.Fatalf("build calls = %d, want 2 (one initial, one after TTL)", calls)
	}
	if third["build"] != 2 {
		t.Fatalf("expected refreshed payload after TTL, got: %v", third)
	}
}

// TestInitialConfigCache_DoesNotCacheEmpty build 返回空（settings 出错时的降级值）不缓存，
// 下一次 get 继续尝试；之后一旦拿到非空结果就正常缓存。
func TestInitialConfigCache_DoesNotCacheEmpty(t *testing.T) {
	t0 := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name  string
		empty map[string]interface{}
	}{
		{"empty map", map[string]interface{}{}},
		{"nil map", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := newInitialConfigCache(5*time.Second, func() time.Time { return t0 })
			calls := 0
			buildEmpty := func() map[string]interface{} {
				calls++
				return tc.empty
			}

			cache.get(buildEmpty)
			cache.get(buildEmpty)
			if calls != 2 {
				t.Fatalf("empty result must not be cached: build calls = %d, want 2", calls)
			}

			// 恢复后正常缓存：第三次 build，第四次命中
			build := countingBuilder(&calls)
			cache.get(build)
			got := cache.get(build)
			if calls != 3 {
				t.Fatalf("non-empty result after failures must be cached: build calls = %d, want 3", calls)
			}
			if got["build"] != 3 {
				t.Fatalf("expected the cached non-empty payload, got: %v", got)
			}
		})
	}
}
