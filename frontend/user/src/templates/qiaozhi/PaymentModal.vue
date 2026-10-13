<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="pay-modal-mask"
      role="dialog"
      aria-modal="true"
      :aria-label="t('qiaozhi.payment.modal.title')"
      @click.self="handleClose"
    >
      <div
        ref="modalRef"
        class="pay-modal"
        :class="stageClass"
        tabindex="-1"
        @keydown="handleKeydown"
      >
        <button
          type="button"
          class="pay-modal__close"
          :aria-label="t('qiaozhi.payment.modal.closeAria')"
          @click="handleClose"
        >
          <X />
        </button>

        <span class="pay-modal__tag">{{ t('qiaozhi.payment.modal.tag') }}</span>
        <h2 class="pay-modal__title">{{ t('qiaozhi.payment.modal.title') }}</h2>

        <div class="pay-modal__sn">
          <span>{{ order?.sn || '—' }}</span>
          <span v-if="showCountdown" class="pay-modal__countdown" :class="{ 'is-danger': countdownExpired }">
            <Clock />
            <b>{{ countdownText }}</b>
          </span>
        </div>

        <div class="pay-modal__time">
          <span>{{ t('qiaozhi.payment.modal.created') }} <b>{{ createdAtText }}</b></span>
          <span>{{ t('qiaozhi.payment.modal.expiresAt') }} <b>{{ expiresAtText }}</b></span>
        </div>

        <div class="pay-modal__sum">
          <div v-for="(line, idx) in summaryLines" :key="idx" class="pay-modal__sum-row">
            <span>{{ line.label }}</span>
            <span>{{ line.value }}</span>
          </div>
          <div class="pay-modal__sum-row pay-modal__sum-row--total">
            <span>{{ t('qiaozhi.payment.modal.total') }}</span>
            <span>{{ totalText }}</span>
          </div>
        </div>

        <div v-if="qrSupported" class="pay-modal__qr-wrap">
          <div class="pay-modal__qr">
            <img v-if="qrImageUrl" :src="qrImageUrl" alt="QR" />
            <div v-else class="pay-modal__qr-placeholder">
              <Loader2 class="pay-modal__spin" />
            </div>
          </div>
        </div>

        <div v-else-if="!orderPastExpire" class="pay-modal__redirect-hint">
          <ExternalLink />
          <span>{{ t('qiaozhi.payment.modal.notQrCapable') }}</span>
        </div>

        <div v-if="cryptoDetails.length" class="pay-modal__crypto">
          <div v-for="item in cryptoDetails" :key="item.key" class="pay-modal__crypto-row">
            <small>{{ item.label }}</small>
            <b>
              {{ item.value }}
              <em v-if="item.detail">({{ item.detail }})</em>
            </b>
          </div>
          <div v-if="order?.walletAddress" class="pay-modal__crypto-action">
            <button type="button" class="pay-modal__copy-btn" @click="handleCopyWalletAddress">
              <Copy />
              {{ t('payment.copyWalletAddress') }}
            </button>
            <span v-if="walletAddressCopied" class="pay-modal__copied">{{ t('payment.copied') }}</span>
          </div>
        </div>

        <div class="pay-modal__steps">
          <span :class="{ on: stepIdx >= 0 }">{{ t('qiaozhi.payment.modal.stepWaiting') }}</span>
          <span :class="{ on: stepIdx >= 1 }">{{ t('qiaozhi.payment.modal.stepConfirming') }}</span>
          <span :class="{ on: stepIdx >= 2 }">{{ t('qiaozhi.payment.modal.stepPaid') }}</span>
        </div>

        <div class="pay-modal__stage">
          <div class="pay-modal__go" :class="goClass">
            <span v-if="stage === 'waiting'">
              {{ orderPastExpire ? t('qiaozhi.payment.modal.orderExpired') : t('qiaozhi.payment.modal.waitingScan') }}
              <template v-if="!orderPastExpire">
                <em class="pay-modal__dot"></em>
                <em class="pay-modal__dot"></em>
                <em class="pay-modal__dot"></em>
              </template>
            </span>
            <span v-else-if="stage === 'running'">{{ t('qiaozhi.payment.modal.confirmingPay') }}</span>
            <span v-else>{{ t('qiaozhi.payment.modal.paymentSuccess') }}</span>
          </div>
          <svg class="pay-modal__ring" viewBox="0 0 80 80" aria-hidden="true">
            <circle class="t" cx="40" cy="40" r="36" />
            <circle class="p" cx="40" cy="40" r="36" pathLength="100" :style="{ '--p': ringProgress }" />
          </svg>
          <svg class="pay-modal__ck" viewBox="0 0 24 24" aria-hidden="true">
            <path d="M5 12.5l4.5 4.5L19 7.5" />
          </svg>
        </div>

        <div v-if="stage === 'waiting' && !orderPastExpire && order?.h5PayUrl" class="pay-modal__link-row">
          <button type="button" class="pay-modal__link-btn" @click="handleOpenPayLink">
            <ExternalLink />
            {{ t('qiaozhi.payment.modal.openPayLink') }}
          </button>
          <button type="button" class="pay-modal__copy-btn" @click="handleCopyPayLink">
            <Copy />
            {{ t('qiaozhi.payment.modal.copyPayLink') }}
          </button>
          <span v-if="copied" class="pay-modal__copied">{{ t('qiaozhi.payment.modal.copied') }}</span>
        </div>

        <div class="pay-modal__rec" :aria-hidden="stage !== 'success'">
          <div class="pay-modal__rec-header">{{ t('qiaozhi.payment.modal.receiptHeader') }}</div>
          <div v-for="(it, idx) in receiptItems" :key="idx" class="pay-modal__rec-item">
            <div class="pay-modal__rec-icon">{{ it.initial }}</div>
            <div class="pay-modal__rec-meta">
              <div class="pay-modal__rec-name">{{ it.name }}</div>
              <div class="pay-modal__rec-sub">{{ t('qiaozhi.payment.modal.autoDelivery') }}</div>
            </div>
            <div class="pay-modal__rec-amt">{{ it.amount }}</div>
          </div>
          <i class="pay-modal__rec-check"><svg viewBox="0 0 30 30"><path d="M8 15.5l5 5L23 10" /></svg></i>
          <b class="pay-modal__rec-title">{{ t('qiaozhi.payment.modal.paymentSuccess') }}</b>
          <p class="pay-modal__rec-sub">
            <template v-if="order?.maskedEmail">
              <span>{{ emailNotifyBefore }}</span>
              <span class="pay-modal__rec-email">{{ order.maskedEmail }}</span>
              <span>{{ emailNotifyAfter }}</span>
              <br />
              <span>{{ t('qiaozhi.payment.modal.emailFallback') }}</span>
            </template>
            <template v-else>
              {{ t('qiaozhi.payment.modal.emailNotifyNoEmail') }}
            </template>
          </p>
          <dl class="pay-modal__rec-dl">
            <dt>{{ t('qiaozhi.payment.modal.orderNoLabel') }}</dt>
            <dd>{{ shortSn }}</dd>
            <dt>{{ t('qiaozhi.payment.modal.amountLabel') }}</dt>
            <dd>{{ totalText }}</dd>
            <dt>{{ t('qiaozhi.payment.modal.timeLabel') }}</dt>
            <dd>{{ receiptTimeText }}</dd>
          </dl>
          <button type="button" class="pay-modal__rec-btn" @click="handleViewDetail">
            {{ t('qiaozhi.payment.modal.viewOrderDetail') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Clock, Copy, ExternalLink, Loader2, X } from 'lucide-vue-next'
import QRCode from 'qrcode'
import { userOrderAPI, guestOrderAPI } from '../../api'

export interface PaymentModalItem {
  title: string
  amount: string
  quantity?: number
}

export interface PaymentModalOrder {
  sn: string
  amount: string
  currency?: string
  items: PaymentModalItem[]
  qrCodeContent?: string
  payUrl?: string
  h5PayUrl?: string
  expireAt?: number | string | null
  createdAt?: number | string | null
  maskedEmail?: string
  interactionMode?: 'qr' | 'redirect' | string
  feeAmount?: string
  walletAddress?: string
  chainAmount?: string
  chain?: string
  tokenId?: string
  guestCredentials?: { email: string; order_password: string } | null
}

const props = defineProps<{
  open: boolean
  order: PaymentModalOrder | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'paid', order: PaymentModalOrder): void
}>()

