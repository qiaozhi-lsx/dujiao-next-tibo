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
        <!-- 成功态下视觉已淡出（opacity:0 + pointer-events:none），inert 把键盘 Tab 也挡在外头；
             crypto-actions 已被 v-if 过滤掉，不需要 inert。 -->
        <div class="pay-modal__amount-card" :inert="stage === 'success' || undefined">
          <div class="pay-modal__amount-top">
            <small class="pay-modal__amount-label">
              {{ t('qiaozhi.payment.modal.amountPaid') }}
              <em v-if="feeText" class="pay-modal__fee-note">({{ t('qiaozhi.payment.modal.feeIncluded') }})</em>
            </small>
            <div class="pay-modal__amount-value">
              <!-- 复制金额按钮紧跟金额本身，不再悬浮在卡片右上角，弹窗顶部也不必为它额外留白。 -->
              <div class="pay-modal__amount-main">
                <b>{{ amountBigText }}</b>
                <button
                  type="button"
                  class="pay-modal__amount-copy"
                  :aria-label="t('qiaozhi.payment.modal.copyAmount')"
                  @click="handleCopyAmount"
                >
                  <CheckCircle2 v-if="amountCopied" />
                  <Copy v-else />
                </button>
              </div>
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
        <div class="pay-modal__pay-card" :inert="stage === 'success' || undefined">
          <div class="pay-modal__pay-head">
            <b>{{ payCardTitle }}</b>
          </div>

          <div v-if="showCountdown" class="pay-modal__cd-num" :class="`is-${cdTone}`" role="timer">
            {{ countdownText }}
          </div>

          <div v-if="qrSupported && !orderPastExpire" class="pay-modal__qr-wrap">
            <div class="pay-modal__qr">
              <img v-if="qrImageUrl" :src="qrImageUrl" :alt="t('qiaozhi.payment.modal.qrAlt')" />
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
            <button type="button" class="pay-modal__expired-btn" :disabled="renewing" @click="handleRenew">
              {{ t('qiaozhi.payment.modal.renewPayment') }}
            </button>
          </div>
        </div>

        <!-- 加密货币操作：切换网络或币种 / 已转账但未到账 -->
        <div
          v-if="hasCryptoPayment && stage === 'waiting' && !orderPastExpire && (switchableChannels.length || txHashSupported)"
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

            <button
              v-if="txHashSupported"
              type="button"
              class="pay-modal__action-btn is-solid"
              @click="toggleTxPanel"
            >
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
                {{ channelDisplayName(ch.name) }}
              </button>
            </div>
            <p v-if="switchBusy" class="pay-modal__panel-msg">{{ t('qiaozhi.payment.modal.switching') }}</p>
            <p v-else-if="switchError" class="pay-modal__panel-msg is-err">{{ switchError }}</p>
          </div>

          <div v-if="txPanelOpen && txHashSupported" class="pay-modal__panel">
            <input
              v-model="txHashInput"
              class="pay-modal__input"
              type="text"
              maxlength="128"
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

        <div class="pay-modal__stage" aria-live="polite">
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
import { useAppStore } from '../../stores/app'
import { channelDisplayName } from '../../utils/cryptoChain'
import { countdownTone } from '../../utils/countdownTone'
import {
  bareAmountText,
  isPaidOrderStatus,
  isTerminalOrderStatus,
  parsePaidAtMs,
  paymentChipLabel as paymentChipLabel_,
  resolvePayableCurrency,
  shouldShowFee,
} from './utils/paymentDisplay'
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
  feePolicy?: string
  payableAmount?: string
  payableCurrency?: string
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
const appStore = useAppStore()

type Stage = 'waiting' | 'running' | 'success'
const stage = ref<Stage>('waiting')
const modalRef = ref<HTMLElement | null>(null)
const qrImageUrl = ref('')
const walletAddressCopied = ref(false)
const walletCopiedTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const switchPanelOpen = ref(false)
const txPanelOpen = ref(false)
const txHashInput = ref('')
// 倒计时用校正后的服务器时钟，避免客户端改时导致提前 / 延后过期。
const now = ref(appStore.getServerTime())
const nowTimer = ref<ReturnType<typeof setInterval> | null>(null)
const pollTimer = ref<ReturnType<typeof setInterval> | null>(null)
const paidEmitTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const animationFrame = ref<number | null>(null)
// 服务端已返回终态（取消 / 过期 / 退款）：不等倒计时走完，直接进入过期态。
const forcedExpired = ref(false)
// 支付成功时刻（服务端 paid_at），收据上的时间冻结在这一刻而不是随 now 继续走。
const paidAt = ref<number | null>(null)
// 「重新发起支付」在途标志：补查 + emit 期间禁用按钮，防连点并发两次补查而重复建单。
const renewing = ref(false)
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

