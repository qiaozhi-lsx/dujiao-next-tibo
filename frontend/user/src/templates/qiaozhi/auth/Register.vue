<template>
  <div ref="rootEl" class="qz-container qz-auth">
    <RouterLink to="/" class="qz-pd-back qz-reveal is-pending" :data-reveal="1">
      <ChevronLeft />
      {{ t('auth.login.backHome') }}
    </RouterLink>

    <section class="qz-auth-card qz-reveal is-pending" :data-reveal="1">
      <!-- 注册关闭 -->
      <div v-if="!registrationEnabled" class="qz-empty">
        <PackageOpen />
        <b>{{ t('auth.register.registrationDisabled') }}</b>
        <RouterLink to="/auth/login" class="qz-btn-primary">
          {{ t('auth.register.hasAccount') }}
        </RouterLink>
      </div>

      <template v-else>
        <header class="qz-auth-hero">
          <span class="qz-eyebrow">{{ t('qiaozhi.register.eyebrow') }}</span>
          <h1>{{ t('auth.register.title') }}</h1>
          <p>{{ t('qiaozhi.register.subtitle') }}</p>
          <small class="qz-auth-brand">{{ brandSiteName }}</small>
        </header>

        <form class="qz-ic" novalidate @submit.prevent="handleRegister">
          <!-- 邮箱 -->
          <div class="qz-ic-field">
            <label for="qz-reg-email">{{ t('auth.register.emailLabel') }}<b>*</b></label>
            <div v-if="emailDomainSelectionRequired" class="qz-auth-email-row">
              <input
                id="qz-reg-email"
                v-model="emailLocalPart"
                type="text"
                autocomplete="username"
                required
                :placeholder="t('auth.register.emailLocalPlaceholder')"
                :aria-invalid="formValidation.hasError('email') || undefined"
                @blur="touchRegistrationEmail"
              />
              <QzSelect
                v-model="selectedEmailDomain"
                class="qz-auth-email-domain"
                :items="allowedEmailDomains.map((domain: string) => ({ value: domain, label: '@' + domain }))"
                @change="touchRegistrationEmail"
              />
            </div>
            <input
              v-else
              id="qz-reg-email"
              v-model="email"
              type="email"
              autocomplete="username"
              required
              :placeholder="t('auth.register.emailPlaceholder')"
              :aria-invalid="formValidation.hasError('email') || undefined"
              @blur="touchRegistrationEmail"
            />
            <p v-if="formValidation.getError('email')" class="qz-ic-error">
              {{ formValidation.getError('email') }}
            </p>
            <p v-else-if="emailDomainSelectionRequired" class="qz-ic-hint">
              {{ t('auth.register.emailDomainSelectHint') }}
            </p>
            <p v-else-if="emailDomainAllowlistEnabled" class="qz-ic-hint">
              {{ allowedEmailDomains.length > 0
                ? t('auth.register.allowedEmailDomainsHint', { domains: allowedEmailDomainsText })
                : t('auth.register.noAllowedEmailDomainsHint') }}
            </p>
          </div>

          <!-- 密码 -->
          <div class="qz-ic-field">
            <label for="qz-reg-password">{{ t('auth.register.passwordLabel') }}<b>*</b></label>
            <div class="qz-auth-password">
              <input
                id="qz-reg-password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="new-password"
                required
                :placeholder="t('auth.register.passwordPlaceholder')"
                :aria-invalid="formValidation.hasError('password') || undefined"
                @blur="formValidation.touchField('password', password)"
              />
              <button
                type="button"
                class="qz-auth-pass-toggle"
                :aria-label="showPassword ? t('auth.common.hidePassword') : t('auth.common.showPassword')"
                @click="showPassword = !showPassword"
              >
                <EyeOff v-if="showPassword" />
                <Eye v-else />
              </button>
            </div>
            <p v-if="formValidation.getError('password')" class="qz-ic-error">
              {{ formValidation.getError('password') }}
            </p>
            <div
              v-else-if="password"
              class="qz-auth-strength"
              :class="`is-${passwordStrength}`"
            >
              <div class="qz-auth-strength-bars">
                <span></span>
                <span></span>
                <span></span>
              </div>
              <small>{{ t(`formValidation.passwordStrength.${passwordStrength}`) }}</small>
            </div>
          </div>

          <!-- 发码场景验证码 -->
          <div v-if="emailVerificationEnabled && sendCodeCaptchaEnabled" class="qz-ic-field">
            <label>{{ t('auth.common.captchaLabel') }}<b>*</b></label>
            <ImageCaptcha
              v-if="captchaProvider === 'image'"
              ref="imageCaptchaRef"
              v-model="captchaPayload"
              :disabled="sending || countdown > 0"
              @config-stale="handleCaptchaConfigStale"
            />
            <TurnstileCaptcha
              v-else-if="captchaProvider === 'turnstile'"
              ref="turnstileRef"
              v-model="turnstileToken"
              :site-key="turnstileSiteKey"
            />
          </div>

          <!-- 邮箱验证码 -->
          <div v-if="emailVerificationEnabled" class="qz-ic-field qz-ic-with-action">
            <label for="qz-reg-code">{{ t('auth.register.codeLabel') }}<b>*</b></label>
            <div class="qz-auth-code-row">
              <input
                id="qz-reg-code"
                v-model="code"
                type="text"
                autocomplete="one-time-code"
                required
                :placeholder="t('auth.register.codePlaceholder')"
              />
              <button
                type="button"
                class="qz-auth-code-btn"
                :disabled="sending || countdown > 0"
                @click="handleSendCode"
              >
                <Loader2 v-if="sending" class="qz-ic-spin" />
                <span>
                  {{ countdown > 0
                    ? t('auth.common.countdown', { seconds: countdown })
                    : t('auth.common.sendCode') }}
                </span>
              </button>
            </div>
          </div>

          <!-- 服务条款 -->
          <label class="qz-auth-agree">
            <input v-model="agreed" type="checkbox" />
            <span>
              {{ t('auth.register.agreementPrefix') }}
              <RouterLink to="/privacy" target="_blank" rel="noopener noreferrer">
                {{ t('footer.privacy') }}
              </RouterLink>
              {{ t('auth.register.agreementAnd') }}
              <RouterLink to="/terms" target="_blank" rel="noopener noreferrer">
                {{ t('footer.terms') }}
              </RouterLink>
            </span>
          </label>

          <p v-if="error" class="qz-ic-alert is-error" role="alert">{{ error }}</p>

          <button
            type="submit"
            class="qz-xcheck-btn qz-auth-submit"
            :disabled="userAuthStore.loading || !agreed"
          >
            <Loader2 v-if="userAuthStore.loading" class="qz-ic-spin" />
            <UserPlus v-else />
            {{ userAuthStore.loading ? t('auth.register.creating') : t('auth.register.create') }}
          </button>
        </form>

        <footer class="qz-auth-foot">
          <span>{{ t('qiaozhi.register.hasAccountPrefix') }}</span>
          <RouterLink to="/auth/login">{{ t('qiaozhi.register.goLogin') }}</RouterLink>
        </footer>
      </template>
    </section>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, Eye, EyeOff, Loader2, PackageOpen, UserPlus } from 'lucide-vue-next'
