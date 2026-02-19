<script setup lang="ts">
import type { Connection } from '@/types/connection'
import StatusBadge from '@/components/common/StatusBadge.vue'

defineProps<{
  connection: Connection
}>()
</script>

<template>
  <div class="border rounded-lg p-4">
    <div class="flex items-center justify-between">
      <h3 class="font-medium">{{ connection.name }}</h3>
      <StatusBadge :status="connection.status" />
    </div>
    <p class="text-sm text-gray-500 mt-1">{{ connection.type }}</p>
  </div>
</template>
