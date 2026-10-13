<template>
  <div class="j8-me-panel">
    <!-- ==================== 安全总览 ==================== -->
    <div class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ t('personalCenter.security.title') }}</h2>
          <small>{{ requiresOldEmailCode ? t('personalCenter.security.subtitle') : t('personalCenter.security.subtitleBindOnly') }}</small>
        </div>
        <div class="j8-me-card-acts">
          <span class="j8-me-pill is-ok">{{ t('personalCenter.tabs.security') }}</span>
        </div>
      </div>
      <p
        v-if="securityAlert"
        class="j8-ic-alert"
        :class="alertClass(securityAlert.level)"
        :role="securityAlert.level === 'error' ? 'alert' : 'status'"
      >
        {{ securityAlert.message }}
      </p>
    </div>

    <!-- ==================== Telegram 绑定 ==================== -->
    <div class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ t('personalCenter.security.telegramTitle') }}</h2>
          <small>{{ telegramEnabled ? t('personalCenter.security.telegramSubtitle') : t('personalCenter.security.telegramDisabledTip') }}</small>
        </div>
        <div class="j8-me-card-acts">
          <span class="j8-me-pill" :class="telegramBound ? 'is-ok' : 'is-warn'">
            {{ telegramBound ? t('personalCenter.security.telegramBound') : t('personalCenter.security.telegramUnbound') }}
          </span>
        </div>
      </div>

      <div class="j8-me-rows">
        <p v-if="userProfileStore.loadingTelegramBinding" class="j8-me-hint">
          {{ t('personalCenter.security.telegramLoading') }}
        </p>

        <template v-else-if="telegramBound">
          <div class="j8-me-ident">
            <span v-if="telegramAvatarUrl" class="j8-me-avatar">
              <img :src="telegramAvatarUrl" alt="Telegram Avatar" />
            </span>
            <div class="j8-me-row-main">
              <b>{{ telegramDisplayName }}</b>
              <small>{{ t('personalCenter.security.telegramBindID', { id: telegramProviderUserId }) }}</small>
              <small>{{ t('personalCenter.security.telegramBindTime', { time: formatDate(userProfileStore.telegramBinding?.auth_at) }) }}</small>
            </div>
          </div>

          <div class="j8-me-card-acts">
            <J8Button
              type="button"
              variant="secondary"
              size="sm"
              :loading="userProfileStore.unbindingTelegram"
              :disabled="!canUnbindTelegram"
              @click="handleUnbindTelegram"
            >
              {{ userProfileStore.unbindingTelegram ? t('personalCenter.security.telegramUnbinding') : t('personalCenter.security.telegramUnbind') }}
            </J8Button>
          </div>
          <p v-if="!canUnbindTelegram" class="j8-me-hint">
            {{ t('personalCenter.security.telegramUnbindDisabledTip') }}
          </p>

          <template v-if="showTelegramMiniAppEntry">
            <p class="j8-me-hint">{{ t('personalCenter.security.telegramMiniAppEntryHint') }}</p>
            <div class="j8-me-card-acts">
              <J8Button type="button" variant="secondary" size="sm" @click="openTelegramMiniAppEntry">
                {{ t('personalCenter.security.telegramMiniAppEntryAction') }}
              </J8Button>
            </div>
          </template>
        </template>

        <template v-else>
          <p class="j8-me-hint">
            {{ telegramEnabled
              ? (showMiniAppBindAction ? t('personalCenter.security.telegramMiniAppBindHint') : t('personalCenter.security.telegramUnboundTip'))
              : t('personalCenter.security.telegramDisabledTip') }}
          </p>

          <div v-if="showMiniAppBindAction" class="j8-me-card-acts">
            <J8Button
              type="button"
              :loading="userProfileStore.bindingTelegram"
              :disabled="miniAppInitData === ''"
              @click="handleTelegramMiniAppBind"
            >
              {{ userProfileStore.bindingTelegram ? t('personalCenter.security.telegramMiniAppBinding') : t('personalCenter.security.telegramMiniAppBindAction') }}
            </J8Button>
          </div>
          <div v-else-if="showTelegramWidget" ref="telegramWidgetRef" class="j8-sec-widget"></div>
          <template v-else-if="showTelegramOidcBind">
            <div class="j8-me-card-acts">
              <J8Button
                type="button"
                :disabled="userProfileStore.bindingTelegram"
                @click="startTelegramOidcBind"
              >
                {{ t('personalCenter.security.telegramOidcBindButton') }}
              </J8Button>
            </div>
            <p class="j8-me-hint">{{ t('personalCenter.security.telegramOidcBindHint') }}</p>
          </template>

          <template v-if="showTelegramMiniAppEntry">
            <p class="j8-me-hint">{{ t('personalCenter.security.telegramMiniAppEntryHint') }}</p>
            <div class="j8-me-card-acts">
              <J8Button type="button" variant="secondary" size="sm" @click="openTelegramMiniAppEntry">
                {{ t('personalCenter.security.telegramMiniAppEntryAction') }}
              </J8Button>
            </div>
          </template>
        </template>
      </div>
    </div>

    <!-- ==================== Google 绑定 ==================== -->
    <div class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ t('personalCenter.security.googleTitle') }}</h2>
          <small>{{ googleAuthorizationEnabled ? t('personalCenter.security.googleSubtitle') : t('personalCenter.security.googleDisabledTip') }}</small>
        </div>
        <div class="j8-me-card-acts">
          <span class="j8-me-pill" :class="googleBound ? 'is-ok' : 'is-warn'">
            {{ googleBound ? t('personalCenter.security.googleBound') : t('personalCenter.security.googleUnbound') }}
          </span>
        </div>
      </div>

      <div class="j8-me-rows">
        <p v-if="userProfileStore.loadingGoogleBinding" class="j8-me-hint">
          {{ t('personalCenter.security.googleLoading') }}
        </p>

        <template v-else-if="googleBound">
          <div class="j8-me-ident">
            <span class="j8-me-avatar">
              <img
                v-if="googleAvatarUrl"
                :src="googleAvatarUrl"
                :alt="t('personalCenter.security.googleAvatarAlt')"
              />
              <template v-else>G</template>
            </span>
            <div class="j8-me-row-main">
              <b>{{ googleDisplayName }}</b>
              <small v-if="googleEmail">{{ googleEmail }}</small>
              <small>{{ t('personalCenter.security.googleBindID', { id: googleProviderUserId }) }}</small>
              <small>{{ t('personalCenter.security.googleBindTime', { time: formatDate(userProfileStore.googleBinding?.auth_at) }) }}</small>
            </div>
          </div>

          <div class="j8-me-card-acts">
            <J8Button
              type="button"
              variant="secondary"
              size="sm"
              :loading="userProfileStore.unbindingGoogle"
              :disabled="!canUnbindGoogle"
              @click="handleUnbindGoogle"
            >
              {{ userProfileStore.unbindingGoogle ? t('personalCenter.security.googleUnbinding') : t('personalCenter.security.googleUnbind') }}
            </J8Button>
          </div>
          <p v-if="!canUnbindGoogle" class="j8-me-hint">
            {{ t('personalCenter.security.googleUnbindDisabledTip') }}
          </p>
        </template>

        <template v-else>
          <p class="j8-me-hint">
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
          <p v-if="userProfileStore.bindingGoogle" class="j8-me-hint" role="status">
            {{ t('personalCenter.security.googleBinding') }}
          </p>
        </template>
      </div>
    </div>

    <!-- ==================== 邮箱换绑 / 绑定 ==================== -->
    <div class="j8-me-card">
      <form class="j8-ic" novalidate @submit.prevent="handleChangeEmail">
        <div class="j8-ic-field">
          <label for="qz-me-sec-current-email">{{ t('personalCenter.security.currentEmailLabel') }}</label>
          <input id="qz-me-sec-current-email" :value="currentEmailDisplay" type="email" disabled />
          <p v-if="!requiresOldEmailCode" class="j8-ic-hint">
            {{ t('personalCenter.security.bindOnlyTip') }}
          </p>
        </div>

        <div class="j8-ic-field">
          <label for="qz-me-sec-new-email">{{ t('personalCenter.security.newEmailLabel') }}</label>
          <input
            id="qz-me-sec-new-email"
            v-model="securityForm.newEmail"
            type="email"
            autocomplete="email"
            :placeholder="t('personalCenter.security.newEmailPlaceholder')"
          />
        </div>

        <div :class="requiresOldEmailCode ? 'j8-ic-row' : ''">
          <div v-if="requiresOldEmailCode" class="j8-ic-field">
            <label for="qz-me-sec-old-code">{{ t('personalCenter.security.oldCodeLabel') }}</label>
            <div class="j8-sec-inline">
              <input
                id="qz-me-sec-old-code"
                v-model="securityForm.oldCode"
                inputmode="numeric"
                :placeholder="t('personalCenter.security.codePlaceholder')"
              />
              <button
                type="button"
                class="j8-me-act"
                :disabled="userProfileStore.sendingCode || oldCodeCooldown > 0"
                @click="handleSendOldCode"
              >
                {{ oldCodeCooldown > 0 ? t('personalCenter.security.countdown', { seconds: oldCodeCooldown }) : t('personalCenter.security.sendOldCode') }}
              </button>
            </div>
          </div>

          <div class="j8-ic-field">
            <label for="qz-me-sec-new-code">{{ t('personalCenter.security.newCodeLabel') }}</label>
            <div class="j8-sec-inline">
              <input
                id="qz-me-sec-new-code"
                v-model="securityForm.newCode"
                inputmode="numeric"
                :placeholder="t('personalCenter.security.codePlaceholder')"
              />
              <button
                type="button"
                class="j8-me-act"
                :disabled="userProfileStore.sendingCode || newCodeCooldown > 0"
                @click="handleSendNewCode"
              >
                {{ newCodeCooldown > 0 ? t('personalCenter.security.countdown', { seconds: newCodeCooldown }) : t('personalCenter.security.sendNewCode') }}
              </button>
            </div>
          </div>
        </div>

        <div class="j8-me-card-acts">
          <J8Button type="submit" :loading="userProfileStore.changingEmail">
            {{ userProfileStore.changingEmail
              ? t('personalCenter.security.submitting')
              : (requiresOldEmailCode ? t('personalCenter.security.submit') : t('personalCenter.security.bindSubmit')) }}
          </J8Button>
        </div>
      </form>
    </div>

    <!-- ==================== 最近登录记录 ==================== -->
    <div class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ t('personalCenter.security.loginLogsTitle') }}</h2>
          <small>{{ t('personalCenter.security.loginLogsTip') }}</small>
        </div>
      </div>

      <p v-if="userProfileStore.loadingLoginLogs" class="j8-me-hint">
        {{ t('personalCenter.security.loginLogsLoading') }}
      </p>
      <p v-else-if="userProfileStore.recentLoginLogs.length === 0" class="j8-me-hint">
        {{ t('personalCenter.security.loginLogsEmpty') }}
      </p>
      <div v-else class="j8-me-rows">
        <div v-for="item in userProfileStore.recentLoginLogs" :key="item.id" class="j8-me-row">
          <div class="j8-me-row-main">
            <b>{{ formatDate(item.created_at) }}</b>
            <small>{{ t('personalCenter.security.loginLogsIp') }}：{{ item.client_ip || '-' }}</small>
            <small>{{ t('personalCenter.security.loginLogsReason') }}：{{ loginReasonLabel(item.fail_reason) }}</small>
          </div>
          <div class="j8-me-row-side">
            <span class="j8-me-pill" :class="loginStatusTone(item.status)">{{ loginStatusLabel(item.status) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- ==================== 登录密码 ==================== -->
    <div v-if="canManagePassword" class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ requiresOldPassword ? t('personalCenter.security.passwordTitle') : t('personalCenter.security.setPasswordTitle') }}</h2>
          <small>{{ requiresOldPassword ? t('personalCenter.security.passwordSubtitle') : t('personalCenter.security.setPasswordSubtitle') }}</small>
        </div>
      </div>

      <form class="j8-ic" novalidate @submit.prevent="handleChangePassword">
        <div v-if="requiresOldPassword" class="j8-ic-field">
          <label for="qz-me-sec-old-password">{{ t('personalCenter.security.currentPasswordLabel') }}</label>
          <input
            id="qz-me-sec-old-password"
            v-model="passwordForm.oldPassword"
            type="password"
            autocomplete="current-password"
            :placeholder="t('personalCenter.security.passwordPlaceholder')"
          />
        </div>

        <div class="j8-ic-row">
          <div class="j8-ic-field">
            <label for="qz-me-sec-new-password">{{ t('personalCenter.security.newPasswordLabel') }}</label>
            <input
              id="qz-me-sec-new-password"
              v-model="passwordForm.newPassword"
              type="password"
              autocomplete="new-password"
              :placeholder="t('personalCenter.security.passwordPlaceholder')"
            />
          </div>
          <div class="j8-ic-field">
            <label for="qz-me-sec-confirm-password">{{ t('personalCenter.security.confirmPasswordLabel') }}</label>
            <input
              id="qz-me-sec-confirm-password"
              v-model="passwordForm.confirmPassword"
              type="password"
              autocomplete="new-password"
              :placeholder="t('personalCenter.security.passwordPlaceholder')"
            />
          </div>
        </div>

        <div class="j8-me-card-acts">
          <J8Button type="submit" :loading="userProfileStore.changingPassword">
            {{ userProfileStore.changingPassword
              ? (requiresOldPassword ? t('personalCenter.security.changePasswordSubmitting') : t('personalCenter.security.setPasswordSubmitting'))
              : (requiresOldPassword ? t('personalCenter.security.changePassword') : t('personalCenter.security.setPassword')) }}
          </J8Button>
        </div>
      </form>
    </div>

    <!-- ==================== 两步验证（2FA） ==================== -->
    <div v-if="canManagePassword" class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ t('personalCenter.security.twofa.title') }}</h2>
          <small>{{ t('personalCenter.security.twofa.subtitle') }}</small>
        </div>
        <div class="j8-me-card-acts">
          <span class="j8-me-pill" :class="{ 'is-ok': status?.enabled }">
            {{ status?.enabled ? t('personalCenter.security.twofa.statusEnabled') : t('personalCenter.security.twofa.statusDisabled') }}
          </span>
        </div>
      </div>

      <p
        v-if="twofaAlert"
        class="j8-ic-alert"
        :class="alertClass(twofaAlert.level)"
        :role="twofaAlert.level === 'error' ? 'alert' : 'status'"
      >
        {{ twofaAlert.message }}
      </p>

      <!-- 已启用：状态 + 关闭 / 重新生成恢复码 -->
      <div v-if="status?.enabled" class="j8-me-rows">
        <div class="j8-sec-sub">
          <p class="j8-me-hint">{{ t('personalCenter.security.twofa.enabledAt', { date: formatDate(status.enabled_at) }) }}</p>
          <p class="j8-me-hint">
            {{ t('personalCenter.security.twofa.recoveryRemaining', {
              remaining: status.recovery_codes_remaining,
              total: status.recovery_codes_total,
            }) }}
          </p>
        </div>

        <div class="j8-ic-row">
          <J8Button type="button" variant="secondary" @click="openRegenerate">
            {{ t('personalCenter.security.twofa.regenerateAction') }}
          </J8Button>
          <J8Button type="button" variant="danger" @click="openDisable">
            {{ t('personalCenter.security.twofa.disableAction') }}
          </J8Button>
        </div>

        <!-- 关闭 2FA 表单 -->
        <div v-if="mode === 'disable'" class="j8-sec-sub">
          <p class="j8-me-hint">{{ t('personalCenter.security.twofa.disableHint') }}</p>
          <J8Segment
            :model-value="disableMode"
            :items="disableModeItems"
            @update:model-value="onDisableModeChange"
          />
          <div class="j8-ic-field">
            <input
              v-if="disableMode === 'code'"
              v-model="disableCode"
              class="j8-sec-code"
              inputmode="numeric"
              maxlength="6"
              :aria-label="t('personalCenter.security.twofa.codePlaceholder')"
              :placeholder="t('personalCenter.security.twofa.codePlaceholder')"
            />
            <input
              v-else
              v-model="disableRecovery"
              autocomplete="off"
              :aria-label="t('personalCenter.security.twofa.recoveryPlaceholder')"
              :placeholder="t('personalCenter.security.twofa.recoveryPlaceholder')"
            />
          </div>
          <div class="j8-ic-row">
            <J8Button type="button" variant="danger" :loading="loading" @click="submitDisable">
              {{ loading ? t('personalCenter.security.twofa.disableSubmitting') : t('personalCenter.security.twofa.disableSubmit') }}
            </J8Button>
            <J8Button type="button" variant="secondary" @click="resetMode">
              {{ t('personalCenter.security.twofa.cancel') }}
            </J8Button>
          </div>
        </div>

        <!-- 重新生成恢复码表单 -->
        <div v-if="mode === 'regenerate'" class="j8-sec-sub">
          <p class="j8-me-hint">{{ t('personalCenter.security.twofa.regenerateHint') }}</p>
          <div class="j8-ic-field">
            <input
              v-model="regenerateCode"
              class="j8-sec-code"
              inputmode="numeric"
              maxlength="6"
              :aria-label="t('personalCenter.security.twofa.codePlaceholder')"
              :placeholder="t('personalCenter.security.twofa.codePlaceholder')"
            />
          </div>
          <div class="j8-ic-row">
            <J8Button type="button" :loading="loading" @click="submitRegenerate">
              {{ loading ? t('personalCenter.security.twofa.regenerateSubmitting') : t('personalCenter.security.twofa.regenerateSubmit') }}
            </J8Button>
            <J8Button type="button" variant="secondary" @click="resetMode">
              {{ t('personalCenter.security.twofa.cancel') }}
            </J8Button>
          </div>
        </div>
      </div>

      <!-- 未启用：绑定流程 -->
      <div v-else class="j8-me-rows">
        <p class="j8-me-hint">{{ t('personalCenter.security.twofa.notEnabledHint') }}</p>

        <div v-if="!setupResult" class="j8-me-card-acts">
          <J8Button type="button" :loading="loading" @click="startSetup">
            {{ loading ? t('personalCenter.security.twofa.startingSetup') : t('personalCenter.security.twofa.startSetup') }}
          </J8Button>
        </div>

        <div v-else class="j8-sec-sub">
          <p class="j8-me-hint">{{ t('personalCenter.security.twofa.setupStep1') }}</p>
          <div v-if="qrcodeDataUrl" class="j8-sec-qr">
            <img :src="qrcodeDataUrl" alt="2FA QR" />
          </div>
          <div class="j8-me-rows">
            <p class="j8-me-hint">{{ t('personalCenter.security.twofa.secret') }}</p>
            <code class="j8-sec-secret">{{ setupResult.secret }}</code>
          </div>

          <p class="j8-me-hint">{{ t('personalCenter.security.twofa.setupStep2') }}</p>
          <div class="j8-ic-field">
            <input
              v-model="enableCode"
              class="j8-sec-code"
              inputmode="numeric"
              maxlength="6"
              :aria-label="t('personalCenter.security.twofa.codePlaceholder')"
              :placeholder="t('personalCenter.security.twofa.codePlaceholder')"
            />
          </div>

          <div class="j8-ic-row">
            <J8Button type="button" :loading="loading" @click="submitEnable">
              {{ loading ? t('personalCenter.security.twofa.enableSubmitting') : t('personalCenter.security.twofa.enableSubmit') }}
            </J8Button>
            <J8Button type="button" variant="secondary" @click="cancelSetup">
              {{ t('personalCenter.security.twofa.cancel') }}
            </J8Button>
          </div>
        </div>
      </div>

      <!-- 恢复码展示弹窗（启用成功 / 重新生成成功后） -->
      <Teleport to="body">
        <div
          v-if="recoveryCodes.length > 0"
          class="j8-sec-mask"
          role="dialog"
          aria-modal="true"
          :aria-label="t('personalCenter.security.twofa.recoveryTitle')"
          @click.self="recoveryCodes = []"
        >
          <div class="j8-sec-modal">
            <h4>{{ t('personalCenter.security.twofa.recoveryTitle') }}</h4>
            <p class="j8-me-hint">{{ t('personalCenter.security.twofa.recoveryWarning') }}</p>
            <div class="j8-sec-codes">
              <code v-for="(c, idx) in recoveryCodes" :key="idx">{{ c }}</code>
            </div>
            <div class="j8-sec-modal-acts">
              <J8Button type="button" variant="secondary" @click="copyRecoveryCodes">
                {{ copied ? t('personalCenter.security.twofa.copied') : t('personalCenter.security.twofa.copy') }}
              </J8Button>
              <J8Button type="button" @click="acknowledgeRecoveryCodes">
                {{ t('personalCenter.security.twofa.acknowledge') }}
              </J8Button>
            </div>
          </div>
        </div>
      </Teleport>
    </div>
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
import J8Button from '../components/J8Button.vue'
import J8Segment, { type J8SegmentItem } from '../components/J8Segment.vue'
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

