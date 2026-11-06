<template>
  <!-- 桌面：侧栏面板 -->
  <aside class="qz-panel qz-cat-panel qz-only-desktop">
    <header class="qz-cat-head">
      <span class="qz-eyebrow">{{ t('qiaozhi.categories.eyebrow') }}</span>
      <h2>{{ t('qiaozhi.catalog.title') }}</h2>
    </header>
    <div class="qz-cat-list">
      <button type="button" class="qz-pill qz-cat-item" :class="{ 'is-active': selectedCategory === null }" :aria-pressed="selectedCategory === null" @click="emit('select', null)">
        <span class="qz-cat-logo hot"><Star /></span>
        <span>{{ t('products.allCategories') }}</span>
        <b class="qz-badge">{{ totalCount }}</b>
      </button>
      <button
        v-for="cat in visibleCategories"
        :key="cat.id"
        type="button"
        class="qz-pill qz-cat-item"
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
        <b class="qz-badge">{{ cat.count }}</b>
      </button>
    </div>
  </aside>

  <!-- 移动端：分类抽屉触发钮 -->
  <button type="button" class="qz-btn qz-btn--secondary qz-btn--block qz-cat-btn qz-only-mobile" @click="drawerOpen = true">
    <FolderOpen />
    <span>{{ activeCategoryName }}</span>
    <ChevronDown />
  </button>

  <Teleport to="body">
    <Transition name="qz-drawer">
      <div v-if="drawerOpen" class="qz-drawer">
        <div class="qz-drawer-backdrop" @click="drawerOpen = false"></div>
        <div class="qz-drawer-panel">
          <header class="qz-cat-head">
            <div>
              <span class="qz-eyebrow">{{ t('qiaozhi.categories.eyebrow') }}</span>
              <h2>{{ t('qiaozhi.catalog.title') }}</h2>
            </div>
            <button type="button" class="qz-icon-btn qz-icon-btn--sm" :aria-label="t('common.cancel')" @click="drawerOpen = false"><X /></button>
          </header>
          <div class="qz-cat-list">
            <button type="button" class="qz-pill qz-cat-item" :class="{ 'is-active': selectedCategory === null }" :aria-pressed="selectedCategory === null" @click="pick(null)">
              <span class="qz-cat-logo hot"><Star /></span>
              <span>{{ t('products.allCategories') }}</span>
              <b class="qz-badge">{{ totalCount }}</b>
            </button>
            <button
              v-for="cat in visibleCategories"
              :key="cat.id"
              type="button"
              class="qz-pill qz-cat-item"
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
              <b class="qz-badge">{{ cat.count }}</b>
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

<style>
/* 分类面板 + 分类抽屉。
   抽屉 Teleport 到 body，不在 .qiaozhi-scope 内，所以这里不用 scoped，而是沿用 base.css 的前缀
   `:is(.qiaozhi-scope, :where(body.qiaozhi-tokens))`：桌面侧栏与抽屉共用同一套 .qz-cat-* 规则，
   body 一侧包在 :where() 里保持 (0,2,0) 特指度（见 __tests__/drawer-scope.spec.ts）。
   表面来自 .qz-panel，条目来自 .qz-pill（选中 .is-active），计数来自 .qz-badge，关闭钮是 .qz-icon-btn--sm，
   手机触发钮是 .qz-btn--secondary.qz-btn--block。 */

