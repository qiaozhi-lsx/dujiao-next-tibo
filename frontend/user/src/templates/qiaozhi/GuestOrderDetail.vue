<template>
  <div ref="rootEl" class="qz-container qz-order">
    <!-- 顶部：返回 + 订单状态 -->
    <div class="qz-order-topbar">
      <RouterLink class="qz-order-back" to="/guest/orders">
        <ArrowLeft />
        {{ t('qiaozhi.orderDetail.backToOrders') }}
      </RouterLink>
      <span v-if="showOrderTitle" class="qz-order-status" :class="toneClass(order?.status)">
        <component :is="statusIcon" />
        {{ statusLabel(order?.status) }}
      </span>
    </div>
    <h1 class="qz-sr-only">{{ t('guestOrderDetail.title') }}</h1>

    <!-- 游客鉴权 -->
    <section v-if="viewState === 'auth'" class="qz-order-auth qz-reveal is-pending">
      <div class="qz-order-auth-head">
        <b>{{ t('guestOrderDetail.authTitle') }}</b>
        <small>{{ t('guestOrderDetail.authHint') }}</small>
      </div>
      <form class="qz-ic" novalidate @submit.prevent="handleAuthSubmit">
        <div class="qz-ic-row">
          <div class="qz-ic-field">
            <label for="qz-order-guest-email">{{ t('guestOrders.emailPlaceholder') }}</label>
            <input
              id="qz-order-guest-email"
              v-model="auth.email"
              type="email"
              autocomplete="email"
              :placeholder="t('guestOrders.emailPlaceholder')"
            />
          </div>
          <div class="qz-ic-field">
            <label for="qz-order-guest-password">{{ t('guestOrders.passwordPlaceholder') }}</label>
            <input
              id="qz-order-guest-password"
              v-model="auth.order_password"
              type="password"
              autocomplete="current-password"
              :placeholder="t('guestOrders.passwordPlaceholder')"
            />
          </div>
        </div>
        <p v-if="authError" class="qz-ic-alert is-error" role="alert">{{ authError }}</p>
        <div class="qz-order-auth-actions">
          <button type="submit" class="qz-order-btn is-primary">
            {{ t('guestOrderDetail.authSubmit') }}
          </button>
          <button type="button" class="qz-order-auth-clear" @click="clearAuth">
            {{ t('guestOrderDetail.authClear') }}
          </button>
        </div>
      </form>
    </section>

    <!-- 骨架 -->
    <div v-else-if="viewState === 'loading'" class="qz-pd-skeleton">
      <div class="qz-pd-sk-line is-lg"></div>
      <div class="qz-pd-sk-line"></div>
      <div class="qz-order-sk-card"></div>
      <div class="qz-order-sk-card"></div>
    </div>

    <!-- 订单不存在 -->
    <div v-else-if="viewState === 'empty' || !order" class="qz-empty qz-reveal is-pending">
      <PackageOpen />
      <b>{{ t('guestOrderDetail.notFound') }}</b>
    </div>

    <template v-else>
      <!-- ORDER · 订单摘要 -->
      <section class="qz-order-card qz-reveal is-pending">
        <div class="qz-order-card-tag">
          <span>{{ t('qiaozhi.orderDetail.sectionOrder') }}</span>
        </div>
        <div class="qz-order-rows">
          <div class="qz-order-row">
            <span>{{ t('orders.orderNo') }}</span>
            <b class="qz-order-no">{{ order.order_no }}</b>
          </div>
          <div class="qz-order-row">
            <span>{{ t('orderDetail.createdAtLabel') }}</span>
            <b>{{ formatDate(order.created_at) }}</b>
          </div>
          <div class="qz-order-row">
            <span>{{ t('orderDetail.amountTotal') }}</span>
            <b>{{ formatPaidTotal(order) }}</b>
          </div>
        </div>
        <div v-if="order.status === 'pending_payment'" class="qz-order-card-actions">
          <RouterLink class="qz-order-btn is-primary" :to="`/pay?guest=1&order_no=${order.order_no}`">
            {{ t('orders.payNow') }}
          </RouterLink>
        </div>
      </section>

      <!-- DELIVERY · 交付详情 -->
      <section class="qz-order-card qz-reveal is-pending">
        <div class="qz-order-card-h2">
          <span>{{ t('qiaozhi.orderDetail.sectionDelivery') }}</span>
          <div v-if="singleFulfillment" class="qz-order-actions">
            <QzButton
              v-if="canCopy(singleFulfillment.fulfillment)"
              size="sm"
              @click="handleCopyFulfillment(singleFulfillment.fulfillment)"
            >
              <Check v-if="fulfillmentCopied" />
              <Copy v-else />
              <span>{{ fulfillmentCopied ? t('orderDetail.fulfillmentCopied') : t('qiaozhi.orderDetail.copyPayload') }}</span>
            </QzButton>
            <QzButton
              v-if="isFulfillmentTruncated(singleFulfillment.fulfillment)"
              size="sm"
              variant="secondary"
              :disabled="fulfillmentDownloading"
              @click="handleDownloadFulfillment(singleFulfillment.orderNo)"
            >
              <Download />
              <span>{{ fulfillmentDownloading ? t('orderDetail.fulfillmentDownloading') : t('orderDetail.fulfillmentDownload') }}</span>
            </QzButton>
          </div>
        </div>

        <p v-if="deliveredEmailHint" class="qz-order-deliver-hint">{{ deliveredEmailHint }}</p>

        <div v-for="unit in deliveryUnits" :key="unit.key" class="qz-order-unit">
          <!-- 子订单头 -->
          <div v-if="unit.isChild" class="qz-order-unit-head">
            <div class="qz-order-unit-order">
              <span>{{ t('orderDetail.childOrderNo') }}</span>
              <b>{{ unit.orderNo }}</b>
              <span>· {{ t('orderDetail.childOrderAmount') }} {{ formatMoney(unit.orderAmount, unit.currency) }}</span>
            </div>
            <span v-if="unit.status && unit.status !== order.status" class="qz-order-badge" :class="toneClass(unit.status)">{{ statusLabel(unit.status) }}</span>
          </div>

          <!-- 多单元时每单自己的复制 / 下载 -->
          <div v-if="deliveryUnits.length > 1 && unit.fulfillment" class="qz-order-actions">
            <QzButton
              v-if="canCopy(unit.fulfillment)"
              size="sm"
              @click="handleCopyFulfillment(unit.fulfillment)"
            >
              <Check v-if="fulfillmentCopied" />
              <Copy v-else />
              <span>{{ fulfillmentCopied ? t('orderDetail.fulfillmentCopied') : t('qiaozhi.orderDetail.copyPayload') }}</span>
            </QzButton>
            <QzButton
              v-if="isFulfillmentTruncated(unit.fulfillment)"
              size="sm"
              variant="secondary"
              :disabled="fulfillmentDownloading"
              @click="handleDownloadFulfillment(unit.orderNo)"
            >
              <Download />
              <span>{{ fulfillmentDownloading ? t('orderDetail.fulfillmentDownloading') : t('orderDetail.fulfillmentDownload') }}</span>
            </QzButton>
          </div>

          <!-- 商品回显 -->
          <div v-if="unit.items.length" class="qz-order-items">
            <div v-for="(item, idx) in unit.items" :key="`${unit.key}-item-${idx}`" class="qz-order-item">
              <div class="qz-order-thumb">
                <img
                  v-if="orderItemImage(item)"
                  :src="orderItemImage(item)"
                  :alt="getLocalizedText(item.title)"
                  loading="lazy"
                  decoding="async"
                />
                <span v-else-if="itemInitial(item)">{{ itemInitial(item) }}</span>
                <ImageIcon v-else />
              </div>
              <div class="qz-order-item-main">
                <div class="qz-order-item-name">{{ getLocalizedText(item.title) }}</div>
                <div class="qz-order-item-tags">
                  <span v-if="orderItemSkuText(item)" class="tag">{{ t('orderDetail.itemSkuLabel') }}：{{ orderItemSkuText(item) }}</span>
                  <span class="tag">× {{ item.quantity }}</span>
                  <span class="tag">{{ fulfillmentTypeLabelText(item.fulfillment_type) }}</span>
                  <span v-for="(tag, tagIdx) in item.tags || []" :key="`tag-${idx}-${tagIdx}`" class="tag">{{ tag }}</span>
                </div>
                <div
                  v-if="manualSubmissionRows(item.manual_form_submission, item.manual_form_schema_snapshot).length"
                  class="qz-order-manual"
                >
                  <b>{{ t('orderDetail.manualSubmissionTitle') }}</b>
                  <div
                    v-for="row in manualSubmissionRows(item.manual_form_submission, item.manual_form_schema_snapshot)"
                    :key="row.key"
                  >
                    <span class="k">{{ row.label }}</span>：{{ row.value }}
                  </div>
                </div>
                <p v-if="unit.items.length === 1 && unitNote(unit)" class="qz-order-item-note">{{ unitNote(unit) }}</p>
              </div>
              <div class="qz-order-item-amt">
                <b :title="t('orderDetail.itemPaidAmountLabel')">{{ formatItemPaidAmount(item, unit.currency) }}</b>
                <span v-if="hasItemDiscount(item)" class="orig">{{ formatMoney(item.original_total_price, unit.currency) }}</span>
                <span v-if="hasDiscountAmount(item.coupon_discount_amount)" class="disc">
                  − {{ formatMoney(item.coupon_discount_amount, unit.currency) }} {{ t('orderDetail.couponDiscountLabel') }}
                </span>
                <span v-if="hasDiscountAmount(item.promotion_discount_amount)" class="disc">
                  − {{ formatMoney(item.promotion_discount_amount, unit.currency) }} {{ t('orderDetail.promotionDiscountLabel') }}
                </span>
                <span v-if="hasDiscountAmount(item.wholesale_discount_amount)" class="disc">
                  − {{ formatMoney(item.wholesale_discount_amount, unit.currency) }} {{ t('orderDetail.wholesaleDiscountLabel') }}
                </span>
                <span v-if="hasDiscountAmount(item.member_discount_amount)" class="disc">
                  − {{ formatMoney(item.member_discount_amount, unit.currency) }} {{ t('orderDetail.memberDiscountLabel') }}
                </span>
                <span v-if="item.quantity > 1" class="unit">
                  {{ t('orderDetail.unitPriceLabel') }} {{ formatMoney(item.original_unit_price, unit.currency) }} × {{ item.quantity }}
                </span>
              </div>
            </div>
          </div>
          <p v-else-if="!unit.fulfillment" class="qz-order-ful-empty">{{ t('orderDetail.noItems') }}</p>

          <!-- 交付内容 -->
          <div v-if="unit.fulfillment" class="qz-order-ful">
            <div class="qz-order-ful-bar">
              <span class="qz-order-ful-meta">{{ fulfillmentInfoText(unit.fulfillment) }}</span>
            </div>
            <p v-if="unit.items.length !== 1 && unitNote(unit)" class="qz-order-item-note">{{ unitNote(unit) }}</p>
            <div v-if="isFulfillmentTruncated(unit.fulfillment)" class="qz-order-trunc">
              <AlertTriangle />
              <span>{{ t('orderDetail.fulfillmentTruncatedHint') }}</span>
            </div>
            <div class="qz-order-payload">
              <template v-if="isFulfillmentTruncated(unit.fulfillment)">{{ unit.fulfillment.payload }}</template>
              <template v-else-if="payloadRows(unit.fulfillment).length">
                <div v-for="(row, rowIdx) in payloadRows(unit.fulfillment)" :key="`${unit.key}-row-${rowIdx}`">
                  <span v-if="row.key" class="k">{{ row.key }}</span>{{ row.key ? ': ' : '' }}{{ row.value }}
                </div>
              </template>
              <template v-else>{{ unit.fulfillment.payload }}</template>
            </div>
          </div>
          <p v-else class="qz-order-ful-empty">{{ t('orderDetail.childFulfillmentEmpty') }}</p>

          <!-- 使用说明 -->
          <template v-if="unit.fulfillment?.status === 'delivered' && instructionBlocks(unit.items).length">
            <div v-for="(block, bi) in instructionBlocks(unit.items)" :key="`${unit.key}-inst-${bi}`" class="qz-order-inst">
              <div class="qz-order-inst-head">
                <Info />
                {{ t('qiaozhi.orderDetail.instructionsTitle') }}
              </div>
              <div class="qz-order-inst-body" v-html="block.html"></div>
            </div>
          </template>
        </div>
      </section>

      <!-- BILLING · 金额明细 -->
      <section class="qz-order-card qz-reveal is-pending">
        <div class="qz-order-card-h2">
          <span>{{ t('qiaozhi.orderDetail.sectionBilling') }}</span>
        </div>
        <div class="qz-order-amount-grid">
          <div class="qz-order-amount-cell">
            <div class="lbl">{{ t('orderDetail.amountOriginal') }}</div>
            <div class="val">{{ formatMoney(order.original_amount, order.currency) }}</div>
          </div>
          <div class="qz-order-amount-cell is-disc">
            <div class="lbl">{{ t('orderDetail.amountDiscount') }}</div>
            <div class="val">{{ formatDiscountMoney(order.discount_amount, order.currency) }}</div>
          </div>
          <div class="qz-order-amount-cell">
            <div class="lbl">{{ t('orderDetail.amountTotal') }}</div>
            <div class="val">{{ formatPaidTotal(order) }}</div>
          </div>
        </div>
        <div
          v-if="hasActualPayment(order) && (hasAmount(order.wallet_paid_amount) || hasAmount(order.online_paid_amount) || hasAmount(order.refunded_amount))"
          class="qz-order-amount-grid is-secondary"
        >
          <div v-if="hasAmount(order.wallet_paid_amount)" class="qz-order-amount-cell">
            <div class="lbl">{{ t('orderDetail.amountWalletPaid') }}</div>
            <div class="val">{{ formatMoney(order.wallet_paid_amount, order.currency) }}</div>
          </div>
          <div v-if="hasAmount(order.online_paid_amount)" class="qz-order-amount-cell">
            <div class="lbl">{{ t('orderDetail.amountOnlinePaid') }}</div>
            <div class="val">{{ formatMoney(order.online_paid_amount, order.currency) }}</div>
          </div>
          <div v-if="hasAmount(order.refunded_amount)" class="qz-order-amount-cell">
            <div class="lbl">{{ t('orderDetail.amountRefunded') }}</div>
            <div class="val">{{ formatMoney(order.refunded_amount, order.currency) }}</div>
          </div>
        </div>
        <div
          v-if="hasDiscountAmount(order.member_discount_amount) || hasDiscountAmount(order.wholesale_discount_amount)"
          class="qz-order-disc-chips"
        >
          <span v-if="hasDiscountAmount(order.member_discount_amount)" class="qz-order-chip is-member">
            {{ t('orderDetail.amountMemberDiscount') }} {{ formatDiscountMoney(order.member_discount_amount, order.currency) }}
          </span>
          <span v-if="hasDiscountAmount(order.wholesale_discount_amount)" class="qz-order-chip is-wholesale">
            {{ t('orderDetail.amountWholesaleDiscount') }} {{ formatDiscountMoney(order.wholesale_discount_amount, order.currency) }}
          </span>
        </div>
      </section>

      <!-- REFUNDS · 退款记录 -->
      <section v-if="showRefundRecordsCard" class="qz-order-card qz-reveal is-pending">
        <div class="qz-order-card-h2">
          <span>{{ t('qiaozhi.orderDetail.sectionRefunds') }}</span>
        </div>
        <div v-if="refundRecords.length" class="qz-order-table-wrap">
          <table class="qz-order-table">
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
        <p v-else class="qz-order-empty-text">{{ t('orderDetail.refundRecordsEmpty') }}</p>
      </section>

      <!-- TIMELINE · 时间线 -->
      <section v-if="showTimeCard" class="qz-order-card qz-reveal is-pending">
        <div class="qz-order-card-h2">
          <span>{{ t('qiaozhi.orderDetail.sectionTimeline') }}</span>
        </div>
        <div class="qz-order-time-grid">
          <div class="qz-order-time-cell">
            <div class="lbl">{{ t('orderDetail.createdAtLabel') }}</div>
            <div class="val">{{ formatDate(order.created_at) }}</div>
          </div>
          <div v-if="order.paid_at" class="qz-order-time-cell">
            <div class="lbl">{{ t('orderDetail.paidAtLabel') }}</div>
            <div class="val">{{ formatDate(order.paid_at) }}</div>
          </div>
          <div v-if="order.expires_at" class="qz-order-time-cell">
            <div class="lbl">{{ t('orderDetail.expiresAtLabel') }}</div>
            <div class="val">{{ formatDate(order.expires_at) }}</div>
          </div>
          <div v-if="order.canceled_at" class="qz-order-time-cell">
            <div class="lbl">{{ t('orderDetail.canceledAtLabel') }}</div>
            <div class="val">{{ formatDate(order.canceled_at) }}</div>
          </div>
        </div>
      </section>

      <!-- 底部帮助 -->
      <p class="qz-order-help">
        {{ t('qiaozhi.orderDetail.helpText') }}
        <a v-if="supportHref" :href="supportHref" target="_blank" rel="noopener noreferrer">{{ t('qiaozhi.orderDetail.helpContact') }}</a>
        <span v-else>{{ t('qiaozhi.orderDetail.helpContact') }}</span>
        {{ t('qiaozhi.orderDetail.helpOrderNo') }} <b>{{ order.order_no }}</b>
      </p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  AlertTriangle, ArrowLeft, Check, Clock, Copy, Download,
  Hourglass, Image as ImageIcon, Info, PackageOpen, X,
} from 'lucide-vue-next'
import { useAppStore } from '../../stores/app'
import { useGuestOrderDetail } from '../../composables/useGuestOrderDetail'
import { setupReveal } from './utils/reveal'
import { buildDeliveryUnits, type DeliveryUnit } from './utils/deliveryUnits'
import QzButton from './components/QzButton.vue'

