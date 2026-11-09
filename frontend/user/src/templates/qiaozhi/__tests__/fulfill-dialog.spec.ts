// 支付后开通弹窗的纯函数用例：订单项判断、步骤构造、轮询节奏、预检错误解析。
// Project uses node:test (not vitest) + no DOM mount.
// Picked up by pnpm run test via the shim at tests/qiaozhiFulfillDialog.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import {
  buildStepViews,
  findFulfillPendingItem,
  isConflictError,
  nextPollDelay,
  normalizeFulfillProgress,
  orderHasFulfillInfoPending,
  orderIsFulfillProcuring,
  parsePrecheckRejection,
  resolvePercent,
  resolveSupportHref,
} from '../utils/fulfillDialog.ts'

const labels = { paid: '支付成功', submit: '提交', process: '处理', complete: '完成' }
const schema = { fields: [{ key: 'session_json', type: 'textarea', required: true }] }

test('订单项：fulfill_info_pending 或「有快照且无提交值」视为待填写', () => {
  assert.equal(orderHasFulfillInfoPending({ items: [{ fulfill_info_pending: true }] }), true)
  assert.equal(orderHasFulfillInfoPending({ items: [{ manual_form_schema_snapshot: schema, manual_form_submission: {} }] }), true)
  assert.equal(orderHasFulfillInfoPending({ items: [{ manual_form_schema_snapshot: schema, manual_form_submission: { session_json: 'x' } }] }), false)
  assert.equal(orderHasFulfillInfoPending({ items: [{ manual_form_schema_snapshot: { fields: [] }, manual_form_submission: {} }] }), false)
  assert.equal(orderHasFulfillInfoPending({ items: [] }), false)
  assert.equal(orderHasFulfillInfoPending(null), false)
  const item = { manual_form_schema_snapshot: schema }
  assert.equal(findFulfillPendingItem({ items: [{}, item] }), item)
})

test('订单项：已提交开通信息且订单在 paid / fulfilling 时为开通中', () => {
  const items = [{ manual_form_schema_snapshot: schema, manual_form_submission: { session_json: 'x' } }]
  assert.equal(orderIsFulfillProcuring({ status: 'paid', items }), true)
  assert.equal(orderIsFulfillProcuring({ status: 'fulfilling', items }), true)
  assert.equal(orderIsFulfillProcuring({ status: 'delivered', items }), false)
  assert.equal(orderIsFulfillProcuring({ status: 'paid', items: [{ manual_form_schema_snapshot: schema, manual_form_submission: {} }] }), false)
})

test('nextPollDelay: 默认 3000，夹在 [2000, 15000]，终态与等待填写不轮询', () => {
  const base = normalizeFulfillProgress({ phase: 'processing' })
  assert.equal(nextPollDelay(base), 3000)
  assert.equal(nextPollDelay({ ...base, poll_after_ms: 500 }), 2000)
  assert.equal(nextPollDelay({ ...base, poll_after_ms: 99999 }), 15000)
  assert.equal(nextPollDelay({ ...base, poll_after_ms: 4500 }), 4500)
  assert.equal(nextPollDelay({ ...base, terminal: true }), null)
  assert.equal(nextPollDelay(normalizeFulfillProgress({ phase: 'completed' })), null)
  assert.equal(nextPollDelay(normalizeFulfillProgress({ phase: 'failed' })), null)
  assert.equal(nextPollDelay(normalizeFulfillProgress({ phase: 'awaiting_info' })), null)
  assert.equal(nextPollDelay(null), null)
})

test('normalizeFulfillProgress: 缺字段补默认值，未知 phase 回落 submitting，percent 夹在 0-100', () => {
  const p = normalizeFulfillProgress({ phase: 'weird', percent: 180, steps: [{ key: 'a', label: 'A' }, { key: '', label: 'x' }] })
  assert.equal(p.phase, 'submitting')
  assert.equal(p.percent, 100)
  assert.deepEqual(p.steps, [{ key: 'a', label: 'A' }])
  assert.equal(p.step, null)
  assert.equal(p.message, '')
})

