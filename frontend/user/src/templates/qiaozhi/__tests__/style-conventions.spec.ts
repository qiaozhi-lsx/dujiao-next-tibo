// Style-convention guard for the qiaozhi template.
//
// The template's design-system prefix was renamed wholesale to `qz` (class names,
// keyframes, <Transition> names, `--qz-press`, `Qz*.vue` components) and the base
// stylesheet to styles/legacy.css. Everything built on top of it uses the new names
// only, so this static check walks every .vue / .ts / .css file under
// templates/qiaozhi and fails with `file:line` for any legacy identifier that creeps
// back in.
//
// The legacy prefix is assembled at runtime so this file does not trip its own rule
// (or a repo-wide `rg` for it). Project uses node:test (no DOM, no vitest); picked up
// by pnpm run test via the shim at tests/qiaozhiStyleConventions.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync, readFileSync } from 'node:fs'
import { join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'

const TEMPLATE_ROOT = fileURLToPath(new URL('..', import.meta.url))
const SOURCE_FILE = /\.(vue|ts|css)$/

const LEGACY = ['j', '8'].join('')
const LEGACY_UPPER = LEGACY.toUpperCase()
// Bare word, class/keyframe/variable prefix, component prefix, and the old stylesheet name.
const LEGACY_PATTERN = new RegExp(String.raw`\b${LEGACY}\b|${LEGACY}-|${LEGACY_UPPER}[A-Z]|ji8\.`)

function sourceFiles(dir: string): string[] {
  const out: string[] = []
  for (const entry of readdirSync(dir, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))) {
    const full = join(dir, entry.name)
    if (entry.isDirectory()) out.push(...sourceFiles(full))
    else if (SOURCE_FILE.test(entry.name)) out.push(full)
  }
  return out
}

test(`no ${LEGACY} identifiers remain`, () => {
  const offenders: string[] = []
  const offendingFiles = new Set<string>()
  for (const file of sourceFiles(TEMPLATE_ROOT)) {
    readFileSync(file, 'utf8')
      .split('\n')
      .forEach((line, index) => {
        if (LEGACY_PATTERN.test(line)) {
          offenders.push(`${relative(TEMPLATE_ROOT, file)}:${index + 1}`)
          offendingFiles.add(file)
        }
      })
  }
  assert.equal(
    offenders.length,
    0,
    `${offendingFiles.size} file(s) under templates/qiaozhi still use the legacy "${LEGACY}" prefix (${offenders.length} line(s)):\n${offenders.join('\n')}`,
  )
})

// ---------------------------------------------------------------------------
// Token contract for styles/tokens.css.
//
// base.css / layout.css and the page <style scoped> blocks consume these names,
// so the list below is the contract: every token must be declared on the scope
// selector pair (`.qiaozhi-scope, body.qiaozhi-tokens` — the body variant is what
// Teleport-ed overlays see) and every colour / shadow token must be redefined on
// the `.dark` pair. Sizes are theme-independent and declared once.
// ---------------------------------------------------------------------------
const TOKENS_CSS = join(TEMPLATE_ROOT, 'styles', 'tokens.css')
const LIGHT_SCOPE = '.qiaozhi-scope, body.qiaozhi-tokens'
const DARK_SCOPE = '.dark .qiaozhi-scope, .dark body.qiaozhi-tokens'
const MOBILE_MEDIA = '@media (max-width: 760px)'

