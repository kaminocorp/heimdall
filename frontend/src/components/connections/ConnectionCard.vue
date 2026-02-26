<script setup lang="ts">
import type { Connection } from '@/types/connection'
import StatusBadge from '@/components/common/StatusBadge.vue'

defineProps<{
  connection: Connection
  testing?: boolean
}>()

const emit = defineEmits<{
  delete: [id: string]
  edit: [connection: Connection]
  test: [id: string]
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
        <span v-if="testing" class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full font-mono text-[10px] font-medium uppercase tracking-wider border border-accent/30 bg-accent/10 text-accent">
          <span class="w-1.5 h-1.5 rounded-full bg-accent animate-pulse" />
          Testing
        </span>
        <StatusBadge v-else :status="connection.status" />
        <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all">
          <button
            @click="emit('test', connection.id)"
            class="font-mono text-xs uppercase tracking-wider text-text-muted hover:text-accent transition-colors cursor-pointer"
          >
            Ping
          </button>
          <button
            @click="emit('edit', connection)"
            class="font-mono text-xs uppercase tracking-wider text-text-muted hover:text-accent transition-colors cursor-pointer"
          >
            Edit
          </button>
          <button
            @click="emit('delete', connection.id)"
            class="font-mono text-xs uppercase tracking-wider text-text-muted hover:text-status-critical transition-colors cursor-pointer"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
