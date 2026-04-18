import { defineAsyncComponent, type Component } from 'vue'

export interface FlowStep {
  id: string
  label: string
  component: Component
}

export type ConnectionCategory = 'ingestion' | 'enrichment' | 'outbound'

export interface PlatformFlow {
  id: string
  name: string
  icon: string
  description: string
  section: 'platform_log' | 'direct_protocol' | 'agent_tool' | 'outbound'
  subtitle: string
  connectorType: string
  direction: 'one_way' | 'two_way'
  available: boolean
  steps: FlowStep[]
}

/** Map a connection type string to its visual category on the Connections page. */
export function typeToCategory(type: string): ConnectionCategory {
  switch (type) {
    case 'postgres':
    case 'mysql':
    case 'github':
      return 'enrichment'
    case 'slack':
    case 'telegram':
    case 'linear':
    case 'trajan':
      return 'outbound'
    default:
      return 'ingestion'
  }
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
const StepFlyioMode = defineAsyncComponent(() => import('./steps/StepFlyioMode.vue'))
const StepFlyioAuth = defineAsyncComponent(() => import('./steps/StepFlyioAuth.vue'))
const StepFlyioDrainSetup = defineAsyncComponent(() => import('./steps/StepFlyioDrainSetup.vue'))
const StepConnectionScope = defineAsyncComponent(() => import('./steps/StepConnectionScope.vue'))

export const flows: PlatformFlow[] = [
  // ── Platform log sources ─────────────────────────────
  // "Where do your logs come from?" — platform-first selection
  {
    id: 'supabase',
    name: 'Supabase',
    icon: 'SB',
    description: 'Database logs, auth events, edge functions',
    section: 'platform_log',
    subtitle: 'Log polling + API',
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
    id: 'flyio',
    name: 'Fly.io',
    icon: 'FI',
    description: 'Ship logs from Fly.io apps via log drain or API polling',
    section: 'platform_log',
    subtitle: 'Log Drain or API Polling',
    connectorType: 'flyio',
    direction: 'one_way',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'flyio_mode', label: 'Mode', component: StepFlyioMode },
      // Scope step is only shown in drain mode — polling is intrinsically
      // app-scoped (one Fly app → one connection). Filtered out by
      // getFlyioSteps when mode === 'polling'.
      { id: 'scope', label: 'Scope', component: StepConnectionScope },
      { id: 'flyio_auth', label: 'Auth', component: StepFlyioAuth },
      { id: 'flyio_drain', label: 'Setup', component: StepFlyioDrainSetup },
      { id: 'test', label: 'Test', component: StepTest },
    ],
  },
  {
    id: 'vercel',
    name: 'Vercel',
    icon: 'VC',
    description: 'Ship logs from Vercel deployments via log drain',
    section: 'platform_log',
    subtitle: 'via Log Drain',
    connectorType: 'webhook_logs',
    direction: 'one_way',
    available: false,
    steps: [],
  },
  {
    id: 'render',
    name: 'Render',
    icon: 'RN',
    description: 'Ship logs from Render services via syslog drain',
    section: 'platform_log',
    subtitle: 'via Syslog drain',
    connectorType: 'syslog',
    direction: 'one_way',
    available: false,
    steps: [],
  },
  {
    id: 'railway',
    name: 'Railway',
    icon: 'RW',
    description: 'Ship logs from Railway services via HTTP drain',
    section: 'platform_log',
    subtitle: 'via HTTP Log Drain',
    connectorType: 'webhook_logs',
    direction: 'one_way',
    available: false,
    steps: [],
  },
  {
    id: 'heroku',
    name: 'Heroku',
    icon: 'HK',
    description: 'Ship logs from Heroku dynos via log drain',
    section: 'platform_log',
    subtitle: 'via Syslog drain',
    connectorType: 'syslog',
    direction: 'one_way',
    available: false,
    steps: [],
  },
  {
    id: 'aws',
    name: 'AWS',
    icon: 'AW',
    description: 'Ship logs from AWS services via CloudWatch',
    section: 'platform_log',
    subtitle: 'via CloudWatch + OTLP',
    connectorType: 'otlp',
    direction: 'one_way',
    available: false,
    steps: [],
  },
  {
    id: 'digitalocean',
    name: 'DigitalOcean',
    icon: 'DO',
    description: 'Ship logs from DigitalOcean apps via log forwarding',
    section: 'platform_log',
    subtitle: 'via Log Forwarding',
    connectorType: 'syslog',
    direction: 'one_way',
    available: false,
    steps: [],
  },

  // ── Direct protocols ─────────────────────────────────
  // "Or connect directly" — for engineers who know the protocol
  {
    id: 'webhook_logs',
    name: 'Webhook',
    icon: 'WH',
    description: 'HTTP endpoint for log ingestion',
    section: 'direct_protocol',
    subtitle: 'HTTP endpoint',
    connectorType: 'webhook_logs',
    direction: 'one_way',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'scope', label: 'Scope', component: StepConnectionScope },
      { id: 'setup', label: 'Setup', component: StepWebhookSetup },
    ],
  },
  {
    id: 'syslog',
    name: 'Syslog',
    icon: 'SL',
    description: 'TCP/TLS listener for syslog protocol',
    section: 'direct_protocol',
    subtitle: 'TCP / TLS listener',
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
    id: 'otlp',
    name: 'OpenTelemetry',
    icon: 'OT',
    description: 'OTLP HTTP endpoint for logs',
    section: 'direct_protocol',
    subtitle: 'OTLP HTTP endpoint',
    connectorType: 'otlp',
    direction: 'one_way',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'setup', label: 'Setup', component: StepOTLPSetup },
    ],
  },

  // ── Agent investigation tools ────────────────────────
  // "Give the agent investigation tools" — on-demand query access
  {
    id: 'postgres',
    name: 'PostgreSQL',
    icon: 'PG',
    description: 'Query your database during investigations',
    section: 'agent_tool',
    subtitle: 'Query during investigations',
    connectorType: 'postgres',
    direction: 'two_way',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'config', label: 'Config', component: StepPostgresConfig },
      { id: 'test', label: 'Test', component: StepTest },
    ],
  },
  {
    id: 'github',
    name: 'GitHub',
    icon: 'GH',
    description: 'Repository access for code investigation',
    section: 'agent_tool',
    subtitle: 'Code search + context',
    connectorType: 'github',
    direction: 'one_way',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'install', label: 'Install', component: StepGitHubInstall },
    ],
  },
  {
    id: 'mysql',
    name: 'MySQL',
    icon: 'MY',
    description: 'Query MySQL databases during investigations',
    section: 'agent_tool',
    subtitle: 'Query during investigations',
    connectorType: 'mysql',
    direction: 'two_way',
    available: false,
    steps: [],
  },

  // ── Outbound channels ──────────────────────────────────
  // "Where should the agent send alerts and reports?"
  {
    id: 'slack',
    name: 'Slack',
    icon: 'SK',
    description: 'Send alerts and reports to Slack channels',
    section: 'outbound',
    subtitle: 'Channel notifications',
    connectorType: 'slack',
    direction: 'one_way',
    available: false,
    steps: [],
  },
  {
    id: 'telegram',
    name: 'Telegram',
    icon: 'TG',
    description: 'Send alerts to Telegram chats and groups',
    section: 'outbound',
    subtitle: 'Bot notifications',
    connectorType: 'telegram',
    direction: 'one_way',
    available: false,
    steps: [],
  },
  {
    id: 'linear',
    name: 'Linear',
    icon: 'LN',
    description: 'Create issues from incidents automatically',
    section: 'outbound',
    subtitle: 'Issue tracking',
    connectorType: 'linear',
    direction: 'one_way',
    available: false,
    steps: [],
  },
  {
    id: 'trajan',
    name: 'Trajan',
    icon: 'TJ',
    description: 'Create tickets in Trajan project management',
    section: 'outbound',
    subtitle: 'Ticket management',
    connectorType: 'trajan',
    direction: 'one_way',
    available: false,
    steps: [],
  },
]

export function getFlowById(id: string): PlatformFlow | undefined {
  return flows.find(f => f.id === id)
}

/**
 * Returns the visible steps for a Fly.io flow based on mode selection.
 * - Drain mode:    Name → Mode → Scope → Drain Setup
 *                  (connection created before Drain Setup)
 * - Polling mode:  Name → Mode → Auth (credentials) → Test
 *                  (no Scope step — polling is intrinsically app-scoped)
 */
export function getFlyioSteps(flow: PlatformFlow, mode: 'drain' | 'polling'): FlowStep[] {
  if (flow.id !== 'flyio') return flow.steps
  return flow.steps.filter(s => {
    if (mode === 'drain') return s.id !== 'flyio_auth' && s.id !== 'test'
    // Polling mode: strip Scope + Drain Setup; it's a single-app API client.
    return s.id !== 'flyio_drain' && s.id !== 'scope'
  })
}

/**
 * Returns the effective connector type for a Fly.io flow based on mode.
 * Drain creates a webhook_logs connection; polling creates a flyio connection.
 */
export function getFlyioConnectorType(mode: 'drain' | 'polling'): string {
  return mode === 'drain' ? 'webhook_logs' : 'flyio'
}
