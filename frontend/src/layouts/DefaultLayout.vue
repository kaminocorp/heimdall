<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import AppHeader from '@/components/common/AppHeader.vue'
import AppSidebar from '@/components/common/AppSidebar.vue'
import OrgSidebar from '@/components/common/OrgSidebar.vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const mobileMenuOpen = ref(false)

const isOrgContext = computed(() => route.meta.context === 'org')

// Close mobile menu when context switches (e.g. org → app via card click).
watch(isOrgContext, () => {
  mobileMenuOpen.value = false
})
</script>

<template>
  <!-- Hide sidebar + layout on login, public, and onboarding pages -->
  <template v-if="route.name === 'login' || route.name === 'onboarding' || route.meta?.public">
    <slot />
  </template>

  <template v-else>
    <div class="h-screen flex flex-col bg-bg-primary overflow-hidden">
      <!-- Top header bar (full width) -->
      <AppHeader />

      <!-- Below header: sidebar + content -->
      <div class="flex flex-1 min-h-0">
        <!-- Desktop sidebar (context-aware, crossfade on switch) -->
        <aside class="hidden lg:flex flex-shrink-0">
          <Transition name="sidebar" mode="out-in">
            <OrgSidebar v-if="isOrgContext" key="org" />
            <AppSidebar v-else key="app" />
          </Transition>
        </aside>

        <!-- Mobile hamburger button -->
        <button
          class="lg:hidden fixed top-14 left-4 z-50 p-2 rounded border border-border bg-bg-elevated text-text-secondary hover:text-text-primary hover:border-border-hover transition-colors cursor-pointer"
          @click="mobileMenuOpen = true"
          aria-label="Open navigation"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>

        <!-- Mobile sidebar overlay -->
        <Teleport to="body">
          <Transition name="overlay">
            <div
              v-if="mobileMenuOpen"
              class="lg:hidden fixed inset-0 z-50 flex"
            >
              <!-- Backdrop -->
              <div
                class="absolute inset-0 bg-black/60 backdrop-blur-sm"
                @click="mobileMenuOpen = false"
              />
              <!-- Slide-over panel -->
              <div class="relative z-10">
                <OrgSidebar v-if="isOrgContext" mobile @close="mobileMenuOpen = false" />
              <AppSidebar v-else mobile @close="mobileMenuOpen = false" />
              </div>
            </div>
          </Transition>
        </Teleport>

        <!-- Main content -->
        <main class="flex-1 min-w-0 overflow-y-auto px-6 py-8 lg:px-8 lg:py-10">
          <slot />
        </main>
      </div>
    </div>
  </template>
</template>

<style scoped>
/* Sidebar context-switch crossfade */
.sidebar-enter-active,
.sidebar-leave-active {
  transition: opacity 0.15s ease;
}
.sidebar-enter-from,
.sidebar-leave-to {
  opacity: 0;
}

.overlay-enter-active,
.overlay-leave-active {
  transition: opacity 0.2s ease;
}
.overlay-enter-active > .relative {
  transition: transform 0.2s ease;
}
.overlay-leave-active > .relative {
  transition: transform 0.15s ease;
}
.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
}
.overlay-enter-from > .relative {
  transform: translateX(-100%);
}
.overlay-leave-to > .relative {
  transform: translateX(-100%);
}
</style>
