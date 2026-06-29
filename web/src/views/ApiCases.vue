<template>
  <div class="workbench">
    <aside class="tree-panel panel-card">
      <div class="tree-head">
        <span>模块目录</span>
        <el-button text type="primary" size="small" @click="selectFolder(null)">全部</el-button>
      </div>
      <el-tree
        v-loading="treeLoading"
        :data="treeData"
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
        highlight-current
        default-expand-all
        @node-click="(node) => selectFolder(node.id === 0 ? null : node.id)"
      />
    </aside>

    <section class="list-panel panel-card">
      <div class="list-toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索接口或用例"
          clearable
          class="search-input"
          :prefix-icon="Search"
        />
        <el-select v-model="gapFilter" placeholder="用例缺口" clearable class="gap-filter" @change="loadApis">
          <el-option label="无用例" value="not_ready" />
          <el-option label="有用例" value="ready" />
        </el-select>
        <template v-if="selectedApi">
          <el-button size="small" :loading="generating" @click="generateCases">AI 生成</el-button>
          <el-button
            size="small"
            type="primary"
            :disabled="!envId || !cases.length"
            :loading="runningAll"
            @click="runAllCases"
          >
            执行全部
          </el-button>
        </template>
        <span class="toolbar-meta">
          <span class="count-badge">{{ filteredApis.length }}</span>
          <span class="count-label">个接口</span>
        </span>
      </div>

      <div class="split-wrap">
        <div class="api-section">
          <p class="section-title">接口</p>
          <el-table
            v-loading="listLoading"
            :data="filteredApis"
            stripe
            highlight-current-row
            height="100%"
            class="api-table"
            :row-class-name="apiRowClass"
            @current-change="onApiSelect"
          >
            <el-table-column label="方法" width="68" align="center">
              <template #default="{ row }">
                <MethodBadge :method="row.method" />
              </template>
            </el-table-column>
            <el-table-column prop="name" label="名称" min-width="100" show-overflow-tooltip />
            <el-table-column label="缺口" width="72" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.scenario_ready" type="success" size="small">{{ row.case_count || '✓' }}</el-tag>
                <el-tag v-else type="warning" size="small">缺</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="case-section">
          <p class="section-title">
            用例
            <span v-if="selectedApi" class="section-sub">{{ selectedApi.name }}</span>
            <el-tag v-if="staleCaseCount" type="danger" size="small" class="section-badge">
              {{ staleCaseCount }} 过期
            </el-tag>
          </p>
          <el-table
            v-loading="casesLoading"
            :data="filteredCases"
            stripe
            highlight-current-row
            height="100%"
            class="case-table"
            empty-text="请选择接口或 AI 生成用例"
            :row-class-name="caseRowClass"
            @current-change="onCaseSelect"
          >
            <el-table-column prop="dataset_key" label="Key" width="100" show-overflow-tooltip />
            <el-table-column prop="name" label="名称" min-width="120" show-overflow-tooltip />
            <el-table-column label="状态" width="88" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.stale" type="danger" size="small">过期</el-tag>
                <el-tag v-else-if="hasTag(row.tags, 'draft')" type="info" size="small">草稿</el-tag>
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column label="断言" width="56" align="center">
              <template #default="{ row }">{{ assertionCount(row) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="56" align="center">
              <template #default="{ row }">
                <el-tooltip content="执行" placement="top">
                  <el-button
                    circle
                    size="small"
                    type="primary"
                    plain
                    :icon="VideoPlay"
                    :disabled="!envId"
                    :loading="runningCaseId === row.id"
                    @click.stop="runCase(row)"
                  />
                </el-tooltip>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </section>

    <aside class="detail-panel panel-card" :class="{ 'detail-panel--empty': !selectedCase }">
      <template v-if="selectedCase && caseForm">
        <div class="detail-toolbar">
          <h3>{{ caseForm.name }}</h3>
          <div class="detail-actions">
            <el-button
              v-if="hasTag(selectedCase.tags, 'draft')"
              size="small"
              type="success"
              plain
              :loading="confirming"
              @click="confirmCase"
            >
              确认用例
            </el-button>
            <el-button size="small" type="primary" :loading="saving" @click="saveCase">保存</el-button>
            <el-button
              size="small"
              :icon="VideoPlay"
              :disabled="!envId"
              :loading="runningCaseId === selectedCase.id"
              @click="runCase(selectedCase)"
            >
              执行
            </el-button>
          </div>
        </div>
        <p class="case-key"><code>{{ caseForm.dataset_key }}</code></p>
        <el-alert
          v-if="selectedCase.stale"
          type="warning"
          :closable="false"
          show-icon
          class="stale-alert"
          title="用例可能已过期"
          :description="selectedCase.stale_reason || '接口定义已更新，请核对断言与请求参数后保存'"
        />
        <div v-if="caseTags.length" class="case-tags">
          <el-tag v-for="t in caseTags" :key="t" size="small" :type="tagType(t)">{{ t }}</el-tag>
        </div>

        <el-form label-position="top" size="small" class="case-form">
          <el-form-item label="名称">
            <el-input v-model="caseForm.name" />
          </el-form-item>
          <el-form-item label="说明">
            <el-input v-model="caseForm.description" type="textarea" :rows="2" />
          </el-form-item>
          <el-form-item label="Variables（JSON）">
            <el-input v-model="caseForm.variables" type="textarea" :rows="3" class="mono-input" />
          </el-form-item>
          <el-form-item label="Body 覆盖">
            <el-input v-model="caseForm.body_override" type="textarea" :rows="3" class="mono-input" />
          </el-form-item>
          <el-form-item label="断言">
            <div class="assert-toolbar">
              <el-button size="small" @click="addAssertion">+ 断言</el-button>
            </div>
            <el-table :data="caseForm.assertions" size="small" border class="assert-table">
              <el-table-column prop="type" label="类型" width="88">
                <template #default="{ row }">
                  <el-select v-model="row.type" size="small">
                    <el-option label="status_code" value="status_code" />
                    <el-option label="json_path" value="json_path" />
                    <el-option label="duration_ms" value="duration_ms" />
                  </el-select>
                </template>
              </el-table-column>
              <el-table-column prop="expression" label="表达式" min-width="80">
                <template #default="{ row }">
                  <el-input v-model="row.expression" size="small" />
                </template>
              </el-table-column>
              <el-table-column prop="operator" label="运算" width="72">
                <template #default="{ row }">
                  <el-input v-model="row.operator" size="small" />
                </template>
              </el-table-column>
              <el-table-column prop="expected" label="期望" width="72">
                <template #default="{ row }">
                  <el-input v-model="row.expected" size="small" />
                </template>
              </el-table-column>
              <el-table-column width="40" align="center">
                <template #default="{ $index }">
                  <el-button link type="danger" size="small" @click="removeAssertion($index)">×</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-form-item>
          <el-button type="danger" plain size="small" @click="deleteCase">删除用例</el-button>
        </el-form>
      </template>
      <div v-else class="detail-empty">
        <el-empty :description="selectedApi ? '选择一条用例进行编辑与执行' : '先选择接口'" />
      </div>
    </aside>

    <RunResultDialog
      v-model="runDialogOpen"
      :run="lastRun"
      :env-name="currentEnvName"
      :api-name="selectedApi?.name"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Search, VideoPlay } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api/client'
