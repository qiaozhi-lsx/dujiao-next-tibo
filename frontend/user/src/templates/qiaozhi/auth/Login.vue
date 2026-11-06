<template>
  <div ref="rootEl" class="qz-container qz-auth">
    <!-- 返回首页 -->
    <RouterLink to="/" class="qz-pd-back qz-reveal is-pending" :data-reveal="1">
      <ChevronLeft />
      {{ t('auth.login.backHome') }}
    </RouterLink>

    <!-- 卡片：单列登录 -->
    <section class="qz-auth-card qz-reveal is-pending" :data-reveal="1">
      <header class="qz-auth-hero">
        <span class="qz-eyebrow">{{ t('qiaozhi.login.eyebrow') }}</span>
        <h1>
          {{ step === 'totp' ? t('auth.login.totp.title') : t('auth.login.title') }}
        </h1>
        <p>
          {{ step === 'totp' ? t('auth.login.totp.subtitle') : t('qiaozhi.login.subtitle') }}
        </p>
        <small class="qz-auth-brand">{{ brandSiteName }}</small>
      </header>

      <!-- 2FA 分步 -->
      <form
        v-if="step === 'totp'"
        class="qz-ic"
        novalidate
        @submit.prevent="handleVerify2FA"
      >
        <p class="qz-ic-alert" role="status">
          {{ t('auth.login.totp.countdown', { seconds: challengeRemainingSeconds }) }}
        </p>

        <div v-if="totpMode === 'code'" class="qz-ic-field">
          <label for="qz-totp-code">{{ t('auth.login.totp.codeLabel') }}<b>*</b></label>
          <input
            id="qz-totp-code"
            v-model="totpCode"
            inputmode="numeric"
            autocomplete="one-time-code"
            maxlength="6"
            :placeholder="t('auth.login.totp.codePlaceholder')"
            class="qz-auth-input-code"
          />
        </div>

        <div v-else class="qz-ic-field">
          <label for="qz-totp-recovery">{{ t('auth.login.totp.recoveryLabel') }}<b>*</b></label>
          <input
            id="qz-totp-recovery"
            v-model="recoveryCode"
            autocomplete="off"
            :placeholder="t('auth.login.totp.recoveryPlaceholder')"
          />
        </div>

        <p class="qz-ic-hint is-right">
          <a
            href="#"
            @click.prevent="totpMode = totpMode === 'code' ? 'recovery' : 'code'"
          >
            {{ totpMode === 'code' ? t('auth.login.totp.useRecovery') : t('auth.login.totp.useCode') }}
          </a>
        </p>

        <p v-if="error" class="qz-ic-alert is-error" role="alert">{{ error }}</p>

        <button
          type="submit"
          class="qz-xcheck-btn qz-auth-submit"
          :disabled="userAuthStore.loading"
        >
          <Loader2 v-if="userAuthStore.loading" class="qz-ic-spin" />
          <ShieldCheck v-else />
          {{ userAuthStore.loading ? t('auth.login.totp.verifying') : t('auth.login.totp.submit') }}
        </button>

        <p class="qz-ic-hint is-right">
          <a href="#" @click.prevent="cancel2FA">{{ t('auth.login.totp.cancel') }}</a>
        </p>
      </form>

      <!-- 密码登录 -->
      <form
        v-else
        class="qz-ic"
        novalidate
        @submit.prevent="handleLogin"
      >
        <div class="qz-ic-field">
          <label for="qz-login-email">{{ t('auth.login.emailLabel') }}<b>*</b></label>
          <input
            id="qz-login-email"
            v-model="email"
            type="email"
            autocomplete="username"
            required
            :placeholder="t('auth.login.emailPlaceholder')"
            :aria-invalid="formValidation.hasError('email') || undefined"
            :aria-describedby="formValidation.hasError('email') ? 'qz-login-email-err' : undefined"
            @blur="formValidation.touchField('email', email)"
          />
          <p v-if="formValidation.getError('email')" id="qz-login-email-err" class="qz-ic-error">
            {{ formValidation.getError('email') }}
          </p>
        </div>

        <div class="qz-ic-field">
          <label for="qz-login-password">{{ t('auth.login.passwordLabel') }}<b>*</b></label>
          <div class="qz-auth-password">
            <input
              id="qz-login-password"
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              autocomplete="current-password"
              required
              :placeholder="t('auth.login.passwordPlaceholder')"
              :aria-invalid="formValidation.hasError('password') || undefined"
              :aria-describedby="formValidation.hasError('password') ? 'qz-login-password-err' : undefined"
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
          <p v-if="formValidation.getError('password')" id="qz-login-password-err" class="qz-ic-error">
            {{ formValidation.getError('password') }}
          </p>
        </div>

        <div v-if="loginCaptchaEnabled" class="qz-ic-field">
          <label>{{ t('auth.common.captchaLabel') }}<b>*</b></label>
          <ImageCaptcha
            v-if="captchaProvider === 'image'"
            ref="imageCaptchaRef"
            v-model="captchaPayload"
            :disabled="userAuthStore.loading"
            @config-stale="handleCaptchaConfigStale"
          />
          <TurnstileCaptcha
            v-else-if="captchaProvider === 'turnstile'"
            ref="turnstileRef"
            v-model="turnstileToken"
            :site-key="turnstileSiteKey"
          />
        </div>

        <div class="qz-auth-row">
          <label class="qz-auth-check">
            <input v-model="rememberMe" type="checkbox" />
            <span>{{ t('auth.login.rememberMe') }}</span>
          </label>
          <RouterLink
            v-if="emailVerificationEnabled"
            to="/auth/forgot"
            class="qz-auth-link"
          >
            {{ t('auth.login.forgot') }}
          </RouterLink>
        </div>

        <p v-if="info" class="qz-ic-alert qz-auth-info" role="status">{{ info }}</p>
        <p v-if="error" class="qz-ic-alert is-error" role="alert">{{ error }}</p>

        <button
          type="submit"
          class="qz-xcheck-btn qz-auth-submit"
          :disabled="userAuthStore.loading"
        >
          <Loader2 v-if="userAuthStore.loading" class="qz-ic-spin" />
          <LogIn v-else />
          {{ userAuthStore.loading ? t('auth.login.submitting') : t('auth.login.submit') }}
        </button>

        <!-- 第三方登录 -->
        <template v-if="showThirdPartyLogin">
          <div class="qz-auth-divider">
            <span></span>
            <em>{{ t('auth.login.socialOr') }}</em>
            <span></span>
          </div>
          <div class="qz-auth-oauth">
            <div v-if="showTelegramWidget" class="qz-auth-oauth-item">
              <div ref="telegramWidgetRef" class="qz-auth-oauth-center"></div>
              <small>{{ t('auth.login.telegramHint') }}</small>
            </div>
            <div v-else-if="showTelegramOidc" class="qz-auth-oauth-item">
              <button
                type="button"
                class="qz-auth-oauth-btn"
                @click="startTelegramOidc"
              >
                {{ t('auth.login.telegramOidcButton') }}
              </button>
              <small>{{ t('auth.login.telegramOidcHint') }}</small>
            </div>
            <div v-else-if="showMiniAppLoginHint" class="qz-auth-oauth-item">
              <small>
                {{ attemptingMiniAppLogin ? t('auth.login.telegramMiniAppLoggingIn') : t('auth.login.telegramMiniAppHint') }}
              </small>
            </div>
            <div v-if="showGoogleLogin" class="qz-auth-oauth-item">
              <GoogleIdentityButton
                :client-id="googleClientID"
                :locale="googleButtonLocale"
                :ux-mode="googleIdentityUXMode"
                :login-uri="googleRedirectLoginURI"
                :prepare-redirect="prepareGoogleRedirectLogin"
                :disabled="userAuthStore.loading"
                :loading-label="t('auth.login.googleLoading')"
                @credential="handleGoogleCredential"
                @error="handleGoogleScriptError"
              />
              <small>{{ t('auth.login.googleHint') }}</small>
            </div>
          </div>
        </template>

        <div v-if="showTelegramMiniAppEntry" class="qz-auth-oauth-item">
          <button
            type="button"
            class="qz-auth-oauth-btn"
            @click="openTelegramMiniAppEntry"
          >
            {{ t('auth.login.telegramMiniAppEntryAction') }}
          </button>
          <small>{{ t('auth.login.telegramMiniAppEntryHint') }}</small>
        </div>
      </form>

      <footer v-if="registrationEnabled && step !== 'totp'" class="qz-auth-foot">
        <span>{{ t('qiaozhi.login.noAccountPrefix') }}</span>
        <RouterLink to="/auth/register">{{ t('qiaozhi.login.goRegister') }}</RouterLink>
      </footer>
    </section>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, Eye, EyeOff, Loader2, LogIn, ShieldCheck } from 'lucide-vue-next'
