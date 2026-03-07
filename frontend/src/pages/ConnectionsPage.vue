<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useConnectionsStore } from '@/stores/connections'
import { useAppStore } from '@/stores/app'
import { listConnectionsByApp } from '@/api/applications'
import type { Connection, CreateConnectionPayload } from '@/types/connection'
import ConnectionList from '@/components/connections/ConnectionList.vue'
import ConnectionForm from '@/components/connections/ConnectionForm.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'

const store = useConnectionsStore()
const appStore = useAppStore()
const showForm = ref(false)
const editingConnection = ref<Connection | null>(null)
const actionError = ref<string | null>(null)

async function fetchAppConnections() {
  const appId = appStore.currentAppId
  if (!appId) return
  store.loading = true
  store.error = null
  try {
    store.connections = await listConnectionsByApp(appId)
  } catch (e: any) {
    store.error = e.response?.data?.error ?? 'Failed to load connections'
  } finally {
    store.loading = false
  }
}

onMounted(fetchAppConnections)
watch(() => appStore.currentAppId, fetchAppConnections)

function openCreate() {
  editingConnection.value = null
  showForm.value = true
}

function openEdit(connection: Connection) {
  editingConnection.value = connection
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editingConnection.value = null
}

async function handleSubmit(payload: Omit<CreateConnectionPayload, 'app_id'>) {
  actionError.value = null
  const isEdit = !!editingConnection.value
  try {
    let connId: string
    if (editingConnection.value) {
      const updated = await store.updateConnection(editingConnection.value.id, {
        ...payload,
        status: 'inactive',
      })
      connId = updated.id
    } else {
      const created = await store.createConnection({
        ...payload,
        app_id: appStore.currentAppId!,
      })
      connId = created.id
    }
    closeForm()
    const result = await store.testConnection(connId)
    if (!result.success) {
      actionError.value = `${isEdit ? 'Connection updated' : 'Connection created'} but test failed: ${result.message}`
    }
    await fetchAppConnections()
  } catch (e: any) {
    actionError.value = e.response?.data?.error ?? `Failed to ${isEdit ? 'update' : 'create'} connection`
  }
}

async function handleTest(id: string) {
  actionError.value = null
  try {
    const result = await store.testConnection(id)
    if (!result.success) {
      actionError.value = `Connection test failed: ${result.message}`
    }
  } catch (e: any) {
    actionError.value = e.response?.data?.error ?? 'Failed to test connection'
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
        @click="openCreate"
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
      :initial-values="editingConnection"
      @submit="handleSubmit"
      @cancel="closeForm"
    />

    <!-- Skeleton loader -->
    <div v-if="store.loading" class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div v-for="n in 3" :key="n" class="border border-border rounded-lg bg-bg-surface p-5 space-y-3">
        <SkeletonBlock width="60%" height="1rem" />
        <SkeletonBlock width="40%" height="0.75rem" />
        <SkeletonBlock width="30%" height="0.75rem" />
      </div>
    </div>
    <div v-else-if="store.error" class="text-status-critical text-sm font-mono">{{ store.error }}</div>
    <ConnectionList v-else :connections="store.connections" :testing-id="store.testingId" @delete="handleDelete" @edit="openEdit" @test="handleTest" />
  </div>
</template>
