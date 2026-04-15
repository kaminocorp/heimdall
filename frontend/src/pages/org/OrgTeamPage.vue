<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { listOrgMembers, inviteOrgMember, updateMemberRole, removeOrgMember } from '@/api/organizations'
import { extractApiError } from '@/utils/apiError'
import type { OrgMember, OrgMemberRole } from '@/types/organization'

const auth = useAuthStore()
const appStore = useAppStore()

// ── State ──
const members = ref<OrgMember[]>([])
const loading = ref(true)
const fetchError = ref<string | null>(null)
const actionError = ref<string | null>(null)

// Invite form
const inviteEmail = ref('')
const inviteRole = ref<OrgMemberRole>('member')
const inviting = ref(false)
const inviteError = ref<string | null>(null)
const inviteSuccess = ref<string | null>(null)

// Confirm removal
const removeTarget = ref<OrgMember | null>(null)
const removing = ref(false)

// ── Computed ──
const callerRole = computed<OrgMemberRole>(() => appStore.organization?.role ?? 'member')
const isOwner = computed(() => callerRole.value === 'owner')
const isAdmin = computed(() => callerRole.value === 'admin' || callerRole.value === 'owner')
const currentUserId = computed(() => auth.user?.id ?? '')

// ── Fetch ──
async function fetchMembers() {
  loading.value = true
  fetchError.value = null
  try {
    members.value = await listOrgMembers()
  } catch (e: unknown) {
    fetchError.value = extractApiError(e, 'Failed to load members')
  } finally {
    loading.value = false
  }
}

watch(() => appStore.organization?.id, () => fetchMembers(), { immediate: true, flush: 'post' })

// ── Invite ──
async function handleInvite() {
  if (!inviteEmail.value.trim()) return
  inviting.value = true
  inviteError.value = null
  inviteSuccess.value = null
  try {
    await inviteOrgMember(inviteEmail.value.trim(), inviteRole.value)
    inviteSuccess.value = `Invited ${inviteEmail.value.trim()}`
    inviteEmail.value = ''
    inviteRole.value = 'member'
    await fetchMembers()
  } catch (e: unknown) {
    inviteError.value = extractApiError(e, 'Failed to invite member')
  } finally {
    inviting.value = false
  }
}

// ── Role change ──
async function handleRoleChange(member: OrgMember, newRole: OrgMemberRole) {
  const oldRole = member.role
  member.role = newRole // optimistic update
  try {
    await updateMemberRole(member.user_id, newRole)
  } catch (e: unknown) {
    member.role = oldRole // revert on failure
    actionError.value = extractApiError(e, 'Failed to update role')
  }
}

// ── Remove ──
function confirmRemove(member: OrgMember) {
  removeTarget.value = member
}

function cancelRemove() {
  removeTarget.value = null
}

async function handleRemove() {
  if (!removeTarget.value) return
  removing.value = true
  try {
    await removeOrgMember(removeTarget.value.user_id)
    removeTarget.value = null
    await fetchMembers()
  } catch (e: unknown) {
    actionError.value = extractApiError(e, 'Failed to remove member')
    removeTarget.value = null
  } finally {
    removing.value = false
  }
}

// ── Helpers ──
function roleLabel(role: OrgMemberRole): string {
  return role.charAt(0).toUpperCase() + role.slice(1)
}

function roleBadgeClass(role: OrgMemberRole): string {
  switch (role) {
    case 'owner':
      return 'border-accent/30 bg-accent/10 text-accent'
    case 'admin':
      return 'border-status-warn/30 bg-status-warn/10 text-status-warn'
    default:
      return 'border-border bg-bg-surface text-text-muted'
  }
}

