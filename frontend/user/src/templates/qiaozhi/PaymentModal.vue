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
        <!-- 金额卡 -->
        <div class="pay-modal__amount-card">
          <button
            type="button"
            class="pay-modal__amount-copy"
            :aria-label="t('qiaozhi.payment.modal.copyAmount')"
            @click="handleCopyAmount"
          >
            <CheckCircle2 v-if="amountCopied" />
            <Copy v-else />
          </button>
          <div class="pay-modal__amount-top">
            <small class="pay-modal__amount-label">
              {{ t('qiaozhi.payment.modal.amountPaid') }}
              <em class="pay-modal__fee-note">({{ t('qiaozhi.payment.modal.feeIncluded') }})</em>
            </small>
            <div class="pay-modal__amount-value">
              <b>{{ amountBigText }}</b>
              <span v-if="paymentChipLabel" class="pay-modal__chain-chip">
                <img v-if="chainLogo" class="pay-modal__chain-logo" :src="chainLogo" alt="" aria-hidden="true" />
                <i v-else class="pay-modal__chain-dot" :class="chainDotClass" aria-hidden="true"></i>
                {{ paymentChipLabel }}
              </span>
            </div>
          </div>
          <div class="pay-modal__amount-meta">
            <div class="pay-modal__meta-row">
              <span>{{ t('qiaozhi.payment.modal.orderAmount') }}</span>
              <b>{{ totalText }}</b>
            </div>
            <div v-if="feeText" class="pay-modal__meta-row">
              <span>{{ t('qiaozhi.payment.modal.fee') }}</span>
              <b>+{{ feeText }}</b>
            </div>
            <div v-if="equivalentText" class="pay-modal__meta-row">
              <span>{{ t('qiaozhi.payment.modal.equivalent') }}</span>
              <b>{{ equivalentText }}</b>
            </div>
            <div class="pay-modal__meta-row">
              <span>{{ t('qiaozhi.payment.modal.orderNoLabel') }}</span>
              <b>{{ order?.sn || '—' }}</b>
            </div>
          </div>
        </div>

        <!-- 支付卡：二维码 / 倒计时 / 收款地址 -->
        <div class="pay-modal__pay-card">
          <div class="pay-modal__pay-head">
            <b>{{ payCardTitle }}</b>
          </div>

          <div v-if="showCountdown" class="pay-modal__cd-num" :class="`is-${cdTone}`">
            {{ countdownText }}
          </div>

          <div v-if="qrSupported && !orderPastExpire" class="pay-modal__qr-wrap">
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

          <div v-if="order?.walletAddress && !orderPastExpire" class="pay-modal__addr">
            <div class="pay-modal__addr-label">{{ t('payment.walletAddress') }}</div>
            <div class="pay-modal__addr-body">
              <b>{{ order.walletAddress }}</b>
              <button
                type="button"
                class="pay-modal__addr-copy"
                :aria-label="t('payment.copyWalletAddress')"
                @click="handleCopyWalletAddress"
              >
                <CheckCircle2 v-if="walletAddressCopied" />
                <Copy v-else />
              </button>
            </div>
            <span v-if="walletAddressCopied" class="pay-modal__copied">{{ t('payment.copied') }}</span>
          </div>

          <div v-if="orderPastExpire" class="pay-modal__expired">
            <CircleX class="pay-modal__expired-icon" aria-hidden="true" />
            <b class="pay-modal__expired-title">{{ t('qiaozhi.payment.modal.payExpired') }}</b>
            <button type="button" class="pay-modal__expired-btn" @click="handleRenew">
              {{ t('qiaozhi.payment.modal.renewPayment') }}
            </button>
          </div>
        </div>

        <!-- 加密货币操作：切换网络或币种 / 已转账但未到账 -->
        <div
          v-if="hasCryptoPayment && stage === 'waiting' && !orderPastExpire"
          class="pay-modal__crypto-actions"
        >
          <div class="pay-modal__action-row">
            <button
              v-if="switchableChannels.length"
              type="button"
              class="pay-modal__action-btn"
              :disabled="switchBusy"
              @click="toggleSwitchPanel"
            >
              <ArrowLeftRight />
              {{ t('qiaozhi.payment.modal.switchNetwork') }}
            </button>

            <button type="button" class="pay-modal__action-btn is-solid" @click="toggleTxPanel">
              <CheckCircle2 />
              {{ t('qiaozhi.payment.modal.txHashBtn') }}
            </button>
          </div>

          <div v-if="switchPanelOpen" class="pay-modal__panel">
            <div class="pay-modal__chip-list">
              <button
                v-for="ch in switchableChannels"
                :key="ch.id"
                type="button"
                class="pay-modal__chip"
                :class="{ 'is-active': ch.id === order?.channelId }"
                :disabled="switchBusy || ch.id === order?.channelId"
                @click="handleSwitchChannel(ch.id)"
              >
                {{ ch.name }}
              </button>
            </div>
            <p v-if="switchBusy" class="pay-modal__panel-msg">{{ t('qiaozhi.payment.modal.switching') }}</p>
            <p v-else-if="switchError" class="pay-modal__panel-msg is-err">{{ switchError }}</p>
          </div>

          <div v-if="txPanelOpen" class="pay-modal__panel">
            <input
              v-model="txHashInput"
              class="pay-modal__input"
              type="text"
              autocomplete="off"
              spellcheck="false"
              :placeholder="t('qiaozhi.payment.modal.txHashPlaceholder')"
              :disabled="txBusy"
              @keydown.enter.prevent="handleSubmitTx"
            />
            <button
              type="button"
              class="pay-modal__action-btn is-solid"
              :disabled="txBusy || !txHashInput.trim()"
              @click="handleSubmitTx"
            >
              {{ txBusy ? t('qiaozhi.payment.modal.txHashSubmitting') : t('qiaozhi.payment.modal.txHashSubmit') }}
            </button>
            <p v-if="txFeedback" class="pay-modal__panel-msg" :class="{ 'is-err': !txFeedback.ok }">
              {{ txFeedback.ok ? t('qiaozhi.payment.modal.txHashOk') : (txFeedback.message || t('qiaozhi.payment.modal.txHashFailed')) }}
            </p>
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
              <Loader2 v-if="!orderPastExpire" class="pay-modal__spin" aria-hidden="true" />
              {{ orderPastExpire ? t('qiaozhi.payment.modal.orderExpired') : t('qiaozhi.payment.modal.waitingScan') }}
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
import { ArrowLeftRight, CheckCircle2, CircleX, Copy, ExternalLink, Loader2 } from 'lucide-vue-next'
import QRCode from 'qrcode'
import { userOrderAPI, guestOrderAPI } from '../../api'
import { chainLabel } from '../../utils/cryptoChain'
import { countdownTone } from '../../utils/countdownTone'
import bscLogo from '../../assets/chains/bsc.png'
import tronLogo from '../../assets/chains/tron.png'
import solanaLogo from '../../assets/chains/solana.png'

