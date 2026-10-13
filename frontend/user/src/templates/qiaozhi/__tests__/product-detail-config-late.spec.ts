// Verifies ProductDetail.vue behavior when onMounted no longer awaits
// appStore.loadConfig() — the `payment_channels` and `captcha` computed values
// must stay correctly reactive when config arrives late (迟到 1s 后再到).
//
// Project uses node:test (not vitest) + no DOM mount; this spec therefore
// reproduces the computed semantics using Vue's standalone reactivity engine
// so we can prove `channels` picks up the late-arriving payment_channels array.
// Picked up by pnpm run test via the shim at tests/productDetailConfigLate.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { computed, nextTick, ref } from 'vue'

type ConfigShape = {
  payment_channels?: Array<{ id: number; name?: string; provider_type?: string; channel_type?: string; fee_bps?: number }>
  captcha?: { provider?: string; scenes?: { guest_create_order?: boolean }; turnstile?: { site_key?: string } }
} | null

// 复刻 ProductDetail.vue 的 channels / captchaProvider computed，
// 这两个 computed 是从 appStore.config 读的，所以首帧 config=null 时
// 它们返回安全的空值；config 到达后应立刻切换。
function makeDetailComputed(appStoreConfig: { value: ConfigShape }) {
  const captchaConfig = computed(() => appStoreConfig.value?.captcha as any)
  const captchaProvider = computed(() => {
    const cfg = captchaConfig.value
    if (!cfg || String(cfg.provider || 'none') === 'none') return 'none'
    if (!cfg?.scenes?.guest_create_order) return 'none'
    return String(cfg.provider)
  })
  const channels = computed<Array<{ id: number; name: string }>>(() => {
    const list = appStoreConfig.value?.payment_channels
    if (!Array.isArray(list)) return []
    return list
      .filter((ch: any) => {
        const providerType = String(ch?.provider_type || '').toLowerCase()
        const channelType = String(ch?.channel_type || '').toLowerCase()
        if (providerType === 'epay') return ['wechat', 'wxpay', 'alipay', 'qqpay'].includes(channelType)
        return true
      })
      .map((ch: any) => ({
        id: Number(ch.id),
        name: String(ch.name || ch.channel_type || ch.provider_type || ''),
      }))
  })
  return { channels, captchaProvider }
}

test('ProductDetail payment_channels: computed is safe empty when config is still null (首帧)', () => {
  const configRef = ref<ConfigShape>(null)
  const { channels, captchaProvider } = makeDetailComputed(configRef)
  assert.deepEqual(channels.value, [])
  assert.equal(captchaProvider.value, 'none')
})

test('ProductDetail payment_channels: computed responds within one tick after config arrives late (~1s)', async () => {
  const configRef = ref<ConfigShape>(null)
  const { channels, captchaProvider } = makeDetailComputed(configRef)

  // 首帧：onMounted 已经触发（component mounted），但 config 还没回来
  assert.deepEqual(channels.value, [])
  assert.equal(captchaProvider.value, 'none')

  // 模拟 1s 后 /public/config 到达，后台把 config 塞给 store
  await new Promise((resolve) => setTimeout(resolve, 1000))
  configRef.value = {
    payment_channels: [
      { id: 1, name: 'Alipay', provider_type: 'epay', channel_type: 'alipay', fee_bps: 60 },
      { id: 2, name: 'Wechat', provider_type: 'epay', channel_type: 'wechat', fee_bps: 60 },
      { id: 3, name: 'Blocked', provider_type: 'epay', channel_type: 'other', fee_bps: 0 }, // epay 下非白名单会被过滤
      { id: 4, name: 'Stripe', provider_type: 'stripe', channel_type: 'card' },
    ],
    captcha: { provider: 'turnstile', scenes: { guest_create_order: true }, turnstile: { site_key: 'sk-xx' } },
  }
  await nextTick()

  // channels 应该立刻反映到最新 config
  assert.equal(channels.value.length, 3, `expected 3 channels after filter, got ${channels.value.length}`)
  assert.deepEqual(channels.value.map((c: { id: number; name: string }) => c.id).sort(), [1, 2, 4])
  // captchaProvider 从 'none' 切换到 'turnstile'
  assert.equal(captchaProvider.value, 'turnstile')
})

test('ProductDetail payment_channels: when payment_channels is missing from config, computed stays []', () => {
  const configRef = ref<ConfigShape>({ captcha: { provider: 'none' } })
  const { channels } = makeDetailComputed(configRef)
  assert.deepEqual(channels.value, [])
})
