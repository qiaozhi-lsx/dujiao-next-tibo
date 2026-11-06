<template>
  <div class="qz-me-panel">
    <!-- 标题行 -->
    <div class="qz-me-card-head">
      <div>
        <h2>{{ t('orders.title') }}</h2>
        <small>{{ t('orders.subtitle') }}</small>
      </div>
      <div class="qz-me-card-acts">
        <span class="qz-me-pill is-neutral">
          {{ t('orders.pageInfo', { page: activePagination.page, total: activePagination.total_page }) }}
        </span>
        <RouterLink class="qz-me-act" to="/products">{{ t('orders.continueShopping') }}</RouterLink>
      </div>
    </div>

    <!-- Tab 切换 -->
    <QzSegment
      :items="tabItems"
      :model-value="activeTab"
      :aria-label="t('orders.title')"
      @change="switchTab($event as 'product' | 'recharge')"
    />

    <!-- 普通订单 Tab -->
    <template v-if="activeTab === 'product'">
      <div class="qz-me-stats">
        <div class="qz-me-stat">
          <b>{{ orderPagination.total }}</b>
          <span>{{ t('orders.stats.totalMatched') }}</span>
        </div>
        <div class="qz-me-stat">
          <b>{{ orders.length }}</b>
          <span>{{ t('orders.stats.currentPage') }}</span>
        </div>
        <div class="qz-me-stat">
          <b>{{ pendingPaymentCount }}</b>
          <span>{{ t('orders.stats.pendingPayment') }}</span>
        </div>
        <div class="qz-me-stat">
          <b>{{ finishedCount }}</b>
          <span>{{ t('orders.stats.finished') }}</span>
        </div>
      </div>

      <!-- 筛选 -->
      <div class="qz-me-card">
        <div class="qz-me-toolbar">
          <div class="qz-ic-field">
            <label for="qz-me-order-keyword">{{ t('orders.filters.keyword') }}</label>
            <input
              id="qz-me-order-keyword"
              v-model="orderFilters.orderNo"
              type="text"
              :placeholder="t('orders.filters.orderNoPlaceholder')"
              @input="handleOrderNoInput"
              @keyup.enter="applyOrderFilters"
            />
          </div>
          <div class="qz-ic-field">
            <label for="qz-me-order-status">{{ t('orders.filters.status') }}</label>
            <QzSelect
              id="qz-me-order-status"
              :model-value="orderStatusProxy"
              :items="orderStatusOptions.map((item: any) => ({ value: item.value || 'all', label: item.label }))"
              @update:model-value="orderStatusProxy = $event as string"
            />
          </div>
          <div class="qz-me-toolbar-acts">
            <button type="button" class="qz-me-act is-primary" @click="applyOrderFilters">
              {{ t('orders.filters.search') }}
            </button>
            <button type="button" class="qz-me-act" @click="resetOrderFilters">
              {{ t('orders.filters.reset') }}
            </button>
            <button type="button" class="qz-me-act" @click="refreshOrdersCurrentPage">
              {{ t('orders.filters.refresh') }}
            </button>
          </div>
        </div>
        <p v-if="hasOrderActiveFilters" class="qz-me-hint">
          {{ t('orders.filters.current', { orderNo: orderFilters.orderNo || t('orders.filters.any'), status: currentOrderStatusLabel }) }}
        </p>
      </div>

      <!-- 订单列表 -->
      <div v-if="orderLoading" class="qz-me-rows">
        <div v-for="i in 3" :key="i" class="qz-me-sk" aria-hidden="true"></div>
      </div>
      <div v-else-if="orders.length === 0" class="qz-me-empty">
        <p class="qz-me-hint">{{ hasOrderActiveFilters ? t('orders.emptyFiltered') : t('orders.empty') }}</p>
        <RouterLink class="qz-me-act" to="/products">{{ t('orders.emptyAction') }}</RouterLink>
      </div>
      <div v-else class="qz-me-rows">
        <div v-for="order in orders" :key="order.order_no" class="qz-me-row">
          <div class="qz-me-row-main">
            <b>{{ formatMoney(order.total_amount, order.currency) }}</b>
            <small>{{ t('orders.orderNo') }}：{{ order.order_no }}</small>
            <small>{{ formatDate(order.created_at) }}</small>
            <div v-if="hasDiscount(order)" class="qz-me-discounts">
              <span v-if="hasDiscountAmount(order.discount_amount)" class="qz-me-pill is-ok">
                {{ t('orderDetail.couponDiscountLabel') }}：{{ formatDiscountMoney(order.discount_amount, order.currency) }}
              </span>
              <span v-if="hasDiscountAmount(order.promotion_discount_amount)" class="qz-me-pill is-danger">
                {{ t('orderDetail.promotionDiscountLabel') }}：{{ formatDiscountMoney(order.promotion_discount_amount, order.currency) }}
              </span>
            </div>
          </div>
          <div class="qz-me-row-side">
            <span class="qz-me-pill" :class="`is-${statusPillTone(order.status)}`">
              {{ statusLabel(t, order.status) }}
            </span>
            <RouterLink class="qz-me-act" :to="`/orders/${order.order_no}`">{{ t('orders.viewDetails') }}</RouterLink>
            <RouterLink v-if="order.status === 'pending_payment'" class="qz-me-act is-primary" :to="`/pay?order_no=${order.order_no}`">
              {{ t('orders.payNow') }}
            </RouterLink>
          </div>
        </div>
      </div>

      <!-- 分页（不滚页：仅翻页请求，不做滚动） -->
      <div v-if="orderPagination.total_page > 1" class="qz-me-pager">
        <button
          type="button"
          :disabled="orderPagination.page <= 1 || orderLoading"
          :aria-label="t('pagination.previous')"
          @click="changeOrderPage(orderPagination.page - 1)"
        >
          <ChevronLeft />
        </button>
        <span>{{ orderPagination.page }} / {{ orderPagination.total_page }}</span>
        <button
          type="button"
          :disabled="orderPagination.page >= orderPagination.total_page || orderLoading"
          :aria-label="t('pagination.next')"
          @click="changeOrderPage(orderPagination.page + 1)"
        >
          <ChevronRight />
        </button>
      </div>
    </template>

    <!-- 充值订单 Tab -->
    <template v-if="activeTab === 'recharge'">
      <div class="qz-me-stats is-triple">
        <div class="qz-me-stat">
          <b>{{ rechargePagination.total }}</b>
          <span>{{ t('orders.stats.totalMatched') }}</span>
        </div>
        <div class="qz-me-stat">
          <b>{{ rechargeOrders.length }}</b>
          <span>{{ t('orders.stats.currentPage') }}</span>
        </div>
        <div class="qz-me-stat">
          <b>{{ rechargePendingCount }}</b>
          <span>{{ t('orders.stats.pendingPayment') }}</span>
        </div>
      </div>

      <!-- 筛选 -->
      <div class="qz-me-card">
        <div class="qz-me-toolbar">
          <div class="qz-ic-field">
            <label for="qz-me-recharge-keyword">{{ t('orders.rechargeFilters.keyword') }}</label>
            <input
              id="qz-me-recharge-keyword"
              v-model="rechargeFilters.rechargeNo"
              type="text"
              :placeholder="t('orders.rechargeFilters.rechargeNoPlaceholder')"
              @input="handleRechargeNoInput"
              @keyup.enter="applyRechargeFilters"
            />
          </div>
          <div class="qz-ic-field">
            <label for="qz-me-recharge-status">{{ t('orders.filters.status') }}</label>
            <QzSelect
              id="qz-me-recharge-status"
              :model-value="rechargeStatusProxy"
              :items="rechargeStatusOptions.map((item: any) => ({ value: item.value || 'all', label: item.label }))"
              @update:model-value="rechargeStatusProxy = $event as string"
            />
          </div>
          <div class="qz-me-toolbar-acts">
            <button type="button" class="qz-me-act is-primary" @click="applyRechargeFilters">
              {{ t('orders.filters.search') }}
            </button>
            <button type="button" class="qz-me-act" @click="resetRechargeFilters">
              {{ t('orders.filters.reset') }}
            </button>
            <button type="button" class="qz-me-act" @click="refreshRechargeCurrentPage">
              {{ t('orders.filters.refresh') }}
            </button>
          </div>
        </div>
      </div>

      <!-- 充值订单列表 -->
      <div v-if="rechargeLoading" class="qz-me-rows">
        <div v-for="i in 3" :key="i" class="qz-me-sk" aria-hidden="true"></div>
      </div>
      <div v-else-if="rechargeOrders.length === 0" class="qz-me-empty">
        <p class="qz-me-hint">{{ t('orders.rechargeEmpty') }}</p>
        <RouterLink class="qz-me-act" to="/me/wallet">{{ t('orders.rechargeEmptyAction') }}</RouterLink>
      </div>
      <div v-else class="qz-me-rows">
        <div v-for="ro in rechargeOrders" :key="ro.recharge_no" class="qz-me-row">
          <div class="qz-me-row-main">
            <b>{{ formatMoney(ro.amount, ro.currency) }}</b>
            <small>{{ t('personalCenter.wallet.rechargeNoLabel') }}：{{ ro.recharge_no }}</small>
            <small>{{ formatDate(ro.created_at) }}</small>
          </div>
          <div class="qz-me-row-side">
            <span class="qz-me-pill" :class="`is-${rechargeStatusTone(ro.status)}`">
              {{ rechargeStatusText(ro.status) }}
            </span>
            <RouterLink class="qz-me-act" :to="`/recharge-orders/${ro.recharge_no}`">{{ t('orders.viewDetails') }}</RouterLink>
            <RouterLink v-if="ro.status === 'pending'" class="qz-me-act is-primary" :to="`/recharge-orders/${ro.recharge_no}`">
              {{ t('orders.payNow') }}
            </RouterLink>
          </div>
        </div>
      </div>

      <!-- 分页（不滚页：仅翻页请求，不做滚动） -->
      <div v-if="rechargePagination.total_page > 1" class="qz-me-pager">
        <button
          type="button"
          :disabled="rechargePagination.page <= 1 || rechargeLoading"
          :aria-label="t('pagination.previous')"
          @click="changeRechargePage(rechargePagination.page - 1)"
        >
          <ChevronLeft />
        </button>
        <span>{{ rechargePagination.page }} / {{ rechargePagination.total_page }}</span>
        <button
          type="button"
          :disabled="rechargePagination.page >= rechargePagination.total_page || rechargeLoading"
          :aria-label="t('pagination.next')"
          @click="changeRechargePage(rechargePagination.page + 1)"
        >
          <ChevronRight />
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { userOrderAPI } from '../../../api'
import { walletAPI } from '../../../api/wallet'
import { type BadgeTone } from '../../../utils/status'
import { debounceAsync } from '../../../utils/debounce'
import { amountToCents } from '../../../utils/money'
import QzSegment, { type QzSegmentItem } from '../components/QzSegment.vue'
import QzSelect from '../components/QzSelect.vue'
import { formatDate, formatMoney, statusLabel, statusPillTone, type MePillTone } from './support'

