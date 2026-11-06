// Regression guard: every static i18n key referenced by the qiaozhi template must
// exist in all three locale files (zh-CN / zh-TW / en-US).
//
// vue-i18n renders a missing key as the key itself, so a typo or a forgotten locale
// entry ships silently (screen readers end up announcing "common.previous"). The
// project uses node:test (no DOM, no vitest), so this is a static check: scan
// src/templates/qiaozhi/**/*.{vue,ts} (minus __tests__) for `t('…')` / `$t("…")` /
// `te('…')` / `tm('…')` calls with a string-literal first argument, flatten the locale
// JSON into dot paths, and assert each key resolves in every locale. Dynamic keys
// (template literals, variables, ternaries) are out of scope. Picked up by
// pnpm run test via the shim at tests/qiaozhiI18nKeys.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync, readFileSync } from 'node:fs'
import { join, relative, sep } from 'node:path'
import { fileURLToPath } from 'node:url'

const TEMPLATE_ROOT = fileURLToPath(new URL('..', import.meta.url))
const LOCALES_DIR = fileURLToPath(new URL('../../../i18n/locales/', import.meta.url))
const LOCALES = ['zh-CN', 'zh-TW', 'en-US'] as const

// `t('key')`, `$t("key")`, `te('key')`, `tm('key')` with a static string-literal key.
// `(?<!\w)` keeps `format(`, `split(`, `.at(` etc. out; `$` and backtick are excluded from
// the key body so template literals / interpolated keys are never collected.
const KEY_CALL = /(?<!\w)(?:\$t|t|te|tm)\(\s*(['"])([^'"`$]+)\1/g

type Usage = { key: string; file: string; line: number }

function sourceFiles(dir: string): string[] {
  const out: string[] = []
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name)
    if (entry.isDirectory()) {
      if (entry.name !== '__tests__') out.push(...sourceFiles(full))
    } else if (/\.(vue|ts)$/.test(entry.name)) {
      out.push(full)
    }
  }
  return out.sort()
}

// First occurrence of each distinct key, in file / line order.
function collectUsages(files: string[]): Map<string, Usage> {
  const firstSeen = new Map<string, Usage>()
  for (const file of files) {
    const source = readFileSync(file, 'utf8')
    for (const match of source.matchAll(KEY_CALL)) {
      const key = match[2] ?? ''
      if (!key || firstSeen.has(key)) continue
      const line = source.slice(0, match.index).split('\n').length
      firstSeen.set(key, { key, file: relative(TEMPLATE_ROOT, file).split(sep).join('/'), line })
    }
  }
  return firstSeen
}

type LocaleIndex = { leaves: Set<string>; namespaces: Set<string> }

// Flatten nested messages into dot paths. `leaves` are translatable values; `namespaces`
// are intermediate nodes, which count as present for `tm('ns')` / `te('ns')` lookups.
function indexLocale(messages: unknown): LocaleIndex {
  const leaves = new Set<string>()
  const namespaces = new Set<string>()
  const walk = (node: unknown, path: string) => {
    if (node !== null && typeof node === 'object') {
      if (path) namespaces.add(path)
      for (const [k, v] of Object.entries(node)) walk(v, path ? `${path}.${k}` : k)
    } else if (path) {
      leaves.add(path)
    }
  }
  walk(messages, '')
  return { leaves, namespaces }
}

const hasKey = (index: LocaleIndex, key: string) => index.leaves.has(key) || index.namespaces.has(key)

test('every static i18n key used by qiaozhi exists in zh-CN, zh-TW and en-US', () => {
  const files = sourceFiles(TEMPLATE_ROOT)
  assert.ok(files.length > 0, `no .vue/.ts sources found under ${TEMPLATE_ROOT}`)

  const usages = collectUsages(files)
  // Sanity check on the scanner itself: QiaozhiLayout.vue always renders this key.
  assert.ok(usages.has('qiaozhi.nav.shop'), `scanner collected ${usages.size} keys but missed qiaozhi.nav.shop`)

  const indexes = LOCALES.map((locale) => ({
    locale,
    index: indexLocale(JSON.parse(readFileSync(join(LOCALES_DIR, `${locale}.json`), 'utf8'))),
  }))

  const problems: string[] = []
  for (const { key, file, line } of usages.values()) {
    const missing = indexes.filter(({ index }) => !hasKey(index, key)).map(({ locale }) => locale)
    if (missing.length) problems.push(`${key} — missing in ${missing.join(', ')} — first used at ${file}:${line}`)
  }

  assert.deepEqual(
    problems,
    [],
    `${problems.length} i18n key(s) used by qiaozhi are missing from locale files:\n  ${problems.join('\n  ')}`,
  )
})
