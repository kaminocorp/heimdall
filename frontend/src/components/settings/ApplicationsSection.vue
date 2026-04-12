<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { listApplicationsWithCounts } from '@/api/applications'
import { formatDate } from '@/utils/format'
import type { ApplicationWithCounts } from '@/types/organization'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import DeleteAppModal from './DeleteAppModal.vue'
import AppWizard from '@/components/app-wizard/AppWizard.vue'

// ———————————————————————————————————————————————————————————————————————
// ApplicationsSection — the meat of the Settings page.
//
// Fetches the enriched `?include=counts` payload once on mount so each row
// can show its connection/schedule counts without a per-row fetch. The
// payload is stored locally (not in the app store) because the app store's
// `applications` list is the lean variant that every other consumer uses,
// and pushing counts into it would risk accidentally coupling the sidebar
// selector to a heavier payload.
//
// Delete flow: click opens DeleteAppModal with the row passed as a prop.
// On successful delete the section refetches — we could surgically splice
// the row out of the local list, but a refetch is simpler and keeps the
// UI consistent in the (unlikely) edge case where server-side state drift
// changed other rows while the modal was open.
//
// Last-app guard: the Delete button renders disabled when the local list
// has exactly one row. The backend has its own 409 guard (Phase 1), so
// this is UI-side defence-in-depth — users should never *see* the button
// enabled when they can't use it.
// ———————————————————————————————————————————————————————————————————————

const app = useAppStore()
const router = useRouter()

const rows = ref<ApplicationWithCounts[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const deleteTarget = ref<ApplicationWithCounts | null>(null)
const wizardOpen = ref(false)

const isOnlyApp = computed(() => rows.value.length <= 1)

async function fetchRows() {
  loading.value = true
  error.value = null
  try {
    rows.value = await listApplicationsWithCounts()
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } }; message?: string }
    error.value = err.response?.data?.error ?? err.message ?? 'Failed to load applications'
  } finally {
    loading.value = false
  }
}

onMounted(fetchRows)

// Click-through navigation: selecting the app first ensures that every
// per-app page (Connections, Schedules, etc.) renders data for the row
// the user actually clicked, not whatever was selected before.
function goToConnections(row: ApplicationWithCounts) {
  app.selectApp(row.id)
  router.push('/connections')
}

function goToSchedules(row: ApplicationWithCounts) {
  app.selectApp(row.id)
  router.push('/schedules')
}

function openDeleteModal(row: ApplicationWithCounts) {
  if (isOnlyApp.value) return // Defensive — button is already disabled.
  deleteTarget.value = row
}

function closeDeleteModal() {
  deleteTarget.value = null
}

async function onDeleted() {
  deleteTarget.value = null
  await fetchRows()
}

function openWizard() {
  wizardOpen.value = true
}

async function onWizardClose() {
  wizardOpen.value = false
  // The wizard may or may not have created a new app (user can discard
  // or finish). Refetching unconditionally is cheap and keeps the list
  // in sync with whatever state the user landed on, without having to
  // care about which path they took through the wizard.
  await fetchRows()
}
</script>

