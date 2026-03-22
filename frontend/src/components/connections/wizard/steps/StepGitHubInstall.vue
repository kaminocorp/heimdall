<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { getGitHubInstallURL } from '@/api/github'
import type { WizardState } from '../flows'

defineProps<{ modelValue: WizardState }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

const appStore = useAppStore()
const installing = ref(false)
const error = ref<string | null>(null)

// GitHub install redirects the browser — no "next" step in the wizard.
// The wizard's Done button is hidden; the install button replaces it.
onMounted(() => {
  emit('valid', false) // Never auto-advance — user must click install
})

async function install() {
  if (!appStore.currentAppId) return
  installing.value = true
  error.value = null
  try {
    const { url } = await getGitHubInstallURL(appStore.currentAppId)
    window.location.href = url
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    error.value = err.response?.data?.error ?? 'Failed to start GitHub App installation'
    installing.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
    <p class="font-mono text-xs text-text-secondary leading-relaxed">
      Heimdall uses a GitHub App for secure, org-scoped access to your repositories.
      Click below to install the app on your GitHub organization.
    </p>

    <div v-if="error" class="rounded border border-status-critical/30 bg-status-critical/10 px-4 py-2">
      <p class="text-sm font-mono text-status-critical">{{ error }}</p>
    </div>

    <button
      type="button"
      :disabled="installing"
      @click="install"
      class="px-4 py-2 bg-action text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-action-hover transition-colors cursor-pointer disabled:opacity-50"
    >
      {{ installing ? 'Redirecting...' : 'Install GitHub App' }}
    </button>

    <p class="font-mono text-xs text-text-tertiary">
      You will be redirected to GitHub. After installation, you'll return to Heimdall to select repositories.
    </p>
  </div>
</template>
