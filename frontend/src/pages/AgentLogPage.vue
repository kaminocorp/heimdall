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

function handleFilter(filters: { severity?: string; connection_id?: string }) {
  activeFilters.value = filters
  logsStore.resetPagination()
  logsStore.fetchLogs(filters)
}
</script>

<template>
  <div>
    <h2 class="text-2xl font-semibold mb-4">Agent Log</h2>

    <div v-if="logsStore.error" class="mb-4 p-3 bg-red-50 text-red-700 rounded text-sm">
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
