<template>
  <div ref="rootEl" class="qz-container qz-pay">
    <!-- 返回 -->
    <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm qz-pay-back" @click="goBack">
      <ChevronLeft />
      {{ t('payment.backToOrders') }}
    </button>

    <!-- 骨架 -->
    <div v-if="loading" class="qz-pay-skeleton">
      <div class="qz-skeleton is-lg"></div>
      <div class="qz-skeleton is-line"></div>
      <div class="qz-skeleton qz-pay-sk-block"></div>
      <div class="qz-skeleton is-sm"></div>
      <div class="qz-skeleton qz-pay-sk-btn"></div>
    </div>

    <!-- 游客补录凭据 -->
    <section v-else-if="showGuestAuthForm" class="qz-pay-detail">
      <h2>{{ t('payment.guestAuthTitle') }}</h2>
      <p class="qz-alert is-warning" style="margin-top: 10px;">
        <b>{{ t('payment.guestAuthHint') }}</b>
      </p>
      <form class="qz-pay-form" style="margin-top: 16px;" novalidate @submit.prevent="handleGuestAuthSubmit">
        <div class="qz-field">
          <label class="qz-field-label" for="qz-pay-guest-email">{{ t('qiaozhi.detail.email') }}<b>*</b></label>
          <input
            id="qz-pay-guest-email"
            v-model="guestAuth.email"
            class="qz-field-control"
            type="email"
            autocomplete="email"
            :placeholder="t('qiaozhi.detail.emailPlaceholder')"
          />
        </div>
        <div class="qz-field">
          <label class="qz-field-label" for="qz-pay-guest-password">{{ t('qiaozhi.detail.orderPassword') }}<b>*</b></label>
          <input
            id="qz-pay-guest-password"
            v-model="guestAuth.order_password"
            class="qz-field-control"
            type="password"
            autocomplete="current-password"
            :placeholder="t('qiaozhi.detail.orderPasswordPlaceholder')"
          />
        </div>
        <p v-if="guestAuthError" class="qz-alert is-danger" role="alert">{{ guestAuthError }}</p>
        <button type="submit" class="qz-btn qz-btn--primary qz-btn--block">
          <LogIn />
          {{ t('payment.guestAuthSubmit') }}
        </button>
      </form>
    </section>

    <!-- 订单不存在 -->
    <div v-else-if="!order" class="qz-empty">
      <PackageOpen />
      <b>{{ t('payment.orderNotFound') }}</b>
      <RouterLink :to="backLink" class="qz-btn qz-btn--secondary">
        <ArrowLeft />
        {{ t('payment.backToOrders') }}
      </RouterLink>
    </div>

    <!-- 过期 / 取消 -->
    <section v-else-if="orderExpired || orderCanceled" class="qz-pay-detail">
      <div class="qz-pay-head">
        <div>
          <span class="qz-eyebrow">{{ t('qiaozhi.payment.pageEyebrow') }}</span>
          <h2>{{ orderCanceled ? t('payment.orderCanceled') : t('payment.orderExpired') }}</h2>
          <p class="qz-pay-sub">{{ order.order_no }}</p>
        </div>
        <RouterLink :to="backLink" class="qz-btn qz-btn--secondary">
          <ArrowLeft />
          {{ t('payment.backToOrders') }}
        </RouterLink>
      </div>
      <div class="qz-pay-meta">
        <div>
          <small>{{ t('payment.orderNo') }}</small>
          <b>{{ order.order_no }}</b>
        </div>
        <div>
          <small>{{ t('payment.orderStatus') }}</small>
          <b>{{ statusLabel(order.status) }}</b>
        </div>
        <div>
          <small>{{ t('orderDetail.amountTotal') }}</small>
          <b>{{ formatMoney(order.total_amount, order.currency) }}</b>
        </div>
      </div>
    </section>

    <template v-else>
      <!-- 订单摘要卡 -->
      <section class="qz-pay-card qz-pay-summary">
        <div class="qz-pay-head">
          <div>
            <span class="qz-eyebrow">{{ t('qiaozhi.payment.eyebrow') }}</span>
            <h1>{{ t('qiaozhi.payment.summaryTitle') }}</h1>
            <p class="qz-pay-sub">
              <span>{{ t('payment.orderNo') }}</span>
              <code>{{ order.order_no }}</code>
            </p>
          </div>
          <div class="qz-pay-chips">
            <span class="qz-badge is-accent">{{ statusLabel(order.status) }}</span>
            <span v-if="showCountdown" class="qz-badge" :class="countdownExpired ? 'is-danger' : 'is-warning'">
              <Clock />
              <span style="font-variant-numeric: tabular-nums;">{{ countdownText }}</span>
            </span>
          </div>
        </div>

        <div class="qz-pay-price-row">
          <div class="qz-pay-price">
            <span class="qz-pay-price-sym">{{ totalParts.symbol }}</span>
            <strong>{{ totalParts.major }}{{ totalParts.minor }}</strong>
          </div>
          <div v-if="customerFeeApplied && customerFeeAmountDisplay" class="qz-badge is-danger">
            {{ t('payment.feeAmountLabel') }} +{{ customerFeeAmountDisplay }}
          </div>
        </div>

        <div class="qz-pay-meta">
          <div>
            <small>{{ t('orderDetail.createdAtLabel') }}</small>
            <b>{{ formatDate(order.created_at) }}</b>
          </div>
          <div v-if="order.expires_at">
            <small>{{ t('payment.expiresAt') }}</small>
            <b>{{ formatDate(order.expires_at) }}</b>
          </div>
          <div v-if="paymentResult && resultChannelName">
            <small>{{ t('payment.methodLabel') }}</small>
            <b>{{ resultChannelName }}</b>
          </div>
          <div v-if="paymentResult && interactionLabel && interactionLabel !== '-'">
            <small>{{ t('payment.interactionLabel') }}</small>
            <b>{{ interactionLabel }}</b>
          </div>
        </div>

        <!-- 商品清单 -->
        <div v-if="orderItems.length" class="qz-pay-items">
          <h3>{{ t('payment.itemsTitle') }}</h3>
          <ul>
            <li v-for="(item, idx) in orderItems" :key="idx">
              <div class="qz-pay-item-main">
                <b>{{ getLocalizedText(item.title) }}</b>
                <small>
                  {{ t('orderDetail.quantityLabel') }}&nbsp;×&nbsp;{{ item.quantity }}
                  <span v-if="orderItemSkuText(item)">&nbsp;·&nbsp;{{ orderItemSkuText(item) }}</span>
                  <span v-if="item.fulfillment_type">&nbsp;·&nbsp;{{ fulfillmentTypeLabelText(item.fulfillment_type) }}</span>
                </small>
              </div>
              <span class="qz-pay-item-amount">{{ formatMoney(item.total_price, order.currency) }}</span>
            </li>
          </ul>
        </div>
      </section>

      <!-- 支付结果视图：QR / 跳转 -->
      <section v-if="showResultView" class="qz-pay-detail qz-pay-result">
        <div class="qz-pay-head">
          <div>
            <span class="qz-eyebrow">{{ t('qiaozhi.payment.scanEyebrow') }}</span>
            <h2>{{ paymentResultTitle }}</h2>
            <p class="qz-pay-sub">{{ paymentGuideTip }}</p>
          </div>
          <div class="qz-pay-head-actions">
            <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm" :disabled="loading" @click="handleRefresh">
              <RefreshCw />
              {{ t('payment.refreshStatus') }}
            </button>
            <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm" @click="handleChangePaymentMethod">
              <RotateCcw />
              {{ t('payment.changeMethod') }}
            </button>
          </div>
        </div>

        <div v-if="paymentAlert" class="qz-alert" :class="paymentAlert.level === 'error' ? 'is-danger' : 'is-warning'">
          {{ paymentAlert.message }}
        </div>

        <div v-if="showQRCode" class="qz-pay-qrbox">
          <div class="qz-pay-qr">
            <img v-if="qrImageUrl" :src="qrImageUrl" alt="QR Code" />
            <div v-else class="qz-pay-qr-placeholder">
              <Loader2 class="qz-pay-spin" />
            </div>
          </div>
          <div class="qz-pay-qr-side">
            <p class="qz-pay-qr-title">{{ paymentGuideTitle }}</p>
            <p class="qz-pay-qr-amount">
              <span>{{ t('payment.payableAmountLabel') }}</span>
              <strong>{{ payableAmountDisplay }}</strong>
            </p>
            <p v-if="qrUsingPayLinkFallback" class="qz-alert is-warning">
              <b>{{ t('payment.qrFallbackHint') }}</b>
            </p>
            <div v-if="pollingActive" class="qz-pay-polling">
              <Loader2 class="qz-pay-spin" />
              <span>{{ t('payment.pollingHint') }}</span>
            </div>
          </div>
        </div>

        <div v-if="hasCryptoPaymentDetails" class="qz-pay-crypto">
          <div v-for="item in cryptoPaymentDetails" :key="item.key" class="qz-pay-crypto-row">
            <small>{{ item.label }}</small>
            <b>
              {{ item.value }}
              <em v-if="item.detail">({{ item.detail }})</em>
            </b>
          </div>
          <div v-if="cryptoWalletAddress" class="qz-pay-crypto-action">
            <button type="button" class="qz-btn qz-btn--secondary" @click="handleCopyWalletAddress">
              <Copy />
              {{ t('payment.copyWalletAddress') }}
            </button>
            <span v-if="walletAddressCopied" class="qz-pay-hint is-ok">{{ t('payment.copied') }}</span>
          </div>
        </div>

        <div v-if="showPayLink" class="qz-pay-link">
          <button type="button" class="qz-btn qz-btn--primary" @click="handleOpenPayLink">
            <ExternalLink />
            {{ t('payment.openPayLink') }}
          </button>
          <button type="button" class="qz-btn qz-btn--secondary" @click="handleCopyPayLink">
            <Copy />
            {{ t('payment.copyPayLink') }}
          </button>
          <span v-if="copied" class="qz-pay-hint is-ok">{{ t('payment.copied') }}</span>
        </div>
        <p v-if="openedPayWindow && payLinkOpenedTip" class="qz-alert is-info">
          <b>{{ payLinkOpenedTip }}</b>
        </p>
        <p v-if="showTelegramPayHint" class="qz-alert is-warning">
          <b>{{ t('payment.telegramExternalHint') }}</b>
        </p>
      </section>

      <!-- 待发起支付：渠道选择 + 下单按钮 -->
      <section v-else class="qz-pay-detail qz-pay-choose">
        <h2>{{ t('payment.channelTitle') }}</h2>

        <div v-if="!configReady" class="qz-alert is-info" style="margin-top: 12px;">
          <b>{{ t('common.loading') }}</b>
        </div>

        <template v-else>
          <!-- 钱包余额开关 -->
          <div v-if="showBalanceOption" class="qz-pay-balance" style="margin-top: 14px;">
            <Wallet />
            <span>
              <b>{{ t('payment.walletBalanceLabel') }}</b>
              <small>{{ walletLoading ? t('common.loading') : walletBalanceDisplay }}</small>
            </span>
            <em v-if="walletOnlyPayment">{{ t('payment.walletOnlyHint') }}</em>
            <label style="margin-left: auto; display: inline-flex; align-items: center; gap: 6px; font-size: var(--qz-text-xs); color: var(--qz-text-muted);">
              <input v-model="useBalance" type="checkbox" :disabled="walletOnlyPayment" />
              {{ t('payment.useBalance') }}
            </label>
          </div>
          <div v-if="useBalance" class="qz-pay-walletinfo">
            <span>{{ t('payment.walletDeductLabel') }}&nbsp;{{ expectedWalletPaidDisplay }}</span>
            <span v-if="!walletOnlyPayment">·&nbsp;{{ t('payment.onlinePayLabel') }}&nbsp;{{ expectedOnlinePayDisplay }}</span>
            <span v-if="walletOnlyPayment && expectedOnlinePayCents > 0" class="is-warn">
              ·&nbsp;{{ t('payment.walletInsufficientHint') }}
            </span>
          </div>

          <!-- 缓存支付 -->
          <div v-if="cachedPayment" class="qz-alert is-warning" style="margin-top: 14px;">
            <div style="font-weight: 800; margin-bottom: 6px;">{{ t('payment.cachedTitle') }}</div>
            <div style="margin-bottom: 8px;">{{ t('payment.cachedHint', { channel: cachedChannelName }) }}</div>
            <div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center;">
              <button type="button" class="qz-btn qz-btn--secondary" @click="restoreCachedPayment">
                <RotateCcw />
                {{ t('payment.useCached') }}
              </button>
              <small style="font-size: var(--qz-text-2xs); opacity: 0.8;">{{ t('payment.cachedCreateHint') }}</small>
            </div>
          </div>

          <!-- 支付方式网格 -->
          <div v-if="!walletOnlyPayment" class="qz-pay-methods" style="margin-top: 14px;">
            <h3><Wallet />{{ t('qiaozhi.detail.payTitle') }}</h3>
            <div v-if="!channels.length" class="qz-alert is-warning">
              {{ showBalanceOption ? t('payment.channelEmptyUseBalance') : t('payment.channelEmpty') }}
            </div>
            <div v-else class="qz-pay-channel-grid">
              <button
                v-for="ch in channels"
                :key="ch.id"
                type="button"
                class="qz-pay-channel"
                :class="{ 'is-active': selectedChannelId === Number(ch.id) }"
                :disabled="isChannelDisabledForAmount(ch)"
                :title="channelAmountLimitHint(ch) || undefined"
                @click="selectedChannelId = Number(ch.id)"
              >
                <i></i>
                <img v-if="ch.icon" :src="ch.icon" alt="" />
                <span>
                  <b>{{ channelDisplayName(ch.name) }}</b>
                  <small v-if="channelAmountLimitHint(ch)">{{ channelAmountLimitHint(ch) }}</small>
                </span>
              </button>
            </div>
          </div>

          <div class="qz-pay-total">
            <span>{{ t('payment.payableAmountLabel') }}</span>
            <strong>{{ payableAmountDisplay }}</strong>
          </div>

          <!-- 提示 -->
          <div v-if="paymentAlert" class="qz-alert" :class="paymentAlert.level === 'error' ? 'is-danger' : 'is-warning'">
            {{ paymentAlert.message }}
          </div>
          <div
            v-else-if="!walletOnlyPayment && requiresOnlineChannel && !selectedChannel"
            class="qz-alert is-warning"
          >
            {{ t('payment.selectChannelError') }}
          </div>
          <div
            v-else-if="walletOnlyPayment && expectedOnlinePayCents > 0"
            class="qz-alert is-warning"
          >
            {{ t('payment.walletInsufficientHint') }}
          </div>
          <div
            v-else-if="!requiresOnlineChannel"
            class="qz-alert is-info"
          >
            <b>{{ t('payment.walletPayOnly') }}</b>
          </div>

          <div class="qz-pay-actions">
            <button type="button" class="qz-btn qz-btn--primary" :disabled="!canSubmitPayment" @click="handlePayment">
              <Loader2 v-if="submitting" class="qz-pay-spin" />
              <CreditCard v-else />
              {{ submitting ? t('payment.submitting') : t('payment.submitButton') }}
            </button>
            <button type="button" class="qz-btn qz-btn--secondary" :disabled="loading" @click="handleRefresh">
              <RefreshCw />
              {{ t('payment.refreshStatus') }}
            </button>
          </div>
        </template>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  ArrowLeft, ChevronLeft, Clock, Copy, CreditCard, ExternalLink,
  Loader2, LogIn, PackageOpen, RefreshCw, RotateCcw, Wallet,
} from 'lucide-vue-next'
import { usePayment } from '../../composables/usePayment'
import { channelDisplayName } from '../../utils/cryptoChain'
import { useQiaozhiPrice } from './utils/price'
import { setupReveal } from './utils/reveal'

