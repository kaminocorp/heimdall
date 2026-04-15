<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { listApplicationsWithCounts } from '@/api/applications'
import { extractApiError } from '@/utils/apiError'
import type { ApplicationWithCounts } from '@/types/organization'
import AppCard from '@/components/org/AppCard.vue'
import AppListRow from '@/components/org/AppListRow.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import AppWizard from '@/components/app-wizard/AppWizard.vue'

const appStore = useAppStore()
const router = useRouter()

// ── State ──
const apps = ref<ApplicationWithCounts[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const search = ref('')
const sortBy = ref<'name' | 'date' | 'status'>('name')
const viewMode = ref<'grid' | 'list'>(
  (localStorage.getItem('heimdall_org_view') as 'grid' | 'list') ?? 'grid'
)
const wizardOpen = ref(false)

// ── Fetch ──
async function fetchApps() {
  loading.value = true
  error.value = null
  try {
    apps.value = await listApplicationsWithCounts()
  } catch (e: unknown) {
    error.value = extractApiError(e, 'Failed to load applications')
  } finally {
    loading.value = false
  }
}

watch(() => appStore.organization?.id, () => fetchApps(), { immediate: true, flush: 'post' })

// ── Computed ──
const filteredApps = computed(() => {
  let result = apps.value

  // Search filter
  if (search.value.trim()) {
    const q = search.value.toLowerCase().trim()
    result = result.filter(a => a.name.toLowerCase().includes(q))
  }

  // Sort
  return [...result].sort((a, b) => {
    switch (sortBy.value) {
      case 'name':
        return a.name.localeCompare(b.name)
      case 'date':
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      case 'status':
        return a.status.localeCompare(b.status)
      default:
        return 0
    }
  })
})

// ── Actions ──
function selectApp(appId: string) {
  appStore.selectApp(appId)
  router.push('/dashboard')
}

function setViewMode(mode: 'grid' | 'list') {
  viewMode.value = mode
  localStorage.setItem('heimdall_org_view', mode)
}

function openWizard() {
  wizardOpen.value = true
}

async function onWizardClose() {
  wizardOpen.value = false
  await fetchApps()
}
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="pb-6 mb-8 border-b border-border">
      <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">
        Projects
      </h2>
      <p class="font-sans text-sm text-text-secondary mt-1">
        All applications in {{ appStore.organization?.name ?? 'your organization' }}.
      </p>
    </div>

    <!-- Toolbar: search + sort + view toggle + new app -->
    <div class="flex flex-wrap items-center gap-3 mb-6">
      <!-- Search -->
      <div class="relative flex-1 min-w-[200px] max-w-sm">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted pointer-events-none" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
        </svg>
        <input
          v-model="search"
          type="text"
          placeholder="Search projects..."
          class="w-full pl-10 pr-4 py-2 rounded border border-border bg-bg-surface font-mono text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent/50 transition-colors"
        />
      </div>

      <!-- Sort dropdown -->
      <select
        v-model="sortBy"
        class="px-3 py-2 rounded border border-border bg-bg-surface font-mono text-xs text-text-secondary focus:outline-none focus:border-accent/50 transition-colors cursor-pointer"
      >
        <option value="name">Sort by name</option>
        <option value="date">Sort by date</option>
        <option value="status">Sort by status</option>
      </select>

      <!-- View toggle -->
      <div class="flex rounded border border-border overflow-hidden">
        <button
          @click="setViewMode('grid')"
          class="px-2.5 py-2 transition-colors cursor-pointer"
          :class="viewMode === 'grid' ? 'bg-accent-subtle text-accent-bright' : 'bg-bg-surface text-text-muted hover:text-text-secondary'"
          title="Grid view"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z" />
          </svg>
        </button>
        <button
          @click="setViewMode('list')"
          class="px-2.5 py-2 transition-colors cursor-pointer"
          :class="viewMode === 'list' ? 'bg-accent-subtle text-accent-bright' : 'bg-bg-surface text-text-muted hover:text-text-secondary'"
          title="List view"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 010 3.75H5.625a1.875 1.875 0 010-3.75z" />
          </svg>
        </button>
      </div>

      <!-- New project button -->
      <button
        @click="openWizard"
        class="ml-auto px-4 py-2 font-mono text-xs font-medium uppercase tracking-wider rounded bg-action text-bg-primary hover:bg-action-hover transition-colors cursor-pointer flex items-center gap-2"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
        </svg>
        New project
      </button>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="n in 3" :key="n" class="rounded-lg border border-border bg-bg-surface p-5 space-y-3">
        <SkeletonBlock width="60%" height="1rem" />
        <SkeletonBlock width="40%" height="0.75rem" />
        <SkeletonBlock width="80%" height="0.75rem" />
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="rounded border border-status-critical/30 bg-status-critical/10 px-5 py-4">
      <p class="font-mono text-sm text-status-critical">{{ error }}</p>
      <button
        @click="fetchApps"
        class="mt-2 font-mono text-xs text-text-secondary hover:text-text-primary underline cursor-pointer"
      >
        Retry
      </button>
    </div>

    <!-- Empty state -->
    <div v-else-if="apps.length === 0" class="rounded-lg border border-border bg-bg-surface p-12 text-center">
      <div class="w-14 h-14 rounded-full bg-accent-subtle mx-auto mb-4 flex items-center justify-center">
        <svg class="w-7 h-7 text-accent" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375" />
        </svg>
      </div>
      <h3 class="font-mono text-sm font-semibold text-text-primary uppercase tracking-wider mb-2">
        No projects yet
      </h3>
      <p class="font-sans text-sm text-text-secondary mb-6 max-w-md mx-auto">
        Create your first project to start monitoring your applications with Heimdall.
      </p>
      <button
        @click="openWizard"
        class="px-6 py-2.5 font-mono text-sm font-medium uppercase tracking-wider rounded bg-action text-bg-primary hover:bg-action-hover transition-colors cursor-pointer"
      >
        Create your first project
      </button>
    </div>

    <!-- No search results -->
    <div v-else-if="filteredApps.length === 0" class="rounded-lg border border-border bg-bg-surface p-8 text-center">
      <p class="font-mono text-sm text-text-muted">
        No projects matching "{{ search }}"
      </p>
    </div>

    <!-- Grid view -->
    <div v-else-if="viewMode === 'grid'" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <AppCard
        v-for="a in filteredApps"
        :key="a.id"
        :app="a"
        @select="selectApp"
      />
    </div>

    <!-- List view -->
    <div v-else class="rounded-lg border border-border bg-bg-surface overflow-hidden">
      <!-- List header -->
      <div class="flex items-center gap-4 px-4 py-2 border-b border-border bg-bg-elevated">
        <div class="flex-1 font-mono text-xs font-medium uppercase tracking-wider text-text-muted">
          Name
        </div>
        <div class="shrink-0 w-28 font-mono text-xs font-medium uppercase tracking-wider text-text-muted text-right">
          Connections
        </div>
        <div class="shrink-0 w-28 font-mono text-xs font-medium uppercase tracking-wider text-text-muted text-right">
          Schedules
        </div>
        <div class="shrink-0 w-28 font-mono text-xs font-medium uppercase tracking-wider text-text-muted text-right hidden md:block">
          Created
        </div>
      </div>
      <AppListRow
        v-for="a in filteredApps"
        :key="a.id"
        :app="a"
        @select="selectApp"
      />
    </div>

    <!-- App wizard modal -->
    <AppWizard v-if="wizardOpen" @close="onWizardClose" />
  </div>
</template>
