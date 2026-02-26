<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const isSignup = ref(false)
const loading = ref(false)
const error = ref('')

async function handleSubmit() {
  loading.value = true
  error.value = ''
  try {
    if (isSignup.value) {
      await auth.signup(email.value, password.value)
      // After signup, Supabase may require email confirmation.
      // If auto-confirm is on, onAuthStateChange will fire and the
      // router guard will redirect. Otherwise show a message.
      if (!auth.isAuthenticated) {
        error.value = 'Check your email to confirm your account.'
        loading.value = false
        return
      }
    } else {
      await auth.login(email.value, password.value)
    }
    router.push('/')
  } catch (e: any) {
    error.value = e.message ?? 'Authentication failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen">
    <form @submit.prevent="handleSubmit" class="w-full max-w-sm space-y-4">
      <h2 class="text-2xl font-semibold text-center">
        {{ isSignup ? 'Create an account' : 'Sign in to Heimdall' }}
      </h2>

      <div v-if="error" class="px-3 py-2 text-sm text-red-700 bg-red-50 border border-red-200 rounded">
        {{ error }}
      </div>

      <div>
        <label class="block text-sm font-medium">Email</label>
        <input v-model="email" type="email" required class="mt-1 block w-full border rounded px-3 py-2" />
      </div>
      <div>
        <label class="block text-sm font-medium">Password</label>
        <input v-model="password" type="password" required minlength="6" class="mt-1 block w-full border rounded px-3 py-2" />
      </div>

      <button type="submit" :disabled="loading" class="w-full px-4 py-2 bg-gray-900 text-white rounded disabled:opacity-50">
        {{ loading ? 'Please wait…' : isSignup ? 'Sign up' : 'Sign in' }}
      </button>

      <p class="text-sm text-center text-gray-500">
        {{ isSignup ? 'Already have an account?' : "Don't have an account?" }}
        <button type="button" class="underline text-gray-700" @click="isSignup = !isSignup; error = ''">
          {{ isSignup ? 'Sign in' : 'Sign up' }}
        </button>
      </p>
    </form>
  </div>
</template>
