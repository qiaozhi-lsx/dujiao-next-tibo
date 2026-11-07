<template>
  <div class="qz-me-panel">
    <!-- ==================== 安全总览 ==================== -->
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.security.title') }}</h2>
          <small>{{ requiresOldEmailCode ? t('personalCenter.security.subtitle') : t('personalCenter.security.subtitleBindOnly') }}</small>
        </div>
        <div class="qz-me-acts">
          <span class="qz-chip">{{ t('personalCenter.tabs.security') }}</span>
        </div>
      </div>
      <p
        v-if="securityAlert"
        class="qz-alert"
        :class="alertClass(securityAlert.level)"
        :role="securityAlert.level === 'error' ? 'alert' : 'status'"
      >
        {{ securityAlert.message }}
      </p>
    </section>

    <!-- ==================== Telegram 绑定 ==================== -->
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.security.telegramTitle') }}</h2>
          <small>{{ telegramEnabled ? t('personalCenter.security.telegramSubtitle') : t('personalCenter.security.telegramDisabledTip') }}</small>
        </div>
        <div class="qz-me-acts">
          <span class="qz-badge" :class="telegramBound ? 'is-success' : 'is-warning'">
            {{ telegramBound ? t('personalCenter.security.telegramBound') : t('personalCenter.security.telegramUnbound') }}
          </span>
        </div>
      </div>

      <div class="qz-me-stack">
        <p v-if="userProfileStore.loadingTelegramBinding" class="qz-field-hint">
          {{ t('personalCenter.security.telegramLoading') }}
        </p>

        <template v-else-if="telegramBound">
          <div class="qz-me-ident">
            <span v-if="telegramAvatarUrl" class="qz-me-avatar">
              <img :src="telegramAvatarUrl" alt="Telegram Avatar" />
            </span>
            <div class="qz-me-ident-text">
              <b>{{ telegramDisplayName }}</b>
              <small>{{ t('personalCenter.security.telegramBindID', { id: telegramProviderUserId }) }}</small>
              <small>{{ t('personalCenter.security.telegramBindTime', { time: formatDate(userProfileStore.telegramBinding?.auth_at) }) }}</small>
            </div>
          </div>

          <div class="qz-me-acts qz-me-acts-start">
            <QzButton
              type="button"
              variant="secondary"
              size="sm"
              :loading="userProfileStore.unbindingTelegram"
              :disabled="!canUnbindTelegram"
              @click="handleUnbindTelegram"
            >
              {{ userProfileStore.unbindingTelegram ? t('personalCenter.security.telegramUnbinding') : t('personalCenter.security.telegramUnbind') }}
            </QzButton>
          </div>
          <p v-if="!canUnbindTelegram" class="qz-field-hint">
            {{ t('personalCenter.security.telegramUnbindDisabledTip') }}
          </p>

          <template v-if="showTelegramMiniAppEntry">
            <p class="qz-field-hint">{{ t('personalCenter.security.telegramMiniAppEntryHint') }}</p>
            <div class="qz-me-acts qz-me-acts-start">
              <QzButton type="button" variant="secondary" size="sm" @click="openTelegramMiniAppEntry">
                {{ t('personalCenter.security.telegramMiniAppEntryAction') }}
              </QzButton>
            </div>
          </template>
        </template>

        <template v-else>
          <p class="qz-field-hint">
            {{ telegramEnabled
              ? (showMiniAppBindAction ? t('personalCenter.security.telegramMiniAppBindHint') : t('personalCenter.security.telegramUnboundTip'))
              : t('personalCenter.security.telegramDisabledTip') }}
          </p>

          <div v-if="showMiniAppBindAction" class="qz-me-acts qz-me-acts-start">
            <QzButton
              type="button"
              :loading="userProfileStore.bindingTelegram"
              :disabled="miniAppInitData === ''"
              @click="handleTelegramMiniAppBind"
            >
              {{ userProfileStore.bindingTelegram ? t('personalCenter.security.telegramMiniAppBinding') : t('personalCenter.security.telegramMiniAppBindAction') }}
            </QzButton>
          </div>
          <div v-else-if="showTelegramWidget" ref="telegramWidgetRef" class="qz-sec-widget"></div>
          <template v-else-if="showTelegramOidcBind">
            <div class="qz-me-acts qz-me-acts-start">
              <QzButton
                type="button"
                :disabled="userProfileStore.bindingTelegram"
                @click="startTelegramOidcBind"
              >
                {{ t('personalCenter.security.telegramOidcBindButton') }}
              </QzButton>
            </div>
            <p class="qz-field-hint">{{ t('personalCenter.security.telegramOidcBindHint') }}</p>
          </template>

          <template v-if="showTelegramMiniAppEntry">
            <p class="qz-field-hint">{{ t('personalCenter.security.telegramMiniAppEntryHint') }}</p>
            <div class="qz-me-acts qz-me-acts-start">
              <QzButton type="button" variant="secondary" size="sm" @click="openTelegramMiniAppEntry">
                {{ t('personalCenter.security.telegramMiniAppEntryAction') }}
              </QzButton>
            </div>
          </template>
        </template>
      </div>
    </section>

    <!-- ==================== Google 绑定 ==================== -->
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.security.googleTitle') }}</h2>
          <small>{{ googleAuthorizationEnabled ? t('personalCenter.security.googleSubtitle') : t('personalCenter.security.googleDisabledTip') }}</small>
        </div>
        <div class="qz-me-acts">
          <span class="qz-badge" :class="googleBound ? 'is-success' : 'is-warning'">
            {{ googleBound ? t('personalCenter.security.googleBound') : t('personalCenter.security.googleUnbound') }}
          </span>
        </div>
      </div>

      <div class="qz-me-stack">
        <p v-if="userProfileStore.loadingGoogleBinding" class="qz-field-hint">
          {{ t('personalCenter.security.googleLoading') }}
        </p>

        <template v-else-if="googleBound">
          <div class="qz-me-ident">
            <span class="qz-me-avatar">
              <img
                v-if="googleAvatarUrl"
                :src="googleAvatarUrl"
                :alt="t('personalCenter.security.googleAvatarAlt')"
              />
              <template v-else>G</template>
            </span>
            <div class="qz-me-ident-text">
              <b>{{ googleDisplayName }}</b>
              <small v-if="googleEmail">{{ googleEmail }}</small>
              <small>{{ t('personalCenter.security.googleBindID', { id: googleProviderUserId }) }}</small>
              <small>{{ t('personalCenter.security.googleBindTime', { time: formatDate(userProfileStore.googleBinding?.auth_at) }) }}</small>
            </div>
          </div>

          <div class="qz-me-acts qz-me-acts-start">
            <QzButton
              type="button"
              variant="secondary"
              size="sm"
              :loading="userProfileStore.unbindingGoogle"
              :disabled="!canUnbindGoogle"
              @click="handleUnbindGoogle"
            >
              {{ userProfileStore.unbindingGoogle ? t('personalCenter.security.googleUnbinding') : t('personalCenter.security.googleUnbind') }}
            </QzButton>
          </div>
          <p v-if="!canUnbindGoogle" class="qz-field-hint">
            {{ t('personalCenter.security.googleUnbindDisabledTip') }}
          </p>
        </template>

        <template v-else>
          <p class="qz-field-hint">
            {{ !googleAuthorizationEnabled
              ? t('personalCenter.security.googleDisabledTip')
              : isTelegramMiniApp
                ? t('personalCenter.security.googleMiniAppUnavailable')
                : t('personalCenter.security.googleUnboundTip') }}
          </p>
          <GoogleIdentityButton
            v-if="googleAuthorizationEnabled && !isTelegramMiniApp"
            :client-id="googleClientID"
            :locale="googleButtonLocale"
            text="continue_with"
            :ux-mode="googleIdentityUXMode"
            :login-uri="googleRedirectLoginURI"
            :prepare-redirect="prepareGoogleRedirectBind"
            :disabled="userProfileStore.bindingGoogle"
            :loading-label="t('personalCenter.security.googleButtonLoading')"
            @credential="handleGoogleBind"
            @error="handleGoogleScriptError"
          />
          <p v-if="userProfileStore.bindingGoogle" class="qz-field-hint" role="status">
            {{ t('personalCenter.security.googleBinding') }}
          </p>
        </template>
      </div>
    </section>

    <!-- ==================== 邮箱换绑 / 绑定 ==================== -->
    <section class="qz-panel qz-me-sec">
      <form class="qz-me-form" novalidate @submit.prevent="handleChangeEmail">
        <div class="qz-field">
          <label class="qz-field-label" for="qz-me-sec-current-email">{{ t('personalCenter.security.currentEmailLabel') }}</label>
          <input id="qz-me-sec-current-email" class="qz-field-control" :value="currentEmailDisplay" type="email" disabled />
          <p v-if="!requiresOldEmailCode" class="qz-field-hint">
            {{ t('personalCenter.security.bindOnlyTip') }}
          </p>
        </div>

        <div class="qz-field">
          <label class="qz-field-label" for="qz-me-sec-new-email">{{ t('personalCenter.security.newEmailLabel') }}</label>
          <input
            id="qz-me-sec-new-email"
            v-model="securityForm.newEmail"
            class="qz-field-control"
            type="email"
            autocomplete="email"
            :placeholder="t('personalCenter.security.newEmailPlaceholder')"
          />
        </div>

        <div :class="requiresOldEmailCode ? 'qz-me-form-row' : ''">
          <div v-if="requiresOldEmailCode" class="qz-field">
            <label class="qz-field-label" for="qz-me-sec-old-code">{{ t('personalCenter.security.oldCodeLabel') }}</label>
            <div class="qz-sec-inline">
              <input
                id="qz-me-sec-old-code"
                v-model="securityForm.oldCode"
                class="qz-field-control"
                inputmode="numeric"
                :placeholder="t('personalCenter.security.codePlaceholder')"
              />
              <button
                type="button"
                class="qz-btn qz-btn--secondary"
                :disabled="userProfileStore.sendingCode || oldCodeCooldown > 0"
                @click="handleSendOldCode"
              >
                {{ oldCodeCooldown > 0 ? t('personalCenter.security.countdown', { seconds: oldCodeCooldown }) : t('personalCenter.security.sendOldCode') }}
              </button>
            </div>
          </div>

          <div class="qz-field">
            <label class="qz-field-label" for="qz-me-sec-new-code">{{ t('personalCenter.security.newCodeLabel') }}</label>
            <div class="qz-sec-inline">
              <input
                id="qz-me-sec-new-code"
                v-model="securityForm.newCode"
                class="qz-field-control"
                inputmode="numeric"
                :placeholder="t('personalCenter.security.codePlaceholder')"
              />
              <button
                type="button"
                class="qz-btn qz-btn--secondary"
                :disabled="userProfileStore.sendingCode || newCodeCooldown > 0"
                @click="handleSendNewCode"
              >
                {{ newCodeCooldown > 0 ? t('personalCenter.security.countdown', { seconds: newCodeCooldown }) : t('personalCenter.security.sendNewCode') }}
              </button>
            </div>
          </div>
        </div>

        <div class="qz-me-acts qz-me-acts-start">
          <QzButton type="submit" :loading="userProfileStore.changingEmail">
            {{ userProfileStore.changingEmail
              ? t('personalCenter.security.submitting')
              : (requiresOldEmailCode ? t('personalCenter.security.submit') : t('personalCenter.security.bindSubmit')) }}
          </QzButton>
        </div>
      </form>
    </section>

    <!-- ==================== 最近登录记录 ==================== -->
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.security.loginLogsTitle') }}</h2>
          <small>{{ t('personalCenter.security.loginLogsTip') }}</small>
        </div>
      </div>

      <p v-if="userProfileStore.loadingLoginLogs" class="qz-field-hint">
        {{ t('personalCenter.security.loginLogsLoading') }}
      </p>
      <p v-else-if="userProfileStore.recentLoginLogs.length === 0" class="qz-field-hint">
        {{ t('personalCenter.security.loginLogsEmpty') }}
      </p>
      <div v-else class="qz-me-rows">
        <div v-for="item in userProfileStore.recentLoginLogs" :key="item.id" class="qz-card qz-me-row">
          <div class="qz-me-row-main">
            <b>{{ formatDate(item.created_at) }}</b>
            <small>{{ t('personalCenter.security.loginLogsIp') }}：{{ item.client_ip || '-' }}</small>
            <small>{{ t('personalCenter.security.loginLogsReason') }}：{{ loginReasonLabel(item.fail_reason) }}</small>
          </div>
          <div class="qz-me-row-side">
            <span class="qz-badge" :class="loginStatusTone(item.status)">{{ loginStatusLabel(item.status) }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ==================== 登录密码 ==================== -->
    <section v-if="canManagePassword" class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ requiresOldPassword ? t('personalCenter.security.passwordTitle') : t('personalCenter.security.setPasswordTitle') }}</h2>
          <small>{{ requiresOldPassword ? t('personalCenter.security.passwordSubtitle') : t('personalCenter.security.setPasswordSubtitle') }}</small>
        </div>
      </div>

      <form class="qz-me-form" novalidate @submit.prevent="handleChangePassword">
        <div v-if="requiresOldPassword" class="qz-field">
          <label class="qz-field-label" for="qz-me-sec-old-password">{{ t('personalCenter.security.currentPasswordLabel') }}</label>
          <input
            id="qz-me-sec-old-password"
            v-model="passwordForm.oldPassword"
            class="qz-field-control"
            type="password"
            autocomplete="current-password"
            :placeholder="t('personalCenter.security.passwordPlaceholder')"
          />
        </div>

        <div class="qz-me-form-row">
          <div class="qz-field">
            <label class="qz-field-label" for="qz-me-sec-new-password">{{ t('personalCenter.security.newPasswordLabel') }}</label>
            <input
              id="qz-me-sec-new-password"
              v-model="passwordForm.newPassword"
              class="qz-field-control"
              type="password"
              autocomplete="new-password"
              :placeholder="t('personalCenter.security.passwordPlaceholder')"
            />
          </div>
          <div class="qz-field">
            <label class="qz-field-label" for="qz-me-sec-confirm-password">{{ t('personalCenter.security.confirmPasswordLabel') }}</label>
            <input
              id="qz-me-sec-confirm-password"
              v-model="passwordForm.confirmPassword"
              class="qz-field-control"
              type="password"
              autocomplete="new-password"
              :placeholder="t('personalCenter.security.passwordPlaceholder')"
            />
          </div>
        </div>

        <div class="qz-me-acts qz-me-acts-start">
          <QzButton type="submit" :loading="userProfileStore.changingPassword">
            {{ userProfileStore.changingPassword
              ? (requiresOldPassword ? t('personalCenter.security.changePasswordSubmitting') : t('personalCenter.security.setPasswordSubmitting'))
              : (requiresOldPassword ? t('personalCenter.security.changePassword') : t('personalCenter.security.setPassword')) }}
          </QzButton>
        </div>
      </form>
    </section>

    <!-- ==================== 两步验证（2FA） ==================== -->
    <section v-if="canManagePassword" class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.security.twofa.title') }}</h2>
          <small>{{ t('personalCenter.security.twofa.subtitle') }}</small>
        </div>
        <div class="qz-me-acts">
          <span class="qz-badge" :class="{ 'is-success': status?.enabled }">
            {{ status?.enabled ? t('personalCenter.security.twofa.statusEnabled') : t('personalCenter.security.twofa.statusDisabled') }}
          </span>
        </div>
      </div>

      <p
        v-if="twofaAlert"
        class="qz-alert"
        :class="alertClass(twofaAlert.level)"
        :role="twofaAlert.level === 'error' ? 'alert' : 'status'"
      >
        {{ twofaAlert.message }}
      </p>

      <!-- 已启用：状态 + 关闭 / 重新生成恢复码 -->
      <div v-if="status?.enabled" class="qz-me-stack">
        <div class="qz-sec-sub">
          <p class="qz-field-hint">{{ t('personalCenter.security.twofa.enabledAt', { date: formatDate(status.enabled_at) }) }}</p>
          <p class="qz-field-hint">
            {{ t('personalCenter.security.twofa.recoveryRemaining', {
              remaining: status.recovery_codes_remaining,
              total: status.recovery_codes_total,
            }) }}
          </p>
        </div>

        <div class="qz-me-acts qz-me-acts-start">
          <QzButton type="button" variant="secondary" @click="openRegenerate">
            {{ t('personalCenter.security.twofa.regenerateAction') }}
          </QzButton>
          <QzButton type="button" variant="danger" @click="openDisable">
            {{ t('personalCenter.security.twofa.disableAction') }}
          </QzButton>
        </div>

        <!-- 关闭 2FA 表单 -->
        <div v-if="mode === 'disable'" class="qz-sec-sub">
          <p class="qz-field-hint">{{ t('personalCenter.security.twofa.disableHint') }}</p>
          <QzSegment
            :model-value="disableMode"
            :items="disableModeItems"
            @update:model-value="onDisableModeChange"
          />
          <div class="qz-field">
            <input
              v-if="disableMode === 'code'"
              v-model="disableCode"
              class="qz-field-control qz-sec-code"
              inputmode="numeric"
              maxlength="6"
              :aria-label="t('personalCenter.security.twofa.codePlaceholder')"
              :placeholder="t('personalCenter.security.twofa.codePlaceholder')"
            />
            <input
              v-else
              v-model="disableRecovery"
              class="qz-field-control"
              autocomplete="off"
              :aria-label="t('personalCenter.security.twofa.recoveryPlaceholder')"
              :placeholder="t('personalCenter.security.twofa.recoveryPlaceholder')"
            />
          </div>
          <div class="qz-me-acts qz-me-acts-start">
            <QzButton type="button" variant="danger" :loading="loading" @click="submitDisable">
              {{ loading ? t('personalCenter.security.twofa.disableSubmitting') : t('personalCenter.security.twofa.disableSubmit') }}
            </QzButton>
            <QzButton type="button" variant="secondary" @click="resetMode">
              {{ t('personalCenter.security.twofa.cancel') }}
            </QzButton>
          </div>
        </div>

        <!-- 重新生成恢复码表单 -->
        <div v-if="mode === 'regenerate'" class="qz-sec-sub">
          <p class="qz-field-hint">{{ t('personalCenter.security.twofa.regenerateHint') }}</p>
          <div class="qz-field">
            <input
              v-model="regenerateCode"
              class="qz-field-control qz-sec-code"
              inputmode="numeric"
              maxlength="6"
              :aria-label="t('personalCenter.security.twofa.codePlaceholder')"
              :placeholder="t('personalCenter.security.twofa.codePlaceholder')"
            />
          </div>
          <div class="qz-me-acts qz-me-acts-start">
            <QzButton type="button" :loading="loading" @click="submitRegenerate">
              {{ loading ? t('personalCenter.security.twofa.regenerateSubmitting') : t('personalCenter.security.twofa.regenerateSubmit') }}
            </QzButton>
            <QzButton type="button" variant="secondary" @click="resetMode">
              {{ t('personalCenter.security.twofa.cancel') }}
            </QzButton>
          </div>
        </div>
      </div>

      <!-- 未启用：绑定流程 -->
      <div v-else class="qz-me-stack">
        <p class="qz-field-hint">{{ t('personalCenter.security.twofa.notEnabledHint') }}</p>

        <div v-if="!setupResult" class="qz-me-acts qz-me-acts-start">
          <QzButton type="button" :loading="loading" @click="startSetup">
            {{ loading ? t('personalCenter.security.twofa.startingSetup') : t('personalCenter.security.twofa.startSetup') }}
          </QzButton>
        </div>

        <div v-else class="qz-sec-sub">
          <p class="qz-field-hint">{{ t('personalCenter.security.twofa.setupStep1') }}</p>
          <div v-if="qrcodeDataUrl" class="qz-sec-qr">
            <img :src="qrcodeDataUrl" alt="2FA QR" />
          </div>
          <div class="qz-me-stack">
            <p class="qz-field-hint">{{ t('personalCenter.security.twofa.secret') }}</p>
            <code class="qz-sec-secret">{{ setupResult.secret }}</code>
          </div>

          <p class="qz-field-hint">{{ t('personalCenter.security.twofa.setupStep2') }}</p>
          <div class="qz-field">
            <input
              v-model="enableCode"
              class="qz-field-control qz-sec-code"
              inputmode="numeric"
              maxlength="6"
              :aria-label="t('personalCenter.security.twofa.codePlaceholder')"
              :placeholder="t('personalCenter.security.twofa.codePlaceholder')"
            />
          </div>

          <div class="qz-me-acts qz-me-acts-start">
            <QzButton type="button" :loading="loading" @click="submitEnable">
              {{ loading ? t('personalCenter.security.twofa.enableSubmitting') : t('personalCenter.security.twofa.enableSubmit') }}
            </QzButton>
            <QzButton type="button" variant="secondary" @click="cancelSetup">
              {{ t('personalCenter.security.twofa.cancel') }}
            </QzButton>
          </div>
        </div>
      </div>

      <!-- 恢复码展示弹窗（启用成功 / 重新生成成功后） -->
      <Teleport to="body">
        <div
          v-if="recoveryCodes.length > 0"
          class="qz-sec-mask"
          role="dialog"
          aria-modal="true"
          :aria-label="t('personalCenter.security.twofa.recoveryTitle')"
          @click.self="recoveryCodes = []"
        >
          <div class="qz-panel qz-sec-modal">
            <h4>{{ t('personalCenter.security.twofa.recoveryTitle') }}</h4>
            <p class="qz-field-hint">{{ t('personalCenter.security.twofa.recoveryWarning') }}</p>
            <div class="qz-sec-codes">
              <code v-for="(c, idx) in recoveryCodes" :key="idx">{{ c }}</code>
            </div>
            <div class="qz-sec-modal-acts">
              <QzButton type="button" variant="secondary" @click="copyRecoveryCodes">
                {{ copied ? t('personalCenter.security.twofa.copied') : t('personalCenter.security.twofa.copy') }}
              </QzButton>
              <QzButton type="button" @click="acknowledgeRecoveryCodes">
                {{ t('personalCenter.security.twofa.acknowledge') }}
              </QzButton>
            </div>
          </div>
        </div>
      </Teleport>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import QRCode from 'qrcode'
