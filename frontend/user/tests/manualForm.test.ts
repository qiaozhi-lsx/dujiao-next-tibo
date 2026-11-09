// 人工交付表单纯函数（src/utils/manualForm.ts）用例：useCheckout 与支付后开通弹窗共用同一份规则。
import test from 'node:test'
import assert from 'node:assert/strict'
import {
  buildManualFormRow,
  buildManualFormValues,
  compileManualRegex,
  normalizeManualFormSchema,
  validateManualFormValues,
} from '../src/utils/manualForm.ts'

const t = (key: string, params?: Record<string, any>) => `${key}${params ? JSON.stringify(params) : ''}`
const ctx = { getLabel: (f: { key: string }) => f.key, t }

test('normalizeManualFormSchema: 丢弃无 key / 未知类型字段，去重选项，数字字段容错', () => {
  const fields = normalizeManualFormSchema({
    fields: [
      { key: 'a', type: 'text', required: true, max_len: '10' },
      { key: '', type: 'text' },
      { key: 'b', type: 'unknown' },
      { key: 'c', type: 'select', options: ['x', ' x ', '', 'y'] },
      { key: 'd', type: 'number', min: 'abc', max: 5 },
    ],
  })
  assert.deepEqual(fields.map((f) => f.key), ['a', 'c', 'd'])
  assert.equal(fields[0]?.max_len, 10)
  assert.deepEqual(fields[1]?.options, ['x', 'y'])
  assert.equal(fields[2]?.min, undefined)
  assert.equal(fields[2]?.max, 5)
  assert.deepEqual(normalizeManualFormSchema(null), [])
})

test('compileManualRegex: 支持 /pattern/flags 与裸 pattern，非法返回 null', () => {
  assert.equal(compileManualRegex('/^a+$/i')?.test('AAA'), true)
  assert.equal(compileManualRegex('^a+$')?.test('aaa'), true)
  assert.equal(compileManualRegex('(['), null)
  assert.equal(compileManualRegex(''), null)
})

test('validateManualFormValues: 必填 / 长度 / 邮箱 / 数字范围 / 选项 / 正则', () => {
  const fields = normalizeManualFormSchema({
    fields: [
      { key: 'req', type: 'text', required: true },
      { key: 'len', type: 'text', max_len: 3 },
      { key: 'mail', type: 'email' },
      { key: 'num', type: 'number', min: 1, max: 5 },
      { key: 'sel', type: 'select', options: ['x', 'y'] },
      { key: 'rx', type: 'text', regex: '^\\d+$' },
      { key: 'chk', type: 'checkbox', required: true, options: ['a', 'b'] },
    ],
  })
  const errors = validateManualFormValues(fields, {
    req: '  ', len: 'abcd', mail: 'nope', num: '9', sel: 'z', rx: 'abc', chk: [],
  }, ctx)
  assert.deepEqual(Object.keys(errors).sort(), ['chk', 'len', 'mail', 'num', 'req', 'rx', 'sel'])
  assert.match(errors.req ?? '', /manualFormFieldRequired/)
  assert.match(errors.len ?? '', /manualFormFieldMaxLength/)
  assert.match(errors.num ?? '', /manualFormFieldNumberRange/)

  const ok = validateManualFormValues(fields, {
    req: 'v', len: 'abc', mail: 'a@b.co', num: '3', sel: 'x', rx: '123', chk: ['a'],
  }, ctx)
  assert.deepEqual(ok, {})
})

test('buildManualFormValues / buildManualFormRow: 规范初值，提交体去空白且省略空值', () => {
  const fields = normalizeManualFormSchema({
    fields: [
      { key: 'a', type: 'text' },
      { key: 'b', type: 'checkbox', options: ['x', 'y'] },
      { key: 'c', type: 'text' },
    ],
  })
  assert.deepEqual(buildManualFormValues(fields), { a: '', b: [], c: '' })
  assert.deepEqual(buildManualFormValues(fields, { a: 5, b: ['x', ''], extra: 1 }), { a: '5', b: ['x'], c: '' })
  assert.deepEqual(buildManualFormRow(fields, { a: '  hi ', b: [' x '], c: '   ' }), { a: 'hi', b: ['x'] })
  assert.deepEqual(buildManualFormRow(fields, { a: '', b: [], c: '' }), {})
})
