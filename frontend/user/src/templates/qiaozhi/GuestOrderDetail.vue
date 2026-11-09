<template>
  <div ref="rootEl" class="qz-container qz-od">
    <!-- 页头：eyebrow + h1（含订单状态）+ 说明 + 返回 -->
    <header class="qz-page-head">
      <div class="qz-page-head-copy">
        <span class="qz-eyebrow">{{ t('guestOrders.title') }}</span>
        <h1>
          <span>{{ t('guestOrderDetail.title') }}</span>
          <span v-if="showOrderTitle" class="qz-badge" :class="toneClass(order?.status)">
            <component :is="statusIcon" />
            {{ statusLabel(order?.status) }}
          </span>
        </h1>
        <p>{{ t('guestOrderDetail.subtitle') }}</p>
      </div>
      <RouterLink class="qz-btn qz-btn--secondary qz-btn--sm" to="/guest/orders">
        <ChevronLeft />
        {{ t('qiaozhi.orderDetail.backToOrders') }}
      </RouterLink>
    </header>

    <!-- 游客鉴权 -->
    <section v-if="viewState === 'auth'" class="qz-panel qz-od-panel qz-od-auth qz-reveal is-pending">
      <div class="qz-od-auth-head">
        <h2>{{ t('guestOrderDetail.authTitle') }}</h2>
        <p>{{ t('guestOrderDetail.authHint') }}</p>
      </div>
      <form class="qz-od-auth-form" novalidate @submit.prevent="handleAuthSubmit">
        <div class="qz-od-auth-fields">
          <div class="qz-field">
            <label class="qz-field-label" for="qz-od-guest-email">{{ t('guestOrders.emailPlaceholder') }}</label>
            <input
              id="qz-od-guest-email"
              v-model="auth.email"
              class="qz-field-control"
              type="email"
              autocomplete="email"
              :placeholder="t('guestOrders.emailPlaceholder')"
              :aria-describedby="authError ? 'qz-od-auth-error' : undefined"
            />
          </div>
          <div class="qz-field">
            <label class="qz-field-label" for="qz-od-guest-password">{{ t('guestOrders.passwordPlaceholder') }}</label>
            <input
              id="qz-od-guest-password"
              v-model="auth.order_password"
              class="qz-field-control"
              type="password"
              autocomplete="current-password"
              :placeholder="t('guestOrders.passwordPlaceholder')"
              :aria-describedby="authError ? 'qz-od-auth-error' : undefined"
            />
          </div>
        </div>
        <p v-if="authError" id="qz-od-auth-error" class="qz-field-error" role="alert">{{ authError }}</p>
        <div class="qz-od-auth-actions">
          <QzButton type="submit" variant="primary">
            {{ t('guestOrderDetail.authSubmit') }}
          </QzButton>
          <button type="button" class="qz-btn qz-btn--ghost" @click="clearAuth">
            {{ t('guestOrderDetail.authClear') }}
          </button>
        </div>
      </form>
    </section>

    <!-- 骨架 -->
    <div v-else-if="viewState === 'loading'" class="qz-od-skeleton" aria-hidden="true">
      <div class="qz-skeleton is-line qz-od-sk-line"></div>
      <div class="qz-skeleton qz-od-sk-card"></div>
      <div class="qz-skeleton qz-od-sk-card"></div>
    </div>

    <!-- 订单不存在 -->
    <div v-else-if="viewState === 'empty' || !order" class="qz-empty qz-reveal is-pending">
      <PackageOpen />
      <b>{{ t('guestOrderDetail.notFound') }}</b>
    </div>

    <template v-else>
      <!-- ORDER · 订单摘要 -->
      <section class="qz-panel qz-od-panel qz-reveal is-pending">
        <div class="qz-od-panel-head">
          <h2 class="qz-eyebrow">{{ t('qiaozhi.orderDetail.sectionOrder') }}</h2>
        </div>
        <dl class="qz-od-rows">
          <div class="qz-od-row">
            <dt>{{ t('orders.orderNo') }}</dt>
            <dd class="qz-od-no">{{ order.order_no }}</dd>
          </div>
          <div class="qz-od-row">
            <dt>{{ t('orderDetail.createdAtLabel') }}</dt>
            <dd>{{ formatDate(order.created_at) }}</dd>
          </div>
          <div class="qz-od-row">
            <dt>{{ t('orderDetail.amountTotal') }}</dt>
            <dd>{{ formatPaidTotal(order) }}</dd>
          </div>
        </dl>
        <div v-if="order.status === 'pending_payment'" class="qz-od-actions-bar">
          <RouterLink class="qz-btn qz-btn--primary qz-btn--sm" :to="`/pay?guest=1&order_no=${order.order_no}`">
            {{ t('orders.payNow') }}
          </RouterLink>
        </div>
        <div v-if="hasFulfillPending || isFulfillProcuring" class="qz-od-actions-bar">
          <p v-if="hasFulfillPending" class="qz-field-hint">{{ t('qiaozhi.fulfill.pendingHint') }}</p>
          <QzButton v-if="hasFulfillPending" size="sm" @click="fulfillOpen = true">
            {{ t('qiaozhi.fulfill.fillInfo') }}
          </QzButton>
          <QzButton v-else size="sm" variant="secondary" @click="fulfillOpen = true">
            {{ t('qiaozhi.fulfill.viewProgress') }}
          </QzButton>
        </div>
      </section>

      <!-- DELIVERY · 交付详情 -->
      <QzOrderDeliveryUnit
        :order="order"
        :fulfillment-copied="fulfillmentCopied"
        :fulfillment-downloading="fulfillmentDownloading"
        :delivery-hint="deliveredEmailHint"
        @copy-fulfillment="handleCopyFulfillment"
        @download-fulfillment="handleDownloadFulfillment"
      />

      <!-- BILLING · 金额明细 -->
      <section class="qz-panel qz-od-panel qz-reveal is-pending">
        <div class="qz-od-panel-head">
          <h2 class="qz-eyebrow">{{ t('qiaozhi.orderDetail.sectionBilling') }}</h2>
        </div>
        <div class="qz-od-grid is-amount">
          <div class="qz-od-cell">
            <span class="lbl">{{ t('orderDetail.amountOriginal') }}</span>
            <b class="val">{{ formatMoney(order.original_amount, order.currency) }}</b>
          </div>
          <div class="qz-od-cell is-disc">
            <span class="lbl">{{ t('orderDetail.amountDiscount') }}</span>
            <b class="val">{{ formatDiscountMoney(order.discount_amount, order.currency) }}</b>
          </div>
          <div class="qz-od-cell">
            <span class="lbl">{{ t('orderDetail.amountTotal') }}</span>
            <b class="val">{{ formatPaidTotal(order) }}</b>
          </div>
        </div>
        <div
          v-if="hasActualPayment(order) && (hasAmount(order.wallet_paid_amount) || hasAmount(order.online_paid_amount) || hasAmount(order.refunded_amount))"
          class="qz-od-grid is-secondary"
        >
          <div v-if="hasAmount(order.wallet_paid_amount)" class="qz-od-cell">
            <span class="lbl">{{ t('orderDetail.amountWalletPaid') }}</span>
            <b class="val">{{ formatMoney(order.wallet_paid_amount, order.currency) }}</b>
          </div>
          <div v-if="hasAmount(order.online_paid_amount)" class="qz-od-cell">
            <span class="lbl">{{ t('orderDetail.amountOnlinePaid') }}</span>
            <b class="val">{{ formatMoney(order.online_paid_amount, order.currency) }}</b>
          </div>
          <div v-if="hasAmount(order.refunded_amount)" class="qz-od-cell">
            <span class="lbl">{{ t('orderDetail.amountRefunded') }}</span>
            <b class="val">{{ formatMoney(order.refunded_amount, order.currency) }}</b>
          </div>
        </div>
        <div
          v-if="hasDiscountAmount(order.member_discount_amount) || hasDiscountAmount(order.wholesale_discount_amount)"
          class="qz-od-disc-chips"
        >
          <span v-if="hasDiscountAmount(order.member_discount_amount)" class="qz-badge is-warning">
            {{ t('orderDetail.amountMemberDiscount') }} {{ formatDiscountMoney(order.member_discount_amount, order.currency) }}
          </span>
          <span v-if="hasDiscountAmount(order.wholesale_discount_amount)" class="qz-badge is-success">
            {{ t('orderDetail.amountWholesaleDiscount') }} {{ formatDiscountMoney(order.wholesale_discount_amount, order.currency) }}
          </span>
        </div>
      </section>

      <!-- REFUNDS · 退款记录 -->
      <section v-if="showRefundRecordsCard" class="qz-panel qz-od-panel qz-reveal is-pending">
        <div class="qz-od-panel-head">
          <h2 class="qz-eyebrow">{{ t('qiaozhi.orderDetail.sectionRefunds') }}</h2>
        </div>
        <div v-if="refundRecords.length" class="qz-od-table-wrap">
          <table class="qz-od-table">
            <thead>
              <tr>
                <th>{{ t('orderDetail.refundRecordTime') }}</th>
                <th>{{ t('orderDetail.refundRecordAmount') }}</th>
                <th>{{ t('orderDetail.refundRecordReason') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(record, idx) in refundRecords" :key="`refund-${idx}`">
                <td class="nowrap">{{ formatDate(record.created_at) }}</td>
                <td class="nowrap amount">{{ formatMoney(record.amount, record.currency || order.currency) }}</td>
                <td>{{ refundReasonText(record.remark) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else class="qz-od-muted">{{ t('orderDetail.refundRecordsEmpty') }}</p>
      </section>

      <!-- TIMELINE · 时间线 -->
      <section v-if="showTimeCard" class="qz-panel qz-od-panel qz-reveal is-pending">
        <div class="qz-od-panel-head">
          <h2 class="qz-eyebrow">{{ t('qiaozhi.orderDetail.sectionTimeline') }}</h2>
        </div>
        <div class="qz-od-grid is-time">
          <div class="qz-od-cell">
            <span class="lbl">{{ t('orderDetail.createdAtLabel') }}</span>
            <b class="val">{{ formatDate(order.created_at) }}</b>
          </div>
          <div v-if="order.paid_at" class="qz-od-cell">
            <span class="lbl">{{ t('orderDetail.paidAtLabel') }}</span>
            <b class="val">{{ formatDate(order.paid_at) }}</b>
          </div>
          <div v-if="order.expires_at" class="qz-od-cell">
            <span class="lbl">{{ t('orderDetail.expiresAtLabel') }}</span>
            <b class="val">{{ formatDate(order.expires_at) }}</b>
          </div>
          <div v-if="order.canceled_at" class="qz-od-cell">
            <span class="lbl">{{ t('orderDetail.canceledAtLabel') }}</span>
            <b class="val">{{ formatDate(order.canceled_at) }}</b>
          </div>
        </div>
      </section>

      <!-- 底部帮助 -->
      <p class="qz-od-help">
        {{ t('qiaozhi.orderDetail.helpText') }}
        <a v-if="supportHref" :href="supportHref" target="_blank" rel="noopener noreferrer">{{ t('qiaozhi.orderDetail.helpContact') }}</a>
        <span v-else>{{ t('qiaozhi.orderDetail.helpContact') }}</span>
        {{ t('qiaozhi.orderDetail.helpOrderNo') }} <b>{{ order.order_no }}</b>
      </p>
    </template>

    <QzFulfillDialog
      v-if="order"
      v-model:open="fulfillOpen"
      :order-no="order.order_no"
      :guest-credentials="{ email: auth.email, order_password: auth.order_password }"
      @view-order="fulfillOpen = false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Check, ChevronLeft, Clock,
  Hourglass, Info, PackageOpen, X,
} from 'lucide-vue-next'
import { useAppStore } from '../../stores/app'
import { useGuestOrderDetail } from '../../composables/useGuestOrderDetail'
import { setupReveal } from './utils/reveal'
import QzButton from './components/QzButton.vue'
import QzFulfillDialog from './components/QzFulfillDialog.vue'
import QzOrderDeliveryUnit from './components/QzOrderDeliveryUnit.vue'
import { orderHasFulfillInfoPending, orderIsFulfillProcuring } from './utils/fulfillDialog'

const { t } = useI18n()
const appStore = useAppStore()

const {
  order, authError, auth, viewState, handleAuthSubmit, clearAuth,
  fulfillmentDownloading, handleDownloadFulfillment,
  statusLabel, statusVariant,
  formatDate, formatMoney, formatDiscountMoney, hasDiscountAmount, hasAmount,
  hasActualPayment, formatPaidTotal,
  refundReasonText, showRefundRecordsCard, refundRecords, showTimeCard,
  fulfillmentCopied, handleCopyFulfillment,
} = useGuestOrderDetail()

// 支付后补填开通信息 / 查看开通进度（上游表单商品）
const fulfillOpen = ref(false)
const hasFulfillPending = computed(() => orderHasFulfillInfoPending(order.value) && order.value?.status === 'paid')
const isFulfillProcuring = computed(() => orderIsFulfillProcuring(order.value))
// 关闭弹窗后重新拉取订单：状态 / 表单提交值可能已变化。
watch(fulfillOpen, (opened, wasOpened) => {
  if (wasOpened && !opened) void handleAuthSubmit()
})

const toneClass = (status?: string | null) => {
  switch (statusVariant(status || '')) {
    case 'success': return 'is-success'
    case 'warning': return 'is-warning'
    case 'info': return 'is-info'
    case 'danger': return 'is-danger'
    case 'accent': return 'is-accent'
    default: return ''
  }
}

const statusIcon = computed(() => {
  switch (statusVariant(order.value?.status)) {
    case 'success': return Check
    case 'warning': return Clock
    case 'info': return Info
    case 'danger': return X
    case 'accent': return Hourglass
    default: return X
  }
})

const showOrderTitle = computed(() => viewState.value === 'detail' && !!order.value)

const maskedEmail = computed(() => {
  const email = String(order.value?.guest_email || auth.value.email || '').trim()
  const at = email.indexOf('@')
  if (at <= 0) return email
  return `${email.slice(0, 1)}***${email.slice(at)}`
})

const hasDeliveredFulfillment = computed(() => {
  const o = order.value
  if (!o) return false
  if (o.fulfillment?.status === 'delivered') return true
  return Array.isArray(o.children) && o.children.some((child: any) => child?.fulfillment?.status === 'delivered')
})

/** 交付成功后展示「卡密已发送至邮箱」提示（原顶部副标题迁入交付卡片）。 */
const deliveredEmailHint = computed(() => {
  if (!maskedEmail.value || !hasDeliveredFulfillment.value) return ''
  return t('qiaozhi.orderDetail.emailDeliveredHint', { email: maskedEmail.value })
})

const supportHref = computed(() => {
  const c = appStore.config?.contact as { telegram?: string; whatsapp?: string; email?: string; support_url?: string } | undefined
  if (c?.support_url) return c.support_url
  if (c?.telegram) return c.telegram
  if (c?.whatsapp) return `https://wa.me/${c.whatsapp.replace(/[^\d]/g, '')}`
  if (c?.email) return `mailto:${c.email}`
  return ''
})

// 入场动画
const rootEl = ref<HTMLElement | null>(null)
let cleanupReveal: (() => void) | null = null
const armReveal = () => {
  cleanupReveal?.()
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
}
watch(viewState, () => { armReveal() })
onMounted(() => { armReveal() })
onUnmounted(() => { cleanupReveal?.() })
</script>

<style scoped>
.qz-od {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  max-width: 760px;
}

/* 页头（含标题旁的状态徽标）用 base.css 的 .qz-page-head / .qz-badge.is-accent */

/* ---------- 游客鉴权 ---------- */
.qz-panel.qz-od-auth {
  display: grid;
  gap: var(--qz-space-4);
  border-style: dashed;
  border-color: var(--qz-border-strong);
}

.qz-od-auth-head {
  display: grid;
  gap: var(--qz-space-1);
}

.qz-od-auth-head h2 {
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: var(--qz-leading-body);
}

.qz-od-auth-head p {
  margin: 0;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-od-auth-form {
  display: grid;
  gap: var(--qz-space-3);
}

.qz-od-auth-fields {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: var(--qz-space-3);
}

.qz-od-auth-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2);
}

/* ---------- 骨架：行与卡块都是 base 的 .qz-skeleton，这里只给尺寸 ---------- */
.qz-od-skeleton {
  display: grid;
  gap: var(--qz-space-4);
}

.qz-skeleton.qz-od-sk-line {
  width: 60%;
}

/* 卡块占位跟面板同圆角：覆盖基类的 control 圆角，写成 (0,3,0) */
.qz-skeleton.qz-od-sk-card {
  height: 150px;
  border-radius: var(--qz-radius-panel);
}

/* ---------- 面板 ---------- */
.qz-od-panel-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--qz-space-2) var(--qz-space-3);
  margin-bottom: var(--qz-space-4);
}

