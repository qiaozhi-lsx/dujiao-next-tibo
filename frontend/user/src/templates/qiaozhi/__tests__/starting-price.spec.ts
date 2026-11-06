// 精选卡「起」价：多 SKU 商品取启用 SKU 的最低价，无 SKU / 全部停用时回退商品价（C10）。
//
// startingPrice 本体在 utils/startingPrice.ts（纯函数，只依赖 utils/money），由 utils/price.ts
// 统一再导出给组件用；spec 直接 import 纯模块，因为 price.ts 顶层引入 Pinia store，node:test
// 下无法加载。Picked up by pnpm run test via the shim at tests/qiaozhiStartingPrice.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { startingPrice } from '../utils/startingPrice.ts'

test('startingPrice: 多个 SKU 取最低 price_amount，而不是最后一个', () => {
  assert.equal(startingPrice({ price_amount: 58.9, skus: [{ price_amount: 28.9 }, { price_amount: 58.9 }] }), 28.9)
  // 顺序无关：最低价排在后面也要取到
  assert.equal(startingPrice({ price_amount: 50, skus: [{ price_amount: 195 }, { price_amount: 50 }, { price_amount: 99 }] }), 50)
})

test('startingPrice: 没有 SKU 时用商品的 price_amount', () => {
  assert.equal(startingPrice({ price_amount: 12.5 }), 12.5)
  assert.equal(startingPrice({ price_amount: '12.50', skus: [] }), '12.50')
  assert.equal(startingPrice({ price_amount: 9, skus: null }), 9)
})

test('startingPrice: 全部 SKU 停用时用商品的 price_amount', () => {
  const product = { price_amount: 88, skus: [{ price_amount: 28.9, is_active: false }, { price_amount: 58.9, is_active: false }] }
  assert.equal(startingPrice(product), 88)
})

test('startingPrice: 停用的 SKU 不参与比较，未带 is_active 的 SKU 视为启用', () => {
  const product = { price_amount: 1, skus: [{ price_amount: 10, is_active: false }, { price_amount: 20, is_active: true }, { price_amount: 30 }] }
  assert.equal(startingPrice(product), 20)
})

test('startingPrice: 金额按分比较（字符串金额也能比），非法金额的 SKU 跳过', () => {
  assert.equal(startingPrice({ price_amount: '0.00', skus: [{ price_amount: '195.00' }, { price_amount: '50.00' }] }), '50.00')
  assert.equal(startingPrice({ price_amount: '9.99', skus: [{ price_amount: 'n/a' }, { price_amount: '100.00' }] }), '100.00')
  assert.equal(startingPrice({ price_amount: '9.99', skus: [{ price_amount: null }, { price_amount: undefined }] }), '9.99')
})

test('startingPrice: 空输入原样回退', () => {
  assert.equal(startingPrice(null), undefined)
  assert.equal(startingPrice(undefined), undefined)
  assert.equal(startingPrice({}), undefined)
})
