<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useConnectionsStore } from '@/stores/connections'
import { useAppStore } from '@/stores/app'
import { extractApiError } from '@/utils/apiError'
import type { Connection, CreateConnectionPayload } from '@/types/connection'
import { typeToCategory, type ConnectionCategory } from '@/components/connections/wizard/flows'
import ConnectionForm from '@/components/connections/ConnectionForm.vue'
import ConnectionTestModal from '@/components/connections/ConnectionTestModal.vue'
import ConnectionWizard from '@/components/connections/wizard/ConnectionWizard.vue'
import SourceSelector from '@/components/connections/SourceSelector.vue'
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
const sourceSelectorConnectionId = ref<string | null>(null)
// Whether the currently open SourceSelector should be in discoverable mode
// (GitHub). Derived from the connection type at open time so we don't
// reactively flip mid-use.
const sourceSelectorDiscoverable = ref(false)
const githubInstalledMessage = ref<string | null>(null)
const testModalConnection = ref<Connection | null>(null)
const selectedConnection = ref<Connection | null>(null)
const bubbleEls = ref<(HTMLElement | null)[]>([])
const nebulaEl = ref<HTMLElement | null>(null)

// Group connections by category for three-lane layout
const ingestionConns = computed(() => store.connections.filter(c => typeToCategory(c.type) === 'ingestion'))
const enrichmentConns = computed(() => store.connections.filter(c => typeToCategory(c.type) === 'enrichment'))
const outboundConns = computed(() => store.connections.filter(c => typeToCategory(c.type) === 'outbound'))

// Category metadata for rendering lane headers
const categories: { key: ConnectionCategory; label: string; sublabel: string; icon: string }[] = [
  { key: 'ingestion', label: 'Ingestion', sublabel: 'Data flowing in', icon: 'down' },
  { key: 'enrichment', label: 'Enrichment', sublabel: 'Agent tools', icon: 'bidirectional' },
  { key: 'outbound', label: 'Outbound', sublabel: 'Alerts flowing out', icon: 'up' },
]

function connectionsForCategory(cat: ConnectionCategory): Connection[] {
  switch (cat) {
    case 'ingestion': return ingestionConns.value
    case 'enrichment': return enrichmentConns.value
    case 'outbound': return outboundConns.value
  }
}

// Build a flat list of all bubble elements with their category, preserving order for FlowLines
function setBubbleRef(el: any, globalIndex: number) {
  bubbleEls.value[globalIndex] = el?.$el ?? el
}

// Ordered list of categories per global bubble index — used by FlowLines
const bubbleCategories = computed<ConnectionCategory[]>(() => {
  const cats: ConnectionCategory[] = []
  for (const conn of ingestionConns.value) cats.push('ingestion')
  for (const conn of enrichmentConns.value) cats.push('enrichment')
  for (const conn of outboundConns.value) cats.push('outbound')
  return cats
})

async function fetchAppConnections() {
  const appId = appStore.currentAppId
  if (!appId) return
  // Reset bubble refs so stale DOM elements from a previous render don't
  // linger when connections are added or removed.
  bubbleEls.value = []
  await store.fetchConnectionsByApp(appId)
}

