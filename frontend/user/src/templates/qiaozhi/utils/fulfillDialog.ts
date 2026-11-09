/**
 * 支付后「开通信息 / 开通进度」弹窗的纯函数：订单项判断、进度步骤构造、轮询节奏、预检错误解析。
 * 与 Vue / DOM / 存储无关，便于 node:test 直接覆盖。
 * 后端响应形状见 internal/modules/procurement/contract/types.go 的 FulfillProgress。
 * 会话（session_json）只存在于弹窗组件的内存里，本文件不接触其内容。
 */

export type FulfillPhase = 'awaiting_info' | 'submitting' | 'processing' | 'failed' | 'completed'

export interface FulfillProgressStep {
  key: string
  label: string
}

export interface FulfillProgress {
  phase: FulfillPhase
  steps: FulfillProgressStep[]
  step: { index: number; total: number; key: string; label: string } | null
  percent: number
  terminal: boolean
  outcome: string
  message: string
  poll_after_ms: number
}

export type FulfillStepState = 'done' | 'current' | 'pending' | 'failed'

export interface FulfillStepView {
  key: string
  label: string
  state: FulfillStepState
}

export interface FulfillStepLabels {
  paid: string
  submit: string
  process: string
  complete: string
}

export const DEFAULT_POLL_MS = 3000
/** 下限：游客读接口按 IP 限流（120/min），轮询间隔不低于 2s 才稳妥。 */
export const MIN_POLL_MS = 2000
export const MAX_POLL_MS = 15000

const hasFields = (schema: unknown): boolean => {
  if (!schema || typeof schema !== 'object') return false
  const fields = (schema as { fields?: unknown }).fields
  return Array.isArray(fields) && fields.length > 0
}

const isNonEmptyObject = (value: unknown): boolean =>
  !!value && typeof value === 'object' && Object.keys(value as object).length > 0

/** 订单项是否带表单快照（上游延后填写的商品）。 */
export const itemHasFormSnapshot = (item: any): boolean => hasFields(item?.manual_form_schema_snapshot)

/** 订单项等待填写开通信息：后端标记 fulfill_info_pending，或「有表单快照且提交值为空」。 */
export const isFulfillInfoPendingItem = (item: any): boolean =>
  Boolean(item?.fulfill_info_pending) || (itemHasFormSnapshot(item) && !isNonEmptyObject(item?.manual_form_submission))

/** 订单的所有订单项：下单是「父订单 + 每个商品一个子订单」，订单项在子订单上，这里把本单与子订单的订单项摊平。 */
export const collectOrderItems = (order: any): any[] => {
  const own = Array.isArray(order?.items) ? order.items : []
  const children = Array.isArray(order?.children) ? order.children : []
  return [...own, ...children.flatMap((child: any) => collectOrderItems(child))]
}

/** 订单里第一个待填写开通信息的订单项，没有则 null。 */
export const findFulfillPendingItem = (order: any): any | null =>
  collectOrderItems(order).find((item: any) => isFulfillInfoPendingItem(item)) || null

/** 后端汇总标记 fulfill_info_pending，或订单项（含子订单）里有待填写项。 */
export const orderHasFulfillInfoPending = (order: any): boolean =>
  Boolean(order?.fulfill_info_pending) || findFulfillPendingItem(order) !== null

/** 已提交开通信息、订单仍在处理（paid / fulfilling）：可查看开通进度。 */
export const orderIsFulfillProcuring = (order: any): boolean => {
  const status = String(order?.status || '')
  if (status !== 'paid' && status !== 'fulfilling') return false
  return collectOrderItems(order).some((item: any) => itemHasFormSnapshot(item) && isNonEmptyObject(item?.manual_form_submission))
}

export const isTerminalProgress = (progress: Pick<FulfillProgress, 'phase' | 'terminal'> | null | undefined): boolean =>
  !!progress && (progress.terminal || progress.phase === 'completed' || progress.phase === 'failed')

/** 下一次轮询间隔（毫秒）；终态与等待填写阶段返回 null（不再轮询）。 */
export const nextPollDelay = (progress: FulfillProgress | null | undefined): number | null => {
  if (!progress || isTerminalProgress(progress) || progress.phase === 'awaiting_info') return null
  const raw = Number(progress.poll_after_ms)
  const ms = Number.isFinite(raw) && raw > 0 ? raw : DEFAULT_POLL_MS
  return Math.min(MAX_POLL_MS, Math.max(MIN_POLL_MS, ms))
}

