<template>
  <div class="qz-container qz-pd-page">
    <!-- 返回：secondary sm 胶囊，只在手机端吸顶（桌面静态，不再压住下方标题） -->
    <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm qz-pd-return" @click="goBack">
      <ChevronLeft />
      {{ t('qiaozhi.detail.back') }}
    </button>

    <!-- 骨架 -->
    <div v-if="loading" class="qz-pd-skeleton" aria-hidden="true">
      <div class="qz-skeleton qz-pd-sk-media"></div>
      <div class="qz-skeleton is-lg"></div>
      <div class="qz-skeleton is-line"></div>
      <div class="qz-skeleton is-sm"></div>
      <div class="qz-skeleton qz-pd-sk-cta"></div>
    </div>

    <template v-else-if="product">
      <div class="qz-pd-card">
        <!-- 左：媒体区（粘性居中，滚动长表单时图片保持可见） -->
        <div class="qz-pd-media">
          <div class="qz-pd-media-sticky">
            <div class="qz-pd-stage">
              <img v-if="currentImage" :src="currentImage" :alt="title" />
              <span v-else class="qz-pd-logo" aria-hidden="true"><b>{{ title.slice(0, 1) }}</b></span>
            </div>
            <div v-if="gallery.length > 1" class="qz-pd-thumbs">
              <button
                v-for="(img, i) in gallery"
                :key="i"
                type="button"
                :class="{ 'is-active': i === currentIndex }"
                :aria-pressed="i === currentIndex"
                @click="currentIndex = i"
              >
                <img :src="img" :alt="`${i + 1}`" loading="lazy" />
              </button>
            </div>
          </div>
        </div>

        <!-- 右：购买区（内嵌下单） -->
        <div class="qz-pd-info">
          <div class="qz-pd-title-row">
            <h1>{{ title }}</h1>
            <button type="button" class="qz-btn qz-btn--ghost qz-btn--sm" @click="shareProduct"><Share2 />{{ t('qiaozhi.detail.share') }}</button>
          </div>

          <div class="qz-pd-chip-row">
            <span class="qz-chip"><Zap />{{ fulfillmentLabel }}</span>
            <span class="qz-chip">{{ purchaseTypeLabel }}</span>
            <span class="qz-chip" :class="{ 'is-danger': soldOut }">{{ stockLabel }}</span>
          </div>

          <div class="qz-pd-price-row">
            <div class="qz-pd-price">
              <span class="qz-pd-price-sym">{{ unitPriceParts.symbol }}</span>
              <strong>{{ unitPriceParts.major }}{{ unitPriceParts.minor }}</strong>
            </div>
            <span v-if="promoTag" class="qz-badge" :class="promoTag.tone">{{ promoTag.text }}</span>
          </div>

          <!-- 规格 -->
          <div v-if="plans.length" class="qz-field">
            <span id="qz-pd-sku-label" class="qz-field-label">{{ t('qiaozhi.detail.spec') }}</span>
            <div class="qz-pd-sku-list" role="radiogroup" aria-labelledby="qz-pd-sku-label" @keydown="onRadioArrow">
              <button
                v-for="plan in plans"
                :key="plan.id"
                type="button"
                role="radio"
                class="qz-pill"
                :class="{ 'is-active': plan.selected }"
                :disabled="plan.disabled"
                :aria-checked="plan.selected"
                @click="selectPlan(plan.id, plan.disabled)"
              >
                <b>{{ plan.name }}</b>
                <small>{{ plan.hint }}</small>
              </button>
            </div>
            <p v-if="requiresSKUSelection" class="qz-field-hint">{{ t('productDetail.skuRequired') }}</p>
          </div>

          <!-- 数量 -->
          <div class="qz-field">
            <span class="qz-field-label">{{ t('productDetail.quantity') }}</span>
            <div>
              <QzQtyStepper v-model="quantity" :min="quantityEffectiveMin" :max="quantityEffectiveLimit" />
            </div>
          </div>

          <template v-if="pitchLine">
            <p id="qz-pd-desc" class="qz-pd-desc">
              <span class="qz-pd-desc-text" :class="{ 'qz-pd-desc-text--clamp': !descExpanded }">{{ pitchLine }}</span>
            </p>
            <button
              type="button"
              class="qz-btn qz-btn--ghost qz-btn--sm qz-pd-desc-toggle"
              :aria-expanded="descExpanded"
              aria-controls="qz-pd-desc"
              @click="descExpanded = !descExpanded"
            >
              {{ descExpanded ? t('qiaozhi.detail.descCollapse') : t('qiaozhi.detail.descExpand') }}
            </button>
          </template>

          <!-- 内嵌下单 -->
          <div class="qz-pd-actions">
            <form class="qz-pd-form" novalidate @submit.prevent="submitOrder">
              <!-- 游客字段 -->
              <template v-if="!userAuthStore.isAuthenticated">
                <div class="qz-pd-form-row">
                  <div class="qz-field">
                    <label for="qz-pd-email" class="qz-field-label">{{ t('qiaozhi.detail.email') }}<b>*</b></label>
                    <input
                      id="qz-pd-email"
                      v-model="email"
                      type="email"
                      class="qz-field-control"
                      autocomplete="email"
                      :placeholder="t('qiaozhi.detail.emailPlaceholder')"
                      :aria-invalid="!!fieldErrors.email"
                      :aria-describedby="fieldErrors.email ? 'qz-pd-email-err' : undefined"
                    />
                    <p v-if="fieldErrors.email" id="qz-pd-email-err" class="qz-field-error">{{ fieldErrors.email }}</p>
                  </div>
                  <div class="qz-field">
                    <label for="qz-pd-password" class="qz-field-label">{{ t('qiaozhi.detail.orderPassword') }}<b>*</b></label>
                    <input
                      id="qz-pd-password"
                      v-model="orderPassword"
                      type="password"
                      class="qz-field-control"
                      autocomplete="new-password"
                      :placeholder="t('qiaozhi.detail.orderPasswordPlaceholder')"
                      :aria-invalid="!!fieldErrors.password"
                      :aria-describedby="fieldErrors.password ? 'qz-pd-password-err' : undefined"
                    />
                    <p v-if="fieldErrors.password" id="qz-pd-password-err" class="qz-field-error">{{ fieldErrors.password }}</p>
                  </div>
                </div>
                <div v-if="captchaProvider === 'image'" class="qz-field">
                  <span class="qz-field-label">{{ t('qiaozhi.detail.captcha') }}<b>*</b></span>
                  <ImageCaptcha ref="imageCaptchaRef" v-model="captchaPayload" @config-stale="handleCaptchaConfigStale" />
                </div>
                <div v-else-if="captchaProvider === 'turnstile'" class="qz-field">
                  <span class="qz-field-label">{{ t('qiaozhi.detail.captcha') }}<b>*</b></span>
                  <TurnstileCaptcha ref="turnstileRef" v-model="turnstileToken" :site-key="turnstileSiteKey" :theme="theme" />
                </div>
                <p class="qz-pd-form-hint">
                  {{ t('qiaozhi.detail.guestHint') }}
                  <RouterLink :to="`/auth/login?redirect=${encodeURIComponent(route.fullPath)}`">{{ t('qiaozhi.detail.loginToBuy') }}</RouterLink>
                </p>
              </template>
              <p v-else class="qz-pd-form-hint is-ok">{{ t('qiaozhi.detail.loggedInHint') }}</p>

              <!-- 支付方式 -->
              <div v-if="channels.length" class="qz-pd-pay">
                <h3 id="qz-pd-pay-label" class="qz-field-label"><Wallet />{{ t('qiaozhi.detail.payTitle') }}</h3>
                <div class="qz-pd-pay-options" role="radiogroup" aria-labelledby="qz-pd-pay-label" @keydown="onRadioArrow">
                  <button
                    v-for="ch in channels"
                    :key="ch.id"
                    type="button"
                    role="radio"
                    class="qz-pd-pay-option"
                    :class="{ 'is-active': selectedChannelId === ch.id }"
                    :aria-checked="selectedChannelId === ch.id"
                    @click="selectedChannelId = ch.id"
                  >
                    <i aria-hidden="true"></i>
                    <!-- 左侧标识：有上传 icon 用之；没传图且文案看不出币种的 USDT 渠道才补通用币种徽标（对读屏隐藏） -->
                    <img v-if="ch.icon" :src="ch.icon" alt="" />
                    <span v-else-if="ch.isUsdt" class="qz-badge" aria-hidden="true">USDT</span>
                    <span>
                      <b>{{ ch.label }}</b>
                      <small v-if="ch.feeText">{{ ch.feeText }}</small>
                    </span>
                  </button>
                </div>
              </div>

              <div class="qz-pd-total">
                <span>{{ t('qiaozhi.detail.totalDue') }}</span>
                <strong>{{ totalParts.symbol }}{{ totalParts.major }}{{ totalParts.minor }}</strong>
              </div>

              <p v-if="formError" class="qz-alert is-danger" role="alert">{{ formError }}</p>
              <p v-else-if="purchaseWarning" class="qz-alert is-warning" role="alert">{{ purchaseWarning }}</p>
              <p v-else-if="cannotPurchaseReason" class="qz-alert is-warning" role="status">{{ cannotPurchaseReason }}</p>

              <div class="qz-pd-buttons">
                <button
                  type="button"
                  class="qz-icon-btn qz-icon-btn--lg"
                  :disabled="!canPurchase"
                  :aria-label="t('productDetail.addToCart')"
                  :title="t('productDetail.addToCart')"
                  @click="onAddToCart"
                >
                  <ShoppingCart />
                </button>
                <QzButton
                  type="submit"
                  size="lg"
                  class="qz-btn--block"
                  :loading="submitting"
                  :disabled="!canPurchase"
                >
                  <CreditCard v-if="!submitting" />
                  <span>{{ submitting ? t('qiaozhi.detail.submitting') : t('qiaozhi.detail.submitAndPay') }}</span>
                </QzButton>
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- 商品详情 -->
      <section v-if="detailHtml" class="qz-pd-detail">
        <h2>{{ t('qiaozhi.detail.details') }}</h2>
        <div class="qz-pd-prose" v-html="detailHtml"></div>
      </section>

      <!-- 相关文章 -->
      <section v-if="relatedPosts.length" class="qz-pd-detail">
        <h2>{{ t('productDetail.relatedPosts') }}</h2>
        <div class="qz-pd-posts">
          <RouterLink v-for="rp in relatedPosts" :key="rp.id" :to="`/blog/${rp.slug}`">
            <b>{{ getLocalizedText(rp.title) }}</b>
            <small>{{ rp.summary ? getLocalizedText(rp.summary).slice(0, 60) : '' }}</small>
          </RouterLink>
        </div>
      </section>
    </template>

    <!-- 错误 -->
    <div v-else class="qz-empty">
      <PackageOpen />
      <b>{{ t('productDetail.notFound') }}</b>
      <button type="button" class="qz-btn qz-btn--primary" @click="loadProduct">{{ t('errorBoundary.retry') }}</button>
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
import { getAffiliateCode, getAffiliateVisitorKey } from '../../utils/affiliate'
import { channelDisplayName, hasUsdtToken } from '../../utils/cryptoChain'
import { useTheme } from '../../utils/theme'
import ImageCaptcha from '../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../components/captcha/TurnstileCaptcha.vue'
import { useQiaozhiPrice } from './utils/price'
import { shouldOpenInlineModal } from './utils/paymentDisplay'
import { buildCreateAndPayPayload, hasManualFormFields } from './utils/orderPayload'
import PaymentModal, { type PaymentModalItem, type PaymentModalOrder } from './PaymentModal.vue'
import QzButton from './components/QzButton.vue'
import QzQtyStepper from './components/QzQtyStepper.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
const userAuthStore = useUserAuthStore()
const { split: splitPrice } = useQiaozhiPrice()
// 站点亮/暗主题：传给 Turnstile，暗色下验证码跟着切深色（组件内监听 theme 变化重渲染）
const { theme } = useTheme()

