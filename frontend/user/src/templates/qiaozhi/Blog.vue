<template>
  <QzDocPage
    plain
    :eyebrow="t('qiaozhi.blog.eyebrow')"
    :title="t('nav.blog')"
    :subtitle="t('blog.subtitle')"
    back-to="/"
  >
    <!-- 搜索：关键字变化由 usePostList 防抖重查 -->
    <div class="qz-field qz-blog-search">
      <label class="qz-sr-only" for="qz-blog-search">{{ t('blog.searchPlaceholder') }}</label>
      <div class="qz-field-control qz-blog-search-control">
        <Search aria-hidden="true" />
        <input
          id="qz-blog-search"
          v-model="searchKeyword"
          type="text"
          autocomplete="off"
          enterkeyhint="search"
          :placeholder="t('blog.searchPlaceholder')"
        />
        <button v-if="searchKeyword" type="button" class="qz-btn qz-btn--ghost qz-btn--sm" @click="searchKeyword = ''">
          {{ t('blog.searchClear') }}
        </button>
      </div>
    </div>

    <!-- 加载骨架 -->
    <div v-if="loading" class="qz-blog-grid" aria-busy="true">
      <span class="qz-sr-only" role="status">{{ t('common.loading') }}</span>
      <div v-for="i in 6" :key="i" class="qz-card qz-blog-skeleton" aria-hidden="true">
        <span class="qz-skeleton qz-blog-skeleton-cover"></span>
        <span class="qz-skeleton is-sm"></span>
        <span class="qz-skeleton is-line"></span>
        <span class="qz-skeleton is-line"></span>
      </div>
    </div>

    <!-- 文章卡片：整卡可点（标题链接铺满卡片） -->
    <div v-else-if="posts.length > 0" class="qz-blog-grid">
      <article v-for="post in posts" :key="post.id" class="qz-card qz-blog-item">
        <!-- 无封面时放占位块，同一行的卡片高度保持一致 -->
        <span class="qz-blog-cover" aria-hidden="true">
          <img v-if="post.thumbnail" :src="getImageUrl(post.thumbnail)" alt="" loading="lazy" decoding="async" />
          <BookOpen v-else />
        </span>
        <div class="qz-blog-body">
          <div class="qz-blog-meta">
            <span class="qz-chip">{{ post.type === 'blog' ? t('nav.blog') : t('nav.notice') }}</span>
            <time :datetime="post.published_at">{{ formatDate(post.published_at) }}</time>
          </div>
          <h2 class="qz-blog-title">
            <RouterLink :to="`/blog/${post.slug}`">{{ getLocalizedText(post.title) }}</RouterLink>
          </h2>
          <p v-if="getLocalizedText(post.summary)" class="qz-blog-summary">
            {{ getLocalizedText(post.summary) }}
          </p>
          <span class="qz-blog-more" aria-hidden="true">
            {{ t('blog.readMore') }}
            <ArrowRight />
          </span>
        </div>
      </article>
    </div>

    <!-- 空态：有关键字时提示无匹配 -->
    <div v-else class="qz-empty">
      <BookOpen />
      <b>{{ searchKeyword.trim() ? t('blog.noResults') : t('blog.empty') }}</b>
    </div>

    <!-- 分页：翻页加载期间保持挂载（按钮禁用），避免焦点丢失 -->
    <nav v-if="totalPages > 1" class="qz-blog-pager" :aria-label="t('pagination.label')">
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
import { ArrowRight, BookOpen, ChevronLeft, ChevronRight, Search } from 'lucide-vue-next'
import { getImageUrl } from '../../utils/image'
import { usePostList } from '../../composables/usePostList'
import QzDocPage from './components/QzDocPage.vue'

const { t } = useI18n()

const {
  loading, posts, currentPage, totalPages, searchKeyword,
  getLocalizedText, formatDate, changePage,
} = usePostList('blog', { title: () => t('nav.blog'), canonicalPath: '/blog' })
</script>

<style scoped>
/* ---------- 搜索 ---------- */
.qz-blog-search {
  max-width: 520px;
}