export interface PaymentModalItem {
  title: string
  amount: string
  quantity?: number
}

export interface PaymentModalChannel {
  id: number
  name: string
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
  payableAmount?: string
  walletAddress?: string
  chainAmount?: string
  chain?: string
  tokenId?: string
  channelId?: number
  channelName?: string
  providerType?: string
  paymentId?: number
  guestCredentials?: { email: string; order_password: string } | null
}

const props = defineProps<{
  open: boolean
  order: PaymentModalOrder | null
  channels?: PaymentModalChannel[]
  switchBusy?: boolean
  switchError?: string
  txBusy?: boolean
  txFeedback?: { ok: boolean; message?: string } | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'paid', order: PaymentModalOrder): void
  (e: 'switch-channel', channelId: number): void
  (e: 'submit-tx-hash', txHash: string): void
  (e: 'renew-payment'): void
}>()

const { t } = useI18n()

type Stage = 'waiting' | 'running' | 'success'
const stage = ref<Stage>('waiting')
const modalRef = ref<HTMLElement | null>(null)
const qrImageUrl = ref('')
const walletAddressCopied = ref(false)
const walletCopiedTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const switchPanelOpen = ref(false)
const txPanelOpen = ref(false)
const txHashInput = ref('')
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
  return Boolean(
    (o.walletAddress || '').trim()
    || (o.qrCodeContent || '').trim()
    || (o.payUrl || '').trim()
    || (o.h5PayUrl || '').trim(),
  )
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

