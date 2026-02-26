<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useAgentStore } from '@/stores/agent'
import { useConnectionsStore } from '@/stores/connections'
import { useLogsStore } from '@/stores/logs'
import StatusBadge from '@/components/common/StatusBadge.vue'

const agentStore = useAgentStore()
const connectionsStore = useConnectionsStore()
const logsStore = useLogsStore()

onMounted(() => {
  agentStore.fetchConfig()
  connectionsStore.fetchConnections()
  logsStore.fetchLogs()
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

    <!-- Status cards row -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
      <!-- System Status -->
      <div class="border border-border rounded-lg bg-bg-surface p-5 glow-active animate-fade-in" :style="{ '--stagger-index': 0 }">
        <div class="font-mono text-[10px] font-medium uppercase tracking-widest text-text-muted mb-4">System Status</div>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Agent</span>
            <div class="flex items-center gap-2">
              <span class="w-1.5 h-1.5 rounded-full bg-status-ok animate-pulse" />
              <span class="font-mono text-xs text-accent uppercase tracking-wider">Active</span>
            </div>
          </div>
          <div class="flex items-center justify-between" v-if="agentStore.config">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Model</span>
            <span class="font-mono text-xs text-text-primary">{{ agentStore.config.model }}</span>
          </div>
          <div class="flex items-center justify-between" v-if="agentStore.config">
            <span class="font-mono text-xs uppercase tracking-wider text-text-secondary">Mode</span>
            <span class="font-mono text-xs text-text-primary">{{ agentStore.config.mode }}</span>
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
    </div>

    <!-- Recent Activity -->
    <div class="border border-border rounded-lg bg-bg-surface p-5 animate-fade-in" :style="{ '--stagger-index': 2 }">
      <div class="font-mono text-[10px] font-medium uppercase tracking-widest text-text-muted mb-4">Recent Activity</div>
      <div v-if="logsStore.loading" class="font-mono text-xs text-text-muted">Loading…</div>
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
