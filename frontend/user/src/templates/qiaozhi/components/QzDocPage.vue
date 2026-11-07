<template>
  <div ref="rootEl" class="qz-container qz-doc">
    <!-- 页头：eyebrow + h1 + 说明 + 返回胶囊（同查订单页页头） -->
    <header class="qz-page-head qz-reveal is-pending">
      <div class="qz-page-head-copy">
        <span class="qz-eyebrow">{{ eyebrow }}</span>
        <h1>{{ title }}</h1>
        <p v-if="subtitle">{{ subtitle }}</p>
      </div>
      <RouterLink v-if="backTo" :to="backTo" class="qz-btn qz-btn--secondary qz-btn--sm">
        <ChevronLeft />
        {{ backLabel || t('auth.login.backHome') }}
      </RouterLink>
    </header>

    <!-- 正文：默认装进 .qz-panel.theme-prose（文档 / 富文本页）；plain 时直接铺在容器里（卡片列表页） -->
    <section v-if="!plain" class="qz-panel theme-prose qz-doc-panel qz-reveal is-pending">
      <slot />
    </section>
    <div v-else class="qz-doc-body qz-reveal is-pending">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft } from 'lucide-vue-next'
import type { RouteLocationRaw } from 'vue-router'
import { setupReveal } from '../utils/reveal'

/**
 * 文档类页面外壳（公告 / 关于 / 条款隐私 / 博客 / 文章详情共用）。
 * - eyebrow / title / subtitle 组成页头，backTo 给出返回胶囊（文案缺省「返回首页」）。
 * - 默认插槽放正文：缺省包在 .qz-panel.theme-prose 里，富文本用 .qz-doc-prose 包一层即可获得排版；
 *   列表页传 plain，正文直接铺在容器里（.qz-card 列表不再套一层面板）。
 */
defineProps<{
  eyebrow: string
  title: string
  subtitle?: string
  backTo?: RouteLocationRaw
  backLabel?: string
  plain?: boolean
}>()

const { t } = useI18n()

// 入场动画：页头与正文容器在挂载时就存在，arm 一次即可
const rootEl = ref<HTMLElement | null>(null)
let cleanupReveal: (() => void) | null = null
onMounted(() => {
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
})
onUnmounted(() => {
  cleanupReveal?.()
})
</script>

<style scoped>
.qz-doc {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  max-width: 920px;
}

/* ---------- 页头：base.css 的 .qz-page-head；标题 / 摘要来自后台富文本，补一条可在任意处断行 ---------- */
.qz-doc .qz-page-head :is(h1, p) {
  overflow-wrap: anywhere;
}

/* ---------- 正文 ----------
   写成 .qz-panel.qz-doc-panel (0,3,0)：阅读面板比 base.css 的 .qz-panel 默认内边距更宽松，不靠加载顺序赢。 */
.qz-panel.qz-doc-panel {
  padding: var(--qz-space-8) var(--qz-space-6);
}

.qz-doc-body {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  min-width: 0;
}

/* ---------- 富文本 ----------
   v-html 注入的节点没有 scoped 属性，用 :deep() 命中；.theme-prose（全局）负责标题 / 正文 / 链接颜色，
   这里只补排版节奏，所有尺寸走 token。 */
.qz-doc-panel :deep(.qz-doc-prose) {
  color: var(--qz-text);
  font-size: var(--qz-text-base);
  line-height: 1.75;
  overflow-wrap: anywhere;
}

.qz-doc-panel :deep(.qz-doc-prose > :first-child) {
  margin-top: 0;
}

.qz-doc-panel :deep(.qz-doc-prose > :last-child) {
  margin-bottom: 0;
}

.qz-doc-panel :deep(.qz-doc-prose p) {
  margin: 0 0 var(--qz-space-4);
}

.qz-doc-panel :deep(.qz-doc-prose :is(h1, h2, h3, h4, h5, h6)) {
  margin: var(--qz-space-6) 0 var(--qz-space-3);
  font-weight: 700;
  line-height: var(--qz-leading-body);
  letter-spacing: var(--qz-tracking-display);
}

.qz-doc-panel :deep(.qz-doc-prose h1) {
  font-size: var(--qz-text-xl);
}

.qz-doc-panel :deep(.qz-doc-prose h2) {
  font-size: var(--qz-text-lg);
}

.qz-doc-panel :deep(.qz-doc-prose :is(h3, h4, h5, h6)) {
  font-size: var(--qz-text-base);
}

.qz-doc-panel :deep(.qz-doc-prose :is(ul, ol)) {
  margin: 0 0 var(--qz-space-4);
  padding-left: 1.5em;
}

.qz-doc-panel :deep(.qz-doc-prose ul) {
  list-style: disc;
}

.qz-doc-panel :deep(.qz-doc-prose ol) {
  list-style: decimal;
}

.qz-doc-panel :deep(.qz-doc-prose li) {
  margin: 0 0 var(--qz-space-1);
}

/* 正文里主色与文字同色，链接必须带下划线才分得出来 */
.qz-doc-panel :deep(.qz-doc-prose a) {
  text-decoration: underline;
  text-underline-offset: 3px;
}

.qz-doc-panel :deep(.qz-doc-prose img) {
  display: block;
  max-width: 100%;
  height: auto;
  margin: var(--qz-space-4) 0;
  border-radius: var(--qz-radius-control);
}

.qz-doc-panel :deep(.qz-doc-prose blockquote) {
  margin: 0 0 var(--qz-space-4);
  padding: var(--qz-space-3) var(--qz-space-4);
  border-left: 3px solid var(--qz-border-strong);
  border-radius: 0 var(--qz-radius-control) var(--qz-radius-control) 0;
  background: var(--qz-surface-muted);
}

.qz-doc-panel :deep(.qz-doc-prose :is(pre, code)) {
  font-family: var(--qz-font-mono);
  font-size: var(--qz-text-sm);
}

.qz-doc-panel :deep(.qz-doc-prose code) {
  padding: 0 0.3em;
  border-radius: var(--qz-radius-xs);
  background: var(--qz-surface-muted);
}

.qz-doc-panel :deep(.qz-doc-prose pre) {
  margin: 0 0 var(--qz-space-4);
  padding: var(--qz-space-4);
  overflow-x: auto;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
}

.qz-doc-panel :deep(.qz-doc-prose pre code) {
  padding: 0;
  background: transparent;
}

.qz-doc-panel :deep(.qz-doc-prose hr) {
  margin: var(--qz-space-6) 0;
  border: 0;
  border-top: 1px solid var(--qz-border);
}

.qz-doc-panel :deep(.qz-doc-prose table) {
  display: block;
  width: 100%;
  margin: 0 0 var(--qz-space-4);
  overflow-x: auto;
  border-collapse: collapse;
  font-size: var(--qz-text-sm);
}

.qz-doc-panel :deep(.qz-doc-prose :is(th, td)) {
  padding: var(--qz-space-2) var(--qz-space-3);
  border: 1px solid var(--qz-border);
  text-align: left;
}

.qz-doc-panel :deep(.qz-doc-prose th) {
  background: var(--qz-surface-muted);
  font-weight: 600;
}

@media (max-width: 760px) {
  .qz-panel.qz-doc-panel {
    padding: var(--qz-space-6) var(--qz-space-4);
  }
}
</style>
