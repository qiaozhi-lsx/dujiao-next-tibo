<template>
  <div ref="rootEl" class="j8-container">
    <!-- 引导区 -->
    <section class="j8-hero is-home">
      <div class="j8-hero-copy">
        <span class="j8-eyebrow">{{ t('tibo.hero.eyebrow') }}</span>
        <h1 class="j8-hero-title">
          <span>{{ brandName }}</span>
          <span class="is-dim">{{ t('tibo.hero.titleB') }}</span>
        </h1>
        <p>{{ t('tibo.hero.subtitle') }}</p>
      </div>
      <p class="j8-shortlink">{{ t('tibo.hero.shortlink') }} <em>{{ siteDomain }}</em></p>
      <div class="j8-hero-actions">
        <label class="j8-search" role="search">
          <Search />
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="t('tibo.hero.searchPlaceholder')"
            :aria-label="t('products.searchLabel')"
          />
          <button v-if="searchQuery" type="button" class="j8-search-clear" :aria-label="t('products.clearFilters')" @click="searchQuery = ''">
            <X />
          </button>
        </label>
        <button type="button" class="j8-share" @click="shareSite">
          <Share2 />
          {{ shared ? t('tibo.hero.shared') : t('tibo.hero.share') }}
        </button>
      </div>
    </section>

    <!-- 为你精选 -->
    <section v-if="featuredProducts.length" class="j8-featured">
      <div class="j8-section-heading">
        <div>
          <span class="j8-eyebrow">FEATURED</span>
          <h2>{{ t('tibo.featured.title') }}</h2>
        </div>
        <span>{{ featuredProducts.length + (toolsCard ? 1 : 0) }} {{ t('tibo.featured.unit') }}</span>
      </div>
      <div class="j8-featured-grid" :data-count="Math.min(featuredProducts.length, 5)">
        <J8FeaturedCard v-for="product in featuredProducts" :key="product.id" :product="product" />
        <J8FeaturedCard v-if="toolsCard" :product="toolsCard.product" :tools="toolsCard.tools" />
      </div>
    </section>

    <!-- 全部商品目录 -->
    <section class="j8-catalog">
      <J8CategoryPanel
        :categories="panelCategories"
        :selected-category="selectedCategory"
        :total-count="allProducts.length"
        @select="selectedCategory = $event"
      />
      <div>
        <div class="j8-section-heading">
          <div>
            <span class="j8-eyebrow">STORE</span>
            <h2>{{ activeCategoryName }}</h2>
          </div>
          <span>{{ filteredProducts.length }} {{ t('tibo.catalog.unit') }}</span>
        </div>

        <div v-if="loading" class="j8-grid">
          <div v-for="i in 6" :key="i" class="j8-skeleton-card"></div>
        </div>
        <div v-else-if="filteredProducts.length" class="j8-grid">
          <J8ProductCard v-for="product in filteredProducts" :key="product.id" :product="product" />
        </div>
        <div v-else class="j8-empty">
          <PackageOpen />
          <b>{{ searchQuery ? t('products.emptyFiltered') : t('products.empty') }}</b>
          <span v-if="searchQuery || selectedCategory">
            <button type="button" class="j8-btn-primary" @click="resetFilters">{{ t('products.clearFilters') }}</button>
          </span>
        </div>
      </div>
    </section>

    <AnnouncementModal
      v-if="activeAnnouncement"
      :announcement="activeAnnouncement"
      :visible="announcementVisible"
      @update:visible="announcementVisible = $event"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { PackageOpen, Search, Share2, X } from 'lucide-vue-next'
import { categoryAPI, productAPI } from '../../api'
import { useAppStore } from '../../stores/app'
import { useLocalized } from '../../composables/useProduct'
import { usePageSeo } from '../../composables/usePageSeo'
import { useAnnouncement, type HomeAnnouncement } from '../../composables/useAnnouncement'
import { buildCategoryGroups } from '../../utils/category'
import { getImageUrl } from '../../utils/image'
import AnnouncementModal from '../../components/AnnouncementModal.vue'
import J8FeaturedCard from './components/J8FeaturedCard.vue'
import J8ProductCard from './components/J8ProductCard.vue'
import J8CategoryPanel, { type PanelCategory } from './components/J8CategoryPanel.vue'
import { setupReveal } from './utils/reveal'
import { useTiboPrice } from './utils/price'

