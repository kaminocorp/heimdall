# Phase 1 Completion — Syslog TLS Listener

**Date:** 2026-04-05
**Roadmap ref:** `docs/executing/ingestion-roadmap.md` — Phase 1, item 1
**Status:** Complete

---

## What was built

A production-ready TCP/TLS syslog listener that accepts RFC 5424 and RFC 3164 messages over persistent TCP connections and inserts them into `log_buffer`. This completes Phase 1 of the ingestion roadmap (Supabase API poller was already done in v0.18.0).

---

## Architecture decisions

### Listener, not poller

Syslog is fundamentally different from the Supabase connector. Supabase uses `PollConnector` (timer-driven pulls from an API), but syslog is a long-running TCP server that accepts inbound connections. This required a new concurrency primitive: the **ListenerManager** — analogous to `Poller` but for persistent network listener goroutines.

### No external dependencies

The implementation uses Go's standard library (`net`, `crypto/tls`, `bufio`, `regexp`) for TCP/TLS listening and RFC 5424/3164 parsing. The syslog protocol is simple enough that a third-party library (e.g. `go-syslog`) would add dependency weight without meaningful benefit. The parser handles both RFC formats via regex, with a graceful fallback for unstructured messages.

### Server-level TLS certificates

TLS cert/key can be provided in two ways:
1. **Per-connection** — stored in the connection's `config` JSONB (`tls_cert`, `tls_key` fields)
2. **Server-level** — via `SYSLOG_TLS_CERT` / `SYSLOG_TLS_KEY` env vars, auto-injected into connections that don't specify their own

This lets a single Heimdall deployment serve multiple syslog connections on different ports while sharing one TLS certificate.

### TCP plaintext mode

For development and environments behind a reverse proxy, the listener supports `protocol: "tcp"` (plaintext) alongside `protocol: "tls"`. Production deployments should use TLS.

---

## Files created

| File | Purpose |
|------|---------|
| `backend/internal/connectors/logs/syslog.go` | Full syslog listener implementation — config parsing, TCP/TLS server, RFC 5424/3164 parsing, log_buffer insertion, graceful shutdown |
| `backend/internal/connectors/logs/syslog_test.go` | Unit tests — config validation, RFC 5424 parsing, RFC 3164 parsing, fallback parsing, severity mapping |
| `backend/internal/connectors/listener.go` | `ListenerManager` — manages long-running network listener goroutines (start/stop/stopAll), analogous to `Poller` |
| `frontend/src/components/connections/wizard/steps/StepSyslogConfig.vue` | Wizard step component — port and protocol configuration with platform-specific setup instructions |

## Files modified

| File | Change |
|------|--------|
| `backend/internal/config/config.go` | Added `SyslogTLSCert` and `SyslogTLSKey` fields loaded from `SYSLOG_TLS_CERT` / `SYSLOG_TLS_KEY` env vars |
| `backend/internal/api/handlers/server.go` | Added `Listener *connectors.ListenerManager` to Server struct and constructor |
| `backend/internal/api/router.go` | Updated `NewRouter` signature to accept and pass `ListenerManager` |
| `backend/internal/api/handlers/connections.go` | Syslog config validation on create, TLS cert injection, listener start on create, listener restart on update, listener stop on delete, syslog-specific test handler |
| `backend/cmd/heimdall/main.go` | Create `ListenerManager`, call `resumeSyslogListeners` on boot, stop listeners on shutdown |
| `frontend/src/components/connections/wizard/flows.ts` | Syslog flow: `available: true`, wired to StepName → StepSyslogConfig → StepTest |
| `frontend/src/components/connections/ConnectionForm.vue` | Fixed syslog edit fields (was describing a remote target; now describes a listener with port + protocol) |

---

## How it works

### Connection lifecycle

1. **Create:** User picks "Syslog" in the wizard → enters port + protocol → connection saved to DB → `ListenerManager.Start()` binds the port and launches the listener goroutine
2. **Listen:** Listener accepts TCP connections in a loop. Each client connection is handled in its own goroutine. Messages are parsed line-by-line via `bufio.Scanner`, classified as RFC 5424 / RFC 3164 / raw, and inserted into `log_buffer` with mapped severity.
3. **Update:** Old listener is stopped (`ListenerManager.Stop()`), new one started with updated config.
4. **Delete:** Listener stopped, connection row deleted (CASCADE removes log_buffer entries).
5. **Server restart:** `resumeSyslogListeners()` queries all active syslog connections and restarts their listeners.

