type ScriptPosition = 'head' | 'body_end'

type CustomScript = {
  name: string
  enabled: boolean
  position: ScriptPosition
  code: string
}

const managedScriptAttr = 'data-site-script-managed'
const managedScriptNameAttr = 'data-site-script-name'
const managedScriptGroupAttr = 'data-site-script-group'

const normalizeScriptPosition = (raw: unknown): ScriptPosition => {
  return raw === 'body_end' ? 'body_end' : 'head'
}

const normalizeScriptEnabled = (raw: unknown): boolean => {
  if (typeof raw === 'boolean') return raw
  if (typeof raw === 'number') return raw !== 0
  if (typeof raw === 'string') {
    const value = raw.trim().toLowerCase()
    return value === '1' || value === 'true' || value === 'yes' || value === 'on'
  }
  return false
}

export const normalizeCustomScripts = (raw: unknown): CustomScript[] => {
  if (!Array.isArray(raw)) return []

  return raw
    .map((item) => {
      if (!item || typeof item !== 'object') return null
      const value = item as Record<string, unknown>
      const code = typeof value.code === 'string' ? value.code.trim() : ''
      if (code === '') return null

      return {
        name: typeof value.name === 'string' ? value.name.trim() : '',
        enabled: normalizeScriptEnabled(value.enabled),
        position: normalizeScriptPosition(value.position),
        code,
      } as CustomScript
    })
    .filter((item): item is CustomScript => !!item && item.enabled)
}

function createManagedScriptNode(code: string, name: string, group: string): HTMLScriptElement {
  const node = document.createElement('script')
  node.type = 'text/javascript'
  node.text = code
  node.setAttribute(managedScriptAttr, '1')
  node.setAttribute(managedScriptNameAttr, name)
  node.setAttribute(managedScriptGroupAttr, group)
  return node
}

function createManagedNodesFromHTML(snippet: string, name: string, group: string): Node[] {
  const template = document.createElement('template')
  template.innerHTML = snippet
  const nodes: Node[] = []

  template.content.childNodes.forEach((child) => {
    if (child.nodeType !== Node.ELEMENT_NODE) {
      return
    }

    const element = child as Element
    if (element.tagName.toLowerCase() === 'script') {
      const sourceScript = element as HTMLScriptElement
      const scriptNode = document.createElement('script')
      Array.from(sourceScript.attributes).forEach((attr) => {
        scriptNode.setAttribute(attr.name, attr.value)
      })
      scriptNode.text = sourceScript.textContent || ''
      scriptNode.setAttribute(managedScriptAttr, '1')
      scriptNode.setAttribute(managedScriptNameAttr, name)
      scriptNode.setAttribute(managedScriptGroupAttr, group)
      nodes.push(scriptNode)
      return
    }

    const clone = element.cloneNode(true) as Element
    clone.setAttribute(managedScriptAttr, '1')
    clone.setAttribute(managedScriptNameAttr, name)
    clone.setAttribute(managedScriptGroupAttr, group)
    nodes.push(clone)
  })

  return nodes
}

export const computeCustomScriptsFingerprint = (rawScripts: unknown): string => {
  return JSON.stringify(normalizeCustomScripts(rawScripts))
}

// 当前已注入 DOM 的脚本集合指纹。stores/app.ts 冷启动时会先用内联的
// window.__PUBLIC_CONFIG__ 调一次 applyCustomScripts，后台 refetch 完整配置后再调一次；
// 内容相同就不能再删掉重插（统计双计数、顶层 const 重复声明报错）。
let appliedFingerprint = ''

export const clearCustomScripts = () => {
  appliedFingerprint = ''
  if (typeof document === 'undefined') return
  document.querySelectorAll(`script[${managedScriptAttr}="1"], [${managedScriptAttr}="1"]:not(script)`).forEach((node) => {
    node.parentElement?.removeChild(node)
  })
}

export const applyCustomScripts = (rawScripts: unknown) => {
  if (typeof document === 'undefined') return

  const fingerprint = computeCustomScriptsFingerprint(rawScripts)
  if (fingerprint === appliedFingerprint) return

  clearCustomScripts()
  appliedFingerprint = fingerprint
  const scripts = normalizeCustomScripts(rawScripts)

  scripts.forEach((item, index) => {
    const name = item.name || `script-${index + 1}`
    const group = `group-${index + 1}`
    const target = item.position === 'body_end' && document.body ? document.body : document.head
    const snippet = item.code.trim()
    if (snippet === '') return

    if (!snippet.includes('<')) {
      target.appendChild(createManagedScriptNode(snippet, name, group))
      return
    }

    const nodes = createManagedNodesFromHTML(snippet, name, group)
    if (nodes.length === 0) {
      target.appendChild(createManagedScriptNode(snippet, name, group))
      return
    }

    nodes.forEach((node) => target.appendChild(node))
  })
}
