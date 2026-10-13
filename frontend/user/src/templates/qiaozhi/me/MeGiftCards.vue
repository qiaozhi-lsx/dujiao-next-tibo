<template>
  <div class="j8-me-panel">
    <div class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ t('personalCenter.giftCard.title') }}</h2>
          <small>{{ t('personalCenter.giftCard.subtitle') }}</small>
        </div>
        <div class="j8-me-card-acts">
          <span class="j8-me-pill is-ok">
            <Gift />
            {{ t('personalCenter.tabs.giftCard') }}
          </span>
        </div>
      </div>

      <p v-if="panelAlert" class="j8-ic-alert panel-alert" :class="panelAlertClass" role="alert">
        {{ panelAlert.message }}
      </p>

      <!-- 兑换结果 -->
      <div v-if="lastRedeem" class="j8-ic-pay gift-success">
        <div class="j8-me-card-acts">
          <span class="j8-me-pill is-ok">
            <Check />
            {{ t('personalCenter.giftCard.successTitle') }}
          </span>
        </div>
        <div class="j8-me-kv">
          <div class="j8-me-kv-row">
            <span>{{ t('personalCenter.giftCard.successCode') }}</span>
            <b class="gift-code">{{ String(lastRedeem.gift_card?.code || '-').toUpperCase() }}</b>
          </div>
          <div class="j8-me-kv-row">
            <span>{{ t('personalCenter.giftCard.successAmount') }}</span>
            <b>{{ redeemedAmountText }}</b>
          </div>
          <div class="j8-me-kv-row">
            <span>{{ t('personalCenter.giftCard.successBalance') }}</span>
            <b>{{ currentBalanceText }}</b>
          </div>
        </div>
      </div>

      <!-- 兑换表单 -->
      <form class="j8-ic" novalidate @submit.prevent="submitRedeem">
        <div class="j8-ic-field">
          <label for="j8-gift-code">{{ t('personalCenter.giftCard.codeLabel') }}</label>
          <input
            id="j8-gift-code"
            v-model="redeemForm.code"
            type="text"
            maxlength="80"
            autocomplete="off"
            class="is-code"
            :placeholder="t('personalCenter.giftCard.codePlaceholder')"
          />
        </div>

        <div v-if="redeemCaptchaEnabled" class="j8-ic-field">
          <label>{{ t('auth.common.captchaLabel') }}</label>
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

        <div class="gift-actions">
          <J8Button type="submit" :loading="submitting">
            {{ submitting ? t('personalCenter.giftCard.redeeming') : t('personalCenter.giftCard.redeemButton') }}
          </J8Button>
          <J8Button type="button" variant="secondary" :disabled="submitting" @click="resetForm">
            {{ t('personalCenter.giftCard.resetButton') }}
          </J8Button>
        </div>
      </form>
    </div>
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
import J8Button from '../components/J8Button.vue'

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

const panelAlertClass = computed(() => {
  const level = panelAlert.value?.level
  if (level === 'error') return 'is-error'
  if (level === 'success') return 'is-ok'
  return ''
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
.j8-me-panel .panel-alert {
  margin: 0 0 12px;
}

.j8-me-panel .panel-alert.is-ok {
  background: var(--accent-soft);
  color: var(--accent-strong);
}

.j8-me-panel .gift-success {
  margin-bottom: 14px;
  animation: gift-success-burst 0.45s ease both;
}

@keyframes gift-success-burst {
  0% {
    opacity: 0;
    transform: translateY(8px) scale(0.98);
  }
  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.j8-me-panel .is-code {
  font-family: var(--font-mono);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.j8-me-panel .gift-code {
  font-family: var(--font-mono);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.j8-me-panel .gift-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding-top: 2px;
}
</style>
