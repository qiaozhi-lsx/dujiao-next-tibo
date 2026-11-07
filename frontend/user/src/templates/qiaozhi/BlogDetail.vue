<template>
  <!-- 加载中：标题未知，先铺一张骨架面板 -->
  <div v-if="loading" class="qz-container qz-post-loading" aria-busy="true">
    <span class="qz-sr-only" role="status">{{ t('common.loading') }}</span>
    <div class="qz-post-skeleton-head" aria-hidden="true">
      <span class="qz-skeleton is-sm"></span>
      <span class="qz-skeleton is-lg"></span>
      <span class="qz-skeleton is-line"></span>
    </div>
    <section class="qz-panel qz-post-skeleton-panel" aria-hidden="true">
      <span class="qz-skeleton qz-post-skeleton-cover"></span>
      <span class="qz-skeleton is-line"></span>
      <span class="qz-skeleton is-line"></span>
      <span class="qz-skeleton is-line"></span>
      <span class="qz-skeleton is-line"></span>
    </section>
  </div>

  <!-- 文章 / 公告正文 -->
  <QzDocPage
    v-else-if="post"
    :eyebrow="post.type === 'notice' ? t('nav.notice') : t('nav.blog')"
    :title="getLocalizedText(post.title)"
    :subtitle="getLocalizedText(post.summary)"
    :back-to="backLink"
    :back-label="backText"
  >
    <article class="qz-post-article">
      <span v-if="post.thumbnail" class="qz-post-cover">
        <img :src="getImageUrl(post.thumbnail)" :alt="getLocalizedText(post.title)" loading="lazy" decoding="async" />
      </span>

      <div class="qz-post-meta">
        <span class="qz-chip">{{ post.type === 'notice' ? t('nav.notice') : t('nav.blog') }}</span>
        <time :datetime="post.published_at">{{ formatDate(post.published_at) }}</time>
      </div>

      <div class="qz-doc-prose" v-html="processHtmlForDisplay(getLocalizedText(post.content))"></div>

      <!-- 相关商品 -->
      <section v-if="relatedProducts.length" class="qz-post-related">
        <h2 class="qz-post-related-title">{{ t('blog.relatedProducts') }}</h2>
        <div class="qz-post-related-grid">
          <RouterLink
            v-for="rp in relatedProducts"
            :key="rp.id"
            :to="`/products/${rp.slug}`"
            class="qz-card qz-post-related-item"
          >
            <span class="qz-post-related-thumb" aria-hidden="true">
              <img v-if="rp.image" :src="getImageUrl(rp.image)" alt="" loading="lazy" decoding="async" />
              <template v-else>{{ getLocalizedText(rp.title).slice(0, 1).toUpperCase() }}</template>
            </span>
            <span class="qz-post-related-copy">
              <b>{{ getLocalizedText(rp.title) }}</b>
              <span>{{ formatPrice(rp.price_amount) }}</span>
            </span>
          </RouterLink>
        </div>
      </section>

      <footer class="qz-post-foot">
        <RouterLink :to="backLink" class="qz-btn qz-btn--secondary">
          <ChevronLeft />
          {{ backText }}
        </RouterLink>
      </footer>
    </article>
  </QzDocPage>

  <!-- 文章不存在 / 加载失败 -->
  <QzDocPage
    v-else
    :eyebrow="t('nav.blog')"
    :title="t('blogDetail.notFound')"
    back-to="/blog"
    :back-label="t('blogDetail.backToBlog')"
  >
    <div class="qz-empty">
      <AlertCircle />
      <b>{{ t('blogDetail.notFound') }}</b>
      <RouterLink to="/blog" class="qz-btn qz-btn--primary">{{ t('blogDetail.backToBlog') }}</RouterLink>
    </div>
  </QzDocPage>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { AlertCircle, ChevronLeft } from 'lucide-vue-next'
import { getImageUrl } from '../../utils/image'
import { processHtmlForDisplay } from '../../utils/content'
import { useBlogDetail } from '../../composables/useBlogDetail'
import QzDocPage from './components/QzDocPage.vue'

const { t } = useI18n()

const {
  loading, post, relatedProducts, getLocalizedText, formatDate, formatPrice, backLink, backText,
} = useBlogDetail()
</script>

