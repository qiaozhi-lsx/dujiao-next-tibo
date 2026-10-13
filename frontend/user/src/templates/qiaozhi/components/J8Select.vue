<template>
  <div ref="rootEl" class="j8-select" :class="attrs.class as string | undefined" :style="attrs.style as any">
    <button
      type="button"
      class="j8-select-btn"
      :class="{ 'is-open': open, 'is-placeholder': showPlaceholder }"
      :disabled="disabled"
      v-bind="buttonAttrs"
      aria-haspopup="listbox"
      :aria-expanded="open"
      :aria-activedescendant="open && highlight >= 0 ? optionDomId(highlight) : undefined"
      @click="toggle"
      @keydown="onKeydown"
      @blur="closePanel"
    >
      <span class="j8-select-label">{{ selectedItem ? selectedItem.label : placeholder }}</span>
      <ChevronDown class="j8-select-chevron" aria-hidden="true" />
    </button>
    <Transition name="j8-select-pop">
      <div v-if="open" class="j8-select-panel" role="listbox">
        <div
          v-for="(item, index) in items"
          :id="optionDomId(index)"
          :key="`${String(item.value)}-${index}`"
          role="option"
          class="j8-select-opt"
          :class="{ 'is-active': item.value === modelValue, 'is-highlight': index === highlight, 'is-disabled': item.disabled }"
          :aria-selected="item.value === modelValue"
          :aria-disabled="item.disabled || undefined"
          @pointerdown.prevent
          @click="choose(item)"
          @pointermove="onHover(index)"
        >
          <span class="j8-select-opt-text">{{ item.label }}</span>
          <Check v-if="item.value === modelValue" class="j8-select-check" aria-hidden="true" />
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useAttrs } from 'vue'
import { Check, ChevronDown } from 'lucide-vue-next'

export interface J8SelectItem {
  value: string | number
  label: string
  disabled?: boolean
}

const props = withDefaults(defineProps<{
  modelValue?: string | number | null
  items: J8SelectItem[]
  placeholder?: string
  disabled?: boolean
}>(), {
  modelValue: null,
  placeholder: '',
  disabled: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
  change: [value: string | number]
}>()

defineOptions({ inheritAttrs: false })

const attrs = useAttrs()
// id / aria-* 等透传到可聚焦的 button；class/style 留在根元素上供父级布局覆盖
const buttonAttrs = computed(() => {
  const { class: _class, style: _style, ...rest } = attrs
  return rest
})

let uid = 0
const instanceId = `j8sel-${(++uid).toString(36)}${Math.random().toString(36).slice(2, 7)}`
const optionDomId = (index: number) => `${instanceId}-opt-${index}`

const rootEl = ref<HTMLElement | null>(null)
const open = ref(false)
const highlight = ref(-1)

const selectedItem = computed<J8SelectItem | null>(() => {
  const found = props.items.find((item) => item.value === props.modelValue)
  return found ?? null
})
const showPlaceholder = computed(() => !selectedItem.value && props.placeholder !== '')

const firstEnabled = (from: number, step: number): number => {
  const count = props.items.length
  if (count === 0) return -1
  let index = from
  for (let hops = 0; hops < count; hops += 1) {
    index = (index + step + count) % count
    if (!props.items[index]?.disabled) return index
  }
  return -1
}

const scrollHighlightIntoView = () => {
  const panel = rootEl.value?.querySelector('.j8-select-panel')
  if (!panel || highlight.value < 0) return
  const el = panel.children[highlight.value] as HTMLElement | undefined
  el?.scrollIntoView({ block: 'nearest' })
}

const openPanel = (startAtEnd = false) => {
  if (props.disabled || open.value) return
  open.value = true
  const fallback = selectedItem.value
    ? props.items.findIndex((item) => item.value === props.modelValue)
    : (startAtEnd ? firstEnabled(props.items.length, -1) : firstEnabled(-1, 1))
  highlight.value = fallback
  void nextTick(() => scrollHighlightIntoView())
}

const closePanel = () => {
  open.value = false
  highlight.value = -1
}

const toggle = () => {
  if (open.value) closePanel()
  else openPanel()
}

const onHover = (index: number) => {
  if (index !== highlight.value && !props.items[index]?.disabled) highlight.value = index
}

const choose = (item: J8SelectItem) => {
  if (item.disabled) return
  closePanel()
  if (item.value !== props.modelValue) {
    emit('update:modelValue', item.value)
    emit('change', item.value)
  }
  const btn = rootEl.value?.querySelector('.j8-select-btn') as HTMLButtonElement | null
  btn?.focus({ preventScroll: true })
}

const move = (step: number) => {
  highlight.value = firstEnabled(highlight.value, step)
  void nextTick(() => scrollHighlightIntoView())
}

const onKeydown = (event: KeyboardEvent) => {
  if (props.disabled) return
  if (!open.value) {
    if (event.key === 'Enter' || event.key === ' ' || event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault()
      openPanel(event.key === 'ArrowUp')
    }
    return
  }
  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      move(1)
      break
    case 'ArrowUp':
      event.preventDefault()
      move(-1)
      break
    case 'Home':
      event.preventDefault()
      highlight.value = firstEnabled(-1, 1)
      void nextTick(() => scrollHighlightIntoView())
      break
    case 'End':
      event.preventDefault()
      highlight.value = firstEnabled(props.items.length, -1)
      void nextTick(() => scrollHighlightIntoView())
      break
    case 'Enter':
    case ' ': {
      event.preventDefault()
      const item = props.items[highlight.value]
      if (item && !item.disabled) choose(item)
      break
    }
    case 'Escape':
      event.preventDefault()
      closePanel()
      break
    case 'Tab':
      closePanel()
      break
    default:
      break
  }
}

const onDocPointerDown = (event: PointerEvent) => {
  if (!open.value) return
  const root = rootEl.value
  if (root && !root.contains(event.target as Node)) closePanel()
}

onMounted(() => {
  document.addEventListener('pointerdown', onDocPointerDown, true)
})
onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocPointerDown, true)
})
</script>
