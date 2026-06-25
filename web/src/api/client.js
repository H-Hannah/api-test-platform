import axios from 'axios'
import { ElMessage } from 'element-plus'

const TOKEN_KEY = 'platform_api_token'

// 开发环境自动填入 Token，避免首次打开即 401
if (import.meta.env.DEV && import.meta.env.VITE_API_TOKEN && !localStorage.getItem(TOKEN_KEY)) {
  localStorage.setItem(TOKEN_KEY, import.meta.env.VITE_API_TOKEN)
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, (token || '').trim())
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

/** 登录页校验令牌（不走全局拦截器，避免未登录时误跳转） */
export async function verifyToken(token) {
  try {
    const res = await axios.get('/api/v1/products', {
      headers: { Authorization: `Bearer ${token.trim()}` },
      timeout: 15000
    })
    return res.data
  } catch (err) {
    let msg = err.response?.data?.error || err.message || '验证失败'
    if (!err.response && (err.code === 'ECONNREFUSED' || /Network Error/i.test(msg))) {
      msg = '无法连接后端（请先运行 go run ./cmd/server）'
    }
    if (err.response?.status === 401 || err.response?.status === 403) {
      msg = '令牌无效，请检查后重试'
    }
    throw new Error(msg)
  }
}

const http = axios.create({
  baseURL: '',
  timeout: 120000
})

http.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const status = err.response?.status
    let msg = err.response?.data?.error || err.message || '请求失败'
    if (!err.response && (err.code === 'ECONNREFUSED' || /Network Error/i.test(msg))) {
      msg = '无法连接后端（请先运行 go run ./cmd/server）'
    }
    if (status === 401 || status === 403) {
      msg = '令牌无效或已过期，请重新登录'
      clearToken()
      if (!window.location.pathname.startsWith('/login')) {
        const redirect = encodeURIComponent(window.location.pathname + window.location.search)
        window.location.href = `/login?redirect=${redirect}`
      }
    }
    ElMessage.error(msg)
    return Promise.reject(new Error(msg))
  }
)

export const api = {
  listProducts: () => http.get('/api/v1/products'),
  folderTree: () => http.get('/api/v1/folders/tree'),
  listAPIs: (params = {}) => http.get('/api/v1/apis', { params }),
  getCoverage: (params = {}) => http.get('/api/v1/products/1/coverage', { params }),
  patchAPIMeta: (id, body) => http.patch(`/api/v1/apis/${id}/meta`, body),
  generateTestData: (body) => http.post('/api/v1/ai/testdata/generate', body),
  importTestData: (body) => http.post('/api/v1/testdata/import', body),
  listTestDatasets: (params = {}) => http.get('/api/v1/testdata/datasets', { params }),
  getTestDataset: (id) => http.get(`/api/v1/testdata/datasets/${id}`),
  deleteTestDataset: (id) => http.delete(`/api/v1/testdata/datasets/${id}`),
  importEnvVarKeys: (envId, body) =>
    http.post(`/api/v1/environments/${envId}/import-var-keys`, body),
  impactAnalyze: (body) => http.post('/api/v1/impact/analyze', body),
  impactPostMRComment: (body) => http.post('/api/v1/impact/post-mr-comment', body),
  impactPreviewMRComment: (body) => http.post('/api/v1/impact/preview-mr-comment', body),
  impactRunPlan: (body) => http.post('/api/v1/impact/run-plan', body),
  loadRequirementPackage: (params) =>
    http.get('/api/v1/docs/requirement-package', { params }),
  loadTestCases: (params) =>
    http.get('/api/v1/docs/testcases', { params }),
  listTestDocsBranches: () => http.get('/api/v1/docs/testcases/branches'),
  listTestDocsCatalog: (params = {}) =>
    http.get('/api/v1/docs/testcases/catalog', { params }),
  getAPI: (id) => http.get(`/api/v1/apis/${id}`),
  deleteAPI: (id) => http.delete(`/api/v1/apis/${id}`),
  runAPI: (id, envId, datasetId) =>
    http.post(`/api/v1/apis/${id}/run`, { env_id: envId, dataset_id: datasetId || 0 }),
  listAPIRuns: (apiId, limit = 20) =>
    http.get(`/api/v1/apis/${apiId}/runs`, { params: { limit } }),
  listScenarios: () => http.get('/api/v1/scenarios'),
  getScenario: (id) => http.get(`/api/v1/scenarios/${id}`),
  runScenario: (id, envId) => http.post(`/api/v1/scenarios/${id}/run`, { env_id: envId }),
  listEnvironments: () => http.get('/api/v1/environments'),
  getEnvironment: (id) => http.get(`/api/v1/environments/${id}`),
  createEnvironment: (body) => http.post('/api/v1/environments', body),
  updateEnvironment: (id, body) => http.put(`/api/v1/environments/${id}`, body),
  deleteEnvironment: (id) => http.delete(`/api/v1/environments/${id}`),
  listRuns: () => http.get('/api/v1/runs'),
  getRun: (id) => http.get(`/api/v1/runs/${id}`)
}
