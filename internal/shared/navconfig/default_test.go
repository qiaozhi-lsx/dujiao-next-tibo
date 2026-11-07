package navconfig

import (
	"reflect"
	"testing"
)

// TestDefault_Shape 默认 nav_config 结构必须与 /api/v1/public/config 的内联 fallback 完全一致:
// builtin 三键全 true、custom_items 为空数组。前端对这个 schema 有硬期待,偏离一字段都会让
// 内联首帧与完整配置不同形。
func TestDefault_Shape(t *testing.T) {
	got := Default()

	want := map[string]any{
		"builtin":      map[string]any{"blog": true, "notice": true, "about": true},
		"custom_items": make([]any, 0),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Default() = %#v, want %#v", got, want)
	}
}

// TestDefault_ReturnsFreshMap 每次调用必须返回独立的 map 与 slice,否则调用方写 custom_items
// 或改 builtin 子键会污染后续调用的默认值。
func TestDefault_ReturnsFreshMap(t *testing.T) {
	first := Default()
	first["custom_items"] = append(first["custom_items"].([]any), "mutated")
	first["builtin"].(map[string]any)["blog"] = false

	second := Default()
	if items, ok := second["custom_items"].([]any); !ok || len(items) != 0 {
		t.Fatalf("second Default() custom_items = %#v, want empty slice (previous call leaked state)", second["custom_items"])
	}
	if builtin, ok := second["builtin"].(map[string]any); !ok || builtin["blog"] != true {
		t.Fatalf("second Default() builtin = %#v, want blog=true (previous call leaked state)", second["builtin"])
	}
}
