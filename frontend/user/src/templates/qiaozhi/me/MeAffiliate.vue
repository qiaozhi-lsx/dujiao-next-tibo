<template>
  <div class="j8-me-panel">
    <!-- 开通 / 数据总览 -->
    <div class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ t('personalCenter.affiliate.title') }}</h2>
          <small>{{ t('personalCenter.affiliate.subtitle') }}</small>
        </div>
        <div class="j8-me-card-acts">
          <span class="j8-me-pill is-ok">
            <Megaphone aria-hidden="true" />
            {{ t('personalCenter.tabs.affiliate') }}
          </span>
        </div>
      </div>

      <p
        v-if="panelAlert"
        class="j8-me-alert"
        :class="alertToneClass(panelAlert.level)"
        :role="panelAlert.level === 'error' ? 'alert' : 'status'"
      >
        {{ panelAlert.message }}
      </p>

      <div v-if="loading" class="j8-me-rows" aria-hidden="true">
        <div v-for="idx in 3" :key="idx" class="j8-me-sk"></div>
      </div>

      <div v-else-if="dashboard?.opened" class="j8-me-opened">
        <div class="j8-me-rows">
          <div class="j8-me-row">
            <div class="j8-me-row-main">
              <small>{{ t('personalCenter.affiliate.affiliateCode') }}</small>
              <b><code>{{ dashboard?.affiliate_code || '-' }}</code></b>
            </div>
            <div class="j8-me-row-side">
              <button type="button" class="j8-me-act" @click="copyPromotionUrl">
                {{ t('personalCenter.affiliate.copyPromotionUrl') }}
              </button>
            </div>
          </div>
        </div>
        <p class="j8-me-hint j8-me-url">{{ promotionUrl }}</p>

        <div class="j8-me-stats">
          <div class="j8-me-stat">
            <b>{{ conversionRateText }}</b>
            <span>{{ t('personalCenter.affiliate.conversionRate') }}</span>
            <small>{{ t('personalCenter.affiliate.conversionDetail', { clicks: dashboard?.click_count || 0, orders: dashboard?.valid_order_count || 0 }) }}</small>
          </div>
          <div class="j8-me-stat">
            <b>{{ dashboard?.pending_commission || '0.00' }}</b>
            <span>{{ t('personalCenter.affiliate.pendingCommission') }}</span>
          </div>
          <div class="j8-me-stat">
            <b>{{ dashboard?.available_commission || '0.00' }}</b>
            <span>{{ t('personalCenter.affiliate.availableCommission') }}</span>
          </div>
          <div class="j8-me-stat">
            <b>{{ dashboard?.withdrawn_commission || '0.00' }}</b>
            <span>{{ t('personalCenter.affiliate.withdrawnCommission') }}</span>
          </div>
        </div>
      </div>

      <div v-else class="j8-me-cta">
        <p class="j8-me-hint">{{ t('personalCenter.affiliate.notOpened') }}</p>
        <J8Button :loading="opening" @click="openAffiliate">
          {{ opening ? t('personalCenter.affiliate.opening') : t('personalCenter.affiliate.openButton') }}
        </J8Button>
      </div>
    </div>

    <!-- 申请提现 -->
    <div v-if="dashboard?.opened" class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ t('personalCenter.affiliate.withdrawTitle') }}</h2>
          <small>{{ t('personalCenter.affiliate.withdrawSubtitle') }}</small>
        </div>
      </div>

      <form class="j8-ic" novalidate @submit.prevent="handleApplyWithdraw">
        <div class="j8-ic-field">
          <label for="j8-aff-amount">{{ t('personalCenter.affiliate.withdrawAmountLabel') }}</label>
          <input
            id="j8-aff-amount"
            v-model="withdrawForm.amount"
            type="text"
            inputmode="decimal"
            :placeholder="t('personalCenter.affiliate.withdrawAmountPlaceholder')"
          />
        </div>
        <div class="j8-ic-row">
          <div class="j8-ic-field">
            <label for="j8-aff-channel">{{ t('personalCenter.affiliate.withdrawChannelLabel') }}</label>
            <select v-if="channelOptions.length > 0" id="j8-aff-channel" v-model="withdrawForm.channel">
              <option value="" disabled>{{ t('personalCenter.affiliate.withdrawChannelPlaceholder') }}</option>
              <option v-for="channel in channelOptions" :key="channel" :value="channel">{{ channel }}</option>
            </select>
            <input
              v-else
              id="j8-aff-channel"
              v-model="withdrawForm.channel"
              type="text"
              :placeholder="t('personalCenter.affiliate.withdrawChannelPlaceholder')"
            />
          </div>
          <div class="j8-ic-field">
            <label for="j8-aff-account">{{ t('personalCenter.affiliate.withdrawAccountLabel') }}</label>
            <input
              id="j8-aff-account"
              v-model="withdrawForm.account"
              type="text"
              :placeholder="t('personalCenter.affiliate.withdrawAccountPlaceholder')"
            />
          </div>
        </div>
        <J8Button type="submit" class="j8-arc-block" :loading="submittingWithdraw" :disabled="submittingWithdraw">
          {{ submittingWithdraw ? t('personalCenter.affiliate.withdrawing') : t('personalCenter.affiliate.withdrawSubmit') }}
        </J8Button>
      </form>
    </div>

    <!-- 佣金记录 -->
    <div v-if="dashboard?.opened" class="j8-me-card">
      <div class="j8-me-card-head">
        <div><h2>{{ t('personalCenter.affiliate.commissionTitle') }}</h2></div>
        <div class="j8-me-card-acts">
          <button
            type="button"
            class="j8-me-act"
            :disabled="commissionsLoading"
            @click="loadCommissions(commissionsPagination.page)"
          >
            {{ t('orders.filters.refresh') }}
          </button>
        </div>
      </div>

      <div v-if="commissionsLoading" class="j8-me-rows" aria-hidden="true">
        <div v-for="idx in 3" :key="idx" class="j8-me-sk"></div>
      </div>
      <p v-else-if="commissions.length === 0" class="j8-me-hint">
        {{ t('personalCenter.affiliate.commissionEmpty') }}
      </p>
      <div v-else class="j8-me-rows">
        <div v-for="item in commissions" :key="item.id" class="j8-me-row">
          <div class="j8-me-row-main">
            <b>{{ item.commission_amount }}</b>
            <small>{{ formatDate(item.created_at) }}</small>
          </div>
          <div class="j8-me-row-side">
            <span class="j8-me-pill" :class="`is-${commissionTone(item.status)}`">
              {{ commissionStatusLabel(item.status) }}
            </span>
          </div>
        </div>
      </div>

      <div v-if="commissionsPagination.total_page > 1" class="j8-me-pager">
        <button
          type="button"
          :disabled="commissionsPagination.page <= 1 || commissionsLoading"
          :aria-label="t('pagination.previous')"
          @click="loadCommissions(commissionsPagination.page - 1)"
        >
          <ChevronLeft />
        </button>
        <span>{{ commissionsPagination.page }} / {{ commissionsPagination.total_page }}</span>
        <button
          type="button"
          :disabled="commissionsPagination.page >= commissionsPagination.total_page || commissionsLoading"
          :aria-label="t('pagination.next')"
          @click="loadCommissions(commissionsPagination.page + 1)"
        >
          <ChevronRight />
        </button>
      </div>
    </div>

    <!-- 提现记录 -->
    <div v-if="dashboard?.opened" class="j8-me-card">
      <div class="j8-me-card-head">
        <div><h2>{{ t('personalCenter.affiliate.withdrawRecordTitle') }}</h2></div>
        <div class="j8-me-card-acts">
          <button
            type="button"
            class="j8-me-act"
            :disabled="withdrawsLoading"
            @click="loadWithdraws(withdrawsPagination.page)"
          >
            {{ t('orders.filters.refresh') }}
          </button>
        </div>
      </div>

      <div v-if="withdrawsLoading" class="j8-me-rows" aria-hidden="true">
        <div v-for="idx in 3" :key="idx" class="j8-me-sk"></div>
      </div>
      <p v-else-if="withdraws.length === 0" class="j8-me-hint">
        {{ t('personalCenter.affiliate.withdrawEmpty') }}
      </p>
      <div v-else class="j8-me-rows">
        <div v-for="item in withdraws" :key="item.id" class="j8-me-row">
          <div class="j8-me-row-main">
            <b>{{ item.amount }}</b>
            <small>{{ item.channel }}</small>
            <small>{{ formatDate(item.created_at) }}</small>
          </div>
          <div class="j8-me-row-side">
            <span class="j8-me-pill" :class="`is-${withdrawTone(item.status)}`">
              {{ withdrawStatusLabel(item.status) }}
            </span>
          </div>
        </div>
      </div>

      <div v-if="withdrawsPagination.total_page > 1" class="j8-me-pager">
        <button
          type="button"
          :disabled="withdrawsPagination.page <= 1 || withdrawsLoading"
          :aria-label="t('pagination.previous')"
          @click="loadWithdraws(withdrawsPagination.page - 1)"
        >
          <ChevronLeft />
        </button>
        <span>{{ withdrawsPagination.page }} / {{ withdrawsPagination.total_page }}</span>
        <button
          type="button"
          :disabled="withdrawsPagination.page >= withdrawsPagination.total_page || withdrawsLoading"
          :aria-label="t('pagination.next')"
          @click="loadWithdraws(withdrawsPagination.page + 1)"
        >
          <ChevronRight />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, ChevronRight, Megaphone } from 'lucide-vue-next'
