<script setup lang="ts">
import { ref, computed } from 'vue'
import { useToast } from '@/composables/useToast'
import {
  listNotificationChannels,
  createNotificationChannel,
  updateNotificationChannel as apiUpdateChannel,
  deleteNotificationChannel,
  testNotificationChannel,
} from '@/api/notifications'
import type { NotificationChannel, NotificationChannelType } from '@/types/notification'
import { extractApiError } from '@/utils/apiError'
import BaseSelect from '@/components/common/BaseSelect.vue'

const props = defineProps<{
  appId: string
  channels: NotificationChannel[]
}>()

const emit = defineEmits<{
  'update:channels': [channels: NotificationChannel[]]
}>()

const toast = useToast()

const showForm = ref(false)
const editingChannelId = ref<string | null>(null)
const saving = ref(false)
const testingChannelId = ref<string | null>(null)
const deleteTarget = ref<NotificationChannel | null>(null)

const formType = ref<NotificationChannelType>('slack')
const formName = ref('')
const formEnabled = ref(true)
const formWebhookUrl = ref('')
const formRecipients = ref('')

const channelTypeOptions: { label: string; value: NotificationChannelType }[] = [
  { label: 'Slack', value: 'slack' },
  { label: 'Discord', value: 'discord' },
  { label: 'Email', value: 'email' },
]

const channelConfig = computed(() => {
  if (formType.value === 'email') {
    return { recipients: formRecipients.value.split(',').map(s => s.trim()).filter(Boolean) }
  }
  return { webhook_url: formWebhookUrl.value }
})

function channelConfigSummary(ch: NotificationChannel): string {
  if (ch.type === 'email') {
    return (ch.config as { recipients: string[] }).recipients.join(', ')
  }
  return (ch.config as { webhook_url: string }).webhook_url
}

function channelTypeLabel(type: string) {
  switch (type) {
    case 'email': return 'Email'
    case 'slack': return 'Slack'
    case 'discord': return 'Discord'
    default: return type
  }
}

function openAdd() {
  editingChannelId.value = null
  formType.value = 'slack'
  formName.value = ''
  formEnabled.value = true
  formWebhookUrl.value = ''
  formRecipients.value = ''
  showForm.value = true
}

function openEdit(ch: NotificationChannel) {
  editingChannelId.value = ch.id
  formType.value = ch.type
  formName.value = ch.name
  formEnabled.value = ch.enabled
  if (ch.type === 'email') {
    const cfg = ch.config as { recipients: string[] }
    formRecipients.value = cfg.recipients.join(', ')
    formWebhookUrl.value = ''
  } else {
    const cfg = ch.config as { webhook_url: string }
    formWebhookUrl.value = cfg.webhook_url
    formRecipients.value = ''
  }
  showForm.value = true
}

function cancelForm() {
  showForm.value = false
  editingChannelId.value = null
}

async function save() {
  saving.value = true
  try {
    if (editingChannelId.value) {
      await apiUpdateChannel(props.appId, editingChannelId.value, {
        name: formName.value,
        config: channelConfig.value,
        enabled: formEnabled.value,
      })
      toast.show('Channel updated', 'success')
    } else {
      await createNotificationChannel(props.appId, {
        type: formType.value,
        name: formName.value,
        config: channelConfig.value,
        enabled: formEnabled.value,
      })
      toast.show('Channel created', 'success')
    }
    showForm.value = false
    editingChannelId.value = null
    try {
      emit('update:channels', await listNotificationChannels(props.appId))
    } catch {
      toast.show('Saved, but failed to refresh the list — please reload', 'error')
    }
  } catch (e: unknown) {
    toast.show(extractApiError(e, 'Failed to save channel'), 'error')
  } finally {
    saving.value = false
  }
}

function confirmDelete(ch: NotificationChannel) {
  deleteTarget.value = ch
}

async function handleDelete() {
  if (!deleteTarget.value) return
  const ch = deleteTarget.value
  deleteTarget.value = null
  try {
    await deleteNotificationChannel(props.appId, ch.id)
    emit('update:channels', props.channels.filter(c => c.id !== ch.id))
    toast.show('Channel deleted', 'success')
  } catch {
    toast.show('Failed to delete channel', 'error')
  }
}

async function test(ch: NotificationChannel) {
  testingChannelId.value = ch.id
  try {
    await testNotificationChannel(props.appId, ch.id)
    toast.show('Test notification sent', 'success')
  } catch (e: unknown) {
    toast.show(extractApiError(e, 'Test failed'), 'error')
  } finally {
    testingChannelId.value = null
  }
}

defineExpose({ resetForm: () => { showForm.value = false; editingChannelId.value = null } })
</script>

