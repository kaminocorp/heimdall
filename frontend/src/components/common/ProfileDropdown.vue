<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const auth = useAuthStore()
const app = useAppStore()
const router = useRouter()

const open = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

defineExpose({ el: dropdownRef, close: () => { open.value = false } })

const userInitial = computed(() => {
  const email = auth.user?.email
  return email ? email[0].toUpperCase() : '?'
})

function toggle() {
  open.value = !open.value
}

async function handleLogout() {
  open.value = false
  app.reset()
  try {
    await auth.logout()
  } catch {
    // Logout failed — state is already cleared, so force-navigate anyway.
  }
  router.push({ name: 'login' })
}

function goSettings() {
  open.value = false
  router.push({ name: 'org-settings' })
}

function goOrg() {
  open.value = false
  router.push('/org')
}
</script>

<template>
  <div ref="dropdownRef" class="relative">
    <button
      @click="toggle"
      class="flex items-center justify-center w-7 h-7 rounded-full border border-border bg-bg-surface text-xs font-mono font-bold text-text-secondary hover:text-text-primary hover:border-border-hover transition-colors cursor-pointer"
      :title="auth.user?.email ?? 'Profile'"
    >
      {{ userInitial }}
    </button>

    <Transition
      enter-active-class="transition duration-100 ease-out"
      enter-from-class="opacity-0 -translate-y-1"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-75 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-1"
    >
      <div v-if="open" class="absolute top-full right-0 mt-1 w-56 border border-border rounded bg-bg-elevated shadow-lg shadow-black/30 z-50">
        <div class="px-3 py-2.5 border-b border-border">
          <div v-if="auth.user" class="font-mono text-xs text-text-muted truncate">
            {{ auth.user.email }}
          </div>
        </div>
        <div class="py-1">
          <button
            @click="goOrg"
            class="w-full text-left px-3 py-2 text-sm font-mono text-text-secondary hover:text-text-primary hover:bg-bg-surface transition-colors cursor-pointer"
          >
            Organisation
          </button>
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
</template>
