/**
 * 商品页内嵌下单（create-and-pay）请求体构造与手动表单商品判断的纯函数。
 * 与 DOM / Vue / 存储无关：分销归因由调用方从 utils/affiliate 读好后传入，便于 node:test 直接覆盖。
 * 后端字段见 internal/modules/order/transport/http/create_handler.go 的
 * CreateOrderAndPayRequest / CreateGuestOrderAndPayRequest。
 */

export type CreateAndPayGuestInput = {
  email: string
  orderPassword: string
  /** image → { captcha_id, captcha_code }；turnstile → { turnstile_token }；无验证码 → {} */
  captchaPayload: Record<string, unknown>
}

export type CreateAndPayPayloadInput = {
  items: unknown[]
  channelId: number | null
  affiliateCode: string
  affiliateVisitorKey: string
  /** 游客下单时传入；登录用户省略 */
  guest?: CreateAndPayGuestInput
}

/**
 * 构造 /orders/create-and-pay 与 /guest/orders/create-and-pay 的请求体。
 * 未选渠道（null）与空归因（''）统一置为 undefined，JSON 序列化时该键不会出现，与 classic useCheckout 一致。
 */
export const buildCreateAndPayPayload = (input: CreateAndPayPayloadInput): Record<string, unknown> => {
  const payload: Record<string, unknown> = {
    items: input.items,
    channel_id: input.channelId || undefined,
    affiliate_code: input.affiliateCode || undefined,
    affiliate_visitor_key: input.affiliateVisitorKey || undefined,
  }
  if (input.guest) {
    payload.email = input.guest.email
    payload.order_password = input.guest.orderPassword
    payload.captcha_payload = input.guest.captchaPayload
  }
  return payload
}

/**
 * 商品是否配置了手动表单字段（product.manual_form_schema.fields 非空）。
 * 这类商品需在结算页填写交付资料，商品页直接下单会被后端拒绝。
 */
export const hasManualFormFields = (schema: unknown): boolean => {
  if (!schema || typeof schema !== 'object') return false
  const fields = (schema as { fields?: unknown }).fields
  return Array.isArray(fields) && fields.length > 0
}
