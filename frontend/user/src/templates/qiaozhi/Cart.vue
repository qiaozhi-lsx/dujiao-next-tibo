<template>
  <div ref="rootEl" class="qz-container qz-pd">
    <!-- 返回条 -->
    <button type="button" class="qz-pd-back" @click="goBack">
      <ChevronLeft />
      {{ t('qiaozhi.cart.back') }}
    </button>

    <!-- 小标题区 -->
    <section class="qz-section-heading">
      <div>
        <span class="qz-eyebrow">{{ t('qiaozhi.cart.eyebrow') }}</span>
        <h2>{{ t('cart.title') }}</h2>
        <p class="qz-cart-subtitle">{{ t('cart.subtitle') }}</p>
      </div>
      <span v-if="cartItems.length">{{ totalItems }} {{ t('qiaozhi.catalog.unit') }}</span>
    </section>

    <!-- 骨架 -->
    <div v-if="loading" class="qz-pd-skeleton" aria-hidden="true">
      <div class="qz-pd-sk-line is-lg"></div>
      <div class="qz-pd-sk-line"></div>
      <div class="qz-pd-sk-line is-sm"></div>
    </div>

    <!-- 空态 -->
    <div v-else-if="!cartItems.length" class="qz-empty">
      <ShoppingCart />
      <b>{{ t('cart.empty') }}</b>
      <span>{{ t('qiaozhi.cart.emptyHint') }}</span>
      <RouterLink to="/products" class="qz-btn-primary">
        {{ t('cart.emptyAction') }}
      </RouterLink>
    </div>

    <!-- 购物车项列表 -->
    <template v-else>
      <div class="qz-cart-layout">
        <div class="qz-cart-items">
          <article
            v-for="item in cartItems"
            :key="cartItemKey(item)"
            class="qz-pd-card qz-cart-card qz-reveal is-pending"
            :data-reveal="1"
          >
            <!-- 左：图 -->
            <div class="qz-pd-media">
              <div class="qz-pd-stage">
                <img
                  v-if="cartItemImage(item)"
                  :src="cartItemImage(item)"
                  :alt="itemTitle(item)"
                  loading="lazy"
                />
                <span v-else class="qz-pd-logo qz-brand-icon">
                  <b>{{ itemTitle(item).slice(0, 1) || '·' }}</b>
                </span>
              </div>
            </div>

            <!-- 右：信息 -->
            <div class="qz-pd-info">
              <div class="qz-cart-head">
                <RouterLink :to="`/products/${item.slug}`" class="qz-cart-title">
                  <h1>{{ itemTitle(item) }}</h1>
                </RouterLink>
                <QzHoldConfirm
                  :label="t('cart.remove')"
                  :confirmed-label="t('cart.removedShort')"
                  :hint="t('cart.holdRemoveHint', { seconds: '1.2' })"
                  @confirm="removeWithUndo(item)"
                />
              </div>

              <div class="qz-pd-chips">
                <span class="is-accent">
                  {{ item.purchaseType === 'guest' ? t('productPurchase.guest') : t('productPurchase.member') }}
                </span>
                <span class="is-accent">
                  {{ item.fulfillmentType === 'auto'
                    ? t('products.fulfillmentType.auto')
                    : t('products.fulfillmentType.manual') }}
                </span>
                <span v-if="itemSkuDisplay(item)">{{ itemSkuDisplay(item) }}</span>
              </div>

              <div class="qz-pd-price-row">
                <div class="qz-pd-price">
                  <span class="qz-pd-price-sym">{{ unitPriceParts(item).symbol }}</span>
                  <strong>{{ unitPriceParts(item).major }}{{ unitPriceParts(item).minor }}</strong>
                </div>
                <div class="qz-cart-subtotal">
                  <small>{{ t('checkout.totalPriceLabel') }}</small>
                  <b>{{ itemSubtotal(item) }}</b>
                </div>
              </div>

              <div class="qz-pd-field">
                <label>{{ t('productDetail.quantity') }}</label>
                <QzQtyStepper
                  :model-value="item.quantity"
                  :min="itemPurchaseMin(item)"
                  :max="itemMaxQuantity(item)"
                  @update:model-value="(q) => updateQty(item, q)"
                />
                <p v-if="itemStockHint(item)" class="qz-pd-hint">{{ itemStockHint(item) }}</p>
                <p v-if="quantityWarning(item)" class="qz-pd-alert is-warm" role="alert">
                  {{ quantityWarning(item) }}
                </p>
              </div>
            </div>
          </article>
        </div>

        <!-- 右侧：竖排结算卡（桌面 sticky 跟随） -->
        <aside class="qz-cart-summary qz-reveal is-pending" :data-reveal="1">
          <h2>{{ t('cart.summaryTitle') }}</h2>
          <div class="qz-pd-note is-accent">
            <b>{{ t('cart.itemsCount') }}</b>
            <em>{{ totalItems }}</em>
          </div>
          <p class="qz-pd-desc">{{ t('cart.disclaimer') }}</p>
          <div class="qz-cart-divider" aria-hidden="true"></div>
          <div class="qz-pd-price-row">
            <span class="qz-cart-total-label">{{ t('cart.totalLabel') }}</span>
            <div class="qz-pd-price">
              <span class="qz-pd-price-sym">{{ totalParts.symbol }}</span>
              <strong>{{ totalParts.major }}{{ totalParts.minor }}</strong>
            </div>
          </div>
          <RouterLink to="/checkout" class="qz-pd-buy">
            <ShoppingBag />
            {{ t('cart.checkout') }}
          </RouterLink>
          <RouterLink to="/products" class="qz-pd-back-link">
            {{ t('cart.emptyAction') }}
          </RouterLink>
        </aside>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, ShoppingBag, ShoppingCart } from 'lucide-vue-next'
