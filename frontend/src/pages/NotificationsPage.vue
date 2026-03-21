<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { useToast } from '@/composables/useToast'
import {
  getNotificationPreferences,
  updateNotificationPreferences,
  listNotificationChannels,
  createNotificationChannel,
  updateNotificationChannel as apiUpdateChannel,
  deleteNotificationChannel,
  testNotificationChannel,
  listNotificationHistory,
} from '@/api/notifications'
import type {
  NotificationPreferences,
  NotificationChannel,
  NotificationChannelType,
  NotificationLogEntry,
} from '@/types/notification'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'

const appStore = useAppStore()
const toast = useToast()

// --- State ---
const loading = ref(false)
const prefs = ref<NotificationPreferences | null>(null)
const channels = ref<NotificationChannel[]>([])
const history = ref<NotificationLogEntry[]>([])

// Preferences editing
const editingPrefs = ref(false)
const savingPrefs = ref(false)
const formEnabled = ref(false)
const formThreshold = ref<string>('warning')
const formCooldown = ref(15)

// Channel form
const showChannelForm = ref(false)
const editingChannelId = ref<string | null>(null)
const savingChannel = ref(false)
const formChannelType = ref<NotificationChannelType>('slack')
const formChannelName = ref('')
const formChannelEnabled = ref(true)
const formWebhookUrl = ref('')
const formRecipients = ref('')

const testingChannelId = ref<string | null>(null)

function channelConfigSummary(ch: NotificationChannel): string {
  if (ch.type === 'email') {
    return (ch.config as { recipients: string[] }).recipients.join(', ')
  }
  return (ch.config as { webhook_url: string }).webhook_url
}

const thresholdOptions = [
  { label: 'Info+', value: 'info' },
  { label: 'Warning+', value: 'warning' },
  { label: 'Error+', value: 'error' },
  { label: 'Critical only', value: 'critical' },
]

const channelTypeOptions: { label: string; value: NotificationChannelType }[] = [
  { label: 'Slack', value: 'slack' },
  { label: 'Discord', value: 'discord' },
  { label: 'Email', value: 'email' },
]

// --- Fetching ---
async function fetchAll() {
  const appId = appStore.currentAppId
  if (!appId) return
  loading.value = true
  try {
    const [p, c, h] = await Promise.all([
      getNotificationPreferences(appId),
      listNotificationChannels(appId),
      listNotificationHistory(appId, 10),
    ])
    prefs.value = p
    channels.value = c
    history.value = h
  } catch {
    toast.show('Failed to load notification settings', 'error')
  } finally {
    loading.value = false
  }
}

// --- Preferences ---
function startEditPrefs() {
  if (!prefs.value) return
  formEnabled.value = prefs.value.enabled
  formThreshold.value = prefs.value.severity_threshold
  formCooldown.value = prefs.value.cooldown_minutes
  editingPrefs.value = true
}

async function savePrefs() {
  const appId = appStore.currentAppId
  if (!appId) return
  savingPrefs.value = true
  try {
    prefs.value = await updateNotificationPreferences(appId, {
      enabled: formEnabled.value,
      severity_threshold: formThreshold.value,
      cooldown_minutes: formCooldown.value,
    })
    editingPrefs.value = false
    toast.show('Preferences saved', 'success')
  } catch {
    toast.show('Failed to save preferences', 'error')
  } finally {
    savingPrefs.value = false
  }
}

// --- Channels ---
function openAddChannel() {
  editingChannelId.value = null
  formChannelType.value = 'slack'
  formChannelName.value = ''
  formChannelEnabled.value = true
  formWebhookUrl.value = ''
  formRecipients.value = ''
  showChannelForm.value = true
}

function openEditChannel(ch: NotificationChannel) {
  editingChannelId.value = ch.id
  formChannelType.value = ch.type
  formChannelName.value = ch.name
  formChannelEnabled.value = ch.enabled
  if (ch.type === 'email') {
    const cfg = ch.config as { recipients: string[] }
    formRecipients.value = cfg.recipients.join(', ')
    formWebhookUrl.value = ''
  } else {
    const cfg = ch.config as { webhook_url: string }
    formWebhookUrl.value = cfg.webhook_url
    formRecipients.value = ''
  }
  showChannelForm.value = true
}

