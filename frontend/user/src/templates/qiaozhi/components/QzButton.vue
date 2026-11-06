<template>
  <button
    ref="buttonEl"
    class="qz-arc-btn"
    :class="[`is-${variant}`, `is-${size}`, { 'is-loading': loading, 'is-inert': inert }]"
    :disabled="disabled || undefined"
    :aria-busy="loading || undefined"
    :aria-disabled="loading || ariaDisabled || undefined"
    v-bind="$attrs"
    @pointerdown="handlePointerDown"
    @click="handleClick"
  >
    <Transition name="qz-arc-loader">
      <span v-if="loading" class="qz-arc-loader" aria-hidden="true"><span class="qz-arc-spinner" /></span>
    </Transition>
    <span ref="slotEl" class="qz-arc-slot" :style="slotStyle">
      <span ref="contentEl" class="qz-arc-content">
        <Transition :name="labelTransitionName()" mode="out-in">
          <span :key="currentKey()" class="qz-arc-label"><slot /></span>
        </Transition>
      </span>
    </span>
  </button>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, useAttrs, useSlots, watch } from 'vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
  disabled?: boolean
}>(), {
  variant: 'primary',
  size: 'md',
  loading: false,
  disabled: false,
})

const emit = defineEmits<{ click: [event: MouseEvent] }>()
const attrs = useAttrs()
const slots = useSlots()

const ariaDisabled = computed(() => attrs['aria-disabled'] === true || attrs['aria-disabled'] === 'true')
const inert = computed(() => props.disabled || props.loading || ariaDisabled.value)

/* ---------- 减小动态偏好 ---------- */
const reduced = ref(false)
let mq: MediaQueryList | null = null
const syncReduced = () => { reduced.value = mq?.matches ?? false }

/* ---------- label 指纹：文本 + 组件名，文案/图标变化才触发交叉淡入 ---------- */
function vnodeKey(node: any): string {
  if (node == null || typeof node === 'boolean') return ''
  if (typeof node === 'string' || typeof node === 'number') return String(node)
  if (Array.isArray(node)) return node.map(vnodeKey).join('')
  if (typeof node !== 'object') return ''
  const type = node.type
  let name = ''
  if (typeof type === 'string') name = type
  else if (type) name = (type as any).__name || (type as any).name || (type as any).displayName || ''
  const inner = vnodeKey(node.children)
  return name ? `<${name}>${inner}` : inner
}

let lastKey = ''
let armUntil = 0
/** 渲染期调用：label 指纹变化时打开 700ms 的宽度形变窗口。 */
function currentKey(): string {
  const key = vnodeKey(slots.default?.() ?? [])
  if (key !== lastKey) {
    lastKey = key
    armUntil = performance.now() + 700
  }
  return key
}

const labelIsIcon = () => !/\S/.test(currentKey().replace(/<[^>]*>/g, ''))
const labelTransitionName = () => {
  if (reduced.value) return 'qz-arc-label-fade'
  return labelIsIcon() ? 'qz-arc-label-icon' : 'qz-arc-label-text'
}

/* ---------- 宽度形变：新 label 出现时槽位弹簧到新宽度，其他重排（字体后载）直接跳变 ---------- */
const slotEl = ref<HTMLElement | null>(null)
const contentEl = ref<HTMLElement | null>(null)
const slotStyle = ref<Record<string, string>>({})
let observer: ResizeObserver | null = null
let measured = false

