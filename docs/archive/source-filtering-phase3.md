# Source Filtering — Phase 3 Completion

**Scope:** Retire the bespoke `github_repos` table and route GitHub repo selection through the same `(connection_sources, app_source_filters)` pair that Phase 1 introduced. One "connect once, pick what to include" pattern for every connector type. One selector component. One set of endpoints.
**Plan:** `docs/executing/source-filtering-and-org-connections.md`
**Builds on:** Phases 1 & 2.

---

## The Problem This Solves

Pre-Phase-3, GitHub had its own table, its own endpoints, and its own Vue component for repo selection. Every other "pick what's on" flow — Fly.io apps, future OTLP services, future syslog hosts — had to invent its own mirror. Phase 1 introduced the generic machinery on the Fly.io side but left GitHub running in parallel. Phase 3 collapses the two.

The user-visible UX is unchanged: click a button, see a list of repos, toggle, save. The underlying plumbing is now the same plumbing that filters Fly apps. New connectors that ship in Phase 4 and beyond will reuse both the tables and the Vue component.

---

## What Changed

### 1. Migration 036 — data move + table drop

**File:** `backend/migrations/036_github_to_generic_sources.up.sql` (+ `.down.sql`)

The migration is idempotent end-to-end:

```sql
INSERT INTO connection_sources (connection_id, source_name, first_seen_at, last_seen_at)
SELECT gr.connection_id, gr.repo_full_name, gr.created_at, gr.created_at
FROM github_repos gr
ON CONFLICT (connection_id, source_name) DO NOTHING;

INSERT INTO app_source_filters (app_id, connection_id, source_name, enabled, created_at)
SELECT c.app_id, gr.connection_id, gr.repo_full_name, gr.enabled, gr.created_at
FROM github_repos gr
JOIN connections c ON c.id = gr.connection_id
WHERE c.app_id IS NOT NULL
ON CONFLICT (app_id, connection_id, source_name) DO NOTHING;

DROP TABLE github_repos;
```

`repo_full_name` becomes `source_name`; `enabled` maps one-for-one. The discarded columns — `repo_id` (bigint) and `default_branch` — were not read by any consumer of the table: the agent's `search_codebase` tool only ever used `repo_full_name` and the parent connection's `config`. Renames will now manifest as "new source" rather than "same row, updated name," which is the same semantics every other source type exhibits (a renamed Fly app is a new source, too).

The down migration recreates the table and re-seeds from the generic rows, synthesising negative placeholder `repo_id`s — a realistic rollback would re-sync from GitHub for authentic ids, so the placeholders are an honest marker rather than a pretence.

### 2. sqlc — delete the github_repos queries, rewrite the agent-facing projection

**Files:** `backend/internal/db/queries/github_repos.sql` (deleted), `backend/internal/db/queries/source_filters.sql` (extended)

Three pre-Phase-3 queries disappear with the file: `ListGitHubReposByConnection`, `UpsertGitHubRepo`, `DeleteGitHubRepo`. None had callers after the handler rewrite.

`ListEnabledGitHubReposByApp` moves into `source_filters.sql` and is rewritten as a projection over the new tables:

```sql
SELECT
    c.id               AS connection_id,
    asf.source_name    AS repo_full_name,
    c.config           AS connection_config
FROM app_source_filters asf
JOIN connections c ON c.id = asf.connection_id
WHERE asf.app_id = $1
  AND asf.enabled = true
  AND c.type = 'github'
  AND c.status = 'active'
ORDER BY asf.source_name;
```

The generated row struct shrinks from 8 fields (`id`, `connection_id`, `repo_full_name`, `repo_id`, `default_branch`, `enabled`, `created_at`, `connection_config`) to 3 (`connection_id`, `repo_full_name`, `connection_config`) because the agent only reads those three. `ListEnabledGitHubReposByApp_QuerySignature` is a compile-time regression guard against anyone silently breaking the consumer.

sqlc's generated `db.GithubRepo` model is gone; no hand-written code referenced it after the handler rewrite.

### 3. Backend handlers — replace 2 with 1

**Files:**
- Deleted: `backend/internal/api/handlers/github_repos.go` (the file that owned `ListGitHubRepos` + `UpdateGitHubRepos` + `TestGitHubConnection`)
- New: `backend/internal/api/handlers/source_filters_discover.go`
- `TestGitHubConnection` moved into `github_install.go` (its sole caller remains `connections_test_handler.go`)

**`DiscoverSources`** (`POST /api/connections/{id}/sources/discover`) is the Phase 3 addition. It's a thin router on connection type:

- `github`: calls `discoverGitHubRepos`, which paginates `GET /installation/repositories` (50 pages × 100 repos = 5,000 ceiling, matching the old `ListGitHubRepos` behaviour) and upserts each `owner/repo` into `connection_sources`.
- `webhook_logs` / `otlp` / etc.: returns 400 with a message pointing callers at the passive flow ("sources appear automatically as traffic arrives").

The handler wraps everything in `UserQueries` so upserts land under a real RLS transaction. Errors from the GitHub API surface as 502 Bad Gateway with a user-facing message — the frontend renders these verbatim in the sync button's error state.