import { useAppStore } from '@/composables/useAppStore'
import { notifyRunResult } from '@/utils/notify'
import { hasTag, parseTags, removeTag } from '@/utils/dataset'
import MethodBadge from '@/components/MethodBadge.vue'
import RunResultDialog from '@/components/RunResultDialog.vue'

const route = useRoute()
const { envId, environments } = useAppStore()

const treeLoading = ref(false)
const listLoading = ref(false)
const casesLoading = ref(false)
const treeData = ref([])
const apis = ref([])
const cases = ref([])
const folderId = ref(null)
const keyword = ref('')
const gapFilter = ref('')
const selectedApi = ref(null)
const selectedCase = ref(null)
const caseForm = ref(null)
const saving = ref(false)
const confirming = ref(false)
const generating = ref(false)
const runningCaseId = ref(null)
const runningAll = ref(false)
const runDialogOpen = ref(false)
const lastRun = ref(null)

const currentEnvName = computed(() => {
  const e = environments.value.find((x) => x.id === envId.value)
  return e?.name || ''
})

const filteredApis = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  return apis.value.filter((a) => {
    if (!k) return true
    return (
      (a.name || '').toLowerCase().includes(k) ||
      (a.path || '').toLowerCase().includes(k)
    )
  })
})

const filteredCases = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  if (!k) return cases.value
  return cases.value.filter((c) => {
    return (
      (c.name || '').toLowerCase().includes(k) ||
      (c.dataset_key || '').toLowerCase().includes(k)
    )
  })
})

