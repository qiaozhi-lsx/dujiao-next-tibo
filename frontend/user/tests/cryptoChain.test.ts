import test from 'node:test'
import assert from 'node:assert/strict'
import { chainLabel } from '../src/utils/cryptoChain.ts'

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
