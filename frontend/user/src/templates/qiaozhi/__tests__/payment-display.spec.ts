// 支付弹窗展示判断（币种 / 手续费 / 已付与终态 / 弹窗分流 / 裸金额）的纯函数用例。
// Project uses node:test (not vitest) + no DOM mount.
// Picked up by pnpm run test via the shim at tests/qiaozhiPaymentDisplay.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import {
  PAID_ORDER_STATUSES,
  TERMINAL_ORDER_STATUSES,
  bareAmountText,
  isPaidOrderStatus,
  isTerminalOrderStatus,
  resolvePayableCurrency,
  shouldOpenInlineModal,
  shouldShowFee,
} from '../utils/paymentDisplay.ts'

test('resolvePayableCurrency: payableCurrency 优先，回退 currency，都没有时为空串', () => {
  assert.equal(resolvePayableCurrency({ currency: 'USD', payableCurrency: 'CNY' }), 'CNY')
  assert.equal(resolvePayableCurrency({ currency: 'USD' }), 'USD')
  assert.equal(resolvePayableCurrency({}), '')
})

test('shouldShowFee: 仅客户承担手续费且金额非零时展示', () => {
  assert.equal(shouldShowFee({ feeAmount: '3.00', feePolicy: 'merchant_absorbed' }), false)
  assert.equal(shouldShowFee({ feeAmount: '3.00', feePolicy: 'customer_surcharge' }), true)
  assert.equal(shouldShowFee({ feeAmount: '3.00', feePolicy: 'legacy_customer_surcharge' }), true)
  assert.equal(shouldShowFee({ feeAmount: '0.00', feePolicy: 'customer_surcharge' }), false)
  assert.equal(shouldShowFee({ feeAmount: '3.00' }), false)
})

test('isPaidOrderStatus / isTerminalOrderStatus: 已付与终态判断', () => {
  assert.equal(isPaidOrderStatus('delivered'), true)
  assert.equal(isPaidOrderStatus('pending'), false)
  assert.equal(isTerminalOrderStatus('canceled'), true)
  assert.equal(isTerminalOrderStatus('expired'), true)
  assert.equal(isTerminalOrderStatus('pending'), false)
})

test('PAID_ORDER_STATUSES / TERMINAL_ORDER_STATUSES: 导出集合内容与约定一致', () => {
  assert.deepEqual([...PAID_ORDER_STATUSES].sort(), ['completed', 'delivered', 'fulfilling', 'paid', 'partially_delivered'])
  assert.deepEqual([...TERMINAL_ORDER_STATUSES].sort(), ['canceled', 'cancelled', 'expired', 'refunded'])
})

test('shouldOpenInlineModal: 仅 qr 模式且有支付载荷、未已付时走内嵌弹窗', () => {
  assert.equal(shouldOpenInlineModal({ qr_code: 'x', interaction_mode: 'qr' }), true)
  assert.equal(shouldOpenInlineModal({ pay_url: 'https://p', interaction_mode: 'redirect' }), false)
  assert.equal(shouldOpenInlineModal({ pay_url: 'https://p', interaction_mode: 'wap' }), false)
  assert.equal(shouldOpenInlineModal({ pay_url: 'https://p', interaction_mode: 'page' }), false)
  assert.equal(shouldOpenInlineModal({ pay_url: 'https://p' }), true)
  assert.equal(shouldOpenInlineModal({ qr_code: 'x', interaction_mode: 'qr', order_paid: true }), false)
  assert.equal(shouldOpenInlineModal(null), false)
})

test('bareAmountText: chainAmount → payableAmount → amount，不带币种', () => {
  assert.equal(bareAmountText({ chainAmount: '13.45', payableAmount: '97.00', amount: '100.00' }), '13.45')
  assert.equal(bareAmountText({ payableAmount: '97.00', amount: '100.00' }), '97.00')
  assert.equal(bareAmountText({ amount: '100.00' }), '100.00')
})
