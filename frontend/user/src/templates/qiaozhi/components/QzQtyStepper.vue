<template>
  <div
    class="qz-qty"
    :class="{ 'is-shake': qtyShake }"
    role="group"
    :aria-label="t('productDetail.quantity')"
    @keydown="handleQtyKeydown"
  >
    <button
      type="button"
      class="qz-qty-btn"
      :aria-disabled="modelValue <= min"
      :aria-label="t('qiaozhi.qty.decrease')"
      @pointerdown="handlePointerDown"
      @click="stepQuantity(-1)"
    >
      <Minus />
    </button>
    <!-- 读屏只听这一处：当前数量整句播报一次；下面的滚动数字只是装饰，整体对辅助技术隐藏 -->
    <span class="qz-sr-only" aria-live="polite" aria-atomic="true">{{ t('productDetail.quantityAria', { count: modelValue }) }}</span>
    <span class="qz-qty-num" aria-hidden="true">
      <span
        class="qz-qty-col"
        :class="{ 'is-hide': rollDigits.hundreds === null }"
        :style="{ transform: `translateY(calc(${-(rollDigits.hundreds || 0)} * var(--qz-qty-row)))` }"
      >
        <span v-for="d in 10" :key="`h-${d}`">{{ d - 1 }}</span>
      </span>
      <span
        class="qz-qty-col"
        :class="{ 'is-hide': rollDigits.tens === null }"
        :style="{ transform: `translateY(calc(${-(rollDigits.tens || 0)} * var(--qz-qty-row)))` }"
      >
        <span v-for="d in 10" :key="`t-${d}`">{{ d - 1 }}</span>
      </span>
      <span
        class="qz-qty-col"
        :style="{ transform: `translateY(calc(${-rollDigits.ones} * var(--qz-qty-row)))` }"
      >
        <span v-for="d in 10" :key="`o-${d}`">{{ d - 1 }}</span>
      </span>
    </span>
    <button
      type="button"
      class="qz-qty-btn is-plus"
      :aria-disabled="max !== null && modelValue >= max"
      :aria-label="t('qiaozhi.qty.increase')"
      @pointerdown="handlePointerDown"
      @click="stepQuantity(1)"
    >
      <Plus />
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Minus, Plus } from 'lucide-vue-next'

const props = withDefaults(defineProps<{
  modelValue: number
  min?: number
  max?: number | null
}>(), {
  min: 1,
  max: null,
})

const emit = defineEmits<{ 'update:modelValue': [value: number] }>()
const { t } = useI18n()

/* ---------- 滚动数字：百/十/个三列，高位为 0 时收起 ---------- */
const rollDigits = computed(() => {
  const q = Math.max(0, Number(props.modelValue) || 0)
  return {
    hundreds: q >= 100 ? Math.floor(q / 100) : null,
    tens: q >= 10 ? Math.floor(q / 10) % 10 : null,
    ones: q % 10,
  }
})

/* ---------- 边界抖动：先复位再重触发，让 CSS 动画每次都能播一遍 ---------- */
const qtyShake = ref(false)
let qtyShakeTimer: ReturnType<typeof setTimeout> | null = null
const triggerQtyShake = () => {
  qtyShake.value = false
  void nextTick(() => {
    qtyShake.value = true
    if (qtyShakeTimer) clearTimeout(qtyShakeTimer)
    qtyShakeTimer = setTimeout(() => { qtyShake.value = false }, 460)
  })
}

/* ---------- 步进：到 min/max 边界抖动提示不变化，否则 emit clamp 后的新值 ---------- */
const stepQuantity = (delta: number) => {
  if (delta < 0 && props.modelValue <= props.min) {
    triggerQtyShake()
    return
  }
  if (delta > 0 && props.max !== null && props.modelValue >= props.max) {
    triggerQtyShake()
    return
  }
  const next = props.modelValue + delta
  emit('update:modelValue', props.max !== null ? Math.min(Math.max(props.min, next), props.max) : Math.max(props.min, next))
}

const handleQtyKeydown = (event: KeyboardEvent) => {
  if (event.key === 'ArrowUp' || event.key === 'ArrowRight') {
    event.preventDefault()
    stepQuantity(1)
  } else if (event.key === 'ArrowDown' || event.key === 'ArrowLeft') {
    event.preventDefault()
    stepQuantity(-1)
  }
}

