<script setup lang="ts">
defineProps<{
  status: string
}>()
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full font-mono text-xs font-medium uppercase tracking-wider border"
    :class="{
      'border-status-ok/30 bg-status-ok/10 text-status-ok': status === 'active',
      'border-white/10 bg-white/5 text-text-muted': status === 'inactive',
      'border-status-critical/30 bg-status-critical/10 text-status-critical': status === 'error',
      'border-status-warn/30 bg-status-warn/10 text-status-warn': status === 'warning',
    }"
  >
    <span
      class="w-1.5 h-1.5 rounded-full"
      :class="{
        'bg-status-ok': status === 'active',
        'bg-text-muted': status === 'inactive',
        'bg-status-critical': status === 'error',
        'bg-status-warn': status === 'warning',
      }"
    />
    {{ status }}
  </span>
</template>
