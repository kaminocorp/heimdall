<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { usePipelineStore } from '@/stores/pipeline'
import { usePipelineStream } from '@/composables/usePipelineStream'
import type { PipelineStage } from '@/types/pipeline'
import PipelineFunnel from '@/components/pipeline/PipelineFunnel.vue'
import PipelineParticles from '@/components/pipeline/PipelineParticles.vue'
import PipelineNodeCard from '@/components/pipeline/PipelineNodeCard.vue'
import PipelineNodeDetail from '@/components/pipeline/PipelineNodeDetail.vue'
import PipelineLogTicker from '@/components/pipeline/PipelineLogTicker.vue'
import PipelineJourneyModal from '@/components/pipeline/PipelineJourneyModal.vue'
import TimeMachineBlock from '@/components/pipeline/TimeMachineBlock.vue'

type StageKey = PipelineStage | 'activity'

const appStore = useAppStore()
const pipelineStore = usePipelineStore()
const { stats, rates } = storeToRefs(pipelineStore)
const route = useRoute()
const router = useRouter()

const funnelRef = ref<InstanceType<typeof PipelineFunnel> | null>(null)
const expandedStage = ref<StageKey | null>(null)
const inspectingLogId = ref<string | null>(null)

// Deep-link support: /pipeline/logs/:logId auto-opens the replay modal for
// the given log. Watching the param keeps the modal state in sync with the
// URL — closing the modal navigates back to /pipeline so the link is idempotent
// (refresh re-opens, close clears it).
watch(
  () => route.params.logId,
  (id) => {
    if (typeof id === 'string' && id) {
      inspectingLogId.value = id
    } else if (!id) {
      inspectingLogId.value = null
    }
  },
  { immediate: true },
)

// usePipelineStream is created lazily once we know currentAppId. When the app
// switches we tear down the previous stream and start a fresh one — both the
// SSE connection and the store buffer reset, so the new app starts with a
// clean slate instead of inheriting the previous app's events.
let streamHandle: ReturnType<typeof usePipelineStream> | null = null
const connectionState = ref<'idle' | 'bootstrapping' | 'streaming' | 'reconnecting' | 'resyncing' | 'error'>('idle')
const lastError = ref<string | null>(null)

async function startStream() {
  const appId = appStore.currentAppId
  if (!appId) return
  if (streamHandle) {
    streamHandle.destroy()
    streamHandle = null
  }
  pipelineStore.reset()
  streamHandle = usePipelineStream({ appId })
  // Forward reactive state out into local refs so the template can render it
  // without poking into the composable's return shape.
  watch(streamHandle.connectionState, (s) => (connectionState.value = s), { immediate: true })
  watch(streamHandle.lastError, (e) => (lastError.value = e), { immediate: true })
  try {
    await streamHandle.start()
  } catch {
    // Errors already surface through lastError; suppress unhandled-promise log.
  }
}

watch(
  () => appStore.currentAppId,
  () => { void startStream() },
  { immediate: true },
)

onBeforeUnmount(() => {
  streamHandle?.destroy()
  streamHandle = null
})

// Stage descriptors — count + rate sourced from store stats / rates. Activity
// inherits assessment counts since it represents the downstream view.
const stageCards = computed(() => [
  { key: 'ingestion' as StageKey, label: 'Ingestion', count: stats.value.ingestion_count, rate: rates.value.ingestion, tone: 'default' as const },
  { key: 'classified' as StageKey, label: 'Lumber', count: stats.value.classified_count, rate: rates.value.classified, tone: 'default' as const },
  { key: 'gate' as StageKey, label: 'Gate', count: stats.value.flagged_count, rate: rates.value.gate, tone: 'warn' as const },
  { key: 'assessment' as StageKey, label: 'Agent', count: stats.value.assessment_count, rate: rates.value.assessment, tone: 'default' as const },
  { key: 'activity' as StageKey, label: 'Activity', count: stats.value.assessment_count, rate: rates.value.assessment, tone: 'default' as const },
])

function toggleExpand(stage: StageKey) {
  expandedStage.value = expandedStage.value === stage ? null : stage
}

function inspectLog(logId: string) {
  inspectingLogId.value = logId
  // Mirror the state into the URL so the replay is shareable. Use replace
  // (not push) so the back button doesn't accumulate one entry per click.
  if (route.name === 'pipeline-log' && route.params.logId === logId) return
  router.replace({ name: 'pipeline-log', params: { logId } })
}

