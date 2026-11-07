<template>
  <div class="qz-me-panel">
    <!-- 余额概览 -->
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.wallet.title') }}</h2>
          <small>{{ t('personalCenter.wallet.subtitle') }}</small>
        </div>
        <div class="qz-me-acts">
          <span class="qz-chip">{{ t('personalCenter.tabs.wallet') }}</span>
        </div>
      </div>

      <p v-if="walletAlert" class="qz-alert" :class="walletAlertClass" role="alert">
        {{ walletAlert.message }}
      </p>

      <div class="qz-stat-grid is-triple">
        <div class="qz-stat">
          <b>{{ balanceDisplay }}</b>
          <span>{{ t('personalCenter.wallet.balanceLabel') }}</span>
        </div>
        <div class="qz-stat">
          <b>{{ pagination.total }}</b>
          <span>{{ t('personalCenter.wallet.transactionsLabel') }}</span>
        </div>
        <div class="qz-stat">
          <b>{{ t('orders.pageInfo', { page: pagination.page, total: pagination.total_page }) }}</b>
          <span>{{ t('personalCenter.wallet.currentPageLabel') }}</span>
        </div>
      </div>
    </section>

    <!-- 充值 -->
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.wallet.rechargeTitle') }}</h2>
          <small>{{ t('personalCenter.wallet.rechargeSubtitle') }}</small>
        </div>
      </div>

      <form class="qz-me-form" novalidate @submit.prevent="handleRecharge">
        <div class="qz-me-form-row">
          <div class="qz-field">
            <label class="qz-field-label" for="qz-wallet-amount">{{ t('personalCenter.wallet.amountLabel') }}</label>
            <input
              id="qz-wallet-amount"
              v-model.trim="rechargeForm.amount"
              class="qz-field-control"
              type="text"
              inputmode="decimal"
              autocomplete="off"
              :placeholder="t('personalCenter.wallet.amountPlaceholder')"
            />
          </div>
          <div class="qz-field">
            <label class="qz-field-label" for="qz-wallet-channel">{{ t('personalCenter.wallet.channelLabel') }}</label>
            <QzSelect
              id="qz-wallet-channel"
              :model-value="rechargeForm.channelId"
              :items="[{ value: 0, label: t('personalCenter.wallet.channelPlaceholder') }, ...channels.map((channel: any) => ({ value: Number(channel.id), label: channelDisplayName(String(channel.name)) }))]"
              :disabled="!hasChannels || channelLoading || recharging"
              @update:model-value="rechargeForm.channelId = $event as number"
            />
          </div>
        </div>

        <div class="qz-field">
          <label class="qz-field-label" for="qz-wallet-remark">{{ t('personalCenter.wallet.remarkLabel') }}</label>
          <input
            id="qz-wallet-remark"
            v-model.trim="rechargeForm.remark"
            class="qz-field-control"
            type="text"
            autocomplete="off"
            :placeholder="t('personalCenter.wallet.remarkPlaceholder')"
          />
        </div>

        <p v-if="selectedChannelAmountHint" class="qz-field-error" role="alert">{{ selectedChannelAmountHint }}</p>

        <!-- 手续费（通道向用户收取时展示） -->
        <div v-if="selectedChannel?.fee_policy === 'customer_surcharge'" class="qz-me-sub">
          <div class="qz-me-kv">
            <div class="qz-me-kv-row"><span>{{ t('payment.feeRateLabel') }}</span><b>{{ selectedChannelFeeRateDisplay }}</b></div>
            <div class="qz-me-kv-row"><span>{{ t('payment.fixedFeeLabel') }}</span><b>{{ selectedChannelFixedFeeDisplay }}</b></div>
            <div class="qz-me-kv-row"><span>{{ t('payment.feeAmountLabel') }}</span><b>{{ selectedChannelFeeAmountDisplay }}</b></div>
          </div>
        </div>

        <p v-if="channelLoading" class="qz-field-hint">{{ t('common.loading') }}</p>
        <p v-else-if="!hasChannels" class="qz-field-hint is-warn">{{ t('payment.channelEmpty') }}</p>

        <QzButton
          type="submit"
          size="lg"
          class="qz-btn--block"
          :loading="recharging"
          :disabled="channelLoading || !hasChannels"
        >
          <CreditCard v-if="!recharging" />
          <span>{{ recharging ? t('personalCenter.wallet.recharging') : t('personalCenter.wallet.rechargeSubmit') }}</span>
        </QzButton>
      </form>
    </section>

    <!-- 交易明细 -->
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.wallet.detailTitle') }}</h2>
        </div>
        <div class="qz-me-acts">
          <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm" :disabled="loading" @click="refreshCurrentPage">
            <RefreshCw />
            {{ t('orders.filters.refresh') }}
          </button>
        </div>
      </div>

      <div v-if="loading" class="qz-me-rows" aria-hidden="true">
        <div v-for="idx in 3" :key="idx" class="qz-skeleton qz-me-sk"></div>
      </div>
      <p v-else-if="transactions.length === 0" class="qz-field-hint">
        {{ t('personalCenter.wallet.empty') }}
      </p>
      <div v-else class="qz-me-rows">
        <div v-for="item in transactions" :key="item.id" class="qz-card qz-me-row">
          <div class="qz-me-row-main">
            <b
              class="qz-me-tx-amount"
              :class="item.direction === 'in' ? 'is-in' : item.direction === 'out' ? 'is-out' : ''"
            >{{ signedAmount(item.direction, item.amount, item.currency) }}</b>
            <small>{{ transactionTypeLabel(item.type) }} · {{ formatDate(item.created_at) }}</small>
            <small v-if="item.remark">{{ item.remark }}</small>
            <small>{{ t('personalCenter.wallet.table.balanceAfter') }}：{{ formatMoney(item.balance_after, item.currency) }}</small>
          </div>
          <div class="qz-me-row-side">
            <span class="qz-badge" :class="badgeToneClass(directionTone(item.direction))">
              {{ directionLabel(item.direction) }}
            </span>
          </div>
        </div>
      </div>

      <nav v-if="pagination.total_page > 1" class="qz-me-pager" :aria-label="t('orders.pageInfo', { page: pagination.page, total: pagination.total_page })">
        <button
          type="button"
          class="qz-icon-btn qz-icon-btn--sm"
          :disabled="pagination.page <= 1 || loading"
          :aria-label="t('pagination.previous')"
          @click="changePage(pagination.page - 1)"
        >
          <ChevronLeft />
        </button>
        <span>{{ pagination.page }} / {{ pagination.total_page }}</span>
        <button
          type="button"
          class="qz-icon-btn qz-icon-btn--sm"
          :disabled="pagination.page >= pagination.total_page || loading"
          :aria-label="t('pagination.next')"
          @click="changePage(pagination.page + 1)"
        >
          <ChevronRight />
        </button>
      </nav>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, ChevronRight, CreditCard, RefreshCw } from 'lucide-vue-next'
