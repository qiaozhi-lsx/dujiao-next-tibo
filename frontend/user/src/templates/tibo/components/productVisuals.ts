import type { Component } from 'vue'
import TiboXPremiumVisual from './visuals/TiboXPremiumVisual.vue'
import TiboDefaultVisual from './visuals/TiboDefaultVisual.vue'

/**
 * 商品专属视觉注册表：按商品 slug 映射到定制的卡片插画组件。
 * 未命中的商品使用 TiboDefaultVisual（信封插画）兜底。
 */
export interface TiboProductVisual {
  visual: Component
  /** 传给视觉组件的附加 props（如 X 卡的 wordmark） */
  visualProps?: Record<string, unknown>
}

const registry: Record<string, TiboProductVisual> = {
  'x-premium': {
    visual: TiboXPremiumVisual,
    visualProps: { wordmark: 'Premium', durationLabel: '3 / 6' },
  },
}

export const resolveProductVisual = (slug: string | undefined | null): TiboProductVisual => {
  if (slug && registry[slug]) return registry[slug]
  return { visual: TiboDefaultVisual }
}