const {
  loading, product, relatedPosts,
  getLocalizedText, getFulfillmentTypeLabel, getPurchaseTypeLabel, getStockStatusLabel,
  hasPromotionPrice, getPromotionPriceAmount,
  hasSkuPromotionPrice, getSkuPromotionPriceAmount,
  activeSkus, selectedSku, selectedSkuId, quantity,
  isSkuPurchasable, skuDisplayText, skuStockText,
  requiresLogin, requiresSKUSelection, canPurchase, cannotPurchaseReason,
  quantityEffectiveMin, quantityEffectiveLimit,
  addToCart, buyNow, loadProduct, purchaseWarning,
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

// 单选组方向键（规格 / 支付方式共用）：←/↑ 上一项、→/↓ 下一项，循环且跳过不可选项；
// 走 click() 复用各组自己的选中逻辑，焦点与选中一起移动。
const onRadioArrow = (event: KeyboardEvent) => {
  if (event.altKey || event.ctrlKey || event.metaKey) return // Alt+← 等浏览器快捷键放行
  const step = ({ ArrowRight: 1, ArrowDown: 1, ArrowLeft: -1, ArrowUp: -1 } as Record<string, number>)[event.key]
  if (!step) return
  const radios = Array.from((event.currentTarget as HTMLElement).querySelectorAll<HTMLButtonElement>('[role="radio"]:not(:disabled)'))
  const index = radios.indexOf(event.target as HTMLButtonElement)
  const next = radios[(index + step + radios.length) % radios.length]
  if (index < 0 || radios.length < 2 || !next) return
  event.preventDefault()
  next.focus()
  next.click()
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
    // 促销标用 base 的状态徽标（.qz-badge.is-danger = 红色 12% 底）
    return { tone: 'is-danger', text: t('products.promotionTag') }
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

// 游客下单的图形验证码若因上游配置已变更而失效（ImageCaptcha 拉不到新图时 emit config-stale），
// 强刷 /public/config 让本页取到最新 captcha / 渠道配置，顺便把已废弃的 payload 清掉。
// 与 Login / Register / Forgot 三页 useLogin / useRegister / useForgot 中的同名 handler 保持一致。
const handleCaptchaConfigStale = async () => {
  await appStore.loadConfig(true)
  captchaPayload.value = {}
  turnstileToken.value = ''
}

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
      const name = String(ch.name || ch.channel_type || ch.provider_type || '')
      const channelType = String(ch?.channel_type || '')
      // name 原样保留给支付弹窗 / 支付记录回填；label 只用于本页支付方式行：
      // 渠道名纯粹是「USDT + 单一链」时统一成「USDT · 链名」，其余一律显示管理端原名（只看名字，不信 channel_type）。
      const label = channelDisplayName(name)
      return {
        id: Number(ch.id),
        name,
        label,
        // 无上传 icon 时的通用币种徽标：只在显示文案里看不出币种（如「泰达币」）而 channel_type 是 usdt 时才补；
        // 文案本身已含 USDT（含统一后的「USDT · 链名」）就不再重复打标。
        isUsdt: !hasUsdtToken(label) && hasUsdtToken(channelType),
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
    feePolicy: String(payData?.fee_policy || ''),
    payableAmount: String(payData?.payable_amount || ''),
    payableCurrency: String(payData?.currency || ''),
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
// 游客且验证码开启时，表单里的图形验证码 / Turnstile token 已在上一次建单时被消费，
// 原样再发必被后端拒绝：先关弹窗、清空并刷新验证码、提示用户重填后再手动提交。
const handleRenewPayment = async () => {
  if (!userAuthStore.isAuthenticated && (captchaProvider.value === 'image' || captchaProvider.value === 'turnstile')) {
    showPaymentModal.value = false
    captchaPayload.value = {}
    turnstileToken.value = ''
    imageCaptchaRef.value?.refresh()
    turnstileRef.value?.reset()
    formError.value = t('qiaozhi.detail.renewNeedsCaptcha')
    return
  }
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
      // /payments 响应不含订单对象：订单侧字段（币种/过期时间/创建时间）沿用切换前值。
      next.currency = next.currency || current.currency
      next.expireAt = next.expireAt ?? current.expireAt ?? null
      next.createdAt = next.createdAt ?? current.createdAt ?? null
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

// 切换网络后是另一笔支付记录：上一渠道的"已提交 / 提交失败"反馈不再适用，清掉。
watch(() => paymentOrder.value?.channelId, () => {
  paymentTxFeedback.value = null
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
  // 手动表单商品需在结算页填写交付资料，商品页直接下单会被后端拒绝：
  // 改走 classic 的 buy-now 路径（buyNowStore + /checkout?mode=buynow）。
  // 放在游客表单校验之前：结算页会自己收集游客邮箱/密码/验证码，这里不必先填一遍。
  if (hasManualFormFields(product.value?.manual_form_schema)) {
    buyNow()
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
    // 分销归因与 classic useCheckout 一致：空值在序列化时不出现。
    const affiliateCode = getAffiliateCode()
    const affiliateVisitorKey = getAffiliateVisitorKey()
    if (userAuthStore.isAuthenticated) {
      const res = await userOrderAPI.createAndPay(buildCreateAndPayPayload({
        items,
        channelId: selectedChannelId.value,
        affiliateCode,
        affiliateVisitorKey,
      }))
      payData = res.data?.data
    } else {
      const res = await guestOrderAPI.createAndPay(buildCreateAndPayPayload({
        items,
        channelId: selectedChannelId.value,
        affiliateCode,
        affiliateVisitorKey,
        guest: {
          email: guestEmailInput,
          orderPassword: guestPasswordInput,
          captchaPayload: captchaProvider.value === 'image'
            ? captchaPayload.value
            : (turnstileToken.value ? { turnstile_token: turnstileToken.value } : {}),
        },
      }))
      saveGuestOrderAuth({ email: guestEmailInput, order_password: guestPasswordInput })
      payData = res.data?.data
    }
    const orderNo = payData?.order_no || ''
    if (!orderNo) throw new Error(t('qiaozhi.detail.errSubmit'))

    // 后端 /orders/create-and-pay 已直接返回支付字段:qr_code / pay_url /
    // expires_at / interaction_mode / wallet_address。扫码类(qr)渠道与带链上收款地址的加密货币单
    // 拿到完整响应就当页弹 Modal,不再跳 /pay(加密单弹窗渲染地址二维码,与 interaction_mode 无关)。
    // 纯跳转类(redirect/wap/page,无地址)渠道走 /pay,由 usePayment 负责自动打开支付链接。
    // 没有支付字段(例如 channel_id 为 0、全额钱包抵扣、通道返错)时同样 fallback 到
    // 旧的 /pay 独立页,保留原流程不丢失。
    if (shouldOpenInlineModal(payData)) {
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
      // 「重新发起支付」会在弹窗打开时调到这里:/pay 是懒加载视图,先关弹窗,
      // 避免 chunk 加载期间再点「重新发起」重复建单。
      showPaymentModal.value = false
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

<style scoped>
/* 商品详情页：控件全部来自 styles/base.css（.qz-btn / .qz-icon-btn / .qz-pill / .qz-chip / .qz-badge /
   .qz-field*），这里只放页面布局与详情页专属元素，只读 --qz-* token（无色值字面量、无 px 字号）。

   选择器一律挂在页面根 .qz-pd-page 下：多一层根类让本页规则对 base.css 的 (0,2,0) 基类在特指度上稳赢，
   不依赖加载顺序；本页用到的每个 .qz-pd-* 元素的全部属性都在这里显式声明（支付页的同类元素已改名
   qz-pay-*，两页互不共享规则）。 */

/* ---------- 页面骨架 ---------- */
.qz-pd-page {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
  max-width: 1080px;
}
.qz-pd-page .qz-pd-return { align-self: flex-start; }

/* 骨架块都是 base.css 的 .qz-skeleton（标题 / 正文 / 小字行由 .is-lg / .is-line / .is-sm 定高，≤760 跟着 h1 换 xl）；
   主图台与购买钮占位（.qz-pd-sk-media / .qz-pd-sk-cta）只在这里给尺寸。 */
.qz-pd-page .qz-pd-skeleton { display: grid; gap: var(--qz-space-4); }
.qz-pd-page .qz-pd-sk-media { aspect-ratio: 4 / 3; max-height: 420px; }
/* 标题骨架 = h1 行盒：页面 h1 用 body 行高（base .is-lg 默认 display 行高），首屏到真实标题不跳高 */
.qz-pd-page .qz-skeleton.is-lg { line-height: var(--qz-leading-body); }
.qz-pd-page .qz-pd-sk-cta { height: var(--qz-control-lg); margin-top: auto; }

/* ---------- 左图右单 ---------- */
.qz-pd-page .qz-pd-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  overflow: visible; /* 媒体列 sticky 需要 */
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-panel);
  background: var(--qz-surface);
  box-shadow: var(--qz-shadow-resting);
}

/* 媒体列透明底；圆角跟随卡片内缘（≤1023 单列时改为上两角），底色一旦加回来也对得上 */
.qz-pd-page .qz-pd-media {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: var(--qz-space-3);
  min-width: 0;
  padding: var(--qz-space-4);
  border-radius: calc(var(--qz-radius-panel) - 1px) 0 0 calc(var(--qz-radius-panel) - 1px);
  background: transparent;
}
.qz-pd-page .qz-pd-media-sticky {
  position: sticky;
  top: calc(var(--qz-header-h) + var(--qz-space-4));
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-3);
}
/* 主图台：4/3（手机 16/9），contain 不裁图，底色用弱化面 */
.qz-pd-page .qz-pd-stage {
  position: relative;
  display: grid;
  place-items: center;
  width: 100%;
  max-height: none;
  margin: 0;
  aspect-ratio: 4 / 3;
  overflow: hidden;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
}
.qz-pd-page .qz-pd-stage > img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
}
/* 无图兜底：首字母圆徽 */
.qz-pd-page .qz-pd-logo {
  display: grid;
  place-items: center;
  width: 28%;
  aspect-ratio: 1;
  border-radius: 50%;
  background: var(--qz-surface);
  color: var(--qz-text-secondary);
  box-shadow: var(--qz-shadow-resting);
}
.qz-pd-page .qz-pd-logo b { font-size: var(--qz-text-3xl); font-weight: 800; line-height: 1; }

.qz-pd-page .qz-pd-thumbs { display: flex; gap: var(--qz-space-2); overflow-x: auto; scrollbar-width: none; }
.qz-pd-page .qz-pd-thumbs button {
  flex: none;
  width: 64px;
  height: 64px;
  padding: 0;
  overflow: hidden;
  border: 2px solid transparent;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  opacity: .65;
  cursor: pointer;
  transition:
    opacity var(--qz-duration-fast) var(--qz-ease-standard),
    border-color var(--qz-duration-fast) var(--qz-ease-standard);
}
.qz-pd-page .qz-pd-thumbs button.is-active { border-color: var(--qz-accent); opacity: 1; }
.qz-pd-page .qz-pd-thumbs img { width: 100%; height: 100%; object-fit: cover; }

.qz-pd-page .qz-pd-info {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-3);
  min-width: 0;
  padding: var(--qz-space-5) var(--qz-space-6);
}

/* 标题行：h1 + 右端幽灵分享钮（负外边距抵掉按钮内边距，让文字与右缘对齐） */
.qz-pd-page .qz-pd-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--qz-space-3);
}
.qz-pd-page .qz-pd-title-row h1 {
  min-width: 0;
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-2xl);
  font-weight: 700;
  line-height: var(--qz-leading-body);
  letter-spacing: var(--qz-tracking-display);
  overflow-wrap: anywhere;
}
.qz-pd-page .qz-pd-title-row .qz-btn { flex: none; margin-right: calc(-1 * var(--qz-space-3)); }

