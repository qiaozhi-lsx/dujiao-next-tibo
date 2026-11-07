<template>
  <div ref="rootEl" class="qz-container qz-cart">
    <!-- 页头：eyebrow + h1 + 说明 + 返回 -->
    <header class="qz-page-head">
      <div class="qz-page-head-copy">
        <span class="qz-eyebrow">{{ t('qiaozhi.cart.eyebrow') }}</span>
        <h1>
          <span>{{ t('cart.title') }}</span>
          <span v-if="cartItems.length" class="qz-chip" aria-hidden="true">{{ totalItems }} {{ t('qiaozhi.catalog.unit') }}</span>
        </h1>
        <p>{{ t('cart.subtitle') }}</p>
      </div>
      <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm" @click="goBack">
        <ChevronLeft />
        {{ t('qiaozhi.cart.back') }}
      </button>
    </header>

    <!-- 骨架 -->
    <div v-if="loading" class="qz-cart-skeleton" aria-hidden="true">
      <div class="qz-skeleton is-lg"></div>
      <div class="qz-skeleton is-line"></div>
      <div class="qz-skeleton is-sm"></div>
    </div>

    <!-- 空态 -->
    <div v-else-if="!cartItems.length" class="qz-empty">
      <ShoppingCart />
      <b>{{ t('cart.empty') }}</b>
      <span>{{ t('qiaozhi.cart.emptyHint') }}</span>
      <RouterLink to="/products" class="qz-btn qz-btn--primary">
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
            class="qz-card qz-cart-item qz-reveal is-pending"
            :data-reveal="1"
          >
            <!-- 图 -->
            <div class="qz-cart-media">
              <img
                v-if="cartItemImage(item)"
                :src="cartItemImage(item)"
                :alt="itemTitle(item)"
                loading="lazy"
              />
              <span v-else class="qz-cart-initial" aria-hidden="true">{{ itemTitle(item).slice(0, 1) || '·' }}</span>
            </div>

            <!-- 标题 + 移除 -->
            <div class="qz-cart-item-head">
              <RouterLink :to="`/products/${item.slug}`" class="qz-cart-item-title">
                <h2>{{ itemTitle(item) }}</h2>
              </RouterLink>
              <QzHoldConfirm
                :label="t('cart.remove')"
                :confirmed-label="t('cart.removedShort')"
                :hint="t('cart.holdRemoveHint', { seconds: '1.2' })"
                @confirm="removeWithUndo(item)"
              />
            </div>

            <!-- 属性 chips -->
            <div class="qz-cart-chips">
              <span class="qz-chip">
                {{ item.purchaseType === 'guest' ? t('productPurchase.guest') : t('productPurchase.member') }}
              </span>
              <span class="qz-chip">
                {{ item.fulfillmentType === 'auto'
                  ? t('products.fulfillmentType.auto')
                  : t('products.fulfillmentType.manual') }}
              </span>
              <span v-if="itemSkuDisplay(item)" class="qz-chip">{{ itemSkuDisplay(item) }}</span>
            </div>

            <!-- 数量 + 价格 -->
            <div class="qz-cart-item-foot">
              <div class="qz-field qz-cart-qty">
                <span class="qz-field-label">{{ t('productDetail.quantity') }}</span>
                <QzQtyStepper
                  :model-value="item.quantity"
                  :min="itemPurchaseMin(item)"
                  :max="itemMaxQuantity(item)"
                  @update:model-value="(q) => updateQty(item, q)"
                />
                <p v-if="itemStockHint(item)" class="qz-field-hint">{{ itemStockHint(item) }}</p>
                <p v-if="quantityWarning(item)" class="qz-cart-warn" role="alert">
                  {{ quantityWarning(item) }}
                </p>
              </div>
              <div class="qz-cart-prices">
                <span class="qz-cart-unit">
                  <span class="qz-cart-price-sym">{{ unitPriceParts(item).symbol }}</span>
                  <strong>{{ unitPriceParts(item).major }}{{ unitPriceParts(item).minor }}</strong>
                </span>
                <span class="qz-cart-subtotal">
                  <small>{{ t('checkout.totalPriceLabel') }}</small>
                  <b>{{ itemSubtotal(item) }}</b>
                </span>
              </div>
            </div>
          </article>
        </div>

        <!-- 右侧：竖排结算卡（桌面 sticky 跟随） -->
        <aside class="qz-panel qz-cart-summary qz-reveal is-pending" :data-reveal="1">
          <h2>{{ t('cart.summaryTitle') }}</h2>
          <dl class="qz-cart-summary-rows">
            <div>
              <dt>{{ t('cart.itemsCount') }}</dt>
              <dd>{{ totalItems }}</dd>
            </div>
          </dl>
          <p class="qz-cart-disclaimer">{{ t('cart.disclaimer') }}</p>
          <div class="qz-cart-total">
            <span>{{ t('cart.totalLabel') }}</span>
            <span class="qz-cart-total-price">
              <span class="qz-cart-price-sym">{{ totalParts.symbol }}</span>
              <strong>{{ totalParts.major }}{{ totalParts.minor }}</strong>
            </span>
          </div>
          <RouterLink to="/checkout" class="qz-btn qz-btn--primary qz-btn--lg qz-btn--block">
            <ShoppingBag />
            {{ t('cart.checkout') }}
          </RouterLink>
          <RouterLink to="/products" class="qz-btn qz-btn--secondary qz-btn--block">
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
.qz-cart {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-5);
}

