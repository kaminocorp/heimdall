<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const { toasts, dismiss } = useToast()

const typeClasses: Record<string, string> = {
  success: 'border-status-ok/30 bg-status-ok/10 text-status-ok',
  error: 'border-status-critical/30 bg-status-critical/10 text-status-critical',
  info: 'border-status-info/30 bg-status-info/10 text-status-info',
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 max-w-sm">
      <TransitionGroup
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="opacity-0 translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 translate-y-2"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          :class="[
            'rounded border px-4 py-3 font-mono text-sm flex items-start gap-3',
            typeClasses[toast.type] || typeClasses.info,
          ]"
        >
          <span class="flex-1">{{ toast.message }}</span>
          <button
            class="opacity-60 hover:opacity-100 transition-opacity text-xs mt-0.5"
            @click="dismiss(toast.id)"
          >
            &times;
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
