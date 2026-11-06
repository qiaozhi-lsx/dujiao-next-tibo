<template>
  <div ref="rootEl" class="qz-container">
    <!-- 引导区 -->
    <section class="qz-hero is-home">
      <div class="qz-hero-copy">
        <span class="qz-eyebrow">{{ t('qiaozhi.hero.eyebrow') }}</span>
        <h1 class="qz-hero-title">
          <span>{{ t('qiaozhi.hero.titleB') }}</span>
        </h1>
        <p>{{ t('qiaozhi.hero.subtitle') }}</p>
      </div>
      <p class="qz-shortlink">{{ t('qiaozhi.hero.shortlink') }} <em>{{ siteDomain }}</em></p>
      <div class="qz-hero-actions">
        <QzSearch
          v-model="searchQuery"
          :items="searchItems"
          :label="t('qiaozhi.search.label')"
          :trigger-label="t('qiaozhi.search.trigger')"
          :placeholder="t('qiaozhi.hero.searchPlaceholder')"
          :empty-title="(q) => t('qiaozhi.search.emptyTitle', { query: q })"
          :empty-hint="t('qiaozhi.search.emptyHint')"
          :results-text="(n) => t('qiaozhi.search.results', { count: n })"
          :group-label="t('qiaozhi.search.groupProducts')"
          :clear-label="t('qiaozhi.search.clear')"
          @select="onSearchSelect"
        />
        <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm" @click="shareSite">
          <Share2 />
          {{ shared ? t('qiaozhi.hero.shared') : t('qiaozhi.hero.share') }}
        </button>
      </div>
    </section>

    <!-- 为你精选 -->
    <section v-if="featuredProducts.length" class="qz-featured">
      <div class="qz-section-head">
        <div>
          <span class="qz-eyebrow">{{ t('qiaozhi.home.featuredEyebrow') }}</span>
          <h2>{{ t('qiaozhi.featured.title') }}</h2>
        </div>
        <span>{{ featuredProducts.length + (toolsCard ? 1 : 0) }} {{ t('qiaozhi.featured.unit') }}</span>
      </div>
      <div class="qz-featured-grid" :data-count="Math.min(featuredProducts.length, 5)">
        <QzFeaturedCard v-for="product in featuredProducts" :key="product.id" :product="product" />
        <QzFeaturedCard v-if="toolsCard" :product="toolsCard.product" :tools="toolsCard.tools" />
      </div>
    </section>

    <!-- 全部商品目录 -->
    <section class="qz-catalog">
      <QzCategoryPanel
        :categories="panelCategories"
        :selected-category="selectedCategory"
        :total-count="allProducts.length"
        @select="selectedCategory = $event"
      />
      <div>
        <div class="qz-section-head">
          <div>
            <span class="qz-eyebrow">{{ t('qiaozhi.home.storeEyebrow') }}</span>
            <h2>{{ activeCategoryName }}</h2>
          </div>
          <span>{{ filteredProducts.length }} {{ t('qiaozhi.catalog.unit') }}</span>
        </div>

        <div v-if="loading" class="qz-grid">
          <div v-for="i in 6" :key="i" class="qz-skeleton-card" aria-hidden="true">
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
          <b>{{ searchQuery ? t('products.emptyFiltered') : t('products.empty') }}</b>
          <span v-if="searchQuery || selectedCategory">
            <button type="button" class="qz-btn qz-btn--primary" @click="resetFilters">{{ t('products.clearFilters') }}</button>
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
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { PackageOpen, Share2 } from 'lucide-vue-next'
import { categoryAPI, productAPI } from '../../api'
import { useAppStore } from '../../stores/app'
import { useLocalized } from '../../composables/useProduct'
import { usePageSeo } from '../../composables/usePageSeo'
import { useAnnouncement, type HomeAnnouncement } from '../../composables/useAnnouncement'
import { buildCategoryGroups } from '../../utils/category'
import { getImageUrl } from '../../utils/image'
import AnnouncementModal from '../../components/AnnouncementModal.vue'
import QzFeaturedCard from './components/QzFeaturedCard.vue'
import QzProductCard from './components/QzProductCard.vue'
import QzCategoryPanel, { type PanelCategory } from './components/QzCategoryPanel.vue'
import QzSearch from './components/QzSearch.vue'
import type { SearchableItem } from './utils/searchRank'
import { pickFeatured } from './utils/featured'
import { setupReveal } from './utils/reveal'
import { useQiaozhiPrice } from './utils/price'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const { getLocalizedText } = useLocalized()
const { format: formatPrice } = useQiaozhiPrice()

const rootEl = ref<HTMLElement | null>(null)
const loading = ref(true)
const allProducts = ref<any[]>([])
const allCategories = ref<any[]>([])
const selectedCategory = ref<number | null>(null)
const searchQuery = ref('')
const shared = ref(false)

// ==================== 展开式搜索 ====================
const searchItems = computed<SearchableItem[]>(() => allProducts.value.map((p: any) => ({
  id: p.id,
  slug: String(p.slug || ''),
  title: getLocalizedText(p.title),
  meta: formatPrice(p.price_amount),
  keywords: [String(p.slug || ''), String(p.id ?? '')],
})))

const onSearchSelect = (item: SearchableItem) => {
  const slug = String(item.slug || '')
  if (slug) void router.push(`/products/${slug}`)
}

