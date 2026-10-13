<template>
  <div ref="rootEl" class="j8-container j8-pd">
    <!-- 返回条 -->
    <button type="button" class="j8-pd-back" @click="goBack">
      <ChevronLeft />
      {{ t('qiaozhi.cart.back') }}
    </button>

    <!-- 小标题区 -->
    <section class="j8-section-heading">
      <div>
        <span class="j8-eyebrow">{{ t('qiaozhi.cart.eyebrow') }}</span>
        <h2>{{ t('cart.title') }}</h2>
        <p class="j8-cart-subtitle">{{ t('cart.subtitle') }}</p>
      </div>
      <span v-if="cartItems.length">{{ totalItems }} {{ t('qiaozhi.catalog.unit') }}</span>
    </section>

    <!-- 骨架 -->
    <div v-if="loading" class="j8-pd-skeleton" aria-hidden="true">
      <div class="j8-pd-sk-line is-lg"></div>
      <div class="j8-pd-sk-line"></div>
      <div class="j8-pd-sk-line is-sm"></div>
    </div>

    <!-- 空态 -->
    <div v-else-if="!cartItems.length" class="j8-empty">
      <ShoppingCart />
      <b>{{ t('cart.empty') }}</b>
      <span>{{ t('qiaozhi.cart.emptyHint') }}</span>
      <RouterLink to="/products" class="j8-btn-primary">
        {{ t('cart.emptyAction') }}
      </RouterLink>
    </div>

    <!-- 购物车项列表 -->
    <template v-else>
      <article
        v-for="item in cartItems"
        :key="cartItemKey(item)"
        class="j8-pd-card j8-cart-card j8-reveal is-pending"
        :data-reveal="1"
      >
        <!-- 左：图 -->
        <div class="j8-pd-media">
          <div class="j8-pd-stage">
            <img
              v-if="cartItemImage(item)"
              :src="cartItemImage(item)"
              :alt="itemTitle(item)"
              loading="lazy"
            />
            <span v-else class="j8-pd-logo j8-brand-icon">
              <b>{{ itemTitle(item).slice(0, 1) || '·' }}</b>
            </span>
          </div>
        </div>

        <!-- 右：信息 -->
        <div class="j8-pd-info">
          <div class="j8-cart-head">
            <RouterLink :to="`/products/${item.slug}`" class="j8-cart-title">
              <h1>{{ itemTitle(item) }}</h1>
            </RouterLink>
            <button
              type="button"
              class="j8-cart-remove"
              :aria-label="t('cart.remove')"
              @click="removeWithUndo(item)"
            >
              <Trash2 />
              <span>{{ t('cart.remove') }}</span>
            </button>
          </div>

          <div class="j8-pd-chips">
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

          <div class="j8-pd-price-row">
            <div class="j8-pd-price">
              <span class="j8-pd-price-sym">{{ unitPriceParts(item).symbol }}</span>
              <strong>{{ unitPriceParts(item).major }}{{ unitPriceParts(item).minor }}</strong>
            </div>
            <div class="j8-cart-subtotal">
              <small>{{ t('checkout.totalPriceLabel') }}</small>
              <b>{{ itemSubtotal(item) }}</b>
            </div>
          </div>

          <div class="j8-pd-field">
            <label>{{ t('productDetail.quantity') }}</label>
            <div class="j8-pd-qty">
              <button
                type="button"
                :disabled="item.quantity <= itemPurchaseMin(item)"
                :aria-label="t('common.previous')"
                @click="updateQty(item, item.quantity - 1)"
              >
                <Minus />
              </button>
              <input
                type="text"
                inputmode="numeric"
                :id="`qz-cart-qty-${item.productId}`"
                :name="`qz-cart-qty-${item.productId}`"
                :value="item.quantity"
                @change="handleQtyChange(item, $event)"
              />
              <button
                type="button"
                :disabled="item.quantity >= itemMaxQuantity(item)"
                :aria-label="t('common.next')"
                @click="updateQty(item, item.quantity + 1)"
              >
                <Plus />
              </button>
            </div>
            <p v-if="itemStockHint(item)" class="j8-pd-hint">{{ itemStockHint(item) }}</p>
            <p v-if="quantityWarning(item)" class="j8-pd-alert is-warm" role="alert">
              {{ quantityWarning(item) }}
            </p>
          </div>
        </div>
      </article>

      <!-- 底部合计条 -->
      <section class="j8-cart-summary j8-reveal is-pending" :data-reveal="1">
        <div class="j8-cart-summary-info">
          <h2>{{ t('cart.summaryTitle') }}</h2>
          <div class="j8-pd-note is-accent">
            <b>{{ t('cart.itemsCount') }}</b>
            <em>{{ totalItems }}</em>
          </div>
          <p class="j8-pd-desc">{{ t('cart.disclaimer') }}</p>
        </div>
        <div class="j8-cart-summary-cta">
          <div class="j8-pd-price-row">
            <span class="j8-cart-total-label">{{ t('cart.totalLabel') }}</span>
            <div class="j8-pd-price">
              <span class="j8-pd-price-sym">{{ totalParts.symbol }}</span>
              <strong>{{ totalParts.major }}{{ totalParts.minor }}</strong>
            </div>
          </div>
          <RouterLink to="/checkout" class="j8-pd-buy">
            <ShoppingBag />
            {{ t('cart.checkout') }}
          </RouterLink>
          <RouterLink to="/products" class="j8-pd-back-link">
            {{ t('cart.emptyAction') }}
          </RouterLink>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, Minus, Plus, ShoppingBag, ShoppingCart, Trash2 } from 'lucide-vue-next'
