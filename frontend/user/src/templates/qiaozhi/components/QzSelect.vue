<template>
  <div ref="rootEl" class="qz-select" :class="attrs.class as string | undefined" :style="attrs.style as any">
    <button
      type="button"
      class="qz-field-control qz-select-btn"
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
      <span class="qz-select-label">{{ selectedItem ? selectedItem.label : placeholder }}</span>
      <ChevronDown class="qz-select-chevron" aria-hidden="true" />
    </button>
    <Transition name="qz-select-pop">
      <div v-if="open" class="qz-select-panel" role="listbox">
        <div
          v-for="(item, index) in items"
          :id="optionDomId(index)"
          :key="`${String(item.value)}-${index}`"
          role="option"
          class="qz-select-opt"
          :class="{ 'is-active': item.value === modelValue, 'is-highlight': index === highlight, 'is-disabled': item.disabled }"
          :aria-selected="item.value === modelValue"
          :aria-disabled="item.disabled || undefined"
          @pointerdown.prevent
          @click="choose(item)"
          @pointermove="onHover(index)"
        >
          <span class="qz-select-opt-text">{{ item.label }}</span>
          <Check v-if="item.value === modelValue" class="qz-select-check" aria-hidden="true" />
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useAttrs } from 'vue'
import { Check, ChevronDown } from 'lucide-vue-next'

export interface QzSelectItem {
  value: string | number
  label: string
  disabled?: boolean
}

const props = withDefaults(defineProps<{
  modelValue?: string | number | null
  items: QzSelectItem[]
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
const instanceId = `qzsel-${(++uid).toString(36)}${Math.random().toString(36).slice(2, 7)}`
const optionDomId = (index: number) => `${instanceId}-opt-${index}`

const rootEl = ref<HTMLElement | null>(null)
const open = ref(false)
const highlight = ref(-1)

const selectedItem = computed<QzSelectItem | null>(() => {
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
  const panel = rootEl.value?.querySelector('.qz-select-panel')
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

const choose = (item: QzSelectItem) => {
  if (item.disabled) return
  closePanel()
  if (item.value !== props.modelValue) {
    emit('update:modelValue', item.value)
    emit('change', item.value)
  }
  const btn = rootEl.value?.querySelector('.qz-select-btn') as HTMLButtonElement | null
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

<style>
/* 自绘下拉：触发钮就是 base.css 的 .qz-field-control（md 高、control 圆角、强边框；hover 边框加深、
   :focus-within / aria-invalid / :disabled 全部由基类接管，箭头按基类的 18px 弱化色），这里只补
   按钮形态（两端对齐、左对齐文案、占位色、打开时箭头翻转）与浮层面板。
   不用 scoped、沿用 base.css 的作用域前缀（浮层若日后 Teleport 到 body 也能命中）；覆盖基类属性的选择器带上
   .qz-field-control，靠特指度赢。面板用 control 圆角，选项圆角按 space-1 内边距同心收一圈。全部取值走 --qz-* token。 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select {
  position: relative;
  min-width: 0;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-field-control.qz-select-btn {
  justify-content: space-between;
  text-align: left;
  cursor: pointer;
}
/* 展开态描边；aria-invalid 时让 base 的 danger 描边赢（与 base 里 [aria-invalid] 压过 :focus-within 的顺序一致） */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-field-control.qz-select-btn.is-open:where(:not([aria-invalid="true"])) { border-color: var(--qz-accent); }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select-btn.is-placeholder .qz-select-label { color: var(--qz-text-muted); }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select-label {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select-chevron { transition: transform var(--qz-duration-fast) var(--qz-ease-standard); }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select-btn.is-open .qz-select-chevron { transform: rotate(180deg); }

/* 浮层面板：贴在触发钮下方，浮起阴影；选项 sm 高（手机 40，可点）、高亮走弱强调底、选中加粗 + 对勾 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select-panel {
  position: absolute;
  top: calc(100% + var(--qz-space-1));
  right: 0;
  left: 0;
  z-index: 60;
  max-height: 280px;
  overflow-y: auto;
  padding: var(--qz-space-1);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface-raised);
  box-shadow: var(--qz-shadow-floating);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select-opt {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--qz-space-2);
  min-height: var(--qz-control-sm);
  padding: 0 var(--qz-space-3);
  border-radius: calc(var(--qz-radius-control) - var(--qz-space-1));
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
  cursor: pointer;
  transition: background-color var(--qz-duration-instant) var(--qz-ease-standard);
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select-opt.is-highlight { background: var(--qz-accent-subtle); }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select-opt.is-active { color: var(--qz-accent-strong); font-weight: 600; }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select-opt.is-disabled { opacity: .5; cursor: not-allowed; }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select-opt-text {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-select-check {
  flex: none;
  width: 16px;
  height: 16px;
  color: var(--qz-accent-strong);
}

/* 面板进出：淡入 + 自上方 space-1 落下 */
.qz-select-pop-enter-active {
  transition:
    opacity var(--qz-duration-fast) var(--qz-ease-standard),
    transform var(--qz-duration-standard) var(--qz-ease-enter);
}
.qz-select-pop-leave-active {
  transition:
    opacity var(--qz-duration-instant) var(--qz-ease-standard),
    transform var(--qz-duration-instant) var(--qz-ease-standard);
}
.qz-select-pop-enter-from,
.qz-select-pop-leave-to {
  opacity: 0;
  transform: translateY(calc(-1 * var(--qz-space-1)));
}

@media (prefers-reduced-motion: reduce) {
  :is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) :is(.qz-select-chevron, .qz-select-opt) { transition: none; }
  .qz-select-pop-enter-active,
  .qz-select-pop-leave-active { transition: none; }
}
</style>
