// Style-convention guard for the qiaozhi template.
//
// The template's design-system prefix was renamed wholesale to `qz` (class names,
// keyframes, <Transition> names, `--qz-press`, `Qz*.vue` components) and the old
// sheet it came with has since been dissolved into tokens.css / base.css / layout.css
// and the page <style> blocks. Everything built on top of it uses the new names
// only, so this static check walks every .vue / .ts / .css file under
// templates/qiaozhi and fails with `file:line` for any old identifier that creeps
// back in — the former prefix, the former sheet name, and the word that named the
// transitional stylesheet (see the "dissolved stylesheet" test below).
//
// The old identifiers are assembled at runtime so this file does not trip its own
// rules (or a repo-wide `rg` for them). Project uses node:test (no DOM, no vitest);
// picked up by pnpm run test via the shim at tests/qiaozhiStyleConventions.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { existsSync, readdirSync, readFileSync } from 'node:fs'
import { join, relative, sep } from 'node:path'
import { fileURLToPath } from 'node:url'

const TEMPLATE_ROOT = fileURLToPath(new URL('..', import.meta.url))
const SOURCE_FILE = /\.(vue|ts|css)$/

const OLD_PREFIX = ['j', '8'].join('')
// The old stylesheet / prose name (the prefix with an `i` in the middle), also assembled at runtime.
const OLD_SHEET = ['j', 'i8'].join('')
// Case-insensitive and unanchored on purpose: any occurrence of either string is a hit,
// whether it is a bare word, a class / component prefix, a comment or the stylesheet name.
const OLD_PREFIX_PATTERN = new RegExp(`${OLD_PREFIX}|${OLD_SHEET}`, 'i')

function sourceFiles(dir: string): string[] {
  const out: string[] = []
  for (const entry of readdirSync(dir, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))) {
    const full = join(dir, entry.name)
    if (entry.isDirectory()) out.push(...sourceFiles(full))
    else if (SOURCE_FILE.test(entry.name)) out.push(full)
  }
  return out
}

test(`no ${OLD_PREFIX} identifiers remain`, () => {
  const offenders: string[] = []
  const offendingFiles = new Set<string>()
  for (const file of sourceFiles(TEMPLATE_ROOT)) {
    readFileSync(file, 'utf8')
      .split('\n')
      .forEach((line, index) => {
        if (OLD_PREFIX_PATTERN.test(line)) {
          offenders.push(`${relative(TEMPLATE_ROOT, file)}:${index + 1}`)
          offendingFiles.add(file)
        }
      })
  }
  assert.equal(
    offenders.length,
    0,
    `${offendingFiles.size} file(s) under templates/qiaozhi still use the old "${OLD_PREFIX}" prefix (${offenders.length} line(s)):\n${offenders.join('\n')}`,
  )
})

// ---------------------------------------------------------------------------
// The transitional stylesheet is gone.
//
// styles/<word>.css carried the old variable bridge and the rules pages had not yet
// taken over. Its bridge now lives in tokens.css (compat section), its base rules in
// base.css, the rest in the pages that used them. The file must not come back, and
// nothing under templates/qiaozhi — code or comments — may refer to it by name: a
// comment that says "still in <word>.css" is a stale map. The word is assembled at
// runtime so this file does not trip its own rule. The one tolerated occurrence is the
// backend fee-policy value `<word>_customer_surcharge` (a wire enum, not a stylesheet).
// ---------------------------------------------------------------------------
const DISSOLVED_SHEET = ['leg', 'acy'].join('')
const DISSOLVED_SHEET_PATTERN = new RegExp(DISSOLVED_SHEET, 'i')
const WIRE_ENUM_PATTERN = new RegExp(`${DISSOLVED_SHEET}_customer_surcharge`, 'g')

