<template>
  <div class="j8-me-panel">
    <!-- 数据一览 -->
    <div class="j8-me-stats">
      <div class="j8-me-stat">
        <b>{{ userProfileStore.loadingOrders ? '—' : userProfileStore.ordersTotal }}</b>
        <span>{{ t('personalCenter.tabs.orders') }}</span>
      </div>
      <div class="j8-me-stat">
        <b><span class="j8-me-stat-text">{{ currentLevelName }}</span></b>
        <span>{{ t('personalCenter.memberLevel.currentLevel') }}</span>
      </div>
      <div class="j8-me-stat">
        <b><span class="j8-me-stat-text">{{ discountText }}</span></b>
        <span>{{ t('personalCenter.memberLevel.discountRate') }}</span>
      </div>
      <div class="j8-me-stat">
        <b><span class="j8-me-stat-text">{{ emailVerifiedLabel }}</span></b>
        <span>{{ t('personalCenter.overview.accountLabel') }}</span>
      </div>
    </div>

    <!-- 会员等级卡 -->
    <div v-if="userProfileStore.memberLevels.length > 0" class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ t('personalCenter.memberLevel.currentLevel') }}</h2>
          <small>{{ currentLevelName }}</small>
        </div>
        <div class="j8-me-card-acts">
          <span class="j8-me-pill is-ok">
            {{ t('personalCenter.memberLevel.discountRate') }}
            {{ discountOffText }}
          </span>
          <span v-if="!userProfileStore.nextLevel" class="j8-me-pill is-info">
            {{ t('personalCenter.memberLevel.highestLevel') }}
          </span>
        </div>
      </div>

      <!-- 升级进度 -->
      <div v-if="userProfileStore.nextLevel" class="j8-me-level-when">
        <p class="j8-me-hint">
          {{ t('personalCenter.memberLevel.nextLevel') }} ·
          <b>{{ nextLevelName }}</b>
          <template v-if="nextLevelDiscount !== ''">（{{ nextLevelDiscount }}）</template>
        </p>
        <div v-if="userProfileStore.upgradeProgress" class="j8-me-progress-list">
          <div v-if="userProfileStore.upgradeProgress.rechargePercent !== null" class="j8-me-progress">
            <div class="j8-me-progress-head">
              <span>{{ t('personalCenter.memberLevel.rechargeProgress') }}</span>
              <b>{{ userProfileStore.upgradeProgress.recharged.toFixed(2) }} / {{ userProfileStore.upgradeProgress.rechargeThreshold.toFixed(2) }}</b>
            </div>
            <div class="j8-me-progress-track">
              <div class="j8-me-progress-fill" :style="{ width: userProfileStore.upgradeProgress.rechargePercent + '%' }"></div>
            </div>
          </div>
          <div v-if="userProfileStore.upgradeProgress.spendPercent !== null" class="j8-me-progress">
            <div class="j8-me-progress-head">
              <span>{{ t('personalCenter.memberLevel.spendProgress') }}</span>
              <b>{{ userProfileStore.upgradeProgress.spent.toFixed(2) }} / {{ userProfileStore.upgradeProgress.spendThreshold.toFixed(2) }}</b>
            </div>
            <div class="j8-me-progress-track">
              <div class="j8-me-progress-fill" :style="{ width: userProfileStore.upgradeProgress.spendPercent + '%' }"></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 最近订单 -->
    <div class="j8-me-card">
      <div class="j8-me-card-head">
        <div><h2>{{ t('personalCenter.overview.recentOrdersTitle') }}</h2></div>
        <div class="j8-me-card-acts">
          <RouterLink class="j8-me-act" to="/me/orders">{{ t('personalCenter.overview.viewAllOrders') }}</RouterLink>
        </div>
      </div>

      <div v-if="userProfileStore.loadingOrders" class="j8-me-rows">
        <div v-for="i in 3" :key="i" class="j8-me-sk" aria-hidden="true"></div>
      </div>
      <p v-else-if="userProfileStore.recentOrders.length === 0" class="j8-me-hint">
        {{ t('personalCenter.overview.emptyOrders') }}
      </p>
      <div v-else class="j8-me-rows">
        <div v-for="order in userProfileStore.recentOrders" :key="order.order_no" class="j8-me-row">
          <div class="j8-me-row-main">
            <b>{{ formatMoney(order.total_amount, order.currency) }}</b>
            <small>{{ t('orders.orderNo') }}：{{ order.order_no }}</small>
            <small>{{ formatDate(order.created_at) }}</small>
          </div>
          <div class="j8-me-row-side">
            <span class="j8-me-pill" :class="`is-${statusPillTone(order.status)}`">
              {{ statusLabel(t, order.status) }}
            </span>
            <RouterLink class="j8-me-act" :to="`/orders/${order.order_no}`">{{ t('orders.viewDetails') }}</RouterLink>
            <RouterLink v-if="order.status === 'pending_payment'" class="j8-me-act is-primary" :to="`/pay?order_no=${order.order_no}`">
              {{ t('orders.payNow') }}
            </RouterLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUserProfileStore } from '../../../stores/userProfile'
import { formatDate, formatMoney, levelName, statusLabel, statusPillTone } from './support'

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
.j8-me-progress-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 10px;
}
</style>
