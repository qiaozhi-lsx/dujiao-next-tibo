<template>
  <div ref="rootEl" class="qz-container qz-checkout">
    <!-- 页头：eyebrow + h1 + 说明 + 返回 -->
    <header class="qz-page-head">
      <div class="qz-page-head-copy">
        <span class="qz-eyebrow">{{ t('qiaozhi.checkout.eyebrow') }}</span>
        <h1>{{ t('checkout.title') }}</h1>
        <p>{{ t('qiaozhi.checkout.subtitle') }}</p>
      </div>
      <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm" @click="goBack">
        <ChevronLeft />
        {{ backLabel }}
      </button>
    </header>

    <!-- 空购物车 -->
    <div v-if="cartItems.length === 0" class="qz-empty">
      <PackageOpen />
      <b>{{ t('checkout.empty') }}</b>
      <span>{{ t('qiaozhi.checkout.emptyHint') }}</span>
      <RouterLink to="/products" class="qz-btn qz-btn--primary">{{ t('checkout.emptyAction') }}</RouterLink>
    </div>

    <!-- 结算主体：左商品 + 表单，右合计 -->
    <div v-else class="qz-checkout-layout">
      <div class="qz-checkout-main">
        <!-- 订单明细 -->
        <section class="qz-panel qz-checkout-card qz-reveal is-pending" :data-reveal="1">
          <header class="qz-checkout-card-head">
            <span class="qz-eyebrow">{{ t('qiaozhi.checkout.sectionItems') }}</span>
            <h2>{{ t('checkout.itemsTitle') }}</h2>
          </header>
          <ul class="qz-checkout-items">
            <li
              v-for="item in cartItems"
              :key="cartItemKey(item)"
              class="qz-card qz-checkout-item"
              :class="{ 'is-warn': itemStockExceeded(item) }"
            >
              <div class="qz-checkout-thumb">
                <img
                  v-if="checkoutItemImage(item)"
                  :src="checkoutItemImage(item)"
                  :alt="getLocalizedText(item.title)"
                  loading="lazy"
                />
                <span v-else class="qz-checkout-initial" aria-hidden="true">{{ (getLocalizedText(item.title) || '·').slice(0, 1) }}</span>
              </div>
              <div class="qz-checkout-item-body">
                <RouterLink
                  class="qz-checkout-item-title"
                  :to="`/products/${item.slug}`"
                >
                  {{ getLocalizedText(item.title) }}
                </RouterLink>
                <div class="qz-checkout-item-chips">
                  <span class="qz-chip">{{ t('checkout.quantityLabel') }} × {{ item.quantity }}</span>
                  <span v-if="itemSkuDisplay(item)" class="qz-chip">{{ itemSkuDisplay(item) }}</span>
                </div>
                <p
                  v-if="itemStockHint(item)"
                  class="qz-checkout-item-stock"
                  :class="{ 'is-warn': itemStockExceeded(item) }"
                >
                  {{ itemStockHint(item) }}
                </p>
              </div>
              <div class="qz-checkout-item-price">
                <span
                  class="qz-checkout-price"
                  :class="{ 'is-discount': checkoutItemHasPriceDiscount(item) }"
                >
                  <strong>
                    {{ checkoutItemPriceParts(item).integer }}<small>{{ checkoutItemPriceParts(item).decimal }}</small>
                  </strong>
                  <span class="qz-checkout-price-cur">{{ checkoutItemCurrency }}</span>
                </span>
                <del
                  v-if="checkoutItemHasPriceDiscount(item)"
                  class="qz-checkout-price-old"
                >
                  {{ checkoutItemOriginalPriceParts(item).integer }}{{ checkoutItemOriginalPriceParts(item).decimal }} {{ checkoutItemCurrency }}
                </del>
              </div>
            </li>
          </ul>
        </section>

        <!-- 交付资料 / 游客 / 优惠码 表单 -->
        <section class="qz-panel qz-checkout-card qz-reveal is-pending" :data-reveal="1">
          <header class="qz-checkout-card-head">
            <span class="qz-eyebrow">{{ t('qiaozhi.checkout.sectionInfo') }}</span>
            <h2>{{ t('qiaozhi.checkout.infoTitle') }}</h2>
          </header>

          <form class="qz-checkout-form" novalidate @submit.prevent="handleSubmit">
            <!-- 交付信息 -->
            <template v-if="manualFormProducts.length">
              <div
                v-for="product in manualFormProducts"
                :key="product.itemKey"
                class="qz-checkout-manual"
              >
                <h3>
                  {{ manualItemTitle(product) }}
                </h3>
                <p class="qz-field-hint">{{ t('checkout.manualFormTip') }}</p>
                <div class="qz-checkout-manual-grid">
                  <div
                    v-for="field in product.fields"
                    :key="`${product.itemKey}-${field.key}`"
                    class="qz-field"
                  >
                    <!-- label 自带 id：radio / checkbox 组没有可被 for 指向的单个控件，组容器用 aria-labelledby 指回这里 -->
                    <label class="qz-field-label" :id="`${manualFieldId(product.itemKey, field.key)}-label`" :for="manualFieldId(product.itemKey, field.key)">
                      {{ getManualFieldLabel(field) }}
                      <b v-if="field.required">*</b>
                    </label>

                    <textarea
                      v-if="field.type === 'textarea'"
                      :id="manualFieldId(product.itemKey, field.key)"
                      class="qz-field-control"
                      :value="getManualFieldValue(product.itemKey, field.key)"
                      rows="3"
                      :placeholder="getManualFieldPlaceholder(field)"
                      :aria-invalid="manualFieldHasError(product.itemKey, field.key)"
                      :aria-describedby="manualFieldHasError(product.itemKey, field.key) ? manualFieldErrId(product.itemKey, field.key) : undefined"
                      @input="setManualFieldValue(product.itemKey, field.key, ($event.target as HTMLTextAreaElement).value)"
                    ></textarea>

                    <QzSelect
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
                      class="qz-checkout-options"
                      role="radiogroup"
                      :aria-labelledby="`${manualFieldId(product.itemKey, field.key)}-label`"
                    >
                      <label v-for="option in field.options" :key="option">
                        <input
                          type="radio"
                          :name="`qz-checkout-radio-${product.itemKey}-${field.key}`"
                          :value="option"
                          :checked="getManualFieldValue(product.itemKey, field.key) === option"
                          @change="setManualFieldValue(product.itemKey, field.key, option)"
                        />
                        <span>{{ option }}</span>
                      </label>
                    </div>

                    <div
                      v-else-if="field.type === 'checkbox'"
                      class="qz-checkout-options"
                      role="group"
                      :aria-labelledby="`${manualFieldId(product.itemKey, field.key)}-label`"
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
                      class="qz-field-control"
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
                      class="qz-field-error"
                    >
                      {{ manualFieldError(product.itemKey, field.key) }}
                    </p>
                  </div>
                </div>
              </div>
            </template>

            <!-- 游客 -->
            <template v-if="!userAuthStore.isAuthenticated">
              <div class="qz-checkout-mode">
                <h3>{{ t('checkout.modeTitle') }}</h3>
                <div class="qz-checkout-mode-tabs" role="tablist">
                  <button
                    type="button"
                    role="tab"
                    class="qz-pill"
                    :aria-selected="checkoutMode === 'guest'"
                    :class="{ 'is-active': checkoutMode === 'guest' }"
                    @click="checkoutMode = 'guest'"
                  >
                    {{ t('checkout.guestPurchase') }}
                  </button>
                  <RouterLink
                    to="/auth/login"
                    role="tab"
                    class="qz-pill"
                  >
                    {{ t('checkout.memberPurchase') }}
                  </RouterLink>
                </div>
              </div>

              <template v-if="checkoutMode === 'guest'">
                <div class="qz-checkout-form-row">
                  <div class="qz-field">
                    <label class="qz-field-label" for="qz-checkout-email">
                      {{ t('qiaozhi.detail.email') }}<b>*</b>
                    </label>
                    <input
                      id="qz-checkout-email"
                      v-model="guestEmail"
                      class="qz-field-control"
                      type="email"
                      autocomplete="email"
                      :placeholder="t('checkout.guestEmailPlaceholder')"
                      :aria-invalid="submitAttempted && !!guestEmail && !guestEmailValid"
                      :aria-describedby="submitAttempted && !!guestEmail && !guestEmailValid ? 'qz-checkout-email-err' : undefined"
                    />
                    <p
                      v-if="submitAttempted && !!guestEmail && !guestEmailValid"
                      id="qz-checkout-email-err"
                      class="qz-field-error"
                    >
                      {{ t('error.email_invalid') }}
                    </p>
                  </div>

                  <div class="qz-field">
                    <label class="qz-field-label" for="qz-checkout-password">
                      {{ t('qiaozhi.detail.orderPassword') }}<b>*</b>
                    </label>
                    <input
                      id="qz-checkout-password"
                      v-model="guestPassword"
                      class="qz-field-control"
                      type="password"
                      autocomplete="new-password"
                      :placeholder="t('checkout.guestPasswordPlaceholder')"
                    />
                  </div>
                </div>

                <div v-if="guestCaptchaEnabled" class="qz-field" role="group" aria-labelledby="qz-checkout-captcha-label">
                  <span id="qz-checkout-captcha-label" class="qz-field-label">{{ t('auth.common.captchaLabel') }}<b>*</b></span>
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

                <div class="qz-checkout-notice">
                  <b>{{ t('checkout.guestInstructions.title') }}</b>
                  <ul>
                    <li>{{ t('checkout.guestInstructions.email') }}</li>
                    <li>{{ t('checkout.guestInstructions.password') }}</li>
                  </ul>
                </div>
              </template>
            </template>
            <p v-else class="qz-field-hint is-ok">{{ t('qiaozhi.detail.loggedInHint') }}</p>

            <!-- 优惠码 -->
            <div v-if="!isResellerTenant" class="qz-field">
              <label class="qz-field-label" for="qz-checkout-coupon">{{ t('checkout.couponTitle') }}</label>
              <input
                id="qz-checkout-coupon"
                v-model="couponCode"
                class="qz-field-control"
                type="text"
                autocomplete="off"
                :placeholder="t('checkout.couponPlaceholder')"
              />
            </div>
          </form>
        </section>
      </div>

      <!-- 右：合计 + 支付（桌面 sticky 跟随） -->
      <aside class="qz-checkout-aside qz-reveal is-pending" :data-reveal="1">
        <div class="qz-panel qz-checkout-summary">
          <header class="qz-checkout-card-head">
            <span class="qz-eyebrow">{{ t('qiaozhi.checkout.sectionSummary') }}</span>
            <h2>{{ t('checkout.submitTitle') }}</h2>
          </header>

          <p class="qz-field-hint">{{ t('checkout.submitHint') }}</p>

          <dl class="qz-checkout-rows">
            <div class="qz-checkout-row">
              <dt>{{ t('cart.itemsCount') }}</dt>
              <dd>{{ totalItems }}</dd>
            </div>
            <div class="qz-checkout-row">
              <dt>{{ t('checkout.previewOriginal') }}</dt>
              <dd>{{ formatPrice(previewOriginal, previewCurrency) }}</dd>
            </div>
            <template v-if="!isResellerTenant">
              <div class="qz-checkout-row">
                <dt>{{ t('checkout.previewCoupon') }}</dt>
                <dd :class="{ 'is-discount': hasPositiveAmount(previewCoupon) }">
                  {{ formatDiscountPrice(previewCoupon, previewCurrency) }}
                </dd>
              </div>
              <div class="qz-checkout-row">
                <dt>{{ t('checkout.previewPromotion') }}</dt>
                <dd :class="{ 'is-discount': hasPositiveAmount(previewPromotion) }">
                  {{ formatDiscountPrice(previewPromotion, previewCurrency) }}
                </dd>
              </div>
              <div class="qz-checkout-row">
                <dt>{{ t('checkout.previewWholesale') }}</dt>
                <dd :class="{ 'is-saving': hasPositiveAmount(previewWholesale) }">
                  {{ formatDiscountPrice(previewWholesale, previewCurrency) }}
                </dd>
              </div>
            </template>
            <div v-if="hasPositiveAmount(previewMemberDiscount)" class="qz-checkout-row">
              <dt>{{ t('checkout.previewMemberDiscount') }}</dt>
              <dd class="is-saving">-{{ formatPrice(previewMemberDiscount, previewCurrency) }}</dd>
            </div>
          </dl>

          <div class="qz-checkout-total">
            <span>{{ t('checkout.previewTotal') }}</span>
            <strong>{{ formatPrice(previewTotal, previewCurrency) }}</strong>
          </div>

          <p v-if="previewLoading || couponRefreshing" class="qz-field-hint">
            {{ previewStatusText }}
          </p>

          <!-- 支付方式 -->
          <div class="qz-checkout-pay">
            <h3 id="qz-checkout-pay-label">
              <Wallet />
              {{ t('checkout.paymentMethod') }}
            </h3>

            <!-- 钱包余额 -->
            <label v-if="showBalanceOption" class="qz-checkout-balance" :class="{ 'is-active': useBalance }">
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
              class="qz-checkout-wallet-info"
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
                class="qz-checkout-channels"
                role="group"
                aria-labelledby="qz-checkout-pay-label"
              >
                <button
                  v-for="channel in paymentChannels"
                  :key="channel.id"
                  type="button"
                  class="qz-checkout-channel"
                  :class="{ 'is-active': selectedChannelId === channel.id && !isChannelDisabledForAmount(channel) }"
                  :disabled="isChannelDisabledForAmount(channel)"
                  :title="isChannelDisabledForAmount(channel) ? channelAmountLimitHint(channel) : ''"
                  @click="handleSelectChannel(channel)"
                >
                  <i aria-hidden="true"></i>
                  <img
                    v-if="channel.icon"
                    :src="getImageUrl(channel.icon)"
                    alt=""
                    loading="lazy"
                  />
                  <span>
                    <b>{{ channelDisplayName(channel.name) }}</b>
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
                class="qz-field-hint"
              >
                {{ t('checkout.noPaymentChannels') }}
              </p>
            </template>

            <p
              v-if="!requiresOnlineChannel"
              class="qz-field-hint is-ok"
            >
              {{ t('checkout.walletCoversAll') }}
            </p>
          </div>

          <!-- 提示 / 错误 -->
          <p
            v-if="checkoutAlert"
            class="qz-alert"
            :class="checkoutAlert.level === 'error' ? 'is-danger' : 'is-warning'"
            role="alert"
          >
            {{ checkoutAlert.message }}
          </p>

          <!-- 提交 -->
          <QzButton
            type="button"
            size="lg"
            class="qz-btn--block"
            :disabled="!canSubmit"
            :loading="submitting"
            @click="handleSubmit"
          >
            <CreditCard v-if="!submitting" />
            {{ submitting ? t('checkout.submitting') : t('checkout.submitButton') }}
          </QzButton>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, CreditCard, PackageOpen, Wallet } from 'lucide-vue-next'
