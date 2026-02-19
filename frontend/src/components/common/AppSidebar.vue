<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <nav class="w-56 border-r border-gray-200 p-4 flex flex-col gap-2">
    <RouterLink to="/" class="px-3 py-2 rounded hover:bg-gray-100">Dashboard</RouterLink>
    <RouterLink to="/connections" class="px-3 py-2 rounded hover:bg-gray-100">Connections</RouterLink>
    <RouterLink to="/agent/config" class="px-3 py-2 rounded hover:bg-gray-100">Agent Config</RouterLink>
    <RouterLink to="/agent/chat" class="px-3 py-2 rounded hover:bg-gray-100">Agent Chat</RouterLink>
    <RouterLink to="/agent/log" class="px-3 py-2 rounded hover:bg-gray-100">Agent Log</RouterLink>
    <RouterLink to="/reports" class="px-3 py-2 rounded hover:bg-gray-100">Reports</RouterLink>
  </nav>
</template>