import type { PageAlert, PageAlertLevel } from '../../../utils/alerts'
import { userProfileAPI } from '../../../api/user'
import { userTotpAPI } from '../../../api/auth'
import type { TelegramAuthPayload } from '../../../api'
import { useAppStore } from '../../../stores/app'
import { useTelegramMiniAppStore } from '../../../stores/telegramMiniApp'
import { useUserProfileStore } from '../../../stores/userProfile'
import { useUserAuthStore } from '../../../stores/userAuth'
import { buildTelegramMiniAppEntryLink, isTelegramUrlEnvironment, openTelegramCompatibleLink } from '../../../utils/telegramMiniApp'
import { canUnbindExternalIdentity } from '../../../utils/externalIdentity'
import { detectGoogleIdentityUXMode } from '../../../utils/googleIdentity'
import {
  createGoogleRedirectIntent,
  createGoogleRedirectPreparedIntent,
  getGoogleRedirectSessionStorage,
  storeGoogleRedirectIntent,
  tryBuildGoogleRedirectCredentialCallbackURL,
} from '../../../utils/googleRedirect'
import GoogleIdentityButton from '../../../components/auth/GoogleIdentityButton.vue'
import QzButton from '../components/QzButton.vue'
import QzSegment, { type QzSegmentItem } from '../components/QzSegment.vue'
import { formatDate } from './support'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const telegramMiniAppStore = useTelegramMiniAppStore()
const userProfileStore = useUserProfileStore()
const userAuthStore = useUserAuthStore()

