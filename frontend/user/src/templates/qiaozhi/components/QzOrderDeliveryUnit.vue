<template>
  <section class="qz-panel qz-od-panel qz-reveal is-pending">
    <div class="qz-od-panel-head">
      <h2 class="qz-eyebrow">{{ t('qiaozhi.orderDetail.sectionDelivery') }}</h2>
      <div v-if="singleFulfillment" class="qz-od-actions">
        <QzButton
          v-if="canCopy(singleFulfillment.fulfillment)"
          size="sm"
          @click="emit('copy-fulfillment', singleFulfillment.fulfillment)"
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
          @click="emit('download-fulfillment', singleFulfillment.orderNo)"
        >
          <Download />
          <span>{{ fulfillmentDownloading ? t('orderDetail.fulfillmentDownloading') : t('orderDetail.fulfillmentDownload') }}</span>
        </QzButton>
      </div>
    </div>

    <p v-if="deliveryHint" class="qz-od-hint">{{ deliveryHint }}</p>

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
          @click="emit('copy-fulfillment', unit.fulfillment)"
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
          @click="emit('download-fulfillment', unit.orderNo)"
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
        <div v-if="isFulfillmentTruncated(unit.fulfillment)" class="qz-alert is-warning qz-od-trunc">
          <AlertTriangle />
          <span>{{ t('orderDetail.fulfillmentTruncatedHint') }}</span>
        </div>
        <!-- 待交付（payload 为空、也没有键值行）时不画空盒子 -->
        <div v-if="unit.fulfillment.payload || payloadRows(unit.fulfillment).length" class="qz-od-payload">
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
</template>
<script setup lang="ts">
import { computed, toRef } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  AlertTriangle, Check, Copy, Download,
  Image as ImageIcon, Info,
} from 'lucide-vue-next'
import { useOrderDisplayHelpers } from '../../../composables/useOrderDisplayHelpers'
import { buildDeliveryUnits, type DeliveryUnit } from '../utils/deliveryUnits'
import QzButton from './QzButton.vue'

interface Props {
  order: any
  fulfillmentCopied: boolean
  fulfillmentDownloading: boolean
  deliveryHint?: string
}
interface Emits {
  (e: 'copy-fulfillment', fulfillment: any): void
  (e: 'download-fulfillment', orderNo: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()

const orderRef = toRef(props, 'order')
const {
  statusLabel, statusVariant,
  fulfillmentTypeLabelText, fulfillmentStatusLabelText,
  isFulfillmentTruncated,
  formatDate, getLocalizedText,
  orderItemImage, orderItemSkuText, manualSubmissionRows,
  formatMoney, hasDiscountAmount, hasItemDiscount, formatItemPaidAmount,
  instructionBlocks, resolvedChildStatus,
} = useOrderDisplayHelpers(orderRef)

const deliveryUnits = computed<DeliveryUnit[]>(() => buildDeliveryUnits(props.order, resolvedChildStatus))

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
</script>

<style scoped>
/* 本组件的根 <section> 挂 qz-panel.qz-od-panel；通用样式在 base.css 的 .qz-panel，这里只补响应式收缩 padding。 */
.qz-panel.qz-od-panel {
  min-width: 0;
}

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

/* 交付成功后的「卡密已发送至邮箱」提示 */
.qz-od-hint {
  margin: calc(-1 * var(--qz-space-2)) 0 var(--qz-space-3);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  overflow-wrap: anywhere;
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
  font-family: var(--qz-font-mono);
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
  border-radius: var(--qz-radius-sm);
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
  border-radius: var(--qz-radius-sm);
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

/* 截断提示是 base 的 .qz-alert.is-warning；这里只把图标和文案排成一行（警示色留给图标） */
.qz-alert.qz-od-trunc {
  display: flex;
  align-items: flex-start;
  gap: var(--qz-space-2);
}

.qz-od-trunc svg {
  flex: none;
  width: 16px;
  height: 16px;
  margin-top: 2px;
  color: var(--qz-warning);
}

.qz-od-payload {
  max-height: 260px;
  overflow-y: auto;
  padding: var(--qz-space-3) var(--qz-space-4);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  color: var(--qz-text);
  font-family: var(--qz-font-mono);
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

@media (max-width: 760px) {
  /* 手机断点：面板收缩 padding，并把金额列换到商品下方整行，不再和名称 / 规格 chip 争宽 */
  .qz-panel.qz-od-panel {
    padding: var(--qz-space-4);
  }

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
}
</style>
