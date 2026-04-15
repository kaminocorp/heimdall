<script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import type { Connection } from '@/types/connection'
import ConnectorLogo from '@/components/icons/ConnectorLogo.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'

const props = defineProps<{
  connection: Connection
}>()

const emit = defineEmits<{
  close: []
  edit: [connection: Connection]
  test: [id: string]
  delete: [id: string]
  'manage-repos': [id: string]
}>()

/* ── Type labels ── */

const typeLabels: Record<string, string> = {
  postgres: 'PostgreSQL',
  supabase: 'Supabase',
  webhook_logs: 'Webhook Logs',
  syslog: 'Syslog',
  github: 'GitHub',
  otlp: 'OpenTelemetry',
  datadog: 'Datadog',
  mysql: 'MySQL',
}

const displayType = computed(() => typeLabels[props.connection.type] ?? props.connection.type)

const directionLabel = computed(() =>
  props.connection.direction === 'two_way' ? 'Two-way' : 'One-way'
)

/* ── Config details (type-specific) ── */

const configDetails = computed(() => {
  const c = props.connection
  const cfg = c.config as Record<string, unknown> | undefined
  if (!cfg) return []

  const details: { label: string; value: string; masked?: boolean }[] = []

  switch (c.type) {
    case 'postgres':
      if (cfg.host) details.push({ label: 'Host', value: String(cfg.host) })
      if (cfg.port) details.push({ label: 'Port', value: String(cfg.port) })
      if (cfg.database) details.push({ label: 'Database', value: String(cfg.database) })
      if (cfg.user) details.push({ label: 'User', value: String(cfg.user) })
      if (cfg.password) details.push({ label: 'Password', value: String(cfg.password), masked: true })
      if (cfg.ssl_mode) details.push({ label: 'SSL Mode', value: String(cfg.ssl_mode) })
      break

    case 'supabase':
      if (cfg.project_ref) details.push({ label: 'Project Ref', value: String(cfg.project_ref) })
      if (cfg.access_token) details.push({ label: 'Access Token', value: String(cfg.access_token), masked: true })
      if (cfg.poll_interval_secs) details.push({ label: 'Poll Interval', value: `${cfg.poll_interval_secs}s` })
      if (Array.isArray(cfg.poll_tables) && cfg.poll_tables.length > 0) {
        details.push({ label: 'Polling Tables', value: `${cfg.poll_tables.length} table${cfg.poll_tables.length === 1 ? '' : 's'}` })
      }
      break

    case 'webhook_logs':
      if (cfg.webhook_token) details.push({ label: 'Bearer Token', value: String(cfg.webhook_token), masked: true })
      break

    case 'syslog':
      if (cfg.port) details.push({ label: 'Port', value: String(cfg.port) })
      if (cfg.protocol) details.push({ label: 'Protocol', value: String(cfg.protocol).toUpperCase() })
      break

    case 'github':
      // GitHub connections show repos via manage-repos action
      break

    case 'otlp':
      // OTLP shows endpoint info
      break
  }

  return details
})

/* ── Metadata ── */

const createdDate = computed(() => {
  const d = new Date(props.connection.created_at)
  return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
})

const updatedDate = computed(() => {
  const d = new Date(props.connection.updated_at)
  return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
})

/* ── Masked field reveal ── */

const revealedFields = ref<Set<string>>(new Set())

function toggleReveal(label: string) {
  if (revealedFields.value.has(label)) {
    revealedFields.value.delete(label)
  } else {
    revealedFields.value.add(label)
  }
}

function maskedValue(value: string, label: string): string {
  if (revealedFields.value.has(label)) return value
  if (value.length <= 8) return '*'.repeat(value.length)
  return value.slice(0, 4) + '*'.repeat(Math.min(value.length - 4, 20))
}

/* ── Delete confirmation ── */

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
  emit('close')
}

/* ── Keyboard ── */

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}

onMounted(() => {
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeydown)
  if (deleteTimer) { clearTimeout(deleteTimer); deleteTimer = null }
})
</script>

