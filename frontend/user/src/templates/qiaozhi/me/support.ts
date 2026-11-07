import type { ComposerTranslation } from 'vue-i18n'
import type { PublicMemberLevel } from '../../../api'
import { orderStatusLabel, orderStatusVariant, type BadgeTone } from '../../../utils/status'

/**
 * qiaozhi 个人中心面板共用工具：状态徽章 tone / 文本格式化 / 会员等级名。
 * tone 对应 base.css 的 `.qz-badge.is-*`（success / warning / danger / info / accent），`neutral` 不带状态类。
 */
export type MeBadgeTone = BadgeTone

export const statusBadgeTone = (status?: string): MeBadgeTone => orderStatusVariant(status)

export const badgeToneClass = (tone: MeBadgeTone) => (tone === 'neutral' ? '' : `is-${tone}`)

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