const securityForm = reactive({
  newEmail: '',
  oldCode: '',
  newCode: '',
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const securityAlert = ref<PageAlert | null>(null)
const oldCodeCooldown = ref(0)
const newCodeCooldown = ref(0)
const telegramWidgetRef = ref<HTMLDivElement | null>(null)
let cooldownTimer: number | null = null
const telegramCallbackName = '__dujiaoSecurityTelegramBind'

/** 提示条 tone（base .qz-alert）：success → is-success、warning → is-warning、error → is-danger */
const alertClass = (level: PageAlertLevel) => {
  if (level === 'error') return 'is-danger'
  if (level === 'success') return 'is-success'
  return 'is-warning'
}

/* ==================== 平台配置 / 绑定状态 ==================== */
const telegramConfig = computed(() => appStore.config?.telegram_auth || null)
const telegramBotUsername = computed(() => String(telegramConfig.value?.bot_username || '').trim())
const telegramMiniAppURL = computed(() => String(telegramConfig.value?.mini_app_url || '').trim())
const telegramLoginMode = computed(() => String(telegramConfig.value?.mode || '').trim())
const telegramEnabled = computed(() => !!telegramConfig.value?.enabled && telegramBotUsername.value !== '')
const telegramBound = computed(() => !!userProfileStore.telegramBinding?.bound)
const googleConfig = computed(() => appStore.config?.google_auth || null)
const googleClientID = computed(() => String(googleConfig.value?.client_id || '').trim())
const googleEnabled = computed(() => !!googleConfig.value?.enabled && googleClientID.value !== '')
const googleButtonLocale = computed(() => String(appStore.locale || '').trim())
const googleBound = computed(() => !!userProfileStore.googleBinding?.bound)
const googleIdentityUXMode = detectGoogleIdentityUXMode()
const googleRedirectLoginURI = googleIdentityUXMode === 'redirect'
  ? tryBuildGoogleRedirectCredentialCallbackURL()
  : ''
const googleRedirectAvailable = googleIdentityUXMode === 'popup' || googleRedirectLoginURI !== ''
const googleAuthorizationEnabled = computed(() => googleEnabled.value && googleRedirectAvailable)
const isTelegramUrlEnv = isTelegramUrlEnvironment()
const isTelegramMiniApp = computed(() => (telegramMiniAppStore.isMiniApp && telegramMiniAppStore.isReady) || isTelegramUrlEnv)
const miniAppInitData = computed(() => String(telegramMiniAppStore.initData || '').trim())
const showMiniAppBindAction = computed(() => telegramEnabled.value && !telegramBound.value && isTelegramMiniApp.value)
const showTelegramWidget = computed(() => telegramLoginMode.value !== 'oidc' && telegramEnabled.value && !telegramBound.value && !isTelegramMiniApp.value)
const showTelegramOidcBind = computed(() => telegramLoginMode.value === 'oidc' && telegramEnabled.value && !telegramBound.value && !isTelegramMiniApp.value)
const telegramMiniAppEntryLink = computed(() => buildTelegramMiniAppEntryLink(telegramBotUsername.value, telegramMiniAppURL.value))
const showTelegramMiniAppEntry = computed(() => !isTelegramMiniApp.value && telegramMiniAppEntryLink.value !== '')
const emailChangeMode = computed(() => userProfileStore.profile?.email_change_mode || 'change_with_old_and_new')
const requiresOldEmailCode = computed(() => emailChangeMode.value !== 'bind_only')
const canManagePassword = computed(() => requiresOldEmailCode.value)
const passwordChangeMode = computed(() => userProfileStore.profile?.password_change_mode || 'change_with_old')
const requiresOldPassword = computed(() => passwordChangeMode.value !== 'set_without_old')
// External identities can create passwordless accounts. Only an explicit
// backend authorization enables unbinding; missing capability data fails closed.
const canUnbindTelegram = computed(() => canUnbindExternalIdentity(userProfileStore.telegramBinding))
const canUnbindGoogle = computed(() => canUnbindExternalIdentity(userProfileStore.googleBinding))
const currentEmailDisplay = computed(() => {
  if (!requiresOldEmailCode.value) {
    return t('personalCenter.security.bindOnlyEmailDisplay')
  }
  return userProfileStore.profile?.email || ''
})
const telegramDisplayName = computed(() => {
  if (userProfileStore.telegramBinding?.username) {
    return `@${userProfileStore.telegramBinding.username}`
  }
  return t('personalCenter.security.telegramDisplayFallback')
})
const googleDisplayName = computed(() => {
  const displayName = String(userProfileStore.googleBinding?.display_name || '').trim()
  if (displayName !== '') return displayName
  const email = String(userProfileStore.googleBinding?.email || '').trim()
  if (email !== '') return email
  const username = String(userProfileStore.googleBinding?.username || '').trim()
  if (username !== '') return username
  return t('personalCenter.security.googleDisplayFallback')
})
const telegramAvatarUrl = computed(() => userProfileStore.telegramBinding?.avatar_url || '')
const telegramProviderUserId = computed(() => userProfileStore.telegramBinding?.provider_user_id || '-')
const googleAvatarUrl = computed(() => userProfileStore.googleBinding?.avatar_url || '')
const googleEmail = computed(() => userProfileStore.googleBinding?.email || userProfileStore.googleBinding?.username || '')
const googleProviderUserId = computed(() => userProfileStore.googleBinding?.provider_user_id || '-')

/* ==================== 登录记录展示 ==================== */
const loginStatusLabel = (status?: string) => {
  const normalized = (status || '').trim() || 'failed'
  return t(`personalCenter.security.loginLogsStatusMap.${normalized}`)
}
const loginStatusTone = (status?: string) => ((status || '').trim() === 'success' ? 'is-success' : 'is-danger')
const loginReasonLabel = (reason?: string) => {
  const normalized = (reason || '').trim()
  if (!normalized) return '-'
  const key = `personalCenter.security.loginLogsReasonMap.${normalized}`
  const translated = t(key)
  return translated === key ? normalized : translated
}

/* ==================== 邮箱换绑 / 绑定 ==================== */
const startCooldown = (kind: 'old' | 'new') => {
  if (kind === 'old') {
    oldCodeCooldown.value = 60
  } else {
    newCodeCooldown.value = 60
  }
  if (cooldownTimer !== null) return
  cooldownTimer = window.setInterval(() => {
    if (oldCodeCooldown.value > 0) {
      oldCodeCooldown.value -= 1
    }
    if (newCodeCooldown.value > 0) {
      newCodeCooldown.value -= 1
    }
    if (oldCodeCooldown.value === 0 && newCodeCooldown.value === 0 && cooldownTimer !== null) {
      window.clearInterval(cooldownTimer)
      cooldownTimer = null
    }
  }, 1000)
}

const handleSendOldCode = async () => {
  securityAlert.value = null
  if (!requiresOldEmailCode.value) {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.bindOnlyOldCodeDisabled'),
    }
    return
  }
  const ok = await userProfileStore.sendChangeEmailCode({ kind: 'old' })
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.sendCodeFailed'),
    }
    return
  }
  startCooldown('old')
  securityAlert.value = {
    level: 'success',
    message: t('personalCenter.security.sendOldCodeSuccess'),
  }
}

