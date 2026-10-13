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
        v-for="order in orders"
        :key="order.order_no"
        class="j8-orders-item j8-reveal is-pending"
      >
        <div class="j8-orders-main">
          <b class="j8-orders-sn">{{ order.order_no }}</b>
          <span class="j8-orders-time">{{ formatDate(order.created_at) }}</span>
        </div>
        <div class="j8-orders-amt">
          <b>{{ formatMoney(order.total_amount, order.currency) }}</b>
          <small v-if="discountText(order)" class="is-rose">{{ discountText(order) }}</small>
        </div>
        <div class="j8-orders-acts">
          <span class="j8-order-badge" :class="toneClass(order.status)">{{ statusLabel(order.status) }}</span>
          <RouterLink
            class="j8-orders-act"
            :to="{ name: 'guest-order-detail', params: { order_no: order.order_no } }"
          >
            {{ t('guestOrders.viewDetails') }}
          </RouterLink>
          <RouterLink
            v-if="order.status === 'pending_payment'"
            class="j8-orders-act is-primary"
            :to="`/pay?guest=1&order_no=${encodeURIComponent(order.order_no)}`"
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
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, ChevronRight, PackageOpen, Search } from 'lucide-vue-next'
import { useGuestOrders } from '../../composables/useGuestOrders'
import { setupReveal } from './utils/reveal'

const { t } = useI18n()

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
