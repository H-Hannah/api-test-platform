<template>
  <span class="method-tag" :class="cls">{{ method || '—' }}</span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  method: { type: String, default: '' }
})

const cls = computed(() => {
  const m = (props.method || '').toUpperCase()
  const map = { GET: 'method-get', POST: 'method-post', PUT: 'method-put', DELETE: 'method-delete', PATCH: 'method-patch' }
  return map[m] || 'method-default'
})
</script>
