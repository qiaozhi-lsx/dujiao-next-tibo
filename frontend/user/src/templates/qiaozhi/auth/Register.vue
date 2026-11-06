<template>
  <div ref="rootEl" class="qz-container qz-auth">
    <RouterLink to="/" class="qz-btn qz-btn--secondary qz-btn--sm qz-auth-back qz-reveal is-pending" :data-reveal="1">
      <ChevronLeft />
      {{ t('auth.login.backHome') }}
    </RouterLink>

    <section class="qz-panel qz-auth-card qz-reveal is-pending" :data-reveal="1">
      <!-- 注册关闭 -->
      <div v-if="!registrationEnabled" class="qz-empty">
        <PackageOpen />
        <b>{{ t('auth.register.registrationDisabled') }}</b>
        <RouterLink to="/auth/login" class="qz-btn qz-btn--primary">
          {{ t('auth.register.hasAccount') }}
        </RouterLink>
      </div>

      <template v-else>
        <header class="qz-auth-hero">
          <span class="qz-eyebrow">{{ t('qiaozhi.register.eyebrow') }}</span>
          <h1>{{ t('auth.register.title') }}</h1>
          <p>{{ t('qiaozhi.register.subtitle') }}</p>
          <span class="qz-chip qz-auth-brand">{{ brandSiteName }}</span>
        </header>

        <form class="qz-auth-form" novalidate @submit.prevent="submitRegister">
          <!-- 邮箱 -->
          <div class="qz-field">
            <label class="qz-field-label" for="qz-reg-email">{{ t('auth.register.emailLabel') }}<b>*</b></label>
            <div v-if="emailDomainSelectionRequired" class="qz-auth-email-row">
              <input
                id="qz-reg-email"
                v-model="emailLocalPart"
                class="qz-field-control"
                type="text"
                autocomplete="username"
                required
                :placeholder="t('auth.register.emailLocalPlaceholder')"
                :aria-invalid="formValidation.hasError('email') || undefined"
                :aria-describedby="emailDescribedBy"
                @blur="touchRegistrationEmail"
              />
              <QzSelect
                v-model="selectedEmailDomain"
                :items="allowedEmailDomains.map((domain: string) => ({ value: domain, label: '@' + domain }))"
                @change="touchRegistrationEmail"
              />
            </div>
            <input
              v-else
              id="qz-reg-email"
              v-model="email"
              class="qz-field-control"
              type="email"
              autocomplete="username"
              required
              :placeholder="t('auth.register.emailPlaceholder')"
              :aria-invalid="formValidation.hasError('email') || undefined"
              :aria-describedby="emailDescribedBy"
              @blur="touchRegistrationEmail"
            />
            <p v-if="formValidation.getError('email')" id="qz-reg-email-err" class="qz-field-error">
              {{ formValidation.getError('email') }}
            </p>
            <p v-else-if="emailDomainSelectionRequired" id="qz-reg-email-hint" class="qz-field-hint">
              {{ t('auth.register.emailDomainSelectHint') }}
            </p>
            <p v-else-if="emailDomainAllowlistEnabled" id="qz-reg-email-hint" class="qz-field-hint">
              {{ allowedEmailDomains.length > 0
                ? t('auth.register.allowedEmailDomainsHint', { domains: allowedEmailDomainsText })
                : t('auth.register.noAllowedEmailDomainsHint') }}
            </p>
          </div>

          <!-- 密码 -->
          <div class="qz-field">
            <label class="qz-field-label" for="qz-reg-password">{{ t('auth.register.passwordLabel') }}<b>*</b></label>
            <div class="qz-field-control qz-auth-password">
              <input
                id="qz-reg-password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="new-password"
                required
                :placeholder="t('auth.register.passwordPlaceholder')"
                :aria-invalid="formValidation.hasError('password') || undefined"
                :aria-describedby="formValidation.hasError('password') ? 'qz-reg-password-err' : undefined"
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
            <p v-if="formValidation.getError('password')" id="qz-reg-password-err" class="qz-field-error">
              {{ formValidation.getError('password') }}
            </p>
            <div
              v-else-if="password"
              class="qz-auth-strength"
              :class="`is-${passwordStrength}`"
            >
              <div class="qz-auth-strength-bars" aria-hidden="true">
                <span></span>
                <span></span>
                <span></span>
              </div>
              <small>{{ t(`formValidation.passwordStrength.${passwordStrength}`) }}</small>
            </div>
          </div>

          <!-- 发码场景验证码 -->
          <div
            v-if="emailVerificationEnabled && sendCodeCaptchaEnabled"
            class="qz-field"
            role="group"
            aria-labelledby="qz-reg-captcha-label"
          >
            <span id="qz-reg-captcha-label" class="qz-field-label">{{ t('auth.common.captchaLabel') }}<b>*</b></span>
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

          <!-- 邮箱验证码 -->
          <div v-if="emailVerificationEnabled" class="qz-field">
            <label class="qz-field-label" for="qz-reg-code">{{ t('auth.register.codeLabel') }}<b>*</b></label>
            <div class="qz-auth-code-row">
              <input
                id="qz-reg-code"
                v-model="code"
                class="qz-field-control"
                type="text"
                autocomplete="one-time-code"
                required
                :placeholder="t('auth.register.codePlaceholder')"
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

          <p v-if="error" class="qz-auth-alert is-error" role="alert">{{ error }}</p>

          <QzButton
            type="submit"
            variant="primary"
            class="qz-btn--block"
            :loading="userAuthStore.loading"
            :disabled="!agreed"
          >
            <UserPlus v-if="!userAuthStore.loading" />
            {{ userAuthStore.loading ? t('auth.register.creating') : t('auth.register.create') }}
          </QzButton>
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
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, Eye, EyeOff, PackageOpen, UserPlus } from 'lucide-vue-next'
import ImageCaptcha from '../../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../../components/captcha/TurnstileCaptcha.vue'
import { useRegister } from '../../../composables/useRegister'
import { useTheme } from '../../../utils/theme'
import { setupReveal } from '../utils/reveal'
import QzButton from '../components/QzButton.vue'
import QzSelect from '../components/QzSelect.vue'

