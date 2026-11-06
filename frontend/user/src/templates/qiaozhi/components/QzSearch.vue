<template>
  <div class="qz-es">
    <div
      ref="shellEl"
      class="qz-es-shell"
      :data-expanded="expanded ? '' : undefined"
      :style="{ width: `${shellWidth}px` }"
    >
      <button
        v-if="!expanded"
        ref="triggerEl"
        type="button"
        class="qz-es-trigger"
        :aria-label="label"
        @click="expand"
      >
        <Search aria-hidden="true" />
        <span v-if="triggerLabel" class="qz-es-trigger-label" aria-hidden="true">{{ triggerLabel }}</span>
      </button>
      <input
        ref="inputEl"
        class="qz-es-input"
        type="text"
        role="combobox"
        :inert="!expanded || undefined"
        :value="query"
        :placeholder="placeholder"
        :aria-label="label"
        :aria-expanded="panelOpen"
        :aria-controls="panelOpen && mode === 'results' ? listboxId : undefined"
        aria-autocomplete="list"
        :aria-activedescendant="panelOpen && activeItem ? optionId(activeItem) : undefined"
        autocomplete="off"
        autocorrect="off"
        spellcheck="false"
        enterkeyhint="search"
        @input="onInput"
        @focus="focused = true"
        @blur="onBlur"
        @keydown="onKeydown"
      />
      <Transition name="qz-es-clear">
        <button
          v-if="expanded && query"
          type="button"
          tabindex="-1"
          class="qz-es-clear"
          :aria-label="clearLabel"
          @mousedown.prevent
          @click="clearQuery"
        >
          <X aria-hidden="true" />
        </button>
      </Transition>
    </div>

    <Transition name="qz-es-panel">
      <div
        v-if="panelOpen"
        ref="panelEl"
        class="qz-es-panel"
        :style="{ width: `${shellWidth}px` }"
        @mousedown.prevent
      >
        <div ref="panelBodyEl" class="qz-es-panel-body">
          <div v-if="mode === 'empty'" class="qz-es-empty">
            <p class="qz-es-empty-title">{{ emptyTitle(trimmedQuery) }}</p>
            <p v-if="emptyHint" class="qz-es-empty-hint">{{ emptyHint }}</p>
          </div>
          <template v-else>
            <div v-if="groupLabel" class="qz-es-group" role="presentation">{{ groupLabel }}</div>
            <div :id="listboxId" role="listbox" :aria-label="label">
              <div
                v-for="row in rows"
                :id="optionId(row.item)"
                :key="row.item.id"
                role="option"
                :aria-selected="row.index === activeIndex"
                class="qz-es-option"
                :class="{ 'is-active': row.index === activeIndex }"
                @pointermove="onHover(row.index)"
                @mousedown.prevent
                @click="choose(row.item)"
              >
                <span class="qz-es-option-title">
                  <span>{{ row.parts.pre }}</span><mark v-if="row.parts.hit" class="qz-es-mark">{{ row.parts.hit }}</mark><span>{{ row.parts.post }}</span>
                </span>
                <span v-if="row.item.meta" class="qz-es-option-meta">{{ row.item.meta }}</span>
              </div>
            </div>
          </template>
        </div>
      </div>
    </Transition>

    <span class="qz-sr-only" role="status" aria-live="polite">{{ announcement }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Search, X } from 'lucide-vue-next'
import { rankSearchItems, type SearchableItem } from '../utils/searchRank'

