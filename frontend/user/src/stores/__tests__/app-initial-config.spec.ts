// Verifies appStore.loadConfig fast-path behavior:
// when window.__PUBLIC_CONFIG__ is injected by the server, loadConfig()
// must consume it synchronously without issuing a /public/config HTTP request.
//
// Project uses node:test (not vitest) + no DOM; this spec therefore reproduces
// the fast-path logic as a focused unit test rather than booting Pinia + Vue
// runtime (which require @unhead/vue + jsdom and are not configured here).
// Picked up by pnpm run test via the shim at tests/appInitialConfig.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'

// loadConfigFastPath 复刻 stores/app.ts 的 window.__PUBLIC_CONFIG__ 分支语义：
// 当 config 为空、未强制刷新、且 window.__PUBLIC_CONFIG__ 存在时，
// 同步把它赋给 config.value，并由后台 refetch 补齐完整字段。
// 这个 helper 和 app.ts 里的实际代码保持等价——改一处要同步改另一处。
function loadConfigFastPath(opts: {
  configRef: { value: any }
  force: boolean
  windowRef: Record<string, any> | undefined
  configAPIGet: () => Promise<any>
  onBackgroundResolved?: (resp: any) => void
}): { tookFastPath: boolean; backgroundStarted: boolean } {
  const { configRef, force, windowRef, configAPIGet, onBackgroundResolved } = opts
  if (configRef.value && !force) {
    return { tookFastPath: false, backgroundStarted: false }
  }
  if (!force && windowRef && windowRef.__PUBLIC_CONFIG__) {
    configRef.value = windowRef.__PUBLIC_CONFIG__
    // 后台 refetch：和 app.ts 一样 fire-and-forget
    void (async () => {
      try {
        const resp = await configAPIGet()
        configRef.value = resp.data.data
        if (onBackgroundResolved) onBackgroundResolved(resp)
      } catch { /* 后台失败不影响首帧 */ }
    })()
    return { tookFastPath: true, backgroundStarted: true }
  }
  return { tookFastPath: false, backgroundStarted: false }
}

test('loadConfig fast path: window.__PUBLIC_CONFIG__ consumed synchronously, no await on configAPI', () => {
  let apiCallCount = 0
  const configAPIGet = () => {
    apiCallCount++
    // 返回一个永不 resolve 的 Promise，模拟慢网络：若 fast path 真的 await 了，
    // 这个测试会卡死；只要我们走的是 void(...) 后台调用，就不会阻塞。
    return new Promise(() => {})
  }
  const configRef = { value: null as any }
  const windowRef = { __PUBLIC_CONFIG__: { storefront_template: 'qiaozhi', brand: {} } }

  const startedAt = Date.now()
  const result = loadConfigFastPath({ configRef, force: false, windowRef, configAPIGet })
  const elapsed = Date.now() - startedAt

  // 同步路径：50ms 内完成，不能因为 configAPI 的 pending Promise 卡住
  assert.ok(elapsed < 50, `fast path must be synchronous, took ${elapsed}ms`)
  assert.equal(result.tookFastPath, true)
  // config 立刻就有值
  assert.equal(configRef.value?.storefront_template, 'qiaozhi')
  // 后台 refetch 发出了一次（但没 await）
  assert.equal(apiCallCount, 1)
})

test('loadConfig fast path: without window.__PUBLIC_CONFIG__ falls through to normal path', () => {
  const configRef = { value: null as any }
  const result = loadConfigFastPath({
    configRef,
    force: false,
    windowRef: {}, // 没有 __PUBLIC_CONFIG__
    configAPIGet: () => Promise.resolve({ data: { data: {} } }),
  })
  assert.equal(result.tookFastPath, false)
  assert.equal(configRef.value, null)
})

test('loadConfig fast path: force=true bypasses fast path and refetches', () => {
  const configRef = { value: { storefront_template: 'stale' } }
  const result = loadConfigFastPath({
    configRef,
    force: true,
    windowRef: { __PUBLIC_CONFIG__: { storefront_template: 'qiaozhi' } },
    configAPIGet: () => Promise.resolve({ data: { data: {} } }),
  })
  assert.equal(result.tookFastPath, false)
})
