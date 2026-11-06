// navSection(path): which primary-nav section a route path belongs to. QiaozhiLayout.vue
// uses one `section` computed for the header pills and the mobile bottom nav, so both
// highlight the same item (the bottom nav used to invert the orders check and never
// highlighted the cart). `/me/orders*` belongs to orders, the rest of `/me/*` to account.
// Project uses node:test (no DOM, no vitest); picked up by pnpm run test via the shim at
// tests/qiaozhiNavSection.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { navSection } from '../utils/navSection.ts'

test('navSection: storefront routes resolve to shop', () => {
  assert.equal(navSection('/'), 'shop')
  assert.equal(navSection('/products'), 'shop')
  assert.equal(navSection('/products/x-premium'), 'shop')
  assert.equal(navSection('/categories/3'), 'shop')
})

test('navSection: guest and member order routes resolve to orders', () => {
  assert.equal(navSection('/guest/orders'), 'orders')
  assert.equal(navSection('/guest/orders/NO123'), 'orders')
  assert.equal(navSection('/me/orders'), 'orders')
  assert.equal(navSection('/me/orders/NO1'), 'orders')
})

test('navSection: cart and checkout resolve to cart', () => {
  assert.equal(navSection('/cart'), 'cart')
  assert.equal(navSection('/checkout'), 'cart')
})

test('navSection: notice / 2fa / blog / about', () => {
  assert.equal(navSection('/notice'), 'notice')
  assert.equal(navSection('/2fa'), 'tfa')
  assert.equal(navSection('/blog/abc'), 'blog')
  assert.equal(navSection('/about'), 'about')
})

test('navSection: personal centre and auth resolve to account (except /me/orders)', () => {
  assert.equal(navSection('/me'), 'account')
  assert.equal(navSection('/me/wallet'), 'account')
  assert.equal(navSection('/auth/login'), 'account')
  assert.equal(navSection('/auth/register'), 'account')
})

test('navSection: unrelated routes resolve to null', () => {
  assert.equal(navSection('/terms'), null)
  assert.equal(navSection('/whatever'), null)
})

test('navSection: prefix matching is segment-aware and tolerates a trailing slash', () => {
  assert.equal(navSection('/cart/'), 'cart')
  assert.equal(navSection('/me/'), 'account')
  assert.equal(navSection('/cartoon'), null)
  assert.equal(navSection('/products-and-more'), null)
})