import ImageCaptcha from '../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../components/captcha/TurnstileCaptcha.vue'
import { useCheckout } from '../../composables/useCheckout'
import { useLocalized } from '../../composables/useProduct'
import { channelDisplayName } from '../../utils/cryptoChain'
import { setupReveal } from './utils/reveal'
import QzButton from './components/QzButton.vue'
import QzSelect from './components/QzSelect.vue'

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

const manualFieldId = (itemKey: string, fieldKey: string) => `qz-mf-${itemKey}-${fieldKey}`
const manualFieldErrId = (itemKey: string, fieldKey: string) => `qz-mf-err-${itemKey}-${fieldKey}`

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
/* 结算页：控件全部来自 styles/base.css（.qz-btn / .qz-pill / .qz-chip / .qz-field* / .qz-card / .qz-panel / .qz-empty），
   这里只放页面布局与结算页专属元素，只读 --qz-* token（无色值字面量、无 px 字号）。 */
.qz-checkout {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-5);
}

/* 页头用 base.css 的 .qz-page-head，提示条用 .qz-alert */

/* ---------- 双栏：左商品 + 表单，右合计 ---------- */
.qz-checkout-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: var(--qz-space-4);
  align-items: start;
}

.qz-checkout-main {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  min-width: 0;
}

.qz-checkout-aside {
  min-width: 0;
}

