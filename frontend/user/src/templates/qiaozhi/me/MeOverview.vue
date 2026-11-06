<template>
  <div class="qz-me-panel">
    <!-- 数据一览 -->
    <div class="qz-me-stats">
      <div class="qz-card qz-me-stat">
        <b>{{ userProfileStore.loadingOrders ? '—' : userProfileStore.ordersTotal }}</b>
        <span>{{ t('personalCenter.tabs.orders') }}</span>
      </div>
      <div class="qz-card qz-me-stat">
        <b><span class="qz-me-stat-text">{{ currentLevelName }}</span></b>
        <span>{{ t('personalCenter.memberLevel.currentLevel') }}</span>
      </div>
      <div class="qz-card qz-me-stat">
        <b><span class="qz-me-stat-text">{{ discountText }}</span></b>
        <span>{{ t('personalCenter.memberLevel.discountRate') }}</span>
      </div>
      <div class="qz-card qz-me-stat">
        <b><span class="qz-me-stat-text">{{ emailVerifiedLabel }}</span></b>
        <span>{{ t('personalCenter.overview.accountLabel') }}</span>
      </div>
    </div>

    <!-- 会员等级卡 -->
    <section v-if="userProfileStore.memberLevels.length > 0" class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.memberLevel.currentLevel') }}</h2>
          <small>{{ currentLevelName }}</small>
        </div>
        <div class="qz-me-acts">
          <span class="qz-chip">
            {{ t('personalCenter.memberLevel.discountRate') }}
            {{ discountOffText }}
          </span>
          <span v-if="!userProfileStore.nextLevel" class="qz-badge is-info">
            {{ t('personalCenter.memberLevel.highestLevel') }}
          </span>
        </div>
      </div>

      <!-- 升级进度 -->
      <div v-if="userProfileStore.nextLevel" class="qz-me-level">
        <p class="qz-field-hint">
          {{ t('personalCenter.memberLevel.nextLevel') }} ·
          <b>{{ nextLevelName }}</b>
          <template v-if="nextLevelDiscount !== ''">（{{ nextLevelDiscount }}）</template>
        </p>
        <div v-if="userProfileStore.upgradeProgress" class="qz-me-progress-list">
          <div v-if="userProfileStore.upgradeProgress.rechargePercent !== null" class="qz-me-progress">
            <div class="qz-me-progress-head">
              <span>{{ t('personalCenter.memberLevel.rechargeProgress') }}</span>
              <b>{{ userProfileStore.upgradeProgress.recharged.toFixed(2) }} / {{ userProfileStore.upgradeProgress.rechargeThreshold.toFixed(2) }}</b>
            </div>
            <div class="qz-me-progress-track">
              <div class="qz-me-progress-fill" :style="{ width: userProfileStore.upgradeProgress.rechargePercent + '%' }"></div>
            </div>
          </div>
          <div v-if="userProfileStore.upgradeProgress.spendPercent !== null" class="qz-me-progress">
            <div class="qz-me-progress-head">
              <span>{{ t('personalCenter.memberLevel.spendProgress') }}</span>
              <b>{{ userProfileStore.upgradeProgress.spent.toFixed(2) }} / {{ userProfileStore.upgradeProgress.spendThreshold.toFixed(2) }}</b>
            </div>
            <div class="qz-me-progress-track">
              <div class="qz-me-progress-fill" :style="{ width: userProfileStore.upgradeProgress.spendPercent + '%' }"></div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 最近订单 -->
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div><h2>{{ t('personalCenter.overview.recentOrdersTitle') }}</h2></div>
        <div class="qz-me-acts">
          <RouterLink class="qz-btn qz-btn--secondary qz-btn--sm" to="/me/orders">{{ t('personalCenter.overview.viewAllOrders') }}</RouterLink>
        </div>
      </div>

      <div v-if="userProfileStore.loadingOrders" class="qz-me-rows" aria-hidden="true">
        <div v-for="i in 3" :key="i" class="qz-me-sk"></div>
      </div>
      <p v-else-if="userProfileStore.recentOrders.length === 0" class="qz-field-hint">
        {{ t('personalCenter.overview.emptyOrders') }}
      </p>
      <div v-else class="qz-me-rows">
        <div v-for="order in userProfileStore.recentOrders" :key="order.order_no" class="qz-card qz-me-row">
          <div class="qz-me-row-main">
            <b>{{ formatMoney(order.total_amount, order.currency) }}</b>
            <small>{{ t('orders.orderNo') }}：{{ order.order_no }}</small>
            <small>{{ formatDate(order.created_at) }}</small>
          </div>
          <div class="qz-me-row-side">
            <span class="qz-badge" :class="badgeToneClass(statusBadgeTone(order.status))">
              {{ statusLabel(t, order.status) }}
            </span>
            <RouterLink class="qz-btn qz-btn--secondary qz-btn--sm" :to="`/orders/${order.order_no}`">{{ t('orders.viewDetails') }}</RouterLink>
            <RouterLink v-if="order.status === 'pending_payment'" class="qz-btn qz-btn--primary qz-btn--sm" :to="`/pay?order_no=${order.order_no}`">
              {{ t('orders.payNow') }}
            </RouterLink>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUserProfileStore } from '../../../stores/userProfile'
