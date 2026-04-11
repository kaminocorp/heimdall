<script setup lang="ts">
import { computed } from 'vue'
import type { InvestigationSchedule } from '@/types/schedule'
import { formatSchedule } from '@/utils/cron-presets'

const props = defineProps<{
  schedule: InvestigationSchedule
  running: boolean
}>()

const emit = defineEmits<{
  edit: [schedule: InvestigationSchedule]
  delete: [schedule: InvestigationSchedule]
  run: [schedule: InvestigationSchedule]
}>()

const cadence = computed(() => formatSchedule(props.schedule))

// lastRunLabel renders the last-run timestamp as either "never" or an
// absolute ISO-ish display ("Apr 11, 14:23"). We deliberately avoid
// relative "5m ago" formatting here because the card is static until the
// user interacts with it — relative times would go stale silently.
const lastRunLabel = computed(() => {
  if (!props.schedule.last_run_at) return 'never'
  const d = new Date(props.schedule.last_run_at)
  return d.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
})

// statusColor maps the last_status into a single-token status color so the
// card border-left stripe and the label badge stay synchronised.
const statusColor = computed(() => {
  if (!props.schedule.enabled) return 'text-text-muted'
  if (props.schedule.last_status === 'success') return 'text-status-ok'
  if (props.schedule.last_status === 'error') return 'text-status-critical'
  return 'text-text-secondary'
})

const statusLabel = computed(() => {
  if (!props.schedule.enabled) return 'Disabled'
  if (props.schedule.last_status === 'success') return 'Success'
  if (props.schedule.last_status === 'error') return 'Error'
  return 'Pending'
})
</script>

<template>
  <div class="border border-border rounded-lg bg-bg-surface px-5 py-4 flex flex-col gap-3">
    <!-- Top row: name + cadence + status + actions -->
    <div class="flex items-start justify-between gap-4">
      <div class="min-w-0 flex-1">
        <div class="flex items-baseline gap-3">
          <h3 class="font-mono text-sm font-bold text-text-primary truncate">
            {{ schedule.name }}
          </h3>
          <span class="font-mono text-xs uppercase tracking-wider" :class="statusColor">
            {{ statusLabel }}
          </span>
        </div>
        <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 font-mono text-xs text-text-muted">
          <span>{{ cadence }}</span>
          <span class="text-text-muted">·</span>
          <span>Last run: {{ lastRunLabel }}</span>
        </div>
      </div>

      <div class="flex items-center gap-1.5 shrink-0">
        <button
          type="button"
          @click="emit('run', schedule)"
          :disabled="running || !schedule.enabled"
          class="px-3 py-1.5 border border-border text-text-secondary font-mono text-xs uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
        >
          {{ running ? 'Running…' : 'Run now' }}
        </button>
        <button
          type="button"
          @click="emit('edit', schedule)"
          class="px-3 py-1.5 border border-border text-text-secondary font-mono text-xs uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
        >
          Edit
        </button>
        <button
          type="button"
          @click="emit('delete', schedule)"
          class="px-3 py-1.5 border border-border text-text-muted font-mono text-xs uppercase tracking-wider rounded hover:border-status-critical/50 hover:text-status-critical transition-colors cursor-pointer"
        >
          Delete
        </button>
      </div>
    </div>

    <!-- Prompt preview -->
    <div class="pt-3 border-t border-border">
      <div class="font-mono text-xs uppercase tracking-wider text-text-muted mb-1">Prompt</div>
      <p class="font-mono text-xs text-text-secondary whitespace-pre-wrap break-words leading-relaxed line-clamp-3">
        {{ schedule.prompt }}
      </p>
    </div>

    <!-- Last summary or error -->
    <div v-if="schedule.last_error" class="pt-3 border-t border-border">
      <div class="font-mono text-xs uppercase tracking-wider text-status-critical mb-1">Last error</div>
      <p class="font-mono text-xs text-text-secondary whitespace-pre-wrap break-words leading-relaxed">
        {{ schedule.last_error }}
      </p>
    </div>
    <div v-else-if="schedule.last_summary" class="pt-3 border-t border-border">
      <div class="font-mono text-xs uppercase tracking-wider text-text-muted mb-1">Last summary</div>
      <p class="font-mono text-xs text-text-secondary whitespace-pre-wrap break-words leading-relaxed line-clamp-3">
        {{ schedule.last_summary }}
      </p>
    </div>
  </div>
</template>
