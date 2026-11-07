// Package navconfig 提供 nav_config 未配置时的默认导航结构。
// web 包与 settings 公开 handler 共用此默认值,确保内联首帧与完整 /api/v1/public/config
// 响应保持同形,否则前端会先按一种导航渲染再翻转。
package navconfig

// Default 返回 nav_config 未配置时的默认导航结构:三个内置导航全开、无自定义项。
// 每次返回新 map,避免调用方共享可变状态(map 引用语义)。
func Default() map[string]any {
	return map[string]any{
		"builtin":      map[string]any{"blog": true, "notice": true, "about": true},
		"custom_items": make([]any, 0),
	}
}
