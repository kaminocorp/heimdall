<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const emit = defineEmits<{
  close: []
}>()

defineProps<{
  mobile?: boolean
}>()

const sections = [
  {
    label: 'Overview',
    items: [
      { name: 'Dashboard', to: '/', routeName: 'dashboard' },
    ],
  },
  {
    label: 'Infrastructure',
    items: [
      { name: 'Connections', to: '/connections', routeName: 'connections' },
    ],
  },
  {
    label: 'Agent',
    items: [
      { name: 'Configuration', to: '/agent/config', routeName: 'agent-config' },
      { name: 'Chat', to: '/agent/chat', routeName: 'agent-chat' },
      { name: 'Log', to: '/agent/log', routeName: 'agent-log' },
    ],
  },
  {
    label: 'Intelligence',
    items: [
      { name: 'Reports', to: '/reports', routeName: 'reports' },
    ],
  },
]

const currentRoute = computed(() => route.name)

function isActive(routeName: string): boolean {
  return currentRoute.value === routeName
}

async function handleLogout() {
  await auth.logout()
  router.push({ name: 'login' })
}

function handleNav() {
  emit('close')
}
</script>

<template>
  <nav
    class="w-60 flex flex-col bg-bg-elevated border-r border-border"
    :class="mobile ? 'h-full' : 'min-h-screen'"
  >
    <!-- Brand header -->
    <div class="px-5 py-5 border-b border-border">
      <div class="flex items-center gap-2.5">
        <div class="w-2 h-2 rounded-full bg-accent animate-pulse" />
        <span class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">
          Heimdall
        </span>
      </div>
      <div class="mt-1.5 pl-[18px] font-mono text-[10px] uppercase tracking-wider text-text-muted">
        Status: Active
      </div>
    </div>

    <!-- Navigation sections -->
    <div class="flex-1 overflow-y-auto py-4 px-3">
      <div v-for="section in sections" :key="section.label" class="mb-5">
        <!-- Section label -->
        <div class="px-2 mb-2 font-mono text-[10px] font-medium uppercase tracking-widest text-text-muted">
          {{ section.label }}
        </div>

        <!-- Nav items -->
        <RouterLink
          v-for="item in section.items"
          :key="item.routeName"
          :to="item.to"
          class="group flex items-center gap-2.5 px-2 py-1.5 rounded text-sm font-mono transition-colors relative"
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
            class="absolute left-0 top-1 bottom-1 w-0.5 rounded-full bg-accent"
          />
          <span class="pl-1">{{ item.name }}</span>
        </RouterLink>
      </div>
    </div>

    <!-- User footer -->
    <div class="mt-auto border-t border-border px-5 py-4">
      <div v-if="auth.user" class="font-mono text-[11px] text-text-muted truncate mb-2">
        {{ auth.user.email }}
      </div>
      <button
        @click="handleLogout"
        class="w-full text-left font-mono text-xs uppercase tracking-wider text-text-muted hover:text-status-critical transition-colors cursor-pointer"
      >
        Sign Out
      </button>
    </div>
  </nav>
</template>
