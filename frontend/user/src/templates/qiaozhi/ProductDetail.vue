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
        <!-- 左：媒体区 -->
        <div class="j8-pd-media">
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
            <div class="j8-pd-qty">
              <button type="button" :disabled="quantity <= quantityEffectiveMin" :aria-label="t('common.previous')" @click="quantity = Math.max(quantityEffectiveMin, quantity - 1)">
                <Minus />
              </button>
              <input inputmode="numeric" :value="quantity" @change="handleQuantityInput($event)" />
              <button type="button" :disabled="quantityEffectiveLimit !== null && quantity >= quantityEffectiveLimit" :aria-label="t('common.next')" @click="quantity += 1">
                <Plus />
              </button>
            </div>
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
                <button type="submit" class="j8-pd-buy" :disabled="!canPurchase || submitting">
                  <Loader2 v-if="submitting" class="j8-ic-spin" />
                  <CreditCard v-else />
                  {{ submitting ? t('qiaozhi.detail.submitting') : t('qiaozhi.detail.submitAndPay') }}
                </button>
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  ChevronLeft, CreditCard, Loader2, Minus, PackageOpen, Plus, Share2, ShoppingCart, Wallet, Zap,
} from 'lucide-vue-next'
import { useAppStore } from '../../stores/app'
import { useUserAuthStore } from '../../stores/userAuth'
import { guestOrderAPI, userOrderAPI } from '../../api'
import { useProductDetail } from '../../composables/useProductDetail'
import { processHtmlForDisplay } from '../../utils/content'
import { getFirstImageUrl, getImageUrl } from '../../utils/image'
import { normalizeSkuId } from '../../utils/sku'
import { saveGuestOrderAuth } from '../../utils/guestOrderAuth'
import ImageCaptcha from '../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../components/captcha/TurnstileCaptcha.vue'
import { useQiaozhiPrice } from './utils/price'

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
  quantityEffectiveMin, quantityEffectiveLimit, handleQuantityInput,
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
    let orderNo = ''
    if (userAuthStore.isAuthenticated) {
      const res = await userOrderAPI.createAndPay({ items, channel_id: selectedChannelId.value || undefined })
      orderNo = res.data?.data?.order_no || ''
    } else {
      const res = await guestOrderAPI.createAndPay({
        items,
        channel_id: selectedChannelId.value || undefined,
        email: email.value.trim(),
        order_password: orderPassword.value,
        captcha_payload: captchaProvider.value === 'image'
          ? captchaPayload.value
          : (turnstileToken.value ? { turnstile_token: turnstileToken.value } : {}),
      })
      saveGuestOrderAuth({ email: email.value.trim(), order_password: orderPassword.value })
      orderNo = res.data?.data?.order_no || ''
    }
    if (!orderNo) throw new Error(t('qiaozhi.detail.errSubmit'))
    const query = userAuthStore.isAuthenticated
      ? `order_no=${encodeURIComponent(orderNo)}`
      : `guest=1&order_no=${encodeURIComponent(orderNo)}`
    router.push(`/pay?${query}`)
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
