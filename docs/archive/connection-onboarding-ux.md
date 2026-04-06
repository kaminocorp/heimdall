# Connection Onboarding UX — Guided Setup Flow

## Problem

The current connection creation form is a generic form with a type dropdown. The user selects "postgres", fills in host/port/user/password fields, and hopes they got it right. This works for developers who already know what they're doing, but it has several problems:

1. **No guidance** — a Supabase user doesn't know whether to use their direct connection string, pooled connection, or the Management API. They just see "PostgreSQL" and a host field.
2. **No platform awareness** — whether you're connecting Supabase, Neon, AWS RDS, or a self-hosted Postgres, you get the same form. But the setup steps are different for each.
3. **No instruction** — for push-based methods (webhook, syslog, OTLP), the user needs to configure something on the *other* platform too. The current form gives no instructions for that.
4. **Flat structure** — all connection types (postgres, webhook_logs, syslog, github) are in a single dropdown. As we add more types (supabase, OTLP, platform API pollers), this list becomes unwieldy and the distinction between "log source" and "query target" gets lost.

## Vision

When a user clicks "+ Add Connection", they enter a **guided wizard** — a multi-step modal that:

1. Shows them a visual grid of platforms/methods to choose from
2. Walks them through platform-specific setup steps with copy-pasteable instructions
3. Tests the connection before finishing
4. Gets them to a working integration in under 2 minutes

Think of it like Stripe's onboarding or Vercel's "Add Integration" flow — visual, guided, and impossible to get wrong.

## Design

### Entry Point

The "+ New Connection" button on `ConnectionsPage.vue` opens a **full-width modal** (not an in-page form). This modal contains the entire wizard flow. The existing in-page `ConnectionForm` is preserved for editing existing connections — editing doesn't need the wizard.

### Step 1: Choose Your Source

A visual grid of platform/method cards. Each card shows a platform logo placeholder (monochrome icon to match the design system), name, and one-line description.

