<template>
  <div ref="rootEl" class="j8-container j8-xcheck-main">
    <!-- 返回条 -->
    <button type="button" class="j8-pd-back" @click="goBack">
      <ChevronLeft />
      {{ backLabel }}
    </button>

    <!-- 顶部文案 -->
    <section class="j8-xcheck-hero">
      <span class="j8-eyebrow">{{ t('qiaozhi.checkout.eyebrow') }}</span>
      <h1>{{ t('checkout.title') }}</h1>
      <p>{{ t('qiaozhi.checkout.subtitle') }}</p>
    </section>

    <!-- 空购物车 -->
    <div v-if="cartItems.length === 0" class="j8-empty">
      <PackageOpen />
      <b>{{ t('checkout.empty') }}</b>
      <span>{{ t('qiaozhi.checkout.emptyHint') }}</span>
      <RouterLink to="/products" class="j8-btn-primary">{{ t('checkout.emptyAction') }}</RouterLink>
    </div>

    <!-- 结算主体 -->
    <div v-else class="j8-xcheck-grid">
      <!-- 左：商品 + 表单 -->
      <div class="j8-xcheck-left">
        <!-- 订单明细 -->
        <section class="j8-pd-card j8-xcheck-card j8-reveal is-pending" :data-reveal="1">
          <header class="j8-xcheck-card-head">
            <span class="j8-eyebrow">{{ t('qiaozhi.checkout.sectionItems') }}</span>
            <h2>{{ t('checkout.itemsTitle') }}</h2>
          </header>
          <ul class="j8-xcheck-items">
            <li
              v-for="item in cartItems"
              :key="cartItemKey(item)"
              class="j8-xcheck-item"
              :class="{ 'is-warn': itemStockExceeded(item) }"
            >
              <div class="j8-xcheck-item-thumb">
                <img
                  v-if="checkoutItemImage(item)"
                  :src="checkoutItemImage(item)"
                  :alt="getLocalizedText(item.title)"
                  loading="lazy"
                />
                <span v-else class="j8-pd-logo j8-brand-icon">
                  <b>{{ (getLocalizedText(item.title) || '·').slice(0, 1) }}</b>
                </span>
              </div>
              <div class="j8-xcheck-item-body">
                <RouterLink
                  class="j8-xcheck-item-title"
                  :to="`/products/${item.slug}`"
                >
                  {{ getLocalizedText(item.title) }}
                </RouterLink>
                <div class="j8-xcheck-item-meta">
                  <span>{{ t('checkout.quantityLabel') }} × {{ item.quantity }}</span>
                  <span v-if="itemSkuDisplay(item)">{{ itemSkuDisplay(item) }}</span>
                </div>
                <p
                  v-if="itemStockHint(item)"
                  class="j8-xcheck-item-stock"
                  :class="{ 'is-warn': itemStockExceeded(item) }"
                >
                  {{ itemStockHint(item) }}
                </p>
              </div>
              <div class="j8-xcheck-item-price">
                <div
                  class="j8-pd-price"
                  :class="{ 'is-discount': checkoutItemHasPriceDiscount(item) }"
                >
                  <strong>
                    {{ checkoutItemPriceParts(item).integer }}<small>{{ checkoutItemPriceParts(item).decimal }}</small>
                  </strong>
                  <span class="j8-xcheck-item-cur">{{ checkoutItemCurrency }}</span>
                </div>
                <del
                  v-if="checkoutItemHasPriceDiscount(item)"
                  class="j8-xcheck-item-old"
                >
                  {{ checkoutItemOriginalPriceParts(item).integer }}{{ checkoutItemOriginalPriceParts(item).decimal }} {{ checkoutItemCurrency }}
                </del>
              </div>
            </li>
          </ul>
        </section>

        <!-- 交付资料 / 游客 / 优惠码 表单 -->
        <section class="j8-pd-card j8-xcheck-card j8-reveal is-pending" :data-reveal="1">
          <header class="j8-xcheck-card-head">
            <span class="j8-eyebrow">{{ t('qiaozhi.checkout.sectionInfo') }}</span>
            <h2>{{ t('qiaozhi.checkout.infoTitle') }}</h2>
          </header>

          <form class="j8-ic" novalidate @submit.prevent="handleSubmit">
            <!-- 交付信息 -->
            <template v-if="manualFormProducts.length">
              <div
                v-for="product in manualFormProducts"
                :key="product.itemKey"
                class="j8-xcheck-manual"
              >
                <h3>
                  {{ manualItemTitle(product) }}
                </h3>
                <p class="j8-ic-hint">{{ t('checkout.manualFormTip') }}</p>
                <div class="j8-xcheck-manual-grid">
                  <div
                    v-for="field in product.fields"
                    :key="`${product.itemKey}-${field.key}`"
                    class="j8-ic-field"
                  >
                    <label :for="manualFieldId(product.itemKey, field.key)">
                      {{ getManualFieldLabel(field) }}
                      <b v-if="field.required">*</b>
                    </label>

                    <textarea
                      v-if="field.type === 'textarea'"
                      :id="manualFieldId(product.itemKey, field.key)"
                      :value="getManualFieldValue(product.itemKey, field.key)"
                      rows="3"
                      :placeholder="getManualFieldPlaceholder(field)"
                      :aria-invalid="manualFieldHasError(product.itemKey, field.key)"
                      :aria-describedby="manualFieldHasError(product.itemKey, field.key) ? manualFieldErrId(product.itemKey, field.key) : undefined"
                      @input="setManualFieldValue(product.itemKey, field.key, ($event.target as HTMLTextAreaElement).value)"
                    ></textarea>

                    <J8Select
                      v-else-if="field.type === 'select'"
                      :id="manualFieldId(product.itemKey, field.key)"
                      :model-value="getManualFieldValue(product.itemKey, field.key)"
                      :items="[{ value: '', label: t('checkout.manualFormSelectPlaceholder') }, ...(field.options || []).map((option: string) => ({ value: option, label: option }))]"
                      :aria-invalid="manualFieldHasError(product.itemKey, field.key)"
                      :aria-describedby="manualFieldHasError(product.itemKey, field.key) ? manualFieldErrId(product.itemKey, field.key) : undefined"
                      @update:model-value="setManualFieldValue(product.itemKey, field.key, String($event))"
                    />

                    <div
                      v-else-if="field.type === 'radio'"
                      class="j8-ic-options"
                      :aria-labelledby="manualFieldId(product.itemKey, field.key)"
                    >
                      <label v-for="option in field.options" :key="option">
                        <input
                          type="radio"
                          :name="`j8-ic-radio-${product.itemKey}-${field.key}`"
                          :value="option"
                          :checked="getManualFieldValue(product.itemKey, field.key) === option"
                          @change="setManualFieldValue(product.itemKey, field.key, option)"
                        />
                        <span>{{ option }}</span>
                      </label>
                    </div>

                    <div
                      v-else-if="field.type === 'checkbox'"
                      class="j8-ic-options"
                      :aria-labelledby="manualFieldId(product.itemKey, field.key)"
                    >
                      <label v-for="option in field.options" :key="option">
                        <input
                          type="checkbox"
                          :value="option"
                          :checked="manualCheckboxChecked(product.itemKey, field.key, option)"
                          @change="toggleManualCheckbox(product.itemKey, field.key, option, ($event.target as HTMLInputElement).checked)"
                        />
                        <span>{{ option }}</span>
                      </label>
                    </div>

                    <input
                      v-else
                      :id="manualFieldId(product.itemKey, field.key)"
                      :type="field.type === 'number' ? 'number' : field.type === 'email' ? 'email' : field.type === 'phone' ? 'tel' : 'text'"
                      :value="getManualFieldValue(product.itemKey, field.key)"
                      :placeholder="getManualFieldPlaceholder(field)"
                      :aria-invalid="manualFieldHasError(product.itemKey, field.key)"
                      :aria-describedby="manualFieldHasError(product.itemKey, field.key) ? manualFieldErrId(product.itemKey, field.key) : undefined"
                      @input="setManualFieldValue(product.itemKey, field.key, ($event.target as HTMLInputElement).value)"
                    />

                    <p
                      v-if="submitAttempted && manualFieldError(product.itemKey, field.key)"
                      :id="manualFieldErrId(product.itemKey, field.key)"
                      class="j8-ic-error"
                    >
                      {{ manualFieldError(product.itemKey, field.key) }}
                    </p>
                  </div>
                </div>
              </div>
            </template>

            <!-- 游客 -->
            <template v-if="!userAuthStore.isAuthenticated">
              <div class="j8-xcheck-mode">
                <h3>{{ t('checkout.modeTitle') }}</h3>
                <div class="j8-xcheck-mode-tabs" role="tablist">
                  <button
                    type="button"
                    role="tab"
                    :aria-selected="checkoutMode === 'guest'"
                    :class="{ 'is-active': checkoutMode === 'guest' }"
                    @click="checkoutMode = 'guest'"
                  >
                    {{ t('checkout.guestPurchase') }}
                  </button>
                  <RouterLink
                    to="/auth/login"
                    role="tab"
                    class="j8-xcheck-mode-link"
                  >
                    {{ t('checkout.memberPurchase') }}
                  </RouterLink>
                </div>
              </div>

              <template v-if="checkoutMode === 'guest'">
                <div class="j8-ic-field">
                  <label for="j8-xcheck-email">
                    {{ t('qiaozhi.detail.email') }}<b>*</b>
                  </label>
                  <input
                    id="j8-xcheck-email"
                    v-model="guestEmail"
                    type="email"
                    autocomplete="email"
                    :placeholder="t('checkout.guestEmailPlaceholder')"
                    :aria-invalid="submitAttempted && !!guestEmail && !guestEmailValid"
                    :aria-describedby="submitAttempted && !!guestEmail && !guestEmailValid ? 'j8-xcheck-email-err' : undefined"
                  />
                  <p
                    v-if="submitAttempted && !!guestEmail && !guestEmailValid"
                    id="j8-xcheck-email-err"
                    class="j8-ic-error"
                  >
                    {{ t('error.email_invalid') }}
                  </p>
                </div>

                <div class="j8-ic-field">
                  <label for="j8-xcheck-password">
                    {{ t('qiaozhi.detail.orderPassword') }}<b>*</b>
                  </label>
                  <input
                    id="j8-xcheck-password"
                    v-model="guestPassword"
                    type="password"
                    autocomplete="new-password"
                    :placeholder="t('checkout.guestPasswordPlaceholder')"
                  />
                </div>

                <div v-if="guestCaptchaEnabled" class="j8-ic-field">
                  <label>{{ t('auth.common.captchaLabel') }}<b>*</b></label>
                  <ImageCaptcha
                    v-if="captchaProvider === 'image'"
                    ref="guestImageCaptchaRef"
                    v-model="guestCaptchaPayload"
                    :disabled="submitting"
                    @config-stale="handleGuestCaptchaConfigStale"
                  />
                  <TurnstileCaptcha
                    v-else-if="captchaProvider === 'turnstile'"
                    ref="guestTurnstileRef"
                    v-model="guestTurnstileToken"
                    :site-key="guestTurnstileSiteKey"
                  />
                </div>

                <div class="j8-xcheck-notice">
                  <b>{{ t('checkout.guestInstructions.title') }}</b>
                  <ul>
                    <li>{{ t('checkout.guestInstructions.email') }}</li>
                    <li>{{ t('checkout.guestInstructions.password') }}</li>
                  </ul>
                </div>
              </template>
            </template>
            <p v-else class="j8-ic-hint is-ok">{{ t('qiaozhi.detail.loggedInHint') }}</p>

            <!-- 优惠码 -->
            <div v-if="!isResellerTenant" class="j8-ic-field j8-ic-with-action">
              <label for="j8-xcheck-coupon">{{ t('checkout.couponTitle') }}</label>
              <input
                id="j8-xcheck-coupon"
                v-model="couponCode"
                type="text"
                autocomplete="off"
                :placeholder="t('checkout.couponPlaceholder')"
              />
            </div>
          </form>
        </section>
      </div>

      <!-- 右：合计 + 支付 -->
      <aside class="j8-xcheck-aside j8-reveal is-pending" :data-reveal="1">
        <div class="j8-xcheck-summary">
          <header class="j8-xcheck-card-head">
            <span class="j8-eyebrow">{{ t('qiaozhi.checkout.sectionSummary') }}</span>
            <h2>{{ t('checkout.submitTitle') }}</h2>
          </header>

          <p class="j8-ic-hint">{{ t('checkout.submitHint') }}</p>

          <dl class="j8-xcheck-rows">
            <div class="j8-xcheck-row">
              <dt>{{ t('cart.itemsCount') }}</dt>
              <dd>{{ totalItems }}</dd>
            </div>
            <div class="j8-xcheck-row">
              <dt>{{ t('checkout.previewOriginal') }}</dt>
              <dd>{{ formatPrice(previewOriginal, previewCurrency) }}</dd>
            </div>
            <template v-if="!isResellerTenant">
              <div class="j8-xcheck-row">
                <dt>{{ t('checkout.previewCoupon') }}</dt>
                <dd :class="{ 'is-discount': hasPositiveAmount(previewCoupon) }">
                  {{ formatDiscountPrice(previewCoupon, previewCurrency) }}
                </dd>
              </div>
              <div class="j8-xcheck-row">
                <dt>{{ t('checkout.previewPromotion') }}</dt>
                <dd :class="{ 'is-discount': hasPositiveAmount(previewPromotion) }">
                  {{ formatDiscountPrice(previewPromotion, previewCurrency) }}
                </dd>
              </div>
              <div class="j8-xcheck-row">
                <dt>{{ t('checkout.previewWholesale') }}</dt>
                <dd :class="{ 'is-saving': hasPositiveAmount(previewWholesale) }">
                  {{ formatDiscountPrice(previewWholesale, previewCurrency) }}
                </dd>
              </div>
            </template>
            <div v-if="hasPositiveAmount(previewMemberDiscount)" class="j8-xcheck-row">
              <dt>{{ t('checkout.previewMemberDiscount') }}</dt>
              <dd class="is-saving">-{{ formatPrice(previewMemberDiscount, previewCurrency) }}</dd>
            </div>
          </dl>

          <div class="j8-ic-total">
            <span>{{ t('checkout.previewTotal') }}</span>
            <strong>{{ formatPrice(previewTotal, previewCurrency) }}</strong>
          </div>

          <p v-if="previewLoading || couponRefreshing" class="j8-ic-hint">
            {{ previewStatusText }}
          </p>

          <!-- 支付方式 -->
          <div class="j8-ic-pay">
            <h3>
              <Wallet />
              {{ t('checkout.paymentMethod') }}
            </h3>

            <!-- 钱包余额 -->
            <label v-if="showBalanceOption" class="j8-ic-balance" :class="{ 'is-active': useBalance }">
              <Wallet />
              <span>
                <b>{{ t('payment.walletBalanceLabel') }}</b>
                <small>
                  {{ walletLoading ? t('common.loading') : formatPrice(walletBalance, previewCurrency) }}
                </small>
              </span>
              <em v-if="walletOnlyPayment">{{ t('payment.walletOnlyHint') }}</em>
              <input
                type="checkbox"
                :checked="useBalance"
                :disabled="walletOnlyPayment"
                @change="useBalance = ($event.target as HTMLInputElement).checked"
              />
            </label>

            <!-- 钱包抵扣摘要 -->
            <ul
              v-if="showBalanceOption && useBalance"
              class="j8-xcheck-wallet-info"
            >
              <li>
                <span>{{ t('payment.walletDeductLabel') }}</span>
                <b>{{ expectedWalletPaidDisplay }}</b>
              </li>
              <li v-if="!walletOnlyPayment">
                <span>{{ t('payment.onlinePayLabel') }}</span>
                <b>{{ expectedOnlinePayDisplay }}</b>
              </li>
              <li
                v-if="walletOnlyPayment && expectedOnlinePayCents > 0"
                class="is-warn"
              >
                {{ t('payment.walletInsufficientHint') }}
              </li>
            </ul>

            <!-- 在线支付渠道 -->
            <template v-if="!walletOnlyPayment">
              <div
                v-if="requiresOnlineChannel && paymentChannels.length > 0"
                class="j8-ic-channels"
              >
                <button
                  v-for="channel in paymentChannels"
                  :key="channel.id"
                  type="button"
                  class="j8-ic-channel"
                  :class="{ 'is-active': selectedChannelId === channel.id && !isChannelDisabledForAmount(channel) }"
                  :disabled="isChannelDisabledForAmount(channel)"
                  :title="isChannelDisabledForAmount(channel) ? channelAmountLimitHint(channel) : ''"
                  @click="handleSelectChannel(channel)"
                >
                  <i></i>
                  <img
                    v-if="channel.icon"
                    :src="getImageUrl(channel.icon)"
                    alt=""
                    loading="lazy"
                  />
                  <span>
                    <b>{{ channel.name }}</b>
                    <small v-if="channel.fee_policy === 'customer_surcharge'">
                      {{ t('payment.feeLabel') }} {{ formatChannelFeeRate(channel) }} · {{ formatChannelFixedFee(channel) }}
                    </small>
                    <small v-else-if="isChannelDisabledForAmount(channel)">
                      {{ channelAmountLimitHint(channel) }}
                    </small>
                  </span>
                </button>
              </div>
              <p
                v-else-if="requiresOnlineChannel && paymentChannels.length === 0"
                class="j8-ic-hint"
              >
                {{ t('checkout.noPaymentChannels') }}
              </p>
            </template>

            <p
              v-if="!requiresOnlineChannel"
              class="j8-ic-hint is-ok"
            >
              {{ t('checkout.walletCoversAll') }}
            </p>
          </div>

          <!-- 提示 / 错误 -->
          <p
            v-if="checkoutAlert"
            class="j8-ic-alert"
            :class="{ 'is-error': checkoutAlert.level === 'error' }"
            role="alert"
          >
            {{ checkoutAlert.message }}
          </p>

          <!-- 提交 -->
          <button
            type="button"
            class="j8-xcheck-btn"
            :disabled="!canSubmit"
            @click="handleSubmit"
          >
            <Loader2 v-if="submitting" class="j8-xcheck-spin" />
            <CreditCard v-else />
            {{ submitting ? t('checkout.submitting') : t('checkout.submitButton') }}
          </button>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, CreditCard, Loader2, PackageOpen, Wallet } from 'lucide-vue-next'
