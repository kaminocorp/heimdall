# Log Ingestion Roadmap — 80% Coverage for Early Users

## Current State

Heimdall has two ingestion paths today:

| Type | Status | How it works |
|------|--------|-------------|
| `webhook_logs` | Implemented | HTTP POST with bearer token → `log_buffer` |
| `syslog` | Stubbed | RFC 5424/3164 listener (code exists, no logic) |

This document identifies the ingestion methods that would give early users ~80% coverage across the platforms and infrastructure they're most likely running.

---

## Who Are Early Users?

- Indie developers, small startups, small-to-mid engineering teams
- Running on: Vercel, Fly.io, Railway, Render, Heroku, AWS, GCP, DigitalOcean, Cloudflare
- Databases: Supabase, PlanetScale, Neon, AWS RDS, Cloud SQL, MongoDB Atlas
- Frameworks: Next.js, Express, Rails, Django, FastAPI, Go services
- May use serverless: Lambda, Edge Functions, Cloudflare Workers

---

## Ingestion Methods — Ranked by Coverage

### 1. HTTP Webhook (already implemented)

**What it is:** Platform pushes log batches to Heimdall's `POST /api/webhooks/logs` endpoint via HTTPS.

**Platforms that support this natively:**

| Platform | Mechanism | Plan/Cost |
|----------|-----------|-----------|
| Vercel | Log Drains (JSON/NDJSON) | Pro ($20/mo), $0.50/GB egress |
| Heroku | HTTPS Log Drains | All paid plans, free feature |
| Supabase | Log Drains | Pro ($25/mo) + $60/mo add-on |
| Cloudflare Workers | Logpush / Tail Workers | Workers Paid, $0.05/M requests |
| AWS Kinesis Firehose | HTTP endpoint delivery | $0.029/GB |
| GCP Pub/Sub | Push subscription to HTTP | 50GB/mo free, then $40/TB |

**What Heimdall needs:** Already done. The only gap is **per-platform payload parsers** — Vercel sends NDJSON, Heroku wraps in syslog framing, Firehose wraps in a delivery envelope. Currently, the webhook expects a specific JSON shape. Adding format detection or platform-specific parsing would unlock all of these without new connector types.

**Coverage:** ~30% of early users could use this today if their platform offers HTTP log drains.

---

### 2. Syslog TLS Listener (stub exists)

**What it is:** Heimdall listens on a TCP/TLS port for syslog-formatted messages (RFC 5424/3164).

**Platforms that support this natively:**

| Platform | Mechanism | Plan/Cost |
|----------|-----------|-----------|
| Render | Log Streams | All plans including free |
| Heroku | Syslog Drains | All paid plans, free feature |
| DigitalOcean | Log Forwarding (rsyslog) | All plans, free |
| Any Linux server | rsyslog / syslog-ng | Built-in to OS |

**What Heimdall needs:** Implement the syslog listener in `connectors/logs/syslog.go`. Go has a solid library (`gopkg.in/mcuadros/go-syslog.v2`). Needs TLS cert management and an RFC 5424 parser. The listener would parse messages and insert into `log_buffer` the same way the webhook handler does.

**Why this matters:** Render is one of the most popular platforms for early-stage startups, and syslog is their **only** log export mechanism — available on all plans including free. This is the cheapest path for Render users (zero cost) and also covers Heroku users who prefer syslog over HTTPS drains.

**Coverage:** +15% (primarily Render, Heroku, DigitalOcean, self-hosted Linux)

---

### 3. OTLP HTTP Receiver

**What it is:** Heimdall exposes an OpenTelemetry Protocol endpoint at `POST /v1/logs` that accepts OTLP-formatted log data over HTTP (JSON or protobuf).

**What sends OTLP:**

| Source | Notes |
|--------|-------|
| Heroku Fir | New generation — telemetry drains emit OTLP natively |
| Neon | Native OTLP export for Postgres logs and metrics |
| Supabase | OTLP as a log drain destination option |
| OpenTelemetry SDKs | All major languages (JS, Python, Go, Java, Rust) |
| Fluent Bit | OTel output plugin |
| Vector | OTel sink |
| OTel Collector | Standard exporter |

**What Heimdall needs:** A single HTTP endpoint that accepts OTLP log payloads. Go has official OpenTelemetry libraries for parsing. Start with HTTP/JSON (simpler), add protobuf and gRPC later.

**Why this matters:** OTLP is becoming the industry default for observability data. Supporting it means any app instrumented with OpenTelemetry can send logs to Heimdall with zero custom integration. It also unlocks Neon (which only exports via OTLP) and Heroku's new Fir generation. One endpoint, many sources.

**Coverage:** +15% (OTel-instrumented apps, Neon, Heroku Fir, forwarding agents)

---

### 4. Platform API Pollers

**What it is:** Heimdall actively polls a platform's management API on an interval to pull logs. No public URL required — Heimdall initiates all requests.

**Platforms with pollable log APIs:**

| Platform | API | Plan/Cost | Rate Limits |
|----------|-----|-----------|-------------|
| Supabase | Management API (BigQuery SQL) | All plans, free | ~120 req/min, 1k rows/query |
| Fly.io | Logs API | All plans, free | Moderate |
| Vercel | REST API (deployment events) | All plans, free | Rate limited |
| Railway | GraphQL API | All plans, free | Token required |
| MongoDB Atlas | Admin API | M10+, free | 30-day retention |
| PlanetScale | Audit Log API | All plans, free | Service token |
| AWS CloudWatch | GetLogEvents / FilterLogEvents | All, $0.01/1k requests | 10 req/sec |
| GCP Cloud Logging | entries.list | All, 50GB/mo free | Quota limited |

