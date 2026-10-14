import test from 'node:test'
import assert from 'node:assert/strict'
import { countdownTone } from '../src/utils/countdownTone.ts'

test('阈值边界精确落档(>300 绿 / ≤300 橙 / ≤180 深橙 / ≤60 红)', () => {
  assert.equal(countdownTone(601), 'safe')
  assert.equal(countdownTone(301), 'safe')
  assert.equal(countdownTone(300), 'warn')
  assert.equal(countdownTone(181), 'warn')
  assert.equal(countdownTone(180), 'late')
  assert.equal(countdownTone(61), 'late')
  assert.equal(countdownTone(60), 'critical')
  assert.equal(countdownTone(0), 'critical')
  assert.equal(countdownTone(-5), 'critical')
})