import { useCart } from '../../composables/useCart'
import type { CartItem } from '../../stores/cart'
import { useQiaozhiPrice } from './utils/price'
import { setupReveal } from './utils/reveal'
import QzHoldConfirm from './components/QzHoldConfirm.vue'
import QzQtyStepper from './components/QzQtyStepper.vue'

const router = useRouter()
const { t } = useI18n()

const {
  getLocalizedText,
  cartItems,
  totalItems,
  totalAmount,
  cartItemKey,
  cartItemImage,
  itemSkuDisplay,
  itemSubtotal,
  itemStockHint,
  quantityWarning,
  itemPurchaseMin,
  itemMaxQuantity,
  removeWithUndo,
  updateQty,
} = useCart()

const { split: splitPrice } = useQiaozhiPrice()

// ------------------------------------------------------------------
// 本地加载态：避免初次挂载时骨架一直显示；仅在挂载前短暂展示
// ------------------------------------------------------------------
const loading = ref(true)

const rootEl = ref<HTMLElement | null>(null)
let cleanupReveal: (() => void) | null = null

const armReveal = () => {
  cleanupReveal?.()
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
}

// 条目数变化（如删除最后一件后"撤销"）会新建带 is-pending 的条目和结算卡，需重新武装
watch(() => cartItems.value.length, () => {
  armReveal()
})

const itemTitle = (item: CartItem): string => {
  const text = getLocalizedText(item.title)
  return typeof text === 'string' ? text : String(text ?? '')
}

const unitPriceParts = (item: CartItem) => splitPrice(item.priceAmount)
const totalParts = computed(() => splitPrice(totalAmount.value))

const goBack = () => {
  if (window.history.length > 1) router.back()
  else router.push('/')
}

onMounted(() => {
  // 让 cart store 的 stock 快照刷新先走一轮，避免 reveal 和骨架抖动
  window.setTimeout(() => {
    loading.value = false
    armReveal()
  }, 120)
})

onUnmounted(() => {
  cleanupReveal?.()
})
</script>

<style scoped>
.qz-cart-subtitle {
  margin-top: 6px;
  font-size: 13px;
  color: var(--muted);
}

.qz-cart-card {
  position: relative;
}

.qz-cart-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.qz-cart-title {
  min-width: 0;
  flex: 1;
  color: inherit;
  text-decoration: none;
}

.qz-cart-title h1 {
  margin: 0;
  overflow-wrap: anywhere;
  transition: color 0.15s;
}

.qz-cart-title:hover h1 {
  color: var(--accent-strong);
}

.qz-cart-subtotal {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
  text-align: right;
}

.qz-cart-subtotal small {
  font-size: 11px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--muted);
}

.qz-cart-subtotal b {
  font-size: 15px;
  font-weight: 800;
  color: var(--text);
  font-variant-numeric: tabular-nums;
}

/* 桌面双栏：左商品卡列表 + 右竖排结算卡 */
.qz-cart-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 16px;
  align-items: start;
}

.qz-cart-items {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

/* 结算卡：独立竖排卡片；桌面 sticky 跟随滚动（列表短于一屏时页面不滚，天然不浮动） */
.qz-cart-summary {
  position: sticky;
  top: calc(var(--site-header-height) + 16px);
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 20px 22px;
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--r-card);
  box-shadow: var(--shadow-soft);
}

.qz-cart-summary h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: var(--text);
}

.qz-cart-summary .qz-pd-note em {
  margin-left: auto;
  font-style: normal;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.qz-cart-divider {
  height: 1px;
  background: var(--line);
}

.qz-cart-total-label {
  font-size: 13px;
  font-weight: 700;
  color: var(--muted);
}

.qz-pd-back-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 42px;
  border-radius: var(--r-pill);
  border: 1px solid var(--line);
  background: var(--surface-2);
  color: var(--text);
  font-size: 14px;
  font-weight: 700;
  text-decoration: none;
}

.qz-pd-back-link:hover {
  color: var(--accent-strong);
  border-color: var(--accent-line);
}

@media (max-width: 1023px) {
  .qz-cart-layout {
    grid-template-columns: minmax(0, 1fr);
    gap: 12px;
  }

  .qz-cart-items {
    gap: 12px;
  }

  .qz-cart-summary {
    position: static;
  }

  .qz-cart-head {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
