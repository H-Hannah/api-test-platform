/** 环境管理默认 URL 变量（token 在测试数据中配置） */
export const ENV_MATRIX_ROWS = [
  { key: 'edgen_url', placeholder: 'https://...' },
  { key: 'quest_edgen_url', placeholder: 'https://...' },
  { key: 'trex_url', placeholder: 'https://...' },
  { key: 'quest_trex_url', placeholder: 'https://...' },
  { key: 'anchor_url', placeholder: 'https://...' }
]

export const ENV_ORDER = { BETA: 0, PRE: 1, PROD: 2 }

const LEGACY_ENV_KEY =
  /^(base_url(_|$)|token$|.*_token$|openreplay_url$|test$)/

export function isLegacyEnvKey(key) {
  const k = (key || '').trim()
  return !k || k === 'base_url' || LEGACY_ENV_KEY.test(k)
}

/** 从默认 URL 变量推导 DB base_url 列 */
export function deriveBaseURL(form, rowDefs = []) {
  const presetKeys = ENV_MATRIX_ROWS.map((r) => r.key)
  const byKey = new Map(rowDefs.map((r) => [(r.key || '').trim(), r]))
  for (const key of presetKeys) {
    const row = byKey.get(key)
    if (!row) continue
    const v = (form.values?.[row.id] || '').trim()
    if (v) return v.replace(/\/+$/, '')
  }
  for (const row of rowDefs) {
    const k = (row.key || '').trim()
    const v = (form.values?.[row.id] || '').trim()
    if (!v || !k.endsWith('_url')) continue
    return v.replace(/\/+$/, '')
  }
  return ''
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

export function buildVariablesPayload(form, rowDefs = []) {
  const vars = {}
  for (const row of rowDefs) {
    const k = (row.key || '').trim()
    if (!k || isLegacyEnvKey(k)) continue
    vars[k] = form.values?.[row.id] ?? ''
  }
  return stringifyVariables(vars)
}
