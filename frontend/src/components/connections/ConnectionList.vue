<script setup lang="ts">
import type { Connection } from '@/types/connection'
import ConnectionCard from './ConnectionCard.vue'

defineProps<{
  connections: Connection[]
}>()

const emit = defineEmits<{
  delete: [id: string]
}>()
</script>

<template>
  <div v-if="connections.length === 0" class="text-text-muted text-sm font-mono">
    No connections yet. Add one to get started.
  </div>
  <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-3">
    <ConnectionCard
      v-for="(conn, i) in connections"
      :key="conn.id"
      :connection="conn"
      class="animate-fade-in"
      :style="{ '--stagger-index': i }"
      @delete="emit('delete', $event)"
    />
  </div>
</template>
