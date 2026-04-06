import { defineAsyncComponent, type Component } from 'vue'

export interface FlowStep {
  id: string
  label: string
  component: Component
}

export interface PlatformFlow {
  id: string
  name: string
  icon: string
  description: string
  category: 'log_source' | 'database' | 'generic'
  connectorType: string
  direction: 'one_way' | 'two_way'
  available: boolean
  steps: FlowStep[]
}

export interface WizardState {
  name: string
  config: Record<string, unknown>
}

// Canonical list of Supabase log tables available for polling.
// Shared between ConnectionForm (edit) and StepSupabaseTables (wizard).
export const supabaseLogTables = [
  { value: 'postgres_logs', label: 'Database queries and errors' },
  { value: 'auth_logs', label: 'Authentication events' },
  { value: 'edge_logs', label: 'API gateway requests' },
  { value: 'function_logs', label: 'Edge Function output' },
  { value: 'storage_logs', label: 'Object storage operations' },
  { value: 'realtime_logs', label: 'WebSocket connections' },
] as const

// Lazy-load step components to keep the flows file lightweight.
const StepName = defineAsyncComponent(() => import('./steps/StepName.vue'))
const StepTest = defineAsyncComponent(() => import('./steps/StepTest.vue'))
const StepSupabaseAuth = defineAsyncComponent(() => import('./steps/StepSupabaseAuth.vue'))
const StepSupabaseTables = defineAsyncComponent(() => import('./steps/StepSupabaseTables.vue'))
const StepPostgresConfig = defineAsyncComponent(() => import('./steps/StepPostgresConfig.vue'))
const StepWebhookSetup = defineAsyncComponent(() => import('./steps/StepWebhookSetup.vue'))
const StepGitHubInstall = defineAsyncComponent(() => import('./steps/StepGitHubInstall.vue'))
const StepSyslogConfig = defineAsyncComponent(() => import('./steps/StepSyslogConfig.vue'))
const StepOTLPSetup = defineAsyncComponent(() => import('./steps/StepOTLPSetup.vue'))

export const flows: PlatformFlow[] = [
  // — Log Sources —
  {
    id: 'supabase',
    name: 'Supabase',
    icon: 'SB',
    description: 'Database logs, auth events, edge functions',
    category: 'log_source',
    connectorType: 'supabase',
    direction: 'one_way',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'auth', label: 'Auth', component: StepSupabaseAuth },
      { id: 'tables', label: 'Tables', component: StepSupabaseTables },
      { id: 'test', label: 'Test', component: StepTest },
    ],
  },
  {
    id: 'webhook_logs',
    name: 'Webhook',
    icon: 'WH',
    description: 'HTTP endpoint for log ingestion',
    category: 'log_source',
    connectorType: 'webhook_logs',
    direction: 'one_way',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'setup', label: 'Setup', component: StepWebhookSetup },
    ],
  },
  // — Databases —
  {
    id: 'postgres',
    name: 'PostgreSQL',
    icon: 'PG',
    description: 'Query your database during investigations',
    category: 'database',
    connectorType: 'postgres',
    direction: 'two_way',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'config', label: 'Config', component: StepPostgresConfig },
      { id: 'test', label: 'Test', component: StepTest },
    ],
  },
  // — Generic —
  {
    id: 'github',
    name: 'GitHub',
    icon: 'GH',
    description: 'Repository access for code investigation',
    category: 'generic',
    connectorType: 'github',
    direction: 'one_way',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'install', label: 'Install', component: StepGitHubInstall },
    ],
  },
  {
    id: 'otlp',
    name: 'OpenTelemetry',
    icon: 'OT',
    description: 'OTLP HTTP endpoint for logs',
    category: 'log_source',
    connectorType: 'otlp',
    direction: 'one_way',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'setup', label: 'Setup', component: StepOTLPSetup },
    ],
  },
  // — Coming soon —
  {
    id: 'datadog',
    name: 'Datadog',
    icon: 'DD',
    description: 'Ingest logs from Datadog',
    category: 'log_source',
    connectorType: 'datadog',
    direction: 'one_way',
    available: false,
    steps: [],
  },
  {
    id: 'syslog',
    name: 'Syslog',
    icon: 'SL',
    description: 'TCP/TLS listener for syslog protocol',
    category: 'log_source',
    connectorType: 'syslog',
    direction: 'one_way',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'config', label: 'Config', component: StepSyslogConfig },
      { id: 'test', label: 'Test', component: StepTest },
    ],
  },
  {
    id: 'mysql',
    name: 'MySQL',
    icon: 'MY',
    description: 'Query MySQL databases during investigations',
    category: 'database',
    connectorType: 'mysql',
    direction: 'two_way',
    available: false,
    steps: [],
  },
]

export function getFlowById(id: string): PlatformFlow | undefined {
  return flows.find(f => f.id === id)
}