import { walletAPI } from '../../../api'
import { channelDisplayName } from '../../../utils/cryptoChain'
import { useAppStore } from '../../../stores/app'
import type { PageAlert } from '../../../utils/alerts'
import {
  amountToCents,
  basisPointsToPercent,
  calculateFeeCents,
  centsToAmount,
  rateToBasisPoints,
} from '../../../utils/money'
import QzButton from '../components/QzButton.vue'
import QzSelect from '../components/QzSelect.vue'
import { badgeToneClass, formatDate, formatMoney, type MeBadgeTone } from './support'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const loading = ref(true)
const recharging = ref(false)
const wallet = ref<any>(null)
const transactions = ref<any[]>([])
const pagination = ref({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})
const walletAlert = ref<PageAlert | null>(null)
const channels = ref<any[]>([])
const channelFetchTimer = ref<number | null>(null)
const channelFetchSeq = ref(0)
const channelLoading = ref(false)
const channelsResolvedAmount = ref('')

const rechargeForm = reactive({
  amount: '',
  channelId: 0,
  remark: '',
})

// 提示条 tone（base .qz-alert）：success → is-success、warning → is-warning、error → is-danger
const walletAlertClass = computed(() => {
  const level = walletAlert.value?.level
  if (level === 'error') return 'is-danger'
  if (level === 'success') return 'is-success'
  return 'is-warning'
})

const hasChannels = computed(() => {
  const amount = rechargeForm.amount.trim()
  const amountCents = amountToCents(amount)
  if (!amount || amountCents === null || amountCents <= 0) {
    return true
  }
  if (channelLoading.value) {
    return true
  }
  if (channelsResolvedAmount.value !== amount) {
    return true
  }
  return channels.value.length > 0
})

const selectedChannel = computed(() => channels.value.find((item: any) => item.id === rechargeForm.channelId) || null)

