<template>
  <div class="qz-me-panel">
    <!-- 开通 / 数据总览 -->
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.affiliate.title') }}</h2>
          <small>{{ t('personalCenter.affiliate.subtitle') }}</small>
        </div>
        <div class="qz-me-acts">
          <span class="qz-chip">
            <Megaphone aria-hidden="true" />
            {{ t('personalCenter.tabs.affiliate') }}
          </span>
        </div>
      </div>

      <p
        v-if="panelAlert"
        class="qz-me-alert"
        :class="alertToneClass(panelAlert.level)"
        :role="panelAlert.level === 'error' ? 'alert' : 'status'"
      >
        {{ panelAlert.message }}
      </p>

      <div v-if="loading" class="qz-me-rows" aria-hidden="true">
        <div v-for="idx in 3" :key="idx" class="qz-me-sk"></div>
      </div>

      <div v-else-if="dashboard?.opened" class="qz-me-opened">
        <div class="qz-me-rows">
          <div class="qz-card qz-me-row">
            <div class="qz-me-row-main">
              <small>{{ t('personalCenter.affiliate.affiliateCode') }}</small>
              <b><code>{{ dashboard?.affiliate_code || '-' }}</code></b>
            </div>
            <div class="qz-me-row-side">
              <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm" @click="copyPromotionUrl">
                {{ t('personalCenter.affiliate.copyPromotionUrl') }}
              </button>
            </div>
          </div>
        </div>
        <p class="qz-field-hint qz-me-url">{{ promotionUrl }}</p>

        <div class="qz-me-stats">
          <div class="qz-card qz-me-stat">
            <b>{{ conversionRateText }}</b>
            <span>{{ t('personalCenter.affiliate.conversionRate') }}</span>
            <small>{{ t('personalCenter.affiliate.conversionDetail', { clicks: dashboard?.click_count || 0, orders: dashboard?.valid_order_count || 0 }) }}</small>
          </div>
          <div class="qz-card qz-me-stat">
            <b>{{ dashboard?.pending_commission || '0.00' }}</b>
            <span>{{ t('personalCenter.affiliate.pendingCommission') }}</span>
          </div>
          <div class="qz-card qz-me-stat">
            <b>{{ dashboard?.available_commission || '0.00' }}</b>
            <span>{{ t('personalCenter.affiliate.availableCommission') }}</span>
          </div>
          <div class="qz-card qz-me-stat">
            <b>{{ dashboard?.withdrawn_commission || '0.00' }}</b>
            <span>{{ t('personalCenter.affiliate.withdrawnCommission') }}</span>
          </div>
        </div>
      </div>

      <div v-else class="qz-me-cta">
        <p class="qz-field-hint">{{ t('personalCenter.affiliate.notOpened') }}</p>
        <QzButton :loading="opening" @click="openAffiliate">
          {{ opening ? t('personalCenter.affiliate.opening') : t('personalCenter.affiliate.openButton') }}
        </QzButton>
      </div>
    </section>

    <!-- 申请提现 -->
    <section v-if="dashboard?.opened" class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.affiliate.withdrawTitle') }}</h2>
          <small>{{ t('personalCenter.affiliate.withdrawSubtitle') }}</small>
        </div>
      </div>

      <form class="qz-me-form" novalidate @submit.prevent="handleApplyWithdraw">
        <div class="qz-field">
          <label class="qz-field-label" for="qz-aff-amount">{{ t('personalCenter.affiliate.withdrawAmountLabel') }}</label>
          <input
            id="qz-aff-amount"
            v-model="withdrawForm.amount"
            class="qz-field-control"
            type="text"
            inputmode="decimal"
            :placeholder="t('personalCenter.affiliate.withdrawAmountPlaceholder')"
          />
        </div>
        <div class="qz-me-form-row">
          <div class="qz-field">
            <label class="qz-field-label" for="qz-aff-channel">{{ t('personalCenter.affiliate.withdrawChannelLabel') }}</label>
            <QzSelect
              v-if="channelOptions.length > 0"
              id="qz-aff-channel"
              :model-value="withdrawForm.channel"
              :items="channelOptions.map((channel: string) => ({ value: channel, label: channel }))"
              :placeholder="t('personalCenter.affiliate.withdrawChannelPlaceholder')"
              @update:model-value="withdrawForm.channel = $event as string"
            />
            <input
              v-else
              id="qz-aff-channel"
              v-model="withdrawForm.channel"
              class="qz-field-control"
              type="text"
              :placeholder="t('personalCenter.affiliate.withdrawChannelPlaceholder')"
            />
          </div>
          <div class="qz-field">
            <label class="qz-field-label" for="qz-aff-account">{{ t('personalCenter.affiliate.withdrawAccountLabel') }}</label>
            <input
              id="qz-aff-account"
              v-model="withdrawForm.account"
              class="qz-field-control"
              type="text"
              :placeholder="t('personalCenter.affiliate.withdrawAccountPlaceholder')"
            />
          </div>
        </div>
        <QzButton type="submit" class="qz-btn--block" :loading="submittingWithdraw" :disabled="submittingWithdraw">
          {{ submittingWithdraw ? t('personalCenter.affiliate.withdrawing') : t('personalCenter.affiliate.withdrawSubmit') }}
        </QzButton>
      </form>
    </section>

    <!-- 佣金记录 -->
    <section v-if="dashboard?.opened" class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div><h2>{{ t('personalCenter.affiliate.commissionTitle') }}</h2></div>
        <div class="qz-me-acts">
          <button
            type="button"
            class="qz-btn qz-btn--secondary qz-btn--sm"
            :disabled="commissionsLoading"
            @click="loadCommissions(commissionsPagination.page)"
          >
            {{ t('orders.filters.refresh') }}
          </button>
        </div>
      </div>

      <div v-if="commissionsLoading" class="qz-me-rows" aria-hidden="true">
        <div v-for="idx in 3" :key="idx" class="qz-me-sk"></div>
      </div>
      <p v-else-if="commissions.length === 0" class="qz-field-hint">
        {{ t('personalCenter.affiliate.commissionEmpty') }}
      </p>
      <div v-else class="qz-me-rows">
        <div v-for="item in commissions" :key="item.id" class="qz-card qz-me-row">
          <div class="qz-me-row-main">
            <b>{{ item.commission_amount }}</b>
            <small>{{ formatDate(item.created_at) }}</small>
          </div>
          <div class="qz-me-row-side">
            <span class="qz-badge" :class="badgeToneClass(commissionTone(item.status))">
              {{ commissionStatusLabel(item.status) }}
            </span>
          </div>
        </div>
      </div>

      <nav v-if="commissionsPagination.total_page > 1" class="qz-me-pager" :aria-label="t('orders.pageInfo', { page: commissionsPagination.page, total: commissionsPagination.total_page })">
        <button
          type="button"
          class="qz-icon-btn qz-icon-btn--sm"
          :disabled="commissionsPagination.page <= 1 || commissionsLoading"
          :aria-label="t('pagination.previous')"
          @click="loadCommissions(commissionsPagination.page - 1)"
        >
          <ChevronLeft />
        </button>
        <span>{{ commissionsPagination.page }} / {{ commissionsPagination.total_page }}</span>
        <button
          type="button"
          class="qz-icon-btn qz-icon-btn--sm"
          :disabled="commissionsPagination.page >= commissionsPagination.total_page || commissionsLoading"
          :aria-label="t('pagination.next')"
          @click="loadCommissions(commissionsPagination.page + 1)"
        >
          <ChevronRight />
        </button>
      </nav>
    </section>

    <!-- 提现记录 -->
    <section v-if="dashboard?.opened" class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div><h2>{{ t('personalCenter.affiliate.withdrawRecordTitle') }}</h2></div>
        <div class="qz-me-acts">
          <button
            type="button"
            class="qz-btn qz-btn--secondary qz-btn--sm"
            :disabled="withdrawsLoading"
            @click="loadWithdraws(withdrawsPagination.page)"
          >
            {{ t('orders.filters.refresh') }}
          </button>
        </div>
      </div>

      <div v-if="withdrawsLoading" class="qz-me-rows" aria-hidden="true">
        <div v-for="idx in 3" :key="idx" class="qz-me-sk"></div>
      </div>
      <p v-else-if="withdraws.length === 0" class="qz-field-hint">
        {{ t('personalCenter.affiliate.withdrawEmpty') }}
      </p>
      <div v-else class="qz-me-rows">
        <div v-for="item in withdraws" :key="item.id" class="qz-card qz-me-row">
          <div class="qz-me-row-main">
            <b>{{ item.amount }}</b>
            <small>{{ item.channel }}</small>
            <small>{{ formatDate(item.created_at) }}</small>
          </div>
          <div class="qz-me-row-side">
            <span class="qz-badge" :class="badgeToneClass(withdrawTone(item.status))">
              {{ withdrawStatusLabel(item.status) }}
            </span>
          </div>
        </div>
      </div>

      <nav v-if="withdrawsPagination.total_page > 1" class="qz-me-pager" :aria-label="t('orders.pageInfo', { page: withdrawsPagination.page, total: withdrawsPagination.total_page })">
        <button
          type="button"
          class="qz-icon-btn qz-icon-btn--sm"
          :disabled="withdrawsPagination.page <= 1 || withdrawsLoading"
          :aria-label="t('pagination.previous')"
          @click="loadWithdraws(withdrawsPagination.page - 1)"
        >
          <ChevronLeft />
        </button>
        <span>{{ withdrawsPagination.page }} / {{ withdrawsPagination.total_page }}</span>
        <button
          type="button"
          class="qz-icon-btn qz-icon-btn--sm"
          :disabled="withdrawsPagination.page >= withdrawsPagination.total_page || withdrawsLoading"
          :aria-label="t('pagination.next')"
          @click="loadWithdraws(withdrawsPagination.page + 1)"
        >
          <ChevronRight />
        </button>
      </nav>
    </section>
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
import QzButton from '../components/QzButton.vue'
import QzSelect from '../components/QzSelect.vue'
import { badgeToneClass, formatDate, type MeBadgeTone } from './support'

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