/* ---------- 按压：按被按按钮宽度写入 --qz-press（与 QzButton 同一手感规则），低动态偏好下不设 ---------- */
let reducedMq: MediaQueryList | null = null
const handlePointerDown = (event: PointerEvent) => {
  const el = event.currentTarget as HTMLElement | null
  if (!el) return
  reducedMq ??= window.matchMedia('(prefers-reduced-motion: reduce)')
  if (reducedMq.matches) return
  const width = el.offsetWidth
  el.style.setProperty('--qz-press', String(width > 220 ? 0.985 : (width > 0 && width <= 48 ? 0.96 : 0.97)))
}

onUnmounted(() => {
  if (qtyShakeTimer) clearTimeout(qtyShakeTimer)
})
</script>

<style>
/* 数量步进器：sm 控件高的胶囊容器 + 两枚 28px 圆钮（.qz-icon-btn 式中性底）+ 列滚动数字。
   只读 --qz-* token；.qz-sr-only 沿用模板全局的视觉隐藏类。
   --qz-qty-row 是一格数字的行高，模板里的 translateY 也按它算，改尺寸只动这一处。 */
.qz-qty {
  --qz-qty-row: 26px;
  display: inline-flex;
  align-items: center;
  gap: var(--qz-space-1);
  height: var(--qz-control-sm);
  padding: 0 var(--qz-space-1);
  border: 1px solid var(--qz-border-strong);
  border-radius: var(--qz-radius-pill);
  background: var(--qz-surface);
  color: var(--qz-text);
}
.qz-qty.is-shake { animation: qz-qty-shake var(--qz-duration-considered) var(--qz-ease-standard); }
@keyframes qz-qty-shake {
  20% { transform: translateX(-6px); }
  40% { transform: translateX(5px); }
  60% { transform: translateX(-3px); }
  80% { transform: translateX(2px); }
}

.qz-qty-btn {
  display: grid;
  place-items: center;
  flex: none;
  width: 28px;
  height: 28px;
  padding: 0;
  border: 0;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-surface-muted);
  color: var(--qz-text);
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  transition:
    transform var(--qz-duration-spring) var(--qz-ease-spring),
    background-color var(--qz-duration-fast) var(--qz-ease-standard),
    color var(--qz-duration-fast) var(--qz-ease-standard);
}
.qz-qty-btn svg { width: 16px; height: 16px; }
@media (hover: hover) and (pointer: fine) {
  .qz-qty-btn:hover:where(:not([aria-disabled="true"])) { background: var(--qz-accent-subtle); }
}
.qz-qty-btn:active:where(:not([aria-disabled="true"])) {
  background: var(--qz-accent-subtle);
  transform: scale(var(--qz-press, .96));
  transition: transform var(--qz-duration-instant) var(--qz-ease-standard);
}
/* 到边界仍可点（点了抖一下提示），所以用 aria-disabled 而不是 disabled：只压灰，不移出焦点序 */
.qz-qty-btn[aria-disabled="true"] { color: var(--qz-text-muted); cursor: not-allowed; }

.qz-qty-num {
  display: flex;
  justify-content: center;
  min-width: 24px;
  height: var(--qz-qty-row);
  padding: 0 var(--qz-space-1);
  overflow: hidden;
  font-size: var(--qz-text-sm);
  font-weight: 700;
  line-height: var(--qz-qty-row);
  font-variant-numeric: tabular-nums;
}
.qz-qty-col {
  display: flex;
  flex-direction: column;
  width: 10px;
  transition:
    transform var(--qz-duration-spring) var(--qz-ease-spring),
    width var(--qz-duration-standard) var(--qz-ease-standard),
    opacity var(--qz-duration-standard) var(--qz-ease-standard);
}
.qz-qty-col > span { height: var(--qz-qty-row); text-align: center; }
.qz-qty-col.is-hide { width: 0; opacity: 0; }

@media (prefers-reduced-motion: reduce) {
  .qz-qty-col, .qz-qty-btn { transition: none; }
  .qz-qty.is-shake { animation: none; }
  .qz-qty-btn:active { transform: none; }
}
</style>
