<template>
  <div class="page-width">
    <!-- Loading -->
    <div v-if="loading" class="tibo-detail-loading reveal-in">
      <div class="skeleton-block skeleton-left"></div>
      <div class="skeleton-block skeleton-right"></div>
    </div>

    <template v-else-if="product">
      <!-- 引导区 -->
      <section class="detail-intro reveal-in">
        <nav class="product-breadcrumb" aria-label="breadcrumb">
          <RouterLink to="/">{{ t('products.allCategories') }}</RouterLink>
          <ChevronRight class="crumb-icon" />
          <span aria-current="page">{{ title }}</span>
        </nav>
        <h1>{{ title }}</h1>
        <p v-if="pitchLead">{{ pitchLead }}</p>
      </section>

      <!-- 购买工作区 -->
      <section class="purchase-workspace reveal-in delay-1" :aria-label="t('tibo.detail.pickPlan')">
        <!-- 左侧：套餐选择 -->
        <div class="product-area">
          <div class="product-topline">
            <span class="section-kicker">{{ t('tibo.detail.pickPlan') }}</span>
            <span class="availability" :class="{ 'is-out': soldOut }">
              <span></span>
              {{ soldOut ? t('products.stockStatus.outOfStock') : t('tibo.detail.available') }}
            </span>
          </div>

          <div class="product-heading">
            <div class="product-mark">
              <img v-if="markImage" :src="markImage" alt="" />
              <Package v-else />
            </div>
            <div>
              <h2>{{ title }}</h2>
              <p>{{ categoryName || t('tibo.detail.headingFallback') }}</p>
            </div>
            <span v-if="categoryName" class="product-badge">
              <Tag class="badge-icon" />
              {{ categoryName }}
            </span>
          </div>

          <div v-if="pitchTitle || pitchSub" class="product-pitch">
            <h3 v-if="pitchTitle">{{ pitchTitle }}</h3>
            <p v-if="pitchSub">{{ pitchSub }}</p>
          </div>

          <!-- 套餐 -->
          <template v-if="plans.length">
            <div class="plan-label">
              <span>{{ t('tibo.detail.planLabel') }}</span>
              <span>{{ t('tibo.detail.oneTime') }}</span>
            </div>
            <div
              role="radiogroup"
              :aria-label="t('productDetail.skuTitle')"
              class="plan-options subscription-plans"
            >
              <button
                v-for="plan in plans"
                :key="plan.id"
                type="button"
                role="radio"
                :aria-checked="plan.selected"
                :disabled="plan.disabled"
                class="plan-option"
                :class="{ 'is-selected': plan.selected }"
                @click="selectPlan(plan)"
              >
                <span class="plan-option-top">
                  <span class="plan-radio" aria-hidden="true"></span>
                  <span>{{ plan.name }}</span>
                </span>
                <p>{{ plan.hint }}</p>
                <span class="plan-price">
                  <span>{{ plan.priceParts.symbol }}</span>
                  <strong>{{ plan.priceParts.major }}</strong>
                  <span>{{ plan.priceParts.minor }} {{ t('tibo.detail.perUnit') }}</span>
                </span>
              </button>
            </div>
            <p v-if="requiresSKUSelection" class="plan-required">{{ t('productDetail.skuRequired') }}</p>
          </template>

          <!-- 特点 -->
          <div class="plan-features">
            <span><Check class="feature-icon" />{{ fulfillmentLabel }}</span>
            <span><Check class="feature-icon" />{{ purchaseTypeLabel }}</span>
            <span><Check class="feature-icon" />{{ t('tibo.detail.featureTrackable') }}</span>
          </div>

          <div class="product-bottom">
            <ShieldCheck />
            <span>{{ t('tibo.detail.bottomNote') }}</span>
            <button type="button" class="text-button" @click="noticeOpen = true">
              {{ t('tibo.notice.title') }}
              <ChevronRight class="bottom-icon" />
            </button>
          </div>
        </div>

        <!-- 右侧：订单摘要 -->
        <aside class="order-summary" :aria-label="t('tibo.detail.orderTitle')">
          <div class="summary-title">
            <h2>{{ t('tibo.detail.orderTitle') }}</h2>
            <Wallet />
          </div>

          <div class="summary-product">
            <div class="product-mark small">
              <img v-if="markImage" :src="markImage" alt="" />
              <Package v-else />
            </div>
            <div>
              <strong>{{ title }}<template v-if="selectedPlanName"> · <span class="state-enter">{{ selectedPlanName }}</span></template></strong>
              <p>{{ categoryName || fulfillmentLabel }}</p>
            </div>
          </div>

          <dl class="summary-details">
            <div>
              <dt>{{ t('tibo.detail.qty') }}</dt>
              <dd>{{ quantity }}</dd>
            </div>
            <div>
              <dt>{{ t('tibo.detail.delivery') }}</dt>
              <dd>{{ fulfillmentLabel }}</dd>
            </div>
            <div>
              <dt>{{ t('tibo.detail.payment') }}</dt>
              <dd>{{ t('tibo.detail.onlinePayment') }}</dd>
            </div>
          </dl>

          <div class="summary-total">
            <div>
              <span>{{ t('tibo.detail.total') }}</span>
              <span class="price-caption">{{ t('tibo.detail.totalCaption') }}</span>
            </div>
            <div class="total-price">
              <small>{{ totalParts.symbol }}</small>{{ totalParts.major }}<span>{{ totalParts.minor }}</span>
            </div>
          </div>

          <div v-if="cannotPurchaseReason" class="summary-warning">{{ cannotPurchaseReason }}</div>

          <button v-if="requiresLogin" type="button" class="summary-cta summary-login" @click="goLogin">
            {{ t('productDetail.loginToBuy') }}
          </button>
          <button v-else type="button" class="summary-cta" :disabled="!canPurchase" @click="onBuyClick">
            {{ t('tibo.detail.buyNow') }}
            <ArrowRight />
          </button>

          <p class="summary-note">
            <LockKeyhole />
            {{ t('tibo.detail.secure') }}
          </p>

          <div class="summary-footnote">
            <CircleHelp />
            <p>
              {{ t('tibo.detail.footNote') }}
              <button type="button" @click="noticeOpen = true">{{ t('tibo.notice.title') }}</button>
            </p>
          </div>
        </aside>
      </section>

      <!-- 交付流程 -->
      <section class="delivery-guide" :aria-label="t('tibo.detail.howItWorks')">
        <article>
          <small>{{ t('tibo.detail.steps.step1k') }}</small>
          <h3>{{ t('tibo.detail.steps.step1t') }}</h3>
          <p>{{ t('tibo.detail.steps.step1d') }}</p>
        </article>
        <article>
          <small>{{ t('tibo.detail.steps.step2k') }}</small>
          <h3>{{ t('tibo.detail.steps.step2t') }}</h3>
          <p>{{ t('tibo.detail.steps.step2d') }}</p>
        </article>
        <article>
          <small>{{ t('tibo.detail.steps.step3k') }}</small>
          <h3>{{ t('tibo.detail.steps.step3t') }}</h3>
          <p>{{ t('tibo.detail.steps.step3d') }}</p>
        </article>
      </section>

      <!-- 商品详情 -->
      <section v-if="descriptionText || product.content" class="product-details reveal-in">
        <h2>{{ t('productDetail.details') }}</h2>
        <div class="details-body">
          <p v-if="descriptionText">{{ descriptionText }}</p>
          <div v-if="product.content" v-html="processHtmlForDisplay(getLocalizedText(product.content))"></div>
        </div>
      </section>
    </template>

    <!-- 错误 -->
    <div v-else class="tibo-empty">
      <AlertCircle class="empty-icon" />
      <p>{{ t('productDetail.notFound') }}</p>
      <button type="button" class="store-action" @click="loadProduct">{{ t('errorBoundary.retry') }}</button>
    </div>

    <!-- 购买须知 -->
    <Teleport to="body">
      <div v-if="noticeOpen" class="tibo-modal-mask" @click.self="noticeOpen = false">
        <div class="tibo-modal" role="dialog" aria-modal="true" :aria-label="t('tibo.notice.title')">
          <h2>{{ t('tibo.notice.title') }}</h2>
          <p class="modal-sub">{{ title }}<template v-if="selectedPlanName"> · {{ selectedPlanName }}</template></p>
          <div class="tibo-notice-emphasis">
            <Info />
            <strong>{{ t('tibo.notice.emphasis') }}</strong>
          </div>
          <ol class="tibo-notice-steps">
            <li>{{ t('tibo.notice.line1') }}</li>
            <li>{{ t('tibo.notice.line2') }}</li>
          </ol>
          <div class="tibo-modal-actions">
            <button type="button" @click="noticeOpen = false">{{ t('tibo.notice.cancel') }}</button>
            <button type="button" class="modal-primary" @click="confirmBuy">
              {{ t('tibo.notice.confirm') }}
              <ArrowRight class="btn-icon" />
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  AlertCircle, ArrowRight, Check, ChevronRight, CircleHelp, Info, LockKeyhole,
  Package, ShieldCheck, Tag, Wallet,
} from 'lucide-vue-next'
import { useProductDetail } from '../../composables/useProductDetail'
import { processHtmlForDisplay } from '../../utils/content'
import { getFirstImageUrl } from '../../utils/image'
import { useLocalized } from '../../composables/useProduct'
import { useTiboPrice } from './utils/price'
import './styles/tibo-detail.css'