/** 提示条 tone：success → 成功色、warning → 警示色（默认）、error → 错误色 */
const alertToneClass = (level: PageAlert['level']) => {
  if (level === 'success') return 'is-success'
  if (level === 'error') return 'is-error'
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

const commissionTone = (status?: string): MeBadgeTone => {
  if (status === AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM) return 'warning'
  if (status === AFFILIATE_COMMISSION_STATUS_AVAILABLE) return 'success'
  if (status === AFFILIATE_COMMISSION_STATUS_WITHDRAWN) return 'info'
  return 'neutral'
}

const withdrawStatusLabel = (status?: string) => {
  if (status === AFFILIATE_WITHDRAW_STATUS_PENDING_REVIEW) return t('personalCenter.affiliate.withdrawStatus.pendingReview')
  if (status === AFFILIATE_WITHDRAW_STATUS_REJECTED) return t('personalCenter.affiliate.withdrawStatus.rejected')
  if (status === AFFILIATE_WITHDRAW_STATUS_PAID) return t('personalCenter.affiliate.withdrawStatus.paid')
  return status || '-'
}

const withdrawTone = (status?: string): MeBadgeTone => {
  if (status === AFFILIATE_WITHDRAW_STATUS_PENDING_REVIEW) return 'warning'
  if (status === AFFILIATE_WITHDRAW_STATUS_PAID) return 'success'
  return 'neutral'
}

onMounted(() => {
  initialize()
})
</script>

<style scoped>
/* 分销面板：区块 .qz-panel + .qz-section-head、统计卡 / 记录行 .qz-card、表单 .qz-field 三件套、
   状态 .qz-badge、翻页 .qz-icon-btn--sm、按钮 .qz-btn / QzButton（styles/base.css）；渠道下拉来自 QzSelect。 */
.qz-me-panel {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
}

/* ---------- 区块面板 ---------- */
.qz-panel.qz-me-sec {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
}

.qz-me-sec > .qz-section-head {
  align-items: center;
  margin-bottom: 0;
}

.qz-section-head small {
  display: block;
  margin-top: var(--qz-space-1);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-me-acts {
  display: flex;
  flex: none;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--qz-space-2);
}

.qz-chip svg {
  flex: none;
  width: 12px;
  height: 12px;
}

.qz-me-sec .qz-field-hint {
  margin: 0;
}

/* ---------- 提示条：默认警示色，is-error 错误色，is-success 成功色 ---------- */
.qz-me-alert {
  margin: 0;
  padding: var(--qz-space-3) var(--qz-space-4);
  border-radius: var(--qz-radius-control);
  background: color-mix(in oklab, var(--qz-warning) 12%, transparent);
  color: var(--qz-warning);
  font-size: var(--qz-text-sm);
  font-weight: 500;
  line-height: var(--qz-leading-body);
  overflow-wrap: anywhere;
}

.qz-me-alert.is-error {
  background: color-mix(in oklab, var(--qz-danger) 12%, transparent);
  color: var(--qz-danger);
}

.qz-me-alert.is-success {
  background: color-mix(in oklab, var(--qz-success) 12%, transparent);
  color: var(--qz-success);
}

/* ---------- 已开通：推广码 + 链接 + 统计 ---------- */
.qz-me-opened {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-3);
}

