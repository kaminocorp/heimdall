<script setup lang="ts">
import { computed, ref, onBeforeUnmount } from 'vue'
import type { Connection } from '@/types/connection'
import StatusBadge from '@/components/common/StatusBadge.vue'

const props = defineProps<{
  connection: Connection
  testing?: boolean
  icon: string
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
  webhook_logs: 'Webhook',
  syslog: 'Syslog',
  github: 'GitHub',
}

const displayType = computed(() => typeLabels[props.connection.type] ?? props.connection.type)
</script>

<template>
  <div
    class="group flex items-center gap-3 px-3 py-2 rounded border border-border bg-bg-surface hover:border-border-hover transition-colors"
    :class="{ 'glow-active': connection.status === 'active' }"
  >
    <!-- Icon badge -->
    <div
      class="shrink-0 w-7 h-7 rounded flex items-center justify-center font-mono text-[10px] font-bold uppercase tracking-wider bg-accent/10 text-accent border border-accent/20"
    >
      {{ icon }}
    </div>

    <!-- Name + type -->
    <div class="min-w-0 flex-1">
      <p class="font-mono text-xs font-medium uppercase tracking-wider text-text-primary truncate">
        {{ connection.name }}
      </p>
      <p class="font-mono text-[10px] text-text-muted truncate">{{ displayType }}</p>
    </div>

    <!-- Status -->
    <span v-if="testing" class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full font-mono text-[9px] font-medium uppercase tracking-wider border border-accent/30 bg-accent/10 text-accent">
      <span class="w-1 h-1 rounded-full bg-accent animate-pulse" />
      Testing
    </span>
    <StatusBadge v-else :status="connection.status" />

    <!-- Hover actions -->
    <div class="flex items-center gap-1.5 sm:opacity-0 sm:group-hover:opacity-100 focus-within:opacity-100 transition-all">
      <button
        v-if="connection.type === 'github'"
        @click.stop="emit('manage-repos', connection.id)"
        class="font-mono text-[10px] uppercase tracking-wider text-text-muted hover:text-accent focus:text-accent focus:outline-none transition-colors cursor-pointer"
      >
        Repos
      </button>
      <button
        @click.stop="emit('test', connection.id)"
        class="font-mono text-[10px] uppercase tracking-wider text-text-muted hover:text-accent focus:text-accent focus:outline-none transition-colors cursor-pointer"
      >
        Ping
      </button>
      <button
        @click.stop="emit('edit', connection)"
        class="font-mono text-[10px] uppercase tracking-wider text-text-muted hover:text-accent focus:text-accent focus:outline-none transition-colors cursor-pointer"
      >
        Edit
      </button>
      <button
        @click.stop="handleDelete"
        class="font-mono text-[10px] uppercase tracking-wider transition-colors cursor-pointer focus:outline-none"
        :class="confirmingDelete
          ? 'text-status-critical font-medium'
          : 'text-text-muted hover:text-status-critical focus:text-status-critical'"
      >
        {{ confirmingDelete ? 'Confirm?' : 'Del' }}
      </button>
    </div>
  </div>
</template>
