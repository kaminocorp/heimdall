<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function handleSubmit() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(email.value, password.value)
    router.push('/dashboard')
  } catch (e: any) {
    error.value = e.message ?? 'Authentication failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-bg-primary flex items-center justify-center px-4 relative overflow-hidden">
    <!-- Background grid pattern -->
    <div class="absolute inset-0 bg-grid opacity-[0.03] pointer-events-none" />

    <div class="w-full max-w-sm relative z-10">
      <!-- Brand -->
      <div class="text-center mb-8">
        <div class="flex items-center justify-center gap-2.5 mb-2">
          <div class="w-2 h-2 rounded-full bg-accent animate-pulse" />
          <span class="font-mono text-xl font-bold uppercase tracking-widest text-text-primary">Heimdall</span>
        </div>
        <p class="font-mono text-xs uppercase tracking-wider text-text-muted">Autonomous Monitoring Agent</p>
      </div>

      <!-- Login card -->
      <form @submit.prevent="handleSubmit" class="border border-border bg-bg-surface backdrop-blur-sm rounded-lg p-6 space-y-5">
        <div v-if="error" class="px-3 py-2 text-sm font-mono border border-status-critical/30 bg-status-critical/10 text-status-critical rounded">
          {{ error }}
        </div>

        <div>
          <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Email</label>
          <input
            v-model="email"
            type="email"
            required
            placeholder="operator@corp.io"
            class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
          />
        </div>
        <div>
          <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Password</label>
          <input
            v-model="password"
            type="password"
            required
            minlength="6"
            placeholder="••••••••••"
            class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
          />
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full px-4 py-2.5 bg-action text-bg-primary font-mono text-sm font-medium uppercase tracking-widest rounded hover:bg-action-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors cursor-pointer"
        >
          {{ loading ? 'Authenticating…' : 'Authenticate' }}
        </button>

      </form>
    </div>
  </div>
</template>

<style scoped>
.bg-grid {
  background-image:
    linear-gradient(rgba(77, 93, 83, 0.4) 1px, transparent 1px),
    linear-gradient(90deg, rgba(77, 93, 83, 0.4) 1px, transparent 1px);
  background-size: 40px 40px;
}
</style>
