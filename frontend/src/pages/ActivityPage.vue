<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useLogsStore } from '@/stores/logs'
import { useConnectionsStore } from '@/stores/connections'
import { useAppStore } from '@/stores/app'
import LogFeed from '@/components/log/LogFeed.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'

const logsStore = useLogsStore()
const connectionsStore = useConnectionsStore()
const appStore = useAppStore()

const activeFilters = ref<{ severity?: string; connection_id?: string }>({})

onMounted(() => {
  logsStore.fetchLogs()
  if (appStore.currentAppId) {
    connectionsStore.fetchConnectionsByApp(appStore.currentAppId)
  }
})

watch(() => appStore.currentAppId, (appId) => {
  logsStore.resetPagination()
  logsStore.fetchLogs(activeFilters.value)
  if (appId) {
    connectionsStore.fetchConnectionsByApp(appId)
  }
})

function handleFilter(filters: { severity?: string; connection_id?: string; source?: string }) {
  if (filters.source) {
    logsStore.setSource(filters.source as 'raw' | 'agent' | 'all')
  } else {
    logsStore.setSource('all')
  }
  activeFilters.value = { severity: filters.severity, connection_id: filters.connection_id }
  logsStore.resetPagination()
  logsStore.fetchLogs(activeFilters.value)
}
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="pb-6 mb-6 border-b border-border">
      <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">Activity</h2>
      <p class="font-sans text-sm text-text-secondary mt-1">Unified chronological feed of all system activity</p>
    </div>

    <div v-if="logsStore.error" class="mb-4 rounded border border-status-critical/30 bg-status-critical/10 px-4 py-2 text-sm font-mono text-status-critical">
      {{ logsStore.error }}
    </div>

    <!-- Skeleton loader -->
    <div v-if="logsStore.loading && logsStore.entries.length === 0" class="border border-border rounded-lg bg-bg-surface p-6 space-y-3">
      <div v-for="n in 8" :key="n" class="flex items-center gap-3 py-1.5">
        <SkeletonBlock width="4rem" height="0.75rem" />
        <SkeletonBlock width="3rem" height="0.75rem" />
        <SkeletonBlock width="100%" height="0.75rem" />
      </div>
    </div>
    <LogFeed
      v-else
      :entries="logsStore.entries"
      :connections="connectionsStore.connections"
      :total="logsStore.total"
      :limit="logsStore.limit"
      :offset="logsStore.offset"
      @filter="handleFilter"
      @next="logsStore.nextPage(activeFilters)"
      @prev="logsStore.prevPage(activeFilters)"
    />
  </div>
</template>
