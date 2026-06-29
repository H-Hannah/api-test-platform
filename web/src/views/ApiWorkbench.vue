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
          placeholder="搜索名称、说明、路径、模块"
          clearable
          class="search-input"
          :prefix-icon="Search"
        />
        <el-select v-model="methodFilter" placeholder="方法" clearable class="method-filter">
          <el-option v-for="m in methodOptions" :key="m" :label="m" :value="m" />
        </el-select>
        <el-select v-model="gapFilter" placeholder="场景缺口" clearable class="gap-filter" @change="loadApis">
          <el-option label="无用例" value="not_ready" />
          <el-option label="有用例" value="ready" />
        </el-select>
        <el-input v-model="mrFilter" placeholder="MR 标签" clearable class="mr-filter" @change="loadApis" />
        <el-button
          size="small"
          type="primary"
          plain
          :disabled="!gapApisCount"
          :loading="batchGenerating"
          @click="batchGenerateCases"
        >
          批量补用例{{ gapApisCount ? ` (${gapApisCount})` : '' }}
        </el-button>
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
          class="api-table"
          empty-text="暂无接口，请用插件录制并入库"
          :row-class-name="rowClassName"
          @current-change="onRowSelect"
        >
          <el-table-column label="方法" width="80" align="center" fixed="left">
            <template #default="{ row }">
              <MethodBadge :method="row.method" />
            </template>
          </el-table-column>
          <el-table-column label="接口" width="160" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="name-cell">
                <span class="name-text">{{ row.name }}</span>
                <span v-if="row.folder_path" class="name-sub">{{ row.folder_path }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="说明" width="260" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="desc-cell">{{ row.description || '—' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="路径" min-width="120" show-overflow-tooltip>
            <template #default="{ row }">
              <code class="path-cell">{{ row.path }}</code>
            </template>
          </el-table-column>
          <el-table-column label="场景缺口" width="88" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.scenario_ready" type="success" size="small">
                就绪<span v-if="row.case_count"> · {{ row.case_count }}</span>
              </el-tag>
              <el-tag v-else type="warning" size="small">缺口</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="更新时间" min-width="128" align="center" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="time-cell">{{ formatUpdatedAt(row.updated_at) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="96" align="center" fixed="right" class-name="col-ops">
            <template #default="{ row }">
              <div class="row-actions">
                <el-tooltip content="AI 生成用例" placement="top">
                  <el-button
                    plain
                    size="small"
                    circle
                    type="primary"
                    :icon="VideoPlay"
                    :loading="generatingCaseId === row.id"
                    @click.stop="generateCases(row)"
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
            <section class="cases-section">
              <el-button
                type="primary"
                size="small"
                :loading="generatingCaseId === detail.id"
                @click="generateCases(detail)"
              >
                AI 生成用例
              </el-button>
              <el-alert
                v-if="staleDatasetCount"
                type="warning"
                :closable="false"
                show-icon
                class="stale-alert"
                :title="`${staleDatasetCount} 条用例可能已过期`"
                description="接口定义已更新，请在「接口用例」页核对后保存"
              />
              <div v-if="apiDatasets.length" class="quick-run-bar">
                <el-select
                  v-model="selectedDatasetId"
                  size="small"
                  class="dataset-select"
                  placeholder="选择用例"
                >
                  <el-option
                    v-for="ds in apiDatasets"
                    :key="ds.id"
                    :label="datasetOptionLabel(ds)"
                    :value="ds.id"
                  />
                </el-select>
                <el-button
                  size="small"
                  type="success"
                  :icon="VideoPlay"
                  :disabled="!envId || !selectedDatasetId"
                  :loading="runningCase"
                  @click="runSelectedCase"
                >
                  执行
                </el-button>
                <router-link :to="casesLink()" class="manage-link">管理 →</router-link>
              </div>
              <div v-if="apiDatasets.length" class="case-bindings">
                <p class="section-label">已绑定用例（{{ apiDatasets.length }}）</p>
                <router-link
                  v-for="ds in apiDatasets"
                  :key="ds.id"
                  :to="casesLink(ds.id)"
                  class="case-link"
                  :class="{ 'case-link--stale': ds.stale }"
                >
                  {{ ds.dataset_key }} · {{ ds.name }}
                  <el-tag v-if="ds.stale" type="danger" size="small">过期</el-tag>
                </router-link>
              </div>
              <p v-else-if="detail.case_count > 0" class="case-hint">
                <router-link :to="casesLink()" class="case-link">查看 {{ detail.case_count }} 条用例 →</router-link>
              </p>
            </section>
          </el-tab-pane>
          <el-tab-pane label="请求" name="request">
            <h4>Headers</h4>
            <pre class="json-block">{{ pretty(detail.headers) }}</pre>
            <h4>Body</h4>
            <pre class="json-block">{{ pretty(detail.body) }}</pre>
          </el-tab-pane>
        </el-tabs>
      </template>

      <div v-else class="detail-empty">
        <el-empty description="点击列表中的接口查看详情" />
      </div>
    </aside>

    <RunResultDialog
      v-model="runDialogOpen"
      :run="lastRun"
      :env-name="currentEnvName"
      :api-name="detail?.name"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Search, Delete, VideoPlay } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api/client'
import { useAppStore } from '@/composables/useAppStore'
import { notifyRunResult } from '@/utils/notify'
import { hasTag } from '@/utils/dataset'
import MethodBadge from '@/components/MethodBadge.vue'
import RunResultDialog from '@/components/RunResultDialog.vue'

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
const generatingCaseId = ref(null)
const batchGenerating = ref(false)
const apiDatasets = ref([])
const selectedDatasetId = ref(null)
const runningCase = ref(false)
const runDialogOpen = ref(false)
const lastRun = ref(null)

const currentEnvName = computed(() => {
  const e = environments.value.find((x) => x.id === envId.value)
  return e?.name || ''
})

const gapApisCount = computed(() => filteredApis.value.filter((a) => !a.scenario_ready).length)

const staleDatasetCount = computed(() => apiDatasets.value.filter((d) => d.stale).length)

function datasetOptionLabel(ds) {
  const base = `${ds.dataset_key} · ${ds.name}`
  if (ds.stale) return `${base} [过期]`
  if (hasTag(ds.tags, 'draft')) return `${base} [草稿]`
  return base
}

const filteredApis = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  const mf = methodFilter.value
  return apis.value.filter((a) => {
    if (mf && (a.method || '').toUpperCase() !== mf) return false
    if (!k) return true
    return (
      (a.name || '').toLowerCase().includes(k) ||
      (a.path || '').toLowerCase().includes(k) ||
      (a.folder_path || '').toLowerCase().includes(k) ||
      (a.description || '').toLowerCase().includes(k)
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

function formatUpdatedAt(val) {
  if (!val) return '—'
  const s = String(val).trim().replace('T', ' ')
  if (s.length >= 16) return s.slice(0, 16)
  return s
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
    await tryFocusApiFromQuery()
  } finally {
    listLoading.value = false
  }
}

async function tryFocusApiFromQuery() {
  const qid = route.query.api_id
  if (!qid) return
  const row = apis.value.find((a) => String(a.id) === String(qid))
  if (row) await focusRow(row)
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

async function refreshDetail() {
  if (!selectedApi.value) return
  detail.value = await api.getAPI(selectedApi.value.id)
  await loadApiDatasets()
}

function casesLink(datasetId) {
  const apiId = selectedApi.value?.id || detail.value?.id
  const query = { api_id: apiId }
  if (datasetId) query.dataset_id = datasetId
  return { path: '/cases', query }
}

async function loadApiDatasets() {
  if (!selectedApi.value?.id) {
    apiDatasets.value = []
    selectedDatasetId.value = null
    return
  }
  try {
    apiDatasets.value = await api.listTestDatasets({ api_id: selectedApi.value.id })
    if (selectedDatasetId.value && !apiDatasets.value.some((d) => d.id === selectedDatasetId.value)) {
      selectedDatasetId.value = apiDatasets.value[0]?.id || null
    } else if (!selectedDatasetId.value && apiDatasets.value.length) {
      selectedDatasetId.value = apiDatasets.value[0].id
    }
  } catch {
    apiDatasets.value = []
    selectedDatasetId.value = null
  }
}

async function runSelectedCase() {
  if (!selectedApi.value?.id || !selectedDatasetId.value || !envId.value) {
    ElMessage.warning('请选择运行环境（顶栏）和用例')
    return
  }
  runningCase.value = true
  try {
    const run = await api.runAPI(selectedApi.value.id, envId.value, selectedDatasetId.value)
    lastRun.value = run
    runDialogOpen.value = true
    notifyRunResult(run, currentEnvName.value)
  } finally {
    runningCase.value = false
  }
}

async function batchGenerateCases() {
  const targets = filteredApis.value.filter((a) => !a.scenario_ready)
  if (!targets.length) {
    ElMessage.info('当前列表没有缺口接口')
    return
  }
  await ElMessageBox.confirm(
    `将为 ${targets.length} 个接口依次 AI 生成用例，可能耗时数分钟，是否继续？`,
    '批量补用例',
    { type: 'info' }
  )
  batchGenerating.value = true
  let ok = 0
  let fail = 0
  try {
    for (const row of targets) {
      try {
        await api.generateAPICases(row.id, {})
        ok++
      } catch {
        fail++
      }
    }
    ElMessage.success(`批量完成：${ok} 成功${fail ? `，${fail} 失败` : ''}`)
    await loadApis()
    if (selectedApi.value) await refreshDetail()
  } finally {
    batchGenerating.value = false
  }
}

async function generateCases(row) {
  if (!row?.id) return
  generatingCaseId.value = row.id
  try {
    const res = await api.generateAPICases(row.id, {})
    const n = res.datasets?.length || 0
    ElMessage.success(n ? `已生成 ${n} 条用例` : '生成完成')
    await loadApis()
    if (selectedApi.value?.id === row.id) {
      detail.value = await api.getAPI(row.id)
      const updated = apis.value.find((a) => a.id === row.id)
      if (updated) selectedApi.value = updated
      await loadApiDatasets()
    }
  } finally {
    generatingCaseId.value = null
  }
}

async function removeApi(row) {
  await ElMessageBox.confirm(`确定删除「${row.name}」？`, '删除接口', { type: 'warning' })
  await api.deleteAPI(row.id)
  ElMessage.success('已删除')
  if (selectedApi.value?.id === row.id) {
    selectedApi.value = null
    detail.value = null
  }
  loadApis()
}

onMounted(() => {
  loadTree()
  loadApis()
})
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

.table-wrap {
  flex: 1;
  min-height: 0;
  padding: 0 8px 8px;
}

.api-table {
  height: 100%;
  width: 100%;
}

:deep(.api-table .el-table__body-wrapper) {
  overflow-x: auto;
}

:deep(.api-table .col-ops .cell) {
  padding: 0 4px;
  overflow: visible;
}

.row-actions {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
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

.desc-cell {
  font-size: 12px;
  color: #6b7280;
  line-height: 1.4;
}

.time-cell {
  font-size: 12px;
  color: #9ca3af;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
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

.cases-section {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--color-border);
}

.quick-run-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  flex-wrap: wrap;
}

.stale-alert {
  margin-top: 12px;
}

.dataset-select {
  flex: 1;
  min-width: 120px;
}

.manage-link {
  font-size: 12px;
  color: var(--color-primary);
  text-decoration: none;
  white-space: nowrap;
}

.section-label {
  margin: 12px 0 6px;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-muted);
}

.case-bindings {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.case-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--color-primary);
  text-decoration: none;
  padding: 6px 8px;
  border-radius: 6px;
  background: #f0f7ff;
  transition: background 0.15s;

  &:hover {
    background: #dbeafe;
  }

  &--stale {
    border: 1px solid var(--el-color-danger-light-5);
  }
}

.case-hint {
  margin: 10px 0 0;
  font-size: 12px;
}

.detail-empty {
  padding: 24px;
  text-align: center;
}

.detail-panel h4 {
  margin: 12px 0 8px;
  font-size: 12px;
  color: var(--color-muted);
  font-weight: 600;
}

.json-block {
  margin: 0 0 12px;
  padding: 10px 12px;
  background: #f8fafc;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 12px;
  line-height: 1.5;
  overflow: auto;
  max-height: 280px;
}
</style>
