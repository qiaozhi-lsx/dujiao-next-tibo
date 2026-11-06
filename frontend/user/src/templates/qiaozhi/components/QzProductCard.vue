<template>
  <RouterLink
    :to="`/products/${product.slug}`"
    class="qz-card qz-product-card qz-reveal is-pending"
    :class="{ 'is-sold-out': soldOut }"
    :data-reveal="1"
  >
    <div class="qz-card-top">
      <span class="qz-card-icon" :class="{ 'is-logo': !cover && !!icon }">
        <span class="qz-brand-icon">
          <img v-if="image" :src="image" alt="" loading="lazy" @error="imageErrored = true" />
          <b v-else>{{ letter }}</b>
        </span>
      </span>
      <span v-if="badge" class="qz-badge" :class="soldOut ? 'is-danger' : 'is-warning'">{{ badge }}</span>
    </div>
    <h3>{{ title }}</h3>
    <div class="qz-card-chips">
      <span class="qz-chip">{{ fulfillmentLabel }}</span>
      <span v-if="purchaseLabel" class="qz-chip">{{ purchaseLabel }}</span>
      <span v-if="soldOut" class="qz-chip is-danger">{{ t('products.stockStatus.outOfStock') }}</span>
    </div>
    <div class="qz-card-bottom">
      <div class="qz-card-price">
        <small>{{ t('qiaozhi.card.finalPrice') }}</small>
        <strong>{{ priceText }}</strong>
        <del v-if="originalPriceText">{{ originalPriceText }}</del>
      </div>
      <span class="qz-card-stock">{{ stockLabel }}</span>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useLocalized, useProductLabels } from '../../../composables/useProduct'
import { getFirstImageUrl, getImageUrl } from '../../../utils/image'
import { useQiaozhiPrice } from '../utils/price'

const props = defineProps<{ product: any }>()

const { t } = useI18n()
const { getLocalizedText } = useLocalized()
const { getFulfillmentTypeLabel, getPurchaseTypeLabel, isSoldOut } = useProductLabels()
const { format: formatPrice } = useQiaozhiPrice()

const title = computed(() => getLocalizedText(props.product?.title))
const soldOut = computed(() => isSoldOut(props.product))
const letter = computed(() => title.value.slice(0, 1).toUpperCase())

const imageErrored = ref(false)
const icon = computed(() => {
  const catIcon = props.product?.category?.icon
  return catIcon ? getImageUrl(catIcon) : ''
})
const image = computed(() => {
  if (imageErrored.value) return icon.value
  return getFirstImageUrl(props.product?.images) || icon.value
})
const cover = computed(() => getFirstImageUrl(props.product?.images))

const badge = computed(() => {
  const tags = props.product?.tags
  if (Array.isArray(tags) && tags.length) return String(tags[0]).slice(0, 8)
  return ''
})

const fulfillmentLabel = computed(() => getFulfillmentTypeLabel(props.product?.fulfillment_type))
const purchaseLabel = computed(() => getPurchaseTypeLabel(props.product?.purchase_type))

const priceText = computed(() => formatPrice(props.product?.price_amount))

const originalPriceText = computed(() => {
  const original = props.product?.original_price_amount
  if (!original) return ''
  return formatPrice(original)
})

const stockLabel = computed(() => {
  if (soldOut.value) return t('products.stockStatus.outOfStock')
  if (props.product?.stock_status === 'low_stock') return t('products.stockStatus.lowStock')
  return t('products.stockStatus.inStock')
})
</script>

<style scoped>
/* 商品卡：表面 / 边框 / 圆角 / 内边距来自 .qz-card，角标与标签来自 .qz-badge / .qz-chip；
   这里只放卡内布局与 hover / 按压。 */
