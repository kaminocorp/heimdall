# Phase 4 Completion — Connection Wizard

Phase 4 replaces the dropdown-based connection creation form with a guided multi-step wizard. This is a pure frontend change — the backend and API are identical.

---

## What Was Built

### 1. Flow Definitions (`flows.ts`)

**File:** `frontend/src/components/connections/wizard/flows.ts`

Declarative flow system that defines each platform's wizard steps:

```typescript
interface PlatformFlow {
  id: string
  name: string
  icon: string            // 2-letter badge (SB, PG, WH, GH)
  description: string
  category: 'log_source' | 'database' | 'generic'
  connectorType: string   // Maps to backend connection type
  direction: 'one_way' | 'two_way'
  available: boolean
  steps: FlowStep[]       // Array of { id, label, component }
}
```

**Available flows:**

| Platform | Steps | Type |
|----------|-------|------|
| Supabase | Name → Auth → Tables → Test | `supabase` |
| Webhook | Name → Setup | `webhook_logs` |
| PostgreSQL | Name → Config → Test | `postgres` |
| GitHub | Name → Install | `github` |

**Coming soon (disabled):** Datadog, Syslog, MySQL — shown with `opacity-40` and "Soon" badge.

Step components are lazy-loaded via `defineAsyncComponent` to keep the initial bundle small.

---

### 2. Wizard Shell (`ConnectionWizard.vue`)

**File:** `frontend/src/components/connections/wizard/ConnectionWizard.vue`

Full-screen modal orchestrating the entire creation flow:

**State machine:**
1. **Platform selection** — `PlatformGrid` shows available connectors grouped by category
2. **Step progression** — Dynamic `<component :is="...">` renders the current step
3. **Connection creation** — Happens automatically before the test step (or on finish for flows without test)
4. **Done** — Emits `created` event with the new connection ID

**Key behaviours:**
- `WizardState` (name + config) is local reactive, not Pinia — ephemeral, scoped to the modal
- Each step emits `valid` to control the Continue button
- Back button returns to previous step or to platform selection
- For flows with a test step, the connection is created before advancing to test
- For flows without a test step (webhook, GitHub), creation happens on "Done" / "Install"
- Error banner shows API errors inline

**Transitions:** Step content slides left/right via `<Transition>` with `translate-x-4` / `-translate-x-4` and opacity fade (200ms ease-out in, 150ms ease-in out).

---

### 3. Step Indicator (`WizardStepIndicator.vue`)

**File:** `frontend/src/components/connections/wizard/WizardStepIndicator.vue`

Dot-line progress bar: `● ─── ○ ─── ○`

- Active dot: `bg-accent`
- Completed: `bg-accent/60`
- Future: `bg-border`
- Labels below each dot in `text-[10px] uppercase tracking-wider`

---

### 4. Platform Grid (`PlatformGrid.vue` + `PlatformCard.vue`)

**Files:**
- `frontend/src/components/connections/wizard/PlatformGrid.vue`
- `frontend/src/components/connections/wizard/PlatformCard.vue`

Cards grouped by category (Log Sources, Databases, Generic). Each card shows:
- 2-letter icon badge with accent border
- Platform name and one-line description
- Unavailable cards: `opacity-40 cursor-not-allowed` with "Soon" badge

---

### 5. Step Components

**Shared steps (used by multiple flows):**

| File | Purpose |
|------|---------|
| `steps/StepName.vue` | Connection name input. Validates non-empty. |
| `steps/StepTest.vue` | Live connection test. Receives `connectionId` prop. Shows testing/success/error phases with elapsed timer. Reuses the visual pattern from `ConnectionTestModal`. |

**Supabase-specific steps:**

| File | Purpose |
|------|---------|
| `steps/StepSupabaseAuth.vue` | Project reference + Personal Access Token inputs. Info text about Management API and PAT generation. |
| `steps/StepSupabaseTables.vue` | Checkbox group for 6 log tables (defaults: `postgres_logs`, `auth_logs`). Poll interval select (15s/30s/60s). Validates at least one table selected. |

**Existing type steps:**

| File | Purpose |
|------|---------|
| `steps/StepPostgresConfig.vue` | Host, port, database, username, password, SSL mode. 2-column grid layout for host/port. Validates required fields. |
| `steps/StepWebhookSetup.vue` | Shows endpoint format and payload example. Explains that token is generated on creation. Uses `CopyableField`. Always valid. |
| `steps/StepGitHubInstall.vue` | GitHub App install button that redirects to GitHub OAuth. Never auto-advances — user must click install. |

---

### 6. CopyableField Component

**File:** `frontend/src/components/common/CopyableField.vue`

Reusable component for values users need to copy. Shows a monospace code value with a Copy/Copied button. Uses `navigator.clipboard.writeText` with graceful fallback. "Copied" feedback lasts 2 seconds.

---

### 7. ConnectionsPage Integration

**File:** `frontend/src/pages/ConnectionsPage.vue` (modified)

- **"+ New Connection"** button now opens the wizard modal (`showWizard = true`)
- **`ConnectionForm`** preserved for **editing** existing connections (inline, not modal)
- Wizard emits `close` and `created` → triggers connection list refresh
- Both `ConnectionTestModal` (for ping) and `ConnectionWizard` (for creation) can coexist

---

## Files Summary

### New Files (16)

| File | Purpose |
|------|---------|
| `wizard/flows.ts` | Flow definitions and types |
| `wizard/ConnectionWizard.vue` | Wizard shell / orchestrator |
| `wizard/WizardStepIndicator.vue` | Dot-line progress bar |
| `wizard/PlatformGrid.vue` | Categorized platform card grid |
| `wizard/PlatformCard.vue` | Individual platform card |
| `wizard/steps/StepName.vue` | Name input (shared) |
| `wizard/steps/StepTest.vue` | Live connection test (shared) |
| `wizard/steps/StepSupabaseAuth.vue` | Supabase auth fields |
| `wizard/steps/StepSupabaseTables.vue` | Supabase table selection |
| `wizard/steps/StepPostgresConfig.vue` | PostgreSQL config fields |
| `wizard/steps/StepWebhookSetup.vue` | Webhook setup info |
| `wizard/steps/StepGitHubInstall.vue` | GitHub App install |
| `common/CopyableField.vue` | Copy-to-clipboard field |

### Modified Files (1)

| File | Change |
|------|--------|
| `pages/ConnectionsPage.vue` | Import wizard, `showWizard` state, `openCreate` opens wizard, wizard modal in template |

---

## Architecture Decisions

**Local state over Pinia:** Wizard state is ephemeral — it only exists while the modal is open. Using a Pinia store would add unnecessary global state that needs manual cleanup. A local `reactive()` object scoped to the wizard component is simpler and self-cleaning.

**Declarative flows over conditional rendering:** Rather than a giant `v-if` tree per connection type, each platform declares its steps as data. The wizard shell renders them generically via `<component :is="...">`. Adding a new connector type = adding a flow definition + step components, zero wizard shell changes.

**Create-before-test:** For flows with a test step, the connection is created on the server *before* advancing to the test step. This way `StepTest` can call the real `POST /connections/{id}/test` endpoint. If creation fails, the user stays on the current step with an error message.

**Lazy step loading:** Step components are loaded via `defineAsyncComponent` so the initial wizard bundle only contains the shell, grid, and indicator. Step code loads on demand when the user selects a platform.

---

## Verification

- `npm run build` — compiles cleanly
- All 25 existing frontend tests pass
- No new dependencies added
