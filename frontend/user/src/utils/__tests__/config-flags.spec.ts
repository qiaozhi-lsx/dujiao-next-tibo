// Verifies optionalFlagOn, the helper that gates "default-on" optional feature
// flags (nav blog / notice / about, email_verification_enabled) behind
// appStore.configComplete. The server-inlined window.__PUBLIC_CONFIG__ is only a
// subset of /public/config and lacks those keys, so reading `value !== false` on
// it yields true for one RTT and then flips to false once the full config lands:
// the blog link, forgot-password link and register code field flash in and vanish.
// The flag must therefore stay off until a full config has been fetched.
//
// Project uses node:test (not vitest) + no DOM.
// Picked up by pnpm run test via the shim at tests/configFlags.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { optionalFlagOn } from '../configFlags.ts'

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
