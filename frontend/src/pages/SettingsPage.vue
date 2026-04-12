<script setup lang="ts">
import ProfileSection from '@/components/settings/ProfileSection.vue'
import OrganisationSection from '@/components/settings/OrganisationSection.vue'
import ApplicationsSection from '@/components/settings/ApplicationsSection.vue'

// Settings page — a flat vertical stack of three sections. No tabs, no
// sub-routes: the sections are small enough that a single scroll is the
// clearest way to see everything at once, and anchor navigation via
// in-page links can be added later if the surface grows.
</script>

<template>
  <div>
    <!-- Page header. Matches the styling used by SchedulesPage and other
         inner pages for visual consistency. -->
    <div class="pb-6 mb-8 border-b border-border">
      <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">
        Settings
      </h2>
      <p class="font-sans text-sm text-text-secondary mt-1">
        Profile, organisation, and application management.
      </p>
    </div>

    <div class="space-y-6 max-w-3xl">
      <ProfileSection />
      <OrganisationSection />
      <ApplicationsSection />
    </div>
  </div>
</template>
