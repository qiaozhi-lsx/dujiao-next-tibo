/** 滚动显现系统：.j8-reveal.is-pending 进入视口后加 .is-in 播放 j8-rise，并按序错峰。 */
export const setupReveal = (root: HTMLElement | null | undefined): (() => void) => {
  if (!root) return () => undefined
  const els = Array.from(root.querySelectorAll<HTMLElement>('.j8-reveal.is-pending'))
  if (!els.length) return () => undefined

  if (!('IntersectionObserver' in window)) {
    els.forEach((el) => {
      el.classList.remove('is-pending')
      el.classList.add('is-in')
    })
    return () => undefined
  }

  els.forEach((el, i) => {
    el.style.setProperty('--d', `${(i % 6) * 55}ms`)
  })

  const io = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (!entry.isIntersecting) continue
        const el = entry.target as HTMLElement
        el.classList.remove('is-pending')
        el.classList.add('is-in')
        io.unobserve(el)
      }
    },
    { rootMargin: '0px 0px -6% 0px', threshold: 0.06 },
  )
  els.forEach((el) => io.observe(el))

  // 兜底：若观察器因异常未触发，1.8s 后强制显现首屏内的卡片，避免内容隐身
  const fallback = window.setTimeout(() => {
    for (const el of els) {
      if (!el.classList.contains('is-pending')) continue
      const rect = el.getBoundingClientRect()
      if (rect.top < window.innerHeight && rect.bottom > 0) {
        el.classList.remove('is-pending')
        el.classList.add('is-in')
        io.unobserve(el)
      }
    }
  }, 1800)

  return () => {
    window.clearTimeout(fallback)
    io.disconnect()
  }
}