/* 页头与骨架行用 base.css 的 .qz-page-head / .qz-skeleton，这里只留骨架的排布 */
.qz-cart-skeleton {
  display: grid;
  gap: var(--qz-space-4);
}

/* ---------- 桌面双栏：左商品卡列表 + 右竖排结算卡 ---------- */
.qz-cart-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: var(--qz-space-4);
  align-items: start;
}

.qz-cart-items {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  min-width: 0;
}

/* 商品卡：图占左列三行，右列依次是标题行 / chips / 数量与价格
   .qz-card.qz-cart-item 写成 (0,3,0)：base.css 的 .qz-card 是 (0,2,0)，串上页面类靠特指度覆盖它的内边距，
   不依赖样式注入顺序。 */
.qz-card.qz-cart-item {
  display: grid;
  grid-template-columns: 148px minmax(0, 1fr);
  grid-template-rows: auto auto 1fr;
  gap: var(--qz-space-3) var(--qz-space-4);
  min-height: 0;
  border-radius: var(--qz-radius-control);
  overflow: visible;
  cursor: auto;
  transform: none;
  transition: none;
}

.qz-card.qz-cart-item::before {
  content: none;
}

.qz-card.qz-cart-item:hover,
.qz-card.qz-cart-item:active {
  border-color: var(--qz-border);
  box-shadow: var(--qz-shadow-resting);
  transform: none;
}

.qz-cart-media {
  display: grid;
  grid-row: 1 / span 3;
  place-items: center;
  aspect-ratio: 1;
  overflow: hidden;
  border-radius: var(--qz-radius-sm);
  background: var(--qz-surface-muted);
}

.qz-cart-media img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.qz-cart-initial {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-3xl);
  font-weight: 800;
  letter-spacing: var(--qz-tracking-display);
}

.qz-cart-item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--qz-space-3);
  min-width: 0;
}

.qz-cart-item-title {
  flex: 1 1 auto;
  min-width: 0;
  color: inherit;
  text-decoration: none;
}

.qz-cart-item-title h2 {
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: var(--qz-leading-body);
  overflow-wrap: anywhere;
}

.qz-cart-item-title:hover h2 {
  text-decoration: underline;
  text-underline-offset: 3px;
}

.qz-cart-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--qz-space-1) var(--qz-space-2);
  min-width: 0;
}