onMounted(() => {
  mq = window.matchMedia('(prefers-reduced-motion: reduce)')
  syncReduced()
  mq.addEventListener?.('change', syncReduced)

  const node = contentEl.value
  const slot = slotEl.value
  if (!node || !slot || typeof ResizeObserver === 'undefined') return
  observer = new ResizeObserver((entries) => {
    const entry = entries[0]
    if (!entry) return
    const next = entry.contentRect.width
    if (!next) return
    const shouldMorph = measured && !reduced.value && performance.now() <= armUntil
    measured = true
    const from = slot.offsetWidth
    slotStyle.value = { width: `${next}px` }
    if (shouldMorph && from > 0 && Math.abs(from - next) > 0.5 && typeof slot.animate === 'function') {
      slot.animate(
        [{ width: `${from}px` }, { width: `${next}px` }],
        { duration: 450, easing: 'cubic-bezier(0.34, 1.45, 0.64, 1)' },
      )
    }
  })
  observer.observe(node)
})

onBeforeUnmount(() => {
  observer?.disconnect()
  mq?.removeEventListener?.('change', syncReduced)
})

/* ---------- 按压：宽按钮浅一点、图标按钮深一点，手感一致 ---------- */
const buttonEl = ref<HTMLButtonElement | null>(null)
const handlePointerDown = () => {
  const el = buttonEl.value
  if (!el || inert.value || reduced.value) return
  const width = el.offsetWidth
  const scale = width > 220 ? 0.985 : (width > 0 && width <= 48 ? 0.96 : 0.97)
  el.style.setProperty('--qz-press', String(scale))
}

/* ---------- loading 时吞掉点击但保留键盘焦点（aria-disabled 而非 disabled） ---------- */
const handleClick = (event: MouseEvent) => {
  if (props.loading) {
    event.preventDefault()
    return
  }
  emit('click', event)
}

watch(() => props.loading, (value) => {
  if (value) armUntil = performance.now() + 700
})
</script>

<style>
.qz-arc-btn {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid transparent;
  border-radius: var(--r-pill, 999px);
  font-weight: 800;
  line-height: 1;
  cursor: pointer;
  white-space: nowrap;
  user-select: none;
  -webkit-tap-highlight-color: transparent;
  transition: transform 0.42s cubic-bezier(0.34, 1.45, 0.64, 1), background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease, opacity 0.2s ease;
}
.qz-arc-btn:active:not(:disabled):not(.is-inert) {
  transform: scale(var(--qz-press, 0.97));
  transition: transform 0.05s ease, background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease, opacity 0.2s ease;
}
.qz-arc-btn[aria-busy="true"] { cursor: progress; }

.qz-arc-btn.is-sm { height: 34px; padding: 0px 13px; font-size: 12.5px; }
.qz-arc-btn.is-md { height: 44px; padding: 0px 22px; font-size: 14px; }
.qz-arc-btn.is-lg { height: 50px; padding: 0px 28px; font-size: 15px; }
.qz-arc-btn.is-sm .qz-arc-content { gap: 6px; }
.qz-arc-btn.is-md .qz-arc-content { gap: 8px; }
.qz-arc-btn.is-lg .qz-arc-content { gap: 8px; }
.qz-arc-btn.is-sm .qz-arc-content svg { width: 14px; height: 14px; }
.qz-arc-btn.is-md .qz-arc-content svg { width: 16px; height: 16px; }
.qz-arc-btn.is-lg .qz-arc-content svg { width: 18px; height: 18px; }

.qz-arc-btn.is-primary { background: var(--accent); border-color: var(--accent); color: var(--on-accent); }
.qz-arc-btn.is-secondary { background: var(--surface); border-color: var(--line-strong); color: var(--text); }
.qz-arc-btn.is-ghost { background: transparent; border-color: transparent; color: var(--text); }
.qz-arc-btn.is-danger { background: var(--surface); border-color: var(--line-strong); color: var(--red); }

@media (hover: hover) and (pointer: fine) {
  .qz-arc-btn.is-primary:hover:not(:disabled):not(.is-inert) { background: var(--accent-strong); border-color: var(--accent-strong); }
  .qz-arc-btn.is-secondary:hover:not(:disabled):not(.is-inert) { background: var(--surface-2); border-color: var(--accent-line); }
  .qz-arc-btn.is-ghost:hover:not(:disabled):not(.is-inert) { background: var(--accent-soft); }
  .qz-arc-btn.is-danger:hover:not(:disabled):not(.is-inert) { background: var(--red-soft); border-color: var(--red); }
}

