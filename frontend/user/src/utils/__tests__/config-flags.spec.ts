// Verifies optionalFlagOn, the helper that gates "default-on" optional feature
// flags (nav blog / notice / about, email_verification_enabled) behind
// appStore.configComplete. The server-inlined window.__PUBLIC_CONFIG__ is only a
// subset of /public/config and lacks those keys, so reading `value !== false` on
// it yields true for one RTT and then flips to false once the full config lands:
// the blog link, forgot-password link and register code field flash in and vanish.
// The flag must therefore stay off until a full config has been fetched.
//
// Project uses node:test (not vitest) + no DOM.
// Picked up by pnpm run test via the shim at tests/configFlags.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { optionalFlagOn } from '../configFlags.ts'

test('optionalFlagOn: incomplete config keeps a missing flag off', () => {
  assert.equal(optionalFlagOn(false, undefined), false)
})

test('optionalFlagOn: incomplete config keeps even an explicit true off', () => {
  assert.equal(optionalFlagOn(false, true), false)
})

test('optionalFlagOn: complete config treats a missing flag as on', () => {
  assert.equal(optionalFlagOn(true, undefined), true)
})

test('optionalFlagOn: complete config honours an explicit false', () => {
  assert.equal(optionalFlagOn(true, false), false)
})

test('optionalFlagOn: complete config honours an explicit true', () => {
  assert.equal(optionalFlagOn(true, true), true)
})
