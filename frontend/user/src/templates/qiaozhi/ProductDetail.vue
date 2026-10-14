<template>
  <div class="j8-container j8-pd">
    <!-- 返回 -->
    <button type="button" class="j8-pd-back" @click="goBack">
      <ChevronLeft />
      {{ t('qiaozhi.detail.back') }}
    </button>

    <!-- 骨架 -->
    <div v-if="loading" class="j8-pd-skeleton">
      <div class="j8-pd-sk-block"></div>
      <div class="j8-pd-sk-line is-lg"></div>
      <div class="j8-pd-sk-line"></div>
      <div class="j8-pd-sk-line is-sm"></div>
      <div class="j8-pd-sk-line is-btn"></div>
    </div>

    <template v-else-if="product">
      <div class="j8-pd-card">
        <!-- 左：媒体区（粘性居中，滚动长表单时图片保持可见） -->
        <div class="j8-pd-media">
          <div class="j8-pd-media-sticky">
            <div class="j8-pd-stage">
              <img v-if="currentImage" :src="currentImage" :alt="title" />
              <span v-else class="j8-pd-logo j8-brand-icon"><b>{{ title.slice(0, 1) }}</b></span>
            </div>
            <div v-if="gallery.length > 1" class="j8-pd-thumbs">
              <button
                v-for="(img, i) in gallery"
                :key="i"
                type="button"
                :class="{ 'is-active': i === currentIndex }"
                @click="currentIndex = i"
              >
                <img :src="img" :alt="`${i + 1}`" loading="lazy" />
              </button>
            </div>
          </div>
        </div>

        <!-- 右：购买区（内嵌下单） -->
        <div class="j8-pd-info">
          <h1>{{ title }}</h1>

          <div class="j8-pd-chips">
            <span class="is-accent"><Zap />{{ fulfillmentLabel }}</span>
            <span class="is-accent">{{ purchaseTypeLabel }}</span>
            <span :class="soldOut ? 'is-danger' : 'is-accent'">{{ stockLabel }}</span>
            <button type="button" class="j8-pd-share" @click="shareProduct"><Share2 />{{ t('qiaozhi.detail.share') }}</button>
          </div>

          <div class="j8-pd-price-row">
            <div class="j8-pd-price">
              <span class="j8-pd-price-sym">{{ unitPriceParts.symbol }}</span>
              <strong>{{ unitPriceParts.major }}{{ unitPriceParts.minor }}</strong>
            </div>
            <div v-if="promoTag" class="j8-pd-price-tag" :class="promoTag.tone">{{ promoTag.text }}</div>
          </div>

          <!-- 规格 -->
          <div v-if="plans.length" class="j8-pd-field">
            <label>{{ t('qiaozhi.detail.spec') }}</label>
            <div class="j8-pd-skus">
              <button
                v-for="plan in plans"
                :key="plan.id"
                type="button"
                :disabled="plan.disabled"
                :class="{ 'is-active': plan.selected, 'is-disabled': plan.disabled }"
                @click="selectPlan(plan.id, plan.disabled)"
              >
                <b>{{ plan.name }}</b>
                <small>{{ plan.hint }}</small>
              </button>
            </div>
            <p v-if="requiresSKUSelection" class="j8-pd-hint">{{ t('productDetail.skuRequired') }}</p>
          </div>

          <!-- 数量 -->
          <div class="j8-pd-field">
            <label>{{ t('productDetail.quantity') }}</label>
            <J8QtyStepper v-model="quantity" :min="quantityEffectiveMin" :max="quantityEffectiveLimit" />
          </div>

          <template v-if="pitchLine">
            <p class="j8-pd-desc" :class="{ 'j8-pd-desc--clamp': !descExpanded }">{{ pitchLine }}</p>
            <button type="button" class="j8-pd-desc-toggle" @click="descExpanded = !descExpanded">
              {{ descExpanded ? t('qiaozhi.detail.descCollapse') : t('qiaozhi.detail.descExpand') }}
            </button>
          </template>

          <!-- 内嵌下单 -->
          <div class="j8-pd-actions">
            <form class="j8-ic" novalidate @submit.prevent="submitOrder">
              <!-- 游客字段 -->
              <template v-if="!userAuthStore.isAuthenticated">
                <div class="j8-ic-row">
                  <div class="j8-ic-field">
                    <label for="j8-ic-email">{{ t('qiaozhi.detail.email') }}<b>*</b></label>
                    <input
                      id="j8-ic-email"
                      v-model="email"
                      type="email"
                      autocomplete="email"
                      :placeholder="t('qiaozhi.detail.emailPlaceholder')"
                      :aria-invalid="!!fieldErrors.email"
                      :aria-describedby="fieldErrors.email ? 'j8-ic-email-err' : undefined"
                    />
                    <p v-if="fieldErrors.email" id="j8-ic-email-err" class="j8-ic-error">{{ fieldErrors.email }}</p>
                  </div>
                  <div class="j8-ic-field">
                    <label for="j8-ic-password">{{ t('qiaozhi.detail.orderPassword') }}<b>*</b></label>
                    <input
                      id="j8-ic-password"
                      v-model="orderPassword"
                      type="password"
                      autocomplete="new-password"
                      :placeholder="t('qiaozhi.detail.orderPasswordPlaceholder')"
                      :aria-invalid="!!fieldErrors.password"
                      :aria-describedby="fieldErrors.password ? 'j8-ic-password-err' : undefined"
                    />
                    <p v-if="fieldErrors.password" id="j8-ic-password-err" class="j8-ic-error">{{ fieldErrors.password }}</p>
                  </div>
                </div>
                <div v-if="captchaProvider === 'image'" class="j8-ic-field">
                  <label>{{ t('qiaozhi.detail.captcha') }}<b>*</b></label>
                  <ImageCaptcha ref="imageCaptchaRef" v-model="captchaPayload" />
                </div>
                <div v-else-if="captchaProvider === 'turnstile'" class="j8-ic-field">
                  <label>{{ t('qiaozhi.detail.captcha') }}<b>*</b></label>
                  <TurnstileCaptcha ref="turnstileRef" v-model="turnstileToken" :site-key="turnstileSiteKey" />
                </div>
                <p class="j8-ic-hint">
                  {{ t('qiaozhi.detail.guestHint') }}
                  <RouterLink :to="`/auth/login?redirect=${encodeURIComponent(route.fullPath)}`">{{ t('qiaozhi.detail.loginToBuy') }}</RouterLink>
                </p>
              </template>
              <p v-else class="j8-ic-hint is-ok">{{ t('qiaozhi.detail.loggedInHint') }}</p>

              <!-- 支付方式 -->
              <div v-if="channels.length" class="j8-ic-pay">
                <h3><Wallet />{{ t('qiaozhi.detail.payTitle') }}</h3>
                <div class="j8-ic-channels">
                  <button
                    v-for="ch in channels"
                    :key="ch.id"
                    type="button"
                    class="j8-ic-channel"
                    :class="{ 'is-active': selectedChannelId === ch.id }"
                    @click="selectedChannelId = ch.id"
                  >
                    <i></i>
                    <img v-if="ch.icon" :src="ch.icon" alt="" />
                    <span>
                      <b>{{ ch.name }}</b>
                      <small v-if="ch.feeText">{{ ch.feeText }}</small>
                    </span>
                  </button>
                </div>
              </div>

              <div class="j8-ic-total">
                <span>{{ t('qiaozhi.detail.totalDue') }}</span>
                <strong>{{ totalParts.symbol }}{{ totalParts.major }}{{ totalParts.minor }}</strong>
              </div>

              <p v-if="formError" class="j8-ic-alert is-error" role="alert">{{ formError }}</p>
              <p v-else-if="cannotPurchaseReason" class="j8-ic-alert" role="status">{{ cannotPurchaseReason }}</p>

              <div class="j8-pd-buttons">
                <button
                  type="button"
                  class="j8-pd-cart-icon"
                  :disabled="!canPurchase"
                  :aria-label="t('productDetail.addToCart')"
                  :title="t('productDetail.addToCart')"
                  @click="onAddToCart"
                >
                  <ShoppingCart />
                </button>
                <J8Button
                  type="submit"
                  size="lg"
                  class="j8-arc-block"
                  :loading="submitting"
                  :disabled="!canPurchase"
                >
                  <CreditCard v-if="!submitting" />
                  <span>{{ submitting ? t('qiaozhi.detail.submitting') : t('qiaozhi.detail.submitAndPay') }}</span>
                </J8Button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- 商品详情 -->
      <div v-if="detailHtml" class="j8-pd-detail">
        <h2>{{ t('qiaozhi.detail.details') }}</h2>
        <div class="j8-pd-prose" v-html="detailHtml"></div>
      </div>

      <!-- 相关文章 -->
      <div v-if="relatedPosts.length" class="j8-pd-detail">
        <h2>{{ t('productDetail.relatedPosts') }}</h2>
        <div class="j8-pd-posts">
          <RouterLink v-for="rp in relatedPosts" :key="rp.id" :to="`/blog/${rp.slug}`">
            <b>{{ getLocalizedText(rp.title) }}</b>
            <small>{{ rp.summary ? getLocalizedText(rp.summary).slice(0, 60) : '' }}</small>
          </RouterLink>
        </div>
      </div>
    </template>

    <!-- 错误 -->
    <div v-else class="j8-empty">
      <PackageOpen />
      <b>{{ t('productDetail.notFound') }}</b>
      <button type="button" class="j8-btn-primary" @click="loadProduct">{{ t('errorBoundary.retry') }}</button>
    </div>

    <!-- 当页支付 Modal -->
    <PaymentModal
      v-model:open="showPaymentModal"
      :order="paymentOrder"
      :channels="cryptoChannels"
      :switch-busy="paymentSwitching"
      :switch-error="paymentSwitchError"
      :tx-busy="paymentTxBusy"
      :tx-feedback="paymentTxFeedback"
      @paid="handlePaymentPaid"
      @switch-channel="handleSwitchPaymentChannel"
      @submit-tx-hash="handleSubmitPaymentTxHash"
      @renew-payment="handleRenewPayment"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  ChevronLeft, CreditCard, PackageOpen, Share2, ShoppingCart, Wallet, Zap,
} from 'lucide-vue-next'
import { useAppStore } from '../../stores/app'
import { useUserAuthStore } from '../../stores/userAuth'
import { guestOrderAPI, paymentAPI, userOrderAPI } from '../../api'
import { useProductDetail } from '../../composables/useProductDetail'
import { processHtmlForDisplay } from '../../utils/content'
import { getFirstImageUrl, getImageUrl } from '../../utils/image'
import { normalizeSkuId } from '../../utils/sku'
import { saveGuestOrderAuth } from '../../utils/guestOrderAuth'
import ImageCaptcha from '../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../components/captcha/TurnstileCaptcha.vue'
import { useQiaozhiPrice } from './utils/price'
import PaymentModal, { type PaymentModalItem, type PaymentModalOrder } from './PaymentModal.vue'
import J8Button from './components/J8Button.vue'
import J8QtyStepper from './components/J8QtyStepper.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
const userAuthStore = useUserAuthStore()
const { split: splitPrice } = useQiaozhiPrice()

