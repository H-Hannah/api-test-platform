<template>
  <div class="workbench">
    <aside class="tree-panel panel-card">
      <div class="tree-head">
        <span>模块目录</span>
        <el-button text type="primary" size="small" @click="selectFolder(null, '全部接口')">全部</el-button>
      </div>
      <el-tree
        v-loading="treeLoading"
        :data="treeData"
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
        highlight-current
        default-expand-all
        @node-click="(node) => selectFolder(node.id === 0 ? null : node.id, node.name)"
      />
    </aside>

    <section class="list-panel panel-card">
      <div class="list-toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索名称、路径、模块"
          clearable
          class="search-input"
          :prefix-icon="Search"
        />
        <el-select v-model="methodFilter" placeholder="方法" clearable class="method-filter">
          <el-option v-for="m in methodOptions" :key="m" :label="m" :value="m" />
        </el-select>
        <el-select v-model="gapFilter" placeholder="场景缺口" clearable class="gap-filter" @change="loadApis">
          <el-option label="未就绪" value="not_ready" />
          <el-option label="已就绪" value="ready" />
          <el-option label="缺 US" value="no_us" />
          <el-option label="缺 TC" value="no_tc" />
          <el-option label="缺 BDD(追溯)" value="no_bdd" />
          <el-option label="缺断言" value="no_assert" />
        </el-select>
        <el-input v-model="mrFilter" placeholder="MR 标签" clearable class="mr-filter" @change="loadApis" />
        <span class="toolbar-meta">
          <span class="count-badge">{{ filteredApis.length }}</span>
          <span class="count-label">条接口</span>
          <span v-if="folderLabel" class="folder-tag">{{ folderLabel }}</span>
        </span>
      </div>

      <div class="table-wrap">
        <el-table
          v-loading="listLoading"
          :data="filteredApis"
          stripe
          highlight-current-row
          table-layout="fixed"
          class="api-table"
          empty-text="暂无接口，请用插件录制并入库"
          :row-class-name="rowClassName"
          @current-change="onRowSelect"
          @row-dblclick="(row) => runApi(row)"
        >
          <el-table-column label="方法" width="70" align="center">
            <template #default="{ row }">
              <MethodBadge :method="row.method" />
            </template>
          </el-table-column>
          <el-table-column label="接口" width="200" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="name-cell">
                <span class="name-text">{{ row.name }}</span>
                <span v-if="row.folder_path" class="name-sub">{{ row.folder_path }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="路径" width="260" show-overflow-tooltip>
            <template #default="{ row }">
              <code class="path-cell">{{ row.path }}</code>
            </template>
          </el-table-column>
          <el-table-column label="场景" width="72" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.scenario_ready" type="success" size="small">就绪</el-tag>
              <el-tag v-else type="warning" size="small">缺口</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="结果" width="100" align="center">
            <template #default="{ row }">
              <span v-if="runningApiId === row.id" class="run-status run-status--loading">执行中</span>
              <button
                v-else-if="lastRunMap[row.id]?.runId"
                type="button"
                class="run-status run-status--link"
                :class="lastRunMap[row.id].status === 'passed' ? 'run-status--pass' : 'run-status--fail'"
                @click.stop="openRunById(lastRunMap[row.id].runId, row)"
              >
                <template v-if="lastRunMap[row.id].status === 'passed'">
                  <span class="result-ico">✅</span><span>成功</span>
                </template>
                <template v-else>
                  <span class="result-ico">❌</span><span>失败</span>
                </template>
              </button>
              <span v-else class="run-status run-status--none">—</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100" align="center" class-name="col-ops">
            <template #default="{ row }">
              <div class="row-actions">
                <el-tooltip content="执行" placement="top">
                  <el-button
                    plain
                    size="small"
                    circle
                    type="primary"
                    :icon="VideoPlay"
                    :loading="runningApiId === row.id"
                    :disabled="!envId"
                    @click.stop="runApi(row)"
                  />
                </el-tooltip>
                <el-tooltip content="删除" placement="top">
                  <el-button
                    type="danger"
                    plain
                    size="small"
                    circle
                    :icon="Delete"
                    @click.stop="removeApi(row)"
                  />
                </el-tooltip>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </section>

    <aside class="detail-panel panel-card" :class="{ 'detail-panel--empty': !selectedApi }">
      <template v-if="selectedApi && detail">
        <div class="detail-toolbar">
          <div class="detail-title">
            <MethodBadge :method="detail.method" />
            <h3>{{ detail.name }}</h3>
          </div>
          <div class="detail-run-bar">
            <el-select
              v-model="selectedDatasetId"
              class="dataset-select"
              placeholder="测试数据（可选）"
              clearable
              size="small"
              :loading="datasetsLoading"
            >
              <el-option
                v-for="ds in apiDatasets"
                :key="ds.id"
                :label="`${ds.dataset_key} · ${ds.name}`"
                :value="ds.id"
              />
            </el-select>
            <el-tooltip content="执行" placement="bottom">
              <el-button
                type="primary"
                size="small"
                circle
                :icon="VideoPlay"
                :loading="runningApiId === selectedApi.id"
                :disabled="!envId"
                @click="runApi(selectedApi)"
              />
            </el-tooltip>
          </div>
        </div>
        <p class="detail-path-line">
          <code>{{ detail.path }}</code>
        </p>
        <p v-if="detail.full_url_template" class="url-template">
          <span class="label">URL</span>
          <code>{{ detail.full_url_template }}</code>
        </p>

        <el-tabs v-model="activeTab" class="detail-tabs">
          <el-tab-pane label="概览" name="overview">
            <dl class="meta-list">
              <dt>模块</dt><dd>{{ detail.folder_path || '—' }}</dd>
              <dt>说明</dt><dd>{{ detail.description || '—' }}</dd>
              <dt>AI 备注</dt><dd>{{ detail.ai_remark || '—' }}</dd>
              <dt>更新</dt><dd>{{ detail.updated_at || '—' }}</dd>
            </dl>
            <section v-if="runHistory.length" class="run-history">
              <h4 class="section-h4">执行历史</h4>
              <p class="history-hint">点击记录查看断言、请求与响应详情</p>
              <ul class="history-list">
                <li
                  v-for="r in runHistory"
                  :key="r.id"
                  class="history-item"
                  @click="openRunById(r.id)"
                >
                  <span
                    class="history-status"
                    :class="r.status === 'passed' ? 'run-status--pass' : 'run-status--fail'"
                  >
                    {{ r.status === 'passed' ? '通过' : '失败' }}
                  </span>
                  <span class="history-env">{{ envNameById(r.env_id) }}</span>
                  <span class="history-time">{{ r.started_at }}</span>
                </li>
              </ul>
            </section>
            <p v-else class="history-empty">暂无执行记录，点击「执行」后将出现在此处</p>
          </el-tab-pane>
          <el-tab-pane label="请求" name="request">
            <h4>Headers</h4>
            <pre class="json-block">{{ pretty(detail.headers) }}</pre>
            <h4>Body</h4>
            <pre class="json-block">{{ pretty(detail.body) }}</pre>
          </el-tab-pane>
          <el-tab-pane label="断言" name="assertions">
            <el-table :data="detail.assertions || []" size="small" border>
              <el-table-column prop="type" label="类型" width="96" />
              <el-table-column prop="expression" label="表达式" min-width="100" show-overflow-tooltip />
              <el-table-column prop="operator" label="运算" width="64" />
              <el-table-column prop="expected" label="期望" min-width="72" show-overflow-tooltip />
            </el-table>
          </el-tab-pane>
          <el-tab-pane label="追溯" name="trace">
            <el-form label-width="88px" class="trace-form">
              <el-form-item label="User Story">
                <el-input v-model="metaForm.user_story" placeholder="US-123 登录后查看绑定" />
              </el-form-item>
              <el-form-item label="测试用例 TC" required>
                <el-input
                  v-model="metaForm.tc_ref"
                  placeholder="TC001 | REQ-xxx | @line-trend"
                />
              </el-form-item>
              <el-form-item label="BDD 追溯（可选）">
                <el-input
                  v-model="metaForm.bdd_ref"
                  placeholder="01-chart.feature:@line-trend"
                />
              </el-form-item>
              <el-form-item label="MR 标签">
                <el-input v-model="metaForm.mr_tags" placeholder="MR-128,MR-130（逗号分隔）" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="metaSaving" @click="saveMeta">保存追溯信息</el-button>
              </el-form-item>
            </el-form>
            <p class="trace-hint">
              场景就绪 = 已填 US + TC + 至少 1 条断言。
              <router-link to="/impact">精准测试</router-link>
            </p>
          </el-tab-pane>
        </el-tabs>
      </template>

      <div v-else class="detail-empty">
        <el-empty description="点击列表中的接口查看详情，或双击行快速执行">
          <template v-if="!envId" #description>
            <p>请先在顶部选择<strong>运行环境</strong></p>
            <p class="hint-sub">再点击「执行」调试单接口</p>
          </template>
        </el-empty>
      </div>
    </aside>

    <RunResultDialog
      v-model="runDialogOpen"
      :run="lastRun"
      :env-name="runDialogEnvName || currentEnvName"
      :api-name="selectedApi?.name"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/composables/useAppStore'
import { Search, VideoPlay, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api/client'
import MethodBadge from '@/components/MethodBadge.vue'
import RunResultDialog from '@/components/RunResultDialog.vue'
import { notifyRunResult } from '@/utils/notify'

const route = useRoute()
const { envId, environments } = useAppStore()

const methodOptions = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']
const gapFilter = ref('')
const mrFilter = ref('')

const treeLoading = ref(false)
const listLoading = ref(false)
const treeData = ref([])
const apis = ref([])
const folderId = ref(null)
const folderLabel = ref('全部接口')
const keyword = ref('')
const methodFilter = ref('')
const selectedApi = ref(null)
const detail = ref(null)
const activeTab = ref('overview')
const runningApiId = ref(null)
const runDialogOpen = ref(false)
const lastRun = ref(null)
const lastRunMap = ref({})
const runHistory = ref([])
const runDialogEnvName = ref('')
const metaForm = ref({ user_story: '', bdd_ref: '', tc_ref: '', mr_tags: '' })
const metaSaving = ref(false)
const apiDatasets = ref([])
const selectedDatasetId = ref(null)
const datasetsLoading = ref(false)

const currentEnvName = computed(() => {
  const e = environments.value.find((x) => x.id === envId.value)
  return e?.name || ''
})

const filteredApis = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  const mf = methodFilter.value
  return apis.value.filter((a) => {
    if (mf && (a.method || '').toUpperCase() !== mf) return false
    if (!k) return true
    return (
      (a.name || '').toLowerCase().includes(k) ||
      (a.path || '').toLowerCase().includes(k) ||
      (a.folder_path || '').toLowerCase().includes(k)
    )
  })
})

