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
    class="w-full text-left flex items-center gap-4 px-4 py-3 border-b border-border transition-colors hover:bg-bg-elevated cursor-pointer group"
    :class="app.status === 'paused' ? 'opacity-70' : ''"
  >
    <!-- Name -->
    <div class="flex-1 min-w-0 flex items-center gap-3">
      <h3 class="font-mono text-sm font-medium text-text-primary truncate group-hover:text-accent-bright transition-colors">
        {{ app.name }}
      </h3>
      <StatusBadge :status="app.status" />
    </div>

    <!-- Connections -->
    <div class="shrink-0 w-28 font-mono text-xs text-text-muted text-right">
      {{ app.connection_count }} {{ app.connection_count === 1 ? 'conn' : 'conns' }}
    </div>

    <!-- Schedules -->
    <div class="shrink-0 w-28 font-mono text-xs text-text-muted text-right">
      {{ app.schedule_count }} {{ app.schedule_count === 1 ? 'schedule' : 'schedules' }}
    </div>

    <!-- Created -->
    <div class="shrink-0 w-28 font-mono text-xs text-text-muted text-right hidden md:block">
      {{ formatDate(app.created_at) }}
    </div>
  </button>
</template>