import ImageCaptcha from '../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../components/captcha/TurnstileCaptcha.vue'
import { useCheckout } from '../../composables/useCheckout'
import { useLocalized } from '../../composables/useProduct'
import { setupReveal } from './utils/reveal'
import J8Select from './components/J8Select.vue'

const { t } = useI18n()
const router = useRouter()
const { getLocalizedText: resolveLocalizedText } = useLocalized()

const {
  userAuthStore,
  getLocalizedText,
  formatPrice,
  getImageUrl,
  // items
  cartItems,
  totalItems,
  cartItemKey,
  checkoutItemImage,
  itemSkuDisplay,
  itemStockExceeded,
  itemStockHint,
  checkoutItemCurrency,
  checkoutItemPriceParts,
  checkoutItemOriginalPriceParts,
  checkoutItemHasPriceDiscount,
  // manual form
  manualFormProducts,
  manualFormData,
  submitAttempted,
  getManualFieldLabel,
  getManualFieldPlaceholder,
  manualFieldError,
  // coupon / reseller
  couponCode,
  isResellerTenant,
  // guest
  checkoutMode,
  guestEmail,
  guestPassword,
  guestEmailValid,
  guestCaptchaEnabled,
  captchaProvider,
  guestCaptchaPayload,
  guestTurnstileToken,
  guestTurnstileSiteKey,
  guestImageCaptchaRef,
  guestTurnstileRef,
  handleGuestCaptchaConfigStale,
  // preview
  previewCurrency,
  previewOriginal,
  previewCoupon,
  previewPromotion,
  previewWholesale,
  previewMemberDiscount,
  previewTotal,
  previewLoading,
  couponRefreshing,
  previewStatusText,
  hasPositiveAmount,
  formatDiscountPrice,
  checkoutAlert,
  // wallet
  showBalanceOption,
  walletLoading,
  walletBalance,
  useBalance,
  walletOnlyPayment,
  expectedWalletPaidDisplay,
  expectedOnlinePayDisplay,
  expectedOnlinePayCents,
  // channels
  requiresOnlineChannel,
  paymentChannels,
  selectedChannelId,
  isChannelDisabledForAmount,
  channelAmountLimitHint,
  handleSelectChannel,
  formatChannelFeeRate,
  formatChannelFixedFee,
  // submit
  submitting,
  canSubmit,
  handleSubmit,
} = useCheckout()

