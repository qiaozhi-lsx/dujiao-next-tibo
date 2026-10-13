<template>
  <div class="j8-seg" role="tablist" :aria-label="ariaLabel">
    <button
      v-for="item in items"
      :key="item.key"
      :ref="(el) => setRef(item.key, el as HTMLButtonElement | null)"
      type="button"
      role="tab"
      :class="{ 'is-active': item.key === modelValue }"
      :aria-selected="item.key === modelValue"
      :tabindex="item.key === modelValue ? 0 : -1"
      @click="select(item.key)"
      @keydown="onKeydown($event, item.key)"
    >
      {{ item.label }}
    </button>
    <span ref="glowEl" class="j8-seg-glow" :class="{ 'is-instant': !glowReady }" aria-hidden="true" />
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

export interface J8SegmentItem {
  key: string
  label: string
}

const props = withDefaults(defineProps<{
  items: J8SegmentItem[]
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