// 过期态：倒计时走完，或服务端已返回终态（forcedExpired）；已进入支付成功流程后不再显示。
const orderPastExpire = computed(() => {
  if (stage.value !== 'waiting') return false
  if (forcedExpired.value) return true
  return expireMs.value !== null && expireMs.value <= now.value
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

// 「已转账但未到账 → 提交交易哈希」只有 epusdt 网关实现了（GatewayTxHashSubmitter）；
// 其他加密渠道（bepusdt / dujiaopay）同样有收款地址，但后端会直接回"不支持"，不给入口。
// providerType 后端有时下发 'Epusdt' / 'EPUSDT'，大小写不敏感比较。
const txHashSupported = computed(() => (props.order?.providerType || '').toLowerCase() === 'epusdt')

// 金额下方的支付方式 chip：纯函数见 utils/paymentDisplay.paymentChipLabel；实现与测试走同一口径。
const paymentChipLabel = computed(() => paymentChipLabel_(props.order || {}))

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
  // 链上数量后跟 tokenId 解析出的币种；网关没给 tokenId 时只显示数量，不臆测币种。
  if (o.chainAmount) return cryptoTokenLabel.value ? `${o.chainAmount} ${cryptoTokenLabel.value}` : o.chainAmount
  // 实付金额 = 含手续费金额；后端未下发时回退订单金额。
  const payable = (o.payableAmount || '').trim()
  return payable ? formatCurrency(payable, resolvePayableCurrency(o)) : formatCurrency(o.amount, o.currency)
})

const feeText = computed(() => {
  const o = props.order
  if (!o || !shouldShowFee(o)) return ''
  return formatCurrency((o.feeAmount || '').trim(), o.currency)
})