// vue-tsc 不把模板字符串 ref 算作使用，显式标记避免告警。
void guestImageCaptchaRef
void guestTurnstileRef

// ========== 入场动画 ==========
const rootEl = ref<HTMLElement | null>(null)
let cleanupReveal: (() => void) | null = null
const armReveal = () => {
  cleanupReveal?.()
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
}

onMounted(() => {
  armReveal()
})
onUnmounted(() => {
  cleanupReveal?.()
})
watch(
  () => [cartItems.value.length, manualFormProducts.value.length],
  () => armReveal(),
)

// ========== 返回按钮 ==========
const backLabel = computed(() => (totalItems.value > 0 ? t('qiaozhi.checkout.backToCart') : t('qiaozhi.detail.back')))
const goBack = () => {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.push(totalItems.value > 0 ? '/cart' : '/products')
}

// ========== 交付资料辅助 ==========
const manualItemTitle = (product: { title: any; skuCount: number }) => {
  const name = resolveLocalizedText(product.title)
  if (!product.skuCount || product.skuCount <= 1) return name
  return `${name} (${t('checkout.manualFormAppliesToSkuCount', { count: product.skuCount })})`
}

const manualFieldId = (itemKey: string, fieldKey: string) => `j8-mf-${itemKey}-${fieldKey}`
const manualFieldErrId = (itemKey: string, fieldKey: string) => `j8-mf-err-${itemKey}-${fieldKey}`

