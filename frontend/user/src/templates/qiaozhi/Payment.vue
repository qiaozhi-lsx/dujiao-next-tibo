<template>
  <div ref="rootEl" class="j8-container j8-pd">
    <!-- 返回 -->
    <button type="button" class="j8-pd-back" @click="goBack">
      <ChevronLeft />
      {{ t('payment.backToOrders') }}
    </button>

    <!-- 骨架 -->
    <div v-if="loading" class="j8-pd-skeleton">
      <div class="j8-pd-sk-line is-lg"></div>
      <div class="j8-pd-sk-line"></div>
      <div class="j8-pd-sk-block"></div>
      <div class="j8-pd-sk-line is-sm"></div>
      <div class="j8-pd-sk-line is-btn"></div>
    </div>

    <!-- 游客补录凭据 -->
    <section v-else-if="showGuestAuthForm" class="j8-pd-detail">
      <h2>{{ t('payment.guestAuthTitle') }}</h2>
      <p class="j8-pd-note is-warm" style="margin-top: 10px;">
        <b>{{ t('payment.guestAuthHint') }}</b>
      </p>
      <form class="j8-ic" style="margin-top: 16px;" novalidate @submit.prevent="handleGuestAuthSubmit">
        <div class="j8-ic-field">
          <label for="j8-pay-guest-email">{{ t('qiaozhi.detail.email') }}<b>*</b></label>
          <input
            id="j8-pay-guest-email"
            v-model="guestAuth.email"
            type="email"
            autocomplete="email"
            :placeholder="t('qiaozhi.detail.emailPlaceholder')"
          />
        </div>
        <div class="j8-ic-field">
          <label for="j8-pay-guest-password">{{ t('qiaozhi.detail.orderPassword') }}<b>*</b></label>
          <input
            id="j8-pay-guest-password"
            v-model="guestAuth.order_password"
            type="password"
            autocomplete="current-password"
            :placeholder="t('qiaozhi.detail.orderPasswordPlaceholder')"
          />
        </div>
        <p v-if="guestAuthError" class="j8-ic-alert is-error" role="alert">{{ guestAuthError }}</p>
        <button type="submit" class="j8-pd-buy">
          <LogIn />
          {{ t('payment.guestAuthSubmit') }}
        </button>
      </form>
    </section>

    <!-- 订单不存在 -->
    <div v-else-if="!order" class="j8-empty">
      <PackageOpen />
      <b>{{ t('payment.orderNotFound') }}</b>
      <RouterLink :to="backLink" class="j8-btn-primary">
        <ArrowLeft />
        {{ t('payment.backToOrders') }}
      </RouterLink>
    </div>

    <!-- 过期 / 取消 -->
    <section v-else-if="orderExpired || orderCanceled" class="j8-pd-detail">
      <div class="j8-pd-pay-head">
        <div>
          <span class="j8-eyebrow">PAYMENT</span>
          <h2>{{ orderCanceled ? t('payment.orderCanceled') : t('payment.orderExpired') }}</h2>
          <p class="j8-pd-pay-sub">{{ order.order_no }}</p>
        </div>
        <RouterLink :to="backLink" class="j8-btn-primary">
          <ArrowLeft />
          {{ t('payment.backToOrders') }}
        </RouterLink>
      </div>
      <div class="j8-pd-pay-meta">
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
      <section class="j8-pd-card j8-pd-pay-card">
        <div class="j8-pd-pay-head">
          <div>
            <span class="j8-eyebrow">{{ t('qiaozhi.payment.eyebrow') }}</span>
            <h1>{{ t('qiaozhi.payment.summaryTitle') }}</h1>
            <p class="j8-pd-pay-sub">
              <span>{{ t('payment.orderNo') }}</span>
              <code>{{ order.order_no }}</code>
            </p>
          </div>
          <div class="j8-pd-chips">
            <span class="is-accent">{{ statusLabel(order.status) }}</span>
            <span v-if="showCountdown" :class="countdownExpired ? 'is-danger' : 'is-warm'">
              <Clock />
              <span style="font-variant-numeric: tabular-nums;">{{ countdownText }}</span>
            </span>
          </div>
        </div>

        <div class="j8-pd-price-row">
          <div class="j8-pd-price">
            <span class="j8-pd-price-sym">{{ totalParts.symbol }}</span>
            <strong>{{ totalParts.major }}{{ totalParts.minor }}</strong>
          </div>
          <div v-if="customerFeeApplied && customerFeeAmountDisplay" class="j8-pd-price-tag is-promo">
            {{ t('payment.feeAmountLabel') }} +{{ customerFeeAmountDisplay }}
          </div>
        </div>

        <div class="j8-pd-pay-meta">
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
        <div v-if="orderItems.length" class="j8-pd-pay-items">
          <h3>{{ t('payment.itemsTitle') }}</h3>
          <ul>
            <li v-for="(item, idx) in orderItems" :key="idx">
              <div class="j8-pd-pay-item-main">
                <b>{{ getLocalizedText(item.title) }}</b>
                <small>
                  {{ t('orderDetail.quantityLabel') }}&nbsp;×&nbsp;{{ item.quantity }}
                  <span v-if="orderItemSkuText(item)">&nbsp;·&nbsp;{{ orderItemSkuText(item) }}</span>
                  <span v-if="item.fulfillment_type">&nbsp;·&nbsp;{{ fulfillmentTypeLabelText(item.fulfillment_type) }}</span>
                </small>
              </div>
              <span class="j8-pd-pay-item-amount">{{ formatMoney(item.total_price, order.currency) }}</span>
            </li>
          </ul>
        </div>
      </section>

      <!-- 支付结果视图：QR / 跳转 -->
      <section v-if="showResultView" class="j8-pd-detail j8-pd-pay-result">
        <div class="j8-pd-pay-head">
          <div>
            <span class="j8-eyebrow">{{ t('qiaozhi.payment.scanEyebrow') }}</span>
            <h2>{{ paymentResultTitle }}</h2>
            <p class="j8-pd-pay-sub">{{ paymentGuideTip }}</p>
          </div>
          <div class="j8-pd-pay-head-actions">
            <button type="button" class="j8-pd-cart" :disabled="loading" @click="handleRefresh">
              <RefreshCw />
              {{ t('payment.refreshStatus') }}
            </button>
            <button type="button" class="j8-pd-cart" @click="handleChangePaymentMethod">
              <RotateCcw />
              {{ t('payment.changeMethod') }}
            </button>
          </div>
        </div>

        <div v-if="paymentAlert" class="j8-pd-alert" :class="paymentAlert.level === 'error' ? 'is-danger' : 'is-warm'">
          {{ paymentAlert.message }}
        </div>

        <div v-if="showQRCode" class="j8-pd-pay-qrbox">
          <div class="j8-pd-pay-qr">
            <img v-if="qrImageUrl" :src="qrImageUrl" alt="QR Code" />
            <div v-else class="j8-pd-pay-qr-placeholder">
              <Loader2 class="j8-ic-spin" />
            </div>
          </div>
          <div class="j8-pd-pay-qr-side">
            <p class="j8-pd-pay-qr-title">{{ paymentGuideTitle }}</p>
            <p class="j8-pd-pay-qr-amount">
              <span>{{ t('payment.payableAmountLabel') }}</span>
              <strong>{{ payableAmountDisplay }}</strong>
            </p>
            <p v-if="qrUsingPayLinkFallback" class="j8-pd-note is-warm">
              <b>{{ t('payment.qrFallbackHint') }}</b>
            </p>
            <div v-if="pollingActive" class="j8-pd-pay-polling">
              <Loader2 class="j8-ic-spin" />
              <span>{{ t('payment.pollingHint') }}</span>
            </div>
          </div>
        </div>

        <div v-if="hasCryptoPaymentDetails" class="j8-pd-pay-crypto">
          <div v-for="item in cryptoPaymentDetails" :key="item.key" class="j8-pd-pay-crypto-row">
            <small>{{ item.label }}</small>
            <b>
              {{ item.value }}
              <em v-if="item.detail">({{ item.detail }})</em>
            </b>
          </div>
          <div v-if="cryptoWalletAddress" class="j8-pd-pay-crypto-action">
            <button type="button" class="j8-pd-cart" @click="handleCopyWalletAddress">
              <Copy />
              {{ t('payment.copyWalletAddress') }}
            </button>
            <span v-if="walletAddressCopied" class="j8-ic-hint is-ok">{{ t('payment.copied') }}</span>
          </div>
        </div>

        <div v-if="showPayLink" class="j8-pd-pay-link">
          <button type="button" class="j8-pd-buy" @click="handleOpenPayLink">
            <ExternalLink />
            {{ t('payment.openPayLink') }}
          </button>
          <button type="button" class="j8-pd-cart" @click="handleCopyPayLink">
            <Copy />
            {{ t('payment.copyPayLink') }}
          </button>
          <span v-if="copied" class="j8-ic-hint is-ok">{{ t('payment.copied') }}</span>
        </div>
        <p v-if="openedPayWindow && payLinkOpenedTip" class="j8-pd-note is-accent">
          <b>{{ payLinkOpenedTip }}</b>
        </p>
        <p v-if="showTelegramPayHint" class="j8-pd-note is-warm">
          <b>{{ t('payment.telegramExternalHint') }}</b>
        </p>
      </section>

      <!-- 待发起支付：渠道选择 + 下单按钮 -->
      <section v-else class="j8-pd-detail j8-pd-pay-channel">
        <h2>{{ t('payment.channelTitle') }}</h2>

        <div v-if="!configReady" class="j8-pd-note is-accent" style="margin-top: 12px;">
          <b>{{ t('common.loading') }}</b>
        </div>

        <template v-else>
          <!-- 钱包余额开关 -->
          <div v-if="showBalanceOption" class="j8-ic-balance" style="margin-top: 14px;">
            <Wallet />
            <span>
              <b>{{ t('payment.walletBalanceLabel') }}</b>
              <small>{{ walletLoading ? t('common.loading') : walletBalanceDisplay }}</small>
            </span>
            <em v-if="walletOnlyPayment">{{ t('payment.walletOnlyHint') }}</em>
            <label style="margin-left: auto; display: inline-flex; align-items: center; gap: 6px; font-size: 12px; color: var(--muted);">
              <input v-model="useBalance" type="checkbox" :disabled="walletOnlyPayment" />
              {{ t('payment.useBalance') }}
            </label>
          </div>
          <div v-if="useBalance" class="j8-pd-pay-walletinfo">
            <span>{{ t('payment.walletDeductLabel') }}&nbsp;{{ expectedWalletPaidDisplay }}</span>
            <span v-if="!walletOnlyPayment">·&nbsp;{{ t('payment.onlinePayLabel') }}&nbsp;{{ expectedOnlinePayDisplay }}</span>
            <span v-if="walletOnlyPayment && expectedOnlinePayCents > 0" class="is-warn">
              ·&nbsp;{{ t('payment.walletInsufficientHint') }}
            </span>
          </div>

          <!-- 缓存支付 -->
          <div v-if="cachedPayment" class="j8-pd-alert is-warm" style="margin-top: 14px;">
            <div style="font-weight: 800; margin-bottom: 6px;">{{ t('payment.cachedTitle') }}</div>
            <div style="margin-bottom: 8px;">{{ t('payment.cachedHint', { channel: cachedChannelName }) }}</div>
            <div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center;">
              <button type="button" class="j8-pd-cart" @click="restoreCachedPayment">
                <RotateCcw />
                {{ t('payment.useCached') }}
              </button>
              <small style="font-size: 11px; opacity: 0.8;">{{ t('payment.cachedCreateHint') }}</small>
            </div>
          </div>

          <!-- 支付方式网格 -->
          <div v-if="!walletOnlyPayment" class="j8-ic-pay" style="margin-top: 14px;">
            <h3><Wallet />{{ t('qiaozhi.detail.payTitle') }}</h3>
            <div v-if="!channels.length" class="j8-ic-alert">
              {{ showBalanceOption ? t('payment.channelEmptyUseBalance') : t('payment.channelEmpty') }}
            </div>
            <div v-else class="j8-ic-channels">
              <button
                v-for="ch in channels"
                :key="ch.id"
                type="button"
                class="j8-ic-channel"
                :class="{ 'is-active': selectedChannelId === Number(ch.id) }"
                :disabled="isChannelDisabledForAmount(ch)"
                :title="channelAmountLimitHint(ch) || undefined"
                @click="selectedChannelId = Number(ch.id)"
              >
                <i></i>
                <img v-if="ch.icon" :src="ch.icon" alt="" />
                <span>
                  <b>{{ ch.name }}</b>
                  <small v-if="channelAmountLimitHint(ch)">{{ channelAmountLimitHint(ch) }}</small>
                </span>
              </button>
            </div>
          </div>

          <div class="j8-pd-pay-total">
            <span>{{ t('payment.payableAmountLabel') }}</span>
            <strong>{{ payableAmountDisplay }}</strong>
          </div>

          <!-- 提示 -->
          <div v-if="paymentAlert" class="j8-pd-alert" :class="paymentAlert.level === 'error' ? 'is-danger' : 'is-warm'">
            {{ paymentAlert.message }}
          </div>
          <div
            v-else-if="!walletOnlyPayment && requiresOnlineChannel && !selectedChannel"
            class="j8-pd-alert is-warm"
          >
            {{ t('payment.selectChannelError') }}
          </div>
          <div
            v-else-if="walletOnlyPayment && expectedOnlinePayCents > 0"
            class="j8-pd-alert is-warm"
          >
            {{ t('payment.walletInsufficientHint') }}
          </div>
          <div
            v-else-if="!requiresOnlineChannel"
            class="j8-pd-note is-accent"
          >
            <b>{{ t('payment.walletPayOnly') }}</b>
          </div>

          <div class="j8-pd-pay-actions">
            <button type="button" class="j8-pd-buy" :disabled="!canSubmitPayment" @click="handlePayment">
              <Loader2 v-if="submitting" class="j8-ic-spin" />
              <CreditCard v-else />
              {{ submitting ? t('payment.submitting') : t('payment.submitButton') }}
            </button>
            <button type="button" class="j8-pd-cart" :disabled="loading" @click="handleRefresh">
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
const totalParts = computed(() => splitPrice(order.value?.total_amount ?? '0'))

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
/* 页面级补充（基于 j8-pd-* / j8-ic-* 的组合排版） */
.j8-pd-pay-card { display: grid; gap: 16px; padding: 24px 24px 20px; grid-template-columns: minmax(0, 1fr); }
.j8-pd-pay-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 14px; flex-wrap: wrap; }
.j8-pd-pay-head > div:first-child { display: grid; gap: 6px; min-width: 0; }
.j8-pd-pay-head h1 { font-size: 24px; font-weight: 900; line-height: 1.3; color: var(--text); }
.j8-pd-pay-head h2 { font-size: 20px; font-weight: 900; line-height: 1.3; color: var(--text); }
.j8-pd-pay-sub { display: inline-flex; align-items: center; gap: 8px; flex-wrap: wrap; font-size: 13px; color: var(--muted); }
.j8-pd-pay-sub code { font-family: inherit; font-weight: 700; color: var(--text); font-variant-numeric: tabular-nums; }
.j8-pd-pay-head-actions { display: inline-flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.j8-pd-pay-head-actions .j8-pd-cart { height: 36px; padding: 0 12px; font-size: 13px; }
.j8-pd-pay-head-actions .j8-pd-cart svg { width: 15px; height: 15px; }

.j8-pd-pay-meta { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 12px; padding-top: 4px; border-top: 1px solid var(--line); }
.j8-pd-pay-meta > div { display: grid; gap: 4px; min-width: 0; padding-top: 10px; }
.j8-pd-pay-meta small { font-size: 11.5px; color: var(--muted); font-weight: 600; }
.j8-pd-pay-meta b { font-size: 13.5px; font-weight: 800; color: var(--text); overflow-wrap: anywhere; }

.j8-pd-pay-items { padding-top: 4px; border-top: 1px solid var(--line); }
.j8-pd-pay-items h3 { margin: 10px 0 8px; font-size: 13px; font-weight: 700; color: var(--muted); }
.j8-pd-pay-items ul { display: grid; gap: 8px; }
.j8-pd-pay-items li { display: flex; align-items: flex-start; justify-content: space-between; gap: 10px; padding: 10px 12px; border-radius: 11px; background: var(--surface-2); }
.j8-pd-pay-item-main { display: grid; gap: 4px; min-width: 0; }
.j8-pd-pay-item-main b { font-size: 13.5px; font-weight: 800; color: var(--text); overflow-wrap: anywhere; }
.j8-pd-pay-item-main small { font-size: 11.5px; color: var(--muted); }
.j8-pd-pay-item-amount { flex: 0 0 auto; font-size: 13px; font-weight: 800; color: var(--text); white-space: nowrap; font-variant-numeric: tabular-nums; }

.j8-pd-pay-result { display: grid; gap: 16px; }
.j8-pd-pay-qrbox { display: grid; grid-template-columns: minmax(0, 320px) minmax(0, 1fr); gap: 20px; align-items: center; padding: 16px; border-radius: 16px; background: var(--surface-2); border: 1px solid var(--line); }
.j8-pd-pay-qr { position: relative; width: 100%; max-width: 300px; aspect-ratio: 1 / 1; margin: 0 auto; padding: 14px; border-radius: 14px; background: #fff; display: grid; place-items: center; box-shadow: var(--shadow-soft); }
.j8-pd-pay-qr img { width: 100%; height: 100%; object-fit: contain; }
.j8-pd-pay-qr-placeholder { display: grid; place-items: center; color: var(--muted); }
.j8-pd-pay-qr-placeholder .j8-ic-spin { width: 32px; height: 32px; }
.j8-pd-pay-qr-side { display: grid; gap: 10px; min-width: 0; }
.j8-pd-pay-qr-title { font-size: 15px; font-weight: 800; color: var(--text); }
.j8-pd-pay-qr-amount { display: flex; align-items: baseline; justify-content: space-between; gap: 10px; padding: 10px 12px; border-radius: 12px; background: var(--surface); border: 1px solid var(--line); }
.j8-pd-pay-qr-amount span { font-size: 12px; color: var(--muted); font-weight: 600; }
.j8-pd-pay-qr-amount strong { font-size: 20px; font-weight: 900; color: var(--text); font-variant-numeric: tabular-nums; }
.j8-pd-pay-polling { display: inline-flex; align-items: center; gap: 8px; padding: 10px 12px; border-radius: 11px; background: var(--accent-soft); color: var(--accent-strong); font-size: 13px; font-weight: 700; }
.j8-pd-pay-polling .j8-ic-spin { width: 16px; height: 16px; }

.j8-pd-pay-crypto { display: grid; gap: 8px; padding: 14px; border-radius: 14px; background: var(--surface-2); border: 1px solid var(--line); }
.j8-pd-pay-crypto-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; padding-bottom: 8px; border-bottom: 1px solid var(--line); }
.j8-pd-pay-crypto-row:last-child { border-bottom: 0; padding-bottom: 0; }
.j8-pd-pay-crypto-row small { flex: 0 0 auto; font-size: 11.5px; color: var(--muted); font-weight: 600; }
.j8-pd-pay-crypto-row b { min-width: 0; text-align: right; font-size: 13.5px; font-weight: 800; color: var(--text); overflow-wrap: anywhere; word-break: break-all; }
.j8-pd-pay-crypto-row em { font-style: normal; font-weight: 500; color: var(--muted); margin-left: 4px; }
.j8-pd-pay-crypto-action { display: inline-flex; align-items: center; gap: 10px; flex-wrap: wrap; }

.j8-pd-pay-link { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.j8-pd-pay-link .j8-pd-buy { width: auto; min-width: 200px; padding: 0 20px; }

.j8-pd-pay-channel { display: grid; gap: 0; }
.j8-pd-pay-channel h2 { font-size: 18px; font-weight: 900; color: var(--text); margin-bottom: 4px; }
.j8-pd-pay-walletinfo { display: flex; flex-wrap: wrap; gap: 4px 8px; margin-top: 8px; font-size: 12px; color: var(--muted); }
.j8-pd-pay-walletinfo .is-warn { color: var(--amber); font-weight: 700; }

.j8-pd-pay-total { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-top: 16px; padding: 14px 16px; border-radius: 14px; background: var(--surface-2); border: 1px solid var(--line); }
.j8-pd-pay-total span { font-size: 13px; color: var(--muted); font-weight: 600; }
.j8-pd-pay-total strong { font-size: 22px; font-weight: 900; color: var(--text); font-variant-numeric: tabular-nums; }

.j8-pd-pay-actions { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); gap: 10px; margin-top: 14px; }
.j8-pd-pay-actions .j8-pd-buy { width: 100%; }

@media (max-width: 720px) {
  .j8-pd-pay-qrbox { grid-template-columns: minmax(0, 1fr); }
  .j8-pd-pay-actions { grid-template-columns: minmax(0, 1fr); }
  .j8-pd-pay-head { align-items: stretch; }
  .j8-pd-pay-head-actions { width: 100%; }
}
</style>