import { badgeToneClass, formatDate, formatMoney, levelName, statusBadgeTone, statusLabel } from './support'

const { t, locale } = useI18n()
const userProfileStore = useUserProfileStore()

const currentLevelName = computed(() => levelName(t, locale.value as string, userProfileStore.currentLevel))
const nextLevelName = computed(() => levelName(t, locale.value as string, userProfileStore.nextLevel))

const emailVerifiedLabel = computed(() =>
  userProfileStore.profile?.email_verified_at
    ? t('personalCenter.overview.emailVerified')
    : t('personalCenter.overview.emailUnverified'),
)

const discountOffText = computed(() => {
  const lvl = userProfileStore.currentLevel
  if (lvl && lvl.discount_rate < 100) return t('personalCenter.memberLevel.discountOff', { n: lvl.discount_rate / 10 })
  return t('personalCenter.memberLevel.noDiscount')
})
const discountText = discountOffText

const nextLevelDiscount = computed(() => {
  const lvl = userProfileStore.nextLevel
  if (lvl && lvl.discount_rate < 100) return t('personalCenter.memberLevel.discountOff', { n: lvl.discount_rate / 10 })
  return ''
})
</script>

<style scoped>
/* 概览面板：统计卡 .qz-card、区块 .qz-panel + .qz-section-head、订单行 .qz-card、状态 .qz-badge（styles/base.css） */
.qz-me-panel {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
}

/* ---------- 统计卡 ---------- */
.qz-me-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--qz-space-3);
}

.qz-card.qz-me-stat {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-1);
}

.qz-me-stat b {
  display: flex;
  align-items: center;
  gap: var(--qz-space-2);
  min-width: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
  font-variant-numeric: tabular-nums;
}

.qz-me-stat-text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.qz-me-stat > span {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  font-weight: 500;
  line-height: var(--qz-leading-body);
}

/* ---------- 区块面板 ---------- */
.qz-panel.qz-me-sec {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
}

.qz-me-sec > .qz-section-head {
  align-items: center;
  margin-bottom: 0;
}

.qz-section-head small {
  display: block;
  margin-top: var(--qz-space-1);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-me-acts {
  display: flex;
  flex: none;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--qz-space-2);
}

.qz-me-sec .qz-field-hint {
  margin: 0;
}

.qz-field-hint b {
  color: var(--qz-text);
  font-weight: 600;
}

/* ---------- 升级进度 ---------- */
.qz-me-level {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-3);
}

.qz-me-progress-list {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-3);
}

.qz-me-progress {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-2);
}

.qz-me-progress-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--qz-space-2);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  font-weight: 500;
  line-height: var(--qz-leading-body);
}

.qz-me-progress-head b {
  color: var(--qz-text);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.qz-me-progress-track {
  height: 6px;
  overflow: hidden;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-border);
}

.qz-me-progress-fill {
  height: 100%;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-accent);
  transition: width var(--qz-duration-considered) var(--qz-ease-standard);
}

/* ---------- 订单行 ---------- */
.qz-me-rows {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-2);
}

.qz-card.qz-me-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2) var(--qz-space-4);
}

.qz-me-row-main {
  display: flex;
  flex: 1 1 220px;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.qz-me-row-main b {
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.qz-me-row-main small {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.qz-me-row-side {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2);
  margin-left: auto;
}

/* 「处理中」用主色底的 badge，基类只带四种状态色 */
.qz-badge.is-accent {
  color: var(--qz-accent-strong);
  background: var(--qz-accent-subtle);
}

/* ---------- 骨架 ---------- */
.qz-me-sk {
  height: 64px;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  animation: qz-me-pulse 1.4s ease-in-out infinite;
}

@keyframes qz-me-pulse {
  50% {
    opacity: 0.55;
  }
}

@media (max-width: 900px) {
  .qz-me-stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .qz-panel.qz-me-sec {
    padding: var(--qz-space-4);
  }

  .qz-me-row-side {
    flex: 1 1 100%;
    justify-content: flex-end;
    margin-left: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .qz-me-sk {
    animation: none;
  }

  .qz-me-progress-fill {
    transition: none;
  }
}
</style>
