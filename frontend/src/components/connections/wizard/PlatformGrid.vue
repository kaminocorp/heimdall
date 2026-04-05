<script setup lang="ts">
import { computed } from 'vue'
import { flows, type PlatformFlow } from './flows'
import PlatformCard from './PlatformCard.vue'

const emit = defineEmits<{
  select: [flowId: string]
}>()

const categories: { key: PlatformFlow['category']; label: string }[] = [
  { key: 'log_source', label: 'Log Sources' },
  { key: 'database', label: 'Databases' },
  { key: 'generic', label: 'Generic' },
]

const grouped = computed(() =>
  categories
    .map(cat => ({
      ...cat,
      flows: flows.filter(f => f.category === cat.key),
    }))
    .filter(cat => cat.flows.length > 0)
)
</script>

<template>
  <div class="space-y-6">
    <div v-for="group in grouped" :key="group.key">
      <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted mb-2">
        {{ group.label }}
      </p>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <PlatformCard
          v-for="flow in group.flows"
          :key="flow.id"
          :flow="flow"
          @select="emit('select', $event)"
        />
      </div>
    </div>
  </div>
</template>