/* ---------- 面板 + 段头（.qz-panel.qz-checkout-* 写成 (0,3,0)，稳压 base 的 .qz-panel 默认值） ---------- */
.qz-panel.qz-checkout-card,
.qz-panel.qz-checkout-summary {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
}

/* 合计卡：桌面 sticky 跟随滚动 */
.qz-panel.qz-checkout-summary {
  position: sticky;
  top: calc(var(--qz-header-h) + var(--qz-space-4));
}

.qz-checkout-card-head {
  display: grid;
  gap: var(--qz-space-1);
}

.qz-checkout-card-head h2 {
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: var(--qz-leading-body);
}

/* ---------- 商品列表：图 | 标题 + chips + 库存 | 价格 ---------- */
.qz-checkout-items {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-2);
  margin: 0;
  padding: 0;
  list-style: none;
}

.qz-card.qz-checkout-item {
  display: grid;
  grid-template-columns: 64px minmax(0, 1fr) auto;
  gap: var(--qz-space-3);
  align-items: center;
  padding: var(--qz-space-3);
}

/* 超出库存 / 限购的行：警示色描边 + 淡警示底 */
.qz-card.qz-checkout-item.is-warn {
  border-color: var(--qz-warning);
  background: color-mix(in oklab, var(--qz-warning) 8%, var(--qz-surface));
}

