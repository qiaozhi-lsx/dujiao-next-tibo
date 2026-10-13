import test from 'node:test'
import assert from 'node:assert/strict'
import { hasActualPayment } from '../src/utils/orderPayment.ts'

test('unpaid statuses never count as actual payment even when intent amounts are present', () => {
  assert.equal(hasActualPayment({ status: 'canceled', paid_at: null, online_paid_amount: '28.90' }), false)
  assert.equal(hasActualPayment({ status: 'pending_payment', paid_at: null }), false)
  assert.equal(hasActualPayment({ status: 'expired', paid_at: null }), false)
})

test('paid_at marks an order as actually paid', () => {
  assert.equal(hasActualPayment({ status: 'delivered', paid_at: '2026-10-14T00:00:00Z' }), true)
  assert.equal(hasActualPayment({ status: 'canceled', paid_at: '2026-10-14T00:00:00Z' }), true)
})

test('paid family statuses count as paid even without paid_at', () => {
  assert.equal(hasActualPayment({ status: 'paid', paid_at: null }), true)
  assert.equal(hasActualPayment({ status: 'completed', paid_at: null }), true)
  assert.equal(hasActualPayment({ status: 'refunded', paid_at: null }), true)
})

test('empty or unknown input is not an actual payment', () => {
  assert.equal(hasActualPayment(null), false)
  assert.equal(hasActualPayment(undefined), false)
  assert.equal(hasActualPayment({}), false)
})
