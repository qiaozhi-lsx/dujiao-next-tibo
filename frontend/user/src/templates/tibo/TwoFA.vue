<template>
  <div class="j8-container">
    <div class="tfa-wrap">
      <div class="tfa-card">
        <span class="j8-eyebrow">TWO-FACTOR CODE</span>
        <h1>{{ t('tibo.tfa.title') }}</h1>
        <p class="tfa-desc">{{ t('tibo.tfa.desc') }}</p>

        <label class="tfa-field">
          <input
            v-model="secretInput"
            type="text"
            spellcheck="false"
            autocomplete="off"
            :placeholder="t('tibo.tfa.placeholder')"
            :aria-label="t('tibo.tfa.title')"
            @keyup.enter="applySecret"
          />
          <button type="button" class="j8-btn-primary" :disabled="!normalizedSecret" @click="applySecret">
            {{ t('tibo.tfa.generate') }}
          </button>
        </label>

        <p v-if="error" class="tfa-error">{{ error }}</p>

        <template v-if="activeSecret">
          <div class="tfa-code-box">
            <div class="tfa-code" role="group" tabindex="0" :aria-label="code ? t('tibo.tfa.codeAria', { code }) : ''">
              <span v-for="(ch, i) in codeChars" :key="i" class="tfa-digit" aria-hidden="true">{{ ch }}</span>
            </div>
            <div class="tfa-progress">
              <i :style="{ width: `${progress * 100}%` }" :class="{ 'is-warn': secondsLeft <= 5 }"></i>
            </div>
            <div class="tfa-meta">
              <span>{{ t('tibo.tfa.remaining', { seconds: secondsLeft }) }}</span>
              <button type="button" class="tfa-copy" @click="copyCode">
                <Check v-if="copied" />
                <Copy v-else />
                {{ copied ? t('tibo.tfa.copied') : t('tibo.tfa.copy') }}
              </button>
            </div>
          </div>
          <p class="tfa-hint">{{ t('tibo.tfa.hint') }}</p>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Check, Copy } from 'lucide-vue-next'

const { t } = useI18n()

const secretInput = ref('')
const activeSecret = ref('')
const error = ref('')
const code = ref('')
const secondsLeft = ref(30)
const copied = ref(false)

let timer: ReturnType<typeof setInterval> | null = null
let copyTimer: ReturnType<typeof setTimeout> | null = null

// ---------- Base32（RFC4648）解码 ----------
const BASE32_ALPHABET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567'

