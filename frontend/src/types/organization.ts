export interface Organization {
  id: string
  name: string
  slug: string
  created_at: string
  updated_at: string
  role?: 'owner' | 'admin' | 'member'
}

export type OrgMemberRole = 'owner' | 'admin' | 'member'

export interface OrgMember {
  user_id: string
  email: string
  role: OrgMemberRole
  created_at: string
}

export interface OrganizationWithRole extends Organization {
  role: OrgMemberRole
}

export interface Application {
  id: string
  org_id: string
  name: string
  status: 'active' | 'paused' | 'archived'
  created_at: string
  updated_at: string
}

// The enriched row returned by GET /api/apps?include=counts. Mirrors the
// backend's `ListApplicationsByOrgWithCountsRow` struct. Used by the
// Settings → Applications section to render per-app summaries without
// firing an N+1 query per row.
export interface ApplicationWithCounts extends Application {
  connection_count: number
  schedule_count: number
}

export interface AppAgentConfig {
  app_id: string
  model: string
  provider: 'anthropic' | 'openrouter'
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