function rowClassName({ row }) {
  return selectedApi.value?.id === row.id ? 'row-current' : ''
}

function pretty(val) {
  if (!val) return '（空）'
  if (typeof val === 'string') {
    try {
      return JSON.stringify(JSON.parse(val), null, 2)
    } catch {
      return val
    }
  }
  return JSON.stringify(val, null, 2)
}

async function loadTree() {
  treeLoading.value = true
  try {
    const nodes = await api.folderTree()
    treeData.value = [{ id: 0, name: '全部接口', children: nodes }]
  } finally {
    treeLoading.value = false
  }
}

async function syncLastRunMapFromReports() {
  try {
    const runs = await api.listRuns()
    const map = { ...lastRunMap.value }
    for (const r of runs) {
      if (r.api_id && !map[r.api_id]) {
        map[r.api_id] = { status: r.status, runId: r.id }
      }
    }
    lastRunMap.value = map
  } catch { /* */ }
}

function listParams() {
  const p = {}
  if (folderId.value) p.folder_id = folderId.value
  if (gapFilter.value) p.gap = gapFilter.value
  if (mrFilter.value.trim()) p.mr_tag = mrFilter.value.trim()
  return p
}

async function loadApis() {
  listLoading.value = true
  try {
    apis.value = await api.listAPIs(listParams())
    await syncLastRunMapFromReports()
    await tryFocusApiFromQuery()
  } finally {
    listLoading.value = false
  }
}