import { affiliateAPI, type AffiliateCommissionData, type AffiliateDashboardData, type AffiliateWithdrawData } from '../../../api'
import {
  AFFILIATE_COMMISSION_STATUS_AVAILABLE,
  AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM,
  AFFILIATE_COMMISSION_STATUS_REJECTED,
  AFFILIATE_COMMISSION_STATUS_WITHDRAWN,
  AFFILIATE_WITHDRAW_STATUS_PAID,
  AFFILIATE_WITHDRAW_STATUS_PENDING_REVIEW,
  AFFILIATE_WITHDRAW_STATUS_REJECTED,
} from '../../../constants/affiliate'
import { useAppStore } from '../../../stores/app'
import type { PageAlert } from '../../../utils/alerts'
import J8Button from '../components/J8Button.vue'
import { formatDate, type MePillTone } from './support'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const opening = ref(false)
const submittingWithdraw = ref(false)
const commissionsLoading = ref(false)
const withdrawsLoading = ref(false)
const dashboard = ref<AffiliateDashboardData | null>(null)
const panelAlert = ref<PageAlert | null>(null)

const commissions = ref<AffiliateCommissionData[]>([])
const withdraws = ref<AffiliateWithdrawData[]>([])

const commissionsPagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})

const withdrawsPagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})

