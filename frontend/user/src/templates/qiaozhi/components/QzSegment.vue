<template>
  <div class="qz-seg" role="tablist" :aria-label="ariaLabel">
    <button
      v-for="item in items"
      :key="item.key"
      :ref="(el) => setRef(item.key, el as HTMLButtonElement | null)"
      type="button"
      role="tab"
      class="qz-pill"
      :class="{ 'is-active': item.key === modelValue }"
      :aria-selected="item.key === modelValue"
      :tabindex="item.key === modelValue ? 0 : -1"
      @click="select(item.key)"
      @keydown="onKeydown($event, item.key)"
    >
      {{ item.label }}
    </button>
    <span ref="glowEl" class="qz-seg-glow" :class="{ 'is-instant': !glowReady }" aria-hidden="true" />
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

export interface QzSegmentItem {
  key: string
  label: string
}

const props = withDefaults(defineProps<{
  items: QzSegmentItem[]
  modelValue: string
  ariaLabel?: string
}>(), {
  ariaLabel: '',
})

const emit = defineEmits<{ 'update:modelValue': [key: string]; change: [key: string] }>()

const glowEl = ref<HTMLElement | null>(null)
const btnEls = new Map<string, HTMLButtonElement>()
const setRef = (key: string, el: HTMLButtonElement | null) => {
  if (el) btnEls.set(key, el)
  else btnEls.delete(key)
}

const glowReady = ref(false)
let observer: ResizeObserver | null = null

const select = (key: string) => {
  if (key === props.modelValue) return
  emit('update:modelValue', key)
  emit('change', key)
}

const syncGlow = () => {
  const glow = glowEl.value
  if (!glow) return
  const el = btnEls.get(props.modelValue)
  if (!el) {
    glow.style.opacity = '0'
    return
  }
  glow.style.opacity = '1'
  glow.style.width = `${el.offsetWidth}px`
  glow.style.height = `${el.offsetHeight}px`
  glow.style.transform = `translate(${el.offsetLeft}px, ${el.offsetTop}px)`
}

watch(() => props.modelValue, async () => {
  await nextTick()
  syncGlow()
})

const onKeydown = (event: KeyboardEvent, key: string) => {
  const keys = props.items.map((item) => item.key)
  const at = keys.indexOf(key)
  if (at < 0 || keys.length === 0) return
  let next = -1
  if (event.key === 'ArrowRight' || event.key === 'ArrowDown') next = (at + 1) % keys.length
  else if (event.key === 'ArrowLeft' || event.key === 'ArrowUp') next = (at - 1 + keys.length) % keys.length
  else if (event.key === 'Home') next = 0
  else if (event.key === 'End') next = keys.length - 1
  if (next < 0) return
  event.preventDefault()
  const targetKey = keys[next]
  if (targetKey === undefined) return
  btnEls.get(targetKey)?.focus({ preventScroll: true })
  select(targetKey)
}

onMounted(async () => {
  await nextTick()
  syncGlow()
  requestAnimationFrame(() => { glowReady.value = true })
  if (typeof ResizeObserver !== 'undefined') {
    const host = glowEl.value?.parentElement ?? null
    if (host) {
      observer = new ResizeObserver(() => syncGlow())
      observer.observe(host)
    }
  }
  if (typeof document !== 'undefined' && document.fonts) {
    void document.fonts.ready.then(() => syncGlow())
  }
})

onBeforeUnmount(() => {
  observer?.disconnect()
})
</script>

<style>
/* 分段控件：条目是 base.css 的 .qz-pill（sm 高、14px/500、hover 弱强调底、按压缩放、.is-active 白字），
   轨道与滑块在这里。轨道 = 弱化底 + 1px 内描边（box-shadow 不占尺寸，space-1 内边距 + sm 条目正好 md 高）；
   选中项的黑底由滑块提供，所以 .is-active 的 .qz-pill 底色压回透明，滑块按 offsetLeft/offsetTop 测量摆位（脚本不变）。
   不用 scoped、沿用 base.css 的作用域前缀；覆盖基类属性的选择器都带上 .qz-pill，靠特指度赢。全部取值走 --qz-* token。 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-seg {
  position: relative;
  display: flex;
  padding: var(--qz-space-1);
  border-radius: var(--qz-radius-pill);
  background: var(--qz-surface-muted);
  box-shadow: inset 0 0 0 1px var(--qz-border);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-seg .qz-pill {
  position: relative;
  z-index: 1;
  flex: 1 1 0;
  min-width: 0;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-seg .qz-pill.is-active { background: transparent; }

/* 滑块：尺寸与位移由脚本写入，首帧 .is-instant 直接摆位，之后弹簧滑动 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-seg-glow {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 0;
  width: 0;
  height: 0;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-accent);
  transition:
    transform var(--qz-duration-spring) var(--qz-ease-spring),
    width var(--qz-duration-spring) var(--qz-ease-spring),
    height var(--qz-duration-spring) var(--qz-ease-spring);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-seg-glow.is-instant { transition: none; }

@media (prefers-reduced-motion: reduce) {
  :is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-seg-glow { transition: none; }
}
</style>