async function tryFocusApiFromQuery() {
  const qid = route.query.api_id
  if (!qid) return
  const row = apis.value.find((a) => String(a.id) === String(qid))
  if (row) {
    await focusRow(row)
    activeTab.value = 'trace'
  }
}

function selectFolder(id, label) {
  folderId.value = id || null
  folderLabel.value = label || '全部接口'
  loadApis()
}

async function focusRow(row) {
  selectedApi.value = row
  await refreshDetail()
}

async function onRowSelect(row) {
  if (!row) return
  await focusRow(row)
}

function envNameById(id) {
  const e = environments.value.find((x) => x.id === id)
  return e?.name || `#${id}`
}

async function loadRunHistory() {
  if (!selectedApi.value?.id) {
    runHistory.value = []
    return
  }
  try {
    runHistory.value = await api.listAPIRuns(selectedApi.value.id)
    if (runHistory.value.length) {
      const latest = runHistory.value[0]
      lastRunMap.value = {
        ...lastRunMap.value,
        [selectedApi.value.id]: { status: latest.status, runId: latest.id }
      }
    }
  } catch {
    runHistory.value = []
  }
}

async function openRunById(runId, row) {
  if (row && selectedApi.value?.id !== row.id) {
    await focusRow(row)
  }
  const run = await api.getRun(runId)
  lastRun.value = run
  runDialogEnvName.value = envNameById(run.env_id)
  runDialogOpen.value = true
}

