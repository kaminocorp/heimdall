export interface Organization {
  id: string
  name: string
  slug: string
  created_at: string
  updated_at: string
}

export interface Application {
  id: string
  org_id: string
  name: string
  status: 'active' | 'paused' | 'archived'
  created_at: string
  updated_at: string
}

export interface AppAgentConfig {
  app_id: string
  model: string
  mode: 'continuous' | 'periodic' | 'off'
  schedule_interval_secs: number
  system_prompt_override: string | null
  created_at: string
  updated_at: string
}

export interface MonitoringStatus {
  mode: 'continuous' | 'periodic' | 'off'
  schedule_interval_secs: number
  last_monitored_at: string | null
}

export interface OnboardingPayload {
  org_name: string
  org_slug: string
  app_name: string
}

export interface OnboardingResponse {
  organization: Organization
  application: Application
}
