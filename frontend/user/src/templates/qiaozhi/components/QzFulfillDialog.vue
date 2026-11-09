<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="qz-fd-mask"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      @click.self="requestClose"
    >
      <div ref="panelRef" class="qz-fd" tabindex="-1" @keydown.esc.stop="requestClose">
        <header class="qz-fd-head">
          <div class="qz-fd-head-text">
            <ol class="qz-fd-stages" :aria-label="t('qiaozhi.fulfill.stagesLabel')">
              <li :class="{ 'is-current': stage === 'form', 'is-done': stage === 'progress' }">
                <span class="qz-fd-stage-dot" aria-hidden="true">1</span>
                {{ t('qiaozhi.fulfill.stageForm') }}
              </li>
              <li :class="{ 'is-current': stage === 'progress' }">
                <span class="qz-fd-stage-dot" aria-hidden="true">2</span>
                {{ t('qiaozhi.fulfill.stageProgress') }}
              </li>
            </ol>
            <h2 :id="titleId" class="qz-fd-title">
              {{ stage === 'form' ? t('qiaozhi.fulfill.formTitle') : t('qiaozhi.fulfill.progressTitle') }}
            </h2>
          </div>
          <button type="button" class="qz-icon-btn qz-icon-btn--sm" :aria-label="t('common.cancel')" @click="requestClose">
            <X />
          </button>
        </header>

        <!-- 载入订单 -->
        <div v-if="stage === 'loading'" class="qz-fd-state" role="status">
          <Loader2 class="qz-fd-spin" aria-hidden="true" />
          <span>{{ t('qiaozhi.fulfill.loading') }}</span>
        </div>

        <div v-else-if="stage === 'error'" class="qz-fd-state qz-fd-state--error" role="alert">
          <CircleX aria-hidden="true" />
          <span>{{ loadError || t('qiaozhi.fulfill.loadFailed') }}</span>
          <QzButton size="sm" variant="secondary" @click="initialize">{{ t('errorBoundary.retry') }}</QzButton>
        </div>

        <!-- 阶段 1：填写开通信息 -->
        <form v-else-if="stage === 'form'" class="qz-fd-form" novalidate @submit.prevent="handleSubmit">
          <p class="qz-fd-lead">{{ t('qiaozhi.fulfill.formLead') }}</p>

          <div v-for="field in fields" :key="field.key" class="qz-field">
            <label class="qz-field-label" :id="`${fieldId(field.key)}-label`" :for="fieldId(field.key)">
              {{ fieldLabel(field) }}
              <b v-if="field.required">*</b>
            </label>

            <textarea
              v-if="field.type === 'textarea'"
              :id="fieldId(field.key)"
              class="qz-field-control qz-fd-textarea"
              :value="values[field.key]"
              rows="5"
              autocomplete="off"
              spellcheck="false"
              :placeholder="fieldPlaceholder(field)"
              :aria-invalid="hasError(field.key)"
              :aria-describedby="hasError(field.key) ? `${fieldId(field.key)}-err` : undefined"
              @input="setValue(field.key, ($event.target as HTMLTextAreaElement).value)"
            ></textarea>

            <QzSelect
              v-else-if="field.type === 'select'"
              :id="fieldId(field.key)"
              :model-value="values[field.key]"
              :items="[{ value: '', label: t('checkout.manualFormSelectPlaceholder') }, ...field.options.map((option) => ({ value: option, label: option }))]"
              :aria-invalid="hasError(field.key)"
              @update:model-value="setValue(field.key, String($event))"
            />

            <div
              v-else-if="field.type === 'radio'"
              class="qz-fd-options"
              role="radiogroup"
              :aria-labelledby="`${fieldId(field.key)}-label`"
            >
              <label v-for="option in field.options" :key="option">
                <input
                  type="radio"
                  :name="`qz-fd-radio-${field.key}`"
                  :value="option"
                  :checked="values[field.key] === option"
                  @change="setValue(field.key, option)"
                />
                <span>{{ option }}</span>
              </label>
            </div>

            <div
              v-else-if="field.type === 'checkbox'"
              class="qz-fd-options"
              role="group"
              :aria-labelledby="`${fieldId(field.key)}-label`"
            >
              <label v-for="option in field.options" :key="option">
                <input
                  type="checkbox"
                  :value="option"
                  :checked="checkboxChecked(field.key, option)"
                  @change="toggleCheckbox(field.key, option, ($event.target as HTMLInputElement).checked)"
                />
                <span>{{ option }}</span>
              </label>
            </div>

            <input
              v-else
              :id="fieldId(field.key)"
              class="qz-field-control"
              autocomplete="off"
              :type="field.type === 'number' ? 'number' : field.type === 'email' ? 'email' : field.type === 'phone' ? 'tel' : 'text'"
              :value="values[field.key]"
              :placeholder="fieldPlaceholder(field)"
              :aria-invalid="hasError(field.key)"
              :aria-describedby="hasError(field.key) ? `${fieldId(field.key)}-err` : undefined"
              @input="setValue(field.key, ($event.target as HTMLInputElement).value)"
            />

            <p v-if="submitAttempted && errors[field.key]" :id="`${fieldId(field.key)}-err`" class="qz-field-error">
              {{ errors[field.key] }}
            </p>
          </div>

          <!-- 上游预检拒绝：内联展示原因与账号摘要，用户可换一个会话重试 -->
          <div v-if="rejection" class="qz-fd-reject" role="alert">
            <TriangleAlert aria-hidden="true" />
            <div>
              <b>{{ rejection.message || t('qiaozhi.fulfill.rejectedFallback') }}</b>
              <dl v-if="rejection.accountEmail || rejection.planType" class="qz-fd-account">
                <div v-if="rejection.accountEmail">
                  <dt>{{ t('qiaozhi.fulfill.accountLabel') }}</dt>
                  <dd>{{ rejection.accountEmail }}</dd>
                </div>
                <div v-if="rejection.planType">
                  <dt>{{ t('qiaozhi.fulfill.planLabel') }}</dt>
                  <dd>{{ rejection.planType }}</dd>
                </div>
              </dl>
              <p>{{ t('qiaozhi.fulfill.rejectedHint') }}</p>
            </div>
          </div>
          <p v-else-if="submitError" class="qz-field-error" role="alert">{{ submitError }}</p>

          <div class="qz-fd-actions">
            <QzButton type="button" variant="ghost" @click="requestClose">{{ t('qiaozhi.fulfill.later') }}</QzButton>
            <QzButton type="submit" :loading="submitting">{{ t('qiaozhi.fulfill.submit') }}</QzButton>
          </div>
          <p class="qz-field-hint">{{ t('qiaozhi.fulfill.laterHint') }}</p>
        </form>

        <!-- 阶段 2：开通进度 -->
        <div v-else class="qz-fd-progress">
          <div v-if="progress?.phase === 'completed'" class="qz-fd-result qz-fd-result--ok" role="status">
            <CheckCircle2 aria-hidden="true" />
            <div>
              <b>{{ t('qiaozhi.fulfill.successTitle') }}</b>
              <p>{{ t('qiaozhi.fulfill.successDesc') }}</p>
            </div>
          </div>
          <div v-else-if="progress?.phase === 'failed'" class="qz-fd-result qz-fd-result--fail" role="alert">
            <CircleX aria-hidden="true" />
            <div>
              <b>{{ t('qiaozhi.fulfill.failTitle') }}</b>
              <p>{{ progress.message || t('qiaozhi.fulfill.failDesc') }}</p>
              <p>{{ t('qiaozhi.fulfill.failHint', { orderNo }) }}</p>
            </div>
          </div>
          <template v-else>
            <p class="qz-fd-lead">{{ progress?.message || t('qiaozhi.fulfill.progressLead') }}</p>
          </template>

          <div
            class="qz-fd-bar"
            role="progressbar"
            :aria-valuenow="percent"
            aria-valuemin="0"
            aria-valuemax="100"
            :aria-label="t('qiaozhi.fulfill.progressTitle')"
          >
            <span class="qz-fd-bar-fill" :class="{ 'is-failed': progress?.phase === 'failed' }" :style="{ width: `${percent}%` }"></span>
          </div>
          <p class="qz-fd-percent">{{ percent }}%</p>

          <ol class="qz-fd-steps">
            <li v-for="step in stepViews" :key="step.key" :class="`is-${step.state}`">
              <span class="qz-fd-step-icon" aria-hidden="true">
                <Check v-if="step.state === 'done'" />
                <Loader2 v-else-if="step.state === 'current'" class="qz-fd-spin" />
                <X v-else-if="step.state === 'failed'" />
                <Circle v-else />
              </span>
              <span class="qz-fd-step-label">{{ step.label }}</span>
              <span v-if="step.state === 'current'" class="qz-sr-only">{{ t('qiaozhi.fulfill.stepCurrent') }}</span>
            </li>
          </ol>

          <p v-if="pollFailed && !isFinished" class="qz-field-hint" role="status">{{ t('qiaozhi.fulfill.pollRetrying') }}</p>
          <p v-if="!isFinished" class="qz-field-hint">{{ t('qiaozhi.fulfill.closeHint') }}</p>

          <div class="qz-fd-actions">
            <QzButton v-if="progress?.phase === 'failed' && supportHref" variant="secondary" @click="openSupport">
              {{ t('qiaozhi.fulfill.contactSupport') }}
            </QzButton>
            <QzButton :variant="isFinished ? 'primary' : 'secondary'" @click="emit('view-order')">
              {{ t('qiaozhi.fulfill.viewOrder') }}
            </QzButton>
            <QzButton v-if="!isFinished" variant="ghost" @click="requestClose">{{ t('qiaozhi.fulfill.close') }}</QzButton>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Check, CheckCircle2, Circle, CircleX, Loader2, TriangleAlert, X } from 'lucide-vue-next'
