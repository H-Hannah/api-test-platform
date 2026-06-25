import { ElNotification } from 'element-plus'

/** AI 入库结果 */
export function notifyIngestResult(result, mode) {
  if (result?.scenario) {
    const sc = result.scenario
    ElNotification({
      title: '场景已保存',
      message: `${sc.name || '场景'} · ${sc.step_count || 0} 步${sc.folder_path ? ' · ' + sc.folder_path : ''}`,
      type: 'success',
      duration: 5000,
      position: 'top-right'
    })
    return
  }
  const apis = result?.apis || []
  if (!apis.length) {
    ElNotification({
      title: '入库完成',
      message: '未返回接口明细，请到 Web 管理端查看',
      type: 'warning',
      duration: 4000,
      position: 'top-right'
    })
    return
  }
  const preview = apis
    .slice(0, 3)
    .map((a) => a.name || a.folder_path || `#${a.id}`)
    .join('、')
  const suffix = apis.length > 3 ? ` 等共 ${apis.length} 个` : ` · 共 ${apis.length} 个`
  ElNotification({
    title: mode === 'scenario' ? '场景已保存' : '接口已入库',
    message: preview + suffix,
    type: 'success',
    duration: 5500,
    position: 'top-right'
  })
}

/** 单次执行结果 */
export function notifyRunResult(run, envName) {
  if (!run) return
  const passed = run.status === 'passed'
  const step = run.steps?.[0]
  const failedAssert = step?.assertion_results
    ? parseAssertionResults(step.assertion_results).filter((a) => !a.passed)
    : []
  let message = envName ? `环境 ${envName} · ` : ''
  if (step?.duration_ms != null) message += `${step.duration_ms}ms`
  if (failedAssert.length) {
    message += ` · ${failedAssert.length} 项断言失败`
  } else if (step?.error_message) {
    message += ` · ${step.error_message}`
  }
  ElNotification({
    title: passed ? '执行通过' : '执行失败',
    message: message || run.summary || run.status,
    type: passed ? 'success' : 'error',
    duration: passed ? 4000 : 8000,
    position: 'top-right'
  })
}

export function parseAssertionResults(raw) {
  if (!raw) return []
  try {
    const data = typeof raw === 'string' ? JSON.parse(raw) : raw
    return Array.isArray(data) ? data : []
  } catch {
    return []
  }
}

export function parseSnapshot(raw) {
  if (!raw) return null
  try {
    return typeof raw === 'string' ? JSON.parse(raw) : raw
  } catch {
    return { raw }
  }
}