const router = useRouter()
const { t } = useI18n()
const { split: splitPrice } = useQiaozhiPrice()

const {
  loading, submitting, order, paymentResult, selectedChannelId, copied, walletAddressCopied,
  openedPayWindow, cachedPayment, guestAuth, guestAuthError, walletLoading, useBalance,
  backLink, showGuestAuthForm, walletOnlyPayment, showBalanceOption, configReady, channels,
  selectedChannel, cachedChannelName, resultChannelName, interactionLabel,
  paymentResultTitle, paymentGuideTitle, paymentGuideTip, showPayLink, showTelegramPayHint, payLinkOpenedTip,
  cryptoWalletAddress, cryptoPaymentDetails, hasCryptoPaymentDetails, qrUsingPayLinkFallback, showQRCode, qrImageUrl,
  orderExpired, orderCanceled, paymentAlert, countdownExpired, countdownText, showCountdown, showResultView, pollingActive, orderItems,
  customerFeeApplied, customerFeeAmountDisplay, payableAmountDisplay, walletBalanceDisplay,
  expectedWalletPaidDisplay, expectedOnlinePayDisplay, expectedOnlinePayCents, requiresOnlineChannel,
  isChannelDisabledForAmount, channelAmountLimitHint, canSubmitPayment,
  formatDate, statusLabel, formatMoney, getLocalizedText, orderItemSkuText, fulfillmentTypeLabelText,
  handleCopyPayLink, handleCopyWalletAddress, handleOpenPayLink, restoreCachedPayment, handleChangePaymentMethod,
  handlePayment, handleGuestAuthSubmit, handleRefresh,
} = usePayment()