const { t } = useI18n()

// ========== Tab 状态 ==========
const activeTab = ref<'product' | 'recharge'>('product')

const tabItems = computed<QzSegmentItem[]>(() => [
  { key: 'product', label: t('orders.tabs.product') },
  { key: 'recharge', label: t('orders.tabs.recharge') },
])

const switchTab = (tab: 'product' | 'recharge') => {
  if (activeTab.value === tab) return
  activeTab.value = tab
  if (tab === 'product' && !orderLoaded.value) {
    loadOrders(1)
  }
  if (tab === 'recharge' && !rechargeLoaded.value) {
    loadRechargeOrders(1)
  }
}

const activePagination = computed(() =>
  activeTab.value === 'product' ? orderPagination.value : rechargePagination.value,
)

// ========== 普通订单 ==========
const orderLoading = ref(true)
const orderLoaded = ref(false)
const orders = ref<any[]>([])
const orderPagination = ref({ page: 1, page_size: 20, total: 0, total_page: 1 })
const orderFilters = reactive({ orderNo: '', status: '' })
const orderStats = ref<Record<string, number>>({})

const orderStatusOptions = computed(() => [
  { value: '', label: t('orders.filters.statusAll') },
  { value: 'pending_payment', label: t('order.status.pending_payment') },
  { value: 'paid', label: t('order.status.paid') },
  { value: 'fulfilling', label: t('order.status.fulfilling') },
  { value: 'partially_delivered', label: t('order.status.partially_delivered') },
  { value: 'partially_refunded', label: t('order.status.partially_refunded') },
  { value: 'delivered', label: t('order.status.delivered') },
  { value: 'completed', label: t('order.status.completed') },
  { value: 'expired', label: t('order.status.expired') },
  { value: 'canceled', label: t('order.status.canceled') },
  { value: 'refunded', label: t('order.status.refunded') },
])