const { t } = useI18n()
const appStore = useAppStore()

const {
  order, authError, auth, viewState, handleAuthSubmit, clearAuth,
  fulfillmentDownloading, handleDownloadFulfillment,
  statusLabel, statusVariant, fulfillmentTypeLabelText, fulfillmentStatusLabelText,
  formatDate, getLocalizedText, formatMoney, formatDiscountMoney, hasDiscountAmount, hasAmount,
  hasActualPayment, formatPaidTotal,
  refundReasonText, showRefundRecordsCard, refundRecords, showTimeCard,
  orderItemImage, orderItemSkuText, manualSubmissionRows,
  hasItemDiscount, formatItemPaidAmount, resolvedChildStatus,
  instructionBlocks, isFulfillmentTruncated,
  fulfillmentCopied, handleCopyFulfillment,
} = useGuestOrderDetail()

const deliveryUnits = computed<DeliveryUnit[]>(() => buildDeliveryUnits(order.value, resolvedChildStatus))

/** 单单元（无子订单或多交付）时，卡头直接放复制 / 下载按钮。 */
const singleFulfillment = computed<{ orderNo: string; fulfillment: any } | null>(() => {
  const list = deliveryUnits.value
  if (list.length !== 1) return null
  const unit = list[0]
  if (!unit || !unit.fulfillment) return null
  return { orderNo: unit.orderNo, fulfillment: unit.fulfillment }
})