.qz-checkout-thumb {
  display: grid;
  place-items: center;
  width: 64px;
  height: 64px;
  overflow: hidden;
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-sm);
  background: var(--qz-surface-muted);
}

.qz-checkout-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.qz-checkout-initial {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xl);
  font-weight: 800;
  letter-spacing: var(--qz-tracking-display);
}

.qz-checkout-item-body {
  display: grid;
  gap: var(--qz-space-1);
  min-width: 0;
}

.qz-checkout-item-title {
  display: block;
  min-width: 0;
  overflow: hidden;
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  line-height: var(--qz-leading-body);
  text-decoration: none;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.qz-checkout-item-title:hover {
  text-decoration: underline;
  text-underline-offset: 3px;
}

.qz-checkout-item-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--qz-space-1) var(--qz-space-2);
  min-width: 0;
}

.qz-checkout-item-stock {
  margin: 0;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-checkout-item-stock.is-warn {
  color: var(--qz-warning);
  font-weight: 500;
}

/* 价格：整数大字 + 小数小字，右对齐；折后价用危险色，原价划线 */
.qz-checkout-item-price {
  display: grid;
  justify-items: end;
  gap: 2px;
  text-align: right;
}

.qz-checkout-price {
  display: inline-flex;
  align-items: baseline;
  gap: var(--qz-space-1);
  color: var(--qz-text);
  font-variant-numeric: tabular-nums;
}

.qz-checkout-price strong {
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: 1;
  letter-spacing: var(--qz-tracking-display);
}

.qz-checkout-price strong small {
  font-size: var(--qz-text-xs);
  font-weight: 700;
}

.qz-checkout-price.is-discount {
  color: var(--qz-danger);
}

.qz-checkout-price-cur {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
  font-weight: 600;
}

.qz-checkout-price-old {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
  font-variant-numeric: tabular-nums;
}

/* ---------- 表单 ---------- */
.qz-checkout-form {
  display: grid;
  gap: var(--qz-space-4);
}

.qz-checkout-form-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--qz-space-3);
}

