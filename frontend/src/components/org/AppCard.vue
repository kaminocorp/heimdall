<script setup lang="ts">
import type { ApplicationWithCounts } from '@/types/organization'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { formatDate } from '@/utils/format'

const props = defineProps<{
  app: ApplicationWithCounts
}>()

const emit = defineEmits<{
  select: [appId: string]
}>()

function handleClick() {
  emit('select', props.app.id)
}
</script>

<template>
  <button
    type="button"
    @click="handleClick"
    class="w-full text-left rounded-lg border border-border bg-bg-surface p-5 transition-all hover:border-border-hover hover:bg-bg-elevated cursor-pointer group"
    :class="app.status === 'paused' ? 'opacity-70' : ''"
  >
    <!-- Header: name + status -->
    <div class="flex items-center justify-between gap-3 mb-3">
      <h3 class="font-mono text-sm font-semibold text-text-primary truncate group-hover:text-accent-bright transition-colors">
        {{ app.name }}
      </h3>
      <StatusBadge :status="app.status" />
    </div>

    <!-- Stats row -->
    <div class="flex items-center gap-4 font-mono text-xs text-text-muted mb-3">
      <div class="flex items-center gap-1.5">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m9.86-2.439a4.5 4.5 0 00-6.364-6.364L4.5 8.25a4.5 4.5 0 006.364 6.364l2.25-2.25" />
        </svg>
        <span>{{ app.connection_count }} {{ app.connection_count === 1 ? 'connection' : 'connections' }}</span>
      </div>
      <div class="flex items-center gap-1.5">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{{ app.schedule_count }} {{ app.schedule_count === 1 ? 'schedule' : 'schedules' }}</span>
      </div>
    </div>

    <!-- Footer: created date -->
    <div class="font-mono text-xs text-text-muted">
      Created {{ formatDate(app.created_at) }}
    </div>
  </button>
</template>
