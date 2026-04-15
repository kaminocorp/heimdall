<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { useLogsStore } from '@/stores/logs'
import { getAppStats, getAppAgentConfig, getMonitoringStatus, listConnectionsByApp } from '@/api/applications'
import type { AppAgentConfig, MonitoringStatus } from '@/types/organization'
import type { Connection } from '@/types/connection'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'

const appStore = useAppStore()
const logsStore = useLogsStore()

const connections = ref<Connection[]>([])
const agentConfig = ref<AppAgentConfig | null>(null)
const monitoring = ref<MonitoringStatus | null>(null)
const stats = ref<{ log_count_24h: number; connection_count: number; active_connections: number } | null>(null)
const statsError = ref('')
const fetchError = ref('')

// Generation counter prevents stale responses from overwriting fresh data
// when the user switches apps rapidly.
let loadGeneration = 0

async function loadData() {
  const appId = appStore.currentAppId
  if (!appId) return

  const gen = ++loadGeneration
  fetchError.value = ''
  statsError.value = ''

  const [agentRes, connRes, logsRes, statsRes, monitorRes] = await Promise.allSettled([
    getAppAgentConfig(appId),
    listConnectionsByApp(appId),
    logsStore.fetchLogs(),
    getAppStats(appId),
    getMonitoringStatus(appId),
  ])

  // Discard results if a newer loadData call has started since we fired.
  if (gen !== loadGeneration) return

  const errors: string[] = []
  if (agentRes.status === 'fulfilled') { agentConfig.value = agentRes.value } else { errors.push('agent config') }
  if (connRes.status === 'fulfilled') { connections.value = connRes.value } else { errors.push('connections') }
  if (logsRes.status === 'rejected') { errors.push('logs') }
  if (statsRes.status === 'fulfilled') { stats.value = statsRes.value } else { statsError.value = 'Failed to load stats' }
  if (monitorRes.status === 'fulfilled') { monitoring.value = monitorRes.value }

  if (errors.length) {
    fetchError.value = `Failed to load: ${errors.join(', ')}`
  }
}

onMounted(loadData)
watch(() => appStore.currentAppId, loadData)

const agentStatus = computed(() => {
  if (!agentConfig.value) return { label: '...', color: 'text-text-muted', dot: 'bg-text-muted' }
  switch (agentConfig.value.mode) {
    case 'continuous': return { label: 'Continuous', color: 'text-accent', dot: 'bg-status-ok glow-pulse' }
    case 'periodic': return { label: 'Periodic', color: 'text-status-info', dot: 'bg-status-info' }
    case 'off': return { label: 'Inactive', color: 'text-text-muted', dot: 'bg-text-muted' }
    default: return { label: agentConfig.value.mode, color: 'text-text-muted', dot: 'bg-text-muted' }
  }
})

const activeCount = computed(() => connections.value.filter(c => c.status === 'active').length)
const inactiveCount = computed(() => connections.value.filter(c => c.status !== 'active').length)
const recentEntries = computed(() => logsStore.entries?.slice(0, 8) ?? [])

