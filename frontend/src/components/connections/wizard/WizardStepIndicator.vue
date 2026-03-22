<script setup lang="ts">
defineProps<{
  steps: string[]
  currentIndex: number
}>()
</script>

<template>
  <div class="flex items-center justify-center gap-1">
    <template v-for="(step, i) in steps" :key="step">
      <!-- Dot -->
      <div class="flex flex-col items-center gap-1.5">
        <div
          class="w-2.5 h-2.5 rounded-full transition-colors duration-200"
          :class="
            i < currentIndex
              ? 'bg-accent/60'
              : i === currentIndex
                ? 'bg-accent'
                : 'bg-border'
          "
        />
        <span
          class="font-mono text-[10px] uppercase tracking-wider transition-colors duration-200"
          :class="
            i === currentIndex
              ? 'text-text-primary'
              : 'text-text-muted'
          "
        >{{ step }}</span>
      </div>
      <!-- Line between dots -->
      <div
        v-if="i < steps.length - 1"
        class="w-8 h-px mb-4 transition-colors duration-200"
        :class="i < currentIndex ? 'bg-accent/40' : 'bg-border'"
      />
    </template>
  </div>
</template>
