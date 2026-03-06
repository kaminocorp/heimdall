<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAgentStore } from '@/stores/agent'
import { useConnectionsStore } from '@/stores/connections'
import { useLogsStore } from '@/stores/logs'
import { getDashboardStats, type DashboardStats } from '@/api/stats'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'

const agentStore = useAgentStore()
const connectionsStore = useConnectionsStore()
const logsStore = useLogsStore()

const stats = ref<DashboardStats | null>(null)
const statsError = ref('')
const fetchError = ref('')

onMounted(async () => {
  try { await agentStore.fetchConfig() } catch { fetchError.value = 'Failed to load agent config' }
  try { await connectionsStore.fetchConnections() } catch { fetchError.value = 'Failed to load connections' }
  try { await logsStore.fetchLogs() } catch { fetchError.value = 'Failed to load logs' }
  try {
    const { data } = await getDashboardStats()
    stats.value = data
  } catch {
    statsError.value = 'Failed to load stats'
  }
})

const agentStatus = computed(() => {
  if (!agentStore.config) return { label: '...', color: 'text-text-muted', dot: 'bg-text-muted' }
  switch (agentStore.config.mode) {
    case 'continuous': return { label: 'Active', color: 'text-accent', dot: 'bg-status-ok animate-pulse' }
    case 'scheduled': return { label: `Scheduled`, color: 'text-status-info', dot: 'bg-status-info' }
    case 'off': return { label: 'Inactive', color: 'text-text-muted', dot: 'bg-text-muted' }
    default: return { label: agentStore.config.mode, color: 'text-text-muted', dot: 'bg-text-muted' }
  }
})

const activeCount = computed(() =>
  connectionsStore.connections.filter(c => c.status === 'active').length
)
const inactiveCount = computed(() =>
  connectionsStore.connections.filter(c => c.status !== 'active').length
)

const recentEntries = computed(() => logsStore.entries.slice(0, 8))
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="pb-6 mb-8 border-b border-border">
      <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">Dashboard</h2>
      <p class="font-sans text-sm text-text-secondary mt-1">System overview and recent activity</p>
    </div>

    <!-- Error banner -->
    <div v-if="fetchError" class="mb-6 rounded border border-status-critical/30 bg-status-critical/10 px-4 py-2 text-sm font-mono text-status-critical flex items-center justify-between">
      <span>{{ fetchError }}</span>
      <button @click="fetchError = ''" class="text-xs opacity-60 hover:opacity-100">&times;</button>
    </div>

    <!-- Status cards row -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
      <!-- System Status -->
      <div class="border border-border rounded-lg bg-bg-surface p-5 glow-active animate-fade-in" :style="{ '--stagger-index': 0 }">
        <div class="font-mono text-[10px] font-medium uppercase tracking-widest text-text-muted mb-4">System Status</div>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Agent</span>
            <div class="flex items-center gap-2">
              <span class="w-1.5 h-1.5 rounded-full" :class="agentStatus.dot" />
              <span class="font-mono text-xs uppercase tracking-wider" :class="agentStatus.color">{{ agentStatus.label }}</span>
            </div>
          </div>
          <div class="flex items-center justify-between" v-if="agentStore.config">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Model</span>
            <span class="font-mono text-xs text-text-primary">{{ agentStore.config.model }}</span>
          </div>
          <div class="flex items-center justify-between" v-if="agentStore.config?.mode === 'scheduled' && agentStore.config.schedule">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Schedule</span>
            <span class="font-mono text-xs text-text-primary">{{ agentStore.config.schedule }}</span>
          </div>
        </div>
      </div>

      <!-- Connections -->
      <div class="border border-border rounded-lg bg-bg-surface p-5 animate-fade-in" :style="{ '--stagger-index': 1 }">
        <div class="font-mono text-[10px] font-medium uppercase tracking-widest text-text-muted mb-4">Connections</div>
        <div class="font-mono text-xs text-text-secondary mb-3">
          {{ activeCount }} active<span v-if="inactiveCount"> · {{ inactiveCount }} inactive</span>
        </div>
        <div v-if="connectionsStore.connections.length === 0" class="font-mono text-xs text-text-muted">
          No connections configured.
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="conn in connectionsStore.connections"
            :key="conn.id"
            class="flex items-center justify-between"
          >
            <span class="font-mono text-xs text-text-primary truncate mr-2">{{ conn.name }}</span>
            <StatusBadge :status="conn.status" />
          </div>
        </div>
      </div>

      <!-- Log Ingestion -->
      <div class="border border-border rounded-lg bg-bg-surface p-5 animate-fade-in" :style="{ '--stagger-index': 2 }">
        <div class="font-mono text-[10px] font-medium uppercase tracking-widest text-text-muted mb-4">Log Ingestion</div>
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
    <div class="border border-border rounded-lg bg-bg-surface p-5 animate-fade-in" :style="{ '--stagger-index': 3 }">
      <div class="font-mono text-[10px] font-medium uppercase tracking-widest text-text-muted mb-4">Recent Activity</div>
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
          <span class="font-mono text-[10px] text-text-muted shrink-0 tabular-nums">
            {{ new Date(entry.timestamp).toLocaleTimeString() }}
          </span>
          <span
            class="font-mono text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded shrink-0"
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
