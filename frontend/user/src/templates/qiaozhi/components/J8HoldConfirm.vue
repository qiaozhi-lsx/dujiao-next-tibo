<template>
  <button
    ref="buttonEl"
    type="button"
    class="j8-hc-btn"
    :data-state="done ? 'done' : holding ? 'holding' : 'idle'"
    :disabled="disabled || undefined"
    :aria-disabled="done || undefined"
    :aria-label="displayText"
    :aria-describedby="!done && hint ? hintId : undefined"
    v-bind="$attrs"
    @pointerdown="onPointerDown"
    @pointermove="onPointerMove"
    @pointerup="release"
    @pointercancel="release"
    @lostpointercapture="release"
    @keydown="onKeydown"
    @keyup="onKeyup"
    @blur="release"
    @contextmenu.prevent
  >
    <!-- 底层正常色面;fill 里再叠一份同款面,文字白色,clip-path 裁切处自然换色 -->
    <span class="j8-hc-face">
      <span class="j8-hc-icon" aria-hidden="true">
        <Transition :name="iconSwapName">
          <span v-if="done" key="done" class="j8-hc-icon-phase">
            <svg
              class="j8-hc-check"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M4 12.5l5 5L20 6.5" pathLength="1" />
            </svg>
          </span>
          <span v-else key="idle" class="j8-hc-icon-phase">
            <slot name="icon"><Trash2 :size="14" :stroke-width="1.75" /></slot>
          </span>
        </Transition>
      </span>
      <span ref="frameEl" class="j8-hc-frame">
        <span ref="measureEl" class="j8-hc-measure">{{ displayText }}</span>
        <Transition :name="swapName">
          <span :key="displayText" class="j8-hc-label">{{ displayText }}</span>
        </Transition>
      </span>
    </span>

    <span ref="fillEl" class="j8-hc-fill" aria-hidden="true">
      <span class="j8-hc-face">
        <span class="j8-hc-icon">
          <Transition :name="iconSwapName">
            <span v-if="done" key="done" class="j8-hc-icon-phase">
              <svg
                class="j8-hc-check"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M4 12.5l5 5L20 6.5" pathLength="1" />
              </svg>
            </span>
            <span v-else key="idle" class="j8-hc-icon-phase">
              <slot name="icon"><Trash2 :size="14" :stroke-width="1.75" /></slot>
            </span>
          </Transition>
        </span>
        <span ref="fillFrameEl" class="j8-hc-frame">
          <span class="j8-hc-measure">{{ displayText }}</span>
          <Transition :name="swapName">
            <span :key="displayText" class="j8-hc-label">{{ displayText }}</span>
          </Transition>
        </span>
      </span>
    </span>
  </button>

  <span v-if="hint" :id="hintId" class="j8-hc-sr">{{ hint }}</span>
  <span class="j8-hc-sr" role="status">{{ done ? confirmedLabel : '' }}</span>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, useId, watch } from 'vue'
import { Trash2 } from 'lucide-vue-next'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{
  /** 常态文案,如「移除」 */
  label: string
  /** 完成文案,如「已移除」 */
  confirmedLabel: string
  /** 长按时长 ms */
  duration?: number
  disabled?: boolean
  /** sr-only 提示全文(父级传入已插值的文案) */
  hint?: string
  /** 外部控制 done 态;不传则内部自管 */
  confirmed?: boolean
}>(), {
  duration: 1200,
  disabled: false,
  hint: '',
})

const emit = defineEmits<{ confirm: [] }>()
defineSlots<{ icon?: () => any }>()

const ownDone = ref(false)
const done = computed(() => props.confirmed ?? ownDone.value)
const holding = ref(false)
const completions = ref(0)
const displayText = computed(() => (done.value ? props.confirmedLabel : props.label))
const hintId = useId()

const buttonEl = ref<HTMLButtonElement | null>(null)
const fillEl = ref<HTMLElement | null>(null)
const frameEl = ref<HTMLElement | null>(null)
const fillFrameEl = ref<HTMLElement | null>(null)
const measureEl = ref<HTMLElement | null>(null)

/* ---------- 减小动态偏好 ---------- */
const reduced = ref(false)
let mq: MediaQueryList | null = null
const syncReduced = () => { reduced.value = mq?.matches ?? false }
const swapName = computed(() => (reduced.value ? 'j8-hc-fade' : 'j8-hc-swap-text'))
const iconSwapName = computed(() => (reduced.value ? 'j8-hc-fade' : 'j8-hc-swap-icon'))

