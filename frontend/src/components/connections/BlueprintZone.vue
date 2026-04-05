<script setup lang="ts">
import type { Connection } from '@/types/connection'
import BlueprintNode from './BlueprintNode.vue'

defineProps<{
  label: string
  iconType: 'server' | 'database' | 'code'
  connections: Connection[]
  emptyPrompt: string
  testingId?: string | null
  iconMap: Record<string, string>
}>()

const emit = defineEmits<{
  delete: [id: string]
  edit: [connection: Connection]
  test: [id: string]
  'manage-repos': [id: string]
  add: []
}>()
</script>

<template>
  <div ref="zoneEl" class="border border-border rounded-lg bg-bg-surface/50 p-4">
    <!-- Zone header -->
    <div class="flex items-center gap-2 mb-3">
      <!-- Server icon -->
      <svg v-if="iconType === 'server'" class="w-4 h-4 text-text-muted shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <rect x="2" y="2" width="20" height="8" rx="2" />
        <rect x="2" y="14" width="20" height="8" rx="2" />
        <circle cx="6" cy="6" r="1" fill="currentColor" />
        <circle cx="6" cy="18" r="1" fill="currentColor" />
      </svg>
      <!-- Database icon -->
      <svg v-else-if="iconType === 'database'" class="w-4 h-4 text-text-muted shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <ellipse cx="12" cy="5" rx="9" ry="3" />
        <path d="M3 5v14c0 1.66 4.03 3 9 3s9-1.34 9-3V5" />
        <path d="M3 12c0 1.66 4.03 3 9 3s9-1.34 9-3" />
      </svg>
      <!-- Code icon -->
      <svg v-else class="w-4 h-4 text-text-muted shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="16 18 22 12 16 6" />
        <polyline points="8 6 2 12 8 18" />
        <line x1="14" y1="4" x2="10" y2="20" />
      </svg>

      <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">
        {{ label }}
      </p>
    </div>

    <!-- Connection nodes -->
    <div v-if="connections.length > 0" class="space-y-2">
      <BlueprintNode
        v-for="(conn, i) in connections"
        :key="conn.id"
        :connection="conn"
        :testing="testingId === conn.id"
        :icon="iconMap[conn.type] ?? '??'"
        class="animate-fade-in"
        :style="{ '--stagger-index': i }"
        @delete="emit('delete', $event)"
        @edit="emit('edit', $event)"
        @test="emit('test', $event)"
        @manage-repos="emit('manage-repos', $event)"
      />
    </div>

    <!-- Empty state -->
    <button
      v-else
      @click="emit('add')"
      class="w-full border border-dashed border-border-hover rounded-lg px-4 py-3 text-center hover:border-accent/40 hover:bg-accent/5 transition-colors cursor-pointer group/empty"
    >
      <p class="font-mono text-xs text-text-muted group-hover/empty:text-text-secondary transition-colors">
        {{ emptyPrompt }}
      </p>
    </button>
  </div>
</template>