const props = withDefaults(defineProps<{
  modelValue?: string
  items?: SearchableItem[]
  label?: string
  triggerLabel?: string
  placeholder?: string
  emptyHint?: string
  groupLabel?: string
  clearLabel?: string
  expandedWidth?: number
  maxResults?: number
  emptyTitle?: (query: string) => string
  resultsText?: (count: number) => string
}>(), {
  modelValue: '',
  items: () => [],
  label: 'Search',
  triggerLabel: '',
  placeholder: '',
  emptyHint: '',
  groupLabel: '',
  clearLabel: 'Clear',
  expandedWidth: 340,
  maxResults: 6,
  emptyTitle: (query: string) => `No results for “${query}”`,
  resultsText: (count: number) => `${count} results`,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  select: [item: SearchableItem]
}>()

let uid = 0
const instanceId = `qzes-${(++uid).toString(36)}${Math.random().toString(36).slice(2, 7)}`
const listboxId = `${instanceId}-listbox`
const optionId = (item: SearchableItem) => `${instanceId}-opt-${String(item.id)}`

/* ---------- 减小动态偏好 ---------- */
const reduced = ref(false)
let mq: MediaQueryList | null = null
const syncReduced = () => { reduced.value = mq?.matches ?? false }

/* ---------- 展开 / 收起 ---------- */
// 折叠态宽度在 mounted 后实测（有标签取自然宽度，无标签取壳高凑正方形）；这里只是首帧兜底，对应 --qz-control-sm 桌面值。
const COLLAPSED_SIZE = 36
const expanded = ref(false)
const focused = ref(false)
const query = ref(props.modelValue)
const shellWidth = ref(COLLAPSED_SIZE)
const collapsedWidth = ref(COLLAPSED_SIZE)
const shellEl = ref<HTMLElement | null>(null)
const inputEl = ref<HTMLInputElement | null>(null)
const triggerEl = ref<HTMLButtonElement | null>(null)
let morphAnim: Animation | null = null

watch(() => props.modelValue, (value) => {
  if (value !== query.value) query.value = value
})
watch(query, (value) => emit('update:modelValue', value))

const expandedTarget = () =>
  Math.min(props.expandedWidth, (typeof window !== 'undefined' ? window.innerWidth : props.expandedWidth) - 96)

const morphTo = (goal: number) => {
  const el = shellEl.value
  const from = el?.offsetWidth ?? goal
  shellWidth.value = goal
  if (!el || reduced.value || from === goal || typeof el.animate !== 'function') return null
  morphAnim = el.animate(
    [{ width: `${from}px` }, { width: `${goal}px` }],
    { duration: 420, easing: 'cubic-bezier(0.34, 1.45, 0.64, 1)' },
  )
  return morphAnim
}

// 折叠宽度实测：临时置宽 auto 并隐藏输入框，读外壳自然宽度后还原；无标签时取壳高（--qz-control-sm）做正方形图标钮
const measureCollapsedWidth = () => {
  if (expanded.value || (morphAnim && morphAnim.playState === 'running')) return
  const el = shellEl.value
  if (!el || !triggerEl.value) return
  const input = inputEl.value
  const prevWidth = el.style.width
  const prevDisplay = input?.style.display ?? ''
  el.style.width = 'auto'
  if (input) input.style.display = 'none'
  const rect = el.getBoundingClientRect()
  const measured = Math.ceil(props.triggerLabel ? rect.width : rect.height)
  if (input) input.style.display = prevDisplay
  el.style.width = prevWidth
  const goal = measured > 0 ? measured : COLLAPSED_SIZE
  collapsedWidth.value = goal
  shellWidth.value = goal
}

watch(() => props.triggerLabel, () => {
  void nextTick(measureCollapsedWidth)
})

const expand = () => {
  if (expanded.value) return
  expanded.value = true
  // 输入框始终在 DOM 中；同步移除 inert 后立即聚焦，触摸键盘也能在同一次点击里弹出
  inputEl.value?.removeAttribute('inert')
  inputEl.value?.focus({ preventScroll: true })
  morphTo(expandedTarget())
}

const collapse = (returnFocus: boolean) => {
  if (!expanded.value) return
  expanded.value = false
  query.value = ''
  active.value = -1
  const anim = morphTo(collapsedWidth.value)
  // 收起动画落定后补测一次：兜住展开期间标签 / 字体变化
  if (anim) void anim.finished.then(measureCollapsedWidth).catch(() => { /* 动画被取消时忽略 */ })
  else void nextTick(measureCollapsedWidth)
  if (returnFocus) void nextTick(() => triggerEl.value?.focus({ preventScroll: true }))
}

/* ---------- 结果 ---------- */
const trimmedQuery = computed(() => query.value.trim())
const results = computed(() => trimmedQuery.value ? rankSearchItems(props.items, trimmedQuery.value, props.maxResults) : [])
const flat = computed(() => results.value)
const mode = computed<'results' | 'empty' | null>(() => {
  if (!trimmedQuery.value) return null
  return flat.value.length ? 'results' : 'empty'
})
const panelOpen = computed(() => expanded.value && focused.value && mode.value !== null)

const active = ref(-1)
const activeIndex = computed(() => Math.min(active.value, flat.value.length - 1))
const activeItem = computed(() => flat.value[activeIndex.value])

const rows = computed(() => flat.value.map((item, index) => ({
  item,
  index,
  parts: matchParts(String(item.title ?? '')),
})))

const matchParts = (text: string) => {
  const q = trimmedQuery.value
  if (!q) return { pre: text, hit: '', post: '' }
  const at = text.toLocaleLowerCase().indexOf(q.toLocaleLowerCase())
  if (at < 0) return { pre: text, hit: '', post: '' }
  return { pre: text.slice(0, at), hit: text.slice(at, at + q.length), post: text.slice(at + q.length) }
}

const announcement = computed(() => {
  if (!panelOpen.value) return ''
  if (mode.value === 'empty') return props.emptyTitle(trimmedQuery.value)
  return props.resultsText(flat.value.length)
})

/* ---------- 交互 ---------- */
const onInput = (event: Event) => {
  query.value = (event.target as HTMLInputElement).value
  active.value = trimmedQuery.value ? 0 : -1
}

const onHover = (index: number) => {
  if (index !== activeIndex.value) active.value = index
}

const onBlur = () => {
  // 切换窗口时保持展开，回来继续输入
  if (typeof document !== 'undefined' && !document.hasFocus()) return
  focused.value = false
  if (expanded.value && !query.value.trim()) collapse(false)
}

const onKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    event.preventDefault()
    collapse(true)
    return
  }
  if (!panelOpen.value || !flat.value.length) return
  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    event.preventDefault()
    const step = event.key === 'ArrowDown' ? 1 : -1
    active.value = activeIndex.value < 0
      ? (step > 0 ? 0 : flat.value.length - 1)
      : (activeIndex.value + step + flat.value.length) % flat.value.length
  } else if (event.key === 'Enter' && activeItem.value) {
    event.preventDefault()
    choose(activeItem.value)
  }
}