function formatInterval(secs: number): string {
  if (secs < 60) return `${secs}s`
  if (secs < 3600) return `${Math.round(secs / 60)}m`
  return `${Math.round(secs / 3600)}h`
}
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="pb-6 mb-8 border-b border-border">
      <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">Dashboard</h2>
      <p class="font-sans text-sm text-text-secondary mt-1">
        {{ appStore.currentApp?.name ?? 'System overview' }} — monitoring &amp; recent activity
      </p>
    </div>

    <!-- Error banner -->
    <div v-if="fetchError" class="mb-6 rounded border border-status-critical/30 bg-status-critical/10 px-4 py-2 text-sm font-mono text-status-critical flex items-center justify-between">
      <span>{{ fetchError }}</span>
      <button @click="fetchError = ''" class="text-xs opacity-60 hover:opacity-100">&times;</button>
    </div>

    <!-- Status cards row -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 mb-10">
      <!-- Monitoring Status -->
      <div class="border border-border rounded-lg bg-bg-surface p-6 animate-fade-in chrome-brackets" :style="{ '--stagger-index': 0 }">
        <div class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted mb-4">Monitoring</div>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Mode</span>
            <div class="flex items-center gap-2">
              <span class="w-1.5 h-1.5 rounded-full" :class="agentStatus.dot" />
              <span class="font-mono text-xs uppercase tracking-wider" :class="agentStatus.color">{{ agentStatus.label }}</span>
            </div>
          </div>
          <div v-if="agentConfig?.mode === 'periodic'" class="flex items-center justify-between">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Interval</span>
            <span class="font-mono text-xs text-text-primary">{{ formatInterval(agentConfig.schedule_interval_secs) }}</span>
          </div>
          <div v-if="monitoring?.last_monitored_at" class="flex items-center justify-between">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Last Check</span>
            <span class="font-mono text-xs text-text-primary tabular-nums">{{ new Date(monitoring.last_monitored_at).toLocaleTimeString() }}</span>
          </div>
        </div>
      </div>

      <!-- System Status -->
      <div class="border border-border rounded-lg bg-bg-surface p-6 animate-fade-in chrome-brackets" :style="{ '--stagger-index': 1 }">
        <div class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted mb-4">Agent</div>
        <div class="space-y-3">
          <div v-if="agentConfig" class="flex items-center justify-between">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Model</span>
            <span class="font-mono text-xs text-text-primary">{{ agentConfig.model }}</span>
          </div>
        </div>
      </div>

      <!-- Connections -->
      <div class="border border-border rounded-lg bg-bg-surface p-6 animate-fade-in chrome-brackets" :style="{ '--stagger-index': 2 }">
        <div class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted mb-4">Connections</div>
        <div class="font-mono text-xs text-text-secondary mb-3">
          {{ activeCount }} active<span v-if="inactiveCount"> · {{ inactiveCount }} inactive</span>
        </div>
        <div v-if="connections.length === 0" class="font-mono text-xs text-text-muted">
          No connections configured.
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="conn in connections"
            :key="conn.id"
            class="flex items-center justify-between"
          >
            <span class="font-mono text-xs text-text-primary truncate mr-2">{{ conn.name }}</span>
            <StatusBadge :status="conn.status" />
          </div>
        </div>
      </div>

      <!-- Log Ingestion -->
      <div class="border border-border rounded-lg bg-bg-surface p-6 animate-fade-in chrome-brackets" :style="{ '--stagger-index': 3 }">
        <div class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted mb-4">Log Ingestion</div>
        <div v-if="statsError" class="font-mono text-xs text-status-critical">{{ statsError }}</div>
        <div v-else-if="stats" class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Last 24h</span>
            <span class="font-mono text-sm text-text-primary tabular-nums">{{ stats.log_count_24h.toLocaleString() }} entries</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Rate</span>
            <span class="font-mono text-sm text-text-primary tabular-nums">~{{ Math.round(stats.log_count_24h / 24) }}/hr</span>
          </div>
        </div>
        <div v-else class="font-mono text-xs text-text-muted">Loading...</div>
      </div>
    </div>

    <!-- Recent Activity -->
    <div class="border border-border rounded-lg bg-bg-surface p-6 animate-fade-in" :style="{ '--stagger-index': 4 }">
      <div class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted mb-4">Recent Activity</div>
      <div v-if="logsStore.loading" class="font-mono text-xs text-text-muted">Loading...</div>
      <div v-else-if="recentEntries.length === 0" class="font-mono text-xs text-text-muted">
        No recent activity.
      </div>
      <div v-else class="space-y-1">
        <div
          v-for="(entry, i) in recentEntries"
          :key="entry.id"
          class="flex items-baseline gap-3 py-1.5 border-b border-border last:border-b-0 animate-fade-in"
          :style="{ '--stagger-index': i }"
        >
          <span class="font-mono text-xs text-text-muted shrink-0 tabular-nums">
            {{ new Date(entry.timestamp).toLocaleTimeString() }}
          </span>
          <span
            class="font-mono text-xs font-bold uppercase tracking-wider px-1.5 py-0.5 rounded shrink-0"
            :class="{
              'text-accent': entry.source === 'agent',
              'text-status-critical': entry.severity === 'critical',
              'text-status-warn': entry.severity === 'warning',
              'text-status-info': entry.severity === 'info' || (!entry.severity && entry.source !== 'agent'),
            }"
          >{{ entry.source === 'agent' ? entry.source_type : entry.severity || 'info' }}</span>
          <span class="font-mono text-xs text-text-secondary truncate">{{ entry.summary }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