/* 已登录 / 余额已覆盖：成功色提示 */
.qz-checkout .qz-field-hint.is-ok {
  color: var(--qz-success);
  font-weight: 500;
}

/* 交付资料分组：弱化底的内嵌区块，字段两列 */
.qz-checkout-manual {
  display: grid;
  gap: var(--qz-space-3);
  padding: var(--qz-space-4);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
}

.qz-checkout-manual h3 {
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 700;
  line-height: var(--qz-leading-body);
}

.qz-checkout-manual-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--qz-space-3);
}

/* 单选 / 多选组：与输入框等高，选项横向流式排布 */
.qz-checkout-options {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--qz-space-2) var(--qz-space-4);
  min-height: var(--qz-control-md);
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
}

.qz-checkout-options label {
  display: inline-flex;
  align-items: center;
  gap: var(--qz-space-2);
  cursor: pointer;
}

.qz-checkout-options input {
  width: 16px;
  height: 16px;
  margin: 0;
  accent-color: var(--qz-accent);
}

/* 购买方式：分段胶囊（.qz-pill 轨道） */
.qz-checkout-mode {
  display: grid;
  gap: var(--qz-space-2);
}

.qz-checkout-mode h3 {
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  line-height: var(--qz-leading-body);
}

.qz-checkout-mode-tabs {
  display: inline-flex;
  flex-wrap: wrap;
  gap: var(--qz-space-1);
  justify-self: start;
  padding: var(--qz-space-1);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-pill);
  background: var(--qz-surface-muted);
}

