<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { useToast } from '@/composables/useToast'
import {
  getNotificationPreferences,
  listNotificationChannels,
  listNotificationHistory,
} from '@/api/notifications'
import type {
  NotificationPreferences,
  NotificationChannel,
  NotificationLogEntry,
} from '@/types/notification'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import NotificationPreferencesSection from '@/components/notifications/NotificationPreferences.vue'
import NotificationChannels from '@/components/notifications/NotificationChannels.vue'
import NotificationHistory from '@/components/notifications/NotificationHistory.vue'

const appStore = useAppStore()
const toast = useToast()

const loading = ref(false)
const prefs = ref<NotificationPreferences | null>(null)
const channels = ref<NotificationChannel[]>([])
const history = ref<NotificationLogEntry[]>([])

const prefsRef = ref<InstanceType<typeof NotificationPreferencesSection>>()
const channelsRef = ref<InstanceType<typeof NotificationChannels>>()

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

onMounted(fetchAll)
watch(() => appStore.currentAppId, () => {
  prefsRef.value?.resetEditing()
  channelsRef.value?.resetForm()
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
      <div class="border border-border rounded-lg bg-bg-surface p-6 space-y-4">
        <div v-for="n in 3" :key="n" class="flex items-baseline justify-between py-2">
          <SkeletonBlock width="5rem" height="0.75rem" />
          <SkeletonBlock width="10rem" height="0.75rem" />
        </div>
      </div>
    </div>

    <div v-else class="space-y-8">
      <NotificationPreferencesSection
        v-if="appStore.currentAppId"
        ref="prefsRef"
        :app-id="appStore.currentAppId"
        :preferences="prefs"
        @update:preferences="prefs = $event"
      />

      <NotificationChannels
        v-if="appStore.currentAppId"
        ref="channelsRef"
        :app-id="appStore.currentAppId"
        :channels="channels"
        @update:channels="channels = $event"
      />

      <NotificationHistory :entries="history" />
    </div>
  </div>
</template>
