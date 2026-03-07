# Phase 8, Step 1 — Organizations & Applications: Completion Notes

Reference: [Phase 8 Implementation Plan](../executing/phase-8-monitoring-mode.md)

---

## What was done

Introduced the foundational data model for multi-app monitoring: Organizations, Applications, and per-application agent configuration. This replaces the flat user-scoped model with an org hierarchy: **User -> Organization -> Application -> Connections**.

---

## Migration 014 — Organizations & Applications

| Action | File | Notes |
|--------|------|-------|
| Created | `backend/migrations/014_organizations_applications.up.sql` | Creates `organizations` table, adds `org_id` to `users`, creates `applications` table, adds `app_id` to `connections` |
| Created | `backend/migrations/014_organizations_applications.down.sql` | Reverses all changes in order |

### Tables created

**`organizations`**
- `id` UUID PK, `name` TEXT, `slug` TEXT UNIQUE, `created_at`, `updated_at`
- Unique index on `slug` for URL-friendly org identifiers

**`applications`**
- `id` UUID PK, `org_id` UUID FK -> organizations (CASCADE), `name` TEXT, `status` TEXT (default `'active'`), `created_at`, `updated_at`
- Index on `org_id` for org-scoped lookups

### Columns added

**`users.org_id`** — `UUID REFERENCES organizations(id) ON DELETE SET NULL`
- Nullable: users without an org haven't completed onboarding yet
- Index on `org_id`

**`connections.app_id`** — `UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE`
- NOT NULL is safe because migration wipes existing test data (`DELETE FROM log_buffer; DELETE FROM connections;`)
- Index on `app_id`

**Design decision:** Clean break on existing data. All connections and log_buffer rows are deleted before adding the NOT NULL `app_id` column. This avoids backfill complexity and is safe because all existing data is test data (per Phase 8 plan, Decision #2).

---

## Migration 015 — Per-Application Agent Config

| Action | File | Notes |
|--------|------|-------|
| Created | `backend/migrations/015_app_agent_config.up.sql` | Creates `app_agent_config` table with RLS policy |
| Created | `backend/migrations/015_app_agent_config.down.sql` | Drops policy and table |

### Table created

**`app_agent_config`**
- `app_id` UUID PK (FK -> applications, CASCADE) — 1:1 with applications
- `model` TEXT (default `'claude-sonnet-4-6'`)
- `mode` TEXT (default `'continuous'`) — continuous, periodic, off
- `schedule_interval_secs` INTEGER (default 60) — simple seconds instead of cron
- `system_prompt_override` TEXT (nullable)
- `created_at`, `updated_at`

### RLS policy

`app_agent_config_org` — users can access configs for apps in their org. Uses a subquery joining `applications` -> `users` via `org_id`, checked against `app_current_user_id()`.

**Design decision:** Uses `schedule_interval_secs` (integer) rather than cron expressions. Both are interval-based; cron adds day/hour specificity we don't need yet. Simpler to validate, display, and reason about.

---

## sqlc Queries

| Action | File | Queries |
|--------|------|---------|
| Created | `backend/internal/db/queries/organizations.sql` | `CreateOrganization`, `GetOrganization`, `GetOrganizationBySlug`, `GetOrganizationByUser`, `UpdateOrganization` |
| Created | `backend/internal/db/queries/applications.sql` | `CreateApplication`, `GetApplication`, `ListApplicationsByOrg`, `UpdateApplication`, `DeleteApplication` |
| Created | `backend/internal/db/queries/app_agent_config.sql` | `GetAppAgentConfig`, `UpsertAppAgentConfig` |
| Updated | `backend/internal/db/queries/users.sql` | `GetUser` now includes `org_id`; added `SetUserOrg` |

### Generated Go code

| File | Status |
|------|--------|
| `backend/internal/db/organizations.sql.go` | New — 5 query methods |
| `backend/internal/db/applications.sql.go` | New — 5 query methods |
| `backend/internal/db/app_agent_config.sql.go` | New — 2 query methods |
| `backend/internal/db/users.sql.go` | Regenerated — `GetUser` returns `GetUserRow` with `OrgID` field, new `SetUserOrg` method |
| `backend/internal/db/connections.sql.go` | Regenerated — `Connection` model now includes `AppID` field |
| `backend/internal/db/models.go` | Regenerated — new `Organization`, `Application`, `AppAgentConfig` models; `User` and `Connection` models updated |

---

## Impact on existing code

- **`GetUser` return type** now includes `org_id` (as `pgtype.UUID`, nullable). The `Me` handler JSON-encodes the result directly, so the API response now includes `"org_id": null` for users without an org. This is beneficial — the frontend can use it to detect onboarding state.
- **`Connection` model** now includes `AppID`. Existing connection queries still work (they don't SELECT `app_id` explicitly since they use `SELECT *`), but `CreateConnection` will need updating in a future step to accept `app_id`.
- **No handler changes** in this step — handlers, routes, and frontend are unchanged. The old global `agent_config` table still exists and works; `app_agent_config` is additive.

---

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `sqlc generate` — clean, no warnings
