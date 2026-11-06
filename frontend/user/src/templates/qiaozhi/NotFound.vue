<template>
  <div class="qz-container qz-404">
    <div class="qz-404-card">
      <div class="qz-404-badge">
        <Info />
        <span>{{ brandSiteName }}</span>
      </div>

      <div class="qz-404-pixel" aria-hidden="true">
        <span v-for="n in 16" :key="n" :class="'px-' + n"></span>
      </div>

      <p class="qz-404-code">404</p>
      <h1 class="qz-404-title">{{ t('notFoundPage.title') }}</h1>
      <p class="qz-404-desc">{{ t('notFoundPage.description', { site: brandSiteName }) }}</p>

      <div class="qz-404-quick">
        <p class="qz-eyebrow">{{ t('notFoundPage.quickLinksTitle') }}</p>
        <div class="qz-404-chip-row">
          <RouterLink class="qz-404-chip" to="/products"><ShoppingBag /><span>{{ t('nav.products') }}</span></RouterLink>
          <RouterLink class="qz-404-chip" to="/notice"><Bell /><span>{{ t('nav.notice') }}</span></RouterLink>
          <RouterLink class="qz-404-chip" to="/blog"><BookOpen /><span>{{ t('nav.blog') }}</span></RouterLink>
        </div>
      </div>

      <div class="qz-404-actions">
        <RouterLink class="qz-404-btn qz-404-btn--primary" to="/"><HomeIcon /><span>{{ t('notFoundPage.backHome') }}</span></RouterLink>
        <button type="button" class="qz-404-btn qz-404-btn--ghost" @click="goBack"><ArrowLeft /><span>{{ t('notFoundPage.backPrevious') }}</span></button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ArrowLeft, Bell, BookOpen, Home as HomeIcon, Info, ShoppingBag } from 'lucide-vue-next'
import { useNotFound } from '../../composables/useNotFound'

const { t } = useI18n()
const { brandSiteName, goBack } = useNotFound()
</script>