/* 信息 chips：钉住 base.css .qz-chip 的尺寸与字色（与 base 同值），页面根类把它提到 (0,3,0)，
   任何同名共享规则都压不过它。 */
.qz-pd-page .qz-pd-chip-row { display: flex; flex-wrap: wrap; gap: var(--qz-space-2); margin-top: calc(-1 * var(--qz-space-1)); }
.qz-pd-page .qz-chip {
  min-height: 24px;
  padding: 0 var(--qz-space-2);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-2xs);
  font-weight: 600;
  line-height: 1;
}
.qz-pd-page .qz-chip svg { width: 12px; height: 12px; }
.qz-pd-page .qz-chip.is-danger { color: var(--qz-danger); background: color-mix(in oklab, var(--qz-danger) 12%, transparent); }

/* 价格：符号与数字之间不留 gap（¥28.90），数字等宽；促销徽标靠行右端 */
.qz-pd-page .qz-pd-price-row { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: var(--qz-space-3); }
.qz-pd-page .qz-pd-price {
  display: inline-flex;
  align-items: baseline;
  flex-wrap: nowrap;
  gap: 0;
  color: var(--qz-text);
  font-variant-numeric: tabular-nums;
}
.qz-pd-page .qz-pd-price-sym { color: var(--qz-text); font-size: var(--qz-text-lg); font-weight: 700; }
.qz-pd-page .qz-pd-price strong {
  color: var(--qz-text);
  font-size: var(--qz-text-3xl);
  font-weight: 800;
  line-height: var(--qz-leading-display);
  letter-spacing: 0;
}

