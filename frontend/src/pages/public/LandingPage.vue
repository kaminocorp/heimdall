<script setup lang="ts">
import HeroMesh from '@/components/public/HeroMesh.vue'
</script>

<template>
  <!-- Hero: mesh background with text overlay -->
  <main class="flex-1 flex items-center justify-center relative overflow-hidden">
      <!-- Animated dot mesh -->
      <HeroMesh />

      <!-- Gradient overlays: clear top for text, fade mesh at edges -->
      <div class="absolute inset-0 pointer-events-none" style="background: linear-gradient(to bottom, var(--bg-primary) 20%, transparent 65%, var(--bg-primary) 95%)" />
      <div class="absolute inset-0 bg-gradient-to-r from-bg-primary/60 via-transparent to-bg-primary/60 pointer-events-none" />

      <!-- Text content -->
      <div class="relative z-10 w-full max-w-5xl text-center px-6 sm:px-10 lg:px-16">
        <h1 class="font-mono text-5xl sm:text-6xl lg:text-8xl font-bold leading-[1.05] tracking-tight uppercase">
          <span class="block">Autonomous system</span>
          <span class="block text-accent">surveillance.</span>
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
</template>
