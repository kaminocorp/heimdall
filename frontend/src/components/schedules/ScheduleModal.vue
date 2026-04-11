<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import type { InvestigationSchedule, ScheduleInput } from '@/types/schedule'
import { cronPresets } from '@/utils/cron-presets'

const props = defineProps<{
  // When editing: the schedule row to seed the form. When creating: null.
  schedule: InvestigationSchedule | null
  submitting: boolean
}>()

const emit = defineEmits<{
  close: []
  submit: [input: ScheduleInput]
}>()

// Form state. The "mode" toggle between interval presets and raw cron lives
// entirely in the modal — it's only used to produce a submit payload, not
// stored on the schedule row itself.
const name = ref('')
const prompt = ref('')
const enabled = ref(true)

// Scheduling mode: 'preset' shows cron preset buttons (which under the hood
// ARE cron expressions — Phase 4's design decision), while 'custom' exposes
// a raw textbox for advanced users. Both submit as cron_expr. The 'interval'
// mode is kept as a third option so existing Phase 3 interval-only schedules
// remain editable without being force-migrated to cron on save.
const mode = ref<'preset' | 'custom' | 'interval'>('preset')
const presetValue = ref(cronPresets[2].value) // default: "Every hour"
const customCron = ref('')
const intervalSecs = ref(300) // default: 5 minutes

// seedForm fills the form state from a schedule row (edit path) or resets
// it to defaults (create path). Called on mount AND whenever props.schedule
// changes (handles the "close modal, reopen with different row" pattern).
function seedForm() {
  if (props.schedule) {
    name.value = props.schedule.name
    prompt.value = props.schedule.prompt
    enabled.value = props.schedule.enabled

    if (props.schedule.cron_expr) {
      // If this cron expression matches a preset, keep the user in preset
      // mode (friendlier). Otherwise drop them into custom mode pre-filled.
      const matched = cronPresets.find((p) => p.value === props.schedule!.cron_expr)
      if (matched) {
        mode.value = 'preset'
        presetValue.value = matched.value
      } else {
        mode.value = 'custom'
        customCron.value = props.schedule.cron_expr
      }
    } else {
      // Legacy interval-only schedule.
      mode.value = 'interval'
      intervalSecs.value = props.schedule.interval_secs
    }
  } else {
    name.value = ''
    prompt.value = ''
    enabled.value = true
    mode.value = 'preset'
    presetValue.value = cronPresets[2].value
    customCron.value = ''
    intervalSecs.value = 300
  }
}

onMounted(() => {
  seedForm()
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeydown)
})

watch(() => props.schedule, seedForm)

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && !props.submitting) emit('close')
}

// buildInput assembles the submit payload based on the current mode. Cron
// modes send cron_expr and omit interval_secs; interval mode sends
// interval_secs and omits cron_expr. This matches the backend's
// exactly-one-of validation contract.
function buildInput(): ScheduleInput | null {
  const trimmedName = name.value.trim()
  const trimmedPrompt = prompt.value.trim()
  if (!trimmedName || !trimmedPrompt) return null

  const base: ScheduleInput = {
    name: trimmedName,
    prompt: trimmedPrompt,
    enabled: enabled.value,
  }

  if (mode.value === 'preset') {
    return { ...base, cron_expr: presetValue.value }
  }
  if (mode.value === 'custom') {
    const expr = customCron.value.trim()
    if (!expr) return null
    return { ...base, cron_expr: expr }
  }
  // interval mode
  return { ...base, interval_secs: intervalSecs.value }
}

const canSubmit = computed(() => {
  if (props.submitting) return false
  if (!name.value.trim() || !prompt.value.trim()) return false
  if (mode.value === 'custom' && !customCron.value.trim()) return false
  return true
})

const heading = computed(() => (props.schedule ? 'Edit Schedule' : 'New Schedule'))