/* ---------- 运动常量:回退 / 补满 / 宽度形变共用同一条 spring 曲线 ---------- */
const SPRING_BEZIER = [0.34, 1.45, 0.64, 1] as const
const SPRING_EASE = `cubic-bezier(${SPRING_BEZIER.join(', ')})`
const SPRING_MS = 350

/** JS 版 cubic-bezier:牛顿迭代 + 二分兜底,rAF 回退/补满用 */
function createBezier(x1: number, y1: number, x2: number, y2: number) {
  const cx = 3 * x1
  const bx = 3 * (x2 - x1) - cx
  const ax = 1 - cx - bx
  const cy = 3 * y1
  const by = 3 * (y2 - y1) - cy
  const ay = 1 - cy - by
  const sampleX = (t: number) => ((ax * t + bx) * t + cx) * t
  const sampleY = (t: number) => ((ay * t + by) * t + cy) * t
  const slopeX = (t: number) => (3 * ax * t + 2 * bx) * t + cx
  return (x: number) => {
    if (x <= 0) return 0
    if (x >= 1) return 1
    let t = x
    for (let i = 0; i < 8; i += 1) {
      const error = sampleX(t) - x
      if (Math.abs(error) < 1e-5) return sampleY(t)
      const slope = slopeX(t)
      if (Math.abs(slope) < 1e-6) break
      t = Math.min(1, Math.max(0, t - error / slope))
    }
    let low = 0
    let high = 1
    t = x
    for (let i = 0; i < 24; i += 1) {
      const error = sampleX(t) - x
      if (Math.abs(error) < 1e-5) break
      if (error > 0) high = t
      else low = t
      t = (low + high) / 2
    }
    return sampleY(t)
  }
}
const springTween = createBezier(...SPRING_BEZIER)

/* ---------- 进度引擎:线性蓄力 / spring 回退 / spring 补满 rAF 驱动 ---------- */
let progress = 0
let mode: 'idle' | 'hold' | 'rewind' | 'settle' = 'idle'
let rafId = 0
let holdBase = 0
let holdStart = 0
let tweenFrom = 0
let tweenStart = 0
let source: 'pointer' | 'key' | null = null
let pointerKind = 'mouse'

function applyProgress() {
  const value = Math.min(1, Math.max(0, progress))
  if (fillEl.value) fillEl.value.style.clipPath = `inset(0 ${((1 - value) * 100).toFixed(3)}% 0 0)`
}

function cancelRaf() {
  if (rafId) cancelAnimationFrame(rafId)
  rafId = 0
  mode = 'idle'
}

function tick(now: number) {
  rafId = 0
  if (mode === 'hold') {
    progress = holdBase + Math.max(0, now - holdStart) / Math.max(1, props.duration)
    if (progress >= 1) {
      progress = 1
      applyProgress()
      mode = 'idle'
      complete()
      return
    }
    applyProgress()
    rafId = requestAnimationFrame(tick)
    return
  }
  if (mode === 'rewind' || mode === 'settle') {
    const t = Math.min(1, (now - tweenStart) / SPRING_MS)
    const eased = springTween(t)
    progress = mode === 'rewind' ? tweenFrom * (1 - eased) : tweenFrom + (1 - tweenFrom) * eased
    if (t >= 1) {
      progress = mode === 'rewind' ? 0 : 1
      applyProgress()
      mode = 'idle'
      return
    }
    applyProgress()
    rafId = requestAnimationFrame(tick)
  }
}

function stopHolding() {
  source = null
  holding.value = false
}

function complete() {
  if (!source) return
  stopHolding()
  if (pointerKind === 'touch') navigator.vibrate?.(12)
  ownDone.value = true
  completions.value += 1
  emit('confirm')
}

/** 从当前进度续力:holdBase 记录按下瞬间的进度,回退中途再按也不归零 */
function begin(from: 'pointer' | 'key') {
  if (done.value || props.disabled || source) return
  cancelRaf()
  const el = buttonEl.value
  if (el && !reduced.value) el.style.setProperty('--j8-press', String(el.offsetWidth > 220 ? 0.985 : 0.97))
  source = from
  holding.value = true
  holdBase = Math.min(1, Math.max(0, progress))
  holdStart = performance.now()
  mode = 'hold'
  rafId = requestAnimationFrame(tick)
}

function rewind() {
  cancelRaf()
  if (progress <= 0) {
    progress = 0
    applyProgress()
    return
  }
  if (reduced.value) {
    progress = 0
    applyProgress()
    return
  }
  tweenFrom = progress
  tweenStart = performance.now()
  mode = 'rewind'
  rafId = requestAnimationFrame(tick)
}

