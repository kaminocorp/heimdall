<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, useTemplateRef, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import type { ApplicationWithCounts } from '@/types/organization'

// ———————————————————————————————————————————————————————————————————————
// DeleteAppModal — two-stage destructive action modal.
//
// Stage 1 ("confirm"): typed-to-confirm pattern. The user must type the
//   exact app name before the Delete button enables. We show an itemised
//   list of everything that will cascade-delete (connections, schedules,
//   logs, reports, conversation history) so the blast radius is explicit.
//
// Stage 2 ("deleted"): post-delete acknowledgement. After the API call
//   resolves, we transition to a success state with a single "Close"
//   button. The user sees a second, calm confirmation that the destructive
//   action actually landed. The parent handles list refresh on close.
//
// Error handling: if the backend returns 409 with code: "last_app" (which
//   should be impossible because the parent disables the button in that
//   case, but we defend anyway), the error text is surfaced inline and
//   the stage stays at "confirm" so the user can cancel out without
//   surprise. Any other error renders similarly — no stage transition.
// ———————————————————————————————————————————————————————————————————————

const props = defineProps<{
  application: ApplicationWithCounts
}>()

const emit = defineEmits<{
  close: []
  // Emitted after the user acknowledges the post-delete success state.
  // The parent uses this to refetch the list — we don't emit it on
  // inline errors, because nothing changed in that case.
  deleted: []
}>()

const app = useAppStore()

type Stage = 'confirm' | 'deleted'
const stage = ref<Stage>('confirm')

const typedName = ref('')
const deleting = ref(false)
const error = ref<string | null>(null)

// The typed-to-confirm check is a case-sensitive exact-match. We explicitly
// don't lowercase or trim because the match is meant to force the user to
// *read* the app name before deleting it — slight leniency defeats the
// purpose of the safeguard.
const canDelete = computed(() => typedName.value === props.application.name)

// Auto-focus the input when the modal mounts so keyboard-first users can
// start typing immediately.
const inputRef = useTemplateRef<HTMLInputElement>('nameInput')
onMounted(async () => {
  await nextTick()
  inputRef.value?.focus()
})

// If the application prop changes while the modal is open (very unlikely —
// it implies the parent swapped to a different row mid-flight), reset the
// input so the user isn't confused by a pre-filled value that no longer
// matches.
watch(() => props.application.id, () => {
  typedName.value = ''
  error.value = null
})

async function handleDelete() {
  if (!canDelete.value || deleting.value) return
  deleting.value = true
  error.value = null
  try {
    await app.deleteApp(props.application.id)
    stage.value = 'deleted'
  } catch (e: unknown) {
    // Extract backend error text. The Phase 1 last-app guard returns
    // { error: "cannot delete last application", code: "last_app" } — we
    // prefer the server's message when present so users see the real
    // reason instead of a generic fallback.
    const err = e as { response?: { data?: { error?: string; code?: string } } }
    error.value = err.response?.data?.error ?? 'Failed to delete application. Please try again.'
  } finally {
    deleting.value = false
  }
}

