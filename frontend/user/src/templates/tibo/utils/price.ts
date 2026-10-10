import { computed } from 'vue'
import { useAppStore } from '../../../stores/app'
import { amountToCents, centsToAmount } from '../../../utils/money'

const CURRENCY_SYMBOLS: Record<string, string> = {
  CNY: '¥',
  USD: '$',
  EUR: '€',
  GBP: '£',
  HKD: 'HK$',
  JPY: '¥',
  TWD: 'NT$',
  SGD: 'S$',
  KRW: '₩',
  AUD: 'A$',
  CAD: 'C$',
}

export const currencySymbol = (code: string): string => {
  const upper = String(code || '').trim().toUpperCase()
  if (!upper) return ''
  return CURRENCY_SYMBOLS[upper] ?? `${upper} `
}

export const formatTiboPrice = (amount: unknown, currency: string): string => {
  if (amount === null || amount === undefined || amount === '') return '-'
  const cents = amountToCents(amount)
  const displayAmount = cents === null ? String(amount) : centsToAmount(cents)
  return `${currencySymbol(currency)}${displayAmount}`
}

/** 拆分为 ¥ / 29 / .90 三段，用于大字号价格排版（对齐 gptibo 的价签样式） */
export const splitTiboPrice = (amount: unknown, currency: string): { symbol: string; major: string; minor: string } => {
  const full = formatTiboPrice(amount, currency)
  const symbol = currencySymbol(currency)
  if (full === '-') return { symbol: '', major: '-', minor: '' }
  const rest = symbol && full.startsWith(symbol) ? full.slice(symbol.length) : full
  const dot = rest.indexOf('.')
  if (dot === -1) return { symbol, major: rest, minor: '' }
  return { symbol, major: rest.slice(0, dot), minor: rest.slice(dot) }
}

export function useTiboPrice() {
  const appStore = useAppStore()
  const currency = computed(() => {
    const raw = String(appStore.config?.currency || '').trim().toUpperCase()
    return /^[A-Z]{3}$/.test(raw) ? raw : 'CNY'
  })

  return {
    currency,
    symbol: computed(() => currencySymbol(currency.value)),
    format: (amount: unknown) => formatTiboPrice(amount, currency.value),
    split: (amount: unknown) => splitTiboPrice(amount, currency.value),
  }
}