const getManualFieldValue = (itemKey: string, fieldKey: string) => {
  const value = manualFormData.value[itemKey]?.[fieldKey]
  if (value == null) return ''
  if (Array.isArray(value)) return value.join(',')
  return String(value)
}

const setManualFieldValue = (itemKey: string, fieldKey: string, value: any) => {
  const next = { ...manualFormData.value }
  const row = { ...(next[itemKey] || {}) }
  row[fieldKey] = value
  next[itemKey] = row
  manualFormData.value = next
}

const manualCheckboxChecked = (itemKey: string, fieldKey: string, option: string) => {
  const value = manualFormData.value[itemKey]?.[fieldKey]
  return Array.isArray(value) && value.includes(option)
}

const toggleManualCheckbox = (itemKey: string, fieldKey: string, option: string, checked: boolean) => {
  const current = manualFormData.value[itemKey]?.[fieldKey]
  const list = Array.isArray(current) ? [...current] : []
  if (checked) {
    if (!list.includes(option)) list.push(option)
  } else {
    const idx = list.indexOf(option)
    if (idx !== -1) list.splice(idx, 1)
  }
  setManualFieldValue(itemKey, fieldKey, list)
}

const manualFieldHasError = (itemKey: string, fieldKey: string) => {
  if (!submitAttempted.value) return false
  return !!manualFieldError(itemKey, fieldKey)
}
</script>