function handleClose() {
  if (deleting.value) return // Don't allow closing mid-request.
  if (stage.value === 'deleted') {
    emit('deleted')
  }
  emit('close')
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') handleClose()
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => document.removeEventListener('keydown', onKeydown))
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
    @click.self="handleClose"
  >
    <div class="relative w-full max-w-md mx-4 border border-border rounded-lg bg-bg-surface shadow-2xl">
      <!-- Stage 1: confirmation -->
      <template v-if="stage === 'confirm'">
        <header class="flex items-center justify-between px-6 py-4 border-b border-border">
          <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-status-critical">
            Delete Application
          </h3>
          <button
            type="button"
            @click="handleClose"
            :disabled="deleting"
            class="font-mono text-xs text-text-muted hover:text-text-primary transition-colors cursor-pointer disabled:cursor-not-allowed"
          >&times;</button>
        </header>

        <div class="px-6 py-5">
          <p class="font-mono text-sm text-text-secondary mb-4">
            This will permanently delete
            <span class="text-text-primary font-medium">{{ application.name }}</span>
            and all of its data:
          </p>

          <ul class="mb-5 space-y-1 font-mono text-xs text-text-secondary pl-4">
            <li class="flex gap-2">
              <span class="text-text-muted">&bull;</span>
              <span>
                <span class="text-text-primary font-medium">{{ application.connection_count }}</span>
                {{ application.connection_count === 1 ? 'connection' : 'connections' }}
              </span>
            </li>
            <li class="flex gap-2">
              <span class="text-text-muted">&bull;</span>
              <span>
                <span class="text-text-primary font-medium">{{ application.schedule_count }}</span>
                {{ application.schedule_count === 1 ? 'schedule' : 'schedules' }}
              </span>
            </li>
            <li class="flex gap-2">
              <span class="text-text-muted">&bull;</span>
              <span class="text-text-primary">all logs</span>
            </li>
            <li class="flex gap-2">
              <span class="text-text-muted">&bull;</span>
              <span class="text-text-primary">all reports</span>
            </li>
            <li class="flex gap-2">
              <span class="text-text-muted">&bull;</span>
              <span class="text-text-primary">all conversation history</span>
            </li>
          </ul>

          <p class="font-mono text-xs text-status-critical mb-4">
            This cannot be undone.
          </p>

          <label class="block">
            <span class="block font-mono text-xs uppercase tracking-wider text-text-muted mb-1.5">
              Type <span class="text-text-primary font-medium">{{ application.name }}</span> to confirm
            </span>
            <input
              ref="nameInput"
              v-model="typedName"
              type="text"
              :disabled="deleting"
              :placeholder="application.name"
              @keydown.enter.prevent="handleDelete"
              class="w-full px-3 py-2 bg-bg-elevated/80 border border-border rounded font-mono text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-status-critical/50 focus:ring-1 focus:ring-status-critical/20 transition-colors disabled:opacity-60"
            />
          </label>

          <!-- Inline error banner. Used for both the last-app guard and
               any other backend/network failure. Stays in-place so users
               don't lose the modal context. -->
          <div
            v-if="error"
            class="mt-4 rounded border border-status-critical/30 bg-status-critical/10 px-3 py-2"
          >
            <p class="font-mono text-xs text-status-critical">{{ error }}</p>
          </div>
        </div>

        <footer class="flex items-center justify-end gap-2 px-6 py-3 border-t border-border">
          <button
            type="button"
            @click="handleClose"
            :disabled="deleting"
            class="px-5 py-2 font-mono text-xs uppercase tracking-wider rounded border border-border text-text-secondary hover:border-border-hover hover:text-text-primary transition-colors cursor-pointer disabled:cursor-not-allowed"
          >
            Cancel
          </button>
          <button
            type="button"
            @click="handleDelete"
            :disabled="!canDelete || deleting"
            class="px-5 py-2 font-mono text-xs font-medium uppercase tracking-wider rounded transition-colors cursor-pointer"
            :class="(!canDelete || deleting)
              ? 'bg-status-critical/30 text-bg-primary cursor-not-allowed'
              : 'bg-status-critical text-bg-primary hover:bg-status-critical/80'"
          >
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </button>
        </footer>
      </template>

      <!-- Stage 2: post-delete confirmation -->
      <template v-else>
        <div class="px-6 py-8 text-center">
          <div class="mx-auto w-12 h-12 rounded-full bg-accent-subtle flex items-center justify-center mb-4">
            <svg class="w-6 h-6 text-accent" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20 6L9 17l-5-5" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </div>
          <h4 class="font-mono text-sm font-medium text-text-primary mb-1">
            Application deleted
          </h4>
          <p class="font-mono text-xs text-text-secondary">
            <span class="text-text-primary font-medium">{{ application.name }}</span> has been removed.
          </p>
        </div>
        <footer class="flex items-center justify-center px-6 py-3 border-t border-border">
          <button
            type="button"
            @click="handleClose"
            class="px-5 py-2 font-mono text-xs font-medium uppercase tracking-wider rounded bg-action text-bg-primary hover:bg-action-hover transition-colors cursor-pointer"
          >
            Close
          </button>
        </footer>
      </template>
    </div>
  </div>
</template>
