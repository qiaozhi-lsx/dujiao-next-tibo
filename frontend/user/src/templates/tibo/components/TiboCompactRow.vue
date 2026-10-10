<template>
  <RouterLink class="store-compact" :to="`/products/${product.slug}`">
    <span class="product-mark">
      <img v-if="cover" :src="cover" alt="" loading="lazy" />
      <Package v-else />
    </span>
    <span>
      <strong>{{ title }}</strong>
      <small>{{ subtitle }}</small>
    </span>
    <ChevronRight class="row-chevron" />
  </RouterLink>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ChevronRight, Package } from 'lucide-vue-next'
import { useLocalized } from '../../../composables/useProduct'
import { getFirstImageUrl, getImageUrl } from '../../../utils/image'
import { useTiboPrice } from '../utils/price'

const props = defineProps<{
  product: any
}>()

const { getLocalizedText } = useLocalized()
const { format: formatPrice } = useTiboPrice()

const title = computed(() => getLocalizedText(props.product?.title))
const subtitle = computed(() => formatPrice(props.product?.price_amount))

const imageErrored = ref(false)
const cover = computed(() => {
  if (imageErrored.value) return ''
  const primary = getFirstImageUrl(props.product?.images)
  if (primary) return primary
  const icon = props.product?.category?.icon
  return icon ? getImageUrl(icon) : ''
})
</script>

<style scoped>
.row-chevron {
  width: 16px;
  height: 16px;
}
</style>
