<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterView, useRouter } from 'vue-router'
import DefaultLayout from './layouts/DefaultLayout.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const ready = ref(false)

onMounted(async () => {
  await auth.init()
  // Re-evaluate the current route now that auth state is known.
  // The router guard skips checks while !initialized, so the initial
  // navigation may have landed on a protected route without a session.
  await router.replace(router.currentRoute.value.fullPath)
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
</template>