const { t } = useI18n()
const appStore = useAppStore()
const { getLocalizedText } = useLocalized()
const { format: formatPrice } = useTiboPrice()

const rootEl = ref<HTMLElement | null>(null)
const loading = ref(true)
const allProducts = ref<any[]>([])
const allCategories = ref<any[]>([])
const selectedCategory = ref<number | null>(null)
const searchQuery = ref('')
const shared = ref(false)

let cleanupReveal: (() => void) | null = null
let shareTimer: ReturnType<typeof setTimeout> | null = null

const brandName = computed(() => String(appStore.config?.brand?.site_name || '').trim() || '商店')
const siteDomain = computed(() => {
  try {
    return window.location.host.toUpperCase()
  } catch {
    return ''
  }
})

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

// ==================== 精选拆分 ====================
const FEATURED_PATTERN = /chatgpt|gpt|gemini|bard|google|^x-|x-premium|twitter|duolingo|duo|muse/i

const featuredProducts = computed(() => {
  const matched = allProducts.value.filter((p) => FEATURED_PATTERN.test(String(p.slug || '')))
  if (matched.length) return matched.slice(0, 5)
  // 没有任何匹配变体的商品时，用前 4 个产品做默认卡
  return allProducts.value.slice(0, 4)
})

const toolsCard = computed(() => {
  if (!allProducts.value.length) return null
  const featuredIds = new Set(featuredProducts.value.map((p) => p.id))
  const rest = allProducts.value.filter((p) => !featuredIds.has(p.id))
  if (!rest.length) return null
  const tools = rest.slice(0, 6).map((p) => ({
    id: p.id,
    icon: getImageUrl(p?.category?.icon || '') || getImageUrl((p?.images || [])[0] || ''),
    letter: getLocalizedText(p?.title).slice(0, 1).toUpperCase(),
    price: formatPrice(p?.price_amount),
    soldOut: p?.stock_status === 'out_of_stock' || p?.is_sold_out === true,
  }))
  // 通栏卡以第一个剩余商品为主（标题/价格），其余作为工具行
  return { product: rest[0], tools }
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
  selectedCategory.value = null
}

// ==================== 分享 ====================
const shareSite = async () => {
  try {
    await navigator.clipboard.writeText(window.location.origin)
    shared.value = true
    if (shareTimer) clearTimeout(shareTimer)
    shareTimer = setTimeout(() => (shared.value = false), 2000)
  } catch {
    /* 剪贴板不可用时忽略 */
  }
}

// ==================== 公告 ====================
const { shouldShow } = useAnnouncement()
const activeAnnouncement = ref<HomeAnnouncement | null>(null)
const announcementVisible = ref(false)
const showAnnouncementIfNeeded = () => {
  const announcement = appStore.config?.announcement as HomeAnnouncement | undefined
  if (announcement && shouldShow(announcement)) {
    activeAnnouncement.value = announcement
    announcementVisible.value = true
  }
}

usePageSeo({ canonicalPath: () => '/' })

const armReveal = () => {
  cleanupReveal?.()
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
}

onMounted(async () => {
  await appStore.loadConfig()
  try {
    const [prodRes, catRes] = await Promise.all([
      productAPI.list({ page: 1, page_size: 100 }),
      categoryAPI.list(),
    ])
    allProducts.value = prodRes.data.data || []
    allCategories.value = catRes.data.data || []
  } catch (err) {
    console.error('Failed to load catalog:', err)
  } finally {
    loading.value = false
    armReveal()
  }
  showAnnouncementIfNeeded()
})

onUnmounted(() => {
  cleanupReveal?.()
  if (shareTimer) clearTimeout(shareTimer)
})
</script>