const showCountdown = computed(() => expireMs.value !== null && stage.value === 'waiting' && !orderPastExpire.value)
const remainSeconds = computed(() => (
  expireMs.value === null ? Number.POSITIVE_INFINITY : Math.floor((expireMs.value - now.value) / 1000)
))
const cdTone = computed(() => countdownTone(remainSeconds.value))
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

// 加密货币支付：币种 / 链 / 链上应付数量 / 收款地址。
const cryptoTokenLabel = computed(() => {
  const tokenId = (props.order?.tokenId || '').trim()
  if (!tokenId) return ''
  const parts = tokenId.split('-').filter(Boolean)
  return String(parts[parts.length - 1] || tokenId).toUpperCase()
})

const hasCryptoPayment = computed(() => Boolean(props.order?.walletAddress))

const cryptoChainLabel = computed(() => (props.order?.chain ? chainLabel(props.order.chain) : ''))

// 金额下方的支付方式 chip：加密单显示链名；支付宝渠道统一品牌名 Alipay；其余显示渠道名。
const paymentChipLabel = computed(() => {
  if (cryptoChainLabel.value) return cryptoChainLabel.value
  const name = (props.order?.channelName || '').trim()
  if (!name) return ''
  return /支付宝|alipay/i.test(name) ? 'Alipay' : name
})

// 有素材的链用真 logo（网关同款），其余链回退色点。
const CHAIN_LOGOS: Record<string, string> = {
  bsc: bscLogo,
  binance: bscLogo,
  tron: tronLogo,
  trc20: tronLogo,
  solana: solanaLogo,
}
const chainLogo = computed(() => CHAIN_LOGOS[(props.order?.chain || '').trim().toLowerCase()] ?? '')

// 渠道 chip 的圆点配色：链/渠道 → 品牌色。
const chainDotClass = computed(() => {
  const key = `${props.order?.chain || ''} ${props.order?.providerType || ''} ${props.order?.channelName || ''}`.toLowerCase()
  if (/tron|trc20/.test(key)) return 'is-tron'
  if (/binance|bsc/.test(key)) return 'is-bsc'
  if (/solana|sol/.test(key)) return 'is-sol'
  if (/epay|alipay|支付宝/.test(key)) return 'is-alipay'
  return 'is-generic'
})

const amountBigText = computed(() => {
  const o = props.order
  if (!o) return '-'
  if (o.chainAmount) return `${o.chainAmount} ${cryptoTokenLabel.value || 'USDT'}`
  // 实付金额 = 含手续费金额；后端未下发时回退订单金额。
  const payable = (o.payableAmount || '').trim()
  return payable ? formatCurrency(payable, o.currency) : formatCurrency(o.amount, o.currency)
})

const feeText = computed(() => {
  const o = props.order
  if (!o) return ''
  const raw = (o.feeAmount || '').trim()
  if (!raw || raw === '0' || raw === '0.00') return ''
  return formatCurrency(raw, o.currency)
})

// 加密货币单：链上数量与法币实付的换算关系。网关按 amount(法币) + 实时汇率报出链上数量，
// 两者是一笔的两种币种口径；补这一行，页面上两个币种才能对上账。
const equivalentText = computed(() => {
  const o = props.order
  if (!o || !o.chainAmount) return ''
  const payable = (o.payableAmount || '').trim()
  if (!payable) return ''
  return formatCurrency(payable, o.currency)
})

const payCardTitle = computed(() => (
  hasCryptoPayment.value
    ? t('qiaozhi.payment.modal.scanOrCopy')
    : t('qiaozhi.payment.modal.scanTitle')
))

