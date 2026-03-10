<script setup lang="ts">
import PublicNav from '@/components/public/PublicNav.vue'
import PublicFooter from '@/components/public/PublicFooter.vue'
import HeroMesh from '@/components/public/HeroMesh.vue'
</script>

<template>
  <div class="min-h-screen flex flex-col bg-bg-primary">
    <PublicNav />

    <!-- Hero: mesh background with text overlay -->
    <main class="flex-1 flex items-center justify-center relative overflow-hidden">
      <!-- Animated dot mesh -->
      <HeroMesh />

      <!-- Gradient overlay to fade mesh edges and improve text legibility -->
      <div class="absolute inset-0 bg-gradient-to-b from-bg-primary via-transparent to-bg-primary pointer-events-none" />
      <div class="absolute inset-0 bg-gradient-to-r from-bg-primary/60 via-transparent to-bg-primary/60 pointer-events-none" />

      <!-- Text content -->
      <div class="relative z-10 w-full max-w-5xl text-center px-6 sm:px-10 lg:px-16">
        <h1 class="font-mono text-5xl sm:text-6xl lg:text-8xl font-bold leading-[1.05] tracking-tight uppercase">
          <span class="block">The all-seeing</span>
          <span class="block text-accent">eye.</span>
        </h1>

        <p class="mt-8 text-lg sm:text-xl text-text-secondary max-w-2xl mx-auto leading-relaxed">
          Autonomous AI monitoring for production systems.
          Watches 24/7. Investigates anomalies. Reports what matters.
        </p>

        <div class="mt-10 flex items-center justify-center gap-4">
          <RouterLink
            to="/login"
            class="px-8 py-3 text-sm font-mono font-medium bg-accent hover:bg-accent-hover text-bg-primary rounded transition-colors"
          >
            Start Monitoring
          </RouterLink>
          <RouterLink
            to="/features"
            class="px-8 py-3 text-sm font-mono font-medium border border-accent-border text-text-secondary hover:text-text-primary hover:border-accent/40 rounded transition-colors"
          >
            See How It Works
          </RouterLink>
        </div>

        <p class="mt-5 text-xs font-mono text-text-muted">
          Set up in under 5 minutes. No credit card required.
        </p>
      </div>
    </main>

    <PublicFooter />
  </div>
</template>