const { t } = useI18n()

type Stage = 'waiting' | 'running' | 'success'
const stage = ref<Stage>('waiting')
const modalRef = ref<HTMLElement | null>(null)
const qrImageUrl = ref('')
const copied = ref(false)
const copiedTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const walletAddressCopied = ref(false)
const walletCopiedTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const now = ref(Date.now())
const nowTimer = ref<ReturnType<typeof setInterval> | null>(null)
const pollTimer = ref<ReturnType<typeof setInterval> | null>(null)
const animationFrame = ref<number | null>(null)
const prefersReducedMotion = typeof window !== 'undefined' && window.matchMedia
  ? window.matchMedia('(prefers-reduced-motion: reduce)').matches
  : false

const ringProgress = ref(0)
const previouslyFocused = ref<HTMLElement | null>(null)

const stageClass = computed(() => ({
  'is-waiting': stage.value === 'waiting',
  'is-run': stage.value === 'running',
  'is-ok': stage.value === 'success',
  'is-rec': stage.value === 'success',
}))

const goClass = computed(() => ({
  waiting: stage.value === 'waiting',
}))

const stepIdx = computed(() => {
  if (stage.value === 'success') return 2
  if (stage.value === 'running') return 1
  return 0
})

const qrSupported = computed(() => {
  const o = props.order
  if (!o) return false
  const mode = (o.interactionMode || '').toLowerCase()
  if (mode === 'redirect') return false
  return Boolean((o.qrCodeContent || '').trim() || (o.payUrl || '').trim() || (o.h5PayUrl || '').trim())
})