const switchableChannels = computed(() => {
  const list = props.channels || []
  return list.filter((ch) => Number.isFinite(ch.id) && ch.id > 0)
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

let qrRenderVersion = 0
const renderQRCode = async () => {
  const version = ++qrRenderVersion
  const o = props.order
  if (!qrSupported.value || !o) {
    qrImageUrl.value = ''
    return
  }
  const payload = (o.walletAddress || '').trim()
    || (o.qrCodeContent || '').trim()
    || (o.payUrl || '').trim()
    || (o.h5PayUrl || '').trim()
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
    const url = await QRCode.toDataURL(payload, {
      width: 220,
      margin: 1,
      errorCorrectionLevel: 'M',
    })
    if (version !== qrRenderVersion) return
    qrImageUrl.value = url
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

// 过期瞬间可能刚完成付款、2s 轮询尚未翻转：重发起前先补查一次状态，
// 已支付则直接进入成功流程，避免旧单已付、又新建一单的重复付款窗口。
const handleRenew = async () => {
  await pollOnce()
  if (stage.value !== 'waiting') return
  emit('renew-payment')
}

const trapFocus = (e: KeyboardEvent) => {
  if (e.key !== 'Tab') return
  const root = modalRef.value
  if (!root) return
  const focusable = Array.from(root.querySelectorAll<HTMLElement>(
    'button:not([disabled]), [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
  )).filter((el) => el.getClientRects().length > 0)
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

const amountCopied = ref(false)
const amountCopiedTimer = ref<ReturnType<typeof setTimeout> | null>(null)

const handleCopyAmount = async () => {
  const text = amountBigText.value.trim()
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    amountCopied.value = true
    if (amountCopiedTimer.value !== null) clearTimeout(amountCopiedTimer.value)
    amountCopiedTimer.value = setTimeout(() => { amountCopied.value = false }, 1500)
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

const toggleSwitchPanel = () => {
  switchPanelOpen.value = !switchPanelOpen.value
  if (switchPanelOpen.value) txPanelOpen.value = false
}

const handleSwitchChannel = (channelId: number) => {
  if (props.switchBusy) return
  if (channelId === props.order?.channelId) return
  emit('switch-channel', channelId)
}

const toggleTxPanel = () => {
  txPanelOpen.value = !txPanelOpen.value
  if (txPanelOpen.value) switchPanelOpen.value = false
}

const handleSubmitTx = () => {
  const hash = txHashInput.value.trim()
  if (!hash || props.txBusy) return
  emit('submit-tx-hash', hash)
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
  walletAddressCopied.value = false
  switchPanelOpen.value = false
  txPanelOpen.value = false
  txHashInput.value = ''
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
    if (walletCopiedTimer.value !== null) {
      clearTimeout(walletCopiedTimer.value)
      walletCopiedTimer.value = null
    }
    if (amountCopiedTimer.value !== null) {
      clearTimeout(amountCopiedTimer.value)
      amountCopiedTimer.value = null
    }
    document.body.style.overflow = ''
    previouslyFocused.value?.focus?.()
  }
}, { immediate: true })

watch(() => props.order, async () => {
  if (!props.open) return
  switchPanelOpen.value = false
  txPanelOpen.value = false
  txHashInput.value = ''
  await renderQRCode()
})

watch(() => props.txFeedback, (feedback) => {
  if (feedback?.ok) txHashInput.value = ''
})

onBeforeUnmount(() => {
  stopNowTimer()
  stopPolling()
  cancelAnimation()
  if (walletCopiedTimer.value !== null) clearTimeout(walletCopiedTimer.value)
  if (amountCopiedTimer.value !== null) clearTimeout(amountCopiedTimer.value)
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
  max-height: min(90vh, 800px);
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(10, 10, 10, .18) transparent;
  background: var(--ux-surface);
  border-radius: 20px;
  box-shadow: 0 30px 60px -36px rgba(40, 30, 80, 0.3), 0 0 0 1px var(--ux-line);
  display: flex; flex-direction: column;
  padding: 20px;
  color: var(--text, #0a0a0a);
  font: 14px/1.5 -apple-system, "Segoe UI", "PingFang SC", sans-serif;
  animation: pay-modal-rise .3s var(--ux-spring);
  outline: none;
}

@keyframes pay-modal-rise { from { transform: translateY(20px) scale(.98); opacity: 0; } }

.pay-modal::-webkit-scrollbar { width: 6px; height: 6px; }
.pay-modal::-webkit-scrollbar-track { background: transparent; }
.pay-modal::-webkit-scrollbar-thumb { background: rgba(10, 10, 10, .14); border-radius: 999px; }
.pay-modal::-webkit-scrollbar-thumb:hover { background: rgba(10, 10, 10, .28); }

/* ---- 金额卡 ---- */
.pay-modal__amount-card {
  position: relative;
  padding: 18px 18px 16px;
  border-radius: 18px;
  background: var(--ux-surface);
  box-shadow: inset 0 0 0 1px var(--ux-line), 0 4px 6px -1px rgba(0, 0, 0, .1), 0 2px 4px -2px rgba(0, 0, 0, .1);
  margin-bottom: 12px;
  transition: opacity .4s, transform .5s var(--ux-ease);
}
.pay-modal__amount-copy {
  position: absolute; top: 16px; right: 16px;
  width: 32px; height: 32px;
  border: 0; border-radius: 8px;
  background: var(--ux-surface-2); color: var(--ux-ink-2);
  display: grid; place-items: center;
  cursor: pointer;
  transition: background-color .15s, color .15s;
}
.pay-modal__amount-copy:hover { background: #ededed; color: #000; }
.pay-modal__amount-copy svg { width: 14px; height: 14px; }
.pay-modal__amount-label {
  display: block;
  font-size: 13.5px; color: var(--ux-ink-3); font-weight: 500;
  margin-bottom: 7px;
}
.pay-modal__fee-note {
  font-style: normal;
  font-size: 12px; font-weight: 400;
  color: var(--ux-ink-3);
}
.pay-modal__amount-value {
  display: flex; flex-direction: column; align-items: flex-start; gap: 7px;
  margin-bottom: 12px;
}
.pay-modal__amount-value b {
  font-size: 36px; font-weight: 700; letter-spacing: -.02em;
  line-height: 1.1;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}
.pay-modal__chain-chip {
  display: inline-flex; align-items: center; gap: 6px;
  height: 26px; padding: 0 11px;
  border-radius: 999px;
  background: var(--ux-surface-2); color: var(--ux-ink-2);
  font-size: 13px; font-weight: 700;
}
.pay-modal__chain-dot { flex: 0 0 auto; width: 11px; height: 11px; border-radius: 50%; background: #a1a1aa; }
.pay-modal__chain-logo { flex: 0 0 auto; width: 18px; height: 18px; border-radius: 50%; }
.pay-modal__chain-dot.is-tron { background: #ef4444; }
.pay-modal__chain-dot.is-bsc { background: #f0b90b; }
.pay-modal__chain-dot.is-sol { background: linear-gradient(135deg, #9945ff, #14f195); }
.pay-modal__chain-dot.is-alipay { background: #1677ff; }
.pay-modal__amount-meta {
  border-top: 1px solid var(--ux-line);
  padding-top: 12px;
  display: flex; flex-direction: column; gap: 8px;
}
.pay-modal__meta-row {
  display: flex; justify-content: space-between; align-items: center; gap: 12px;
  font-size: 14px; color: var(--ux-ink-3);
}
.pay-modal__meta-row b {
  color: var(--ux-ink-2); font-weight: 500;
  font-size: 14px;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere; text-align: right;
}

/* ---- 支付卡 ---- */
.pay-modal__pay-card {
  border-radius: 18px;
  padding: 18px 18px 0;
  margin-bottom: 12px;
  background: var(--ux-surface);
  box-shadow: inset 0 0 0 1px var(--ux-line), 0 4px 6px -1px rgba(0, 0, 0, .1), 0 2px 4px -2px rgba(0, 0, 0, .1);
  transition: opacity .4s, transform .5s var(--ux-ease);
}
.pay-modal__pay-head {
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 6px;
}
.pay-modal__pay-head b { font-size: 15.5px; font-weight: 500; }

.pay-modal__cd-num {
  text-align: center;
  margin: 6px 0 12px;
  font-size: 30px; font-weight: 700;
  letter-spacing: .04em;
  color: #22c55e;
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
  transition: color .3s;
}
.pay-modal__cd-num.is-safe { color: #22c55e; }
.pay-modal__cd-num.is-warn { color: #f59e0b; }
.pay-modal__cd-num.is-late { color: #ea580c; }
.pay-modal__cd-num.is-critical { color: #b02020; }

.pay-modal__qr-wrap {
  display: flex; justify-content: center;
  margin: 0 0 12px;
  transition: filter .4s, opacity .4s;
}
.pay-modal__qr {
  padding: 8px; background: #fff;
  border-radius: 8px;
}
.pay-modal__qr img { display: block; width: 120px; height: 120px; }
.pay-modal__qr-placeholder {
  width: 120px; height: 120px;
  display: grid; place-items: center;
  color: var(--ux-ink-3);
}
.pay-modal__spin { width: 24px; height: 24px; animation: pay-modal-spin 1s linear infinite; }
@keyframes pay-modal-spin { to { transform: rotate(360deg); } }

.is-run .pay-modal__qr-wrap,
.is-ok .pay-modal__qr-wrap { filter: grayscale(1); opacity: .3; }

.pay-modal__redirect-hint {
  display: flex; align-items: flex-start; gap: 8px;
  padding: 10px 12px; margin-bottom: 16px;
  border-radius: 10px;
  background: #fff7e6; color: var(--ux-warn);
  font-size: 12px; line-height: 1.5;
}
.pay-modal__redirect-hint svg { width: 14px; height: 14px; margin-top: 2px; flex-shrink: 0; }

.pay-modal__addr {
  margin-top: 12px;
  padding: 14px 0 16px;
  border-top: 1px solid var(--ux-line);
}
.pay-modal__addr-label {
  font-size: 12px; color: var(--ux-ink-3); font-weight: 600;
  letter-spacing: .04em;
  margin-bottom: 6px;
}
.pay-modal__addr-body {
  display: flex; align-items: center; gap: 10px;
}
.pay-modal__addr-body b {
  flex: 1; min-width: 0;
  font-size: 14px; font-weight: 500;
  line-height: 1.6;
  color: var(--ux-ink-2);
  overflow-wrap: anywhere; word-break: break-all;
}
.pay-modal__addr-copy {
  flex-shrink: 0;
  width: 32px; height: 32px;
  border: 0; border-radius: 8px;
  background: var(--ux-surface-2); color: var(--ux-ink-2);
  display: grid; place-items: center;
  cursor: pointer;
  transition: background-color .15s, color .15s;
}
.pay-modal__addr-copy svg { width: 14px; height: 14px; }
.pay-modal__addr-copy:hover { background: #ededed; color: #000; }

/* ---- 加密货币操作区 ---- */
.pay-modal__crypto-actions {
  display: flex; flex-direction: column; gap: 10px;
  margin-bottom: 12px;
}
.pay-modal__action-row {
  display: flex; gap: 10px;
}
.pay-modal__action-btn {
  display: inline-flex; align-items: center; justify-content: center; gap: 8px;
  flex: 1 1 0; min-width: 0; height: 44px;
  border: 0; border-radius: 14px;
  cursor: pointer;
  font-size: 15.5px; font-weight: 500;
  background: var(--ux-surface); color: var(--ux-ink-2);
  box-shadow: inset 0 0 0 1px var(--ux-line), 0 1px 2px rgba(0, 0, 0, .05);
}
.pay-modal__action-btn:hover:not(:disabled) { color: #000; background: var(--ux-surface-2); }
.pay-modal__action-btn.is-solid {
  background: #343434; color: #fbfbfb;
  box-shadow: none;
}
.pay-modal__action-btn.is-solid:hover:not(:disabled) { color: #fff; background: #262626; }
.pay-modal__action-btn:disabled { opacity: .55; cursor: not-allowed; }
.pay-modal__action-btn svg { width: 17px; height: 17px; }

.pay-modal__panel {
  padding: 12px;
  border-radius: 12px;
  background: var(--ux-surface-2);
  display: flex; flex-direction: column; gap: 8px;
}
.pay-modal__chip-list {
  display: flex; flex-wrap: wrap; gap: 8px;
}
.pay-modal__chip {
  padding: 7px 12px;
  border: 0; border-radius: 999px;
  background: var(--ux-surface); color: var(--ux-ink-2);
  box-shadow: inset 0 0 0 1px var(--ux-line-2);
  font-size: 12.5px; font-weight: 700;
  cursor: pointer;
}
.pay-modal__chip.is-active {
  background: var(--ux-accent); color: #fff;
  box-shadow: none;
}
.pay-modal__chip:disabled { cursor: default; opacity: .9; }
.pay-modal__input {
  width: 100%; height: 40px;
  border: 0; border-radius: 10px;
  padding: 0 12px;
  background: var(--ux-surface);
  box-shadow: inset 0 0 0 1px var(--ux-line-2);
  font-size: 12.5px; color: var(--text, #0a0a0a);
  outline: none;
}
.pay-modal__input:focus { box-shadow: inset 0 0 0 1.5px var(--ux-accent); }
.pay-modal__panel-msg {
  margin: 0;
  font-size: 11.5px; color: var(--ux-mint); font-weight: 600;
  line-height: 1.5;
}
.pay-modal__panel-msg.is-err { color: #b02020; }

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
  min-height: 34px; margin-top: 4px;
  overflow: hidden;
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
  gap: 8px;
  height: auto;
  background: transparent; color: var(--ux-ink-3);
  box-shadow: none;
  font-size: 14px; font-weight: 500;
}
.pay-modal__go.waiting span { display: inline-flex; align-items: center; gap: 8px; }
.pay-modal__go.waiting .pay-modal__spin { width: 16px; height: 16px; color: var(--ux-ink-3); }
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

.pay-modal__copied {
  font-size: 11px; color: var(--ux-mint); font-weight: 600;
}

/* 过期态：对齐支付网关收银台（红 ✗ + 标题 + 重新发起入口），不显示二维码与地址。 */
.pay-modal__expired {
  display: flex; flex-direction: column; align-items: center;
  gap: 10px; padding: 34px 0 26px;
}
.pay-modal__expired-icon { width: 46px; height: 46px; color: #d4342a; stroke-width: 1.6; }
.pay-modal__expired-title { font-size: 16px; font-weight: 700; letter-spacing: -.01em; }
.pay-modal__expired-btn {
  margin-top: 4px; width: 100%; height: 44px;
  border: 0; border-radius: 14px; cursor: pointer;
  background: var(--ux-accent); color: #fff;
  font-size: 14.5px; font-weight: 500;
}

.pay-modal__rec {
  position: absolute;
  left: 16px; right: 16px; bottom: 16px; top: 60px;
  display: none;
  flex-direction: column;
  padding: 22px 18px 16px;
  border-radius: 22px;
  background: var(--ux-surface-2);
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
.is-rec .pay-modal__rec {
  display: flex;
  animation: pay-modal-rec-in .6s var(--ux-spring) both;
}
@keyframes pay-modal-rec-in { from { transform: translateY(60px); opacity: 0; } }

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
  font-size: 17px; font-weight: 700;
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
.pay-modal__rec-dl dt { color: var(--ux-ink-3); font-weight: 600; }
.pay-modal__rec-dl dd {
  margin: 0; text-align: right;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.pay-modal__rec-btn {
  width: 100%; height: 44px;
  border: 0; border-radius: 12px;
  background: var(--ux-accent); color: #fff;
  font-size: 13.5px; font-weight: 700;
  cursor: pointer;
}

.is-rec .pay-modal__amount-card,
.is-rec .pay-modal__pay-card,
.is-rec .pay-modal__crypto-actions,
.is-rec .pay-modal__steps,
.is-rec .pay-modal__stage {
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