.qz-cart-item-foot {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--qz-space-3) var(--qz-space-4);
  align-self: end;
  min-width: 0;
}

/* 数量字段是 .qz-field 网格：步进器按自身宽度靠左，不随列拉伸 */
.qz-cart-qty {
  flex: 1 1 200px;
  justify-items: start;
}

.qz-cart-warn {
  margin: 0;
  color: var(--qz-warning);
  font-size: var(--qz-text-xs);
  font-weight: 500;
  line-height: var(--qz-leading-body);
}

/* 价格：单价大字 + 小计小字，右对齐；符号紧贴数字 */
.qz-cart-prices {
  display: grid;
  justify-items: end;
  gap: var(--qz-space-1);
  text-align: right;
}

.qz-cart-unit,
.qz-cart-total-price {
  display: inline-flex;
  align-items: baseline;
  color: var(--qz-text);
  font-weight: 700;
  letter-spacing: var(--qz-tracking-display);
  font-variant-numeric: tabular-nums;
}

.qz-cart-price-sym {
  font-size: var(--qz-text-sm);
}

.qz-cart-unit strong {
  font-size: var(--qz-text-xl);
  line-height: 1;
}

.qz-cart-subtotal {
  display: inline-flex;
  align-items: baseline;
  gap: var(--qz-space-1);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

/* <small> 不能继承 UA 的 `smaller`（12 × 0.8 = 9.6px，跌破 11px 地板），显式钉在 xs */
.qz-cart-subtotal small {
  font-size: var(--qz-text-xs);
}

.qz-cart-subtotal b {
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

/* ---------- 结算卡：桌面 sticky 跟随滚动（列表短于一屏时页面不滚，天然不浮动）
   .qz-panel.qz-cart-summary 写成 (0,3,0)：base.css 的 .qz-panel 是 (0,2,0)，串上页面类靠特指度覆盖它，
   不依赖样式注入顺序。 */
.qz-panel.qz-cart-summary {
  position: sticky;
  top: calc(var(--qz-header-h) + var(--qz-space-4));
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  overflow: visible;
}

.qz-panel.qz-cart-summary h2 {
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: var(--qz-leading-body);
}

.qz-cart-summary-rows {
  display: grid;
  gap: var(--qz-space-2);
  margin: 0;
}

.qz-cart-summary-rows div {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--qz-space-3);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
}

.qz-cart-summary-rows dd {
  margin: 0;
  color: var(--qz-text);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.qz-cart-disclaimer {
  margin: 0;
  padding: var(--qz-space-3) var(--qz-space-4);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  white-space: pre-line;
}

.qz-cart-total {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--qz-space-3);
  padding-top: var(--qz-space-4);
  border-top: 1px solid var(--qz-border);
}

.qz-cart-total > span:first-child {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  font-weight: 600;
}

.qz-cart-total-price .qz-cart-price-sym {
  font-size: var(--qz-text-base);
}

.qz-cart-total-price strong {
  font-size: var(--qz-text-2xl);
  line-height: 1;
}

@media (max-width: 1023px) {
  .qz-cart-layout {
    grid-template-columns: minmax(0, 1fr);
    gap: var(--qz-space-3);
  }

  .qz-cart-items {
    gap: var(--qz-space-3);
  }

  .qz-panel.qz-cart-summary {
    position: static;
  }
}

@media (max-width: 760px) {
  /* 手机：图缩到 88px 只陪标题行，chips 与数量/价格整行铺开 */
  .qz-card.qz-cart-item {
    grid-template-columns: 88px minmax(0, 1fr);
    grid-template-rows: auto;
    gap: var(--qz-space-3);
    padding: var(--qz-space-3);
  }

  .qz-cart-media {
    grid-row: auto;
  }

  .qz-cart-chips,
  .qz-cart-item-foot {
    grid-column: 1 / -1;
  }

  .qz-cart-item-head {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
