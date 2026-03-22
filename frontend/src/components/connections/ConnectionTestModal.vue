<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import type { Connection } from '@/types/connection'
import { useConnectionsStore } from '@/stores/connections'

const props = defineProps<{
  connection: Connection
}>()

const emit = defineEmits<{
  close: []
}>()

const store = useConnectionsStore()

const phase = ref<'testing' | 'success' | 'error'>('testing')
const message = ref('')
const elapsed = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const connectionMeta = computed(() => {
  const c = props.connection
  const details: { label: string; value: string }[] = [
    { label: 'Name', value: c.name },
    { label: 'Type', value: c.type },
    { label: 'Direction', value: c.direction.replace('_', '-') },
  ]
  if (c.type === 'postgres' && c.config) {
    const cfg = c.config as Record<string, unknown>
    if (cfg.host) details.push({ label: 'Host', value: String(cfg.host) })
    if (cfg.port) details.push({ label: 'Port', value: String(cfg.port) })
    if (cfg.database) details.push({ label: 'Database', value: String(cfg.database) })
    if (cfg.user) details.push({ label: 'User', value: String(cfg.user) })
    if (cfg.ssl_mode) details.push({ label: 'SSL', value: String(cfg.ssl_mode) })
  }
  return details
})

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && phase.value !== 'testing') emit('close')
}

onMounted(async () => {
  document.addEventListener('keydown', onKeydown)
  timer = setInterval(() => { elapsed.value += 100 }, 100)

  try {
    const result = await store.testConnection(props.connection.id)
    if (result.success) {
      phase.value = 'success'
      message.value = result.message
    } else {
      phase.value = 'error'
      message.value = result.message
    }
  } catch (e: any) {
    phase.value = 'error'
    message.value = e.response?.data?.message ?? e.message ?? 'Unexpected error during connection test'
  } finally {
    if (timer) clearInterval(timer)
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeydown)
  if (timer) {
    clearInterval(timer)
    timer = null
  }
})

const elapsedDisplay = computed(() => (elapsed.value / 1000).toFixed(1) + 's')

</script>

<template>
  <!-- Backdrop -->
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" @click.self="phase !== 'testing' && emit('close')">
    <!-- Modal -->
    <div class="w-full max-w-lg mx-4 border border-border rounded-lg bg-bg-surface shadow-2xl">
      <!-- Header -->
      <div class="flex items-center justify-between px-5 py-4 border-b border-border">
        <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">Connection Test</h3>
        <button
          v-if="phase !== 'testing'"
          @click="emit('close')"
          class="font-mono text-xs text-text-muted hover:text-text-primary transition-colors cursor-pointer"
        >&times;</button>
      </div>

      <!-- Connection details -->
      <div class="px-5 py-4 border-b border-border space-y-1.5">
        <div v-for="detail in connectionMeta" :key="detail.label" class="flex items-baseline justify-between">
          <span class="font-mono text-[10px] uppercase tracking-wider text-text-muted">{{ detail.label }}</span>
          <span class="font-mono text-xs text-text-secondary truncate ml-4 max-w-[70%] text-right">{{ detail.value }}</span>
        </div>
      </div>

      <!-- Test status -->
      <div class="px-5 py-5">
        <!-- Testing phase -->
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

        <!-- Success phase -->
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

        <!-- Error phase -->
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

      <!-- Footer -->
      <div class="flex justify-end px-5 py-3 border-t border-border">
        <button
          @click="emit('close')"
          :disabled="phase === 'testing'"
          class="px-4 py-1.5 font-mono text-xs uppercase tracking-wider rounded transition-colors cursor-pointer"
          :class="phase === 'testing'
            ? 'text-text-muted border border-border cursor-not-allowed'
            : 'text-text-primary border border-border hover:border-border-hover'"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>