const expireMs = computed(() => {
  const raw = props.order?.expireAt
  if (raw === null || raw === undefined || raw === '') return null
  const ts = typeof raw === 'number' ? raw : new Date(raw).getTime()
  return Number.isNaN(ts) ? null : ts
})

const orderPastExpire = computed(() => {
  if (expireMs.value === null) return false
  return expireMs.value <= now.value && stage.value === 'waiting'
})

const showCountdown = computed(() => expireMs.value !== null && stage.value === 'waiting')
const countdownExpired = computed(() => expireMs.value !== null && expireMs.value - now.value <= 0)
const countdownText = computed(() => {
  if (expireMs.value === null) return '--:--'
  const remain = Math.max(0, Math.floor((expireMs.value - now.value) / 1000))
  const m = String(Math.floor(remain / 60)).padStart(2, '0')
  const s = String(remain % 60).padStart(2, '0')
  return `${m}:${s}`
})

const formatCurrency = (amount?: string | number | null, currency?: string) => {
  if (amount === null || amount === undefined || amount === '') return '-'
  const text = typeof amount === 'number' ? String(amount) : amount
  if (!currency) return text
  return `${text} ${currency}`
}

const totalText = computed(() => formatCurrency(props.order?.amount, props.order?.currency))

const summaryLines = computed(() => {
  const o = props.order
  if (!o || !Array.isArray(o.items)) return [] as Array<{ label: string; value: string }>
  const lines = o.items.map((it) => {
    const label = it.quantity && it.quantity > 1 ? `${it.title} × ${it.quantity}` : it.title
    return { label, value: formatCurrency(it.amount, o.currency) }
  })
  if (o.feeAmount && Number(o.feeAmount) > 0) {
    lines.push({ label: t('qiaozhi.payment.modal.fee'), value: formatCurrency(o.feeAmount, o.currency) })
  }
  return lines
})

// 加密货币支付：币种 / 链 / 链上应付数量 / 收款地址。金额与地址仅在 api 返回 crypto 字段时展示。
const formatCryptoChain = (value: string) => {
  const normalized = value.trim().toLowerCase()
  const labels: Record<string, string> = {
    tron: 'TRON',
    trc20: 'TRON',
    base: 'Base',
    ethereum: 'Ethereum',
    eth: 'Ethereum',
    bsc: 'BNB Smart Chain',
    binance: 'BNB Smart Chain',
    polygon: 'Polygon',
    solana: 'Solana',
  }
  return labels[normalized] || value
}

const cryptoTokenLabel = computed(() => {
  const tokenId = (props.order?.tokenId || '').trim()
  if (!tokenId) return ''
  const parts = tokenId.split('-').filter(Boolean)
  return String(parts[parts.length - 1] || tokenId).toUpperCase()
})
const cryptoTokenDetail = computed(() => {
  const tokenId = (props.order?.tokenId || '').trim()
  if (!tokenId) return ''
  return tokenId.toUpperCase() === cryptoTokenLabel.value ? '' : tokenId
})
const cryptoDetails = computed(() => {
  const o = props.order
  const details: Array<{ key: string; label: string; value: string; detail?: string }> = []
  if (!o) return details
  if (cryptoTokenLabel.value) {
    details.push({ key: 'token', label: t('payment.cryptoToken'), value: cryptoTokenLabel.value, detail: cryptoTokenDetail.value })
  }
  if (o.chain) {
    details.push({ key: 'chain', label: t('payment.cryptoChain'), value: formatCryptoChain(o.chain) })
  }
  if (o.chainAmount) {
    details.push({ key: 'amount', label: t('payment.cryptoAmount'), value: o.chainAmount })
  }
  if (o.walletAddress) {
    details.push({ key: 'wallet_address', label: t('payment.walletAddress'), value: o.walletAddress })
  }
  return details
})

const receiptItems = computed(() => {
  const o = props.order
  if (!o || !Array.isArray(o.items) || !o.items.length) return []
  return o.items.map((it) => ({
    initial: (it.title || '?').trim().slice(0, 1).toUpperCase(),
    name: it.title,
    amount: formatCurrency(it.amount, o.currency),
  }))
})