<template>
  <section>
    <div class="flex items-center justify-between mb-3">
      <h3 class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Channels</h3>
      <button
        v-if="!showForm"
        @click="openAdd"
        class="px-3 py-1 border border-accent-border/50 text-text-secondary font-mono text-xs uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
      >
        + Add Channel
      </button>
    </div>

    <!-- Channel form -->
    <form v-if="showForm" @submit.prevent="save" class="border border-border rounded-lg bg-bg-surface p-6 space-y-6 mb-4">
      <div v-if="!editingChannelId">
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Type</label>
        <BaseSelect v-model="formType" :options="channelTypeOptions" />
      </div>

      <div>
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Name</label>
        <input
          v-model="formName"
          type="text"
          required
          placeholder="e.g. Ops Slack"
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors"
        />
      </div>

      <!-- Webhook URL (Slack / Discord) -->
      <div v-if="formType !== 'email'">
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Webhook URL</label>
        <input
          v-model="formWebhookUrl"
          type="url"
          required
          :placeholder="formType === 'slack' ? 'https://hooks.slack.com/services/...' : 'https://discord.com/api/webhooks/...'"
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors"
        />
      </div>

      <!-- Recipients (Email) -->
      <div v-if="formType === 'email'">
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Recipients</label>
        <input
          v-model="formRecipients"
          type="text"
          required
          placeholder="ops@company.com, dev@company.com"
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors"
        />
        <p class="mt-1 font-mono text-xs text-text-muted">Comma-separated email addresses</p>
      </div>

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

      <div class="flex gap-2 pt-2">
        <button
          type="submit"
          :disabled="saving"
          class="px-5 py-2.5 bg-action text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-action-hover transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ saving ? 'Saving...' : (editingChannelId ? 'Update Channel' : 'Add Channel') }}
        </button>
        <button
          type="button"
          @click="cancelForm"
          :disabled="saving"
          class="px-5 py-2.5 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
        >
          Cancel
        </button>
      </div>
    </form>

    <!-- Channel list -->
    <div v-if="channels.length === 0 && !showForm" class="border border-border rounded-lg bg-bg-surface p-8 text-center">
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
            <span class="font-mono text-xs uppercase tracking-wider px-1.5 py-0.5 rounded border border-border text-text-muted">
              {{ channelTypeLabel(ch.type) }}
            </span>
            <span
              class="w-1.5 h-1.5 rounded-full"
              :class="ch.enabled ? 'bg-accent' : 'bg-text-muted'"
            />
          </div>
          <p class="font-mono text-xs text-text-muted mt-0.5 truncate">
            {{ channelConfigSummary(ch) }}
          </p>
        </div>

        <div class="flex items-center gap-1.5 flex-shrink-0">
          <button
            @click="test(ch)"
            :disabled="testingChannelId === ch.id"
            class="px-2.5 py-1 border border-border text-text-secondary font-mono text-xs uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer disabled:opacity-50"
          >
            {{ testingChannelId === ch.id ? 'Sending...' : 'Test' }}
          </button>
          <button
            @click="openEdit(ch)"
            class="px-2.5 py-1 border border-border text-text-secondary font-mono text-xs uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer"
          >
            Edit
          </button>
          <button
            @click="confirmDelete(ch)"
            class="px-2.5 py-1 border border-border text-text-secondary font-mono text-xs uppercase tracking-wider rounded hover:border-status-critical/50 hover:text-status-critical transition-colors cursor-pointer"
          >
            Del
          </button>
        </div>
      </div>
    </div>
    <!-- Delete confirmation modal -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition duration-100 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div v-if="deleteTarget" class="fixed inset-0 z-50 flex items-center justify-center">
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="deleteTarget = null" />
          <div class="relative z-10 w-full max-w-md mx-4 rounded-lg border border-border bg-bg-elevated shadow-xl shadow-black/40 p-6">
            <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary mb-2">
              Delete channel
            </h3>
            <p class="font-sans text-sm text-text-secondary mb-6">
              Are you sure you want to delete
              <span class="font-mono text-text-primary">{{ deleteTarget.name }}</span>?
              Notifications will no longer be sent to this channel.
            </p>
            <div class="flex items-center justify-end gap-3">
              <button
                @click="deleteTarget = null"
                class="px-4 py-2 font-mono text-xs uppercase tracking-wider rounded border border-border text-text-secondary hover:text-text-primary hover:border-border-hover transition-colors cursor-pointer"
              >
                Cancel
              </button>
              <button
                @click="handleDelete"
                class="px-4 py-2 font-mono text-xs uppercase tracking-wider rounded bg-status-critical text-white hover:bg-status-critical/80 transition-colors cursor-pointer"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </section>
</template>