/** 归一化接口返回，缺字段补默认值，避免模板处处判空。 */
export const normalizeFulfillProgress = (raw: any): FulfillProgress => {
  const phases: FulfillPhase[] = ['awaiting_info', 'submitting', 'processing', 'failed', 'completed']
  const phase = phases.includes(raw?.phase) ? raw.phase : 'submitting'
  const steps = Array.isArray(raw?.steps)
    ? raw.steps
      .map((s: any) => ({ key: String(s?.key || ''), label: String(s?.label || '') }))
      .filter((s: FulfillProgressStep) => s.key && s.label)
    : []
  const step = raw?.step && typeof raw.step === 'object'
    ? {
      index: Number(raw.step.index) || 0,
      total: Number(raw.step.total) || steps.length,
      key: String(raw.step.key || ''),
      label: String(raw.step.label || ''),
    }
    : null
  const percent = Math.min(100, Math.max(0, Math.round(Number(raw?.percent) || 0)))
  return {
    phase,
    steps,
    step,
    percent,
    terminal: Boolean(raw?.terminal),
    outcome: String(raw?.outcome || ''),
    message: String(raw?.message || ''),
    poll_after_ms: Number(raw?.poll_after_ms) || 0,
  }
}

/**
 * 步骤列表：「支付成功」恒为已完成，其后是上游步骤（带 done / current / pending）。
 * 上游不提供步骤（旧上游 / 未带 order_progress 能力）时降级为 提交 → 处理 → 完成 三步。
 */
export const buildStepViews = (progress: FulfillProgress | null | undefined, labels: FulfillStepLabels): FulfillStepView[] => {
  const paid: FulfillStepView = { key: 'paid', label: labels.paid, state: 'done' }
  const phase = progress?.phase || 'submitting'
  const failed = phase === 'failed'
  const completed = phase === 'completed'

  if (progress && progress.steps.length > 0) {
    const currentIndex = progress.step?.index || (completed ? progress.steps.length + 1 : 1)
    const views = progress.steps.map((s, i): FulfillStepView => {
      const index = i + 1
      let state: FulfillStepState = 'pending'
      if (completed || index < currentIndex) state = 'done'
      else if (index === currentIndex) state = failed ? 'failed' : 'current'
      return { key: s.key, label: s.label, state }
    })
    return [paid, ...views]
  }

  const generic: Array<{ key: string; label: string }> = [
    { key: 'submit', label: labels.submit },
    { key: 'process', label: labels.process },
    { key: 'complete', label: labels.complete },
  ]
  let currentIndex = 1
  if (phase === 'processing') currentIndex = 2
  if (completed) currentIndex = 4
  const views = generic.map((s, i): FulfillStepView => {
    const index = i + 1
    let state: FulfillStepState = 'pending'
    if (index < currentIndex) state = 'done'
    else if (index === currentIndex) state = failed ? 'failed' : 'current'
    return { ...s, state }
  })
  return [paid, ...views]
}

/** 总进度百分比：优先上游给的 percent；降级时按阶段估算。 */
export const resolvePercent = (progress: FulfillProgress | null | undefined): number => {
  if (!progress) return 0
  if (progress.phase === 'completed') return 100
  if (progress.percent > 0) return progress.percent
  if (progress.phase === 'processing') return 50
  if (progress.phase === 'submitting') return 20
  return 0
}

export interface PrecheckRejection {
  reasonCode: string
  message: string
  accountEmail: string
  planType: string
  isActive: boolean
}

/** 从 API 错误（带 status / data）解析 422 预检拒绝；不是预检拒绝返回 null。 */
export const parsePrecheckRejection = (error: unknown): PrecheckRejection | null => {
  const err = error as { status?: number; data?: any; message?: string } | null
  if (!err || err.status !== 422 || !err.data || typeof err.data !== 'object') return null
  const data = err.data
  const account = data.account && typeof data.account === 'object' ? data.account : null
  return {
    reasonCode: String(data.reason_code || ''),
    message: String(data.message || err.message || ''),
    accountEmail: account ? String(account.email_masked || '') : '',
    planType: account ? String(account.plan_type || '') : '',
    isActive: account ? Boolean(account.is_active) : false,
  }
}

export const isConflictError = (error: unknown): boolean => (error as { status?: number } | null)?.status === 409

/** 与布局悬浮客服入口同一优先级的客服链接。 */
export const resolveSupportHref = (contact: { telegram?: string; whatsapp?: string; email?: string; support_url?: string } | null | undefined): string => {
  if (!contact) return ''
  if (contact.support_url) return contact.support_url
  if (contact.telegram) return contact.telegram
  if (contact.whatsapp) return `https://wa.me/${contact.whatsapp.replace(/[^\d]/g, '')}`
  if (contact.email) return `mailto:${contact.email}`
  return ''
}
