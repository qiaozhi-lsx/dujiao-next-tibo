// 商品页内嵌下单的请求体构造（分销归因 / 渠道 / 游客字段）与手动表单商品判断的纯函数用例。
// Project uses node:test (not vitest) + no DOM mount.
// Picked up by pnpm run test via the shim at tests/qiaozhiOrderPayload.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { buildCreateAndPayPayload, hasManualFormFields, isFulfillAfterPayment } from '../utils/orderPayload.ts'

test('buildCreateAndPayPayload: 登录用户只带 items / channel_id / 分销归因，不带游客字段', () => {
  const payload = buildCreateAndPayPayload({
    items: [{ product_id: 1 }],
    channelId: 3,
    affiliateCode: 'AFF',
    affiliateVisitorKey: 'vk',
  })
  assert.deepEqual(payload, {
    items: [{ product_id: 1 }],
    channel_id: 3,
    affiliate_code: 'AFF',
    affiliate_visitor_key: 'vk',
  })
  assert.equal('email' in payload, false)
  assert.equal('order_password' in payload, false)
  assert.equal('captcha_payload' in payload, false)
})

test('buildCreateAndPayPayload: 无归因 / 未选渠道时字段为 undefined，JSON 序列化后不出现', () => {
  const payload = buildCreateAndPayPayload({
    items: [{ product_id: 1 }],
    channelId: null,
    affiliateCode: '',
    affiliateVisitorKey: '',
  })
  assert.equal(payload.affiliate_code, undefined)
  assert.equal(payload.affiliate_visitor_key, undefined)
  assert.equal(payload.channel_id, undefined)
  const json = JSON.stringify(payload)
  assert.equal(json.includes('affiliate_code'), false)
  assert.equal(json.includes('affiliate_visitor_key'), false)
  assert.equal(json.includes('channel_id'), false)
  assert.deepEqual(JSON.parse(json), { items: [{ product_id: 1 }] })
})

test('buildCreateAndPayPayload: 游客下单附带 email / order_password / captcha_payload', () => {
  const payload = buildCreateAndPayPayload({
    items: [{ product_id: 1, sku_id: 2, quantity: 1 }],
    channelId: 3,
    affiliateCode: 'AFF',
    affiliateVisitorKey: 'vk',
    guest: {
      email: 'guest@example.com',
      orderPassword: 'secret-1',
      captchaPayload: { turnstile_token: 'tok' },
    },
  })
  assert.deepEqual(payload, {
    items: [{ product_id: 1, sku_id: 2, quantity: 1 }],
    channel_id: 3,
    affiliate_code: 'AFF',
    affiliate_visitor_key: 'vk',
    email: 'guest@example.com',
    order_password: 'secret-1',
    captcha_payload: { turnstile_token: 'tok' },
  })
})

test('hasManualFormFields: schema 含至少一个字段时为 true', () => {
  assert.equal(hasManualFormFields({ fields: [{ key: 'a', type: 'text' }] }), true)
})

test('hasManualFormFields: 空字段 / 缺失 / 空对象 / null 均为 false', () => {
  assert.equal(hasManualFormFields({ fields: [] }), false)
  assert.equal(hasManualFormFields(undefined), false)
  assert.equal(hasManualFormFields({}), false)
  assert.equal(hasManualFormFields(null), false)
})

test('buildCreateAndPayPayload: deferManualForm 为 true 时带 defer_manual_form，否则不出现', () => {
  const deferred = buildCreateAndPayPayload({
    items: [{ product_id: 1 }], channelId: 3, affiliateCode: '', affiliateVisitorKey: '', deferManualForm: true,
  })
  assert.equal(deferred.defer_manual_form, true)
  const normal = buildCreateAndPayPayload({
    items: [{ product_id: 1 }], channelId: 3, affiliateCode: '', affiliateVisitorKey: '',
  })
  assert.equal('defer_manual_form' in normal, false)
})

test('isFulfillAfterPayment: 仅 fulfill_after_payment === true 的商品为 true', () => {
  assert.equal(isFulfillAfterPayment({ fulfill_after_payment: true }), true)
  assert.equal(isFulfillAfterPayment({ fulfill_after_payment: false }), false)
  assert.equal(isFulfillAfterPayment({ manual_form_schema: { fields: [{ key: 'a' }] } }), false)
  assert.equal(isFulfillAfterPayment(null), false)
  assert.equal(isFulfillAfterPayment(undefined), false)
})
