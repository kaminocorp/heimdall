<script setup lang="ts">
import { onMounted } from 'vue'
import { useConnectionsStore } from '@/stores/connections'
import ConnectionList from '@/components/connections/ConnectionList.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const store = useConnectionsStore()

onMounted(() => {
  store.fetchConnections()
})
</script>

<template>
  <div>
    <h2 class="text-2xl font-semibold mb-4">Connections</h2>
    <LoadingSpinner v-if="store.loading" />
    <ConnectionList v-else :connections="store.connections" />
  </div>
</template>
