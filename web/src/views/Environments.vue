<template>
  <div class="env-page panel-card">
    <div class="page-toolbar">
      <p class="toolbar-hint">
        配置各环境 URL 与自定义变量（key / value）
      </p>
      <div class="toolbar-actions">
        <template v-if="editing">
          <el-button size="small" type="warning" @click="addRow"> 新增</el-button>
          <el-button size="small" type="success" :loading="saving" @click="saveAll">保存</el-button>
          <el-button size="small" type="info" @click="cancelEdit">取消</el-button>
        </template>
        <el-button v-else type="primary" size="small" @click="startEdit" icon="Edit">编辑变量</el-button>
      </div>
    </div>

    <div v-loading="loading" class="matrix-wrap">
      <table v-if="forms.length" class="env-matrix">
        <thead>
          <tr>
            <th class="col-label">变量</th>
            <th v-for="f in forms" :key="f.id" class="col-env" :class="envToneClass(f.name)">
              <div class="env-head">
                <span class="env-head-name">{{ f.name }}</span>
                <el-tag v-if="f.is_default" size="small" type="success">默认运行环境</el-tag>
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(row, rowIdx) in rowDefs"
            :key="row.id"
            :class="rowIdx % 2 === 0 ? 'row-odd' : 'row-even'"
          >
            <td class="col-label">
              <template v-if="editing">
                <div class="custom-key-cell">
                  <el-input
                    v-model="row.key"
                    placeholder="变量名"
                    size="small"
                    class="custom-key-input key-input-bold"
                  />
                  <el-button
                    link
                    type="danger"
                    size="small"
                    title="删除此行"
                    @click="removeRow(row.id)"
                  >
                    ×
                  </el-button>
                </div>
              </template>
              <template v-else>
                <code class="var-key-only">{{ row.key || '—' }}</code>
              </template>
            </td>
            <td v-for="f in forms" :key="f.id + row.id" class="col-env" :class="envToneClass(f.name)">
              <el-input
                v-if="editing"
                v-model="f.values[row.id]"
                :placeholder="row.placeholder || '值'"
                size="small"
              />
              <span v-else class="cell-text" :class="{ muted: !(f.values[row.id] ?? '').trim() }">
                {{ displayCell(f, row) }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>

      <el-empty v-else-if="!loading" description="暂无环境" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Edit } from '@element-plus/icons-vue'
import { useAppStore } from '@/composables/useAppStore'
import { ElMessage } from 'element-plus'
import { api } from '@/api/client'
import {
  ENV_MATRIX_ROWS,
  ENV_ORDER,
  deriveBaseURL,
  parseVariables,
  buildVariablesPayload,
  isLegacyEnvKey
} from '@/utils/envPresets'

const { loadEnvironments } = useAppStore()

const loading = ref(false)
const editing = ref(false)
const saving = ref(false)
const forms = ref([])
const rowDefs = ref([])
let tableSnapshot = null

function envToneClass(name) {
  const n = (name || '').toUpperCase()
  if (n === 'BETA') return 'env-beta'
  if (n === 'PRE') return 'env-pre'
  if (n === 'PROD') return 'env-prod'
  return 'env-other'
}

function sortEnvs(list) {
  return [...list].sort((a, b) => {
    const oa = ENV_ORDER[a.name?.toUpperCase()] ?? 9
    const ob = ENV_ORDER[b.name?.toUpperCase()] ?? 9
    if (oa !== ob) return oa - ob
    return (a.id || 0) - (b.id || 0)
  })
}

function newRowId() {
  return `r-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`
}

function rowFromPreset(r) {
  return {
    id: `d-${r.key}`,
    key: r.key,
    placeholder: r.placeholder || ''
  }
}

function buildRowDefs(loadedForms) {
  const dbKeys = new Set()
  for (const f of loadedForms) {
    for (const k of Object.keys(f.rawVars)) {
      if (!isLegacyEnvKey(k)) dbKeys.add(k)
    }
  }

  const rows = []
  const seen = new Set()

  for (const r of ENV_MATRIX_ROWS) {
    rows.push(rowFromPreset(r))
    seen.add(r.key)
  }

  for (const k of [...dbKeys].sort((a, b) => a.localeCompare(b))) {
    if (seen.has(k)) continue
    rows.push({
      id: `r-${k}`,
      key: k,
      placeholder: ''
    })
  }
  return rows
}

function applyValues(loadedForms, defs) {
  for (const f of loadedForms) {
    const values = {}
    for (const row of defs) {
      values[row.id] = f.rawVars[row.key] ?? ''
    }
    f.values = values
    delete f.rawVars
  }
}

function envToForm(e) {
  return {
    id: e.id,
    name: e.name,
    is_default: !!e.is_default,
    rawVars: parseVariables(e.variables),
    values: {}
  }
}

function snapshotTable() {
  return {
    rowDefs: rowDefs.value.map((r) => ({ ...r })),
    forms: forms.value.map((f) => ({
      id: f.id,
      values: { ...f.values }
    }))
  }
}

function restoreTable(snap) {
  rowDefs.value = snap.rowDefs.map((r) => ({ ...r }))
  for (const item of snap.forms) {
    const f = forms.value.find((x) => x.id === item.id)
    if (!f) continue
    f.values = { ...item.values }
  }
}

function displayCell(f, row) {
  const v = (f.values[row.id] ?? '').trim()
  return v || '—'
}

