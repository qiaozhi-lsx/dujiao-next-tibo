import test from 'node:test'
import assert from 'node:assert/strict'
import { rankSearchItems } from '../src/templates/qiaozhi/utils/searchRank.ts'

const items = [
  { id: 1, title: 'Standard plan' },
  { id: 2, title: 'Premium plan' },
  { id: 3, title: 'plan Pro' },
  { id: 4, title: 'X Premium', keywords: ['plan'] },
  { id: 5, title: 'Other', meta: 'plan guide' },
  { id: 6, title: 'Unrelated' },
]

test('title starts beat word starts, which beat inside-word matches, which beat keyword and meta matches', () => {
  const result = rankSearchItems(items, 'plan', 10).map((item) => item.id)
  assert.deepEqual(result, [3, 1, 2, 4, 5])
})

test('non-matching items are dropped and ties keep the original order', () => {
  const result = rankSearchItems(items, 'premium', 10).map((item) => item.id)
  assert.deepEqual(result, [2, 4])
})

test('limit truncates the ranked list and empty queries return nothing', () => {
  assert.deepEqual(rankSearchItems(items, 'plan', 2).map((item) => item.id), [3, 1])
  assert.deepEqual(rankSearchItems(items, '   ', 10), [])
})

test('matching is case-insensitive', () => {
  assert.deepEqual(rankSearchItems(items, 'PREMIUM', 10).map((item) => item.id), [2, 4])
})
