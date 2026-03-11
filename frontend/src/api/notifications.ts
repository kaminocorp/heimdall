import client from './client'
import type {
  NotificationPreferences,
  UpdatePreferencesPayload,
  NotificationChannel,
  CreateChannelPayload,
  UpdateChannelPayload,
  NotificationLogEntry,
} from '@/types/notification'

export function getNotificationPreferences(appId: string) {
  return client.get<NotificationPreferences>(`/apps/${appId}/notifications/preferences`).then(r => r.data)
}

export function updateNotificationPreferences(appId: string, prefs: UpdatePreferencesPayload) {
  return client.put<NotificationPreferences>(`/apps/${appId}/notifications/preferences`, prefs).then(r => r.data)
}

export function listNotificationChannels(appId: string) {
  return client.get<NotificationChannel[]>(`/apps/${appId}/notifications/channels`).then(r => r.data)
}

export function createNotificationChannel(appId: string, channel: CreateChannelPayload) {
  return client.post<NotificationChannel>(`/apps/${appId}/notifications/channels`, channel).then(r => r.data)
}

export function updateNotificationChannel(appId: string, channelId: string, channel: UpdateChannelPayload) {
  return client.put<NotificationChannel>(`/apps/${appId}/notifications/channels/${channelId}`, channel).then(r => r.data)
}

export function deleteNotificationChannel(appId: string, channelId: string) {
  return client.delete(`/apps/${appId}/notifications/channels/${channelId}`)
}

export function testNotificationChannel(appId: string, channelId: string) {
  return client.post<{ status: string }>(`/apps/${appId}/notifications/channels/${channelId}/test`).then(r => r.data)
}

export function listNotificationHistory(appId: string, limit = 20, offset = 0) {
  return client.get<NotificationLogEntry[]>(`/apps/${appId}/notifications/history`, {
    params: { limit, offset },
  }).then(r => r.data)
}
