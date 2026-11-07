import { nextTick, onBeforeUnmount, onMounted, ref, watch, type Ref } from 'vue'

/**
 * 高亮滑块测位 composable（QzTabs / QzSegment 共用）
 *
 * 原理：测激活 item 的 offsetWidth/Height/Left/Top，写入 glowEl 的 opacity/width/height/transform。
 * 首帧 glowReady=false 让模板套上 .is-instant（无过渡直接摆位），下一帧翻 true 开启弹簧动画。
 * 监听 host（默认 glowEl.parentElement）尺寸变化，自动补摆；字体后载也补一次。
 */
export interface UseGlowHighlightOptions {
  /** 滑块元素 ref */
  glowEl: Ref<HTMLElement | null>
  /** 主 host ref；可选，默认 glowEl.parentElement。resize 时会触发 syncGlow */
  listEl?: Ref<HTMLElement | null>
  /** 额外要观察的 host（如 QzTabs 的 viewport），resize 时也触发 syncGlow + onResize */
  extraHosts?: Ref<HTMLElement | null>[]
  /** 根据当前状态拿到激活 item 的 HTMLElement，无激活返回 null */
  getActive: () => HTMLElement | null
  /** 任何会影响激活位置的依赖；watch 的 source（返回值变化即触发 syncGlow） */
  deps: () => unknown
  /** resize 时额外副作用（如 QzTabs 的 updateEdges） */
  onResize?: () => void
}

export function useGlowHighlight(opts: UseGlowHighlightOptions) {
  const glowReady = ref(false)
  let observer: ResizeObserver | null = null

  function syncGlow() {
    const glow = opts.glowEl.value
    if (!glow) return
    const el = opts.getActive()
    if (!el) {
      glow.style.opacity = '0'
      return
    }
    glow.style.opacity = '1'
    glow.style.width = `${el.offsetWidth}px`
    glow.style.height = `${el.offsetHeight}px`
    glow.style.transform = `translate(${el.offsetLeft}px, ${el.offsetTop}px)`
  }

  onMounted(async () => {
    await nextTick()
    syncGlow()
    requestAnimationFrame(() => { glowReady.value = true })
    if (typeof ResizeObserver !== 'undefined') {
      const primary = opts.listEl?.value ?? opts.glowEl.value?.parentElement ?? null
      observer = new ResizeObserver(() => {
        syncGlow()
        opts.onResize?.()
      })
      if (primary) observer.observe(primary)
      if (opts.extraHosts) {
        for (const h of opts.extraHosts) {
          if (h.value) observer.observe(h.value)
        }
      }
    }
    // 字体后载会改变标签宽度，就绪后补摆一次
    if (typeof document !== 'undefined' && document.fonts) {
      void document.fonts.ready.then(() => syncGlow())
    }
  })

  onBeforeUnmount(() => {
    observer?.disconnect()
    observer = null
  })

  // flush:'post' 保证 DOM 已更新再测位
  watch(() => opts.deps(), () => syncGlow(), { flush: 'post' })

  return { glowReady, syncGlow }
}
