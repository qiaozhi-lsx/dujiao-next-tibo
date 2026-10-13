<template>
  <div ref="rootEl" class="j8-container j8-auth">
    <!-- 返回首页 -->
    <RouterLink to="/" class="j8-pd-back j8-reveal is-pending" :data-reveal="1">
      <ChevronLeft />
      {{ t('auth.login.backHome') }}
    </RouterLink>

    <!-- 卡片：单列登录 -->
    <section class="j8-auth-card j8-reveal is-pending" :data-reveal="1">
      <header class="j8-auth-hero">
        <span class="j8-eyebrow">{{ t('qiaozhi.login.eyebrow') }}</span>
        <h1>
          {{ step === 'totp' ? t('auth.login.totp.title') : t('auth.login.title') }}
        </h1>
        <p>
          {{ step === 'totp' ? t('auth.login.totp.subtitle') : t('qiaozhi.login.subtitle') }}
        </p>
        <small class="j8-auth-brand">{{ brandSiteName }}</small>
      </header>

      <!-- 2FA 分步 -->
      <form
        v-if="step === 'totp'"
        class="j8-ic"
        novalidate
        @submit.prevent="handleVerify2FA"
      >
        <p class="j8-ic-alert" role="status">
          {{ t('auth.login.totp.countdown', { seconds: challengeRemainingSeconds }) }}
        </p>

        <div v-if="totpMode === 'code'" class="j8-ic-field">
          <label for="qz-totp-code">{{ t('auth.login.totp.codeLabel') }}<b>*</b></label>
          <input
            id="qz-totp-code"
            v-model="totpCode"
            inputmode="numeric"
            autocomplete="one-time-code"
            maxlength="6"
            :placeholder="t('auth.login.totp.codePlaceholder')"
            class="j8-auth-input-code"
          />
        </div>

        <div v-else class="j8-ic-field">
          <label for="qz-totp-recovery">{{ t('auth.login.totp.recoveryLabel') }}<b>*</b></label>
          <input
            id="qz-totp-recovery"
            v-model="recoveryCode"
            autocomplete="off"
            :placeholder="t('auth.login.totp.recoveryPlaceholder')"
          />
        </div>

        <p class="j8-ic-hint is-right">
          <a
            href="#"
            @click.prevent="totpMode = totpMode === 'code' ? 'recovery' : 'code'"
          >
            {{ totpMode === 'code' ? t('auth.login.totp.useRecovery') : t('auth.login.totp.useCode') }}
          </a>
        </p>

        <p v-if="error" class="j8-ic-alert is-error" role="alert">{{ error }}</p>

        <button
          type="submit"
          class="j8-xcheck-btn j8-auth-submit"
          :disabled="userAuthStore.loading"
        >
          <Loader2 v-if="userAuthStore.loading" class="j8-ic-spin" />
          <ShieldCheck v-else />
          {{ userAuthStore.loading ? t('auth.login.totp.verifying') : t('auth.login.totp.submit') }}
        </button>

        <p class="j8-ic-hint is-right">
          <a href="#" @click.prevent="cancel2FA">{{ t('auth.login.totp.cancel') }}</a>
        </p>
      </form>

      <!-- 密码登录 -->
      <form
        v-else
        class="j8-ic"
        novalidate
        @submit.prevent="handleLogin"
      >
        <div class="j8-ic-field">
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
          <p v-if="formValidation.getError('email')" id="qz-login-email-err" class="j8-ic-error">
            {{ formValidation.getError('email') }}
          </p>
        </div>

        <div class="j8-ic-field">
          <label for="qz-login-password">{{ t('auth.login.passwordLabel') }}<b>*</b></label>
          <div class="j8-auth-password">
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
              class="j8-auth-pass-toggle"
              :aria-label="showPassword ? t('auth.common.hidePassword') : t('auth.common.showPassword')"
              @click="showPassword = !showPassword"
            >
              <EyeOff v-if="showPassword" />
              <Eye v-else />
            </button>
          </div>
          <p v-if="formValidation.getError('password')" id="qz-login-password-err" class="j8-ic-error">
            {{ formValidation.getError('password') }}
          </p>
        </div>

        <div v-if="loginCaptchaEnabled" class="j8-ic-field">
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

        <div class="j8-auth-row">
          <label class="j8-auth-check">
            <input v-model="rememberMe" type="checkbox" />
            <span>{{ t('auth.login.rememberMe') }}</span>
          </label>
          <RouterLink
            v-if="emailVerificationEnabled"
            to="/auth/forgot"
            class="j8-auth-link"
          >
            {{ t('auth.login.forgot') }}
          </RouterLink>
        </div>

        <p v-if="info" class="j8-ic-alert j8-auth-info" role="status">{{ info }}</p>
        <p v-if="error" class="j8-ic-alert is-error" role="alert">{{ error }}</p>

        <button
          type="submit"
          class="j8-xcheck-btn j8-auth-submit"
          :disabled="userAuthStore.loading"
        >
          <Loader2 v-if="userAuthStore.loading" class="j8-ic-spin" />
          <LogIn v-else />
          {{ userAuthStore.loading ? t('auth.login.submitting') : t('auth.login.submit') }}
        </button>

        <!-- 第三方登录 -->
        <template v-if="showThirdPartyLogin">
          <div class="j8-auth-divider">
            <span></span>
            <em>{{ t('auth.login.socialOr') }}</em>
            <span></span>
          </div>
          <div class="j8-auth-oauth">
            <div v-if="showTelegramWidget" class="j8-auth-oauth-item">
              <div ref="telegramWidgetRef" class="j8-auth-oauth-center"></div>
              <small>{{ t('auth.login.telegramHint') }}</small>
            </div>
            <div v-else-if="showTelegramOidc" class="j8-auth-oauth-item">
              <button
                type="button"
                class="j8-auth-oauth-btn"
                @click="startTelegramOidc"
              >
                {{ t('auth.login.telegramOidcButton') }}
              </button>
              <small>{{ t('auth.login.telegramOidcHint') }}</small>
            </div>
            <div v-else-if="showMiniAppLoginHint" class="j8-auth-oauth-item">
              <small>
                {{ attemptingMiniAppLogin ? t('auth.login.telegramMiniAppLoggingIn') : t('auth.login.telegramMiniAppHint') }}
              </small>
            </div>
            <div v-if="showGoogleLogin" class="j8-auth-oauth-item">
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

        <div v-if="showTelegramMiniAppEntry" class="j8-auth-oauth-item">
          <button
            type="button"
            class="j8-auth-oauth-btn"
            @click="openTelegramMiniAppEntry"
          >
            {{ t('auth.login.telegramMiniAppEntryAction') }}
          </button>
          <small>{{ t('auth.login.telegramMiniAppEntryHint') }}</small>
        </div>
      </form>

      <footer v-if="registrationEnabled && step !== 'totp'" class="j8-auth-foot">
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
  font-weight: 700;
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
.j8-auth-input-code {
  text-align: center;
  letter-spacing: 0.4em;
  font-weight: 700;
}
.j8-auth-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: -4px;
  font-size: 12px;
  color: var(--muted);
}
.j8-auth-check {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.j8-auth-check input {
  width: 15px;
  height: 15px;
  accent-color: var(--accent);
}
.j8-auth-link {
  color: var(--muted);
  font-weight: 600;
  text-decoration: none;
}
.j8-auth-link:hover {
  color: var(--accent-strong);
}
.j8-auth-submit {
  width: 100%;
}
.j8-auth-info {
  background: var(--accent-soft);
  color: var(--accent-strong);
}
.j8-auth-divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 4px;
  color: var(--muted);
  font-size: 11px;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}
.j8-auth-divider span {
  flex: 1 1 auto;
  height: 1px;
  background: var(--line);
}
.j8-auth-divider em {
  font-style: normal;
  font-weight: 700;
}
.j8-auth-oauth {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.j8-auth-oauth-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.j8-auth-oauth-item small {
  text-align: center;
  font-size: 11px;
  color: var(--muted);
  line-height: 1.5;
}
.j8-auth-oauth-center {
  display: flex;
  justify-content: center;
}
.j8-auth-oauth-btn {
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
.j8-auth-oauth-btn:hover {
  border-color: var(--accent-line);
  background: var(--surface-2);
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
}
</style>