.qz-me-url {
  overflow-wrap: anywhere;
}

.qz-me-cta {
  display: grid;
  gap: var(--qz-space-3);
  justify-items: start;
}

/* ---------- 统计卡 ---------- */
.qz-me-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--qz-space-3);
}

.qz-card.qz-me-stat {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-1);
}

.qz-me-stat b {
  display: flex;
  align-items: center;
  gap: var(--qz-space-2);
  min-width: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.qz-me-stat > span {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  font-weight: 500;
  line-height: var(--qz-leading-body);
}

.qz-me-stat > small {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
  line-height: var(--qz-leading-body);
}

/* ---------- 记录行 ---------- */
.qz-me-rows {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-2);
}

.qz-card.qz-me-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2) var(--qz-space-4);
}

.qz-me-row-main {
  display: flex;
  flex: 1 1 220px;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.qz-me-row-main b {
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.qz-me-row-main small {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.qz-me-row-main code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: var(--qz-text-sm);
  letter-spacing: 0.04em;
}

.qz-me-row-side {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2);
  margin-left: auto;
}

/* ---------- 提现表单 ---------- */
.qz-me-form {
  display: grid;
  gap: var(--qz-space-4);
}

.qz-me-form-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--qz-space-4);
}

/* ---------- 分页 ---------- */
.qz-me-pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--qz-space-3);
  padding-top: var(--qz-space-1);
}

.qz-me-pager span {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

/* ---------- 骨架 ---------- */
.qz-me-sk {
  height: 64px;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  animation: qz-me-pulse 1.4s ease-in-out infinite;
}

@keyframes qz-me-pulse {
  50% {
    opacity: 0.55;
  }
}

@media (max-width: 900px) {
  .qz-me-stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .qz-panel.qz-me-sec {
    padding: var(--qz-space-4);
  }

  .qz-me-form-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .qz-me-row-side {
    flex: 1 1 100%;
    justify-content: flex-end;
    margin-left: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .qz-me-sk {
    animation: none;
  }
}
</style>
