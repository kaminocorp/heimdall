# Source Filtering & Org-Level Connections

**Status:** Not started
**Owner:** TBD
**Prereqs:** None — phases are independently shippable

---

## Problem

Fly.io's Log Shipper is org-wide: it drains logs from **every** Fly app into a single HTTP endpoint. But on Heimdall's side, each webhook connection is scoped to a specific Application. The result: if a user has five Fly apps, all five apps' logs land in whichever Heimdall Application the webhook connection belongs to — polluting the feed with irrelevant entries.

This is not unique to Fly.io. Any multi-source integration (GitHub repos, future Kubernetes clusters, AWS accounts) faces the same structural mismatch: **one integration produces logs from many sources, but a Heimdall Application should only see the sources it cares about.**

### Current State

1. **Connections are always app-scoped.** The `connections` table requires `app_id` — every connection belongs to exactly one Heimdall Application.
2. **No filtering exists at ingestion time.** The webhook handler accepts all entries that authenticate with a valid token and stores them under the connection's `app_id`.
3. **The Fly.io parser already extracts the source app name.** `fly.app.name` lands in both `source_type` (`"flyio/trajan"`) and `payload.app_name` — the data is available, we just ignore it.
4. **GitHub repos use a bespoke `github_repos` table** for the same "connect once, select what to include" pattern. This works but isn't reusable.

### What Users Need

The same pattern GitHub already uses, generalised:

| GitHub | Fly.io | Generic Concept |
|--------|--------|-----------------|
| Install GitHub App (org-wide) | Set up Fly Log Shipper (org-wide) | Create connection |
| `github_repos` table | — (missing) | Source discovery + filtering |
| List repos via GitHub API | Detect apps from traffic + manual add | Source discovery |
| Toggle repos on/off | Toggle Fly apps on/off | Source selection |
| Code search scoped to enabled repos | Ingestion filtered to enabled apps | Filtered consumption |

Additionally, users should be able to choose whether a connection is **scoped to one app** (current behaviour) or **shared across the entire organisation** — with per-app source selection in the latter case.

---

## Vision

A user connects an integration once, Heimdall discovers the available sources within that integration, and the user toggles which sources feed into which Heimdall Application. This works identically whether the integration is Fly.io, GitHub, Kubernetes, or anything else.

**Principles:**

