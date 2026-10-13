<template>
  <div class="j8-es">
    <div
      ref="shellEl"
      class="j8-es-shell"
      :data-expanded="expanded ? '' : undefined"
      :style="{ width: `${shellWidth}px` }"
    >
      <button
        v-if="!expanded"
        ref="triggerEl"
        type="button"
        class="j8-es-trigger"
        :aria-label="label"
        @click="expand"
      >
        <Search aria-hidden="true" />
      </button>
      <input
        ref="inputEl"
        class="j8-es-input"
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
      <Transition name="j8-es-clear">
        <button
          v-if="expanded && query"
          type="button"
          tabindex="-1"
          class="j8-es-clear"
          :aria-label="clearLabel"
          @mousedown.prevent
          @click="clearQuery"
        >
          <X aria-hidden="true" />
        </button>
      </Transition>
    </div>

    <Transition name="j8-es-panel">
      <div
        v-if="panelOpen"
        ref="panelEl"
        class="j8-es-panel"
        :style="{ width: `${shellWidth}px` }"
        @mousedown.prevent
      >
        <div ref="panelBodyEl" class="j8-es-panel-body">
          <div v-if="mode === 'empty'" class="j8-es-empty">
            <p class="j8-es-empty-title">{{ emptyTitle(trimmedQuery) }}</p>
            <p v-if="emptyHint" class="j8-es-empty-hint">{{ emptyHint }}</p>
          </div>
          <template v-else>
            <div v-if="groupLabel" class="j8-es-group" role="presentation">{{ groupLabel }}</div>
            <div :id="listboxId" role="listbox" :aria-label="label">
              <div
                v-for="row in rows"
                :id="optionId(row.item)"
                :key="row.item.id"
                role="option"
                :aria-selected="row.index === activeIndex"
                class="j8-es-option"
                :class="{ 'is-active': row.index === activeIndex }"
                @pointermove="onHover(row.index)"
                @mousedown.prevent
                @click="choose(row.item)"
              >
                <span class="j8-es-option-title">
                  <span>{{ row.parts.pre }}</span><mark v-if="row.parts.hit" class="j8-es-mark">{{ row.parts.hit }}</mark><span>{{ row.parts.post }}</span>
                </span>
                <span v-if="row.item.meta" class="j8-es-option-meta">{{ row.item.meta }}</span>
              </div>
            </div>
          </template>
        </div>
      </div>
    </Transition>

    <span class="j8-sr-only" role="status" aria-live="polite">{{ announcement }}</span>
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
const instanceId = `j8es-${(++uid).toString(36)}${Math.random().toString(36).slice(2, 7)}`
const listboxId = `${instanceId}-listbox`
const optionId = (item: SearchableItem) => `${instanceId}-opt-${String(item.id)}`

/* ---------- 减小动态偏好 ---------- */
const reduced = ref(false)
let mq: MediaQueryList | null = null
const syncReduced = () => { reduced.value = mq?.matches ?? false }

/* ---------- 展开 / 收起 ---------- */
const COLLAPSED_SIZE = 46
const expanded = ref(false)
const focused = ref(false)
const query = ref(props.modelValue)
const shellWidth = ref(COLLAPSED_SIZE)
const shellEl = ref<HTMLElement | null>(null)
const inputEl = ref<HTMLInputElement | null>(null)
const triggerEl = ref<HTMLButtonElement | null>(null)

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
  if (!el || reduced.value || from === goal || typeof el.animate !== 'function') return
  el.animate(
    [{ width: `${from}px` }, { width: `${goal}px` }],
    { duration: 420, easing: 'cubic-bezier(0.34, 1.45, 0.64, 1)' },
  )
}

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
  morphTo(COLLAPSED_SIZE)
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
})

onBeforeUnmount(() => {
  panelObserver?.disconnect()
  mq?.removeEventListener?.('change', syncReduced)
})

defineExpose({ expand, collapse, focus: () => inputEl.value?.focus() })
</script>

<style>
.j8-es { position: relative; }
.j8-es-shell {
  position: relative;
  display: flex;
  align-items: center;
  height: 46px;
  overflow: hidden;
  border: 1px solid var(--line-strong);
  border-radius: var(--r-pill, 999px);
  background: var(--surface);
  box-shadow: var(--shadow-soft);
  transition: border-color 0.15s;
}
.j8-es-shell:hover { border-color: var(--accent-line); }
.j8-es-shell:focus-within { border-color: var(--accent); }
.j8-es-trigger {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  border: 0;
  border-radius: inherit;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  transition: background-color 0.15s;
}
.j8-es-trigger:hover { background: var(--surface-2); }
.j8-es-trigger svg { width: 18px; height: 18px; }
.j8-es-input {
  flex: 1 1 0%;
  min-width: 0;
  height: 100%;
  padding: 0 12px;
  border: 0;
  background: transparent;
  color: var(--text);
  font-size: 13px;
  outline: none;
  opacity: 0;
  transition: opacity 0.08s ease;
}
.j8-es-shell[data-expanded] .j8-es-input {
  opacity: 1;
  transition: opacity 0.2s ease 0.08s;
}
.j8-es-input::placeholder { color: var(--muted); }
.j8-es-clear {
  flex: 0 0 auto;
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  margin-right: 7px;
  padding: 0;
  border: 0;
  border-radius: 50%;
  background: var(--surface-2);
  color: var(--muted);
  cursor: pointer;
}
.j8-es-clear:hover { background: var(--surface-3); color: var(--text); }
.j8-es-clear svg { width: 15px; height: 15px; }
.j8-es-clear-enter-active { transition: opacity 0.18s ease, transform 0.3s cubic-bezier(0.34, 1.45, 0.64, 1); }
.j8-es-clear-leave-active { transition: opacity 0.1s ease, transform 0.1s ease; }
.j8-es-clear-enter-from,
.j8-es-clear-leave-to { opacity: 0; transform: scale(0.6); }

.j8-es-panel {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  z-index: 45;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: var(--r-card);
  background: var(--surface);
  box-shadow: var(--shadow);
}
.j8-es-panel-body { max-height: 320px; overflow-y: auto; padding: 8px; }
.j8-es-panel-enter-active { transition: opacity 0.16s ease; }
.j8-es-panel-leave-active { transition: opacity 0.1s ease; }
.j8-es-panel-enter-from,
.j8-es-panel-leave-to { opacity: 0; }
.j8-es-group {
  padding: 6px 10px 4px;
  color: var(--muted);
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.j8-es-option {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 12px;
  color: var(--text);
  font-size: 13px;
  cursor: pointer;
  transition: background-color 0.12s;
}
.j8-es-option.is-active { background: var(--accent-soft); }
.j8-es-option-title { min-width: 0; overflow-wrap: anywhere; }
.j8-es-mark { background: transparent; color: var(--accent-strong); font-weight: 900; }
.j8-es-option-meta { flex: 0 0 auto; color: var(--muted); font-size: 12px; font-variant-numeric: tabular-nums; }
.j8-es-empty { padding: 14px 12px 10px; }
.j8-es-empty-title { margin: 0 0 4px; font-size: 13px; font-weight: 700; color: var(--text); }
.j8-es-empty-hint { margin: 0; font-size: 12px; color: var(--muted); }

@media (prefers-reduced-motion: reduce) {
  .j8-es-input,
  .j8-es-shell[data-expanded] .j8-es-input { transition-duration: 0.001ms; transition-delay: 0s; }
}
</style>
