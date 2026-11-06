<template>
  <div ref="rootEl" class="qz-container qz-od">
    <!-- 页头：eyebrow + h1（含订单状态）+ 说明 + 返回 -->
    <header class="qz-od-head">
      <div class="qz-od-head-copy">
        <span class="qz-eyebrow">{{ t('orders.title') }}</span>
        <h1>
          <span>{{ t('orderDetail.title') }}</span>
          <span v-if="showOrderTitle" class="qz-badge qz-od-status" :class="toneClass(order?.status)">
            <component :is="statusIcon" />
            {{ statusLabel(order?.status) }}
          </span>
        </h1>
        <p>{{ t('orderDetail.subtitle') }}</p>
      </div>
      <RouterLink class="qz-btn qz-btn--secondary qz-btn--sm" to="/me/orders">
        <ChevronLeft />
        {{ t('qiaozhi.orderDetail.backToOrders') }}
      </RouterLink>
    </header>

    <!-- 骨架 -->
    <div v-if="loading" class="qz-od-skeleton" aria-hidden="true">
      <div class="qz-od-sk-line"></div>
      <div class="qz-od-sk-card"></div>
      <div class="qz-od-sk-card"></div>
    </div>

    <!-- 订单不存在 -->
    <div v-else-if="!order" class="qz-empty qz-reveal is-pending">
      <PackageOpen />
      <b>{{ t('orderDetail.notFound') }}</b>
      <QzButton variant="primary" @click="debouncedLoadOrder()">
        {{ t('errorBoundary.retry') }}
      </QzButton>
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
          <RouterLink class="qz-btn qz-btn--primary qz-btn--sm" :to="`/pay?order_no=${order.order_no}`">
            {{ t('orderDetail.payNow') }}
          </RouterLink>
          <QzButton variant="danger" size="sm" @click="cancelOrder">
            {{ t('orderDetail.cancel') }}
          </QzButton>
        </div>
      </section>

      <!-- DELIVERY · 交付详情 -->
      <section class="qz-panel qz-od-panel qz-reveal is-pending">
        <div class="qz-od-panel-head">
          <h2 class="qz-eyebrow">{{ t('qiaozhi.orderDetail.sectionDelivery') }}</h2>
          <div v-if="singleFulfillment" class="qz-od-actions">
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

        <div v-for="unit in deliveryUnits" :key="unit.key" class="qz-od-unit">
          <!-- 子订单头 -->
          <div v-if="unit.isChild" class="qz-od-unit-head">
            <div class="qz-od-unit-order">
              <span>{{ t('orderDetail.childOrderNo') }}</span>
              <b>{{ unit.orderNo }}</b>
              <span>· {{ t('orderDetail.childOrderAmount') }} {{ formatMoney(unit.orderAmount, unit.currency) }}</span>
            </div>
            <span v-if="unit.status && unit.status !== order.status" class="qz-badge" :class="toneClass(unit.status)">{{ statusLabel(unit.status) }}</span>
          </div>

          <!-- 多单元时每单自己的复制 / 下载 -->
          <div v-if="deliveryUnits.length > 1 && unit.fulfillment" class="qz-od-actions">
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
          <div v-if="unit.items.length" class="qz-od-items">
            <div v-for="(item, idx) in unit.items" :key="`${unit.key}-item-${idx}`" class="qz-od-item">
              <div class="qz-od-thumb">
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
              <div class="qz-od-item-main">
                <div class="qz-od-item-name">{{ getLocalizedText(item.title) }}</div>
                <div class="qz-od-item-tags">
                  <span v-if="orderItemSkuText(item)" class="qz-chip">{{ t('orderDetail.itemSkuLabel') }}：{{ orderItemSkuText(item) }}</span>
                  <span class="qz-chip">× {{ item.quantity }}</span>
                  <span class="qz-chip">{{ fulfillmentTypeLabelText(item.fulfillment_type) }}</span>
                  <span v-for="(tag, tagIdx) in item.tags || []" :key="`tag-${idx}-${tagIdx}`" class="qz-chip">{{ tag }}</span>
                </div>
                <div
                  v-if="manualSubmissionRows(item.manual_form_submission, item.manual_form_schema_snapshot).length"
                  class="qz-od-manual"
                >
                  <b>{{ t('orderDetail.manualSubmissionTitle') }}</b>
                  <div
                    v-for="row in manualSubmissionRows(item.manual_form_submission, item.manual_form_schema_snapshot)"
                    :key="row.key"
                  >
                    <span class="k">{{ row.label }}</span>：{{ row.value }}
                  </div>
                </div>
                <p v-if="unit.items.length === 1 && unitNote(unit)" class="qz-od-note">{{ unitNote(unit) }}</p>
              </div>
              <div class="qz-od-item-amt">
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
          <p v-else-if="!unit.fulfillment" class="qz-od-muted">{{ t('orderDetail.noItems') }}</p>

          <!-- 交付内容 -->
          <div v-if="unit.fulfillment" class="qz-od-ful">
            <span class="qz-od-ful-meta">{{ fulfillmentInfoText(unit.fulfillment) }}</span>
            <p v-if="unit.items.length !== 1 && unitNote(unit)" class="qz-od-note">{{ unitNote(unit) }}</p>
            <div v-if="isFulfillmentTruncated(unit.fulfillment)" class="qz-od-trunc">
              <AlertTriangle />
              <span>{{ t('orderDetail.fulfillmentTruncatedHint') }}</span>
            </div>
            <div class="qz-od-payload">
              <template v-if="isFulfillmentTruncated(unit.fulfillment)">{{ unit.fulfillment.payload }}</template>
              <template v-else-if="payloadRows(unit.fulfillment).length">
                <div v-for="(row, rowIdx) in payloadRows(unit.fulfillment)" :key="`${unit.key}-row-${rowIdx}`">
                  <span v-if="row.key" class="k">{{ row.key }}</span>{{ row.key ? ': ' : '' }}{{ row.value }}
                </div>
              </template>
              <template v-else>{{ unit.fulfillment.payload }}</template>
            </div>
          </div>
          <p v-else class="qz-od-muted">{{ t('orderDetail.childFulfillmentEmpty') }}</p>

          <!-- 使用说明 -->
          <template v-if="unit.fulfillment?.status === 'delivered' && instructionBlocks(unit.items).length">
            <div v-for="(block, bi) in instructionBlocks(unit.items)" :key="`${unit.key}-inst-${bi}`" class="qz-od-inst">
              <div class="qz-od-inst-head">
                <Info />
                {{ t('qiaozhi.orderDetail.instructionsTitle') }}
              </div>
              <div class="qz-od-inst-body" v-html="block.html"></div>
            </div>
          </template>
        </div>
      </section>

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
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  AlertTriangle, Check, ChevronLeft, Clock, Copy, Download,
  Hourglass, Image as ImageIcon, Info, PackageOpen, X,
} from 'lucide-vue-next'
import { useAppStore } from '../../stores/app'
import { useOrderDetail } from '../../composables/useOrderDetail'
import { setupReveal } from './utils/reveal'
import { buildDeliveryUnits, type DeliveryUnit } from './utils/deliveryUnits'
import QzButton from './components/QzButton.vue'