<style scoped>
/* ---------- 加载骨架（与 QzDocPage 同宽同节奏）：块都是 base 的 .qz-skeleton，这里只排布局与行宽 ---------- */
.qz-post-loading {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  max-width: 920px;
}

/* 页头骨架：eyebrow（.is-sm，收窄到 eyebrow 的宽度，(0,4,0) 压过基类的 45%）→ 标题（.is-lg）→ 副标题（.is-line） */
.qz-post-skeleton-head {
  display: grid;
  gap: var(--qz-space-3);
}

.qz-post-skeleton-head > .qz-skeleton.is-sm {
  width: 72px;
}

.qz-post-skeleton-head > .qz-skeleton.is-lg {
  width: 60%;
}

.qz-post-skeleton-head > .qz-skeleton.is-line {
  width: 44%;
}

/* (0,3,0) 压过 base.css 的 .qz-panel 内边距，与 QzDocPage 的阅读面板一致 */
.qz-panel.qz-post-skeleton-panel {
  display: grid;
  gap: var(--qz-space-3);
  padding: var(--qz-space-8) var(--qz-space-6);
}

.qz-post-skeleton-panel > .qz-skeleton:nth-child(3) {
  width: 92%;
}

.qz-post-skeleton-panel > .qz-skeleton:nth-child(5) {
  width: 58%;
}

.qz-post-skeleton-panel > .qz-post-skeleton-cover {
  height: 220px;
  margin-bottom: var(--qz-space-3);
}

/* ---------- 正文 ---------- */
.qz-post-article {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-5);
}

.qz-post-cover {
  display: block;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
}

.qz-post-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.qz-post-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2);
  padding-bottom: var(--qz-space-4);
  border-bottom: 1px solid var(--qz-border);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  font-variant-numeric: tabular-nums;
}

/* ---------- 相关商品 ---------- */
.qz-post-related {
  display: grid;
  gap: var(--qz-space-4);
  margin-top: var(--qz-space-3);
  padding-top: var(--qz-space-6);
  border-top: 1px solid var(--qz-border);
}

.qz-post-related-title {
  margin: 0;
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: var(--qz-leading-body);
  letter-spacing: var(--qz-tracking-display);
}

.qz-post-related-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--qz-space-3);
}

/* 商品卡：(0,3,0) 压过 base.css 的 .qz-card 内边距，也压过全局 .theme-prose a 的主色 + 悬停下划线 */
.qz-card.qz-post-related-item {
  display: flex;
  align-items: center;
  gap: var(--qz-space-3);
  padding: var(--qz-space-3);
  color: var(--qz-text);
  text-decoration: none;
  transition:
    border-color var(--qz-duration-fast) var(--qz-ease-standard),
    box-shadow var(--qz-duration-fast) var(--qz-ease-standard);
}

@media (hover: hover) and (pointer: fine) {
  .qz-card.qz-post-related-item:hover {
    border-color: var(--qz-border-strong);
    box-shadow: var(--qz-shadow-raised);
    color: var(--qz-text);
    text-decoration: none;
  }
}

.qz-post-related-thumb {
  display: grid;
  flex: none;
  place-items: center;
  width: 48px;
  height: 48px;
  overflow: hidden;
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-sm);
  background: var(--qz-surface-muted);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  font-weight: 700;
}

.qz-post-related-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.qz-post-related-copy {
  display: grid;
  gap: var(--qz-space-1);
  min-width: 0;
}

.qz-post-related-copy b {
  overflow: hidden;
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  line-height: var(--qz-leading-body);
  white-space: nowrap;
  text-overflow: ellipsis;
}

.qz-post-related-copy span {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  font-variant-numeric: tabular-nums;
}

/* ---------- 页脚返回 ---------- */
.qz-post-foot {
  display: flex;
  justify-content: center;
  margin-top: var(--qz-space-3);
  padding-top: var(--qz-space-6);
  border-top: 1px solid var(--qz-border);
}

@media (max-width: 760px) {
  .qz-post-skeleton-panel {
    padding: var(--qz-space-6) var(--qz-space-4);
  }

  .qz-post-related-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .qz-post-foot .qz-btn {
    width: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .qz-post-related-item {
    transition: none;
  }
}
</style>
