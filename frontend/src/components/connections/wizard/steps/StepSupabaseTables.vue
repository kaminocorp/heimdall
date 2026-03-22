<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { supabaseLogTables, type WizardState } from '../flows'
import BaseSelect from '@/components/common/BaseSelect.vue'

const props = defineProps<{ modelValue: WizardState }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

const tables = supabaseLogTables

const intervalOptions = [
  { value: '15', label: '15 seconds' },
  { value: '30', label: '30 seconds' },
  { value: '60', label: '60 seconds' },
]

const selectedTables = ref<string[]>(
  (props.modelValue.config.poll_tables as string[]) ?? ['postgres_logs', 'auth_logs']
)
const pollInterval = ref(String(props.modelValue.config.poll_interval_secs ?? '30'))

// Sync inward when parent resets state (e.g. platform re-selection).
watch(() => props.modelValue.config, (cfg) => {
  const newTables = (cfg.poll_tables as string[]) ?? ['postgres_logs', 'auth_logs']
  const newInterval = String(cfg.poll_interval_secs ?? '30')
  if (JSON.stringify(newTables) !== JSON.stringify(selectedTables.value)) selectedTables.value = newTables
  if (newInterval !== pollInterval.value) pollInterval.value = newInterval
}, { deep: true })

function sync() {
  emit('update:modelValue', {
    ...props.modelValue,
    config: {
      ...props.modelValue.config,
      poll_tables: selectedTables.value,
      poll_interval_secs: Number(pollInterval.value),
    },
  })
  emit('valid', selectedTables.value.length > 0)
}

watch([selectedTables, pollInterval], sync, { deep: true })
onMounted(sync)
</script>

<template>
  <div class="space-y-5">
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-2">
        Log Tables
      </label>
      <div class="space-y-2">
        <label
          v-for="table in tables"
          :key="table.value"
          class="flex items-center gap-2.5 cursor-pointer group/check"
        >
          <input
            type="checkbox"
            :value="table.value"
            v-model="selectedTables"
            class="w-3.5 h-3.5 rounded border-border bg-bg-elevated text-accent focus:ring-accent/30 focus:ring-offset-0 cursor-pointer"
          />
          <span class="font-mono text-sm text-text-primary group-hover/check:text-accent transition-colors">
            {{ table.value }}
          </span>
          <span class="font-mono text-xs text-text-muted">&mdash; {{ table.label }}</span>
        </label>
      </div>
      <p v-if="selectedTables.length === 0" class="mt-2 font-mono text-xs text-status-critical">
        Select at least one table.
      </p>
    </div>

    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">
        Poll Interval
      </label>
      <BaseSelect
        :modelValue="pollInterval"
        @update:modelValue="pollInterval = String($event)"
        :options="intervalOptions"
      />
    </div>
  </div>
</template>