const QZ_COLOR_TOKENS = [
  '--qz-bg', '--qz-surface', '--qz-surface-raised', '--qz-surface-muted',
  '--qz-text', '--qz-text-secondary', '--qz-text-muted',
  '--qz-border', '--qz-border-subtle', '--qz-border-strong',
  '--qz-accent', '--qz-accent-strong', '--qz-accent-subtle', '--qz-on-accent',
  '--qz-success', '--qz-warning', '--qz-danger', '--qz-info',
  '--qz-header-bg',
]
const QZ_SHADOW_TOKENS = ['--qz-shadow-resting', '--qz-shadow-raised', '--qz-shadow-floating']
const QZ_SIZE_TOKENS = [
  ...[1, 2, 3, 4, 5, 6, 8, 10, 12, 16, 20, 24].map((step) => `--qz-space-${step}`),
  '--qz-radius-control', '--qz-radius-panel', '--qz-radius-surface', '--qz-radius-pill',
  '--qz-control-sm', '--qz-control-md', '--qz-control-lg',
  '--qz-text-2xs', '--qz-text-xs', '--qz-text-sm', '--qz-text-base', '--qz-text-lg', '--qz-text-xl', '--qz-text-2xl', '--qz-text-3xl',
  '--qz-leading-display', '--qz-leading-body', '--qz-tracking-display',
  '--qz-ease-enter', '--qz-ease-standard', '--qz-ease-exit', '--qz-ease-spring',
  '--qz-duration-instant', '--qz-duration-fast', '--qz-duration-standard', '--qz-duration-considered', '--qz-duration-spring',
  '--qz-focus-outline',
  '--qz-container', '--qz-header-h',
]

interface CssRule {
  media: string | null
  selector: string
  declarations: Map<string, string>
}

// Minimal CSS walker: enough for a token sheet (comments, plain rules, one level of @media).
// Values never contain `;`, so splitting declarations on it is safe here.
function parseRules(css: string, media: string | null = null): CssRule[] {
  const rules: CssRule[] = []
  let cursor = 0
  for (;;) {
    const open = css.indexOf('{', cursor)
    if (open === -1) break
    let depth = 1
    let close = open + 1
    while (close < css.length && depth > 0) {
      if (css[close] === '{') depth += 1
      else if (css[close] === '}') depth -= 1
      close += 1
    }
    const prelude = css.slice(cursor, open).trim().replace(/\s+/g, ' ')
    const body = css.slice(open + 1, close - 1)
    if (prelude.startsWith('@')) {
      rules.push(...parseRules(body, prelude))
    } else {
      const declarations = new Map<string, string>()
      for (const chunk of body.split(';')) {
        const colon = chunk.indexOf(':')
        if (colon === -1) continue
        declarations.set(chunk.slice(0, colon).trim(), chunk.slice(colon + 1).trim())
      }
      rules.push({ media, selector: prelude, declarations })
    }
    cursor = close
  }
  return rules
}

function declarationsFor(rules: CssRule[], selector: string, media: string | null): Map<string, string> {
  const merged = new Map<string, string>()
  for (const rule of rules) {
    if (rule.selector !== selector || rule.media !== media) continue
    for (const [name, value] of rule.declarations) merged.set(name, value)
  }
  return merged
}

