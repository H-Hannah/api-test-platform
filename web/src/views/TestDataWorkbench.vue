<template>
  <div class="td-page panel-card">
    <div class="page-toolbar">
      <span class="toolbar-title">测试数据</span>
      <span class="spacer"></span>
      <router-link to="/apis" class="link-next">接口执行 →</router-link>
    </div>

    <el-row :gutter="16" class="body-row">
      <el-col :span="10">
        <section class="input-panel">
          <el-alert
            type="info"
            :closable="false"
            show-icon
            class="phase-alert"
            title="需求文档(edgen-product-docs) + 后端设计(osp-wiki) + 测试用例(qa-doc-generator) → AI 生成测试集"
          />

          <div class="setup-block">
            <div class="setup-label">用例目录 · qa-doc-generator（选分支/版本/需求）</div>
            <div class="setup-row">
              <el-select
                v-model="docForm.branch"
                placeholder="分支"
                filterable
                :loading="loadingBranches"
                class="sel-branch"
                @change="onBranchChange"
              >
                <el-option v-for="b in branches" :key="b" :label="b" :value="b" />
              </el-select>
              <el-select
                v-model="docForm.version"
                placeholder="版本"
                filterable
                :loading="loadingCatalog"
                :disabled="!docForm.branch"
                class="sel-version"
                @change="onVersionChange"
              >
                <el-option v-for="v in versions" :key="v" :label="v" :value="v" />
              </el-select>
              <el-select
                v-model="docForm.requirement_id"
                placeholder="需求"
                filterable
                :loading="loadingReqs"
                :disabled="!docForm.version"
                class="sel-req"
              >
                <el-option v-for="r in requirements" :key="r" :label="r" :value="r" />
              </el-select>
              <el-button :loading="loadingDocs" :disabled="!canLoadDocs" @click="loadDocs">
                加载文档
              </el-button>
            </div>
            <div v-if="docLoaded" class="doc-status">
              <el-tag v-if="form.prd_text.trim()" type="success" size="small">
                需求文档{{ prdSourceLabel ? ` · ${prdSourceLabel}` : '' }}
              </el-tag>
              <el-tag v-if="form.be_tech_text.trim()" type="success" size="small">
                后端方案{{ beSourceLabel ? ` · ${beSourceLabel}` : '' }}
              </el-tag>
              <el-tag v-if="form.cases_json.trim()" type="success" size="small">
                测试用例{{ tcSourceLabel ? ` · ${tcSourceLabel}` : '' }} {{ caseCount ? `(${caseCount})` : '' }}
              </el-tag>
              <el-tag v-for="(w, i) in docWarnings" :key="i" type="warning" size="small">{{ w }}</el-tag>
            </div>
          </div>

          <el-form label-position="top">
            <el-form-item label="需求名称">
              <el-input v-model="form.requirement_name" placeholder="自动填充，可改" />
            </el-form-item>

            <el-collapse v-model="docCollapse">
              <el-collapse-item title="需求文档（可编辑）" name="prd">
                <el-input v-model="form.prd_text" type="textarea" :rows="5" placeholder="PRD / 产品说明" />
              </el-collapse-item>
              <el-collapse-item title="后端技术方案（可编辑）" name="be">
                <el-input v-model="form.be_tech_text" type="textarea" :rows="6" placeholder="API 设计、错误码、数据模型" />
              </el-collapse-item>
              <el-collapse-item title="测试用例 JSON（可编辑）" name="tc">
                <el-input v-model="form.cases_json" type="textarea" :rows="5" placeholder="test-docs JSON 数组" />
              </el-collapse-item>
            </el-collapse>

            <el-form-item label="已知接口（每行 METHOD /path，可选）">
              <el-input v-model="form.api_hints_text" type="textarea" :rows="2" placeholder="GET /v2/trackers" />
            </el-form-item>

            <el-form-item label="补充说明">
              <el-input v-model="form.hint" type="textarea" :rows="2" placeholder="例：重点覆盖 Tracker 创建与 Brief 卡片流" />
            </el-form-item>

            <el-button type="primary" :loading="generating" :disabled="!canGenerate" @click="generate">
              AI 生成测试集
            </el-button>
          </el-form>
        </section>

        <section class="list-panel">
          <div class="list-head">
            <h3 class="sub-title">已导入数据集</h3>
            <el-button link type="primary" size="small" :disabled="!form.version || !form.requirement_id" @click="loadDatasets">
              刷新
            </el-button>
          </div>
          <el-table v-loading="listLoading" :data="datasets" size="small" empty-text="生成后点「导入平台」">
            <el-table-column prop="dataset_key" label="键" width="88" show-overflow-tooltip />
            <el-table-column prop="name" label="名称" min-width="100" show-overflow-tooltip />
            <el-table-column prop="obtain_type" label="来源" width="72" />
            <el-table-column label="" width="48" align="center">
              <template #default="{ row }">
                <el-button link type="danger" size="small" @click.stop="removeDataset(row)">删</el-button>
              </template>
            </el-table-column>
          </el-table>
        </section>
      </el-col>

      <el-col :span="14">
        <section v-if="result" class="preview-panel">
          <div class="preview-head">
            <h3>{{ result.requirement_name || result.requirement_id }}</h3>
            <el-tag :type="result.gate_passed ? 'success' : 'warning'" size="small">
              {{ result.gate_passed ? '门禁通过' : '门禁待完善' }}
            </el-tag>
          </div>

          <p v-if="result.git_output_hint" class="git-hint">
            <strong>建议 Git 路径：</strong>{{ result.git_yaml_path || result.git_output_hint }}
          </p>

          <div v-if="result.gate_reasons?.length" class="gate-reasons">
            <el-alert type="warning" :closable="false" show-icon title="门禁提示">
              <ul>
                <li v-for="(r, i) in result.gate_reasons" :key="i">{{ r }}</li>
              </ul>
            </el-alert>
          </div>

          <div class="stats-row" v-if="result.stats">
            <div class="stat">
              <span class="stat-num">{{ result.stats.total_collections || 0 }}</span>
              <span class="stat-label">测试集</span>
            </div>
            <div class="stat">
              <span class="stat-num">{{ result.stats.total_datasets }}</span>
              <span class="stat-label">数据集</span>
            </div>
            <div class="stat">
              <span class="stat-num">{{ result.stats.env_key_count }}</span>
              <span class="stat-label">环境变量键</span>
            </div>
          </div>

          <div class="action-row">
            <el-button size="small" :disabled="!result.spec_yaml" @click="copyText(result.spec_yaml)">复制 YAML</el-button>
            <el-button size="small" type="primary" plain :loading="importing" :disabled="!result.datasets?.length" @click="importToPlatform">
              导入平台
            </el-button>
            <el-button
              size="small"
              type="success"
              plain
              :loading="importingKeys"
              :disabled="!result.env_keys?.length || !envId"
              @click="importEnvKeys"
            >
              导入环境变量键
            </el-button>
          </div>

          <el-tabs v-model="previewTab">
            <el-tab-pane :label="`测试集 (${result.collections?.length || 0})`" name="collections">
              <div v-for="coll in result.collections || []" :key="coll.collection_key" class="collection-block">
                <h4 class="coll-title">{{ coll.name }}</h4>
                <p v-if="coll.description" class="coll-desc">{{ coll.description }}</p>
                <el-table :data="coll.datasets || []" size="small" border>
                  <el-table-column prop="dataset_key" label="键" width="96" />
                  <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
                  <el-table-column prop="obtain_type" label="来源" width="72" />
                  <el-table-column label="TC" width="56">
                    <template #default="{ row }">{{ row.tc_refs?.length || 0 }}</template>
                  </el-table-column>
                  <el-table-column label="API" width="56">
                    <template #default="{ row }">{{ row.api_bindings?.length || 0 }}</template>
                  </el-table-column>
                </el-table>
              </div>
              <el-empty v-if="!result.collections?.length" description="无测试集分组" />
            </el-tab-pane>
            <el-tab-pane label="YAML 规格" name="yaml">
              <pre class="artifact-block">{{ result.spec_yaml }}</pre>
            </el-tab-pane>
            <el-tab-pane label="数据集 JSON" name="json">
              <pre class="artifact-block">{{ datasetsJson }}</pre>
            </el-tab-pane>
            <el-tab-pane label="环境变量键" name="keys">
              <ul class="key-list">
                <li v-for="k in result.env_keys || []" :key="k"><code>{{ k }}</code></li>
              </ul>
              <p v-if="!result.env_keys?.length" class="muted">无</p>
            </el-tab-pane>
          </el-tabs>

          <p v-if="result.coverage_notes" class="coverage">{{ result.coverage_notes }}</p>
        </section>
        <el-empty v-else description="从 GitLab 加载三类文档后生成测试集" />
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, watch, inject, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api/client'

