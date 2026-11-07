// 带 .ts 后缀：本模块被 node:test 直接加载（见 __tests__/starting-price.spec.ts）。
import { amountToCents } from '../../../utils/money.ts'

/**
 * 商品列表卡片展示的「起」价：启用 SKU 里最低的 price_amount；没有 SKU（或全部停用、金额无效）
 * 时回退到商品自身的 price_amount。返回原始金额值（number / string），交给 formatQiaozhiPrice /
 * splitQiaozhiPrice 格式化。比较按分进行，所以 "50.00" 与 195 这类混合写法也能正确排序。
 *
 * `is_active === false` 视为停用；列表接口若省略该字段则视为启用。
 */
export const startingPrice = (product: any): unknown => {
  const skus: any[] = Array.isArray(product?.skus) ? product.skus : []
  let lowest: unknown
  let lowestCents = Number.POSITIVE_INFINITY
  for (const sku of skus) {
    if (!sku || sku.is_active === false) continue
    const cents = amountToCents(sku.price_amount)
    if (cents === null || cents >= lowestCents) continue
    lowestCents = cents
    lowest = sku.price_amount
  }
  return lowest === undefined ? product?.price_amount : lowest
}
