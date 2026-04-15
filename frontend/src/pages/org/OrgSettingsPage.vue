<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { updateOrganization, deleteOrganization } from '@/api/organizations'
import { extractApiError } from '@/utils/apiError'
import CopyableField from '@/components/common/CopyableField.vue'
import { formatDate } from '@/utils/format'

const appStore = useAppStore()
const router = useRouter()

// ── State ──
const org = computed(() => appStore.organization)
const callerRole = computed(() => org.value?.role ?? 'member')
const isOwner = computed(() => callerRole.value === 'owner')
const isAdmin = computed(() => callerRole.value === 'admin' || callerRole.value === 'owner')

// Edit form
const editName = ref('')
const editSlug = ref('')
const saving = ref(false)
const saveError = ref<string | null>(null)
const saveSuccess = ref(false)

// Delete flow
const showDeleteConfirm = ref(false)
const deleteConfirmInput = ref('')
const deleting = ref(false)
const deleteError = ref<string | null>(null)

const deleteConfirmMatch = computed(() =>
  deleteConfirmInput.value === org.value?.slug
)

// ── Init — seed form when org data arrives (handles late-loading store) ──
watch(org, (newOrg) => {
  if (newOrg) {
    editName.value = newOrg.name
    editSlug.value = newOrg.slug
  }
}, { immediate: true })

// Sanitize slug on input: lowercase, alphanumeric + hyphens only, max 48 chars.
watch(editSlug, (val) => {
  const sanitized = val
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '')
    .replace(/-+/g, '-')
    .slice(0, 48)
  if (sanitized !== val) {
    editSlug.value = sanitized
  }
})

// ── Save ──
async function handleSave() {
  saving.value = true
  saveError.value = null
  saveSuccess.value = false
  try {
    const updated = await updateOrganization({
      name: editName.value.trim(),
      slug: editSlug.value.trim(),
    })
    // Update both store refs so the header and dropdown reflect the change.
    appStore.updateOrg({ name: updated.name, slug: updated.slug })
    saveSuccess.value = true
    setTimeout(() => { saveSuccess.value = false }, 3000)
  } catch (e: unknown) {
    saveError.value = extractApiError(e, 'Failed to save')
  } finally {
    saving.value = false
  }
}

// ── Delete ──
function openDeleteConfirm() {
  showDeleteConfirm.value = true
  deleteConfirmInput.value = ''
  deleteError.value = null
}

function cancelDelete() {
  showDeleteConfirm.value = false
  deleteConfirmInput.value = ''
  deleteError.value = null
}