const route = useRoute()
const { envId } = inject('appStore')

const docForm = ref({ branch: '', version: '', requirement_id: '' })
const branches = ref([])
const versions = ref([])
const requirements = ref([])
const loadingBranches = ref(false)
const loadingCatalog = ref(false)
const loadingReqs = ref(false)
const loadingDocs = ref(false)
const docLoaded = ref(false)
const docWarnings = ref([])
const caseCount = ref(0)
const prdSourceLabel = ref('')
const beSourceLabel = ref('')
const tcSourceLabel = ref('')
const docCollapse = ref(['prd', 'be', 'tc'])

const form = ref({
  version: '',
  requirement_id: '',
  requirement_name: '',
  prd_text: '',
  be_tech_text: '',
  cases_json: '',
  api_hints_text: '',
  hint: ''
})

const generating = ref(false)
const importing = ref(false)
const importingKeys = ref(false)
const listLoading = ref(false)
const result = ref(null)
const previewTab = ref('collections')
const datasets = ref([])

const canLoadDocs = computed(
  () => docForm.value.branch && docForm.value.version.trim() && docForm.value.requirement_id.trim()
)

const canGenerate = computed(
  () =>
    form.value.version.trim() &&
    form.value.requirement_id.trim() &&
    form.value.prd_text.trim() &&
    form.value.be_tech_text.trim() &&
    form.value.cases_json.trim()
)

