<template>
  <div ref="rootEl" class="qz-container qz-tfa">
    <section class="qz-panel qz-tfa-card qz-reveal is-pending">
      <!-- 页头：eyebrow + h1 + 说明 + 返回（与购物车 / 查订单页同款胶囊） -->
      <header class="qz-page-head">
        <div class="qz-page-head-copy">
          <span class="qz-eyebrow">{{ t('qiaozhi.tfa.eyebrow') }}</span>
          <h1>{{ t('qiaozhi.tfa.title') }}</h1>
          <p>{{ t('qiaozhi.tfa.desc') }}</p>
        </div>
        <RouterLink to="/" class="qz-btn qz-btn--secondary qz-btn--sm">
          <ChevronLeft />
          {{ t('auth.login.backHome') }}
        </RouterLink>
      </header>

      <!-- 输入区 -->
      <div class="qz-field">
        <div class="qz-tfa-field-head">
          <label class="qz-field-label" for="tfa-secret">{{ t('qiaozhi.tfa.inputLabel') }}</label>
          <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm" @click="pasteFromClipboard">
            <ClipboardPaste />
            {{ t('qiaozhi.tfa.paste') }}
          </button>
        </div>
        <textarea
          id="tfa-secret"
          ref="secretEl"
          v-model="secretInput"
          class="qz-field-control qz-tfa-secret"
          rows="3"
          spellcheck="false"
          autocomplete="off"
          :placeholder="t('qiaozhi.tfa.placeholder')"
          :aria-invalid="secretInvalid || undefined"
          :aria-describedby="secretInvalid ? 'tfa-secret-error' : undefined"
        ></textarea>
      </div>

      <!-- 验证码展示 -->
      <div class="qz-tfa-display" :class="{ 'is-ready': parsed }">
        <p v-if="!secretInput.trim()" class="qz-tfa-status">{{ t('qiaozhi.tfa.idle') }}</p>
        <p v-else-if="!parsed" id="tfa-secret-error" class="qz-tfa-status is-error" role="alert">{{ t('qiaozhi.tfa.invalid') }}</p>

        <template v-else>
          <div class="qz-tfa-code" role="group" tabindex="0" :aria-label="t('qiaozhi.tfa.codeAria', { code })">
            <span v-for="(ch, i) in codeChars" :key="i" class="qz-tfa-digit" aria-hidden="true">{{ ch }}</span>
          </div>

          <div class="qz-tfa-ring" :aria-label="t('qiaozhi.tfa.remaining', { seconds: secondsLeft })" role="timer">
            <svg viewBox="0 0 36 36" aria-hidden="true">
              <circle class="qz-tfa-ring-bg" cx="18" cy="18" r="15" />
              <circle
                class="qz-tfa-ring-fg"
                cx="18"
                cy="18"
                r="15"
                :style="{ strokeDashoffset: ringOffset }"
              />
            </svg>
            <span class="qz-tfa-ring-num" aria-hidden="true">{{ secondsLeft }}</span>
          </div>
        </template>
      </div>

      <!-- 复制 -->
      <QzButton variant="primary" size="lg" class="qz-btn--block" :disabled="!parsed" @click="copyCode">
        <Check v-if="copied" />
        <Copy v-else />
        {{ copied ? t('qiaozhi.tfa.copied') : t('qiaozhi.tfa.copyFull') }}
      </QzButton>

      <p class="qz-tfa-hint">{{ t('qiaozhi.tfa.hint') }}</p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Check, ChevronLeft, ClipboardPaste, Copy } from 'lucide-vue-next'
import { setupReveal } from './utils/reveal'
import QzButton from './components/QzButton.vue'

const { t } = useI18n()

const secretInput = ref('')
const secretEl = ref<HTMLTextAreaElement | null>(null)
const code = ref('')
const secondsLeft = ref(30)
const copied = ref(false)

let timer: ReturnType<typeof setInterval> | null = null
let copyTimer: ReturnType<typeof setTimeout> | null = null
let currentWindow = -1

// ---------- Base32（RFC4648）解码 ----------
const BASE32_ALPHABET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567'

const decodeBase32 = (raw: string): Uint8Array | null => {
  const cleaned = raw.toUpperCase().replace(/[\s-]/g, '').replace(/=+$/, '')
  if (!cleaned || cleaned.length < 8 || /[^A-Z2-7]/.test(cleaned)) return null
  let bits = 0
  let value = 0
  const out: number[] = []
  for (const ch of cleaned) {
    const idx = BASE32_ALPHABET.indexOf(ch)
    if (idx === -1) return null
    value = (value << 5) | idx
    bits += 5
    if (bits >= 8) {
      out.push((value >>> (bits - 8)) & 0xff)
      bits -= 8
    }
  }
  return out.length ? new Uint8Array(out) : null
}

