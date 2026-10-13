/**
 * 展开式搜索的结果排序（参考 uiarc.dev ExpandingSearch 的 rank 规则）。
 * 标题开头 > 词首（前面是空格/横杠/斜杠）> 词中 > 关键词/副文案命中；同分保持原始顺序。
 */
export interface SearchableItem {
  id: string | number
  title: string
  meta?: string
  keywords?: string[]
  [key: string]: unknown
}

export const rankSearchItems = <T extends SearchableItem>(items: T[], query: string, limit: number): T[] => {
  const q = query.trim().toLocaleLowerCase()
  if (!q) return []
  const scored: { item: T; score: number; order: number }[] = []
  items.forEach((item, order) => {
    const title = String(item.title || '').toLocaleLowerCase()
    const at = title.indexOf(q)
    const extra = [item.meta ?? '', ...(item.keywords ?? [])].some((word) => String(word).toLocaleLowerCase().includes(q))
    const prev = at > 0 ? title[at - 1] ?? '' : ''
    const score = at === 0 ? 0 : at > 0 && /[\s\-/]/.test(prev) ? 1 : at > 0 ? 2 : extra ? 3 : -1
    if (score >= 0) scored.push({ item, score, order })
  })
  return scored
    .sort((a, b) => a.score - b.score || a.order - b.order)
    .slice(0, Math.max(1, limit))
    .map((entry) => entry.item)
}
