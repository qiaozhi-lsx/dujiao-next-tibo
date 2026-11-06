<template>
  <div ref="rootEl" class="qz-container qz-orders">
    <!-- 页头：eyebrow + h1 + 说明 + 返回 -->
    <header class="qz-orders-head qz-reveal is-pending">
      <div class="qz-orders-head-copy">
        <span class="qz-eyebrow">{{ t('qiaozhi.guestOrders.eyebrow') }}</span>
        <h1>{{ t('guestOrders.title') }}</h1>
        <p>{{ t('guestOrders.subtitle') }}</p>
      </div>
      <RouterLink to="/" class="qz-btn qz-btn--secondary qz-btn--sm">
        <ChevronLeft />
        {{ t('auth.login.backHome') }}
      </RouterLink>
    </header>

    <!-- 查询表单 -->
    <section class="qz-panel qz-orders-search qz-reveal is-pending">
      <div v-if="hasSavedAuth" class="qz-orders-saved">
        <span>{{ t('guestOrders.savedHint', { email: savedAuth.email || '-' }) }}</span>
        <button type="button" class="qz-btn qz-btn--ghost qz-btn--sm" @click="clearSavedAuth">
          {{ t('guestOrders.clearSaved') }}
        </button>
      </div>
      <form class="qz-orders-form" novalidate @submit.prevent="submitSearch">
        <div class="qz-field">
          <label class="qz-field-label" for="qz-orders-email">{{ t('guestOrders.emailPlaceholder') }}<b>*</b></label>
          <input
            id="qz-orders-email"
            v-model="email"
            class="qz-field-control"
            type="email"
            autocomplete="email"
            required
            :aria-describedby="error ? 'qz-orders-error' : undefined"
          />
        </div>
        <div class="qz-field">
          <label class="qz-field-label" for="qz-orders-password">{{ t('guestOrders.passwordPlaceholder') }}<b>*</b></label>
          <input
            id="qz-orders-password"
            v-model="orderPassword"
            class="qz-field-control"
            type="password"
            autocomplete="current-password"
            required
            :aria-describedby="error ? 'qz-orders-error' : undefined"
          />
        </div>
        <div class="qz-field">
          <label class="qz-field-label" for="qz-orders-no">{{ t('guestOrders.orderNoPlaceholder') }}</label>
          <input
            id="qz-orders-no"
            v-model="orderNo"
            class="qz-field-control"
            type="text"
            autocomplete="off"
            :aria-describedby="error ? 'qz-orders-error' : undefined"
          />
        </div>
        <QzButton type="submit" variant="primary" :loading="loading" class="qz-orders-submit">
          <Search v-if="!loading" />
          {{ loading ? t('guestOrders.searching') : t('guestOrders.search') }}
        </QzButton>
      </form>
      <p class="qz-field-hint qz-orders-tip">{{ t('guestOrders.tip') }}</p>
      <p v-if="error" id="qz-orders-error" class="qz-field-error" role="alert">{{ error }}</p>
    </section>

    <!-- 空态：未查询前给引导，查询后给结果提示 -->
    <div v-if="orders.length === 0 && !loading" class="qz-empty qz-reveal is-pending">
      <PackageOpen />
      <b>{{ emptyText }}</b>
    </div>

    <!-- 订单列表 -->
    <div v-else class="qz-orders-list">
      <article
        v-for="row in rows"
        :key="row.order.order_no"
        class="qz-card qz-orders-item qz-reveal is-pending"
      >
        <div v-if="row.product" class="qz-orders-prod">
          <span class="qz-orders-thumb">
            <img
              v-if="row.product.image"
              :src="row.product.image"
              :alt="row.product.title"
              loading="lazy"
              decoding="async"
            />
            <template v-else>{{ row.product.initial }}</template>
          </span>
          <b class="qz-orders-pname">{{ row.product.title }}</b>
          <span v-if="row.product.extraCount" class="qz-orders-pextra">
            {{ t('guestOrders.moreItems', { count: row.product.extraCount }) }}
          </span>
          <span v-else-if="row.product.quantity > 1" class="qz-orders-pextra">× {{ row.product.quantity }}</span>
        </div>
        <div class="qz-orders-main">
          <b class="qz-orders-sn">{{ row.order.order_no }}</b>
          <span class="qz-orders-time">{{ formatDate(row.order.created_at) }}</span>
        </div>
        <div class="qz-orders-amt">
          <b>{{ formatMoney(row.order.total_amount, row.order.currency) }}</b>
          <small v-if="discountText(row.order)">{{ discountText(row.order) }}</small>
        </div>
        <div class="qz-orders-acts">
          <span class="qz-badge" :class="toneClass(row.order.status)">{{ statusLabel(row.order.status) }}</span>
          <RouterLink
            class="qz-btn qz-btn--secondary qz-btn--sm"
            :to="{ name: 'guest-order-detail', params: { order_no: row.order.order_no } }"
          >
            {{ t('guestOrders.viewDetails') }}
          </RouterLink>
          <RouterLink
            v-if="row.order.status === 'pending_payment'"
            class="qz-btn qz-btn--primary qz-btn--sm"
            :to="`/pay?guest=1&order_no=${encodeURIComponent(row.order.order_no)}`"
          >
            {{ t('guestOrders.payNow') }}
          </RouterLink>
        </div>
      </article>
    </div>

    <!-- 分页 -->
    <nav v-if="pagination.total_page > 1" class="qz-orders-pager" :aria-label="t('guestOrders.pageInfo', { page: pagination.page, total: pagination.total_page })">
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
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, ChevronRight, PackageOpen, Search } from 'lucide-vue-next'
import { useAppStore } from '../../stores/app'
import { useGuestOrders } from '../../composables/useGuestOrders'
import { setupReveal } from './utils/reveal'
import { getImageUrl } from '../../utils/image'
import { getLocalizedText } from '../../utils/resellerSiteConfig'
import QzButton from './components/QzButton.vue'

