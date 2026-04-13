# Connection Wizard Redesign — Implementation Plan

**Status:** Not started  
**Goal:** Redesign the wizard's platform selection grid from a flat 3-category list (Log Sources / Databases / Generic) into a structured 3-section layout that communicates *what happens* when you add a connection, not just *what type* it is.

---

## Why this exists

The current wizard opens with a `PlatformGrid` that groups connectors into "Log Sources," "Databases," and "Generic." These categories are technically accurate but fail the user in two ways:

1. **Platform-blind.** A developer on Fly.io has to know that Fly.io uses syslog drains, then find "Syslog" under Log Sources. A developer on Supabase has to know that Supabase log polling is a separate connector from Supabase's underlying PostgreSQL. The user must translate from "where my stuff runs" to "what protocol Heimdall supports" — the wizard should do that translation.

2. **No consequence labelling.** Adding a PostgreSQL connection doesn't make logs flow — it gives the agent a queryable database for investigations. Adding Supabase *does* make logs flow. The current grid treats both as equivalent items in different categories, with no indication that one feeds the monitoring loop and the other doesn't.

### The new layout

Three sections with explicit purpose descriptions:

| Section | Purpose | Items |
|---------|---------|-------|
| **Where do your logs come from?** | Platform-first selection — pick your infra, we determine the connector | Supabase, Fly.io, Vercel, Render, Railway, Heroku, AWS, DigitalOcean |
| **Or connect directly** | Protocol-level escape hatch for experienced engineers | Webhook (HTTP), Syslog (TCP/TLS), OpenTelemetry (OTLP) |
| **Agent investigation tools** | On-demand query access — explicitly not log sources | PostgreSQL, GitHub, MySQL |

The first section maps platforms to their underlying connector:

| Platform | Underlying connector | Status |
|----------|---------------------|--------|
| Supabase | `supabase` (log poller) | Live |
| Fly.io | `syslog` (TCP/TLS drain) | Coming soon |
| Vercel | `webhook_logs` (log drain → webhook) | Coming soon |
| Render | `syslog` (TCP/TLS drain) | Coming soon |
| Railway | `webhook_logs` (HTTP log drain) | Coming soon |
| Heroku | `syslog` (log drain) | Coming soon |
| AWS | `otlp` (CloudWatch → OTLP) | Coming soon |
| DigitalOcean | `syslog` (log forwarding) | Coming soon |

Coming-soon platforms don't need backend work — they're greyed-out cards with a "Soon" badge, identical to how Datadog/MySQL work today. When they become available, each will start the appropriate existing connector flow (syslog, webhook, or otlp) with platform-specific instructions in the setup steps.

---

## What changes and what doesn't

### Changes (this plan)

| Component | Change |
|-----------|--------|
| `flows.ts` | New category system (`platform_log`, `direct_protocol`, `agent_tool`), new platform entries |
| `PlatformGrid.vue` | Rewrite to 3-section layout with headers, subtitles, and a visual divider |
| `PlatformCard.vue` | Minor — add subtitle/protocol hint, adjust sizing for denser grid |
| `ConnectorLogo.vue` | Add logos for Fly.io, Vercel, Render, Railway, Heroku, AWS, DigitalOcean |

### Untouched

| Component | Why |
|-----------|-----|
| `ConnectionWizard.vue` | The state machine, step progression, creation logic, and cleanup are all flow-agnostic — they work off `PlatformFlow.steps` regardless of how the flow was categorised |
| All step components (`StepName`, `StepSupabaseAuth`, etc.) | Steps are per-connector, not per-category — nothing about them changes |
| `WizardStepIndicator.vue` | Step indicator is flow-agnostic |
| `ConnectionForm.vue` | Edit form is separate from the wizard entirely |
| Backend | No API changes — all new platforms map to existing connector types |

---

## Implementation Phases

### Phase 1 — Restructure `flows.ts`

**Goal:** New category taxonomy and platform entries.

**File:** `frontend/src/components/connections/wizard/flows.ts`

