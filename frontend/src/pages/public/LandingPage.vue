<script setup lang="ts">
import PublicNav from '@/components/public/PublicNav.vue'
import PublicFooter from '@/components/public/PublicFooter.vue'
</script>

<template>
  <div class="min-h-screen flex flex-col bg-bg-primary">
    <PublicNav />

    <!-- Hero: centered in remaining space -->
    <main class="flex-1 flex items-center justify-center px-6">
      <div class="max-w-3xl text-center">
        <h1 class="font-mono text-4xl sm:text-5xl lg:text-6xl font-bold leading-tight tracking-tight">
          <span class="block">Watches everything.</span>
          <span class="block text-accent">Investigates automatically.</span>
          <span class="block">Reports what matters.</span>
        </h1>

        <p class="mt-6 text-base sm:text-lg text-text-secondary max-w-2xl mx-auto leading-relaxed">
          Heimdall monitors your production systems 24/7, autonomously investigates anomalies,
          and delivers actionable incident reports — so you know exactly what went wrong and how to fix it.
        </p>

        <div class="mt-8 flex items-center justify-center gap-4">
          <RouterLink
            to="/login"
            class="px-6 py-2.5 text-sm font-mono font-medium bg-accent hover:bg-accent-hover text-bg-primary rounded transition-colors"
          >
            Start Monitoring
          </RouterLink>
          <RouterLink
            to="/features"
            class="px-6 py-2.5 text-sm font-mono font-medium border border-accent-border text-text-secondary hover:text-text-primary hover:border-accent/40 rounded transition-colors"
          >
            See How It Works
          </RouterLink>
        </div>

        <p class="mt-4 text-xs font-mono text-text-muted">
          Set up in under 5 minutes. No credit card required.
        </p>
      </div>
    </main>

    <PublicFooter />
  </div>
</template>