const handleSendNewCode = async () => {
  securityAlert.value = null
  const newEmail = securityForm.newEmail.trim()
  if (!newEmail) {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.newEmailRequired'),
    }
    return
  }
  const ok = await userProfileStore.sendChangeEmailCode({ kind: 'new', new_email: newEmail })
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.sendCodeFailed'),
    }
    return
  }
  startCooldown('new')
  securityAlert.value = {
    level: 'success',
    message: t('personalCenter.security.sendNewCodeSuccess'),
  }
}

const handleChangeEmail = async () => {
  securityAlert.value = null
  const requiresOldCode = requiresOldEmailCode.value
  const oldCode = securityForm.oldCode.trim()
  const payload = {
    new_email: securityForm.newEmail.trim(),
    new_code: securityForm.newCode.trim(),
    ...(requiresOldCode ? { old_code: oldCode } : {}),
  }
  if (!payload.new_email || !payload.new_code || (requiresOldCode && !oldCode)) {
    securityAlert.value = {
      level: 'warning',
      message: requiresOldCode
        ? t('personalCenter.security.changeEmailRequired')
        : t('personalCenter.security.bindEmailRequired'),
    }
    return
  }

  const ok = await userProfileStore.changeEmail(payload)
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.changeEmailFailed'),
    }
    return
  }

  securityForm.newEmail = ''
  securityForm.oldCode = ''
  securityForm.newCode = ''
  oldCodeCooldown.value = 0
  newCodeCooldown.value = 0
  securityAlert.value = {
    level: 'success',
    message: requiresOldCode
      ? t('personalCenter.security.changeEmailSuccess')
      : t('personalCenter.security.bindEmailSuccess'),
  }
}