// 大价签三段式：优先用订单原币种金额（paymentResult.amount 在换汇渠道下是结算币种，这里只做视觉主金额）
const totalParts = computed(() => splitPrice(order.value?.total_amount ?? '0', order.value?.currency))

// 入场动画
const rootEl = ref<HTMLElement | null>(null)
let cleanupReveal: (() => void) | null = null
const armReveal = () => {
  cleanupReveal?.()
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
}
watch(loading, (v) => { if (!v) armReveal() })
onMounted(() => { if (!loading.value) armReveal() })
onUnmounted(() => { cleanupReveal?.() })

const goBack = () => {
  if (window.history.length > 1) router.back()
  else router.push(backLink.value)
}
</script>

<style scoped>
/* 支付页：控件走 base.css 基类——返回钮 / 刷新 / 换方式 / 复制 / 提交是 .qz-btn（--primary / --secondary /
   --sm / --block），状态与倒计时是 .qz-badge.is-*，提示条是 .qz-alert.is-*，游客补录表单是 .qz-field 四件套，
   骨架是 .qz-skeleton(.is-lg / .is-line / .is-sm)。本块只剩本页的排版：页面根 / 卡片与详情面 / 价签行 /
   元信息 / 商品清单 / 二维码 / 链上明细 / 渠道选择 / 操作区，class 统一 qz-pay-*。颜色 / 字号 / 圆角 / 阴影 /
   过渡一律读 --qz-* token。
   特指度：根元素同时带 .qz-container（layout.css，(0,2,0)），这里写成 .qz-container.qz-pay 提到 (0,3,0)；
   覆盖基类的返回钮同样串成 .qz-btn.qz-pay-back，不依赖注入顺序。 */

