// 首页「为你精选」的选品与精选卡艺术区变体。纯函数、零依赖：node:test 直接加载（见 __tests__/featured.spec.ts）。

/** 精选卡艺术区变体。`tools` 只由 Home 的通栏卡通过 prop 指定，不参与标签 / slug 判定。 */
export type ArtVariant = 'chatgpt' | 'gemini' | 'x' | 'duo' | 'muse' | 'tools' | 'default'

/** 精选判定只读这两个字段；列表接口返回的完整商品对象可以直接传入，返回的仍是原对象。 */
export interface FeaturedProduct {
  slug?: unknown
  tags?: unknown
}

/** 旧的精选判定：slug 命中这些品牌关键词。没有任何商品打 featured 标签时沿用，现网展示保持不变。 */
export const FEATURED_PATTERN = /chatgpt|gpt|gemini|bard|google|^x-|x-premium|twitter|duolingo|duo|muse/i

/** 后台给商品打上其中之一即进入精选（大小写、两端空白不敏感）。 */
const FEATURED_TAGS = new Set(['featured', '精选'])
/** `art:<variant>` 指定插画；未列出的名字忽略，回落到 slug 正则。 */
const ART_TAG_PREFIX = 'art:'
const ART_TAG_VARIANTS = new Set<ArtVariant>(['chatgpt', 'gemini', 'x', 'duo', 'muse', 'default'])

const FEATURED_LIMIT = 5
const FALLBACK_LIMIT = 4

const normalizeTags = (tags: unknown): string[] =>
  Array.isArray(tags) ? tags.map((tag) => String(tag ?? '').trim()).filter(Boolean) : []

const isFeaturedTag = (tag: string): boolean => FEATURED_TAGS.has(tag.toLowerCase())
const isArtTag = (tag: string): boolean => tag.toLowerCase().startsWith(ART_TAG_PREFIX)

/**
 * 首页精选：打了 featured / 精选 标签的商品按列表原顺序取前 5；一个都没打时按 FEATURED_PATTERN 正则
 * 取前 5；正则也没命中时取前 4 做默认卡。
 */
export const pickFeatured = <T extends FeaturedProduct>(products: readonly T[]): T[] => {
  const tagged = products.filter((product) => normalizeTags(product?.tags).some(isFeaturedTag))
  if (tagged.length) return tagged.slice(0, FEATURED_LIMIT)
  const matched = products.filter((product) => FEATURED_PATTERN.test(String(product?.slug || '')))
  if (matched.length) return matched.slice(0, FEATURED_LIMIT)
  return products.slice(0, FALLBACK_LIMIT)
}

/** 艺术区变体：tags 里的 `art:<variant>` 优先（取第一个认识的），否则按 slug 关键词映射。 */
export const artVariant = (slug: string, tags: readonly unknown[] = []): ArtVariant => {
  for (const tag of normalizeTags(tags)) {
    if (!isArtTag(tag)) continue
    const name = tag.slice(ART_TAG_PREFIX.length).trim().toLowerCase() as ArtVariant
    if (ART_TAG_VARIANTS.has(name)) return name
  }
  const s = String(slug || '').toLowerCase()
  if (/chatgpt|gpt/.test(s)) return 'chatgpt'
  if (/gemini|bard|google/.test(s)) return 'gemini'
  if (/^x-|x-premium|twitter|tweet/.test(s)) return 'x'
  if (/duolingo|duo|lingo/.test(s)) return 'duo'
  if (/muse/.test(s)) return 'muse'
  return 'default'
}

/** 卡片上展示的标签：去掉 `art:*` 这类只给程序看的控制标签，其余按原顺序保留。 */
export const displayTags = (tags: unknown): string[] => normalizeTags(tags).filter((tag) => !isArtTag(tag))