Routes in `router.go` lose `GET/PUT /{id}/github/repos` and gain `POST /{id}/sources/discover`.

### 4. Frontend — one selector, two modes

**Files:**
- Deleted: `frontend/src/components/connections/GitHubRepoSelector.vue`
- Deleted: `frontend/src/types/github.ts` (the `GitHubRepo` interface)
- Changed: `frontend/src/api/github.ts` (retains only `getGitHubInstallURL`; repo functions removed)
- Changed: `frontend/src/api/sources.ts` (adds `discoverSources`)
- Changed: `frontend/src/components/connections/SourceSelector.vue` (new `discoverable` prop)
- Changed: `frontend/src/components/connections/ConnectionDetailModal.vue` (`manage-repos` emit collapsed into `manage-sources`; "Repos" label used for GitHub, "Sources" for everything else)
- Changed: `frontend/src/components/connections/wizard/ConnectionWizard.vue` (emits `manage-sources`)
- Changed: `frontend/src/pages/ConnectionsPage.vue` (removes GitHubRepoSelector import + state; picks discoverable mode based on connection type)

**`discoverable` is a UX mode flip.** When the prop is true, the selector:

- **Hides polling.** Discovery is a user-initiated action, not a background sweep. Hammering GitHub every 10s is unnecessary and rude.
- **Hides the manual-add form.** You can't grant yourself GitHub repo access by typing a name; the only meaningful action is "sync what we can see."
- **Hides the staleness banner.** Repos don't "go quiet" the way a Fly app does; the tiered dots still render per-repo for visual consistency but the org-wide banner would just confuse.
- **Shows a "Sync" button** in the header that triggers `POST /sources/discover` and refreshes the list. On success the button briefly reads `Synced (N)` to confirm.
- **Auto-syncs on mount if the local list is empty.** First-time opens do the right thing; subsequent opens don't re-sync unless the user asks.

Everything else — toggle rows, bulk save, the `appId` prop for org-scoped filter editing — works identically in both modes.

The existing `manage-repos` event was renamed `manage-sources` everywhere; there's no backwards-compat shim because the emitter (`ConnectionDetailModal` / `ConnectionWizard`) and the handler (`ConnectionsPage`) are all in the same repo and ship together.

### 5. Agent tool — no source-level changes

**File:** `backend/internal/agent/tools_codebase.go`

The agent calls `ListEnabledGitHubReposByApp` the same way it always did: get rows, pull `RepoFullName`, pass `[0].ConnectionConfig` into the codebase connector. The connector itself (`connectors/codebase/github.go`) reads only `installation_id` from the config and the list of repo full-names. Zero lines changed in `tools_codebase.go`, zero lines changed in `codebase/github.go`. That's exactly what I wanted out of the query-shape change — the caller couldn't care less where the data came from.

### 6. Tests

**File:** `backend/internal/api/handlers/source_filters_github_test.go`

Six new tests:

- **`TestSourceFilters_GitHubEnableFlowBackendE2E`** — end-to-end: create a GitHub connection directly in SQL (bypassing the OAuth flow), seed a `connection_sources` row like the discover endpoint would, enable it via `PUT /sources`, then call `ListEnabledGitHubReposByApp` the way the agent does and confirm the row + connection config come through.
- **`TestSourceFilters_GitHubDisabledNotReturned`** — the inverse: `enabled=false` rows stay invisible to the agent, so the tool never queries a repo the user opted out of.
- **`TestDiscoverSources_RejectsNonGitHub`** — POST `/sources/discover` on a webhook_logs connection returns 400 with a message naming the connection type.
- **`TestDiscoverSources_GitHubAppNotConfigured`** — a GitHub connection whose server has no `s.GitHub` configured returns 502 with an informative message rather than crashing.
- **`TestSourceFilters_RoutesDroppedForGitHubRepos`** — `GET /connections/{id}/github/repos` returns 404; there's no route for it anymore.
- **`TestListEnabledGitHubReposByApp_QuerySignature`** — compile-time guard that the three fields the agent reads (`ConnectionID`, `RepoFullName`, `ConnectionConfig`) still exist on the generated row struct.

Plus all Phase 1 + Phase 2 tests continue to pass (`TestIngestWebhookLogs*`, `TestSourceFilters_*`, `TestCreateConnection_OrgScoped`, etc.) — the filter infrastructure they depend on is untouched by Phase 3; this phase just adds another shape of client on top.

