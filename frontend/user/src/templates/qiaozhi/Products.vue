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
        <label class="qz-field-control qz-search" role="search">
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
        <div class="qz-section-head">
          <div>
            <span class="qz-eyebrow">{{ t('qiaozhi.products.catalog.eyebrow') }}</span>
            <h2>{{ activeCategoryName }}</h2>
          </div>
          <span>{{ filteredProducts.length }} {{ t('qiaozhi.products.catalog.unit') }}</span>
        </div>

        <div v-if="loading" class="qz-grid">
          <div v-for="i in 8" :key="i" class="qz-skeleton-card" aria-hidden="true">
            <span class="qz-skeleton-art"><span class="qz-skeleton"></span></span>
            <span class="qz-skeleton qz-skeleton-title"></span>
            <span class="qz-skeleton qz-skeleton-spec"></span>
            <div class="qz-skeleton-bottom">
              <span class="qz-skeleton qz-skeleton-price"></span>
              <span class="qz-skeleton qz-skeleton-action"></span>
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
            <button type="button" class="qz-btn qz-btn--primary" @click="resetFilters">{{ t('products.clearFilters') }}</button>
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

<style scoped>
/* 列表页专属版式：hero（页 h1 用 2xl）、搜索框（.qz-field-control 上的宽度与清除钮）、目录两栏
   （.qz-catalog：分类面板 + 列表）、商品网格与加载骨架。骨架占位块是 base.css 的 .qz-skeleton（qz-pulse），
   这里只给卡片外形与各占位块的尺寸。标题入场关键帧 qz-text-in 在本块底部（Home.vue 同款）。 */

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
  font-size: var(--qz-text-2xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
}
.qz-hero-title > span {
  display: inline-block;
  animation: qz-text-in var(--qz-duration-considered) var(--qz-ease-enter) backwards;
}
.qz-hero-actions {
  display: flex;
  flex: none;
  align-items: center;
  gap: var(--qz-space-2);
}

/* ---------- 搜索框：基类给高度 / 边框 / 焦点，这里只定宽与清除钮（带上基类名靠特指度覆盖 width:100%） ---------- */
.qz-field-control.qz-search { width: min(380px, 35vw); }
.qz-search input::-webkit-search-cancel-button,
.qz-search input::-webkit-search-decoration,
.qz-search input::-webkit-search-results-button,
.qz-search input::-webkit-search-results-decoration {
  appearance: none;
  display: none;
}
.qz-search-clear {
  display: grid;
  flex: none;
  place-items: center;
  width: 28px;
  height: 28px;
  margin-right: calc(-1 * var(--qz-space-2));
  padding: 0;
  border: 0;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-surface-muted);
  color: var(--qz-text-muted);
  cursor: pointer;
  transition: color var(--qz-duration-fast) var(--qz-ease-standard), background-color var(--qz-duration-fast) var(--qz-ease-standard);
}
.qz-search .qz-search-clear svg { width: 14px; height: 14px; color: inherit; }
@media (hover: hover) and (pointer: fine) {
  .qz-search-clear:hover { color: var(--qz-text); background: var(--qz-accent-subtle); }
}
.qz-search-clear:active { color: var(--qz-text); background: var(--qz-accent-subtle); }

/* ---------- 目录两栏：分类面板 + 列表（≤1120 窄面板，≤760 单列、面板收成抽屉） ---------- */
.qz-catalog {
  display: grid;
  grid-template-columns: 250px minmax(0, 1fr);
  gap: var(--qz-space-4);
  align-items: start;
}
.qz-catalog > div { min-width: 0; }

/* ---------- 商品网格 ---------- */
.qz-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--qz-space-4);
}

/* ---------- 加载骨架：商品卡外形 + base .qz-skeleton 占位块（艺术区 / 标题 / 规格 / 价格 / 按钮） ---------- */
.qz-skeleton-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 280px;
  padding: var(--qz-space-5);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface);
}
.qz-skeleton-card .qz-skeleton { border-radius: var(--qz-radius-xs); }
.qz-skeleton-art { display: grid; place-items: center; height: 160px; margin-top: 6px; }
.qz-skeleton-art > .qz-skeleton { width: 54%; height: 92px; opacity: .6; border-radius: 14px; }
.qz-skeleton-title { width: 48%; height: 24px; }
.qz-skeleton-spec { width: 62%; height: 12px; margin-top: 2px; }
.qz-skeleton-bottom { display: flex; align-items: center; justify-content: space-between; margin-top: auto; }
.qz-skeleton-price { width: 28%; height: 22px; }
.qz-skeleton-card .qz-skeleton-action { width: 86px; height: 36px; border-radius: var(--qz-radius-pill); }

@media (max-width: 1120px) {
  .qz-catalog { grid-template-columns: 220px minmax(0, 1fr); }
}

@media (max-width: 760px) {
  .qz-catalog { display: block; }
  .qz-skeleton-card { min-height: 220px; padding: var(--qz-space-4); }
  .qz-skeleton-art { height: 110px; }
  .qz-skeleton-art > .qz-skeleton { height: 70px; }

  .qz-hero {
    display: grid;
    gap: var(--qz-space-4);
    margin-bottom: var(--qz-space-4);
  }
  .qz-hero-actions { display: grid; grid-template-columns: minmax(0, 1fr); }
  .qz-field-control.qz-search { width: 100%; }
  .qz-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--qz-space-2);
  }
}

/* 标题入场：上浮 + 去模糊 */
@keyframes qz-text-in {
  0% { opacity: 0; transform: translateY(14px); filter: blur(4px); }
  100% { opacity: 1; transform: none; filter: none; }
}
</style>