function startEdit() {
  tableSnapshot = snapshotTable()
  editing.value = true
}

function cancelEdit() {
  if (tableSnapshot) {
    restoreTable(tableSnapshot)
  }
  tableSnapshot = null
  editing.value = false
}

async function load() {
  loading.value = true
  try {
    const raw = await api.listEnvironments()
    const loaded = sortEnvs(raw).map(envToForm)
    const defs = buildRowDefs(loaded)
    applyValues(loaded, defs)
    rowDefs.value = defs
    forms.value = loaded
    tableSnapshot = null
    editing.value = false
  } finally {
    loading.value = false
  }
}

function buildPayload(f) {
  return {
    name: f.name,
    base_url: deriveBaseURL(f, rowDefs.value),
    is_default: f.is_default,
    variables: buildVariablesPayload(f, rowDefs.value)
  }
}

async function saveAll() {
  if (!forms.value.length) return
  saving.value = true
  try {
    for (const f of forms.value) {
      await api.updateEnvironment(f.id, buildPayload(f))
    }
    tableSnapshot = null
    editing.value = false
    ElMessage.success('已保存')
    await loadEnvironments?.()
  } finally {
    saving.value = false
  }
}

function addRow() {
  if (!editing.value) return
  const id = newRowId()
  rowDefs.value.push({ id, key: '', placeholder: '' })
  for (const f of forms.value) {
    f.values[id] = ''
  }
}

function removeRow(id) {
  rowDefs.value = rowDefs.value.filter((r) => r.id !== id)
  for (const f of forms.value) {
    delete f.values[id]
  }
}

onMounted(load)
</script>

<style scoped lang="scss">
.env-page {
  flex: 1;
  margin: 12px 16px 16px;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.page-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.toolbar-hint {
  margin: 0;
  flex: 1;
  font-size: 13px;
  color: var(--color-muted);
  line-height: 1.5;
}

.toolbar-actions {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 8px;
}

.matrix-wrap {
  flex: 1;
  min-height: 0;
  overflow: auto;
  border-radius: 10px;
  box-shadow: var(--shadow-md);
  background: #f4f6fa;
}

.env-matrix {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 13px;

  th,
  td {
    border-bottom: 1px solid rgba(51, 65, 85, 0.08);
    padding: 11px 14px;
    vertical-align: middle;
    transition: background-color 0.15s ease;
  }

  thead th {
    position: sticky;
    top: 0;
    z-index: 2;
    background: linear-gradient(180deg, #e9eef6 0%, #dfe6f0 100%);
    border-bottom: 2px solid #c8d2e0;
    font-weight: 700;
    color: #1e3a5f;
    box-shadow: 0 2px 6px rgba(30, 58, 95, 0.06);
  }

  .col-label {
    width: 200px;
    min-width: 200px;
    position: sticky;
    left: 0;
    z-index: 1;
    border-right: 1px solid rgba(51, 65, 85, 0.1);
    font-weight: 600;
    box-shadow: 4px 0 8px -4px rgba(30, 58, 95, 0.08);
  }

  thead .col-label {
    z-index: 3;
    background: linear-gradient(180deg, #dde4ee 0%, #d2dae6 100%);
  }

  .col-env {
    min-width: 220px;
  }

  tbody tr.row-odd td.col-label {
    background: #e3e9f2;
    color: #1e3a5f;
  }

  tbody tr.row-even td.col-label {
    background: #d5dde9;
    color: #1e3a5f;
  }

  tbody tr.row-odd td.env-beta {
    background: #f0f5ff;
  }

  tbody tr.row-even td.env-beta {
    background: #e3edff;
  }

  tbody tr.row-odd td.env-pre {
    background: #f3f1fe;
  }

  tbody tr.row-even td.env-pre {
    background: #e8e4fc;
  }

  tbody tr.row-odd td.env-prod {
    background: #edf9f6;
  }

  tbody tr.row-even td.env-prod {
    background: #dff3ee;
  }

  tbody tr.row-odd td.env-other {
    background: #f4f6f8;
  }

  tbody tr.row-even td.env-other {
    background: #e9edf2;
  }

  tbody tr:hover td.col-label {
    background: #cdd6e4;
  }

  tbody tr:hover td.env-beta {
    background: #d6e4ff;
  }

  tbody tr:hover td.env-pre {
    background: #ddd6fc;
  }

  tbody tr:hover td.env-prod {
    background: #ccefdf;
  }

  tbody tr:hover td.env-other {
    background: #dde3eb;
  }

  tbody tr:last-child td {
    border-bottom: none;
  }
}

.env-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.env-head-name {
  font-weight: 700;
  font-size: 14px;
  letter-spacing: 0.02em;
}

.var-key-only {
  font-size: 13px;
  font-weight: 700;
  font-family: ui-monospace, monospace;
  color: var(--color-text);
}

.custom-key-cell {
  display: flex;
  align-items: center;
  gap: 4px;
}

.custom-key-input {
  flex: 1;
  min-width: 0;
}

.key-input-bold :deep(.el-input__inner) {
  font-weight: 700;
  font-family: ui-monospace, monospace;
}

.cell-text {
  display: block;
  line-height: 1.5;
  word-break: break-all;
  font-size: 13px;
  color: #2c3e50;

  &.muted {
    color: #94a3b8;
    font-style: italic;
  }
}
</style>