// ---------- 密钥解析（Base32 裸串 或 otpauth:// URI，含 digits/period/algorithm） ----------
interface TotpParams {
  secret: string
  digits: number
  period: number
  algorithm: 'SHA-1' | 'SHA-256' | 'SHA-512'
}

const parseSecret = (raw: string): TotpParams | null => {
  const text = (raw || '').trim()
  if (!text) return null

  if (/^otpauth:\/\//i.test(text)) {
    try {
      const url = new URL(text)
      const secret = (url.searchParams.get('secret') || '').replace(/[\s-]/g, '').toUpperCase()
      if (!decodeBase32(secret)) return null
      const rawDigits = Number(url.searchParams.get('digits') || 6)
      const rawPeriod = Number(url.searchParams.get('period') || 30)
      const rawAlgo = (url.searchParams.get('algorithm') || 'SHA1').toUpperCase()
      const algorithm: TotpParams['algorithm'] =
        rawAlgo === 'SHA256' ? 'SHA-256' : rawAlgo === 'SHA512' ? 'SHA-512' : 'SHA-1'
      return {
        secret,
        digits: rawDigits === 8 ? 8 : 6,
        period: Number.isFinite(rawPeriod) && rawPeriod > 0 ? Math.min(Math.round(rawPeriod), 300) : 30,
        algorithm,
      }
    } catch {
      return null
    }
  }

  const clean = text.replace(/[\s-]/g, '').toUpperCase()
  if (!/^[A-Z2-7]+=*$/.test(clean)) return null
  if (!decodeBase32(clean)) return null
  return { secret: clean.replace(/=+$/, ''), digits: 6, period: 30, algorithm: 'SHA-1' }
}

const parsed = computed(() => parseSecret(secretInput.value))
const secretInvalid = computed(() => Boolean(secretInput.value.trim()) && !parsed.value)

// ---------- TOTP（WebCrypto HMAC） ----------
const generateCode = async (params: TotpParams, windowIndex: number): Promise<string> => {
  const keyBytes = decodeBase32(params.secret)
  if (!keyBytes) return ''
  const msg = new Uint8Array(8)
  let counter = windowIndex
  for (let i = 7; i >= 0; i--) {
    msg[i] = counter & 0xff
    counter = Math.floor(counter / 256)
  }
  const key = await crypto.subtle.importKey(
    'raw',
    keyBytes as unknown as BufferSource,
    { name: 'HMAC', hash: params.algorithm },
    false,
    ['sign'],
  )
  const digest = new Uint8Array(await crypto.subtle.sign('HMAC', key, msg as unknown as BufferSource))
  const offset = (digest[digest.length - 1] ?? 0) & 0x0f
  const binary =
    (((digest[offset] ?? 0) & 0x7f) << 24) |
    (((digest[offset + 1] ?? 0) & 0xff) << 16) |
    (((digest[offset + 2] ?? 0) & 0xff) << 8) |
    ((digest[offset + 3] ?? 0) & 0xff)
  return String(binary % 10 ** params.digits).padStart(params.digits, '0')
}

const refresh = async (force = false) => {
  const params = parsed.value
  if (!params) {
    code.value = ''
    currentWindow = -1
    return
  }
  const now = Math.floor(Date.now() / 1000)
  const windowIndex = Math.floor(now / params.period)
  secondsLeft.value = params.period - (now % params.period)
  if (force || windowIndex !== currentWindow) {
    currentWindow = windowIndex
    code.value = await generateCode(params, windowIndex)
  }
}

// 输入即解析：粘贴/敲入立即出码，无需点按钮
watch(secretInput, () => void refresh(true))

// ---------- 视觉 ----------
const codeChars = computed(() => {
  const text = code.value || ''.padStart(6, '-')
  if (text.length === 8) return [text.slice(0, 4), text.slice(4)]
  return text.split('')
})

const RING_LENGTH = 94.2
// 倒计时圆环按剩余时间绘制：每个周期从满圈开始，随后逆时针消退（此前公式画反，呈涨满方向）。
const ringOffset = computed(() => {
  const period = parsed.value?.period || 30
  return RING_LENGTH * (1 - secondsLeft.value / period)
})

