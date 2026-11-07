<template>
  <div ref="rootEl" class="qz-container qz-auth">
    <!-- 返回首页 -->
    <RouterLink to="/" class="qz-btn qz-btn--secondary qz-btn--sm qz-auth-back qz-reveal is-pending" :data-reveal="1">
      <ChevronLeft />
      {{ t('auth.login.backHome') }}
    </RouterLink>

    <!-- 卡片：单列登录 -->
    <section class="qz-panel qz-auth-card qz-reveal is-pending" :data-reveal="1">
      <header class="qz-auth-hero">
        <span class="qz-eyebrow">{{ t('qiaozhi.login.eyebrow') }}</span>
        <h1>
          {{ step === 'totp' ? t('auth.login.totp.title') : t('auth.login.title') }}
        </h1>
        <p>
          {{ step === 'totp' ? t('auth.login.totp.subtitle') : t('qiaozhi.login.subtitle') }}
        </p>
        <span class="qz-chip qz-auth-brand">{{ brandSiteName }}</span>
      </header>

      <!-- 2FA 分步 -->
      <form
        v-if="step === 'totp'"
        class="qz-auth-form"
        novalidate
        @submit.prevent="submitVerify2FA"
      >
        <p class="qz-alert is-warning" role="status">
          {{ t('auth.login.totp.countdown', { seconds: challengeRemainingSeconds }) }}
        </p>

        <div v-if="totpMode === 'code'" class="qz-field">
          <label class="qz-field-label" for="qz-totp-code">{{ t('auth.login.totp.codeLabel') }}<b>*</b></label>
          <input
            id="qz-totp-code"
            v-model="totpCode"
            class="qz-field-control qz-auth-code-input"
            inputmode="numeric"
            autocomplete="one-time-code"
            maxlength="6"
            :placeholder="t('auth.login.totp.codePlaceholder')"
          />
        </div>

        <div v-else class="qz-field">
          <label class="qz-field-label" for="qz-totp-recovery">{{ t('auth.login.totp.recoveryLabel') }}<b>*</b></label>
          <input
            id="qz-totp-recovery"
            v-model="recoveryCode"
            class="qz-field-control"
            autocomplete="off"
            :placeholder="t('auth.login.totp.recoveryPlaceholder')"
          />
        </div>

        <p class="qz-auth-switch">
          <button type="button" class="qz-auth-link" @click="totpMode = totpMode === 'code' ? 'recovery' : 'code'">
            {{ totpMode === 'code' ? t('auth.login.totp.useRecovery') : t('auth.login.totp.useCode') }}
          </button>
        </p>

        <p v-if="error" class="qz-alert is-danger" role="alert">{{ error }}</p>

        <QzButton type="submit" variant="primary" class="qz-btn--block" :loading="userAuthStore.loading">
          <ShieldCheck v-if="!userAuthStore.loading" />
          {{ userAuthStore.loading ? t('auth.login.totp.verifying') : t('auth.login.totp.submit') }}
        </QzButton>

        <p class="qz-auth-switch">
          <button type="button" class="qz-auth-link" @click="cancel2FA">{{ t('auth.login.totp.cancel') }}</button>
        </p>
      </form>

      <!-- 密码登录 -->
      <form
        v-else
        class="qz-auth-form"
        novalidate
        @submit.prevent="submitLogin"
      >
        <div class="qz-field">
          <label class="qz-field-label" for="qz-login-email">{{ t('auth.login.emailLabel') }}<b>*</b></label>
          <input
            id="qz-login-email"
            v-model="email"
            class="qz-field-control"
            type="email"
            autocomplete="username"
            required
            :placeholder="t('auth.login.emailPlaceholder')"
            :aria-invalid="formValidation.hasError('email') || undefined"
            :aria-describedby="formValidation.hasError('email') ? 'qz-login-email-err' : undefined"
            @blur="formValidation.touchField('email', email)"
          />
          <p v-if="formValidation.getError('email')" id="qz-login-email-err" class="qz-field-error">
            {{ formValidation.getError('email') }}
          </p>
        </div>

        <div class="qz-field">
          <label class="qz-field-label" for="qz-login-password">{{ t('auth.login.passwordLabel') }}<b>*</b></label>
          <div class="qz-field-control qz-auth-password">
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
          <p v-if="formValidation.getError('password')" id="qz-login-password-err" class="qz-field-error">
            {{ formValidation.getError('password') }}
          </p>
        </div>

        <div v-if="loginCaptchaEnabled" class="qz-field" role="group" aria-labelledby="qz-login-captcha-label">
          <span id="qz-login-captcha-label" class="qz-field-label">{{ t('auth.common.captchaLabel') }}<b>*</b></span>
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
            :theme="theme"
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

        <p v-if="info" class="qz-alert is-info" role="status">{{ info }}</p>
        <p v-if="error" class="qz-alert is-danger" role="alert">{{ error }}</p>

        <QzButton type="submit" variant="primary" class="qz-btn--block" :loading="userAuthStore.loading">
          <LogIn v-if="!userAuthStore.loading" />
          {{ userAuthStore.loading ? t('auth.login.submitting') : t('auth.login.submit') }}
        </QzButton>

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
                class="qz-btn qz-btn--secondary qz-btn--block"
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
            class="qz-btn qz-btn--secondary qz-btn--block"
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
import { ChevronLeft, Eye, EyeOff, LogIn, ShieldCheck } from 'lucide-vue-next'
import ImageCaptcha from '../../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../../components/captcha/TurnstileCaptcha.vue'
import GoogleIdentityButton from '../../../components/auth/GoogleIdentityButton.vue'
import { useLogin } from '../../../composables/useLogin'
import { useTheme } from '../../../utils/theme'
import { setupReveal } from '../utils/reveal'
import QzButton from '../components/QzButton.vue'