```
┌─────────────────────────────────────────────────────────────┐
│  Add Connection                                        [×]  │
│─────────────────────────────────────────────────────────────│
│                                                             │
│  What are you connecting?                                   │
│                                                             │
│  ┌─── LOG SOURCES ──────────────────────────────────────┐   │
│  │                                                       │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐│  │
│  │  │ Supabase │ │  Vercel  │ │  Fly.io  │ │  Render  ││  │
│  │  │ DB logs, │ │ Deploys, │ │ App logs │ │ App logs ││  │
│  │  │ auth,    │ │ function │ │ via API  │ │ via      ││  │
│  │  │ edge fn  │ │ logs     │ │          │ │ syslog   ││  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘│  │
│  │                                                       │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐│  │
│  │  │  Heroku  │ │  Railway │ │  AWS     │ │  GCP     ││  │
│  │  │ Log      │ │ App logs │ │CloudWatch│ │ Cloud    ││  │
│  │  │ drains   │ │          │ │ logs     │ │ Logging  ││  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘│  │
│  │                                                       │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌─── DATABASES ────────────────────────────────────────┐   │
│  │                                                       │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐│  │
│  │  │PostgreSQL│ │   Neon   │ │Planet-   │ │ MongoDB  ││  │
│  │  │ Direct   │ │ Postgres │ │Scale     │ │ Atlas    ││  │
│  │  │ connect  │ │ + OTLP   │ │          │ │          ││  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘│  │
│  │                                                       │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌─── GENERIC ──────────────────────────────────────────┐   │
│  │                                                       │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐│  │
│  │  │ Webhook  │ │  Syslog  │ │  OTLP    │ │  GitHub  ││  │
│  │  │ HTTP     │ │ TCP/TLS  │ │ OpenTel  │ │ Codebase ││  │
│  │  │ endpoint │ │ listener │ │ receiver │ │ access   ││  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘│  │
│  │                                                       │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Key design decisions:**
- Cards are grouped by category: Log Sources, Databases, Generic
- Each card is clickable and shows a hover state with `border-accent/50`
- Platform-specific cards (Supabase, Vercel, etc.) map to underlying connector types but present platform-specific setup flows
- Generic cards (Webhook, Syslog, OTLP) are for advanced users or platforms we don't explicitly support
- Cards that aren't yet implemented show a subtle "Coming soon" badge — they're visible but not clickable, so users can see the roadmap

**Mapping cards to connector types:**

| Card | Underlying Type | Setup Flow |
|------|----------------|------------|
| Supabase | `supabase` (new) | API poller with PAT + project ref |
| Vercel | `webhook_logs` | Webhook setup with Vercel-specific instructions |
| Fly.io | `flyio` (new, API poller) | API token + app name |
| Render | `syslog` | Syslog with Render-specific instructions |
| Heroku | `webhook_logs` or `syslog` | Choice between HTTPS drain and syslog drain |
| Railway | `webhook_logs` | Webhook with Railway-specific instructions |
| AWS CloudWatch | `webhook_logs` | Firehose → HTTP instructions |
| GCP Cloud Logging | `webhook_logs` | Pub/Sub push subscription instructions |
| PostgreSQL | `postgres` | Direct connection form |
| Neon | `postgres` + optional OTLP | Connection string + OTLP instructions |
| PlanetScale | `planetscale` (new, API poller) | API token |
| MongoDB Atlas | `mongodb_atlas` (new, API poller) | API key pair |
| Webhook | `webhook_logs` | Generic, no platform-specific instructions |
| Syslog | `syslog` | Generic listener setup |
| OTLP | `otlp` (new) | Generic OTLP endpoint info |
| GitHub | `github` | Existing OAuth flow |

### Step 2: Platform-Specific Guided Setup

After selecting a platform, the modal transitions to a step-by-step setup flow. Each platform has its own sequence of steps. The modal shows a **step indicator** at the top and **Back / Continue** buttons at the bottom.

Every step-based flow follows the same structural pattern but with platform-specific content.

#### Example: Supabase

```
Step 1 of 4 — Connection Name
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  [Supabase icon]  Supabase                                  │
│                                                             │
│  ● ─── ○ ─── ○ ─── ○                                       │
│  Name   Auth  Tables  Test                                  │
│                                                             │
│  Give this connection a name so you can identify it later.  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ e.g. Production Supabase                            │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│                                    [Back]  [Continue →]     │
└─────────────────────────────────────────────────────────────┘
```

```
Step 2 of 4 — Authentication
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  [Supabase icon]  Supabase                                  │
│                                                             │
│  ○ ─── ● ─── ○ ─── ○                                       │
│  Name   Auth  Tables  Test                                  │
│                                                             │
│  Heimdall uses the Supabase Management API to pull logs     │
│  from your project. This requires two things:               │
│                                                             │
│  PROJECT REFERENCE                                          │
│  Find this in your Supabase dashboard under                 │
│  Settings → General → Reference ID.                         │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ e.g. abcdefghijklmnopqrst                           │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  PERSONAL ACCESS TOKEN                                      │
│  Generate one at supabase.com/dashboard/account/tokens.     │
│  Heimdall only reads logs — it never modifies your project. │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ ••••••••••••••••••••••••••••••••                     │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  ┌ ℹ ─────────────────────────────────────────────────┐     │
│  │ This uses the same API that powers your Supabase    │     │
│  │ dashboard Logs Explorer. No Log Drain add-on        │     │
│  │ needed — works on all plans including Free.         │     │
│  └────────────────────────────────────────────────────┘     │
│                                                             │
│                                    [Back]  [Continue →]     │
└─────────────────────────────────────────────────────────────┘
```

```
Step 3 of 4 — Log Tables
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  [Supabase icon]  Supabase                                  │
│                                                             │
│  ○ ─── ○ ─── ● ─── ○                                       │
│  Name   Auth  Tables  Test                                  │
│                                                             │
│  Which logs should Heimdall monitor?                        │
│                                                             │
│  ☑  postgres_logs     Database queries and errors           │
│  ☑  auth_logs         Sign-ins, failures, token issues      │
│  ☐  edge_logs         API gateway request/response          │
│  ☐  function_logs     Edge Function console output          │
│  ☐  storage_logs      Object storage operations             │
│  ☐  realtime_logs     WebSocket connections                 │
│                                                             │
│  POLL INTERVAL                                              │
│  [15s] [30s] [60s]  ← preset buttons, 30s selected         │
│                                                             │
│                                    [Back]  [Continue →]     │
└─────────────────────────────────────────────────────────────┘
```

```
Step 4 of 4 — Test Connection
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  [Supabase icon]  Supabase                                  │
│                                                             │
│  ○ ─── ○ ─── ○ ─── ●                                       │
│  Name   Auth  Tables  Test                                  │
│                                                             │
│  Testing connection to your Supabase project...             │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  [pulsing dot]  Connecting to Management API...     │    │
│  │                                          2.3s       │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  — then on success: —                                       │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  [green dot]  Connected — logs are flowing          │    │
│  │  Fetched 47 log entries from postgres_logs          │    │
│  │                                          1.8s       │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│                                    [Back]  [Done ✓]         │
└─────────────────────────────────────────────────────────────┘
```

#### Example: Webhook (push-based, user configures the other side)

```
Step 1 of 3 — Connection Name
(same name input as above)
```

```
Step 2 of 3 — Endpoint & Token
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  [Webhook icon]  HTTP Webhook                               │
│                                                             │
│  ○ ─── ● ─── ○                                              │
│  Name   Setup  Test                                         │
│                                                             │
│  Heimdall has generated a webhook endpoint for you.         │
│  Configure your application or platform to send logs here.  │
│                                                             │
│  ENDPOINT                                                   │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ https://heimdall-backend.fly.dev/api/webhooks/logs  │ 📋│
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  BEARER TOKEN                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ whk_a3f8b2c1d4e5f6...                              │ 📋│
│  └─────────────────────────────────────────────────────┘    │
│  ⚠ Copy this token now — it won't be shown again.          │
│                                                             │
│  PAYLOAD FORMAT                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ POST /api/webhooks/logs                             │    │
│  │ Authorization: Bearer whk_a3f8b2c1d4e5...           │    │
│  │ Content-Type: application/json                      │    │
│  │                                                     │    │
│  │ {                                                   │    │
│  │   "source_type": "application/error",               │    │
│  │   "severity": "error",                              │    │
│  │   "payload": { ... }                                │    │
│  │ }                                                   │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│                                    [Back]  [Continue →]     │
└─────────────────────────────────────────────────────────────┘
```

#### Example: Vercel (webhook, but with platform-specific instructions)

```
Step 2 of 3 — Configure Vercel Log Drain
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  [Vercel icon]  Vercel                                      │
│                                                             │
│  ○ ─── ● ─── ○                                              │
│  Name   Setup  Test                                         │
│                                                             │
│  Set up a Log Drain in Vercel to send logs to Heimdall.     │
│                                                             │
│  1. Open your Vercel dashboard → Project Settings            │
│     → Observability → Log Drains                             │
│                                                             │
│  2. Click "Add Log Drain" and select:                        │
│     Delivery format: NDJSON                                 │
│     Environments: Production (or all)                       │
│     Sources: All (or select specific)                        │
│                                                             │
│  3. Paste this endpoint:                                     │
│     ┌──────────────────────────────────────────────────┐    │
│     │ https://heimdall-backend.fly.dev/api/webhooks/   │    │
│     │ logs                                             │ 📋│
│     └──────────────────────────────────────────────────┘    │
│                                                             │
│  4. Add this header:                                         │
│     ┌──────────────────────────────────────────────────┐    │
│     │ Authorization: Bearer whk_a3f8b2c1d4e5f6...     │ 📋│
│     └──────────────────────────────────────────────────┘    │
│                                                             │
│  5. Click "Create Log Drain" in Vercel.                      │
│                                                             │
│  ┌ ℹ ─────────────────────────────────────────────────┐     │
│  │ Vercel Log Drains require a Pro plan ($20/mo).      │     │
│  │ Logs typically start arriving within 30 seconds.    │     │
│  └────────────────────────────────────────────────────┘     │
│                                                             │
│                                    [Back]  [Continue →]     │
└─────────────────────────────────────────────────────────────┘
```

#### Example: Render (syslog)

```
Step 2 of 3 — Configure Render Log Stream
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  [Render icon]  Render                                      │
│                                                             │
│  ○ ─── ● ─── ○                                              │
│  Name   Setup  Test                                         │
│                                                             │
│  Set up a Log Stream in Render to send logs to Heimdall     │
│  via syslog.                                                │
│                                                             │
│  1. Open your Render dashboard → Account Settings            │
│     → Log Streams                                            │
│                                                             │
│  2. Click "Add Log Stream" and enter:                        │
│                                                             │
│     Endpoint:                                                │
│     ┌──────────────────────────────────────────────────┐    │
│     │ tls://heimdall-backend.fly.dev:6514             │ 📋│
│     └──────────────────────────────────────────────────┘    │
│                                                             │
│     Token (paste into the "Token" field):                    │
│     ┌──────────────────────────────────────────────────┐    │
│     │ whk_a3f8b2c1d4e5f6...                           │ 📋│
│     └──────────────────────────────────────────────────┘    │
│                                                             │
│  3. Select which services to include, then save.             │
│                                                             │
│  ┌ ℹ ─────────────────────────────────────────────────┐     │
│  │ Log Streams are available on all Render plans       │     │
│  │ including Free. Logs arrive in real-time via TLS.   │     │
│  └────────────────────────────────────────────────────┘     │
│                                                             │
│                                    [Back]  [Continue →]     │
└─────────────────────────────────────────────────────────────┘
```

#### Example: PostgreSQL (direct connection)

```
Step 2 of 3 — Connection Details
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  [Postgres icon]  PostgreSQL                                │
│                                                             │
│  ○ ─── ● ─── ○                                              │
│  Name   Config  Test                                        │
│                                                             │
│  Enter your database connection details. Heimdall connects  │
│  in read-only mode — it cannot modify your data.            │
│                                                             │
│  HOST                                                       │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ db.example.com                                      │    │
│  └─────────────────────────────────────────────────────┘    │
│  PORT                        DATABASE                       │
│  ┌────────────┐              ┌─────────────────────────┐    │
│  │ 5432       │              │ mydb                    │    │
│  └────────────┘              └─────────────────────────┘    │
│  USERNAME                    PASSWORD                       │
│  ┌────────────────────┐      ┌─────────────────────────┐    │
│  │ postgres           │      │ ••••••••                │    │
│  └────────────────────┘      └─────────────────────────┘    │
│  SSL MODE                                                   │
│  [Require ▾]                                                │
│                                                             │
│  DIRECTION                                                  │
│  ● Two-way (ingest + query) — recommended                   │
│  ○ One-way (ingest only)                                    │
│                                                             │
│                                    [Back]  [Continue →]     │
└─────────────────────────────────────────────────────────────┘
```

### Step 3: Test & Finish

Every flow ends with a test step. The test is automatic — it starts as soon as the user reaches this step. The test behavior varies by type:

| Type | What the test does |
|------|-------------------|
| `supabase` | Calls the Management API with a trivial query |
| `postgres` | Attempts a TCP connection with 5s timeout |
| `webhook_logs` | Creates the connection; shows endpoint + token; marks pass (no remote to test) |
| `syslog` | Creates the listener; marks pass (waits for first message) |
| `otlp` | Creates the receiver; marks pass (waits for first payload) |
| `github` | Tests API rate limit endpoint |

On success, the modal shows a green confirmation and a "Done" button that closes the modal and refreshes the connections list.

On failure, the modal shows the error message with a "Retry" button and a "Back" button to fix the configuration.

### Step Flow Definitions

Each platform card maps to a flow definition — an array of step objects:

```typescript
interface WizardStep {
  id: string
  label: string                          // shown in step indicator
  component: Component                   // Vue component to render
  props?: Record<string, unknown>        // static props
  validate?: (state: WizardState) => boolean  // can the user proceed?
}

