<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { useToast } from '@/composables/useToast'
import { getAppAgentConfig, updateAppAgentConfig } from '@/api/applications'
import type { AppAgentConfig } from '@/types/organization'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'

const appStore = useAppStore()
const toast = useToast()

const config = ref<AppAgentConfig | null>(null)
const loading = ref(false)
const editing = ref(false)
const saving = ref(false)

// Form state
const formModel = ref('')
const formMode = ref<'continuous' | 'periodic' | 'off'>('off')
const formInterval = ref(60)
const formPrompt = ref('')

const knownModels = [
  'claude-sonnet-4-6',
  'claude-haiku-4-5-20251001',
  'claude-opus-4-6',
]

const intervalPresets = [
  { label: '30s', value: 30 },
  { label: '1m', value: 60 },
  { label: '5m', value: 300 },
  { label: '15m', value: 900 },
]

async function fetchConfig() {
  const appId = appStore.currentAppId
  if (!appId) return
  loading.value = true
  try {
    config.value = await getAppAgentConfig(appId)
  } catch {
    toast.show('Failed to load agent config', 'error')
  } finally {
    loading.value = false
  }
}

function startEdit() {
  if (!config.value) return
  formModel.value = config.value.model
  formMode.value = config.value.mode
  formInterval.value = config.value.schedule_interval_secs
  formPrompt.value = config.value.system_prompt_override ?? ''
  editing.value = true
}

function cancelEdit() {
  editing.value = false
}

async function saveConfig() {
  const appId = appStore.currentAppId
  if (!appId) return
  saving.value = true
  try {
    config.value = await updateAppAgentConfig(appId, {
      model: formModel.value,
      mode: formMode.value,
      schedule_interval_secs: formInterval.value,
      system_prompt_override: formPrompt.value || null,
    })
    editing.value = false
    toast.show('Configuration saved', 'success')
  } catch {
    toast.show('Failed to save configuration', 'error')
  } finally {
    saving.value = false
  }
}

function formatInterval(secs: number): string {
  if (secs < 60) return `${secs} seconds`
  if (secs < 3600) return `${Math.round(secs / 60)} minutes`
  return `${Math.round(secs / 3600)} hours`
}

onMounted(fetchConfig)
watch(() => appStore.currentAppId, () => {
  editing.value = false
  fetchConfig()
})
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="pb-6 mb-8 border-b border-border flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">Agent Configuration</h2>
        <p class="font-sans text-sm text-text-secondary mt-1">
          Per-application model settings and monitoring behavior
          <span v-if="appStore.currentApp" class="text-text-muted">— {{ appStore.currentApp.name }}</span>
        </p>
      </div>
      <button
        v-if="config && !editing"
        @click="startEdit"
        class="px-4 py-2 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
      >
        Edit Configuration
      </button>
    </div>

    <!-- Skeleton loader -->
    <div v-if="loading" class="border border-border rounded-lg bg-bg-surface p-5 space-y-4">
      <div v-for="n in 4" :key="n" class="flex items-baseline justify-between py-2">
        <SkeletonBlock width="5rem" height="0.75rem" />
        <SkeletonBlock width="10rem" height="0.75rem" />
      </div>
    </div>

    <!-- Edit mode -->
    <form v-else-if="editing" @submit.prevent="saveConfig" class="border border-border rounded-lg bg-bg-surface p-5 space-y-5">
      <div>
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Model</label>
        <input
          v-model="formModel"
          type="text"
          list="known-models"
          required
          placeholder="claude-sonnet-4-6"
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
        />
        <datalist id="known-models">
          <option v-for="m in knownModels" :key="m" :value="m" />
        </datalist>
      </div>

      <div>
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Monitoring Mode</label>
        <BaseSelect
          v-model="formMode"
          :options="[
            { value: 'continuous', label: 'Continuous' },
            { value: 'periodic', label: 'Periodic' },
            { value: 'off', label: 'Off' },
          ]"
        />
      </div>

      <div v-if="formMode === 'periodic'">
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Check Interval</label>
        <div class="flex gap-2 mb-2">
          <button
            v-for="preset in intervalPresets"
            :key="preset.value"
            type="button"
            @click="formInterval = preset.value"
            class="px-3 py-1 border rounded font-mono text-xs uppercase tracking-wider transition-colors cursor-pointer"
            :class="formInterval === preset.value
              ? 'border-accent bg-accent-subtle text-text-primary'
              : 'border-border text-text-secondary hover:border-accent/50'
            "
          >
            {{ preset.label }}
          </button>
        </div>
        <input
          v-model.number="formInterval"
          type="number"
          min="10"
          max="86400"
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
        />
        <p class="mt-1 font-mono text-[10px] text-text-muted">Interval in seconds (min 10, max 86400)</p>
      </div>

      <div>
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">System Prompt Override</label>
        <textarea
          v-model="formPrompt"
          rows="6"
          placeholder="Custom system prompt for the agent (leave empty for default)"
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors resize-y"
        />
      </div>

      <div class="flex gap-2 pt-2">
        <button
          type="submit"
          :disabled="saving"
          class="px-4 py-2 bg-action text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-action-hover transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ saving ? 'Saving...' : 'Save Configuration' }}
        </button>
        <button
          type="button"
          @click="cancelEdit"
          :disabled="saving"
          class="px-4 py-2 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
        >
          Cancel
        </button>
      </div>
    </form>

    <!-- Display mode -->
    <div v-else-if="config" class="border border-border rounded-lg bg-bg-surface p-5">
      <div class="space-y-4">
        <div class="flex items-baseline justify-between py-2 border-b border-border">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Model</span>
          <span class="font-mono text-sm text-text-primary">{{ config.model }}</span>
        </div>
        <div class="flex items-baseline justify-between py-2 border-b border-border">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Monitoring Mode</span>
          <span class="font-mono text-sm text-text-primary capitalize">{{ config.mode }}</span>
        </div>
        <div v-if="config.mode === 'periodic'" class="flex items-baseline justify-between py-2 border-b border-border">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Check Interval</span>
          <span class="font-mono text-sm text-text-primary">{{ formatInterval(config.schedule_interval_secs) }}</span>
        </div>
        <div class="flex items-baseline justify-between py-2">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">System Prompt</span>
          <span class="font-mono text-sm text-text-primary">{{ config.system_prompt_override ? 'Custom' : 'Default' }}</span>
        </div>
        <div v-if="config.system_prompt_override" class="pt-2 border-t border-border">
          <pre class="font-mono text-xs text-text-secondary whitespace-pre-wrap leading-relaxed">{{ config.system_prompt_override }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>
