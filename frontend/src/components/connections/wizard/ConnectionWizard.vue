<script setup lang="ts">
import { ref, computed, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useConnectionsStore } from '@/stores/connections'
import { useAppStore } from '@/stores/app'
import { getFlowById, type PlatformFlow, type WizardState } from './flows'
import PlatformGrid from './PlatformGrid.vue'
import WizardStepIndicator from './WizardStepIndicator.vue'

const emit = defineEmits<{
  close: []
  created: [connectionId: string]
}>()

const store = useConnectionsStore()
const appStore = useAppStore()

// — State machine —
const selectedFlow = ref<PlatformFlow | null>(null)
const currentStepIndex = ref(0)
const stepValid = ref(false)
const creating = ref(false)
const error = ref<string | null>(null)
const createdConnectionId = ref<string | null>(null)

const state = reactive<WizardState>({
  name: '',
  config: {},
})

// — Derived —
const currentStep = computed(() => selectedFlow.value?.steps[currentStepIndex.value] ?? null)
const stepLabels = computed(() => selectedFlow.value?.steps.map(s => s.label) ?? [])
const isFirstStep = computed(() => currentStepIndex.value === 0)
const isLastStep = computed(() =>
  selectedFlow.value ? currentStepIndex.value === selectedFlow.value.steps.length - 1 : false
)
const isTestStep = computed(() => currentStep.value?.id === 'test')

// — Actions —
function selectPlatform(flowId: string) {
  const flow = getFlowById(flowId)
  if (!flow) return
  selectedFlow.value = flow
  currentStepIndex.value = 0
  stepValid.value = false
  state.name = ''
  state.config = {}
  error.value = null
  createdConnectionId.value = null
}

function goBack() {
  if (currentStepIndex.value > 0) {
    currentStepIndex.value--
    stepValid.value = true // Previous steps were already validated.
  } else {
    // Back to platform selection.
    selectedFlow.value = null
  }
}

async function goNext() {
  if (!selectedFlow.value) return

  // If we're about to enter the test step, create the connection first.
  const nextStep = selectedFlow.value.steps[currentStepIndex.value + 1]
  if (nextStep?.id === 'test' && !createdConnectionId.value) {
    await createConnection()
    if (error.value) return
  }

  if (!isLastStep.value) {
    currentStepIndex.value++
    stepValid.value = false
  }
}

async function createConnection() {
  if (!selectedFlow.value || !appStore.currentAppId) return

  creating.value = true
  error.value = null

  try {
    const conn = await store.createConnection({
      app_id: appStore.currentAppId,
      name: state.name,
      type: selectedFlow.value.connectorType,
      direction: selectedFlow.value.direction,
      config: state.config,
    })
    createdConnectionId.value = conn.id
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    error.value = err.response?.data?.error ?? 'Failed to create connection'
  } finally {
    creating.value = false
  }
}

async function finish() {
  // For flows without a test step (webhook, github), create the connection now.
  if (!createdConnectionId.value) {
    await createConnection()
    if (error.value) return
  }

  if (createdConnectionId.value) {
    emit('created', createdConnectionId.value)
  }
  emit('close')
}

async function handleClose() {
  // If a connection was created but the user is abandoning the wizard,
  // clean it up to avoid orphaned records.
  if (createdConnectionId.value) {
    try {
      await store.deleteConnection(createdConnectionId.value)
    } catch {
      // Best-effort cleanup — don't block close on failure.
    }
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
  <!-- Backdrop -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
    @click.self="handleClose"
  >
    <!-- Modal -->
    <div class="w-full max-w-xl mx-4 border border-border rounded-lg bg-bg-surface shadow-2xl max-h-[85vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between px-5 py-4 border-b border-border shrink-0">
        <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">
          {{ selectedFlow ? selectedFlow.name + ' Connection' : 'New Connection' }}
        </h3>
        <button
          @click="handleClose"
          class="font-mono text-xs text-text-muted hover:text-text-primary transition-colors cursor-pointer"
        >&times;</button>
      </div>

      <!-- Step indicator -->
      <div v-if="selectedFlow" class="px-5 py-3 border-b border-border shrink-0">
        <WizardStepIndicator :steps="stepLabels" :currentIndex="currentStepIndex" />
      </div>

      <!-- Body -->
      <div class="px-5 py-5 overflow-y-auto flex-1">
        <!-- Error banner -->
        <div v-if="error" class="mb-4 rounded border border-status-critical/30 bg-status-critical/10 px-4 py-2">
          <p class="text-sm font-mono text-status-critical">{{ error }}</p>
        </div>

        <!-- Platform selection -->
        <template v-if="!selectedFlow">
          <p class="font-mono text-xs text-text-secondary mb-4">Choose a platform to connect.</p>
          <PlatformGrid @select="selectPlatform" />
        </template>

        <!-- Step content with slide transitions -->
        <template v-else>
          <Transition
            enter-active-class="transition-all duration-200 ease-out"
            enter-from-class="opacity-0 translate-x-4"
            enter-to-class="opacity-100 translate-x-0"
            leave-active-class="transition-all duration-150 ease-in"
            leave-from-class="opacity-100 translate-x-0"
            leave-to-class="opacity-0 -translate-x-4"
            mode="out-in"
          >
            <component
              :is="currentStep!.component"
              :key="currentStep!.id"
              v-model="state"
              :connectionId="createdConnectionId"
              @valid="stepValid = $event"
            />
          </Transition>
        </template>
      </div>

      <!-- Footer -->
      <div v-if="selectedFlow" class="flex items-center justify-between px-5 py-3 border-t border-border shrink-0">
        <button
          @click="goBack"
          class="px-4 py-1.5 font-mono text-xs uppercase tracking-wider rounded border border-border text-text-secondary hover:border-border-hover hover:text-text-primary transition-colors cursor-pointer"
        >
          Back
        </button>

        <div class="flex gap-2">
          <!-- For test step or last step: Done button -->
          <button
            v-if="isLastStep"
            @click="finish"
            :disabled="!stepValid && isTestStep"
            class="px-4 py-1.5 font-mono text-xs font-medium uppercase tracking-wider rounded transition-colors cursor-pointer"
            :class="(!stepValid && isTestStep)
              ? 'bg-action/30 text-bg-primary cursor-not-allowed'
              : 'bg-action text-bg-primary hover:bg-action-hover'"
          >
            Done
          </button>
          <!-- For non-last steps: Continue button -->
          <button
            v-else
            @click="goNext"
            :disabled="!stepValid || creating"
            class="px-4 py-1.5 font-mono text-xs font-medium uppercase tracking-wider rounded transition-colors cursor-pointer"
            :class="(!stepValid || creating)
              ? 'bg-action/30 text-bg-primary cursor-not-allowed'
              : 'bg-action text-bg-primary hover:bg-action-hover'"
          >
            {{ creating ? 'Creating...' : 'Continue' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