interface PlatformFlow {
  id: string
  name: string
  icon: string                           // icon component or SVG path
  description: string                    // one-line subtitle on the card
  category: 'log_source' | 'database' | 'generic'
  connectorType: string                  // underlying connection type for the API
  available: boolean                     // false = "coming soon" badge
  steps: WizardStep[]
}
```

This lets us define each platform's flow declaratively. The wizard modal renders steps generically — it doesn't know about Supabase vs. Vercel, it just renders the current step's component with the current state.

---

## Component Architecture

```
ConnectionsPage.vue
  ├─ ConnectionWizard.vue (new — the modal)
  │  ├─ WizardStepIndicator.vue (new — the ● ─── ○ ─── ○ bar)
  │  ├─ PlatformGrid.vue (new — Step 1 card grid)
  │  │  └─ PlatformCard.vue (new — individual clickable card)
  │  ├─ Step components (new, one per step type):
  │  │  ├─ StepName.vue — connection name input (shared across all flows)
  │  │  ├─ StepSupabaseAuth.vue — project ref + PAT inputs
  │  │  ├─ StepSupabaseTables.vue — table checkboxes + poll interval
  │  │  ├─ StepPostgresConfig.vue — host/port/db/user/pass/ssl
  │  │  ├─ StepWebhookSetup.vue — shows endpoint + token + payload format
  │  │  ├─ StepPlatformInstructions.vue — platform-specific setup guide (Vercel, Render, Heroku, etc.)
  │  │  ├─ StepGitHubInstall.vue — wraps existing OAuth redirect flow
  │  │  └─ StepTest.vue — connection test (reuses ConnectionTestModal logic)
  │  └─ Flow definitions:
  │     └─ flows.ts — exports PlatformFlow[] array
  ├─ ConnectionForm.vue (existing — kept for editing)
  ├─ ConnectionList.vue (existing)
  └─ ConnectionTestModal.vue (existing — kept for ping/retest from connection cards)