Frontend tests: 58 existing tests pass; no new ones added for Phase 3 (the selector's behaviour is exercised by the same unit tests that covered it before, and the `discoverable` branches are thin UI).

---

## What Did NOT Change (and why)

### No change to `connection_sources` schema

Earlier sketches added a `metadata JSONB` column to carry `default_branch` and `repo_id` through the migration. That turned out to be overkill — neither field is read by any consumer. Keeping the table lean matters: the metadata shape would have become a junk drawer for connector-specific debris, and every future connector would have been tempted to shove its own fields in.

If Phase 4 (or beyond) needs per-connector structured data on a source row, that's a separate migration with a concrete motivating consumer. We won't add the column speculatively.

### No automatic sync on connection creation

Right now the user has to click "Sync" (or open the selector, which auto-syncs when empty). A future nicety would be to trigger discovery inside the OAuth callback so the repo list is pre-populated by the time the user lands on the connections page. Out of scope for Phase 3 — the current UX matches the pre-Phase-3 behaviour (where `ListGitHubRepos` also fetched on open) and the incremental value isn't worth adding a new backend dependency into the callback path.

### No backwards-compat endpoint aliases

`/api/connections/{id}/github/repos` is gone, not redirected. There are no external consumers — Heimdall is still pre-1.0, the endpoint only ever served our own frontend — so a 404 is the honest answer. Documented in the migration notes.

### No tests around the migration rollback

The up-migration is idempotent (`ON CONFLICT DO NOTHING`), and the down migration is documented as destructive-if-reversed-on-production — synthesising negative `repo_id`s to rebuild the UNIQUE constraint. There's no automated test exercising the full down/up cycle because doing so in the test DB would interfere with every other test's data.

### No change to the agent's system prompt

The Phase 1 system-prompt injection ("paused connections can't be queried") is unchanged. Phase 3 doesn't alter how the agent reasons about connections, only how it locates GitHub repo names internally.

---

## Risks & Mitigations

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| An installed GitHub org has more than 5,000 repos and hits the pagination cap | Low | Same ceiling as pre-Phase-3. If it matters, raise `maxPages` in `discoverGitHubRepos`. |
| User disables a repo, agent already has a tool trace with its name | Negligible | The agent queries on every invocation; disabled repos vanish on the next run. |
| GitHub renames a repo (`org/old` → `org/new`) between syncs | Expected | Both names end up in `connection_sources`; the old one goes stale (no traffic → `last_seen_at` doesn't refresh). The enabled filter stays on the old name until the user manually migrates. This is the same behaviour a renamed Fly app would exhibit. Low-severity UX issue — the user notices and fixes. |
| Future rollback of 036 against a DB that's accumulated new org-scoped GitHub connections (hypothetical) | Nonexistent today (supportsOrgScope excludes github) | Down migration's WHERE filter captures only `c.type = 'github'`; rollback handles whatever state is present. |

---

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `npx vue-tsc --noEmit` — clean
- `make migrate-up` — migration 036 applied cleanly on the local Supabase instance
- Backend tests: 6 Phase 3 + 5 Phase 2 + 13 Phase 1 = 24 relevant tests, all pass
- Frontend tests: 58 unit tests, all pass
- 11 pre-existing unrelated handler-test failures remain (tracked separately, unchanged by Phase 3)

---

## File Map

```
backend/
  migrations/036_github_to_generic_sources.up.sql      NEW
  migrations/036_github_to_generic_sources.down.sql    NEW
  internal/db/queries/github_repos.sql                 DELETED
  internal/db/queries/source_filters.sql               CHANGED — ListEnabledGitHubReposByApp rewritten here, 3-field projection
  internal/db/github_repos.sql.go                      DELETED (regenerated away)
  internal/db/models.go                                REGENERATED — GithubRepo struct removed
  internal/db/source_filters.sql.go                    REGENERATED — ListEnabledGitHubReposByApp uses the new tables
  internal/api/handlers/github_repos.go                DELETED
  internal/api/handlers/github_install.go              CHANGED — TestGitHubConnection moved here + context import added
  internal/api/handlers/source_filters_discover.go     NEW — DiscoverSources handler + discoverGitHubRepos helper
  internal/api/handlers/source_filters_github_test.go  NEW — 6 regression/flow tests
  internal/api/router.go                               CHANGED — /github/repos routes removed, /sources/discover added
  internal/api/handlers/testhelpers_test.go            CHANGED — mirror router changes

frontend/
  src/types/github.ts                                  DELETED
  src/components/connections/GitHubRepoSelector.vue    DELETED
  src/api/github.ts                                    CHANGED — repo-listing functions removed; install-URL retained
  src/api/sources.ts                                   CHANGED — discoverSources added
  src/components/connections/SourceSelector.vue        CHANGED — discoverable prop, Sync button, mount-time auto-sync, manual-add/banner gated off in discoverable mode
  src/components/connections/ConnectionDetailModal.vue CHANGED — manage-repos collapsed into manage-sources, SOURCE_FILTERED_TYPES includes 'github'
  src/components/connections/wizard/ConnectionWizard.vue CHANGED — emits manage-sources
  src/pages/ConnectionsPage.vue                        CHANGED — removes GitHubRepoSelector, picks discoverable mode based on connection.type
```

---

## Handoff

Phase 4 remains. It extends source-name extraction to OTLP (`service.name`), syslog (hostname), and generic webhooks (configurable JSONPath or `source_type` fallback). None of that depends on Phase 3 internals — it just broadens what gets upserted into `connection_sources` at ingestion time, then the existing selector/filter machinery handles the rest. Phase 3's `discoverable` hatch is also a useful hook if any future connector grows its own "list what's accessible" API (a Vercel projects endpoint, for example).
