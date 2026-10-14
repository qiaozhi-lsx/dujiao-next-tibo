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
