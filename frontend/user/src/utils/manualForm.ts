/**
 * 人工交付表单（manual_form_schema）的纯函数：schema 归一化、字段校验、提交体构造。
 * 与 Vue / DOM / 存储无关，结算页（useCheckout）与支付后开通弹窗共用，便于 node:test 直接覆盖。
 * 规则需与后端 internal/modules/catalog/product/manualform 保持一致。
 */

export interface ManualFormField {
  key: string
  type: string
  required: boolean
  label?: Record<string, string>
  placeholder?: Record<string, string>
  regex?: string
  min?: number
  max?: number
  max_len?: number
  options: string[]
}

export type ManualFormValues = Record<string, any>

export const manualFieldTypes = new Set(['text', 'textarea', 'phone', 'email', 'number', 'select', 'radio', 'checkbox'])
export const manualEmailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
export const manualPhonePattern = /^\+?[0-9\-()\s]{6,20}$/

const findLastUnescapedSlash = (value: string): number => {
  for (let index = value.length - 1; index > 0; index -= 1) {
    if (value[index] !== '/') {
      continue
    }
    let backslashes = 0
    for (let cursor = index - 1; cursor >= 0 && value[cursor] === '\\'; cursor -= 1) {
      backslashes += 1
    }
    if (backslashes % 2 === 0) {
      return index
    }
  }
  return -1
}

export const compileManualRegex = (rawRegex?: string): RegExp | null => {
  const text = String(rawRegex || '').trim()
  if (!text) {
    return null
  }

  if (text.startsWith('/')) {
    const lastSlashIndex = findLastUnescapedSlash(text)
    if (lastSlashIndex > 0) {
      const pattern = text.slice(1, lastSlashIndex)
      const flags = text.slice(lastSlashIndex + 1)
      if (/^[gimsuy]*$/.test(flags)) {
        try {
          return new RegExp(pattern, flags)
        } catch {
          return null
        }
      }
    }
  }

  try {
    return new RegExp(text)
  } catch {
    return null
  }
}

export const normalizeManualFormSchema = (rawSchema: any): ManualFormField[] => {
  const rawFields = Array.isArray(rawSchema?.fields) ? rawSchema.fields : []
  return rawFields
    .map((rawField: any) => {
      const key = String(rawField?.key || '').trim()
      const type = String(rawField?.type || '').trim()
      if (!key || !manualFieldTypes.has(type)) {
        return null
      }
      const options = Array.isArray(rawField?.options)
        ? rawField.options.map((item: any) => String(item || '').trim()).filter(Boolean)
        : []
      const minValue = Number(rawField?.min)
      const maxValue = Number(rawField?.max)
      const maxLenValue = Number(rawField?.max_len)
      return {
        key,
        type,
        required: Boolean(rawField?.required),
        label: rawField?.label || undefined,
        placeholder: rawField?.placeholder || undefined,
        regex: String(rawField?.regex || '').trim() || undefined,
        min: Number.isFinite(minValue) ? minValue : undefined,
        max: Number.isFinite(maxValue) ? maxValue : undefined,
        max_len: Number.isFinite(maxLenValue) ? maxLenValue : undefined,
        options: Array.from(new Set(options)),
      } as ManualFormField
    })
    .filter(Boolean) as ManualFormField[]
}

/** 以已有输入为基础，为每个字段生成规范的初始值（checkbox → 数组，其余 → 字符串）。 */
export const buildManualFormValues = (fields: ManualFormField[], current: ManualFormValues = {}): ManualFormValues => {
  const values: ManualFormValues = {}
  fields.forEach((field) => {
    const currentValue = current[field.key]
    if (field.type === 'checkbox') {
      values[field.key] = Array.isArray(currentValue)
        ? currentValue.map((item: any) => String(item)).filter(Boolean)
        : []
    } else {
      values[field.key] = currentValue == null ? '' : String(currentValue)
    }
  })
  return values
}

