<template>
  <button
    ref="buttonEl"
    type="button"
    class="qz-btn qz-btn--secondary qz-btn--sm qz-hc-btn"
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
    <span class="qz-hc-face">
      <span class="qz-hc-icon" aria-hidden="true">
        <Transition :name="iconSwapName">
          <span v-if="done" key="done" class="qz-hc-icon-phase">
            <svg
              class="qz-hc-check"
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
          <span v-else key="idle" class="qz-hc-icon-phase">
            <slot name="icon"><Trash2 :size="16" :stroke-width="1.75" /></slot>
          </span>
        </Transition>
      </span>
      <span ref="frameEl" class="qz-hc-frame">
        <span ref="measureEl" class="qz-hc-measure">{{ displayText }}</span>
        <Transition :name="swapName">
          <span :key="displayText" class="qz-hc-label">{{ displayText }}</span>
        </Transition>
      </span>
    </span>

    <span ref="fillEl" class="qz-hc-fill" aria-hidden="true">
      <span class="qz-hc-face">
        <span class="qz-hc-icon">
          <Transition :name="iconSwapName">
            <span v-if="done" key="done" class="qz-hc-icon-phase">
              <svg
                class="qz-hc-check"
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
            <span v-else key="idle" class="qz-hc-icon-phase">
              <slot name="icon"><Trash2 :size="16" :stroke-width="1.75" /></slot>
            </span>
          </Transition>
        </span>
        <span ref="fillFrameEl" class="qz-hc-frame">
          <span class="qz-hc-measure">{{ displayText }}</span>
          <Transition :name="swapName">
            <span :key="displayText" class="qz-hc-label">{{ displayText }}</span>
          </Transition>
        </span>
      </span>
    </span>
  </button>

  <span v-if="hint" :id="hintId" class="qz-sr-only">{{ hint }}</span>
  <span class="qz-sr-only" role="status">{{ done ? confirmedLabel : '' }}</span>
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
const swapName = computed(() => (reduced.value ? 'qz-hc-fade' : 'qz-hc-swap-text'))
const iconSwapName = computed(() => (reduced.value ? 'qz-hc-fade' : 'qz-hc-swap-icon'))

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
  if (el && !reduced.value) el.style.setProperty('--qz-press', String(el.offsetWidth > 220 ? 0.985 : 0.97))
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
/* 长按确认：外观骨架是 base.css 的 .qz-btn--secondary.qz-btn--sm（sm 高、14px/500、surface 底 + 边框、hover 弱化底、
   松开后的弹簧回弹、disabled 保形只淡字、键盘焦点环），这里只叠三样：危险色（hover / 按住 / done）、按住时的瞬时缩放、
   从左到右揭开的填充层。文案 / 图标切换的曲线与 QzButton 的 label 形变保持同一组数值，两种按钮并排时手感一致。
   不用 scoped、沿用 base.css 的作用域前缀；覆盖基类属性的选择器都带上 .qz-btn，靠特指度赢。颜色 / 尺寸全部走 --qz-* token。 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-btn.qz-hc-btn {
  /* 长按期间不能被选中文字、不能弹出 iOS 长按菜单、不能双击缩放 */
  -webkit-user-select: none;
  user-select: none;
  touch-action: manipulation;
}

