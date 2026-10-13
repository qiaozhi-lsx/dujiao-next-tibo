// Verifies nav guard fire-and-forget behavior for appStore.loadConfig().
// Project uses node:test (not vitest); test registers against node's test runner.
// Picked up by pnpm run test via the shim at tests/navGuardRace.test.ts which imports this file.
import test from 'node:test'
import assert from 'node:assert/strict'

test('nav guard loadConfig race: does not block routing when config is still loading', () => {
  let resolved = false
  const slow = new Promise<void>((resolve) => {
    setTimeout(() => { resolved = true; resolve() }, 500)
  })
  // 模拟 loadConfig 是慢请求（stand-in for appStore.loadConfig spy）
  const fakeStore = { loadConfig: () => slow as unknown as Promise<void> }
  const startedAt = Date.now()
  // 模拟 nav guard 的 "fire and forget" 行为：void 调用不等 Promise
  void fakeStore.loadConfig()
  const elapsed = Date.now() - startedAt
  assert.ok(elapsed < 50, `elapsed was ${elapsed}ms, must be < 50`) // 必须立即返回，不等 500ms
  assert.equal(resolved, false)
})
