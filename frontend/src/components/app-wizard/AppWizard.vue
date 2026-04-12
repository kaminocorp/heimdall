<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useToast } from '@/composables/useToast'
import ConnectionWizard from '@/components/connections/wizard/ConnectionWizard.vue'
import WizardStepIndicator from '@/components/connections/wizard/WizardStepIndicator.vue'
import StepAppDetails from './steps/StepAppDetails.vue'
import StepConnectorChoice from './steps/StepConnectorChoice.vue'
import StepConfirm from './steps/StepConfirm.vue'

// —————————————————————————————————————————————————————————————————————
// AppWizard — "Create a new application" flow
//
// Steps: Details (name+description) → Connector? (add/skip) → Confirm.
//
// Eager-create semantics:
//   The application is POSTed to the backend *at the end of step 1*, not at
//   the end of the wizard. Two reasons for this choice:
//     1. The optional step-2 connector needs a valid app_id FK to attach to.
//     2. It mirrors ConnectionWizard's existing "create on enter, rollback
//        on discard" pattern, so the two wizards feel consistent.
//
//   The trade-off is that closing the browser mid-wizard leaves an orphan
//   app the user has to clean up manually. This is intentional — a "draft"
//   column with a TTL sweep would be over-engineering for a symptom users
//   can trivially fix themselves via Settings → Applications.
//
// Selection semantics:
//   The draft app is NOT auto-selected when created. If the user discards,
//   their previously-active app stays selected and their route stays put.
//   The selection is only committed when the user hits "Go to dashboard"
//   on the final step. This is why `app.createApp()` is called with
//   `{ select: false }` here but `{ select: true }` everywhere else.
// —————————————————————————————————————————————————————————————————————

const emit = defineEmits<{
  close: []
}>()

const app = useAppStore()
const router = useRouter()
const toast = useToast()

// Three step IDs, matching the stepLabels below. 'connector' is a dedicated
// sub-state: when active, the AppWizard modal shell is hidden and
// ConnectionWizard takes over the screen so the user doesn't see two
// stacked backdrops.
type StepId = 'details' | 'choice' | 'connector' | 'confirm'
const stepId = ref<StepId>('details')
const stepLabels = ['Details', 'Connector', 'Confirm']
const stepIndexMap: Record<StepId, number> = {
  details: 0,
  choice: 1,
  connector: 1,
  confirm: 2,
}
const currentStepIndex = computed(() => stepIndexMap[stepId.value])

// Form state for step 1.
const details = reactive({ name: '', description: '' })
const detailsValid = ref(false)

// Draft state.
const draftAppId = ref<string | null>(null)
const connectorAdded = ref(false)
const creating = ref(false)
const error = ref<string | null>(null)

// Discard confirmation overlay.
const showDiscardConfirm = ref(false)

const hasDraft = computed(() => draftAppId.value !== null)

// Step 1 → step 2 transition: create the app eagerly. If the API call fails
// we stay on step 1 and surface the error inline.
async function advanceFromDetails() {
  if (!detailsValid.value || creating.value) return
  creating.value = true
  error.value = null
  try {
    const created = await app.createApp(details.name.trim(), { select: false })
    draftAppId.value = created.id
    stepId.value = 'choice'
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    error.value = err.response?.data?.error ?? 'Failed to create application'
  } finally {
    creating.value = false
  }
}

// Step 2 fork handlers.
function handleChoice(choice: 'add' | 'skip') {
  if (choice === 'add') {
    stepId.value = 'connector'
  } else {
    stepId.value = 'confirm'
  }
}

// ConnectionWizard callbacks. `created` fires when the user finishes the
// inner wizard with a successful connection; `close` fires when they close
// it without completing (e.g. back-out or discard inside ConnectionWizard).
// In the close-without-create case we jump back to the choice step so the
// user can pick "Skip for now" or retry the connector — we do NOT abandon
// the whole AppWizard flow, because they still have a live draft app.
function onConnectorCreated() {
  connectorAdded.value = true
  stepId.value = 'confirm'
}

function onConnectorClose() {
  if (connectorAdded.value) {
    // Shouldn't happen in practice because @created handles the success
    // path, but if a caller emits close after create we still honour it.
    stepId.value = 'confirm'
  } else {
    stepId.value = 'choice'
  }
}

// Final commit. Clears draft state so the discard path becomes a no-op
// (the wizard closes cleanly without rolling anything back), then selects
// the new app and routes to its dashboard.
async function finish() {
  if (!draftAppId.value) return
  const committedId = draftAppId.value
  const committedName = details.name.trim()
  draftAppId.value = null
  app.selectApp(committedId)
  // Non-blocking success toast. The wizard already shows an explicit
  // success screen at step 3, so the toast is a secondary acknowledgement
  // that survives the route transition and catches the user's eye on
  // the dashboard. Delete flow is handled in-place by DeleteAppModal's
  // own stage-2 confirmation, so it intentionally doesn't toast.
  toast.show(`Application "${committedName}" created`, 'success')
  emit('close')
  router.push('/dashboard')
}

// —————————————————————————————————————————————————————————————————————
// Discard / close handling
//
// There are two layers:
//   1. If a draft exists, show the confirmation overlay first.
//   2. The confirmed path calls deleteApp() to roll back the eager-created
//      row, then closes.
// Closing while still on step 1 (no draft) is a clean no-op close.
// —————————————————————————————————————————————————————————————————————

