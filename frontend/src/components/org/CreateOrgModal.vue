<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { createOrganization } from '@/api/organizations'
import { extractApiError } from '@/utils/apiError'

const emit = defineEmits<{
  close: []
}>()

const appStore = useAppStore()

const name = ref('')
const slug = ref('')
const autoSlug = ref(true)
const creating = ref(false)
const error = ref<string | null>(null)

// Auto-generate slug from name while autoSlug is true.
watch(name, (val) => {
  if (autoSlug.value) {
    slug.value = val
      .toLowerCase()
      .replace(/[^a-z0-9\s-]/g, '')
      .replace(/\s+/g, '-')
      .replace(/-+/g, '-')
      .slice(0, 48)
  }
})

function onSlugInput() {
  autoSlug.value = false
}

const canSubmit = computed(() => name.value.trim() && slug.value.trim() && !creating.value)

async function handleSubmit() {
  if (!canSubmit.value) return
  creating.value = true
  error.value = null
  try {
    const org = await createOrganization(name.value.trim(), slug.value.trim())
    await appStore.addOrg(org)
    emit('close')
  } catch (e: unknown) {
    error.value = extractApiError(e, 'Failed to create organisation')
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <!-- Backdrop -->
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="emit('close')" />

      <!-- Dialog -->
      <div class="relative z-10 w-full max-w-md mx-4 rounded-lg border border-border bg-bg-elevated shadow-xl shadow-black/40 p-6">
        <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary mb-4">
          New organisation
        </h3>

        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div>
            <label class="block font-mono text-xs text-text-muted uppercase tracking-wider mb-1.5">
              Organisation name
            </label>
            <input
              v-model="name"
              type="text"
              placeholder="Acme Corp"
              required
              autofocus
              class="w-full px-3 py-2 rounded border border-border bg-bg-primary font-mono text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent/50 transition-colors"
            />
          </div>

          <div>
            <label class="block font-mono text-xs text-text-muted uppercase tracking-wider mb-1.5">
              Slug
            </label>
            <input
              v-model="slug"
              @input="onSlugInput"
              type="text"
              placeholder="acme-corp"
              required
              class="w-full px-3 py-2 rounded border border-border bg-bg-primary font-mono text-sm text-text-secondary placeholder:text-text-muted focus:outline-none focus:border-accent/50 transition-colors"
            />
            <p class="mt-1 font-mono text-xs text-text-muted">
              Used in URLs. Lowercase, hyphens only.
            </p>
          </div>

          <div v-if="error" class="rounded border border-status-critical/30 bg-status-critical/10 px-3 py-2">
            <p class="font-mono text-xs text-status-critical">{{ error }}</p>
          </div>

          <div class="flex items-center justify-end gap-3 pt-2">
            <button
              type="button"
              @click="emit('close')"
              class="px-4 py-2 font-mono text-xs uppercase tracking-wider rounded border border-border text-text-secondary hover:text-text-primary hover:border-border-hover transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="!canSubmit"
              class="px-4 py-2 font-mono text-xs font-medium uppercase tracking-wider rounded bg-action text-bg-primary hover:bg-action-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
              {{ creating ? 'Creating...' : 'Create organisation' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>
