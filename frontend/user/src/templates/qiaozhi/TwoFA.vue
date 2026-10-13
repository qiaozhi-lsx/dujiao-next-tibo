<template>
  <div ref="rootEl" class="j8-container">
    <div class="tfa-wrap">
      <div class="tfa-card j8-reveal is-pending">
        <span class="j8-eyebrow">TWO-FACTOR AUTHENTICATOR</span>
        <h1>{{ t('qiaozhi.tfa.title') }}</h1>
        <p class="tfa-desc">{{ t('qiaozhi.tfa.desc') }}</p>

        <!-- 输入区 -->
        <div class="tfa-field-head">
          <label for="tfa-secret">{{ t('qiaozhi.tfa.inputLabel') }}</label>
          <button type="button" class="tfa-paste" @click="pasteFromClipboard">
            <ClipboardPaste />
            {{ t('qiaozhi.tfa.paste') }}
          </button>
        </div>
        <textarea
          id="tfa-secret"
          ref="secretEl"
          v-model="secretInput"
          rows="3"
          spellcheck="false"
          autocomplete="off"
          :placeholder="t('qiaozhi.tfa.placeholder')"
        ></textarea>

        <!-- 验证码展示 -->
        <div class="tfa-display" :class="{ 'is-ready': parsed }">
          <p v-if="!secretInput.trim()" class="tfa-status is-idle">{{ t('qiaozhi.tfa.idle') }}</p>
          <p v-else-if="!parsed" class="tfa-status is-error" role="alert">{{ t('qiaozhi.tfa.invalid') }}</p>

          <template v-else>
            <div class="tfa-code" role="group" tabindex="0" :aria-label="t('qiaozhi.tfa.codeAria', { code })">
              <span v-for="(ch, i) in codeChars" :key="i" class="tfa-digit" aria-hidden="true">{{ ch }}</span>
            </div>

            <div class="tfa-ring-row">
              <div class="tfa-ring">
                <svg viewBox="0 0 36 36" aria-hidden="true">
                  <circle class="ring-bg" cx="18" cy="18" r="15" />
                  <circle
                    class="ring-fg"
                    cx="18"
                    cy="18"
                    r="15"
                    :style="{ strokeDashoffset: ringOffset }"
                  />
                </svg>
                <span class="tfa-ring-num">{{ secondsLeft }}</span>
              </div>
            </div>
          </template>
        </div>

        <!-- 复制 -->
        <button type="button" class="tfa-copy-btn" :disabled="!parsed" @click="copyCode">
          <Check v-if="copied" />
          <Copy v-else />
          {{ copied ? t('qiaozhi.tfa.copied') : t('qiaozhi.tfa.copyFull') }}
        </button>

        <p class="tfa-hint">{{ t('qiaozhi.tfa.hint') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Check, ClipboardPaste, Copy } from 'lucide-vue-next'
import { setupReveal } from './utils/reveal'

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
.tfa-wrap {
  display: grid;
  place-items: start center;
  padding: 40px 0 60px;
}

.tfa-card {
  width: min(560px, 100%);
  padding: 30px 28px 26px;
  border: 1px solid var(--line);
  border-radius: var(--r-card);
  background: var(--surface);
  box-shadow: var(--shadow-soft);
}

.tfa-card h1 {
  margin: 6px 0 0;
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.tfa-desc {
  margin: 10px 0 20px;
  color: var(--muted);
  font-size: 13px;
  line-height: 1.7;
}

.tfa-field-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 8px;
}

.tfa-field-head label {
  font-size: 13px;
  font-weight: 700;
  color: var(--text);
}

.tfa-paste {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 12px;
  border: 1px solid var(--accent);
  border-radius: 9px;
  background: var(--accent);
  color: var(--on-accent);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  transition: background 0.15s, transform 0.15s;
}

.tfa-paste:hover {
  background: var(--accent-strong);
  transform: translateY(-1px);
}

.tfa-paste :deep(svg) {
  width: 14px;
  height: 14px;
}

.tfa-card textarea {
  width: 100%;
  min-height: 92px;
  padding: 14px 16px;
  border: 1.5px solid var(--line-strong);
  border-radius: 16px;
  background: var(--surface);
  color: var(--text);
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.6;
  resize: vertical;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.tfa-card textarea:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 4px var(--accent-soft);
}

.tfa-card textarea::placeholder {
  color: var(--muted);
}

.tfa-display {
  margin-top: 18px;
  padding: 26px 20px 22px;
  border: 1.5px solid var(--line);
  border-radius: 20px;
  background: linear-gradient(160deg, var(--surface), var(--surface-2));
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 18px;
  min-height: 190px;
  transition: border-color 0.3s;
}

.tfa-display.is-ready {
  border-color: var(--accent-line);
}

.tfa-status {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.tfa-status.is-idle {
  color: var(--accent);
  animation: tfa-pulse 2s ease-in-out infinite;
}

.tfa-status.is-error {
  color: var(--red);
}

@keyframes tfa-pulse {
  0%,
  100% {
    opacity: 0.6;
  }

  50% {
    opacity: 1;
  }
}

.tfa-code {
  display: flex;
  justify-content: center;
  gap: 10px;
  outline: none;
}

.tfa-code:focus-visible {
  outline: 3px solid var(--accent-soft);
  outline-offset: 6px;
  border-radius: 10px;
}

.tfa-digit {
  font-family: var(--font-mono);
  font-size: 46px;
  font-weight: 800;
  line-height: 1;
  letter-spacing: -0.02em;
  font-variant-numeric: tabular-nums;
  color: var(--accent);
  filter: drop-shadow(0 2px 8px rgba(10, 10, 10, 0.12));
  animation: tfa-code-in 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes tfa-code-in {
  from {
    opacity: 0;
    transform: translateY(8px) scale(0.98);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.tfa-ring-row {
  display: flex;
  align-items: center;
  justify-content: center;
}

.tfa-ring {
  position: relative;
  width: 42px;
  height: 42px;
}

.tfa-ring svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.ring-bg {
  fill: none;
  stroke: var(--surface-3);
  stroke-width: 3;
}

.ring-fg {
  fill: none;
  stroke: var(--accent);
  stroke-width: 3;
  stroke-linecap: round;
  stroke-dasharray: 94.2;
  transition: stroke-dashoffset 0.22s linear;
}

.tfa-ring-num {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  font-size: 12px;
  font-weight: 800;
  color: var(--text);
  font-variant-numeric: tabular-nums;
}

.tfa-copy-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  min-height: 50px;
  margin-top: 16px;
  border: 0;
  border-radius: 14px;
  background: var(--accent);
  color: var(--on-accent);
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  box-shadow: var(--shadow-primary);
  transition: transform 0.15s, box-shadow 0.15s, opacity 0.15s;
}

.tfa-copy-btn :deep(svg) {
  width: 18px;
  height: 18px;
}

.tfa-copy-btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.tfa-copy-btn:active:not(:disabled) {
  transform: translateY(0) scale(0.985);
}

.tfa-copy-btn:disabled {
  background: var(--surface-3);
  color: var(--muted);
  box-shadow: none;
  cursor: not-allowed;
}

.tfa-hint {
  margin: 14px 0 0;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.7;
}
</style>