import { guestOrderAPI, userOrderAPI } from '../../../api'
import { useAppStore } from '../../../stores/app'
import {
  buildManualFormRow,
  buildManualFormValues,
  normalizeManualFormSchema,
  validateManualFormValues,
  type ManualFormField,
} from '../../../utils/manualForm'
import {
  buildStepViews,
  findFulfillPendingItem,
  isConflictError,
  isTerminalProgress,
  nextPollDelay,
  normalizeFulfillProgress,
  parsePrecheckRejection,
  resolvePercent,
  resolveSupportHref,
  type FulfillProgress,
  type PrecheckRejection,
} from '../utils/fulfillDialog'
import QzButton from './QzButton.vue'
import QzSelect from './QzSelect.vue'

type Stage = 'loading' | 'error' | 'form' | 'progress'

const props = defineProps<{
  open: boolean
  orderNo: string
  /** 游客订单的邮箱 + 订单密码（调用方从下单结果 / 本地游客凭证取）；登录用户传 null。 */
  guestCredentials?: { email: string; order_password: string } | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'view-order'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()

const titleId = `qz-fd-title-${Math.random().toString(36).slice(2, 8)}`
const panelRef = ref<HTMLElement | null>(null)

const stage = ref<Stage>('loading')
const loadError = ref('')
const fields = ref<ManualFormField[]>([])
// 表单值（含用户粘贴的会话）只存在于组件内存：不写 localStorage / sessionStorage / URL / 日志。
const values = ref<Record<string, any>>({})
const errors = ref<Record<string, string>>({})
const submitAttempted = ref(false)
const submitting = ref(false)
const submitError = ref('')
const rejection = ref<PrecheckRejection | null>(null)

const progress = ref<FulfillProgress | null>(null)
const pollFailed = ref(false)
let pollTimer: ReturnType<typeof setTimeout> | undefined
let pollRun = 0

const isGuest = computed(() => Boolean(props.guestCredentials?.email))
const guestAuth = () => ({
  email: props.guestCredentials?.email || '',
  order_password: props.guestCredentials?.order_password || '',
})

const percent = computed(() => resolvePercent(progress.value))
const isFinished = computed(() => isTerminalProgress(progress.value))
const stepViews = computed(() => buildStepViews(progress.value, {
  paid: t('qiaozhi.fulfill.stepPaid'),
  submit: t('qiaozhi.fulfill.stepSubmit'),
  process: t('qiaozhi.fulfill.stepProcess'),
  complete: t('qiaozhi.fulfill.stepComplete'),
}))
const supportHref = computed(() => resolveSupportHref(appStore.config?.contact as any))

// ---------- 字段渲染辅助 ----------
const localized = (map: Record<string, string> | undefined, fallback = '') => {
  if (!map) return fallback
  const locale = appStore.locale
  return map[locale] || map['zh-CN'] || map['en-US'] || fallback
}
const fieldId = (key: string) => `qz-fd-f-${key}`
const fieldLabel = (field: ManualFormField) => localized(field.label, field.key)
const fieldPlaceholder = (field: ManualFormField) => localized(field.placeholder, '')
const hasError = (key: string) => submitAttempted.value && Boolean(errors.value[key])

const revalidate = () => {
  errors.value = validateManualFormValues(fields.value, values.value, { getLabel: fieldLabel, t })
}
const setValue = (key: string, value: string) => {
  values.value = { ...values.value, [key]: value }
  rejection.value = null
  submitError.value = ''
  revalidate()
}
const checkboxChecked = (key: string, option: string) =>
  Array.isArray(values.value[key]) && values.value[key].includes(option)
const toggleCheckbox = (key: string, option: string, checked: boolean) => {
  const current: string[] = Array.isArray(values.value[key]) ? [...values.value[key]] : []
  const next = checked ? Array.from(new Set([...current, option])) : current.filter((item) => item !== option)
  values.value = { ...values.value, [key]: next }
  revalidate()
}

// ---------- 加载 ----------
const fetchOrder = async () => {
  const res = isGuest.value
    ? await guestOrderAPI.detail(props.orderNo, guestAuth())
    : await userOrderAPI.detail(props.orderNo)
  return res.data?.data
}

const callProgress = async () => {
  const res = isGuest.value
    ? await guestOrderAPI.fulfillProgress(props.orderNo, guestAuth())
    : await userOrderAPI.fulfillProgress(props.orderNo)
  return normalizeFulfillProgress(res.data?.data)
}

const resetState = () => {
  stopPolling()
  stage.value = 'loading'
  loadError.value = ''
  fields.value = []
  values.value = {}
  errors.value = {}
  submitAttempted.value = false
  submitting.value = false
  submitError.value = ''
  rejection.value = null
  progress.value = null
  pollFailed.value = false
}

const enterForm = (item: any) => {
  fields.value = normalizeManualFormSchema(item?.manual_form_schema_snapshot)
  values.value = buildManualFormValues(fields.value)
  revalidate()
  stage.value = 'form'
  nextTick(() => panelRef.value?.querySelector<HTMLElement>('textarea, input')?.focus())
}

const initialize = async () => {
  if (!props.orderNo) return
  resetState()
  try {
    const order = await fetchOrder()
    const pending = findFulfillPendingItem(order)
    if (pending && normalizeManualFormSchema(pending.manual_form_schema_snapshot).length > 0) {
      enterForm(pending)
      return
    }
    await enterProgress()
  } catch (err: any) {
    loadError.value = err?.message || ''
    stage.value = 'error'
  }
}

// ---------- 提交 ----------
const handleSubmit = async () => {
  if (submitting.value) return
  submitAttempted.value = true
  submitError.value = ''
  rejection.value = null
  revalidate()
  if (Object.keys(errors.value).length > 0) return
  submitting.value = true
  try {
    const payload = buildManualFormRow(fields.value, values.value)
    if (isGuest.value) {
      await guestOrderAPI.submitFulfillInfo(props.orderNo, guestAuth(), payload)
    } else {
      await userOrderAPI.submitFulfillInfo(props.orderNo, payload)
    }
    // 提交成功即清掉内存里的会话。
    values.value = {}
    await enterProgress()
  } catch (err: any) {
    const rejected = parsePrecheckRejection(err)
    if (rejected) {
      rejection.value = rejected
      // 留在表单阶段，保留输入，方便用户直接改会话再提交。
    } else if (isConflictError(err)) {
      // 已提交过（例如另一端已填）：直接看进度。
      await enterProgress()
    } else {
      submitError.value = err?.message || t('qiaozhi.fulfill.submitFailed')
    }
  } finally {
    submitting.value = false
  }
}

// ---------- 进度轮询 ----------
const stopPolling = () => {
  pollRun += 1
  if (pollTimer !== undefined) {
    clearTimeout(pollTimer)
    pollTimer = undefined
  }
}

const pollOnce = async (run: number) => {
  let delay: number | null = 3000
  try {
    const next = await callProgress()
    if (run !== pollRun) return
    pollFailed.value = false
    if (next.phase === 'awaiting_info' && stage.value === 'progress') {
      // 还没有提交过开通信息：回到表单。
      const order = await fetchOrder()
      if (run !== pollRun) return
      enterForm(findFulfillPendingItem(order))
      return
    }
    progress.value = next
    delay = nextPollDelay(next)
  } catch {
    if (run !== pollRun) return
    pollFailed.value = true
  }
  if (run !== pollRun || delay === null) return
  pollTimer = setTimeout(() => { void pollOnce(run) }, delay)
}

async function enterProgress() {
  stopPolling()
  stage.value = 'progress'
  pollFailed.value = false
  progress.value = null
  const run = pollRun
  await pollOnce(run)
}

// ---------- 关闭 / 客服 ----------
const requestClose = () => {
  // 关闭只是收起弹窗，不取消任何开通流程。
  stopPolling()
  values.value = {}
  emit('update:open', false)
}

const openSupport = () => {
  if (supportHref.value) window.open(supportHref.value, '_blank', 'noopener,noreferrer')
}

watch(() => [props.open, props.orderNo] as const, ([opened]) => {
  if (opened) {
    void initialize()
  } else {
    resetState()
  }
}, { immediate: true })

watch(() => appStore.locale, () => {
  if (stage.value === 'form') revalidate()
})

onBeforeUnmount(() => {
  stopPolling()
})
</script>

<style scoped>
.qz-fd-mask {
  position: fixed; inset: 0;
  z-index: 1200;
  display: flex; align-items: center; justify-content: center;
  padding: var(--qz-space-4);
  background: color-mix(in oklab, var(--qz-text) 50%, transparent);
  backdrop-filter: blur(4px);
  animation: qz-fd-fade var(--qz-duration-standard) var(--qz-ease-standard);
}
.dark .qz-fd-mask { background: color-mix(in oklab, var(--qz-bg) 72%, transparent); }
@keyframes qz-fd-fade { from { opacity: 0; } }

.qz-fd {
  width: min(460px, 100%);
  max-height: calc(100dvh - 32px);
  overflow-y: auto;
  display: flex; flex-direction: column; gap: var(--qz-space-4);
  padding: var(--qz-space-5);
  color: var(--qz-text);
  background: var(--qz-surface);
  border-radius: var(--qz-radius-panel);
  box-shadow: var(--qz-shadow-floating), 0 0 0 1px var(--qz-border);
  font-family: var(--qz-font-body);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
  outline: none;
  animation: qz-fd-rise var(--qz-duration-considered) var(--qz-ease-enter);
}
@keyframes qz-fd-rise { from { transform: translateY(20px) scale(.98); opacity: 0; } }

.qz-fd-head { display: flex; align-items: flex-start; justify-content: space-between; gap: var(--qz-space-3); }
.qz-fd-head-text { display: grid; gap: var(--qz-space-2); min-width: 0; }
.qz-fd-title { margin: 0; font-size: var(--qz-text-lg); font-weight: 650; }

.qz-fd-stages { display: flex; gap: var(--qz-space-3); margin: 0; padding: 0; list-style: none; color: var(--qz-text-muted); font-size: var(--qz-text-xs); }
.qz-fd-stages li { display: inline-flex; align-items: center; gap: var(--qz-space-2); }
.qz-fd-stages li.is-current { color: var(--qz-text); font-weight: 600; }
.qz-fd-stages li.is-done { color: var(--qz-success); }
.qz-fd-stage-dot {
  display: inline-grid; place-items: center;
  width: 18px; height: 18px;
  border-radius: var(--qz-radius-pill);
  border: 1px solid currentColor;
  font-size: var(--qz-text-2xs);
}
.qz-fd-stages li.is-current .qz-fd-stage-dot { color: var(--qz-on-accent); background: var(--qz-accent); border-color: var(--qz-accent); }

.qz-fd-lead { margin: 0; color: var(--qz-text-secondary); }

.qz-fd-state { display: flex; align-items: center; justify-content: center; gap: var(--qz-space-3); padding: var(--qz-space-8) 0; color: var(--qz-text-secondary); }
.qz-fd-state svg { width: 20px; height: 20px; }
.qz-fd-state--error { flex-direction: column; color: var(--qz-danger); text-align: center; }

.qz-fd-form, .qz-fd-progress { display: grid; gap: var(--qz-space-4); }
.qz-fd-textarea { font-family: var(--qz-font-mono); font-size: var(--qz-text-xs); min-height: 120px; }

.qz-fd-options { display: flex; flex-wrap: wrap; gap: var(--qz-space-3); }
.qz-fd-options label { display: inline-flex; align-items: center; gap: var(--qz-space-2); cursor: pointer; }

.qz-fd-reject {
  display: flex; gap: var(--qz-space-3);
  padding: var(--qz-space-3) var(--qz-space-4);
  color: var(--qz-text);
  background: color-mix(in oklab, var(--qz-danger) 10%, var(--qz-surface));
  border: 1px solid color-mix(in oklab, var(--qz-danger) 40%, var(--qz-border));
  border-radius: var(--qz-radius-control);
}
.qz-fd-reject > svg { flex: none; width: 18px; height: 18px; margin-top: 2px; color: var(--qz-danger); }
.qz-fd-reject p { margin: var(--qz-space-2) 0 0; color: var(--qz-text-secondary); font-size: var(--qz-text-xs); }
.qz-fd-account { display: grid; gap: var(--qz-space-1); margin: var(--qz-space-2) 0 0; font-size: var(--qz-text-xs); }
.qz-fd-account > div { display: flex; gap: var(--qz-space-2); }
.qz-fd-account dt { color: var(--qz-text-muted); }
.qz-fd-account dd { margin: 0; font-weight: 600; }

.qz-fd-actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: var(--qz-space-3); }

