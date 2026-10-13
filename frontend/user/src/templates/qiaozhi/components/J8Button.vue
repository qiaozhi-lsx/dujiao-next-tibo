<template>
  <button
    ref="buttonEl"
    class="j8-arc-btn"
    :class="[`is-${variant}`, `is-${size}`, { 'is-loading': loading, 'is-inert': inert }]"
    :disabled="disabled || undefined"
    :aria-busy="loading || undefined"
    :aria-disabled="loading || ariaDisabled || undefined"
    v-bind="$attrs"
    @pointerdown="handlePointerDown"
    @click="handleClick"
  >
    <Transition name="j8-arc-loader">
      <span v-if="loading" class="j8-arc-loader" aria-hidden="true"><span class="j8-arc-spinner" /></span>
    </Transition>
    <span ref="slotEl" class="j8-arc-slot" :style="slotStyle">
      <span ref="contentEl" class="j8-arc-content">
        <Transition :name="labelTransitionName()" mode="out-in">
          <span :key="currentKey()" class="j8-arc-label"><slot /></span>
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
  if (reduced.value) return 'j8-arc-label-fade'
  return labelIsIcon() ? 'j8-arc-label-icon' : 'j8-arc-label-text'
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
  el.style.setProperty('--j8-press', String(scale))
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
.j8-arc-btn {
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
.j8-arc-btn:active:not(:disabled):not(.is-inert) {
  transform: scale(var(--j8-press, 0.97));
  transition: transform 0.05s ease, background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease, opacity 0.2s ease;
}
.j8-arc-btn[aria-busy="true"] { cursor: progress; }

.j8-arc-btn.is-sm { height: 34px; padding: 0px 13px; font-size: 12.5px; }
.j8-arc-btn.is-md { height: 44px; padding: 0px 22px; font-size: 14px; }
.j8-arc-btn.is-lg { height: 50px; padding: 0px 28px; font-size: 15px; }
.j8-arc-btn.is-sm .j8-arc-content { gap: 6px; }
.j8-arc-btn.is-md .j8-arc-content { gap: 8px; }
.j8-arc-btn.is-lg .j8-arc-content { gap: 8px; }
.j8-arc-btn.is-sm .j8-arc-content svg { width: 14px; height: 14px; }
.j8-arc-btn.is-md .j8-arc-content svg { width: 16px; height: 16px; }
.j8-arc-btn.is-lg .j8-arc-content svg { width: 18px; height: 18px; }

.j8-arc-btn.is-primary { background: var(--accent); border-color: var(--accent); color: var(--on-accent); }
.j8-arc-btn.is-secondary { background: var(--surface); border-color: var(--line-strong); color: var(--text); }
.j8-arc-btn.is-ghost { background: transparent; border-color: transparent; color: var(--text); }
.j8-arc-btn.is-danger { background: var(--surface); border-color: var(--line-strong); color: var(--red); }

@media (hover: hover) and (pointer: fine) {
  .j8-arc-btn.is-primary:hover:not(:disabled):not(.is-inert) { background: var(--accent-strong); border-color: var(--accent-strong); }
  .j8-arc-btn.is-secondary:hover:not(:disabled):not(.is-inert) { background: var(--surface-2); border-color: var(--accent-line); }
  .j8-arc-btn.is-ghost:hover:not(:disabled):not(.is-inert) { background: var(--accent-soft); }
  .j8-arc-btn.is-danger:hover:not(:disabled):not(.is-inert) { background: var(--red-soft); border-color: var(--red); }
}

/* 禁用：主按钮整体褪色；次/危险保留描边变灰底；ghost 变灰字——不可用仍读得出是按钮 */
.j8-arc-btn:disabled { cursor: not-allowed; }
.j8-arc-btn.is-primary:disabled { opacity: 0.4; }
.j8-arc-btn.is-secondary:disabled,
.j8-arc-btn.is-danger:disabled { background: var(--surface-2); color: var(--muted); }
.j8-arc-btn.is-ghost:disabled { color: var(--muted); }

.j8-arc-btn.j8-arc-block { width: 100%; }

/* label 槽：宽度由脚本测出并弹簧形变；内容永远按自然宽度渲染，不换行 */
.j8-arc-slot { display: inline-grid; place-items: center; }
.j8-arc-content { display: inline-flex; align-items: center; white-space: nowrap; }
.j8-arc-label { display: inline-flex; align-items: center; gap: inherit; }

/* 文案切换：新字从下方轻微上浮（带模糊）进场，旧字加速上移退场 */
.j8-arc-label-text-enter-active { transition: opacity 0.2s ease, transform 0.24s cubic-bezier(0.2, 0.8, 0.2, 1), filter 0.2s ease; }
.j8-arc-label-text-leave-active { transition: opacity 0.13s ease, transform 0.13s ease, filter 0.13s ease; }
.j8-arc-label-text-enter-from { opacity: 0; transform: translateY(4px); filter: blur(4px); }
.j8-arc-label-text-leave-to { opacity: 0; transform: translateY(-3px); filter: blur(4px); }

/* 图标切换：缩放弹入 */
.j8-arc-label-icon-enter-active { transition: opacity 0.2s ease, transform 0.34s cubic-bezier(0.34, 1.45, 0.64, 1), filter 0.2s ease; }
.j8-arc-label-icon-leave-active { transition: opacity 0.12s ease, transform 0.12s ease, filter 0.12s ease; }
.j8-arc-label-icon-enter-from,
.j8-arc-label-icon-leave-to { opacity: 0; transform: scale(0.6); filter: blur(2px); }

/* 低动态：纯淡入淡出 */
.j8-arc-label-fade-enter-active,
.j8-arc-label-fade-leave-active { transition: opacity 0.15s ease; }
.j8-arc-label-fade-enter-from,
.j8-arc-label-fade-leave-to { opacity: 0; }

/* loading 转圈 */
.j8-arc-loader { display: inline-grid; place-items: center; margin-right: 8px; }
.j8-arc-spinner {
  width: 15px;
  height: 15px;
  border-radius: 50%;
  border: 2px solid color-mix(in srgb, currentColor 30%, transparent);
  border-top-color: currentColor;
  animation: j8-arc-spin 0.7s linear infinite;
}
@keyframes j8-arc-spin { to { transform: rotate(360deg); } }
.j8-arc-loader-enter-active { transition: opacity 0.18s ease, transform 0.3s cubic-bezier(0.34, 1.45, 0.64, 1); }
.j8-arc-loader-leave-active { transition: opacity 0.12s ease, transform 0.12s ease; }
.j8-arc-loader-enter-from,
.j8-arc-loader-leave-to { opacity: 0; transform: scale(0.6); }

@media (prefers-reduced-motion: reduce) {
  .j8-arc-btn,
  .j8-arc-btn:active:not(:disabled):not(.is-inert) { transform: none; transition-duration: 0.001ms; }
  .j8-arc-label-text-enter-from,
  .j8-arc-label-text-leave-to,
  .j8-arc-label-icon-enter-from,
  .j8-arc-label-icon-leave-to { transform: none; filter: none; }
}
</style>
