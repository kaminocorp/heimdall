<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()

const emit = defineEmits<{
  close: []
}>()

defineProps<{
  mobile?: boolean
}>()

const sections = [
  {
    label: 'Organisation',
    items: [
      { name: 'Projects', to: '/org', routeName: 'org-overview' },
      { name: 'Team', to: '/org/team', routeName: 'org-team' },
    ],
  },
  {
    label: 'Management',
    items: [
      { name: 'Billing', to: '/org/billing', routeName: 'org-billing' },
      { name: 'Settings', to: '/org/settings', routeName: 'org-settings' },
    ],
  },
]

const currentRoute = computed(() => route.name)

function isActive(routeName: string): boolean {
  return currentRoute.value === routeName
}

function handleNav() {
  emit('close')
}
</script>

<template>
  <nav
    class="w-56 flex flex-col bg-bg-elevated border-r border-border"
    :class="mobile ? 'h-full' : 'h-full'"
  >
    <!-- Navigation sections -->
    <div class="flex-1 overflow-y-auto py-4 px-3">
      <div v-for="section in sections" :key="section.label" class="mb-5">
        <!-- Section label -->
        <div class="px-2 mb-2 font-mono text-xs font-medium uppercase tracking-widest text-accent">
          {{ section.label }}
        </div>

        <!-- Nav items -->
        <RouterLink
          v-for="item in section.items"
          :key="item.routeName"
          :to="item.to"
          class="group flex items-center gap-2.5 px-3 py-2 rounded text-sm font-mono transition-colors relative"
          :class="
            isActive(item.routeName)
              ? 'bg-accent-subtle text-text-primary'
              : 'text-text-secondary hover:text-text-primary hover:bg-bg-surface'
          "
          @click="handleNav"
        >
          <!-- Active indicator bar -->
          <div
            v-if="isActive(item.routeName)"
            class="absolute left-0 top-1 bottom-1 w-0.5 rounded-full bg-accent-bright"
            style="box-shadow: 0 0 6px var(--accent-glow)"
          />
          <span class="pl-1">{{ item.name }}</span>
        </RouterLink>
      </div>
    </div>
  </nav>
</template>
