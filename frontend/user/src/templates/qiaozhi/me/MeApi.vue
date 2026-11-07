<template>
  <div class="qz-me-panel">
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.apiPanel.title') }}</h2>
          <small>{{ t('personalCenter.apiPanel.subtitle') }}</small>
        </div>
        <div class="qz-me-acts">
          <span class="qz-chip">
            <Key aria-hidden="true" />
            {{ t('personalCenter.tabs.api') }}
          </span>
        </div>
      </div>

      <p
        v-if="panelAlert"
        class="qz-alert"
        :class="alertToneClass(panelAlert.level)"
        :role="panelAlert.level === 'error' ? 'alert' : 'status'"
      >
        {{ panelAlert.message }}
      </p>

      <!-- Loading -->
      <div v-if="loading" class="qz-me-rows" aria-hidden="true">
        <div v-for="idx in 3" :key="idx" class="qz-skeleton qz-me-sk"></div>
      </div>

      <!-- 无凭证 -->
      <div v-else-if="!credential" class="qz-me-cta">
        <p class="qz-field-hint">{{ t('personalCenter.apiPanel.noCredential') }}</p>
        <QzButton :loading="submitting" @click="handleApply">
          {{ submitting ? t('personalCenter.apiPanel.applying') : t('personalCenter.apiPanel.apply') }}
        </QzButton>
      </div>

      <!-- 审核中 -->
      <div v-else-if="credential.status === 'pending_review'" class="qz-me-rows">
        <div class="qz-card qz-me-row">
          <div class="qz-me-row-main">
            <b>{{ t('personalCenter.apiPanel.pendingTitle') }}</b>
            <small>{{ t('personalCenter.apiPanel.pendingDesc') }}</small>
          </div>
          <div class="qz-me-row-side">
            <span class="qz-badge is-warning" aria-hidden="true"><AlertTriangle /></span>
          </div>
        </div>
      </div>

      <!-- 已拒绝 -->
      <div v-else-if="credential.status === 'rejected'" class="qz-me-rows">
        <div class="qz-card qz-me-row">
          <div class="qz-me-row-main">
            <b>{{ t('personalCenter.apiPanel.rejectedTitle') }}</b>
            <small v-if="credential.reject_reason">
              {{ t('personalCenter.apiPanel.rejectReason', { reason: credential.reject_reason }) }}
            </small>
          </div>
          <div class="qz-me-row-side">
            <span class="qz-badge is-danger" aria-hidden="true"><XCircle /></span>
            <QzButton size="sm" :loading="submitting" @click="handleApply">
              {{ submitting ? t('personalCenter.apiPanel.applying') : t('personalCenter.apiPanel.reapply') }}
            </QzButton>
          </div>
        </div>
      </div>

      <!-- 已通过 / 使用中 -->
      <div v-else class="qz-me-rows">
        <!-- 首次生成提示（Secret 从未查看过） -->
        <div v-if="!hasViewedSecret && !newSecret" class="qz-card qz-me-row">
          <div class="qz-me-row-main">
            <b>{{ t('personalCenter.apiPanel.approvedNoticeTitle') }}</b>
            <small>{{ t('personalCenter.apiPanel.approvedNoticeDesc') }}</small>
          </div>
          <div class="qz-me-row-side">
            <QzButton size="sm" :loading="submitting" @click="handleFirstGenerate">
              {{ submitting ? t('personalCenter.apiPanel.regenerating') : t('personalCenter.apiPanel.generateSecret') }}
            </QzButton>
          </div>
        </div>

        <!-- API Key -->
        <div class="qz-card qz-me-row">
          <div class="qz-me-row-main">
            <small>API Key</small>
            <b><code>{{ credential.api_key || '-' }}</code></b>
          </div>
          <div class="qz-me-row-side">
            <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm" @click="copyToClipboard(credential.api_key || '')">
              {{ t('personalCenter.apiPanel.copy') }}
            </button>
          </div>
        </div>

        <!-- API Secret -->
        <div class="qz-card qz-me-row">
          <div class="qz-me-row-main">
            <small>API Secret</small>
            <b><code>{{ maskedSecret }}</code></b>
            <small>{{ t('personalCenter.apiPanel.secretHint') }}</small>
          </div>
          <div class="qz-me-row-side">
            <button type="button" class="qz-btn qz-btn--secondary qz-btn--sm" :disabled="submitting" @click="handleRegenerate">
              {{ t('personalCenter.apiPanel.regenerate') }}
            </button>
          </div>
        </div>

        <!-- 新生成的 Secret（仅显示一次） -->
        <div v-if="newSecret" class="qz-card qz-me-row qz-me-fresh">
          <div class="qz-me-row-main">
            <b><code>{{ newSecret }}</code></b>
            <small>{{ t('personalCenter.apiPanel.newSecretWarning') }}</small>
          </div>
          <div class="qz-me-row-side">
            <span class="qz-badge is-success">
              <Check aria-hidden="true" />
              {{ t('personalCenter.apiPanel.newSecretTitle') }}
            </span>
            <button type="button" class="qz-btn qz-btn--primary qz-btn--sm" @click="copyToClipboard(newSecret)">
              {{ t('personalCenter.apiPanel.copySecret') }}
            </button>
          </div>
        </div>

        <!-- 启用 / 禁用 -->
        <div class="qz-card qz-me-row">
          <div class="qz-me-row-main">
            <b>{{ t('personalCenter.apiPanel.statusLabel') }}</b>
            <small>{{ credential.is_active ? t('personalCenter.apiPanel.statusEnabled') : t('personalCenter.apiPanel.statusDisabled') }}</small>
          </div>
          <div class="qz-me-row-side">
            <button
              type="button"
              class="qz-me-switch"
              :class="{ 'is-on': credential.is_active }"
              role="switch"
              :aria-checked="credential.is_active"
              :disabled="submitting"
              :aria-label="t('personalCenter.apiPanel.statusLabel')"
              @click="handleToggleStatus"
            >
              <span aria-hidden="true"></span>
            </button>
          </div>
        </div>
      </div>

      <!-- 重新生成确认弹窗 -->
      <Teleport to="body">
        <div
          v-if="showConfirm"
          class="qz-me-modal-mask"
          role="dialog"
          aria-modal="true"
          :aria-label="t('personalCenter.apiPanel.regenerateTitle')"
          @click.self="showConfirm = false"
        >
          <div ref="confirmEl" class="qz-panel qz-me-modal" tabindex="-1" @keydown.esc="showConfirm = false">
            <h3>{{ t('personalCenter.apiPanel.regenerateTitle') }}</h3>
            <p>{{ t('personalCenter.apiPanel.regenerateDesc') }}</p>
            <div class="qz-me-modal-acts">
              <button type="button" class="qz-btn qz-btn--secondary" @click="showConfirm = false">
                {{ t('personalCenter.apiPanel.cancel') }}
              </button>
              <QzButton variant="danger" :loading="submitting" @click="confirmRegenerate">
                {{ submitting ? t('personalCenter.apiPanel.regenerating') : t('personalCenter.apiPanel.regenerateConfirm') }}
              </QzButton>
            </div>
          </div>
        </div>
      </Teleport>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlertTriangle, Check, Key, XCircle } from 'lucide-vue-next'