1. **Connect once, filter per app.** One Fly.io drain, one GitHub App install — no duplicate connections for the same integration.
2. **Discover automatically, add manually.** Auto-detection of sources from incoming traffic is convenient but fragile (silent apps won't appear). Users can always manually add source names.
3. **Drop by default.** No data is stored until the user explicitly enables a source. This prevents feed pollution and gives users full control.
4. **Explicit, simple UX.** Org-level vs. app-level scoping must be crystal clear in the UI. No hidden behaviour, no implicit routing. The user always knows where their data goes.
5. **Generic abstraction.** The filtering system works across all connection types, not just Fly.io. GitHub repos migrate to the same system over time.

---

## Architecture

### Data Model

Three changes to the database:

#### 1. Connection Scoping: `org_id` column + nullable `app_id`

```sql
-- Migration 033: Add org-level connection support
ALTER TABLE connections ADD COLUMN org_id UUID REFERENCES organizations(id);
ALTER TABLE connections ALTER COLUMN app_id DROP NOT NULL;

-- Backfill org_id from existing app-scoped connections
UPDATE connections c
SET org_id = a.org_id
FROM applications a
WHERE c.app_id = a.id;

ALTER TABLE connections ALTER COLUMN org_id SET NOT NULL;

-- Invariant: app-scoped connections must have app_id, org-scoped must not
ALTER TABLE connections ADD CONSTRAINT chk_connection_scope
  CHECK (
    (app_id IS NOT NULL) OR  -- app-scoped: specific app
    (app_id IS NULL)         -- org-scoped: shared across org
  );
```

- `org_id` — always set, links to the owning organisation
- `app_id` — if set, connection is app-scoped (current behaviour). If NULL, connection is org-scoped (visible to all apps in the org)

#### 2. Source Discovery: `connection_sources` table

```sql
-- Migration 033: Source discovery ledger
CREATE TABLE connection_sources (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id   UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    source_name     TEXT NOT NULL,
    first_seen_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (connection_id, source_name)
);

CREATE INDEX idx_connection_sources_conn ON connection_sources(connection_id);

ALTER TABLE connection_sources ENABLE ROW LEVEL SECURITY;
CREATE POLICY connection_sources_owner ON connection_sources
  FOR ALL
  USING (
    connection_id IN (
      SELECT id FROM connections WHERE user_id = app_current_user_id()
    )
  );
```

This table is the **auto-discovery ledger**. Every time ingestion encounters a new source identifier (e.g. `fly.app.name = "trajan"`), it upserts a row here. It answers: "What sources has this connection ever seen?"

Not user-facing directly — it feeds the source selector UI.

#### 3. Per-App Source Selection: `app_source_filters` table

```sql
-- Migration 033: Per-app source filtering
CREATE TABLE app_source_filters (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id          UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    connection_id   UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    source_name     TEXT NOT NULL,
    enabled         BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (app_id, connection_id, source_name)
);

CREATE INDEX idx_app_source_filters_app_conn ON app_source_filters(app_id, connection_id);
CREATE INDEX idx_app_source_filters_lookup ON app_source_filters(connection_id, source_name)
  WHERE enabled = true;

ALTER TABLE app_source_filters ENABLE ROW LEVEL SECURITY;
CREATE POLICY app_source_filters_owner ON app_source_filters
  FOR ALL
  USING (
    app_id IN (
      SELECT a.id FROM applications a
      JOIN org_members om ON om.org_id = a.org_id
      WHERE om.user_id = app_current_user_id()
    )
  );
```

This is the **user's toggle list**. It answers: "For this Heimdall Application, from this connection, which sources should be accepted?"

- `enabled = false` by default → **drop by default** semantics
- Users toggle sources on/off from the source selector UI
- Manually added sources appear here with `enabled = true` even before being seen in `connection_sources`

### Ingestion Flow

Current flow:

```
Webhook arrives → authenticate by token → find connection → insert all entries with connection.app_id
```

New flow:

```
Webhook arrives
  → authenticate by token → find connection
  → parse payload → extract source_name (e.g. fly.app.name)
  → UPSERT connection_sources (discovery ledger)
  → query app_source_filters:
      if connection.app_id IS NOT NULL (app-scoped):
        → check app_source_filters for (connection.app_id, connection.id, source_name)
        → if no filter row exists OR enabled = false → DROP entry
        → if enabled = true → INSERT into log_buffer with connection.app_id
      if connection.app_id IS NULL (org-scoped):
        → find ALL app_source_filters rows for (connection.id, source_name) WHERE enabled = true
        → for each matching app → INSERT into log_buffer with that app's app_id
  → return 202 (entries accepted for processing)
```

**Performance note:** The filter lookup adds one indexed query per webhook batch (not per entry — source names repeat within a batch). The `idx_app_source_filters_lookup` partial index on `(connection_id, source_name) WHERE enabled = true` keeps this fast.

### Source Name Extraction (Generic)

Each connection type defines how to extract a `source_name` from its parsed entries:

| Connection Type | Source Name | Extracted From |
|-----------------|------------|----------------|
| `webhook_logs` (Fly.io) | `"trajan"`, `"my-api-prod"` | `fly.app.name` from parsed payload |
| `webhook_logs` (generic) | `source_type` field value | `source_type` field in payload |
| `github` | `"owner/repo"` | Repository full name |
| `otlp` | `service.name` | OpenTelemetry resource attribute |
| `syslog` | hostname | Syslog header |

This mapping lives in the parser layer — each parser returns a `source_name` alongside the existing parsed entry fields.

### API Endpoints

```
GET    /api/connections/{id}/sources          — List discovered sources + filter state for current app
PUT    /api/connections/{id}/sources          — Upsert source filter selections
POST   /api/connections/{id}/sources          — Manually add a source name
DELETE /api/connections/{id}/sources/{name}   — Remove a manually added source
```

The `GET` endpoint merges `connection_sources` (discovery) with `app_source_filters` (user selections) — same merge pattern used by `ListGitHubRepos`.

### Frontend Components

**`SourceSelector.vue`** — Generic source selector component (replaces the role of `GitHubRepoSelector.vue` for new connection types):

```
┌─ Sources ─────────────────────────────────────────────┐
│                                                        │
│  ● trajan              last seen 2 min ago    [ON ]   │
│  ● my-api-staging      last seen 1 hr ago     [OFF]   │
│  ○ payment-service     (manually added)        [ON ]   │
│                                                        │
│  ┌──────────────────────┐ [Add]                       │
│  │ Add app name...      │                             │
│  └──────────────────────┘                             │
│                                                        │
│  ● = auto-discovered   ○ = manually added             │
│  Sources are disabled by default until you enable them │
└────────────────────────────────────────────────────────┘
```

- Auto-discovered sources show `last_seen_at` timestamp
- Manually added sources show "(manually added)" label
- Toggle switches for enable/disable
- Text input + "Add" button for manual source names
- Accessible from connection detail modal → "Manage Sources" button

**Connection scoping in wizard** — New step when creating multi-source connections:

```
┌─ Connection Scope ─────────────────────────────────────┐
│                                                         │
│  Where should this connection be available?             │
│                                                         │
│  ┌─────────────────────────────────────────────┐       │
│  │  ◉  This app only — Trajan                  │       │
│  │     Logs are scoped to this application.     │       │
│  └─────────────────────────────────────────────┘       │
│  ┌─────────────────────────────────────────────┐       │
│  │  ○  Entire organisation — Kamino Corp        │       │
│  │     Available to all apps. Each app picks    │       │
│  │     which sources to include.                │       │
│  └─────────────────────────────────────────────┘       │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**Connections page** — Org-scoped connections show an "Org-wide" badge and appear in every app's connection list with a visual distinction (e.g. muted border, org icon).

---

## Implementation Phases

Each phase is independently shippable. Later phases build on earlier ones but earlier phases provide standalone value.

---

### Phase 1 — Source Discovery & Filtering (App-Scoped)

**Goal:** Solve the immediate Fly.io problem. Users can filter which Fly apps' logs are stored, using the existing app-scoped connection model.

**No changes to connection scoping** — connections stay app-scoped. The `app_source_filters` table uses the connection's existing `app_id`. This is the minimal viable fix.

#### Task 1.1 — Database migration

**File:** `backend/migrations/033_source_filtering.up.sql` (and `.down.sql`)

Create both tables (`connection_sources` and `app_source_filters`) with RLS policies. The `org_id` column on `connections` is **not** added in this phase.

#### Task 1.2 — sqlc queries for source filtering

**File:** `backend/internal/db/queries/source_filters.sql`

Queries needed:
- `UpsertConnectionSource` — upsert into `connection_sources` (discovery)
- `ListConnectionSources` — list all discovered sources for a connection
- `ListAppSourceFilters` — list filters for a (app_id, connection_id) pair
- `UpsertAppSourceFilter` — upsert a filter row (enable/disable)
- `DeleteAppSourceFilter` — remove a manually added filter
- `ListEnabledSourceNames` — return enabled source names for a (app_id, connection_id) pair (used at ingestion time for fast lookup)

Run `make sqlc-generate` after writing the queries.

#### Task 1.3 — Source name extraction in parsers

**File:** `backend/internal/api/handlers/webhook_parsers.go`

Update the `ParsedEntry` struct (or equivalent return type) to include a `SourceName` field. Populate it:
- Fly.io parser: `entry.Fly.App.Name`
- Generic parser: `entry.SourceType` (fallback)

This is a non-breaking change — `SourceName` is a new field alongside existing ones.

#### Task 1.4 — Ingestion filtering in webhook handler

**File:** `backend/internal/api/handlers/webhooks.go`

In `ingestEntries` (or the function that processes parsed entries):

1. After parsing, extract `source_name` from each entry
2. Upsert `connection_sources` (batch upsert for all unique source names in the batch)
3. Load the set of enabled source names for `(connection.app_id, connection.id)` (single query, cached for the duration of the request)
4. Filter: only entries whose `source_name` is in the enabled set proceed to `InsertLogEntry`
5. If no filters exist at all for this connection, **drop all entries** (drop-by-default)

Return 202 regardless — the caller doesn't need to know which entries were filtered.

#### Task 1.5 — API endpoints for source management

**File:** `backend/internal/api/handlers/source_filters.go` (new)

Implement four endpoints:
- `GET /api/connections/{id}/sources` — merge `connection_sources` + `app_source_filters` for the current app
- `PUT /api/connections/{id}/sources` — bulk upsert filter selections (array of `{source_name, enabled}`)
- `POST /api/connections/{id}/sources` — add a manual source name (creates both a `connection_sources` row and an `app_source_filters` row with `enabled = true`)
- `DELETE /api/connections/{id}/sources/{name}` — remove a manually added source filter

Register routes in `router.go`. All endpoints require JWT auth + ownership check on the connection.

#### Task 1.6 — Frontend: SourceSelector component

**File:** `frontend/src/components/connections/SourceSelector.vue` (new)

Generic component that:
- Fetches discovered sources + filter state via `GET /api/connections/{id}/sources`
- Renders the toggle list (auto-discovered vs. manually added, last-seen timestamps)
- Provides manual "Add source" input
- Calls `PUT` to save toggle changes
- Emits events on save/cancel

#### Task 1.7 — Frontend: integrate into connection detail modal

**File:** `frontend/src/components/connections/ConnectionDetailModal.vue`

Add a "Manage Sources" button for connection types that support source filtering (initially `webhook_logs` with Fly.io detection). Opens `SourceSelector` in a modal or inline panel.

#### Task 1.8 — Frontend: integrate into Fly.io wizard flow

**File:** `frontend/src/components/connections/wizard/ConnectionWizard.vue` and relevant step files

After the drain setup step completes and logs start arriving, optionally show a "Select which Fly apps to include" step using `SourceSelector`. If no sources have been discovered yet, show the manual add interface with a note: "Sources will appear here automatically as logs arrive."

#### Task 1.9 — Tests

**Files:** Handler tests, parser tests, integration tests

- Parser tests: verify `source_name` extraction for Fly.io, generic, and edge cases (missing app name)
- Handler tests: verify filtering logic (enabled source passes, disabled source dropped, unknown source dropped, no filters = all dropped)
- API tests: CRUD operations on source filters
- Frontend: component tests for SourceSelector

---

### Phase 2 — Org-Level Connections

**Goal:** Users can create connections at the organisation level, shared across all apps. Each app selects which sources to include independently.

**Depends on:** Phase 1 (source filtering infrastructure)

#### Task 2.1 — Database migration: org-scoped connections

**File:** `backend/migrations/034_org_connections.up.sql`

- Add `org_id` column to `connections` (NOT NULL, backfilled from `applications.org_id`)
- Make `app_id` nullable
- Add check constraint ensuring `org_id` is always present
- Update RLS policies to handle org-scoped connections (visible to all org members)

#### Task 2.2 — Backend: update connection queries and handlers

**Files:** `backend/internal/db/queries/connections.sql`, `backend/internal/api/handlers/connections.go`

- `ListConnectionsByApp` — also return org-scoped connections for the app's org
- `CreateConnection` — accept optional `app_id`; if omitted, create as org-scoped
- Authorization: org-scoped connections require org membership check instead of app ownership

#### Task 2.3 — Backend: multi-app routing in webhook handler

**File:** `backend/internal/api/handlers/webhooks.go`

When an org-scoped connection receives a webhook:
1. Extract `source_name` as before
2. Query `app_source_filters` for **all apps** that have `enabled = true` for this `(connection_id, source_name)`
3. Insert into `log_buffer` once per matching app (with that app's `app_id`)

#### Task 2.4 — Frontend: connection scoping in wizard

**File:** New wizard step component

Add a "Connection Scope" step to the wizard for connection types that support org-level scoping. Two-option radio: "This app only" vs. "Entire organisation." Clear one-line descriptions for each.

#### Task 2.5 — Frontend: org-wide badge on connections page

**File:** `frontend/src/components/connections/ConnectionsPage.vue`

Org-scoped connections render with an "Org-wide" badge. They appear in every app's connection list but are visually distinct (e.g. different border style, org icon).

#### Task 2.6 — Frontend: per-app source selector for org connections

Org-scoped connections show "Manage Sources" in the detail modal. The selector is scoped to the currently active Heimdall Application — each app manages its own filter independently.

#### Task 2.7 — Tests

- Backend: org-scoped connection creation, multi-app routing, authorization
- Frontend: scoping step, org badge rendering, per-app filter isolation

---

### Phase 3 — Migrate GitHub Repos to Generic Source Filters

**Goal:** Unify the `github_repos` table with the generic `app_source_filters` system. One pattern for all "connect once, select what to include" flows.

**Depends on:** Phase 2 (org-level connections)

#### Task 3.1 — Data migration

Migrate existing `github_repos` rows into `connection_sources` + `app_source_filters`:
- `repo_full_name` → `source_name`
- `enabled` → `app_source_filters.enabled`
- Repo discovery via GitHub API continues to upsert into `connection_sources`

#### Task 3.2 — Backend: replace GitHub repo endpoints

Replace `ListGitHubRepos` / `UpdateGitHubRepos` with the generic source filter endpoints. The GitHub API repo-fetching logic moves into a "source discovery" hook that populates `connection_sources` on demand (not just from traffic).

#### Task 3.3 — Frontend: replace GitHubRepoSelector with SourceSelector

Replace `GitHubRepoSelector.vue` with `SourceSelector.vue` configured for GitHub context. The component is generic — it just needs a connection ID and renders the same toggle list.

#### Task 3.4 — Drop `github_repos` table

Migration to remove the `github_repos` table after data is migrated and all references are updated.

#### Task 3.5 — Tests

- Verify GitHub repo selection still works end-to-end through the new tables
- Verify migration correctness (no data loss, enabled state preserved)

---

### Phase 4 — Extended Source Extraction

**Goal:** Extend source name extraction to all connection types, not just Fly.io webhooks.

**Depends on:** Phase 1

This phase can run in parallel with Phases 2–3. It broadens the filtering capability to more connection types.

#### Task 4.1 — OTLP source extraction

Extract `service.name` from OpenTelemetry resource attributes as `source_name`.

#### Task 4.2 — Syslog source extraction

Extract hostname from syslog headers as `source_name`.

#### Task 4.3 — Generic webhook source extraction

For non-Fly.io webhook payloads, use the `source_type` field as `source_name`. If absent, use a configurable JSONPath expression stored in the connection config.

#### Task 4.4 — Source selector availability

Enable the "Manage Sources" button in the connection detail modal for all connection types that now support source extraction, not just Fly.io webhooks.

---

## Risk & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Drop-by-default blocks new users** | User sets up drain, no logs appear, confused | Wizard flow includes source selection step. Connection detail shows "0 sources enabled" warning banner with link to selector |
| **Ingestion latency from filter lookup** | Added DB query on hot path | Single indexed lookup per batch (not per entry). Cache enabled set in-memory for request duration. Partial index on `enabled = true` |
| **Org-scoped connection + many apps = fan-out writes** | One webhook batch → N inserts (one per enabled app) | Batch inserts. Monitor fan-out ratio. Consider async processing if fan-out exceeds threshold |
| **Auto-discovery misses silent apps** | Source never appears in selector | Manual "Add source" input always available. Discovery is a convenience, not a requirement |
| **Migration complexity for GitHub repos** | Phase 3 touches auth + API surface | Phase 3 is optional and deferred. Phases 1–2 deliver full value without it |

---

## Open Questions

1. **Should disabled entries be counted?** When a source is filtered out, should we still increment a "filtered" counter visible in the UI? This helps users know the drain is working even if sources aren't enabled yet.

2. **Batch size limits for org fan-out.** If an org has 20 apps and 15 have "trajan" enabled, one webhook batch produces 15x the inserts. Should we cap this, or is it fine at realistic scale?

3. **Retention for `connection_sources`.** Should auto-discovered sources that haven't been seen in N days be pruned? Or kept indefinitely as a historical record?

4. **Default scope for new connections.** Should the wizard default to "This app only" or "Entire organisation"? Recommendation: default to "This app only" — it matches current behaviour and is the safer default.