onMounted(async () => {
  await fetchAppConnections()

  // Handle ?github=installed redirect from callback.
  if (route.query.github === 'installed') {
    githubInstalledMessage.value = 'GitHub App installed successfully. Select which repositories Heimdall can access.'
    // Prefer the explicit connection_id param (Phase 6 Tier 3.4) so we open
    // the selector for the exact install that just finished, even when the
    // user has multiple GitHub connections. Falls back to first-match for
    // backwards compatibility with older callback links in flight.
    const targetId =
      typeof route.query.connection_id === 'string' ? route.query.connection_id : undefined
    const ghConn = targetId
      ? store.connections.find(c => c.id === targetId && c.type === 'github')
      : store.connections.find(c => c.type === 'github')
    if (ghConn) {
      sourceSelectorDiscoverable.value = true
      sourceSelectorConnectionId.value = ghConn.id
    }
    // Clean up query params.
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

async function handlePause(id: string) {
  actionError.value = null
  try {
    await store.pauseConnection(id)
  } catch (e: unknown) {
    actionError.value = extractApiError(e, 'Failed to pause connection')
  }
}

async function handleResume(id: string) {
  actionError.value = null
  try {
    await store.resumeConnection(id)
  } catch (e: unknown) {
    actionError.value = extractApiError(e, 'Failed to resume connection')
  }
}

async function handleDelete(id: string) {
  actionError.value = null
  try {
    await store.deleteConnection(id)
  } catch (e: unknown) {
    actionError.value = extractApiError(e, 'Failed to delete connection')
  }
}

function openSourceSelector(connectionId: string) {
  // Discoverable mode is picked from the connection's type at open time —
  // GitHub connections need the sync + no-manual-add UX; webhook/Fly.io
  // drains don't.
  const conn = store.connections.find(c => c.id === connectionId)
  sourceSelectorDiscoverable.value = conn?.type === 'github'
  sourceSelectorConnectionId.value = connectionId
}

function closeSourceSelector() {
  sourceSelectorConnectionId.value = null
  sourceSelectorDiscoverable.value = false
  // Clear the install banner if it was still visible from the GitHub
  // callback redirect — we used to do this on closeRepoSelector; the
  // source selector now owns the post-install flow too.
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

    <!-- Source selector — unified across webhook and GitHub connections in
         Phase 3. `discoverable` flips the UX for GitHub (manual sync + no
         free-form add). For org-scoped connections `appId` ensures the
         backend filters for the currently visible Heimdall app. -->
    <SourceSelector
      v-if="sourceSelectorConnectionId"
      :connection-id="sourceSelectorConnectionId"
      :app-id="appStore.currentAppId ?? undefined"
      :discoverable="sourceSelectorDiscoverable"
      class="mb-6"
      @close="closeSourceSelector"
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
      <!-- Three-lane layout: categorised bubbles → flow lines → nebula -->
      <div class="relative">
        <!-- Category lanes -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 relative z-10">
          <div v-for="cat in categories" :key="cat.key" class="category-lane">
            <!-- Lane header -->
            <div class="flex items-center gap-2 mb-3 px-1">
              <!-- Direction icon -->
              <svg v-if="cat.icon === 'down'" class="w-3.5 h-3.5 text-text-muted shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 5v14" /><path d="m19 12-7 7-7-7" />
              </svg>
              <svg v-else-if="cat.icon === 'bidirectional'" class="w-3.5 h-3.5 text-text-muted shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M8 3 4 7l4 4" /><path d="M4 7h16" /><path d="m16 21 4-4-4-4" /><path d="M20 17H4" />
              </svg>
              <svg v-else class="w-3.5 h-3.5 text-text-muted shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 19V5" /><path d="m5 12 7-7 7 7" />
              </svg>
              <div>
                <h3 class="font-mono text-sm font-bold uppercase tracking-widest text-text-primary leading-tight">
                  {{ cat.label }}
                </h3>
                <p class="font-mono text-[11px] uppercase tracking-wider text-text-muted leading-tight">
                  {{ cat.sublabel }}
                </p>
              </div>
            </div>

            <!-- Bubbles in this category — 2-column grid for consistent alignment -->
            <div class="grid grid-cols-2 gap-3 min-h-[80px]">
              <ConnectionBubble
                v-for="(conn, i) in connectionsForCategory(cat.key)"
                :key="conn.id"
                :ref="(el: any) => {
                  const offset = cat.key === 'ingestion' ? 0
                    : cat.key === 'enrichment' ? ingestionConns.length
                    : ingestionConns.length + enrichmentConns.length
                  setBubbleRef(el, offset + i)
                }"
                :connection="conn"
                :testing="store.testingId === conn.id"
                :index="i"
                @click="selectedConnection = conn"
              />
              <!-- Empty lane hint -->
              <div
                v-if="connectionsForCategory(cat.key).length === 0"
                class="col-span-2 flex items-center justify-center text-center py-4"
              >
                <span class="font-mono text-[10px] uppercase tracking-wider text-text-muted/50">
                  No {{ cat.label.toLowerCase() }} connections
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Spacer between bubbles and nebula -->
        <div class="h-16" />

        <!-- Flow lines (SVG overlay, hidden on mobile) -->
        <FlowLines
          v-if="store.connections.length > 0"
          :bubble-els="bubbleEls"
          :nebula-el="nebulaEl"
          :count="store.connections.length"
          :categories="bubbleCategories"
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
      @pause="(id) => { selectedConnection = null; handlePause(id) }"
      @resume="(id) => { selectedConnection = null; handleResume(id) }"
      @manage-sources="(id) => { selectedConnection = null; openSourceSelector(id) }"
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
      @manage-sources="(id: string) => { closeWizard(); openSourceSelector(id) }"
    />
  </div>
</template>
