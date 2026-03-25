<script setup lang="ts">
defineProps<{
  modelValue: 'blueprint' | 'list'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: 'blueprint' | 'list']
}>()
</script>

<template>
  <div class="flex border border-border rounded overflow-hidden">
    <button
      @click="emit('update:modelValue', 'blueprint')"
      class="px-3 py-1.5 font-mono text-xs uppercase tracking-wider transition-colors cursor-pointer"
      :class="modelValue === 'blueprint'
        ? 'bg-accent/20 text-text-primary'
        : 'text-text-muted hover:text-text-secondary'"
    >
      Blueprint
    </button>
    <button
      @click="emit('update:modelValue', 'list')"
      class="px-3 py-1.5 font-mono text-xs uppercase tracking-wider transition-colors cursor-pointer border-l border-border"
      :class="modelValue === 'list'
        ? 'bg-accent/20 text-text-primary'
        : 'text-text-muted hover:text-text-secondary'"
    >
      List
    </button>
  </div>
</template>