/* 禁用：主按钮整体褪色；次/危险保留描边变灰底；ghost 变灰字——不可用仍读得出是按钮 */
.qz-arc-btn:disabled { cursor: not-allowed; }
.qz-arc-btn.is-primary:disabled { opacity: 0.4; }
.qz-arc-btn.is-secondary:disabled,
.qz-arc-btn.is-danger:disabled { background: var(--surface-2); color: var(--muted); }
.qz-arc-btn.is-ghost:disabled { color: var(--muted); }

.qz-arc-btn.qz-arc-block { width: 100%; }

/* label 槽：宽度由脚本测出并弹簧形变；内容永远按自然宽度渲染，不换行 */
.qz-arc-slot { display: inline-grid; place-items: center; }
.qz-arc-content { display: inline-flex; align-items: center; white-space: nowrap; }
.qz-arc-label { display: inline-flex; align-items: center; gap: inherit; }

/* 文案切换：新字从下方轻微上浮（带模糊）进场，旧字加速上移退场 */
.qz-arc-label-text-enter-active { transition: opacity 0.2s ease, transform 0.24s cubic-bezier(0.2, 0.8, 0.2, 1), filter 0.2s ease; }
.qz-arc-label-text-leave-active { transition: opacity 0.13s ease, transform 0.13s ease, filter 0.13s ease; }
.qz-arc-label-text-enter-from { opacity: 0; transform: translateY(4px); filter: blur(4px); }
.qz-arc-label-text-leave-to { opacity: 0; transform: translateY(-3px); filter: blur(4px); }

/* 图标切换：缩放弹入 */
.qz-arc-label-icon-enter-active { transition: opacity 0.2s ease, transform 0.34s cubic-bezier(0.34, 1.45, 0.64, 1), filter 0.2s ease; }
.qz-arc-label-icon-leave-active { transition: opacity 0.12s ease, transform 0.12s ease, filter 0.12s ease; }
.qz-arc-label-icon-enter-from,
.qz-arc-label-icon-leave-to { opacity: 0; transform: scale(0.6); filter: blur(2px); }

/* 低动态：纯淡入淡出 */
.qz-arc-label-fade-enter-active,
.qz-arc-label-fade-leave-active { transition: opacity 0.15s ease; }
.qz-arc-label-fade-enter-from,
.qz-arc-label-fade-leave-to { opacity: 0; }

/* loading 转圈 */
.qz-arc-loader { display: inline-grid; place-items: center; margin-right: 8px; }
.qz-arc-spinner {
  width: 15px;
  height: 15px;
  border-radius: 50%;
  border: 2px solid color-mix(in srgb, currentColor 30%, transparent);
  border-top-color: currentColor;
  animation: qz-arc-spin 0.7s linear infinite;
}
@keyframes qz-arc-spin { to { transform: rotate(360deg); } }
.qz-arc-loader-enter-active { transition: opacity 0.18s ease, transform 0.3s cubic-bezier(0.34, 1.45, 0.64, 1); }
.qz-arc-loader-leave-active { transition: opacity 0.12s ease, transform 0.12s ease; }
.qz-arc-loader-enter-from,
.qz-arc-loader-leave-to { opacity: 0; transform: scale(0.6); }

@media (prefers-reduced-motion: reduce) {
  .qz-arc-btn,
  .qz-arc-btn:active:not(:disabled):not(.is-inert) { transform: none; transition-duration: 0.001ms; }
  .qz-arc-label-text-enter-from,
  .qz-arc-label-text-leave-to,
  .qz-arc-label-icon-enter-from,
  .qz-arc-label-icon-leave-to { transform: none; filter: none; }
}
</style>
