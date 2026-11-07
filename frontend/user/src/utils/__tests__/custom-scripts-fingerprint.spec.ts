// Verifies the fingerprint dedupe for custom scripts. On cold start
// stores/app.ts calls applyCustomScripts() twice: first with the inlined
// window.__PUBLIC_CONFIG__, then again after the background /public/config
// refetch. When the script set is unchanged the second call must not strip and
// re-inject the nodes (double-counted analytics, "Identifier has already been
// declared" for snippets with a top-level const).
//
// Project uses node:test (not vitest) + no DOM. The fingerprint tests are pure;
// the dedupe test installs a minimal in-memory `document` stand-in that only
// implements what applyCustomScripts / clearCustomScripts touch.
// Picked up by pnpm run test via the shim at tests/customScriptsFingerprint.test.ts.
import test from 'node:test'
import assert from 'node:assert/strict'
import {
  applyCustomScripts,
  clearCustomScripts,
  computeCustomScriptsFingerprint,
} from '../customScripts.ts'

const baseScripts = [
  { name: 'analytics', enabled: true, position: 'head', code: "console.log('analytics')" },
  { name: 'probe', enabled: true, position: 'body_end', code: 'const __qzProbe = (window.__qzProbe || 0) + 1' },
]

const cloneScripts = (scripts: unknown) => JSON.parse(JSON.stringify(scripts))

test('computeCustomScriptsFingerprint: same input yields the same string', () => {
  assert.equal(computeCustomScriptsFingerprint(baseScripts), computeCustomScriptsFingerprint(baseScripts))
  // inline config and background refetch are different object instances with equal content
  assert.equal(computeCustomScriptsFingerprint(cloneScripts(baseScripts)), computeCustomScriptsFingerprint(baseScripts))
})

test('computeCustomScriptsFingerprint: undefined, null and [] share one fingerprint', () => {
  const empty = computeCustomScriptsFingerprint([])
  assert.equal(computeCustomScriptsFingerprint(undefined), empty)
  assert.equal(computeCustomScriptsFingerprint(null), empty)
})

test('computeCustomScriptsFingerprint: changing any code field changes the fingerprint', () => {
  const original = computeCustomScriptsFingerprint(baseScripts)
  baseScripts.forEach((_, changedIndex) => {
    const mutated = baseScripts.map((item, index) =>
      index === changedIndex ? { ...item, code: `${item.code}; void 0` } : item,
    )
    assert.notEqual(
      computeCustomScriptsFingerprint(mutated),
      original,
      `code change at index ${changedIndex} must alter the fingerprint`,
    )
  })
})

test('computeCustomScriptsFingerprint: key order does not matter when content is equal', () => {
  const reordered = [
    { code: "console.log('analytics')", position: 'head', enabled: true, name: 'analytics' },
    { position: 'body_end', code: 'const __qzProbe = (window.__qzProbe || 0) + 1', name: 'probe', enabled: true },
  ]
  assert.equal(computeCustomScriptsFingerprint(reordered), computeCustomScriptsFingerprint(baseScripts))
})

// ---- applyCustomScripts dedupe (minimal document stand-in) ----

type FakeNode = {
  tagName: string
  type?: string
  text?: string
  attributes: Record<string, string>
  parentElement: FakeContainer | null
  setAttribute: (name: string, value: string) => void
}

type FakeContainer = {
  children: FakeNode[]
  appendChild: (node: FakeNode) => FakeNode
  removeChild: (node: FakeNode) => FakeNode
}

function installFakeDocument() {
  let appendCount = 0
  const makeContainer = (): FakeContainer => {
    const container: FakeContainer = {
      children: [],
      appendChild(node) {
        appendCount += 1
        node.parentElement = container
        container.children.push(node)
        return node
      },
      removeChild(node) {
        container.children = container.children.filter((child) => child !== node)
        node.parentElement = null
        return node
      },
    }
    return container
  }
  const head = makeContainer()
  const body = makeContainer()
  const fakeDocument = {
    head,
    body,
    createElement(tagName: string): FakeNode {
      const node: FakeNode = {
        tagName,
        attributes: {},
        parentElement: null,
        setAttribute(name, value) {
          node.attributes[name] = value
        },
      }
      return node
    },
    // every managed node matches the managed-attribute selector, so the selector is ignored
    querySelectorAll() {
      return [...head.children, ...body.children]
    },
  }
  ;(globalThis as any).document = fakeDocument
  return {
    get appendCount() {
      return appendCount
    },
    get mountedCount() {
      return head.children.length + body.children.length
    },
    uninstall() {
      delete (globalThis as any).document
    },
  }
}

test('applyCustomScripts: identical script set is not re-injected, changed set is', () => {
  const dom = installFakeDocument()
  try {
    clearCustomScripts() // reset module state left by earlier calls

    // cold start: inline window.__PUBLIC_CONFIG__
    applyCustomScripts(baseScripts)
    assert.equal(dom.appendCount, 2)
    assert.equal(dom.mountedCount, 2)

    // background /public/config refetch: same content, different object instances
    applyCustomScripts(cloneScripts(baseScripts))
    assert.equal(dom.appendCount, 2, 'unchanged scripts must not be re-injected')
    assert.equal(dom.mountedCount, 2)

    // admin changed the scripts → strip and re-inject the whole set
    const changed = [...baseScripts, { name: 'extra', enabled: true, position: 'head', code: 'window.__extra = 1' }]
    applyCustomScripts(changed)
    assert.equal(dom.appendCount, 5)
    assert.equal(dom.mountedCount, 3)

    // explicit clear resets the fingerprint so the same set can be applied again
    clearCustomScripts()
    assert.equal(dom.mountedCount, 0)
    applyCustomScripts(changed)
    assert.equal(dom.appendCount, 8)
    assert.equal(dom.mountedCount, 3)
  } finally {
    clearCustomScripts()
    dom.uninstall()
  }
})