const hasOrderActiveFilters = computed(() => Boolean(orderFilters.orderNo || orderFilters.status))
const currentOrderStatusLabel = computed(() => {
  const selected = orderStatusOptions.value.find((item) => item.value === orderFilters.status)
  return selected?.label || t('orders.filters.statusAll')
})
const pendingPaymentCount = computed(() => orderStats.value['pending_payment'] || 0)
const finishedCount = computed(
  () =>
    (orderStats.value['delivered'] || 0) +
    (orderStats.value['completed'] || 0) +
    (orderStats.value['partially_refunded'] || 0) +
    (orderStats.value['refunded'] || 0),
)

const loadOrders = async (page = 1) => {
  orderLoading.value = true
  try {
    const response = await userOrderAPI.list({
      page,
      page_size: orderPagination.value.page_size,
      status: orderFilters.status || undefined,
      order_no: orderFilters.orderNo || undefined,
    })
    orders.value = response.data.data || []
    orderPagination.value = response.data.pagination || orderPagination.value
    orderLoaded.value = true
  } catch {
    orders.value = []
  } finally {
    orderLoading.value = false
  }
  loadOrderStats()
}

// 按状态聚合的全量统计（不受分页与状态筛选影响，仅复用关键词筛选）
const loadOrderStats = async () => {
  try {
    const response = await userOrderAPI.stats({ order_no: orderFilters.orderNo || undefined })
    orderStats.value = response.data.data?.by_status || {}
  } catch {
    orderStats.value = {}
  }
}