const { t } = useI18n()
const { getLocalizedText } = useLocalized()
const { split: splitPrice } = useTiboPrice()

const noticeOpen = ref(false)

const {
  loading, product,
  getFulfillmentTypeLabel, getPurchaseTypeLabel, getStockStatusLabel,
  hasPromotionPrice, getPromotionPriceAmount,
  hasSkuPromotionPrice, getSkuPromotionPriceAmount,
  normalizeSkuId,
  activeSkus, selectedSku, selectedSkuId, quantity,
  isSkuPurchasable, skuDisplayText, skuStockText,
  requiresLogin, requiresSKUSelection, canPurchase, cannotPurchaseReason,
  categoryName,
  buyNow, goLogin, loadProduct,
} = useProductDetail()

const title = computed(() => getLocalizedText(product.value?.title))
const fulfillmentLabel = computed(() => getFulfillmentTypeLabel(product.value?.fulfillment_type))
const purchaseTypeLabel = computed(() => getPurchaseTypeLabel(product.value?.purchase_type))
const soldOut = computed(() => getStockStatusLabel(product.value) === t('products.stockStatus.outOfStock'))

const markImage = computed(() => {
  const primary = getFirstImageUrl(product.value?.images)
  if (primary) return primary
  return ''
})

const descriptionText = computed(() => {
  const raw = getLocalizedText(product.value?.description)
  return raw ? raw.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').trim() : ''
})

