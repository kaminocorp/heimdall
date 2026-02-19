<script setup lang="ts">
defineProps<{
  status: string
}>()
</script>

<template>
  <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium" :class="{
    'bg-green-100 text-green-800': status === 'active',
    'bg-gray-100 text-gray-800': status === 'inactive',
    'bg-red-100 text-red-800': status === 'error',
  }">
    {{ status }}
  </span>
</template>