**Changes:**

1. Update the `PlatformFlow` interface to support the new sections:

```ts
export interface PlatformFlow {
  id: string
  name: string
  icon: string               // still used as fallback if no ConnectorLogo match
  description: string
  section: 'platform_log' | 'direct_protocol' | 'agent_tool'  // replaces `category`
  subtitle?: string          // e.g. "via Syslog drain" — shown on the card
  connectorType: string
  direction: 'one_way' | 'two_way'
  available: boolean
  steps: FlowStep[]
}
```

2. Keep the existing `category` field as a deprecated alias during transition (the `ConnectionForm` edit flow doesn't use it, but the old `BlueprintView` did — now deleted, so safe to remove).

3. Add new platform entries:

```ts
// Platform log sources
{ id: 'supabase',      section: 'platform_log',    subtitle: 'Log polling + API',     available: true,  ... }
{ id: 'flyio',         section: 'platform_log',    subtitle: 'via Syslog drain',      available: false, connectorType: 'syslog', ... }
{ id: 'vercel',        section: 'platform_log',    subtitle: 'via Log Drain',         available: false, connectorType: 'webhook_logs', ... }
{ id: 'render',        section: 'platform_log',    subtitle: 'via Syslog drain',      available: false, connectorType: 'syslog', ... }
{ id: 'railway',       section: 'platform_log',    subtitle: 'via HTTP Log Drain',    available: false, connectorType: 'webhook_logs', ... }
{ id: 'heroku',        section: 'platform_log',    subtitle: 'via Syslog drain',      available: false, connectorType: 'syslog', ... }
{ id: 'aws',           section: 'platform_log',    subtitle: 'via CloudWatch + OTLP', available: false, connectorType: 'otlp', ... }
{ id: 'digitalocean',  section: 'platform_log',    subtitle: 'via Log Forwarding',    available: false, connectorType: 'syslog', ... }

// Direct protocols
{ id: 'webhook_logs',  section: 'direct_protocol', subtitle: 'HTTP endpoint',         available: true,  ... }
{ id: 'syslog',        section: 'direct_protocol', subtitle: 'TCP / TLS listener',    available: true,  ... }
{ id: 'otlp',          section: 'direct_protocol', subtitle: 'OTLP HTTP endpoint',    available: true,  ... }

// Agent investigation tools
{ id: 'postgres',      section: 'agent_tool',      subtitle: 'Query during investigations', available: true, ... }
{ id: 'github',        section: 'agent_tool',      subtitle: 'Code search & context',       available: true, ... }
{ id: 'mysql',         section: 'agent_tool',      subtitle: 'Query during investigations', available: false, ... }
```

4. Remove the old `datadog` entry — it doesn't fit cleanly into any section and was never wired up. If Datadog support is added later, it would be a `platform_log` entry.

**Acceptance criteria:**
- [ ] All existing flows keep their `id`, `connectorType`, `direction`, `steps` unchanged
- [ ] New platform entries have `available: false` and empty `steps: []`
- [ ] `section` field replaces `category` on all entries
- [ ] `subtitle` field present on all entries
- [ ] `getFlowById()` still works
- [ ] `supabaseLogTables` export unchanged
- [ ] `vue-tsc --noEmit` clean

---

### Phase 2 — Add Platform Logos to `ConnectorLogo.vue`

**Goal:** Official SVG marks for the new platforms.

**File:** `frontend/src/components/icons/ConnectorLogo.vue`

**New logos needed:**

| Platform | Source | Mark |
|----------|--------|------|
| Fly.io | Simple Icons | Bird/paper plane mark |
| Vercel | Simple Icons | Triangle mark |
| Render | Simple Icons | R mark |
| Railway | Simple Icons | Railway mark |
| Heroku | Simple Icons | H mark |
| AWS | Simple Icons | Arrow/smile mark |
| DigitalOcean | Simple Icons | DO droplet mark |

Same pattern as Phase 2 of the Ingestion redesign: fetch official paths from Simple Icons, add `v-else-if` branches with `fill="currentColor"`.

**Acceptance criteria:**
- [ ] All 7 new platforms render recognisable logos
- [ ] Existing logos unchanged
- [ ] `vue-tsc --noEmit` clean

---

### Phase 3 — Redesign `PlatformGrid.vue`

**Goal:** Three-section layout with headers, subtitles, and visual hierarchy.

**File:** `frontend/src/components/connections/wizard/PlatformGrid.vue`

**Current structure (to be replaced):**
```
LOG SOURCES
  [Card] [Card] [Card] ...

DATABASES
  [Card] [Card] ...

GENERIC
  [Card] ...
```

**New structure:**
```
WHERE DO YOUR LOGS COME FROM?
Pick your platform — we'll handle the wiring.

  [Supabase] [Fly.io]  [Vercel]  [Render]
  [Railway]  [Heroku]  [AWS]     [DigitalOcean]

──────────── or ────────────

CONNECT DIRECTLY
Already know the protocol? Skip the platform.

  [Webhook HTTP]  [Syslog TCP/TLS]  [OpenTelemetry OTLP]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

AGENT INVESTIGATION TOOLS
These don't send logs — they let the agent query
your systems during investigations and chat.

  [PostgreSQL]  [GitHub]  [MySQL]
```

**Implementation details:**

- Three `<section>` blocks, each with:
  - Header (`h3`, mono, uppercase, tracking-wider)
  - Subtitle (`p`, sans, text-secondary, smaller)
  - Grid of `PlatformCard` components
- Between sections 1 and 2: a visual "or" divider (horizontal line with centred "or" text, common UI pattern)
- Between sections 2 and 3: a stronger visual break — thicker border or subtle background change to signal "this is fundamentally different"
- Section 3 gets a distinct visual treatment — slightly muted background, maybe a subtle icon (magnifying glass / query) in the header to reinforce "investigation, not monitoring"

**Grid layout per section:**
- Section 1 (platforms): `grid-cols-2 sm:grid-cols-3 md:grid-cols-4` — denser, more items
- Section 2 (protocols): `grid-cols-3` — exactly 3, evenly spaced
- Section 3 (tools): `grid-cols-3` — exactly 3, evenly spaced

**Computed grouping:**
```ts
const platformLogs = computed(() => flows.filter(f => f.section === 'platform_log'))
const directProtocols = computed(() => flows.filter(f => f.section === 'direct_protocol'))
const agentTools = computed(() => flows.filter(f => f.section === 'agent_tool'))
```

**Acceptance criteria:**
- [ ] Three visually distinct sections render correctly
- [ ] Headers and subtitles communicate purpose
- [ ] Available platforms are clickable and emit `select`
- [ ] Coming-soon platforms show "Soon" badge and are disabled
- [ ] Sections 1 and 2 are visually connected (both about log ingestion)
- [ ] Section 3 is visually separated (different purpose)
- [ ] Responsive: 2-col on mobile, 3-4 col on desktop
- [ ] `vue-tsc --noEmit` and `vite build` clean

---

### Phase 4 — Update `PlatformCard.vue`

**Goal:** Cards show the subtitle/protocol hint and use `ConnectorLogo` instead of 2-letter badges.

**File:** `frontend/src/components/connections/wizard/PlatformCard.vue`

**Current card anatomy:**
```
┌────────────┐
│  [SB]      │   ← 2-letter icon badge
│  Supabase  │   ← Name
│  Database   │   ← Description
│  logs...   │
└────────────┘
```

**New card anatomy:**
```
┌────────────┐
│  [logo]    │   ← ConnectorLogo (native brand SVG)
│  Supabase  │   ← Name
│  via Log   │   ← Subtitle (from flow.subtitle)
│  polling   │
└────────────┘
```

**Changes:**
- Replace the 2-letter badge `<div>` with `<ConnectorLogo :type="flow.id" :size="28" />`
- Show `flow.subtitle` instead of `flow.description` (subtitle is shorter, more informative at grid scale)
- Keep the "Soon" badge for `!flow.available`
- Keep disabled styling for coming-soon
- Slightly reduce card padding to fit denser grids

**Acceptance criteria:**
- [ ] Cards show native brand logos
- [ ] Subtitle visible beneath name
- [ ] "Soon" badge still works
- [ ] Cards fit well in 4-col grid
- [ ] `vue-tsc --noEmit` clean

---

### Phase 5 — Verify & Polish

**Goal:** End-to-end verification that all existing wizard flows still work, and visual polish.

**Tasks:**

1. **Smoke test every available flow:**
   - Supabase: select → Name → Auth → Tables → Test → created
   - Webhook: select → Name → Setup → created
   - Syslog: select → Name → Config → Test → created
   - OTLP: select → Name → Setup → created
   - PostgreSQL: select → Name → Config → Test → created
   - GitHub: select → Name → Install → redirect

2. **Verify coming-soon platforms:**
   - All 7 new platforms render with logo, name, subtitle, "Soon" badge
   - Click does nothing (disabled)
   - No console errors

3. **Visual polish:**
   - Section spacing feels balanced
   - "or" divider between sections 1-2 is subtle but clear
   - Section 3 separator is visually stronger
   - Cards align consistently across sections
   - Mobile layout stacks gracefully

4. **Cleanup:**
   - Remove any references to the old `category` field in `flows.ts`
   - Verify `ConnectionForm.vue` edit flow still works (it reads `connectorType`, not `category`)

5. **Type-check and build:**
   - `vue-tsc --noEmit` clean
   - `vite build` clean
   - `vitest run` all tests pass

---

## Files Summary

### Modified files (4)
| File | Phase | Change |
|------|-------|--------|
| `frontend/src/components/connections/wizard/flows.ts` | 1 | New section taxonomy, platform entries, subtitle field |
| `frontend/src/components/icons/ConnectorLogo.vue` | 2 | Add 7 platform logos |
| `frontend/src/components/connections/wizard/PlatformGrid.vue` | 3 | Rewrite to 3-section layout |
| `frontend/src/components/connections/wizard/PlatformCard.vue` | 4 | ConnectorLogo + subtitle display |

### Untouched files
| File | Why |
|------|-----|
| `ConnectionWizard.vue` | State machine is flow-agnostic — works off `PlatformFlow.steps` |
| All `Step*.vue` components | Steps are per-connector-type, not per-category |
| `WizardStepIndicator.vue` | Step indicator is flow-agnostic |
| `ConnectionForm.vue` | Edit form reads `connectorType`, not `category` |
| Backend (all Go code) | No API changes — new platforms map to existing connectors |

---

## Execution Order

Phases 1 and 2 are independent and can be done in parallel. Phase 3 depends on Phase 1 (needs the new section grouping). Phase 4 depends on Phase 2 (needs new logos). Phase 5 depends on all of 1–4.

```
Phase 1 (flows.ts) ──→ Phase 3 (PlatformGrid) ──┐
                                                   ├──→ Phase 5 (Verify & Polish)
Phase 2 (Logos) ──────→ Phase 4 (PlatformCard) ──┘
```

---

## Risk Notes

1. **`category` field removal** — `flows.ts` exports `flows` which is imported by `PlatformGrid.vue` and `BlueprintView.vue`. BlueprintView was deleted in 0.39.0, so the only consumer of `category` is `PlatformGrid` itself. Safe to replace with `section`.

2. **`ConnectionForm.vue` uses `flows`** — Only to build the type-to-category lookup for display in edit mode. This uses `connectorType`, not `category`. The form will still find the flow by `connectorType` regardless of `section` changes.

3. **Platform → connector mapping ambiguity** — When a coming-soon platform becomes available, its flow will need platform-specific setup instructions (e.g. "Go to your Fly.io dashboard → Monitoring → Add log drain"). This is a future concern — the step components will need either conditional content based on platform ID or new platform-specific step components. Not in scope for this plan.
