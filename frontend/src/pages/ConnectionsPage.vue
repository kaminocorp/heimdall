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
    <!-- Page header -->
    <div class="flex items-center justify-between pb-6 mb-8 border-b border-border">
      <div>
        <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">Connections</h2>
        <p class="font-sans text-sm text-text-secondary mt-1">Manage your infrastructure integrations</p>
      </div>
      <button
        v-if="!showForm"
        @click="showForm = true"
        class="px-4 py-2 bg-accent text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-accent-hover transition-colors cursor-pointer"
      >
        + New Connection
      </button>
    </div>

    <div v-if="actionError" class="mb-4 rounded border border-status-critical/30 bg-status-critical/10 px-4 py-2 text-sm font-mono text-status-critical">
      {{ actionError }}
    </div>

    <ConnectionForm
      v-if="showForm"
      class="mb-6"
      @submit="handleCreate"
      @cancel="showForm = false"
    />

    <LoadingSpinner v-if="store.loading" />
    <div v-else-if="store.error" class="text-status-critical text-sm font-mono">{{ store.error }}</div>
    <ConnectionList v-else :connections="store.connections" @delete="handleDelete" />
  </div>
</template>