const choose = (item: SearchableItem) => {
  collapse(true)
  emit('select', item)
}

const clearQuery = () => {
  query.value = ''
  active.value = -1
  inputEl.value?.focus({ preventScroll: true })
}

/* ---------- 结果面板：宽度随壳、首开从 0 展开、内容变化时高度弹到位 ---------- */
const panelEl = ref<HTMLElement | null>(null)
const panelBodyEl = ref<HTMLElement | null>(null)
let panelObserver: ResizeObserver | null = null
let panelHeight = 0

watch(panelOpen, async (open) => {
  panelObserver?.disconnect()
  panelObserver = null
  if (!open) return
  await nextTick()
  const panel = panelEl.value
  const body = panelBodyEl.value
  if (!panel || !body) return
  panelHeight = 0
  panel.style.height = '0px'
  const apply = () => {
    const next = body.offsetHeight + 2
    if (next === panelHeight) return
    const from = panelHeight
    panelHeight = next
    panel.style.height = `${next}px`
    if (reduced.value || from === 0 || typeof panel.animate !== 'function') return
    panel.animate(
      [{ height: `${from}px` }, { height: `${next}px` }],
      { duration: 380, easing: 'cubic-bezier(0.34, 1.45, 0.64, 1)' },
    )
  }
  apply()
  if (typeof ResizeObserver !== 'undefined') {
    panelObserver = new ResizeObserver(apply)
    panelObserver.observe(body)
  }
})