### Message parsing

The parser tries formats in order:

1. **RFC 5424** — `<PRI>VERSION SP TIMESTAMP SP HOSTNAME SP APP-NAME SP PROCID SP MSGID SP MSG`
   - Extracts: facility, severity, timestamp, hostname, app name, process ID, message ID, message
   - Handles NILVALUE (`-`) → empty string

2. **RFC 3164** (BSD syslog) — `<PRI>TIMESTAMP SP HOSTNAME SP MSG`
   - Extracts: facility, severity, BSD timestamp, hostname, message
   - Timestamp pattern: `Mon DD HH:MM:SS`

3. **Fallback** — Any line that doesn't match either format is treated as an unstructured message with severity=6 (Informational).

### Severity mapping

| Syslog numeric | Syslog name | Heimdall severity |
|----------------|-------------|-------------------|
| 0–2 | Emergency, Alert, Critical | `critical` |
| 3 | Error | `error` |
| 4 | Warning | `warning` |
| 5–6 | Notice, Informational | `info` |
| 7 | Debug | `debug` |

### Payload format

Each syslog message is stored as a JSON object in `log_buffer.payload`:

```json
{
  "facility": 1,
  "severity": 5,
  "timestamp": "Apr  5 12:00:00",
  "hostname": "myhost",
  "app_name": "myapp",
  "proc_id": "1234",
  "msg_id": "ID47",
  "message": "Something happened",
  "raw": "<13>Apr  5 12:00:00 myhost myapp[1234]: Something happened"
}
```

The `raw` field preserves the original line for debugging. The `source_type` is always `"syslog"`.

---

## Configuration

### Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SYSLOG_TLS_CERT` | No | PEM-encoded TLS certificate for syslog listeners (shared across all syslog connections) |
| `SYSLOG_TLS_KEY` | No | PEM-encoded TLS private key |

### Connection config JSONB

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | int | `6514` | TCP port to listen on |
| `protocol` | string | `"tls"` | `"tcp"` or `"tls"` |
| `tls_cert` | string | — | Per-connection PEM cert (overrides server-level) |
| `tls_key` | string | — | Per-connection PEM key |

---

## Test coverage

| Test | What it validates |
|------|-------------------|
| `TestNewSyslog_Defaults` | Default port (6514) and protocol (tls) |
| `TestNewSyslog_InvalidPort` | Port range validation (1–65535) |
| `TestNewSyslog_InvalidProtocol` | Only tcp/tls accepted |
| `TestNewSyslog_ValidTCP` | TCP plaintext config |
| `TestParseSyslogMessage_RFC5424` | Full RFC 5424 field extraction |
| `TestParseSyslogMessage_RFC5424_NilValues` | NILVALUE `-` handling |
| `TestParseSyslogMessage_RFC3164` | BSD syslog timestamp and hostname extraction |
| `TestParseSyslogMessage_Fallback` | Unstructured messages default to severity 6 |
| `TestMapSyslogSeverity` | All 8 syslog severity levels → Heimdall mapping |

---

## Phase 1 status

| Item | Status |
|------|--------|
| Supabase API poller | Done (v0.18.0) |
| Syslog TLS listener | **Done** (this implementation) |
| **Phase 1** | **Complete** |

### Platforms now unlocked

| Platform | Ingestion path | Cost to user |
|----------|---------------|-------------|
| Render | Syslog Log Streams → Heimdall syslog listener | Free (all plans) |
| Heroku | Syslog drain → Heimdall syslog listener | Free (paid plans) |
| DigitalOcean | rsyslog forwarding → Heimdall syslog listener | Free (all plans) |
| Any Linux server | rsyslog / syslog-ng → Heimdall syslog listener | Free |
| Supabase | API poller (v0.18.0) | Free (all plans) |

---

## What's next (Phase 2)

Per `docs/executing/ingestion-roadmap.md`:

1. **OTLP HTTP receiver** — `POST /v1/logs` accepting OpenTelemetry log payloads (unlocks Neon, Heroku Fir, OTel-instrumented apps)
2. **JS/TS SDK** — `@heimdall/sdk` npm package for direct instrumentation (unlocks serverless, any Node.js app)
