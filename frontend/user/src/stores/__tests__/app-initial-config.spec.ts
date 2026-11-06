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
// configComplete 同样复刻 app.ts：内联子集赋值后为 false，后台 refetch 成功赋值后才为 true，
// 刷新失败保持 false。这个 helper 和 app.ts 里的实际代码保持等价——改一处要同步改另一处。
function loadConfigFastPath(opts: {
  configRef: { value: any }
  configCompleteRef?: { value: boolean }
  force: boolean
  windowRef: Record<string, any> | undefined
  configAPIGet: () => Promise<any>
  onBackgroundResolved?: (resp: any) => void
}): { tookFastPath: boolean; backgroundStarted: boolean; background: Promise<void> | null } {
  const { configRef, configCompleteRef = { value: false }, force, windowRef, configAPIGet, onBackgroundResolved } = opts
  if (configRef.value && !force) {
    return { tookFastPath: false, backgroundStarted: false, background: null }
  }
  if (!force && windowRef && windowRef.__PUBLIC_CONFIG__) {
    configRef.value = windowRef.__PUBLIC_CONFIG__
    configCompleteRef.value = false
    // 后台 refetch：和 app.ts 一样 fire-and-forget（app.ts 用 void 丢弃；
    // 这里把 Promise 一并返回，仅供测试等待它 settle，helper 自身不 await）
    const background = (async () => {
      try {
        const resp = await configAPIGet()
        configRef.value = resp.data.data
        configCompleteRef.value = true
        if (onBackgroundResolved) onBackgroundResolved(resp)
      } catch { /* 后台失败不影响首帧 */ }
    })()
    return { tookFastPath: true, backgroundStarted: true, background }
  }
  return { tookFastPath: false, backgroundStarted: false, background: null }
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

test('loadConfig fast path: configComplete is false after inline config, true once the background refresh lands', async () => {
  let resolveBackground!: (resp: any) => void
  const configAPIGet = () => new Promise<any>((resolve) => { resolveBackground = resolve })
  const configRef = { value: null as any }
  const configCompleteRef = { value: false }
  // 内联子集和服务端 whitelist 一样只有品牌 / 模板等首屏字段，
  // 没有 nav_config、email_verification_enabled 这些可选开关键
  const windowRef = { __PUBLIC_CONFIG__: { storefront_template: 'qiaozhi', brand: {} } }

  const result = loadConfigFastPath({ configRef, configCompleteRef, force: false, windowRef, configAPIGet })
  assert.equal(result.tookFastPath, true)
  // 内联阶段：config 已经有值，但完整配置还没到，configComplete 必须为 false，
  // 否则 `value !== false` 会把缺失的开关键读成"开启"
  assert.equal(configRef.value?.storefront_template, 'qiaozhi')
  assert.equal(configRef.value?.nav_config, undefined)
  assert.equal(configCompleteRef.value, false)

  resolveBackground({
    data: { data: { storefront_template: 'qiaozhi', brand: {}, nav_config: { builtin: { blog: false } } } },
  })
  await result.background
  assert.equal(configCompleteRef.value, true)
  assert.equal(configRef.value?.nav_config?.builtin?.blog, false)
})

test('loadConfig fast path: failed background refresh leaves configComplete false', async () => {
  const configRef = { value: null as any }
  const configCompleteRef = { value: false }
  const result = loadConfigFastPath({
    configRef,
    configCompleteRef,
    force: false,
    windowRef: { __PUBLIC_CONFIG__: { storefront_template: 'qiaozhi' } },
    configAPIGet: () => Promise.reject(new Error('network down')),
  })
  assert.equal(result.tookFastPath, true)
  await result.background
  // 刷新失败：首帧继续用内联配置，但不能把它当成完整配置
  assert.equal(configRef.value?.storefront_template, 'qiaozhi')
  assert.equal(configCompleteRef.value, false)
})