/* ==================== 登录密码 ==================== */
const handleChangePassword = async () => {
  securityAlert.value = null
  const oldPassword = passwordForm.oldPassword.trim()
  const newPassword = passwordForm.newPassword.trim()
  const confirmPassword = passwordForm.confirmPassword.trim()
  const needOldPassword = requiresOldPassword.value

  if (!newPassword || !confirmPassword || (needOldPassword && !oldPassword)) {
    securityAlert.value = {
      level: 'warning',
      message: needOldPassword
        ? t('personalCenter.security.changePasswordRequired')
        : t('personalCenter.security.setPasswordRequired'),
    }
    return
  }

  if (newPassword !== confirmPassword) {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.passwordMismatch'),
    }
    return
  }

  const payload = {
    ...(needOldPassword ? { old_password: oldPassword } : {}),
    new_password: newPassword,
  }
  const ok = await userProfileStore.changePassword(payload)

  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.changePasswordFailed'),
    }
    return
  }

  passwordForm.oldPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
  securityAlert.value = {
    level: 'success',
    message: needOldPassword
      ? t('personalCenter.security.changePasswordSuccess')
      : t('personalCenter.security.setPasswordSuccess'),
  }

  userAuthStore.logout('/auth/login?reason=password_changed')
}

/* ==================== Telegram 绑定 ==================== */
const buildTelegramPayload = (raw: any): TelegramAuthPayload | null => {
  const id = Number(raw?.id)
  const authDate = Number(raw?.auth_date)
  const hash = String(raw?.hash || '').trim()
  if (!Number.isFinite(id) || id <= 0 || !Number.isFinite(authDate) || authDate <= 0 || hash === '') {
    return null
  }
  return {
    id,
    first_name: String(raw?.first_name || '').trim(),
    last_name: String(raw?.last_name || '').trim(),
    username: String(raw?.username || '').trim(),
    photo_url: String(raw?.photo_url || '').trim(),
    auth_date: authDate,
    hash,
  }
}