.qz-product-card {
  display: flex;
  flex-direction: column;
  min-height: 270px;
  overflow: hidden;
  text-decoration: none;
  cursor: pointer;
  transition:
    transform var(--qz-duration-standard) var(--qz-ease-standard),
    border-color var(--qz-duration-fast) var(--qz-ease-standard),
    box-shadow var(--qz-duration-fast) var(--qz-ease-standard),
    opacity var(--qz-duration-fast) var(--qz-ease-standard);
}
@media (hover: hover) and (pointer: fine) {
  .qz-product-card:hover {
    transform: translateY(-3px);
    border-color: var(--qz-border-strong);
    box-shadow: var(--qz-shadow-raised);
  }
}
.qz-product-card:active {
  transform: translateY(-1px) scale(0.985);
  transition-duration: var(--qz-duration-instant);
}
.qz-product-card.is-sold-out { opacity: 0.72; box-shadow: none; }
.qz-product-card.is-sold-out:hover { opacity: 0.9; }

.qz-card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--qz-space-2);
}
.qz-card-icon {
  display: grid;
  flex: none;
  place-items: center;
  width: 62px;
  height: 62px;
  overflow: hidden;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  color: var(--qz-accent-strong);
  font-size: var(--qz-text-base);
  font-weight: 800;
  text-transform: uppercase;
}
.qz-card-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  background: var(--qz-surface);
}
.qz-card-icon.is-logo img {
  box-sizing: border-box;
  padding: var(--qz-space-3);
  background: transparent;
}
/* 暗色下给品牌 logo 垫一块浅底（--qz-text 在暗色是近白），深色 logo 才看得见 */
.dark .qz-card-icon.is-logo img { background: var(--qz-text); }
.qz-card-icon > .qz-brand-icon > svg { width: 55%; height: 55%; }
.qz-card-icon > .qz-brand-icon > b { font-size: var(--qz-text-xl); }

.qz-product-card h3 {
  display: -webkit-box;
  min-height: calc(2 * var(--qz-text-base) * var(--qz-leading-body));
  margin: var(--qz-space-4) 0 var(--qz-space-1);
  overflow: hidden;
  color: var(--qz-text);
  font-size: var(--qz-text-base);
  font-weight: 700;
  line-height: var(--qz-leading-body);
  word-break: break-word;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.qz-card-chips {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-1) var(--qz-space-2);
  min-height: 24px;
  margin-top: var(--qz-space-1);
}
.qz-card-chips .qz-chip { overflow: hidden; text-overflow: ellipsis; }
.qz-card-chips .qz-chip.is-danger {
  color: var(--qz-danger);
  background: color-mix(in oklab, var(--qz-danger) 12%, transparent);
}

.qz-card-bottom {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--qz-space-2);
  margin-top: auto;
  padding-top: var(--qz-space-4);
  border-top: 1px solid var(--qz-border);
}
.qz-card-price {
  display: flex;
  align-items: baseline;
  gap: var(--qz-space-1);
  min-width: 0;
}
.qz-card-price small {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
  white-space: nowrap;
}
.qz-card-price strong {
  color: var(--qz-text);
  font-size: var(--qz-text-xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
  font-variant-numeric: tabular-nums;
}
.qz-card-price del {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
  white-space: nowrap;
}
.qz-card-stock {
  flex: none;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-2xs);
  white-space: nowrap;
}
.qz-product-card.is-sold-out .qz-card-stock { color: var(--qz-danger); }

@media (max-width: 760px) {
  /* 带上基类名，靠特指度覆盖 .qz-card 的 padding */
  .qz-card.qz-product-card {
    min-height: 190px;
    padding: var(--qz-space-3);
  }
  .qz-card-icon {
    width: 42px;
    height: 42px;
    font-size: var(--qz-text-sm);
  }
  .qz-card-icon.is-logo img { padding: var(--qz-space-2); }
  .qz-card-icon > .qz-brand-icon > b { font-size: var(--qz-text-base); }
  .qz-product-card h3 {
    min-height: calc(2 * var(--qz-text-sm) * var(--qz-leading-body));
    margin-top: var(--qz-space-3);
    font-size: var(--qz-text-sm);
  }
  .qz-card-chips { gap: var(--qz-space-1); }
  .qz-card-bottom {
    gap: var(--qz-space-1);
    padding-top: var(--qz-space-2);
  }
  .qz-card-price {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0 var(--qz-space-1);
  }
  .qz-card-price small { grid-column: 1 / -1; }
  .qz-card-price strong { font-size: var(--qz-text-lg); }
  .qz-card-price del { display: none; }
}

@media (max-width: 360px) {
  .qz-card-stock { display: none; }
}
</style>
