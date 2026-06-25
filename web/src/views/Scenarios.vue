<template>
  <div class="scenarios-page panel-card">
    <div class="page-toolbar">
      <el-input v-model="keyword" placeholder="搜索场景名称" clearable style="width: 220px" :prefix-icon="Search" />
      <span class="spacer"></span>
      <el-button type="primary" :disabled="!selected || !envId" :loading="running" @click="runSelected">
        <el-icon><VideoPlay /></el-icon>
        执行场景
      </el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="filtered"
      highlight-current-row
      height="100%"
      @current-change="onSelect"
    >
      <el-table-column prop="name" label="场景名称" min-width="180" show-overflow-tooltip />
      <el-table-column prop="folder_path" label="所属模块" width="180" show-overflow-tooltip />
      <el-table-column prop="description" label="说明" min-width="200" show-overflow-tooltip />
      <el-table-column prop="created_at" label="创建时间" width="170" />
      <el-table-column label="操作" width="100" align="center">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click.stop="openDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-drawer v-model="drawerOpen" :title="detail?.name || '场景详情'" size="480px" direction="rtl">
      <template v-if="detail">
        <p class="desc">{{ detail.description || '暂无说明' }}</p>
        <p class="path-hint" v-if="detail.folder_path">模块：{{ detail.folder_path }}</p>
        <h4 class="steps-title">步骤（{{ detail.steps?.length || 0 }}）</h4>
        <el-timeline>
          <el-timeline-item
            v-for="step in detail.steps"
            :key="step.id"
            :timestamp="`步骤 ${step.step_order}`"
            placement="top"
          >
            <div class="step-card">
              <div class="step-head">
                <MethodBadge :method="step.method" />
                <span class="step-name">{{ step.name }}</span>
              </div>
              <code class="step-path">{{ step.path }}</code>
            </div>
          </el-timeline-item>
        </el-timeline>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, inject, onMounted } from 'vue'
import { Search, VideoPlay } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { api } from '@/api/client'
import { notifyRunResult } from '@/utils/notify'
import MethodBadge from '@/components/MethodBadge.vue'

const { envId, environments } = inject('appStore')

const loading = ref(false)
const list = ref([])
const keyword = ref('')
const selected = ref(null)
const detail = ref(null)
const drawerOpen = ref(false)
const running = ref(false)

const filtered = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  if (!k) return list.value
  return list.value.filter((s) => (s.name || '').toLowerCase().includes(k))
})

async function load() {
  loading.value = true
  try {
    list.value = await api.listScenarios()
  } finally {
    loading.value = false
  }
}

function onSelect(row) {
  selected.value = row
}

async function openDetail(row) {
  selected.value = row
  detail.value = await api.getScenario(row.id)
  drawerOpen.value = true
}

async function runSelected() {
  if (!selected.value || !envId.value) return
  running.value = true
  try {
    const run = await api.runScenario(selected.value.id, envId.value)
    const full = run.id ? await api.getRun(run.id) : run
    const envName = environments.value.find((e) => e.id === envId.value)?.name || ''
    notifyRunResult(full, envName)
  } finally {
    running.value = false
  }
}

onMounted(load)
</script>

<style scoped lang="scss">
.scenarios-page {
  flex: 1;
  margin: 12px 16px 16px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.desc {
  margin: 0 0 8px;
  color: var(--color-muted);
  font-size: 13px;
  line-height: 1.5;
}
.path-hint {
  margin: 0 0 16px;
  font-size: 12px;
  color: var(--color-primary);
}
.steps-title {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
}
.step-card {
  padding: 10px 12px;
  background: #f8f9fb;
  border-radius: 6px;
  border: 1px solid var(--color-border);
}
.step-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.step-name {
  font-weight: 500;
}
.step-path {
  font-size: 12px;
  color: var(--color-muted);
  word-break: break-all;
}
</style>