/* 危险色：细指针 hover、以及任何输入方式按住蓄力时都显示；done / disabled 不响应 */
@media (hover: hover) and (pointer: fine) {
  :is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-btn.qz-hc-btn:hover:where(:not(:disabled, [aria-disabled="true"])) {
    color: var(--qz-danger);
    background: color-mix(in oklab, var(--qz-danger) 12%, transparent);
    border-color: var(--qz-danger);
    box-shadow: none;
  }
}
/* 按住：瞬时缩放（与 .qz-btn:active 同一手感，--qz-press 由脚本按宽度写入），松开由基类的弹簧 transition 弹回 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-btn.qz-hc-btn[data-state="holding"] {
  color: var(--qz-danger);
  background: color-mix(in oklab, var(--qz-danger) 12%, transparent);
  border-color: var(--qz-danger);
  transform: scale(var(--qz-press, .97));
  transition: transform var(--qz-duration-instant) var(--qz-ease-standard);
}

/* done：整体变红（填充层已铺满），由 aria-disabled 表达不可再操作，保持可聚焦 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-btn.qz-hc-btn[data-state="done"] {
  border-color: var(--qz-danger);
  cursor: default;
}

/* 面：图标 + 文案；fill 里再叠一份同款。图标盒与 .qz-btn--sm 的 16px 字形同大 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-hc-face {
  display: inline-flex;
  align-items: center;
  gap: var(--qz-space-2);
  white-space: nowrap;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-hc-icon {
  flex: none;
  display: inline-grid;
  place-items: center;
  width: 16px;
  height: 16px;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-hc-icon-phase {
  grid-area: 1 / 1;
  display: grid;
  place-items: center;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-hc-icon svg { display: block; }

/* 文案框：measure 撑出自然宽，label 绝对居中；宽度由脚本 WAAPI 弹簧形变 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-hc-frame { position: relative; display: inline-block; }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-hc-measure { display: inline-block; visibility: hidden; white-space: nowrap; }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-hc-label {
  position: absolute;
  inset: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
}

/* 危险色填充层：整层定位，clip-path 从左到右揭开；面上的字用 --qz-on-accent（亮色白字压深红、暗色深字压粉红，两边都够对比） */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-hc-fill {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: inherit;
  background: var(--qz-danger);
  color: var(--qz-on-accent);
  clip-path: inset(0 100% 0 0);
}

/* 手绘对勾：svg 没写 width/height，显式撑满 16px 图标格；pathLength=1 时 dasharray / offset 即 0..1 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-hc-check { width: 16px; height: 16px; }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-hc-check path {
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  animation: qz-hc-draw 0.32s cubic-bezier(0.2, 0.8, 0.2, 1) 0.08s forwards;
}
@keyframes qz-hc-draw { to { stroke-dashoffset: 0; } }

/* 文案切换：新字自下方上浮进场，旧字上移退场（与 QzButton 同一组时长 / 缓动 token） */
.qz-hc-swap-text-enter-active { transition: opacity var(--qz-duration-standard) var(--qz-ease-standard), transform var(--qz-duration-standard) var(--qz-ease-enter), filter var(--qz-duration-standard) var(--qz-ease-standard); }
.qz-hc-swap-text-leave-active { transition: opacity var(--qz-duration-instant) var(--qz-ease-standard), transform var(--qz-duration-instant) var(--qz-ease-standard), filter var(--qz-duration-instant) var(--qz-ease-standard); }
.qz-hc-swap-text-enter-from { opacity: 0; transform: translateY(0.3em); filter: blur(4px); }
.qz-hc-swap-text-leave-to { opacity: 0; transform: translateY(-0.3em); filter: blur(4px); }

/* 图标切换：缩放弹入 / 缩出 */
.qz-hc-swap-icon-enter-active { transition: opacity var(--qz-duration-standard) var(--qz-ease-standard), transform var(--qz-duration-spring) var(--qz-ease-spring), filter var(--qz-duration-standard) var(--qz-ease-standard); }
.qz-hc-swap-icon-leave-active { transition: opacity var(--qz-duration-instant) var(--qz-ease-standard), transform var(--qz-duration-instant) var(--qz-ease-standard), filter var(--qz-duration-instant) var(--qz-ease-standard); }
.qz-hc-swap-icon-enter-from,
.qz-hc-swap-icon-leave-to { opacity: 0; transform: scale(0.6); filter: blur(2px); }

/* 低动态：纯淡入淡出 */
.qz-hc-fade-enter-active,
.qz-hc-fade-leave-active { transition: opacity var(--qz-duration-fast) var(--qz-ease-standard); }
.qz-hc-fade-enter-from,
.qz-hc-fade-leave-to { opacity: 0; }

@media (prefers-reduced-motion: reduce) {
  :is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-btn.qz-hc-btn[data-state="holding"] { transform: none; transition: none; }
  .qz-hc-swap-text-enter-from,
  .qz-hc-swap-text-leave-to,
  .qz-hc-swap-icon-enter-from,
  .qz-hc-swap-icon-leave-to { transform: none; filter: none; }
  :is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-hc-check path { animation: none; stroke-dashoffset: 0; }
}
</style>
