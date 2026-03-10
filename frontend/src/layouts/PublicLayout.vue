<script setup lang="ts">
import PublicNav from '@/components/public/PublicNav.vue'
import PublicFooter from '@/components/public/PublicFooter.vue'
</script>

<template>
  <div class="min-h-screen flex flex-col bg-bg-primary">
    <PublicNav />
    <RouterView />
    <PublicFooter />
  </div>
</template>