test('tokens.css defines the qz token contract on both scopes', () => {
  const css = readFileSync(TOKENS_CSS, 'utf8').replace(/\/\*[\s\S]*?\*\//g, '')
  const rules = parseRules(css)
  const light = declarationsFor(rules, LIGHT_SCOPE, null)
  const dark = declarationsFor(rules, DARK_SCOPE, null)
  const mobile = declarationsFor(rules, LIGHT_SCOPE, MOBILE_MEDIA)

  const allTokens = [...QZ_COLOR_TOKENS, ...QZ_SHADOW_TOKENS, ...QZ_SIZE_TOKENS]
  assert.deepEqual(
    allTokens.filter((name) => !light.has(name)),
    [],
    `tokens missing from "${LIGHT_SCOPE}"`,
  )
  const themed = [...QZ_COLOR_TOKENS, ...QZ_SHADOW_TOKENS]
  assert.deepEqual(
    themed.filter((name) => !dark.has(name)),
    [],
    `colour / shadow tokens missing from "${DARK_SCOPE}"`,
  )
  assert.deepEqual(
    QZ_SIZE_TOKENS.filter((name) => dark.has(name)),
    [],
    `size tokens are theme-independent and must not be redefined on "${DARK_SCOPE}"`,
  )

  // Literal values other tasks hard-depend on (Arc control heights, type floor, container).
  assert.equal(light.get('--qz-control-sm'), '36px')
  assert.equal(light.get('--qz-control-md'), '44px')
  assert.equal(light.get('--qz-control-lg'), '50px')
  assert.equal(light.get('--qz-text-2xs'), '11px')
  assert.equal(light.get('--qz-container'), '1200px')

  // ≤760px overrides: touch-friendly small controls and the compact header.
  assert.equal(mobile.get('--qz-control-sm'), '40px', `${MOBILE_MEDIA} must set --qz-control-sm: 40px`)
  assert.equal(mobile.get('--qz-header-h'), '60px', `${MOBILE_MEDIA} must set --qz-header-h: 60px`)
})

// ---------------------------------------------------------------------------
// Base classes in styles/base.css consume tokens only.
//
// Colours come from --qz-* (color-mix() of tokens is fine), every font-size is a
// --qz-text-* step and every control height is a --qz-control-* step. The only fixed
// heights are glyphs inside controls (16 / 18), badge (22), chip (24) and the
// empty-state icon (40). QzButton.vue keeps just its label morph / loader CSS; its
// size and variant looks live in base.css as .qz-btn--*.
// ---------------------------------------------------------------------------
const BASE_CSS = join(TEMPLATE_ROOT, 'styles', 'base.css')
const QZ_BUTTON_VUE = join(TEMPLATE_ROOT, 'components', 'QzButton.vue')
const BASE_PX_HEIGHTS = new Set(['16px', '18px', '22px', '24px', '40px'])
const QZ_BUTTON_STATE_CLASSES = ['is-sm', 'is-md', 'is-lg', 'is-primary', 'is-secondary', 'is-ghost', 'is-danger']

const lineOf = (text: string, index: number) => text.slice(0, index).split('\n').length

test('base.css only consumes tokens', () => {
  const css = readFileSync(BASE_CSS, 'utf8').replace(/\/\*[\s\S]*?\*\//g, '')

  const colourLiterals = [...css.matchAll(/#[0-9a-f]{3,8}\b|\b(?:rgba?|oklch)\(/gi)].map(
    (match) => `base.css:${lineOf(css, match.index)} ${match[0]}`,
  )
  assert.deepEqual(colourLiterals, [], 'base.css must not contain #hex / rgb() / rgba() / oklch() literals; use --qz-* tokens')

  const fontSizes = [...css.matchAll(/font-size\s*:\s*([^;}]+)/g)]
    .filter(([, value = '']) => !value.trim().startsWith('var(--qz-text-'))
    .map((match) => `base.css:${lineOf(css, match.index)} font-size: ${(match[1] ?? '').trim()}`)
  assert.deepEqual(fontSizes, [], 'every font-size in base.css must be a --qz-text-* token')

  const heights = [...css.matchAll(/(?<![\w-])(min-height|height)\s*:\s*([^;}]+)/g)]
    .filter(([, , value = '']) => {
      const trimmed = value.trim()
      return /^[\d.]+px$/.test(trimmed) ? !BASE_PX_HEIGHTS.has(trimmed) : !trimmed.startsWith('var(--qz-control-')
    })
    .map((match) => `base.css:${lineOf(css, match.index)} ${match[1]}: ${(match[2] ?? '').trim()}`)
  assert.deepEqual(
    heights,
    [],
    `height / min-height in base.css must be var(--qz-control-*) or one of ${[...BASE_PX_HEIGHTS].join(' / ')}`,
  )
})

test('QzButton.vue has no size/variant css of its own', () => {
  const sfc = readFileSync(QZ_BUTTON_VUE, 'utf8')
  const style = [...sfc.matchAll(/<style[^>]*>([\s\S]*?)<\/style>/g)].map(([, body = '']) => body).join('\n')
  assert.ok(style.trim().length > 0, 'QzButton.vue should keep a <style> block for its label morph and loader')

  const leaked = QZ_BUTTON_STATE_CLASSES.filter((name) => style.includes(name))
  assert.deepEqual(leaked, [], 'QzButton size / variant styling belongs to .qz-btn--* in styles/base.css')
})