/* ---------- 页面根：比容器窄一档，纵向堆叠 ---------- */
.qz-container.qz-pay { width: min(1080px, 94vw); display: flex; flex-direction: column; gap: 16px; padding-top: 24px; }

/* ---------- 返回钮：.qz-btn--secondary --sm 的外观，这里只管位置（桌面端随滚动吸顶、半透明毛玻璃底） ---------- */
.qz-btn.qz-pay-back { align-self: flex-start; }
@media (min-width: 1024px) {
  .qz-btn.qz-pay-back {
    position: sticky;
    top: calc(var(--qz-header-h) + 12px);
    z-index: 38;
    background: color-mix(in srgb, var(--qz-surface) 94%, transparent);
    backdrop-filter: blur(14px);
  }
}

/* ---------- 骨架：行用 base.css 的 .qz-skeleton 变体，这里只给方块与按钮占位一个尺寸 ---------- */
.qz-pay-skeleton { display: grid; gap: 16px; }
.qz-pay-sk-block { aspect-ratio: 1 / 1; max-height: 520px; }
.qz-pay-sk-btn { min-height: var(--qz-control-md); margin-top: auto; }

/* ---------- 卡片面 / 详情面 ---------- */
.qz-pay-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  overflow: visible;
  background: var(--qz-surface);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  box-shadow: var(--qz-shadow-resting);
}
.qz-pay-detail {
  padding: 24px 28px 28px;
  background: var(--qz-surface);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  box-shadow: var(--qz-shadow-resting);
}
.qz-pay-detail > h2 { margin-bottom: 14px; font-size: var(--qz-text-lg); font-weight: 700; color: var(--qz-text); }