/* 规格：.qz-pill 取 md 高；选中 = 1px 强调色描边 + 弱强调底（无内阴影）；不可选 = 虚线边 + 弱化字 */
.qz-pd-page .qz-pd-sku-list { display: flex; flex-wrap: wrap; gap: var(--qz-space-2); }
.qz-pd-page .qz-pd-sku-list .qz-pill {
  min-height: var(--qz-control-md);
  max-width: 100%;
  padding-inline: var(--qz-space-4);
  border: 1px solid var(--qz-border-strong);
  color: var(--qz-text);
  text-align: left;
  transition:
    color var(--qz-duration-fast) var(--qz-ease-standard),
    background-color var(--qz-duration-fast) var(--qz-ease-standard),
    border-color var(--qz-duration-fast) var(--qz-ease-standard);
}
.qz-pd-page .qz-pd-sku-list .qz-pill.is-active {
  border-color: var(--qz-accent);
  background: var(--qz-accent-subtle);
  color: var(--qz-text);
  box-shadow: none;
}
.qz-pd-page .qz-pd-sku-list .qz-pill:disabled {
  border-style: dashed;
  background: transparent;
  color: var(--qz-text-muted);
  cursor: not-allowed;
}
.qz-pd-page .qz-pd-sku-list .qz-pill b { min-width: 0; overflow: hidden; font-weight: 600; text-overflow: ellipsis; }
.qz-pd-page .qz-pd-sku-list .qz-pill small { color: var(--qz-text-muted); font-size: var(--qz-text-xs); }
.qz-pd-page .qz-pd-sku-list .qz-pill small::before { content: "·"; margin-right: var(--qz-space-1); }
.qz-pd-page .qz-pd-sku-list .qz-pill.is-active small { color: var(--qz-text-secondary); }