const formatDateText = (raw?: number | string | null) => {
  if (raw === null || raw === undefined || raw === '') return '—'
  const ts = typeof raw === 'number' ? raw : new Date(raw).getTime()
  if (Number.isNaN(ts)) return '—'
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const createdAtText = computed(() => formatDateText(props.order?.createdAt))
const expiresAtText = computed(() => formatDateText(props.order?.expireAt))
const receiptTimeText = computed(() => formatDateText(now.value))

const shortSn = computed(() => {
  const sn = props.order?.sn || ''
  if (sn.length <= 12) return sn
  return `${sn.slice(0, 4)}...${sn.slice(-6)}`
})

// emailNotify 文案含 {email} 占位,拆成前/后两段,中间插可点击 span。
const emailNotifyBefore = computed(() => {
  const raw = t('qiaozhi.payment.modal.emailNotify', { email: '\u0000EMAIL\u0000' })
  const [before] = raw.split('\u0000EMAIL\u0000')
  return before || ''
})
const emailNotifyAfter = computed(() => {
  const raw = t('qiaozhi.payment.modal.emailNotify', { email: '\u0000EMAIL\u0000' })
  const parts = raw.split('\u0000EMAIL\u0000')
  return parts.length > 1 ? parts[1] : ''
})

const renderQRCode = async () => {
  const o = props.order
  if (!qrSupported.value || !o) {
    qrImageUrl.value = ''
    return
  }
  const payload = (o.qrCodeContent || '').trim() || (o.payUrl || '').trim() || (o.h5PayUrl || '').trim()
  if (!payload) {
    qrImageUrl.value = ''
    return
  }
  if (payload.startsWith('data:image/')) {
    qrImageUrl.value = payload
    return
  }
  if (/^https?:\/\/.+\.(png|jpe?g|gif|webp|svg)(\?.*)?$/i.test(payload)) {
    qrImageUrl.value = payload
    return
  }
  try {
    qrImageUrl.value = await QRCode.toDataURL(payload, {
      width: 220,
      margin: 1,
      errorCorrectionLevel: 'M',
    })
  } catch {
    qrImageUrl.value = ''
  }
}

const stopNowTimer = () => {
  if (nowTimer.value !== null) {
    clearInterval(nowTimer.value)
    nowTimer.value = null
  }
}

const startNowTimer = () => {
  stopNowTimer()
  now.value = Date.now()
  nowTimer.value = setInterval(() => {
    now.value = Date.now()
  }, 1000)
}

const stopPolling = () => {
  if (pollTimer.value !== null) {
    clearInterval(pollTimer.value)
    pollTimer.value = null
  }
}

const cancelAnimation = () => {
  if (animationFrame.value !== null) {
    cancelAnimationFrame(animationFrame.value)
    animationFrame.value = null
  }
}

const runSuccessAnimation = () => {
  cancelAnimation()
  const duration = prefersReducedMotion ? 300 : 2600
  const delay = prefersReducedMotion ? 0 : 350
  const start = performance.now()
  const tick = (ts: number) => {
    const elapsed = Math.max(0, ts - start - delay)
    const k = Math.max(0, Math.min(1, elapsed / duration))
    const eased = 1 - Math.pow(1 - k, 2.2)
    ringProgress.value = Number(eased.toFixed(4))
    if (k < 1) {
      animationFrame.value = requestAnimationFrame(tick)
      return
    }
    animationFrame.value = null
    stage.value = 'success'
    const order = props.order
    const emitPaidDelay = prefersReducedMotion ? 400 : 1500
    setTimeout(() => {
      if (order) emit('paid', order)
    }, emitPaidDelay)
  }
  animationFrame.value = requestAnimationFrame(tick)
}

const triggerSuccessFromPolling = () => {
  stopPolling()
  if (stage.value !== 'waiting') return
  stage.value = 'running'
  runSuccessAnimation()
}

const PAID_STATUSES = new Set(['paid', 'fulfilling', 'partially_delivered', 'delivered', 'completed'])

const pollOnce = async () => {
  const o = props.order
  if (!o || !o.sn) return
  try {
    let resp
    if (o.guestCredentials && o.guestCredentials.email) {
      resp = await guestOrderAPI.detail(o.sn, {
        email: o.guestCredentials.email,
        order_password: o.guestCredentials.order_password,
      }, { silentBusinessError: true })
    } else {
      resp = await userOrderAPI.detail(o.sn, { silentBusinessError: true })
    }
    const data = resp?.data?.data
    const status = String(data?.status || '')
    if (PAID_STATUSES.has(status)) {
      triggerSuccessFromPolling()
    }
  } catch {
    // 轮询失败静默忽略；下一次 2s 继续
  }
}

const startPolling = () => {
  stopPolling()
  pollTimer.value = setInterval(() => {
    void pollOnce()
  }, 2000)
}

const trapFocus = (e: KeyboardEvent) => {
  if (e.key !== 'Tab') return
  const root = modalRef.value
  if (!root) return
  const focusable = Array.from(root.querySelectorAll<HTMLElement>(
    'button:not([disabled]), [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
  ))
  if (!focusable.length) return
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (!first || !last) return
  const active = document.activeElement as HTMLElement | null
  if (e.shiftKey && active === first) {
    e.preventDefault()
    last.focus()
  } else if (!e.shiftKey && active === last) {
    e.preventDefault()
    first.focus()
  }
}

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    e.preventDefault()
    handleClose()
    return
  }
  trapFocus(e)
}

const handleClose = () => {
  emit('update:open', false)
}