/* ---------- 状态徽标行 / 大价签 ---------- */
.qz-pay-chips { display: flex; flex-wrap: wrap; gap: 6px; margin-top: -6px; }
.qz-pay-price-row { display: flex; align-items: center; justify-content: space-between; gap: 14px; flex-wrap: wrap; }
.qz-pay-price { display: flex; align-items: baseline; flex-wrap: wrap; gap: 4px 8px; }
.qz-pay-price-sym { font-size: var(--qz-text-lg); font-weight: 800; color: var(--qz-text); }
.qz-pay-price strong {
  font-size: var(--qz-text-3xl);
  font-weight: 800;
  line-height: 1;
  color: var(--qz-text);
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.01em;
}

/* ---------- 游客补录凭据表单（字段用 .qz-field / .qz-field-label / .qz-field-control） ---------- */
.qz-pay-form { display: grid; gap: 10px; }
.qz-pay-hint { margin: -4px 0 0; font-size: var(--qz-text-xs); line-height: 1.55; color: var(--qz-text-muted); }
.qz-pay-hint.is-ok { margin: 0; color: var(--qz-accent-strong); }

/* ---------- 钱包余额行 / 支付方式 ---------- */
.qz-pay-methods {
  display: grid;
  gap: 8px;
  padding: 10px 12px;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  border: 1px solid var(--qz-border);
}
.qz-pay-methods h3 { display: flex; align-items: center; gap: 6px; margin: 0; font-size: var(--qz-text-xs); font-weight: 700; color: var(--qz-text-muted); }
.qz-pay-methods h3 svg { width: 15px; height: 15px; }
.qz-pay-channel-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(170px, 1fr)); gap: 8px; }
.qz-pay-channel,
.qz-pay-balance {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: var(--qz-control-md);
  padding: 7px 10px;
  text-align: left;
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface);
  color: var(--qz-text);
  cursor: pointer;
  transition: border-color var(--qz-duration-fast) var(--qz-ease-standard), background-color var(--qz-duration-fast) var(--qz-ease-standard);
}
/* 渠道按钮是支付方式卡里的内层行：圆角降一档 */
.qz-pay-channel { border-radius: var(--qz-radius-sm); }
.qz-pay-channel:hover:not(:disabled),
.qz-pay-balance:hover { border-color: var(--qz-border-strong); }
.qz-pay-channel:disabled { opacity: 0.5; cursor: not-allowed; }
.qz-pay-channel.is-active { border-color: var(--qz-accent); background: var(--qz-accent-subtle); }
.qz-pay-channel > i { flex: 0 0 auto; width: 16px; height: 16px; border-radius: 50%; border: 2px solid var(--qz-border-strong); background: var(--qz-surface); }
.qz-pay-channel.is-active > i { border: 5px solid var(--qz-accent); }
.qz-pay-channel img { flex: 0 0 auto; width: 24px; height: 24px; border-radius: var(--qz-radius-xs); object-fit: contain; }
.qz-pay-channel span,
.qz-pay-balance span { display: grid; min-width: 0; }
.qz-pay-channel b,
.qz-pay-balance b { overflow: hidden; font-size: var(--qz-text-sm); font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
.qz-pay-channel small,
.qz-pay-balance small { font-size: var(--qz-text-2xs); color: var(--qz-text-muted); }
.qz-pay-balance input { accent-color: var(--qz-accent); width: 16px; height: 16px; }
.qz-pay-balance svg { flex: 0 0 auto; width: 20px; height: 20px; color: var(--qz-accent-strong); }
.qz-pay-balance em { margin-left: auto; font-size: var(--qz-text-2xs); font-style: normal; color: var(--qz-warning); text-align: right; }
.qz-pay-spin { width: 16px; height: 16px; animation: qz-pay-spin 1s linear infinite; }
@keyframes qz-pay-spin { to { transform: rotate(360deg); } }

@media (max-width: 1023px) {
  .qz-container.qz-pay { padding-top: 16px; gap: 12px; padding-bottom: calc(96px + env(safe-area-inset-bottom)); }
  .qz-pay-card { grid-template-columns: minmax(0, 1fr); }
  .qz-pay-sk-block { max-height: 260px; }
}
@media (max-width: 640px) {
  .qz-pay-price strong { font-size: var(--qz-text-2xl); }
  .qz-pay-detail { padding: 18px 16px 20px; }
}

/* ---------- 本页排版 ---------- */
.qz-pay-summary { display: grid; gap: 16px; padding: 24px 24px 20px; grid-template-columns: minmax(0, 1fr); }
.qz-pay-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 14px; flex-wrap: wrap; }
.qz-pay-head > div:first-child { display: grid; gap: 6px; min-width: 0; }
.qz-pay-head h1 { font-size: var(--qz-text-xl); font-weight: 700; line-height: 1.3; color: var(--qz-text); }
.qz-pay-head h2 { font-size: var(--qz-text-lg); font-weight: 700; line-height: 1.3; color: var(--qz-text); }
.qz-pay-sub { display: inline-flex; align-items: center; gap: 8px; flex-wrap: wrap; font-size: var(--qz-text-sm); color: var(--qz-text-muted); }
.qz-pay-sub code { font-family: inherit; font-weight: 700; color: var(--qz-text); font-variant-numeric: tabular-nums; }
.qz-pay-head-actions { display: inline-flex; align-items: center; gap: 8px; flex-wrap: wrap; }