const { t } = useI18n()
const appStore = useAppStore()

const {
  savedAuth, email, orderPassword, orderNo, loading, error, orders, pagination,
  hasSavedAuth, clearSaved, handleSearch, emptyMessage, changePage,
  statusLabel, statusVariant, formatMoney, formatDiscountMoney, hasDiscountAmount, formatDate,
} = useGuestOrders()

// 提交钮在 loading 时只是 aria-disabled（保留焦点），回车隐式提交要在这里拦住，避免重复查询
const submitSearch = () => {
  if (loading.value) return
  void handleSearch()
}

// 未发起过查询时空态给引导文案；首次进入 loading（手动查询或带已存凭据自动查询）后切回结果提示
const searched = ref(false)
watch(loading, (value) => {
  if (value) searched.value = true
})
const clearSavedAuth = () => {
  clearSaved()
  searched.value = false
}
const emptyText = computed(() => (searched.value ? emptyMessage.value : t('qiaozhi.guestOrders.emptyBeforeQuery')))

const toneClass = (status?: string | null) => {
  switch (statusVariant(status || '')) {
    case 'success': return 'is-success'
    case 'warning': return 'is-warning'
    case 'info': return 'is-info'
    case 'danger': return 'is-danger'
    case 'accent': return 'is-accent'
    default: return ''
  }
}

const discountText = (order: Record<string, any>) => {
  const coupon = order?.discount_amount
  const promotion = order?.promotion_discount_amount
  if (!hasDiscountAmount(coupon) && !hasDiscountAmount(promotion)) return ''
  const total = (Number(coupon) || 0) + (Number(promotion) || 0)
  if (total <= 0) return ''
  return formatDiscountMoney(total.toFixed(2), order?.currency)
}

interface OrderProductBrief {
  image: string
  initial: string
  title: string
  quantity: number
  extraCount: number
}

/** 列表行商品摘要：缩略图 + 首个商品名 + 数量 / 其余件数提示。 */
const resolveOrderProduct = (order: Record<string, any>): OrderProductBrief | null => {
  const items = Array.isArray(order?.items) ? order.items : []
  const first = items[0]
  if (!first) return null
  const title = getLocalizedText(first.title, appStore.locale)
  return {
    image: getImageUrl(String(first?.sku_snapshot?.image || '').trim()),
    initial: title.slice(0, 1).toUpperCase(),
    title,
    quantity: Number(first.quantity) || 1,
    extraCount: items.length > 1 ? items.length - 1 : 0,
  }
}

const rows = computed(() => orders.value.map((order: Record<string, any>) => ({
  order,
  product: resolveOrderProduct(order),
})))

// 入场动画
const rootEl = ref<HTMLElement | null>(null)
let cleanupReveal: (() => void) | null = null
const armReveal = () => {
  cleanupReveal?.()
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
}
// 列表为异步加载：数据到达/翻页/清空后元素才渲染出来（带 is-pending），
// 需重新 arm，否则新节点永远停留在 opacity:0（表现为"查不到订单"）。
watch([loading, () => orders.value.length], () => { armReveal() })
onMounted(() => { armReveal() })
onUnmounted(() => { cleanupReveal?.() })
</script>

<style scoped>
.qz-orders {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  max-width: 920px;
}

