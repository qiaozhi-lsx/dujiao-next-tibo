<template>
  <div class="j8-me">
    <!-- 账户头卡 -->
    <header class="j8-me-profile">
      <div class="j8-me-ident">
        <span class="j8-me-avatar" aria-hidden="true">{{ displayInitial }}</span>
        <div class="j8-me-idtext">
          <h1 class="j8-me-name">{{ userProfileStore.displayName }}</h1>
          <p class="j8-me-mail">{{ userProfileStore.profile?.email || t('personalCenter.subtitle') }}</p>
        </div>
      </div>
      <div class="j8-me-badges">
        <span class="j8-me-pill" :class="emailVerifiedVariant === 'success' ? 'is-ok' : 'is-warn'">
          <ShieldCheck aria-hidden="true" />
          {{ emailVerifiedLabel }}
        </span>
        <span v-if="userProfileStore.currentLevel" class="j8-me-pill is-ok">
          <img
            v-if="isImagePath(userProfileStore.currentLevel.icon)"
            :src="getImageUrl(userProfileStore.currentLevel.icon!)"
            alt=""
          />
          <span v-else-if="userProfileStore.currentLevel.icon" aria-hidden="true">{{ userProfileStore.currentLevel.icon }}</span>
          <Crown v-else aria-hidden="true" />
          {{ levelName(userProfileStore.currentLevel) }}
        </span>
      </div>
    </header>

    <div class="j8-me-layout">
      <!-- 左侧 Tabs 导航 -->
      <aside class="j8-me-nav">
        <J8Tabs
          :items="navItems"
          :model-value="currentSection"
          :orientation="wideNav ? 'vertical' : 'horizontal'"
          :aria-label="t('personalCenter.title')"
          :scroll-left-label="t('pagination.previous')"
          :scroll-right-label="t('pagination.next')"
          @select="switchSection($event as PersonalSection)"
        />
      </aside>

      <!-- 面板内容 -->
      <section class="j8-me-content">
        <p v-if="globalAlert" class="j8-me-alert">{{ globalAlert.message }}</p>
        <Transition :name="transitionName" mode="out-in">
          <component :is="panelComponent" :key="currentSection" />
        </Transition>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import { Crown, ShieldCheck } from 'lucide-vue-next'
import { getImageUrl } from '../../utils/image'
import J8Tabs, { type J8TabItem } from './components/J8Tabs.vue'
import { usePersonalCenter, type PersonalSection } from '../../composables/usePersonalCenter'
import MeOverview from './me/MeOverview.vue'
import MeOrders from './me/MeOrders.vue'
import MeWallet from './me/MeWallet.vue'
import MeGiftCards from './me/MeGiftCards.vue'
import MeAffiliate from './me/MeAffiliate.vue'
import MeApi from './me/MeApi.vue'
import MeSecurity from './me/MeSecurity.vue'
import MeProfile from './me/MeProfile.vue'

const props = withDefaults(defineProps<{ section?: PersonalSection }>(), {
  section: 'overview',
})

const { t } = useI18n()

const {
  userProfileStore, visibleSectionItems, currentSection, globalAlert, displayInitial,
  switchSection, emailVerifiedLabel, emailVerifiedVariant, isImagePath, levelName,
} = usePersonalCenter(() => props.section)

const navItems = computed<J8TabItem[]>(() => visibleSectionItems.value.map((item) => ({
  key: item.key,
  label: t(item.label),
  icon: item.icon,
})))

// 桌面竖排 / 窄屏横排
const wideNav = ref(true)
let mq: MediaQueryList | null = null
const syncWide = () => { wideNav.value = mq?.matches ?? true }

// 面板切换方向：目标在当前位置之后 → 前进，否则回退
const PANELS: Record<PersonalSection, Component> = {
  overview: MeOverview,
  orders: MeOrders,
  wallet: MeWallet,
  giftCard: MeGiftCards,
  affiliate: MeAffiliate,
  reseller: MeOverview,
  api: MeApi,
  security: MeSecurity,
  profile: MeProfile,
}
const panelComponent = computed(() => PANELS[currentSection.value] ?? MeOverview)

const order = () => visibleSectionItems.value.map((item) => item.key)
let previousIndex = order().indexOf(currentSection.value)
const direction = ref<'fwd' | 'back'>('fwd')
watch(currentSection, (next) => {
  const keys = order()
  const nextAt = keys.indexOf(next)
  if (previousIndex >= 0 && nextAt >= 0 && nextAt < previousIndex) direction.value = 'back'
  else direction.value = 'fwd'
  previousIndex = nextAt
})
const transitionName = computed(() => (direction.value === 'back' ? 'j8-me-back' : 'j8-me-fwd'))

onMounted(() => {
  mq = window.matchMedia('(min-width: 1024px)')
  syncWide()
  mq.addEventListener?.('change', syncWide)
})
onUnmounted(() => {
  mq?.removeEventListener?.('change', syncWide)
})
</script>