.qz-pay-meta { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 12px; padding-top: 4px; border-top: 1px solid var(--qz-border); }
.qz-pay-meta > div { display: grid; gap: 4px; min-width: 0; padding-top: 10px; }
.qz-pay-meta small { font-size: var(--qz-text-xs); color: var(--qz-text-muted); font-weight: 600; }
.qz-pay-meta b { font-size: var(--qz-text-sm); font-weight: 800; color: var(--qz-text); overflow-wrap: anywhere; }

.qz-pay-items { padding-top: 4px; border-top: 1px solid var(--qz-border); }
.qz-pay-items h3 { margin: 10px 0 8px; font-size: var(--qz-text-xs); font-weight: 700; color: var(--qz-text-muted); }
.qz-pay-items ul { display: grid; gap: 8px; }
.qz-pay-items li { display: flex; align-items: flex-start; justify-content: space-between; gap: 10px; padding: 10px 12px; border-radius: var(--qz-radius-control); background: var(--qz-surface-muted); }
.qz-pay-item-main { display: grid; gap: 4px; min-width: 0; }
.qz-pay-item-main b { font-size: var(--qz-text-sm); font-weight: 800; color: var(--qz-text); overflow-wrap: anywhere; }
.qz-pay-item-main small { font-size: var(--qz-text-xs); color: var(--qz-text-muted); }
.qz-pay-item-amount { flex: 0 0 auto; font-size: var(--qz-text-sm); font-weight: 800; color: var(--qz-text); white-space: nowrap; font-variant-numeric: tabular-nums; }

