<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { formatDate } from '@/utils/format'

// Organisation section — a read-only summary of the user's org. Member
// management is explicitly out of scope for this phase (see plan doc):
// there's no members table or API endpoint yet, and the one-user-per-org
// invariant that onboarding currently enforces means a member count would
// always show "1". We show only the fields that have meaning today:
// name, slug, and created_at.
//
// When multi-user orgs land, this is the right place to surface an invite
// button, a member list, and role management — keeping everything
// organisation-related in one section.

const app = useAppStore()

const org = computed(() => app.organization)
const name = computed(() => org.value?.name ?? '—')
const slug = computed(() => org.value?.slug ?? '—')
const createdAt = computed(() =>
  org.value?.created_at ? formatDate(org.value.created_at) : '—'
)
</script>

<template>
  <section class="border border-border rounded-lg bg-bg-surface">
    <header class="px-6 py-4 border-b border-border">
      <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">
        Organisation
      </h3>
      <p class="mt-1 font-mono text-xs text-text-muted">
        Your workspace's shared context.
      </p>
    </header>

    <div class="px-6 py-5">
      <div class="grid grid-cols-[auto_1fr] gap-x-6 gap-y-3 text-sm font-mono">
        <div class="text-xs uppercase tracking-wider text-text-muted self-center">Name</div>
        <div class="text-text-primary truncate">{{ name }}</div>

        <div class="text-xs uppercase tracking-wider text-text-muted self-center">Slug</div>
        <div class="text-text-secondary truncate">{{ slug }}</div>

        <div class="text-xs uppercase tracking-wider text-text-muted self-center">Created</div>
        <div class="text-text-secondary">{{ createdAt }}</div>
      </div>
    </div>
  </section>
</template>