/** 警示条 tone：success → 品牌色底、warning → 琥珀底、error → 红底 */
const alertClass = (level: PageAlertLevel) => {
  if (level === 'error') return 'is-error'
  if (level === 'success') return 'j8-ok-alert'
  return ''
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
const loginStatusTone = (status?: string) => ((status || '').trim() === 'success' ? 'is-ok' : 'is-danger')
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

const disableModeItems = computed<J8SegmentItem[]>(() => [
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
/* 成功态警示条（复用 j8-ic-alert 底色体系，绿/品牌色 tone 与登录页一致） */
.j8-ic-alert.j8-ok-alert {
  background: var(--accent-soft);
  color: var(--accent-strong);
}

/* 绑定身份头像：复用 .j8-me-avatar 尺寸，图片铺满裁圆角 */
.j8-me-avatar {
  overflow: hidden;
}
.j8-me-avatar > img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* 输入框 + 按钮同排（发送验证码） */
.j8-sec-inline {
  display: flex;
  align-items: center;
  gap: 8px;
}
.j8-sec-inline > input {
  flex: 1 1 auto;
  min-width: 0;
}
.j8-sec-inline > .j8-me-act {
  flex: 0 0 auto;
}
button.j8-me-act {
  cursor: pointer;
}
button.j8-me-act:disabled {
  cursor: not-allowed;
}

/* Telegram 官方登录组件挂载点 */
.j8-sec-widget {
  display: flex;
  align-items: center;
  min-height: 44px;
}

/* 6 位动态码输入居中 */
.j8-sec-code {
  text-align: center;
  letter-spacing: 0.4em;
  font-weight: 700;
}

/* 功能区内的浅底子面板（状态信息 / 关闭 2FA / 重新生成 / 绑定流程） */
.j8-sec-sub {
  display: grid;
  gap: 10px;
  padding: 14px;
  border-radius: 14px;
  background: var(--surface-2);
  border: 1px solid var(--line);
}

/* 2FA 二维码 + 密钥 */
.j8-sec-qr {
  display: flex;
  justify-content: center;
}
.j8-sec-qr img {
  width: 190px;
  height: 190px;
  padding: 8px;
  border: 1px solid var(--line-strong);
  border-radius: 14px;
  background: #fff;
}
.j8-sec-secret {
  display: block;
  padding: 10px 12px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: var(--surface);
  color: var(--text);
  text-align: center;
  font-family: var(--font-mono);
  font-size: 13px;
  letter-spacing: 0.06em;
  overflow-wrap: anywhere;
}

/* 恢复码弹窗（Teleport 到 body，走 qiaozhi-tokens 变量） */
.j8-sec-mask {
  position: fixed;
  inset: 0;
  z-index: 1200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: rgba(10, 10, 10, 0.5);
  backdrop-filter: blur(4px);
}
.j8-sec-modal {
  width: min(460px, 100%);
  max-height: calc(100vh - 32px);
  overflow-y: auto;
  padding: 22px;
  border-radius: 20px;
  background: var(--surface);
  border: 1px solid var(--line);
  box-shadow: var(--shadow);
}
.j8-sec-modal h4 {
  margin: 0 0 8px;
  font-size: 16px;
  font-weight: 700;
  color: var(--text);
}
.j8-sec-codes {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-top: 14px;
}
.j8-sec-codes code {
  padding: 9px 10px;
  border: 1px solid var(--line);
  border-radius: 10px;
  background: var(--surface-2);
  color: var(--text);
  text-align: center;
  font-family: var(--font-mono);
  font-size: 12.5px;
}
.j8-sec-modal-acts {
  display: flex;
  gap: 8px;
  margin-top: 16px;
}
.j8-sec-modal-acts > * {
  flex: 1 1 0;
}
</style>
