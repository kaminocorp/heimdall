<script setup lang="ts">
import type { ChatMessage } from '@/types/agent'

defineProps<{
  message: ChatMessage
}>()
</script>

<template>
  <div class="flex" :class="message.role === 'user' ? 'justify-end' : 'justify-start'">
    <div class="max-w-2xl w-full">
      <!-- Role label -->
      <div
        class="font-mono text-xs font-medium uppercase tracking-widest mb-1"
        :class="message.role === 'user' ? 'text-right text-accent' : 'text-text-muted'"
      >
        {{ message.role === 'user' ? 'Operator' : 'Heimdall' }}
      </div>
      <!-- Message bubble -->
      <div
        class="rounded-lg px-4 py-3 text-sm border"
        :class="[
          message.role === 'user'
            ? 'bg-accent-subtle border-accent-border/50 text-text-primary'
            : 'bg-bg-surface border-border text-text-secondary animate-reveal',
        ]"
      >
        {{ message.content }}
      </div>
    </div>
  </div>
</template>
