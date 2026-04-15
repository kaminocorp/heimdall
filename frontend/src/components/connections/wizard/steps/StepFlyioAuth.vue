<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import type { WizardState } from '../flows'

const props = defineProps<{ modelValue: WizardState }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

const appName = ref((props.modelValue.config.app_name as string) ?? '')
const apiToken = ref((props.modelValue.config.api_token as string) ?? '')

watch(() => props.modelValue.config, (cfg) => {
  const newApp = (cfg.app_name as string) ?? ''
  const newToken = (cfg.api_token as string) ?? ''
  if (newApp !== appName.value) appName.value = newApp
  if (newToken !== apiToken.value) apiToken.value = newToken
}, { deep: true })

function sync() {
  emit('update:modelValue', {
    ...props.modelValue,
    config: { ...props.modelValue.config, app_name: appName.value, api_token: apiToken.value },
  })
  emit('valid', appName.value.trim().length > 0 && apiToken.value.trim().length > 0)
}

watch([appName, apiToken], sync)
onMounted(sync)
</script>

<template>
  <div class="space-y-4">
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">
        Fly App Name
      </label>
      <input
        v-model="appName"
        type="text"
        placeholder="e.g. my-app-prod"
        autofocus
        class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors"
      />
      <p class="mt-1 font-mono text-[11px] text-text-tertiary">
        The name shown in your Fly.io dashboard and <code class="text-text-secondary">fly.toml</code>.
      </p>
    </div>
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">
        API Token
      </label>
      <input
        v-model="apiToken"
        type="password"
        placeholder="fo1_..."
        class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors"
      />
      <p class="mt-1 font-mono text-[11px] text-text-tertiary">
        Generate a token at
        <span class="text-text-secondary">fly.io/dashboard/personal/access-tokens</span>
        or run <code class="text-text-secondary">fly tokens create</code>.
      </p>
    </div>
  </div>
</template>
