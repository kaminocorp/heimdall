<script setup lang="ts">
import { computed, ref, onBeforeUnmount } from 'vue'
import type { Connection } from '@/types/connection'
import StatusBadge from '@/components/common/StatusBadge.vue'

const props = defineProps<{
  connection: Connection
  testing?: boolean
}>()

const emit = defineEmits<{
  delete: [id: string]
  edit: [connection: Connection]
  test: [id: string]
  'manage-repos': [id: string]
}>()

const confirmingDelete = ref(false)
let deleteTimer: ReturnType<typeof setTimeout> | null = null

function handleDelete() {
  if (!confirmingDelete.value) {
    confirmingDelete.value = true
    // Auto-reset after 3 seconds if user doesn't confirm.
    deleteTimer = setTimeout(() => { confirmingDelete.value = false }, 3000)
    return
  }
  if (deleteTimer) { clearTimeout(deleteTimer); deleteTimer = null }
  confirmingDelete.value = false
  emit('delete', props.connection.id)
}

onBeforeUnmount(() => {
  if (deleteTimer) { clearTimeout(deleteTimer); deleteTimer = null }
})

const typeLabels: Record<string, string> = {
  postgres: 'PostgreSQL',
  supabase: 'Supabase',
  webhook_logs: 'Webhook Logs',
  syslog: 'Syslog',
  github: 'GitHub',
}

const displayType = computed(() => typeLabels[props.connection.type] ?? props.connection.type)

const subtitle = computed(() => {
  if (props.connection.type === 'supabase') {
    const tables = props.connection.config?.poll_tables
    if (Array.isArray(tables) && tables.length > 0) {
      return `Polling ${tables.length} table${tables.length === 1 ? '' : 's'}`
    }
    return 'Polling logs'
  }
  return props.connection.direction
})
</script>

<template>
  <div class="group border border-border rounded-lg p-5 bg-bg-surface hover:border-border-hover transition-colors" :class="{ 'glow-active': connection.status === 'active' }">
    <div class="flex items-start justify-between">
      <div>
        <h3 class="font-mono text-sm font-medium uppercase tracking-wider text-text-primary">{{ connection.name }}</h3>
        <p class="font-mono text-xs text-text-muted mt-1">{{ displayType }} &middot; {{ subtitle }}</p>
      </div>
      <div class="flex items-center gap-3">
        <span v-if="testing" class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full font-mono text-[10px] font-medium uppercase tracking-wider border border-accent/30 bg-accent/10 text-accent">
          <span class="w-1.5 h-1.5 rounded-full bg-accent animate-pulse" />
          Testing
        </span>
        <StatusBadge v-else :status="connection.status" />
        <div class="flex items-center gap-2 sm:opacity-0 sm:group-hover:opacity-100 focus-within:opacity-100 transition-all">
          <button
            v-if="connection.type === 'github'"
            @click="emit('manage-repos', connection.id)"
            class="font-mono text-xs uppercase tracking-wider text-text-muted hover:text-accent focus:text-accent focus:outline-none transition-colors cursor-pointer"
          >
            Repos
          </button>
          <button
            @click="emit('test', connection.id)"
            class="font-mono text-xs uppercase tracking-wider text-text-muted hover:text-accent focus:text-accent focus:outline-none transition-colors cursor-pointer"
          >
            Ping
          </button>
          <button
            @click="emit('edit', connection)"
            class="font-mono text-xs uppercase tracking-wider text-text-muted hover:text-accent focus:text-accent focus:outline-none transition-colors cursor-pointer"
          >
            Edit
          </button>
          <button
            @click="handleDelete"
            class="font-mono text-xs uppercase tracking-wider transition-colors cursor-pointer focus:outline-none"
            :class="confirmingDelete
              ? 'text-status-critical font-medium'
              : 'text-text-muted hover:text-status-critical focus:text-status-critical'"
          >
            {{ confirmingDelete ? 'Confirm?' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
