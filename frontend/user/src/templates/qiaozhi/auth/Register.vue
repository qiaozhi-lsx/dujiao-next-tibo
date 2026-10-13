<template>
  <div ref="rootEl" class="j8-container j8-auth">
    <RouterLink to="/" class="j8-pd-back j8-reveal is-pending" :data-reveal="1">
      <ChevronLeft />
      {{ t('auth.login.backHome') }}
    </RouterLink>

    <section class="j8-auth-card j8-reveal is-pending" :data-reveal="1">
      <!-- 注册关闭 -->
      <div v-if="!registrationEnabled" class="j8-empty">
        <PackageOpen />
        <b>{{ t('auth.register.registrationDisabled') }}</b>
        <RouterLink to="/auth/login" class="j8-btn-primary">
          {{ t('auth.register.hasAccount') }}
        </RouterLink>
      </div>

      <template v-else>
        <header class="j8-auth-hero">
          <span class="j8-eyebrow">{{ t('qiaozhi.register.eyebrow') }}</span>
          <h1>{{ t('auth.register.title') }}</h1>
          <p>{{ t('qiaozhi.register.subtitle') }}</p>
          <small class="j8-auth-brand">{{ brandSiteName }}</small>
        </header>

        <form class="j8-ic" novalidate @submit.prevent="handleRegister">
          <!-- 邮箱 -->
          <div class="j8-ic-field">
            <label for="qz-reg-email">{{ t('auth.register.emailLabel') }}<b>*</b></label>
            <div v-if="emailDomainSelectionRequired" class="j8-auth-email-row">
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
              <select
                v-model="selectedEmailDomain"
                class="j8-auth-email-domain"
                @change="touchRegistrationEmail"
              >
                <option v-for="domain in allowedEmailDomains" :key="domain" :value="domain">
                  @{{ domain }}
                </option>
              </select>
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
            <p v-if="formValidation.getError('email')" class="j8-ic-error">
              {{ formValidation.getError('email') }}
            </p>
            <p v-else-if="emailDomainSelectionRequired" class="j8-ic-hint">
              {{ t('auth.register.emailDomainSelectHint') }}
            </p>
            <p v-else-if="emailDomainAllowlistEnabled" class="j8-ic-hint">
              {{ allowedEmailDomains.length > 0
                ? t('auth.register.allowedEmailDomainsHint', { domains: allowedEmailDomainsText })
                : t('auth.register.noAllowedEmailDomainsHint') }}
            </p>
          </div>

          <!-- 密码 -->
          <div class="j8-ic-field">
            <label for="qz-reg-password">{{ t('auth.register.passwordLabel') }}<b>*</b></label>
            <div class="j8-auth-password">
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
                class="j8-auth-pass-toggle"
                :aria-label="showPassword ? t('auth.common.hidePassword') : t('auth.common.showPassword')"
                @click="showPassword = !showPassword"
              >
                <EyeOff v-if="showPassword" />
                <Eye v-else />
              </button>
            </div>
            <p v-if="formValidation.getError('password')" class="j8-ic-error">
              {{ formValidation.getError('password') }}
            </p>
            <div
              v-else-if="password"
              class="j8-auth-strength"
              :class="`is-${passwordStrength}`"
            >
              <div class="j8-auth-strength-bars">
                <span></span>
                <span></span>
                <span></span>
              </div>
              <small>{{ t(`formValidation.passwordStrength.${passwordStrength}`) }}</small>
            </div>
          </div>

          <!-- 发码场景验证码 -->
          <div v-if="emailVerificationEnabled && sendCodeCaptchaEnabled" class="j8-ic-field">
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
          <div v-if="emailVerificationEnabled" class="j8-ic-field j8-ic-with-action">
            <label for="qz-reg-code">{{ t('auth.register.codeLabel') }}<b>*</b></label>
            <div class="j8-auth-code-row">
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
                class="j8-auth-code-btn"
                :disabled="sending || countdown > 0"
                @click="handleSendCode"
              >
                <Loader2 v-if="sending" class="j8-ic-spin" />
                <span>
                  {{ countdown > 0
                    ? t('auth.common.countdown', { seconds: countdown })
                    : t('auth.common.sendCode') }}
                </span>
              </button>
            </div>
          </div>

          <!-- 服务条款 -->
          <label class="j8-auth-agree">
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

          <p v-if="error" class="j8-ic-alert is-error" role="alert">{{ error }}</p>

          <button
            type="submit"
            class="j8-xcheck-btn j8-auth-submit"
            :disabled="userAuthStore.loading || !agreed"
          >
            <Loader2 v-if="userAuthStore.loading" class="j8-ic-spin" />
            <UserPlus v-else />
            {{ userAuthStore.loading ? t('auth.register.creating') : t('auth.register.create') }}
          </button>
        </form>

        <footer class="j8-auth-foot">
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
.j8-auth {
  display: flex;
  flex-direction: column;
  gap: 18px;
  max-width: 460px;
}
.j8-auth-card {
  display: flex;
  flex-direction: column;
  gap: 22px;
  padding: 32px 28px;
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--r-card);
  box-shadow: var(--shadow-soft);
}
.j8-auth-hero {
  display: flex;
  flex-direction: column;
  gap: 8px;
  text-align: center;
}
.j8-auth-hero h1 {
  margin: 2px 0 0;
  font-size: 26px;
  font-weight: 900;
  letter-spacing: -0.01em;
  color: var(--text);
}
.j8-auth-hero p {
  margin: 0;
  font-size: 13px;
  color: var(--muted);
  line-height: 1.6;
}
.j8-auth-brand {
  margin-top: 6px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--muted);
}
.j8-auth-email-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(9rem, auto);
  gap: 8px;
}
.j8-auth-email-domain {
  background: var(--surface);
}
.j8-auth-password {
  position: relative;
}
.j8-auth-password input {
  padding-right: 42px !important;
}
.j8-auth-pass-toggle {
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
.j8-auth-pass-toggle:hover {
  color: var(--text);
  background: var(--surface-2);
}
.j8-auth-pass-toggle svg {
  width: 16px;
  height: 16px;
}
.j8-auth-strength {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 4px;
}
.j8-auth-strength-bars {
  display: flex;
  flex: 1 1 auto;
  gap: 4px;
}
.j8-auth-strength-bars span {
  flex: 1 1 0;
  height: 3px;
  border-radius: 999px;
  background: var(--surface-3);
  transition: background 0.15s;
}
.j8-auth-strength.is-weak .j8-auth-strength-bars span:nth-child(1) {
  background: var(--red);
}
.j8-auth-strength.is-medium .j8-auth-strength-bars span:nth-child(-n+2) {
  background: var(--amber);
}
.j8-auth-strength.is-strong .j8-auth-strength-bars span {
  background: var(--accent-strong);
}
.j8-auth-strength small {
  font-size: 11px;
  font-weight: 700;
  color: var(--muted);
}
.j8-auth-strength.is-weak small {
  color: var(--red);
}
.j8-auth-strength.is-medium small {
  color: var(--amber);
}
.j8-auth-strength.is-strong small {
  color: var(--accent-strong);
}
.j8-auth-code-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 8px;
}
.j8-auth-code-btn {
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
.j8-auth-code-btn:hover:not(:disabled) {
  border-color: var(--accent-line);
  background: var(--surface-2);
}
.j8-auth-code-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.j8-auth-code-btn svg {
  width: 14px;
  height: 14px;
}
.j8-auth-agree {
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
.j8-auth-agree input {
  margin-top: 3px;
  width: 15px;
  height: 15px;
  flex: 0 0 auto;
  accent-color: var(--accent);
}
.j8-auth-agree a {
  color: var(--accent-strong);
  font-weight: 700;
  text-decoration: none;
}
.j8-auth-agree a:hover {
  text-decoration: underline;
}
.j8-auth-submit {
  width: 100%;
}
.j8-auth-foot {
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
.j8-auth-foot a {
  color: var(--accent-strong);
  font-weight: 700;
  text-decoration: none;
}
.j8-auth-foot a:hover {
  text-decoration: underline;
}
@media (max-width: 540px) {
  .j8-auth-card {
    padding: 26px 20px;
  }
  .j8-auth-hero h1 {
    font-size: 22px;
  }
  .j8-auth-email-row {
    grid-template-columns: 1fr;
  }
  .j8-auth-code-row {
    grid-template-columns: 1fr;
  }
}
</style>