const {
  loading, product, relatedPosts,
  getLocalizedText, getFulfillmentTypeLabel, getPurchaseTypeLabel, getStockStatusLabel,
  hasPromotionPrice, getPromotionPriceAmount,
  hasSkuPromotionPrice, getSkuPromotionPriceAmount,
  activeSkus, selectedSku, selectedSkuId, quantity,
  isSkuPurchasable, skuDisplayText, skuStockText,
  requiresLogin, requiresSKUSelection, canPurchase, cannotPurchaseReason,
  quantityEffectiveMin, quantityEffectiveLimit,
  addToCart, loadProduct,
} = useProductDetail()

const title = computed(() => getLocalizedText(product.value?.title))
const soldOut = computed(() => product.value?.stock_status === 'out_of_stock' || product.value?.is_sold_out === true)
const fulfillmentLabel = computed(() => getFulfillmentTypeLabel(product.value?.fulfillment_type))
const purchaseTypeLabel = computed(() => getPurchaseTypeLabel(product.value?.purchase_type))
const stockLabel = computed(() => (soldOut.value ? t('products.stockStatus.outOfStock') : getStockStatusLabel(product.value)))

// ==================== 图库 ====================
const gallery = computed(() => {
  const list: string[] = []
  const raw = product.value?.images
  if (Array.isArray(raw)) {
    for (const item of raw) {
      const url = typeof item === 'string' ? item : item?.url
      if (url) list.push(getImageUrl(url))
    }
  }
  if (!list.length) {
    const primary = getFirstImageUrl(raw)
    if (primary) list.push(primary)
  }
  return list
})
const currentIndex = ref(0)
const currentImage = computed(() => gallery.value[currentIndex.value] || '')