/* 简介折叠：外层只管内边距 / 底色，行数裁切放在内层文本上，折叠态底部渐隐，不露半行 */
.qz-pd-page .qz-pd-desc {
  position: relative;
  margin: 0;
  padding: var(--qz-space-2) var(--qz-space-3);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: 1.55;
  white-space: normal;
}
.qz-pd-page .qz-pd-desc-text { display: block; white-space: pre-line; }
/* 链上基类名：(0,4,0) 压过 .qz-pd-desc-text 的 display/white-space，不依赖源顺序 */
.qz-pd-page .qz-pd-desc-text.qz-pd-desc-text--clamp {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  white-space: normal;
  mask-image: linear-gradient(currentColor 70%, transparent);
}
.qz-pd-page .qz-pd-desc-toggle { align-self: flex-start; margin-top: calc(-1 * var(--qz-space-2)); }

/* ---------- 内嵌下单 ---------- */
.qz-pd-page .qz-pd-actions {
  display: grid;
  gap: var(--qz-space-3);
  margin-top: auto;
  padding-top: var(--qz-space-3);
  border-top: 1px solid var(--qz-border);
}
.qz-pd-page .qz-pd-form { display: grid; gap: var(--qz-space-3); }
.qz-pd-page .qz-pd-form-row { display: grid; grid-template-columns: 1fr 1fr; gap: var(--qz-space-3); }
.qz-pd-page .qz-pd-form-hint {
  margin: 0;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}
