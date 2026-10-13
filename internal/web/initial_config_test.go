package web

import (
	"strings"
	"testing"
)

// TestInjectInitialConfigEscapesScriptTag 验证注入的 JSON 不会被 </script> 污染破坏。
func TestInjectInitialConfigEscapesScriptTag(t *testing.T) {
	payload := map[string]interface{}{
		"storefront_template": "qiaozhi",
		"site_name":           "ai.1iiu.com</script><script>alert(1)</script>",
	}
	got, err := renderInitialConfigScript(payload)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(got, "window.__PUBLIC_CONFIG__") {
		t.Fatalf("expected window.__PUBLIC_CONFIG__ assignment, got: %s", got)
	}
	// 必须对 </script> 进行转义，避免 XSS
	if strings.Contains(got, "</script>") && !strings.Contains(got, `<\/script>`) {
		t.Fatalf("script tag not escaped: %s", got)
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
