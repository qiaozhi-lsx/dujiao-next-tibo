// Verifies optionalFlagOn, the helper that gates "default-on" optional feature
// flags (nav blog / notice / about, email_verification_enabled) behind
// appStore.configComplete. On an old backend the server-inlined
// window.__PUBLIC_CONFIG__ is only a subset of /public/config and lacks those keys,
// so reading `value !== false` on it yields true for one RTT and then flips to
// false once the full config lands: the blog link, forgot-password link and
// register code field flash in and vanish. The flag must therefore stay off until
// a full config has been fetched — or until the inline payload itself carries the
// keys (see inlineConfigHasOptionalFlags below).
//
// Project uses node:test (not vitest) + no DOM.
// Picked up by pnpm run test via the shim at tests/configFlags.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { inlineConfigHasOptionalFlags, optionalFlagOn } from '../configFlags.ts'

test('optionalFlagOn: incomplete config keeps a missing flag off', () => {
  assert.equal(optionalFlagOn(false, undefined), false)
})

test('optionalFlagOn: incomplete config keeps even an explicit true off', () => {
  assert.equal(optionalFlagOn(false, true), false)
})

test('optionalFlagOn: complete config treats a missing flag as on', () => {
  assert.equal(optionalFlagOn(true, undefined), true)
})

test('optionalFlagOn: complete config honours an explicit false', () => {
  assert.equal(optionalFlagOn(true, false), false)
})

test('optionalFlagOn: complete config honours an explicit true', () => {
  assert.equal(optionalFlagOn(true, true), true)
})

// composables/useNavConfig.ts 的 builtinNavItems（classic Navbar / VaultLayout 共用）对
// 博客 / 公告 / 关于三项用的就是这一条规则：`optionalFlagOn(appStore.configComplete, builtin?.[key])`。
// 该 composable 依赖 useI18n / pinia / lucide，node:test 里无法直接调用，这里复刻那段过滤循环
// 作为行为说明——改一处要同步改另一处。
const builtinNavKeys = ['blog', 'notice', 'about']
const visibleBuiltinNavKeys = (complete: boolean, builtin: Record<string, boolean> | undefined): string[] =>
  builtinNavKeys.filter((key) => optionalFlagOn(complete, builtin?.[key]))

test('builtinNavItems rule: inline config (nav_config absent) yields no builtin items', () => {
  // 旧写法 `builtin && builtin[key] === false` 在这里会放行全部三项，然后在完整配置到达时收回
  assert.deepEqual(visibleBuiltinNavKeys(false, undefined), [])
})

test('builtinNavItems rule: complete config without nav_config shows all three', () => {
  assert.deepEqual(visibleBuiltinNavKeys(true, undefined), ['blog', 'notice', 'about'])
})

test('builtinNavItems rule: complete config drops only the explicitly disabled items', () => {
  assert.deepEqual(visibleBuiltinNavKeys(true, { blog: false, notice: true }), ['notice', 'about'])
})

test('builtinNavItems rule: items never flash — empty during the inline phase, settled once the full config lands', () => {
  const inlinePhase = visibleBuiltinNavKeys(false, undefined)
  const completePhase = visibleBuiltinNavKeys(true, { blog: true, notice: false, about: true })
  assert.deepEqual(inlinePhase, [])
  assert.deepEqual(completePhase, ['blog', 'about'])
})

// inlineConfigHasOptionalFlags：stores/app.ts 的 fast path 用它决定内联的 window.__PUBLIC_CONFIG__
// 能否直接当作"可选开关键已就位"（configComplete = true）。新后端把 nav_config /
// registration_enabled / email_verification_enabled 三键一并内联，前端首帧即可信任；
// 旧后端的内联子集不带这三键，仍要等后台完整配置。只看键是否存在（`in`），不看值：
// 键在、值为 null / false 也算就位（false 正是"后台显式关闭"，必须被 optionalFlagOn 读到）。
const fullInlinePayload = () => ({
  storefront_template: 'qiaozhi',
  brand: {},
  nav_config: { builtin: { blog: false } },
  registration_enabled: true,
  email_verification_enabled: false,
})

test('inlineConfigHasOptionalFlags: payload carrying all three keys is trusted', () => {
  assert.equal(inlineConfigHasOptionalFlags(fullInlinePayload()), true)
})

test('inlineConfigHasOptionalFlags: old-server payload without the keys is not trusted', () => {
  assert.equal(inlineConfigHasOptionalFlags({ storefront_template: 'qiaozhi', brand: {}, seo: {} }), false)
  assert.equal(inlineConfigHasOptionalFlags({}), false)
})

test('inlineConfigHasOptionalFlags: any single missing key is enough to distrust the payload', () => {
  for (const missing of ['nav_config', 'registration_enabled', 'email_verification_enabled'] as const) {
    const payload: Record<string, unknown> = fullInlinePayload()
    delete payload[missing]
    assert.equal(inlineConfigHasOptionalFlags(payload), false, `missing ${missing} must yield false`)
  }
})

test('inlineConfigHasOptionalFlags: null and non-object payloads are not trusted', () => {
  assert.equal(inlineConfigHasOptionalFlags(null), false)
  assert.equal(inlineConfigHasOptionalFlags(undefined), false)
  assert.equal(inlineConfigHasOptionalFlags('nav_config'), false)
  assert.equal(inlineConfigHasOptionalFlags(0), false)
  assert.equal(inlineConfigHasOptionalFlags(true), false)
})

test('inlineConfigHasOptionalFlags: keys present with null / false values still count as present', () => {
  assert.equal(
    inlineConfigHasOptionalFlags({ nav_config: null, registration_enabled: false, email_verification_enabled: false }),
    true,
  )
  assert.equal(
    inlineConfigHasOptionalFlags({ nav_config: undefined, registration_enabled: null, email_verification_enabled: null }),
    true,
  )
})

test('inlineConfigHasOptionalFlags + optionalFlagOn: trusted inline payload honours an inline false on the first frame', () => {
  // 端到端的组合语义：内联带键 → complete 为 true → 内联里的显式 false 立即生效、缺省的 about 立即开启，
  // 不再出现"首帧全关、一个 RTT 后才出现"的晚一拍
  const payload = fullInlinePayload()
  const complete = inlineConfigHasOptionalFlags(payload)
  assert.equal(complete, true)
  assert.deepEqual(visibleBuiltinNavKeys(complete, payload.nav_config.builtin), ['notice', 'about'])
  assert.equal(optionalFlagOn(complete, payload.email_verification_enabled), false)
})
