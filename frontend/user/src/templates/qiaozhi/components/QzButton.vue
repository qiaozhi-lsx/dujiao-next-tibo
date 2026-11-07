<template>
  <button
    ref="buttonEl"
    :class="['qz-btn', `qz-btn--${variant}`, size !== 'md' && `qz-btn--${size}`, { 'is-loading': loading, 'is-inert': inert }]"
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
/* 外观（尺寸 / 变体 / hover / active / disabled / focus）全部在 styles/base.css 的 .qz-btn / .qz-btn--* 里，这里只管 label 形变与 loader */

/* label 槽：宽度由脚本测出并弹簧形变；内容永远按自然宽度渲染，不换行；图标与文字间距沿用 .qz-btn 的 gap */
.qz-arc-slot { display: inline-grid; place-items: center; gap: inherit; }
.qz-arc-content { display: inline-flex; align-items: center; gap: inherit; white-space: nowrap; }
.qz-arc-label { display: inline-flex; align-items: center; gap: inherit; }

/* 文案切换：新字从下方轻微上浮（带模糊）进场，旧字加速上移退场。时长 / 缓动全部读 --qz-duration-* / --qz-ease-*：
   进场 standard + enter 曲线，退场 instant，图标弹入走 --qz-ease-spring（临界阻尼，不过冲）配 --qz-duration-spring */
.qz-arc-label-text-enter-active { transition: opacity var(--qz-duration-standard) var(--qz-ease-standard), transform var(--qz-duration-standard) var(--qz-ease-enter), filter var(--qz-duration-standard) var(--qz-ease-standard); }
.qz-arc-label-text-leave-active { transition: opacity var(--qz-duration-instant) var(--qz-ease-standard), transform var(--qz-duration-instant) var(--qz-ease-standard), filter var(--qz-duration-instant) var(--qz-ease-standard); }
.qz-arc-label-text-enter-from { opacity: 0; transform: translateY(4px); filter: blur(4px); }
.qz-arc-label-text-leave-to { opacity: 0; transform: translateY(-3px); filter: blur(4px); }

/* 图标切换：缩放弹入 */
.qz-arc-label-icon-enter-active { transition: opacity var(--qz-duration-standard) var(--qz-ease-standard), transform var(--qz-duration-spring) var(--qz-ease-spring), filter var(--qz-duration-standard) var(--qz-ease-standard); }
.qz-arc-label-icon-leave-active { transition: opacity var(--qz-duration-instant) var(--qz-ease-standard), transform var(--qz-duration-instant) var(--qz-ease-standard), filter var(--qz-duration-instant) var(--qz-ease-standard); }
.qz-arc-label-icon-enter-from,
.qz-arc-label-icon-leave-to { opacity: 0; transform: scale(0.6); filter: blur(2px); }

/* 低动态：纯淡入淡出 */
.qz-arc-label-fade-enter-active,
.qz-arc-label-fade-leave-active { transition: opacity var(--qz-duration-fast) var(--qz-ease-standard); }
.qz-arc-label-fade-enter-from,
.qz-arc-label-fade-leave-to { opacity: 0; }

/* loading 转圈（与 label 的间距由 .qz-btn 的 gap 提供） */
.qz-arc-loader { display: inline-grid; place-items: center; }
.qz-arc-spinner {
  width: 15px;
  height: 15px;
  border-radius: 50%;
  border: 2px solid color-mix(in srgb, currentColor 30%, transparent);
  border-top-color: currentColor;
  animation: qz-arc-spin 0.7s linear infinite;
}
@keyframes qz-arc-spin { to { transform: rotate(360deg); } }
.qz-arc-loader-enter-active { transition: opacity var(--qz-duration-fast) var(--qz-ease-standard), transform var(--qz-duration-spring) var(--qz-ease-spring); }
.qz-arc-loader-leave-active { transition: opacity var(--qz-duration-instant) var(--qz-ease-standard), transform var(--qz-duration-instant) var(--qz-ease-standard); }
.qz-arc-loader-enter-from,
.qz-arc-loader-leave-to { opacity: 0; transform: scale(0.6); }

@media (prefers-reduced-motion: reduce) {
  .qz-arc-label-text-enter-from,
  .qz-arc-label-text-leave-to,
  .qz-arc-label-icon-enter-from,
  .qz-arc-label-icon-leave-to { transform: none; filter: none; }
}
</style>