onMounted(() => {
  mq = window.matchMedia('(prefers-reduced-motion: reduce)')
  syncReduced()
  mq.addEventListener?.('change', syncReduced)
  measureCollapsedWidth()
  // 字体后加载会改变文字宽度，字体就绪后补测一次
  if (typeof document !== 'undefined' && document.fonts) {
    void document.fonts.ready.then(measureCollapsedWidth)
  }
})

onBeforeUnmount(() => {
  panelObserver?.disconnect()
  mq?.removeEventListener?.('change', syncReduced)
})

defineExpose({ expand, collapse, focus: () => inputEl.value?.focus() })
</script>

<style>
/* 展开式搜索。壳体与触发钮对齐 .qz-pill（base.css）：sm 高、14px/500、图标 16、hover 底 --qz-accent-subtle、
   按压缩放、键盘焦点环；壳体多一圈 --qz-border，和同排的 .qz-btn--secondary.qz-btn--sm「分享」钮同高同边框。
   触发钮在 overflow:hidden 的壳里，焦点环画在壳上（:has），自身的环关掉。全部取值走 --qz-* token。 */
.qz-es { position: relative; }
.qz-es-shell {
  position: relative;
  display: flex;
  align-items: center;
  height: var(--qz-control-sm);
  overflow: hidden;
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-pill);
  background: var(--qz-surface);
  color: var(--qz-text-secondary);
  transition:
    transform var(--qz-duration-spring) var(--qz-ease-spring),
    color var(--qz-duration-fast) var(--qz-ease-standard),
    background-color var(--qz-duration-fast) var(--qz-ease-standard),
    border-color var(--qz-duration-fast) var(--qz-ease-standard);
}
.qz-es-shell:not([data-expanded]) { justify-content: center; }
@media (hover: hover) and (pointer: fine) {
  .qz-es-shell:not([data-expanded]):hover { color: var(--qz-text); background: var(--qz-accent-subtle); }
}
.qz-es-shell:not([data-expanded]):active {
  color: var(--qz-text);
  background: var(--qz-accent-subtle);
  transform: scale(var(--qz-press, 0.97));
  transition: transform var(--qz-duration-instant) var(--qz-ease-standard);
}
.qz-es-shell[data-expanded] { color: var(--qz-text); }
.qz-es-shell[data-expanded]:focus-within { border-color: var(--qz-accent); }
/* 键盘焦点：环画在壳上，和分享钮的 .qz-btn:focus-visible 一致 */
.qz-es-shell:has(> .qz-es-trigger:focus-visible) {
  outline: 2px solid var(--qz-focus-outline);
  outline-offset: 2px;
}
.qz-es .qz-es-shell .qz-es-trigger:focus-visible { outline: none; }
.qz-es-trigger {
  display: inline-flex;
  flex: 0 0 auto;
  align-self: stretch;
  align-items: center;
  justify-content: center;
  gap: var(--qz-space-2);
  padding: 0 var(--qz-space-3);
  border: 0;
  border-radius: inherit;
  background: transparent;
  color: inherit;
  font: inherit;
  font-size: var(--qz-text-sm);
  font-weight: 500;
  line-height: var(--qz-leading-body);
  white-space: nowrap;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}
