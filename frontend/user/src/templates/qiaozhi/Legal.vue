<template>
  <QzDocPage :eyebrow="t('qiaozhi.legal.eyebrow')" :title="title" back-to="/">
    <!-- 站点配置加载中 -->
    <div v-if="loading" aria-busy="true">
      <span class="qz-sr-only" role="status">{{ t('common.loading') }}</span>
      <div class="qz-legal-skeleton" aria-hidden="true">
        <span v-for="i in 6" :key="i" class="qz-skeleton is-line"></span>
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
/* 骨架行都是 base 的 .qz-skeleton.is-line，这里只排布局与行宽（首行当小标题，略宽间距） */
.qz-legal-skeleton {
  display: grid;
  gap: var(--qz-space-3);
}

.qz-legal-skeleton > .qz-skeleton:first-child {
  width: 38%;
  margin-bottom: var(--qz-space-2);
}

.qz-legal-skeleton > .qz-skeleton:nth-child(3) {
  width: 92%;
}

.qz-legal-skeleton > .qz-skeleton:nth-child(5) {
  width: 64%;
}

.qz-legal-skeleton > .qz-skeleton:last-child {
  width: 46%;
}
</style>