const EPAY_ALLOWED_CHANNEL_TYPES = new Set(['wechat', 'wxpay', 'alipay', 'qqpay'])

const channelLimitMeta = (channel?: any) => {
  const minCents = amountToCents(String(channel?.min_amount ?? ''))
  const maxCents = amountToCents(String(channel?.max_amount ?? ''))
  return {
    minCents,
    maxCents,
    hasMin: minCents !== null && minCents > 0,
    hasMax: maxCents !== null && maxCents > 0,
  }
}

const isSupportedEpayChannel = (channel: any) => {
  const providerType = String(channel?.provider_type || '').toLowerCase()
  if (providerType !== 'epay') return true
  const channelType = String(channel?.channel_type || '').toLowerCase()
  return EPAY_ALLOWED_CHANNEL_TYPES.has(channelType)
}

const isChannelOutOfRange = (channel: any, targetAmountCents: number) => {
  const meta = channelLimitMeta(channel)
  if (!meta.hasMin && !meta.hasMax) return false
  const lessThanMin = meta.hasMin && meta.minCents !== null && targetAmountCents < meta.minCents
  const greaterThanMax = meta.hasMax && meta.maxCents !== null && targetAmountCents > meta.maxCents
  return Boolean(lessThanMin || greaterThanMax)
}

const shouldHideChannelForAmount = (channel: any, targetAmountCents: number) => {
  if (!Boolean(channel?.hide_amount_out_range)) return false
  return isChannelOutOfRange(channel, targetAmountCents)
}

const getAllowedChannelIdSet = () => {
  const allowedIds = appStore.config?.wallet_recharge_channel_ids
  if (!Array.isArray(allowedIds) || allowedIds.length === 0) return null
  return new Set(allowedIds.map(Number))
}

const isChannelAllowedByConfig = (channel: any, allowedIdSet: Set<number> | null) => {
  if (!allowedIdSet) return true
  return allowedIdSet.has(Number(channel?.id))
}

const mapChannel = (channel: any) => ({
  id: Number(channel.id),
  name: String(channel.name || channel.channel_type || channel.id),
  channel_type: String(channel.channel_type || ''),
  fee_policy: String(channel.fee_policy || ''),
  fee_rate: String(channel.fee_rate ?? '0'),
  fixed_fee: String(channel.fixed_fee ?? '0'),
  min_amount: String(channel.min_amount ?? '0'),
  max_amount: String(channel.max_amount ?? '0'),
  hide_amount_out_range: Boolean(channel.hide_amount_out_range),
})

const normalizeChannels = (list: any[], targetAmountCents: number) => {
  const allowedIdSet = getAllowedChannelIdSet()
  return list
    .filter((channel: any) => isSupportedEpayChannel(channel))
    .filter((channel: any) => !shouldHideChannelForAmount(channel, targetAmountCents))
    .filter((channel: any) => isChannelAllowedByConfig(channel, allowedIdSet))
    .map(mapChannel)
    .filter((channel: any) => Number.isFinite(channel.id) && channel.id > 0)
}

const getSelectedChannelAmountHint = (channel: any, amountCents: number) => {
  if (!isChannelOutOfRange(channel, amountCents)) return ''

  const meta = channelLimitMeta(channel)
  if (meta.hasMin && meta.hasMax && meta.minCents !== null && meta.maxCents !== null) {
    return t('payment.channelAmountLimitHint', {
      min: formatMoney(centsToAmount(meta.minCents), selectedChannelCurrency.value),
      max: formatMoney(centsToAmount(meta.maxCents), selectedChannelCurrency.value),
    })
  }
  if (meta.hasMin && meta.minCents !== null) {
    return t('payment.channelAmountMinHint', {
      min: formatMoney(centsToAmount(meta.minCents), selectedChannelCurrency.value),
    })
  }
  if (meta.hasMax && meta.maxCents !== null) {
    return t('payment.channelAmountMaxHint', {
      max: formatMoney(centsToAmount(meta.maxCents), selectedChannelCurrency.value),
    })
  }
  return ''
}

const selectedChannelAmountHint = computed(() => {
  const channel = selectedChannel.value
  if (!channel) return ''
  const amountCents = amountToCents(rechargeForm.amount.trim())
  if (amountCents === null || amountCents <= 0) return ''
  return getSelectedChannelAmountHint(channel, amountCents)
})

