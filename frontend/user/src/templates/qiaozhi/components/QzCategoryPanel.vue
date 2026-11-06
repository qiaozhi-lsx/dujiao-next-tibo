<template>
  <!-- 桌面：侧栏面板 -->
  <aside class="qz-panel qz-only-desktop">
    <header>
      <span class="qz-eyebrow">CATEGORIES</span>
      <h2>{{ t('qiaozhi.catalog.title') }}</h2>
    </header>
    <div class="qz-cat-list">
      <button type="button" class="qz-cat-item" :class="{ 'is-active': selectedCategory === null }" :aria-pressed="selectedCategory === null" @click="emit('select', null)">
        <span class="qz-cat-logo hot"><Star /></span>
        <span>{{ t('products.allCategories') }}</span>
        <b>{{ totalCount }}</b>
      </button>
      <button
        v-for="cat in visibleCategories"
        :key="cat.id"
        type="button"
        class="qz-cat-item"
        :class="{ 'is-active': selectedCategory === cat.id, 'is-child': cat.parentId > 0 }"
        :aria-pressed="selectedCategory === cat.id"
        @click="emit('select', cat.id)"
      >
        <span class="qz-cat-logo">
          <span class="qz-brand-icon">
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
  <button type="button" class="qz-cat-btn qz-only-mobile" @click="drawerOpen = true">
    <FolderOpen />
    <span>{{ activeCategoryName }}</span>
    <ChevronDown />
  </button>

  <Teleport to="body">
    <Transition name="qz-drawer">
      <div v-if="drawerOpen" class="qz-drawer">
        <div class="qz-drawer-backdrop" @click="drawerOpen = false"></div>
        <div class="qz-drawer-panel">
          <header>
            <div>
              <span class="qz-eyebrow">CATEGORIES</span>
              <h2>{{ t('qiaozhi.catalog.title') }}</h2>
            </div>
            <button type="button" class="qz-drawer-close" :aria-label="t('common.cancel')" @click="drawerOpen = false"><X /></button>
          </header>
          <div class="qz-cat-list">
            <button type="button" class="qz-cat-item" :class="{ 'is-active': selectedCategory === null }" :aria-pressed="selectedCategory === null" @click="pick(null)">
              <span class="qz-cat-logo hot"><Star /></span>
              <span>{{ t('products.allCategories') }}</span>
              <b>{{ totalCount }}</b>
            </button>
            <button
              v-for="cat in visibleCategories"
              :key="cat.id"
              type="button"
              class="qz-cat-item"
              :class="{ 'is-active': selectedCategory === cat.id, 'is-child': cat.parentId > 0 }"
              :aria-pressed="selectedCategory === cat.id"
              @click="pick(cat.id)"
            >
              <span class="qz-cat-logo">
                <span class="qz-brand-icon">
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
