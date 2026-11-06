<template>
  <QzDocPage :eyebrow="t('qiaozhi.legal.eyebrow')" :title="title" back-to="/">
    <!-- 站点配置加载中 -->
    <div v-if="loading" aria-busy="true">
      <span class="qz-sr-only" role="status">{{ t('common.loading') }}</span>
      <div class="qz-legal-skeleton" aria-hidden="true">
        <span v-for="i in 6" :key="i"></span>
      </div>
    </div>

    <!-- 条款 / 隐私富文本 -->
    <div v-else-if="content" class="qz-doc-prose" v-html="content"></div>

    <!-- 未配置 -->
    <div v-else class="qz-empty">
      <FileText />
      <b>{{ t('common.noContent') }}</b>
    </div>
  </QzDocPage>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { FileText } from 'lucide-vue-next'
import { useLegal } from '../../composables/useLegal'
import QzDocPage from './components/QzDocPage.vue'

const { t } = useI18n()

const props = defineProps<{
  type: 'terms' | 'privacy'
}>()

const { loading, title, content } = useLegal(() => props.type)
</script>

<style scoped>
.qz-legal-skeleton {
  display: grid;
  gap: var(--qz-space-3);
}

.qz-legal-skeleton > span {
  display: block;
  width: 100%;
  height: 14px;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-border);
  animation: qz-legal-pulse 1.2s ease-in-out infinite;
}

.qz-legal-skeleton > span:first-child {
  width: 38%;
  height: 20px;
  margin-bottom: var(--qz-space-2);
}

.qz-legal-skeleton > span:nth-child(3) {
  width: 92%;
}

.qz-legal-skeleton > span:nth-child(5) {
  width: 64%;
}

.qz-legal-skeleton > span:last-child {
  width: 46%;
}

@keyframes qz-legal-pulse {
  50% {
    opacity: 0.45;
  }
}

@media (prefers-reduced-motion: reduce) {
  .qz-legal-skeleton > span {
    animation: none;
  }
}
</style>