test('the dissolved stylesheet does not exist and nothing refers to it', () => {
  assert.ok(!existsSync(join(TEMPLATE_ROOT, 'styles', `${DISSOLVED_SHEET}.css`)), `styles/${DISSOLVED_SHEET}.css must stay deleted`)
  const offenders: string[] = []
  for (const file of sourceFiles(TEMPLATE_ROOT)) {
    readFileSync(file, 'utf8')
      .split('\n')
      .forEach((line, index) => {
        if (DISSOLVED_SHEET_PATTERN.test(line.replace(WIRE_ENUM_PATTERN, ''))) {
          offenders.push(`${relative(TEMPLATE_ROOT, file)}:${index + 1}`)
        }
      })
  }
  assert.deepEqual(offenders, [], `${offenders.length} line(s) under templates/qiaozhi still mention "${DISSOLVED_SHEET}":\n${offenders.join('\n')}`)
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

// Third-party brand colours (contact panel glyphs / open buttons) plus the ink that sits on
// those saturated fills. Declared on both scopes like every colour token, but identical in
// dark: a brand colour does not follow the theme, and only the dark ink clears 4.5:1 on it.
const QZ_BRAND_TOKENS = ['--qz-brand-telegram', '--qz-brand-whatsapp', '--qz-brand-wechat', '--qz-brand-qq', '--qz-on-brand']
// Same contract for the QR quiet-zone frame: white in both themes so a code on a dark surface keeps its light margin.
const QZ_FIXED_TOKENS = [...QZ_BRAND_TOKENS, '--qz-qr-frame']
// Status tones: always paired with a label or icon; light values are ≥ 4.5:1 on the white surface, dark are Arc's pastels.
const QZ_STATUS_TOKENS = ['--qz-success', '--qz-warning', '--qz-danger', '--qz-info']
const QZ_COLOR_TOKENS = [
  '--qz-bg', '--qz-surface', '--qz-surface-raised', '--qz-surface-muted',
  '--qz-text', '--qz-text-secondary', '--qz-text-muted',
  '--qz-border', '--qz-border-subtle', '--qz-border-strong',
  '--qz-accent', '--qz-accent-strong', '--qz-accent-subtle', '--qz-on-accent',
  ...QZ_STATUS_TOKENS,
  '--qz-header-bg',
  ...QZ_FIXED_TOKENS,
]
const QZ_SHADOW_TOKENS = ['--qz-shadow-resting', '--qz-shadow-raised', '--qz-shadow-floating']
const QZ_SIZE_TOKENS = [
  ...[1, 2, 3, 4, 5, 6, 8, 10, 12, 16, 20, 24].map((step) => `--qz-space-${step}`),
  '--qz-radius-control', '--qz-radius-panel', '--qz-radius-surface', '--qz-radius-pill', '--qz-radius-sm', '--qz-radius-xs',
  '--qz-control-sm', '--qz-control-md', '--qz-control-lg',
  '--qz-text-2xs', '--qz-text-xs', '--qz-text-sm', '--qz-text-base', '--qz-text-lg', '--qz-text-xl', '--qz-text-2xl', '--qz-text-3xl',
  '--qz-leading-display', '--qz-leading-body', '--qz-tracking-display', '--qz-font-body', '--qz-font-mono',
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

  // Muted text on the white surface must keep ≥ 4.5:1 (WCAG AA): oklch 55% grey is ~4.9:1, Arc's 59% was ~4.1:1.
  assert.equal(light.get('--qz-text-muted'), 'oklch(55% 0 0)', 'light --qz-text-muted must stay at oklch(55% 0 0)')

  // Brand colours and the QR frame are theme-independent values declared on both scopes.
  assert.deepEqual(
    QZ_FIXED_TOKENS.filter((name) => light.get(name) !== dark.get(name)),
    [],
    'brand / QR-frame tokens must have the same value on the light and dark scopes',
  )
  // White on the brand fills is 2.0–2.8:1; the fixed dark ink is ≥ 7.1:1 on every one of them.
  assert.equal(light.get('--qz-on-brand'), '#0a0a0a', '--qz-on-brand must stay the dark ink (#0a0a0a) in both themes')

  // Status tones are themed: four names, each with its own value on both scopes (the dark pastels differ from light).
  for (const name of QZ_STATUS_TOKENS) {
    assert.ok(light.has(name) && dark.has(name), `${name} must be declared on both scopes`)
    assert.notEqual(light.get(name), dark.get(name), `${name} must have a dark value of its own`)
  }

  // Featured-card artwork palette: the family exists (not enumerated by name — QzFeaturedCard.vue owns the
  // list) and every member declared on one scope is declared on the other, so dark mode never falls back
  // to an undefined var().
  const artLight = [...light.keys()].filter((name) => name.startsWith('--qz-art-')).sort()
  const artDark = [...dark.keys()].filter((name) => name.startsWith('--qz-art-')).sort()
  assert.ok(artLight.length >= 10, `expected the --qz-art-* family on "${LIGHT_SCOPE}", found ${artLight.length}`)
  assert.deepEqual(artDark, artLight, 'every --qz-art-* token must be declared on both the light and dark scopes')

  // ≤760px overrides: touch-friendly small controls and the compact header.
  assert.equal(mobile.get('--qz-control-sm'), '40px', `${MOBILE_MEDIA} must set --qz-control-sm: 40px`)
  assert.equal(mobile.get('--qz-header-h'), '60px', `${MOBILE_MEDIA} must set --qz-header-h: 60px`)
})

// ---------------------------------------------------------------------------
// Compat bridge in tokens.css: the classic `--ui-*` names (and `--font-mono`) that shared
// classic components rendered inside the template still read — Toast / ConfirmDialog /
// Loading / AnnouncementModal / the captcha widgets / shadcn Button + Card through the Tailwind
// @theme map, `.theme-prose` / `.theme-overlay` in src/style.css, the `* { border-color }`
// and `body { background }` globals. Each alias is declared once on the scope pair (the
// values are var() references, so they follow the theme by themselves) and resolves to
// `--qz-*` only: no colour literal, no raw length — a classic value leaking through here is
// exactly the fallback the bridge exists to prevent.
// ---------------------------------------------------------------------------
const BRIDGE_REQUIRED = [
  '--font-mono',
  '--ui-bg-page', '--ui-bg-elevated', '--ui-bg-soft', '--ui-bg-muted', '--ui-bg-overlay-strong',
  '--ui-text-primary', '--ui-text-secondary', '--ui-text-muted', '--ui-text-on-accent',
  '--ui-border', '--ui-border-strong',
  '--ui-accent', '--ui-accent-hover', '--ui-accent-soft',
  '--ui-success', '--ui-success-soft', '--ui-warning', '--ui-warning-soft', '--ui-danger', '--ui-info', '--ui-info-soft',
  '--ui-radius-sm', '--ui-radius-md', '--ui-radius-lg', '--ui-radius-xl',
  '--ui-shadow-1', '--ui-shadow-2', '--ui-shadow-3', '--ui-shadow-card',
]

test('tokens.css compat bridge maps every classic alias onto --qz-* tokens', () => {
  const css = readFileSync(TOKENS_CSS, 'utf8').replace(/\/\*[\s\S]*?\*\//g, '')
  const rules = parseRules(css)
  const light = declarationsFor(rules, LIGHT_SCOPE, null)
  const dark = declarationsFor(rules, DARK_SCOPE, null)

  const aliases = [...light].filter(([name]) => !name.startsWith('--qz-'))
  assert.deepEqual(
    BRIDGE_REQUIRED.filter((name) => !light.has(name)),
    [],
    `compat aliases missing from "${LIGHT_SCOPE}"`,
  )
  const offenders = aliases
    .filter(([, value]) => !/var\(--qz-/.test(value) || /#[0-9a-f]{3,8}\b|\b(?:rgba?|hsla?|oklch)\(|\d(?:px|rem|em)\b/i.test(value))
    .map(([name, value]) => `${name}: ${value}`)
  assert.deepEqual(offenders, [], 'every compat alias must resolve to --qz-* tokens only (no colour literal, no raw length)')
  assert.deepEqual(
    [...dark.keys()].filter((name) => !name.startsWith('--qz-')),
    [],
    `compat aliases are var() references and follow the theme by themselves; do not redeclare them on "${DARK_SCOPE}"`,
  )
  assert.equal(light.get('--ui-text-secondary'), 'var(--qz-text-secondary)', '.theme-prose body copy must read the template secondary text')
  assert.equal(light.get('--ui-success'), 'var(--qz-success)')
  assert.match(light.get('--ui-success-soft') ?? '', /var\(--qz-success\)/, '--ui-success-soft must be derived from --qz-success')
  assert.equal(light.get('--ui-warning'), 'var(--qz-warning)')
  assert.equal(light.get('--ui-danger'), 'var(--qz-danger)')
  assert.equal(light.get('--ui-info'), 'var(--qz-info)')
  assert.equal(light.get('--font-mono'), 'var(--qz-font-mono)')
})

// ---------------------------------------------------------------------------
// Base classes in styles/base.css consume tokens only.
//
// Colours come from --qz-* (color-mix() of tokens is fine), every font-size is a
// --qz-text-* step and every control height is a --qz-control-* step. The only fixed
// heights are glyphs inside badges / chips (12) and controls (16 / 18), badge (22), chip (24),
// the empty-state icon (40) and the 1px box of the visually-hidden utility (.qz-sr-only); the two
// relative fills are the layout root (100vh) and the brand glyph filling its host box (100%).
// QzButton.vue keeps just its label morph / loader CSS; its size and variant looks live
// in base.css as .qz-btn--*, so its <style> must not select .qz-btn at all.
// ---------------------------------------------------------------------------
const BASE_CSS = join(TEMPLATE_ROOT, 'styles', 'base.css')
const QZ_BUTTON_VUE = join(TEMPLATE_ROOT, 'components', 'QzButton.vue')
const BASE_FIXED_HEIGHTS = new Set(['1px', '12px', '16px', '18px', '22px', '24px', '40px', '100%', '100vh'])
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
      return !BASE_FIXED_HEIGHTS.has(trimmed) && !trimmed.startsWith('var(--qz-control-')
    })
    .map((match) => `base.css:${lineOf(css, match.index)} ${match[1]}: ${(match[2] ?? '').trim()}`)
  assert.deepEqual(
    heights,
    [],
    `height / min-height in base.css must be var(--qz-control-*) or one of ${[...BASE_FIXED_HEIGHTS].join(' / ')}`,
  )
})

