/**
 * 可选功能开关（后台可关闭、默认开启：博客 / 公告 / 关于导航、邮箱验证码等）的读取规则。
 *
 * 这类开关在 /public/config 里的语义是"缺省或 true 即开启，只有显式 false 才关闭"，
 * 所以直接写 `value !== false` 很自然。但冷启动时 appStore.config 先是服务端内联的
 * window.__PUBLIC_CONFIG__ 子集（brand / seo / 模板等首屏字段），这些开关键根本不存在，
 * `!== false` 会在一个 RTT 内读成 true，完整配置回来后再翻成 false ——
 * 导航链接、忘记密码入口、验证码字段就闪现一下又消失。
 *
 * 因此必须同时满足：完整配置已就位（appStore.configComplete）且值不是显式 false。
 * 代价是开关为开时要等完整配置到达才显示；"晚一拍出现"比"先出现再消失"可接受。
 */
export const optionalFlagOn = (complete: boolean, value: unknown): boolean => complete && value !== false