function requestClose() {
  if (hasDraft.value) {
    showDiscardConfirm.value = true
  } else {
    emit('close')
  }
}

function cancelDiscard() {
  showDiscardConfirm.value = false
}

async function confirmDiscard() {
  showDiscardConfirm.value = false
  if (draftAppId.value) {
    try {
      await app.deleteApp(draftAppId.value)
    } catch {
      // Best-effort cleanup. If the delete fails (network, 409 last-app
      // guard in an impossible edge case), the user will see the orphan
      // in Settings → Applications and can clean it up there. We still
      // close the wizard so the user isn't trapped.
    }
    draftAppId.value = null
  }
  emit('close')
}

// Global keybindings — matches ConnectionWizard's conventions. We only
// listen while the AppWizard modal shell is visible; when the connector
// sub-flow takes over, ConnectionWizard installs its own Esc handler and
// we don't want to fight with it.
function onKeydown(e: KeyboardEvent) {
  if (stepId.value === 'connector') return
  if (e.key === 'Escape') requestClose()
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => document.removeEventListener('keydown', onKeydown))
</script>

<template>
  <!-- Connector sub-flow: hand off to ConnectionWizard and hide our shell
       so the user doesn't see two stacked modal backdrops. When it closes
       or completes, control returns to AppWizard. -->
  <ConnectionWizard
    v-if="stepId === 'connector' && draftAppId"
    :appId="draftAppId"
    @created="onConnectorCreated"
    @close="onConnectorClose"
  />

  <!-- Backdrop + modal shell. Hidden during the connector sub-flow. -->
  <div
    v-else
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
    @click.self="requestClose"
  >
    <div class="relative w-full max-w-xl mx-4 border border-border rounded-lg bg-bg-surface shadow-2xl max-h-[85vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-border shrink-0">
        <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">
          New Application
        </h3>
        <button
          @click="requestClose"
          class="font-mono text-xs text-text-muted hover:text-text-primary transition-colors cursor-pointer"
        >&times;</button>
      </div>

      <!-- Step indicator -->
      <div class="px-6 py-3 border-b border-border shrink-0">
        <WizardStepIndicator :steps="stepLabels" :currentIndex="currentStepIndex" />
      </div>

      <!-- Body -->
      <div class="px-6 py-6 overflow-y-auto flex-1">
        <div v-if="error" class="mb-4 rounded border border-status-critical/30 bg-status-critical/10 px-4 py-2">
          <p class="text-sm font-mono text-status-critical">{{ error }}</p>
        </div>

        <StepAppDetails
          v-if="stepId === 'details'"
          v-model="details"
          :disabled="creating"
          @valid="detailsValid = $event"
          @submit="advanceFromDetails"
        />

        <StepConnectorChoice
          v-else-if="stepId === 'choice'"
          @choose="handleChoice"
        />

        <StepConfirm
          v-else-if="stepId === 'confirm'"
          :appName="details.name"
          :connectorAdded="connectorAdded"
        />
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-between px-6 py-3 border-t border-border shrink-0">
        <!-- Discard is the left-side action on every step where a draft exists. -->
        <button
          v-if="hasDraft"
          @click="requestClose"
          class="px-5 py-2 font-mono text-xs uppercase tracking-wider rounded border border-border text-text-muted hover:border-status-critical/50 hover:text-status-critical transition-colors cursor-pointer"
        >
          Discard
        </button>
        <span v-else />

        <div class="flex gap-2">
          <button
            v-if="stepId === 'details'"
            @click="advanceFromDetails"
            :disabled="!detailsValid || creating"
            class="px-5 py-2 font-mono text-xs font-medium uppercase tracking-wider rounded transition-colors cursor-pointer"
            :class="(!detailsValid || creating)
              ? 'bg-action/30 text-bg-primary cursor-not-allowed'
              : 'bg-action text-bg-primary hover:bg-action-hover'"
          >
            {{ creating ? 'Creating...' : 'Continue' }}
          </button>

          <button
            v-else-if="stepId === 'confirm'"
            @click="finish"
            class="px-5 py-2 font-mono text-xs font-medium uppercase tracking-wider rounded bg-action text-bg-primary hover:bg-action-hover transition-colors cursor-pointer"
          >
            Go to Dashboard
          </button>
        </div>
      </div>

      <!-- Discard confirmation overlay -->
      <div
        v-if="showDiscardConfirm"
        class="absolute inset-0 z-10 flex items-center justify-center bg-black/50 rounded-lg"
      >
        <div class="border border-border rounded-lg bg-bg-elevated p-6 max-w-xs text-center shadow-xl">
          <p class="font-mono text-sm font-medium text-text-primary mb-1">Discard new application?</p>
          <p class="font-mono text-xs text-text-secondary mb-4">
            {{ details.name || 'The new app' }} and any connectors created during this session will be deleted.
          </p>
          <div class="flex justify-center gap-3">
            <button
              @click="cancelDiscard"
              class="px-5 py-2 font-mono text-xs uppercase tracking-wider rounded border border-border text-text-secondary hover:border-border-hover hover:text-text-primary transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              @click="confirmDiscard"
              class="px-5 py-2 font-mono text-xs font-medium uppercase tracking-wider rounded bg-status-critical text-bg-primary hover:bg-status-critical/80 transition-colors cursor-pointer"
            >
              Discard
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