<style scoped>
/* qiaozhi Checkout 页面独立版式 —— 复用 ji8.css 的 j8-* 原子，仅在此补齐布局 */
.j8-xcheck-main { padding-top: 24px; padding-bottom: 70px; }

.j8-xcheck-grid {
  display: grid;
  gap: 22px;
  align-items: start;
  grid-template-columns: minmax(0, 1fr);
}
@media (min-width: 960px) {
  .j8-xcheck-grid { grid-template-columns: minmax(0, 1.4fr) minmax(300px, 0.9fr); }
}

.j8-xcheck-left { display: grid; gap: 18px; min-width: 0; }

:deep(.j8-xcheck-card) {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
  padding: 22px 22px 24px;
}

.j8-xcheck-card-head { display: grid; gap: 4px; }
.j8-xcheck-card-head h2 {
  margin: 0;
  color: var(--text);
  font-size: 20px;
  font-weight: 700;
  line-height: 1.25;
}

/* 商品列表 */
.j8-xcheck-items { list-style: none; margin: 0; padding: 0; display: grid; gap: 10px; }
.j8-xcheck-item {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr) auto;
  gap: 14px;
  align-items: center;
  padding: 12px;
  border: 1px solid var(--line);
  border-radius: 14px;
  background: var(--surface-2);
}
.j8-xcheck-item.is-warn {
  border-color: color-mix(in srgb, var(--amber) 55%, var(--line));
  background: var(--amber-soft);
}
.j8-xcheck-item-thumb {
  width: 72px; height: 72px; border-radius: 12px; overflow: hidden;
  display: grid; place-items: center;
  background: var(--surface); border: 1px solid var(--line);
}
.j8-xcheck-item-thumb img { width: 100%; height: 100%; object-fit: cover; }
.j8-xcheck-item-thumb .j8-pd-logo { font-size: 22px; }