const { t } = useI18n()
const appStore = useAppStore()

const {
  loading, order, debouncedLoadOrder, cancelOrder,
  fulfillmentDownloading, handleDownloadFulfillment,
  statusLabel, statusVariant, fulfillmentTypeLabelText, fulfillmentStatusLabelText,
  formatDate, getLocalizedText, formatMoney, formatDiscountMoney, hasDiscountAmount, hasAmount,
  hasActualPayment, formatPaidTotal,
  refundReasonText, showRefundRecordsCard, refundRecords, showTimeCard,
  orderItemImage, orderItemSkuText, manualSubmissionRows,
  hasItemDiscount, formatItemPaidAmount, resolvedChildStatus,
  instructionBlocks, isFulfillmentTruncated,
  fulfillmentCopied, handleCopyFulfillment,
} = useOrderDetail()

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

const showOrderTitle = computed(() => !loading.value && !!order.value)

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
watch(loading, (value) => { if (!value) armReveal() })
onMounted(() => { if (!loading.value) armReveal() })
onUnmounted(() => { cleanupReveal?.() })
</script>

<style scoped>
.qz-od {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  max-width: 760px;
}

/* ---------- 页头 ---------- */
.qz-od-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--qz-space-4);
}

.qz-od-head-copy {
  min-width: 0;
}

