<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')

async function handleLogin() {
  // TODO: implement actual login API call
  auth.login('placeholder-token')
  router.push('/')
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen">
    <form @submit.prevent="handleLogin" class="w-full max-w-sm space-y-4">
      <h2 class="text-2xl font-semibold text-center">Sign in to Heimdall</h2>
      <div>
        <label class="block text-sm font-medium">Email</label>
        <input v-model="email" type="email" class="mt-1 block w-full border rounded px-3 py-2" />
      </div>
      <div>
        <label class="block text-sm font-medium">Password</label>
        <input v-model="password" type="password" class="mt-1 block w-full border rounded px-3 py-2" />
      </div>
      <button type="submit" class="w-full px-4 py-2 bg-gray-900 text-white rounded">Sign in</button>
    </form>
  </div>
</template>