function onSubmit() {
  const input = buildInput()
  if (!input) return
  emit('submit', input)
}
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-start justify-center bg-black/60 backdrop-blur-sm overflow-y-auto py-10"
    @click.self="!submitting && emit('close')"
  >
    <div class="w-full max-w-xl mx-4 border border-border rounded-lg bg-bg-surface shadow-2xl">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-border">
        <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">
          {{ heading }}
        </h3>
        <button
          type="button"
          @click="emit('close')"
          :disabled="submitting"
          class="font-mono text-lg text-text-muted hover:text-text-primary transition-colors cursor-pointer disabled:cursor-not-allowed"
        >&times;</button>
      </div>

      <!-- Body -->
      <form @submit.prevent="onSubmit" class="px-6 py-5 space-y-5">
        <!-- Name -->
        <div>
          <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Name</label>
          <input
            v-model="name"
            type="text"
            maxlength="100"
            required
            placeholder="e.g. Hourly pg_stat check"
            class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors"
          />
        </div>

        <!-- Prompt -->
        <div>
          <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Prompt</label>
          <textarea
            v-model="prompt"
            rows="5"
            maxlength="5000"
            required
            placeholder="What should the agent investigate on every run? e.g. Query pg_stat_statements for queries taking >1s in the last hour and summarise the worst offenders."
            class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors resize-y"
          />
          <p class="mt-1 font-mono text-xs text-text-muted">Sent to the agent as the first user message on every run.</p>
        </div>

        <!-- Schedule mode toggle -->
        <div>
          <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Schedule</label>
          <div class="flex gap-1 mb-3">
            <button
              type="button"
              @click="mode = 'preset'"
              class="px-3 py-1.5 border rounded font-mono text-xs uppercase tracking-wider transition-colors cursor-pointer"
              :class="mode === 'preset'
                ? 'border-accent bg-accent-subtle text-text-primary'
                : 'border-border text-text-secondary hover:border-accent/50'"
            >
              Preset
            </button>
            <button
              type="button"
              @click="mode = 'custom'"
              class="px-3 py-1.5 border rounded font-mono text-xs uppercase tracking-wider transition-colors cursor-pointer"
              :class="mode === 'custom'
                ? 'border-accent bg-accent-subtle text-text-primary'
                : 'border-border text-text-secondary hover:border-accent/50'"
            >
              Custom cron
            </button>
            <button
              type="button"
              @click="mode = 'interval'"
              class="px-3 py-1.5 border rounded font-mono text-xs uppercase tracking-wider transition-colors cursor-pointer"
              :class="mode === 'interval'
                ? 'border-accent bg-accent-subtle text-text-primary'
                : 'border-border text-text-secondary hover:border-accent/50'"
            >
              Interval
            </button>
          </div>

          <!-- Preset mode -->
          <div v-if="mode === 'preset'" class="space-y-1.5">
            <button
              v-for="preset in cronPresets"
              :key="preset.value"
              type="button"
              @click="presetValue = preset.value"
              class="w-full flex items-center justify-between px-3 py-2 border rounded font-mono text-xs transition-colors cursor-pointer"
              :class="presetValue === preset.value
                ? 'border-accent bg-accent-subtle text-text-primary'
                : 'border-border text-text-secondary hover:border-accent/50'"
            >
              <span>{{ preset.label }}</span>
              <span class="text-text-muted">{{ preset.value }}</span>
            </button>
          </div>

          <!-- Custom cron mode -->
          <div v-else-if="mode === 'custom'">
            <input
              v-model="customCron"
              type="text"
              maxlength="200"
              placeholder="*/10 * * * *"
              class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors"
            />
            <p class="mt-1 font-mono text-xs text-text-muted">
              Five-field cron expression (minute hour day month weekday). The backend validates this on submit.
            </p>
          </div>

          <!-- Interval mode -->
          <div v-else>
            <input
              v-model.number="intervalSecs"
              type="number"
              min="60"
              max="86400"
              class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors"
            />
            <p class="mt-1 font-mono text-xs text-text-muted">
              Interval in seconds. Minimum 60, maximum 86400 (1 day).
            </p>
          </div>
        </div>

        <!-- Enabled toggle -->
        <div class="flex items-center gap-3">
          <input
            id="schedule-enabled"
            v-model="enabled"
            type="checkbox"
            class="h-4 w-4 accent-accent cursor-pointer"
          />
          <label for="schedule-enabled" class="font-mono text-xs uppercase tracking-wider text-text-secondary cursor-pointer">
            Enabled
          </label>
        </div>

        <!-- Footer -->
        <div class="flex gap-2 pt-2 border-t border-border -mx-6 -mb-5 px-6 py-3">
          <button
            type="submit"
            :disabled="!canSubmit"
            class="px-5 py-2 bg-action text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-action-hover transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ submitting ? 'Saving…' : schedule ? 'Save changes' : 'Create schedule' }}
          </button>
          <button
            type="button"
            @click="emit('close')"
            :disabled="submitting"
            class="px-5 py-2 border border-border text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