// ==================== 套餐 ====================
const plans = computed(() =>
  (activeSkus.value || []).map((sku: any) => {
    const id = normalizeSkuId(sku.id)
    return {
      id,
      disabled: !isSkuPurchasable(sku),
      selected: normalizeSkuId(selectedSkuId.value) === id,
      name: skuDisplayText(sku),
      hint: skuStockText(sku),
    }
  }),
)
const selectPlan = (id: number, disabled: boolean) => {
  if (disabled) return
  selectedSkuId.value = id
}

// ==================== 价格 ====================
const unitPrice = computed(() => {
  const sku = selectedSku.value
  if (sku) return hasSkuPromotionPrice(sku) ? getSkuPromotionPriceAmount(sku) : sku.price_amount
  return hasPromotionPrice(product.value) ? getPromotionPriceAmount(product.value) : product.value?.price_amount
})
const unitPriceParts = computed(() => splitPrice(unitPrice.value))
const totalAmount = computed(() => {
  const base = Number(String(unitPrice.value ?? '0').replace(/[^\d.-]/g, '')) || 0
  return (base * quantity.value).toFixed(2)
})
const totalParts = computed(() => splitPrice(totalAmount.value))
const promoTag = computed(() => {
  const sku = selectedSku.value
  if ((sku && hasSkuPromotionPrice(sku)) || (!sku && hasPromotionPrice(product.value))) {
    return { tone: 'is-promo', text: t('products.promotionTag') }
  }
  return null
})

