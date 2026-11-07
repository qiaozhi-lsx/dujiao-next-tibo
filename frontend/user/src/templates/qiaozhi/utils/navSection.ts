/**
 * 路由路径 → 主导航分区。顶栏胶囊和手机端底部导航共用同一个结果，保证两处高亮一致。
 *
 * 前缀表按"更具体的在前"排序：`/me/orders` 归查订单，必须先于 `/me` → 个人中心判断。
 * 匹配按路径段进行（`/cart` 命中 `/cart` 与 `/cart/...`，不命中 `/cartoon`）。
 */
export type NavSection = 'shop' | 'orders' | 'cart' | 'notice' | 'tfa' | 'account' | 'blog' | 'about' | null

const PREFIXES: ReadonlyArray<readonly [prefix: string, section: NonNullable<NavSection>]> = [
  ['/me/orders', 'orders'],
  ['/guest/orders', 'orders'],
  ['/orders', 'orders'],
  ['/products', 'shop'],
  ['/categories', 'shop'],
  ['/cart', 'cart'],
  ['/checkout', 'cart'],
  ['/notice', 'notice'],
  ['/2fa', 'tfa'],
  ['/blog', 'blog'],
  ['/about', 'about'],
  ['/me', 'account'],
  ['/auth', 'account'],
  ['/recharge-orders', 'account'],
]

const normalize = (path: string): string => {
  const trimmed = path.replace(/\/+$/, '')
  return trimmed || '/'
}

export const navSection = (path: string): NavSection => {
  const clean = normalize(path)
  if (clean === '/') return 'shop'
  for (const [prefix, section] of PREFIXES) {
    if (clean === prefix || clean.startsWith(`${prefix}/`)) return section
  }
  return null
}
