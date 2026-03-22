<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useConnectionsStore } from '@/stores/connections'
import { useAppStore } from '@/stores/app'
import { listConnectionsByApp } from '@/api/applications'
import type { Connection, CreateConnectionPayload } from '@/types/connection'
import ConnectionList from '@/components/connections/ConnectionList.vue'
import ConnectionForm from '@/components/connections/ConnectionForm.vue'
import ConnectionTestModal from '@/components/connections/ConnectionTestModal.vue'
import ConnectionWizard from '@/components/connections/wizard/ConnectionWizard.vue'
import GitHubRepoSelector from '@/components/connections/GitHubRepoSelector.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'

const store = useConnectionsStore()
const appStore = useAppStore()
const route = useRoute()
const router = useRouter()
const showForm = ref(false)
const showWizard = ref(false)
const editingConnection = ref<Connection | null>(null)
const actionError = ref<string | null>(null)
const repoSelectorConnectionId = ref<string | null>(null)
const githubInstalledMessage = ref<string | null>(null)
const testModalConnection = ref<Connection | null>(null)

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

onMounted(async () => {
  await fetchAppConnections()

  // Handle ?github=installed redirect from callback.
  if (route.query.github === 'installed') {
    githubInstalledMessage.value = 'GitHub App installed successfully. Select which repositories Heimdall can access.'
    // Find the newly created GitHub connection and open repo selector.
    const ghConn = store.connections.find(c => c.type === 'github')
    if (ghConn) {
      repoSelectorConnectionId.value = ghConn.id
    }
    // Clean up query param.
    router.replace({ query: {} })
  }
})

watch(() => appStore.currentAppId, fetchAppConnections)

function openCreate() {
  showWizard.value = true
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
  try {
    let conn: Connection
    if (editingConnection.value) {
      conn = await store.updateConnection(editingConnection.value.id, {
        ...payload,
        status: 'inactive',
      })
    } else {
      conn = await store.createConnection({
        ...payload,
        app_id: appStore.currentAppId!,
      })
    }
    closeForm()
    testModalConnection.value = conn
  } catch (e: any) {
    const isEdit = !!editingConnection.value
    actionError.value = e.response?.data?.error ?? `Failed to ${isEdit ? 'update' : 'create'} connection`
  }
}

function handleTest(id: string) {
  actionError.value = null
  const conn = store.connections.find(c => c.id === id)
  if (conn) testModalConnection.value = conn
}

function closeWizard() {
  showWizard.value = false
  fetchAppConnections()
}

function closeTestModal() {
  testModalConnection.value = null
  fetchAppConnections()
}

async function handleDelete(id: string) {
  actionError.value = null
  try {
    await store.deleteConnection(id)
  } catch (e: any) {
    actionError.value = e.response?.data?.error ?? 'Failed to delete connection'
  }
}

function openRepoSelector(connectionId: string) {
  repoSelectorConnectionId.value = connectionId
}

function closeRepoSelector() {
  repoSelectorConnectionId.value = null
  githubInstalledMessage.value = null
  fetchAppConnections()
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
        class="px-4 py-2 bg-action text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-action-hover transition-colors cursor-pointer"
      >
        + New Connection
      </button>
    </div>

    <!-- GitHub installed success banner -->
    <div v-if="githubInstalledMessage" class="mb-4 rounded border border-accent/30 bg-accent/10 px-4 py-2 text-sm font-mono text-accent">
      {{ githubInstalledMessage }}
    </div>

    <div v-if="actionError" class="mb-4 rounded border border-status-critical/30 bg-status-critical/10 px-4 py-2 text-sm font-mono text-status-critical">
      {{ actionError }}
    </div>

    <!-- Repo selector -->
    <GitHubRepoSelector
      v-if="repoSelectorConnectionId"
      :connection-id="repoSelectorConnectionId"
      class="mb-6"
      @close="closeRepoSelector"
    />

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
    <ConnectionList v-else :connections="store.connections" :testing-id="store.testingId" @delete="handleDelete" @edit="openEdit" @test="handleTest" @manage-repos="openRepoSelector" />

    <!-- Connection test modal -->
    <ConnectionTestModal
      v-if="testModalConnection"
      :connection="testModalConnection"
      @close="closeTestModal"
    />

    <!-- Connection wizard modal -->
    <ConnectionWizard
      v-if="showWizard"
      @close="closeWizard"
      @created="closeWizard"
    />
  </div>
</template>