// 说明行：描述首句
const pitchLine = computed(() => {
  const raw = getLocalizedText(product.value?.description)
  if (!raw) return ''
  return raw.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').trim().slice(0, 80)
})
const descExpanded = ref(false)

// 详情 HTML
const detailHtml = computed(() => {
  const content = getLocalizedText(product.value?.content)
  const desc = getLocalizedText(product.value?.description)
  if (content) return processHtmlForDisplay(content)
  if (desc) return `<p>${desc.replace(/<[^>]+>/g, ' ')}</p>`
  return ''
})

// ==================== 内嵌下单 ====================
const email = ref('')
const orderPassword = ref('')
const captchaPayload = ref<{ captcha_id?: string; captcha_code?: string }>({})
const turnstileToken = ref('')
const imageCaptchaRef = ref<{ refresh: () => void } | null>(null)
const turnstileRef = ref<{ reset: () => void } | null>(null)
const formError = ref('')
const submitting = ref(false)
const selectedChannelId = ref<number | null>(null)

const captchaConfig = computed(() => appStore.config?.captcha as any)
const captchaProvider = computed(() => {
  const cfg = captchaConfig.value
  if (!cfg || String(cfg.provider || 'none') === 'none') return 'none'
  if (!cfg?.scenes?.guest_create_order) return 'none'
  return String(cfg.provider)
})
const turnstileSiteKey = computed(() => String(captchaConfig.value?.turnstile?.site_key || ''))

