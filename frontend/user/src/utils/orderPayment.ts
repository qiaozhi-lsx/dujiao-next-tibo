/**
 * 订单实付判定（订单详情展示用）。
 *
 * 后端在「创建支付」时就会把 online_paid_amount / wallet_paid_amount 写入订单
 * （此时只是支付意图金额，钱还没到账），所以这两个字段不能直接当作实付依据：
 * 已取消 / 待支付 / 已过期的订单也可能带着非零金额。以 paid_at 为准，并按
 * 「未支付家族」状态兜底，未支付时不展示任何实付金额。
 */
const UNPAID_STATUSES = ['pending_payment', 'canceled', 'expired']

export const hasActualPayment = (order: any): boolean => {
  if (!order) return false
  if (order.paid_at) return true
  const status = String(order.status || '').trim()
  return Boolean(status) && !UNPAID_STATUSES.includes(status)
}