import { apiCredentialAPI } from '../../../api'
import type { PageAlert } from '../../../utils/alerts'
import QzButton from '../components/QzButton.vue'

const { t } = useI18n()

interface CredentialData {
  id: number
  api_key: string
  api_secret_masked?: string
  status: string
  is_active: boolean
  reject_reason?: string
  created_at?: string
  updated_at?: string
}

const loading = ref(true)
const submitting = ref(false)
const credential = ref<CredentialData | null>(null)
const panelAlert = ref<PageAlert | null>(null)
const newSecret = ref('')
const showConfirm = ref(false)
const hasViewedSecret = ref(false)
const confirmEl = ref<HTMLElement | null>(null)

const maskedSecret = computed(() => {
  if (credential.value?.api_secret_masked) {
    return credential.value.api_secret_masked
  }
  return '••••••••••••'
})

const secretViewedKey = 'api_secret_viewed'

/** 提示条 tone（base .qz-alert）：success → is-success、warning → is-warning、error → is-danger */
const alertToneClass = (level: PageAlert['level']) => {
  if (level === 'success') return 'is-success'
  if (level === 'error') return 'is-danger'
  return 'is-warning'
}

const loadCredential = async () => {
  loading.value = true
  panelAlert.value = null
  try {
    const response = await apiCredentialAPI.getMy()
    const data = response.data.data
    // 后端返回 {status: "none"} 表示无凭证
    if (!data || data.status === 'none') {
      credential.value = null
    } else {
      // 将 api_secret_tail 映射为前端需要的 api_secret_masked 格式
      if (data.api_secret_tail) {
        data.api_secret_masked = '••••••••' + data.api_secret_tail
      }
      credential.value = data
      // 检查用户是否已生成过密钥
      hasViewedSecret.value = localStorage.getItem(secretViewedKey) === String(data.id)
    }
  } catch (err: any) {
    // 404 or no credential is normal
    credential.value = null
  } finally {
    loading.value = false
  }
}

