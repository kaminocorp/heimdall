<script setup lang="ts">
</script>

<template>
  <div class="flex flex-col items-center justify-center min-h-[60vh]">
    <h1 class="font-mono text-6xl font-bold text-text-muted mb-2">404</h1>
    <p class="font-mono text-sm uppercase tracking-wider text-text-muted mb-8">Target not found</p>
    <RouterLink
      to="/"
      class="px-5 py-2.5 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors"
    >
      Return to Dashboard
    </RouterLink>
  </div>
</template>