<template>
  <section class="border border-border rounded-lg bg-bg-surface">
    <header class="px-6 py-4 border-b border-border flex items-start justify-between gap-4">
      <div>
        <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">
          Applications
        </h3>
        <p class="mt-1 font-mono text-xs text-text-muted">
          All applications in your organisation.
        </p>
      </div>
      <button
        type="button"
        @click="openWizard"
        class="shrink-0 px-4 py-1.5 font-mono text-xs font-medium uppercase tracking-wider rounded bg-action text-bg-primary hover:bg-action-hover transition-colors cursor-pointer"
      >
        + New application
      </button>
    </header>

    <!-- Loading skeleton -->
    <div v-if="loading" class="px-6 py-5 space-y-3">
      <div v-for="n in 2" :key="n" class="border border-border/60 rounded p-4 space-y-2">
        <SkeletonBlock width="30%" height="1rem" />
        <SkeletonBlock width="60%" height="0.75rem" />
      </div>
    </div>

    <!-- Error banner -->
    <div v-else-if="error" class="px-6 py-5">
      <div class="rounded border border-status-critical/30 bg-status-critical/10 px-4 py-3">
        <p class="font-mono text-xs text-status-critical">{{ error }}</p>
      </div>
    </div>

    <!-- Empty (shouldn't happen normally — users always have >= 1 app because
         of the last-app guard, but render something sensible just in case). -->
    <div v-else-if="rows.length === 0" class="px-6 py-10 text-center">
      <p class="font-mono text-xs text-text-muted">
        No applications yet. Create your first one to get started.
      </p>
    </div>

    <!-- Rows -->
    <ul v-else class="divide-y divide-border">
      <li
        v-for="row in rows"
        :key="row.id"
        class="px-6 py-4 flex items-start gap-6"
      >
        <!-- Identity block -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-3 mb-1">
            <h4 class="font-mono text-sm font-medium text-text-primary truncate">
              {{ row.name }}
            </h4>
            <StatusBadge :status="row.status" />
          </div>
          <p class="font-mono text-xs text-text-muted">
            Created {{ formatDate(row.created_at) }}
          </p>

          <!-- Count click-throughs — clicking selects the app and routes
               to the corresponding per-app page, pre-filtered implicitly
               via `currentAppId`. -->
          <div class="mt-3 flex flex-wrap gap-4 font-mono text-xs">
            <button
              type="button"
              @click="goToConnections(row)"
              class="text-text-secondary hover:text-accent-bright underline-offset-2 hover:underline transition-colors cursor-pointer"
            >
              <span class="text-text-primary font-medium">{{ row.connection_count }}</span>
              {{ row.connection_count === 1 ? 'connection' : 'connections' }}
            </button>
            <button
              type="button"
              @click="goToSchedules(row)"
              class="text-text-secondary hover:text-accent-bright underline-offset-2 hover:underline transition-colors cursor-pointer"
            >
              <span class="text-text-primary font-medium">{{ row.schedule_count }}</span>
              {{ row.schedule_count === 1 ? 'schedule' : 'schedules' }}
            </button>
          </div>
        </div>

        <!-- Delete column -->
        <div class="flex flex-col items-end gap-1.5 shrink-0">
          <button
            type="button"
            @click="openDeleteModal(row)"
            :disabled="isOnlyApp"
            class="px-4 py-1.5 font-mono text-xs uppercase tracking-wider rounded border transition-colors"
            :class="isOnlyApp
              ? 'border-border/40 text-text-muted cursor-not-allowed'
              : 'border-border text-text-secondary hover:border-status-critical/50 hover:text-status-critical cursor-pointer'"
          >
            Delete
          </button>
          <!-- Explicit helper text for the last-app case. Renders only
               when the button is disabled for *this* specific reason, so
               users get a targeted explanation instead of hunting for a
               tooltip or guessing why they can't click. -->
          <p
            v-if="isOnlyApp"
            class="max-w-[200px] text-right font-mono text-[10px] leading-snug text-text-muted"
          >
            Your organisation must have at least one application.
            Create another first.
          </p>
        </div>
      </li>
    </ul>

    <!-- Delete modal, rendered at section level so it's reachable from
         any row. `v-if` keeps it unmounted until a target is chosen. -->
    <DeleteAppModal
      v-if="deleteTarget"
      :application="deleteTarget"
      @close="closeDeleteModal"
      @deleted="onDeleted"
    />

    <!-- New-app wizard, reused from Phase 2. Same behaviour as the
         sidebar's "+ New application" sentinel — discard cleans up, and
         a successful create lands on /dashboard while refetching our list. -->
    <AppWizard v-if="wizardOpen" @close="onWizardClose" />
  </section>
</template>