function closeJourney() {
  inspectingLogId.value = null
  // If we arrived via deep-link, step back to the plain /pipeline URL so
  // the next modal open pushes a fresh param cleanly.
  if (route.name === 'pipeline-log') {
    router.replace({ name: 'pipeline' })
  }
}

// Funnel exposes geometry; particles consume it. Wrap in a computed so the
// template stays declarative.
const geometry = computed(() => {
  const f = funnelRef.value
  if (!f) return null
  return {
    width: f.width,
    height: f.height,
    cy: f.cy,
    flaggedHalf: f.flaggedHalf,
    safeHalf: f.safeHalf,
    stagePxs: f.stagePoints.map((p: { px: number }) => p.px),
  }
})

// Empty-state messaging branches on whether the user even has sources
// enabled — a freshly onboarded user will see zero ingestion forever until
// they enable a source. Surface the right next step rather than a spinner.
const isEmpty = computed(() => stats.value.ingestion_count === 0)
const hasNoApps = computed(() => appStore.applications.length === 0)
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="pb-6 mb-6 border-b border-border flex items-center justify-between">
      <div>
        <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">Pipeline</h2>
        <p class="font-sans text-sm text-text-secondary mt-1">
          Every log's journey from ingestion to activity, live.
        </p>
      </div>
      <div class="flex items-center gap-2 font-mono text-[10px] uppercase tracking-widest">
        <span
          class="w-1.5 h-1.5 rounded-full"
          :class="{
            'bg-accent-bright glow-pulse': connectionState === 'streaming',
            'bg-status-warn': connectionState === 'reconnecting' || connectionState === 'resyncing',
            'bg-status-critical': connectionState === 'error',
            'bg-text-muted': connectionState === 'idle' || connectionState === 'bootstrapping',
          }"
        />
        <span class="text-text-muted">{{ connectionState }}</span>
      </div>
    </div>

    <div v-if="lastError" class="mb-4 rounded border border-status-critical/30 bg-status-critical/10 px-4 py-2 text-sm font-mono text-status-critical">
      {{ lastError }}
    </div>

    <!-- No apps yet — the user hasn't onboarded an app to scope to. -->
    <div v-if="hasNoApps" class="border border-border rounded bg-bg-surface px-6 py-8 text-center">
      <div class="font-mono text-sm text-text-secondary">
        Create an application to start watching its pipeline.
      </div>
    </div>

    <template v-else>
      <TimeMachineBlock
        v-if="appStore.currentAppId"
        :app-id="appStore.currentAppId"
        :initial-log-id="inspectingLogId"
        @inspect="inspectLog"
      />

      <!-- Funnel + particles + cards stack. The funnel reports its computed
           geometry; particles use it; node cards float above each anchor x. -->
      <div class="relative border border-border rounded bg-bg-elevated p-4">
        <!-- Stage cards row, planted above each stage anchor x. Use flex with
             evenly-distributed children so cards land roughly above their
             SVG counterparts without re-implementing the funnel's pad math. -->
        <div class="flex justify-between items-end gap-2 mb-2 px-[6%]">
          <PipelineNodeCard
            v-for="card in stageCards"
            :key="card.key"
            :label="card.label"
            :count="card.count"
            :rate="card.rate"
            :tone="card.tone"
            :active="rates[card.key === 'activity' ? 'assessment' : card.key] > 0"
            :expanded="expandedStage === card.key"
            @toggle="toggleExpand(card.key)"
          />
        </div>

        <div class="relative">
          <PipelineFunnel ref="funnelRef" />
          <PipelineParticles
            v-if="geometry"
            :width="geometry.width"
            :height="geometry.height"
            :cy="geometry.cy"
            :flagged-half="geometry.flaggedHalf"
            :safe-half="geometry.safeHalf"
            :stage-pxs="geometry.stagePxs"
          />
        </div>

        <PipelineNodeDetail
          v-if="expandedStage"
          :stage="expandedStage"
          @inspect="inspectLog"
        />
      </div>

      <!-- Empty state inline beneath the funnel. The funnel still renders so
           the user sees the structure before any traffic arrives. -->
      <div
        v-if="isEmpty && connectionState === 'streaming'"
        class="mt-4 border border-border rounded bg-bg-surface px-4 py-3 font-mono text-xs text-text-muted"
      >
        No ingestion events in the last hour. Enable a source on a connection to start the flow.
      </div>

      <div class="mt-4">
        <PipelineLogTicker @inspect="inspectLog" />
      </div>
    </template>

    <PipelineJourneyModal
      v-if="appStore.currentAppId"
      :app-id="appStore.currentAppId"
      :log-id="inspectingLogId"
      @close="closeJourney"
    />
  </div>
</template>
