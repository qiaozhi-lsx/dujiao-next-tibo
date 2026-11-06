<template>
  <div class="qz-container qz-404">
    <section class="qz-panel qz-404-card">
      <span class="qz-chip qz-404-brand">
        <Info />
        <span>{{ brandSiteName }}</span>
      </span>

      <div class="qz-404-pixel" aria-hidden="true">
        <span v-for="n in 16" :key="n" :class="'px-' + n"></span>
      </div>

      <p class="qz-404-code">404</p>
      <h1 class="qz-404-title">{{ t('notFoundPage.title') }}</h1>
      <p class="qz-404-desc">{{ t('notFoundPage.description', { site: brandSiteName }) }}</p>

      <div class="qz-404-actions">
        <RouterLink class="qz-btn qz-btn--primary" to="/"><HomeIcon /><span>{{ t('notFoundPage.backHome') }}</span></RouterLink>
        <button type="button" class="qz-btn qz-btn--ghost" @click="goBack"><ArrowLeft /><span>{{ t('notFoundPage.backPrevious') }}</span></button>
      </div>

      <nav class="qz-404-quick" :aria-label="t('notFoundPage.quickLinksTitle')">
        <p class="qz-eyebrow">{{ t('notFoundPage.quickLinksTitle') }}</p>
        <div class="qz-404-chips">
          <RouterLink class="qz-chip qz-404-chip" to="/products"><ShoppingBag /><span>{{ t('nav.products') }}</span></RouterLink>
          <RouterLink class="qz-chip qz-404-chip" to="/notice"><Bell /><span>{{ t('nav.notice') }}</span></RouterLink>
          <RouterLink v-if="blogEnabled" class="qz-chip qz-404-chip" to="/blog"><BookOpen /><span>{{ t('nav.blog') }}</span></RouterLink>
        </div>
      </nav>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ArrowLeft, Bell, BookOpen, Home as HomeIcon, Info, ShoppingBag } from 'lucide-vue-next'
import { useNavConfig } from '../../composables/useNavConfig'
import { useNotFound } from '../../composables/useNotFound'

const { t } = useI18n()
const { brandSiteName, goBack } = useNotFound()
const { blogEnabled } = useNavConfig()
</script>

<style scoped>
/* 404 是所有未知路由的回落点：只依赖 tokens.css / base.css（.qz-panel / .qz-eyebrow / .qz-btn / .qz-chip），
   页面自己只补版式。 */
.qz-404 {
  display: grid;
  place-items: center;
}

/* .qz-panel.qz-404-card 写成 (0,3,0)：legacy.css 里旧分类侧栏的 .qz-panel 规则（sticky、≤760 隐藏、overflow hidden）
   在它被删掉之前不能影响这张卡。 */
.qz-panel.qz-404-card {
  position: static;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--qz-space-4);
  width: min(560px, 100%);
  padding: var(--qz-space-10) var(--qz-space-8);
  overflow: visible;
  text-align: center;
}

.qz-404-brand svg {
  width: 12px;
  height: 12px;
}

/* 像素块：4×4 网格，主色描出一圈 */
.qz-404-pixel {
  display: grid;
  grid-template-columns: repeat(4, 12px);
  gap: 4px;
  margin: var(--qz-space-2) 0;
}

.qz-404-pixel span {
  width: 12px;
  height: 12px;
  border-radius: 3px;
  background: var(--qz-accent-subtle);
}

.qz-404-pixel :is(.px-2, .px-3, .px-5, .px-8, .px-9, .px-12, .px-14, .px-15) {
  background: var(--qz-accent);
}

.qz-404-code {
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-3xl);
  font-weight: 800;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
  font-variant-numeric: tabular-nums;
}

.qz-404-title {
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-2xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
}

.qz-404-desc {
  max-width: 44ch;
  margin: 0;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
}

.qz-404-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: var(--qz-space-2);
  margin-top: var(--qz-space-2);
}

.qz-404-quick {
  display: grid;
  justify-items: center;
  gap: var(--qz-space-3);
  width: 100%;
  margin-top: var(--qz-space-2);
  padding-top: var(--qz-space-5);
  border-top: 1px solid var(--qz-border);
}

.qz-404-quick .qz-eyebrow {
  margin: 0;
}

.qz-404-chips {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: var(--qz-space-2);
}

/* 快捷链接沿用 chip 的外观，抬到 sm 控件高度做可点目标 */
.qz-chip.qz-404-chip {
  min-height: var(--qz-control-sm);
  padding: 0 var(--qz-space-3);
  color: var(--qz-text);
  font-size: var(--qz-text-xs);
  text-decoration: none;
  transition: background-color var(--qz-duration-fast) var(--qz-ease-standard);
}

.qz-chip.qz-404-chip svg {
  width: 14px;
  height: 14px;
  color: var(--qz-text-muted);
}

@media (hover: hover) and (pointer: fine) {
  .qz-chip.qz-404-chip:hover {
    background: var(--qz-accent-subtle);
  }
}

.qz-chip.qz-404-chip:active {
  background: var(--qz-accent-subtle);
}

@media (max-width: 760px) {
  .qz-panel.qz-404-card {
    padding: var(--qz-space-8) var(--qz-space-5);
  }

  .qz-404-actions {
    flex-direction: column;
    width: 100%;
  }

  .qz-404-actions .qz-btn {
    width: 100%;
  }
}
</style>
