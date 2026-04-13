<script setup lang="ts">
import { ref, computed } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import AppWizard from '@/components/app-wizard/AppWizard.vue'

const auth = useAuthStore()
const app = useAppStore()
const router = useRouter()

// ── Org dropdown ──
const orgOpen = ref(false)
const orgRef = ref<HTMLElement | null>(null)

function toggleOrg() {
  orgOpen.value = !orgOpen.value
  appOpen.value = false
  profileOpen.value = false
}

// ── App dropdown ──
const appOpen = ref(false)
const appRef = ref<HTMLElement | null>(null)
const wizardOpen = ref(false)

function toggleApp() {
  appOpen.value = !appOpen.value
  orgOpen.value = false
  profileOpen.value = false
}

function handleSelectApp(appId: string) {
  app.selectApp(appId)
  appOpen.value = false
}

function openNewApp() {
  appOpen.value = false
  wizardOpen.value = true
}

// ── Profile dropdown ──
const profileOpen = ref(false)
const profileRef = ref<HTMLElement | null>(null)

function toggleProfile() {
  profileOpen.value = !profileOpen.value
  orgOpen.value = false
  appOpen.value = false
}

async function handleLogout() {
  profileOpen.value = false
  app.reset()
  await auth.logout()
  router.push({ name: 'login' })
}

function goSettings() {
  profileOpen.value = false
  router.push({ name: 'settings' })
}

// ── Close on outside click ──
function onClickOutside(e: MouseEvent) {
  const target = e.target as Node
  if (orgRef.value && !orgRef.value.contains(target)) orgOpen.value = false
  if (appRef.value && !appRef.value.contains(target)) appOpen.value = false
  if (profileRef.value && !profileRef.value.contains(target)) profileOpen.value = false
}

import { onMounted, onBeforeUnmount } from 'vue'
onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))

// ── Computed ──
const currentAppName = computed(() => app.currentApp?.name ?? 'Select app')
const userInitial = computed(() => {
  const email = auth.user?.email
  return email ? email[0].toUpperCase() : '?'
})
</script>