.qz-od-head h1 {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2) var(--qz-space-3);
  margin: var(--qz-space-1) 0 0;
  color: var(--qz-text);
  font-size: var(--qz-text-2xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
}

.qz-od-head p {
  margin: var(--qz-space-2) 0 0;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
}

.qz-od-head > .qz-btn {
  flex: none;
}

/* 标题旁的状态徽标：基类只带四种状态色，「处理中」用主色底 */
.qz-od-status {
  letter-spacing: normal;
}

.qz-od-status svg {
  width: 12px;
  height: 12px;
}

.qz-badge.is-accent {
  color: var(--qz-accent-strong);
  background: var(--qz-accent-subtle);
}

/* ---------- 骨架 ---------- */
.qz-od-skeleton {
  display: grid;
  gap: var(--qz-space-4);
}

.qz-od-sk-line,
.qz-od-sk-card {
  background: var(--qz-surface-muted);
  animation: qz-od-pulse 1.4s ease-in-out infinite;
}

.qz-od-sk-line {
  width: 60%;
  height: 18px;
  border-radius: var(--qz-radius-pill);
}

.qz-od-sk-card {
  height: 150px;
  border-radius: var(--qz-radius-panel);
}

@keyframes qz-od-pulse {
  50% {
    opacity: 0.55;
  }
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

.qz-od-actions {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--qz-space-2);
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
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
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

/* ---------- 交付单元 ---------- */
.qz-od-unit {
  display: grid;
  gap: var(--qz-space-3);
}

.qz-od-unit + .qz-od-unit {
  margin-top: var(--qz-space-4);
  padding-top: var(--qz-space-4);
  border-top: 1px dashed var(--qz-border-strong);
}

.qz-od-unit-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--qz-space-2) var(--qz-space-3);
}

.qz-od-unit-order {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: var(--qz-space-2);
  min-width: 0;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-od-unit-order b {
  color: var(--qz-text);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-weight: 600;
  overflow-wrap: anywhere;
}

/* 商品回显 */
.qz-od-items {
  display: grid;
  gap: var(--qz-space-2);
}

.qz-od-item {
  display: flex;
  align-items: flex-start;
  gap: var(--qz-space-3);
  padding: var(--qz-space-3);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
}

.qz-od-thumb {
  display: grid;
  flex: none;
  place-items: center;
  width: 52px;
  height: 52px;
  overflow: hidden;
  border: 1px solid var(--qz-border);
  border-radius: 12px;
  background: var(--qz-surface);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-lg);
  font-weight: 700;
}

.qz-od-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.qz-od-thumb svg {
  width: 20px;
  height: 20px;
}

.qz-od-item-main {
  display: grid;
  flex: 1 1 auto;
  gap: var(--qz-space-1);
  min-width: 0;
}

.qz-od-item-name {
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  line-height: var(--qz-leading-body);
  overflow-wrap: anywhere;
}

.qz-od-item-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--qz-space-1);
}

/* 商品卡本身已是 muted 底，chip 换回 surface 底才分得开；规格文案可能很长，允许折行而不是压到金额列上 */
.qz-od-item-tags .qz-chip {
  padding-block: var(--qz-space-1);
  background: var(--qz-surface);
  line-height: var(--qz-leading-body);
  white-space: normal;
  overflow-wrap: anywhere;
}

.qz-od-note {
  margin: 0;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: 1.6;
  overflow-wrap: anywhere;
}

.qz-od-manual {
  display: grid;
  gap: 2px;
  margin-top: var(--qz-space-1);
  padding: var(--qz-space-2) var(--qz-space-3);
  border: 1px solid var(--qz-border);
  border-radius: 12px;
  background: var(--qz-surface);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  overflow-wrap: anywhere;
}

.qz-od-manual > b {
  color: var(--qz-text-muted);
  font-weight: 700;
}

.qz-od-manual .k {
  color: var(--qz-text);
  font-weight: 600;
}

.qz-od-item-amt {
  display: grid;
  flex: none;
  justify-items: end;
  gap: 2px;
  text-align: right;
}

