<template>
  <div ref="rootEl" class="qz-container qz-ro">
    <!-- 页头：eyebrow + h1 + 说明 + 返回 -->
    <header class="qz-page-head">
      <div class="qz-page-head-copy">
        <span class="qz-eyebrow">{{ t('personalCenter.wallet.title') }}</span>
        <h1>{{ t('rechargeOrder.title') }}</h1>
        <p>{{ t('rechargeOrder.subtitle') }}</p>
      </div>
      <RouterLink class="qz-btn qz-btn--secondary qz-btn--sm" to="/me/orders">
        <ChevronLeft />
        {{ t('rechargeOrder.backList') }}
      </RouterLink>
    </header>

    <!-- 骨架 -->
    <div v-if="loading" class="qz-ro-skeleton" aria-hidden="true">
      <div class="qz-skeleton qz-ro-sk-card"></div>
      <div class="qz-skeleton qz-ro-sk-card"></div>
    </div>

    <!-- 充值单不存在 -->
    <div v-else-if="!recharge" class="qz-empty qz-reveal is-pending">
      <AlertCircle />
      <b>{{ t('rechargeOrder.notFound') }}</b>
      <QzButton variant="primary" @click="loadDetail()">
        {{ t('errorBoundary.retry') }}
      </QzButton>
    </div>

    <template v-else>
      <!-- 头部信息：单号 / 创建时间 / 充值金额 / 状态 -->
      <section class="qz-panel qz-ro-panel qz-ro-summary qz-reveal is-pending">
        <div class="qz-ro-summary-main">
          <span class="qz-eyebrow">{{ t('personalCenter.wallet.rechargeNoLabel') }}</span>
          <b class="qz-ro-no">{{ recharge.recharge_no }}</b>
          <small>{{ t('rechargeOrder.createdAtLabel') }}：{{ formatDate(recharge.created_at) }}</small>
        </div>
        <div class="qz-ro-summary-amt">
          <span class="qz-eyebrow">{{ t('rechargeOrder.rechargeAmount') }}</span>
          <b>{{ formatMoney(recharge.amount, recharge.currency) }}</b>
        </div>
        <span class="qz-badge" :class="toneClass(recharge.status)">{{ rechargeStatusText(recharge.status) }}</span>
      </section>

      <!-- 金额明细 -->
      <section class="qz-panel qz-ro-panel qz-reveal is-pending">
        <h2>{{ t('rechargeOrder.amountTitle') }}</h2>
        <div class="qz-ro-grid">
          <div class="qz-ro-cell">
            <span class="lbl">{{ t('rechargeOrder.rechargeAmount') }}</span>
            <b class="val">{{ formatMoney(recharge.amount, recharge.currency) }}</b>
          </div>
          <div v-if="customerFeeApplied" class="qz-ro-cell is-fee">
            <span class="lbl">{{ t('payment.feeAmountLabel') }}</span>
            <b class="val">{{ formatMoney(recharge.fee_amount, recharge.currency) }}</b>
          </div>
          <div class="qz-ro-cell is-total">
            <span class="lbl">{{ t('personalCenter.wallet.payAmountLabel') }}</span>
            <b class="val">{{ formatMoney(recharge.payable_amount, recharge.currency) }}</b>
          </div>
        </div>
      </section>

      <!-- 时间信息 -->
      <section class="qz-panel qz-ro-panel qz-reveal is-pending">
        <h2>{{ t('rechargeOrder.timeTitle') }}</h2>
        <div class="qz-ro-grid">
          <div class="qz-ro-cell">
            <span class="lbl">{{ t('rechargeOrder.createdAtLabel') }}</span>
            <b class="val">{{ formatDate(recharge.created_at) }}</b>
          </div>
          <div v-if="recharge.paid_at" class="qz-ro-cell">
            <span class="lbl">{{ t('rechargeOrder.paidAtLabel') }}</span>
            <b class="val">{{ formatDate(recharge.paid_at) }}</b>
          </div>
          <div v-if="payment?.expires_at" class="qz-ro-cell">
            <span class="lbl">{{ t('payment.expiresAt') }}</span>
            <b class="val">{{ formatDate(payment.expires_at) }}</b>
          </div>
        </div>
      </section>

      <!-- 备注 -->
      <section v-if="recharge.remark" class="qz-panel qz-ro-panel qz-reveal is-pending">
        <h2>{{ t('rechargeOrder.remarkLabel') }}</h2>
        <p class="qz-ro-remark">{{ recharge.remark }}</p>
      </section>

      <!-- 支付区域（仅待支付状态） -->
      <section v-if="isPending" class="qz-panel qz-ro-panel qz-reveal is-pending">
        <h2>{{ t('rechargeOrder.paymentTitle') }}</h2>
        <p class="qz-ro-hint">{{ t('personalCenter.wallet.pendingHint') }}</p>
        <div class="qz-ro-pay">
          <div v-if="showQRCode" class="qz-ro-qr">
            <span class="qz-ro-qr-title">{{ t('payment.qrTitle') }}</span>
            <img :src="qrImageUrl" alt="Recharge QR" />
            <p v-if="qrUsingPayLinkFallback" class="qz-ro-hint">{{ t('payment.qrFallbackHint') }}</p>
          </div>
          <div class="qz-ro-pay-side">
            <div v-if="hasCryptoPaymentDetails" class="qz-ro-crypto">
              <dl>
                <div v-for="item in cryptoPaymentDetails" :key="item.key">
                  <dt>{{ item.label }}</dt>
                  <dd>
                    {{ item.value }}
                    <span v-if="item.detail">({{ item.detail }})</span>
                  </dd>
                </div>
              </dl>
              <div v-if="cryptoWalletAddress" class="qz-ro-crypto-copy">
                <QzButton variant="secondary" size="sm" @click="handleCopyWalletAddress">
                  {{ t('payment.copyWalletAddress') }}
                </QzButton>
                <span v-if="walletAddressCopied" class="qz-ro-copied" role="status">{{ t('payment.copied') }}</span>
              </div>
            </div>
            <div class="qz-ro-pay-actions">
              <QzButton v-if="payLink" variant="primary" size="sm" @click="handleOpenPayLink">
                <ExternalLink />
                {{ t('payment.openPayLink') }}
              </QzButton>
              <QzButton variant="secondary" size="sm" :loading="checkingPayment" @click="checkPayment">
                {{ checkingPayment ? t('personalCenter.wallet.checkingPayStatus') : t('personalCenter.wallet.checkPayStatus') }}
              </QzButton>
            </div>
            <p v-if="showTelegramPayHint" class="qz-ro-hint">{{ t('payment.telegramExternalHint') }}</p>
          </div>
        </div>
      </section>

      <!-- 支付成功提示 -->
      <section v-if="recharge.status === 'success'" class="qz-panel qz-ro-panel qz-ro-success qz-reveal is-pending">
        <CheckCircle2 />
        <p>{{ t('personalCenter.wallet.rechargeSuccess') }}</p>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AlertCircle, CheckCircle2, ChevronLeft, ExternalLink } from 'lucide-vue-next'
