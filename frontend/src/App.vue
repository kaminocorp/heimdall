<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterView } from 'vue-router'
import DefaultLayout from './layouts/DefaultLayout.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const ready = ref(false)

onMounted(async () => {
  await auth.init()
  ready.value = true
})
</script>

<template>
  <template v-if="ready">
    <DefaultLayout>
      <RouterView />
    </DefaultLayout>
  </template>
  <div v-else class="flex items-center justify-center min-h-screen text-gray-400">
    Loading…
  </div>
</template>