const handleApply = async () => {
  submitting.value = true
  panelAlert.value = null
  try {
    await apiCredentialAPI.apply()
    panelAlert.value = {
      level: 'success',
      message: t('personalCenter.apiPanel.applySuccess'),
    }
    await loadCredential()
  } catch (err: any) {
    panelAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.apiPanel.applyFailed'),
    }
  } finally {
    submitting.value = false
  }
}

const handleFirstGenerate = async () => {
  submitting.value = true
  panelAlert.value = null
  try {
    const response = await apiCredentialAPI.regenerate()
    const data = response.data.data
    if (data?.api_secret) {
      newSecret.value = data.api_secret
    }
    if (credential.value) {
      localStorage.setItem(secretViewedKey, String(credential.value.id))
      hasViewedSecret.value = true
    }
    await loadCredential()
    panelAlert.value = {
      level: 'success',
      message: t('personalCenter.apiPanel.generateSuccess'),
    }
  } catch (err: any) {
    panelAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.apiPanel.regenerateFailed'),
    }
  } finally {
    submitting.value = false
  }
}

const handleRegenerate = () => {
  newSecret.value = ''
  showConfirm.value = true
}

const confirmRegenerate = async () => {
  submitting.value = true
  panelAlert.value = null
  try {
    const response = await apiCredentialAPI.regenerate()
    const data = response.data.data
    if (data?.api_secret) {
      newSecret.value = data.api_secret
    }
    if (credential.value) {
      localStorage.setItem(secretViewedKey, String(credential.value.id))
      hasViewedSecret.value = true
    }
    // Reload credential to get updated masked secret
    await loadCredential()
    panelAlert.value = {
      level: 'success',
      message: t('personalCenter.apiPanel.regenerateSuccess'),
    }
  } catch (err: any) {
    panelAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.apiPanel.regenerateFailed'),
    }
  } finally {
    submitting.value = false
    showConfirm.value = false
  }
}

const handleToggleStatus = async () => {
  if (!credential.value) return
  submitting.value = true
  panelAlert.value = null
  const newStatus = !credential.value.is_active
  try {
    await apiCredentialAPI.updateStatus({ is_active: newStatus })
    credential.value.is_active = newStatus
    panelAlert.value = {
      level: 'success',
      message: newStatus ? t('personalCenter.apiPanel.enabled') : t('personalCenter.apiPanel.disabled'),
    }
  } catch (err: any) {
    panelAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.apiPanel.toggleFailed'),
    }
  } finally {
    submitting.value = false
  }
}

const copyToClipboard = async (text: string) => {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    panelAlert.value = {
      level: 'success',
      message: t('personalCenter.apiPanel.copied'),
    }
  } catch {
    panelAlert.value = {
      level: 'error',
      message: t('personalCenter.apiPanel.copyFailed'),
    }
  }
}