import { useRechargeOrderDetail } from '../../composables/useRechargeOrderDetail'
import { setupReveal } from './utils/reveal'
import QzButton from './components/QzButton.vue'

const { t } = useI18n()

const {
  loading, checkingPayment, recharge, payment, walletAddressCopied, qrImageUrl,
  isPending, payLink, showTelegramPayHint, qrUsingPayLinkFallback, showQRCode,
  cryptoWalletAddress, cryptoPaymentDetails, hasCryptoPaymentDetails, customerFeeApplied,
  rechargeStatusText, rechargeStatusVariant, formatMoney, formatDate,
  loadDetail, checkPayment, handleOpenPayLink, handleCopyWalletAddress,
} = useRechargeOrderDetail()

/** BadgeTone → .qz-badge 状态色（success / danger，其余待支付态走 warning）。 */
const toneClass = (status?: string) => {
  switch (rechargeStatusVariant(status)) {
    case 'success': return 'is-success'
    case 'danger': return 'is-danger'
    default: return 'is-warning'
  }
}

// 入场动画：详情为异步加载；轮询把状态从待支付翻到已到账时会新渲染成功卡（带 is-pending），需重新 arm
const rootEl = ref<HTMLElement | null>(null)
let cleanupReveal: (() => void) | null = null
const armReveal = () => {
  cleanupReveal?.()
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
}
watch([loading, () => recharge.value?.status], ([isLoading]) => { if (!isLoading) armReveal() })
onMounted(() => { if (!loading.value) armReveal() })
onUnmounted(() => { cleanupReveal?.() })
</script>

<style scoped>
.qz-ro {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  max-width: 760px;
}

/* 页头用 base.css 的 .qz-page-head */

/* ---------- 骨架：卡块是 base 的 .qz-skeleton，这里只给尺寸 ---------- */
.qz-ro-skeleton {
  display: grid;
  gap: var(--qz-space-4);
}