// 卖点两行：描述按首句拆
const pitchLead = computed(() => {
  const text = descriptionText.value
  if (!text) return ''
  const cut = text.search(/[。.!?！？]/)
  return cut === -1 || cut > 30 ? text.slice(0, 30) : text.slice(0, cut + 1)
})
const pitchTitle = computed(() => pitchLead.value.slice(0, 24))
const pitchSub = computed(() => {
  const text = descriptionText.value
  if (!text || text.length <= pitchTitle.value.length) return ''
  return text.slice(pitchTitle.value.length).slice(0, 60)
})

// 套餐列表
const plans = computed(() => {
  const skus = activeSkus.value || []
  return skus.map((sku: any) => {
    const id = normalizeSkuId(sku.id)
    const promoPrice = hasSkuPromotionPrice(sku) ? getSkuPromotionPriceAmount(sku) : null
    const price = promoPrice ?? sku.price_amount
    return {
      id,
      raw: sku,
      name: skuDisplayText(sku),
      hint: skuStockText(sku),
      disabled: !isSkuPurchasable(sku),
      selected: normalizeSkuId(selectedSkuId.value) === id,
      priceParts: splitPrice(price),
    }
  })
})

const selectPlan = (plan: { id: number; disabled: boolean }) => {
  if (plan.disabled) return
  selectedSkuId.value = plan.id
}

const selectedPlanName = computed(() => {
  const sku = selectedSku.value
  return sku ? skuDisplayText(sku) : ''
})

// 最终单价（含促销）
const unitPrice = computed(() => {
  const sku = selectedSku.value
  if (sku) {
    return hasSkuPromotionPrice(sku) ? getSkuPromotionPriceAmount(sku) : sku.price_amount
  }
  return hasPromotionPrice(product.value) ? getPromotionPriceAmount(product.value) : product.value?.price_amount
})

const totalParts = computed(() => splitPrice(unitPrice.value))

const onBuyClick = () => {
  noticeOpen.value = true
}

const confirmBuy = () => {
  noticeOpen.value = false
  buyNow()
}
</script>

<style scoped>
.crumb-icon,
.bottom-icon,
.badge-icon,
.feature-icon,
.btn-icon {
  width: 13px;
  height: 13px;
}

.feature-icon {
  color: var(--foreground);
}

.tibo-detail-loading {
  display: grid;
  grid-template-columns: minmax(0, 1.85fr) minmax(290px, 1fr);
  gap: 0;
  padding: 30px 0;
}

.skeleton-block {
  min-height: 420px;
  background: linear-gradient(120deg, var(--card), var(--muted), var(--card));
  background-size: 200% 100%;
  animation: tibo-skeleton 1.6s ease-in-out infinite;
  border-radius: 20px;
  margin: 0 6px;
}

.plan-required {
  margin-top: 8px;
  color: var(--warning);
  font-size: 12px;
}

.summary-warning {
  margin-top: 12px;
  padding: 10px 12px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--warning) 12%, transparent);
  color: var(--warning);
  font-size: 11px;
}

.product-details {
  padding: 30px 0 40px;
}

.product-details h2 {
  font-size: 21px;
  font-weight: 550;
  letter-spacing: -0.6px;
  margin-bottom: 14px;
}

.details-body {
  color: var(--muted-foreground);
  font-size: 13px;
  line-height: 1.9;
}

.details-body :deep(img) {
  border-radius: 12px;
  max-width: 100%;
}

.tibo-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 64px 0;
  color: var(--muted-foreground);
  text-align: center;
}

.empty-icon {
  width: 40px;
  height: 40px;
  opacity: 0.55;
}

@keyframes tibo-skeleton {
  0% {
    background-position: 200% 0;
  }

  100% {
    background-position: -200% 0;
  }
}
</style>
