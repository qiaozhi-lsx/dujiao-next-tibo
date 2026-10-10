<template>
  <RouterLink
    :to="`/products/${product.slug}`"
    class="j8-card j8-reveal is-pending"
    :class="{ 'is-sold-out': soldOut }"
    :data-reveal="1"
  >
    <div class="j8-card-top">
      <span class="j8-card-icon" :class="{ 'is-logo': !cover && !!icon }">
        <span class="j8-brand-icon">
          <img v-if="image" :src="image" alt="" loading="lazy" @error="imageErrored = true" />
          <b v-else>{{ letter }}</b>
        </span>
      </span>
      <span v-if="badge" class="j8-card-badge" :class="{ 'is-sold': soldOut }">{{ badge }}</span>
    </div>
    <h3>{{ title }}</h3>
    <div class="j8-card-chips">
      <span :class="chipTone(0)">{{ fulfillmentLabel }}</span>
      <span v-if="purchaseLabel" :class="chipTone(1)">{{ purchaseLabel }}</span>
      <span v-if="soldOut" class="tone-3">{{ t('products.stockStatus.outOfStock') }}</span>
    </div>
    <div class="j8-card-bottom">
      <div class="j8-card-price">
        <small>{{ t('tibo.card.finalPrice') }}</small>
        <strong>{{ priceText }}</strong>
        <del v-if="originalPriceText">{{ originalPriceText }}</del>
      </div>
      <span class="j8-card-stock">{{ stockLabel }}</span>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useLocalized, useProductLabels } from '../../../composables/useProduct'
import { getFirstImageUrl, getImageUrl } from '../../../utils/image'
import { useTiboPrice } from '../utils/price'

const props = defineProps<{ product: any }>()

const { t } = useI18n()
const { getLocalizedText } = useLocalized()
const { getFulfillmentTypeLabel, getPurchaseTypeLabel, isSoldOut } = useProductLabels()
const { format: formatPrice } = useTiboPrice()

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

const chipTone = (i: number) => (i === 0 ? 'tone-1' : 'tone-2')

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