<template>
  <header class="h-12 flex items-center border-b border-border bg-bg-elevated px-4 flex-shrink-0 z-40">
    <!-- Left: Logo + breadcrumbs -->
    <div class="flex items-center gap-1 min-w-0">
      <!-- Logo icon -->
      <RouterLink to="/dashboard" class="flex items-center gap-2.5 flex-shrink-0 mr-1">
        <!-- Heimdall eye icon — inline SVG so we can color it with the accent -->
        <svg class="w-5 h-5 text-accent" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
          <g transform="translate(100,100)" fill="currentColor">
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(-10)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(10)"/>
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(50)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(70)"/>
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(110)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(130)"/>
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(170)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(190)"/>
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(230)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(250)"/>
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(290)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(310)"/>
            <circle r="5.5"/>
          </g>
        </svg>
        <span class="hidden sm:inline font-mono text-xs font-semibold uppercase tracking-[0.25em] text-text-primary">
          H E I M D A L L
        </span>
      </RouterLink>

      <!-- Separator -->
      <span class="text-text-muted text-lg font-light mx-1 select-none">/</span>

      <!-- Org breadcrumb -->
      <div ref="orgRef" class="relative">
        <button
          @click="toggleOrg"
          class="flex items-center gap-1.5 px-2 py-1 rounded text-sm font-mono text-text-secondary hover:text-text-primary hover:bg-bg-surface transition-colors cursor-pointer"
        >
          <!-- Org icon -->
          <svg class="w-3.5 h-3.5 text-text-muted flex-shrink-0" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />
          </svg>
          <span class="truncate max-w-[160px]">{{ app.organization?.name ?? 'Organization' }}</span>
          <svg class="w-3 h-3 text-text-muted flex-shrink-0" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M3 4.5L6 7.5L9 4.5" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>

        <!-- Org dropdown (currently single-org, but ready for multi-org) -->
        <Transition
          enter-active-class="transition duration-100 ease-out"
          enter-from-class="opacity-0 -translate-y-1"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition duration-75 ease-in"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 -translate-y-1"
        >
          <div v-if="orgOpen" class="absolute top-full left-0 mt-1 w-56 border border-border rounded bg-bg-elevated shadow-lg shadow-black/30 z-50">
            <div class="py-1">
              <button
                v-if="app.organization"
                class="w-full text-left px-3 py-2 text-sm font-mono text-accent-bright bg-accent-subtle cursor-default"
              >
                {{ app.organization.name }}
              </button>
            </div>
          </div>
        </Transition>
      </div>

      <!-- Separator -->
      <span class="text-text-muted text-lg font-light mx-1 select-none">/</span>

      <!-- App breadcrumb -->
      <div ref="appRef" class="relative">
        <button
          @click="toggleApp"
          class="flex items-center gap-1.5 px-2 py-1 rounded text-sm font-mono text-text-secondary hover:text-text-primary hover:bg-bg-surface transition-colors cursor-pointer"
        >
          <!-- App icon -->
          <svg class="w-3.5 h-3.5 text-text-muted flex-shrink-0" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" />
          </svg>
          <span class="truncate max-w-[160px]">{{ currentAppName }}</span>
          <svg class="w-3 h-3 text-text-muted flex-shrink-0" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M3 4.5L6 7.5L9 4.5" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>

        <!-- App dropdown -->
        <Transition
          enter-active-class="transition duration-100 ease-out"
          enter-from-class="opacity-0 -translate-y-1"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition duration-75 ease-in"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 -translate-y-1"
        >
          <div v-if="appOpen" class="absolute top-full left-0 mt-1 w-56 border border-border rounded bg-bg-elevated shadow-lg shadow-black/30 z-50">
            <div class="py-1 max-h-64 overflow-y-auto">
              <button
                v-for="a in app.applications"
                :key="a.id"
                @click="handleSelectApp(a.id)"
                class="w-full text-left px-3 py-2 text-sm font-mono transition-colors cursor-pointer"
                :class="a.id === app.currentAppId
                  ? 'text-accent-bright bg-accent-subtle'
                  : 'text-text-secondary hover:text-text-primary hover:bg-bg-surface'"
              >
                {{ a.name }}
              </button>
            </div>
            <div class="border-t border-border py-1">
              <button
                @click="openNewApp"
                class="w-full text-left px-3 py-2 text-sm font-mono text-text-muted hover:text-text-primary hover:bg-bg-surface transition-colors cursor-pointer"
              >
                + New application
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </div>

    <!-- Right: Agent chat + Profile -->
    <div class="ml-auto flex items-center gap-2">
      <!-- Agent chat button -->
      <RouterLink
        to="/agent/chat"
        class="flex items-center justify-center w-7 h-7 rounded-full border border-border bg-bg-surface text-text-secondary hover:text-accent hover:border-accent/40 transition-colors"
        title="Agent Chat"
      >
        <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <!-- Hexagon node shape — symmetrical, technical, non-cringey -->
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 2L3 7v10l9 5 9-5V7l-9-5z" />
          <circle cx="12" cy="12" r="2.5" fill="currentColor" stroke="none" />
        </svg>
      </RouterLink>

      <div ref="profileRef" class="relative">
        <button
          @click="toggleProfile"
          class="flex items-center justify-center w-7 h-7 rounded-full border border-border bg-bg-surface text-xs font-mono font-bold text-text-secondary hover:text-text-primary hover:border-border-hover transition-colors cursor-pointer"
          :title="auth.user?.email ?? 'Profile'"
        >
          {{ userInitial }}
        </button>

        <!-- Profile dropdown -->
        <Transition
          enter-active-class="transition duration-100 ease-out"
          enter-from-class="opacity-0 -translate-y-1"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition duration-75 ease-in"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 -translate-y-1"
        >
          <div v-if="profileOpen" class="absolute top-full right-0 mt-1 w-56 border border-border rounded bg-bg-elevated shadow-lg shadow-black/30 z-50">
            <!-- User info -->
            <div class="px-3 py-2.5 border-b border-border">
              <div v-if="auth.user" class="font-mono text-xs text-text-muted truncate">
                {{ auth.user.email }}
              </div>
            </div>
            <div class="py-1">
              <button
                @click="goSettings"
                class="w-full text-left px-3 py-2 text-sm font-mono text-text-secondary hover:text-text-primary hover:bg-bg-surface transition-colors cursor-pointer"
              >
                Settings
              </button>
              <button
                @click="handleLogout"
                class="w-full text-left px-3 py-2 text-sm font-mono text-text-secondary hover:text-status-critical hover:bg-bg-surface transition-colors cursor-pointer"
              >
                Sign out
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </div>

    <!-- App wizard modal (teleported) -->
    <AppWizard v-if="wizardOpen" @close="wizardOpen = false" />
  </header>
</template>