.qz-fd-result { display: flex; gap: var(--qz-space-3); padding: var(--qz-space-4); border-radius: var(--qz-radius-control); }
.qz-fd-result > svg { flex: none; width: 24px; height: 24px; }
.qz-fd-result b { display: block; font-size: var(--qz-text-base); }
.qz-fd-result p { margin: var(--qz-space-1) 0 0; color: var(--qz-text-secondary); }
.qz-fd-result--ok { background: color-mix(in oklab, var(--qz-success) 10%, var(--qz-surface)); }
.qz-fd-result--ok > svg { color: var(--qz-success); }
.qz-fd-result--fail { background: color-mix(in oklab, var(--qz-danger) 10%, var(--qz-surface)); }
.qz-fd-result--fail > svg { color: var(--qz-danger); }

.qz-fd-bar { height: 8px; overflow: hidden; background: var(--qz-surface-muted); border-radius: var(--qz-radius-pill); }
.qz-fd-bar-fill {
  display: block; height: 100%;
  background: var(--qz-accent);
  border-radius: var(--qz-radius-pill);
  transition: width var(--qz-duration-considered) var(--qz-ease-standard);
}
.qz-fd-bar-fill.is-failed { background: var(--qz-danger); }
.qz-fd-percent { margin: calc(var(--qz-space-2) * -1) 0 0; color: var(--qz-text-muted); font-size: var(--qz-text-xs); text-align: right; font-variant-numeric: tabular-nums; }

