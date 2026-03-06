<script setup lang="ts">
withDefaults(defineProps<{
  width?: string
  height?: string
  rounded?: boolean
}>(), {
  width: '100%',
  height: '1rem',
  rounded: false,
})
</script>

<template>
  <div
    class="animate-pulse bg-bg-surface"
    :class="rounded ? 'rounded-full' : 'rounded'"
    :style="{ width, height }"
  />
</template>
