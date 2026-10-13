<template>
  <div class="j8-me-panel">
    <div class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ t('personalCenter.apiPanel.title') }}</h2>
          <small>{{ t('personalCenter.apiPanel.subtitle') }}</small>
        </div>
        <div class="j8-me-card-acts">
          <span class="j8-me-pill is-ok">
            <Key aria-hidden="true" />
            {{ t('personalCenter.tabs.api') }}
          </span>
        </div>
      </div>

      <p
        v-if="panelAlert"
        class="j8-me-alert"
        :class="alertToneClass(panelAlert.level)"
        :role="panelAlert.level === 'error' ? 'alert' : 'status'"
      >
        {{ panelAlert.message }}
      </p>

      <!-- Loading -->
      <div v-if="loading" class="j8-me-rows" aria-hidden="true">
        <div v-for="idx in 3" :key="idx" class="j8-me-sk"></div>
      </div>

      <!-- 无凭证 -->
      <div v-else-if="!credential" class="j8-me-cta">
        <p class="j8-me-hint">{{ t('personalCenter.apiPanel.noCredential') }}</p>
        <J8Button :loading="submitting" @click="handleApply">
          {{ submitting ? t('personalCenter.apiPanel.applying') : t('personalCenter.apiPanel.apply') }}
        </J8Button>
      </div>

      <!-- 审核中 -->
      <div v-else-if="credential.status === 'pending_review'" class="j8-me-rows">
        <div class="j8-me-row">
          <div class="j8-me-row-main">
            <b>{{ t('personalCenter.apiPanel.pendingTitle') }}</b>
            <small>{{ t('personalCenter.apiPanel.pendingDesc') }}</small>
          </div>
          <div class="j8-me-row-side">
            <span class="j8-me-pill is-warn" aria-hidden="true"><AlertTriangle /></span>
          </div>
        </div>
      </div>

      <!-- 已拒绝 -->
      <div v-else-if="credential.status === 'rejected'" class="j8-me-rows">
        <div class="j8-me-row">
          <div class="j8-me-row-main">
            <b>{{ t('personalCenter.apiPanel.rejectedTitle') }}</b>
            <small v-if="credential.reject_reason">
              {{ t('personalCenter.apiPanel.rejectReason', { reason: credential.reject_reason }) }}
            </small>
          </div>
          <div class="j8-me-row-side">
            <span class="j8-me-pill is-danger" aria-hidden="true"><XCircle /></span>
            <J8Button size="sm" :loading="submitting" @click="handleApply">
              {{ submitting ? t('personalCenter.apiPanel.applying') : t('personalCenter.apiPanel.reapply') }}
            </J8Button>
          </div>
        </div>
      </div>

      <!-- 已通过 / 使用中 -->
      <div v-else class="j8-me-rows">
        <!-- 首次生成提示（Secret 从未查看过） -->
        <div v-if="!hasViewedSecret && !newSecret" class="j8-me-row">
          <div class="j8-me-row-main">
            <b>{{ t('personalCenter.apiPanel.approvedNoticeTitle') }}</b>
            <small>{{ t('personalCenter.apiPanel.approvedNoticeDesc') }}</small>
          </div>
          <div class="j8-me-row-side">
            <J8Button size="sm" :loading="submitting" @click="handleFirstGenerate">
              {{ submitting ? t('personalCenter.apiPanel.regenerating') : t('personalCenter.apiPanel.generateSecret') }}
            </J8Button>
          </div>
        </div>

        <!-- API Key -->
        <div class="j8-me-row">
          <div class="j8-me-row-main">
            <small>API Key</small>
            <b><code>{{ credential.api_key || '-' }}</code></b>
          </div>
          <div class="j8-me-row-side">
            <button type="button" class="j8-me-act" @click="copyToClipboard(credential.api_key || '')">
              {{ t('personalCenter.apiPanel.copy') }}
            </button>
          </div>
        </div>

        <!-- API Secret -->
        <div class="j8-me-row">
          <div class="j8-me-row-main">
            <small>API Secret</small>
            <b><code>{{ maskedSecret }}</code></b>
            <small>{{ t('personalCenter.apiPanel.secretHint') }}</small>
          </div>
          <div class="j8-me-row-side">
            <button type="button" class="j8-me-act" :disabled="submitting" @click="handleRegenerate">
              {{ t('personalCenter.apiPanel.regenerate') }}
            </button>
          </div>
        </div>

        <!-- 新生成的 Secret（仅显示一次） -->
        <div v-if="newSecret" class="j8-me-row j8-me-fresh">
          <div class="j8-me-row-main">
            <b><code>{{ newSecret }}</code></b>
            <small>{{ t('personalCenter.apiPanel.newSecretWarning') }}</small>
          </div>
          <div class="j8-me-row-side">
            <span class="j8-me-pill is-ok">
              <Check aria-hidden="true" />
              {{ t('personalCenter.apiPanel.newSecretTitle') }}
            </span>
            <button type="button" class="j8-me-act is-primary" @click="copyToClipboard(newSecret)">
              {{ t('personalCenter.apiPanel.copySecret') }}
            </button>
          </div>
        </div>

        <!-- 启用 / 禁用 -->
        <div class="j8-me-row">
          <div class="j8-me-row-main">
            <b>{{ t('personalCenter.apiPanel.statusLabel') }}</b>
            <small>{{ credential.is_active ? t('personalCenter.apiPanel.statusEnabled') : t('personalCenter.apiPanel.statusDisabled') }}</small>
          </div>
          <div class="j8-me-row-side">
            <button
              type="button"
              class="j8-me-switch"
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
          class="j8-me-modal-mask"
          role="dialog"
          aria-modal="true"
          :aria-label="t('personalCenter.apiPanel.regenerateTitle')"
          @click.self="showConfirm = false"
        >
          <div ref="confirmEl" class="j8-me-modal" tabindex="-1" @keydown.esc="showConfirm = false">
            <h3>{{ t('personalCenter.apiPanel.regenerateTitle') }}</h3>
            <p>{{ t('personalCenter.apiPanel.regenerateDesc') }}</p>
            <div class="j8-me-modal-acts">
              <button type="button" class="j8-me-act" @click="showConfirm = false">
                {{ t('personalCenter.apiPanel.cancel') }}
              </button>
              <J8Button variant="danger" :loading="submitting" @click="confirmRegenerate">
                {{ submitting ? t('personalCenter.apiPanel.regenerating') : t('personalCenter.apiPanel.regenerateConfirm') }}
              </J8Button>
            </div>
          </div>
        </div>
      </Teleport>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlertTriangle, Check, Key, XCircle } from 'lucide-vue-next'