import ImageCaptcha from '../../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../../components/captcha/TurnstileCaptcha.vue'
import GoogleIdentityButton from '../../../components/auth/GoogleIdentityButton.vue'
import { useLogin } from '../../../composables/useLogin'
import { setupReveal } from '../utils/reveal'

const { t } = useI18n()

const {
  userAuthStore, brandSiteName,
  email, password, showPassword, rememberMe,
  step, totpMode, totpCode, recoveryCode, challengeRemainingSeconds, handleVerify2FA, cancel2FA,
  error, info, formValidation,
  loginCaptchaEnabled, captchaProvider, captchaPayload, turnstileToken, turnstileSiteKey,
  imageCaptchaRef, turnstileRef, handleCaptchaConfigStale,
  registrationEnabled, emailVerificationEnabled,
  showTelegramWidget, telegramWidgetRef, showTelegramOidc, startTelegramOidc,
  showMiniAppLoginHint, attemptingMiniAppLogin, showTelegramMiniAppEntry, openTelegramMiniAppEntry,
  googleClientID, googleButtonLocale, googleIdentityUXMode, googleRedirectLoginURI,
  prepareGoogleRedirectLogin, showGoogleLogin, showThirdPartyLogin,
  handleGoogleCredential, handleGoogleScriptError,
  handleLogin,
} = useLogin()

