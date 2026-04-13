<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import BaseSelect from '@/components/common/BaseSelect.vue'
import AppWizard from '@/components/app-wizard/AppWizard.vue'

// Sentinel value used to represent the "+ New App" row inside the BaseSelect
// dropdown. It's a string that can't collide with a real UUID. When the
// selector emits this value we intercept it in handleSelectApp and open the
// wizard modal *without* committing it as the current app — the user's
// previously-selected app stays live in case they discard mid-wizard.
const NEW_APP_SENTINEL = '__new_app__'

const auth = useAuthStore()
const app = useAppStore()
const route = useRoute()
const router = useRouter()

const wizardOpen = ref(false)

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
      { name: 'Dashboard', to: '/dashboard', routeName: 'dashboard' },
      { name: 'Activity', to: '/activity', routeName: 'activity' },
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
      { name: 'Schedules', to: '/schedules', routeName: 'schedules' },
      { name: 'Notifications', to: '/notifications', routeName: 'notifications' },
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

function handleSelectApp(value: string | number) {
  const v = String(value)
  if (v === NEW_APP_SENTINEL) {
    // Intercept: open the wizard but DON'T commit the sentinel as the
    // current app. Because BaseSelect is one-way (reads modelValue from
    // props), not calling selectApp means the dropdown re-reads the old
    // currentAppId on its next render and visually snaps back — no glitch.
    wizardOpen.value = true
    return
  }
  app.selectApp(v)
}

function closeWizard() {
  wizardOpen.value = false
}

// Appends the "+ New application" sentinel as the last option so users
// always see it at the bottom of the list regardless of how many apps
// they have. `BaseSelect` renders options in order, so appending here
// is the same as ordering visually.
const appOptions = computed(() => [
  ...app.applications.map(a => ({ value: a.id, label: a.name })),
  { value: NEW_APP_SENTINEL, label: '+ New application' },
])

async function handleLogout() {
  app.reset()
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
    <div class="px-6 py-6 border-b border-border">
      <div class="flex items-center gap-2.5">
        <div class="w-2 h-2 rounded-full bg-accent glow-pulse" />
        <span class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary brand-glow">
          Heimdall
        </span>
      </div>
      <div class="mt-1.5 pl-[18px] font-mono text-xs uppercase tracking-wider text-text-muted">
        Status: Active
      </div>
    </div>

    <!-- App selector -->
    <div v-if="app.applications.length > 0" class="px-3 py-3 border-b border-border">
      <label class="block font-mono text-xs uppercase tracking-widest text-text-muted mb-1.5 px-2">
        Application
      </label>
      <BaseSelect
        :modelValue="app.currentAppId ?? ''"
        @update:modelValue="handleSelectApp"
        :options="appOptions"
        size="sm"
      />
    </div>

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

    <!-- User footer -->
    <div class="mt-auto border-t border-border px-6 py-5">
      <!-- Settings link — placed above the org/email block so account-
           level actions feel grouped with account identity, rather than
           mixed into the per-app navigation above. -->
      <RouterLink
        :to="{ name: 'settings' }"
        class="block mb-3 font-mono text-xs uppercase tracking-wider transition-colors cursor-pointer"
        :class="isActive('settings')
          ? 'text-text-primary'
          : 'text-text-muted hover:text-text-primary'"
        @click="handleNav"
      >
        Settings
      </RouterLink>
      <div v-if="app.organization" class="font-mono text-xs uppercase tracking-wider text-text-muted mb-1.5">
        {{ app.organization.name }}
      </div>
      <div v-if="auth.user" class="font-mono text-xs text-text-muted truncate mb-2">
        {{ auth.user.email }}
      </div>
      <button
        @click="handleLogout"
        class="w-full text-left font-mono text-xs uppercase tracking-wider text-text-muted hover:text-status-critical transition-colors cursor-pointer"
      >
        Sign Out
      </button>
    </div>

    <!-- New-app wizard, rendered at the sidebar level so it's reachable
         from any route. Teleporting to <body> would be cleaner long-term,
         but the wizard is fixed-positioned with z-50 so it escapes the
         sidebar's flow regardless. -->
    <AppWizard v-if="wizardOpen" @close="closeWizard" />
  </nav>
</template>