const handleOpenPayLink = () => {
  const o = props.order
  if (!o) return
  const link = (o.h5PayUrl || o.payUrl || '').trim()
  if (!link) return
  const isMobile = /Android|iPhone|iPad|iPod|Mobile/i.test(navigator.userAgent)
  if (isMobile) {
    window.location.href = link
  } else {
    const w = window.open('', '_blank')
    if (w) {
      w.opener = null
      w.location.replace(link)
    } else {
      window.location.href = link
    }
  }
}

const handleCopyPayLink = async () => {
  const o = props.order
  if (!o) return
  const link = (o.h5PayUrl || o.payUrl || '').trim()
  if (!link) return
  try {
    await navigator.clipboard.writeText(link)
    copied.value = true
    if (copiedTimer.value !== null) clearTimeout(copiedTimer.value)
    copiedTimer.value = setTimeout(() => { copied.value = false }, 1500)
  } catch {
    // 复制失败静默
  }
}

const handleCopyWalletAddress = async () => {
  const address = (props.order?.walletAddress || '').trim()
  if (!address) return
  try {
    await navigator.clipboard.writeText(address)
    walletAddressCopied.value = true
    if (walletCopiedTimer.value !== null) clearTimeout(walletCopiedTimer.value)
    walletCopiedTimer.value = setTimeout(() => { walletAddressCopied.value = false }, 1500)
  } catch {
    // 复制失败静默
  }
}

const handleViewDetail = () => {
  const order = props.order
  if (!order) return
  emit('paid', order)
}

const resetState = () => {
  stage.value = 'waiting'
  ringProgress.value = 0
  qrImageUrl.value = ''
  copied.value = false
  walletAddressCopied.value = false
}

watch(() => props.open, async (opened) => {
  if (opened) {
    resetState()
    startNowTimer()
    await renderQRCode()
    if (props.order?.sn) startPolling()
    previouslyFocused.value = (document.activeElement as HTMLElement) || null
    await nextTick()
    modalRef.value?.focus()
    document.body.style.overflow = 'hidden'
  } else {
    stopNowTimer()
    stopPolling()
    cancelAnimation()
    if (copiedTimer.value !== null) {
      clearTimeout(copiedTimer.value)
      copiedTimer.value = null
    }
    if (walletCopiedTimer.value !== null) {
      clearTimeout(walletCopiedTimer.value)
      walletCopiedTimer.value = null
    }
    document.body.style.overflow = ''
    previouslyFocused.value?.focus?.()
  }
}, { immediate: true })

watch(() => props.order, async () => {
  if (!props.open) return
  await renderQRCode()
})

onBeforeUnmount(() => {
  stopNowTimer()
  stopPolling()
  cancelAnimation()
  if (copiedTimer.value !== null) clearTimeout(copiedTimer.value)
  if (walletCopiedTimer.value !== null) clearTimeout(walletCopiedTimer.value)
  document.body.style.overflow = ''
})
</script>

<style scoped>
.pay-modal-mask {
  position: fixed; inset: 0;
  background: rgba(10, 10, 10, 0.5);
  backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center;
  z-index: 1200;
  animation: pay-modal-fade .25s ease;
}

@keyframes pay-modal-fade { from { opacity: 0; } }

