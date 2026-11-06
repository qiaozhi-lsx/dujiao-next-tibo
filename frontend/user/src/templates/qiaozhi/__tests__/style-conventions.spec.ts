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