function release() {
  if (!source) return
  stopHolding()
  rewind()
}

function settle() {
  cancelRaf()
  if (progress >= 1) {
    applyProgress()
    return
  }
  if (reduced.value) {
    progress = 1
    applyProgress()
    return
  }
  tweenFrom = progress
  tweenStart = performance.now()
  mode = 'settle'
  rafId = requestAnimationFrame(tick)
}

/* ---------- 指针:仅主键蓄力,滑出按钮 24px 外取消 ---------- */
function onPointerDown(event: PointerEvent) {
  if (!event.isPrimary || event.button !== 0) return
  if (done.value || props.disabled) return
  pointerKind = event.pointerType
  try {
    buttonEl.value?.setPointerCapture(event.pointerId)
  } catch {
    /* 忽略异常指针 */
  }
  begin('pointer')
}

function onPointerMove(event: PointerEvent) {
  if (source !== 'pointer') return
  const box = buttonEl.value?.getBoundingClientRect()
  if (!box) return
  const slack = 24
  if (
    event.clientX < box.left - slack ||
    event.clientX > box.right + slack ||
    event.clientY < box.top - slack ||
    event.clientY > box.bottom + slack
  ) release()
}

/* ---------- 键盘:空格 / Enter 按住蓄力,忽略 repeat;keyup 或 blur 取消 ---------- */
function onKeydown(event: KeyboardEvent) {
  if (event.key !== ' ' && event.key !== 'Enter') return
  event.preventDefault()
  if (event.repeat) return
  begin('key')
}

function onKeyup(event: KeyboardEvent) {
  if (event.key !== ' ' && event.key !== 'Enter') return
  event.preventDefault()
  if (source === 'key') release()
}

/* ---------- 宽度形变:文案变化时 WAAPI 弹簧到新宽,字体后载 / 首次测量直接跳变 ---------- */
let frameWidth = 0
let lastText: string | null = null

function setFrameWidth(next: number, from: number | null) {
  const nodes: Array<HTMLElement | null> = [frameEl.value, fillFrameEl.value]
  const shouldAnimate = from !== null && !reduced.value && Math.abs(next - from) > 0.5
  for (const el of nodes) {
    if (!el) continue
    el.style.width = `${next}px`
    if (shouldAnimate && from !== null && typeof el.animate === 'function') {
      el.animate([{ width: `${from}px` }, { width: `${next}px` }], { duration: SPRING_MS, easing: SPRING_EASE })
    }
  }
  frameWidth = next
}

/* ---------- 挂载 ---------- */
let observer: ResizeObserver | null = null

onMounted(() => {
  mq = window.matchMedia('(prefers-reduced-motion: reduce)')
  syncReduced()
  mq.addEventListener?.('change', syncReduced)

  const node = measureEl.value
  if (node && typeof ResizeObserver !== 'undefined') {
    observer = new ResizeObserver((entries) => {
      const entry = entries[0]
      // 布局宽度而非绘制盒:done 文案可能在被按压的缩放态下被测量
      const next = entry?.borderBoxSize?.[0]?.inlineSize ?? node.offsetWidth
      if (!next) return
      const text = node.textContent ?? ''
      const from = frameWidth > 0 ? frameWidth : null
      const morph = lastText !== null && lastText !== text && from !== null
      lastText = text
      setFrameWidth(next, morph ? from : null)
    })
    observer.observe(node)
  }

  applyProgress()
  /* 挂载即 confirmed=true:填充补满,和参考实现的首帧同步一致 */
  if (done.value && progress < 1 && !source) settle()
})

onBeforeUnmount(() => {
  cancelRaf()
  observer?.disconnect()
  mq?.removeEventListener?.('change', syncReduced)
})

/* ---------- done 同步:外部置 done 补满;复位或父级未接受确认则回收填充 ---------- */
watch([done, completions], () => {
  if (done.value && progress < 1 && !source) settle()
  else if (!done.value && progress > 0 && !source) rewind()
})

watch(() => props.disabled, (value) => {
  if (value) release()
})
</script>

<style>
/* 长按确认:高 34px 与购物车工具按钮一致;三态 idle / hover / holding + done */
.j8-hc-btn {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 34px;
  padding: 0px 13px;
  border: 1px solid var(--line-strong);
  border-radius: var(--r-pill, 999px);
  background: var(--surface);
  color: var(--text);
  font-size: 12.5px;
  font-weight: 800;
  line-height: 1;
  white-space: nowrap;
  cursor: pointer;
  user-select: none;
  -webkit-user-select: none;
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
  transition: transform 0.42s cubic-bezier(0.34, 1.45, 0.64, 1), background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease, opacity 0.2s ease;
}