/* ---------- 页头 ---------- */
.qz-orders-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--qz-space-4);
}

.qz-orders-head-copy {
  min-width: 0;
}

.qz-orders-head h1 {
  margin: var(--qz-space-1) 0 0;
  color: var(--qz-text);
  font-size: var(--qz-text-2xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
}

.qz-orders-head p {
  margin: var(--qz-space-2) 0 0;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
}

.qz-orders-head > .qz-btn {
  flex: none;
}

/* ---------- 查询表单 ----------
   .qz-panel.qz-orders-search 写成 (0,3,0)：legacy.css 里旧分类侧栏的 .qz-panel 规则（sticky、≤760 隐藏、overflow hidden）
   在它被删掉之前不能影响这块面板。 */
.qz-panel.qz-orders-search {
  position: static;
  display: block;
  overflow: visible;
}

.qz-orders-saved {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--qz-space-2) var(--qz-space-3);
  margin-bottom: var(--qz-space-4);
  padding: var(--qz-space-1) var(--qz-space-1) var(--qz-space-1) var(--qz-space-4);
  border: 1px dashed var(--qz-border-strong);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-orders-saved > span {
  min-width: 0;
  overflow-wrap: anywhere;
}

.qz-orders-form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1fr) auto;
  gap: var(--qz-space-3);
  align-items: end;
}

.qz-orders-tip {
  margin: var(--qz-space-3) 0 0;
}

.qz-orders-search .qz-field-error {
  margin: var(--qz-space-2) 0 0;
}

/* ---------- 订单列表 ---------- */
.qz-orders-list {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-2);
}

/* .qz-card.qz-orders-item 写成 (0,3,0)：legacy.css 里旧商品卡的 .qz-card 规则（min-height 270、hover 上浮、::before 圆斑、pointer）
   在它被删掉之前不能影响订单行。 */
.qz-card.qz-orders-item {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2) var(--qz-space-4);
  min-height: 0;
  border-radius: var(--qz-radius-control);
  overflow: visible;
  cursor: auto;
  transform: none;
  transition: border-color var(--qz-duration-fast) var(--qz-ease-standard);
}

.qz-card.qz-orders-item::before {
  content: none;
}

.qz-card.qz-orders-item:hover,
.qz-card.qz-orders-item:active {
  border-color: var(--qz-border-strong);
  box-shadow: var(--qz-shadow-resting);
  transform: none;
}

.qz-orders-prod {
  display: flex;
  flex: 0 1 auto;
  align-items: center;
  gap: var(--qz-space-2);
  min-width: 0;
  max-width: 100%;
}

.qz-orders-thumb {
  display: grid;
  flex: none;
  place-items: center;
  width: 40px;
  height: 40px;
  overflow: hidden;
  border: 1px solid var(--qz-border);
  border-radius: 12px;
  background: var(--qz-surface-muted);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  font-weight: 700;
}

.qz-orders-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.qz-orders-pname {
  min-width: 0;
  max-width: 230px;
  overflow: hidden;
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.qz-orders-pextra {
  flex: none;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  font-weight: 600;
}

.qz-orders-main {
  display: flex;
  flex: 1 1 240px;
  flex-wrap: wrap;
  align-items: baseline;
  gap: var(--qz-space-2);
  min-width: 0;
}

.qz-orders-sn {
  color: var(--qz-text);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.qz-orders-time {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.qz-orders-amt {
  display: flex;
  flex: none;
  align-items: baseline;
  gap: var(--qz-space-2);
}

.qz-orders-amt b {
  color: var(--qz-text);
  font-size: var(--qz-text-base);
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.qz-orders-amt small {
  color: var(--qz-danger);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.qz-orders-acts {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2);
  margin-left: auto;
}

/* 「处理中」用主色底的 badge，基类只带四种状态色 */
.qz-badge.is-accent {
  color: var(--qz-accent-strong);
  background: var(--qz-accent-subtle);
}

/* ---------- 分页 ---------- */
.qz-orders-pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--qz-space-3);
  padding-top: var(--qz-space-1);
}

.qz-orders-pager span {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

@media (max-width: 760px) {
  .qz-orders-head {
    flex-direction: column-reverse;
    align-items: flex-start;
    gap: var(--qz-space-3);
  }

  .qz-orders-form {
    grid-template-columns: minmax(0, 1fr);
  }

  .qz-orders-submit {
    width: 100%;
  }

  .qz-card.qz-orders-item {
    padding: var(--qz-space-3);
  }

  .qz-orders-acts {
    flex: 1 1 100%;
    justify-content: flex-end;
    margin-left: 0;
  }
}
</style>