import { apiCredentialAPI } from '../../../api'
import type { PageAlert } from '../../../utils/alerts'
import J8Button from '../components/J8Button.vue'

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

const alertToneClass = (level: PageAlert['level']) => {
  if (level === 'success') return 'is-ok'
  if (level === 'warning') return 'is-warn'
  return ''
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
.j8-me-cta {
  display: grid;
  gap: 12px;
  justify-items: start;
}

.j8-me-card .j8-me-alert {
  margin-bottom: 12px;
}

.j8-me-alert.is-ok {
  background: var(--accent-soft);
  color: var(--accent-strong);
}

.j8-me-alert.is-warn {
  background: var(--amber-soft);
  color: var(--amber);
}

.j8-me-row-main code {
  font-family: var(--font-mono, ui-monospace, SFMono-Regular, Menlo, Consolas, monospace);
  font-size: 12.5px;
}

/* 新 Secret 一次性展示的入场 */
.j8-me-fresh {
  animation: j8-me-fresh-in 0.45s ease both;
}

@keyframes j8-me-fresh-in {
  from {
    opacity: 0;
    transform: translateY(8px) scale(0.99);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

/* 启用 / 禁用开关 */
.j8-me-switch {
  position: relative;
  display: inline-flex;
  align-items: center;
  width: 46px;
  height: 26px;
  padding: 2px;
  border: 0;
  border-radius: 999px;
  background: var(--line-strong);
  transition: background 0.2s ease;
}

.j8-me-switch > span {
  display: inline-block;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--surface);
  box-shadow: 0 1px 3px rgba(10, 10, 10, 0.25);
  transform: translateX(0);
  transition: transform 0.2s ease;
}

.j8-me-switch.is-on {
  background: var(--accent);
}

.j8-me-switch.is-on > span {
  transform: translateX(20px);
}

.j8-me-switch:disabled {
  opacity: 0.6;
}

/* 重新生成确认弹窗 */
.j8-me-modal-mask {
  position: fixed;
  inset: 0;
  z-index: 1200;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(10, 10, 10, 0.5);
  backdrop-filter: blur(4px);
  animation: j8-me-mask-in 0.2s ease;
}

@keyframes j8-me-mask-in {
  from {
    opacity: 0;
  }
}

.j8-me-modal {
  width: min(380px, calc(100vw - 32px));
  padding: 20px;
  border-radius: 20px;
  background: var(--surface, #fff);
  color: var(--text, #0a0a0a);
  box-shadow: 0 30px 60px -36px rgba(40, 30, 80, 0.3), 0 0 0 1px var(--line, rgba(10, 10, 10, 0.08));
  outline: none;
  animation: j8-me-modal-rise 0.3s cubic-bezier(0.34, 1.45, 0.64, 1);
}

@keyframes j8-me-modal-rise {
  from {
    transform: translateY(16px) scale(0.98);
    opacity: 0;
  }
}

.j8-me-modal h3 {
  font-size: 15.5px;
  font-weight: 900;
  letter-spacing: -0.01em;
}

.j8-me-modal p {
  margin-top: 8px;
  font-size: 12.5px;
  line-height: 1.65;
  color: var(--muted, #71717a);
}

.j8-me-modal-acts {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 18px;
}

@media (prefers-reduced-motion: reduce) {
  .j8-me-fresh,
  .j8-me-modal-mask,
  .j8-me-modal {
    animation-duration: 0.001ms;
  }
}
</style>
