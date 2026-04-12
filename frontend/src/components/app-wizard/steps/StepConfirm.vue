<script setup lang="ts">
// Step 3 — terminal success screen. Rendered after either "Skip for now"
// or a completed connector sub-flow. This step doesn't do any work itself;
// the "Go to dashboard" button dispatches back up to AppWizard which
// commits the selection and routes away.
defineProps<{
  appName: string
  connectorAdded: boolean
}>()
</script>

<template>
  <div class="text-center py-4">
    <div class="mx-auto w-12 h-12 rounded-full bg-accent-subtle flex items-center justify-center mb-4">
      <svg class="w-6 h-6 text-accent" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M20 6L9 17l-5-5" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </div>

    <h4 class="font-mono text-sm font-medium text-text-primary mb-1">
      You're all set.
    </h4>
    <p class="font-mono text-xs text-text-secondary">
      <span class="text-text-primary font-medium">{{ appName }}</span> is ready to go.
    </p>
    <p v-if="connectorAdded" class="mt-1 font-mono text-xs text-text-muted">
      Your connector is live and waiting for data.
    </p>
    <p v-else class="mt-1 font-mono text-xs text-text-muted">
      Head over to Connections whenever you're ready to wire up a data source.
    </p>
  </div>
</template>