const withdrawForm = reactive({
  amount: '',
  channel: '',
  account: '',
})

const channelOptions = computed(() => {
  const channels = appStore.config?.affiliate?.withdraw_channels
  if (!Array.isArray(channels)) return []
  return channels.map((item: unknown) => String(item || '').trim()).filter(Boolean)
})

const promotionUrl = computed(() => {
  if (!dashboard.value?.affiliate_code) return '-'
  const path = dashboard.value.promotion_path || `/?aff=${dashboard.value.affiliate_code}`
  const origin = typeof window !== 'undefined' ? window.location.origin : ''
  return `${origin}${path}`
})

const conversionRateText = computed(() => {
  const value = Number(dashboard.value?.conversion_rate || 0)
  if (!Number.isFinite(value)) return '0.00%'
  return `${value.toFixed(2)}%`
})

const alertToneClass = (level: PageAlert['level']) => {
  if (level === 'success') return 'is-ok'
  if (level === 'warning') return 'is-warn'
  return ''
}

const loadDashboard = async () => {
  try {
    const response = await affiliateAPI.dashboard()
    dashboard.value = response.data.data || null
  } catch (err: any) {
    panelAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.affiliate.errors.loadFailed'),
    }
  }
}

const loadCommissions = async (page = 1) => {
  commissionsLoading.value = true
  try {
    const response = await affiliateAPI.commissions({
      page,
      page_size: commissionsPagination.page_size,
    })
    commissions.value = response.data.data || []
    Object.assign(commissionsPagination, response.data.pagination || commissionsPagination)
  } catch {
    commissions.value = []
  } finally {
    commissionsLoading.value = false
  }
}

const loadWithdraws = async (page = 1) => {
  withdrawsLoading.value = true
  try {
    const response = await affiliateAPI.withdraws({
      page,
      page_size: withdrawsPagination.page_size,
    })
    withdraws.value = response.data.data || []
    Object.assign(withdrawsPagination, response.data.pagination || withdrawsPagination)
  } catch {
    withdraws.value = []
  } finally {
    withdrawsLoading.value = false
  }
}

const reloadOpenedData = async () => {
  if (!dashboard.value?.opened) return
  await Promise.all([loadCommissions(1), loadWithdraws(1)])
}

const initialize = async () => {
  loading.value = true
  panelAlert.value = null
  if (!appStore.config) void appStore.loadConfig()
  await loadDashboard()
  await reloadOpenedData()
  loading.value = false
}

const openAffiliate = async () => {
  opening.value = true
  panelAlert.value = null
  try {
    await affiliateAPI.open()
    await loadDashboard()
    await reloadOpenedData()
    panelAlert.value = {
      level: 'success',
      message: t('personalCenter.affiliate.openSuccess'),
    }
  } catch (err: any) {
    panelAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.affiliate.errors.openFailed'),
    }
  } finally {
    opening.value = false
  }
}