```

### Shared Step Components

Most steps are reusable across flows:

| Component | Used by |
|-----------|---------|
| `StepName` | Every flow (always step 1) |
| `StepTest` | Every flow (always last step) |
| `StepPlatformInstructions` | Vercel, Render, Heroku, Railway, AWS, GCP — parameterised with markdown-like instruction content from the flow definition |
| `StepWebhookSetup` | Webhook, Vercel, Heroku, Railway (shows endpoint + token after connection is created) |
| `StepPostgresConfig` | PostgreSQL, Neon (direct connection fields) |

Platform-specific steps (Supabase auth, Supabase tables) are one-off components, but they're small — just a form with a few inputs.

### Flow Definitions File

`flows.ts` is the single source of truth for all platform flows. Adding a new platform means adding a new entry to this array — no new routes, no modal changes.

```typescript
// Simplified example
export const platformFlows: PlatformFlow[] = [
  {
    id: 'supabase',
    name: 'Supabase',
    icon: 'supabase',
    description: 'Database logs, auth events, edge functions',
    category: 'log_source',
    connectorType: 'supabase',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'auth', label: 'Auth', component: StepSupabaseAuth },
      { id: 'tables', label: 'Tables', component: StepSupabaseTables },
      { id: 'test', label: 'Test', component: StepTest },
    ],
  },
  {
    id: 'vercel',
    name: 'Vercel',
    icon: 'vercel',
    description: 'Deploy logs, serverless function traces',
    category: 'log_source',
    connectorType: 'webhook_logs',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      {
        id: 'setup',
        label: 'Setup',
        component: StepPlatformInstructions,
        props: { platform: 'vercel' },
      },
      { id: 'test', label: 'Test', component: StepTest },
    ],
  },
  {
    id: 'postgres',
    name: 'PostgreSQL',
    icon: 'postgres',
    description: 'Direct database connection for agent queries',
    category: 'database',
    connectorType: 'postgres',
    available: true,
    steps: [
      { id: 'name', label: 'Name', component: StepName },
      { id: 'config', label: 'Config', component: StepPostgresConfig },
      { id: 'test', label: 'Test', component: StepTest },
    ],
  },
  // ... more flows
]
```

---

## Wizard State Management

The wizard uses a local reactive state object (not a Pinia store — it's ephemeral, scoped to the modal lifecycle):

```typescript
interface WizardState {
  // Navigation
  selectedPlatform: PlatformFlow | null
  currentStepIndex: number