async function handleDelete() {
  if (!deleteConfirmMatch.value || !org.value) return
  deleting.value = true
  deleteError.value = null
  try {
    await deleteOrganization(org.value.slug)
    // Re-initialize rather than reset — if the user belongs to other orgs,
    // init() selects the first one; if none remain it sets needsOnboarding.
    await appStore.init()
    if (appStore.needsOnboarding) {
      router.push({ name: 'onboarding' })
    } else {
      router.push({ name: 'org-overview' })
    }
  } catch (e: unknown) {
    deleteError.value = extractApiError(e, 'Failed to delete')
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="pb-6 mb-8 border-b border-border">
      <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">
        Organisation Settings
      </h2>
      <p class="font-sans text-sm text-text-secondary mt-1">
        General settings for {{ org?.name ?? 'your organization' }}.
      </p>
    </div>

    <div class="space-y-8 max-w-2xl">
      <!-- General section -->
      <section class="rounded-lg border border-border bg-bg-surface">
        <header class="px-6 py-4 border-b border-border">
          <h3 class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">
            General
          </h3>
        </header>

        <div class="px-6 py-5 space-y-5">
          <!-- Organisation ID (read-only, copyable) -->
          <div>
            <label class="block font-mono text-xs text-text-muted uppercase tracking-wider mb-1.5">
              Organisation ID
            </label>
            <CopyableField :value="org?.id ?? ''" />
          </div>

          <!-- Name (editable for admin+) -->
          <div>
            <label class="block font-mono text-xs text-text-muted uppercase tracking-wider mb-1.5">
              Name
            </label>
            <input
              v-if="isAdmin"
              v-model="editName"
              type="text"
              class="w-full px-3 py-2 rounded border border-border bg-bg-primary font-mono text-sm text-text-primary focus:outline-none focus:border-accent/50 transition-colors"
            />
            <div v-else class="font-mono text-sm text-text-primary px-3 py-2 rounded border border-border bg-bg-primary">
              {{ org?.name ?? '—' }}
            </div>
          </div>

          <!-- Slug (editable for admin+) -->
          <div>
            <label class="block font-mono text-xs text-text-muted uppercase tracking-wider mb-1.5">
              Slug
            </label>
            <input
              v-if="isAdmin"
              v-model="editSlug"
              type="text"
              class="w-full px-3 py-2 rounded border border-border bg-bg-primary font-mono text-sm text-text-secondary focus:outline-none focus:border-accent/50 transition-colors"
            />
            <div v-else class="font-mono text-sm text-text-secondary px-3 py-2 rounded border border-border bg-bg-primary">
              {{ org?.slug ?? '—' }}
            </div>
          </div>

          <!-- Created -->
          <div>
            <label class="block font-mono text-xs text-text-muted uppercase tracking-wider mb-1.5">
              Created
            </label>
            <div class="font-mono text-sm text-text-secondary">
              {{ org?.created_at ? formatDate(org.created_at) : '—' }}
            </div>
          </div>

          <!-- Save button (admin+ only) -->
          <div v-if="isAdmin" class="flex items-center gap-3 pt-2">
            <button
              @click="handleSave"
              :disabled="saving"
              class="px-5 py-2 font-mono text-xs font-medium uppercase tracking-wider rounded bg-action text-bg-primary hover:bg-action-hover transition-colors disabled:opacity-50 cursor-pointer"
            >
              {{ saving ? 'Saving...' : 'Save changes' }}
            </button>
            <span v-if="saveSuccess" class="font-mono text-xs text-status-ok">Saved</span>
            <span v-if="saveError" class="font-mono text-xs text-status-critical">{{ saveError }}</span>
          </div>
        </div>
      </section>

      <!-- Danger zone (owner only) -->
      <section v-if="isOwner" class="rounded-lg border border-status-critical/30 bg-bg-surface">
        <header class="px-6 py-4 border-b border-status-critical/20">
          <h3 class="font-mono text-xs font-medium uppercase tracking-wider text-status-critical">
            Danger zone
          </h3>
        </header>

        <div class="px-6 py-5">
          <div class="flex items-start justify-between gap-6">
            <div>
              <h4 class="font-mono text-sm font-medium text-text-primary mb-1">
                Delete this organisation
              </h4>
              <p class="font-sans text-xs text-text-secondary max-w-md">
                Permanently delete this organisation and all of its data, including all applications,
                connections, logs, agent configurations, schedules, and reports. This action cannot be undone.
              </p>
            </div>
            <button
              @click="openDeleteConfirm"
              class="shrink-0 px-4 py-2 font-mono text-xs font-medium uppercase tracking-wider rounded border border-status-critical/40 text-status-critical hover:bg-status-critical/10 transition-colors cursor-pointer"
            >
              Delete organisation
            </button>
          </div>
        </div>
      </section>
    </div>

    <!-- Delete confirmation modal -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition duration-100 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="showDeleteConfirm"
          class="fixed inset-0 z-50 flex items-center justify-center"
        >
          <!-- Backdrop -->
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="cancelDelete" />

          <!-- Dialog -->
          <div class="relative z-10 w-full max-w-md mx-4 rounded-lg border border-status-critical/30 bg-bg-elevated shadow-xl shadow-black/40 p-6">
            <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-status-critical mb-2">
              Delete organisation
            </h3>
            <p class="font-sans text-sm text-text-secondary mb-4">
              This will permanently delete
              <span class="font-mono font-medium text-text-primary">{{ org?.name }}</span>
              and all of its data. This action cannot be undone.
            </p>
            <p class="font-sans text-sm text-text-secondary mb-3">
              Type <code class="font-mono text-text-primary bg-bg-primary px-1.5 py-0.5 rounded">{{ org?.slug }}</code> to confirm.
            </p>
            <input
              v-model="deleteConfirmInput"
              type="text"
              :placeholder="org?.slug"
              class="w-full px-3 py-2 rounded border border-border bg-bg-primary font-mono text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-status-critical/50 transition-colors mb-4"
              @keyup.enter="deleteConfirmMatch && handleDelete()"
            />

            <div v-if="deleteError" class="mb-4 rounded border border-status-critical/30 bg-status-critical/10 px-3 py-2">
              <p class="font-mono text-xs text-status-critical">{{ deleteError }}</p>
            </div>

            <div class="flex items-center justify-end gap-3">
              <button
                @click="cancelDelete"
                class="px-4 py-2 font-mono text-xs uppercase tracking-wider rounded border border-border text-text-secondary hover:text-text-primary hover:border-border-hover transition-colors cursor-pointer"
              >
                Cancel
              </button>
              <button
                @click="handleDelete"
                :disabled="!deleteConfirmMatch || deleting"
                class="px-4 py-2 font-mono text-xs uppercase tracking-wider rounded bg-status-critical text-white transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                :class="deleteConfirmMatch ? 'hover:bg-status-critical/80' : ''"
              >
                {{ deleting ? 'Deleting...' : 'I understand, delete this organisation' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
