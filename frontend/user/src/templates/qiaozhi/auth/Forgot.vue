<template>
  <div ref="rootEl" class="qz-container qz-auth">
    <!-- 返回首页 -->
    <RouterLink to="/" class="qz-btn qz-btn--secondary qz-btn--sm qz-auth-back qz-reveal is-pending">
      <ChevronLeft />
      {{ t('auth.login.backHome') }}
    </RouterLink>

    <!-- 卡片：找回密码 -->
    <section class="qz-panel qz-auth-card qz-reveal is-pending">
      <header class="qz-auth-hero">
        <span class="qz-eyebrow">{{ t('qiaozhi.forgot.eyebrow') }}</span>
        <h1>{{ t('auth.forgot.title') }}</h1>
        <p>{{ t('auth.forgot.subtitle') }}</p>
        <span class="qz-chip qz-auth-brand">{{ brandSiteName }}</span>
      </header>

      <!-- 邮箱验证关闭：无法自助重置，只给回登录的出口 -->
      <div v-if="!emailVerificationEnabled" class="qz-auth-form">
        <p class="qz-alert is-danger" role="alert">{{ t('auth.forgot.disabled') }}</p>
        <RouterLink to="/auth/login" class="qz-btn qz-btn--secondary qz-btn--block">
          {{ t('auth.forgot.backLogin') }}
        </RouterLink>
      </div>

      <!-- 重置表单：原生 required 校验与 classic 一致 -->
      <form v-else class="qz-auth-form" @submit.prevent="submitReset">
        <div class="qz-field">
          <label class="qz-field-label" for="qz-forgot-email">{{ t('auth.forgot.emailLabel') }}<b>*</b></label>
          <input
            id="qz-forgot-email"
            v-model="email"
            class="qz-field-control"
            type="email"
            autocomplete="email"
            required
            :placeholder="t('auth.forgot.emailPlaceholder')"
          />
        </div>

        <!-- 发码场景验证码 -->
        <div v-if="sendCodeCaptchaEnabled" class="qz-field" role="group" aria-labelledby="qz-forgot-captcha-label">
          <span id="qz-forgot-captcha-label" class="qz-field-label">{{ t('auth.common.captchaLabel') }}<b>*</b></span>
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
            :theme="theme"
          />
        </div>

        <!-- 邮箱验证码 + 发送 -->
        <div class="qz-field">
          <label class="qz-field-label" for="qz-forgot-code">{{ t('auth.forgot.codeLabel') }}<b>*</b></label>
          <div class="qz-auth-code-row">
            <input
              id="qz-forgot-code"
              v-model="code"
              class="qz-field-control"
              type="text"
              autocomplete="one-time-code"
              required
              :placeholder="t('auth.forgot.codePlaceholder')"
            />
            <QzButton
              type="button"
              variant="secondary"
              :loading="sending"
              :disabled="countdown > 0"
              @click="sendCode"
            >
              {{ countdown > 0
                ? t('auth.common.countdown', { seconds: countdown })
                : t('auth.common.sendCode') }}
            </QzButton>
          </div>
        </div>

        <div class="qz-field">
          <label class="qz-field-label" for="qz-forgot-password">{{ t('auth.forgot.newPasswordLabel') }}<b>*</b></label>
          <input
            id="qz-forgot-password"
            v-model="newPassword"
            class="qz-field-control"
            type="password"
            autocomplete="new-password"
            required
            :placeholder="t('auth.forgot.newPasswordPlaceholder')"
          />
        </div>

        <p v-if="error" class="qz-alert is-danger" role="alert">{{ error }}</p>

        <QzButton type="submit" variant="primary" class="qz-btn--block" :loading="userAuthStore.loading">
          <RotateCw v-if="!userAuthStore.loading" />
          {{ userAuthStore.loading ? t('auth.forgot.submitting') : t('auth.forgot.submit') }}
        </QzButton>
      </form>

      <footer v-if="emailVerificationEnabled" class="qz-auth-foot">
        <RouterLink to="/auth/login">{{ t('auth.forgot.backLogin') }}</RouterLink>
      </footer>
    </section>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, RotateCw } from 'lucide-vue-next'
import ImageCaptcha from '../../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../../components/captcha/TurnstileCaptcha.vue'
import { useForgot } from '../../../composables/useForgot'
import { useTheme } from '../../../utils/theme'
import { setupReveal } from '../utils/reveal'
import QzButton from '../components/QzButton.vue'

const { t } = useI18n()
const { theme } = useTheme()

const {
  userAuthStore, brandSiteName, emailVerificationEnabled,
  email, code, newPassword, error, sending, countdown,
  captchaPayload, turnstileToken, imageCaptchaRef, turnstileRef,
  captchaProvider, sendCodeCaptchaEnabled, turnstileSiteKey,
  handleCaptchaConfigStale, handleSendCode, handleReset,
} = useForgot()

// 字符串模板 ref 显式占用，避免 vue-tsc 误报未使用
void imageCaptchaRef
void turnstileRef

// 按钮在 loading 时只是 aria-disabled（保留焦点），回车隐式提交 / 连点要在这里拦住，避免重复请求
const submitReset = () => {
  if (userAuthStore.loading) return
  void handleReset()
}
const sendCode = () => {
  if (sending.value || countdown.value > 0) return
  void handleSendCode()
}

// 入场动画
const rootEl = ref<HTMLElement | null>(null)
let cleanupReveal: (() => void) | null = null
onMounted(() => {
  void nextTick(() => {
    cleanupReveal = setupReveal(rootEl.value)
  })
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

/* (0,3,0) 压过 base.css 的 .qz-panel 内边距，与登录 / 注册卡一致 */
.qz-panel.qz-auth-card {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-6);
  padding: var(--qz-space-8) var(--qz-space-6);
}

/* ---------- 页头 ---------- */
.qz-auth-hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--qz-space-2);
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

.qz-auth-code-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--qz-space-2);
}

/* 提示条（错误 / 功能关闭）用 base.css 的 .qz-alert.is-danger */

/* ---------- 页脚 ---------- */
.qz-auth-foot {
  display: flex;
  justify-content: center;
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
}

@media (max-width: 540px) {
  .qz-auth-code-row {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
