<script setup lang="ts">
import type { Connection } from '@/types/connection'
import StatusBadge from '@/components/common/StatusBadge.vue'

defineProps<{
  connection: Connection
}>()

const emit = defineEmits<{
  delete: [id: string]
}>()
</script>

<template>
  <div class="border rounded-lg p-4">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="font-medium">{{ connection.name }}</h3>
        <p class="text-sm text-gray-500 mt-1">{{ connection.type }} &middot; {{ connection.direction }}</p>
      </div>
      <div class="flex items-center gap-3">
        <StatusBadge :status="connection.status" />
        <button
          @click="emit('delete', connection.id)"
          class="text-sm text-red-600 hover:text-red-800"
        >
          Delete
        </button>
      </div>
    </div>
  </div>
</template>
