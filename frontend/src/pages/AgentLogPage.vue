<script setup lang="ts">
import { onMounted } from 'vue'
import { useLogsStore } from '@/stores/logs'
import LogFeed from '@/components/log/LogFeed.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const store = useLogsStore()

onMounted(() => {
  store.fetchLogs()
})

function handleFilter(filters: Record<string, string>) {
  store.fetchLogs(filters)
}
</script>

<template>
  <div>
    <h2 class="text-2xl font-semibold mb-4">Agent Log</h2>
    <LoadingSpinner v-if="store.loading" />
    <LogFeed v-else :entries="store.entries" @filter="handleFilter" />
  </div>
</template>