const debouncedLoadOrders = debounceAsync(loadOrders, 300)

const changeOrderPage = (page: number) => {
  if (page < 1 || page > orderPagination.value.total_page) return
  debouncedLoadOrders(page)
}
const applyOrderFilters = () => loadOrders(1)
const handleOrderNoInput = () => debouncedLoadOrders(1)
const handleOrderStatusChange = () => loadOrders(1)
const orderStatusProxy = computed({
  get: () => orderFilters.status || 'all',
  set: (v: string) => {
    orderFilters.status = v === 'all' ? '' : v
    handleOrderStatusChange()
  },
})
const resetOrderFilters = () => {
  orderFilters.orderNo = ''
  orderFilters.status = ''
  loadOrders(1)
}
const refreshOrdersCurrentPage = () => loadOrders(orderPagination.value.page)

// ========== 充值订单 ==========
const rechargeLoading = ref(false)
const rechargeLoaded = ref(false)
const rechargeOrders = ref<any[]>([])
const rechargePagination = ref({ page: 1, page_size: 20, total: 0, total_page: 1 })
const rechargeFilters = reactive({ rechargeNo: '', status: '' })
const rechargeStats = ref<Record<string, number>>({})

const rechargeStatusOptions = computed(() => [
  { value: '', label: t('orders.filters.statusAll') },
  { value: 'pending', label: t('personalCenter.wallet.rechargeStatus.pending') },
  { value: 'success', label: t('personalCenter.wallet.rechargeStatus.success') },
  { value: 'failed', label: t('personalCenter.wallet.rechargeStatus.failed') },
  { value: 'expired', label: t('personalCenter.wallet.rechargeStatus.expired') },
])

const rechargePendingCount = computed(() => rechargeStats.value['pending'] || 0)