const datasetsJson = computed(() =>
  result.value?.datasets ? JSON.stringify(result.value.datasets, null, 2) : ''
)

function parseApiHints(text) {
  return String(text || '')
    .split('\n')
    .map((l) => l.trim())
    .filter(Boolean)
}

function syncFormFromDoc() {
  form.value.version = docForm.value.version
  form.value.requirement_id = docForm.value.requirement_id
}

async function fetchBranches() {
  loadingBranches.value = true
  try {
    branches.value = await api.listTestDocsBranches()
    if (!docForm.value.branch && branches.value.length) {
      docForm.value.branch = branches.value[0]
      await onBranchChange()
    }
  } finally {
    loadingBranches.value = false
  }
}

async function onBranchChange() {
  docForm.value.version = ''
  docForm.value.requirement_id = ''
  versions.value = []
  requirements.value = []
  if (!docForm.value.branch) return
  loadingCatalog.value = true
  try {
    const cat = await api.listTestDocsCatalog({ ref: docForm.value.branch })
    versions.value = cat.versions || []
  } finally {
    loadingCatalog.value = false
  }
}

async function onVersionChange() {
  docForm.value.requirement_id = ''
  requirements.value = []
  if (!docForm.value.version || !docForm.value.branch) return
  loadingReqs.value = true
  try {
    const cat = await api.listTestDocsCatalog({
      ref: docForm.value.branch,
      version: docForm.value.version
    })
    requirements.value = cat.requirements || []
  } finally {
    loadingReqs.value = false
  }
}

async function loadDocs() {
  if (!canLoadDocs.value) return
  loadingDocs.value = true
  docWarnings.value = []
  try {
    const pkg = await api.loadRequirementPackage({
      ref: docForm.value.branch,
      version: docForm.value.version.trim(),
      requirement_id: docForm.value.requirement_id.trim()
    })
    syncFormFromDoc()
    form.value.prd_text = pkg.prd_text || ''
    form.value.be_tech_text = pkg.be_tech_text || ''
    form.value.cases_json = pkg.cases_json || ''
    form.value.requirement_name = pkg.requirement_name || docForm.value.requirement_id
    caseCount.value = pkg.case_count || 0
    prdSourceLabel.value = pkg.prd_source || ''
    beSourceLabel.value = pkg.be_tech_source || ''
    tcSourceLabel.value = pkg.tc_source || ''
    docWarnings.value = pkg.warnings || []
    docLoaded.value = true
    if (docWarnings.value.length) ElMessage.warning('部分文档未加载，可手动补充')
    else ElMessage.success('三类文档已加载')
    await loadDatasets()
  } finally {
    loadingDocs.value = false
  }
}

function applyRouteQuery() {
  const q = route.query
  if (q.branch) docForm.value.branch = String(q.branch)
  if (q.version) {
    docForm.value.version = String(q.version)
    form.value.version = String(q.version)
  }
  if (q.requirement_id) {
    docForm.value.requirement_id = String(q.requirement_id)
    form.value.requirement_id = String(q.requirement_id)
  }
  if (q.requirement_name) form.value.requirement_name = String(q.requirement_name)
  if (q.prd_text) form.value.prd_text = String(q.prd_text)
  if (q.be_tech_text) form.value.be_tech_text = String(q.be_tech_text)
  if (q.cases_json) form.value.cases_json = String(q.cases_json)
}

async function generate() {
  if (!canGenerate.value) return
  generating.value = true
  try {
    const res = await api.generateTestData({
      version: form.value.version.trim(),
      requirement_id: form.value.requirement_id.trim(),
      requirement_name: form.value.requirement_name.trim() || form.value.requirement_id.trim(),
      prd_text: form.value.prd_text,
      be_tech_text: form.value.be_tech_text,
      cases_json: form.value.cases_json,
      api_hints: parseApiHints(form.value.api_hints_text),
      hint: form.value.hint
    })
    result.value = res
    previewTab.value = 'collections'
    if (res.gate_passed) ElMessage.success(`已生成 ${res.stats?.total_collections || 0} 个测试集`)
    else ElMessage.warning('已生成，请查看门禁提示')
  } finally {
    generating.value = false
  }
}

