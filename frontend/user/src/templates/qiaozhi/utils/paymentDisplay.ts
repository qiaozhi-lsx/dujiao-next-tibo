/**
 * 支付弹窗（PaymentModal）展示判断的纯函数：币种选取、手续费是否展示、
 * 订单已付 / 终态判断、创单后走内嵌弹窗还是独立支付页、不带币种的裸金额。
 * 与 DOM / Vue 无关，便于 node:test 直接覆盖。
 */
import { isCustomerSurchargePayment, resolvePaymentPresentationMode } from '../../../utils/paymentResumePolicy.ts'

const trimmed = (value: unknown): string => String(value || '').trim()

/** 实付币种优先（如钱包抵扣 / 换汇后的币种），缺省回退订单币种。 */
export const resolvePayableCurrency = (o: { currency?: string; payableCurrency?: string }): string => {
  return trimmed(o.payableCurrency) || trimmed(o.currency)
}

/** 仅"客户承担手续费"且手续费金额非零时才展示手续费行。 */
export const shouldShowFee = (o: { feeAmount?: string; feePolicy?: string }): boolean => {
  const raw = trimmed(o.feeAmount)
  if (!raw || raw === '0' || raw === '0.00') return false
  return isCustomerSurchargePayment({ fee_policy: o.feePolicy })
}

export const PAID_ORDER_STATUSES: ReadonlySet<string> = new Set([
  'paid',
  'fulfilling',
  'partially_delivered',
  'delivered',
  'completed',
])

export const isPaidOrderStatus = (s: string): boolean => PAID_ORDER_STATUSES.has(s)

export const TERMINAL_ORDER_STATUSES: ReadonlySet<string> = new Set(['canceled', 'cancelled', 'expired', 'refunded'])

export const isTerminalOrderStatus = (s: string): boolean => TERMINAL_ORDER_STATUSES.has(s)

/**
 * 创单后是否走内嵌支付弹窗：有支付载荷（二维码或支付链接）、订单未已付（如钱包全额抵扣的 order_paid），
 * 且交互模式解析为 qr（空模式视为 qr，与 classic 一致）；redirect / wap / page 走独立支付页。
 */
export const shouldOpenInlineModal = (
  payData: { qr_code?: unknown; pay_url?: unknown; order_paid?: unknown; interaction_mode?: unknown } | null | undefined,
): boolean => {
  if (!payData || payData.order_paid) return false
  if (!trimmed(payData.qr_code) && !trimmed(payData.pay_url)) return false
  return resolvePaymentPresentationMode(payData.interaction_mode) === 'qr'
}

/** 大字金额的裸数值：链上数量 → 实付金额 → 订单金额，不带币种。 */
export const bareAmountText = (o: { chainAmount?: string; payableAmount?: string; amount: string }): string => {
  return trimmed(o.chainAmount) || trimmed(o.payableAmount) || trimmed(o.amount)
}
