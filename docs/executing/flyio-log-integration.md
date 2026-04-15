# Fly.io Log Integration — Research & Proposal

## Current State

Heimdall already has partial Fly.io support, but it's disconnected between frontend and backend:

| Layer | State | Detail |
|-------|-------|--------|
| **Backend connector** | Exists (`connectors/logs/flyio.go`) | Polls Machines API per-machine. Type: `flyio`. |
| **Backend validation** | Exists (`connections_validate.go:41`) | Validates `flyio` config (app_name, api_token). |
| **Poller factory** | Wired (`factory.go:24`) | `flyio` → `logs.NewFlyio`, starts with 30s default interval. |
| **Frontend wizard** | Stub only (`wizard/flows.ts:89`) | `available: false`, empty steps, **mislabelled** as `connectorType: 'syslog'` instead of `'flyio'`. |

The backend connector works in isolation, but the frontend never exposes it. Additionally, the current backend approach has significant limitations (see below).

---

## How the Existing Backend Poller Works

**File:** `backend/internal/connectors/logs/flyio.go`

1. Lists all machines via `GET /v1/apps/{app}/machines` (Machines API at `api.machines.dev`)
2. Filters to `started` / `running` machines only
3. For each machine, fetches `GET /v1/apps/{app}/machines/{id}/logs?limit=200`
4. Deduplicates by tracking a cursor timestamp (last-seen timestamp + 1ns)
5. Inserts into `log_buffer` with `source_type = "flyio/{app_name}"`

**Config fields:** `app_name`, `api_token`, `poll_interval_secs` (default 30s, min 15s)

### Limitations of the Current Approach

| Issue | Severity | Detail |
|-------|----------|--------|
| **N+1 API calls per poll** | High | One call to list machines + one call per running machine. A 10-machine app makes 11 API calls every 30 seconds. |
| **Misses stopped/crashed machine logs** | High | Only polls `started` and `running` machines. Crash logs — arguably the most important — are silently dropped. |
| **Per-machine endpoint is undocumented** | Medium | `/v1/apps/{app}/machines/{id}/logs` is not in Fly.io's public docs. Could break without notice. |
| **No app-level logs endpoint** | Medium | Fly.io has a separate, more stable app-level logs API (see Option A below) that returns logs across all machines. |
| **200-entry cap per machine per poll** | Medium | High-throughput machines could exceed 200 log lines in 30 seconds, causing silent data loss. |
| **Rate limiting risk** | Medium | 429 responses are handled (logs warning, retries next interval) but N+1 calls increase the risk surface. |
| **No historical backfill** | Low | Cursor starts at `now() - 5min` on first connect. No way to request older logs. |

---

## Fly.io Log Export Options

Fly.io offers five ways to get logs out. Here's how each maps to Heimdall:

### Option A — App-Level Logs API (Poller, improved)

**Fly.io endpoint:** `GET https://api.machines.dev/v1/apps/{app}/logs` (or the older `GET /api/v1/apps/{app_name}/logs`)

Fly.io has an app-level logs endpoint that returns NDJSON (newline-delimited JSON) across all machines. This is what `fly logs` uses under the hood.

**How it maps to Heimdall:**
- Replace the current per-machine N+1 polling with a single API call per poll cycle
- Returns logs from all machines (including recently stopped/crashed)
- Supports `start_time` parameter for cursor-based polling
- ~15 days of history available for initial backfill
- NDJSON format (line-by-line JSON parsing)

**Effort:** Small — rewrite `Poll()` and `pollMachineLogs()` in `flyio.go` (~100 lines changed). No new infrastructure, no DB migration, no frontend changes beyond enabling the wizard.

**Trade-offs:**
- Still polling (30s latency floor)
- Endpoint is "mostly stable" but not officially documented — Fly considers it semi-public since `flyctl` depends on it
- Rate limits still apply (but 1 call vs N+1 is a major improvement)

---