// 加密货币单：链上数量与法币实付的换算关系。网关按 amount(法币) + 实时汇率报出链上数量，
// 两者是一笔的两种币种口径；补这一行，页面上两个币种才能对上账。
const equivalentText = computed(() => {
  const o = props.order
  if (!o || !o.chainAmount) return ''
  const payable = (o.payableAmount || '').trim()
  if (!payable) return ''
  return formatCurrency(payable, resolvePayableCurrency(o))
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

const receiptTimeText = computed(() => formatDateText(paidAt.value ?? now.value))

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
  now.value = appStore.getServerTime()
  nowTimer.value = setInterval(() => {
    now.value = appStore.getServerTime()
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

const clearPaidEmitTimer = () => {
  if (paidEmitTimer.value !== null) {
    clearTimeout(paidEmitTimer.value)
    paidEmitTimer.value = null
  }
}

// order 由触发方在确认已付的那一刻捕获传入：动画结束时才读 props.order 的话，
// 动画期间若订单被换掉（重新发起支付的窄窗口），会对新的未付单 emit paid。
const runSuccessAnimation = (order: PaymentModalOrder | null) => {
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
    const emitPaidDelay = prefersReducedMotion ? 400 : 1500
    // 句柄保存下来：动画期间用户关闭弹窗 / 组件卸载时要取消，否则关掉后仍会触发 paid 跳转。
    paidEmitTimer.value = setTimeout(() => {
      paidEmitTimer.value = null
      if (order) emit('paid', order)
    }, emitPaidDelay)
  }
  animationFrame.value = requestAnimationFrame(tick)
}

const triggerSuccessFromPolling = (paidAtRaw?: string | number | null) => {
  // 此刻的 props.order 就是轮询确认已付的那一单（pollOnce 已校验 sn 未变），先捕获再起动画。
  const order = props.order
  stopPolling()
  if (stage.value !== 'waiting') return
  // 收据时间冻结在服务端支付时刻；解析不了就取当前服务器时间。倒计时钟随之停掉。
  paidAt.value = parsePaidAtMs(paidAtRaw, appStore.getServerTime())
  stopNowTimer()
  stage.value = 'running'
  runSuccessAnimation(order)
}

const pollOnce = async () => {
  const o = props.order
  if (!o || !o.sn) return
  const sn = o.sn
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
    // 等待响应期间可能已关窗或已换单（重新发起支付）：迟到的响应不能再起动画 / 作用到新单上。
    if (!props.open || props.order?.sn !== sn) return
    const data = resp?.data?.data
    const status = String(data?.status || '')
    if (isPaidOrderStatus(status)) {
      triggerSuccessFromPolling(data?.paid_at)
    } else if (isTerminalOrderStatus(status)) {
      // 后台取消 / 过期 / 退款等终态：不必等倒计时走完，停轮询并切到过期态。
      stopPolling()
      forcedExpired.value = true
    }
  } catch {
    // 轮询失败静默忽略；下一次 5s 继续
  }
}

const startPolling = () => {
  stopPolling()
  // 起轮询即刻跑一次：轮询间隔 5s，不即时首发时用户要白等 5 秒才看到「已支付」反馈。
  void pollOnce()
  // 5s 一次，与 classic usePayment 一致；游客订单详情按 IP 限流 120 次/分钟，
  // 2s 轮询在多标签页或同一出口 IP 下容易触顶被封 60s。
  pollTimer.value = setInterval(() => {
    void pollOnce()
  }, 5000)
}

// 过期态下轮询已停，而过期瞬间可能刚完成付款：重发起前先补查一次状态，
// 已支付则直接进入成功流程，避免旧单已付、又新建一单的重复付款窗口。
const handleRenew = async () => {
  if (renewing.value) return
  const sn = props.order?.sn
  renewing.value = true
  try {
    await pollOnce()
    // 补查期间关了窗或已换单（上一次重发起已建出新单），补查结果已被 pollOnce 丢弃，
    // 不能据"stage 仍是 waiting"判定未付再建一单。
    if (!props.open || stage.value !== 'waiting') return
    if (props.order?.sn !== sn) return
    emit('renew-payment')
  } finally {
    renewing.value = false
  }
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
  const o = props.order
  if (!o) return
  const text = bareAmountText(o)
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
  forcedExpired.value = false
  paidAt.value = null
  renewing.value = false
}

watch(() => props.open, async (opened) => {
  if (opened) {
    resetState()
    startNowTimer()
    await renderQRCode()
    // 二维码渲染是异步的：期间若弹窗已被关闭（else 分支已清理），不能再起轮询 / 锁滚动。
    if (!props.open) return
    // 打开时已过期的单不起轮询：orderPastExpire 的 watch 只在 false→true 时停轮询，这里等不到翻转。
    if (props.order?.sn && !orderPastExpire.value) startPolling()
    previouslyFocused.value = (document.activeElement as HTMLElement) || null
    await nextTick()
    if (!props.open) return
    modalRef.value?.focus()
    document.body.style.overflow = 'hidden'
  } else {
    stopNowTimer()
    stopPolling()
    cancelAnimation()
    clearPaidEmitTimer()
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

watch(() => props.order, async (next, prev) => {
  if (!props.open) return
  switchPanelOpen.value = false
  txPanelOpen.value = false
  txHashInput.value = ''
  // 重新发起支付时弹窗不关、直接换成新订单：上一单的终态标记要清掉，
  // 过期时已停掉的轮询也要为新单重新拉起（切换网络同单不换号，不走这里）。
  if (next?.sn && next.sn !== prev?.sn) {
    forcedExpired.value = false
    if (stage.value === 'waiting' && !orderPastExpire.value) startPolling()
  }
  await renderQRCode()
})

// 进入过期态（倒计时走完或服务端终态）后停止轮询；重发起前由 handleRenew 补查一次。
watch(orderPastExpire, (expired) => {
  if (expired) stopPolling()
})

watch(() => props.txFeedback, (feedback) => {
  if (feedback?.ok) txHashInput.value = ''
})

onBeforeUnmount(() => {
  stopNowTimer()
  stopPolling()
  cancelAnimation()
  clearPaidEmitTimer()
  if (walletCopiedTimer.value !== null) clearTimeout(walletCopiedTimer.value)
  if (amountCopiedTimer.value !== null) clearTimeout(amountCopiedTimer.value)
  document.body.style.overflow = ''
})
</script>

<style scoped>
/* 弹窗 Teleport 到 body：颜色 / 字号 / 圆角 / 阴影 / 时长全部读 body.qiaozhi-tokens 上的 --qz-* token
   （暗色随 .dark 自动翻转），不再维护一套 --ux-* 局部变量。 */
.pay-modal-mask {
  position: fixed; inset: 0;
  background: color-mix(in oklab, var(--qz-text) 50%, transparent);
  backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center;
  z-index: 1200;
  animation: pay-modal-fade var(--qz-duration-standard) var(--qz-ease-standard);
}
/* 暗色里 --qz-text 是近白，遮罩改用页面底色压暗（与分类抽屉遮罩同一做法） */
.dark .pay-modal-mask { background: color-mix(in oklab, var(--qz-bg) 72%, transparent); }

@keyframes pay-modal-fade { from { opacity: 0; } }

.pay-modal {
  position: relative;
  width: min(380px, calc(100vw - 32px));
  /* 目标是整个弹窗在视口内一屏放下、不出现内部滚动：等待支付态内容约 690px（英文约 720px），
     展开操作面板后最高约 840px，上下各留 16px。高度由内容决定、不再额外写死上限，
     只有极矮视口（横屏手机等）才回落到内部滚动兜底。 */
  max-height: calc(100vh - 32px);
  max-height: calc(100dvh - 32px);
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--qz-border-strong) transparent;
  background: var(--qz-surface);
  border-radius: var(--qz-radius-panel);
  box-shadow: var(--qz-shadow-floating), 0 0 0 1px var(--qz-border);
  display: flex; flex-direction: column;
  padding: 20px;
  color: var(--qz-text);
  font-family: -apple-system, "Segoe UI", "PingFang SC", sans-serif;
  font-size: var(--qz-text-sm);
  line-height: 1.5;
  animation: pay-modal-rise var(--qz-duration-considered) var(--qz-ease-enter);
  outline: none;
}

@keyframes pay-modal-rise { from { transform: translateY(20px) scale(.98); opacity: 0; } }

.pay-modal::-webkit-scrollbar { width: 6px; height: 6px; }
.pay-modal::-webkit-scrollbar-track { background: transparent; }
.pay-modal::-webkit-scrollbar-thumb { background: var(--qz-border-strong); border-radius: var(--qz-radius-pill); }
.pay-modal::-webkit-scrollbar-thumb:hover { background: var(--qz-text-muted); }

/* ---- 金额卡 ---- */
.pay-modal__amount-card {
  position: relative;
  padding: 16px 16px 14px;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface);
  box-shadow: inset 0 0 0 1px var(--qz-border), var(--qz-shadow-raised);
  margin-bottom: 12px;
  transition: opacity var(--qz-duration-considered) var(--qz-ease-standard), transform var(--qz-duration-considered) var(--qz-ease-standard);
}
.pay-modal__amount-main {
  display: flex; align-items: center; gap: 10px;
  min-width: 0; max-width: 100%;
}
.pay-modal__amount-main b { min-width: 0; }
.pay-modal__amount-copy {
  flex: 0 0 auto;
  width: 28px; height: 28px;
  border: 0; border-radius: var(--qz-radius-pill);
  background: var(--qz-surface-muted); color: var(--qz-text-secondary);
  display: grid; place-items: center;
  cursor: pointer;
  transition: background-color var(--qz-duration-fast) var(--qz-ease-standard), color var(--qz-duration-fast) var(--qz-ease-standard);
}
.pay-modal__amount-copy:hover { background: var(--qz-accent-subtle); color: var(--qz-text); }
.pay-modal__amount-copy svg { width: 14px; height: 14px; }
.pay-modal__amount-label {
  display: block;
  font-size: var(--qz-text-sm); color: var(--qz-text-muted); font-weight: 500;
  margin-bottom: 5px;
}
.pay-modal__fee-note {
  font-style: normal;
  font-size: var(--qz-text-xs); font-weight: 400;
  color: var(--qz-text-muted);
}
.pay-modal__amount-value {
  display: flex; flex-direction: column; align-items: flex-start; gap: 6px;
  margin-bottom: 10px;
}
.pay-modal__amount-value b {
  font-size: var(--qz-text-3xl); font-weight: 700; letter-spacing: -.02em;
  line-height: 1.1;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}
.pay-modal__chain-chip {
  display: inline-flex; align-items: center; gap: 6px;
  height: 26px; padding: 0 11px;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-surface-muted); color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs); font-weight: 700;
}
/* 链 / 渠道色点只在没有 logo 素材时兜底显示；颜色是 tokens.css 的固定品牌色 --qz-brand-tron / -bsc / -sol(-2) / -alipay
   （两个主题同值，品牌色不随主题变）。 */
.pay-modal__chain-dot { flex: 0 0 auto; width: 11px; height: 11px; border-radius: 50%; background: var(--qz-text-muted); }
.pay-modal__chain-logo { flex: 0 0 auto; width: 18px; height: 18px; border-radius: 50%; }
.pay-modal__chain-dot.is-tron { background: var(--qz-brand-tron); }
.pay-modal__chain-dot.is-bsc { background: var(--qz-brand-bsc); }
.pay-modal__chain-dot.is-sol { background: linear-gradient(135deg, var(--qz-brand-sol), var(--qz-brand-sol-2)); }
.pay-modal__chain-dot.is-alipay { background: var(--qz-brand-alipay); }
.pay-modal__amount-meta {
  border-top: 1px solid var(--qz-border);
  padding-top: 10px;
  display: flex; flex-direction: column; gap: 6px;
}
.pay-modal__meta-row {
  display: flex; justify-content: space-between; align-items: center; gap: 12px;
  font-size: var(--qz-text-sm); color: var(--qz-text-muted);
}
.pay-modal__meta-row b {
  color: var(--qz-text-secondary); font-weight: 500;
  font-size: var(--qz-text-sm);
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere; text-align: right;
}

/* ---- 支付卡 ---- */
.pay-modal__pay-card {
  border-radius: var(--qz-radius-control);
  padding: 16px 16px 0;
  margin-bottom: 12px;
  background: var(--qz-surface);
  box-shadow: inset 0 0 0 1px var(--qz-border), var(--qz-shadow-raised);
  transition: opacity var(--qz-duration-considered) var(--qz-ease-standard), transform var(--qz-duration-considered) var(--qz-ease-standard);
}
.pay-modal__pay-head {
  display: flex; align-items: center; justify-content: center;
  text-align: center;
  margin-bottom: 6px;
}
.pay-modal__pay-head b { font-size: var(--qz-text-base); font-weight: 500; }

/* 倒计时在标题下方独立一行居中；字号与上下留白比最初（30px / 6 0 12）收紧，好让弹窗一屏放下。
   四档颜色沿用 countdownTone：safe 成功色 → warn 警告色 → late 警告 / 危险各半 → critical 危险色。 */
.pay-modal__cd-num {
  text-align: center;
  margin: 2px 0 8px;
  font-size: var(--qz-text-xl); font-weight: 700; line-height: 1.1;
  letter-spacing: .04em;
  color: var(--qz-success);
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
  transition: color var(--qz-duration-standard) var(--qz-ease-standard);
}
.pay-modal__cd-num.is-safe { color: var(--qz-success); }
.pay-modal__cd-num.is-warn { color: var(--qz-warning); }
.pay-modal__cd-num.is-late { color: color-mix(in oklab, var(--qz-warning) 50%, var(--qz-danger)); }
.pay-modal__cd-num.is-critical { color: var(--qz-danger); }

.pay-modal__qr-wrap {
  display: flex; justify-content: center;
  margin: 0 0 8px;
  transition: filter var(--qz-duration-considered) var(--qz-ease-standard), opacity var(--qz-duration-considered) var(--qz-ease-standard);
}
.pay-modal__qr {
  padding: 8px; background: var(--qz-qr-frame);
  border-radius: var(--qz-radius-control);
}
.pay-modal__qr img { display: block; width: 120px; height: 120px; }
.pay-modal__qr-placeholder {
  width: 120px; height: 120px;
  display: grid; place-items: center;
  color: var(--qz-text-muted);
}
.pay-modal__spin { width: 24px; height: 24px; animation: pay-modal-spin 1s linear infinite; }
@keyframes pay-modal-spin { to { transform: rotate(360deg); } }

.is-run .pay-modal__qr-wrap,
.is-ok .pay-modal__qr-wrap { filter: grayscale(1); opacity: .3; }

.pay-modal__redirect-hint {
  display: flex; align-items: flex-start; gap: 8px;
  padding: 10px 12px; margin-bottom: 16px;
  border-radius: var(--qz-radius-control);
  background: color-mix(in oklab, var(--qz-warning) 12%, transparent); color: var(--qz-warning);
  font-size: var(--qz-text-xs); line-height: 1.5;
}
.pay-modal__redirect-hint svg { width: 14px; height: 14px; margin-top: 2px; flex-shrink: 0; }

.pay-modal__addr {
  margin-top: 8px;
  padding: 10px 0 12px;
  border-top: 1px solid var(--qz-border);
}
.pay-modal__addr-label {
  font-size: var(--qz-text-xs); color: var(--qz-text-muted); font-weight: 600;
  letter-spacing: .04em;
  margin-bottom: 6px;
}
.pay-modal__addr-body {
  display: flex; align-items: center; gap: 10px;
}
.pay-modal__addr-body b {
  flex: 1; min-width: 0;
  font-size: var(--qz-text-sm); font-weight: 500;
  line-height: 1.6;
  color: var(--qz-text-secondary);
  overflow-wrap: anywhere; word-break: break-all;
}
.pay-modal__addr-copy {
  flex-shrink: 0;
  width: 32px; height: 32px;
  border: 0; border-radius: var(--qz-radius-pill);
  background: var(--qz-surface-muted); color: var(--qz-text-secondary);
  display: grid; place-items: center;
  cursor: pointer;
  transition: background-color var(--qz-duration-fast) var(--qz-ease-standard), color var(--qz-duration-fast) var(--qz-ease-standard);
}
.pay-modal__addr-copy svg { width: 14px; height: 14px; }
.pay-modal__addr-copy:hover { background: var(--qz-accent-subtle); color: var(--qz-text); }

/* ---- 加密货币操作区 ---- */
.pay-modal__crypto-actions {
  display: flex; flex-direction: column; gap: 8px;
  margin-bottom: 12px;
}
.pay-modal__action-row {
  display: flex; gap: 10px;
}
.pay-modal__action-btn {
  display: inline-flex; align-items: center; justify-content: center; gap: 6px;
  /* 两个按钮平分一行（手机上每个只有约 154px 宽），英文等长文案会折成两行：高度跟随内容而不是写死 44px，
     否则第二行会溢出按钮边框。字号与弹窗正文一致（14px），中文文案在手机宽度下也能保持一行。 */
  flex: 1 1 0; min-width: 0; min-height: var(--qz-control-md);
  padding: 8px 10px;
  border: 0; border-radius: var(--qz-radius-pill);
  cursor: pointer;
  font-size: var(--qz-text-sm); font-weight: 500; line-height: 1.2; text-align: center;
  background: var(--qz-surface); color: var(--qz-text-secondary);
  box-shadow: inset 0 0 0 1px var(--qz-border), var(--qz-shadow-resting);
}
.pay-modal__action-btn:hover:not(:disabled) { color: var(--qz-text); background: var(--qz-surface-muted); }
.pay-modal__action-btn.is-solid {
  background: var(--qz-accent-strong); color: var(--qz-on-accent);
  box-shadow: none;
}
.pay-modal__action-btn.is-solid:hover:not(:disabled) { color: var(--qz-on-accent); background: var(--qz-accent); }
.pay-modal__action-btn:disabled { opacity: .55; cursor: not-allowed; }
.pay-modal__action-btn svg { width: 17px; height: 17px; flex: 0 0 auto; }

.pay-modal__panel {
  padding: 10px;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  display: flex; flex-direction: column; gap: 6px;
}
.pay-modal__chip-list {
  display: flex; flex-wrap: wrap; gap: 8px;
}
.pay-modal__chip {
  padding: 7px 12px;
  border: 0; border-radius: var(--qz-radius-pill);
  background: var(--qz-surface); color: var(--qz-text-secondary);
  box-shadow: inset 0 0 0 1px var(--qz-border-strong);
  font-size: var(--qz-text-xs); font-weight: 700;
  cursor: pointer;
}
.pay-modal__chip.is-active {
  background: var(--qz-accent); color: var(--qz-on-accent);
  box-shadow: none;
}
.pay-modal__chip:disabled { cursor: default; opacity: .9; }
.pay-modal__input {
  width: 100%; height: 40px;
  border: 0; border-radius: var(--qz-radius-control);
  padding: 0 12px;
  background: var(--qz-surface);
  box-shadow: inset 0 0 0 1px var(--qz-border-strong);
  font-size: var(--qz-text-xs); color: var(--qz-text);
  outline: none;
}
.pay-modal__input:focus { box-shadow: inset 0 0 0 1.5px var(--qz-accent); }
.pay-modal__panel-msg {
  margin: 0;
  font-size: var(--qz-text-xs); color: var(--qz-success); font-weight: 600;
  line-height: 1.5;
}
.pay-modal__panel-msg.is-err { color: var(--qz-danger); }

.pay-modal__steps {
  display: flex; gap: 14px; justify-content: center;
  height: 0; margin: 0; overflow: hidden;
  font-size: var(--qz-text-xs); font-weight: 700; color: var(--qz-text-muted);
  opacity: 0;
  transition: opacity var(--qz-duration-standard) var(--qz-ease-standard), height var(--qz-duration-standard) var(--qz-ease-standard), margin var(--qz-duration-standard) var(--qz-ease-standard);
}
.is-run .pay-modal__steps,
.is-ok .pay-modal__steps { opacity: 1; height: 20px; margin: 4px 0 14px; }
.pay-modal__steps span {
  display: inline-flex; align-items: center; gap: 5px;
  transition: color var(--qz-duration-standard) var(--qz-ease-standard);
}
.pay-modal__steps span::before {
  content: "";
  width: 7px; height: 7px; border-radius: 50%;
  background: var(--qz-border-strong);
  transition: background-color var(--qz-duration-standard) var(--qz-ease-standard), transform var(--qz-duration-spring) var(--qz-ease-spring);
}
.pay-modal__steps span.on { color: var(--qz-text); }
.pay-modal__steps span.on::before { background: var(--qz-accent); transform: scale(1.3); }
.is-ok .pay-modal__steps span.on::before { background: var(--qz-success); }

.pay-modal__stage {
  position: relative;
  display: flex; align-items: center; justify-content: center;
  min-height: 34px; margin-top: 0;
  overflow: hidden;
  transition: min-height var(--qz-duration-considered) var(--qz-ease-standard);
}
.is-run .pay-modal__stage,
.is-ok .pay-modal__stage { min-height: 90px; }

.pay-modal__go {
  position: relative;
  display: flex; align-items: center; justify-content: center;
  width: 100%; height: 52px;
  border: 0; border-radius: var(--qz-radius-pill);
  background: var(--qz-accent); color: var(--qz-on-accent);
  font-size: var(--qz-text-sm); font-weight: 800;
  overflow: hidden;
  transition: width var(--qz-duration-spring) var(--qz-ease-spring), height var(--qz-duration-spring) var(--qz-ease-spring),
              border-radius var(--qz-duration-spring) var(--qz-ease-spring), background-color var(--qz-duration-considered) var(--qz-ease-standard);
}
.pay-modal__go.waiting {
  gap: 8px;
  height: auto;
  background: transparent; color: var(--qz-text-muted);
  box-shadow: none;
  font-size: var(--qz-text-sm); font-weight: 500;
}
.pay-modal__go.waiting span { display: inline-flex; align-items: center; gap: 8px; }
.pay-modal__go.waiting .pay-modal__spin { width: 16px; height: 16px; color: var(--qz-text-muted); }
.is-run .pay-modal__go,
.is-ok .pay-modal__go {
  width: 56px; height: 56px; border-radius: 50%;
  pointer-events: none;
  background: var(--qz-accent); color: var(--qz-on-accent);
  box-shadow: none;
}
.is-run .pay-modal__go span,
.is-ok .pay-modal__go span { opacity: 0; }
.is-ok .pay-modal__go { background: var(--qz-success); }

.pay-modal__ring {
  pointer-events: none;
  position: absolute;
  width: 80px; height: 80px;
  top: 50%; left: 50%;
  margin: -40px 0 0 -40px;
  transform: rotate(-90deg);
  opacity: 0;
  transition: opacity var(--qz-duration-standard) var(--qz-ease-standard);
}
.is-run .pay-modal__ring,
.is-ok .pay-modal__ring { opacity: 1; transition-delay: var(--qz-duration-standard); }
.pay-modal__ring circle { fill: none; stroke-width: 5; stroke-linecap: round; }
.pay-modal__ring .t { stroke: var(--qz-border); }
.pay-modal__ring .p {
  stroke: var(--qz-accent);
  stroke-dasharray: 100;
  stroke-dashoffset: calc(100 - var(--p, 0) * 100);
  transition: stroke var(--qz-duration-considered) var(--qz-ease-standard);
}
.is-ok .pay-modal__ring .p { stroke: var(--qz-success); }

.pay-modal__ck {
  position: absolute;
  top: 50%; left: 50%;
  margin: -12px 0 0 -12px;
  width: 24px; height: 24px;
  fill: none; stroke: var(--qz-on-accent);
  stroke-width: 3.4; stroke-linecap: round; stroke-linejoin: round;
  pointer-events: none;
}
.pay-modal__ck path {
  stroke-dasharray: 24;
  stroke-dashoffset: 24;
  transition: stroke-dashoffset var(--qz-duration-considered) var(--qz-ease-standard) var(--qz-duration-fast);
}
.is-ok .pay-modal__ck path { stroke-dashoffset: 0; }

.pay-modal__copied {
  font-size: var(--qz-text-2xs); color: var(--qz-success); font-weight: 600;
}

/* 过期态：对齐支付网关收银台（红 ✗ + 标题 + 重新发起入口），不显示二维码与地址。 */
.pay-modal__expired {
  display: flex; flex-direction: column; align-items: center;
  gap: 10px; padding: 34px 0 26px;
}
.pay-modal__expired-icon { width: 46px; height: 46px; color: var(--qz-danger); stroke-width: 1.6; }
.pay-modal__expired-title { font-size: var(--qz-text-base); font-weight: 700; letter-spacing: -.01em; }
.pay-modal__expired-btn {
  margin-top: 4px; width: 100%; height: var(--qz-control-md);
  border: 0; border-radius: var(--qz-radius-pill); cursor: pointer;
  background: var(--qz-accent); color: var(--qz-on-accent);
  font-size: var(--qz-text-sm); font-weight: 500;
}
.pay-modal__expired-btn:disabled { opacity: .55; cursor: not-allowed; }

.pay-modal__rec {
  position: absolute;
  /* 顶部内边距已从 50px 收到 20px，收据层随之上移，四边留白接近一致。 */
  left: 16px; right: 16px; bottom: 16px; top: 24px;
  display: none;
  flex-direction: column;
  padding: 22px 18px 16px;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  text-align: center;
  z-index: 2;
  overflow-y: auto;
}
.pay-modal__rec::before {
  content: "";
  position: absolute; left: 0; right: 0; top: -8px;
  height: 16px;
  background: radial-gradient(circle 8px at 8px 0, transparent 98%, var(--qz-surface-muted)) 0 8px/16px 8px repeat-x;
}
.is-rec .pay-modal__rec {
  display: flex;
  animation: pay-modal-rec-in var(--qz-duration-spring) var(--qz-ease-spring) both;
}
@keyframes pay-modal-rec-in { from { transform: translateY(60px); opacity: 0; } }

.pay-modal__rec-header {
  font-size: var(--qz-text-2xs); letter-spacing: 1.3px;
  color: var(--qz-text-muted); font-weight: 600;
  padding-bottom: 12px;
  border-bottom: 1px dashed var(--qz-border-strong);
  margin-bottom: 16px;
}
.pay-modal__rec-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 12px;
  background: var(--qz-surface);
  border-radius: var(--qz-radius-control);
  margin-bottom: 14px;
  text-align: left;
  box-shadow: inset 0 0 0 1px var(--qz-border);
}
.pay-modal__rec-icon {
  width: 36px; height: 36px; flex-shrink: 0;
  border-radius: var(--qz-radius-pill);
  background: linear-gradient(135deg, var(--qz-accent), var(--qz-accent-strong));
  display: grid; place-items: center;
  color: var(--qz-on-accent); font-size: var(--qz-text-sm); font-weight: 800;
}
.pay-modal__rec-meta { flex: 1; min-width: 0; text-align: left; }
.pay-modal__rec-name {
  font-size: var(--qz-text-sm); font-weight: 700;
  color: var(--qz-text); margin-bottom: 2px;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.pay-modal__rec-sub {
  font-size: var(--qz-text-2xs); color: var(--qz-text-muted);
  display: flex; align-items: center; gap: 5px;
}
/* 状态点画成 6px 圆（而非 7px 的 "●" 字形），模板字号地板是 11px */
.pay-modal__rec-sub::before {
  content: ""; flex: none; width: 6px; height: 6px; border-radius: 50%; background: var(--qz-success);
}
.pay-modal__rec-amt {
  font-size: var(--qz-text-sm); font-weight: 800;
  color: var(--qz-text);
  font-variant-numeric: tabular-nums;
}

.pay-modal__rec-check {
  width: 52px; height: 52px;
  margin: 0 auto 10px;
  border-radius: 50%;
  display: grid; place-items: center;
  background: var(--qz-success); color: var(--qz-on-accent);
  box-shadow: 0 0 0 7px color-mix(in oklab, var(--qz-success) 16%, transparent);
}
.pay-modal__rec-check svg {
  width: 24px; height: 24px;
  fill: none; stroke: currentColor;
  stroke-width: 3; stroke-linecap: round; stroke-linejoin: round;
}

.pay-modal__rec-title {
  display: block;
  font-size: var(--qz-text-lg); font-weight: 700;
  letter-spacing: -.03em;
}

.pay-modal__rec > p {
  margin: 4px 0 12px;
  font-size: var(--qz-text-xs); color: var(--qz-text-muted); line-height: 1.55;
}
.pay-modal__rec-email { color: var(--qz-text-secondary); font-weight: 600; }

.pay-modal__rec-dl {
  display: grid; grid-template-columns: auto 1fr;
  gap: 6px 12px;
  margin: 0 0 16px;
  text-align: left;
  font-size: var(--qz-text-xs);
}
.pay-modal__rec-dl dt { color: var(--qz-text-muted); font-weight: 600; }
.pay-modal__rec-dl dd {
  margin: 0; text-align: right;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.pay-modal__rec-btn {
  width: 100%; height: var(--qz-control-md);
  border: 0; border-radius: var(--qz-radius-pill);
  background: var(--qz-accent); color: var(--qz-on-accent);
  font-size: var(--qz-text-sm); font-weight: 700;
  cursor: pointer;
}

.is-rec .pay-modal__amount-card,
.is-rec .pay-modal__pay-card,
.is-rec .pay-modal__crypto-actions,
.is-rec .pay-modal__steps,
.is-rec .pay-modal__stage {
  opacity: 0; transform: translateY(-16px);
  pointer-events: none;
  transition: opacity var(--qz-duration-considered) var(--qz-ease-standard), transform var(--qz-duration-considered) var(--qz-ease-standard);
}

/* 矮视口（如 iPhone Safari 带工具条时约 664px 高）的紧凑模式：缩小大字金额与二维码、再收一轮内边距，
   让等待支付态仍尽量一屏放下而不触发内部滚动。 */
@media (max-height: 740px) {
  .pay-modal { padding: 16px; }
  .pay-modal__amount-card { padding: 10px 14px 8px; margin-bottom: 8px; }
  .pay-modal__amount-value { margin-bottom: 6px; }
  .pay-modal__amount-value b { font-size: var(--qz-text-2xl); }
  .pay-modal__amount-meta { padding-top: 6px; gap: 4px; }
  .pay-modal__pay-card { padding: 14px 14px 0; margin-bottom: 8px; }
  .pay-modal__pay-head { margin-bottom: 4px; }
  .pay-modal__cd-num { margin: 0 0 6px; font-size: var(--qz-text-lg); }
  .pay-modal__qr-wrap { margin-bottom: 6px; }
  .pay-modal__qr img,
  .pay-modal__qr-placeholder { width: 96px; height: 96px; }
  .pay-modal__addr { margin-top: 6px; padding: 6px 0 8px; }
  .pay-modal__crypto-actions { margin-bottom: 8px; }
  .pay-modal__stage { min-height: 30px; }
}

@media (prefers-reduced-motion: reduce) {
  .pay-modal, .pay-modal *, .pay-modal-mask, .pay-modal-mask * {
    animation-duration: .001ms !important;
    transition-duration: .001ms !important;
  }
}
</style>
