<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useLogsStore } from '@/stores/logs'
import { useConnectionsStore } from '@/stores/connections'
import LogFeed from '@/components/log/LogFeed.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const logsStore = useLogsStore()
const connectionsStore = useConnectionsStore()

const activeFilters = ref<{ severity?: string; connection_id?: string }>({})

onMounted(() => {
  logsStore.fetchLogs()
  connectionsStore.fetchConnections()
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
      <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">Agent Log</h2>
      <p class="font-sans text-sm text-text-secondary mt-1">Unified chronological feed of all system activity</p>
    </div>

    <div v-if="logsStore.error" class="mb-4 rounded border border-status-critical/30 bg-status-critical/10 px-4 py-2 text-sm font-mono text-status-critical">
      {{ logsStore.error }}
    </div>

    <LoadingSpinner v-if="logsStore.loading && logsStore.entries.length === 0" />
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
