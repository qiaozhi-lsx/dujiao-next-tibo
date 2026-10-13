/**
 * 订单详情「交付单元」构建（qiaozhi 订单详情页专用）。
 *
 * 后端数据模型里，主订单的 items 是全部子订单商品的聚合快照：子订单存在时，
 * 主订单 items 与 children[].items 是同一批商品。若两者都渲染成单元，页面会
 * 出现重复商品（同一商品显示两遍），因此有子订单时主订单只在其自身存在独立
 * 交付（order.fulfillment）时才追加一个不含商品明细的单元。
 */
export interface DeliveryUnit {
  key: string
  orderNo: string
  orderAmount: string | undefined
  currency: string
  status: string | null
  isChild: boolean
  items: any[]
  fulfillment: any | null
}

export type ChildStatusResolver = (child: any) => string | null

export const buildDeliveryUnits = (
  order: any,
  resolveChildStatus: ChildStatusResolver,
): DeliveryUnit[] => {
  if (!order) return []
  const units: DeliveryUnit[] = []
  const children = Array.isArray(order.children) ? order.children : []
  for (const child of children) {
    units.push({
      key: `child-${String(child?.id ?? child?.order_no ?? units.length)}`,
      orderNo: String(child?.order_no || order.order_no || ''),
      orderAmount: child?.total_amount ?? undefined,
      currency: child?.currency || order.currency || '',
      status: child ? resolveChildStatus(child) : null,
      isChild: true,
      items: Array.isArray(child?.items) ? child.items : [],
      fulfillment: child?.fulfillment || null,
    })
  }
  const items = Array.isArray(order.items) ? order.items : []
  if (!children.length) {
    units.push({
      key: 'main',
      orderNo: String(order.order_no || ''),
      orderAmount: undefined,
      currency: order.currency || '',
      status: order.status || null,
      isChild: false,
      items,
      fulfillment: order.fulfillment || null,
    })
  } else if (order.fulfillment || (items.length && !children.some((child: any) => Array.isArray(child?.items) && child.items.length))) {
    units.push({
      key: 'main',
      orderNo: String(order.order_no || ''),
      orderAmount: undefined,
      currency: order.currency || '',
      status: order.status || null,
      isChild: false,
      items: order.fulfillment ? [] : items,
      fulfillment: order.fulfillment || null,
    })
  }
  return units
}
