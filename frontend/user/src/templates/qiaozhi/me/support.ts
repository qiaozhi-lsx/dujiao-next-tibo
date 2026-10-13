import type { ComposerTranslation } from 'vue-i18n'
import type { PublicMemberLevel } from '../../../api'
import { orderStatusLabel, orderStatusVariant } from '../../../utils/status'

/**
 * qiaozhi 个人中心面板共用工具：状态胶囊 tone / 文本格式化 / 会员等级名。
 * tone 与 ji8.css 的 .j8-me-pill.is-* 对应。
 */
export type MePillTone = 'ok' | 'warn' | 'danger' | 'info' | 'neutral'

const TONE_MAP: Record<string, MePillTone> = {
  success: 'ok',
  accent: 'ok',
  warning: 'warn',
  danger: 'danger',
  info: 'info',
  neutral: 'neutral',
}

export const statusPillTone = (status?: string): MePillTone => TONE_MAP[orderStatusVariant(status)] || 'neutral'

export const statusLabel = (t: ComposerTranslation, status?: string) => orderStatusLabel(t, status)

export const formatMoney = (amount?: string | number | null, currency?: string | null) => {
  if (amount === null || amount === undefined || amount === '') return '-'
  if (currency === null || currency === undefined || currency === '') return String(amount)
  return `${amount} ${currency}`
}

export const formatDate = (raw?: string | null) => {
  if (!raw) return '-'
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw
  return date.toLocaleString()
}

export const isImagePath = (icon?: string | null) =>
  Boolean(icon && (icon.startsWith('/uploads/') || icon.startsWith('http')))

export const levelName = (
  t: ComposerTranslation,
  locale: string,
  level?: PublicMemberLevel | null,
) => {
  if (!level) return t('personalCenter.memberLevel.defaultLevel')
  return level.name[locale] || level.name['zh-CN'] || level.name['en'] || level.slug
    || t('personalCenter.memberLevel.defaultLevel')
}