/* 游客说明 */
.qz-checkout-notice {
  padding: var(--qz-space-3) var(--qz-space-4);
  border-radius: var(--qz-radius-control);
  background: var(--qz-accent-subtle);
  color: var(--qz-text);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-checkout-notice b {
  display: block;
  margin-bottom: var(--qz-space-1);
  font-weight: 600;
}

.qz-checkout-notice ul {
  display: grid;
  gap: var(--qz-space-1);
  margin: 0;
  padding-left: var(--qz-space-4);
  color: var(--qz-text-secondary);
  list-style: disc;
}

/* ---------- 合计 ---------- */
.qz-checkout-rows {
  display: grid;
  gap: var(--qz-space-2);
  margin: 0;
}

.qz-checkout-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--qz-space-3);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
}

.qz-checkout-row dt,
.qz-checkout-row dd {
  margin: 0;
}

.qz-checkout-row dd {
  color: var(--qz-text);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.qz-checkout-row dd.is-discount {
  color: var(--qz-danger);
}

.qz-checkout-row dd.is-saving {
  color: var(--qz-success);
}

.qz-checkout-total {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--qz-space-3);
  padding-top: var(--qz-space-4);
  border-top: 1px solid var(--qz-border);
}

.qz-checkout-total span {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  font-weight: 600;
}

.qz-checkout-total strong {
  color: var(--qz-text);
  font-size: var(--qz-text-2xl);
  font-weight: 700;
  line-height: 1;
  letter-spacing: var(--qz-tracking-display);
  font-variant-numeric: tabular-nums;
}

/* ---------- 支付方式：弱化面板内的卡片式单选行；选中 = 1px 强调色描边 + 弱强调底 ---------- */
.qz-checkout-pay {
  display: grid;
  gap: var(--qz-space-2);
  padding: var(--qz-space-3);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
}

.qz-checkout-pay h3 {
  display: flex;
  align-items: center;
  gap: var(--qz-space-1);
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  line-height: var(--qz-leading-body);
}

.qz-checkout-pay h3 svg {
  width: 16px;
  height: 16px;
  color: var(--qz-text-muted);
}

.qz-checkout-balance,
.qz-checkout-channel {
  display: flex;
  align-items: center;
  gap: var(--qz-space-2);
  min-height: var(--qz-control-md);
  padding: var(--qz-space-2) var(--qz-space-3);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface);
  color: var(--qz-text);
  box-shadow: var(--qz-shadow-resting);
  font: inherit;
  font-size: var(--qz-text-sm);
  text-align: left;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  transition:
    border-color var(--qz-duration-fast) var(--qz-ease-standard),
    background-color var(--qz-duration-fast) var(--qz-ease-standard);
}

