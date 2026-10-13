<template>
  <div
    class="j8-tabs"
    :data-orientation="orientation"
    :data-overflow="edges.overflow ? '' : undefined"
  >
    <button
      v-if="isHorizontal && edges.overflow"
      type="button"
      class="j8-tabs-arrow is-left"
      :disabled="!edges.left"
      :aria-label="scrollLeftLabel"
      @click="scrollByPage(-1)"
    >
      <ChevronLeft aria-hidden="true" />
    </button>
    <div ref="viewportEl" class="j8-tabs-viewport" @scroll.passive="updateEdges">
      <div ref="listEl" class="j8-tabs-list" role="tablist" :aria-orientation="orientation" :aria-label="ariaLabel">
        <button
          v-for="item in items"
          :key="item.key"
          :ref="(el) => setTabRef(item.key, el as HTMLButtonElement | null)"
          type="button"
          role="tab"
          class="j8-tabs-tab"
          :class="{ 'is-active': item.key === modelValue }"
          :aria-selected="item.key === modelValue"
          :tabindex="item.key === modelValue ? 0 : -1"
          @click="select(item.key)"
          @keydown="onKeydown($event, item.key)"
        >
          <component :is="item.icon" v-if="item.icon" aria-hidden="true" />
          <span class="j8-tabs-label">{{ item.label }}</span>
        </button>
        <span ref="glowEl" class="j8-tabs-glow" :class="{ 'is-instant': !glowReady }" aria-hidden="true" />
      </div>
    </div>
    <button
      v-if="isHorizontal && edges.overflow"
      type="button"
      class="j8-tabs-arrow is-right"
      :disabled="!edges.right"
      :aria-label="scrollRightLabel"
      @click="scrollByPage(1)"
    >
      <ChevronRight aria-hidden="true" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch, type Component } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

export interface J8TabItem {
  key: string
  label: string
  icon?: Component
}

const props = withDefaults(defineProps<{
  items: J8TabItem[]
  modelValue: string
  orientation?: 'vertical' | 'horizontal'
  ariaLabel?: string
  scrollLeftLabel?: string
  scrollRightLabel?: string
}>(), {
  orientation: 'vertical',
  ariaLabel: '',
  scrollLeftLabel: 'Scroll left',
  scrollRightLabel: 'Scroll right',
})

const emit = defineEmits<{ select: [key: string] }>()

const isHorizontal = computed(() => props.orientation === 'horizontal')

const listEl = ref<HTMLElement | null>(null)
const viewportEl = ref<HTMLElement | null>(null)
const glowEl = ref<HTMLElement | null>(null)
const tabEls = new Map<string, HTMLButtonElement>()
const setTabRef = (key: string, el: HTMLButtonElement | null) => {
  if (el) tabEls.set(key, el)
  else tabEls.delete(key)
}

const edges = ref({ overflow: false, left: false, right: false })
const reduced = ref(false)
const glowReady = ref(false) // 首帧无过渡直接摆位，之后启用弹簧滑动
let listObserver: ResizeObserver | null = null

const select = (key: string) => {
  if (key !== props.modelValue) emit('select', key)
}

const updateEdges = () => {
  const scroll = viewportEl.value
  if (!scroll || !isHorizontal.value) {
    if (edges.value.overflow) edges.value = { overflow: false, left: false, right: false }
    return
  }
  const max = Math.max(0, scroll.scrollWidth - scroll.clientWidth)
  const next = {
    overflow: scroll.scrollWidth > scroll.clientWidth + 1,
    left: scroll.scrollLeft > 1,
    right: scroll.scrollLeft < max - 1,
  }
  const prev = edges.value
  if (prev.overflow !== next.overflow || prev.left !== next.left || prev.right !== next.right) {
    edges.value = next
  }
}

const scrollByPage = (dir: number) => {
  const scroll = viewportEl.value
  if (!scroll) return
  scroll.scrollBy({ left: dir * scroll.clientWidth * 0.75, behavior: reduced.value ? 'auto' : 'smooth' })
}

// 激活项不在可视区内时滚到可见
const reveal = (el: HTMLElement | null) => {
  const scroll = viewportEl.value
  if (!scroll || !el || !isHorizontal.value) return
  const frame = scroll.getBoundingClientRect()
  const item = el.getBoundingClientRect()
  const max = Math.max(0, scroll.scrollWidth - scroll.clientWidth)
  const left = frame.left + (scroll.scrollLeft > 1 ? 36 : 0)
  const right = frame.right - (scroll.scrollLeft < max - 1 ? 36 : 0)
  const delta = item.left < left ? item.left - left : item.right > right ? item.right - right : 0
  if (delta) scroll.scrollBy({ left: delta, behavior: reduced.value ? 'auto' : 'smooth' })
}

// 高亮滑块摆到激活项上
const syncGlow = () => {
  const glow = glowEl.value
  if (!glow) return
  const el = tabEls.get(props.modelValue)
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
  reveal(tabEls.get(props.modelValue) ?? null)
  updateEdges()
})

const onKeydown = (event: KeyboardEvent, key: string) => {
  const keys = props.items.map((item) => item.key)
  const at = keys.indexOf(key)
  if (at < 0 || keys.length === 0) return
  let next = -1
  const vertical = !isHorizontal.value
  if ((vertical && event.key === 'ArrowDown') || (!vertical && event.key === 'ArrowRight')) {
    next = (at + 1) % keys.length
  } else if ((vertical && event.key === 'ArrowUp') || (!vertical && event.key === 'ArrowLeft')) {
    next = (at - 1 + keys.length) % keys.length
  } else if (event.key === 'Home') {
    next = 0
  } else if (event.key === 'End') {
    next = keys.length - 1
  }
  if (next < 0) return
  event.preventDefault()
  const targetKey = keys[next]
  if (targetKey === undefined) return
  tabEls.get(targetKey)?.focus({ preventScroll: true })
  select(targetKey) // 导航语义：焦点移动即切换
}

onMounted(async () => {
  reduced.value = typeof window !== 'undefined'
    && window.matchMedia('(prefers-reduced-motion: reduce)').matches
  await nextTick()
  syncGlow()
  requestAnimationFrame(() => { glowReady.value = true })
  updateEdges()
  if (typeof ResizeObserver !== 'undefined' && listEl.value) {
    listObserver = new ResizeObserver(() => {
      syncGlow()
      updateEdges()
    })
    listObserver.observe(listEl.value)
    if (viewportEl.value) listObserver.observe(viewportEl.value)
  }
  // 字体后载会改变标签宽度，就绪后补摆一次
  if (typeof document !== 'undefined' && document.fonts) {
    void document.fonts.ready.then(() => syncGlow())
  }
})

onBeforeUnmount(() => {
  listObserver?.disconnect()
})
</script>