async function refreshDetail() {
  if (!selectedApi.value) return
  detail.value = await api.getAPI(selectedApi.value.id)
  metaForm.value = {
    user_story: detail.value.user_story || '',
    bdd_ref: detail.value.bdd_ref || '',
    tc_ref: detail.value.tc_ref || '',
    mr_tags: detail.value.mr_tags || ''
  }
  applyQuerySuggestions()
  await loadRunHistory()
  await loadApiDatasets()
}

async function loadApiDatasets() {
  if (!selectedApi.value?.id) {
    apiDatasets.value = []
    selectedDatasetId.value = null
    return
  }
  datasetsLoading.value = true
  try {
    apiDatasets.value = await api.listTestDatasets({ api_id: selectedApi.value.id })
    if (selectedDatasetId.value && !apiDatasets.value.some((d) => d.id === selectedDatasetId.value)) {
      selectedDatasetId.value = null
    }
  } catch {
    apiDatasets.value = []
  } finally {
    datasetsLoading.value = false
  }
}

async function saveMeta() {
  if (!selectedApi.value?.id) return
  metaSaving.value = true
  try {
    detail.value = await api.patchAPIMeta(selectedApi.value.id, { ...metaForm.value })
    metaForm.value = {
      user_story: detail.value.user_story || '',
      bdd_ref: detail.value.bdd_ref || '',
      tc_ref: detail.value.tc_ref || '',
      mr_tags: detail.value.mr_tags || ''
    }
    ElMessage.success('追溯信息已保存')
    await loadApis()
    const row = apis.value.find((a) => a.id === selectedApi.value.id)
    if (row) selectedApi.value = row
  } finally {
    metaSaving.value = false
  }
}

async function runApi(row) {
  if (!row?.id) return
  if (!envId.value) {
    ElMessage.warning('请先在顶部选择运行环境（BETA / PRE / PROD）')
    return
  }
  runningApiId.value = row.id
  try {
    const run = await api.runAPI(row.id, envId.value, selectedDatasetId.value || 0)
    lastRun.value = run
    lastRunMap.value = { ...lastRunMap.value, [row.id]: { status: run.status, runId: run.id } }
    runDialogEnvName.value = currentEnvName.value
    if (selectedApi.value?.id !== row.id) {
      await focusRow(row)
    } else {
      await loadRunHistory()
    }
    runDialogOpen.value = true
    notifyRunResult(run, currentEnvName.value)
    if (run.status !== 'passed') {
      activeTab.value = 'assertions'
    }
  } finally {
    runningApiId.value = null
  }
}

async function removeApi(row) {
  await ElMessageBox.confirm(`确定删除「${row.name}」？`, '删除接口', { type: 'warning' })
  await api.deleteAPI(row.id)
  ElMessage.success('已删除')
  const next = { ...lastRunMap.value }
  delete next[row.id]
  lastRunMap.value = next
  if (selectedApi.value?.id === row.id) {
    selectedApi.value = null
    detail.value = null
    lastRun.value = null
  }
  loadApis()
}

function applyQuerySuggestions() {
  const us = route.query.suggest_us
  const bdd = route.query.suggest_bdd
  const tc = route.query.suggest_tc
  if (us && metaForm.value && !metaForm.value.user_story) metaForm.value.user_story = String(us)
  if (tc && metaForm.value && !metaForm.value.tc_ref) metaForm.value.tc_ref = String(tc)
  if (bdd && metaForm.value && !metaForm.value.bdd_ref) metaForm.value.bdd_ref = String(bdd)
  if (us || bdd || tc) activeTab.value = 'trace'
}

onMounted(() => {
  loadTree()
  loadApis()
})

watch(
  () => route.query,
  () => {
    if (detail.value) applyQuerySuggestions()
  }
)
</script>

<style scoped lang="scss">
.workbench {
  flex: 1;
  display: flex;
  gap: 12px;
  padding: 12px 16px 16px;
  min-height: 0;
}

.tree-panel {
  width: var(--tree-w);
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.tree-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  font-size: 13px;
  font-weight: 600;
  border-bottom: 1px solid var(--color-border);
}

.tree-panel :deep(.el-tree) {
  flex: 1;
  overflow: auto;
  padding: 8px;
  background: transparent;
}

.list-panel {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.list-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
  flex-wrap: nowrap;
  min-width: 0;
}

