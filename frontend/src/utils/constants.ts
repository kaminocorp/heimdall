export const APP_NAME = 'Heimdall'

export const CONNECTION_TYPES = ['postgres', 'webhook_logs', 'syslog', 'github'] as const

export const SEVERITY_LEVELS = ['info', 'warning', 'critical'] as const

export const INVESTIGATION_STATUSES = ['open', 'investigating', 'resolved', 'dismissed'] as const