const decodeBase32 = (raw: string): Uint8Array | null => {
  const cleaned = raw.toUpperCase().replace(/[\s-]/g, '').replace(/=+$/, '')
  if (!cleaned || /[^A-Z2-7]/.test(cleaned)) return null
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

const normalizedSecret = computed(() => {
  const raw = secretInput.value.trim()
  if (!raw) return ''
  // 支持 otpauth:// 链接
  const match = raw.match(/[?&]secret=([A-Za-z2-7\s=+-]+)/)
  const candidate = match?.[1] ?? raw
  return decodeBase32(candidate) ? candidate : ''
})

// ---------- TOTP（HMAC-SHA1） ----------
const generateCode = async (): Promise<string> => {
  const keyBytes = decodeBase32(activeSecret.value)
  if (!keyBytes) return ''
  const counter = Math.floor(Date.now() / 1000 / 30)
  const msg = new Uint8Array(8)
  let counterValue = counter
  for (let i = 7; i >= 0; i--) {
    msg[i] = counterValue & 0xff
    counterValue = Math.floor(counterValue / 256)
  }
  const key = await crypto.subtle.importKey('raw', keyBytes as unknown as BufferSource, { name: 'HMAC', hash: 'SHA-1' }, false, ['sign'])
  const digest = new Uint8Array(await crypto.subtle.sign('HMAC', key, msg as unknown as BufferSource))
  const offset = (digest[digest.length - 1] ?? 0) & 0x0f
  const binary =
    (((digest[offset] ?? 0) & 0x7f) << 24) |
    (((digest[offset + 1] ?? 0) & 0xff) << 16) |
    (((digest[offset + 2] ?? 0) & 0xff) << 8) |
    ((digest[offset + 3] ?? 0) & 0xff)
  return String(binary % 1_000_000).padStart(6, '0')
}

const tick = async () => {
  secondsLeft.value = 30 - (Math.floor(Date.now() / 1000) % 30)
  code.value = await generateCode()
}

const progress = computed(() => secondsLeft.value / 30)
const codeChars = computed(() => (code.value ? code.value.split('') : '------'.split('')))

const applySecret = async () => {
  error.value = ''
  const raw = secretInput.value.trim()
  if (!raw) return
  if (!normalizedSecret.value) {
    error.value = t('tibo.tfa.invalid')
    return
  }
  const match = raw.match(/[?&]secret=([A-Za-z2-7\s=+-]+)/)
  activeSecret.value = (match?.[1] ?? raw).toUpperCase().replace(/[\s-]/g, '').replace(/=+$/, '')
  await tick()
  if (timer) clearInterval(timer)
  timer = setInterval(() => void tick(), 1000)
}

const copyCode = async () => {
  if (!code.value) return
  try {
    await navigator.clipboard.writeText(code.value)
    copied.value = true
    if (copyTimer) clearTimeout(copyTimer)
    copyTimer = setTimeout(() => (copied.value = false), 1600)
  } catch {
    /* 忽略 */
  }
}

onMounted(() => {
  void tick()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  if (copyTimer) clearTimeout(copyTimer)
})
</script>

<style scoped>
.tfa-wrap {
  display: grid;
  place-items: start center;
  padding: 40px 0 60px;
}

.tfa-card {
  width: min(520px, 100%);
  padding: 28px 26px 26px;
  border: 1px solid var(--line);
  border-radius: var(--r-card);
  background: var(--surface);
  box-shadow: var(--shadow-soft);
}

.tfa-card h1 {
  margin: 6px 0 0;
  font-size: 24px;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.tfa-desc {
  margin: 10px 0 18px;
  color: var(--muted);
  font-size: 13px;
  line-height: 1.7;
}

.tfa-field {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 8px;
}

.tfa-field input {
  min-height: 44px;
  padding: 10px 14px;
  border: 1px solid var(--line-strong);
  border-radius: 12px;
  background: var(--surface);
  color: var(--text);
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.tfa-field input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 4px var(--accent-soft);
}

.tfa-error {
  margin: 10px 0 0;
  color: var(--red);
  font-size: 12px;
}

.tfa-code-box {
  margin-top: 22px;
}

.tfa-code {
  display: flex;
  justify-content: center;
  gap: 10px;
}

.tfa-digit {
  display: grid;
  place-items: center;
  width: 54px;
  height: 72px;
  border: 1px solid var(--line);
  border-radius: 14px;
  background: var(--surface-2);
  color: var(--accent-strong);
  font-size: 34px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.tfa-progress {
  height: 6px;
  margin-top: 16px;
  border-radius: 999px;
  background: var(--surface-3);
  overflow: hidden;
}

.tfa-progress i {
  display: block;
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, var(--accent), #67dfd4);
  transition: width 0.9s linear, background 0.2s;
}

.tfa-progress i.is-warn {
  background: linear-gradient(90deg, var(--red), #f0a06a);
}

.tfa-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
  color: var(--muted);
  font-size: 12px;
}

.tfa-copy {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 12px;
  border: 1px solid var(--accent-line);
  border-radius: 999px;
  background: var(--accent-soft);
  color: var(--accent-strong);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.tfa-copy :deep(svg) {
  width: 14px;
  height: 14px;
}

.tfa-hint {
  margin: 16px 0 0;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.7;
}
</style>