<template>
  <!-- Backdrop -->
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" @click.self="emit('close')">
    <!-- Modal -->
    <div class="w-full max-w-lg mx-4 border border-border rounded-lg bg-bg-surface shadow-2xl">

      <!-- Header -->
      <div class="flex items-start justify-between px-6 py-5 border-b border-border">
        <div class="flex items-center gap-4">
          <div class="flex items-center justify-center w-12 h-12 rounded-lg bg-bg-elevated border border-border">
            <ConnectorLogo
              :type="connection.type"
              :size="28"
              :class="connection.status === 'active' ? 'text-accent' : 'text-text-muted'"
            />
          </div>
          <div>
            <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">
              {{ connection.name }}
            </h3>
            <p class="font-mono text-xs text-text-muted mt-0.5">
              {{ displayType }} &middot; {{ directionLabel }}
            </p>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <StatusBadge :status="connection.status" />
          <button
            @click="emit('close')"
            class="font-mono text-lg text-text-muted hover:text-text-primary transition-colors cursor-pointer leading-none"
          >&times;</button>
        </div>
      </div>

      <!-- Config details -->
      <div v-if="configDetails.length > 0" class="px-6 py-4 border-b border-border space-y-2">
        <div v-for="detail in configDetails" :key="detail.label" class="flex items-baseline justify-between gap-4">
          <span class="font-mono text-xs uppercase tracking-wider text-text-muted shrink-0">{{ detail.label }}</span>
          <div class="flex items-center gap-2 min-w-0">
            <span class="font-mono text-xs text-text-secondary truncate text-right">
              {{ detail.masked ? maskedValue(detail.value, detail.label) : detail.value }}
            </span>
            <button
              v-if="detail.masked"
              @click="toggleReveal(detail.label)"
              class="font-mono text-[0.6rem] uppercase tracking-wider text-text-muted hover:text-accent transition-colors cursor-pointer shrink-0"
            >
              {{ revealedFields.has(detail.label) ? 'Hide' : 'Show' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Metadata -->
      <div class="px-6 py-4 border-b border-border space-y-2">
        <div class="flex items-baseline justify-between">
          <span class="font-mono text-xs uppercase tracking-wider text-text-muted">Created</span>
          <span class="font-mono text-xs text-text-secondary">{{ createdDate }}</span>
        </div>
        <div class="flex items-baseline justify-between">
          <span class="font-mono text-xs uppercase tracking-wider text-text-muted">Updated</span>
          <span class="font-mono text-xs text-text-secondary">{{ updatedDate }}</span>
        </div>
        <div class="flex items-baseline justify-between">
          <span class="font-mono text-xs uppercase tracking-wider text-text-muted">ID</span>
          <span class="font-mono text-[0.6rem] text-text-muted select-all">{{ connection.id }}</span>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex items-center justify-between px-6 py-4">
        <div class="flex items-center gap-2">
          <button
            v-if="connection.type === 'github'"
            @click="emit('manage-repos', connection.id); emit('close')"
            class="px-3 py-1.5 font-mono text-xs uppercase tracking-wider text-text-secondary border border-border rounded hover:border-border-hover hover:text-text-primary transition-colors cursor-pointer"
          >
            Repos
          </button>
          <button
            @click="emit('test', connection.id); emit('close')"
            class="px-3 py-1.5 font-mono text-xs uppercase tracking-wider text-text-secondary border border-border rounded hover:border-border-hover hover:text-text-primary transition-colors cursor-pointer"
          >
            Ping
          </button>
          <button
            @click="emit('edit', connection); emit('close')"
            class="px-3 py-1.5 font-mono text-xs uppercase tracking-wider text-text-secondary border border-border rounded hover:border-border-hover hover:text-text-primary transition-colors cursor-pointer"
          >
            Edit
          </button>
        </div>
        <button
          @click="handleDelete"
          class="px-3 py-1.5 font-mono text-xs uppercase tracking-wider rounded transition-colors cursor-pointer"
          :class="confirmingDelete
            ? 'text-status-critical border border-status-critical/40 bg-status-critical/10 font-medium'
            : 'text-text-muted border border-border hover:text-status-critical hover:border-status-critical/30'"
        >
          {{ confirmingDelete ? 'Confirm Delete?' : 'Delete' }}
        </button>
      </div>
    </div>
  </div>
</template>
