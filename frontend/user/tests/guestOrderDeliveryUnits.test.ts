import test from 'node:test'
import assert from 'node:assert/strict'
import { buildDeliveryUnits } from '../src/templates/qiaozhi/utils/deliveryUnits.ts'

const resolveChildStatus = (child: any) => (child?.status ?? null)

const xPremiumItems = () => ([{ title: { 'zh-CN': 'X Premium' }, quantity: 1 }])

test('orders with child orders render the child units without duplicating aggregated main items', () => {
  const order = {
    order_no: 'DJ-1',
    status: 'canceled',
    currency: 'CNY',
    total_amount: '28.90',
    items: xPremiumItems(),
    children: [{
      order_no: 'DJ-1-01',
      status: 'canceled',
      total_amount: '28.90',
      currency: 'CNY',
      items: xPremiumItems(),
    }],
  }
  const units = buildDeliveryUnits(order, resolveChildStatus)
  assert.equal(units.length, 1)
  assert.equal(units[0].isChild, true)
  assert.equal(units[0].orderNo, 'DJ-1-01')
  assert.equal(units[0].items.length, 1)
  assert.equal(units[0].status, 'canceled')
})

test('orders without child orders render a single main unit with items and fulfillment', () => {
  const order = {
    order_no: 'DJ-2',
    status: 'delivered',
    currency: 'CNY',
    items: xPremiumItems(),
    fulfillment: { status: 'delivered', payload: 'card-123' },
  }
  const units = buildDeliveryUnits(order, resolveChildStatus)
  assert.equal(units.length, 1)
  assert.equal(units[0].isChild, false)
  assert.equal(units[0].items.length, 1)
  assert.ok(units[0].fulfillment)
})

test('a main order fulfillment alongside child orders adds a fulfillment-only unit without items', () => {
  const order = {
    order_no: 'DJ-3',
    status: 'delivered',
    currency: 'CNY',
    items: xPremiumItems(),
    fulfillment: { status: 'delivered', payload: 'card-456' },
    children: [{
      order_no: 'DJ-3-01',
      status: 'delivered',
      total_amount: '1.00',
      currency: 'CNY',
      items: xPremiumItems(),
    }],
  }
  const units = buildDeliveryUnits(order, resolveChildStatus)
  assert.equal(units.length, 2)
  assert.equal(units[0].isChild, true)
  assert.equal(units[1].isChild, false)
  assert.equal(units[1].items.length, 0)
  assert.ok(units[1].fulfillment)
})

test('child orders without items fall back to the aggregated main items', () => {
  const order = {
    order_no: 'DJ-4',
    status: 'canceled',
    currency: 'CNY',
    items: xPremiumItems(),
    children: [{
      order_no: 'DJ-4-01',
      status: 'canceled',
      total_amount: '1.00',
      currency: 'CNY',
      items: [],
    }],
  }
  const units = buildDeliveryUnits(order, resolveChildStatus)
  assert.equal(units.length, 2)
  assert.equal(units[1].items.length, 1)
})

test('empty input yields no delivery units', () => {
  assert.deepEqual(buildDeliveryUnits(null, resolveChildStatus), [])
  assert.deepEqual(buildDeliveryUnits(undefined, resolveChildStatus), [])
})
