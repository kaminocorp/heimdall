<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import { listSources, updateSources, addSource, deleteSource, discoverSources } from '@/api/sources'
import { classifyStaleness, type SourceFilterItem, type StalenessTier } from '@/types/source'
import { extractApiError } from '@/utils/apiError'

const props = withDefaults(defineProps<{
  connectionId: string
  // Embedded mode — when true the selector is rendered as an inline panel
  // (e.g. inside the Fly.io wizard) with no Close/Cancel buttons. The parent
  // owns the navigation; the selector just saves to the backend.
  embedded?: boolean
  // appId — required for org-scoped connections so the backend knows which
  // app's filter rows to return/write. App-scoped connections can omit it;
  // the backend derives the target app from the connection itself.
  appId?: string
  // discoverable — set for connector types with an explicit "list what we
  // have access to" upstream API (today: GitHub). When true:
  //   - polling is disabled (sources only arrive via explicit sync)
  //   - manual-add is hidden (users can only enable repos their install
  //     actually grants)
  //   - a "Sync" button fetches the upstream list and refreshes locally
  //   - an initial sync runs on mount if the local list is empty
  // Defaults to false so webhook/Fly.io flows are unaffected.
  discoverable?: boolean
}>(), { embedded: false, discoverable: false })

const emit = defineEmits<{
  close: []
}>()

// Working copy of the list — mutated locally as the user toggles, saved via
// a single bulk PUT on Save. This lets users experiment without hitting the
// network on every click.
const sources = ref<SourceFilterItem[]>([])
const loading = ref(true)
const saving = ref(false)
const error = ref<string | null>(null)

// Manual-add input. Creates the filter row with enabled=true immediately
// (separate request) so the user sees it appear without having to hit Save.
const newSourceName = ref('')
const adding = ref(false)

// Sync state — used by discoverable connectors (GitHub). Tracks whether a
// sync is in flight (to disable the button) and the count of sources found
// on the last successful sync (shown briefly as confirmation).
const syncing = ref(false)
const lastSyncDiscovered = ref<number | null>(null)

// Polling while open — catches new sources discovered by ingestion in real
// time. 10s cadence is cheap and user-visible during initial drain setup.
let pollTimer: ReturnType<typeof setInterval> | null = null

// Ticking timestamp used to force staleness recomputation once a minute.
// classifyStaleness reads Date.now() directly; without a reactive tick the
// dot colours would freeze at render time.
const nowTick = ref(Date.now())
let nowInterval: ReturnType<typeof setInterval> | null = null

// requestOpts is spread into every API call — the appId forwards to the
// backend as ?app_id=, which is a noop for app-scoped connections but the
// disambiguator for org-scoped ones.
const requestOpts = computed(() => (props.appId ? { appId: props.appId } : undefined))

async function load() {
  try {
    const items = await listSources(props.connectionId, requestOpts.value)
    // Sort: enabled first, then most-recently-seen, then alphabetical. This
    // keeps active configuration at the top where users look for it.
    sources.value = items.sort((a, b) => {
      if (a.enabled !== b.enabled) return a.enabled ? -1 : 1
      const aSeen = a.last_seen_at ? new Date(a.last_seen_at).getTime() : 0
      const bSeen = b.last_seen_at ? new Date(b.last_seen_at).getTime() : 0
      if (aSeen !== bSeen) return bSeen - aSeen
      return a.source_name.localeCompare(b.source_name)
    })
  } catch (e: unknown) {
    error.value = extractApiError(e, 'Failed to load sources')
  } finally {
    loading.value = false
  }
}

function toggle(src: SourceFilterItem) {
  src.enabled = !src.enabled
}

async function save() {
  saving.value = true
  error.value = null
  try {
    await updateSources(
      props.connectionId,
      sources.value.map(s => ({ source_name: s.source_name, enabled: s.enabled })),
      requestOpts.value,
    )
    // Standalone usage: closing the panel is the "done" signal. Embedded
    // usage: stay put — the parent's own navigation controls when to move on.
    if (!props.embedded) emit('close')
  } catch (e: unknown) {
    error.value = extractApiError(e, 'Failed to save source selection')
  } finally {
    saving.value = false
  }
}

