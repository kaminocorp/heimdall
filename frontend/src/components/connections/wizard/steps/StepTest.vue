<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { useConnectionsStore } from '@/stores/connections'
import { extractApiError } from '@/utils/apiError'
import type { WizardState } from '../flows'

const props = defineProps<{
  modelValue: WizardState
  connectionId?: string | null
}>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

const store = useConnectionsStore()
const phase = ref<'testing' | 'success' | 'error'>('testing')
const message = ref('')
const elapsed = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  if (!props.connectionId) {
    phase.value = 'error'
    message.value = 'No connection ID — cannot test.'
    emit('valid', false)
    return
  }

  timer = setInterval(() => { elapsed.value += 100 }, 100)

  try {
    const result = await store.testConnection(props.connectionId)
    if (result.success) {
      phase.value = 'success'
      message.value = result.message
      emit('valid', true)
    } else {
      phase.value = 'error'
      message.value = result.message
      emit('valid', false)
    }
  } catch (e: unknown) {
    phase.value = 'error'
    message.value = extractApiError(e, 'Unexpected error during connection test')
    emit('valid', false)
  } finally {
    if (timer) clearInterval(timer)
  }
})

onBeforeUnmount(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
})

const elapsedDisplay = computed(() => (elapsed.value / 1000).toFixed(1) + 's')
</script>

<template>
  <div class="space-y-4">
    <!-- Testing -->
    <div v-if="phase === 'testing'" class="space-y-3">
      <div class="flex items-center gap-3">
        <span class="relative flex h-2.5 w-2.5">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent opacity-75"></span>
          <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-accent"></span>
        </span>
        <span class="font-mono text-sm text-text-primary">Testing connection...</span>
        <span class="font-mono text-xs text-text-muted tabular-nums ml-auto">{{ elapsedDisplay }}</span>
      </div>
      <div class="w-full h-0.5 bg-border rounded-full overflow-hidden">
        <div class="h-full bg-accent rounded-full animate-pulse" style="width: 60%"></div>
      </div>
    </div>

    <!-- Success -->
    <div v-else-if="phase === 'success'" class="space-y-3">
      <div class="flex items-center gap-3">
        <span class="inline-flex h-2.5 w-2.5 rounded-full bg-status-ok"></span>
        <span class="font-mono text-sm text-status-ok font-medium">Test passed</span>
        <span class="font-mono text-xs text-text-muted tabular-nums ml-auto">{{ elapsedDisplay }}</span>
      </div>
      <div class="rounded border border-status-ok/20 bg-status-ok/5 px-3 py-2">
        <p class="font-mono text-xs text-text-secondary">{{ message }}</p>
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="phase === 'error'" class="space-y-3">
      <div class="flex items-center gap-3">
        <span class="inline-flex h-2.5 w-2.5 rounded-full bg-status-critical"></span>
        <span class="font-mono text-sm text-status-critical font-medium">Test failed</span>
        <span class="font-mono text-xs text-text-muted tabular-nums ml-auto">{{ elapsedDisplay }}</span>
      </div>
      <div class="rounded border border-status-critical/20 bg-status-critical/5 px-3 py-2">
        <p class="font-mono text-xs text-text-secondary break-all whitespace-pre-wrap">{{ message }}</p>
      </div>
    </div>
  </div>
</template>