### Option B — Fly Log Shipper → Heimdall Webhook

**How it works:** Deploy [flyio/log-shipper](https://github.com/superfly/fly-log-shipper) as a Fly Machine in the user's org. It's a pre-built Vector container that connects to Fly's internal NATS log stream and forwards to configurable sinks.

Configure the HTTP sink to POST to Heimdall's existing webhook endpoint (`POST /api/webhooks/logs`).

**Setup for the user:**
```bash
fly launch --image flyio/log-shipper:latest --no-public-ips
fly secrets set ACCESS_TOKEN=$(fly auth token)
fly secrets set HTTP_URL=https://<heimdall-host>/api/webhooks/logs
fly secrets set HTTP_TOKEN=<heimdall-webhook-token>
```

**How it maps to Heimdall:**
- Logs arrive at the existing `POST /api/webhooks/logs` endpoint
- Connection type would be `webhook_logs` (already fully supported)
- Would need a webhook parser for the Vector/Log Shipper JSON format (or configure Vector to output Heimdall's native format)
- Near-real-time delivery (sub-second, NATS-backed)

**Effort:** Small-to-medium — may need a new webhook parser in `webhook_parsers.go` for the Vector HTTP sink format, plus documentation / wizard step. No new connector code; reuses the webhook pipeline.

**Trade-offs:**
- Near-real-time (much lower latency than polling)
- Requires the user to deploy and manage a Log Shipper machine in their Fly org (~$2/month)
- More operational overhead for the user vs a pure API-based approach
- Very reliable — NATS-backed with at-least-once delivery

---

### Option C — Fly Log Shipper → Heimdall Syslog

Same as Option B, but configure the Log Shipper's syslog sink to send RFC 5424 messages to Heimdall's syslog TLS listener.

**How it maps to Heimdall:**
- Logs arrive at Heimdall's existing syslog listener (`connectors/logs/syslog.go`)
- Connection type would be `syslog` (already fully supported)
- No new backend code needed at all

**Effort:** Minimal backend work — documentation and wizard guidance only. Heimdall's syslog listener already handles RFC 5424.

**Trade-offs:**
- Same operational overhead as Option B (user deploys Log Shipper)
- Syslog is a well-defined protocol — no format compatibility risk
- Requires Heimdall's syslog listener to be publicly reachable (needs a public IP or Fly private networking)
- Loses structured JSON metadata (syslog flattens to message string)

---

### Option D — Fly Log Shipper → Heimdall OTLP

Same as Option B, but route through Heimdall's OTLP endpoint (`POST /api/v1/logs`).

**How it maps to Heimdall:**
- Vector doesn't have a native OTLP sink, so this would require an intermediate OpenTelemetry Collector or a custom Vector transform
- Connection type would be `otlp` (already supported)

**Effort:** Medium-high — requires either a sidecar OTel Collector or custom Vector config. Not a natural fit.

**Trade-offs:**
- Adds unnecessary complexity for no clear benefit over Option B (webhook) or C (syslog)
- OTLP format is heavier than needed for simple log forwarding
- **Not recommended** as a primary path

---

### Option E — Direct NATS Connection

Connect directly to Fly.io's internal NATS cluster at `nats://[fdaa::3]:4223`.

**How it maps to Heimdall:**
- Would require a new `StreamConnector` implementation using a NATS client library
- Heimdall's backend would need to be inside Fly's private network (WireGuard VPN) or deployed on Fly itself
- Lowest possible latency — raw message-level streaming

**Effort:** High — new NATS dependency, new connector type, WireGuard networking complexity, deployment constraints.

**Trade-offs:**
- Only works if Heimdall backend is on Fly.io (it is today — but this couples the architecture)
- NATS endpoint is internal/undocumented — no stability guarantee
- **Not recommended** unless Heimdall is permanently Fly-exclusive

---

## Recommendation — Dual-Mode Wizard

**Offer both drain and polling as first-class options in a single Fly.io wizard flow.**

Fly.io doesn't have a native log drain API (unlike Heroku's `heroku drains:add`). The drain mechanism is the **Fly Log Shipper** — a pre-built Vector container the user deploys in their Fly org. This means drain setup requires ~4 CLI commands on the user's side. We can't fully automate it, but we can make it seamless with clear copy-paste instructions.

### Wizard Flow

```
Step 1: Name        → Connection name
Step 2: Mode        → "Log Drain (recommended)" or "API Polling"
Step 3a (drain):    → Creates webhook_logs connection, shows URL + token + setup instructions
Step 3b (polling):  → Collects app_name + api_token for flyio connection
Step 4:             → Test connection
```

The key insight: **the drain path creates a `webhook_logs` connection** (reusing existing infra), while the **polling path creates a `flyio` connection** (using the existing poller). The wizard dynamically switches `connectorType` based on the user's mode choice.

### Drain Path (Log Shipper → Webhook)

When the user selects "Log Drain", Heimdall:
1. Creates a `webhook_logs` connection with an auto-generated bearer token
2. Shows the webhook URL and token
3. Shows copy-paste CLI commands to deploy the Fly Log Shipper:

```bash
fly launch --image flyio/log-shipper:latest --no-public-ips
fly secrets set ACCESS_TOKEN=$(fly auth token)
fly secrets set HTTP_URL=https://<heimdall-host>/api/webhooks/logs
fly secrets set HTTP_TOKEN=<generated-token>
```

**Backend work:** May need a Vector HTTP sink webhook parser in `webhook_parsers.go` if the default format doesn't auto-detect. Otherwise, zero backend changes — the webhook pipeline already exists.

### Polling Path (API Polling)

When the user selects "API Polling", Heimdall:
1. Collects `app_name` and `api_token`
2. Creates a `flyio` connection
3. Starts the poller (improved to use app-level endpoint)

**Backend work:** Rewrite `Poll()` in `flyio.go` to use the app-level logs endpoint instead of N+1 per-machine calls.

### Task Breakdown

1. **Rewrite Fly.io API poller** — Replace N+1 per-machine polling with single app-level endpoint call
2. **Redesign wizard flow** — Mode selection step, dynamic connectorType switching
3. **Create wizard components** — StepFlyioMode (drain/polling toggle), StepFlyioAuth (polling creds), StepFlyioDrainSetup (webhook URL + instructions)
4. **Vector webhook parser** — Test and add if needed
5. **Test end-to-end** — Both paths, builds, existing tests

### Decisions Made

- **No backfill** — start from now, not historical
- **Multiple Fly apps** — supported naturally (one connection per Fly app, users can add multiple)
- **All regions** — no region filtering for now
- **Drain is primary** — recommended in the UI, polling is the zero-setup fallback

---

## Implementation Detail: App-Level Logs Endpoint (for polling path)

The app-level endpoint returns NDJSON (one JSON object per line):

```
GET https://api.machines.dev/v1/apps/{app}/logs?start_time=2026-04-15T00:00:00Z
Authorization: Bearer {token}
```

Each line:
```json
{"timestamp":"2026-04-15T12:00:00.123Z","message":"request completed","level":"info","instance":"d896d7c609dd68","region":"lhr","meta":{"event":{"provider":"app"}}}
```

Key fields: `timestamp`, `message`, `level`, `instance` (machine ID), `region`, `meta`.

The existing `normalizeSev()` function in `flyio.go` already handles the level mapping correctly.

## Implementation Detail: Vector HTTP Sink Format (for drain path)

Vector's HTTP sink sends JSON POST requests. Default format (when using `encoding.codec = "json"`):

```json
{"host":"my-machine","message":"log line here","source_type":"stdin","timestamp":"2026-04-15T12:00:00.123Z"}
```

This may or may not auto-detect via the existing webhook parsers. Needs testing — if not, add a Vector/Log Shipper parser.