.qz-pay-result { display: grid; gap: 16px; }
.qz-pay-qrbox { display: grid; grid-template-columns: minmax(0, 320px) minmax(0, 1fr); gap: 20px; align-items: center; padding: 16px; border-radius: var(--qz-radius-control); background: var(--qz-surface-muted); border: 1px solid var(--qz-border); }
.qz-pay-qr { position: relative; width: 100%; max-width: 300px; aspect-ratio: 1 / 1; margin: 0 auto; padding: 14px; border-radius: var(--qz-radius-control); background: var(--qz-qr-frame); display: grid; place-items: center; box-shadow: var(--qz-shadow-resting); }
.qz-pay-qr img { width: 100%; height: 100%; object-fit: contain; }
.qz-pay-qr-placeholder { display: grid; place-items: center; color: var(--qz-text-muted); }
.qz-pay-qr-placeholder .qz-pay-spin { width: 32px; height: 32px; }
.qz-pay-qr-side { display: grid; gap: 10px; min-width: 0; }
.qz-pay-qr-title { font-size: var(--qz-text-base); font-weight: 800; color: var(--qz-text); }
.qz-pay-qr-amount { display: flex; align-items: baseline; justify-content: space-between; gap: 10px; padding: 10px 12px; border-radius: var(--qz-radius-control); background: var(--qz-surface); border: 1px solid var(--qz-border); }
.qz-pay-qr-amount span { font-size: var(--qz-text-xs); color: var(--qz-text-muted); font-weight: 600; }
.qz-pay-qr-amount strong { font-size: var(--qz-text-xl); font-weight: 800; color: var(--qz-text); font-variant-numeric: tabular-nums; }
.qz-pay-polling { display: inline-flex; align-items: center; gap: 8px; padding: 10px 12px; border-radius: var(--qz-radius-pill); background: var(--qz-accent-subtle); color: var(--qz-accent-strong); font-size: var(--qz-text-sm); font-weight: 700; }
.qz-pay-polling .qz-pay-spin { width: 16px; height: 16px; }

