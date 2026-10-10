<template>
  <article class="store-product">
    <div class="store-product-heading">
      <h3>{{ title }}</h3>
      <span v-if="tag">{{ tag }}</span>
    </div>
    <p v-if="spec" class="store-spec">{{ spec }}</p>
    <component :is="visualEntry.visual" v-bind="visualEntry.visualProps" :label="title" />
    <div class="store-product-bottom">
      <span class="store-starting-price">{{ priceText }}<small v-if="priceSuffix"> {{ priceSuffix }}</small></span>
      <RouterLink class="store-action" :to="`/products/${product.slug}`">{{ t('tibo.card.choosePlan') }}</RouterLink>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useLocalized } from '../../../composables/useProduct'
import { useTiboPrice } from '../utils/price'
import { resolveProductVisual } from './productVisuals'

const props = defineProps<{
  product: any
}>()

const { t } = useI18n()
const { getLocalizedText } = useLocalized()
const { format: formatPrice } = useTiboPrice()

const title = computed(() => getLocalizedText(props.product?.title))
const tag = computed(() => getLocalizedText(props.product?.category?.name))
const priceText = computed(() => formatPrice(props.product?.price_amount))
const priceSuffix = computed(() => t('tibo.card.priceFrom'))
const visualEntry = computed(() => resolveProductVisual(props.product?.slug))

const spec = computed(() => {
  const raw = getLocalizedText(props.product?.description)
  if (!raw) return ''
  const text = raw.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').trim()
  return text.length > 60 ? `${text.slice(0, 60)}…` : text
})
</script>
