import test from 'node:test'
import assert from 'node:assert/strict'
import { chainLabel, channelDisplayName, hasUsdtToken, usdtDisplayLabel } from '../src/utils/cryptoChain.ts'

test('channelDisplayName: 可规范化的渠道名统一成「USDT · 链名」，其余原样返回', () => {
  assert.equal(channelDisplayName('USDT-BEP20'), 'USDT · BSC')
  assert.equal(channelDisplayName('USDT-SOL'), 'USDT · Solana')
  assert.equal(channelDisplayName('USDT-TRC20（推荐）'), 'USDT-TRC20（推荐）')
  assert.equal(channelDisplayName('支付宝'), '支付宝')
  assert.equal(channelDisplayName(''), '')
})

test('BEP20 链文案与网关对齐为 BSC', () => {
  assert.equal(chainLabel('binance'), 'BSC')
  assert.equal(chainLabel('bsc'), 'BSC')
  assert.equal(chainLabel('BSC'), 'BSC')
})

test('tron / solana 文案与网关一致', () => {
  assert.equal(chainLabel('tron'), 'TRON')
  assert.equal(chainLabel('trc20'), 'TRON')
  assert.equal(chainLabel('solana'), 'Solana')
})

test('未知或空链名原样返回,不抛错', () => {
  assert.equal(chainLabel(''), '')
  assert.equal(chainLabel('newchain'), 'newchain')
})

test('usdtDisplayLabel: 渠道名只有币种 + 单一链时改写为「USDT · 链名」', () => {
  assert.equal(usdtDisplayLabel('USDT-TRC20'), 'USDT · TRON')
  assert.equal(usdtDisplayLabel('usdt_trc20'), 'USDT · TRON')
  assert.equal(usdtDisplayLabel('USDT (BEP20)'), 'USDT · BSC')
  assert.equal(usdtDisplayLabel('TRC20 USDT'), 'USDT · TRON')
  assert.equal(usdtDisplayLabel('USDT · TRON'), 'USDT · TRON')
  assert.equal(usdtDisplayLabel('USDT（ERC20）'), 'USDT · Ethereum')
  assert.equal(usdtDisplayLabel('USDT/SOL'), 'USDT · Solana')
  assert.equal(usdtDisplayLabel('[USDT] Polygon'), 'USDT · Polygon')
})

test('usdtDisplayLabel: 带附加文字或无链的渠道名不改写,返回空串', () => {
  assert.equal(usdtDisplayLabel('泰达币'), '')
  assert.equal(usdtDisplayLabel('USDT-TRC20（推荐）'), '')
  assert.equal(usdtDisplayLabel('USDT'), '')
  assert.equal(usdtDisplayLabel('USDT BEP20 快速到账'), '')
  assert.equal(usdtDisplayLabel('TRON 支付'), '')
  assert.equal(usdtDisplayLabel('Alipay'), '')
  assert.equal(usdtDisplayLabel(''), '')
})

test('usdtDisplayLabel: 同时写了多条链的渠道名不改写', () => {
  assert.equal(usdtDisplayLabel('USDT (TRC20/BEP20/ERC20)'), '')
  assert.equal(usdtDisplayLabel('USDT TRC20 BEP20'), '')
})

test('usdtDisplayLabel: provider 名 epusdt / bepusdt 与 USDT0 不算 USDT', () => {
  assert.equal(usdtDisplayLabel('epusdt TRC20'), '')
  assert.equal(usdtDisplayLabel('bepusdt-trc20'), '')
  assert.equal(usdtDisplayLabel('USDT0 Polygon'), '')
})

test('hasUsdtToken: 渠道名 / channel_type 里独立成词的 usdt 才算(徽标用)', () => {
  assert.equal(hasUsdtToken('USDT'), true)
  assert.equal(hasUsdtToken('泰达币USDT'), true)
  assert.equal(hasUsdtToken('usdt-trc20'), true)
  assert.equal(hasUsdtToken('tron-usdt'), true)
  assert.equal(hasUsdtToken('epusdt'), false)
  assert.equal(hasUsdtToken('bepusdt'), false)
  assert.equal(hasUsdtToken('polygon-usdt0'), false)
  assert.equal(hasUsdtToken(''), false)
  // 全局正则需要可重复调用
  assert.equal(hasUsdtToken('USDT'), true)
  assert.equal(hasUsdtToken('USDT'), true)
})