.qz-pay-crypto { display: grid; gap: 8px; padding: 14px; border-radius: var(--qz-radius-control); background: var(--qz-surface-muted); border: 1px solid var(--qz-border); }
.qz-pay-crypto-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; padding-bottom: 8px; border-bottom: 1px solid var(--qz-border); }
.qz-pay-crypto-row:last-child { border-bottom: 0; padding-bottom: 0; }
.qz-pay-crypto-row small { flex: 0 0 auto; font-size: var(--qz-text-xs); color: var(--qz-text-muted); font-weight: 600; }
.qz-pay-crypto-row b { min-width: 0; text-align: right; font-size: var(--qz-text-sm); font-weight: 800; color: var(--qz-text); overflow-wrap: anywhere; word-break: break-all; }
.qz-pay-crypto-row em { font-style: normal; font-weight: 500; color: var(--qz-text-muted); margin-left: 4px; }
.qz-pay-crypto-action { display: inline-flex; align-items: center; gap: 10px; flex-wrap: wrap; }

.qz-pay-link { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.qz-pay-link .qz-btn--primary { min-width: 200px; }

.qz-pay-choose { display: grid; gap: 0; }
.qz-pay-choose h2 { font-size: var(--qz-text-lg); font-weight: 700; color: var(--qz-text); margin-bottom: 4px; }
.qz-pay-walletinfo { display: flex; flex-wrap: wrap; gap: 4px 8px; margin-top: 8px; font-size: var(--qz-text-xs); color: var(--qz-text-muted); }
.qz-pay-walletinfo .is-warn { color: var(--qz-warning); font-weight: 700; }

.qz-pay-total { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-top: 16px; padding: 14px 16px; border-radius: var(--qz-radius-control); background: var(--qz-surface-muted); border: 1px solid var(--qz-border); }
.qz-pay-total span { font-size: var(--qz-text-sm); color: var(--qz-text-muted); font-weight: 600; }
.qz-pay-total strong { font-size: var(--qz-text-xl); font-weight: 800; color: var(--qz-text); font-variant-numeric: tabular-nums; }

.qz-pay-actions { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); gap: 10px; margin-top: 14px; }

@media (max-width: 720px) {
  .qz-pay-qrbox { grid-template-columns: minmax(0, 1fr); }
  .qz-pay-actions { grid-template-columns: minmax(0, 1fr); }
  .qz-pay-head { align-items: stretch; }
  .qz-pay-head-actions { width: 100%; }
}
</style>