.qz-fd-steps { display: grid; gap: var(--qz-space-3); margin: 0; padding: 0; list-style: none; }
.qz-fd-steps li { display: flex; align-items: center; gap: var(--qz-space-3); color: var(--qz-text-muted); }
.qz-fd-step-icon { display: inline-grid; place-items: center; flex: none; width: 22px; height: 22px; border-radius: var(--qz-radius-pill); border: 1px solid var(--qz-border-strong); }
.qz-fd-step-icon svg { width: 14px; height: 14px; }
.qz-fd-step-label { min-width: 0; overflow-wrap: anywhere; }
.qz-fd-steps li.is-done { color: var(--qz-text); }
.qz-fd-steps li.is-done .qz-fd-step-icon { color: var(--qz-on-accent); background: var(--qz-success); border-color: var(--qz-success); }
.qz-fd-steps li.is-current { color: var(--qz-text); font-weight: 600; }
.qz-fd-steps li.is-current .qz-fd-step-icon { color: var(--qz-accent); border-color: var(--qz-accent); }
.qz-fd-steps li.is-failed { color: var(--qz-danger); font-weight: 600; }
.qz-fd-steps li.is-failed .qz-fd-step-icon { color: var(--qz-on-accent); background: var(--qz-danger); border-color: var(--qz-danger); }

.qz-fd-spin { animation: qz-fd-spin 1s linear infinite; }
@keyframes qz-fd-spin { to { transform: rotate(360deg); } }

@media (prefers-reduced-motion: reduce) {
  .qz-fd-mask, .qz-fd { animation: none; }
  .qz-fd-bar-fill { transition: none; }
  .qz-fd-spin { animation: none; }
}
</style>