/* 按住:瞬时缩放,松开由基础 transition 弹回 */
.j8-hc-btn[data-state="holding"] {
  transform: scale(var(--j8-press, 0.97));
  transition: transform 0.05s ease, background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease, opacity 0.2s ease;
}

/* done:整体变红,由 aria-disabled 表达不可再操作,保持可聚焦 */
.j8-hc-btn[data-state="done"] { border-color: var(--red); cursor: default; }

.j8-hc-btn:disabled { cursor: not-allowed; background: var(--surface-2); border-color: var(--line); color: var(--muted); }

@media (hover: hover) and (pointer: fine) {
  .j8-hc-btn:hover:not(:disabled):not([data-state="done"]) {
    color: var(--red);
    background: var(--red-soft);
    border-color: var(--red);
  }
}

/* 面:图标 + 文案;fill 里再叠一份同款 */
.j8-hc-face { display: inline-flex; align-items: center; gap: 6px; white-space: nowrap; }
.j8-hc-icon { flex: 0 0 auto; width: 14px; height: 14px; display: inline-grid; place-items: center; }
.j8-hc-icon-phase { grid-area: 1 / 1; display: grid; place-items: center; }
.j8-hc-icon svg { display: block; width: 14px; height: 14px; }

/* 文案框:measure 撑出自然宽,label 绝对居中;宽度由脚本 WAAPI 弹簧形变 */
.j8-hc-frame { position: relative; display: inline-block; }
.j8-hc-measure { display: inline-block; visibility: hidden; white-space: nowrap; }
.j8-hc-label { position: absolute; inset: 0px; display: inline-flex; align-items: center; justify-content: center; white-space: nowrap; }

/* 红色填充面:整层定位,clip-path 从左到右揭开,顶层文字白色 */
.j8-hc-fill {
  position: absolute;
  inset: 0px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: inherit;
  background: var(--red);
  color: #fff;
  clip-path: inset(0 100% 0 0);
}

/* 手绘对勾:pathLength=1 时 dasharray / offset 即 0..1 */
.j8-hc-check path {
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  animation: j8-hc-draw 0.32s cubic-bezier(0.2, 0.8, 0.2, 1) 0.08s forwards;
}
@keyframes j8-hc-draw { to { stroke-dashoffset: 0; } }

/* 文案切换:新字自下方上浮进场,旧字上移退场 */
.j8-hc-swap-text-enter-active { transition: opacity 0.2s ease, transform 0.24s cubic-bezier(0.2, 0.8, 0.2, 1), filter 0.2s ease; }
.j8-hc-swap-text-leave-active { transition: opacity 0.13s ease, transform 0.13s ease, filter 0.13s ease; }
.j8-hc-swap-text-enter-from { opacity: 0; transform: translateY(0.3em); filter: blur(4px); }
.j8-hc-swap-text-leave-to { opacity: 0; transform: translateY(-0.3em); filter: blur(4px); }

/* 图标切换:缩放弹入 / 缩出 */
.j8-hc-swap-icon-enter-active { transition: opacity 0.2s ease, transform 0.34s cubic-bezier(0.34, 1.45, 0.64, 1), filter 0.2s ease; }
.j8-hc-swap-icon-leave-active { transition: opacity 0.12s ease, transform 0.12s ease, filter 0.12s ease; }
.j8-hc-swap-icon-enter-from,
.j8-hc-swap-icon-leave-to { opacity: 0; transform: scale(0.6); filter: blur(2px); }

/* 低动态:纯淡入淡出 */
.j8-hc-fade-enter-active,
.j8-hc-fade-leave-active { transition: opacity 0.15s ease; }
.j8-hc-fade-enter-from,
.j8-hc-fade-leave-to { opacity: 0; }

/* sr-only:hint 说明与 done 播报 */
.j8-hc-sr { position: absolute; width: 1px; height: 1px; margin: -1px; padding: 0px; border: 0px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }

@media (prefers-reduced-motion: reduce) {
  .j8-hc-btn,
  .j8-hc-btn[data-state="holding"] { transform: none; transition-duration: 0.001ms; }
  .j8-hc-swap-text-enter-from,
  .j8-hc-swap-text-leave-to,
  .j8-hc-swap-icon-enter-from,
  .j8-hc-swap-icon-leave-to { transform: none; filter: none; }
  .j8-hc-check path { animation: none; stroke-dashoffset: 0; }
}
</style>