  // Collected data (accumulated across steps)
  name: string
  connectorType: string
  direction: 'one_way' | 'two_way'
  config: Record<string, unknown>

  // Step-specific transient state
  testing: boolean
  testResult: { success: boolean; message: string } | null
  createdConnection: Connection | null   // set after API call
}
```

Each step component receives `state` as a prop and emits `update:state` to modify it. The wizard shell handles navigation (Back/Continue), validation (calling each step's `validate` function), and the final API call.

### When Does the API Call Happen?

For **pull-based** connectors (Supabase, Fly.io API pollers, Postgres), the connection is created when the user clicks "Continue" on the last config step (just before the test step). The test step then tests the already-created connection.

For **push-based** connectors (Webhook, Syslog, OTLP), the connection is created during the setup step — because the setup step needs to display the generated token/endpoint to the user. The test step then waits for the first incoming message (or just confirms the connection was created successfully).

For **GitHub**, the connection creation happens on the GitHub side via OAuth redirect. The wizard handles the return URL and picks up where it left off.

---

## Styling

The wizard modal follows the existing modal pattern from `ConnectionTestModal.vue`:

- **Backdrop**: `fixed inset-0 z-50 bg-black/60 backdrop-blur-sm`
- **Modal**: `w-full max-w-2xl mx-4 border border-border rounded-lg bg-bg-elevated shadow-2xl`
- **Header**: platform icon + name, step indicator, close button
- **Body**: step content area with consistent padding (`px-6 py-5`)
- **Footer**: `border-t border-border` with Back/Continue buttons using existing button styles

Platform cards use:
- `border border-border rounded-lg bg-bg-surface p-4 cursor-pointer`
- Hover: `hover:border-accent/50 hover:bg-bg-surface-hover transition-colors`
- Selected: `border-accent bg-accent-subtle`
- Coming soon: `opacity-40 cursor-not-allowed` with a small badge

Step indicator dots:
- Active: `bg-accent` with a subtle glow
- Completed: `bg-accent/60`
- Future: `bg-border`
- Connector lines: `h-px bg-border` between dots

---

## Copy-to-Clipboard

Several steps show values the user needs to copy (endpoint URLs, tokens, code snippets). Each copyable field has a clipboard button that:

1. Copies to clipboard via `navigator.clipboard.writeText()`
2. Shows a brief "Copied" tooltip or changes the icon to a checkmark for 2 seconds
3. Uses the same monospace styling as code blocks: `bg-bg-primary border border-border rounded px-3 py-2 font-mono text-xs`

Implement as a small `CopyableField.vue` component:

```vue
<CopyableField :value="webhookUrl" label="Endpoint" />
```

---

## Platform Icons

For the MVP, use simple monochrome SVG icons or single-letter/abbreviation badges that match the design system. Full-color logos can come later. Options:

1. **Letter badges**: "SB" for Supabase, "V" for Vercel, "PG" for PostgreSQL — styled as `w-8 h-8 rounded bg-accent-subtle border border-accent-border text-accent font-mono text-xs font-bold flex items-center justify-center`
2. **Simple SVGs**: monochrome outlines from Simple Icons or similar, rendered in `text-text-secondary`

Letter badges are faster to implement and fit the brutalist aesthetic. Platform SVGs can replace them later without any structural changes.

---

## Migration from Current UX

The transition should be backward-compatible:

1. **New connections**: always use the wizard modal
2. **Editing connections**: keep the existing in-page `ConnectionForm` — editing doesn't need guidance, the user already knows what they configured
3. **Ping/retest**: keep the existing `ConnectionTestModal` — it's triggered from connection cards, not from the wizard
4. **GitHub post-install redirect**: the wizard handles this — when returning with `?github=installed`, the wizard opens directly to the test step for the GitHub flow

The old `ConnectionForm` type dropdown (postgres, webhook_logs, syslog, github) is no longer the entry point for creation, but the form component itself is reused for editing.

---

## Implementation Plan

### Phase 1: Wizard Shell + Platform Grid
1. Create `ConnectionWizard.vue` — modal shell with step navigation, state management, Back/Continue/Done buttons
2. Create `WizardStepIndicator.vue` — the dot-line progress indicator
3. Create `PlatformGrid.vue` + `PlatformCard.vue` — the card selection grid
4. Create `flows.ts` — flow definitions for existing types (postgres, webhook_logs, github)
5. Wire "+ New Connection" button to open the wizard instead of showing the inline form
6. Create `CopyableField.vue` — reusable copy-to-clipboard component

### Phase 2: Step Components for Existing Types
7. Create `StepName.vue` — shared name input step
8. Create `StepPostgresConfig.vue` — extract config fields from existing `ConnectionForm`
9. Create `StepWebhookSetup.vue` — show generated endpoint + token + payload format
10. Create `StepTest.vue` — extract test logic from `ConnectionTestModal`
11. Create `StepGitHubInstall.vue` — wrap existing OAuth redirect
12. Create `StepPlatformInstructions.vue` — parameterised instruction renderer

### Phase 3: New Platform Flows
13. Create `StepSupabaseAuth.vue` — project ref + PAT inputs
14. Create `StepSupabaseTables.vue` — table checkboxes + poll interval
15. Add Vercel, Render, Heroku flow definitions to `flows.ts` (these use `StepPlatformInstructions` with different content — no new components needed)
16. Mark unimplemented platforms as "coming soon" in the grid

### Phase 4: Polish
17. Add keyboard navigation to the platform grid (arrow keys)
18. Add transition animations between steps (slide left/right)
19. Add "coming soon" badge and tooltip to unavailable platforms
20. Test all flows end-to-end

---

## Questions

1. **Backend URL in instructions**: The webhook setup step shows the Heimdall backend URL (e.g. `https://heimdall-backend.fly.dev/api/webhooks/logs`). Should this be configurable via env var / API, or is it safe to hardcode for now? Self-hosted users would need a different URL.

2. **Token visibility**: The current webhook flow auto-generates a token on connection creation and stores it in config. The token is retrievable via the API (it's in the connection's `config` JSON). Should we mask it after first display (show once, then hide), or is it acceptable to let users view it again from the connection details?

3. **"Coming soon" cards**: Should we show unimplemented platforms in the grid with a "Coming soon" badge (builds anticipation, shows the roadmap), or hide them entirely until implemented (cleaner, no broken promises)?

4. **Multi-connection suggestion**: When a user sets up a Supabase log poller, should the wizard suggest also adding a direct Postgres connection for agent investigation? This could be a simple prompt at the end: "Want to also connect your database for deeper investigation?" If yes, it would launch a second wizard flow for Postgres.

5. **Editing via wizard**: Should editing an existing connection also use the wizard (pre-populated with current values), or is the current inline form sufficient for edits? The wizard would be overkill for changing a password, but it would maintain consistency.

6. **Platform detection**: Some setup steps could auto-detect configuration. For example, if the user's Heimdall instance is on Fly.io, we could pre-fill the webhook URL. Similarly, for Supabase, we could validate the project ref format (20 lowercase chars) before the user clicks Continue. How much validation/auto-detection is worth building now vs. later?