.qz-pd-page .qz-pd-form-hint a {
  margin-left: var(--qz-space-1);
  color: var(--qz-text);
  font-weight: 600;
  text-decoration: underline;
  text-underline-offset: 2px;
}
.qz-pd-page .qz-pd-form-hint.is-ok { color: var(--qz-success); }

/* 支付方式：弱化面板内的卡片式单选行；选中 = 1px 强调色描边 + 弱强调底 */
.qz-pd-page .qz-pd-pay {
  display: grid;
  gap: var(--qz-space-2);
  padding: var(--qz-space-3);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
}
.qz-pd-page .qz-pd-pay h3 { display: flex; align-items: center; gap: var(--qz-space-1); margin: 0; }
.qz-pd-page .qz-pd-pay h3 svg { width: 16px; height: 16px; color: var(--qz-text-muted); }
.qz-pd-page .qz-pd-pay-options { display: grid; grid-template-columns: repeat(auto-fill, minmax(170px, 1fr)); gap: var(--qz-space-2); }
.qz-pd-page .qz-pd-pay-option {
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
  .qz-pd-page .qz-pd-pay-option:hover:where(:not(.is-active)) { border-color: var(--qz-border-strong); }
}
.qz-pd-page .qz-pd-pay-option.is-active { border-color: var(--qz-accent); background: var(--qz-accent-subtle); }
.qz-pd-page .qz-pd-pay-option > i {
  flex: none;
  width: 16px;
  height: 16px;
  border: 2px solid var(--qz-border-strong);
  border-radius: 50%;
  background: var(--qz-surface);
  transition: border var(--qz-duration-fast) var(--qz-ease-standard);
}
.qz-pd-page .qz-pd-pay-option.is-active > i { border: 5px solid var(--qz-accent); }
.qz-pd-page .qz-pd-pay-option img { flex: none; width: 24px; height: 24px; border-radius: var(--qz-radius-xs); object-fit: contain; }
/* 无 icon 的 USDT 渠道：通用币种徽标占 icon 位，不参与伸缩 */
.qz-pd-page .qz-pd-pay-option > .qz-badge { flex: none; }
.qz-pd-page .qz-pd-pay-option > span:not(.qz-badge) { display: grid; min-width: 0; }
/* 标签最多两行再截断（两列格里「USDT · Ethereum」一行放不下时折行而不是省略号） */
.qz-pd-page .qz-pd-pay-option b { display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; font-weight: 600; overflow-wrap: anywhere; }
.qz-pd-page .qz-pd-pay-option small { color: var(--qz-text-muted); font-size: var(--qz-text-xs); }