let cleanupReveal: (() => void) | null = null
let shareTimer: ReturnType<typeof setTimeout> | null = null

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
// 后台给商品打 featured / 精选 标签即入选（前 5）；一个都没打时沿用 slug 品牌正则，再回退前 4 做默认卡。
const featuredProducts = computed(() => pickFeatured(allProducts.value))

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
// 内联首屏配置不含 announcement，完整 /public/config 稍后才到；监听配置而非在 onMounted 里只调一次。
// 公告非空即置位，后台刷新再次替换 config 时不会重复弹。
let announcementHandled = false
watch(
  () => appStore.config?.announcement,
  (announcement) => {
    if (announcementHandled || !announcement) return
    announcementHandled = true
    showAnnouncementIfNeeded()
  },
  { immediate: true },
)

usePageSeo({ canonicalPath: () => '/' })

const armReveal = () => {
  cleanupReveal?.()
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
}

// 搜索 / 分类切换后重新武装：新渲染的卡片带 is-pending，不重新观察会一直停在 opacity: 0
watch([searchQuery, selectedCategory], () => {
  armReveal()
})

onMounted(async () => {
  try {
    const [prodRes, catRes] = await Promise.all([
      productAPI.list({ page: 1, page_size: 20 }),
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
})

onUnmounted(() => {
  cleanupReveal?.()
  if (shareTimer) clearTimeout(shareTimer)
})
</script>

<style scoped>
/* 首页专属版式：hero、精选区与商品网格。卡片 / 面板 / 搜索 / 分类面板各自在组件内取样式；
   `.qz-catalog`（侧栏 + 列表两栏）与骨架屏仍在 legacy.css。关键帧 qz-text-in 来自 legacy.css。 */

/* ---------- Hero ---------- */
.qz-hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--qz-space-8);
  margin-bottom: var(--qz-space-5);
}
.qz-hero-copy { min-width: 0; }
.qz-hero-copy p {
  max-width: 650px;
  margin: 0;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
}
.qz-hero-title {
  margin: var(--qz-space-1) 0 var(--qz-space-2);
  color: var(--qz-text);
  font-size: var(--qz-text-3xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
}
.qz-hero-title > span { display: inline-block; }
/* 入场：eyebrow / 标题 / 说明 / 操作区依次淡入 */
.qz-hero .qz-eyebrow,
.qz-hero-title > span,
.qz-shortlink,
.qz-hero-copy p,
.qz-hero-actions {
  animation: qz-text-in var(--qz-duration-considered) var(--qz-ease-enter) backwards;
}
.qz-hero-copy p { animation-delay: var(--qz-duration-instant); }
.qz-hero-actions { animation-delay: var(--qz-duration-fast); }

/* 手机端用「本站短链」替代标题文案 */
.qz-shortlink {
  display: none;
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-xl);
  font-weight: 700;
  line-height: var(--qz-leading-body);
  letter-spacing: var(--qz-tracking-display);
  white-space: nowrap;
}
.qz-shortlink em { color: var(--qz-accent-strong); font-style: italic; }

/* 搜索（QzSearch，sm 高）与分享（.qz-btn--secondary.qz-btn--sm）同排同高 */
.qz-hero-actions {
  display: flex;
  flex: none;
  align-items: center;
  gap: var(--qz-space-2);
}

/* ---------- 为你精选 ---------- */
.qz-featured { margin: var(--qz-space-1) 0 var(--qz-space-8); }
.qz-featured-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--qz-space-4);
}
/* 1 / 2 / 4 张精选卡时两列更整齐（data-count 在网格元素上） */
.qz-featured-grid:is([data-count="1"], [data-count="2"], [data-count="4"]) {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
.qz-featured-grid > .is-wide-full { grid-column: 1 / -1; }

/* ---------- 商品网格 ---------- */
.qz-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--qz-space-4);
}

@media (max-width: 1023.98px) {
  .qz-featured-grid,
  .qz-featured-grid:is([data-count="1"], [data-count="2"], [data-count="4"]) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .qz-hero {
    display: grid;
    gap: var(--qz-space-4);
    margin-bottom: var(--qz-space-4);
  }
  .qz-hero-copy { display: none; }
  .qz-shortlink { display: block; }
  .qz-hero-actions {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: var(--qz-space-2);
  }

  /* 精选卡横向滑动，吸附到每张卡 */
  .qz-featured { margin: 0 calc(-1 * var(--qz-space-3)) var(--qz-space-5); }
  .qz-featured .qz-section-head { padding: 0 var(--qz-space-3); }
  .qz-featured-grid,
  .qz-featured-grid:is([data-count="1"], [data-count="2"], [data-count="4"]) {
    display: flex;
    gap: var(--qz-space-3);
    overflow-x: auto;
    padding: var(--qz-space-1) var(--qz-space-3) var(--qz-space-4);
    scroll-snap-type: x mandatory;
    scroll-padding: 0 var(--qz-space-3);
    overscroll-behavior-x: contain;
    scrollbar-width: none;
  }
  .qz-featured-grid::-webkit-scrollbar { display: none; }
  .qz-featured-grid > .qz-fcard {
    flex: 0 0 84%;
    scroll-snap-align: start;
  }

  .qz-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--qz-space-2);
  }
}
</style>