test('buildStepViews: 上游步骤按 step.index 标记 done / current / pending', () => {
  const progress = normalizeFulfillProgress({
    phase: 'processing',
    steps: [{ key: 'create', label: '提交订单' }, { key: 'charging', label: '扣款确认' }, { key: 'complete', label: '开通完成' }],
    step: { index: 2, total: 3, key: 'charging', label: '扣款确认' },
  })
  const views = buildStepViews(progress, labels)
  assert.deepEqual(views.map((v) => [v.key, v.state]), [
    ['paid', 'done'], ['create', 'done'], ['charging', 'current'], ['complete', 'pending'],
  ])
})

test('buildStepViews: 完成时全部 done；失败时当前步为 failed', () => {
  const steps = [{ key: 'create', label: 'A' }, { key: 'charging', label: 'B' }]
  const done = buildStepViews(normalizeFulfillProgress({ phase: 'completed', steps }), labels)
  assert.ok(done.every((v) => v.state === 'done'))
  const failed = buildStepViews(normalizeFulfillProgress({
    phase: 'failed', steps, step: { index: 2, total: 2, key: 'charging', label: 'B' },
  }), labels)
  assert.deepEqual(failed.map((v) => v.state), ['done', 'done', 'failed'])
})

test('buildStepViews: 上游无步骤时降级为 提交 → 处理 → 完成', () => {
  assert.deepEqual(
    buildStepViews(normalizeFulfillProgress({ phase: 'submitting' }), labels).map((v) => v.state),
    ['done', 'current', 'pending', 'pending'],
  )
  assert.deepEqual(
    buildStepViews(normalizeFulfillProgress({ phase: 'processing' }), labels).map((v) => v.state),
    ['done', 'done', 'current', 'pending'],
  )
  assert.deepEqual(
    buildStepViews(normalizeFulfillProgress({ phase: 'completed' }), labels).map((v) => v.state),
    ['done', 'done', 'done', 'done'],
  )
  assert.equal(buildStepViews(null, labels)[0]?.label, '支付成功')
})

test('resolvePercent: 优先上游 percent，降级按阶段估算', () => {
  assert.equal(resolvePercent(normalizeFulfillProgress({ phase: 'processing', percent: 60 })), 60)
  assert.equal(resolvePercent(normalizeFulfillProgress({ phase: 'processing' })), 50)
  assert.equal(resolvePercent(normalizeFulfillProgress({ phase: 'submitting' })), 20)
  assert.equal(resolvePercent(normalizeFulfillProgress({ phase: 'completed' })), 100)
  assert.equal(resolvePercent(null), 0)
})

test('parsePrecheckRejection: 仅 422 且带 data 才解析；409 单独判断', () => {
  const rejection = parsePrecheckRejection({
    status: 422,
    message: '提交的会话暂不能用于开通',
    data: { reason_code: 'already_subscribed', message: '已订阅', account: { email_masked: 'a***@x.com', plan_type: 'plus', is_active: true } },
  })
  assert.deepEqual(rejection, {
    reasonCode: 'already_subscribed', message: '已订阅', accountEmail: 'a***@x.com', planType: 'plus', isActive: true,
  })
  assert.equal(parsePrecheckRejection({ status: 422, message: 'x', data: { reason_code: 'invalid_session', message: '' } })?.message, 'x')
  assert.equal(parsePrecheckRejection({ status: 500, data: {} }), null)
  assert.equal(parsePrecheckRejection(new Error('boom')), null)
  assert.equal(isConflictError({ status: 409 }), true)
  assert.equal(isConflictError({ status: 422 }), false)
})

test('resolveSupportHref: support_url > telegram > whatsapp > email', () => {
  assert.equal(resolveSupportHref({ support_url: 'https://s', telegram: 't' }), 'https://s')
  assert.equal(resolveSupportHref({ telegram: 'https://t.me/x' }), 'https://t.me/x')
  assert.equal(resolveSupportHref({ whatsapp: '+1 (555) 12' }), 'https://wa.me/155512')
  assert.equal(resolveSupportHref({ email: 'a@b.c' }), 'mailto:a@b.c')
  assert.equal(resolveSupportHref(undefined), '')
})
