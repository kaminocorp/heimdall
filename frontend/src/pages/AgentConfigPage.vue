<script setup lang="ts">
import { onMounted } from 'vue'
import { useAgentStore } from '@/stores/agent'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const store = useAgentStore()

onMounted(() => {
  store.fetchConfig()
})
</script>

<template>
  <div>
    <h2 class="text-2xl font-semibold mb-4">Agent Configuration</h2>
    <LoadingSpinner v-if="store.loading" />
    <div v-else-if="store.config" class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-500">Model</label>
        <p>{{ store.config.model }}</p>
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-500">Mode</label>
        <p>{{ store.config.mode }}</p>
      </div>
    </div>
  </div>
</template>