function cancelChannelForm() {
  showChannelForm.value = false
  editingChannelId.value = null
}

const channelConfig = computed(() => {
  if (formChannelType.value === 'email') {
    return { recipients: formRecipients.value.split(',').map(s => s.trim()).filter(Boolean) }
  }
  return { webhook_url: formWebhookUrl.value }
})

async function saveChannel() {
  const appId = appStore.currentAppId
  if (!appId) return
  savingChannel.value = true
  try {
    if (editingChannelId.value) {
      await apiUpdateChannel(appId, editingChannelId.value, {
        name: formChannelName.value,
        config: channelConfig.value,
        enabled: formChannelEnabled.value,
      })
      toast.show('Channel updated', 'success')
    } else {
      await createNotificationChannel(appId, {
        type: formChannelType.value,
        name: formChannelName.value,
        config: channelConfig.value,
        enabled: formChannelEnabled.value,
      })
      toast.show('Channel created', 'success')
    }
    showChannelForm.value = false
    editingChannelId.value = null
    channels.value = await listNotificationChannels(appId)
  } catch (e: any) {
    const msg = e?.response?.data?.error || 'Failed to save channel'
    toast.show(msg, 'error')
  } finally {
    savingChannel.value = false
  }
}

async function removeChannel(ch: NotificationChannel) {
  const appId = appStore.currentAppId
  if (!appId) return
  try {
    await deleteNotificationChannel(appId, ch.id)
    channels.value = channels.value.filter(c => c.id !== ch.id)
    toast.show('Channel deleted', 'success')
  } catch {
    toast.show('Failed to delete channel', 'error')
  }
}

async function testChannel(ch: NotificationChannel) {
  const appId = appStore.currentAppId
  if (!appId) return
  testingChannelId.value = ch.id
  try {
    await testNotificationChannel(appId, ch.id)
    toast.show('Test notification sent', 'success')
  } catch (e: any) {
    const msg = e?.response?.data?.error || 'Test failed'
    toast.show(msg, 'error')
  } finally {
    testingChannelId.value = null
  }
}

function channelTypeLabel(type: string) {
  switch (type) {
    case 'email': return 'Email'
    case 'slack': return 'Slack'
    case 'discord': return 'Discord'
    default: return type
  }
}