function initials(email: string): string {
  return email.charAt(0).toUpperCase()
}
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="pb-6 mb-8 border-b border-border">
      <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">
        Team
      </h2>
      <p class="font-sans text-sm text-text-secondary mt-1">
        Manage members of {{ appStore.organization?.name ?? 'your organization' }}.
      </p>
    </div>

    <!-- Invite section (admin+ only) -->
    <div v-if="isAdmin" class="mb-8 rounded-lg border border-border bg-bg-surface p-5">
      <h3 class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted mb-4">
        Invite member
      </h3>
      <form @submit.prevent="handleInvite" class="flex flex-wrap items-end gap-3">
        <div class="flex-1 min-w-[200px]">
          <label class="block font-mono text-xs text-text-muted mb-1.5">Email address</label>
          <input
            v-model="inviteEmail"
            type="email"
            placeholder="colleague@company.com"
            required
            class="w-full px-3 py-2 rounded border border-border bg-bg-primary font-mono text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent/50 transition-colors"
          />
        </div>
        <div class="w-36">
          <label class="block font-mono text-xs text-text-muted mb-1.5">Role</label>
          <select
            v-model="inviteRole"
            class="w-full px-3 py-2 rounded border border-border bg-bg-primary font-mono text-sm text-text-secondary focus:outline-none focus:border-accent/50 transition-colors cursor-pointer"
          >
            <option value="member">Member</option>
            <option value="admin">Admin</option>
            <option v-if="isOwner" value="owner">Owner</option>
          </select>
        </div>
        <button
          type="submit"
          :disabled="inviting || !inviteEmail.trim()"
          class="px-5 py-2 font-mono text-xs font-medium uppercase tracking-wider rounded bg-action text-bg-primary hover:bg-action-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
        >
          {{ inviting ? 'Inviting...' : 'Send invite' }}
        </button>
      </form>

      <!-- Invite feedback -->
      <div v-if="inviteError" class="mt-3 rounded border border-status-critical/30 bg-status-critical/10 px-3 py-2">
        <p class="font-mono text-xs text-status-critical">{{ inviteError }}</p>
      </div>
      <div v-if="inviteSuccess" class="mt-3 rounded border border-status-ok/30 bg-status-ok/10 px-3 py-2">
        <p class="font-mono text-xs text-status-ok">{{ inviteSuccess }}</p>
      </div>

      <!-- Role descriptions -->
      <div class="mt-4 grid grid-cols-1 sm:grid-cols-3 gap-3">
        <div class="px-3 py-2 rounded border border-border/50 bg-bg-primary">
          <div class="font-mono text-xs font-medium text-text-primary mb-0.5">Member</div>
          <div class="font-sans text-xs text-text-muted">View all data. Cannot manage team or settings.</div>
        </div>
        <div class="px-3 py-2 rounded border border-border/50 bg-bg-primary">
          <div class="font-mono text-xs font-medium text-text-primary mb-0.5">Admin</div>
          <div class="font-sans text-xs text-text-muted">Invite and remove members. Edit org settings.</div>
        </div>
        <div class="px-3 py-2 rounded border border-border/50 bg-bg-primary">
          <div class="font-mono text-xs font-medium text-text-primary mb-0.5">Owner</div>
          <div class="font-sans text-xs text-text-muted">Full control. Change roles. Transfer ownership.</div>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <div v-for="n in 3" :key="n" class="flex items-center gap-4 px-4 py-3 border border-border rounded-lg bg-bg-surface animate-pulse">
        <div class="w-8 h-8 rounded-full bg-border/30" />
        <div class="flex-1 space-y-2">
          <div class="h-3 w-48 bg-border/30 rounded" />
          <div class="h-2.5 w-24 bg-border/20 rounded" />
        </div>
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="fetchError && members.length === 0" class="rounded border border-status-critical/30 bg-status-critical/10 px-5 py-4">
      <p class="font-mono text-sm text-status-critical">{{ fetchError }}</p>
      <button
        @click="fetchMembers"
        class="mt-2 font-mono text-xs text-text-secondary hover:text-text-primary underline cursor-pointer"
      >
        Retry
      </button>
    </div>

    <!-- Member list -->
    <div v-else class="rounded-lg border border-border bg-bg-surface overflow-hidden">
      <!-- Header -->
      <div class="flex items-center gap-4 px-5 py-3 border-b border-border bg-bg-elevated">
        <div class="flex-1 font-mono text-xs font-medium uppercase tracking-wider text-text-muted">
          {{ members.length }} {{ members.length === 1 ? 'member' : 'members' }}
        </div>
      </div>

      <!-- Rows -->
      <div
        v-for="member in members"
        :key="member.user_id"
        class="flex items-center gap-4 px-5 py-4 border-b border-border last:border-b-0"
        :class="member.user_id === currentUserId ? 'bg-accent-subtle/20' : ''"
      >
        <!-- Avatar -->
        <div class="w-9 h-9 rounded-full border border-border bg-bg-primary flex items-center justify-center font-mono text-sm font-bold text-text-secondary shrink-0">
          {{ initials(member.email) }}
        </div>

        <!-- Info -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <span class="font-mono text-sm text-text-primary truncate">
              {{ member.email }}
            </span>
            <span
              v-if="member.user_id === currentUserId"
              class="font-mono text-xs text-text-muted"
            >
              (you)
            </span>
          </div>
          <div class="font-mono text-xs text-text-muted mt-0.5">
            Joined {{ new Date(member.created_at).toLocaleDateString() }}
          </div>
        </div>

        <!-- Role badge / selector -->
        <div class="shrink-0">
          <!-- Owner can change roles via dropdown -->
          <select
            v-if="isOwner && member.user_id !== currentUserId"
            :value="member.role"
            @change="handleRoleChange(member, ($event.target as HTMLSelectElement).value as OrgMemberRole)"
            class="px-3 py-1.5 rounded border font-mono text-xs uppercase tracking-wider cursor-pointer focus:outline-none focus:border-accent/50 transition-colors"
            :class="roleBadgeClass(member.role)"
          >
            <option value="owner">Owner</option>
            <option value="admin">Admin</option>
            <option value="member">Member</option>
          </select>
          <!-- Non-owners or viewing self: read-only badge -->
          <span
            v-else
            class="inline-flex px-3 py-1.5 rounded border font-mono text-xs font-medium uppercase tracking-wider"
            :class="roleBadgeClass(member.role)"
          >
            {{ roleLabel(member.role) }}
          </span>
        </div>

        <!-- Remove button (admin+ only, not for self if sole member) -->
        <div class="shrink-0 w-20 text-right">
          <button
            v-if="isAdmin && member.user_id !== currentUserId && member.role !== 'owner'"
            @click="confirmRemove(member)"
            class="px-3 py-1.5 rounded border border-border font-mono text-xs text-text-muted hover:text-status-critical hover:border-status-critical/40 transition-colors cursor-pointer"
          >
            Remove
          </button>
        </div>
      </div>
    </div>

    <!-- Inline error banner (for non-fatal errors like failed role change) -->
    <div v-if="actionError" class="mt-4 rounded border border-status-critical/30 bg-status-critical/10 px-4 py-3">
      <div class="flex items-center justify-between">
        <p class="font-mono text-xs text-status-critical">{{ actionError }}</p>
        <button @click="actionError = null" class="font-mono text-xs text-text-muted hover:text-text-secondary cursor-pointer">Dismiss</button>
      </div>
    </div>

    <!-- Remove confirmation modal -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition duration-100 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="removeTarget"
          class="fixed inset-0 z-50 flex items-center justify-center"
        >
          <!-- Backdrop -->
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="cancelRemove" />

          <!-- Dialog -->
          <div class="relative z-10 w-full max-w-md mx-4 rounded-lg border border-border bg-bg-elevated shadow-xl shadow-black/40 p-6">
            <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary mb-2">
              Remove member
            </h3>
            <p class="font-sans text-sm text-text-secondary mb-6">
              Are you sure you want to remove
              <span class="font-mono text-text-primary">{{ removeTarget.email }}</span>
              from the organisation? They will lose access to all projects.
            </p>
            <div class="flex items-center justify-end gap-3">
              <button
                @click="cancelRemove"
                class="px-4 py-2 font-mono text-xs uppercase tracking-wider rounded border border-border text-text-secondary hover:text-text-primary hover:border-border-hover transition-colors cursor-pointer"
              >
                Cancel
              </button>
              <button
                @click="handleRemove"
                :disabled="removing"
                class="px-4 py-2 font-mono text-xs uppercase tracking-wider rounded bg-status-critical text-white hover:bg-status-critical/80 transition-colors disabled:opacity-50 cursor-pointer"
              >
                {{ removing ? 'Removing...' : 'Remove' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