.qz-od-item-amt > b {
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.qz-od-item-amt .orig {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
  font-weight: 600;
  text-decoration: line-through;
}

.qz-od-item-amt .disc {
  color: var(--qz-danger);
  font-size: var(--qz-text-2xs);
  font-weight: 600;
}

.qz-od-item-amt .unit {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
}

/* 交付内容 */
.qz-od-ful {
  display: grid;
  gap: var(--qz-space-2);
}

.qz-od-ful-meta {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  line-height: var(--qz-leading-body);
}

.qz-od-muted {
  margin: 0;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-od-trunc {
  display: flex;
  align-items: flex-start;
  gap: var(--qz-space-2);
  padding: var(--qz-space-2) var(--qz-space-3);
  border: 1px solid color-mix(in oklab, var(--qz-warning) 30%, transparent);
  border-radius: 12px;
  background: color-mix(in oklab, var(--qz-warning) 12%, transparent);
  color: var(--qz-warning);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  line-height: var(--qz-leading-body);
}

.qz-od-trunc svg {
  flex: none;
  width: 14px;
  height: 14px;
  margin-top: 1px;
}

.qz-od-payload {
  max-height: 260px;
  overflow-y: auto;
  padding: var(--qz-space-3) var(--qz-space-4);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  color: var(--qz-text);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: var(--qz-text-xs);
  line-height: 1.65;
  white-space: pre-wrap;
  word-break: break-all;
}

.qz-od-payload .k {
  color: var(--qz-text-muted);
}

/* 使用说明：富文本由 v-html 注入，没有 scoped 属性，用 :deep() 命中 */
.qz-od-inst {
  padding: var(--qz-space-3) var(--qz-space-4);
  border: 1px solid color-mix(in oklab, var(--qz-info) 22%, transparent);
  border-radius: var(--qz-radius-control);
  background: color-mix(in oklab, var(--qz-info) 10%, transparent);
}

.qz-od-inst + .qz-od-inst {
  margin-top: var(--qz-space-2);
}

.qz-od-inst-head {
  display: flex;
  align-items: center;
  gap: var(--qz-space-2);
  margin-bottom: var(--qz-space-2);
  color: var(--qz-info);
  font-size: var(--qz-text-sm);
  font-weight: 700;
  line-height: var(--qz-leading-body);
}

.qz-od-inst-head svg {
  flex: none;
  width: 15px;
  height: 15px;
}

.qz-od-inst-body {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  line-height: 1.65;
  overflow-wrap: anywhere;
}

.qz-od-inst-body :deep(:is(h1, h2, h3, h4, h5, h6)) {
  margin: var(--qz-space-2) 0 var(--qz-space-1);
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 700;
}

.qz-od-inst-body :deep(p) {
  margin: var(--qz-space-1) 0;
}

.qz-od-inst-body :deep(strong) {
  color: var(--qz-text);
}

.qz-od-inst-body :deep(a) {
  color: var(--qz-info);
  text-decoration: underline;
}

.qz-od-inst-body :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: var(--qz-space-2);
}

.qz-od-inst-body :deep(ul),
.qz-od-inst-body :deep(ol) {
  margin: var(--qz-space-1) 0;
  padding-left: 1.4em;
}

.qz-od-inst-body :deep(ul) {
  list-style: disc;
}

.qz-od-inst-body :deep(ol) {
  list-style: decimal;
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
  border-radius: 12px 0 0 12px;
}

.qz-od-table th:last-child {
  border-radius: 0 12px 12px 0;
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

  .qz-od-head {
    flex-direction: column-reverse;
    align-items: flex-start;
    gap: var(--qz-space-3);
  }

  .qz-panel.qz-od-panel {
    padding: var(--qz-space-4);
  }

  /* 手机：金额列换到商品下方整行，不再和名称 / 规格 chip 争宽 */
  .qz-od-item {
    flex-wrap: wrap;
  }

  .qz-od-item-main {
    flex: 1 1 0;
  }

  .qz-od-item-amt {
    flex: 1 1 100%;
    padding-top: var(--qz-space-2);
    border-top: 1px dashed var(--qz-border-strong);
  }

  .qz-od-grid.is-amount {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
