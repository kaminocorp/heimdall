<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { formatDate } from '@/utils/format'

// Profile section — shows what Heimdall knows about the signed-in user.
// There's no separate /auth/me fetch required because the Supabase session
// already carries email + created_at on `auth.user`; dropping an extra HTTP
// roundtrip here keeps the page snappy.
//
// Sign-out is duplicated from the sidebar footer intentionally: users
// looking for "account actions" naturally visit Settings first, and the
// sidebar footer version is easy to miss inside the mobile drawer.

const auth = useAuthStore()
const app = useAppStore()
const router = useRouter()

const email = computed(() => auth.user?.email ?? '—')
const createdAt = computed(() =>
  auth.user?.created_at ? formatDate(auth.user.created_at) : '—'
)

async function handleLogout() {
  // Mirror the sidebar's teardown: reset app store before logging out so the
  // next login lands on a clean slate instead of inheriting stale state.
  app.reset()
  await auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <section class="border border-border rounded-lg bg-bg-surface">
    <header class="px-6 py-4 border-b border-border">
      <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">
        Profile
      </h3>
      <p class="mt-1 font-mono text-xs text-text-muted">
        The account you're signed in with.
      </p>
    </header>

    <div class="px-6 py-5 space-y-4">
      <div class="grid grid-cols-[auto_1fr] gap-x-6 gap-y-3 text-sm font-mono">
        <div class="text-xs uppercase tracking-wider text-text-muted self-center">Email</div>
        <div class="text-text-primary truncate">{{ email }}</div>

        <div class="text-xs uppercase tracking-wider text-text-muted self-center">Joined</div>
        <div class="text-text-secondary">{{ createdAt }}</div>
      </div>

      <div class="pt-3 border-t border-border">
        <button
          type="button"
          @click="handleLogout"
          class="px-5 py-2 font-mono text-xs uppercase tracking-wider rounded border border-border text-text-secondary hover:border-status-critical/50 hover:text-status-critical transition-colors cursor-pointer"
        >
          Sign out
        </button>
      </div>
    </div>
  </section>
</template>