const handleApplyWithdraw = async () => {
  panelAlert.value = null
  if (!withdrawForm.amount.trim()) {
    panelAlert.value = {
      level: 'warning',
      message: t('personalCenter.affiliate.errors.withdrawAmountRequired'),
    }
    return
  }
  if (!withdrawForm.channel.trim()) {
    panelAlert.value = {
      level: 'warning',
      message: t('personalCenter.affiliate.errors.withdrawChannelRequired'),
    }
    return
  }
  if (!withdrawForm.account.trim()) {
    panelAlert.value = {
      level: 'warning',
      message: t('personalCenter.affiliate.errors.withdrawAccountRequired'),
    }
    return
  }

  submittingWithdraw.value = true
  try {
    await affiliateAPI.applyWithdraw({
      amount: withdrawForm.amount.trim(),
      channel: withdrawForm.channel.trim(),
      account: withdrawForm.account.trim(),
    })
    withdrawForm.amount = ''
    withdrawForm.account = ''
    panelAlert.value = {
      level: 'success',
      message: t('personalCenter.affiliate.withdrawSuccess'),
    }
    await Promise.all([loadDashboard(), loadWithdraws(1)])
  } catch (err: any) {
    panelAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.affiliate.errors.withdrawFailed'),
    }
  } finally {
    submittingWithdraw.value = false
  }
}

const copyPromotionUrl = async () => {
  if (!dashboard.value?.affiliate_code || !promotionUrl.value || promotionUrl.value === '-') return
  try {
    await navigator.clipboard.writeText(promotionUrl.value)
    panelAlert.value = {
      level: 'success',
      message: t('personalCenter.affiliate.copySuccess'),
    }
  } catch {
    panelAlert.value = {
      level: 'error',
      message: t('personalCenter.affiliate.errors.copyFailed'),
    }
  }
}

const commissionStatusLabel = (status?: string) => {
  if (status === AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM) return t('personalCenter.affiliate.commissionStatus.pendingConfirm')
  if (status === AFFILIATE_COMMISSION_STATUS_AVAILABLE) return t('personalCenter.affiliate.commissionStatus.available')
  if (status === AFFILIATE_COMMISSION_STATUS_REJECTED) return t('personalCenter.affiliate.commissionStatus.rejected')
  if (status === AFFILIATE_COMMISSION_STATUS_WITHDRAWN) return t('personalCenter.affiliate.commissionStatus.withdrawn')
  return status || '-'
}

const commissionTone = (status?: string): MePillTone => {
  if (status === AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM) return 'warn'
  if (status === AFFILIATE_COMMISSION_STATUS_AVAILABLE) return 'ok'
  if (status === AFFILIATE_COMMISSION_STATUS_WITHDRAWN) return 'info'
  return 'neutral'
}

const withdrawStatusLabel = (status?: string) => {
  if (status === AFFILIATE_WITHDRAW_STATUS_PENDING_REVIEW) return t('personalCenter.affiliate.withdrawStatus.pendingReview')
  if (status === AFFILIATE_WITHDRAW_STATUS_REJECTED) return t('personalCenter.affiliate.withdrawStatus.rejected')
  if (status === AFFILIATE_WITHDRAW_STATUS_PAID) return t('personalCenter.affiliate.withdrawStatus.paid')
  return status || '-'
}

const withdrawTone = (status?: string): MePillTone => {
  if (status === AFFILIATE_WITHDRAW_STATUS_PENDING_REVIEW) return 'warn'
  if (status === AFFILIATE_WITHDRAW_STATUS_PAID) return 'ok'
  return 'neutral'
}

onMounted(() => {
  initialize()
})
</script>

<style scoped>
.j8-me-opened {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.j8-me-cta {
  display: grid;
  gap: 12px;
  justify-items: start;
}

.j8-me-card .j8-me-alert {
  margin-bottom: 12px;
}

.j8-me-alert.is-ok {
  background: var(--accent-soft);
  color: var(--accent-strong);
}

.j8-me-alert.is-warn {
  background: var(--amber-soft);
  color: var(--amber);
}

.j8-me-url {
  overflow-wrap: anywhere;
}

.j8-me-stat small {
  font-size: 11px;
  font-weight: 650;
  color: var(--muted);
}

.j8-me-row-main code {
  font-family: var(--font-mono, ui-monospace, SFMono-Regular, Menlo, Consolas, monospace);
  font-size: 12.5px;
}

.j8-me-card .j8-me-pager {
  margin-top: 10px;
}
</style>
