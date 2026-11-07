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
        <p v-if="globalAlert" class="qz-alert is-danger" role="alert">{{ globalAlert.message }}</p>
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
   （.qz-panel / .qz-badge(.is-accent) / .qz-eyebrow / .qz-alert），八个面板共用的 .qz-me-* 结构词汇在下面的非 scoped
   <style> 块里只写一份，各 me/*.vue 的 <style scoped> 只剩面板专属样式。 */
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

<style>
/* me/* 面板共享词汇（.qz-me-*）：八个面板的模板共用同一套结构类，规则只在这里写一份。
   不用 scoped（子组件内部元素不带本组件的 scope id），沿用 base.css 的作用域前缀
   `:is(.qiaozhi-scope, :where(body.qiaozhi-tokens))`，单类 (0,2,0)、复合类 (0,3,0)。
   通用件已是 base.css 基类，这里不再重复：提示条 .qz-alert.is-*、统计格 .qz-stat-grid(.is-triple) / .qz-stat、
   骨架 .qz-skeleton（qz-pulse）、.qz-badge.is-accent、徽标 / chip 内 12px 图标、.qz-section-head small、
   .qz-field-hint / .qz-field-error 零外边距。面板专属样式（开关、弹窗、礼品卡结果、升级进度、筛选工具条、
   .qz-sec-*）留在各自的 <style scoped>；面板要覆盖这里的规则时串上页面类提到 (0,3,0)，不要依赖注入顺序。 */

/* ---------- 面板根：区块纵向堆叠 ---------- */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-panel {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
}

/* ---------- 区块面板：.qz-panel 内纵向流；区块头相对基类 .qz-section-head 改为居中对齐、外边距交给 gap ---------- */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-panel.qz-me-sec {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-sec > .qz-section-head {
  align-items: center;
  margin-bottom: 0;
}

/* ---------- 区块头右侧动作组；-start / form-acts 修饰符改为左对齐（表单提交行、安全面板操作行） ---------- */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-acts {
  display: flex;
  flex: none;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--qz-space-2);
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-acts.qz-me-acts-start,
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-acts.qz-me-form-acts {
  justify-content: flex-start;
}

/* ---------- 记录行：.qz-card 横排 = 主体（金额 / 标题 + 小字）+ 右侧动作；骨架按行高 64px 占位 ---------- */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-rows {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-2);
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-card.qz-me-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2) var(--qz-space-4);
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-row-main {
  display: flex;
  flex: 1 1 220px;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-row-main b {
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-row-main small {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

/* 推广码 / API Key / Secret */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-row-main code {
  font-family: var(--qz-font-mono);
  font-size: var(--qz-text-sm);
  letter-spacing: 0.04em;
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-row-side {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2);
  margin-left: auto;
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-skeleton.qz-me-sk {
  height: 64px;
}

/* ---------- 表单：单列栅格 + 两列行 ---------- */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-form {
  display: grid;
  gap: var(--qz-space-4);
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-form-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--qz-space-4);
}

/* ---------- 翻页：两枚 .qz-icon-btn--sm 夹页码 ---------- */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--qz-space-3);
  padding-top: var(--qz-space-1);
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-pager span {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

/* ---------- 浅底子面板 + 键值行（手续费说明、礼品卡兑换结果） ---------- */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-sub {
  display: grid;
  gap: var(--qz-space-3);
  padding: var(--qz-space-3) var(--qz-space-4);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-kv {
  display: grid;
  gap: var(--qz-space-2);
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-kv-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--qz-space-3);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-kv-row > span {
  color: var(--qz-text-muted);
  white-space: nowrap;
}

:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-kv-row > b {
  min-width: 0;
  color: var(--qz-text);
  font-weight: 600;
  text-align: right;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

/* ---------- 引导块（未开通分销 / 无 API 凭证）：提示 + 按钮 ---------- */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-cta {
  display: grid;
  gap: var(--qz-space-3);
  justify-items: start;
}

@media (max-width: 760px) {
  :is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-panel.qz-me-sec {
    padding: var(--qz-space-4);
  }

  :is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-form-row {
    grid-template-columns: minmax(0, 1fr);
  }

  :is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-me-row-side {
    flex: 1 1 100%;
    justify-content: flex-end;
    margin-left: 0;
  }
}
</style>