const loadPaymentChannels = async (seq: number, amount: string) => {
  if (seq !== channelFetchSeq.value) return
  const amountCents = amountToCents(amount)
  if (!amount || amountCents === null || amountCents <= 0) {
    if (seq !== channelFetchSeq.value) return
    channels.value = []
    channelsResolvedAmount.value = ''
    return
  }

  try {
    const response = await walletAPI.getPaymentChannels(amount)
    if (seq !== channelFetchSeq.value) return
    const list = Array.isArray(response.data.data) ? response.data.data : []
    channels.value = normalizeChannels(list, amountCents)
  } catch {
    if (seq !== channelFetchSeq.value) return
    channels.value = []
  } finally {
    if (seq === channelFetchSeq.value) {
      channelsResolvedAmount.value = amount
      channelLoading.value = false
    }
  }
}

const scheduleLoadPaymentChannels = () => {
  const amount = rechargeForm.amount.trim()
  const amountCents = amountToCents(amount)
  if (!amount || amountCents === null || amountCents <= 0) {
    channelFetchSeq.value += 1
    if (channelFetchTimer.value) {
      window.clearTimeout(channelFetchTimer.value)
      channelFetchTimer.value = null
    }
    channelLoading.value = false
    channelsResolvedAmount.value = ''
    channels.value = []
    return
  }

  const seq = channelFetchSeq.value + 1
  channelFetchSeq.value = seq
  channelLoading.value = true

  if (channelFetchTimer.value) {
    window.clearTimeout(channelFetchTimer.value)
    channelFetchTimer.value = null
  }
  channelFetchTimer.value = window.setTimeout(() => {
    channelFetchTimer.value = null
    void loadPaymentChannels(seq, amount)
  }, 300)
}

const selectedChannelCurrency = computed(() => String(appStore.config?.currency || 'CNY'))
const selectedChannelFeeRateDisplay = computed(() => {
  const rate = rateToBasisPoints(selectedChannel.value?.fee_rate)
  if (rate === null) return '0.00%'
  return `${basisPointsToPercent(rate)}%`
})
const selectedChannelFixedFeeDisplay = computed(() => {
  return formatMoney(String(selectedChannel.value?.fixed_fee ?? '0.00'), selectedChannelCurrency.value)
})
const selectedChannelFeeAmountDisplay = computed(() => {
  const amountCents = amountToCents(rechargeForm.amount)
  if (amountCents === null || amountCents <= 0) return formatMoney('0.00', selectedChannelCurrency.value)
  const rate = rateToBasisPoints(selectedChannel.value?.fee_rate) || 0
  const fixedFeeCents = amountToCents(selectedChannel.value?.fixed_fee) || 0
  const variableFeeCents = calculateFeeCents(amountCents, rate) || 0
  return formatMoney(centsToAmount(variableFeeCents + fixedFeeCents), selectedChannelCurrency.value)
})
const balanceDisplay = computed(() => formatMoney(wallet.value?.balance, String(appStore.config?.currency || 'CNY')))

const directionLabel = (direction?: string) => {
  if (direction === 'in') return t('personalCenter.wallet.directionIn')
  if (direction === 'out') return t('personalCenter.wallet.directionOut')
  return direction || '-'
}

const directionTone = (direction?: string): MeBadgeTone => {
  if (direction === 'in') return 'success'
  if (direction === 'out') return 'danger'
  return 'warning'
}

const transactionTypeLabel = (type?: string) => {
  const key = `personalCenter.wallet.types.${type || ''}`
  const translated = t(key)
  if (translated === key) return type || '-'
  return translated
}

const signedAmount = (direction: string, amount?: string, currency?: string) => {
  const base = formatMoney(amount, currency)
  if (base === '-') return base
  if (direction === 'out') return `-${base}`
  return `+${base}`
}

const loadWallet = async () => {
  const response = await walletAPI.account()
  wallet.value = response.data.data
}

const loadTransactions = async (page = 1) => {
  loading.value = true
  try {
    const response = await walletAPI.transactions({
      page,
      page_size: pagination.value.page_size,
    })
    transactions.value = response.data.data || []
    pagination.value = response.data.pagination || pagination.value
  } catch {
    transactions.value = []
  } finally {
    loading.value = false
  }
}

