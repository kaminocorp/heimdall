<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterView, useRouter } from 'vue-router'
import DefaultLayout from './layouts/DefaultLayout.vue'
import ToastContainer from '@/components/common/ToastContainer.vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const auth = useAuthStore()
const app = useAppStore()
const router = useRouter()
const ready = ref(false)

onMounted(async () => {
  try {
    await auth.init()
    // If authenticated, load org + apps (determines onboarding redirect)
    if (auth.isAuthenticated) {
      await app.init()
    }
    // Re-evaluate the current route now that auth state is known.
    // The router guard skips checks while !initialized, so the initial
    // navigation may have landed on a protected route without a session.
    await router.replace(router.currentRoute.value.fullPath)
  } catch (e) {
    console.error('Initialization failed:', e)
    // Auth failed — send to login so the user can retry
    await router.replace({ name: 'login' })
  }
  ready.value = true
})
</script>

<template>
  <template v-if="ready">
    <DefaultLayout>
      <RouterView />
    </DefaultLayout>
  </template>
  <div v-else class="flex items-center justify-center min-h-screen bg-bg-primary text-text-muted font-mono text-sm uppercase tracking-wider">
    Initializing…
  </div>
  <ToastContainer />
</template>
