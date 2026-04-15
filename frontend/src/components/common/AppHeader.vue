<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { RouterLink, useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import AppWizard from '@/components/app-wizard/AppWizard.vue'
import CreateOrgModal from '@/components/org/CreateOrgModal.vue'
import OrgDropdown from './OrgDropdown.vue'
import AppDropdown from './AppDropdown.vue'
import ProfileDropdown from './ProfileDropdown.vue'

const app = useAppStore()
const router = useRouter()
const route = useRoute()

const isOrgContext = computed(() => route.meta.context === 'org')

// ── Refs for outside-click handling ──
const orgDropdownRef = ref<InstanceType<typeof OrgDropdown> | null>(null)
const appDropdownRef = ref<InstanceType<typeof AppDropdown> | null>(null)
const profileDropdownRef = ref<InstanceType<typeof ProfileDropdown> | null>(null)

// ── Modals ──
const wizardOpen = ref(false)
const createOrgOpen = ref(false)

// ── Org dropdown handlers ──
async function handleSelectOrg(orgId: string) {
  if (orgId === app.organization?.id) {
    router.push('/org')
    return
  }
  await app.selectOrg(orgId)
  router.push('/org')
}

function openCreateOrg() {
  createOrgOpen.value = true
}

function onCreateOrgClose() {
  createOrgOpen.value = false
  router.push('/org')
}

// ── Close on outside click ──
function onClickOutside(e: MouseEvent) {
  const target = e.target as Node
  if (orgDropdownRef.value?.el && !orgDropdownRef.value.el.contains(target)) orgDropdownRef.value.close()
  if (appDropdownRef.value?.el && !appDropdownRef.value.el.contains(target)) appDropdownRef.value.close()
  if (profileDropdownRef.value?.el && !profileDropdownRef.value.el.contains(target)) profileDropdownRef.value.close()
}

onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))
</script>

<template>
  <header class="h-12 flex items-center border-b border-border bg-bg-elevated px-4 flex-shrink-0 z-40">
    <!-- Left: Logo + breadcrumbs -->
    <div class="flex items-center gap-1 min-w-0">
      <!-- Logo icon -->
      <RouterLink :to="isOrgContext ? '/org' : '/dashboard'" class="flex items-center gap-2.5 flex-shrink-0 mr-1">
        <svg class="w-5 h-5 text-accent" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
          <g transform="translate(100,100)" fill="currentColor">
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(-10)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(10)"/>
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(50)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(70)"/>
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(110)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(130)"/>
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(170)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(190)"/>
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(230)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(250)"/>
            <polygon points="-4,-17 4,-17 4,-72 -4,-83" transform="rotate(290)"/>
            <polygon points="-4,-17 4,-17 4,-83 -4,-72" transform="rotate(310)"/>
            <circle r="5.5"/>
          </g>
        </svg>
        <span class="hidden sm:inline font-mono text-xs font-semibold uppercase tracking-[0.25em] text-text-primary">
          H E I M D A L L
        </span>
      </RouterLink>

      <span class="text-text-muted text-lg font-light mx-1 select-none">/</span>

      <OrgDropdown ref="orgDropdownRef" :is-org-context="isOrgContext" @select="handleSelectOrg" @create="openCreateOrg" />

      <!-- App breadcrumb — only shown when in app context -->
      <template v-if="!isOrgContext">
        <span class="text-text-muted text-lg font-light mx-1 select-none">/</span>
        <AppDropdown ref="appDropdownRef" @create="wizardOpen = true" />
      </template>
    </div>

    <!-- Right: Agent chat + Profile -->
    <div class="ml-auto flex items-center gap-2">
      <RouterLink
        to="/agent/chat"
        class="flex items-center justify-center w-7 h-7 rounded-full border border-border bg-bg-surface text-text-secondary hover:text-accent hover:border-accent/40 transition-colors"
        title="Agent Chat"
      >
        <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 2L3 7v10l9 5 9-5V7l-9-5z" />
          <circle cx="12" cy="12" r="2.5" fill="currentColor" stroke="none" />
        </svg>
      </RouterLink>

      <ProfileDropdown ref="profileDropdownRef" />
    </div>

    <!-- App wizard modal (teleported) -->
    <AppWizard v-if="wizardOpen" @close="wizardOpen = false" />

    <!-- Create org modal -->
    <CreateOrgModal v-if="createOrgOpen" @close="onCreateOrgClose" />
  </header>
</template>