const clearTelegramWidget = () => {
  const widgetEl = telegramWidgetRef.value
  if (widgetEl) {
    widgetEl.innerHTML = ''
  }
}

const renderTelegramWidget = () => {
  if (telegramLoginMode.value === 'oidc') {
    clearTelegramWidget()
    return
  }
  const widgetEl = telegramWidgetRef.value
  if (!showTelegramWidget.value || !widgetEl) {
    clearTelegramWidget()
    return
  }
  clearTelegramWidget()
  const script = document.createElement('script')
  script.async = true
  script.src = 'https://telegram.org/js/telegram-widget.js?22'
  script.setAttribute('data-telegram-login', telegramBotUsername.value)
  script.setAttribute('data-size', 'large')
  script.setAttribute('data-userpic', 'false')
  script.setAttribute('data-request-access', 'write')
  script.setAttribute('data-onauth', `${telegramCallbackName}(user)`)
  script.onerror = () => {
    securityAlert.value = {
      level: 'error',
      message: t('personalCenter.security.telegramWidgetLoadFailed'),
    }
  }
  widgetEl.appendChild(script)
}

const refreshExternalIdentityBindings = async (): Promise<boolean> => {
  const [telegramLoaded, googleLoaded] = await Promise.all([
    userProfileStore.loadTelegramBinding(),
    userProfileStore.loadGoogleBinding(),
  ])
  return telegramLoaded && googleLoaded
}

const finishExternalIdentityMutation = async (successMessage: string) => {
  const refreshed = await refreshExternalIdentityBindings()
  securityAlert.value = refreshed
    ? {
        level: 'success',
        message: successMessage,
      }
    : {
        level: 'warning',
        message: t('personalCenter.security.externalIdentityRefreshFailed'),
      }
}

const openTelegramMiniAppEntry = () => {
  if (telegramMiniAppEntryLink.value === '') return
  openTelegramCompatibleLink(telegramMiniAppEntryLink.value)
}

const handleTelegramBind = async (raw: any) => {
  securityAlert.value = null
  const payload = buildTelegramPayload(raw)
  if (!payload) {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.telegramInvalidPayload'),
    }
    return
  }
  const ok = await userProfileStore.bindTelegram(payload)
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.telegramBindFailed'),
    }
    return
  }
  await finishExternalIdentityMutation(t('personalCenter.security.telegramBindSuccess'))
  renderTelegramWidget()
}

const handleTelegramMiniAppBind = async () => {
  securityAlert.value = null
  if (miniAppInitData.value === '') {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.telegramMiniAppInitDataMissing'),
    }
    return
  }

  const ok = await userProfileStore.bindTelegramMiniApp(miniAppInitData.value)
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.telegramBindFailed'),
    }
    return
  }

  await finishExternalIdentityMutation(t('personalCenter.security.telegramBindSuccess'))
}

const startTelegramOidcBind = async () => {
  securityAlert.value = null
  try {
    sessionStorage.setItem('tg_oidc_intent', 'bind')
    const res = await userProfileAPI.telegramOidcBindStart()
    const url = String(res?.data?.data?.auth_url || '')
    if (!url) {
      securityAlert.value = {
        level: 'error',
        message: t('personalCenter.security.telegramOidcBindFailed'),
      }
      return
    }
    window.location.href = url
  } catch (err: any) {
    securityAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.security.telegramOidcBindFailed'),
    }
  }
}

const handleUnbindTelegram = async () => {
  securityAlert.value = null
  if (!canUnbindTelegram.value) {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.telegramUnbindDisabledTip'),
    }
    return
  }

  const ok = await userProfileStore.unbindTelegram()
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.telegramUnbindFailed'),
    }
    return
  }
  await finishExternalIdentityMutation(t('personalCenter.security.telegramUnbindSuccess'))
  renderTelegramWidget()
}

/* ==================== Google 绑定 ==================== */
const handleGoogleBind = async (credential: string) => {
  if (userProfileStore.bindingGoogle) return
  securityAlert.value = null
  const normalizedCredential = String(credential || '').trim()
  if (normalizedCredential === '') {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.googleInvalidCredential'),
    }
    return
  }

  const ok = await userProfileStore.bindGoogle(normalizedCredential)
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.googleBindFailed'),
    }
    return
  }
  await finishExternalIdentityMutation(t('personalCenter.security.googleBindSuccess'))
}

const handleGoogleScriptError = () => {
  securityAlert.value = {
    level: 'error',
    message: t('personalCenter.security.googleWidgetLoadFailed'),
  }
}

const prepareGoogleRedirectBind = async () => {
  const response = await userProfileAPI.googleRedirectBindIntent()
  const preparedIntent = createGoogleRedirectPreparedIntent(response.data.data)
  if (!preparedIntent) {
    throw new Error('Google redirect state is invalid')
  }
  const intent = createGoogleRedirectIntent(
    'bind',
    '/me/security',
    preparedIntent.issuedAt,
  )
  storeGoogleRedirectIntent(getGoogleRedirectSessionStorage(), intent)
  return preparedIntent
}