.qz-pd-page .qz-pd-total {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--qz-space-3);
  padding-top: var(--qz-space-2);
  border-top: 1px solid var(--qz-border);
}
.qz-pd-page .qz-pd-total span { color: var(--qz-text-secondary); font-size: var(--qz-text-sm); }
.qz-pd-page .qz-pd-total strong {
  color: var(--qz-text);
  font-size: var(--qz-text-xl);
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

/* 表单提示条（不可购买原因 / 校验错误）用 base.css 的 .qz-alert.is-warning / .is-danger */

/* 购物车图标钮（lg 圆钮）+ lg 主按钮：同为 --qz-control-lg 高，一行等高 */
.qz-pd-page .qz-pd-buttons {
  display: grid;
  grid-template-columns: var(--qz-control-lg) minmax(0, 1fr);
  gap: var(--qz-space-2);
  align-items: center;
}

/* ---------- 商品详情 / 相关文章 ---------- */
.qz-pd-page .qz-pd-detail {
  padding: var(--qz-space-6);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-panel);
  background: var(--qz-surface);
  box-shadow: var(--qz-shadow-resting);
}
.qz-pd-page .qz-pd-detail > h2 {
  margin: 0 0 var(--qz-space-3);
  color: var(--qz-text);
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: var(--qz-leading-body);
}
/* 富文本由 v-html 注入，没有 scoped 属性，用 :deep() 命中 */
.qz-pd-page .qz-pd-prose { color: var(--qz-text); font-size: var(--qz-text-sm); line-height: 1.75; overflow-wrap: anywhere; }
.qz-pd-page .qz-pd-prose :deep(p) { margin: 0 0 var(--qz-space-3); }
.qz-pd-page .qz-pd-prose :deep(p:last-child) { margin-bottom: 0; }
.qz-pd-page .qz-pd-prose :deep(img) { max-width: 100%; height: auto; border-radius: var(--qz-radius-control); }
.qz-pd-page .qz-pd-prose :deep(ul),
.qz-pd-page .qz-pd-prose :deep(ol) { margin: 0 0 var(--qz-space-3); padding-left: 1.4em; }
.qz-pd-page .qz-pd-prose :deep(ul) { list-style: disc; }
.qz-pd-page .qz-pd-prose :deep(ol) { list-style: decimal; }
.qz-pd-page .qz-pd-prose :deep(h1),
.qz-pd-page .qz-pd-prose :deep(h2),
.qz-pd-page .qz-pd-prose :deep(h3),
.qz-pd-page .qz-pd-prose :deep(h4) { margin: var(--qz-space-4) 0 var(--qz-space-2); font-weight: 700; line-height: var(--qz-leading-body); }
.qz-pd-page .qz-pd-prose :deep(h1) { font-size: var(--qz-text-xl); }
.qz-pd-page .qz-pd-prose :deep(h2) { font-size: var(--qz-text-lg); }
.qz-pd-page .qz-pd-prose :deep(h3),
.qz-pd-page .qz-pd-prose :deep(h4) { font-size: var(--qz-text-base); }
.qz-pd-page .qz-pd-prose :deep(a) { color: var(--qz-text); text-decoration: underline; text-underline-offset: 2px; }

