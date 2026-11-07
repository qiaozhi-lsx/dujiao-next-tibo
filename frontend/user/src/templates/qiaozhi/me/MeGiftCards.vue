<template>
  <div class="qz-me-panel">
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.giftCard.title') }}</h2>
          <small>{{ t('personalCenter.giftCard.subtitle') }}</small>
        </div>
        <div class="qz-me-acts">
          <span class="qz-chip">
            <Gift aria-hidden="true" />
            {{ t('personalCenter.tabs.giftCard') }}
          </span>
        </div>
      </div>

      <p v-if="panelAlert" class="qz-alert" :class="panelAlertClass" role="alert">
        {{ panelAlert.message }}
      </p>

      <!-- 兑换结果 -->
      <div v-if="lastRedeem" class="qz-me-sub qz-me-gift-success">
        <div class="qz-me-acts qz-me-gift-success-head">
          <span class="qz-badge is-success">
            <Check aria-hidden="true" />
            {{ t('personalCenter.giftCard.successTitle') }}
          </span>
        </div>
        <div class="qz-me-kv">
          <div class="qz-me-kv-row">
            <span>{{ t('personalCenter.giftCard.successCode') }}</span>
            <b class="qz-me-gift-code">{{ String(lastRedeem.gift_card?.code || '-').toUpperCase() }}</b>
          </div>
          <div class="qz-me-kv-row">
            <span>{{ t('personalCenter.giftCard.successAmount') }}</span>
            <b>{{ redeemedAmountText }}</b>
          </div>
          <div class="qz-me-kv-row">
            <span>{{ t('personalCenter.giftCard.successBalance') }}</span>
            <b>{{ currentBalanceText }}</b>
          </div>
        </div>
      </div>

      <!-- 兑换表单 -->
      <form class="qz-me-form" novalidate @submit.prevent="submitRedeem">
        <div class="qz-field">
          <label class="qz-field-label" for="qz-gift-code">{{ t('personalCenter.giftCard.codeLabel') }}</label>
          <input
            id="qz-gift-code"
            v-model="redeemForm.code"
            type="text"
            maxlength="80"
            autocomplete="off"
            class="qz-field-control qz-me-gift-input"
            :placeholder="t('personalCenter.giftCard.codePlaceholder')"
          />
        </div>

        <div v-if="redeemCaptchaEnabled" class="qz-field" role="group" aria-labelledby="qz-gift-captcha-label">
          <span id="qz-gift-captcha-label" class="qz-field-label">{{ t('auth.common.captchaLabel') }}</span>
          <ImageCaptcha
            v-if="captchaProvider === 'image'"
            ref="imageCaptchaRef"
            v-model="captchaPayload"
            :disabled="submitting"
            @config-stale="handleCaptchaConfigStale"
          />
          <TurnstileCaptcha
            v-else-if="captchaProvider === 'turnstile'"
            ref="turnstileRef"
            v-model="turnstileToken"
            :site-key="turnstileSiteKey"
          />
        </div>

        <div class="qz-me-acts qz-me-form-acts">
          <QzButton type="submit" :loading="submitting">
            {{ submitting ? t('personalCenter.giftCard.redeeming') : t('personalCenter.giftCard.redeemButton') }}
          </QzButton>
          <QzButton type="button" variant="secondary" :disabled="submitting" @click="resetForm">
            {{ t('personalCenter.giftCard.resetButton') }}
          </QzButton>
        </div>
      </form>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Check, Gift } from 'lucide-vue-next'
import { giftCardAPI, type CaptchaPayload, type GiftCardRedeemResult } from '../../../api'
import { useAppStore } from '../../../stores/app'
import type { PageAlert } from '../../../utils/alerts'
import ImageCaptcha from '../../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../../components/captcha/TurnstileCaptcha.vue'
import QzButton from '../components/QzButton.vue'

const { t } = useI18n()
const appStore = useAppStore()

const redeemForm = reactive({
  code: '',
})
const submitting = ref(false)
const panelAlert = ref<PageAlert | null>(null)
const lastRedeem = ref<GiftCardRedeemResult | null>(null)

const captchaPayload = ref<CaptchaPayload>({})
const turnstileToken = ref('')
const imageCaptchaRef = ref<{ refresh: () => void } | null>(null)
const turnstileRef = ref<{ reset: () => void } | null>(null)

const captchaConfig = computed(() => appStore.config?.captcha || null)
const captchaProvider = computed(() => String(captchaConfig.value?.provider || 'none'))
const redeemCaptchaEnabled = computed(() => {
  return !!captchaConfig.value?.scenes?.gift_card_redeem && captchaProvider.value !== 'none'
})
const turnstileSiteKey = computed(() => String(captchaConfig.value?.turnstile?.site_key || ''))

// 提示条 tone（base .qz-alert）：success → is-success、warning → is-warning、error → is-danger
const panelAlertClass = computed(() => {
  const level = panelAlert.value?.level
  if (level === 'error') return 'is-danger'
  if (level === 'success') return 'is-success'
  return 'is-warning'
})

