<template>
  <div class="reports-page panel-card">
    <div class="page-toolbar">
      <span class="toolbar-label">最近执行记录</span>
      <span class="spacer"></span>
      <el-button @click="load">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <el-table v-loading="loading" :data="runs" height="100%" @row-click="openDetail">
      <el-table-column prop="id" label="ID" width="72" />
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="row.scenario_id ? 'warning' : 'primary'">
            {{ row.scenario_id ? '场景' : '接口' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <span :class="statusClass(row.status)">{{ row.status }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="summary" label="摘要" min-width="200" show-overflow-tooltip />
      <el-table-column prop="started_at" label="开始时间" width="170" />
      <el-table-column prop="finished_at" label="结束时间" width="170" />
    </el-table>

    <el-drawer v-model="drawerOpen" title="执行详情" size="520px" direction="rtl">
      <template v-if="detail">
        <div class="run-summary">
          <el-tag :type="detail.status === 'passed' ? 'success' : detail.status === 'failed' ? 'danger' : 'info'">
            {{ detail.status }}
          </el-tag>
          <span>{{ detail.summary }}</span>
        </div>
        <el-table :data="detail.steps || []" size="small" border class="steps-table">
          <el-table-column prop="step_order" label="#" width="48" />
          <el-table-column prop="name" label="步骤" min-width="120" />
          <el-table-column label="状态" width="88">
            <template #default="{ row }">
              <span :class="statusClass(row.status)">{{ row.status }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="duration_ms" label="耗时(ms)" width="96" />
          <el-table-column label="错误" min-width="120" show-overflow-tooltip>
            <template #default="{ row }">{{ row.error_message || '—' }}</template>
          </el-table-column>
        </el-table>
        <div v-if="selectedStep" class="step-detail">
          <h4>请求 / 响应</h4>
          <pre class="json-block">{{ pretty(selectedStep.request_snapshot) }}</pre>
          <pre class="json-block">{{ pretty(selectedStep.response_snapshot) }}</pre>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { api } from '@/api/client'

const loading = ref(false)
const runs = ref([])
const detail = ref(null)
const drawerOpen = ref(false)
const selectedStep = ref(null)

function statusClass(s) {
  if (s === 'passed') return 'status-pass'
  if (s === 'failed') return 'status-fail'
  return 'status-running'
}

function pretty(val) {
  if (!val) return '（空）'
  try {
    return JSON.stringify(JSON.parse(val), null, 2)
  } catch {
    return val
  }
}

async function load() {
  loading.value = true
  try {
    runs.value = await api.listRuns()
  } finally {
    loading.value = false
  }
}

async function openDetail(row) {
  detail.value = await api.getRun(row.id)
  selectedStep.value = detail.value.steps?.[0] || null
  drawerOpen.value = true
}

onMounted(load)
</script>

<style scoped lang="scss">
.reports-page {
  flex: 1;
  margin: 12px 16px 16px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.toolbar-label {
  font-size: 13px;
  color: var(--color-muted);
}

.run-summary {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  font-size: 13px;
}

.steps-table {
  margin-bottom: 16px;
}

.step-detail h4 {
  margin: 0 0 8px;
  font-size: 12px;
  color: var(--color-muted);
}
</style>
