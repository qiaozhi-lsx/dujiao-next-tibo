// Verifies estimateServerNow, the pure helper stores/app.ts uses to derive the
// server clock from a /public/config response. The endpoint is served from the
// CDN edge for up to 10s (Cache-Control: public, max-age=0, s-maxage=10), so the
// body's server_time can be up to `Age` seconds stale; the helper adds the Age
// header back plus half the measured round trip.
//
// Project uses node:test (not vitest) + no DOM.
// Picked up by pnpm run test via the shim at tests/serverTime.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { estimateServerNow } from '../serverTime.ts'

test('estimateServerNow: adds Age seconds and half the round trip', () => {
  assert.equal(
    estimateServerNow({ serverTime: 1000, ageHeader: '7', requestTime: 0, responseTime: 100 }),
    8050,
  )
})

test('estimateServerNow: missing Age header counts as zero', () => {
  assert.equal(estimateServerNow({ serverTime: 1000, requestTime: 0, responseTime: 100 }), 1050)
  // fetch Headers.get() returns null for an absent header
  assert.equal(
    estimateServerNow({ serverTime: 1000, ageHeader: null, requestTime: 0, responseTime: 100 }),
    1050,
  )
})

test('estimateServerNow: missing server_time yields null', () => {
  assert.equal(estimateServerNow({ requestTime: 0, responseTime: 100 }), null)
  assert.equal(estimateServerNow({ ageHeader: '7', requestTime: 0, responseTime: 100 }), null)
})

test('estimateServerNow: malformed or negative Age header is ignored', () => {
  assert.equal(
    estimateServerNow({ serverTime: 1000, ageHeader: 'garbage', requestTime: 0, responseTime: 100 }),
    1050,
  )
  assert.equal(
    estimateServerNow({ serverTime: 1000, ageHeader: '-5', requestTime: 0, responseTime: 100 }),
    1050,
  )
})
