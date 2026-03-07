<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useToast } from '@/composables/useToast'

const app = useAppStore()
const router = useRouter()
const toast = useToast()

const orgName = ref('')
const orgSlug = ref('')
const appName = ref('My Application')
const submitting = ref(false)

function generateSlug() {
  orgSlug.value = orgName.value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
}

async function handleSubmit() {
  if (!orgName.value || !orgSlug.value) return
  submitting.value = true
  try {
    await app.onboard(orgName.value, orgSlug.value, appName.value || 'My Application')
    router.push({ name: 'dashboard' })
  } catch (e: any) {
    toast.show(e.response?.data?.error ?? 'Failed to create organization', 'error')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-bg-primary flex items-center justify-center p-6">
    <div class="w-full max-w-md">
      <div class="mb-8 text-center">
        <div class="flex items-center justify-center gap-2.5 mb-4">
          <div class="w-2 h-2 rounded-full bg-accent animate-pulse" />
          <span class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">
            Heimdall
          </span>
        </div>
        <h1 class="font-mono text-lg font-bold text-text-primary">Set Up Your Organization</h1>
        <p class="mt-2 text-sm text-text-secondary">
          Create your organization and first application to start monitoring.
        </p>
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-5">
        <div>
          <label class="block font-mono text-xs uppercase tracking-wider text-text-muted mb-1.5">
            Organization Name
          </label>
          <input
            v-model="orgName"
            @blur="!orgSlug && generateSlug()"
            type="text"
            required
            placeholder="Acme Corp"
            class="w-full px-3 py-2 bg-bg-surface border border-border rounded text-sm text-text-primary placeholder-text-muted focus:outline-none focus:border-accent font-mono"
          />
        </div>

        <div>
          <label class="block font-mono text-xs uppercase tracking-wider text-text-muted mb-1.5">
            Slug
          </label>
          <input
            v-model="orgSlug"
            type="text"
            required
            placeholder="acme-corp"
            pattern="[a-z0-9\-]+"
            class="w-full px-3 py-2 bg-bg-surface border border-border rounded text-sm text-text-primary placeholder-text-muted focus:outline-none focus:border-accent font-mono"
          />
          <p class="mt-1 text-[10px] text-text-muted font-mono">Lowercase letters, numbers, and hyphens only</p>
        </div>

        <div>
          <label class="block font-mono text-xs uppercase tracking-wider text-text-muted mb-1.5">
            First Application Name
          </label>
          <input
            v-model="appName"
            type="text"
            placeholder="My Application"
            class="w-full px-3 py-2 bg-bg-surface border border-border rounded text-sm text-text-primary placeholder-text-muted focus:outline-none focus:border-accent font-mono"
          />
          <p class="mt-1 text-[10px] text-text-muted font-mono">The production system you want to monitor</p>
        </div>

        <button
          type="submit"
          :disabled="submitting || !orgName || !orgSlug"
          class="w-full py-2.5 bg-accent text-bg-primary font-mono text-sm font-bold uppercase tracking-wider rounded hover:brightness-110 transition-all disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
        >
          {{ submitting ? 'Creating...' : 'Get Started' }}
        </button>
      </form>
    </div>
  </div>
</template>
