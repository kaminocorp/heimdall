<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import type { WizardState } from '../flows'

const props = defineProps<{ modelValue: WizardState }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

const projectRef = ref((props.modelValue.config.project_ref as string) ?? '')
const accessToken = ref((props.modelValue.config.access_token as string) ?? '')

// Sync inward when parent resets state (e.g. platform re-selection).
watch(() => props.modelValue.config, (cfg) => {
  const newRef = (cfg.project_ref as string) ?? ''
  const newToken = (cfg.access_token as string) ?? ''
  if (newRef !== projectRef.value) projectRef.value = newRef
  if (newToken !== accessToken.value) accessToken.value = newToken
}, { deep: true })

function sync() {
  emit('update:modelValue', {
    ...props.modelValue,
    config: { ...props.modelValue.config, project_ref: projectRef.value, access_token: accessToken.value },
  })
  emit('valid', projectRef.value.trim().length > 0 && accessToken.value.trim().length > 0)
}

watch([projectRef, accessToken], sync)
onMounted(sync)
</script>

<template>
  <div class="space-y-4">
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">
        Project Reference
      </label>
      <input
        v-model="projectRef"
        type="text"
        placeholder="e.g. abcdefghijklmnopqrst"
        autofocus
        class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
      />
    </div>
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">
        Personal Access Token
      </label>
      <input
        v-model="accessToken"
        type="password"
        placeholder="sbp_..."
        class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
      />
    </div>
    <p class="font-mono text-xs text-text-tertiary leading-relaxed">
      Uses the Supabase Management API — works on all plans including Free.
      Generate a Personal Access Token at
      <span class="text-text-secondary">supabase.com/dashboard/account/tokens</span>.
    </p>
  </div>
</template>