const handleUnbindGoogle = async () => {
  securityAlert.value = null
  if (!canUnbindGoogle.value) {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.googleUnbindDisabledTip'),
    }
    return
  }

  const ok = await userProfileStore.unbindGoogle()
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.googleUnbindFailed'),
    }
    return
  }
  await finishExternalIdentityMutation(t('personalCenter.security.googleUnbindSuccess'))
}

/* ==================== 两步验证（2FA） ==================== */
interface TwoFactorStatus {
  enabled: boolean
  enabled_at?: string | null
  recovery_codes_remaining: number
  recovery_codes_total: number
}

interface TwoFactorSetupResult {
  secret: string
  otpauth_url: string
  expires_at: string
}

const status = ref<TwoFactorStatus | null>(null)
const setupResult = ref<TwoFactorSetupResult | null>(null)
const qrcodeDataUrl = ref('')
const enableCode = ref('')
const recoveryCodes = ref<string[]>([])
const copied = ref(false)
const loading = ref(false)
const twofaAlert = ref<PageAlert | null>(null)

const mode = ref<'idle' | 'disable' | 'regenerate'>('idle')
const disableMode = ref<'code' | 'recovery'>('code')
const disableCode = ref('')
const disableRecovery = ref('')
const regenerateCode = ref('')

const disableModeItems = computed<QzSegmentItem[]>(() => [
  { key: 'code', label: t('personalCenter.security.twofa.useCode') },
  { key: 'recovery', label: t('personalCenter.security.twofa.useRecovery') },
])

const onDisableModeChange = (key: string) => {
  disableMode.value = key === 'recovery' ? 'recovery' : 'code'
}

const refreshTwoFactorStatus = async () => {
  try {
    const res = await userTotpAPI.status()
    status.value = res.data.data
  } catch (err: any) {
    twofaAlert.value = { level: 'error', message: err?.message || t('personalCenter.security.twofa.loadFailed') }
  }
}

const startSetup = async () => {
  twofaAlert.value = null
  loading.value = true
  try {
    const res = await userTotpAPI.setup()
    setupResult.value = res.data.data
    qrcodeDataUrl.value = await QRCode.toDataURL(setupResult.value!.otpauth_url, { width: 240, margin: 1 })
    enableCode.value = ''
  } catch (err: any) {
    twofaAlert.value = { level: 'error', message: err?.message || t('personalCenter.security.twofa.setupFailed') }
  } finally {
    loading.value = false
  }
}

const cancelSetup = () => {
  setupResult.value = null
  qrcodeDataUrl.value = ''
  enableCode.value = ''
  twofaAlert.value = null
}

const submitEnable = async () => {
  twofaAlert.value = null
  const code = enableCode.value.trim()
  if (code === '') {
    twofaAlert.value = { level: 'warning', message: t('personalCenter.security.twofa.codeRequired') }
    return
  }
  loading.value = true
  try {
    const res = await userTotpAPI.enable({ code })
    const data = res.data.data || {}
    recoveryCodes.value = data.recovery_codes || []
    if (data.token) {
      // 启用 2FA 后旧 token 已失效，立即用服务端返回的新 token 替换，避免被中间件登出
      localStorage.setItem('user_token', data.token)
      userAuthStore.token = data.token
    }
    setupResult.value = null
    qrcodeDataUrl.value = ''
    enableCode.value = ''
    twofaAlert.value = { level: 'success', message: t('personalCenter.security.twofa.enableSuccess') }
    await refreshTwoFactorStatus()
  } catch (err: any) {
    twofaAlert.value = { level: 'error', message: err?.message || t('personalCenter.security.twofa.enableFailed') }
  } finally {
    loading.value = false
  }
}

const openDisable = () => {
  mode.value = 'disable'
  disableMode.value = 'code'
  disableCode.value = ''
  disableRecovery.value = ''
  twofaAlert.value = null
}

const openRegenerate = () => {
  mode.value = 'regenerate'
  regenerateCode.value = ''
  twofaAlert.value = null
}

const resetMode = () => {
  mode.value = 'idle'
  disableCode.value = ''
  disableRecovery.value = ''
  regenerateCode.value = ''
}

const submitDisable = async () => {
  twofaAlert.value = null
  const payload: { code?: string; recovery_code?: string } = {}
  if (disableMode.value === 'code') {
    const code = disableCode.value.trim()
    if (code === '') {
      twofaAlert.value = { level: 'warning', message: t('personalCenter.security.twofa.codeRequired') }
      return
    }
    payload.code = code
  } else {
    const rc = disableRecovery.value.trim()
    if (rc === '') {
      twofaAlert.value = { level: 'warning', message: t('personalCenter.security.twofa.recoveryRequired') }
      return
    }
    payload.recovery_code = rc
  }
  loading.value = true
  try {
    await userTotpAPI.disable(payload)
    twofaAlert.value = { level: 'success', message: t('personalCenter.security.twofa.disableSuccess') }
    resetMode()
    await refreshTwoFactorStatus()
  } catch (err: any) {
    twofaAlert.value = { level: 'error', message: err?.message || t('personalCenter.security.twofa.disableFailed') }
  } finally {
    loading.value = false
  }
}

const submitRegenerate = async () => {
  twofaAlert.value = null
  const code = regenerateCode.value.trim()
  if (code === '') {
    twofaAlert.value = { level: 'warning', message: t('personalCenter.security.twofa.codeRequired') }
    return
  }
  loading.value = true
  try {
    const res = await userTotpAPI.regenerateRecoveryCodes({ code })
    recoveryCodes.value = res.data.data?.recovery_codes || []
    resetMode()
    twofaAlert.value = { level: 'success', message: t('personalCenter.security.twofa.regenerateSuccess') }
    await refreshTwoFactorStatus()
  } catch (err: any) {
    twofaAlert.value = { level: 'error', message: err?.message || t('personalCenter.security.twofa.regenerateFailed') }
  } finally {
    loading.value = false
  }
}