.pay-modal {
  --ux-accent: var(--accent, #0a0a0a);
  --ux-mint: #167d3b;
  --ux-ink-2: #57535f;
  --ux-ink-3: #8e8a97;
  --ux-surface: var(--surface, #fff);
  --ux-surface-2: #f5f5f5;
  --ux-surface-3: #ebebeb;
  --ux-line: rgba(10, 10, 10, 0.08);
  --ux-line-2: rgba(10, 10, 10, 0.16);
  --ux-warn: #b06a00;
  --ux-ease: cubic-bezier(.2, .8, .2, 1);
  --ux-spring: cubic-bezier(.34, 1.45, .64, 1);

  position: relative;
  width: min(380px, calc(100vw - 32px));
  max-height: calc(100vh - 32px);
  overflow: hidden;
  background: var(--ux-surface);
  border-radius: 24px;
  box-shadow: 0 30px 60px -36px rgba(40, 30, 80, 0.3), 0 0 0 1px var(--ux-line);
  display: flex; flex-direction: column;
  padding: 22px 20px 20px;
  color: var(--text, #0a0a0a);
  font: 14px/1.5 -apple-system, "Segoe UI", "PingFang SC", sans-serif;
  animation: pay-modal-rise .3s var(--ux-spring);
  outline: none;
}

@keyframes pay-modal-rise { from { transform: translateY(20px) scale(.98); opacity: 0; } }

.pay-modal__close {
  position: absolute; top: 14px; right: 14px;
  width: 30px; height: 30px;
  border: 0; background: transparent; border-radius: 8px;
  cursor: pointer; color: var(--ux-ink-3);
  display: grid; place-items: center;
  z-index: 3;
}
.pay-modal__close:hover { background: #f3f3f3; color: #000; }
.pay-modal__close svg { width: 15px; height: 15px; }

.pay-modal__tag {
  display: block;
  font-size: 10px; letter-spacing: 1.2px;
  color: var(--ux-ink-3); font-weight: 600;
  margin-bottom: 4px;
  transition: opacity .4s, transform .5s var(--ux-ease);
}
.pay-modal__title {
  font-size: 17px; font-weight: 800;
  margin: 0 0 14px;
  transition: opacity .4s, transform .5s var(--ux-ease);
}

.pay-modal__sn {
  display: flex; justify-content: space-between; align-items: center;
  gap: 10px;
  font-size: 11px; color: var(--ux-ink-3);
  margin-bottom: 10px;
  font-variant-numeric: tabular-nums;
  transition: opacity .4s;
}
.pay-modal__sn > span:first-child {
  min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.pay-modal__countdown {
  flex-shrink: 0;
  padding: 3px 9px; background: #fff7e6; color: var(--ux-warn);
  border-radius: 999px; font-size: 11px; font-weight: 600;
  display: inline-flex; align-items: center; gap: 4px;
  font-variant-numeric: tabular-nums;
}
.pay-modal__countdown.is-danger { background: #fde8e8; color: #b02020; }
.pay-modal__countdown svg { width: 11px; height: 11px; }

.pay-modal__time {
  display: flex; justify-content: space-between; gap: 10px;
  font-size: 10.5px; color: var(--ux-ink-3);
  margin-bottom: 12px;
  font-variant-numeric: tabular-nums;
  transition: opacity .4s;
}
.pay-modal__time b { color: var(--ux-ink-2); font-weight: 600; }

.pay-modal__sum {
  padding: 12px 14px; border-radius: 14px;
  background: var(--ux-surface-2);
  margin-bottom: 12px;
  display: flex; flex-direction: column; gap: 4px;
  transition: opacity .4s, transform .5s var(--ux-ease);
}
.pay-modal__sum-row {
  display: flex; justify-content: space-between; gap: 10px;
  font-size: 12.5px; color: var(--ux-ink-2); font-weight: 600;
}
.pay-modal__sum-row span:first-child {
  min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.pay-modal__sum-row--total {
  padding-top: 8px; margin-top: 4px;
  border-top: 1px dashed var(--ux-line-2);
  font-size: 15px; color: var(--text, #0a0a0a); font-weight: 800;
}

.pay-modal__crypto {
  padding: 12px 14px; border-radius: 14px;
  background: var(--ux-surface-2);
  margin-bottom: 12px;
  display: flex; flex-direction: column; gap: 8px;
  transition: opacity .4s, transform .5s var(--ux-ease);
}
.pay-modal__crypto-row {
  display: flex; align-items: flex-start; justify-content: space-between; gap: 12px;
}
.pay-modal__crypto-row small {
  flex: 0 0 auto; font-size: 11.5px; color: var(--ux-ink-3); font-weight: 600;
  padding-top: 1px;
}
.pay-modal__crypto-row b {
  min-width: 0; text-align: right; font-size: 13px; font-weight: 800;
  color: var(--text, #0a0a0a); overflow-wrap: anywhere; word-break: break-all;
}
.pay-modal__crypto-row em {
  font-style: normal; font-weight: 500; color: var(--ux-ink-3); margin-left: 4px;
}
.pay-modal__crypto-action {
  display: inline-flex; align-items: center; gap: 10px; flex-wrap: wrap;
}

.pay-modal__qr-wrap {
  display: flex; justify-content: center;
  margin: 4px 0 12px;
  transition: filter .4s, opacity .4s;
}
.pay-modal__qr {
  padding: 10px; background: #fff;
  border: 1px solid var(--ux-line-2);
  border-radius: 12px;
}
.pay-modal__qr img { display: block; width: 150px; height: 150px; }
.pay-modal__qr-placeholder {
  width: 150px; height: 150px;
  display: grid; place-items: center;
  color: var(--ux-ink-3);
}
.pay-modal__spin { width: 24px; height: 24px; animation: pay-modal-spin 1s linear infinite; }
@keyframes pay-modal-spin { to { transform: rotate(360deg); } }

.is-run .pay-modal__qr-wrap,
.is-ok .pay-modal__qr-wrap { filter: grayscale(1); opacity: .3; }

.pay-modal__redirect-hint {
  display: flex; align-items: flex-start; gap: 8px;
  padding: 10px 12px; margin-bottom: 12px;
  border-radius: 10px;
  background: #fff7e6; color: var(--ux-warn);
  font-size: 12px; line-height: 1.5;
}
.pay-modal__redirect-hint svg { width: 14px; height: 14px; margin-top: 2px; flex-shrink: 0; }

.pay-modal__steps {
  display: flex; gap: 14px; justify-content: center;
  height: 0; margin: 0; overflow: hidden;
  font-size: 11.5px; font-weight: 700; color: var(--ux-ink-3);
  opacity: 0;
  transition: opacity .3s, height .3s, margin .3s;
}
.is-run .pay-modal__steps,
.is-ok .pay-modal__steps { opacity: 1; height: 20px; margin: 4px 0 14px; }
.pay-modal__steps span {
  display: inline-flex; align-items: center; gap: 5px;
  transition: color .3s;
}
.pay-modal__steps span::before {
  content: "";
  width: 7px; height: 7px; border-radius: 50%;
  background: var(--ux-line-2);
  transition: background-color .3s, transform .4s var(--ux-spring);
}
.pay-modal__steps span.on { color: var(--text, #0a0a0a); }
.pay-modal__steps span.on::before { background: var(--ux-accent); transform: scale(1.3); }
.is-ok .pay-modal__steps span.on::before { background: var(--ux-mint); }

.pay-modal__stage {
  position: relative;
  display: flex; align-items: center; justify-content: center;
  min-height: 56px; margin-top: 4px;
  transition: min-height .4s var(--ux-ease);
}
.is-run .pay-modal__stage,
.is-ok .pay-modal__stage { min-height: 90px; }

.pay-modal__go {
  position: relative;
  display: flex; align-items: center; justify-content: center;
  width: 100%; height: 52px;
  border: 0; border-radius: 14px;
  background: var(--ux-accent); color: #fff;
  font-size: 14.5px; font-weight: 800;
  overflow: hidden;
  transition: width .5s var(--ux-spring), height .5s var(--ux-spring),
              border-radius .5s var(--ux-spring), background-color .4s;
}
.pay-modal__go.waiting {
  background: var(--ux-surface-2); color: var(--ux-ink-3);
  box-shadow: inset 0 0 0 1px var(--ux-line-2);
}
.pay-modal__go.waiting .pay-modal__dot {
  display: inline-block; width: 6px; height: 6px; border-radius: 50%;
  background: var(--ux-ink-3); margin-left: 4px;
  animation: pay-modal-dot 1.2s infinite ease-in-out;
}
.pay-modal__go.waiting .pay-modal__dot:nth-of-type(2) { animation-delay: .2s; }
.pay-modal__go.waiting .pay-modal__dot:nth-of-type(3) { animation-delay: .4s; }
@keyframes pay-modal-dot {
  0%, 80%, 100% { opacity: .3; transform: scale(.8); }
  40% { opacity: 1; transform: scale(1.1); }
}
.is-run .pay-modal__go,
.is-ok .pay-modal__go {
  width: 56px; height: 56px; border-radius: 50%;
  pointer-events: none;
  background: var(--ux-accent); color: #fff;
  box-shadow: none;
}
.is-run .pay-modal__go span,
.is-ok .pay-modal__go span { opacity: 0; }
.is-ok .pay-modal__go { background: var(--ux-mint); }

.pay-modal__ring {
  pointer-events: none;
  position: absolute;
  width: 80px; height: 80px;
  top: 50%; left: 50%;
  margin: -40px 0 0 -40px;
  transform: rotate(-90deg);
  opacity: 0;
  transition: opacity .3s;
}
.is-run .pay-modal__ring,
.is-ok .pay-modal__ring { opacity: 1; transition-delay: .25s; }
.pay-modal__ring circle { fill: none; stroke-width: 5; stroke-linecap: round; }
.pay-modal__ring .t { stroke: var(--ux-surface-3); }
.pay-modal__ring .p {
  stroke: var(--ux-accent);
  stroke-dasharray: 100;
  stroke-dashoffset: calc(100 - var(--p, 0) * 100);
  transition: stroke .4s;
}
.is-ok .pay-modal__ring .p { stroke: var(--ux-mint); }

.pay-modal__ck {
  position: absolute;
  top: 50%; left: 50%;
  margin: -12px 0 0 -12px;
  width: 24px; height: 24px;
  fill: none; stroke: #fff;
  stroke-width: 3.4; stroke-linecap: round; stroke-linejoin: round;
  pointer-events: none;
}
.pay-modal__ck path {
  stroke-dasharray: 24;
  stroke-dashoffset: 24;
  transition: stroke-dashoffset .45s var(--ux-ease) .2s;
}
.is-ok .pay-modal__ck path { stroke-dashoffset: 0; }

.pay-modal__link-row {
  display: flex; align-items: center; gap: 8px;
  margin-top: 10px; flex-wrap: wrap;
}
.pay-modal__link-btn,
.pay-modal__copy-btn {
  display: inline-flex; align-items: center; gap: 5px;
  height: 36px; padding: 0 12px;
  border: 0; border-radius: 10px;
  cursor: pointer;
  font-size: 12.5px; font-weight: 700;
}
.pay-modal__link-btn {
  background: var(--ux-accent); color: #fff;
}
.pay-modal__copy-btn {
  background: var(--ux-surface-2); color: var(--ux-ink-2);
  box-shadow: inset 0 0 0 1px var(--ux-line);
}
.pay-modal__link-btn svg, .pay-modal__copy-btn svg { width: 13px; height: 13px; }
.pay-modal__copied {
  font-size: 11px; color: var(--ux-mint); font-weight: 600;
}

.pay-modal__rec {
  position: absolute;
  left: 16px; right: 16px; bottom: 16px; top: 60px;
  display: flex; flex-direction: column;
  padding: 22px 18px 16px;
  border-radius: 22px;
  background: var(--ux-surface-2);
  transform: translateY(110%);
  transition: transform .6s var(--ux-spring);
  text-align: center;
  z-index: 2;
  overflow-y: auto;
}
.pay-modal__rec::before {
  content: "";
  position: absolute; left: 0; right: 0; top: -8px;
  height: 16px;
  background: radial-gradient(circle 8px at 8px 0, transparent 98%, var(--ux-surface-2)) 0 8px/16px 8px repeat-x;
}
.is-rec .pay-modal__rec { transform: none; }

.pay-modal__rec-header {
  font-size: 10px; letter-spacing: 1.3px;
  color: var(--ux-ink-3); font-weight: 600;
  padding-bottom: 12px;
  border-bottom: 1px dashed var(--ux-line-2);
  margin-bottom: 16px;
}
.pay-modal__rec-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 12px;
  background: var(--ux-surface);
  border-radius: 12px;
  margin-bottom: 14px;
  text-align: left;
  box-shadow: inset 0 0 0 1px var(--ux-line);
}
.pay-modal__rec-icon {
  width: 36px; height: 36px; flex-shrink: 0;
  border-radius: 8px;
  background: linear-gradient(135deg, #0f0f0f, #2a2a2a);
  display: grid; place-items: center;
  color: #fff; font-size: 14px; font-weight: 800;
}
.pay-modal__rec-meta { flex: 1; min-width: 0; text-align: left; }
.pay-modal__rec-name {
  font-size: 13px; font-weight: 700;
  color: var(--text, #0a0a0a); margin-bottom: 2px;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.pay-modal__rec-sub {
  font-size: 11px; color: var(--ux-ink-3);
  display: flex; align-items: center; gap: 5px;
}
.pay-modal__rec-sub::before {
  content: "●"; color: var(--ux-mint); font-size: 7px;
}
.pay-modal__rec-amt {
  font-size: 14px; font-weight: 800;
  color: var(--text, #0a0a0a);
  font-variant-numeric: tabular-nums;
}

.pay-modal__rec-check {
  width: 52px; height: 52px;
  margin: 0 auto 10px;
  border-radius: 50%;
  display: grid; place-items: center;
  background: var(--ux-mint); color: #fff;
  box-shadow: 0 0 0 7px color-mix(in srgb, var(--ux-mint) 16%, transparent);
}
.pay-modal__rec-check svg {
  width: 24px; height: 24px;
  fill: none; stroke: currentColor;
  stroke-width: 3; stroke-linecap: round; stroke-linejoin: round;
}

.pay-modal__rec-title {
  display: block;
  font-size: 17px; font-weight: 850;
  letter-spacing: -.03em;
}

.pay-modal__rec > p {
  margin: 4px 0 12px;
  font-size: 12px; color: var(--ux-ink-3); line-height: 1.55;
}
.pay-modal__rec-email { color: var(--ux-ink-2); font-weight: 600; }

.pay-modal__rec-dl {
  display: grid; grid-template-columns: auto 1fr;
  gap: 6px 12px;
  margin: 0 0 16px;
  text-align: left;
  font-size: 12px;
}
.pay-modal__rec-dl dt { color: var(--ux-ink-3); font-weight: 650; }
.pay-modal__rec-dl dd {
  margin: 0; text-align: right;
  font-weight: 750;
  font-variant-numeric: tabular-nums;
}

.pay-modal__rec-btn {
  width: 100%; height: 44px;
  border: 0; border-radius: 12px;
  background: var(--ux-accent); color: #fff;
  font-size: 13.5px; font-weight: 750;
  cursor: pointer;
}

.is-rec .pay-modal__tag,
.is-rec .pay-modal__title,
.is-rec .pay-modal__sn,
.is-rec .pay-modal__time,
.is-rec .pay-modal__sum,
.is-rec .pay-modal__qr-wrap,
.is-rec .pay-modal__redirect-hint,
.is-rec .pay-modal__crypto,
.is-rec .pay-modal__steps,
.is-rec .pay-modal__stage,
.is-rec .pay-modal__link-row {
  opacity: 0; transform: translateY(-16px);
  pointer-events: none;
  transition: opacity .4s, transform .5s var(--ux-ease);
}

@media (prefers-reduced-motion: reduce) {
  .pay-modal, .pay-modal *, .pay-modal-mask, .pay-modal-mask * {
    animation-duration: .001ms !important;
    transition-duration: .001ms !important;
  }
}
</style>

