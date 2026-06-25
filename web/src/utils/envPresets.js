/** 统一环境变量：各业务线常用多服务 base_url（与 runner buildRunVars 对齐） */
const SERVICE_KEY_DEFS = [
  { key: 'base_url_edgen', label: 'Edgen API', hint: '主业务 API 域名' },
  { key: 'base_url_trex', label: 'Trex API', hint: '主业务 API 域名' },
  { key: 'base_url_quest', label: 'Quest', hint: '任务/活动服务' },
  { key: 'base_url_anchor', label: 'Anchor', hint: 'Anchor 服务' },
  { key: 'base_url_openreplay', label: 'OpenReplay', hint: '回放/监控' },
  { key: 'base_url_example', label: 'Example API', hint: '示例域名' }
]

export function serviceKeysForProduct(_productName) {
  return SERVICE_KEY_DEFS
}

export function parseVariables(raw) {
  if (!raw || raw === '{}') return {}
  try {
    const o = typeof raw === 'string' ? JSON.parse(raw) : raw
    if (o && typeof o === 'object' && !Array.isArray(o)) {
      const out = {}
      for (const [k, v] of Object.entries(o)) {
        out[k] = v == null ? '' : String(v)
      }
      return out
    }
  } catch {
    /* ignore */
  }
  return {}
}

export function stringifyVariables(obj) {
  const clean = {}
  for (const [k, v] of Object.entries(obj || {})) {
    if (!k || !String(k).trim()) continue
    clean[k] = v == null ? '' : String(v)
  }
  return JSON.stringify(clean)
}

export function tokenConfigured(vars) {
  return !!(vars?.token && String(vars.token).trim())
}

export function buildVariablesPayload(form) {
  const vars = { ...form.vars }
  vars.base_url = (form.base_url || '').replace(/\/+$/, '')
  if (form.token !== undefined) {
    vars.token = form.token || ''
  }
  for (const row of form.customRows || []) {
    const k = (row.key || '').trim()
    if (!k) continue
    vars[k] = row.value ?? ''
  }
  return stringifyVariables(vars)
}