// 弹窗打开时聚焦面板，Esc 才能生效
watch(showConfirm, async (open) => {
  if (!open) return
  await nextTick()
  confirmEl.value?.focus()
})

onMounted(() => {
  loadCredential()
})
</script>

<style scoped>
/* API 面板：区块 .qz-panel + .qz-section-head、提示条 .qz-alert、状态 .qz-badge、按钮 .qz-btn / QzButton、骨架 .qz-skeleton
   来自 styles/base.css；.qz-me-panel / .qz-me-sec / .qz-me-acts / .qz-me-cta / .qz-me-rows / .qz-me-row-*（含 code 等宽）/
   .qz-me-sk 来自 PersonalCenter.vue 的 me 共享块。这里只剩新 Secret 入场、启用开关与确认弹窗。
   确认弹窗 Teleport 到 body：scoped 规则跟随元素生效，token 由 body.qiaozhi-tokens 提供，弹窗盒子复用 .qz-panel。 */

/* 新 Secret 一次性展示的入场 */
.qz-me-fresh {
  animation: qz-me-fresh-in var(--qz-duration-considered) var(--qz-ease-enter) both;
}

@keyframes qz-me-fresh-in {
  from {
    opacity: 0;
    transform: translateY(var(--qz-space-2)) scale(0.99);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

/* ---------- 启用 / 禁用开关 ---------- */
.qz-me-switch {
  position: relative;
  display: inline-flex;
  flex: none;
  align-items: center;
  width: 46px;
  height: 26px;
  padding: 2px;
  border: 0;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-border-strong);
  cursor: pointer;
  transition: background-color var(--qz-duration-standard) var(--qz-ease-standard);
}

.qz-me-switch > span {
  display: inline-block;
  width: 22px;
  height: 22px;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-surface);
  box-shadow: var(--qz-shadow-raised);
  transform: translateX(0);
  transition: transform var(--qz-duration-standard) var(--qz-ease-standard);
}

.qz-me-switch.is-on {
  background: var(--qz-accent);
}

.qz-me-switch.is-on > span {
  transform: translateX(20px);
}

.qz-me-switch:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ---------- 重新生成确认弹窗（Teleport） ---------- */
.qz-me-modal-mask {
  position: fixed;
  inset: 0;
  z-index: 1200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--qz-space-4);
  background: color-mix(in oklab, var(--qz-text) 55%, transparent);
  backdrop-filter: blur(4px);
  animation: qz-me-mask-in var(--qz-duration-standard) var(--qz-ease-standard);
}

/* 暗色里 --qz-text 是近白，遮罩改用页面底色压暗 */
.dark .qz-me-modal-mask {
  background: color-mix(in oklab, var(--qz-bg) 72%, transparent);
}

@keyframes qz-me-mask-in {
  from {
    opacity: 0;
  }
}

.qz-panel.qz-me-modal {
  width: min(380px, 100%);
  box-shadow: var(--qz-shadow-floating);
  outline: none;
  animation: qz-me-modal-rise var(--qz-duration-standard) var(--qz-ease-enter);
}

@keyframes qz-me-modal-rise {
  from {
    opacity: 0;
    transform: translateY(var(--qz-space-4)) scale(0.98);
  }
}

.qz-me-modal h3 {
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-lg);
  font-weight: 700;
  line-height: var(--qz-leading-body);
  letter-spacing: var(--qz-tracking-display);
}

.qz-me-modal p {
  margin: var(--qz-space-2) 0 0;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
}

.qz-me-modal-acts {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--qz-space-2);
  margin-top: var(--qz-space-5);
}

@media (prefers-reduced-motion: reduce) {
  .qz-me-fresh,
  .qz-me-modal-mask,
  .qz-panel.qz-me-modal {
    animation: none;
  }

  .qz-me-switch,
  .qz-me-switch > span {
    transition: none;
  }
}
</style>
