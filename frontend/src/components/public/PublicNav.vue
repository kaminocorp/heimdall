<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()
const mobileOpen = ref(false)
</script>

<template>
  <nav class="fixed top-0 left-0 right-0 z-50 border-b border-border backdrop-blur-sm bg-bg-primary/80">
    <div class="w-full px-6 lg:px-10 h-[76px] flex items-center justify-between">
      <!-- Brand -->
      <RouterLink to="/" class="font-mono font-semibold text-lg tracking-[0.3em] text-text-primary uppercase">
        H E I M D A L L
      </RouterLink>

      <!-- Desktop: center nav (absolute to guarantee true centering) -->
      <div class="hidden md:flex items-center gap-8 uppercase text-[12.5px] font-mono tracking-widest absolute left-1/2 -translate-x-1/2">
        <RouterLink
          to="/features"
          class="transition-colors"
          :class="route.name === 'features' ? 'text-accent' : 'text-text-secondary hover:text-text-primary'"
        >
          Platform
        </RouterLink>
        <RouterLink
          to="/pricing"
          class="transition-colors"
          :class="route.name === 'pricing' ? 'text-accent' : 'text-text-secondary hover:text-text-primary'"
        >
          Pricing
        </RouterLink>
      </div>

      <!-- Desktop: right actions -->
      <div class="hidden md:flex items-center gap-4 uppercase text-[12.5px] font-mono tracking-widest">
        <a
          href="https://github.com"
          target="_blank"
          rel="noopener noreferrer"
          class="flex items-center gap-2 px-4 py-2 border border-border-hover text-text-secondary hover:text-text-primary hover:border-text-muted rounded transition-colors"
        >
          <svg class="w-[15px] h-[15px]" fill="currentColor" viewBox="0 0 24 24"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
          GitHub
        </a>
        <RouterLink
          to="/login"
          class="text-text-secondary hover:text-text-primary transition-colors"
        >
          Login
        </RouterLink>
        <RouterLink
          to="/login"
          class="px-5 py-2 border border-accent/50 text-accent hover:bg-accent hover:text-bg-primary rounded transition-colors"
        >
          Get Started
        </RouterLink>
      </div>

      <!-- Mobile hamburger -->
      <button
        class="md:hidden p-2 text-text-secondary hover:text-text-primary transition-colors"
        aria-label="Open menu"
        @click="mobileOpen = !mobileOpen"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16"/></svg>
      </button>
    </div>

    <!-- Mobile menu -->
    <div v-if="mobileOpen" class="md:hidden border-t border-border bg-bg-primary/95 backdrop-blur-sm">
      <div class="px-6 py-4 flex flex-col gap-3 uppercase text-[12.5px] font-mono tracking-widest">
        <RouterLink to="/features" class="text-text-secondary hover:text-text-primary transition-colors" @click="mobileOpen = false">Platform</RouterLink>
        <RouterLink to="/pricing" class="text-text-secondary hover:text-text-primary transition-colors" @click="mobileOpen = false">Pricing</RouterLink>
        <a href="https://github.com" target="_blank" rel="noopener noreferrer" class="text-text-secondary hover:text-text-primary transition-colors">GitHub</a>
        <div class="border-t border-border pt-3 mt-1 flex flex-col gap-3">
          <RouterLink to="/login" class="text-text-secondary hover:text-text-primary transition-colors" @click="mobileOpen = false">Login</RouterLink>
          <RouterLink to="/login" class="text-center px-4 py-2 border border-accent/50 text-accent hover:bg-accent hover:text-bg-primary rounded transition-colors" @click="mobileOpen = false">Get Started</RouterLink>
        </div>
      </div>
    </div>
  </nav>
</template>
