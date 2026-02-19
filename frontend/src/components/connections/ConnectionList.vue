<script setup lang="ts">
import type { Connection } from '@/types/connection'
import ConnectionCard from './ConnectionCard.vue'

defineProps<{
  connections: Connection[]
}>()
</script>

<template>
  <div class="space-y-3">
    <ConnectionCard v-for="conn in connections" :key="conn.id" :connection="conn" />
  </div>
</template>
