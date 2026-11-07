<template>
  <QzDocPage
    plain
    :eyebrow="t('qiaozhi.notice.eyebrow')"
    :title="t('nav.notice')"
    :subtitle="t('notice.subtitle')"
    back-to="/"
  >
    <!-- 加载骨架 -->
    <div v-if="loading" class="qz-notice-list" aria-busy="true">
      <span class="qz-sr-only" role="status">{{ t('common.loading') }}</span>
      <div v-for="i in 6" :key="i" class="qz-card qz-notice-skeleton" aria-hidden="true">
        <span class="qz-skeleton qz-notice-skeleton-thumb"></span>
        <span class="qz-notice-skeleton-lines">
          <span class="qz-skeleton is-line"></span>
          <span class="qz-skeleton is-sm"></span>
        </span>
      </div>
    </div>

    <!-- 公告列表：整卡可点（标题链接铺满卡片） -->
    <div v-else-if="notices.length > 0" class="qz-notice-list">
      <article v-for="notice in notices" :key="notice.id" class="qz-card qz-notice-item">
        <span class="qz-notice-thumb" aria-hidden="true">
          <img
            v-if="notice.thumbnail"
            :src="getImageUrl(notice.thumbnail)"
            alt=""
            loading="lazy"
            decoding="async"
          />
          <Bell v-else />
        </span>
        <div class="qz-notice-main">
          <div class="qz-notice-meta">
            <span class="qz-chip">{{ t('nav.notice') }}</span>
            <time :datetime="notice.published_at">{{ formatDate(notice.published_at) }}</time>
          </div>
          <h2 class="qz-notice-title">
            <RouterLink :to="`/blog/${notice.slug}`">{{ getLocalizedText(notice.title) }}</RouterLink>
          </h2>
          <p v-if="getLocalizedText(notice.summary)" class="qz-notice-summary">
            {{ getLocalizedText(notice.summary) }}
          </p>
        </div>
        <ChevronRight class="qz-notice-arrow" aria-hidden="true" />
      </article>
    </div>

    <!-- 空态 -->
    <div v-else class="qz-empty">
      <Bell />
      <b>{{ t('notice.empty') }}</b>
    </div>

    <!-- 分页：翻页加载期间保持挂载（按钮禁用），避免焦点丢失 -->
    <nav v-if="totalPages > 1" class="qz-notice-pager" :aria-label="t('pagination.label')">
      <button
        type="button"
        class="qz-icon-btn qz-icon-btn--sm"
        :disabled="currentPage <= 1 || loading"
        :aria-label="t('pagination.previous')"
        @click="changePage(currentPage - 1)"
      >
        <ChevronLeft />
      </button>
      <span aria-live="polite">{{ currentPage }} / {{ totalPages }}</span>
      <button
        type="button"
        class="qz-icon-btn qz-icon-btn--sm"
        :disabled="currentPage >= totalPages || loading"
        :aria-label="t('pagination.next')"
        @click="changePage(currentPage + 1)"
      >
        <ChevronRight />
      </button>
    </nav>
  </QzDocPage>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Bell, ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { getImageUrl } from '../../utils/image'
import { usePostList } from '../../composables/usePostList'
import QzDocPage from './components/QzDocPage.vue'

const { t } = useI18n()

const {
  loading, posts: notices, currentPage, totalPages,
  getLocalizedText, formatDate, changePage,
} = usePostList('notice', { title: () => t('nav.notice'), canonicalPath: '/notice' })
</script>

<style scoped>
.qz-notice-list {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-2);
}

/* ---------- 列表项 ---------- */
.qz-notice-item {
  display: flex;
  align-items: center;
  gap: var(--qz-space-4);
  transition:
    border-color var(--qz-duration-fast) var(--qz-ease-standard),
    box-shadow var(--qz-duration-fast) var(--qz-ease-standard);
}

@media (hover: hover) and (pointer: fine) {
  .qz-notice-item:hover {
    border-color: var(--qz-border-strong);
    box-shadow: var(--qz-shadow-raised);
  }

  .qz-notice-item:hover .qz-notice-arrow {
    color: var(--qz-text);
    transform: translateX(2px);
  }
}

/* 键盘焦点：铺满卡片的链接把焦点环画在整张卡上 */
.qz-notice-item:has(a:focus-visible) {
  outline: 2px solid var(--qz-focus-outline);
  outline-offset: 2px;
}

.qz-notice-item .qz-notice-title a:focus-visible {
  outline: none;
}

.qz-notice-thumb {
  display: grid;
  flex: none;
  place-items: center;
  width: 56px;
  height: 56px;
  overflow: hidden;
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-sm);
  background: var(--qz-surface-muted);
  color: var(--qz-text-secondary);
}

.qz-notice-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.qz-notice-thumb svg {
  width: 22px;
  height: 22px;
}

.qz-notice-main {
  display: grid;
  flex: 1 1 auto;
  gap: var(--qz-space-1);
  min-width: 0;
}

.qz-notice-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  font-variant-numeric: tabular-nums;
}

.qz-notice-title {
  margin: 0;
  font-size: var(--qz-text-base);
  font-weight: 600;
  line-height: var(--qz-leading-body);
}

.qz-notice-title a {
  color: var(--qz-text);
  text-decoration: none;
  overflow-wrap: anywhere;
}

/* 链接铺满卡片（.qz-card 自带 position: relative） */
.qz-notice-title a::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: var(--qz-radius-control);
}

.qz-notice-summary {
  display: -webkit-box;
  margin: 0;
  overflow: hidden;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
}

.qz-notice-arrow {
  flex: none;
  width: 18px;
  height: 18px;
  color: var(--qz-text-muted);
  transition:
    color var(--qz-duration-fast) var(--qz-ease-standard),
    transform var(--qz-duration-fast) var(--qz-ease-standard);
}

/* ---------- 骨架：块都是 base 的 .qz-skeleton（标题行 .is-line、摘要行 .is-sm），这里只排布局 ---------- */
.qz-notice-skeleton {
  display: flex;
  align-items: center;
  gap: var(--qz-space-4);
  min-height: 88px;
}

/* 缩略图占位跟真图同尺寸同圆角：覆盖基类的 control 圆角，写成 (0,3,0) */
.qz-skeleton.qz-notice-skeleton-thumb {
  flex: none;
  width: 56px;
  height: 56px;
  border-radius: var(--qz-radius-sm);
}

.qz-notice-skeleton-lines {
  display: grid;
  flex: 1 1 auto;
  gap: var(--qz-space-2);
}

/* ---------- 分页 ---------- */
.qz-notice-pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--qz-space-3);
  padding-top: var(--qz-space-1);
}

.qz-notice-pager span {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

@media (max-width: 760px) {
  /* (0,3,0) 压过 base.css 的 .qz-card 内边距 */
  .qz-card.qz-notice-item,
  .qz-card.qz-notice-skeleton {
    gap: var(--qz-space-3);
    padding: var(--qz-space-3);
  }

  .qz-notice-thumb,
  .qz-skeleton.qz-notice-skeleton-thumb {
    width: 44px;
    height: 44px;
    border-radius: var(--qz-radius-sm);
  }

  .qz-notice-arrow {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .qz-notice-item,
  .qz-notice-arrow {
    transition: none;
  }
}
</style>
