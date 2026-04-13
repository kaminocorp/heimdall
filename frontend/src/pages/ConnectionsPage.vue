<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useConnectionsStore } from '@/stores/connections'
import { useAppStore } from '@/stores/app'
import { extractApiError } from '@/utils/apiError'
import type { Connection, CreateConnectionPayload } from '@/types/connection'
import ConnectionForm from '@/components/connections/ConnectionForm.vue'
import ConnectionTestModal from '@/components/connections/ConnectionTestModal.vue'
import ConnectionWizard from '@/components/connections/wizard/ConnectionWizard.vue'
import GitHubRepoSelector from '@/components/connections/GitHubRepoSelector.vue'
import AgentNebula from '@/components/connections/AgentNebula.vue'
import ConnectionBubble from '@/components/connections/ConnectionBubble.vue'
import ConnectionDetailModal from '@/components/connections/ConnectionDetailModal.vue'
import FlowLines from '@/components/connections/FlowLines.vue'
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
const selectedConnection = ref<Connection | null>(null)
const bubbleEls = ref<(HTMLElement | null)[]>([])
const nebulaEl = ref<HTMLElement | null>(null)

async function fetchAppConnections() {
  const appId = appStore.currentAppId
  if (!appId) return
  await store.fetchConnectionsByApp(appId)
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
  } catch (e: unknown) {
    const isEdit = !!editingConnection.value
    actionError.value = extractApiError(e, `Failed to ${isEdit ? 'update' : 'create'} connection`)
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
  } catch (e: unknown) {
    actionError.value = extractApiError(e, 'Failed to delete connection')
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
        class="px-5 py-2.5 bg-action text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-action-hover transition-colors cursor-pointer"
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
    <div v-if="store.loading" class="flex flex-wrap justify-center gap-4">
      <div v-for="n in 3" :key="n" class="flex flex-col items-center gap-2 p-4 min-w-[120px] border border-border rounded-xl bg-bg-surface">
        <SkeletonBlock width="48px" height="48px" />
        <SkeletonBlock width="80px" height="0.7rem" />
        <SkeletonBlock width="60px" height="0.6rem" />
      </div>
    </div>
    <div v-else-if="store.error" class="text-status-critical text-sm font-mono">{{ store.error }}</div>
    <template v-else>
      <!-- Ingestion view: Bubbles → Flow Lines → Nebula -->
      <div class="relative">
        <!-- Connection bubbles -->
        <div class="flex flex-wrap justify-center gap-4 relative z-10">
          <ConnectionBubble
            v-for="(conn, i) in store.connections"
            :key="conn.id"
            :ref="(el: any) => { bubbleEls[i] = el?.$el ?? el }"
            :connection="conn"
            :testing="store.testingId === conn.id"
            :index="i"
            @click="selectedConnection = conn"
          />
        </div>

        <!-- Spacer between bubbles and nebula -->
        <div class="h-16" />

        <!-- Flow lines (SVG overlay, hidden on mobile) -->
        <FlowLines
          v-if="store.connections.length > 0"
          :bubble-els="bubbleEls"
          :nebula-el="nebulaEl"
          :count="store.connections.length"
        />

        <!-- Agent nebula -->
        <div ref="nebulaEl">
          <AgentNebula :dormant="store.connections.length === 0" />
        </div>

        <!-- Empty state overlay -->
        <div
          v-if="store.connections.length === 0 && !store.loading"
          class="absolute inset-0 flex flex-col items-center justify-center z-10 pointer-events-none"
        >
          <p class="font-mono text-sm text-text-secondary mb-2">No connections yet</p>
          <p class="font-sans text-xs text-text-muted mb-4">Add your first integration to start monitoring</p>
          <button
            @click="openCreate"
            class="pointer-events-auto px-5 py-2 bg-action text-bg-primary font-mono text-xs font-medium uppercase tracking-wider rounded hover:bg-action-hover transition-colors cursor-pointer"
          >
            + New Connection
          </button>
        </div>
      </div>
    </template>

    <!-- Connection detail modal -->
    <ConnectionDetailModal
      v-if="selectedConnection"
      :connection="selectedConnection"
      @close="selectedConnection = null"
      @edit="(conn) => { selectedConnection = null; openEdit(conn) }"
      @test="(id) => { selectedConnection = null; handleTest(id) }"
      @delete="(id) => { selectedConnection = null; handleDelete(id) }"
      @manage-repos="(id) => { selectedConnection = null; openRepoSelector(id) }"
    />

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
      @manage-repos="(id: string) => { closeWizard(); openRepoSelector(id) }"
    />
  </div>
</template>
