# Schema-Drift Audit

## Concern

We use **golang-migrate**, which tracks only `(version, dirty)` in `schema_migrations` — it does **not** checksum applied migrations. If a migration file was edited after being applied to a database, the DB and the repo will silently disagree forever; re-running `migrate up` won't notice, because `version N` is already marked applied.

This doc captures the investigation into whether that happened in this repo.

## Approach

Two independent checks, in increasing cost:

1. **Git audit (cheap).** Migrations should be append-only. Any migration file whose git history shows more than one commit is a yellow flag — it *may* have been edited after being applied. Also look at what changed: whitespace-only or comment-only edits are harmless; DDL changes (columns, constraints, indexes, policies) are real drift candidates.

2. **Rebuild-and-diff (authoritative).** Apply all migrations to a fresh Postgres, `pg_dump --schema-only` it, `pg_dump` prod (or staging, or any long-lived DB of interest), and diff the two. Any difference is drift, regardless of cause — it also catches out-of-band DDL run directly against prod, which the git audit cannot see. *Not run in this investigation — requires a live DB and was out of scope for a read-only pass.*

## Findings — Git Audit

Ran (abbreviated):

```bash
for f in backend/migrations/*.sql; do
  commits=$(git log --oneline --follow -- "$f" | wc -l)
  [ "$commits" -gt 1 ] && echo "MODIFIED: $f"
done
```

**Result:** 4 files (all `.up.sql`, no `.down.sql`) have more than one commit. All four were modified in the **same single commit** (`bf9c7bf`, 2026-02-20 05:18 +0800) approximately 13 hours after the initial scaffolding commit (`0825e65`, 2026-02-19 16:18 +0800).

| Migration | Change |
|---|---|
| `001_create_connections.up.sql` | **+** Added `CREATE INDEX idx_connections_status ON connections (status)`. |
| `002_create_agent_config.up.sql` | **~** Changed primary key from `id UUID DEFAULT gen_random_uuid()` to `id INTEGER DEFAULT 1 CHECK (id = 1)` (singleton-row pattern). |
| `004_create_conversations.up.sql` | **~** Added `ON DELETE SET NULL` to the `investigation_id` FK. |
| `005_create_log_buffer.up.sql` | **~** Added `ON DELETE CASCADE` to the `connection_id` FK. |

All four are **semantically meaningful schema changes**, not cosmetic. None of the 70 migrations from `006` onward show a second commit.

### Risk assessment

Whether these edits caused drift depends entirely on a single question:

> **Was any Postgres database — local, staging, or production — created with migration versions 001/002/004/005 applied *before* commit `bf9c7bf` on 2026-02-20?**

Two scenarios:

- **If no such DB exists** (the scaffolding commit was never deployed; it was a 13-hour in-progress state on one developer's machine, superseded by `bf9c7bf` before any DB was provisioned): **no drift.** The "edit" is effectively a rebase of an unused local state.
- **If any such DB exists** (e.g. a Supabase project was created from `0825e65` and carried forward): **confirmed drift on those four objects.** Specifically:
  - `connections` would be missing `idx_connections_status`.
  - `agent_config.id` would be `UUID`, not `INTEGER` with the singleton `CHECK`.
  - `conversations.investigation_id` FK would be `NO ACTION` on delete, not `SET NULL`.
  - `log_buffer.connection_id` FK would be `NO ACTION` on delete, not `CASCADE`.

The timing is suggestive — both commits fall inside the initial 24-hour scaffolding period, and the project was not publicly announced until the `0.9.0 — Public Website` milestone (2026-02-27) — but git cannot *prove* no pre-existing DB was touched. **Only the rebuild-and-diff check can resolve this definitively.**

### No other drift signals in git

- No down-migrations modified (zero hits).
- No migrations 006–037 modified. The append-only discipline has held since the project's second day.

## Recommended next steps

1. **Confirm DB provenance.** For each environment with a live database (prod on Supabase, any staging, any long-lived local dev DB that hasn't been reset), check `SELECT version FROM schema_migrations;` and ask: was this DB first provisioned before or after 2026-02-20 05:18 +0800? If after — we're done; no drift possible from this audit.
2. **If any "before" DB exists, run the rebuild-and-diff:**
   ```bash
   # Fresh expected schema
   docker run -d --name drift-pg -e POSTGRES_PASSWORD=x -p 55432:5432 postgres:16
   DATABASE_URL="postgres://postgres:x@localhost:55432/postgres?sslmode=disable" make migrate-up
   pg_dump --schema-only --no-owner --no-privileges \
     "postgres://postgres:x@localhost:55432/postgres" > /tmp/expected.sql

   # Actual schema (replace with real URL)
   pg_dump --schema-only --no-owner --no-privileges "$PROD_URL" > /tmp/actual.sql

   diff -u /tmp/expected.sql /tmp/actual.sql
   ```
   Or use [`migra`](https://github.com/djrobstep/migra) for a semantic diff that emits the reconciliation SQL directly.
3. **Focus the diff on high-signal sections.** `pg_dump` output has noise (extension versions, `search_path` lines, `COMMENT ON EXTENSION` ordering). What matters is:
   - `CREATE TABLE` column types, NOT NULL, defaults
   - `CHECK` constraints
   - Foreign-key `ON DELETE` / `ON UPDATE` actions
   - `CREATE INDEX`
   - `CREATE POLICY` (RLS — we have a lot of these as of 0.43.0 and 0.45.1)
   - Triggers, functions

## Prevention going forward

- **Append-only CI guard.** Add a check that fails any PR which modifies a migration file whose version is ≤ the max version already on `master`. Cheap, catches the exact failure mode that golang-migrate can't detect.
- **Optional: switch tooling.** `dbmate`, `atlas`, or Flyway all checksum applied migrations and refuse to run if a file's hash has drifted from what was recorded at apply-time. Larger change, not urgent.
- **Periodic rebuild-and-diff in CI.** Cron a weekly job that builds the expected schema from migrations and diffs it against a staging snapshot. Catches out-of-band DDL that never went through migrations at all.