import { useCart } from '../../composables/useCart'
import type { CartItem } from '../../stores/cart'
import { useQiaozhiPrice } from './utils/price'
import { setupReveal } from './utils/reveal'

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
  handleQtyChange,
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
.j8-cart-subtitle {
  margin-top: 6px;
  font-size: 13px;
  color: var(--muted);
}

.j8-cart-card {
  position: relative;
}

.j8-cart-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.j8-cart-title {
  min-width: 0;
  flex: 1;
  color: inherit;
  text-decoration: none;
}

.j8-cart-title h1 {
  margin: 0;
  overflow-wrap: anywhere;
  transition: color 0.15s;
}

.j8-cart-title:hover h1 {
  color: var(--accent-strong);
}

.j8-cart-remove {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
  padding: 8px 12px;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: var(--surface-2);
  color: var(--muted);
  font-size: 12.5px;
  font-weight: 700;
  transition: color 0.15s, border-color 0.15s, background 0.15s;
}

.j8-cart-remove svg {
  width: 14px;
  height: 14px;
}

.j8-cart-remove:hover {
  color: var(--red, #dc2626);
  border-color: var(--red, #dc2626);
  background: var(--surface);
}

.j8-cart-subtotal {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
  text-align: right;
}

.j8-cart-subtotal small {
  font-size: 11px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--muted);
}

.j8-cart-subtotal b {
  font-size: 15px;
  font-weight: 800;
  color: var(--text);
  font-variant-numeric: tabular-nums;
}

/* 合计条：独立卡片，不走 j8-pd-card 两栏 */
.j8-cart-summary {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(260px, 360px);
  gap: 24px;
  padding: 24px 28px;
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--r-card);
  box-shadow: var(--shadow-soft);
}

.j8-cart-summary-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.j8-cart-summary-info h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 900;
  color: var(--text);
}

.j8-cart-summary-info .j8-pd-note em {
  margin-left: auto;
  font-style: normal;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.j8-cart-summary-cta {
  display: flex;
  flex-direction: column;
  gap: 10px;
  justify-content: center;
}

.j8-cart-total-label {
  font-size: 13px;
  font-weight: 700;
  color: var(--muted);
}

.j8-pd-back-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 42px;
  border-radius: 11px;
  border: 1px solid var(--line);
  background: var(--surface-2);
  color: var(--text);
  font-size: 14px;
  font-weight: 700;
  text-decoration: none;
  transition: color 0.15s, border-color 0.15s;
}

.j8-pd-back-link:hover {
  color: var(--accent-strong);
  border-color: var(--accent-line);
}

@media (max-width: 1023px) {
  .j8-cart-summary {
    grid-template-columns: minmax(0, 1fr);
    padding: 18px;
  }

  .j8-cart-head {
    flex-direction: column;
    align-items: stretch;
  }

  .j8-cart-remove {
    align-self: flex-end;
  }
}
</style>