.qz-od-panel-head h2 {
  margin: 0;
}

/* 订单摘要行 */
.qz-od-rows {
  display: grid;
  gap: var(--qz-space-1);
  margin: 0;
}

.qz-od-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--qz-space-3);
  padding: var(--qz-space-1) 0;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
}

.qz-od-row dd {
  min-width: 0;
  margin: 0;
  color: var(--qz-text);
  font-weight: 600;
  text-align: right;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.qz-od-no {
  font-family: var(--qz-font-mono);
  font-size: var(--qz-text-xs);
  letter-spacing: 0.02em;
}

.qz-od-actions-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2);
  margin-top: var(--qz-space-4);
  padding-top: var(--qz-space-3);
  border-top: 1px dashed var(--qz-border-strong);
}

/* 空数据提示：面板里以 muted 文案收尾（如无退款记录） */
.qz-od-muted {
  margin: 0;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

/* ---------- 金额 / 时间网格 ---------- */
.qz-od-grid {
  display: grid;
  gap: var(--qz-space-2);
}

.qz-od-grid.is-amount {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.qz-od-grid.is-secondary {
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  margin-top: var(--qz-space-2);
}

.qz-od-grid.is-time {
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
}

.qz-od-cell {
  display: grid;
  gap: var(--qz-space-1);
  padding: var(--qz-space-3) var(--qz-space-4);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
}

.qz-od-cell .lbl {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
  font-weight: 700;
  letter-spacing: 0.04em;
}

.qz-od-cell .val {
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.qz-od-grid.is-amount .qz-od-cell .val {
  font-size: var(--qz-text-base);
}

.qz-od-cell.is-disc .val {
  color: var(--qz-danger);
}

.qz-od-disc-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--qz-space-2);
  margin-top: var(--qz-space-2);
}

/* ---------- 退款记录表 ---------- */
.qz-od-table-wrap {
  overflow-x: auto;
}

.qz-od-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-od-table th {
  padding: var(--qz-space-2) var(--qz-space-3);
  background: var(--qz-surface-muted);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
  font-weight: 700;
  letter-spacing: 0.04em;
  text-align: left;
}

.qz-od-table th:first-child {
  border-radius: var(--qz-radius-sm) 0 0 var(--qz-radius-sm);
}

.qz-od-table th:last-child {
  border-radius: 0 var(--qz-radius-sm) var(--qz-radius-sm) 0;
}

.qz-od-table td {
  padding: var(--qz-space-2) var(--qz-space-3);
  border-top: 1px solid var(--qz-border);
  color: var(--qz-text-secondary);
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.qz-od-table td.nowrap {
  white-space: nowrap;
}

.qz-od-table .amount {
  color: var(--qz-text);
  font-weight: 700;
}

/* ---------- 底部帮助 ---------- */
.qz-od-help {
  margin: 0;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: 1.7;
  text-align: center;
}

.qz-od-help a,
.qz-od-help b {
  color: var(--qz-text);
  font-weight: 700;
}

.qz-od-help a {
  text-decoration: underline;
  text-underline-offset: 3px;
}

@media (max-width: 760px) {
  .qz-od {
    gap: var(--qz-space-3);
  }

  .qz-panel.qz-od-panel {
    padding: var(--qz-space-4);
  }

  .qz-od-auth-fields {
    grid-template-columns: minmax(0, 1fr);
  }

  .qz-od-grid.is-amount {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