const staleCaseCount = computed(() => cases.value.filter((c) => c.stale).length)
const draftCaseCount = computed(() => cases.value.filter((c) => hasTag(c.tags, 'draft')).length)

const caseTags = computed(() => (selectedCase.value ? parseTags(selectedCase.value.tags) : []))

function tagType(tag) {
  if (tag === 'draft') return 'info'
  if (tag === 'ai-inferred') return 'warning'
  return ''
}

function apiRowClass({ row }) {
  return selectedApi.value?.id === row.id ? 'row-current' : ''
}

function caseRowClass({ row }) {
  return selectedCase.value?.id === row.id ? 'row-current' : ''
}

function assertionCount(row) {
  try {
    const list = JSON.parse(row.assertions || '[]')
    return Array.isArray(list) ? list.length : 0
  } catch {
    return 0
  }
}

function parseAssertions(raw) {
  try {
    const list = JSON.parse(raw || '[]')
    return Array.isArray(list) ? list.map((a) => ({ ...a })) : []
  } catch {
    return []
  }
}

function loadCaseForm(ds) {
  caseForm.value = {
    dataset_key: ds.dataset_key,
    name: ds.name,
    description: ds.description || '',
    variables: ds.variables || '{}',
    body_override: ds.body_override || '',
    assertions: parseAssertions(ds.assertions)
  }
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

async function loadApis() {
  listLoading.value = true
  try {
    const p = {}
    if (folderId.value) p.folder_id = folderId.value
    if (gapFilter.value) p.gap = gapFilter.value
    apis.value = await api.listAPIs(p)
    await applyRouteQuery()
  } finally {
    listLoading.value = false
  }
}

async function loadCases() {
  if (!selectedApi.value?.id) {
    cases.value = []
    return
  }
  casesLoading.value = true
  try {
    cases.value = await api.listTestDatasets({ api_id: selectedApi.value.id })
  } finally {
    casesLoading.value = false
  }
}

function selectFolder(id) {
  folderId.value = id || null
  selectedApi.value = null
  selectedCase.value = null
  caseForm.value = null
  cases.value = []
  loadApis()
}

async function onApiSelect(row) {
  if (!row) return
  selectedApi.value = row
  selectedCase.value = null
  caseForm.value = null
  await loadCases()
}

async function onCaseSelect(row) {
  if (!row) return
  selectedCase.value = row
  loadCaseForm(row)
}

async function applyRouteQuery() {
  const apiId = route.query.api_id
  if (!apiId) return
  const row = apis.value.find((a) => String(a.id) === String(apiId))
  if (!row) return
  selectedApi.value = row
  await loadCases()
  const dsId = route.query.dataset_id
  if (dsId) {
    const c = cases.value.find((x) => String(x.id) === String(dsId))
    if (c) {
      selectedCase.value = c
      loadCaseForm(c)
    }
  }
}

async function generateCases() {
  if (!selectedApi.value?.id) return
  generating.value = true
  try {
    const res = await api.generateAPICases(selectedApi.value.id, {})
    const n = res.datasets?.length || 0
    ElMessage.success(n ? `已生成 ${n} 条用例` : '生成完成')
    await loadApis()
    const updated = apis.value.find((a) => a.id === selectedApi.value.id)
    if (updated) selectedApi.value = updated
    await loadCases()
  } finally {
    generating.value = false
  }
}

function addAssertion() {
  caseForm.value.assertions.push({
    type: 'json_path',
    expression: '$.code',
    operator: 'eq',
    expected: '0'
  })
}

function removeAssertion(idx) {
  caseForm.value.assertions.splice(idx, 1)
}

async function confirmCase() {
  if (!selectedCase.value?.id) return
  confirming.value = true
  try {
    const tags = removeTag(selectedCase.value.tags, 'draft')
    const updated = await api.updateTestDataset(selectedCase.value.id, {
      tags: JSON.stringify(tags)
    })
    ElMessage.success('已确认，草稿标记已移除')
    await loadCases()
    selectedCase.value = updated
    loadCaseForm(updated)
  } finally {
    confirming.value = false
  }
}

async function saveCase() {
  if (!selectedCase.value?.id || !caseForm.value) return
  saving.value = true
  try {
    const updated = await api.updateTestDataset(selectedCase.value.id, {
      name: caseForm.value.name.trim(),
      description: caseForm.value.description,
      variables: caseForm.value.variables,
      body_override: caseForm.value.body_override,
      assertions: JSON.stringify(caseForm.value.assertions)
    })
    ElMessage.success('已保存')
    await loadCases()
    selectedCase.value = updated
    loadCaseForm(updated)
  } finally {
    saving.value = false
  }
}

async function deleteCase() {
  if (!selectedCase.value?.id) return
  await ElMessageBox.confirm(`删除用例「${selectedCase.value.name}」？`, '确认', { type: 'warning' })
  await api.deleteTestDataset(selectedCase.value.id)
  ElMessage.success('已删除')
  selectedCase.value = null
  caseForm.value = null
  await loadCases()
  await loadApis()
}

async function runCase(row) {
  if (!selectedApi.value?.id || !row?.id || !envId.value) {
    ElMessage.warning('请先选择运行环境（顶栏）')
    return
  }
  runningCaseId.value = row.id
  try {
    const run = await api.runAPI(selectedApi.value.id, envId.value, row.id)
    lastRun.value = run
    runDialogOpen.value = true
    notifyRunResult(run, currentEnvName.value)
  } finally {
    runningCaseId.value = null
  }
}

async function runAllCases() {
  if (!selectedApi.value?.id || !envId.value || !cases.value.length) return
  if (draftCaseCount.value > 0) {
    await ElMessageBox.confirm(
      `含 ${draftCaseCount.value} 条 AI 草稿用例，确认全部执行？`,
      '执行全部',
      { type: 'warning' }
    )
  }
  runningAll.value = true
  let passed = 0
  let failed = 0
  try {
    for (const c of cases.value) {
      runningCaseId.value = c.id
      const run = await api.runAPI(selectedApi.value.id, envId.value, c.id)
      if (run.status === 'passed') passed++
      else failed++
      lastRun.value = run
    }
    runDialogOpen.value = true
    ElMessage[failed ? 'warning' : 'success'](`执行完成：${passed} 通过，${failed} 失败`)
  } finally {
    runningCaseId.value = null
    runningAll.value = false
  }
}

onMounted(() => {
  loadTree()
  loadApis()
})

watch(
  () => route.query,
  () => {
    if (apis.value.length) applyRouteQuery()
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
  flex-wrap: wrap;
}

.search-input { width: 180px; }
.gap-filter { width: 100px; }

.toolbar-meta {
  margin-left: auto;
  font-size: 12px;
  color: var(--color-muted);
  display: flex;
  align-items: center;
  gap: 4px;
}

.count-badge {
  font-weight: 700;
  color: var(--color-primary);
  font-size: 14px;
}

.split-wrap {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 0 8px 8px;
  gap: 8px;
}

.api-section,
.case-section {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.section-title {
  margin: 8px 0 4px;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-muted);
  padding: 0 4px;
}

.section-sub {
  font-weight: 400;
  color: var(--color-primary);
  margin-left: 6px;
}

.api-table,
.case-table {
  flex: 1;
  min-height: 0;
}

:deep(.row-current td) {
  background: #f0f7ff !important;
}

.detail-panel {
  width: 360px;
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
  gap: 8px;
  padding: 12px 12px 0;

  h3 {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
    line-height: 1.35;
    flex: 1;
    min-width: 0;
    word-break: break-word;
  }
}

.detail-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.case-key {
  margin: 6px 12px 0;
  code { font-size: 11px; color: var(--color-muted); }
}

.stale-alert {
  margin: 8px 12px 0;
}

.case-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin: 8px 12px 0;
}

.section-badge {
  margin-left: 6px;
  vertical-align: middle;
}

.case-form {
  flex: 1;
  overflow: auto;
  padding: 8px 12px 12px;
}

.mono-input :deep(textarea) {
  font-family: ui-monospace, monospace;
  font-size: 12px;
}

.assert-toolbar {
  margin-bottom: 6px;
}

.assert-table {
  width: 100%;
}

.detail-empty {
  padding: 24px;
}
</style>