test('QzButton.vue has no size/variant css of its own', () => {
  const sfc = readFileSync(QZ_BUTTON_VUE, 'utf8')
  // Comments may describe where the looks live; only rules count.
  const style = [...sfc.matchAll(/<style[^>]*>([\s\S]*?)<\/style>/g)]
    .map(([, body = '']) => body.replace(/\/\*[\s\S]*?\*\//g, ''))
    .join('\n')
  assert.ok(style.trim().length > 0, 'QzButton.vue should keep a <style> block for its label morph and loader')

  const leaked = QZ_BUTTON_STATE_CLASSES.filter((name) => style.includes(name))
  assert.deepEqual(leaked, [], 'QzButton size / variant styling belongs to .qz-btn--* in styles/base.css')
  assert.ok(!/\.qz-btn\b/.test(style), 'QzButton.vue must not select .qz-btn (or .qz-btn--*) itself; that CSS lives in styles/base.css')
})

// ---------------------------------------------------------------------------
// Page / component <style> blocks consume tokens only.
//
// Same colour and font-size rules as base.css, applied to every <style> block of every
// .vue file under templates/qiaozhi — no exemption list: no #hex / rgb() / rgba() / hsl() /
// oklch() literals (color-mix() of tokens is fine) and every font-size is a --qz-text-*
// step. Heights are not checked here: pages size art, icons and media freely.
// ---------------------------------------------------------------------------
function vueFiles(dir: string): string[] {
  const out: string[] = []
  for (const entry of readdirSync(dir, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))) {
    const full = join(dir, entry.name)
    if (entry.isDirectory()) {
      if (entry.name !== '__tests__') out.push(...vueFiles(full))
    } else if (entry.name.endsWith('.vue')) {
      out.push(full)
    }
  }
  return out
}

test('page styles consume tokens only', () => {
  const offenders: string[] = []
  let checked = 0
  for (const file of vueFiles(TEMPLATE_ROOT)) {
    const name = relative(TEMPLATE_ROOT, file).split(sep).join('/')
    const sfc = readFileSync(file, 'utf8')
    for (const block of sfc.matchAll(/<style[^>]*>([\s\S]*?)<\/style>/g)) {
      checked += 1
      const body = (block[1] ?? '').replace(/\/\*[\s\S]*?\*\//g, (comment) => comment.replace(/[^\n]/g, ' '))
      const blockLine = lineOf(sfc, block.index) // line of the <style> tag; body starts on the same line
      const at = (index: number) => `${name}:${blockLine + lineOf(body, index) - 1}`

      for (const match of body.matchAll(/#[0-9a-f]{3,8}\b|\b(?:rgba?|hsla?|oklch)\(/gi)) {
        offenders.push(`${at(match.index)} colour literal ${match[0]} — use a --qz-* token`)
      }
      for (const match of body.matchAll(/font-size\s*:\s*([^;}]+)/g)) {
        const value = (match[1] ?? '').trim()
        if (!value.startsWith('var(--qz-text-')) offenders.push(`${at(match.index)} font-size: ${value} — use a --qz-text-* token`)
      }
    }
  }
  assert.ok(checked > 0, 'expected at least one <style> block under templates/qiaozhi')
  assert.deepEqual(offenders, [], `${offenders.length} page-style declaration(s) bypass the qz tokens:\n${offenders.join('\n')}`)
})

// ---------------------------------------------------------------------------
// Type floor: nothing renders below 11px (--qz-text-2xs) anywhere in the template.
//
// Every <style> block of every .vue file and every sheet under styles/ is scanned, both
// for `font-size:` and for a `font:` shorthand that carries the size (`font: 700 10px/1 …`).
// Same regexes as the repo-wide `rg` check; comments are blanked first so a note about an
// old size does not count.
// ---------------------------------------------------------------------------
const FONT_SIZE_BELOW_FLOOR = /font-size:\s*(?:[0-9]|10)(?:\.\d+)?px/
const FONT_SHORTHAND_BELOW_FLOOR = /font:\s*[^;]*\b(?:[0-9]|10)(?:\.\d+)?px/

function blankComments(css: string): string {
  return css.replace(/\/\*[\s\S]*?\*\//g, (comment) => comment.replace(/[^\n]/g, ' '))
}

function styleSheets(): string[] {
  const dir = join(TEMPLATE_ROOT, 'styles')
  return readdirSync(dir)
    .filter((name) => name.endsWith('.css'))
    .sort()
    .map((name) => join(dir, name))
}

test('no font-size below 11px anywhere', () => {
  const offenders: string[] = []
  const report = (name: string, css: string, firstLine: number) => {
    blankComments(css)
      .split('\n')
      .forEach((line, index) => {
        const match = FONT_SIZE_BELOW_FLOOR.exec(line) ?? FONT_SHORTHAND_BELOW_FLOOR.exec(line)
        if (match) offenders.push(`${name}:${firstLine + index} ${match[0]}`)
      })
  }

  const sheets = styleSheets()
  assert.deepEqual(
    sheets.map((file) => relative(TEMPLATE_ROOT, file).split(sep).join('/')),
    ['styles/base.css', 'styles/layout.css', 'styles/tokens.css'],
    'styles/ holds exactly tokens / base / layout',
  )
  for (const file of sheets) report(relative(TEMPLATE_ROOT, file).split(sep).join('/'), readFileSync(file, 'utf8'), 1)

  let blocks = 0
  for (const file of vueFiles(TEMPLATE_ROOT)) {
    const name = relative(TEMPLATE_ROOT, file).split(sep).join('/')
    const sfc = readFileSync(file, 'utf8')
    for (const block of sfc.matchAll(/<style[^>]*>([\s\S]*?)<\/style>/g)) {
      blocks += 1
      report(name, block[1] ?? '', lineOf(sfc, block.index)) // body starts on the <style> tag's line
    }
  }
  assert.ok(blocks > 0, 'expected at least one <style> block under templates/qiaozhi')
  assert.deepEqual(offenders, [], `${offenders.length} font-size declaration(s) below the 11px floor (use var(--qz-text-2xs) or larger):\n${offenders.join('\n')}`)
})

// ---------------------------------------------------------------------------
// Every qz-* class a template puts on an element has a rule it can reach.
//
// This is the guard that replaces the old shared sheet: a class whose only rules lived
// there would now render unstyled, and nothing else fails loudly. Classes are collected
// from static class="…" attributes, from string literals inside :class="…" bindings and
// from static <Transition name> values (as `<name>-enter-active`). A rule source is one of:
//   - styles/base.css and styles/layout.css;
//   - the file's own <style> blocks, scoped or not;
//   - any other component's NON-scoped <style> block (shared vocabularies: PersonalCenter's
//     .qz-me-* for the me/* panels, QzButton, QzCategoryPanel, …);
//   - `:deep(…)` selectors in any other component's scoped block (slotted markup styled by
//     the component that renders the slot — QzDocPage's .qz-doc-prose).
// A scoped rule in some other file is deliberately NOT a source: Vue's scope attribute
// keeps it from ever matching. `is-*` state classes carry no prefix and are not checked.
// A selector only counts for the classes in its SUBJECT compound (`.qz-a .qz-b` styles
// .qz-b, not .qz-a); for `:deep(…)` the subject of the selector INSIDE the parentheses.
// ---------------------------------------------------------------------------
const STYLE_BLOCK = /<style([^>]*)>([\s\S]*?)<\/style>/g

// Every selector list in a sheet, split on top-level commas, at-rule preludes descended into.
function selectorsOf(css: string): string[] {
  const out: string[] = []
  for (const rule of parseRules(css.replace(/\/\*[\s\S]*?\*\//g, ''))) {
    if (rule.selector.startsWith('@')) continue
    let depth = 0
    let current = ''
    for (const ch of rule.selector) {
      if (ch === '(') depth += 1
      else if (ch === ')') depth -= 1
      if (ch === ',' && depth === 0) {
        out.push(current.trim())
        current = ''
      } else {
        current += ch
      }
    }
    if (current.trim()) out.push(current.trim())
  }
  return out
}

// Split a selector on top-level separators (combinators or commas), honouring (), [] and
// backslash escapes so `:is(.a, .b)`, `[data-x="a b"]` and `.z-\[100\]` stay intact.
function splitTopLevel(selector: string, isSeparator: (ch: string) => boolean): string[] {
  const parts: string[] = []
  let depth = 0
  let current = ''
  for (let i = 0; i < selector.length; i += 1) {
    const ch = selector[i] ?? ''
    if (ch === '\\') {
      current += ch + (selector[i + 1] ?? '')
      i += 1
      continue
    }
    if (ch === '(' || ch === '[') depth += 1
    else if (ch === ')' || ch === ']') depth -= 1
    if (depth === 0 && isSeparator(ch)) {
      if (current.trim()) parts.push(current.trim())
      current = ''
    } else {
      current += ch
    }
  }
  if (current.trim()) parts.push(current.trim())
  return parts
}

const isCombinator = (ch: string) => ch === ' ' || ch === '>' || ch === '+' || ch === '~' || ch === '\n' || ch === '\t'
const splitSelectorList = (selector: string) => splitTopLevel(selector, (ch) => ch === ',')

// Classes a single complex selector actually styles: the ones in its SUBJECT (last) compound.
// `.a .b` styles .b, never .a; `.a:not(.b)` styles .a only; `:is(.a, .b)`, `:where(…)` and
// `:deep(…)` in the subject compound contribute the subject classes of the selectors listed
// inside them. Pseudo-classes / pseudo-elements and attribute selectors never name a class.
function subjectClasses(selector: string): Set<string> {
  const compounds = splitTopLevel(selector, isCombinator)
  const subject = compounds[compounds.length - 1] ?? ''
  const out = new Set<string>()
  let i = 0
  while (i < subject.length) {
    const ch = subject[i] ?? ''
    if (ch === '\\') {
      i += 2
    } else if (ch === '.') {
      const match = /^[\w-]+/.exec(subject.slice(i + 1))
      if (match) out.add(match[0])
      i += 1 + (match?.[0].length ?? 0)
    } else if (ch === '[') {
      const close = subject.indexOf(']', i)
      i = close < 0 ? subject.length : close + 1
    } else if (ch === ':') {
      const name = /^::?[\w-]+/.exec(subject.slice(i))?.[0] ?? ':'
      i += name.length
      if (subject[i] === '(') {
        let depth = 0
        const start = i + 1
        for (; i < subject.length; i += 1) {
          if (subject[i] === '(') depth += 1
          else if (subject[i] === ')') depth -= 1
          if (depth === 0) break
        }
        const inner = subject.slice(start, i)
        i += 1
        if (name === ':is' || name === ':where' || name === ':deep' || name === ':global') {
          for (const part of splitSelectorList(inner)) subjectClasses(part).forEach((cls) => out.add(cls))
        }
      }
    } else {
      i += 1
    }
  }
  return out
}

// Classes a scoped block exposes to markup it does not own: only what sits INSIDE
// `:deep(…)` / `:global(…)` — and there, again, only the subject compound of each inner selector.
function piercedClasses(selector: string): Set<string> {
  const out = new Set<string>()
  const pattern = /:(?:deep|global)\(/g
  let match: RegExpExecArray | null
  while ((match = pattern.exec(selector))) {
    let depth = 1
    let end = match.index + match[0].length
    for (; end < selector.length && depth > 0; end += 1) {
      if (selector[end] === '(') depth += 1
      else if (selector[end] === ')') depth -= 1
    }
    const inner = selector.slice(match.index + match[0].length, end - 1)
    for (const part of splitSelectorList(inner)) subjectClasses(part).forEach((cls) => out.add(cls))
  }
  return out
}

const styledClasses = (selectors: string[], extract: (selector: string) => Set<string>) =>
  new Set(selectors.flatMap((selector) => [...extract(selector)]))

function templateClasses(sfc: string): Map<string, string> {
  const start = sfc.indexOf('<template>')
  const end = sfc.lastIndexOf('</template>')
  const template = start >= 0 && end > start ? sfc.slice(start, end) : ''
  const classes = new Map<string, string>() // class → how it was found
  const add = (cls: string, how: string) => {
    if (cls.startsWith('qz-') && !cls.includes('${')) classes.set(cls, how)
  }
  for (const [, attr = ''] of template.matchAll(/(?<![:\w-])class="([^"]*)"/g)) attr.split(/\s+/).forEach((cls) => add(cls, 'class'))
  for (const [, binding = ''] of template.matchAll(/:class="([^"]*)"/g)) {
    for (const [, single, backtick] of binding.matchAll(/'([^']*)'|`([^`]*)`/g)) (single ?? backtick ?? '').split(/\s+/).forEach((cls) => add(cls, ':class'))
  }
  for (const [, name = ''] of template.matchAll(/<Transition(?:Group)?\b[^>]*\sname="([^"]+)"/g)) add(`${name}-enter-active`, '<Transition name>')
  return classes
}

test('every qz-* class in a template has a rule in base / layout / a reachable component style', () => {
  const shared = styledClasses(
    [BASE_CSS, join(TEMPLATE_ROOT, 'styles', 'layout.css')].flatMap((file) => selectorsOf(readFileSync(file, 'utf8'))),
    subjectClasses,
  )
  const files = vueFiles(TEMPLATE_ROOT).map((file) => {
    const sfc = readFileSync(file, 'utf8')
    const blocks = [...sfc.matchAll(STYLE_BLOCK)].map(([, attrs = '', body = '']) => ({ scoped: /\sscoped\b/.test(attrs), selectors: selectorsOf(body) }))
    // Everything the file's own markup can rely on (scope attribute is always present there).
    const own = styledClasses(blocks.flatMap((block) => block.selectors), subjectClasses)
    // What other files' markup can rely on: unscoped blocks whole, scoped blocks only through :deep() / :global().
    const exported = new Set(blocks.flatMap((block) => [...styledClasses(block.selectors, block.scoped ? piercedClasses : subjectClasses)]))
    return { file, sfc, own, exported }
  })

  const unstyled: string[] = []
  let checked = 0
  for (const { file, sfc, own } of files) {
    const name = relative(TEMPLATE_ROOT, file).split(sep).join('/')
    const others = files.filter((other) => other.file !== file).map((other) => other.exported)
    for (const [cls, how] of templateClasses(sfc)) {
      checked += 1
      const styled = shared.has(cls) || own.has(cls) || others.some((set) => set.has(cls))
      if (!styled) unstyled.push(`${name}: .${cls} (${how})`)
    }
  }
  assert.ok(checked > 50, `expected to check qz-* classes across the template, checked ${checked}`)
  assert.deepEqual(
    unstyled,
    [],
    `${unstyled.length} qz-* class(es) have no rule in base.css / layout.css / their own <style> / another component's unscoped or :deep() style:\n${unstyled.join('\n')}`,
  )
})
