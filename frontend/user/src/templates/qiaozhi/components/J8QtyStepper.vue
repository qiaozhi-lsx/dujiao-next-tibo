<template>
  <div
    class="j8-qty"
    :class="{ 'is-shake': qtyShake }"
    role="group"
    :aria-label="t('productDetail.quantity')"
    @keydown="handleQtyKeydown"
  >
    <button
      type="button"
      class="j8-qty-btn"
      :aria-disabled="modelValue <= min"
      :aria-label="t('common.previous')"
      @pointerdown="handlePointerDown"
      @click="stepQuantity(-1)"
    >
      <Minus />
    </button>
    <span class="j8-qty-num" aria-live="polite" aria-atomic="true">
      <span class="j8-sr-only">{{ t('productDetail.quantityAria', { count: modelValue }) }}</span>
      <span
        class="j8-qty-col"
        :class="{ 'is-hide': rollDigits.hundreds === null }"
        :style="{ transform: `translateY(${-(rollDigits.hundreds || 0) * 26}px)` }"
        aria-hidden="true"
      >
        <span v-for="d in 10" :key="`h-${d}`">{{ d - 1 }}</span>
      </span>
      <span
        class="j8-qty-col"
        :class="{ 'is-hide': rollDigits.tens === null }"
        :style="{ transform: `translateY(${-(rollDigits.tens || 0) * 26}px)` }"
        aria-hidden="true"
      >
        <span v-for="d in 10" :key="`t-${d}`">{{ d - 1 }}</span>
      </span>
      <span
        class="j8-qty-col"
        :style="{ transform: `translateY(${-rollDigits.ones * 26}px)` }"
        aria-hidden="true"
      >
        <span v-for="d in 10" :key="`o-${d}`">{{ d - 1 }}</span>
      </span>
    </span>
    <button
      type="button"
      class="j8-qty-btn is-plus"
      :aria-disabled="max !== null && modelValue >= max"
      :aria-label="t('common.next')"
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

/* ---------- 按压：按被按按钮宽度写入 --j8-press（与 J8Button 同一手感规则），低动态偏好下不设 ---------- */
let reducedMq: MediaQueryList | null = null
const handlePointerDown = (event: PointerEvent) => {
  const el = event.currentTarget as HTMLElement | null
  if (!el) return
  reducedMq ??= window.matchMedia('(prefers-reduced-motion: reduce)')
  if (reducedMq.matches) return
  const width = el.offsetWidth
  el.style.setProperty('--j8-press', String(width > 220 ? 0.985 : (width > 0 && width <= 48 ? 0.96 : 0.97)))
}

onUnmounted(() => {
  if (qtyShakeTimer) clearTimeout(qtyShakeTimer)
})
</script>

<style>
/* 样式全部在 templates/qiaozhi/styles/ji8.css：.j8-qty / .j8-qty-btn / .j8-qty-num / .j8-qty-col / .j8-sr-only（含 is-shake / is-hide 与 reduced-motion），组件不重复定义 */
</style>
