<template>
  <div class="j8-container j8-404">
    <div class="j8-404-card">
      <div class="j8-404-badge">
        <Info />
        <span>{{ brandSiteName }}</span>
      </div>

      <div class="j8-404-pixel" aria-hidden="true">
        <span v-for="n in 16" :key="n" :class="'px-' + n"></span>
      </div>

      <p class="j8-404-code">404</p>
      <h1 class="j8-404-title">{{ t('notFoundPage.title') }}</h1>
      <p class="j8-404-desc">{{ t('notFoundPage.description', { site: brandSiteName }) }}</p>

      <div class="j8-404-quick">
        <p class="j8-eyebrow">{{ t('notFoundPage.quickLinksTitle') }}</p>
        <div class="j8-404-chip-row">
          <RouterLink class="j8-404-chip" to="/products"><ShoppingBag /><span>{{ t('nav.products') }}</span></RouterLink>
          <RouterLink class="j8-404-chip" to="/notice"><Bell /><span>{{ t('nav.notice') }}</span></RouterLink>
          <RouterLink class="j8-404-chip" to="/blog"><BookOpen /><span>{{ t('nav.blog') }}</span></RouterLink>
        </div>
      </div>

      <div class="j8-404-actions">
        <RouterLink class="j8-404-btn j8-404-btn--primary" to="/"><HomeIcon /><span>{{ t('notFoundPage.backHome') }}</span></RouterLink>
        <button type="button" class="j8-404-btn j8-404-btn--ghost" @click="goBack"><ArrowLeft /><span>{{ t('notFoundPage.backPrevious') }}</span></button>
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
