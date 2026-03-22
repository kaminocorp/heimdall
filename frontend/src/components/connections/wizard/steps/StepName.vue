<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import type { WizardState } from '../flows'

const props = defineProps<{ modelValue: WizardState }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

const name = ref(props.modelValue.name)

// Sync inward when parent resets state (e.g. platform re-selection).
watch(() => props.modelValue.name, (v) => {
  if (v !== name.value) name.value = v
})

watch(name, (v) => {
  emit('update:modelValue', { ...props.modelValue, name: v })
  emit('valid', v.trim().length > 0)
})

onMounted(() => {
  emit('valid', name.value.trim().length > 0)
})
</script>

<template>
  <div class="space-y-4">
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">
        Connection Name
      </label>
      <input
        v-model="name"
        type="text"
        placeholder="e.g. Production Supabase"
        autofocus
        class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
      />
    </div>
    <p class="font-mono text-xs text-text-tertiary">
      A descriptive name to identify this connection in your dashboard.
    </p>
  </div>
</template>
