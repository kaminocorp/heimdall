<script setup lang="ts">
import type { NotificationLogEntry } from '@/types/notification'

defineProps<{
  entries: NotificationLogEntry[]
}>()

function severityIndicator(severity: string) {
  switch (severity) {
    case 'critical': return 'bg-status-critical'
    case 'error': return 'bg-status-error'
    case 'warning': return 'bg-status-warning'
    default: return 'bg-accent'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'sent': return 'text-accent'
    case 'failed': return 'text-status-critical'
    default: return 'text-text-muted'
  }
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}
</script>

<template>
  <section>
    <h3 class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted mb-3">Recent Notifications</h3>

    <div v-if="entries.length === 0" class="border border-border rounded-lg bg-bg-surface p-8 text-center">
      <p class="font-mono text-sm text-text-muted">No notifications sent yet</p>
    </div>

    <div v-else class="border border-border rounded-lg bg-bg-surface overflow-hidden">
      <div
        v-for="entry in entries"
        :key="entry.id"
        class="flex items-center gap-3 px-4 py-3 border-b border-border last:border-b-0"
      >
        <span class="w-2 h-2 rounded-full flex-shrink-0" :class="severityIndicator(entry.severity)" />
        <span class="font-mono text-xs text-text-primary flex-1 min-w-0 truncate">{{ entry.summary }}</span>
        <span class="font-mono text-xs uppercase tracking-wider text-text-muted flex-shrink-0">{{ entry.channel_name }}</span>
        <span class="font-mono text-xs uppercase tracking-wider flex-shrink-0" :class="statusLabel(entry.status)">{{ entry.status }}</span>
        <span class="font-mono text-xs text-text-muted flex-shrink-0 w-16 text-right">{{ timeAgo(entry.created_at) }}</span>
      </div>
    </div>
  </section>
</template>
