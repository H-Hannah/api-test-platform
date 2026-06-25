<template>
  <el-tooltip
    v-if="clip"
    :content="full"
    placement="top"
    :show-after="200"
    popper-class="overflow-text-tooltip"
  >
    <span class="overflow-text overflow-text--clip" :class="klass">{{ short }}</span>
  </el-tooltip>
  <span v-else class="overflow-text" :class="klass">{{ short }}</span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  value: { type: [String, Number], default: '' },
  max: { type: Number, default: 48 },
  empty: { type: String, default: '—' },
  klass: { type: String, default: '' }
})

const full = computed(() => {
  const t = props.value == null ? '' : String(props.value)
  return t
})

const clip = computed(() => full.value.length > props.max)

const short = computed(() => {
  const t = full.value
  if (!t) return props.empty
  if (t.length <= props.max) return t
  return t.slice(0, props.max) + '...'
})
</script>

<style scoped>
.overflow-text {
  display: inline-block;
  max-width: 100%;
  vertical-align: bottom;
}

.overflow-text--clip {
  cursor: help;
  border-bottom: 1px dotted var(--color-muted, #8f959e);
}
</style>

<style>
.overflow-text-tooltip {
  max-width: 420px;
  word-break: break-all;
  white-space: pre-wrap;
  font-family: ui-monospace, monospace;
  font-size: 12px;
  line-height: 1.45;
}
</style>