.search-input {
  width: 220px;
  flex-shrink: 1;
  min-width: 120px;
}
.method-filter {
  width: 88px;
  flex-shrink: 0;
}
.gap-filter {
  width: 108px;
  flex-shrink: 0;
}
.mr-filter {
  width: 100px;
  flex-shrink: 0;
}

.toolbar-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--color-muted);
  flex-shrink: 0;
  white-space: nowrap;
}

.count-badge {
  font-weight: 700;
  color: var(--color-primary);
  font-size: 14px;
}

.folder-tag {
  padding: 2px 8px;
  background: #f0f5ff;
  color: #3370ff;
  border-radius: 4px;
  font-size: 11px;
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spacer {
  flex: 1;
  min-width: 8px;
}

.table-wrap {
  flex: 1;
  min-height: 0;
  padding: 0 8px 8px;
}

.api-table {
  height: 100%;
  width: 100%;
}

:deep(.api-table table) {
  table-layout: fixed;
  width: 100% !important;
}

:deep(.api-table .el-table__header-wrapper table),
:deep(.api-table .el-table__body-wrapper table) {
  width: 100% !important;
}

:deep(.api-table .col-ops .cell) {
  padding: 0 4px;
  overflow: visible;
}

.row-actions {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.name-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.name-text {
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.name-sub {
  font-size: 11px;
  color: var(--color-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.path-cell {
  font-size: 12px;
  color: #4b5563;
  background: transparent;
}

.run-status {
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
  &--pass { color: #16a34a; }
  &--fail { color: #dc2626; }
  &--loading { color: #2563eb; font-size: 10px; }
  &--none { color: #c0c4cc; font-weight: 400; }
}

.run-status--link {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  border: none;
  background: none;
  padding: 0;
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
  &:hover {
    opacity: 0.85;
  }
}

.result-ico {
  font-size: 14px;
  line-height: 1;
  text-decoration: none;
}

:deep(.row-current) {
  td {
    background: #f0f7ff !important;
  }
}

.detail-panel {
  width: var(--detail-w);
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  &--empty {
    align-items: center;
    justify-content: center;
  }
}

.detail-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  padding: 14px 14px 0;
  flex-shrink: 0;
}

.detail-run-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.dataset-select {
  width: 168px;
}

.detail-title {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;

  h3 {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    line-height: 1.35;
    word-break: break-word;
  }
}

.detail-path-line {
  margin: 8px 14px 0;
  padding: 0;
  code {
    font-size: 12px;
    color: var(--color-muted);
    word-break: break-all;
  }
}

.url-template {
  margin: 8px 14px 0;
  font-size: 11px;
  line-height: 1.45;
  color: var(--color-muted);

  .label {
    display: block;
    margin-bottom: 4px;
    font-weight: 600;
  }
  code {
    color: #1d4ed8;
    word-break: break-all;
  }
}

.trace-form {
  padding: 4px 0;
}

.trace-hint {
  margin: 0;
  font-size: 12px;
  color: var(--color-muted);

  a {
    color: var(--color-primary);
  }
}

.detail-tabs {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 0 8px 12px;

  :deep(.el-tabs__content) {
    flex: 1;
    overflow: auto;
    padding: 0 6px;
  }
}

.meta-list {
  margin: 0;
  dt {
    font-size: 12px;
    color: var(--color-muted);
    margin: 10px 0 4px;
    &:first-child { margin-top: 0; }
  }
  dd {
    margin: 0;
    font-size: 13px;
    line-height: 1.5;
  }
}

.section-h4 {
  margin: 16px 0 6px;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-muted);
}

.history-hint {
  margin: 0 0 8px;
  font-size: 11px;
  color: var(--color-muted);
}

.history-empty {
  margin: 16px 0 0;
  font-size: 12px;
  color: var(--color-muted);
}

.history-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.history-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  margin-bottom: 4px;
  border-radius: 6px;
  border: 1px solid var(--color-border);
  cursor: pointer;
  font-size: 12px;
  transition: background 0.15s, border-color 0.15s;

  &:hover {
    background: #f0f7ff;
    border-color: #93c5fd;
  }
}

.history-status {
  font-weight: 600;
  min-width: 32px;
}

.history-env {
  color: var(--color-primary);
  background: #eff6ff;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
}

.history-time {
  margin-left: auto;
  color: var(--color-muted);
  font-size: 11px;
}

.detail-empty {
  padding: 24px;
  text-align: center;
  .hint-sub {
    margin-top: 4px;
    font-size: 12px;
    color: var(--color-muted);
  }
}

.detail-panel h4 {
  margin: 12px 0 8px;
  font-size: 12px;
  color: var(--color-muted);
  font-weight: 600;
}
</style>
