<template>
  <div class="qz-container qz-me">
    <!-- 账户头卡 -->
    <header class="qz-panel qz-me-profile">
      <div class="qz-me-ident">
        <span class="qz-me-avatar" aria-hidden="true">{{ displayInitial }}</span>
        <div class="qz-me-idtext">
          <span class="qz-eyebrow">{{ t('personalCenter.title') }}</span>
          <h1 class="qz-me-name">{{ userProfileStore.displayName }}</h1>
          <p class="qz-me-mail">{{ userProfileStore.profile?.email || t('personalCenter.subtitle') }}</p>
        </div>
      </div>
      <div class="qz-me-badges">
        <span class="qz-badge" :class="emailVerifiedVariant === 'success' ? 'is-success' : 'is-warning'">
          <ShieldCheck aria-hidden="true" />
          {{ emailVerifiedLabel }}
        </span>
        <span v-if="userProfileStore.currentLevel" class="qz-badge is-accent">
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

    <div class="qz-me-layout">
      <!-- 左侧 Tabs 导航 -->
      <aside class="qz-panel qz-me-nav">
        <QzTabs
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
      <section class="qz-me-content">
        <p v-if="globalAlert" class="qz-me-alert" role="alert">{{ globalAlert.message }}</p>
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
import QzTabs, { type QzTabItem } from './components/QzTabs.vue'
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

const navItems = computed<QzTabItem[]>(() => visibleSectionItems.value.map((item) => ({
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
const transitionName = computed(() => (direction.value === 'back' ? 'qz-me-back' : 'qz-me-fwd'))

onMounted(() => {
  mq = window.matchMedia('(min-width: 1024px)')
  syncWide()
  mq.addEventListener?.('change', syncWide)
})
onUnmounted(() => {
  mq?.removeEventListener?.('change', syncWide)
})
</script>

<style scoped>
/* 个人中心壳：头卡 + 两栏（侧栏 Tabs / 面板）。控件全部来自 styles/base.css
   （.qz-panel / .qz-badge / .qz-eyebrow），面板内部样式在各 me/*.vue 自己的 <style scoped>。 */
.qz-me {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
}

/* ---------- 账户头卡 ---------- */
.qz-panel.qz-me-profile {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--qz-space-3) var(--qz-space-5);
}

.qz-me-ident {
  display: flex;
  align-items: center;
  gap: var(--qz-space-4);
  min-width: 0;
}

.qz-me-avatar {
  display: grid;
  flex: none;
  place-items: center;
  width: 52px;
  height: 52px;
  border-radius: var(--qz-radius-control);
  background: var(--qz-accent-subtle);
  color: var(--qz-accent-strong);
  font-size: var(--qz-text-xl);
  font-weight: 700;
}

.qz-me-idtext {
  display: grid;
  gap: var(--qz-space-1);
  min-width: 0;
}

.qz-me-name {
  margin: 0;
  overflow: hidden;
  color: var(--qz-text);
  font-size: var(--qz-text-xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.qz-me-mail {
  margin: 0;
  overflow: hidden;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.qz-me-badges {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2);
}

.qz-me-badges .qz-badge svg,
.qz-me-badges .qz-badge img {
  flex: none;
  width: 12px;
  height: 12px;
  object-fit: contain;
}

/* 会员等级用主色底的 badge（基类只带四种状态色） */
.qz-badge.is-accent {
  color: var(--qz-accent-strong);
  background: var(--qz-accent-subtle);
}

/* ---------- 两栏：侧栏 Tabs（桌面 sticky）+ 面板 ---------- */
.qz-me-layout {
  display: grid;
  grid-template-columns: 250px minmax(0, 1fr);
  gap: var(--qz-space-4);
  align-items: start;
}

.qz-panel.qz-me-nav {
  position: sticky;
  top: calc(var(--qz-header-h) + var(--qz-space-4));
  padding: var(--qz-space-2);
}

.qz-me-content {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  min-width: 0;
}

/* 全局加载失败提示（usePersonalCenter 只产生 error 级别） */
.qz-me-alert {
  margin: 0;
  padding: var(--qz-space-3) var(--qz-space-4);
  border-radius: var(--qz-radius-control);
  background: color-mix(in oklab, var(--qz-danger) 12%, transparent);
  color: var(--qz-danger);
  font-size: var(--qz-text-sm);
  font-weight: 500;
  line-height: var(--qz-leading-body);
  overflow-wrap: anywhere;
}

/* ---------- 面板切换过渡（方向滑动；子组件根元素带本组件 scope） ---------- */
.qz-me-fwd-enter-active,
.qz-me-back-enter-active {
  transition:
    opacity var(--qz-duration-fast) var(--qz-ease-standard),
    transform var(--qz-duration-standard) var(--qz-ease-standard);
}

.qz-me-fwd-leave-active,
.qz-me-back-leave-active {
  transition:
    opacity var(--qz-duration-instant) var(--qz-ease-standard),
    transform var(--qz-duration-instant) var(--qz-ease-standard);
}

.qz-me-fwd-enter-from {
  opacity: 0;
  transform: translateX(var(--qz-space-3));
}

.qz-me-fwd-leave-to {
  opacity: 0;
  transform: translateX(calc(-1 * var(--qz-space-2)));
}

.qz-me-back-enter-from {
  opacity: 0;
  transform: translateX(calc(-1 * var(--qz-space-3)));
}

.qz-me-back-leave-to {
  opacity: 0;
  transform: translateX(var(--qz-space-2));
}

@media (max-width: 1023px) {
  .qz-me-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .qz-panel.qz-me-nav {
    position: static;
    padding: var(--qz-space-1);
  }
}

@media (max-width: 760px) {
  .qz-panel.qz-me-profile {
    padding: var(--qz-space-4);
  }

  .qz-me-avatar {
    width: 44px;
    height: 44px;
    font-size: var(--qz-text-lg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .qz-me-fwd-enter-active,
  .qz-me-back-enter-active,
  .qz-me-fwd-leave-active,
  .qz-me-back-leave-active {
    transition: none;
  }

  .qz-me-fwd-enter-from,
  .qz-me-fwd-leave-to,
  .qz-me-back-enter-from,
  .qz-me-back-leave-to {
    transform: none;
  }
}
</style>
