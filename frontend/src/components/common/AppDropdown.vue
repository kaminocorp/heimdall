<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

const open = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

defineExpose({ el: dropdownRef, close: () => { open.value = false } })

const emit = defineEmits<{
  select: [appId: string]
  create: []
}>()

const currentAppName = computed(() => app.currentApp?.name ?? 'Select app')

function toggle() {
  open.value = !open.value
}

function handleSelect(appId: string) {
  app.selectApp(appId)
  open.value = false
  emit('select', appId)
}

function handleCreate() {
  open.value = false
  emit('create')
}
</script>

<template>
  <div ref="dropdownRef" class="relative">
    <button
      @click="toggle"
      class="flex items-center gap-1.5 px-2 py-1 rounded text-sm font-mono text-text-secondary hover:text-text-primary hover:bg-bg-surface transition-colors cursor-pointer"
    >
      <svg class="w-3.5 h-3.5 text-text-muted flex-shrink-0" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" />
      </svg>
      <span class="truncate max-w-[160px]">{{ currentAppName }}</span>
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
      <div v-if="open" class="absolute top-full left-0 mt-1 w-56 border border-border rounded bg-bg-elevated shadow-lg shadow-black/30 z-50">
        <div class="py-1 max-h-64 overflow-y-auto">
          <button
            v-for="a in app.applications"
            :key="a.id"
            @click="handleSelect(a.id)"
            class="w-full text-left px-3 py-2 text-sm font-mono transition-colors cursor-pointer"
            :class="a.id === app.currentAppId
              ? 'text-accent-bright bg-accent-subtle'
              : 'text-text-secondary hover:text-text-primary hover:bg-bg-surface'"
          >
            {{ a.name }}
          </button>
        </div>
        <div class="border-t border-border py-1">
          <button
            @click="handleCreate"
            class="w-full text-left px-3 py-2 text-sm font-mono text-text-muted hover:text-text-primary hover:bg-bg-surface transition-colors cursor-pointer"
          >
            + New application
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>