/* 卡块占位跟面板同圆角：覆盖基类的 control 圆角，写成 (0,3,0) */
.qz-skeleton.qz-ro-sk-card {
  height: 150px;
  border-radius: var(--qz-radius-panel);
}

/* ---------- 面板 ---------- */
.qz-ro-panel h2 {
  margin: 0 0 var(--qz-space-4);
  color: var(--qz-text);
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: var(--qz-leading-body);
}

/* 头部信息卡：单号 / 金额 / 状态徽标横排 */
.qz-panel.qz-ro-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--qz-space-4) var(--qz-space-6);
}

.qz-ro-summary-main,
.qz-ro-summary-amt {
  display: grid;
  gap: var(--qz-space-1);
  min-width: 0;
}

.qz-ro-no {
  color: var(--qz-text);
  font-family: var(--qz-font-mono);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  overflow-wrap: anywhere;
}

.qz-ro-summary-main small {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  font-variant-numeric: tabular-nums;
}

.qz-ro-summary-amt b {
  color: var(--qz-text);
  font-size: var(--qz-text-xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
  font-variant-numeric: tabular-nums;
}

/* 金额 / 时间网格 */
.qz-ro-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: var(--qz-space-2);
}

.qz-ro-cell {
  display: grid;
  gap: var(--qz-space-1);
  padding: var(--qz-space-3) var(--qz-space-4);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
}

.qz-ro-cell .lbl {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
  font-weight: 700;
  letter-spacing: 0.04em;
}

.qz-ro-cell .val {
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.qz-ro-cell.is-fee .val {
  color: var(--qz-warning);
}

.qz-ro-cell.is-total .val {
  font-size: var(--qz-text-base);
}

.qz-ro-remark {
  margin: 0;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
  white-space: pre-line;
  overflow-wrap: anywhere;
}

.qz-ro-hint {
  margin: 0;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

/* ---------- 支付区：左二维码 / 右链上信息与操作；无二维码时右栏铺满 ---------- */
.qz-ro-pay {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: var(--qz-space-4);
  margin-top: var(--qz-space-3);
}

.qz-ro-qr {
  display: grid;
  justify-items: center;
  align-content: start;
  gap: var(--qz-space-3);
  padding: var(--qz-space-4);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  text-align: center;
}

.qz-ro-qr-title {
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
}

/* 二维码外框固定白底（--qz-qr-frame 两套主题同值），暗色下码的静区不会跟着变暗 */
.qz-ro-qr img {
  width: 208px;
  height: 208px;
  object-fit: contain;
  border-radius: var(--qz-radius-sm);
  background: var(--qz-qr-frame);
}

.qz-ro-pay-side {
  display: grid;
  align-content: start;
  gap: var(--qz-space-3);
  padding: var(--qz-space-4);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
}

.qz-ro-crypto {
  display: grid;
  gap: var(--qz-space-2);
  padding: var(--qz-space-3);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-sm);
  background: var(--qz-surface-muted);
}

.qz-ro-crypto dl {
  display: grid;
  gap: var(--qz-space-2);
  margin: 0;
}

.qz-ro-crypto dl > div {
  display: grid;
  gap: 2px;
  padding-bottom: var(--qz-space-2);
  border-bottom: 1px solid var(--qz-border);
}

.qz-ro-crypto dl > div:last-child {
  padding-bottom: 0;
  border-bottom: 0;
}

.qz-ro-crypto dt {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-ro-crypto dd {
  min-width: 0;
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  line-height: var(--qz-leading-body);
  word-break: break-all;
}

.qz-ro-crypto dd span {
  color: var(--qz-text-muted);
  font-weight: 400;
}

.qz-ro-crypto-copy,
.qz-ro-pay-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2);
}

.qz-ro-copied {
  color: var(--qz-success);
  font-size: var(--qz-text-xs);
  font-weight: 600;
}

/* ---------- 支付成功 ---------- */
.qz-panel.qz-ro-success {
  display: flex;
  align-items: center;
  gap: var(--qz-space-3);
  border-color: color-mix(in oklab, var(--qz-success) 40%, var(--qz-border));
}

.qz-ro-success svg {
  flex: none;
  width: 24px;
  height: 24px;
  color: var(--qz-success);
}

.qz-ro-success p {
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  line-height: var(--qz-leading-body);
}

@media (max-width: 760px) {
  .qz-ro {
    gap: var(--qz-space-3);
  }

  .qz-panel.qz-ro-panel {
    padding: var(--qz-space-4);
  }
}
</style>
