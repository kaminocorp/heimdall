<script setup lang="ts">
import { ref } from 'vue'

const severity = ref('')
const emit = defineEmits<{
  filter: [filters: Record<string, string>]
}>()

function handleFilter() {
  emit('filter', { severity: severity.value })
}
</script>

<template>
  <div class="flex gap-2">
    <select v-model="severity" @change="handleFilter" class="border rounded px-3 py-1 text-sm">
      <option value="">All severities</option>
      <option value="info">Info</option>
      <option value="warning">Warning</option>
      <option value="error">Error</option>
    </select>
  </div>
</template>