const copyRecoveryCodes = async () => {
  try {
    await navigator.clipboard.writeText(recoveryCodes.value.join('\n'))
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch (err) {
    /* ignore */
  }
}

const acknowledgeRecoveryCodes = () => {
  recoveryCodes.value = []
}

watch(setupResult, async (val) => {
  if (val) {
    qrcodeDataUrl.value = await QRCode.toDataURL(val.otpauth_url, { width: 240, margin: 1 })
  } else {
    qrcodeDataUrl.value = ''
  }
})

/* ==================== 生命周期 ==================== */
onMounted(async () => {
  await Promise.all([
    appStore.loadConfig(),
    userProfileStore.loadRecentLoginLogs(10),
    userProfileStore.loadTelegramBinding(),
    userProfileStore.loadGoogleBinding(),
  ])
  const win = window as Window & Record<string, any>
  win[telegramCallbackName] = handleTelegramBind
  // 等绑定状态渲染到 DOM 后再注入 Telegram 官方脚本（内联后 ref 依赖本次渲染完成）
  await nextTick()
  renderTelegramWidget()

  if (route.query.tgBound === '1') {
    await finishExternalIdentityMutation(t('personalCenter.security.telegramBoundOk'))
    const nextQuery = { ...route.query }
    delete nextQuery.tgBound
    router.replace({ path: route.path, query: nextQuery })
  } else if (route.query.googleBound === '1') {
    await finishExternalIdentityMutation(t('personalCenter.security.googleBindSuccess'))
    const nextQuery = { ...route.query }
    delete nextQuery.googleBound
    router.replace({ path: route.path, query: nextQuery })
  }
})

onMounted(() => {
  if (canManagePassword.value) {
    void refreshTwoFactorStatus()
  }
})

// 2FA 区随 canManagePassword 出现时才拉取状态（等价经典版子组件挂载时机）
watch(canManagePassword, (value) => {
  if (value) {
    void refreshTwoFactorStatus()
  }
})

onUnmounted(() => {
  const win = window as Window & Record<string, any>
  delete win[telegramCallbackName]
  clearTelegramWidget()
  if (cooldownTimer !== null) {
    window.clearInterval(cooldownTimer)
    cooldownTimer = null
  }
})

watch([showTelegramWidget, telegramBotUsername], () => {
  renderTelegramWidget()
}, { flush: 'post' })
</script>

<style scoped>
/* 安全面板：区块 .qz-panel + .qz-section-head、提示条 .qz-alert、表单 .qz-field 三件套、状态 .qz-badge、按钮 .qz-btn / QzButton
   来自 styles/base.css；.qz-me-panel / .qz-me-sec / .qz-me-acts(.qz-me-acts-start) / .qz-me-form(-row) / .qz-me-rows / .qz-me-row-*
   来自 PersonalCenter.vue 的 me 共享块；关闭 2FA 的方式切换来自 QzSegment。这里只剩已绑定身份块、验证码同排、2FA 子面板与恢复码弹窗。
   恢复码弹窗 Teleport 到 body：scoped 规则跟随元素生效，token 由 body.qiaozhi-tokens 提供，弹窗盒子复用 .qz-panel。 */

/* 纵向内容流（绑定信息 / 提示 / 操作） */
.qz-me-stack {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-3);
}

/* ---------- 已绑定身份：头像 + 文本 ---------- */
.qz-me-ident {
  display: flex;
  align-items: center;
  gap: var(--qz-space-3);
  min-width: 0;
}

.qz-me-avatar {
  display: grid;
  flex: none;
  place-items: center;
  width: 52px;
  height: 52px;
  overflow: hidden;
  border-radius: var(--qz-radius-control);
  background: var(--qz-accent-subtle);
  color: var(--qz-accent-strong);
  font-size: var(--qz-text-xl);
  font-weight: 700;
}

.qz-me-avatar > img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.qz-me-ident-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.qz-me-ident-text b {
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  font-weight: 600;
  overflow-wrap: anywhere;
}

.qz-me-ident-text small {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

/* Telegram 官方登录组件挂载点 */
.qz-sec-widget {
  display: flex;
  align-items: center;
  min-height: var(--qz-control-md);
}

/* 输入框 + 发送验证码按钮同排（按钮与输入同为 md 高） */
.qz-sec-inline {
  display: flex;
  align-items: center;
  gap: var(--qz-space-2);
}

.qz-sec-inline > .qz-field-control {
  flex: 1 1 auto;
  min-width: 0;
}

.qz-sec-inline > .qz-btn {
  flex: none;
}

/* 6 位动态码输入居中 */
.qz-field-control.qz-sec-code {
  font-weight: 600;
  letter-spacing: 0.4em;
  text-align: center;
}

/* ---------- 2FA：浅底子面板（状态信息 / 关闭 / 重新生成 / 绑定流程） ---------- */
.qz-sec-sub {
  display: grid;
  gap: var(--qz-space-3);
  padding: var(--qz-space-4);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
}

/* 二维码 + 密钥；二维码外框固定白底（--qz-qr-frame），暗色下保住静区 */
.qz-sec-qr {
  display: flex;
  justify-content: center;
}

.qz-sec-qr img {
  width: 190px;
  height: 190px;
  padding: var(--qz-space-2);
  border: 1px solid var(--qz-border-strong);
  border-radius: var(--qz-radius-control);
  background: var(--qz-qr-frame);
}

.qz-sec-secret {
  display: block;
  padding: var(--qz-space-3) var(--qz-space-4);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface);
  color: var(--qz-text);
  font-family: var(--qz-font-mono);
  font-size: var(--qz-text-sm);
  letter-spacing: 0.06em;
  text-align: center;
  overflow-wrap: anywhere;
}

/* ---------- 恢复码弹窗（Teleport） ---------- */
.qz-sec-mask {
  position: fixed;
  inset: 0;
  z-index: 1200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--qz-space-4);
  background: color-mix(in oklab, var(--qz-text) 55%, transparent);
  backdrop-filter: blur(4px);
}

/* 暗色里 --qz-text 是近白，遮罩改用页面底色压暗 */
.dark .qz-sec-mask {
  background: color-mix(in oklab, var(--qz-bg) 72%, transparent);
}

.qz-panel.qz-sec-modal {
  width: min(460px, 100%);
  max-height: calc(100vh - var(--qz-space-8));
  overflow-y: auto;
  box-shadow: var(--qz-shadow-floating);
}

.qz-sec-modal h4 {
  margin: 0 0 var(--qz-space-2);
  color: var(--qz-text);
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: var(--qz-leading-body);
  letter-spacing: var(--qz-tracking-display);
}

.qz-sec-codes {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--qz-space-2);
  margin-top: var(--qz-space-4);
}

.qz-sec-codes code {
  padding: var(--qz-space-2) var(--qz-space-3);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-muted);
  color: var(--qz-text);
  font-family: var(--qz-font-mono);
  font-size: var(--qz-text-xs);
  text-align: center;
  overflow-wrap: anywhere;
}

.qz-sec-modal-acts {
  display: flex;
  gap: var(--qz-space-2);
  margin-top: var(--qz-space-4);
}

.qz-sec-modal-acts > * {
  flex: 1 1 0;
}
</style>