// ---------- 操作 ----------
const pasteFromClipboard = async () => {
  try {
    const text = await navigator.clipboard.readText()
    if (text && text.trim()) {
      secretInput.value = text.trim()
      return
    }
  } catch {
    /* 浏览器拒绝读取剪贴板时聚焦输入框由用户手动粘贴 */
  }
  secretEl.value?.focus()
}

const copyCode = async () => {
  if (!code.value) return
  try {
    await navigator.clipboard.writeText(code.value)
    copied.value = true
    if (copyTimer) clearTimeout(copyTimer)
    copyTimer = setTimeout(() => (copied.value = false), 1400)
  } catch {
    /* 忽略 */
  }
}

// 入场动画
const rootEl = ref<HTMLElement | null>(null)
let cleanupReveal: (() => void) | null = null
const armReveal = () => {
  cleanupReveal?.()
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
}

onMounted(() => {
  void refresh(true)
  timer = setInterval(() => void refresh(), 200)
  armReveal()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  if (copyTimer) clearTimeout(copyTimer)
  cleanupReveal?.()
})
</script>

<style scoped>
.qz-tfa {
  display: grid;
  place-items: start center;
}

/* .qz-panel.qz-tfa-card 写成 (0,3,0)：legacy.css 里旧分类侧栏的 .qz-panel 规则（sticky、≤760 隐藏、overflow hidden）
   在它被删掉之前不能影响这张卡。 */
.qz-panel.qz-tfa-card {
  position: static;
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-5);
  width: min(560px, 100%);
  padding: var(--qz-space-8) var(--qz-space-6);
  overflow: visible;
}

/* 页头（文案 + 返回胶囊）用 base.css 的 .qz-page-head，手机端胶囊挪到文案上方也由它负责 */

/* ---------- 输入区：label 与粘贴钮同行 ---------- */
.qz-tfa-field-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--qz-space-3);
}

.qz-field-control.qz-tfa-secret {
  min-height: 92px;
  font-family: var(--qz-font-mono);
  line-height: 1.6;
}

/* ---------- 验证码展示 ---------- */
.qz-tfa-display {
  display: grid;
  place-items: center;
  gap: var(--qz-space-4);
  min-height: 180px;
  padding: var(--qz-space-6) var(--qz-space-5);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  transition: border-color var(--qz-duration-standard) var(--qz-ease-standard);
}

.qz-tfa-display.is-ready {
  border-color: var(--qz-border-strong);
}

.qz-tfa-status {
  margin: 0;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-base);
  font-weight: 600;
  line-height: var(--qz-leading-body);
  text-align: center;
}

.qz-tfa-status.is-error {
  color: var(--qz-danger);
}

.qz-tfa-code {
  display: flex;
  justify-content: center;
  gap: var(--qz-space-2);
  border-radius: var(--qz-radius-control);
}

.qz-tfa-digit {
  color: var(--qz-text);
  font-family: var(--qz-font-mono);
  font-size: var(--qz-text-3xl);
  font-weight: 800;
  line-height: 1;
  letter-spacing: var(--qz-tracking-display);
  font-variant-numeric: tabular-nums;
  animation: qz-tfa-code-in var(--qz-duration-considered) var(--qz-ease-enter);
}

@keyframes qz-tfa-code-in {
  from {
    opacity: 0;
    transform: translateY(8px) scale(0.98);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* 倒计时圆环 */
.qz-tfa-ring {
  position: relative;
  width: 44px;
  height: 44px;
}

.qz-tfa-ring svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.qz-tfa-ring-bg {
  fill: none;
  stroke: var(--qz-border-strong);
  stroke-width: 3;
}

.qz-tfa-ring-fg {
  fill: none;
  stroke: var(--qz-accent);
  stroke-width: 3;
  stroke-linecap: round;
  stroke-dasharray: 94.2;
  transition: stroke-dashoffset var(--qz-duration-standard) linear;
}

.qz-tfa-ring-num {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  color: var(--qz-text);
  font-size: var(--qz-text-xs);
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.qz-tfa-hint {
  margin: 0;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  text-align: center;
}

@media (max-width: 760px) {
  .qz-panel.qz-tfa-card {
    gap: var(--qz-space-4);
    padding: var(--qz-space-6) var(--qz-space-5);
  }
}

@media (prefers-reduced-motion: reduce) {
  .qz-tfa-digit {
    animation: none;
  }
}
</style>
