// Regression guard for the mobile category drawer in QzCategoryPanel.vue.
//
// The drawer is `<Teleport to="body">`, so it renders outside the `.qiaozhi-scope`
// layout root. Every rule that styles it therefore has to start from `body.qiaozhi-tokens`
// (the class QiaozhiLayout.vue adds to <body> on mount) — a plain
// `.qiaozhi-scope .qz-drawer…` prefix never matches and the drawer degrades to a
// column of unstyled buttons.
//
// The drawer's rules live in the component's own (unscoped) <style> block, written with
// the same `:is(.qiaozhi-scope, :where(body.qiaozhi-tokens))` prefix base.css uses; the
// primitives it composes (.qz-eyebrow, .qz-pill, .qz-badge, .qz-icon-btn) and the shared
// brand glyph (.qz-brand-icon) come from styles/base.css.
//
// Project uses node:test (no DOM, no vitest), so this is a static check: collect the
// static classes used inside the Teleport block and assert each one has at least one
// selector reachable from body scope in the component's own styles or base.css. Picked up
// by pnpm run test via the shim at tests/qiaozhiDrawerScope.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const read = (relative: string) => readFileSync(new URL(relative, import.meta.url), 'utf8')

const BODY_SCOPE = 'body.qiaozhi-tokens'
const PANEL_SFC = '../components/QzCategoryPanel.vue'
const STYLE_SOURCES = ['../styles/base.css']

// Static `class="…"` attributes only; Vue `:class="{ … }"` bindings carry runtime state classes.
const STATIC_CLASS_ATTR = /(?<![:\w-])class="([^"]+)"/g

// `is-*` state classes and Vue <Transition name="qz-drawer"> classes are not drawer markup.
const isStateOrTransitionClass = (cls: string) => cls.startsWith('is-') || /^qz-drawer-(enter|leave)/.test(cls)

function teleportedClasses(sfc: string): string[] {
  const start = sfc.indexOf('<Teleport to="body">')
  const end = sfc.indexOf('</Teleport>', start)
  assert.ok(start >= 0 && end > start, 'QzCategoryPanel.vue should teleport the drawer to body')
  const classes = new Set<string>()
  for (const [, attr = ''] of sfc.slice(start, end).matchAll(STATIC_CLASS_ATTR)) {
    for (const cls of attr.split(/\s+/)) {
      if (cls && !isStateOrTransitionClass(cls)) classes.add(cls)
    }
  }
  return [...classes]
}

// Every <style> block of an SFC (the drawer rules must not be `scoped` either way — see header).
const styleBlocks = (sfc: string) => [...sfc.matchAll(/<style[^>]*>([\s\S]*?)<\/style>/g)].map(([, body = '']) => body).join('\n')

// Split a selector list on top-level commas only — `:is(a, b) c` is a single selector.
function splitSelectorList(prelude: string): string[] {
  const out: string[] = []
  let depth = 0
  let current = ''
  for (const ch of prelude) {
    if (ch === '(') depth++
    else if (ch === ')') depth--
    if (ch === ',' && depth === 0) {
      out.push(current.trim())
      current = ''
    } else {
      current += ch
    }
  }
  if (current.trim()) out.push(current.trim())
  return out
}

// Every `prelude {` in the stylesheet, minus at-rule preludes (@media, @keyframes…).
function selectors(css: string): string[] {
  const out: string[] = []
  for (const [, rawPrelude = ''] of css.replace(/\/\*[\s\S]*?\*\//g, '').matchAll(/([^{}]+)\{/g)) {
    const prelude = rawPrelude.trim()
    if (!prelude || prelude.startsWith('@')) continue
    out.push(...splitSelectorList(prelude))
  }
  return out
}

// `.dark` may lead (theme ancestor on <html>); after that the selector must be rooted at body scope.
const startsFromBodyScope = (selector: string) => {
  const rooted = selector.replace(/^\.dark\s+/, '')
  return (
    rooted.startsWith(BODY_SCOPE) ||
    (rooted.startsWith(':is(') && rooted.slice(0, rooted.indexOf(')')).includes(BODY_SCOPE))
  )
}

const mentionsClass = (selector: string, cls: string) =>
  new RegExp(`\\.${cls.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}(?![\\w-])`).test(selector)

test('teleported drawer markup is styled from body scope', () => {
  const sfc = read(PANEL_SFC)
  const classes = teleportedClasses(sfc)
  assert.ok(classes.includes('qz-drawer'), `expected drawer classes inside <Teleport>, got: ${classes.join(', ')}`)

  const componentCss = styleBlocks(sfc)
  assert.ok(componentCss.includes('.qz-drawer'), 'QzCategoryPanel.vue should style its drawer in its own <style> block')
  assert.ok(!/<style[^>]*\sscoped/.test(sfc), 'QzCategoryPanel.vue styles must stay unscoped: the drawer rules are shared with the desktop panel and prefixed by scope')

  const bodyScoped = [componentCss, ...STYLE_SOURCES.map(read)].flatMap((css) => selectors(css)).filter(startsFromBodyScope)
  const unreachable = classes.filter((cls) => !bodyScoped.some((selector) => mentionsClass(selector, cls)))
  assert.deepEqual(unreachable, [], `drawer classes with no body-scoped rule in QzCategoryPanel.vue / base.css: ${unreachable.join(', ')}`)

  // The regression itself: a drawer rule rooted at `.qiaozhi-scope` alone can never match the teleported DOM.
  const drawerOnlyScoped = selectors(componentCss).filter(
    (selector) => /\.qz-drawer(?![\w-])/.test(selector) && selector.startsWith('.qiaozhi-scope'),
  )
  assert.deepEqual(drawerOnlyScoped, [], `drawer rules rooted at .qiaozhi-scope only: ${drawerOnlyScoped.join(' | ')}`)
})