const channels = computed(() => {
  const list = appStore.config?.payment_channels
  if (!Array.isArray(list)) return []
  return list
    .filter((ch: any) => {
      const providerType = String(ch?.provider_type || '').toLowerCase()
      const channelType = String(ch?.channel_type || '').toLowerCase()
      if (providerType === 'epay') return ['wechat', 'wxpay', 'alipay', 'qqpay'].includes(channelType)
      return true
    })
    .map((ch: any) => {
      const feeBps = Number(ch?.fee_bps ?? ch?.feeBps ?? 0)
      return {
        id: Number(ch.id),
        name: String(ch.name || ch.channel_type || ch.provider_type || ''),
        providerType: String(ch?.provider_type || '').toLowerCase(),
        icon: ch.icon || ch.logo ? getImageUrl(String(ch.icon || ch.logo)) : '',
        feeText: feeBps > 0 ? t('qiaozhi.detail.fee', { rate: (feeBps / 100).toFixed(1) }) : '',
      }
    })
})

const buildItems = () => [{
  product_id: product.value?.id,
  sku_id: normalizeSkuId(selectedSkuId.value) || undefined,
  quantity: quantity.value,
  fulfillment_type: product.value?.fulfillment_type || undefined,
}]

const fieldErrors = ref<{ email?: string; password?: string }>({})

// 当页支付 Modal 状态
const showPaymentModal = ref(false)
const paymentOrder = ref<PaymentModalOrder | null>(null)

const maskEmail = (raw: string): string => {
  const mail = (raw || '').trim()
  if (!mail || !mail.includes('@')) return mail
  const [local, domain] = mail.split('@')
  if (!local) return mail
  if (local.length <= 2) return `${local.slice(0, 1)}***@${domain}`
  return `${local.slice(0, 1)}***${local.slice(-1)}@${domain}`
}

const resolveItemTitle = (raw: unknown): string => {
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object') return getLocalizedText(raw as Record<string, string>)
  return ''
}

const buildPaymentOrderPayload = (
  payData: any,
  orderNo: string,
  guestCredentials: { email: string; order_password: string } | null,
): PaymentModalOrder => {
  const order = payData?.order || {}
  const currency = String(order.currency || '')
  const rawItems = Array.isArray(order.items) ? order.items : []
  const items: PaymentModalItem[] = rawItems.length
    ? rawItems.map((it: any) => ({
      title: resolveItemTitle(it?.title) || resolveItemTitle(it?.product_title) || title.value || '',
      amount: String(it?.total_price ?? it?.unit_price ?? order.total_amount ?? '0'),
      quantity: Number(it?.quantity || 0) || undefined,
    }))
    : [{
      title: title.value || '',
      amount: String(order.total_amount || payData?.online_pay_amount || '0'),
      quantity: quantity.value,
    }]
  const expireRaw = payData?.expires_at || order.expires_at || null
  const createdRaw = order.created_at || null
  const resolvedEmail = userAuthStore.isAuthenticated
    ? (userAuthStore.user?.email || '')
    : (guestCredentials?.email || '')
  // create 响应的 channel_name/provider_type 可能为空（omitempty），按 channel_id 从渠道列表回填。
  const channelId = Number(payData?.channel_id || 0) || undefined
  const channelInfo = channelId
    ? channels.value.find((ch: any) => Number(ch.id) === channelId)
    : undefined
  return {
    sn: orderNo,
    amount: String(order.total_amount || payData?.online_pay_amount || '0'),
    currency,
    items,
    qrCodeContent: String(payData?.qr_code || ''),
    payUrl: String(payData?.pay_url || ''),
    h5PayUrl: String(payData?.pay_url || ''),
    expireAt: expireRaw,
    createdAt: createdRaw,
    maskedEmail: resolvedEmail ? maskEmail(resolvedEmail) : '',
    interactionMode: String(payData?.interaction_mode || ''),
    channelName: String(payData?.channel_name || channelInfo?.name || ''),
    providerType: String(payData?.provider_type || channelInfo?.providerType || ''),
    feeAmount: String(payData?.fee_amount || ''),
    payableAmount: String(payData?.payable_amount || ''),
    walletAddress: String(payData?.wallet_address || ''),
    chainAmount: String(payData?.chain_amount || ''),
    chain: String(payData?.chain || ''),
    tokenId: String(payData?.token_id || ''),
    channelId,
    paymentId: Number(payData?.payment_id || 0) || undefined,
    guestCredentials,
  }
}

