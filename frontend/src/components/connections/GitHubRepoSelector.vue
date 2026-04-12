<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { GitHubRepo } from '@/types/github'
import { listGitHubRepos, updateGitHubRepos } from '@/api/github'
import { extractApiError } from '@/utils/apiError'

const props = defineProps<{
  connectionId: string
}>()

const emit = defineEmits<{
  close: []
}>()

const repos = ref<GitHubRepo[]>([])
const loading = ref(true)
const saving = ref(false)
const error = ref<string | null>(null)

onMounted(async () => {
  try {
    repos.value = await listGitHubRepos(props.connectionId)
  } catch (e: unknown) {
    error.value = extractApiError(e, 'Failed to load repositories')
  } finally {
    loading.value = false
  }
})

function toggleRepo(repo: GitHubRepo) {
  repo.enabled = !repo.enabled
}

async function save() {
  saving.value = true
  error.value = null
  try {
    await updateGitHubRepos(props.connectionId, repos.value)
    emit('close')
  } catch (e: unknown) {
    error.value = extractApiError(e, 'Failed to save repository selection')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="border border-border rounded-lg p-6 bg-bg-surface space-y-4">
    <div class="flex items-center justify-between">
      <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">Manage Repositories</h3>
      <button @click="emit('close')" class="text-text-muted hover:text-text-primary transition-colors text-sm font-mono cursor-pointer">Close</button>
    </div>

    <p class="font-mono text-xs text-text-secondary">
      Select which repositories Heimdall can search during investigations.
    </p>

    <div v-if="error" class="rounded border border-status-critical/30 bg-status-critical/10 px-3 py-2 text-sm font-mono text-status-critical">
      {{ error }}
    </div>

    <div v-if="loading" class="text-sm font-mono text-text-muted py-4">Loading repositories...</div>

    <div v-else-if="repos.length === 0" class="text-sm font-mono text-text-muted py-4">
      No repositories found. The GitHub App may not have access to any repos.
    </div>

    <div v-else class="space-y-1 max-h-80 overflow-y-auto">
      <div
        v-for="repo in repos"
        :key="repo.repo_id"
        class="flex items-center justify-between px-3 py-2 rounded hover:bg-bg-elevated/50 transition-colors cursor-pointer"
        @click="toggleRepo(repo)"
      >
        <div class="flex items-center gap-3 min-w-0">
          <div
            class="w-4 h-4 rounded border flex items-center justify-center flex-shrink-0 transition-colors"
            :class="repo.enabled ? 'bg-accent border-accent' : 'border-border'"
          >
            <svg v-if="repo.enabled" class="w-3 h-3 text-bg-primary" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M2 6l3 3 5-5" />
            </svg>
          </div>
          <span class="font-mono text-sm text-text-primary truncate">{{ repo.repo_full_name }}</span>
        </div>
        <span class="font-mono text-xs text-text-muted flex-shrink-0 ml-2">{{ repo.default_branch }}</span>
      </div>
    </div>

    <div class="flex gap-2 pt-2 border-t border-border">
      <button
        :disabled="saving"
        @click="save"
        class="px-5 py-2.5 bg-action text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-action-hover transition-colors cursor-pointer disabled:opacity-50"
      >
        {{ saving ? 'Saving...' : 'Save Selection' }}
      </button>
      <button
        @click="emit('close')"
        class="px-5 py-2.5 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
      >
        Cancel
      </button>
    </div>
  </div>
</template>