const handleRecharge = async () => {
  walletAlert.value = null
  const amount = rechargeForm.amount.trim()
  const amountCents = amountToCents(amount)
  if (!amount || amountCents === null || amountCents <= 0) {
    walletAlert.value = {
      level: 'warning',
      message: t('personalCenter.wallet.errors.invalidAmount'),
    }
    return
  }
  if (!Number.isFinite(rechargeForm.channelId) || rechargeForm.channelId <= 0) {
    walletAlert.value = {
      level: 'warning',
      message: t('personalCenter.wallet.errors.channelRequired'),
    }
    return
  }
  if (selectedChannelAmountHint.value) {
    walletAlert.value = {
      level: 'warning',
      message: selectedChannelAmountHint.value,
    }
    return
  }

  recharging.value = true
  try {
    const response = await walletAPI.recharge({
      amount,
      channel_id: rechargeForm.channelId,
      remark: rechargeForm.remark.trim() || undefined,
    })
    const payload = response.data.data || {}
    const rechargeNo = payload?.recharge?.recharge_no || payload?.recharge_no || ''
    rechargeForm.amount = ''
    rechargeForm.remark = ''
    if (rechargeNo) {
      router.push(`/recharge-orders/${encodeURIComponent(rechargeNo)}`)
    } else {
      walletAlert.value = {
        level: 'success',
        message: t('personalCenter.wallet.createPaymentSuccess'),
      }
    }
  } catch (err: any) {
    walletAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.wallet.errors.rechargeFailed'),
    }
  } finally {
    recharging.value = false
  }
}

const changePage = (page: number) => {
  if (page < 1 || page > pagination.value.total_page) return
  loadTransactions(page)
}

const refreshCurrentPage = async () => {
  await Promise.all([
    loadWallet(),
    loadTransactions(pagination.value.page),
  ])
}

// 支付网关回调可能带 recharge_no 回到 /me/wallet，重定向到充值详情页
const redirectRechargeReturn = () => {
  const query = route.query as Record<string, unknown>
  const rechargeNo = String(query.recharge_no || '').trim()
  const orderNo = String(query.order_no || '').trim()
  const targetNo = rechargeNo || (/^WR/i.test(orderNo) ? orderNo : '')
  if (targetNo) {
    router.replace(`/recharge-orders/${encodeURIComponent(targetNo)}${window.location.search}`)
  }
}

const initialize = async () => {
  walletAlert.value = null
  try {
    if (!appStore.config) void appStore.loadConfig()
    await Promise.all([
      loadWallet(),
      loadTransactions(),
    ])
    redirectRechargeReturn()
  } catch (err: any) {
    walletAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.wallet.errors.loadFailed'),
    }
  }
}

watch(() => rechargeForm.amount, () => {
  scheduleLoadPaymentChannels()
}, { immediate: true })

watch(
  channels,
  (list) => {
    if (list.length === 0) {
      rechargeForm.channelId = 0
      return
    }
    if (!list.some((item: any) => item.id === rechargeForm.channelId)) {
      const first = list[0]
      if (!first) {
        rechargeForm.channelId = 0
        return
      }
      rechargeForm.channelId = first.id
    }
  },
  { immediate: true }
)

onMounted(() => {
  void initialize()
})

onUnmounted(() => {
  channelFetchSeq.value += 1
  if (channelFetchTimer.value) {
    window.clearTimeout(channelFetchTimer.value)
    channelFetchTimer.value = null
  }
  channelLoading.value = false
})
</script>

<style scoped>
/* 钱包面板：区块 .qz-panel + .qz-section-head、提示条 .qz-alert、统计格 .qz-stat-grid.is-triple / .qz-stat、表单 .qz-field 三件套、
   方向 .qz-badge、翻页 .qz-icon-btn--sm、骨架 .qz-skeleton 来自 styles/base.css；.qz-me-panel / .qz-me-sec / .qz-me-acts /
   .qz-me-form(-row) / .qz-me-sub / .qz-me-kv(-row) / .qz-me-rows / .qz-me-row-* / .qz-me-pager / .qz-me-sk 来自
   PersonalCenter.vue 的 me 共享块；通道下拉来自 QzSelect。这里只剩无通道的警示提示与交易金额的方向色。
   .qz-me-tx-amount.is-* 写成 (0,3,0)，压过共享块的 .qz-me-row-main b (0,2,1)。 */
.qz-field-hint.is-warn {
  color: var(--qz-warning);
}

.qz-me-tx-amount.is-in {
  color: var(--qz-success);
}

.qz-me-tx-amount.is-out {
  color: var(--qz-danger);
}
</style>