/* ---------- 桌面侧栏：贴顶、内边距收到 space-3 让条目的 hover 底色贴边 ----------
   覆盖基类属性的选择器都把基类写进去（.qz-panel.qz-cat-panel / .qz-pill.qz-cat-item / .qz-btn.qz-cat-btn），
   靠特指度而不是样式表加载顺序赢过 base.css。 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-panel.qz-cat-panel {
  position: sticky;
  top: calc(var(--qz-header-h) + var(--qz-space-4));
  padding: var(--qz-space-3);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-head {
  padding: var(--qz-space-2) var(--qz-space-2) var(--qz-space-3);
  border-bottom: 1px solid var(--qz-border);
  margin-bottom: var(--qz-space-2);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-head h2 {
  margin: var(--qz-space-1) 0 0;
  color: var(--qz-text);
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: var(--qz-leading-body);
}

/* ---------- 分类列表（侧栏与抽屉共用） ---------- */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-list {
  display: grid;
  gap: var(--qz-space-1);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-pill.qz-cat-item {
  display: grid;
  grid-template-columns: 32px minmax(0, 1fr) auto;
  align-items: center;
  width: 100%;
  padding: var(--qz-space-1) var(--qz-space-2);
  border-radius: var(--qz-radius-control);
  text-align: left;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-item.is-child { padding-left: var(--qz-space-6); }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-item > span:not(.qz-cat-logo) {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-item.is-active .qz-badge {
  color: var(--qz-on-accent);
  background: color-mix(in oklab, var(--qz-on-accent) 18%, transparent);
}

/* 分类图标：圆形底板，品牌 logo 内缩一圈 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-logo {
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  overflow: hidden;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-surface-muted);
  color: var(--qz-accent-strong);
  font-size: var(--qz-text-2xs);
  font-weight: 800;
  text-transform: uppercase;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-logo svg,
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-logo > .qz-brand-icon > svg { width: 16px; height: 16px; }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-logo > .qz-brand-icon > img {
  box-sizing: border-box;
  padding: var(--qz-space-1);
}
/* 暗色下给品牌 logo 垫一块浅底（--qz-text 在暗色是近白），深色 logo 才看得见 */
.dark :is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-logo > .qz-brand-icon > img { background: var(--qz-text); }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-logo.hot {
  color: var(--qz-warning);
  background: color-mix(in oklab, var(--qz-warning) 12%, transparent);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-item.is-active .qz-cat-logo {
  color: var(--qz-on-accent);
  background: color-mix(in oklab, var(--qz-on-accent) 18%, transparent);
}

/* ---------- 手机：触发钮（左对齐标签 + 右侧箭头） ---------- */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-btn.qz-cat-btn {
  justify-content: flex-start;
  margin-bottom: var(--qz-space-4);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-btn > span {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-align: left;
  text-overflow: ellipsis;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-cat-btn svg:last-child { color: var(--qz-text-muted); }

/* ---------- 手机：分类抽屉（Teleport 到 body） ---------- */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-drawer {
  position: fixed;
  inset: 0;
  z-index: 75;
  display: flex;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-drawer-backdrop {
  position: absolute;
  inset: 0;
  background: color-mix(in oklab, var(--qz-text) 55%, transparent);
  backdrop-filter: blur(5px);
}
/* 暗色里 --qz-text 是近白，遮罩改用页面底色压暗 */
.dark :is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-drawer-backdrop {
  background: color-mix(in oklab, var(--qz-bg) 72%, transparent);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-drawer-panel {
  position: relative;
  display: flex;
  flex-direction: column;
  width: min(84vw, 360px);
  height: 100%;
  border-right: 1px solid var(--qz-border);
  background: var(--qz-surface);
  color: var(--qz-text);
  box-shadow: var(--qz-shadow-floating);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-drawer-panel > .qz-cat-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--qz-space-3);
  margin-bottom: 0;
  padding: var(--qz-space-5) var(--qz-space-4);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-drawer-panel > .qz-cat-head > div { min-width: 0; }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-drawer-panel > .qz-cat-list {
  flex: 1 1 0%;
  overflow-y: auto;
  padding: var(--qz-space-3) var(--qz-space-3) max(var(--qz-space-6), env(safe-area-inset-bottom));
}

/* 抽屉过渡（Vue <Transition name="qz-drawer">） */
.qz-drawer-enter-active .qz-drawer-panel,
.qz-drawer-leave-active .qz-drawer-panel { transition: transform var(--qz-duration-standard) var(--qz-ease-standard); }
.qz-drawer-enter-from .qz-drawer-panel,
.qz-drawer-leave-to .qz-drawer-panel { transform: translateX(-100%); }
.qz-drawer-enter-active .qz-drawer-backdrop,
.qz-drawer-leave-active .qz-drawer-backdrop { transition: opacity var(--qz-duration-fast) var(--qz-ease-standard); }
.qz-drawer-enter-from .qz-drawer-backdrop,
.qz-drawer-leave-to .qz-drawer-backdrop { opacity: 0; }
</style>
