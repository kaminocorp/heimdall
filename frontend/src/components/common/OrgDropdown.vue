<script setup lang="ts">
import { ref } from 'vue'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

const open = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

defineExpose({ el: dropdownRef, close: () => { open.value = false } })

const emit = defineEmits<{
  select: [orgId: string]
  create: []
}>()

function toggle() {
  open.value = !open.value
}

function handleSelect(orgId: string) {
  open.value = false
  emit('select', orgId)
}

function handleCreate() {
  open.value = false
  emit('create')
}

defineProps<{
  isOrgContext: boolean
}>()
</script>

<template>
  <div ref="dropdownRef" class="relative">
    <button
      @click="toggle"
      class="flex items-center gap-1.5 px-2 py-1 rounded text-sm font-mono transition-colors cursor-pointer"
      :class="isOrgContext
        ? 'text-text-primary hover:bg-bg-surface'
        : 'text-text-secondary hover:text-text-primary hover:bg-bg-surface'"
    >
      <svg class="w-3.5 h-3.5 text-text-muted flex-shrink-0" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />
      </svg>
      <span class="truncate max-w-[160px]">{{ app.organization?.name ?? 'Organization' }}</span>
      <svg class="w-3 h-3 text-text-muted flex-shrink-0" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M3 4.5L6 7.5L9 4.5" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>

    <Transition
      enter-active-class="transition duration-100 ease-out"
      enter-from-class="opacity-0 -translate-y-1"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-75 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-1"
    >
      <div v-if="open" class="absolute top-full left-0 mt-1 w-64 border border-border rounded bg-bg-elevated shadow-lg shadow-black/30 z-50">
        <div class="py-1 max-h-64 overflow-y-auto">
          <button
            v-for="o in app.organizations"
            :key="o.id"
            @click="handleSelect(o.id)"
            class="w-full text-left px-3 py-2 text-sm font-mono transition-colors cursor-pointer flex items-center justify-between gap-2"
            :class="o.id === app.organization?.id
              ? 'text-accent-bright bg-accent-subtle'
              : 'text-text-secondary hover:text-text-primary hover:bg-bg-surface'"
          >
            <span class="truncate">{{ o.name }}</span>
            <span class="shrink-0 font-mono text-xs uppercase tracking-wider px-1.5 py-0.5 rounded border"
              :class="{
                'border-accent/30 bg-accent/10 text-accent': o.role === 'owner',
                'border-status-warn/30 bg-status-warn/10 text-status-warn': o.role === 'admin',
                'border-border bg-bg-surface text-text-muted': o.role === 'member',
              }"
            >
              {{ o.role }}
            </span>
          </button>
        </div>
        <div class="border-t border-border py-1">
          <button
            @click="handleCreate"
            class="w-full text-left px-3 py-2 text-sm font-mono text-text-muted hover:text-text-primary hover:bg-bg-surface transition-colors cursor-pointer"
          >
            + New organisation
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>
