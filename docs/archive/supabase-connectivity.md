# Supabase Connectivity — API Polling Connector

## Context

Heimdall currently supports two ways to ingest logs from Supabase:

1. **Direct Postgres connection** (port 5432) — gives the agent a `query_database` tool for investigation, but doesn't passively receive logs.
2. **Webhook Logs + Supabase Log Drains** — Supabase pushes logs to Heimdall's `POST /api/webhooks/logs` endpoint in real time. However, Log Drains cost **$60/connection/month** and are only available on Pro+.

Neither option gives cost-conscious users passive log ingestion. This spec introduces a third path: **polling the Supabase Management API** to pull logs on an interval — available on all plans, no add-on required.

## The Supabase Management API

Supabase exposes the same log backend that powers their dashboard Logs Explorer:

```
GET https://api.supabase.com/v1/projects/{ref}/analytics/endpoints/logs.all
```

### Authentication

- **Personal Access Token (PAT)** — generated at `supabase.com/dashboard/account/tokens`
- Header: `Authorization: Bearer <PAT>`
- This is distinct from the project's `anon` / `service_role` keys — it operates at the account level

### Query Interface

The endpoint accepts a `sql` parameter with BigQuery-standard SQL. Available tables:

| Table | Content |
|-------|---------|
| `postgres_logs` | Database query and error logs |
| `auth_logs` | Authentication/authorization events |
| `edge_logs` | API gateway request/response logs |
| `function_logs` | Edge Function console output |
| `function_edge_logs` | Edge Function invocation metadata |
| `storage_logs` | Object storage operations |
| `realtime_logs` | WebSocket connection logs |

Example query:
```sql
SELECT timestamp, event_message, metadata
FROM postgres_logs
WHERE timestamp > '2026-03-21T12:00:00Z'
ORDER BY timestamp ASC
LIMIT 500
```

### Constraints

| Constraint | Detail |
|-----------|--------|
| Max query window | 24 hours per request |
| Rate limit | ~120 req/min globally; analytics endpoints may be stricter |
| Row limit | 1,000 rows per query |
| Log retention | Free: 1 day, Pro ($25/mo): 7 days, Team ($599/mo): 28 days |

## Design

### New Connection Type: `supabase`

A new connector type that combines API polling for log ingestion with optional direct Postgres access for agent investigation.

### Connection Config

```json
{
  "project_ref": "abcdefghijklmnopqrst",
  "access_token": "<personal-access-token>",
  "poll_tables": ["postgres_logs", "auth_logs"],
  "poll_interval_secs": 30
}
```

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `project_ref` | Yes | — | 20-char Supabase project reference ID |
| `access_token` | Yes | — | Personal Access Token for Management API |
| `poll_tables` | No | `["postgres_logs"]` | Which log tables to poll |
| `poll_interval_secs` | No | `30` | Polling interval in seconds (min 15) |

### Polling Loop

A background goroutine per active Supabase connection:

1. **Tick** every `poll_interval_secs`
2. For each table in `poll_tables`:
   - Query: `SELECT timestamp, event_message, metadata FROM {table} WHERE timestamp > $cursor ORDER BY timestamp ASC LIMIT 500`
   - Parse response rows
3. **Insert** each row into `log_buffer` via `InsertLogEntry()`:
   - `source_type`: table name (e.g. `supabase/postgres_logs`, `supabase/auth_logs`)
   - `severity`: derive from log level if present, otherwise omit
   - `payload`: full log row as JSONB
4. **Advance cursor** to the latest `timestamp` seen
5. Existing monitoring loop picks up entries from `log_buffer` as usual

### Cursor Persistence

Store the last-polled timestamp per table in the connection's config or in `monitoring_state`. On restart, resume from the stored cursor to avoid re-ingesting logs.

### Rate Limit Handling

- Respect `X-RateLimit-Remaining` / `X-RateLimit-Reset` response headers
- On HTTP 429, back off and retry after the reset window
- With a 30s default interval and 2 tables, that's ~4 req/min — well within the 120 req/min global limit

### Error Handling

- API errors (401, 403, 5xx) → log the error, set connection status to `error`, continue polling on next tick
- PAT expiry → surface in the UI via connection status; user regenerates token in Supabase dashboard

## User Flow

1. Navigate to Connections → Add Connection
2. Select type **Supabase**
3. Enter project reference ID and Personal Access Token
4. Optionally select which log tables to poll (default: `postgres_logs`)
5. Heimdall starts polling immediately; logs appear in the Agent Log within one polling interval

### Complementary Setup

For full observability + investigation capability, users would create **two** connections:

| Connection | Type | Purpose |
|-----------|------|---------|
| Supabase Logs | `supabase` | Passive log ingestion via API polling |
| Production DB | `postgres` | Agent investigation via `query_database` tool |

## Frontend Changes

### ConnectionForm.vue

Add `supabase` to the type selector. Config fields:

- **Project Reference** — text input, required
- **Access Token** — password input, required
- **Log Tables** — multi-select checkboxes for the 7 available tables
- **Poll Interval** — number input with presets (15s, 30s, 60s)

### Connection Test

Test the connection by making a single API call:
```
GET /v1/projects/{ref}/analytics/endpoints/logs.all?sql=SELECT 1
```
If it returns 200, the PAT is valid and the project ref is correct.

## Backend Changes

| Area | Change |
|------|--------|
| `connectors/` | New `supabase/` package implementing a polling connector |
| `handlers/connections.go` | Add `supabase` to type validation; test via Management API health check |
| `agent/monitor.go` | No changes — the monitoring loop already reads from `log_buffer` regardless of how entries arrived |
| Config | No new env vars — credentials stored per-connection in JSONB config |

## Why Not Just Use Webhook Logs?

Users *can* use the existing `webhook_logs` connector with Supabase Log Drains today. But:

- Log Drains cost **$60/connection/month** on top of the Pro plan
- Many early-stage users won't pay for this when they're evaluating Heimdall
- The Management API is **free on all plans** (limited by retention, not access)
- A first-class `supabase` connector type is a better UX than telling users to configure a webhook + a separate Log Drain

## Trade-offs

| | Log Drain (webhook) | API Polling (this spec) |
|---|---|---|
| Latency | Real-time (push) | Polling interval (15–60s) |
| Cost | $60/connection/month | Free |
| Setup complexity | Configure in both Supabase + Heimdall | Configure in Heimdall only |
| Reliability | Supabase guarantees delivery | Polling may miss logs if retention < poll gap |
| Rate limits | None (Supabase pushes) | ~120 req/min shared |

For a monitoring agent that already processes logs in 15-second batches, the polling delay is negligible. The cost savings make this the default recommendation for most users.