/* .qz-field-control 做外壳：图标 + 裸 input + 清除钮 */
.qz-field-control.qz-blog-search-control {
  padding-right: var(--qz-space-1);
}

/* ---------- 卡片网格 ---------- */
.qz-blog-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--qz-space-4);
}

/* (0,3,0) 压过 base.css 的 .qz-card 内边距：封面贴边更近 */
.qz-card.qz-blog-item {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-3);
  padding: var(--qz-space-3);
  transition:
    border-color var(--qz-duration-fast) var(--qz-ease-standard),
    box-shadow var(--qz-duration-fast) var(--qz-ease-standard);
}

@media (hover: hover) and (pointer: fine) {
  .qz-blog-item:hover {
    border-color: var(--qz-border-strong);
    box-shadow: var(--qz-shadow-raised);
  }

  .qz-blog-item:hover .qz-blog-more {
    color: var(--qz-text);
  }

  .qz-blog-item:hover .qz-blog-more svg {
    transform: translateX(2px);
  }
}

/* 键盘焦点：铺满卡片的链接把焦点环画在整张卡上 */
.qz-blog-item:has(a:focus-visible) {
  outline: 2px solid var(--qz-focus-outline);
  outline-offset: 2px;
}

.qz-blog-item .qz-blog-title a:focus-visible {
  outline: none;
}

.qz-blog-cover {
  display: grid;
  place-items: center;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  border-radius: calc(var(--qz-radius-control) - var(--qz-space-3));
  background: var(--qz-surface-muted);
  color: var(--qz-text-muted);
}

.qz-blog-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.qz-blog-cover svg {
  width: 28px;
  height: 28px;
}

.qz-blog-body {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: var(--qz-space-2);
  min-width: 0;
  padding: var(--qz-space-1) var(--qz-space-1) var(--qz-space-1);
}

.qz-blog-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  font-variant-numeric: tabular-nums;
}

.qz-blog-title {
  display: -webkit-box;
  margin: 0;
  overflow: hidden;
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: var(--qz-leading-body);
  letter-spacing: var(--qz-tracking-display);
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
}

.qz-blog-title a {
  color: var(--qz-text);
  text-decoration: none;
  overflow-wrap: anywhere;
}

/* 链接铺满卡片（.qz-card 自带 position: relative） */
.qz-blog-title a::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: var(--qz-radius-control);
}

.qz-blog-summary {
  display: -webkit-box;
  flex: 1 1 auto;
  margin: 0;
  overflow: hidden;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  line-clamp: 3;
}

.qz-blog-more {
  display: inline-flex;
  align-items: center;
  gap: var(--qz-space-1);
  margin-top: auto;
  padding-top: var(--qz-space-3);
  border-top: 1px solid var(--qz-border);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  transition: color var(--qz-duration-fast) var(--qz-ease-standard);
}

.qz-blog-more svg {
  width: 14px;
  height: 14px;
  transition: transform var(--qz-duration-fast) var(--qz-ease-standard);
}

/* ---------- 骨架：块都是 base 的 .qz-skeleton（meta 行 .is-sm、标题 / 摘要行 .is-line），这里只排布局 ---------- */
.qz-card.qz-blog-skeleton {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-3);
  min-height: 260px;
  padding: var(--qz-space-3);
}

/* 封面占位跟真封面同比例同圆角：覆盖基类的 control 圆角，写成 (0,3,0) */
.qz-skeleton.qz-blog-skeleton-cover {
  aspect-ratio: 16 / 9;
  border-radius: calc(var(--qz-radius-control) - var(--qz-space-3));
}

/* ---------- 分页 ---------- */
.qz-blog-pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--qz-space-3);
  padding-top: var(--qz-space-1);
}

.qz-blog-pager span {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

@media (max-width: 760px) {
  .qz-blog-search {
    max-width: none;
  }

  .qz-blog-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (prefers-reduced-motion: reduce) {
  .qz-blog-item,
  .qz-blog-more,
  .qz-blog-more svg {
    transition: none;
  }
}
</style>
