<script setup lang="ts">
import { ref } from 'vue'
import { useToast } from '@/composables/useToast'
import { updateNotificationPreferences } from '@/api/notifications'
import type { NotificationPreferences } from '@/types/notification'
import BaseSelect from '@/components/common/BaseSelect.vue'

const props = defineProps<{
  appId: string
  preferences: NotificationPreferences | null
}>()

const emit = defineEmits<{
  'update:preferences': [prefs: NotificationPreferences]
}>()

const toast = useToast()

const editing = ref(false)
const saving = ref(false)
const formEnabled = ref(false)
const formThreshold = ref<string>('warning')
const formCooldown = ref(15)

const thresholdOptions = [
  { label: 'Info+', value: 'info' },
  { label: 'Warning+', value: 'warning' },
  { label: 'Error+', value: 'error' },
  { label: 'Critical only', value: 'critical' },
]

function startEdit() {
  if (!props.preferences) return
  formEnabled.value = props.preferences.enabled
  formThreshold.value = props.preferences.severity_threshold
  formCooldown.value = props.preferences.cooldown_minutes
  editing.value = true
}

async function save() {
  saving.value = true
  try {
    const updated = await updateNotificationPreferences(props.appId, {
      enabled: formEnabled.value,
      severity_threshold: formThreshold.value,
      cooldown_minutes: formCooldown.value,
    })
    emit('update:preferences', updated)
    editing.value = false
    toast.show('Preferences saved', 'success')
  } catch {
    toast.show('Failed to save preferences', 'error')
  } finally {
    saving.value = false
  }
}

function cancel() {
  editing.value = false
}

defineExpose({ resetEditing: () => { editing.value = false } })
</script>

<template>
  <section>
    <div class="flex items-center justify-between mb-3">
      <h3 class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Preferences</h3>
      <button
        v-if="preferences && !editing"
        @click="startEdit"
        class="px-3 py-1 border border-accent-border/50 text-text-secondary font-mono text-xs uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
      >
        Edit
      </button>
    </div>

    <!-- Edit mode -->
    <form v-if="editing" @submit.prevent="save" class="border border-border rounded-lg bg-bg-surface p-6 space-y-6">
      <div class="flex items-center gap-3">
        <label class="font-mono text-xs font-medium uppercase tracking-wider text-text-secondary">Enabled</label>
        <button
          type="button"
          @click="formEnabled = !formEnabled"
          class="relative w-10 h-5 rounded-full transition-colors cursor-pointer"
          :class="formEnabled ? 'bg-accent' : 'bg-bg-elevated border border-border'"
        >
          <span
            class="absolute top-0.5 w-4 h-4 rounded-full bg-text-primary transition-transform"
            :class="formEnabled ? 'translate-x-5' : 'translate-x-0.5'"
          />
        </button>
      </div>

      <div>
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Severity Threshold</label>
        <BaseSelect v-model="formThreshold" :options="thresholdOptions" />
      </div>

      <div>
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Cooldown (minutes)</label>
        <input
          v-model.number="formCooldown"
          type="number"
          min="1"
          max="1440"
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors"
        />
        <p class="mt-1 font-mono text-xs text-text-muted">Suppress duplicate notifications for this many minutes (1–1440)</p>
      </div>

      <div class="flex gap-2 pt-2">
        <button
          type="submit"
          :disabled="saving"
          class="px-5 py-2.5 bg-action text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-action-hover transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ saving ? 'Saving...' : 'Save' }}
        </button>
        <button
          type="button"
          @click="cancel"
          :disabled="saving"
          class="px-5 py-2.5 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
        >
          Cancel
        </button>
      </div>
    </form>

    <!-- Display mode -->
    <div v-else-if="preferences" class="border border-border rounded-lg bg-bg-surface p-6">
      <div class="space-y-3">
        <div class="flex items-baseline justify-between py-1.5 border-b border-border">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Status</span>
          <span class="font-mono text-sm" :class="preferences.enabled ? 'text-accent' : 'text-text-muted'">
            {{ preferences.enabled ? 'Enabled' : 'Disabled' }}
          </span>
        </div>
        <div class="flex items-baseline justify-between py-1.5 border-b border-border">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Threshold</span>
          <span class="font-mono text-sm text-text-primary capitalize">{{ preferences.severity_threshold }}+</span>
        </div>
        <div class="flex items-baseline justify-between py-1.5">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Cooldown</span>
          <span class="font-mono text-sm text-text-primary">{{ preferences.cooldown_minutes }} min</span>
        </div>
      </div>
    </div>
  </section>
</template>
