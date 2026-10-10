<template>
  <div>
    <!-- 引导区 -->
    <section class="store-intro page-width reveal-in">
      <div>
        <h1>{{ t('tibo.hero.titleA') }}<span>{{ t('tibo.hero.titleB') }}</span></h1>
        <p>{{ t('tibo.hero.subtitle') }}</p>
      </div>
      <div v-if="supportHref" class="store-intro-support">
        <span>{{ t('tibo.intro.supportQuestion') }}</span>
        <a :href="supportHref" target="_blank" rel="noopener noreferrer" class="text-button">
          <LifeBuoy class="intro-icon" />
          {{ t('tibo.intro.supportAction') }}
          <ArrowRight class="intro-icon" />
        </a>
      </div>
    </section>

    <!-- 分类 + 商品 -->
    <div class="store-home page-width">
      <TiboCategorySidebar
        :category-groups="categoryGroups"
        :selected-category="selectedCategory"
        @select="onSelectCategory"
      />

      <div class="store-content">
        <div class="store-toolbar">
          <h2>{{ t('tibo.home.featuredTitle') }}</h2>
          <label class="store-search">
            <Search class="search-icon" />
            <input
              v-model="searchQuery"
              type="search"
              :placeholder="t('tibo.home.searchPlaceholder')"
              :aria-label="t('products.searchLabel')"
            />
          </label>
        </div>

        <!-- 骨架 -->
        <div v-if="loading" class="store-featured">
          <div v-for="i in 4" :key="i" class="store-product tibo-skeleton"></div>
        </div>

        <!-- 精选大卡 -->
        <div v-else-if="featuredProducts.length" class="store-featured store-product-group">
          <TiboProductCard v-for="product in featuredProducts" :key="product.id" :product="product" />
        </div>

        <!-- 空态 -->
        <div v-else class="tibo-empty">
          <PackageOpen class="empty-icon" />
          <p>{{ searchQuery || selectedCategory ? t('products.emptyFiltered') : t('products.empty') }}</p>
          <button v-if="searchQuery || selectedCategory" type="button" class="text-button" @click="resetFilters">
            {{ t('products.clearFilters') }}
          </button>
        </div>

        <!-- 更多商品紧凑行 -->
        <div v-if="!loading && compactProducts.length" class="store-compact-grid store-product-group">
          <TiboCompactRow v-for="product in compactProducts" :key="product.id" :product="product" />
        </div>

        <!-- 底部说明 -->
        <p class="store-order-note">
          <ReceiptText class="note-icon" />
          <span>{{ t('tibo.home.orderNote') }}</span>
          <RouterLink :to="ordersPath" class="text-button">{{ t('tibo.home.orderNoteLink') }}</RouterLink>
        </p>
      </div>
    </div>

    <AnnouncementModal
      v-if="activeAnnouncement"
      :announcement="activeAnnouncement"
      :visible="announcementVisible"
      @update:visible="announcementVisible = $event"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowRight, LifeBuoy, PackageOpen, ReceiptText, Search } from 'lucide-vue-next'
import { useAppStore } from '../../stores/app'
import { useUserAuthStore } from '../../stores/userAuth'
import { useProductList } from '../../composables/useProductList'
import { usePageSeo } from '../../composables/usePageSeo'
import { useAnnouncement, type HomeAnnouncement } from '../../composables/useAnnouncement'
import AnnouncementModal from '../../components/AnnouncementModal.vue'
import TiboCategorySidebar from './components/TiboCategorySidebar.vue'
import TiboProductCard from './components/TiboProductCard.vue'
import TiboCompactRow from './components/TiboCompactRow.vue'

const { t } = useI18n()
const appStore = useAppStore()
const userAuthStore = useUserAuthStore()

// 卡片模式：分类/搜索均在首页内过滤，不跳转 /categories 路由
const {
  loading,
  products,
  selectedCategory,
  searchQuery,
  categoryGroups,
  selectCategory,
  clearSearch,
  initialize,
  cleanup,
} = useProductList({ pageSize: 30, homeRouteName: 'home', categoryRouteName: 'home' })

// 前 6 个商品用大卡，其余进紧凑行
const featuredProducts = computed(() => products.value.slice(0, 6))
const compactProducts = computed(() => products.value.slice(6))

const onSelectCategory = (categoryId: number | null) => {
  selectCategory(categoryId)
}

const resetFilters = () => {
  clearSearch()
  selectCategory(null)
}

const ordersPath = computed(() => (userAuthStore.isAuthenticated ? '/me/orders' : '/guest/orders'))

const contact = computed(() => appStore.config?.contact as { telegram?: string; email?: string; support_url?: string } | undefined)
const supportHref = computed(() => {
  const c = contact.value
  if (c?.support_url) return c.support_url
  if (c?.telegram) return c.telegram
  if (c?.email) return `mailto:${c.email}`
  return ''
})

// ==================== 公告弹窗（版本化可关闭） ====================
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

onMounted(async () => {
  await appStore.loadConfig()
  await initialize()
  showAnnouncementIfNeeded()
})

onUnmounted(() => cleanup())
</script>

<style scoped>
.intro-icon {
  width: 13px;
  height: 13px;
}

.search-icon {
  width: 15px;
  height: 15px;
}

.note-icon {
  width: 14px;
  height: 14px;
}

.empty-icon {
  width: 40px;
  height: 40px;
  opacity: 0.55;
}

.tibo-skeleton {
  min-height: 340px;
  background: linear-gradient(120deg, var(--card), var(--muted), var(--card));
  background-size: 200% 100%;
  animation: tibo-skeleton 1.6s ease-in-out infinite;
}

.tibo-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 56px 0;
  color: var(--muted-foreground);
  font-size: 13px;
}

@keyframes tibo-skeleton {
  0% {
    background-position: 200% 0;
  }

  100% {
    background-position: -200% 0;
  }
}
</style>