import ImageCaptcha from '../../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../../components/captcha/TurnstileCaptcha.vue'
import { useRegister } from '../../../composables/useRegister'
import { setupReveal } from '../utils/reveal'
import QzSelect from '../components/QzSelect.vue'

const { t } = useI18n()

const {
  userAuthStore, brandSiteName,
  email, emailLocalPart, selectedEmailDomain, password, showPassword, code, agreed,
  passwordStrength, error, sending, countdown,
  captchaPayload, turnstileToken, imageCaptchaRef, turnstileRef,
  captchaProvider, sendCodeCaptchaEnabled, turnstileSiteKey,
  registrationEnabled, emailVerificationEnabled,
  emailDomainAllowlistEnabled, allowedEmailDomains, allowedEmailDomainsText, emailDomainSelectionRequired,
  touchRegistrationEmail, formValidation, handleCaptchaConfigStale, handleSendCode, handleRegister,
} = useRegister()

void imageCaptchaRef
void turnstileRef

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
</script>

<style scoped>
.qz-auth {
  display: flex;
  flex-direction: column;
  gap: 18px;
  max-width: 460px;
}
.qz-auth-card {
  display: flex;
  flex-direction: column;
  gap: 22px;
  padding: 32px 28px;
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--r-card);
  box-shadow: var(--shadow-soft);
}
.qz-auth-hero {
  display: flex;
  flex-direction: column;
  gap: 8px;
  text-align: center;
}
.qz-auth-hero h1 {
  margin: 2px 0 0;
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--text);
}
.qz-auth-hero p {
  margin: 0;
  font-size: 13px;
  color: var(--muted);
  line-height: 1.6;
}
.qz-auth-brand {
  margin-top: 6px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--muted);
}
.qz-auth-email-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(9rem, auto);
  gap: 8px;
}
.qz-auth-email-domain {
  background: var(--surface);
}
.qz-auth-password {
  position: relative;
}
.qz-auth-password input {
  padding-right: 42px !important;
}
.qz-auth-pass-toggle {
  position: absolute;
  top: 50%;
  right: 10px;
  transform: translateY(-50%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border: 0;
  background: transparent;
  border-radius: 8px;
  color: var(--muted);
  cursor: pointer;
  transition: color 0.15s, background 0.15s;
}
.qz-auth-pass-toggle:hover {
  color: var(--text);
  background: var(--surface-2);
}
.qz-auth-pass-toggle svg {
  width: 16px;
  height: 16px;
}
.qz-auth-strength {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 4px;
}
.qz-auth-strength-bars {
  display: flex;
  flex: 1 1 auto;
  gap: 4px;
}
.qz-auth-strength-bars span {
  flex: 1 1 0;
  height: 3px;
  border-radius: 999px;
  background: var(--surface-3);
  transition: background 0.15s;
}
.qz-auth-strength.is-weak .qz-auth-strength-bars span:nth-child(1) {
  background: var(--red);
}
.qz-auth-strength.is-medium .qz-auth-strength-bars span:nth-child(-n+2) {
  background: var(--amber);
}
.qz-auth-strength.is-strong .qz-auth-strength-bars span {
  background: var(--accent-strong);
}
.qz-auth-strength small {
  font-size: 11px;
  font-weight: 700;
  color: var(--muted);
}
.qz-auth-strength.is-weak small {
  color: var(--red);
}
.qz-auth-strength.is-medium small {
  color: var(--amber);
}
.qz-auth-strength.is-strong small {
  color: var(--accent-strong);
}
.qz-auth-code-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 8px;
}
.qz-auth-code-btn {
  min-height: 46px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 0 16px;
  border: 1px solid var(--line-strong);
  border-radius: 12px;
  background: var(--surface);
  color: var(--text);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
  transition: border-color 0.15s, background 0.15s, opacity 0.15s;
}
.qz-auth-code-btn:hover:not(:disabled) {
  border-color: var(--accent-line);
  background: var(--surface-2);
}
.qz-auth-code-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.qz-auth-code-btn svg {
  width: 14px;
  height: 14px;
}
.qz-auth-agree {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 14px;
  background: var(--surface-2);
  border: 1px solid var(--line);
  border-radius: 12px;
  font-size: 13px;
  color: var(--muted);
  line-height: 1.6;
  cursor: pointer;
}
.qz-auth-agree input {
  margin-top: 3px;
  width: 15px;
  height: 15px;
  flex: 0 0 auto;
  accent-color: var(--accent);
}
.qz-auth-agree a {
  color: var(--accent-strong);
  font-weight: 700;
  text-decoration: none;
}
.qz-auth-agree a:hover {
  text-decoration: underline;
}
.qz-auth-submit {
  width: 100%;
}
.qz-auth-foot {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 6px;
  margin-top: 4px;
  padding-top: 18px;
  border-top: 1px solid var(--line);
  font-size: 13px;
  color: var(--muted);
}
.qz-auth-foot a {
  color: var(--accent-strong);
  font-weight: 700;
  text-decoration: none;
}
.qz-auth-foot a:hover {
  text-decoration: underline;
}
@media (max-width: 540px) {
  .qz-auth-card {
    padding: 26px 20px;
  }
  .qz-auth-hero h1 {
    font-size: 22px;
  }
  .qz-auth-email-row {
    grid-template-columns: 1fr;
  }
  .qz-auth-code-row {
    grid-template-columns: 1fr;
  }
}
</style>
