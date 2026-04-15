import client from './client'
import type {
  NotificationPreferences,
  UpdatePreferencesPayload,
  NotificationChannel,
  CreateChannelPayload,
  UpdateChannelPayload,
  NotificationLogEntry,
} from '@/types/notification'

export async function getNotificationPreferences(appId: string): Promise<NotificationPreferences> {
  const { data } = await client.get<NotificationPreferences>(`/apps/${appId}/notifications/preferences`)
  return data
}

export async function updateNotificationPreferences(appId: string, prefs: UpdatePreferencesPayload): Promise<NotificationPreferences> {
  const { data } = await client.put<NotificationPreferences>(`/apps/${appId}/notifications/preferences`, prefs)
  return data
}

export async function listNotificationChannels(appId: string): Promise<NotificationChannel[]> {
  const { data } = await client.get<NotificationChannel[]>(`/apps/${appId}/notifications/channels`)
  return data
}

export async function createNotificationChannel(appId: string, channel: CreateChannelPayload): Promise<NotificationChannel> {
  const { data } = await client.post<NotificationChannel>(`/apps/${appId}/notifications/channels`, channel)
  return data
}

export async function updateNotificationChannel(appId: string, channelId: string, channel: UpdateChannelPayload): Promise<NotificationChannel> {
  const { data } = await client.put<NotificationChannel>(`/apps/${appId}/notifications/channels/${channelId}`, channel)
  return data
}

export async function deleteNotificationChannel(appId: string, channelId: string): Promise<void> {
  await client.delete(`/apps/${appId}/notifications/channels/${channelId}`)
}

export async function testNotificationChannel(appId: string, channelId: string): Promise<{ status: string }> {
  const { data } = await client.post<{ status: string }>(`/apps/${appId}/notifications/channels/${channelId}/test`)
  return data
}

export async function listNotificationHistory(appId: string, limit = 20, offset = 0): Promise<NotificationLogEntry[]> {
  const { data } = await client.get<NotificationLogEntry[]>(`/apps/${appId}/notifications/history`, {
    params: { limit, offset },
  })
  return data
}