async function addNew() {
  const name = newSourceName.value.trim()
  if (!name || adding.value) return
  adding.value = true
  error.value = null
  try {
    await addSource(props.connectionId, name, requestOpts.value)
    newSourceName.value = ''
    await load()
  } catch (e: unknown) {
    error.value = extractApiError(e, 'Failed to add source')
  } finally {
    adding.value = false
  }
}

async function remove(src: SourceFilterItem) {
  error.value = null
  try {
    await deleteSource(props.connectionId, src.source_name, requestOpts.value)
    await load()
  } catch (e: unknown) {
    error.value = extractApiError(e, 'Failed to remove source')
  }
}

async function sync() {
  if (syncing.value) return
  syncing.value = true
  error.value = null
  try {
    const result = await discoverSources(props.connectionId)
    lastSyncDiscovered.value = result.discovered
    await load()
  } catch (e: unknown) {
    error.value = extractApiError(e, 'Failed to sync sources')
  } finally {
    syncing.value = false
  }
}

// Pre-computed view list — one entry per source with its staleness tier.
// Using computed + nowTick ensures the dots update as time passes without
// the component having to re-fetch.
const view = computed(() => {
  // Touch nowTick so this recomputes when it ticks.
  void nowTick.value
  return sources.value.map(s => ({
    ...s,
    tier: classifyStaleness(s.last_seen_at) as StalenessTier,
  }))
})

const staleEnabledCount = computed(() =>
  view.value.filter(s => s.enabled && s.tier === 'stale').length,
)

const enabledCount = computed(() => view.value.filter(s => s.enabled).length)

function relativeTime(iso: string | null): string {
  if (!iso) return '(no traffic yet)'
  const ageMs = nowTick.value - new Date(iso).getTime()
  if (ageMs < 60_000) return 'just now'
  const mins = Math.floor(ageMs / 60_000)
  if (mins < 60) return `${mins} min ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} hour${hours === 1 ? '' : 's'} ago`
  const days = Math.floor(hours / 24)
  return `${days} day${days === 1 ? '' : 's'} ago`
}

function tierColor(tier: StalenessTier): string {
  // Matches the techno-brutalist palette: status-ok for active, status-warn
  // for quiet, status-critical for stale, muted for never-seen.
  switch (tier) {
    case 'active': return 'bg-status-ok'
    case 'quiet': return 'bg-status-warn'
    case 'stale': return 'bg-status-critical'
    case 'never': return 'bg-text-muted/40'
  }
}

onMounted(async () => {
  await load()
  // For discoverable connectors (GitHub): no passive traffic will ever
  // populate the list, so if we loaded an empty list trigger an initial
  // sync. Also skip the polling interval — the user explicitly re-syncs
  // via the button rather than us hammering GitHub every 10s.
  if (props.discoverable) {
    if (sources.value.length === 0) {
      await sync()
    }
  } else {
    pollTimer = setInterval(() => { load() }, 10_000)
  }
  nowInterval = setInterval(() => { nowTick.value = Date.now() }, 60_000)
})

onBeforeUnmount(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (nowInterval) clearInterval(nowInterval)
})
</script>

