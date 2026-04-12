<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { useSchedulesStore } from '@/stores/schedules'
import { useToast } from '@/composables/useToast'
import { extractApiError } from '@/utils/apiError'
import type { InvestigationSchedule, ScheduleInput } from '@/types/schedule'
import ScheduleCard from '@/components/schedules/ScheduleCard.vue'
import ScheduleModal from '@/components/schedules/ScheduleModal.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'

const appStore = useAppStore()
const schedulesStore = useSchedulesStore()
const toast = useToast()

const showModal = ref(false)
// editingSchedule is null for the create flow, or a row reference for edit.
// The modal keys off this prop to decide its heading and form seeding.
const editingSchedule = ref<InvestigationSchedule | null>(null)
const submitting = ref(false)

async function load() {
  const appId = appStore.currentAppId
  if (!appId) return
  await schedulesStore.fetchSchedules(appId)
}

onMounted(load)

// React to app switching — clear any in-progress modal and refetch.
watch(() => appStore.currentAppId, () => {
  showModal.value = false
  editingSchedule.value = null
  load()
})

function openCreate() {
  editingSchedule.value = null
  showModal.value = true
}

function openEdit(schedule: InvestigationSchedule) {
  editingSchedule.value = schedule
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editingSchedule.value = null
}



async function handleSubmit(input: ScheduleInput) {
  const appId = appStore.currentAppId
  if (!appId) return
  submitting.value = true
  try {
    if (editingSchedule.value) {
      await schedulesStore.updateSchedule(appId, editingSchedule.value.id, input)
      toast.show('Schedule updated', 'success')
    } else {
      await schedulesStore.createSchedule(appId, input)
      toast.show('Schedule created', 'success')
    }
    closeModal()
  } catch (e: unknown) {
    toast.show(extractApiError(e, 'Failed to save schedule'), 'error')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(schedule: InvestigationSchedule) {
  const appId = appStore.currentAppId
  if (!appId) return
  // Native confirm is sufficient for a destructive action that's easy to
  // redo — users can simply recreate a schedule. A custom confirmation
  // modal would be over-engineering here.
  if (!window.confirm(`Delete schedule "${schedule.name}"? This cannot be undone.`)) return
  try {
    await schedulesStore.deleteSchedule(appId, schedule.id)
    toast.show('Schedule deleted', 'success')
  } catch (e: unknown) {
    toast.show(extractApiError(e, 'Failed to delete schedule'), 'error')
  }
}

async function handleRunNow(schedule: InvestigationSchedule) {
  const appId = appStore.currentAppId
  if (!appId) return
  try {
    await schedulesStore.runNow(appId, schedule.id)
    toast.show('Run complete — check Activity for details', 'success')
  } catch (e: unknown) {
    toast.show(extractApiError(e, 'Run failed'), 'error')
  }
}

const hasSchedules = computed(() => schedulesStore.schedules.length > 0)
</script>

<template>
  <div>
    <!-- Header -->
    <div class="pb-6 mb-8 border-b border-border flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">
          Scheduled Investigations
        </h2>
        <p class="font-sans text-sm text-text-secondary mt-1">
          Cron-triggered agent investigations — runs a stored prompt on a schedule
          <span v-if="appStore.currentApp" class="text-text-muted">— {{ appStore.currentApp.name }}</span>
        </p>
      </div>
      <button
        type="button"
        @click="openCreate"
        class="px-5 py-2.5 bg-action text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-action-hover transition-colors cursor-pointer"
      >
        New Schedule
      </button>
    </div>

    <!-- Loading skeleton -->
    <div v-if="schedulesStore.loading && !hasSchedules" class="space-y-3">
      <div v-for="n in 3" :key="n" class="border border-border rounded-lg bg-bg-surface p-5 space-y-3">
        <SkeletonBlock width="40%" height="1rem" />
        <SkeletonBlock width="60%" height="0.75rem" />
        <SkeletonBlock width="100%" height="2rem" />
      </div>
    </div>

    <!-- Error banner -->
    <div
      v-else-if="schedulesStore.error"
      class="border border-status-critical/30 bg-status-critical/5 rounded-lg px-5 py-4"
    >
      <div class="font-mono text-xs uppercase tracking-wider text-status-critical mb-1">
        Load failed
      </div>
      <p class="font-mono text-sm text-text-secondary">{{ schedulesStore.error }}</p>
    </div>

    <!-- Empty state -->
    <div
      v-else-if="!hasSchedules"
      class="border border-border rounded-lg bg-bg-surface px-6 py-12 text-center"
    >
      <div class="font-mono text-sm uppercase tracking-wider text-text-muted mb-2">
        No schedules yet
      </div>
      <p class="font-sans text-sm text-text-secondary max-w-md mx-auto mb-5">
        Create a schedule to run an agent investigation on a cron. Useful for
        slow-query checks, GitHub TODO audits, or daily health summaries —
        anything you'd want the agent to look at on a regular cadence.
      </p>
      <button
        type="button"
        @click="openCreate"
        class="px-5 py-2.5 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
      >
        Create your first schedule
      </button>
    </div>

    <!-- Schedule list -->
    <div v-else class="space-y-3">
      <ScheduleCard
        v-for="schedule in schedulesStore.schedules"
        :key="schedule.id"
        :schedule="schedule"
        :running="schedulesStore.runningId === schedule.id"
        @edit="openEdit"
        @delete="handleDelete"
        @run="handleRunNow"
      />
    </div>

    <!-- Create/edit modal -->
    <ScheduleModal
      v-if="showModal"
      :schedule="editingSchedule"
      :submitting="submitting"
      @close="closeModal"
      @submit="handleSubmit"
    />
  </div>
</template>