@media (hover: hover) and (pointer: fine) {
  .qz-checkout-balance:hover:where(:not(.is-active)),
  .qz-checkout-channel:hover:where(:not(.is-active, :disabled)) {
    border-color: var(--qz-border-strong);
  }
}

.qz-checkout-balance.is-active,
.qz-checkout-channel.is-active {
  border-color: var(--qz-accent);
  background: var(--qz-accent-subtle);
}

.qz-checkout-channel:disabled {
  opacity: 0.5;
  box-shadow: none;
  cursor: not-allowed;
}

.qz-checkout-balance > svg {
  flex: none;
  width: 18px;
  height: 18px;
  color: var(--qz-text-secondary);
}

.qz-checkout-balance > span,
.qz-checkout-channel > span {
  display: grid;
  min-width: 0;
}

.qz-checkout-balance b,
.qz-checkout-channel b {
  overflow: hidden;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.qz-checkout-balance small,
.qz-checkout-channel small {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
}

.qz-checkout-balance em {
  margin-left: auto;
  color: var(--qz-warning);
  font-size: var(--qz-text-2xs);
  font-style: normal;
  font-weight: 600;
  text-align: right;
}

.qz-checkout-balance input {
  flex: none;
  width: 16px;
  height: 16px;
  margin: 0 0 0 auto;
  accent-color: var(--qz-accent);
}

.qz-checkout-channels {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: var(--qz-space-2);
}

/* 单选圆点：选中后 5px 粗描边收成实心点 */
.qz-checkout-channel > i {
  flex: none;
  width: 16px;
  height: 16px;
  border: 2px solid var(--qz-border-strong);
  border-radius: 50%;
  background: var(--qz-surface);
  transition: border var(--qz-duration-fast) var(--qz-ease-standard);
}

.qz-checkout-channel.is-active > i {
  border: 5px solid var(--qz-accent);
}

.qz-checkout-channel img {
  flex: none;
  width: 24px;
  height: 24px;
  border-radius: var(--qz-radius-xs);
  object-fit: contain;
}

/* 钱包抵扣摘要 */
.qz-checkout-wallet-info {
  display: grid;
  gap: var(--qz-space-1);
  margin: 0;
  padding: var(--qz-space-2) var(--qz-space-3);
  border: 1px dashed var(--qz-border-strong);
  border-radius: var(--qz-radius-control);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  list-style: none;
}

.qz-checkout-wallet-info li {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--qz-space-2);
}

.qz-checkout-wallet-info li b {
  color: var(--qz-text);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.qz-checkout-wallet-info li.is-warn {
  justify-content: flex-start;
  color: var(--qz-warning);
  font-weight: 500;
}

/* ---------- 响应式 ---------- */
@media (max-width: 1023px) {
  .qz-checkout-layout {
    grid-template-columns: minmax(0, 1fr);
    gap: var(--qz-space-3);
  }

  .qz-checkout-main {
    gap: var(--qz-space-3);
  }

  .qz-panel.qz-checkout-summary {
    position: static;
  }
}

@media (max-width: 760px) {
  .qz-panel.qz-checkout-card,
  .qz-panel.qz-checkout-summary {
    padding: var(--qz-space-4);
  }

  .qz-checkout-form-row,
  .qz-checkout-manual-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  /* 手机：图缩到 56px，价格换到标题下一行左对齐 */
  .qz-card.qz-checkout-item {
    grid-template-columns: 56px minmax(0, 1fr);
  }

  .qz-checkout-thumb {
    width: 56px;
    height: 56px;
  }

  .qz-checkout-item-price {
    grid-column: 2;
    justify-items: start;
    text-align: left;
  }
}

@media (prefers-reduced-motion: reduce) {
  .qz-checkout-balance,
  .qz-checkout-channel,
  .qz-checkout-channel > i {
    transition: none;
  }
}
</style>
