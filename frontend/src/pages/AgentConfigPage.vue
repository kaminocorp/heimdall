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
    <!-- Page header -->
    <div class="pb-6 mb-8 border-b border-border">
      <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">Agent Configuration</h2>
      <p class="font-sans text-sm text-text-secondary mt-1">Model settings and agent behavior</p>
    </div>

    <LoadingSpinner v-if="store.loading" />
    <div v-else-if="store.config" class="border border-border rounded-lg bg-bg-surface p-5">
      <div class="space-y-4">
        <div class="flex items-baseline justify-between py-2 border-b border-border last:border-b-0">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Model</span>
          <span class="font-mono text-sm text-text-primary">{{ store.config.model }}</span>
        </div>
        <div class="flex items-baseline justify-between py-2 border-b border-border last:border-b-0">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Mode</span>
          <span class="font-mono text-sm text-text-primary">{{ store.config.mode }}</span>
        </div>
        <div class="flex items-baseline justify-between py-2 border-b border-border last:border-b-0">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Schedule</span>
          <span class="font-mono text-sm text-text-primary">{{ store.config.schedule ?? '—' }}</span>
        </div>
        <div class="flex items-baseline justify-between py-2">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">System Prompt</span>
          <span class="font-mono text-sm text-text-primary">{{ store.config.system_prompt_override ? 'Custom' : 'Default' }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