function severityIndicator(severity: string) {
  switch (severity) {
    case 'critical': return 'bg-status-critical'
    case 'error': return 'bg-status-error'
    case 'warning': return 'bg-status-warning'
    default: return 'bg-accent'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'sent': return 'text-accent'
    case 'failed': return 'text-status-critical'
    default: return 'text-text-muted'
  }
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

onMounted(fetchAll)
watch(() => appStore.currentAppId, () => {
  editingPrefs.value = false
  showChannelForm.value = false
  fetchAll()
})
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="pb-6 mb-8 border-b border-border flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">Notifications</h2>
        <p class="font-sans text-sm text-text-secondary mt-1">
          Alert channels and escalation preferences
          <span v-if="appStore.currentApp" class="text-text-muted">— {{ appStore.currentApp.name }}</span>
        </p>
      </div>
    </div>

    <!-- Skeleton loader -->
    <div v-if="loading" class="space-y-6">
      <div class="border border-border rounded-lg bg-bg-surface p-5 space-y-4">
        <div v-for="n in 3" :key="n" class="flex items-baseline justify-between py-2">
          <SkeletonBlock width="5rem" height="0.75rem" />
          <SkeletonBlock width="10rem" height="0.75rem" />
        </div>
      </div>
    </div>

    <div v-else class="space-y-8">
      <!-- ═══ PREFERENCES ═══ -->
      <section>
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Preferences</h3>
          <button
            v-if="prefs && !editingPrefs"
            @click="startEditPrefs"
            class="px-3 py-1 border border-accent-border/50 text-text-secondary font-mono text-xs uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
          >
            Edit
          </button>
        </div>

        <!-- Edit mode -->
        <form v-if="editingPrefs" @submit.prevent="savePrefs" class="border border-border rounded-lg bg-bg-surface p-5 space-y-5">
          <div class="flex items-center gap-3">
            <label class="font-mono text-xs font-medium uppercase tracking-wider text-text-secondary">Enabled</label>
            <button
              type="button"
              @click="formEnabled = !formEnabled"
              class="relative w-10 h-5 rounded-full transition-colors cursor-pointer"
              :class="formEnabled ? 'bg-accent' : 'bg-bg-elevated border border-border'"
            >
              <span
                class="absolute top-0.5 w-4 h-4 rounded-full bg-text-primary transition-transform"
                :class="formEnabled ? 'translate-x-5' : 'translate-x-0.5'"
              />
            </button>
          </div>

          <div>
            <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Severity Threshold</label>
            <BaseSelect v-model="formThreshold" :options="thresholdOptions" />
          </div>

          <div>
            <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Cooldown (minutes)</label>
            <input
              v-model.number="formCooldown"
              type="number"
              min="1"
              max="1440"
              class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
            />
            <p class="mt-1 font-mono text-[10px] text-text-muted">Suppress duplicate notifications for this many minutes (1–1440)</p>
          </div>

          <div class="flex gap-2 pt-2">
            <button
              type="submit"
              :disabled="savingPrefs"
              class="px-4 py-2 bg-accent text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-accent-hover transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ savingPrefs ? 'Saving...' : 'Save' }}
            </button>
            <button
              type="button"
              @click="editingPrefs = false"
              :disabled="savingPrefs"
              class="px-4 py-2 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
            >
              Cancel
            </button>
          </div>
        </form>

        <!-- Display mode -->
        <div v-else-if="prefs" class="border border-border rounded-lg bg-bg-surface p-5">
          <div class="space-y-3">
            <div class="flex items-baseline justify-between py-1.5 border-b border-border">
              <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Status</span>
              <span class="font-mono text-sm" :class="prefs.enabled ? 'text-accent' : 'text-text-muted'">
                {{ prefs.enabled ? 'Enabled' : 'Disabled' }}
              </span>
            </div>
            <div class="flex items-baseline justify-between py-1.5 border-b border-border">
              <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Threshold</span>
              <span class="font-mono text-sm text-text-primary capitalize">{{ prefs.severity_threshold }}+</span>
            </div>
            <div class="flex items-baseline justify-between py-1.5">
              <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Cooldown</span>
              <span class="font-mono text-sm text-text-primary">{{ prefs.cooldown_minutes }} min</span>
            </div>
          </div>
        </div>
      </section>

      <!-- ═══ CHANNELS ═══ -->
      <section>
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Channels</h3>
          <button
            v-if="!showChannelForm"
            @click="openAddChannel"
            class="px-3 py-1 border border-accent-border/50 text-text-secondary font-mono text-xs uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
          >
            + Add Channel
          </button>
        </div>

        <!-- Channel form -->
        <form v-if="showChannelForm" @submit.prevent="saveChannel" class="border border-border rounded-lg bg-bg-surface p-5 space-y-5 mb-4">
          <div v-if="!editingChannelId">
            <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Type</label>
            <BaseSelect v-model="formChannelType" :options="channelTypeOptions" />
          </div>

          <div>
            <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Name</label>
            <input
              v-model="formChannelName"
              type="text"
              required
              placeholder="e.g. Ops Slack"
              class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
            />
          </div>

          <!-- Webhook URL (Slack / Discord) -->
          <div v-if="formChannelType !== 'email'">
            <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Webhook URL</label>
            <input
              v-model="formWebhookUrl"
              type="url"
              required
              :placeholder="formChannelType === 'slack' ? 'https://hooks.slack.com/services/...' : 'https://discord.com/api/webhooks/...'"
              class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
            />
          </div>

          <!-- Recipients (Email) -->
          <div v-if="formChannelType === 'email'">
            <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Recipients</label>
            <input
              v-model="formRecipients"
              type="text"
              required
              placeholder="ops@company.com, dev@company.com"
              class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
            />
            <p class="mt-1 font-mono text-[10px] text-text-muted">Comma-separated email addresses</p>
          </div>

          <div class="flex items-center gap-3">
            <label class="font-mono text-xs font-medium uppercase tracking-wider text-text-secondary">Enabled</label>
            <button
              type="button"
              @click="formChannelEnabled = !formChannelEnabled"
              class="relative w-10 h-5 rounded-full transition-colors cursor-pointer"
              :class="formChannelEnabled ? 'bg-accent' : 'bg-bg-elevated border border-border'"
            >
              <span
                class="absolute top-0.5 w-4 h-4 rounded-full bg-text-primary transition-transform"
                :class="formChannelEnabled ? 'translate-x-5' : 'translate-x-0.5'"
              />
            </button>
          </div>

          <div class="flex gap-2 pt-2">
            <button
              type="submit"
              :disabled="savingChannel"
              class="px-4 py-2 bg-accent text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-accent-hover transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ savingChannel ? 'Saving...' : (editingChannelId ? 'Update Channel' : 'Add Channel') }}
            </button>
            <button
              type="button"
              @click="cancelChannelForm"
              :disabled="savingChannel"
              class="px-4 py-2 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
            >
              Cancel
            </button>
          </div>
        </form>

        <!-- Channel list -->
        <div v-if="channels.length === 0 && !showChannelForm" class="border border-border rounded-lg bg-bg-surface p-8 text-center">
          <p class="font-mono text-sm text-text-muted">No notification channels configured</p>
        </div>

        <div v-else class="space-y-2">
          <div
            v-for="ch in channels"
            :key="ch.id"
            class="border border-border rounded-lg bg-bg-surface p-4 flex items-center justify-between gap-4"
          >
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="font-mono text-sm font-medium text-text-primary">{{ ch.name }}</span>
                <span class="font-mono text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded border border-border text-text-muted">
                  {{ channelTypeLabel(ch.type) }}
                </span>
                <span
                  class="w-1.5 h-1.5 rounded-full"
                  :class="ch.enabled ? 'bg-accent' : 'bg-text-muted'"
                />
              </div>
              <p class="font-mono text-[11px] text-text-muted mt-0.5 truncate">
                {{ channelConfigSummary(ch) }}
              </p>
            </div>

            <div class="flex items-center gap-1.5 flex-shrink-0">
              <button
                @click="testChannel(ch)"
                :disabled="testingChannelId === ch.id"
                class="px-2.5 py-1 border border-border text-text-secondary font-mono text-[10px] uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer disabled:opacity-50"
              >
                {{ testingChannelId === ch.id ? 'Sending...' : 'Test' }}
              </button>
              <button
                @click="openEditChannel(ch)"
                class="px-2.5 py-1 border border-border text-text-secondary font-mono text-[10px] uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
              >
                Edit
              </button>
              <button
                @click="removeChannel(ch)"
                class="px-2.5 py-1 border border-border text-text-secondary font-mono text-[10px] uppercase tracking-wider rounded hover:border-status-critical/50 hover:text-status-critical transition-colors cursor-pointer"
              >
                Del
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- ═══ RECENT NOTIFICATIONS ═══ -->
      <section>
        <h3 class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted mb-3">Recent Notifications</h3>

        <div v-if="history.length === 0" class="border border-border rounded-lg bg-bg-surface p-8 text-center">
          <p class="font-mono text-sm text-text-muted">No notifications sent yet</p>
        </div>

        <div v-else class="border border-border rounded-lg bg-bg-surface overflow-hidden">
          <div
            v-for="entry in history"
            :key="entry.id"
            class="flex items-center gap-3 px-4 py-3 border-b border-border last:border-b-0"
          >
            <span class="w-2 h-2 rounded-full flex-shrink-0" :class="severityIndicator(entry.severity)" />
            <span class="font-mono text-xs text-text-primary flex-1 min-w-0 truncate">{{ entry.summary }}</span>
            <span class="font-mono text-[10px] uppercase tracking-wider text-text-muted flex-shrink-0">{{ entry.channel_name }}</span>
            <span class="font-mono text-[10px] uppercase tracking-wider flex-shrink-0" :class="statusLabel(entry.status)">{{ entry.status }}</span>
            <span class="font-mono text-[10px] text-text-muted flex-shrink-0 w-16 text-right">{{ timeAgo(entry.created_at) }}</span>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
