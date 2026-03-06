<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAgentStore } from '@/stores/agent'
import { useToast } from '@/composables/useToast'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'

const store = useAgentStore()
const toast = useToast()

const editing = ref(false)
const saving = ref(false)

// Form state
const formModel = ref('')
const formMode = ref<'continuous' | 'scheduled' | 'off'>('continuous')
const formSchedule = ref('')
const formPrompt = ref('')

const knownModels = [
  'claude-sonnet-4-5-20250929',
  'claude-haiku-4-5-20251001',
  'claude-opus-4-20250115',
]

function startEdit() {
  if (!store.config) return
  formModel.value = store.config.model
  formMode.value = store.config.mode
  formSchedule.value = store.config.schedule ?? ''
  formPrompt.value = store.config.system_prompt_override ?? ''
  editing.value = true
}

function cancelEdit() {
  editing.value = false
}

async function saveConfig() {
  saving.value = true
  try {
    await store.updateConfig({
      model: formModel.value,
      mode: formMode.value,
      schedule: formSchedule.value || null,
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

onMounted(() => {
  store.fetchConfig()
})
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="pb-6 mb-8 border-b border-border flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">Agent Configuration</h2>
        <p class="font-sans text-sm text-text-secondary mt-1">Model settings and agent behavior</p>
      </div>
      <button
        v-if="store.config && !editing"
        @click="startEdit"
        class="px-4 py-2 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
      >
        Edit Configuration
      </button>
    </div>

    <!-- Skeleton loader -->
    <div v-if="store.loading" class="border border-border rounded-lg bg-bg-surface p-5 space-y-4">
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
          placeholder="claude-sonnet-4-5-20250929"
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
        />
        <datalist id="known-models">
          <option v-for="m in knownModels" :key="m" :value="m" />
        </datalist>
      </div>

      <div>
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Mode</label>
        <select
          v-model="formMode"
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
        >
          <option value="continuous">Continuous</option>
          <option value="scheduled">Scheduled</option>
          <option value="off">Off</option>
        </select>
      </div>

      <div v-if="formMode === 'scheduled'">
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Schedule</label>
        <input
          v-model="formSchedule"
          type="text"
          placeholder="*/5 * * * * (cron expression)"
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
        />
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
          class="px-4 py-2 bg-accent text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-accent-hover transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
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
    <div v-else-if="store.config" class="border border-border rounded-lg bg-bg-surface p-5">
      <div class="space-y-4">
        <div class="flex items-baseline justify-between py-2 border-b border-border">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Model</span>
          <span class="font-mono text-sm text-text-primary">{{ store.config.model }}</span>
        </div>
        <div class="flex items-baseline justify-between py-2 border-b border-border">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Mode</span>
          <span class="font-mono text-sm text-text-primary capitalize">{{ store.config.mode }}</span>
        </div>
        <div class="flex items-baseline justify-between py-2 border-b border-border">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Schedule</span>
          <span class="font-mono text-sm text-text-primary">{{ store.config.schedule ?? '—' }}</span>
        </div>
        <div class="flex items-baseline justify-between py-2">
          <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">System Prompt</span>
          <span class="font-mono text-sm text-text-primary">{{ store.config.system_prompt_override ? 'Custom' : 'Default' }}</span>
        </div>
        <div v-if="store.config.system_prompt_override" class="pt-2 border-t border-border">
          <pre class="font-mono text-xs text-text-secondary whitespace-pre-wrap leading-relaxed">{{ store.config.system_prompt_override }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>