// 字符串模板 ref 显式占用，避免 vue-tsc 误报未使用
void imageCaptchaRef
void turnstileRef
void telegramWidgetRef

// —— 入场动画：错峰 reveal ——
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

// step 切换时重新铺一遍 reveal（totp 表单换入）
watch(step, () => {
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
.qz-auth-input-code {
  text-align: center;
  letter-spacing: 0.4em;
  font-weight: 700;
}
.qz-auth-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: -4px;
  font-size: 12px;
  color: var(--muted);
}
.qz-auth-check {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.qz-auth-check input {
  width: 15px;
  height: 15px;
  accent-color: var(--accent);
}
.qz-auth-link {
  color: var(--muted);
  font-weight: 600;
  text-decoration: none;
}
.qz-auth-link:hover {
  color: var(--accent-strong);
}
.qz-auth-submit {
  width: 100%;
}
.qz-auth-info {
  background: var(--accent-soft);
  color: var(--accent-strong);
}
.qz-auth-divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 4px;
  color: var(--muted);
  font-size: 11px;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}
.qz-auth-divider span {
  flex: 1 1 auto;
  height: 1px;
  background: var(--line);
}
.qz-auth-divider em {
  font-style: normal;
  font-weight: 700;
}
.qz-auth-oauth {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.qz-auth-oauth-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.qz-auth-oauth-item small {
  text-align: center;
  font-size: 11px;
  color: var(--muted);
  line-height: 1.5;
}
.qz-auth-oauth-center {
  display: flex;
  justify-content: center;
}
.qz-auth-oauth-btn {
  width: 100%;
  min-height: 46px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 0 16px;
  border: 1px solid var(--line-strong);
  border-radius: 12px;
  background: var(--surface);
  color: var(--text);
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}
.qz-auth-oauth-btn:hover {
  border-color: var(--accent-line);
  background: var(--surface-2);
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
}
</style>
