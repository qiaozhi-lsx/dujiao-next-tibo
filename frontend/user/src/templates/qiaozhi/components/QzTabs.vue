<template>
  <div
    class="qz-tabs"
    :data-orientation="orientation"
    :data-overflow="edges.overflow ? '' : undefined"
  >
    <button
      v-if="isHorizontal && edges.overflow"
      type="button"
      class="qz-icon-btn qz-icon-btn--sm qz-tabs-arrow is-left"
      :disabled="!edges.left"
      :aria-label="scrollLeftLabel"
      @click="scrollByPage(-1)"
    >
      <ChevronLeft aria-hidden="true" />
    </button>
    <div ref="viewportEl" class="qz-tabs-viewport" @scroll.passive="updateEdges">
      <div ref="listEl" class="qz-tabs-list" role="tablist" :aria-orientation="orientation" :aria-label="ariaLabel">
        <button
          v-for="item in items"
          :key="item.key"
          :ref="(el) => setTabRef(item.key, el as HTMLButtonElement | null)"
          type="button"
          role="tab"
          class="qz-pill qz-tabs-tab"
          :class="{ 'is-active': item.key === modelValue }"
          :aria-selected="item.key === modelValue"
          :tabindex="item.key === modelValue ? 0 : -1"
          @click="select(item.key)"
          @keydown="onKeydown($event, item.key)"
        >
          <component :is="item.icon" v-if="item.icon" aria-hidden="true" />
          <span class="qz-tabs-label">{{ item.label }}</span>
        </button>
        <span ref="glowEl" class="qz-tabs-glow" :class="{ 'is-instant': !glowReady }" aria-hidden="true" />
      </div>
    </div>
    <button
      v-if="isHorizontal && edges.overflow"
      type="button"
      class="qz-icon-btn qz-icon-btn--sm qz-tabs-arrow is-right"
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

export interface QzTabItem {
  key: string
  label: string
  icon?: Component
}

const props = withDefaults(defineProps<{
  items: QzTabItem[]
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

<style>
/* Tabs 导航：条目是 base.css 的 .qz-pill（sm 高、14px/500、hover 弱强调底、按压缩放、.is-active 白字），
   溢出箭头是 .qz-icon-btn--sm；只有滑块（.qz-tabs-glow）与排布在这里。选中项的黑底由滑块提供，
   所以 .is-active 的 .qz-pill 底色压回透明，滑块按 offsetLeft/offsetTop 测量摆位（脚本不变）。
   不用 scoped、沿用 base.css 的作用域前缀：覆盖基类属性的选择器都带上 .qz-pill，靠特指度而不是加载顺序赢。
   全部取值走 --qz-* token；键盘焦点环是全局规则，视口留 space-1 的出血位让环不被 overflow 裁掉。 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-tabs {
  display: flex;
  align-items: center;
  gap: var(--qz-space-1);
  min-width: 0;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-tabs-viewport {
  flex: 1 1 auto;
  min-width: 0;
  margin: calc(-1 * var(--qz-space-1));
  padding: var(--qz-space-1);
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-tabs-viewport::-webkit-scrollbar { display: none; }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-tabs-list {
  position: relative;
  display: flex;
  gap: var(--qz-space-1);
  width: 100%;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-tabs[data-orientation="vertical"] .qz-tabs-list { flex-direction: column; }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-tabs[data-orientation="horizontal"] .qz-tabs-list { width: max-content; }

/* 条目：竖排占满一行、内容靠左；横排按内容宽 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-pill.qz-tabs-tab {
  position: relative;
  z-index: 1;
  flex: 1 0 auto;
  justify-content: flex-start;
  text-align: left;
}
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-tabs[data-orientation="horizontal"] .qz-pill.qz-tabs-tab { flex: 0 0 auto; }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-pill.qz-tabs-tab.is-active { background: transparent; }
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-tabs-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 滑块：尺寸与位移由脚本写入，首帧 .is-instant 直接摆位，之后弹簧滑动 */
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-tabs-glow {
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
:is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-tabs-glow.is-instant { transition: none; }

@media (prefers-reduced-motion: reduce) {
  :is(.qiaozhi-scope, :where(body.qiaozhi-tokens)) .qz-tabs-glow { transition: none; }
}
</style>
