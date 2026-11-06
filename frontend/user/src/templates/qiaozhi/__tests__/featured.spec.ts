// 首页「为你精选」的选品与精选卡艺术区变体（P6）。
//
// pickFeatured：打了 featured / 精选 标签的商品按列表原顺序取前 5；一个都没打时沿用旧的 slug 品牌正则
// （现网行为不变），正则也没命中时取前 4。artVariant：tags 里的 art:<variant> 优先，否则沿用 slug 正则。
// displayTags：art:* 只给程序看，不进卡片上的标签 / 芯片。
// 纯函数模块，不依赖 Vue / Pinia，node:test 直接加载。Picked up by pnpm run test via the shim at
// tests/qiaozhiFeatured.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { FEATURED_PATTERN, artVariant, displayTags, pickFeatured } from '../utils/featured.ts'

const slugs = (list: { slug?: unknown }[]) => list.map((p) => p.slug)

test('pickFeatured: 没有 featured 标签时沿用 slug 正则（现网行为）', () => {
  assert.deepEqual(slugs(pickFeatured([{ slug: 'x-premium', tags: [] }])), ['x-premium'])
  const list = [
    { slug: 'netflix', tags: [] },
    { slug: 'chatgpt-plus', tags: [] },
    { slug: 'spotify', tags: [] },
    { slug: 'gemini-advanced', tags: [] },
    { slug: 'duolingo-super', tags: [] },
    { slug: 'muse-ai', tags: [] },
    { slug: 'x-premium', tags: [] },
    { slug: 'google-one', tags: [] },
  ]
  // 命中的按原顺序取前 5，未命中的（netflix / spotify）不进精选
  assert.deepEqual(slugs(pickFeatured(list)), ['chatgpt-plus', 'gemini-advanced', 'duolingo-super', 'muse-ai', 'x-premium'])
  // 正则本身保持与 Home.vue 旧常量一致
  assert.equal(String(FEATURED_PATTERN), String(/chatgpt|gpt|gemini|bard|google|^x-|x-premium|twitter|duolingo|duo|muse/i))
})

test('pickFeatured: 有 featured 标签时只取打标商品，正则命中的不再自动入选', () => {
  const list = [
    { slug: 'abc', tags: ['featured'] },
    { slug: 'x-premium', tags: [] },
  ]
  assert.deepEqual(slugs(pickFeatured(list)), ['abc'])
})

test('pickFeatured: 标签支持「精选」与任意大小写 / 两端空白，按原顺序最多 5 个', () => {
  const list = [
    { slug: 'p1', tags: ['精选'] },
    { slug: 'p2', tags: ['热卖', ' Featured '] },
    { slug: 'p3', tags: [] },
    { slug: 'p4', tags: ['FEATURED'] },
    { slug: 'p5', tags: ['featured'] },
    { slug: 'p6', tags: ['featured'] },
    { slug: 'p7', tags: ['featured'] },
    { slug: 'p8', tags: ['featured'] },
  ]
  assert.deepEqual(slugs(pickFeatured(list)), ['p1', 'p2', 'p4', 'p5', 'p6'])
})

test('pickFeatured: 无标签且正则无命中时取前 4', () => {
  const list = [
    { slug: 'netflix', tags: [] },
    { slug: 'spotify', tags: [] },
    { slug: 'office', tags: null },
    { slug: 'canva' },
    { slug: 'notion', tags: ['热卖'] },
    { slug: 'figma', tags: [] },
  ]
  assert.deepEqual(slugs(pickFeatured(list)), ['netflix', 'spotify', 'office', 'canva'])
  assert.deepEqual(pickFeatured([]), [])
})

test('pickFeatured: 返回原对象引用，调用方按 id 继续拆通栏卡', () => {
  const a = { id: 1, slug: 'abc', tags: ['featured'] }
  const b = { id: 2, slug: 'def', tags: [] }
  const picked = pickFeatured([a, b])
  assert.equal(picked.length, 1)
  assert.equal(picked[0], a)
})

test('artVariant: art:<variant> 标签优先于 slug', () => {
  assert.equal(artVariant('foo', ['art:gemini']), 'gemini')
  assert.equal(artVariant('chatgpt-plus', ['热卖', 'art:x']), 'x')
  // 大小写与空白不敏感
  assert.equal(artVariant('foo', [' ART:Duo ']), 'duo')
  assert.equal(artVariant('foo', ['art:muse']), 'muse')
  assert.equal(artVariant('foo', ['art:chatgpt']), 'chatgpt')
  // 显式回落到默认插画（slug 本会命中 gpt）
  assert.equal(artVariant('chatgpt-plus', ['art:default']), 'default')
})

test('artVariant: 没有 art 标签（或标签不认识）时沿用 slug 正则', () => {
  assert.equal(artVariant('chatgpt-plus', []), 'chatgpt')
  assert.equal(artVariant('gpt-4-team', []), 'chatgpt')
  assert.equal(artVariant('gemini-advanced', []), 'gemini')
  assert.equal(artVariant('google-one', []), 'gemini')
  assert.equal(artVariant('x-premium', []), 'x')
  assert.equal(artVariant('twitter-blue', []), 'x')
  assert.equal(artVariant('duolingo-super', []), 'duo')
  assert.equal(artVariant('muse-ai', []), 'muse')
  assert.equal(artVariant('netflix', []), 'default')
  // 不认识的变体名忽略；tools 只能由通栏卡 prop 指定，不接受标签
  assert.equal(artVariant('x-premium', ['art:unknown']), 'x')
  assert.equal(artVariant('netflix', ['art:tools']), 'default')
  // slug 大小写不敏感，tags 缺省
  assert.equal(artVariant('ChatGPT-Plus'), 'chatgpt')
  // 正则分支的优先级锁定（与旧 QzFeaturedCard 完全一致）：gpt 先于 google，gemini 先于 ^x-
  assert.equal(artVariant('google-gpt', []), 'chatgpt')
  assert.equal(artVariant('x-gemini', []), 'gemini')
})

test('displayTags: 去掉 art:* 控制标签，其余按原顺序保留', () => {
  assert.deepEqual(displayTags(['art:gemini', '热卖', 'featured', 'ART:x', '限时']), ['热卖', 'featured', '限时'])
  assert.deepEqual(displayTags(['  ', '', 'ok']), ['ok'])
  assert.deepEqual(displayTags(undefined), [])
  assert.deepEqual(displayTags('featured'), [])
})
