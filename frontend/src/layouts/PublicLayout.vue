<script setup lang="ts">
import PublicNav from '@/components/public/PublicNav.vue'
import PublicFooter from '@/components/public/PublicFooter.vue'
</script>

<template>
  <div class="min-h-screen flex flex-col bg-bg-primary">
    <PublicNav />
    <div class="flex flex-col flex-1 pt-[76px]">
      <RouterView />
    </div>
    <PublicFooter />
  </div>
</template>