**What Heimdall needs:** A generic polling framework that can be specialized per platform. Each poller needs: auth configuration, cursor management, rate limit handling, and a response parser that maps platform-specific log formats into `log_buffer` entries. See [supabase-connectivity.md](./supabase-connectivity.md) for the Supabase-specific design.

**Why this matters:** This is the most user-friendly approach. The user enters credentials in Heimdall's UI and logs start flowing — no webhook URL to configure on the platform side, no public endpoint needed. Critical for users running Heimdall locally or behind a firewall. Also the only free option for Supabase users who don't want to pay $60/mo for Log Drains.

**Coverage:** +10% (users without public URLs, cost-sensitive Supabase users, Fly.io users)

---

### 5. Application SDK / Direct Instrumentation

**What it is:** The user adds a lightweight library to their application that sends structured logs directly to Heimdall's webhook endpoint with batching and retry logic.

**Applicable to:** Every framework and language — Next.js, Express, Rails, Django, FastAPI, Go, any custom app.

**What Heimdall needs:** Thin SDK wrappers around the existing webhook endpoint. A JS/TS SDK would cover the largest segment of early users (Next.js, Express, serverless functions). The SDK would handle:
- Batching (buffer N entries or M milliseconds, whichever comes first)
- Retry with backoff on failure
- Graceful degradation (don't crash if Heimdall is unreachable)
- Structured log format with severity, source_type, and arbitrary payload

Example usage:
```typescript
import { Heimdall } from '@heimdall/sdk'

const monitor = new Heimdall({ token: 'whk_...' })

monitor.log('info', 'user.signup', { userId: '123', plan: 'pro' })
monitor.error('payment.failed', { orderId: '456', error: err.message })
```

**Why this matters:** Works on every platform, even those without log drain support. Gives developers the most control over what gets sent. Particularly important for serverless environments (Lambda, Edge Functions) where there's no filesystem and no syslog.

**Coverage:** +10% (serverless users, platforms without log drains, users who want custom instrumentation)

---

## Coverage Summary

| Tier | Method | New Work | Additional Coverage |
|------|--------|----------|-------------------|
| **Tier 1** | HTTP Webhook | Per-platform payload parsers | ~30% (already built) |
| **Tier 1** | Syslog TLS Listener | Implement stub | +15% |
| **Tier 1** | OTLP HTTP Receiver | New endpoint | +15% |
| **Tier 2** | Platform API Pollers | Supabase first, then generalize | +10% |
| **Tier 2** | JS/TS SDK | New package | +10% |
| | | **Total** | **~80%** |

---

## Platform Coverage Matrix

After implementing all Tier 1 + Tier 2 methods, here's how each platform maps:

| Platform | Primary Path | Secondary Path | Cost to User |
|----------|-------------|---------------|-------------|
| **Vercel** | Webhook (Log Drain) | API Poller | Pro ($20/mo) / Free |
| **Fly.io** | API Poller | fly-log-shipper → Webhook | Free |
| **Railway** | API Poller | Locomotive → Webhook | Free |
| **Render** | Syslog | — | Free |
| **Heroku (Cedar)** | Webhook or Syslog | — | Free (on paid plan) |
| **Heroku (Fir)** | OTLP | — | Free (on paid plan) |
| **AWS Lambda** | SDK | Firehose → Webhook | Free / low |
| **AWS RDS** | CloudWatch → Firehose → Webhook | API Poller (CloudWatch) | ~$0.03/GB |
| **GCP Cloud SQL** | Log Sink → Pub/Sub → Webhook | API Poller (Cloud Logging) | 50GB/mo free |
| **DigitalOcean** | Syslog | — | Free |
| **Cloudflare Workers** | Webhook (Logpush) | SDK | Workers Paid |
| **Supabase** | API Poller | Webhook (Log Drain, $60) | Free / $60 |
| **Neon** | OTLP | — | Plan-dependent |
| **PlanetScale** | API Poller | — | Free |
| **MongoDB Atlas** | API Poller | — | Free (M10+) |
| **Self-hosted** | Syslog / SDK | Fluent Bit → Webhook/OTLP | Free |
| **Any app** | SDK | — | Free |

---

## Implementation Order

### Phase 1: Unlock free ingestion for top platforms
1. **Syslog TLS listener** — unlocks Render (free), Heroku, DigitalOcean
2. **Supabase API poller** — unlocks free Supabase log ingestion (see [supabase-connectivity.md](./supabase-connectivity.md))

### Phase 2: Industry-standard protocol + developer SDK
3. **OTLP HTTP receiver** — unlocks Neon, Heroku Fir, OTel-instrumented apps
4. **JS/TS SDK** — unlocks serverless, any Node.js app, custom instrumentation

### Phase 3: Polish and extend
5. **Webhook payload parsers** — Vercel NDJSON, Firehose envelope, Pub/Sub envelope
6. **Additional API pollers** — Fly.io, Vercel, Railway, MongoDB Atlas
7. **Python SDK** — Django, FastAPI coverage
8. **Go SDK** — Go service coverage

---

## What NOT to Build

- **Kafka/NATS consumers** — too niche for early users, complex to maintain
- **Custom file-tailing agent** — Fluent Bit / Vector already do this; document how to forward to Heimdall instead
- **Database-specific connectors** — the existing Postgres `QueryConnector` can already run `pg_stat_statements` queries; no new connector type needed, just pre-built diagnostic query templates
- **gRPC OTLP receiver** — HTTP/JSON covers the same use cases; add gRPC only if users request it
