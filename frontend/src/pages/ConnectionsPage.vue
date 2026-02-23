<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useConnectionsStore } from '@/stores/connections'
import type { CreateConnectionPayload } from '@/types/connection'
import ConnectionList from '@/components/connections/ConnectionList.vue'
import ConnectionForm from '@/components/connections/ConnectionForm.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const store = useConnectionsStore()
const showForm = ref(false)
const actionError = ref<string | null>(null)

onMounted(() => {
  store.fetchConnections()
})

async function handleCreate(payload: CreateConnectionPayload) {
  actionError.value = null
  try {
    await store.createConnection(payload)
    showForm.value = false
  } catch (e: any) {
    actionError.value = e.response?.data?.error ?? 'Failed to create connection'
  }
}

async function handleDelete(id: string) {
  actionError.value = null
  try {
    await store.deleteConnection(id)
  } catch (e: any) {
    actionError.value = e.response?.data?.error ?? 'Failed to delete connection'
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-2xl font-semibold">Connections</h2>
      <button
        v-if="!showForm"
        @click="showForm = true"
        class="px-4 py-2 bg-gray-900 text-white rounded"
      >
        New Connection
      </button>
    </div>

    <div v-if="actionError" class="mb-4 p-3 bg-red-50 text-red-700 rounded text-sm">
      {{ actionError }}
    </div>

    <ConnectionForm
      v-if="showForm"
      class="mb-6"
      @submit="handleCreate"
      @cancel="showForm = false"
    />

    <LoadingSpinner v-if="store.loading" />
    <div v-else-if="store.error" class="text-red-600 text-sm">{{ store.error }}</div>
    <ConnectionList v-else :connections="store.connections" @delete="handleDelete" />
  </div>
</template>
