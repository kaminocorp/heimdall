<script setup lang="ts">
import type { Connection } from '@/types/connection'
import StatusBadge from '@/components/common/StatusBadge.vue'

defineProps<{
  connection: Connection
}>()

const emit = defineEmits<{
  delete: [id: string]
}>()
</script>

<template>
  <div class="group border border-border rounded-lg p-5 bg-bg-surface hover:border-border-hover transition-colors" :class="{ 'glow-active': connection.status === 'active' }">
    <div class="flex items-start justify-between">
      <div>
        <h3 class="font-mono text-sm font-medium uppercase tracking-wider text-text-primary">{{ connection.name }}</h3>
        <p class="font-mono text-xs text-text-muted mt-1">{{ connection.type }} &middot; {{ connection.direction }}</p>
      </div>
      <div class="flex items-center gap-3">
        <StatusBadge :status="connection.status" />
        <button
          @click="emit('delete', connection.id)"
          class="font-mono text-xs uppercase tracking-wider text-text-muted opacity-0 group-hover:opacity-100 hover:text-status-critical transition-all cursor-pointer"
        >
          Delete
        </button>
      </div>
    </div>
  </div>
</template>
