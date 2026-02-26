<script setup lang="ts">
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

async function handleLogout() {
  await auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <nav class="w-56 border-r border-gray-200 p-4 flex flex-col gap-2">
    <RouterLink to="/" class="px-3 py-2 rounded hover:bg-gray-100">Dashboard</RouterLink>
    <RouterLink to="/connections" class="px-3 py-2 rounded hover:bg-gray-100">Connections</RouterLink>
    <RouterLink to="/agent/config" class="px-3 py-2 rounded hover:bg-gray-100">Agent Config</RouterLink>
    <RouterLink to="/agent/chat" class="px-3 py-2 rounded hover:bg-gray-100">Agent Chat</RouterLink>
    <RouterLink to="/agent/log" class="px-3 py-2 rounded hover:bg-gray-100">Agent Log</RouterLink>
    <RouterLink to="/reports" class="px-3 py-2 rounded hover:bg-gray-100">Reports</RouterLink>

    <div class="mt-auto pt-4 border-t border-gray-200">
      <div v-if="auth.user" class="px-3 py-1 text-xs text-gray-400 truncate">{{ auth.user.email }}</div>
      <button @click="handleLogout" class="w-full text-left px-3 py-2 rounded text-sm text-gray-500 hover:bg-gray-100 hover:text-gray-700">
        Sign out
      </button>
    </div>
  </nav>
</template>
