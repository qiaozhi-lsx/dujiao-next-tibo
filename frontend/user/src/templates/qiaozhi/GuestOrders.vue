<template>
  <div ref="rootEl" class="j8-container j8-orders-page">
    <!-- 标题 -->
    <div class="j8-orders-hero j8-reveal is-pending">
      <h1 class="j8-orders-title">{{ t('guestOrders.title') }}</h1>
      <p class="j8-orders-sub">{{ t('guestOrders.subtitle') }}</p>
    </div>

    <!-- 查询表单 -->
    <section class="j8-orders-search j8-reveal is-pending">
      <div v-if="hasSavedAuth" class="j8-orders-saved">
        <span>{{ t('guestOrders.savedHint', { email: savedAuth.email || '-' }) }}</span>
        <button type="button" class="j8-orders-clear" @click="clearSaved">{{ t('guestOrders.clearSaved') }}</button>
      </div>
      <form class="j8-orders-form" novalidate @submit.prevent="handleSearch">
        <input
          v-model="email"
          type="email"
          autocomplete="email"
          :placeholder="t('guestOrders.emailPlaceholder')"
        />
        <input
          v-model="orderPassword"
          type="password"
          autocomplete="current-password"
          :placeholder="t('guestOrders.passwordPlaceholder')"
        />
        <input v-model="orderNo" type="text" :placeholder="t('guestOrders.orderNoPlaceholder')" />
        <button type="submit" class="j8-orders-submit" :disabled="loading">
          <Search />
          {{ loading ? t('guestOrders.searching') : t('guestOrders.search') }}
        </button>
      </form>
      <p class="j8-orders-tip">{{ t('guestOrders.tip') }}</p>
      <p v-if="error" class="j8-orders-error" role="alert">{{ error }}</p>
    </section>

    <!-- 空态 -->
    <div v-if="orders.length === 0 && !loading" class="j8-empty j8-reveal is-pending">
      <PackageOpen />
      <b>{{ emptyMessage }}</b>
    </div>

    <!-- 订单列表 -->
    <div v-else class="j8-orders-list">
      <article
        v-for="row in rows"
        :key="row.order.order_no"
        class="j8-orders-item j8-reveal is-pending"
      >
        <div v-if="row.product" class="j8-orders-prod">
          <span class="j8-orders-thumb">
            <img
              v-if="row.product.image"
              :src="row.product.image"
              :alt="row.product.title"
              loading="lazy"
              decoding="async"
            />
            <template v-else>{{ row.product.initial }}</template>
          </span>
          <b class="j8-orders-pname">{{ row.product.title }}</b>
          <span v-if="row.product.extraCount" class="j8-orders-pextra">
            {{ t('guestOrders.moreItems', { count: row.product.extraCount }) }}
          </span>
          <span v-else-if="row.product.quantity > 1" class="j8-orders-pqty">× {{ row.product.quantity }}</span>
        </div>
        <div class="j8-orders-main">
          <b class="j8-orders-sn">{{ row.order.order_no }}</b>
          <span class="j8-orders-time">{{ formatDate(row.order.created_at) }}</span>
        </div>
        <div class="j8-orders-amt">
          <b>{{ formatMoney(row.order.total_amount, row.order.currency) }}</b>
          <small v-if="discountText(row.order)" class="is-rose">{{ discountText(row.order) }}</small>
        </div>
        <div class="j8-orders-acts">
          <span class="j8-order-badge" :class="toneClass(row.order.status)">{{ statusLabel(row.order.status) }}</span>
          <RouterLink
            class="j8-orders-act"
            :to="{ name: 'guest-order-detail', params: { order_no: row.order.order_no } }"
          >
            {{ t('guestOrders.viewDetails') }}
          </RouterLink>
          <RouterLink
            v-if="row.order.status === 'pending_payment'"
            class="j8-orders-act is-primary"
            :to="`/pay?guest=1&order_no=${encodeURIComponent(row.order.order_no)}`"
          >
            {{ t('guestOrders.payNow') }}
          </RouterLink>
        </div>
      </article>
    </div>

    <!-- 分页 -->
    <div v-if="pagination.total_page > 1" class="j8-orders-pager">
      <button
        type="button"
        :disabled="pagination.page <= 1 || loading"
        :aria-label="t('pagination.previous')"
        @click="changePage(pagination.page - 1)"
      >
        <ChevronLeft />
      </button>
      <span>{{ pagination.page }} / {{ pagination.total_page }}</span>
      <button
        type="button"
        :disabled="pagination.page >= pagination.total_page || loading"
        :aria-label="t('pagination.next')"
        @click="changePage(pagination.page + 1)"
      >
        <ChevronRight />
      </button>
    </div>
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

const { t } = useI18n()
const appStore = useAppStore()

const {
  savedAuth, email, orderPassword, orderNo, loading, error, orders, pagination,
  hasSavedAuth, clearSaved, handleSearch, emptyMessage, changePage,
  statusLabel, statusVariant, formatMoney, formatDiscountMoney, hasDiscountAmount, formatDate,
} = useGuestOrders()

const toneClass = (status?: string | null) => {
  switch (statusVariant(status || '')) {
    case 'success': return 'is-ok'
    case 'warning': return 'is-warn'
    case 'info': return 'is-info'
    case 'danger': return 'is-danger'
    case 'accent': return 'is-accent'
    default: return 'is-neutral'
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
