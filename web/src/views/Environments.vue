<template>
  <div class="env-page panel-card">
    <div class="page-toolbar">
      <p class="toolbar-hint">
        在此维护各环境的域名与 <code>token</code>，执行接口时会替换
        <code v-pre>{{base_url_*}}</code>、<code v-pre>{{token}}</code> 等占位符。
      </p>
      <span class="spacer"></span>
      <el-button type="primary" @click="openCreate">
        <el-icon><Plus /></el-icon>
        新建环境
      </el-button>
    </div>

    <el-table v-loading="loading" :data="list" height="100%" empty-text="暂无环境，请新建">
      <el-table-column prop="name" label="环境名称" width="120">
        <template #default="{ row }">
          <span class="env-name">{{ row.name }}</span>
          <el-tag v-if="row.is_default" size="small" type="success" class="def-tag">默认</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="base_url" label="主站 base_url" min-width="220" show-overflow-tooltip />
      <el-table-column label="鉴权 token" width="110" align="center">
        <template #default="{ row }">
          <el-tag :type="row._tokenOk ? 'success' : 'warning'" size="small">
            {{ row._tokenOk ? '已配置' : '未配置' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="服务域名" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">
          <span class="muted">{{ row._serviceSummary || '—' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" align="center" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" size="small" :disabled="list.length <= 1" @click="onDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-drawer
      v-model="drawerOpen"
      :title="editingId ? `编辑环境 · ${form.name}` : '新建环境'"
      size="520px"
      direction="rtl"
      destroy-on-close
      @closed="resetForm"
    >
      <el-form label-width="108px" class="env-form" @submit.prevent>
        <el-form-item label="环境名称" required>
          <el-input v-model="form.name" placeholder="如 BETA、PRE、PROD" :disabled="!!editingId && isBuiltinName(form.name)" />
        </el-form-item>
        <el-form-item label="主站 base_url" required>
          <el-input v-model="form.base_url" placeholder="https://api.example.com" />
          <p class="field-hint">写入 <code>base_url</code> 列，并同步到 variables.base_url</p>
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="form.is_default" />
        </el-form-item>

        <el-divider content-position="left">鉴权</el-divider>
        <el-form-item label="token">
          <el-input
            v-model="form.token"
            type="password"
            show-password
            placeholder="Bearer 令牌，对应接口头中的 {{token}}"
            autocomplete="new-password"
          />
          <p class="field-hint">仅保存在本环境 variables，不会写入接口定义表</p>
        </el-form-item>

        <el-divider v-if="serviceKeys.length" content-position="left">多服务域名</el-divider>
        <el-form-item
          v-for="sk in serviceKeys"
          :key="sk.key"
          :label="sk.label"
        >
          <el-input v-model="form.vars[sk.key]" :placeholder="form.base_url || '留空则执行时回退到主站 base_url'" />
          <p v-if="sk.hint" class="field-hint">{{ sk.hint }} · <code>{{ sk.key }}</code></p>
        </el-form-item>

        <el-divider content-position="left">其它变量</el-divider>
        <div class="custom-vars">
          <div v-for="(row, idx) in form.customRows" :key="idx" class="custom-row">
            <el-input v-model="row.key" placeholder="变量名" class="custom-key" />
            <el-input v-model="row.value" placeholder="值" class="custom-val" />
            <el-button link type="danger" @click="form.customRows.splice(idx, 1)">移除</el-button>
          </div>
          <el-button link type="primary" @click="addCustomRow">+ 添加变量</el-button>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="drawerOpen = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '@/composables/useAppStore'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api/client'
import {
  serviceKeysForProduct,
  parseVariables,
  tokenConfigured,
  buildVariablesPayload
} from '@/utils/envPresets'

const RESERVED_KEYS = new Set(['base_url', 'token'])

const { loadEnvironments } = useAppStore()

const loading = ref(false)
const list = ref([])
const drawerOpen = ref(false)
const saving = ref(false)
const editingId = ref(null)

const form = ref(emptyForm())

const serviceKeys = computed(() => serviceKeysForProduct())

function emptyForm() {
  return {
    name: '',
    base_url: '',
    is_default: false,
    token: '',
    vars: {},
    customRows: []
  }
}

function isBuiltinName(name) {
  return ['BETA', 'PRE', 'PROD'].includes((name || '').toUpperCase())
}

function enrichRow(e) {
  const vars = parseVariables(e.variables)
  const keys = serviceKeysForProduct()
  const parts = keys
    .map((sk) => vars[sk.key])
    .filter((v) => v && String(v).trim())
  return {
    ...e,
    _tokenOk: tokenConfigured(vars),
    _serviceSummary: parts.length ? `${parts.length} 项已填` : '使用主站回退'
  }
}

async function load() {
  loading.value = true
  try {
    const raw = await api.listEnvironments()
    list.value = raw.map(enrichRow)
  } finally {
    loading.value = false
  }
}

function resetForm() {
  editingId.value = null
  form.value = emptyForm()
}

function fillFormFromEnv(row) {
  const vars = parseVariables(row.variables)
  const keys = serviceKeysForProduct()
  const varsForm = {}
  for (const sk of keys) {
    varsForm[sk.key] = vars[sk.key] || ''
  }
  const customRows = []
  for (const [k, v] of Object.entries(vars)) {
    if (RESERVED_KEYS.has(k) || keys.some((sk) => sk.key === k)) continue
    customRows.push({ key: k, value: v })
  }
  form.value = {
    name: row.name,
    base_url: row.base_url || vars.base_url || '',
    is_default: !!row.is_default,
    token: vars.token || '',
    vars: varsForm,
    customRows
  }
}

function openEdit(row) {
  editingId.value = row.id
  fillFormFromEnv(row)
  drawerOpen.value = true
}

function openCreate() {
  resetForm()
  form.value.name = ''
  form.value.is_default = list.value.length === 0
  drawerOpen.value = true
}

function addCustomRow() {
  form.value.customRows.push({ key: '', value: '' })
}

async function save() {
  const name = (form.value.name || '').trim()
  const baseUrl = (form.value.base_url || '').trim()
  if (!name) {
    ElMessage.warning('请填写环境名称')
    return
  }
  if (!baseUrl) {
    ElMessage.warning('请填写主站 base_url')
    return
  }
  const payload = {
    name,
    base_url: baseUrl.replace(/\/+$/, ''),
    is_default: form.value.is_default,
    variables: buildVariablesPayload(form.value)
  }
  saving.value = true
  try {
    if (editingId.value) {
      await api.updateEnvironment(editingId.value, payload)
      ElMessage.success('环境已更新')
    } else {
      await api.createEnvironment(payload)
      ElMessage.success('环境已创建')
    }
    drawerOpen.value = false
    await load()
    await loadEnvironments?.()
  } finally {
    saving.value = false
  }
}

async function onDelete(row) {
  try {
    await ElMessageBox.confirm(`确定删除环境「${row.name}」？`, '删除环境', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
    await api.deleteEnvironment(row.id)
    ElMessage.success('已删除')
    await load()
    await loadEnvironments?.()
  } catch {
    /* cancel */
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

.toolbar-hint {
  margin: 0;
  font-size: 13px;
  color: var(--color-muted);
  max-width: 520px;
  line-height: 1.5;

  code {
    font-size: 12px;
    background: #f3f4f6;
    padding: 1px 4px;
    border-radius: 3px;
  }
}

.env-name {
  font-weight: 600;
  margin-right: 6px;
}

.def-tag {
  vertical-align: middle;
}

.muted {
  color: var(--color-muted);
  font-size: 13px;
}

.env-form {
  padding: 0 4px 16px;
}

.field-hint {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--color-muted);
  line-height: 1.4;

  code {
    font-size: 11px;
  }
}

.custom-vars {
  width: 100%;
  padding-left: 108px;
}

.custom-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}

.custom-key {
  width: 140px;
  flex-shrink: 0;
}

.custom-val {
  flex: 1;
}
</style>