const handlePaymentPaid = (order: PaymentModalOrder) => {
  showPaymentModal.value = false
  if (!order?.sn) return
  const sn = encodeURIComponent(order.sn)
  if (order.guestCredentials?.email) {
    router.push(`/guest/orders/${sn}`)
  } else {
    router.push(`/orders/${sn}`)
  }
}

// 过期后「重新发起支付」：过期订单会被后台置为 canceled，无法复用旧单，
// 走同一套下单链路重新建单（含校验/验证码/错误回显）。
const handleRenewPayment = async () => {
  await submitOrder()
  if (formError.value) {
    showPaymentModal.value = false
  }
}

// 弹窗内可切换的加密货币渠道（同店 USDT 多链）。
const cryptoChannels = computed(() =>
  channels.value
    .filter((ch) => ch.providerType === 'epusdt')
    .map((ch) => ({ id: ch.id, name: ch.name })),
)
const paymentSwitching = ref(false)
const paymentSwitchError = ref('')
const paymentTxBusy = ref(false)
const paymentTxFeedback = ref<{ ok: boolean; message?: string } | null>(null)

// 切换网络或币种：为同一订单在目标渠道重新发起支付（旧待支付记录由后端作废）。
const handleSwitchPaymentChannel = async (channelId: number) => {
  const current = paymentOrder.value
  if (!current?.sn || paymentSwitching.value) return
  paymentSwitching.value = true
  paymentSwitchError.value = ''
  try {
    let payData: any = null
    if (userAuthStore.isAuthenticated) {
      const res = await paymentAPI.create({ order_no: current.sn, channel_id: channelId })
      payData = res.data?.data
    } else {
      const cred = current.guestCredentials
      if (!cred) return
      const res = await guestOrderAPI.createPayment({
        order_no: current.sn,
        channel_id: channelId,
        email: cred.email,
        order_password: cred.order_password,
      })
      payData = res.data?.data
    }
    if (payData) {
      const next = buildPaymentOrderPayload(payData, current.sn, current.guestCredentials ?? null)
      // /payments 响应不含订单对象（Payment.ExpiredAt 建单时为空），切换渠道不改订单过期时间，沿用旧单 expireAt。
      next.expireAt = next.expireAt ?? current.expireAt ?? null
      paymentOrder.value = next
    }
  } catch (e: any) {
    paymentSwitchError.value = e?.message || t('qiaozhi.detail.errSubmit')
  } finally {
    paymentSwitching.value = false
  }
}

// 已转账但未到账：把链上交易哈希提交给网关人工核验，核验通过后订单自动完成。
const handleSubmitPaymentTxHash = async (txHash: string) => {
  const current = paymentOrder.value
  if (!current?.paymentId || paymentTxBusy.value) return
  paymentTxBusy.value = true
  paymentTxFeedback.value = null
  try {
    if (userAuthStore.isAuthenticated) {
      await paymentAPI.submitTxHash(current.paymentId, { block_transaction_id: txHash })
    } else {
      const cred = current.guestCredentials
      if (!cred) return
      await guestOrderAPI.submitTxHash(current.paymentId, {
        block_transaction_id: txHash,
        email: cred.email,
        order_password: cred.order_password,
      })
    }
    paymentTxFeedback.value = { ok: true }
  } catch (e: any) {
    paymentTxFeedback.value = { ok: false, message: e?.message || '' }
  } finally {
    paymentTxBusy.value = false
  }
}

watch(showPaymentModal, (opened) => {
  if (opened) {
    paymentSwitchError.value = ''
    paymentTxFeedback.value = null
  }
})

