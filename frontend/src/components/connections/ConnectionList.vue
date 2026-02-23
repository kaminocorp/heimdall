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
  <div v-if="connections.length === 0" class="text-gray-500 text-sm">
    No connections yet. Add one to get started.
  </div>
  <div v-else class="space-y-3">
    <ConnectionCard
      v-for="conn in connections"
      :key="conn.id"
      :connection="conn"
      @delete="emit('delete', $event)"
    />
  </div>
</template>