const toneClass = (status?: string | null) => {
  switch (statusVariant(status || '')) {
    case 'success': return 'is-ok'
    case 'warning': return 'is-warn'
    case 'info': return 'is-info'
    case 'danger': return 'is-danger'
    case 'accent': return 'is-accent'
    default: return 'is-neutral'
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

const canCopy = (fulfillment: any) => fulfillment?.status === 'delivered' && !isFulfillmentTruncated(fulfillment)

const itemInitial = (item: any) => String(getLocalizedText(item?.title) || '').trim().slice(0, 1).toUpperCase()

interface PayloadRow {
  key: string
  value: string
}

const unitNote = (unit: DeliveryUnit) => {
  const data = unit.fulfillment?.delivery_data
  return String(data?.note || '').trim()
}

const payloadRows = (fulfillment: any): PayloadRow[] => {
  const data = fulfillment?.delivery_data
  const rows: PayloadRow[] = []
  if (data && typeof data === 'object' && Array.isArray(data.entries)) {
    for (const entry of data.entries) {
      const key = String(entry?.key || '').trim()
      const value = String(entry?.value || '').trim()
      if (!key && !value) continue
      rows.push({ key, value })
    }
  }
  return rows
}

const payloadMetaText = (fulfillment: any) => {
  const count = Number(fulfillment?.payload_line_count || 0)
  if (!count) return ''
  if (fulfillment?.delivered_at) {
    return t('qiaozhi.orderDetail.payloadMeta', { count, time: formatDate(fulfillment.delivered_at) })
  }
  return t('orderDetail.fulfillmentTotalLines', { count })
}

const fulfillmentInfoText = (fulfillment: any) => {
  const parts = [fulfillmentTypeLabelText(fulfillment?.type), fulfillmentStatusLabelText(fulfillment?.status)]
  const meta = payloadMetaText(fulfillment)
  if (meta) parts.push(meta)
  return parts.filter(Boolean).join(' · ')
}

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
