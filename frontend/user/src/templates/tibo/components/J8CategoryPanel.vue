<template>
  <!-- 桌面：侧栏面板 -->
  <aside class="j8-panel j8-only-desktop">
    <header>
      <span class="j8-eyebrow">CATEGORIES</span>
      <h2>{{ t('tibo.catalog.title') }}</h2>
    </header>
    <div class="j8-cat-list">
      <button type="button" class="j8-cat-item" :class="{ 'is-active': selectedCategory === null }" @click="emit('select', null)">
        <span class="j8-cat-logo hot"><Star /></span>
        <span>{{ t('products.allCategories') }}</span>
        <b>{{ totalCount }}</b>
      </button>
      <button
        v-for="cat in visibleCategories"
        :key="cat.id"
        type="button"
        class="j8-cat-item"
        :class="{ 'is-active': selectedCategory === cat.id, 'is-child': cat.parentId > 0 }"
        @click="emit('select', cat.id)"
      >
        <span class="j8-cat-logo">
          <span class="j8-brand-icon">
            <img v-if="cat.icon" :src="cat.icon" alt="" loading="lazy" />
            <b v-else>{{ cat.name.slice(0, 1) }}</b>
          </span>
        </span>
        <span>{{ cat.name }}</span>
        <b>{{ cat.count }}</b>
      </button>
    </div>
  </aside>

  <!-- 移动端：分类抽屉 -->
  <button type="button" class="j8-cat-btn j8-only-mobile" @click="drawerOpen = true">
    <FolderOpen />
    <span>{{ activeCategoryName }}</span>
    <ChevronDown />
  </button>

  <Teleport to="body">
    <Transition name="j8-drawer">
      <div v-if="drawerOpen" class="j8-drawer">
        <div class="j8-drawer-backdrop" @click="drawerOpen = false"></div>
        <div class="j8-drawer-panel">
          <header>
            <div>
              <span class="j8-eyebrow">CATEGORIES</span>
              <h2>{{ t('tibo.catalog.title') }}</h2>
            </div>
            <button type="button" class="j8-drawer-close" :aria-label="t('common.cancel')" @click="drawerOpen = false"><X /></button>
          </header>
          <div class="j8-cat-list">
            <button type="button" class="j8-cat-item" :class="{ 'is-active': selectedCategory === null }" @click="pick(null)">
              <span class="j8-cat-logo hot"><Star /></span>
              <span>{{ t('products.allCategories') }}</span>
              <b>{{ totalCount }}</b>
            </button>
            <button
              v-for="cat in visibleCategories"
              :key="cat.id"
              type="button"
              class="j8-cat-item"
              :class="{ 'is-active': selectedCategory === cat.id, 'is-child': cat.parentId > 0 }"
              @click="pick(cat.id)"
            >
              <span class="j8-cat-logo">
                <span class="j8-brand-icon">
                  <img v-if="cat.icon" :src="cat.icon" alt="" loading="lazy" />
                  <b v-else>{{ cat.name.slice(0, 1) }}</b>
                </span>
              </span>
              <span>{{ cat.name }}</span>
              <b>{{ cat.count }}</b>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronDown, FolderOpen, Star, X } from 'lucide-vue-next'

export interface PanelCategory {
  id: number
  name: string
  icon: string
  count: number
  parentId: number
}

const props = defineProps<{
  categories: PanelCategory[]
  selectedCategory: number | null
  totalCount: number
}>()

const emit = defineEmits<{
  select: [categoryId: number | null]
}>()

const { t } = useI18n()
const drawerOpen = ref(false)

const visibleCategories = computed(() => props.categories.filter((c) => c.count > 0 || true))

const activeCategoryName = computed(() => {
  if (props.selectedCategory === null) return t('products.allCategories')
  return props.categories.find((c) => c.id === props.selectedCategory)?.name || t('products.allCategories')
})

const pick = (id: number | null) => {
  emit('select', id)
  drawerOpen.value = false
}
</script>