export interface ManualFormValidateContext {
  /** 字段展示名（已按语言解析） */
  getLabel: (field: ManualFormField) => string
  /** i18n 翻译函数，使用 checkout.manualFormField* 系列 key */
  t: (key: string, params?: Record<string, any>) => string
}

/** 校验一组字段，返回 fieldKey → 错误文案（无错误则为空对象）。 */
export const validateManualFormValues = (
  fields: ManualFormField[],
  values: ManualFormValues,
  ctx: ManualFormValidateContext,
): Record<string, string> => {
  const errors: Record<string, string> = {}
  const setError = (field: ManualFormField, message: string) => {
    if (!errors[field.key]) {
      errors[field.key] = message
    }
  }

  fields.forEach((field) => {
    const fieldLabel = ctx.getLabel(field)
    const rawValue = values[field.key]
    if (field.type === 'checkbox') {
      const list = Array.isArray(rawValue)
        ? rawValue.map((item: any) => String(item).trim()).filter(Boolean)
        : []
      if (field.required && list.length === 0) {
        setError(field, ctx.t('checkout.manualFormFieldRequired', { name: fieldLabel }))
        return
      }
      if (list.length > 0 && field.options.length > 0 && list.some((item) => !field.options.includes(item))) {
        setError(field, ctx.t('checkout.manualFormFieldOptionInvalid', { name: fieldLabel }))
      }
      return
    }

    const text = rawValue == null ? '' : String(rawValue).trim()
    if (field.required && !text) {
      setError(field, ctx.t('checkout.manualFormFieldRequired', { name: fieldLabel }))
      return
    }
    if (!text) {
      return
    }

    if ((field.type === 'text' || field.type === 'textarea' || field.type === 'phone' || field.type === 'email') && field.max_len && text.length > field.max_len) {
      setError(field, ctx.t('checkout.manualFormFieldMaxLength', { name: fieldLabel, max: field.max_len }))
      return
    }
    if ((field.type === 'phone' && !manualPhonePattern.test(text)) || (field.type === 'email' && !manualEmailPattern.test(text))) {
      setError(field, ctx.t('checkout.manualFormFieldInvalid', { name: fieldLabel }))
      return
    }
    if (field.type === 'number') {
      const numberValue = Number(text)
      if (!Number.isFinite(numberValue)) {
        setError(field, ctx.t('checkout.manualFormFieldNumberInvalid', { name: fieldLabel }))
        return
      }
      if ((field.min !== undefined && numberValue < field.min) || (field.max !== undefined && numberValue > field.max)) {
        setError(field, ctx.t('checkout.manualFormFieldNumberRange', { name: fieldLabel }))
        return
      }
    }
    if ((field.type === 'select' || field.type === 'radio') && field.options.length > 0 && !field.options.includes(text)) {
      setError(field, ctx.t('checkout.manualFormFieldOptionInvalid', { name: fieldLabel }))
      return
    }
    if (field.regex) {
      const regex = compileManualRegex(field.regex)
      if (!regex || !regex.test(text)) {
        setError(field, ctx.t('checkout.manualFormFieldInvalid', { name: fieldLabel }))
      }
    }
  })
  return errors
}

/** 构造提交体里的一行：去空白，空值与空数组不出现。 */
export const buildManualFormRow = (fields: ManualFormField[], values: ManualFormValues): Record<string, any> => {
  const row: Record<string, any> = {}
  fields.forEach((field) => {
    const rawValue = values[field.key]
    if (field.type === 'checkbox') {
      const list = Array.isArray(rawValue)
        ? rawValue.map((item: any) => String(item).trim()).filter(Boolean)
        : []
      if (list.length > 0) {
        row[field.key] = list
      }
      return
    }
    const text = rawValue == null ? '' : String(rawValue).trim()
    if (text) {
      row[field.key] = text
    }
  })
  return row
}