async function importToPlatform() {
  if (!result.value?.datasets?.length) return
  importing.value = true
  try {
    const res = await api.importTestData({
      version: result.value.version,
      requirement_id: result.value.requirement_id,
      requirement_name: result.value.requirement_name,
      spec_yaml: result.value.spec_yaml,
      env_keys: result.value.env_keys,
      datasets: result.value.datasets
    })
    ElMessage.success(`已导入 ${res.imported} 条数据集`)
    await loadDatasets()
  } finally {
    importing.value = false
  }
}

async function importEnvKeys() {
  if (!result.value?.env_keys?.length || !envId.value) return
  importingKeys.value = true
  try {
    const res = await api.importEnvVarKeys(envId.value, { keys: result.value.env_keys })
    ElMessage.success(`已添加 ${res.added} 个环境变量键（共 ${res.total} 项）`)
  } finally {
    importingKeys.value = false
  }
}

async function loadDatasets() {
  if (!form.value.version.trim() || !form.value.requirement_id.trim()) {
    datasets.value = []
    return
  }
  listLoading.value = true
  try {
    datasets.value = await api.listTestDatasets({
      version: form.value.version.trim(),
      requirement_id: form.value.requirement_id.trim()
    })
  } finally {
    listLoading.value = false
  }
}

async function removeDataset(row) {
  await ElMessageBox.confirm(`删除数据集「${row.name}」？`, '确认', { type: 'warning' })
  await api.deleteTestDataset(row.id)
  ElMessage.success('已删除')
  await loadDatasets()
}

async function copyText(text) {
  const t = text || ''
  if (!t.trim()) {
    ElMessage.warning('内容为空')
    return
  }
  try {
    await navigator.clipboard.writeText(t)
    ElMessage.success('已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}

watch([() => form.value.version, () => form.value.requirement_id], loadDatasets)
watch(() => route.query, applyRouteQuery, { deep: true })

onMounted(async () => {
  applyRouteQuery()
  await fetchBranches()
  if (docForm.value.version) await onVersionChange()
  if (canLoadDocs.value && !form.value.prd_text) await loadDocs()
  else loadDatasets()
})
</script>

<style scoped lang="scss">
.td-page {
  flex: 1;
  margin: 12px 16px 16px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: auto;
}

.page-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border);
}

.toolbar-title {
  font-size: 14px;
  font-weight: 600;
}

.spacer {
  flex: 1;
}

.link-next {
  font-size: 13px;
  color: var(--color-primary);
  text-decoration: none;
}

.body-row {
  flex: 1;
  min-height: 0;
  padding: 16px;
}

.input-panel,
.preview-panel,
.list-panel {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.phase-alert {
  margin-bottom: 16px;
}

.setup-block {
  margin-bottom: 16px;
}

.setup-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-muted);
  margin-bottom: 8px;
}

.setup-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.sel-branch {
  width: 140px;
}
.sel-version {
  width: 100px;
}
.sel-req {
  flex: 1;
  min-width: 120px;
}

.doc-status {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.sub-title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.list-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.preview-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;

  h3 {
    margin: 0;
    font-size: 16px;
  }
}

.git-hint {
  font-size: 13px;
  color: var(--color-muted);
  margin: 0 0 12px;
}

.gate-reasons ul {
  margin: 0;
  padding-left: 18px;
}

.stats-row {
  display: flex;
  gap: 24px;
  margin-bottom: 12px;
}

.stat {
  text-align: center;
}

.stat-num {
  display: block;
  font-size: 20px;
  font-weight: 700;
}

.stat-label {
  font-size: 12px;
  color: var(--color-muted);
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.collection-block {
  margin-bottom: 20px;
}

.coll-title {
  margin: 0 0 4px;
  font-size: 14px;
}

.coll-desc {
  margin: 0 0 8px;
  font-size: 13px;
  color: var(--color-muted);
}

.artifact-block {
  margin: 0;
  padding: 12px;
  background: var(--color-bg);
  border-radius: 6px;
  font-size: 12px;
  line-height: 1.5;
  max-height: 420px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

.key-list {
  margin: 0;
  padding-left: 20px;
}

.coverage {
  margin-top: 12px;
  font-size: 13px;
  color: var(--color-muted);
}

.muted {
  color: var(--color-muted);
  font-size: 13px;
}
</style>
