/**
 * 可选功能开关（后台可关闭、默认开启：博客 / 公告 / 关于导航、邮箱验证码等）的读取规则。
 *
 * 这类开关在 /public/config 里的语义是"缺省或 true 即开启，只有显式 false 才关闭"，
 * 所以直接写 `value !== false` 很自然。但冷启动时 appStore.config 先是服务端内联的
 * window.__PUBLIC_CONFIG__ 子集（brand / seo / 模板等首屏字段）；旧后端的内联子集不带这些开关键，
 * `!== false` 会在一个 RTT 内读成 true，完整配置回来后再翻成 false ——
 * 导航链接、忘记密码入口、验证码字段就闪现一下又消失。
 *
 * 因此必须同时满足：可选开关键已就位（appStore.configComplete）且值不是显式 false。
 * 新后端把 nav_config / registration_enabled / email_verification_enabled 一并内联，
 * 由 inlineConfigHasOptionalFlags 判定三键齐全即视为就位，首帧就能按真实值渲染；
 * 旧后端不带键时仍等完整配置到达——"晚一拍出现"比"先出现再消失"可接受。
 */
export const optionalFlagOn = (complete: boolean, value: unknown): boolean => complete && value !== false

/** 内联 payload 必须同时带上的可选开关键；与服务端 initialConfigFieldWhitelist 里的三键对应 */
const inlineOptionalFlagKeys = ['nav_config', 'registration_enabled', 'email_verification_enabled'] as const

/**
 * 内联的 window.__PUBLIC_CONFIG__ 是否已带齐可选开关键，可直接当作"可选开关已就位"。
 *
 * 只看键是否存在（`in`），不看值：键在、值为 null / false 同样算就位 —— false 正是
 * "后台显式关闭"，必须让 optionalFlagOn 读到。缺任一键（旧后端）或 payload 不是对象都返回 false。
 */
export const inlineConfigHasOptionalFlags = (payload: unknown): boolean =>
    typeof payload === 'object' && payload !== null && inlineOptionalFlagKeys.every((key) => key in payload)