.qz-es-trigger svg { flex: 0 0 auto; width: 16px; height: 16px; }
.qz-es-input {
  flex: 1 1 0%;
  min-width: 0;
  height: 100%;
  padding: 0 var(--qz-space-3);
  border: 0;
  background: transparent;
  color: var(--qz-text);
  font: inherit;
  font-size: var(--qz-text-sm);
  outline: none;
  opacity: 0;
  transition: opacity var(--qz-duration-instant) var(--qz-ease-standard);
}
.qz-es-shell:not([data-expanded]) .qz-es-input {
  flex: 0 0 0%;
  width: 0;
  padding: 0;
}
.qz-es-shell[data-expanded] .qz-es-input {
  opacity: 1;
  transition: opacity var(--qz-duration-fast) var(--qz-ease-standard) var(--qz-duration-instant);
}
.qz-es-input::placeholder { color: var(--qz-text-muted); }
.qz-es-clear {
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  width: 28px;
  height: 28px;
  margin-right: var(--qz-space-1);
  padding: 0;
  border: 0;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-surface-muted);
  color: var(--qz-text-muted);
  cursor: pointer;
  transition:
    color var(--qz-duration-fast) var(--qz-ease-standard),
    background-color var(--qz-duration-fast) var(--qz-ease-standard);
}
@media (hover: hover) and (pointer: fine) {
  .qz-es-clear:hover { color: var(--qz-text); background: var(--qz-accent-subtle); }
}
.qz-es-clear:active { color: var(--qz-text); background: var(--qz-accent-subtle); }
.qz-es-clear svg { width: 14px; height: 14px; }
.qz-es-clear-enter-active { transition: opacity var(--qz-duration-fast) var(--qz-ease-standard), transform var(--qz-duration-standard) var(--qz-ease-enter); }
.qz-es-clear-leave-active { transition: opacity var(--qz-duration-instant) var(--qz-ease-standard), transform var(--qz-duration-instant) var(--qz-ease-standard); }
.qz-es-clear-enter-from,
.qz-es-clear-leave-to { opacity: 0; transform: scale(0.6); }

/* ---------- 结果面板 ---------- */
.qz-es-panel {
  position: absolute;
  top: calc(100% + var(--qz-space-2));
  left: 0;
  z-index: 45;
  overflow: hidden;
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface);
  color: var(--qz-text);
  box-shadow: var(--qz-shadow-floating);
}
.qz-es-panel-body { max-height: 320px; overflow-y: auto; padding: var(--qz-space-2); }
.qz-es-panel-enter-active { transition: opacity var(--qz-duration-fast) var(--qz-ease-standard); }
.qz-es-panel-leave-active { transition: opacity var(--qz-duration-instant) var(--qz-ease-standard); }
.qz-es-panel-enter-from,
.qz-es-panel-leave-to { opacity: 0; }
.qz-es-group {
  padding: var(--qz-space-1) var(--qz-space-2);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}
.qz-es-option {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--qz-space-2);
  padding: var(--qz-space-2) var(--qz-space-3);
  border-radius: var(--qz-space-3);
  color: var(--qz-text);
  font-size: var(--qz-text-sm);
  line-height: var(--qz-leading-body);
  cursor: pointer;
  transition: background-color var(--qz-duration-instant) var(--qz-ease-standard);
}
.qz-es-option.is-active { background: var(--qz-accent-subtle); }
.qz-es-option-title { min-width: 0; overflow-wrap: anywhere; }
.qz-es-mark { background: transparent; color: var(--qz-text); font-weight: 700; }
.qz-es-option-meta { flex: 0 0 auto; color: var(--qz-text-muted); font-size: var(--qz-text-xs); font-variant-numeric: tabular-nums; }
.qz-es-empty { padding: var(--qz-space-3) var(--qz-space-3) var(--qz-space-2); }
.qz-es-empty-title { margin: 0 0 var(--qz-space-1); color: var(--qz-text); font-size: var(--qz-text-sm); font-weight: 600; }
.qz-es-empty-hint { margin: 0; color: var(--qz-text-muted); font-size: var(--qz-text-xs); }

@media (prefers-reduced-motion: reduce) {
  .qz-es-shell,
  .qz-es-clear,
  .qz-es-option { transition: none; }
  .qz-es-shell:not([data-expanded]):active { transform: none; }
  .qz-es-input,
  .qz-es-shell[data-expanded] .qz-es-input { transition-duration: 0.001ms; transition-delay: 0s; }
}
</style>