.j8-xcheck-item-body { min-width: 0; display: grid; gap: 4px; }
.j8-xcheck-item-title {
  display: block; min-width: 0;
  color: var(--text); font-size: 14px; font-weight: 700;
  text-decoration: none;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.j8-xcheck-item-title:hover { color: var(--accent-strong); }
.j8-xcheck-item-meta {
  display: flex; flex-wrap: wrap; gap: 4px 12px;
  color: var(--muted); font-size: 12px;
}
.j8-xcheck-item-stock { margin: 0; font-size: 12px; color: var(--muted); }
.j8-xcheck-item-stock.is-warn { color: var(--amber); }

.j8-xcheck-item-price { text-align: right; min-width: 86px; }
.j8-xcheck-item-price .j8-pd-price {
  display: inline-flex; align-items: baseline; gap: 2px;
  color: var(--text); font-variant-numeric: tabular-nums;
}
.j8-xcheck-item-price .j8-pd-price strong {
  font-size: 18px; font-weight: 800; line-height: 1;
}
.j8-xcheck-item-price .j8-pd-price small { font-size: 12px; font-weight: 700; }
.j8-xcheck-item-price .j8-pd-price.is-discount { color: var(--red); }
.j8-xcheck-item-cur { margin-left: 3px; color: var(--muted); font-size: 11px; font-weight: 700; }
.j8-xcheck-item-old {
  display: block; margin-top: 2px;
  color: var(--muted); font-size: 11px; text-decoration: line-through;
}

/* 交付资料分组 */
.j8-xcheck-manual {
  display: grid; gap: 10px;
  padding: 14px; border-radius: 14px;
  background: var(--surface-2); border: 1px solid var(--line);
}
.j8-xcheck-manual h3 {
  margin: 0; color: var(--text); font-size: 14px; font-weight: 800;
}
.j8-xcheck-manual-grid {
  display: grid; gap: 12px;
  grid-template-columns: 1fr;
}
@media (min-width: 640px) {
  .j8-xcheck-manual-grid { grid-template-columns: 1fr 1fr; }
}

/* 游客 / 模式 */
.j8-xcheck-mode { display: grid; gap: 10px; }
.j8-xcheck-mode h3 { margin: 0; color: var(--muted); font-size: 13px; font-weight: 700; }
.j8-xcheck-mode-tabs { display: flex; flex-wrap: wrap; gap: 10px; }
.j8-xcheck-mode-tabs button,
.j8-xcheck-mode-tabs .j8-xcheck-mode-link {
  display: inline-flex; align-items: center;
  height: 40px; padding: 0 16px;
  border: 1px solid var(--line); border-radius: 12px;
  background: var(--surface); color: var(--text);
  font-size: 13px; font-weight: 700; text-decoration: none;
  cursor: pointer; transition: border-color 0.15s, background 0.15s, color 0.15s;
}
.j8-xcheck-mode-tabs button:hover,
.j8-xcheck-mode-tabs .j8-xcheck-mode-link:hover {
  border-color: var(--accent-line); color: var(--accent-strong);
}
.j8-xcheck-mode-tabs button.is-active {
  border-color: var(--accent); background: var(--accent-soft); color: var(--accent-strong);
}

/* 游客说明 */
.j8-xcheck-notice {
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid var(--accent-line);
  background: var(--accent-soft);
  color: var(--accent-strong);
  font-size: 12.5px;
  line-height: 1.6;
}
.j8-xcheck-notice b { display: block; margin-bottom: 4px; font-size: 13px; }
.j8-xcheck-notice ul { margin: 0; padding-left: 18px; display: grid; gap: 3px; }

/* 右侧合计 */
.j8-xcheck-aside { min-width: 0; }
.j8-xcheck-summary {
  position: sticky;
  top: calc(var(--site-header-height, 72px) + 16px);
  display: grid;
  gap: 14px;
  padding: 22px;
  border: 1px solid var(--line);
  border-radius: var(--r-card, 20px);
  background: var(--surface);
  box-shadow: var(--shadow-soft);
}

.j8-xcheck-rows { display: grid; gap: 8px; margin: 0; }
.j8-xcheck-row {
  display: flex; align-items: baseline; justify-content: space-between; gap: 12px;
  color: var(--muted); font-size: 13px;
}
.j8-xcheck-row dt { margin: 0; }
.j8-xcheck-row dd {
  margin: 0; color: var(--text);
  font-variant-numeric: tabular-nums; font-feature-settings: 'tnum';
}
.j8-xcheck-row dd.is-discount { color: var(--red); }
.j8-xcheck-row dd.is-saving { color: var(--accent-strong); }

/* 钱包抵扣摘要 */
.j8-xcheck-wallet-info {
  list-style: none; margin: 0; padding: 10px 12px;
  display: grid; gap: 4px;
  border: 1px dashed var(--line);
  border-radius: 10px;
  color: var(--muted); font-size: 12px;
}
.j8-xcheck-wallet-info li { display: flex; align-items: baseline; justify-content: space-between; gap: 10px; }
.j8-xcheck-wallet-info li b { color: var(--text); font-weight: 700; }
.j8-xcheck-wallet-info li.is-warn { color: var(--amber); justify-content: flex-start; }

.j8-xcheck-btn { width: 100%; }

/* 窄屏下取消 sticky，避免跟内容打架 */
@media (max-width: 960px) {
  .j8-xcheck-summary { position: static; }
}
</style>
