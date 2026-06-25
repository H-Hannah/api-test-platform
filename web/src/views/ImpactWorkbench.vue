<template>
  <div class="impact-page panel-card">
    <section class="setup">
      <div class="setup-block">
        <div class="setup-label">测试用例 · qa-doc-generator</div>
        <div class="setup-row">
          <el-select
            v-model="tcForm.branch"
            placeholder="分支"
            filterable
            :loading="loadingBranches"
            class="sel-branch"
            @change="onBranchChange"
          >
            <el-option v-for="b in branches" :key="b" :label="b" :value="b" />
          </el-select>
          <el-select
            v-model="tcForm.version"
            placeholder="版本"
            filterable
            :loading="loadingCatalog"
            :disabled="!tcForm.branch"
            class="sel-version"
            @change="onVersionChange"
          >
            <el-option v-for="v in versions" :key="v" :label="v" :value="v" />
          </el-select>
          <el-select
            v-model="tcForm.requirement_id"
            placeholder="需求"
            filterable
            :loading="loadingReqs"
            :disabled="!tcForm.version"
            class="sel-req"
          >
            <el-option v-for="r in requirements" :key="r" :label="r" :value="r" />
          </el-select>
          <el-button :loading="loadingTC" :disabled="!canLoadTC" @click="loadTC">加载</el-button>
          <el-tag v-if="tcLoaded" type="success" size="small">{{ tcLoaded.case_count }} 条</el-tag>
          <span class="setup-row-spacer"></span>
          <el-button type="primary" :loading="analyzing" :disabled="!canAnalyze" @click="analyze">
            AI 分析
          </el-button>
        </div>
      </div>

      <div class="setup-block">
        <div class="setup-label">变更来源</div>
        <el-radio-group v-model="changeMode" class="mode-row">
          <el-radio value="mr">GitLab MR</el-radio>
          <el-radio value="desc">口述变更</el-radio>
        </el-radio-group>
        <el-input
          v-if="changeMode === 'mr'"
          v-model="form.gitlab_mr_url"
          placeholder="https://gitlab.com/group/project/-/merge_requests/123"
          clearable
        />
        <el-input
          v-else
          v-model="form.change_description"
          placeholder="口述变更，如：Redis 超时调整、Nacos 开关、/v2/platform/bind 校验加强"
          clearable
        />
      </div>
    </section>

    <section v-if="result" class="results">
      <p class="summary">{{ result.summary }}</p>

      <div v-if="result.ai_summary" class="ai-block">
        <h4>{{ changeMode === 'desc' ? '测试点解读' : '变更解读' }}</h4>
        <div class="ai-body" v-html="aiSummaryHtml" />
      </div>

      <div class="stats">
        <span v-if="result.source !== 'description'"><strong>{{ result.stats.changed_file_count }}</strong> 文件</span>
        <span><strong>{{ result.stats.inferred_api_count }}</strong> 接口</span>
        <span><strong>{{ result.stats.recommended_tc }}</strong> 用例</span>
        <span><strong>{{ result.stats.recommended_api }}</strong> 平台接口</span>
        <el-button
          type="success"
          size="small"
          :loading="running"
          :disabled="!envId || !result.run_plan?.api_ids?.length"
          @click="runRecommended"
        >
          执行推荐接口
        </el-button>
        <el-button
          v-if="changeMode === 'mr'"
          type="primary"
          size="small"
          plain
          :loading="previewingComment"
          :disabled="!form.gitlab_mr_url.trim()"
          @click="openMRCommentPreview"
        >
          提交到 MR 评论
        </el-button>
        <span v-if="!envId" class="hint">请先选择运行环境</span>
      </div>

      <el-tabs v-model="resultTab">
        <el-tab-pane :label="`推荐用例 (${result.recommended_tcs?.length || 0})`" name="tc">
          <p v-if="result.recommended_tc_reason" class="tc-reason">{{ result.recommended_tc_reason }}</p>
          <el-table :data="result.recommended_tcs || []" size="small" border empty-text="无强相关用例（得分 >1）">
            <el-table-column prop="tc_id" label="ID" width="72" />
            <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
            <el-table-column prop="priority" label="优先级" width="72" />
            <el-table-column prop="score" label="分" width="56" />
          </el-table>
        </el-tab-pane>
        <el-tab-pane :label="`推荐接口 (${result.recommended_apis?.length || 0})`" name="api">
          <el-table :data="result.recommended_apis || []" size="small" border empty-text="无">
            <el-table-column label="方法" width="64">
              <template #default="{ row }"><MethodBadge :method="row.method" /></template>
            </el-table-column>
            <el-table-column prop="path" label="路径" min-width="140" show-overflow-tooltip />
            <el-table-column prop="name" label="名称" min-width="100" show-overflow-tooltip />
            <el-table-column label="就绪" width="64" align="center">
              <template #default="{ row }">
                <el-tag :type="row.scenario_ready ? 'success' : 'warning'" size="small">
                  {{ row.scenario_ready ? '是' : '否' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="score" label="分" width="56" />
            <el-table-column label="" width="56">
              <template #default="{ row }">
                <router-link :to="{ path: '/apis', query: { api_id: row.api_id } }">打开</router-link>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane
          v-if="result.source !== 'description'"
          :label="`变更文件 (${result.changed_files?.length || 0})`"
          name="files"
        >
          <ul class="file-list">
            <li v-for="(f, i) in result.changed_files" :key="i">
              <code>{{ f.path }}</code>
              <el-tag v-if="f.status" size="small" type="info">{{ f.status }}</el-tag>
            </li>
          </ul>
        </el-tab-pane>
        <el-tab-pane v-if="result.gaps?.length" :label="`提示 (${result.gaps.length})`" name="gaps">
          <ul class="gap-list">
            <li v-for="(g, i) in result.gaps" :key="i">
              <el-tag size="small" type="warning">{{ g.type }}</el-tag>
              {{ g.action }}
            </li>
          </ul>
        </el-tab-pane>
      </el-tabs>
    </section>

    <el-empty v-else class="empty" description="选择分支 → 版本 → 需求并加载用例，再填写变更来源后 AI 分析" />

    <RunResultDialog v-if="lastBatchRun" v-model="batchDialogOpen" :run="lastBatchRun" :env-name="envName" />

    <el-dialog v-model="mrCommentDialogOpen" title="MR 评论预览" width="720px" destroy-on-close>
      <p class="mr-preview-hint">确认后将作为评论发布到 GitLab MR：<code>{{ form.gitlab_mr_url }}</code></p>
      <pre class="mr-preview">{{ mrCommentMarkdown }}</pre>
      <template #footer>
        <el-button @click="mrCommentDialogOpen = false">取消</el-button>
        <el-button type="primary" :loading="postingComment" @click="confirmPostMRComment">确认提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, inject, onMounted } from 'vue'
import { marked } from 'marked'
import { ElMessage } from 'element-plus'
import { api } from '@/api/client'
import MethodBadge from '@/components/MethodBadge.vue'
import RunResultDialog from '@/components/RunResultDialog.vue'

const { envId, environments } = inject('appStore')

const tcForm = ref({ branch: '', version: '', requirement_id: '' })
const branches = ref([])
const versions = ref([])
const requirements = ref([])
const loadingBranches = ref(false)
const loadingCatalog = ref(false)
const loadingReqs = ref(false)

const changeMode = ref('mr')
const form = ref({
  gitlab_mr_url: '',
  change_description: ''
})
const resultTab = ref('tc')
const loadingTC = ref(false)
const analyzing = ref(false)
const running = ref(false)
const postingComment = ref(false)
const previewingComment = ref(false)
const mrCommentDialogOpen = ref(false)
const mrCommentMarkdown = ref('')
const tcLoaded = ref(null)
const casesJson = ref('')
const result = ref(null)
const batchDialogOpen = ref(false)
const lastBatchRun = ref(null)

const envName = computed(() => environments.value.find((e) => e.id === envId.value)?.name || '')

const canLoadTC = computed(
  () => !!tcForm.value.branch && !!tcForm.value.version && !!tcForm.value.requirement_id.trim()
)

const canAnalyze = computed(() => {
  if (!casesJson.value) return false
  if (changeMode.value === 'mr') return !!form.value.gitlab_mr_url.trim()
  return !!form.value.change_description.trim()
})

const aiSummaryHtml = computed(() => {
  const md = result.value?.ai_summary || ''
  return md ? marked.parse(md) : ''
})

function catalogParams(extra = {}) {
  return { ref: tcForm.value.branch, ...extra }
}

async function fetchBranches() {
  loadingBranches.value = true
  try {
    const res = await api.listTestDocsBranches()
    branches.value = (res.branches || []).map((b) => b.name || b)
    const def = res.default || branches.value[0] || ''
    if (def && branches.value.includes(def)) {
      tcForm.value.branch = def
      await fetchVersions()
    }
  } catch (e) {
    branches.value = []
    ElMessage.warning(e.message || '加载分支失败')
  } finally {
    loadingBranches.value = false
  }
}

async function fetchVersions() {
  if (!tcForm.value.branch) {
    versions.value = []
    return
  }
  loadingCatalog.value = true
  try {
    const cat = await api.listTestDocsCatalog(catalogParams())
    versions.value = cat.versions || []
    if (tcForm.value.version && !versions.value.includes(tcForm.value.version)) {
      tcForm.value.version = ''
    }
    if (!tcForm.value.version && versions.value.length) {
      tcForm.value.version = versions.value[0]
      await fetchRequirements()
    }
  } catch {
    versions.value = []
  } finally {
    loadingCatalog.value = false
  }
}

async function fetchRequirements() {
  const ver = tcForm.value.version
  if (!ver || !tcForm.value.branch) {
    requirements.value = []
    return
  }
  loadingReqs.value = true
  try {
    const cat = await api.listTestDocsCatalog(catalogParams({ version: ver }))
    requirements.value = cat.requirements || []
    if (tcForm.value.requirement_id && !requirements.value.includes(tcForm.value.requirement_id)) {
      tcForm.value.requirement_id = ''
    }
  } catch {
    requirements.value = []
  } finally {
    loadingReqs.value = false
  }
}

function onBranchChange() {
  tcForm.value.version = ''
  tcForm.value.requirement_id = ''
  tcLoaded.value = null
  casesJson.value = ''
  versions.value = []
  requirements.value = []
  fetchVersions()
}

function onVersionChange() {
  tcForm.value.requirement_id = ''
  tcLoaded.value = null
  casesJson.value = ''
  fetchRequirements()
}

async function loadTC() {
  if (!canLoadTC.value) return
  loadingTC.value = true
  try {
    tcLoaded.value = await api.loadTestCases({
      ref: tcForm.value.branch,
      version: tcForm.value.version,
      requirement_id: tcForm.value.requirement_id.trim()
    })
    casesJson.value = tcLoaded.value.cases_json || ''
    ElMessage.success(`已加载 ${tcLoaded.value.case_count} 条用例`)
  } finally {
    loadingTC.value = false
  }
}

function buildAnalyzeBody() {
  const body = {
    version: tcLoaded.value?.version || tcForm.value.version,
    requirement_id: tcLoaded.value?.requirement_id || tcForm.value.requirement_id,
    cases_json: casesJson.value,
    use_ai: true
  }
  if (changeMode.value === 'mr') {
    body.gitlab_mr_url = form.value.gitlab_mr_url.trim()
  } else {
    body.change_description = form.value.change_description.trim()
  }
  return body
}

async function analyze() {
  analyzing.value = true
  try {
    result.value = await api.impactAnalyze(buildAnalyzeBody())
    resultTab.value = 'tc'
    ElMessage.success('AI 分析完成')
  } finally {
    analyzing.value = false
  }
}

function mrCommentPayload() {
  return {
    gitlab_mr_url: form.value.gitlab_mr_url.trim(),
    version: tcForm.value.version,
    requirement_id: tcForm.value.requirement_id,
    tc_docs_branch: tcForm.value.branch,
    result: result.value
  }
}

async function openMRCommentPreview() {
  if (!result.value || !form.value.gitlab_mr_url.trim()) return
  previewingComment.value = true
  try {
    const res = await api.impactPreviewMRComment(mrCommentPayload())
    mrCommentMarkdown.value = res.markdown || ''
    mrCommentDialogOpen.value = true
  } finally {
    previewingComment.value = false
  }
}

async function confirmPostMRComment() {
  postingComment.value = true
  try {
    const res = await api.impactPostMRComment(mrCommentPayload())
    ElMessage.success('已提交到 MR 评论')
    mrCommentDialogOpen.value = false
    if (res.note_url) {
      window.open(res.note_url, '_blank', 'noopener')
    }
  } finally {
    postingComment.value = false
  }
}

async function runRecommended() {
  if (!envId.value || !result.value?.run_plan) return
  running.value = true
  try {
    const res = await api.impactRunPlan({
      env_id: envId.value,
      api_ids: result.value.run_plan.api_ids || [],
      scenario_ids: result.value.run_plan.scenario_ids || []
    })
    ElMessage.success(`通过 ${res.passed} / 失败 ${res.failed}`)
    if (res.runs?.length) {
      const last = res.runs[res.runs.length - 1]
      if (last.run_id) {
        try {
          lastBatchRun.value = await api.getRun(last.run_id)
          batchDialogOpen.value = true
        } catch {
          /* ignore */
        }
      }
    }
  } finally {
    running.value = false
  }
}

onMounted(fetchBranches)
</script>

<style scoped lang="scss">
.impact-page {
  flex: 1;
  margin: 12px 16px 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: auto;
}

.setup {
  padding: 16px 20px;
  border-bottom: 1px solid var(--color-border);
}

.setup-block + .setup-block {
  margin-top: 14px;
}

.setup-label {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
}

.setup-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.setup-row-spacer {
  flex: 1;
  min-width: 12px;
}

.sel-branch {
  width: 200px;
}

.sel-version {
  width: 110px;
}

.sel-req {
  width: 180px;
}

.mode-row {
  margin-bottom: 10px;
}

.results {
  padding: 0 20px 20px;
}

.summary {
  margin: 0 0 12px;
  font-size: 14px;
}

.ai-block {
  margin-bottom: 16px;

  h4 {
    margin: 0 0 8px;
    font-size: 13px;
    font-weight: 600;
  }
}

.ai-body {
  font-size: 13px;
  padding: 12px 14px;
  background: var(--color-bg);
  border-radius: 8px;
  line-height: 1.6;
}

.stats {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
  margin-bottom: 14px;
  font-size: 13px;
  color: var(--color-muted);

  strong {
    color: var(--color-text);
    font-weight: 700;
  }
}

.file-list,
.gap-list {
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: 13px;

  li {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
  }
}

.tc-reason {
  margin: 0 0 10px;
  font-size: 13px;
  line-height: 1.55;
  color: var(--color-muted);
}

.empty {
  padding: 48px 0;
}

.mr-preview-hint {
  margin: 0 0 12px;
  font-size: 13px;
  color: var(--color-muted);
  word-break: break-all;
}

.mr-preview {
  margin: 0;
  max-height: 420px;
  overflow: auto;
  padding: 12px 14px;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  background: var(--color-bg);
  border-radius: 8px;
  border: 1px solid var(--color-border);
}
</style>
