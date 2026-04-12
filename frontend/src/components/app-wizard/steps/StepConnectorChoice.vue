<script setup lang="ts">
// Step 2 — binary fork: add a connector now, or skip and land on the
// confirmation screen. We don't render the full PlatformGrid here because
// the inner connector flow is handled by the embedded ConnectionWizard
// (launched by AppWizard when the user picks "Add a connector"). This keeps
// this step dumb — it just emits the user's intent.
const emit = defineEmits<{
  choose: [choice: 'add' | 'skip']
}>()
</script>

<template>
  <div>
    <p class="font-mono text-xs text-text-secondary mb-6">
      Your application is ready. Would you like to add a connector now? Connectors wire log sources,
      databases, and code repositories into Heimdall so the agent has something to watch.
    </p>

    <div class="grid gap-3">
      <button
        type="button"
        @click="emit('choose', 'add')"
        class="group flex items-start gap-4 px-5 py-4 rounded border border-border hover:border-accent/50 hover:bg-accent-subtle/20 text-left transition-colors cursor-pointer"
      >
        <div class="flex-shrink-0 w-8 h-8 rounded flex items-center justify-center bg-accent-subtle text-accent font-mono text-base">
          +
        </div>
        <div class="min-w-0">
          <div class="font-mono text-sm font-medium text-text-primary group-hover:text-accent-bright transition-colors">
            Add a connector
          </div>
          <div class="mt-0.5 font-mono text-xs text-text-muted">
            Walk through the connector wizard to wire up a log source, database, or codebase.
          </div>
        </div>
      </button>

      <button
        type="button"
        @click="emit('choose', 'skip')"
        class="group flex items-start gap-4 px-5 py-4 rounded border border-border hover:border-border-hover text-left transition-colors cursor-pointer"
      >
        <div class="flex-shrink-0 w-8 h-8 rounded flex items-center justify-center bg-bg-surface text-text-muted font-mono text-base">
          &rsaquo;
        </div>
        <div class="min-w-0">
          <div class="font-mono text-sm font-medium text-text-primary">
            Skip for now
          </div>
          <div class="mt-0.5 font-mono text-xs text-text-muted">
            Create the application without a connector. You can add one from the Connections page later.
          </div>
        </div>
      </button>
    </div>
  </div>
</template>
