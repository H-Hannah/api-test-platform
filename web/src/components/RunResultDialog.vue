<template>
  <el-dialog
    v-model="open"
    :title="dialogTitle"
    width="520px"
    align-center
    class="app-dialog run-result-dialog"
    destroy-on-close
    @closed="$emit('closed')"
  >
    <template v-if="run">
      <div class="run-summary-bar">
        <el-tag :type="run.status === 'passed' ? 'success' : 'danger'" size="small">
          {{ run.status === 'passed' ? '通过' : '失败' }}
        </el-tag>
        <el-tag v-if="envName" class="env-tag" size="small">{{ envName }}</el-tag>
        <span v-if="step?.duration_ms != null" class="duration">{{ step.duration_ms }} ms</span>
      </div>

      <p v-if="step?.error_message" class="err-msg">{{ step.error_message }}</p>

      <h4 class="section-title">断言结果</h4>
      <el-table :data="assertions" size="small" border empty-text="无断言">
        <el-table-column label="结果" width="64" align="center">
          <template #default="{ row }">
            <span :class="row.passed ? 'status-pass' : 'status-fail'">{{ row.passed ? '✓' : '✗' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="96" />
        <el-table-column prop="expression" label="表达式" min-width="100" show-overflow-tooltip />
        <el-table-column label="实际 / 期望" min-width="160">
          <template #default="{ row }">
            <div class="assert-pair">
              <OverflowText :value="row.actual" :max="40" klass="mono" />
              <span class="muted sep"> / </span>
              <OverflowText :value="row.expected" :max="24" klass="mono" />
            </div>
          </template>
        </el-table-column>
      </el-table>

      <h4 class="section-title">请求</h4>
      <pre class="json-block">{{ pretty(requestSnap) }}</pre>

      <h4 class="section-title">响应</h4>
      <pre class="json-block">{{ pretty(responseSnap) }}</pre>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'
import { parseAssertionResults, parseSnapshot } from '@/utils/notify'
import OverflowText from '@/components/OverflowText.vue'

const open = defineModel({ type: Boolean, default: false })
const props = defineProps({
  run: { type: Object, default: null },
  envName: { type: String, default: '' },
  apiName: { type: String, default: '' }
})
defineEmits(['closed'])

const step = computed(() => props.run?.steps?.[0])
const assertions = computed(() =>
  parseAssertionResults(step.value?.assertion_results)
)
const requestSnap = computed(() => parseSnapshot(step.value?.request_snapshot))
const responseSnap = computed(() => formatResponseDisplay(parseSnapshot(step.value?.response_snapshot)))

const dialogTitle = computed(() => {
  const name = props.apiName || step.value?.name || '接口'
  return `执行结果 · ${name}`
})

/** 展示响应：优先 json 业务体，兼容旧版 { status, body: "..." } */
function formatResponseDisplay(snap) {
  if (!snap) return null
  if (snap.json != null) {
    return {
      status_code: snap.status_code ?? snap.status,
      ...(typeof snap.json === 'object' && snap.json !== null ? snap.json : { data: snap.json }),
      ...(snap.truncated ? { _truncated: true } : {})
    }
  }
  return snap
}

function pretty(val) {
  if (val == null) return '（空）'
  try {
    return JSON.stringify(val, null, 2)
  } catch {
    return String(val)
  }
}
</script>

<style scoped>
.run-summary-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}
.env-tag {
  font-size: 12px;
  color: var(--color-primary);
  background: #eff6ff;
  padding: 2px 8px;
  border-radius: 4px;
}
.duration {
  font-size: 12px;
  color: var(--color-muted);
  margin-left: auto;
}
.err-msg {
  margin: 0 0 12px;
  padding: 8px 10px;
  background: #fef2f2;
  border-radius: 6px;
  font-size: 12px;
  color: #b91c1c;
}
.section-title {
  margin: 16px 0 8px;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-muted);
}
.section-title:first-of-type {
  margin-top: 0;
}
.mono {
  font-family: ui-monospace, monospace;
  font-size: 11px;
}
.muted {
  color: var(--color-muted);
}

.assert-pair {
  display: flex;
  align-items: center;
  gap: 0;
  min-width: 0;
  line-height: 1.4;
}

.assert-pair .sep {
  flex-shrink: 0;
  padding: 0 2px;
}

.assert-pair :deep(.overflow-text) {
  min-width: 0;
}
</style>