const { t } = useI18n()
const { theme } = useTheme()

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

// 提交钮在 loading 时只是 aria-disabled（保留焦点），回车隐式提交要在这里拦住，避免重复请求
const submitLogin = () => {
  if (userAuthStore.loading) return
  void handleLogin()
}
const submitVerify2FA = () => {
  if (userAuthStore.loading) return
  void handleVerify2FA()
}

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
  gap: var(--qz-space-4);
  max-width: 460px;
}

.qz-auth-back {
  align-self: flex-start;
}

/* .qz-panel.qz-auth-card 写成 (0,3,0)：base.css 的 .qz-panel 是 (0,2,0)，串上页面类靠特指度覆盖它的内边距，
   不依赖样式注入顺序。 */
.qz-panel.qz-auth-card {
  position: static;
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-6);
  padding: var(--qz-space-8) var(--qz-space-6);
  overflow: visible;
}

/* ---------- 页头（同样串上 .qz-panel 提到 (0,3,0)） ---------- */
.qz-panel > .qz-auth-hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--qz-space-2);
  padding: 0;
  border-bottom: 0;
  text-align: center;
}

.qz-auth-hero h1 {
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-2xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
}

.qz-auth-hero p {
  margin: 0;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
}

.qz-auth-brand {
  margin-top: var(--qz-space-1);
}

/* ---------- 表单 ---------- */
.qz-auth-form {
  display: grid;
  gap: var(--qz-space-4);
}

.qz-auth-code-input {
  font-weight: 600;
  letter-spacing: 0.4em;
  text-align: center;
}

/* 密码框：.qz-field-control 做外壳，里面是裸 input + 显隐按钮 */
.qz-field-control.qz-auth-password {
  padding-right: var(--qz-space-1);
}

.qz-auth-pass-toggle {
  display: grid;
  flex: none;
  place-items: center;
  width: 36px;
  height: 36px;
  padding: 0;
  border: 0;
  border-radius: var(--qz-radius-pill);
  background: transparent;
  color: var(--qz-text-muted);
  cursor: pointer;
  transition:
    color var(--qz-duration-fast) var(--qz-ease-standard),
    background-color var(--qz-duration-fast) var(--qz-ease-standard);
}

.qz-field-control .qz-auth-pass-toggle svg {
  width: 18px;
  height: 18px;
  color: inherit;
}

@media (hover: hover) and (pointer: fine) {
  .qz-auth-pass-toggle:hover {
    color: var(--qz-text);
    background: var(--qz-surface-muted);
  }
}

.qz-auth-pass-toggle:active {
  color: var(--qz-text);
  background: var(--qz-surface-muted);
}

/* 记住我 / 忘记密码 */
.qz-auth-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--qz-space-2) var(--qz-space-3);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-auth-check {
  display: inline-flex;
  align-items: center;
  gap: var(--qz-space-2);
  cursor: pointer;
}

.qz-auth-check input {
  width: 16px;
  height: 16px;
  margin: 0;
  accent-color: var(--qz-accent);
}

/* 文字型链接 / 按钮：忘记密码、切换恢复码、取消 */
.qz-auth-link {
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--qz-text-secondary);
  font: inherit;
  font-size: var(--qz-text-xs);
  font-weight: 600;
  line-height: var(--qz-leading-body);
  text-decoration: underline;
  text-underline-offset: 3px;
  cursor: pointer;
  transition: color var(--qz-duration-fast) var(--qz-ease-standard);
}

.qz-auth-link:hover {
  color: var(--qz-text);
}

.qz-auth-switch {
  margin: 0;
  text-align: right;
}

/* 提示条（倒计时 / 错误 / 信息）用 base.css 的 .qz-alert.is-warning / .is-danger / .is-info */

/* ---------- 第三方登录 ---------- */
.qz-auth-divider {
  display: flex;
  align-items: center;
  gap: var(--qz-space-3);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.qz-auth-divider span {
  flex: 1 1 auto;
  height: 1px;
  background: var(--qz-border);
}

.qz-auth-divider em {
  font-style: normal;
}

.qz-auth-oauth {
  display: grid;
  gap: var(--qz-space-4);
}

.qz-auth-oauth-item {
  display: grid;
  gap: var(--qz-space-2);
}

.qz-auth-oauth-item small {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  text-align: center;
}

.qz-auth-oauth-center {
  display: flex;
  justify-content: center;
}

/* ---------- 页脚 ---------- */
.qz-auth-foot {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: var(--qz-space-1) var(--qz-space-2);
  padding-top: var(--qz-space-5);
  border-top: 1px solid var(--qz-border);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
}

.qz-auth-foot a {
  color: var(--qz-text);
  font-weight: 600;
  text-decoration: underline;
  text-underline-offset: 3px;
}

@media (max-width: 760px) {
  .qz-panel.qz-auth-card {
    gap: var(--qz-space-5);
    padding: var(--qz-space-6) var(--qz-space-5);
  }
  /* 与 base.css 的 .qz-page-head h1 同步降一档：手机端全站页 h1 统一为 xl */
  .qz-auth-hero h1 {
    font-size: var(--qz-text-xl);
  }
}
</style>
