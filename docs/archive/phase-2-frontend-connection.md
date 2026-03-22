# Phase 2 Completion — Supabase Frontend Connection Creation

Phase 2 adds Supabase as a connection type in the frontend UI, allowing users to create and manage Supabase connections through the existing form and card patterns.

---

## What Was Built

### 1. Supabase Type in ConnectionForm

**File:** `frontend/src/components/connections/ConnectionForm.vue` (modified)

**Type dropdown:** Added `{ value: 'supabase', label: 'Supabase' }` to the type options, positioned second after PostgreSQL.

**Config fields** added to the `configFields` record:

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `project_ref` | text | yes | — | Supabase project reference ID |
| `access_token` | password | yes | — | Personal Access Token (`sbp_...`) |
| `poll_interval_secs` | select | no | 30 | Options: 15s, 30s, 60s |

**Poll tables checkbox group:** A dedicated template section renders when type is `supabase`, showing checkboxes for all six Supabase log tables:

| Table | Description |
|-------|-------------|
| `postgres_logs` | Database queries and errors |
| `auth_logs` | Authentication events |
| `edge_logs` | API gateway requests |
| `function_logs` | Edge Function output |
| `storage_logs` | Object storage operations |
| `realtime_logs` | WebSocket connections |

Default selection: `postgres_logs` and `auth_logs`.

**Automatic behaviours:**

- Selecting Supabase type auto-locks direction to `one_way` (ingestion-only).
- Checkbox state resets to defaults on type change.
- When editing an existing Supabase connection, `selectedTables` is populated from `config.poll_tables`.
- On submit, `poll_tables` (string array) and `poll_interval_secs` (cast to number) are merged into the config payload.

**Helper text:** Below the access token field:
> Uses the Supabase Management API — works on all plans including Free. Generate a Personal Access Token at supabase.com/dashboard/account/tokens.

**Why a separate template block** instead of using the generic `activeFields` loop: The poll tables field is a multi-select checkbox group (array value), which doesn't fit the existing scalar `FieldDef` type. Rather than complicating the shared field system, Supabase gets its own template section that renders both the standard fields and the checkbox group.

---

### 2. ConnectionCard Supabase Display

**File:** `frontend/src/components/connections/ConnectionCard.vue` (modified)

**Type label mapping:** Added a `typeLabels` record that maps raw type strings to display names (`postgres` -> `PostgreSQL`, `supabase` -> `Supabase`, etc.). All connection types now show human-readable names in cards.

**Supabase subtitle:** Instead of showing the direction (`one_way`), Supabase cards show the polled table count:
- `"Polling 2 tables"` (plural)
- `"Polling 1 table"` (singular)
- `"Polling logs"` (fallback if no tables in config)

Other connection types continue showing their direction as before.

---

## Files Summary

### Modified Files

| File | Change |
|------|--------|
| `frontend/src/components/connections/ConnectionForm.vue` | Added Supabase type, config fields, poll_tables checkbox group, helper text, auto direction lock |
| `frontend/src/components/connections/ConnectionCard.vue` | Human-readable type labels, Supabase-specific subtitle with table count |

### No New Files

Phase 2 modifies existing components only — no new files created.

---

## Verification

- `npm run build` — compiles cleanly (vue-tsc + vite)
- `eslint src/` — zero new errors (1 pre-existing `no-explicit-any` in unrelated GitHub install handler)
- `ConnectionCard.vue` — passes lint with zero errors

---

## What's Next

- **Phase 3:** End-to-end validation with a real Supabase project, error scenario testing, backend + frontend unit tests
- **Phase 4:** Connection creation wizard (independent of Phase 3)