.qz-pd-page .qz-pd-posts { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: var(--qz-space-3); }
.qz-pd-page .qz-pd-posts a {
  display: grid;
  gap: var(--qz-space-1);
  padding: var(--qz-space-3);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  color: var(--qz-text);
  text-decoration: none;
  transition:
    border-color var(--qz-duration-fast) var(--qz-ease-standard),
    background-color var(--qz-duration-fast) var(--qz-ease-standard);
}
@media (hover: hover) and (pointer: fine) {
  .qz-pd-page .qz-pd-posts a:hover { border-color: var(--qz-border-strong); background: var(--qz-surface-muted); }
}
.qz-pd-page .qz-pd-posts b { font-size: var(--qz-text-sm); font-weight: 600; line-height: var(--qz-leading-body); }
.qz-pd-page .qz-pd-posts small { color: var(--qz-text-muted); font-size: var(--qz-text-xs); line-height: var(--qz-leading-body); }

/* ---------- 响应式 ---------- */
@media (max-width: 1023px) {
  .qz-pd-page { gap: var(--qz-space-3); }
  .qz-pd-page .qz-pd-card { grid-template-columns: minmax(0, 1fr); }
  .qz-pd-page .qz-pd-media { border-radius: calc(var(--qz-radius-panel) - 1px) calc(var(--qz-radius-panel) - 1px) 0 0; }
  .qz-pd-page .qz-pd-media-sticky { position: static; }
  /* 单列时主图台上限随视口走（最高 480），不再被 320 硬顶压扁 4/3；骨架图块同高，加载前后不跳 */
  .qz-pd-page .qz-pd-stage,
  .qz-pd-page .qz-pd-sk-media { max-height: min(60vh, 480px); }
}
@media (max-width: 760px) {
  /* 手机端返回钮吸在顶栏下方；桌面保持静态 */
  .qz-pd-page .qz-pd-return {
    position: sticky;
    top: calc(var(--qz-header-h) + var(--qz-space-2));
    z-index: 5;
  }
  /* 页面 h1：桌面 2xl（28），手机 xl（22）；标题骨架（.qz-skeleton.is-lg）由 base 同步换挡 */
  .qz-pd-page .qz-pd-title-row h1 { font-size: var(--qz-text-xl); }
}
@media (max-width: 640px) {
  .qz-pd-page .qz-pd-media { padding: var(--qz-space-3); }
  .qz-pd-page .qz-pd-stage { aspect-ratio: 16 / 9; }
  .qz-pd-page .qz-pd-info { padding: var(--qz-space-4) var(--qz-space-3); }
  .qz-pd-page .qz-pd-price strong { font-size: var(--qz-text-2xl); }
  .qz-pd-page .qz-pd-form-row { grid-template-columns: 1fr; }
  /* 手机两列支付格里徽标会把标签挤到省略号，收起徽标（标签本身已含 USDT） */
  .qz-pd-page .qz-pd-pay-option > .qz-badge { display: none; }
  .qz-pd-page .qz-pd-detail { padding: var(--qz-space-4) var(--qz-space-3) var(--qz-space-5); }
}

@media (prefers-reduced-motion: reduce) {
  .qz-pd-page .qz-pd-thumbs button,
  .qz-pd-page .qz-pd-sku-list .qz-pill,
  .qz-pd-page .qz-pd-pay-option,
  .qz-pd-page .qz-pd-pay-option > i,
  .qz-pd-page .qz-pd-posts a { transition: none; }
}
</style>