const redeemedAmountText = computed(() => {
  const rawAmount = String(lastRedeem.value?.wallet_delta || lastRedeem.value?.gift_card?.amount || '').trim()
  const currency = String(lastRedeem.value?.gift_card?.currency || appStore.config?.currency || 'CNY').trim()
  if (!rawAmount) return '-'
  return `${rawAmount} ${currency}`
})

const currentBalanceText = computed(() => {
  const balance = String(lastRedeem.value?.wallet?.balance || '').trim()
  const currency = String(lastRedeem.value?.gift_card?.currency || appStore.config?.currency || 'CNY').trim()
  if (!balance) return '-'
  return `${balance} ${currency}`
})

const getCaptchaPayload = (): CaptchaPayload | undefined => {
  if (!redeemCaptchaEnabled.value) return undefined
  if (captchaProvider.value === 'image') {
    return {
      captcha_id: captchaPayload.value.captcha_id || '',
      captcha_code: captchaPayload.value.captcha_code || '',
    }
  }
  if (captchaProvider.value === 'turnstile') {
    return {
      turnstile_token: turnstileToken.value || '',
    }
  }
  return undefined
}

const resetCaptcha = () => {
  captchaPayload.value = {}
  turnstileToken.value = ''
  imageCaptchaRef.value?.refresh()
  turnstileRef.value?.reset()
}

const resetForm = () => {
  redeemForm.code = ''
  panelAlert.value = null
  lastRedeem.value = null
  resetCaptcha()
}

const handleCaptchaConfigStale = async () => {
  await appStore.loadConfig(true)
  captchaPayload.value = {}
  turnstileToken.value = ''
}

const ensureCaptchaPassed = () => {
  if (!redeemCaptchaEnabled.value) return true
  if (captchaProvider.value === 'image') {
    return Boolean(captchaPayload.value.captcha_id && captchaPayload.value.captcha_code)
  }
  if (captchaProvider.value === 'turnstile') {
    return Boolean(turnstileToken.value)
  }
  return false
}

const submitRedeem = async () => {
  panelAlert.value = null
  const code = String(redeemForm.code || '').trim().toUpperCase()
  if (!code) {
    panelAlert.value = {
      level: 'warning',
      message: t('personalCenter.giftCard.errors.codeRequired'),
    }
    return
  }
  if (!ensureCaptchaPassed()) {
    panelAlert.value = {
      level: 'warning',
      message: t('auth.common.captchaRequired'),
    }
    return
  }

  submitting.value = true
  try {
    const response = await giftCardAPI.redeem({
      code,
      captcha_payload: getCaptchaPayload(),
    })
    const payload = response.data.data || ({} as GiftCardRedeemResult)
    lastRedeem.value = payload
    panelAlert.value = {
      level: 'success',
      message: t('personalCenter.giftCard.redeemSuccess', {
        amount: String(payload.wallet_delta || payload.gift_card?.amount || ''),
        currency: String(payload.gift_card?.currency || appStore.config?.currency || 'CNY'),
      }),
    }
    redeemForm.code = ''
    resetCaptcha()
  } catch (err: any) {
    panelAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.giftCard.errors.redeemFailed'),
    }
    if (captchaProvider.value === 'image') {
      imageCaptchaRef.value?.refresh()
    }
    if (captchaProvider.value === 'turnstile') {
      turnstileRef.value?.reset()
      turnstileToken.value = ''
    }
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  if (!appStore.config) {
    void appStore.loadConfig()
  }
})
</script>

<style scoped>
/* 礼品卡面板：区块 .qz-panel + .qz-section-head、提示条 .qz-alert、表单 .qz-field 三件套、结果 .qz-badge、按钮 QzButton
   来自 styles/base.css；.qz-me-panel / .qz-me-sec / .qz-me-acts(.qz-me-form-acts) / .qz-me-sub / .qz-me-kv(-row) / .qz-me-form
   来自 PersonalCenter.vue 的 me 共享块。这里只剩兑换结果的入场、结果头左对齐与卡密的等宽大写。
   .qz-me-acts.qz-me-gift-success-head 写成 (0,3,0)，压过共享块的 .qz-me-acts (0,2,0)。 */

/* ---------- 兑换结果：.qz-me-sub 入场 ---------- */
.qz-me-gift-success {
  animation: qz-me-gift-in var(--qz-duration-considered) var(--qz-ease-enter) both;
}

@keyframes qz-me-gift-in {
  from {
    opacity: 0;
    transform: translateY(var(--qz-space-2)) scale(0.98);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

.qz-me-acts.qz-me-gift-success-head {
  justify-content: flex-start;
}

.qz-me-gift-code {
  font-family: var(--qz-font-mono);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

/* ---------- 兑换表单：卡密输入等宽大写 ---------- */
.qz-field-control.qz-me-gift-input {
  font-family: var(--qz-font-mono);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

@media (prefers-reduced-motion: reduce) {
  .qz-me-gift-success {
    animation: none;
  }
}
</style>
