<template>
  <div ref="rootEl" class="qz-container">
    <!-- 商品引导区 -->
    <section class="qz-hero">
      <div class="qz-hero-copy">
        <span class="qz-eyebrow">{{ t('qiaozhi.products.hero.eyebrow') }}</span>
        <h1 class="qz-hero-title">
          <span>{{ t('qiaozhi.products.hero.title') }}</span>
        </h1>
        <p>{{ t('qiaozhi.products.hero.subtitle') }}</p>
      </div>
      <div class="qz-hero-actions">
        <label class="qz-search" role="search">
          <Search />
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="t('qiaozhi.products.hero.searchPlaceholder')"
            :aria-label="t('products.searchLabel')"
          />
          <button
            v-if="searchQuery"
            type="button"
            class="qz-search-clear"
            :aria-label="t('products.clearFilters')"
            @click="searchQuery = ''"
          >
            <X />
          </button>
        </label>
      </div>
    </section>

    <!-- 全部商品目录 -->
    <section class="qz-catalog">
      <QzCategoryPanel
        :categories="panelCategories"
        :selected-category="selectedCategory"
        :total-count="allProducts.length"
        @select="onSelectCategory"
      />
      <div>
        <div class="qz-section-heading">
          <div>
            <span class="qz-eyebrow">{{ t('qiaozhi.products.catalog.eyebrow') }}</span>
            <h2>{{ activeCategoryName }}</h2>
          </div>
          <span>{{ filteredProducts.length }} {{ t('qiaozhi.products.catalog.unit') }}</span>
        </div>

        <div v-if="loading" class="qz-grid">
          <div v-for="i in 8" :key="i" class="qz-skeleton-card" aria-hidden="true">
            <span class="qz-skeleton-art"><span></span></span>
            <span class="qz-skeleton-title"></span>
            <span class="qz-skeleton-spec"></span>
            <div class="qz-skeleton-bottom">
              <span class="qz-skeleton-price"></span>
              <span class="qz-skeleton-action"></span>
            </div>
          </div>
        </div>
        <div v-else-if="filteredProducts.length" class="qz-grid">
          <QzProductCard v-for="product in filteredProducts" :key="product.id" :product="product" />
        </div>
        <div v-else class="qz-empty">
          <PackageOpen />
          <b>{{ searchQuery || selectedCategory !== null ? t('products.emptyFiltered') : t('products.empty') }}</b>
          <span v-if="searchQuery || selectedCategory !== null">
            <button type="button" class="qz-btn-primary" @click="resetFilters">{{ t('products.clearFilters') }}</button>
          </span>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { PackageOpen, Search, X } from 'lucide-vue-next'
import { categoryAPI, productAPI } from '../../api'
import { useLocalized } from '../../composables/useProduct'
import { usePageSeo } from '../../composables/usePageSeo'
import { buildCategoryGroups } from '../../utils/category'
import { getImageUrl } from '../../utils/image'
import QzProductCard from './components/QzProductCard.vue'
import QzCategoryPanel, { type PanelCategory } from './components/QzCategoryPanel.vue'
import { setupReveal } from './utils/reveal'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { getLocalizedText } = useLocalized()

const rootEl = ref<HTMLElement | null>(null)
const loading = ref(true)
const allProducts = ref<any[]>([])
const allCategories = ref<any[]>([])
const selectedCategory = ref<number | null>(null)
const searchQuery = ref('')

let cleanupReveal: (() => void) | null = null
let initializing = true

// ==================== 分类面板（含数量角标） ====================
const panelCategories = computed<PanelCategory[]>(() => {
  const counts = new Map<number, number>()
  for (const p of allProducts.value) {
    const cid = Number(p.category_id) || 0
    counts.set(cid, (counts.get(cid) || 0) + 1)
  }
  const groups = buildCategoryGroups(allCategories.value)
  const out: PanelCategory[] = []
  for (const group of groups as any[]) {
    out.push({
      id: group.id,
      name: getLocalizedText(group.name),
      icon: group.icon ? getImageUrl(group.icon) : '',
      count: counts.get(group.id) || 0,
      parentId: 0,
    })
    for (const child of group.children || []) {
      out.push({
        id: child.id,
        name: getLocalizedText(child.name),
        icon: child.icon ? getImageUrl(child.icon) : '',
        count: counts.get(child.id) || 0,
        parentId: group.id,
      })
    }
  }
  return out
})

