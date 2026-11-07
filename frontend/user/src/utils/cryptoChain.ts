/**
 * 链标识展示文案。口径与支付网关（EPUSDT 收银台）对齐：
 * BEP20 显示 "BSC"（不是 "BNB Smart Chain"）、TRON 全大写、Solana 首字母大写。
 */
const CHAIN_LABELS: Record<string, string> = {
  tron: 'TRON',
  trc20: 'TRON',
  base: 'Base',
  ethereum: 'Ethereum',
  eth: 'Ethereum',
  bsc: 'BSC',
  binance: 'BSC',
  polygon: 'Polygon',
  solana: 'Solana',
}

export const chainLabel = (value: string): string => {
  const normalized = (value || '').trim().toLowerCase()
  return CHAIN_LABELS[normalized] ?? value
}

/**
 * 独立成词的 "USDT"：前后不能紧挨字母 / 数字，所以 provider 名 epusdt / bepusdt、
 * 另一种代币 USDT0 都不算。字符串必须是未归一化的原文（CJK 等非 ASCII 字符视为分隔）。
 */
const USDT_WORD = /(^|[^a-z0-9])usdt(?![a-z0-9])/gi

export const hasUsdtToken = (text: string): boolean => {
  USDT_WORD.lastIndex = 0
  return USDT_WORD.test(text || '')
}

// 渠道名里常见的链写法 → chainLabel 的键；短词一律要求词边界，避免 subscribe / electron 之类误伤。
const USDT_CHAIN_ALIASES: Array<[RegExp, string]> = [
  [/trc-?20|\btron\b/gi, 'tron'],
  [/bep-?20|\bbsc\b|\bbinance\b/gi, 'bsc'],
  [/erc-?20|\beth(?:ereum)?\b/gi, 'ethereum'],
  [/\bsol(?:ana)?\b/gi, 'solana'],
  [/\bpolygon\b/gi, 'polygon'],
  [/\bbase\b/gi, 'base'],
]
// 币种与链之间允许出现的连接符 / 括号；去掉它们之后不能再剩任何字符。
const USDT_SEPARATORS = /[-_/·\s()（）[\]]/g

/**
 * 支付方式行的 USDT 渠道展示名。只在渠道名"纯粹是币种 + 单一链"时返回「USDT · 链名」
 * （USDT-TRC20 / usdt_trc20 / USDT (BEP20) / TRC20 USDT），其余情况返回空串、由调用方保留管理端原名：
 * 名字里没有独立的 USDT、认不出链、写了多条链（TRC20/BEP20/ERC20）或带有附加文字（（推荐）、快速到账）。
 * 只看渠道名，不看 channel_type：epusdt 等网关的真实网络在后台配置里，channel_type 可能是陈旧的 legacy 值。
 */
/** 渠道名的统一展示文案：能规范成「USDT · 链名」就用它，否则原样返回管理端填写的名字。 */
export const channelDisplayName = (name: string): string => usdtDisplayLabel(name) || name

export const usdtDisplayLabel = (name: string): string => {
  const raw = (name || '').trim()
  if (!raw || !hasUsdtToken(raw)) return ''
  const chains = USDT_CHAIN_ALIASES.filter(([pattern]) => {
    pattern.lastIndex = 0
    return pattern.test(raw)
  })
  const hit = chains.length === 1 ? chains[0] : undefined
  if (!hit) return ''
  const [chainPattern, chain] = hit
  const rest = raw
    .replace(USDT_WORD, '$1')
    .replace(chainPattern, ' ')
    .replace(USDT_SEPARATORS, '')
  return rest ? '' : `USDT · ${chainLabel(chain)}`
}