const loadRechargeOrders = async (page = 1) => {
  rechargeLoading.value = true
  try {
    const response = await walletAPI.rechargeOrders({
      page,
      page_size: rechargePagination.value.page_size,
      status: rechargeFilters.status || undefined,
      recharge_no: rechargeFilters.rechargeNo || undefined,
    })
    rechargeOrders.value = response.data.data || []
    rechargePagination.value = response.data.pagination || rechargePagination.value
    rechargeLoaded.value = true
  } catch {
    rechargeOrders.value = []
  } finally {
    rechargeLoading.value = false
  }
  loadRechargeStats()
}

// 按状态聚合的全量统计（不受分页与状态筛选影响，仅复用关键词筛选）
const loadRechargeStats = async () => {
  try {
    const response = await walletAPI.rechargeStats({ recharge_no: rechargeFilters.rechargeNo || undefined })
    rechargeStats.value = response.data.data?.by_status || {}
  } catch {
    rechargeStats.value = {}
  }
}

const debouncedLoadRechargeOrders = debounceAsync(loadRechargeOrders, 300)

const changeRechargePage = (page: number) => {
  if (page < 1 || page > rechargePagination.value.total_page) return
  debouncedLoadRechargeOrders(page)
}
const applyRechargeFilters = () => loadRechargeOrders(1)
const handleRechargeNoInput = () => debouncedLoadRechargeOrders(1)
const handleRechargeStatusChange = () => loadRechargeOrders(1)
const rechargeStatusProxy = computed({
  get: () => rechargeFilters.status || 'all',
  set: (v: string) => {
    rechargeFilters.status = v === 'all' ? '' : v
    handleRechargeStatusChange()
  },
})
const resetRechargeFilters = () => {
  rechargeFilters.rechargeNo = ''
  rechargeFilters.status = ''
  loadRechargeOrders(1)
}
const refreshRechargeCurrentPage = () => loadRechargeOrders(rechargePagination.value.page)

const rechargeStatusText = (status?: string) => {
  const normalized = String(status || '').toLowerCase()
  const key = `personalCenter.wallet.rechargeStatus.${normalized}`
  const translated = t(key)
  if (translated === key) return normalized || '-'
  return translated
}

const rechargeStatusVariant = (status?: string): BadgeTone => {
  const normalized = String(status || '').toLowerCase()
  if (normalized === 'success') return 'success'
  if (normalized === 'failed' || normalized === 'expired') return 'danger'
  return 'warning'
}

// 充值状态 → qz-me-pill tone（success→ok、failed/expired→danger、其余→warn）
const rechargeStatusTone = (status?: string): MePillTone => {
  switch (rechargeStatusVariant(status)) {
    case 'success': return 'ok'
    case 'danger': return 'danger'
    default: return 'warn'
  }
}

// ========== 共用工具 ==========
const formatDiscountMoney = (amount?: string, currency?: string) => {
  return hasDiscountAmount(amount) ? `-${formatMoney(amount, currency)}` : formatMoney(amount, currency)
}

const hasDiscountAmount = (amount?: string) => {
  if (amount === null || amount === undefined || amount === '') return false
  const valueCents = amountToCents(amount)
  return valueCents !== null && valueCents > 0
}

const hasDiscount = (order: any) => {
  if (!order) return false
  return hasDiscountAmount(order.discount_amount) || hasDiscountAmount(order.promotion_discount_amount)
}

onMounted(() => {
  debouncedLoadOrders(1)
})

onUnmounted(() => {
  debouncedLoadOrders.cancel()
  debouncedLoadRechargeOrders.cancel()
})
</script>

<style scoped>
/* 顶部标题行直接作为面板头部：块间距交给 .qz-me-panel 的 gap，抵消卡片语境下的下边距 */
.qz-me-panel > .qz-me-card-head {
  margin-bottom: 0;
}

/* 充值统计只有三格：使用三列，并在 900px 断点与全局两列回退保持一致 */
.qz-me-panel .qz-me-stats.is-triple {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}
@media (max-width: 900px) {
  .qz-me-panel .qz-me-stats.is-triple {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

/* 筛选提示与工具条之间补充间距（.qz-me-hint 自身 margin 为 0） */
.qz-me-panel .qz-me-toolbar + .qz-me-hint {
  margin-top: 10px;
}

/* 空态：说明文案 + 行动链接纵向排列，链接保持胶囊自身宽度 */
.qz-me-empty {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
}
</style>