const activeCategoryName = computed(() => {
  if (selectedCategory.value === null) return t('products.allCategories')
  return panelCategories.value.find((c) => c.id === selectedCategory.value)?.name || t('products.allCategories')
})

// ==================== 筛选 ====================
const filteredProducts = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase()
  return allProducts.value.filter((p) => {
    if (selectedCategory.value !== null && Number(p.category_id) !== selectedCategory.value) return false
    if (!keyword) return true
    const title = getLocalizedText(p?.title).toLowerCase()
    const desc = getLocalizedText(p?.description).toLowerCase()
    return title.includes(keyword) || desc.includes(keyword) || String(p?.slug || '').toLowerCase().includes(keyword)
  })
})

const resetFilters = () => {
  searchQuery.value = ''
  if (selectedCategory.value !== null) {
    onSelectCategory(null)
  }
}

// ==================== 分类切换 + 路由同步 ====================
const onSelectCategory = (categoryId: number | null) => {
  selectedCategory.value = categoryId
  if (initializing) return
  if (categoryId === null) {
    if (route.name === 'category-products') {
      void router.replace({ name: 'products' })
    }
    return
  }
  const matched = allCategories.value.find((c) => c.id === categoryId)
  if (matched?.slug && route.params.slug !== matched.slug) {
    void router.replace({ name: 'category-products', params: { slug: matched.slug } })
  }
}

const syncSelectedCategoryFromRoute = () => {
  if (route.name !== 'category-products') {
    if (selectedCategory.value !== null) selectedCategory.value = null
    return
  }
  const slugParam = route.params.slug as string | undefined
  if (!slugParam || allCategories.value.length === 0) return
  const matched = allCategories.value.find((c) => c.slug === slugParam)
  if (matched && selectedCategory.value !== matched.id) {
    selectedCategory.value = matched.id
  }
}

// ==================== Reveal 入场动画 ====================
const armReveal = () => {
  cleanupReveal?.()
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
}

// 搜索 / 分类切换后重新编排错峰入场
watch([searchQuery, selectedCategory], () => {
  if (initializing || loading.value) return
  armReveal()
})

watch(
  () => route.params.slug,
  () => {
    if (initializing) return
    if (allCategories.value.length === 0) return
    syncSelectedCategoryFromRoute()
  },
)

watch(
  () => route.name,
  (name) => {
    if (initializing) return
    if (name === 'products' && selectedCategory.value !== null) {
      selectedCategory.value = null
    }
  },
)

// ==================== SEO ====================
const seoCategoryName = computed(() => {
  if (selectedCategory.value === null) return ''
  const cat = allCategories.value.find((c) => c.id === selectedCategory.value)
  return cat ? getLocalizedText(cat.name) : ''
})
usePageSeo({
  canonicalPath: () => route.path,
  title: () => {
    if (route.name === 'category-products') {
      return seoCategoryName.value || t('nav.products')
    }
    return t('nav.products')
  },
})

onMounted(async () => {
  try {
    const [prodRes, catRes] = await Promise.all([
      productAPI.list({ page: 1, page_size: 50 }),
      categoryAPI.list(),
    ])
    allProducts.value = prodRes.data.data || []
    allCategories.value = catRes.data.data || []
    syncSelectedCategoryFromRoute()
  } catch (err) {
    console.error('Failed to load products:', err)
  } finally {
    loading.value = false
    initializing = false
    armReveal()
  }
})

onUnmounted(() => {
  cleanupReveal?.()
})
</script>