const { t } = useI18n()
const { theme } = useTheme()

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

// 邮箱框：有错读错误，否则读域名提示（两条提示共用一个 id，同一时刻只渲染一条）
const emailDescribedBy = computed(() => {
  if (formValidation.hasError('email')) return 'qz-reg-email-err'
  if (emailDomainSelectionRequired.value || emailDomainAllowlistEnabled.value) return 'qz-reg-email-hint'
  return undefined
})

// 按钮在 loading 时只是 aria-disabled（保留焦点），回车隐式提交 / 连点要在这里拦住，避免重复请求
const submitRegister = () => {
  if (userAuthStore.loading) return
  void handleRegister()
}
const sendCode = () => {
  if (sending.value || countdown.value > 0) return
  void handleSendCode()
}

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
  gap: var(--qz-space-4);
  max-width: 460px;
}

.qz-auth-back {
  align-self: flex-start;
}

/* .qz-panel.qz-auth-card 写成 (0,3,0)：legacy.css 里旧分类侧栏的 .qz-panel 规则（sticky、≤760 隐藏、overflow hidden）
   在它被删掉之前不能影响这张卡。 */
.qz-panel.qz-auth-card {
  position: static;
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-6);
  padding: var(--qz-space-8) var(--qz-space-6);
  overflow: visible;
}

/* ---------- 页头（同样压过 legacy 的 .qz-panel > header 内边距 / 底线） ---------- */
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

.qz-auth-email-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(9rem, auto);
  gap: var(--qz-space-2);
}

.qz-auth-code-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--qz-space-2);
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

/* 密码强度：三段条 + 文字，状态色成对出现（条 + 文字） */
.qz-auth-strength {
  display: flex;
  align-items: center;
  gap: var(--qz-space-3);
}

.qz-auth-strength-bars {
  display: flex;
  flex: 1 1 auto;
  gap: var(--qz-space-1);
}

.qz-auth-strength-bars span {
  flex: 1 1 0;
  height: 4px;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-border);
  transition: background-color var(--qz-duration-fast) var(--qz-ease-standard);
}

.qz-auth-strength.is-weak .qz-auth-strength-bars span:nth-child(1) {
  background: var(--qz-danger);
}

.qz-auth-strength.is-medium .qz-auth-strength-bars span:nth-child(-n+2) {
  background: var(--qz-warning);
}

.qz-auth-strength.is-strong .qz-auth-strength-bars span {
  background: var(--qz-success);
}

.qz-auth-strength small {
  flex: none;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  font-weight: 600;
  line-height: var(--qz-leading-body);
}

.qz-auth-strength.is-weak small {
  color: var(--qz-danger);
}

.qz-auth-strength.is-medium small {
  color: var(--qz-warning);
}

.qz-auth-strength.is-strong small {
  color: var(--qz-success);
}

/* 服务条款 */
.qz-auth-agree {
  display: flex;
  align-items: flex-start;
  gap: var(--qz-space-3);
  padding: var(--qz-space-3) var(--qz-space-4);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  cursor: pointer;
}

.qz-auth-agree input {
  flex: none;
  width: 16px;
  height: 16px;
  margin: 1px 0 0;
  accent-color: var(--qz-accent);
}

.qz-auth-agree a {
  color: var(--qz-text);
  font-weight: 600;
  text-decoration: underline;
  text-underline-offset: 3px;
}

/* 错误提示条 */
.qz-auth-alert {
  margin: 0;
  padding: var(--qz-space-3) var(--qz-space-4);
  border-radius: var(--qz-radius-control);
  font-size: var(--qz-text-sm);
  font-weight: 500;
  line-height: var(--qz-leading-body);
  overflow-wrap: anywhere;
}

.qz-auth-alert.is-error {
  background: color-mix(in oklab, var(--qz-danger) 12%, transparent);
  color: var(--qz-danger);
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
}

@media (max-width: 540px) {
  .qz-auth-email-row,
  .qz-auth-code-row {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