<template>
  <div class="border border-border rounded-lg p-6 bg-bg-surface space-y-4">
    <div class="flex items-center justify-between gap-3">
      <div class="min-w-0">
        <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">Manage Sources</h3>
        <p class="font-mono text-[11px] text-text-muted mt-1">
          {{ discoverable
            ? 'Sync pulls the latest accessible list from upstream. Toggle what to include.'
            : 'Only enabled sources are stored. Disabled sources are dropped at ingestion.' }}
        </p>
      </div>
      <div class="flex items-center gap-2 shrink-0">
        <!-- Manual sync button: only rendered for discoverable connectors
             (GitHub). Shows the count from the last successful sync for a
             beat so the user sees something happened. -->
        <button
          v-if="discoverable"
          @click="sync"
          :disabled="syncing"
          class="px-3 py-1.5 font-mono text-xs uppercase tracking-wider rounded border border-accent/40 text-accent hover:bg-accent/10 transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        >{{ syncing ? 'Syncing…' : (lastSyncDiscovered !== null ? `Synced (${lastSyncDiscovered})` : 'Sync') }}</button>
        <button
          v-if="!embedded"
          @click="emit('close')"
          class="text-text-muted hover:text-text-primary transition-colors text-sm font-mono cursor-pointer"
        >Close</button>
      </div>
    </div>

    <div v-if="error" class="rounded border border-status-critical/30 bg-status-critical/10 px-3 py-2 text-sm font-mono text-status-critical">
      {{ error }}
    </div>

    <div v-if="!discoverable && staleEnabledCount > 0" class="rounded border border-status-warn/40 bg-status-warn/10 px-3 py-2 font-mono text-[11px] text-status-warn">
      ⚠ {{ staleEnabledCount }} enabled source{{ staleEnabledCount === 1 ? '' : 's' }} stale (no traffic in 24h+).
      This may mean the app was removed or stopped.
    </div>

    <div v-if="loading" class="text-sm font-mono text-text-muted py-4">Loading sources…</div>

    <div v-else-if="view.length === 0" class="space-y-2 py-2">
      <p class="font-mono text-xs text-text-secondary">
        {{ discoverable
          ? 'No sources found. Click Sync to fetch from upstream.'
          : 'No sources discovered yet. They\'ll appear here as logs arrive, or add one manually below.' }}
      </p>
    </div>

    <div v-else class="space-y-1 max-h-80 overflow-y-auto">
      <div
        v-for="src in view"
        :key="src.source_name"
        class="group flex items-center justify-between px-3 py-2 rounded hover:bg-bg-elevated/50 transition-colors"
      >
        <div class="flex items-center gap-3 min-w-0 cursor-pointer flex-1" @click="toggle(src)">
          <!-- Staleness dot -->
          <span
            :class="['w-2 h-2 rounded-full shrink-0', tierColor(src.tier)]"
            :title="'Status: ' + src.tier"
          />
          <!-- Toggle checkbox -->
          <div
            class="w-4 h-4 rounded border flex items-center justify-center flex-shrink-0 transition-colors"
            :class="src.enabled ? 'bg-accent border-accent' : 'border-border'"
          >
            <svg v-if="src.enabled" class="w-3 h-3 text-bg-primary" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M2 6l3 3 5-5" />
            </svg>
          </div>
          <span class="font-mono text-sm text-text-primary truncate">{{ src.source_name }}</span>
        </div>
        <div class="flex items-center gap-3 shrink-0">
          <span class="font-mono text-[10px] text-text-muted">{{ relativeTime(src.last_seen_at) }}</span>
          <button
            @click="remove(src)"
            class="font-mono text-[10px] uppercase tracking-wider text-text-muted opacity-0 group-hover:opacity-100 hover:text-status-critical transition-opacity cursor-pointer"
          >Remove</button>
        </div>
      </div>
    </div>

    <!-- Manual add — hidden for discoverable connectors where arbitrary
         source names aren't meaningful (you can't grant yourself access to
         a GitHub repo by typing its name). Discoverable flows use the Sync
         button at the top instead. -->
    <form v-if="!discoverable" @submit.prevent="addNew" class="flex items-center gap-2 pt-3 border-t border-border">
      <input
        v-model="newSourceName"
        type="text"
        placeholder="Add source name…"
        class="flex-1 bg-bg-elevated border border-border rounded px-3 py-2 font-mono text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent"
      />
      <button
        type="submit"
        :disabled="adding || !newSourceName.trim()"
        class="px-4 py-2 border border-accent/40 text-accent font-mono text-xs uppercase tracking-wider rounded hover:bg-accent/10 transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
      >{{ adding ? 'Adding…' : 'Add' }}</button>
    </form>

    <!-- Footer actions -->
    <div class="flex items-center justify-between gap-2 pt-3 border-t border-border">
      <p class="font-mono text-[11px] text-text-muted">
        {{ enabledCount }} of {{ view.length }} enabled
      </p>
      <div class="flex gap-2">
        <button
          v-if="!embedded"
          @click="emit('close')"
          class="px-4 py-2 border border-accent-border/50 text-text-secondary font-mono text-xs uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
        >Cancel</button>
        <button
          :disabled="saving"
          @click="save"
          class="px-5 py-2 bg-action text-bg-primary font-mono text-xs font-medium uppercase tracking-wider rounded hover:bg-action-hover transition-colors cursor-pointer disabled:opacity-50"
        >{{ saving ? 'Saving…' : 'Save Selection' }}</button>
      </div>
    </div>
  </div>
</template>
