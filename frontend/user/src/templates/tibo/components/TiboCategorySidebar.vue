<template>
  <aside class="store-sidebar" :aria-label="t('products.categories')">
    <span class="store-sidebar-label">{{ t('tibo.home.browse') }}</span>
    <div class="store-categories" role="group" :aria-label="t('products.categories')">
      <button type="button" :aria-pressed="selectedCategory === null" @click="emit('select', null)">
        <LayoutGrid class="cat-icon" />
        <span class="cat-text">{{ t('products.allCategories') }}</span>
      </button>
      <button
        v-for="cat in flatCategories"
        :key="cat.id"
        type="button"
        :aria-pressed="selectedCategory === cat.id"
        :style="cat.parentId ? { paddingLeft: '22px' } : undefined"
        @click="emit('select', cat.id)"
      >
        <img v-if="cat.icon" :src="getImageUrl(cat.icon)" alt="" class="cat-logo" loading="lazy" />
        <Sparkles v-else-if="!cat.parentId" class="cat-icon" />
        <Minus v-else class="cat-icon" />
        <span class="cat-text">{{ cat.name }}</span>
      </button>
    </div>
    <div class="store-sidebar-links">
      <RouterLink :to="ordersPath">
        <History class="cat-icon" />
        <span>{{ t('tibo.home.myOrders') }}</span>
      </RouterLink>
      <a v-if="supportHref" :href="supportHref" target="_blank" rel="noopener noreferrer">
        <CircleHelp class="cat-icon" />
        <span>{{ t('tibo.nav.support') }}</span>
      </a>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { CircleHelp, History, LayoutGrid, Minus, Sparkles } from 'lucide-vue-next'
import { useAppStore } from '../../../stores/app'
import { useUserAuthStore } from '../../../stores/userAuth'
import { getImageUrl } from '../../../utils/image'
import { useLocalized } from '../../../composables/useProduct'
import type { CategoryGroup, PublicCategory } from '../../../utils/category'

const props = defineProps<{
  categoryGroups: CategoryGroup[]
  selectedCategory: number | null
}>()

const emit = defineEmits<{
  select: [categoryId: number | null]
}>()

const { t } = useI18n()
const appStore = useAppStore()
const userAuthStore = useUserAuthStore()
const { getLocalizedText } = useLocalized()

interface FlatCategory {
  id: number
  name: string
  icon?: string | null
  parentId: number
}

const flatCategories = computed<FlatCategory[]>(() => {
  const out: FlatCategory[] = []
  for (const group of props.categoryGroups) {
    out.push({ id: group.id, name: getLocalizedText(group.name), icon: group.icon, parentId: 0 })
    for (const child of group.children as PublicCategory[]) {
      out.push({ id: child.id, name: getLocalizedText(child.name), icon: child.icon, parentId: group.id })
    }
  }
  return out
})

const ordersPath = computed(() => (userAuthStore.isAuthenticated ? '/me/orders' : '/guest/orders'))

const contact = computed(() => appStore.config?.contact as { telegram?: string; email?: string; support_url?: string } | undefined)
const supportHref = computed(() => {
  const c = contact.value
  if (c?.support_url) return c.support_url
  if (c?.telegram) return c.telegram
  if (c?.email) return `mailto:${c.email}`
  return ''
})
</script>

<style scoped>
.cat-icon {
  width: 16px;
  height: 16px;
  flex: none;
}

.cat-logo {
  width: 16px;
  height: 16px;
  flex: none;
  border-radius: 4px;
  object-fit: cover;
}

.cat-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.store-sidebar-links a {
  color: var(--muted-foreground);
}

.store-sidebar-links a:hover {
  color: var(--foreground);
}
</style>
