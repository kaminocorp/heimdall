<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue'

const props = defineProps<{ value: string }>()
const copied = ref(false)
let resetTimer: ReturnType<typeof setTimeout> | null = null

async function copy() {
  try {
    await navigator.clipboard.writeText(props.value)
  } catch {
    // Fallback for non-HTTPS environments (e.g. localhost dev).
    const textarea = document.createElement('textarea')
    textarea.value = props.value
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
  }
  copied.value = true
  if (resetTimer) clearTimeout(resetTimer)
  resetTimer = setTimeout(() => { copied.value = false }, 2000)
}

onBeforeUnmount(() => {
  if (resetTimer) clearTimeout(resetTimer)
})
</script>

<template>
  <div class="flex items-center gap-2 bg-bg-elevated/80 border border-border rounded px-3 py-2">
    <code class="flex-1 font-mono text-xs text-text-primary truncate">{{ value }}</code>
    <button
      type="button"
      @click="copy"
      class="shrink-0 font-mono text-xs uppercase tracking-wider px-2 py-0.5 rounded border transition-colors cursor-pointer"
      :class="copied
        ? 'border-status-ok/30 text-status-ok bg-status-ok/10'
        : 'border-border text-text-muted hover:text-text-primary hover:border-border-hover'"
    >
      {{ copied ? 'Copied' : 'Copy' }}
    </button>
  </div>
</template>
