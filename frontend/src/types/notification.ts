export interface NotificationPreferences {
  app_id: string
  enabled: boolean
  severity_threshold: 'info' | 'warning' | 'error' | 'critical'
  cooldown_minutes: number
  created_at?: string
  updated_at?: string
}

export interface UpdatePreferencesPayload {
  enabled: boolean
  severity_threshold: 'info' | 'warning' | 'error' | 'critical'
  cooldown_minutes: number
}

export type NotificationChannelType = 'email' | 'slack' | 'discord'

export interface NotificationChannel {
  id: string
  app_id: string
  type: NotificationChannelType
  name: string
  config: EmailConfig | SlackConfig | DiscordConfig
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface EmailConfig {
  recipients: string[]
}

export interface SlackConfig {
  webhook_url: string
}

export interface DiscordConfig {
  webhook_url: string
}

export type ChannelConfig = EmailConfig | SlackConfig | DiscordConfig

export interface CreateChannelPayload {
  type: NotificationChannelType
  name: string
  config: ChannelConfig
  enabled?: boolean
}

export interface UpdateChannelPayload {
  name: string
  config: ChannelConfig
  enabled: boolean
}

export interface NotificationLogEntry {
  id: string
  app_id: string
  channel_id: string
  channel_type: string
  channel_name: string
  agent_log_id: string | null
  severity: string
  summary: string
  status: 'pending' | 'sent' | 'failed'
  error_message: string | null
  sent_at: string | null
  created_at: string
}