const validateForm = (): string => {
  fieldErrors.value = {}
  if (!userAuthStore.isAuthenticated) {
    const mail = email.value.trim()
    if (!mail || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(mail)) fieldErrors.value.email = t('qiaozhi.detail.errEmail')
    if (orderPassword.value.trim().length < 6) fieldErrors.value.password = t('qiaozhi.detail.errPassword')
    if (Object.keys(fieldErrors.value).length) return ''
    if (captchaProvider.value === 'image' && (!captchaPayload.value.captcha_id || !captchaPayload.value.captcha_code)) {
      return t('auth.common.captchaRequired')
    }
    if (captchaProvider.value === 'turnstile' && !turnstileToken.value) {
      return t('auth.common.captchaRequired')
    }
  }
  if (channels.value.length && selectedChannelId.value === null) {
    return t('qiaozhi.detail.errChannel')
  }
  return ''
}

const submitOrder = async () => {
  if (submitting.value) return
  if (!canPurchase.value) {
    formError.value = cannotPurchaseReason.value || ''
    return
  }
  if (requiresLogin.value) {
    router.push(`/auth/login?redirect=${encodeURIComponent(route.fullPath)}`)
    return
  }
  const err = validateForm()
  if (Object.keys(fieldErrors.value).length) {
    return
  }
  if (err) {
    formError.value = err
    return
  }
  formError.value = ''
  submitting.value = true
  try {
    const items = buildItems()
    let payData: any = null
    const guestEmailInput = email.value.trim()
    const guestPasswordInput = orderPassword.value
    if (userAuthStore.isAuthenticated) {
      const res = await userOrderAPI.createAndPay({ items, channel_id: selectedChannelId.value || undefined })
      payData = res.data?.data
    } else {
      const res = await guestOrderAPI.createAndPay({
        items,
        channel_id: selectedChannelId.value || undefined,
        email: guestEmailInput,
        order_password: guestPasswordInput,
        captcha_payload: captchaProvider.value === 'image'
          ? captchaPayload.value
          : (turnstileToken.value ? { turnstile_token: turnstileToken.value } : {}),
      })
      saveGuestOrderAuth({ email: guestEmailInput, order_password: guestPasswordInput })
      payData = res.data?.data
    }
    const orderNo = payData?.order_no || ''
    if (!orderNo) throw new Error(t('qiaozhi.detail.errSubmit'))

    // 后端 /orders/create-and-pay 已直接返回支付字段:qr_code / pay_url /
    // expires_at / interaction_mode。拿到完整响应就当页弹 Modal,不再跳 /pay。
    // 没有支付字段(例如 channel_id 为 0、全额钱包抵扣、通道返错)时 fallback 到
    // 旧的 /pay 独立页,保留原流程不丢失。
    if (payData && (payData.qr_code || payData.pay_url) && !payData.order_paid) {
      paymentOrder.value = buildPaymentOrderPayload(
        payData,
        orderNo,
        userAuthStore.isAuthenticated ? null : { email: guestEmailInput, order_password: guestPasswordInput },
      )
      showPaymentModal.value = true
    } else {
      const query = userAuthStore.isAuthenticated
        ? `order_no=${encodeURIComponent(orderNo)}`
        : `guest=1&order_no=${encodeURIComponent(orderNo)}`
      router.push(`/pay?${query}`)
    }
  } catch (e: any) {
    formError.value = e?.message || t('qiaozhi.detail.errSubmit')
    imageCaptchaRef.value?.refresh()
    turnstileRef.value?.reset()
  } finally {
    submitting.value = false
  }
}

const onAddToCart = () => {
  if (!canPurchase.value) return
  addToCart()
}

const shareProduct = async () => {
  try {
    await navigator.clipboard.writeText(window.location.href)
  } catch {
    /* 忽略 */
  }
}

const goBack = () => {
  if (window.history.length > 1) router.back()
  else router.push('/')
}

onMounted(() => {
  if (!appStore.config) void appStore.loadConfig()
})
</script>
