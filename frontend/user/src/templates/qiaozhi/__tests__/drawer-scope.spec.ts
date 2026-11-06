// Regression guard for the mobile category drawer in QzCategoryPanel.vue.
//
// The drawer is `<Teleport to="body">`, so it renders outside the `.qiaozhi-scope`
// layout root. Its legacy.css rules therefore have to start from `body.qiaozhi-tokens`
// (the class QiaozhiLayout.vue adds to <body> on mount) — a plain
// `.qiaozhi-scope .qz-drawer…` prefix never matches and the drawer degrades to a
// column of unstyled buttons.
//
// Project uses node:test (no DOM, no vitest), so this is a static check: collect the
// static classes used inside the Teleport block and assert each one has at least one
// legacy.css selector reachable from body scope. Picked up by pnpm run test via the shim
// at tests/qiaozhiDrawerScope.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const read = (relative: string) => readFileSync(new URL(relative, import.meta.url), 'utf8')

const BODY_SCOPE = 'body.qiaozhi-tokens'

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

const startsFromBodyScope = (selector: string) =>
  selector.startsWith(BODY_SCOPE) ||
  (selector.startsWith(':is(') && selector.slice(0, selector.indexOf(')')).includes(BODY_SCOPE))

const mentionsClass = (selector: string, cls: string) =>
  new RegExp(`\\.${cls.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}(?![\\w-])`).test(selector)

test('teleported drawer markup is styled from body scope', () => {
  const classes = teleportedClasses(read('../components/QzCategoryPanel.vue'))
  assert.ok(classes.includes('qz-drawer'), `expected drawer classes inside <Teleport>, got: ${classes.join(', ')}`)

  const bodyScoped = selectors(read('../styles/legacy.css')).filter(startsFromBodyScope)
  const unreachable = classes.filter((cls) => !bodyScoped.some((selector) => mentionsClass(selector, cls)))
  assert.deepEqual(unreachable, [], `drawer classes with no body-scoped rule in legacy.css: ${unreachable.join(', ')}`)
})
