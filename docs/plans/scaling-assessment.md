# Scaling Assessment — Monitoring Pipeline

## Context

Heimdall's monitoring loop is a single-process goroutine that polls all active applications every 15 seconds, classifies logs via the Lumber ONNX model, and escalates flagged entries to Claude for assessment. Concurrency is capped at 10 simultaneous apps via a buffered-channel semaphore.

This document assesses whether the current architecture can support **hundreds of paying customers**, each with **2–3 apps** and **6–7 connections per app** (~500+ active applications).

---

## Current Architecture

```
┌─────────────────────────────────────────────┐
│  Single Go Process                          │
│                                             │
│  Monitor() goroutine                        │
│    ↓ every 15s                              │
│  ListActiveApplications (ALL apps, global)  │
│    ↓                                        │
│  monitorTick()                              │
│    ├─ goroutine per app (sem capped at 10)  │
│    │   ├─ fetch logs since cursor            │
│    │   ├─ classify via Lumber ONNX           │
│    │   ├─ if flagged → Claude API call       │
│    │   └─ emit heartbeat + advance cursor    │
│    └─ wg.Wait() — blocks until all complete  │
└─────────────────────────────────────────────┘
```

Key files:
- `backend/internal/agent/monitor.go` — tick loop, semaphore, per-app processing
- `backend/internal/agent/agent.go` — single Agent struct, Start/Stop lifecycle
- `backend/internal/db/queries/monitoring.sql` — `ListActiveApplications` (tenant-blind)

---

## Identified Bottlenecks

### 1. Single-Process, Single-Node Monitoring

**Severity: Critical**

There is one `Agent` struct, one `Monitor()` goroutine, one server. All 500 apps are processed by one instance. There is no mechanism to distribute work across multiple backend instances — running two instances would double-process every app (duplicate Claude calls, duplicate agent_log entries).

### 2. Semaphore Throughput Ceiling

**Severity: Critical**

With `maxConcurrentApps = 10` and Claude API calls averaging 2–10 seconds per escalation:

| Scenario | Apps | Flagged % | Time per app | Tick duration |
|----------|------|-----------|-------------|---------------|
| Current  | 20   | 20%       | ~3s avg     | ~6s           |
| Target   | 500  | 20%       | ~3s avg     | ~150s         |

At 500 apps, a single tick takes **~150 seconds** but fires every 15 seconds. The `wg.Wait()` blocks the next tick, creating a growing backlog. Apps would be monitored minutes late.

### 3. Tenant-Blind Query

**Severity: High**

`ListActiveApplications` returns every active app across all orgs in one query. At scale this means:
- Large result sets loaded into memory every tick
- No ability to prioritize or partition by tenant
- No isolation — one org's slow processing delays all others

### 4. Claude API Rate Limits and Cost

**Severity: High**

If 20% of 500 apps have flagged logs per cycle, that's 100 Claude API calls per tick (~400/minute at steady state). This will:
- Hit Anthropic rate limits
- Incur significant cost at scale
- Create latency spikes when rate-limited

### 5. Database Query Volume

**Severity: Medium**

Each app in a tick cycle runs 3–5 queries (monitoring state, log fetch, cursor upsert, heartbeat emit, optional config fetch). At 500 apps, that's ~2,000+ queries per 15-second cycle. PostgreSQL can handle this, but:
- `ListLogsSinceForApp` joins `log_buffer` with `connections` — performance degrades as these tables grow
- Missing indexes on `connections(app_id, status)` and `applications(status)` (deferred from Phase 8)

### 6. ONNX Classifier Thread Safety

**Severity: Medium**

The Lumber classifier is called concurrently from up to 10 goroutines. Thread safety for concurrent ONNX inference was flagged as a P2 deferred item in Phase 8. At higher concurrency this becomes a real risk for panics or corrupted results.

---

## Proposed Solutions

### Short-Term (supports ~50–100 apps)

These changes keep the single-process model but improve throughput.

| Change | Impact | Effort |
|--------|--------|--------|
| Raise `maxConcurrentApps` to 25–50 | Higher throughput per tick | Trivial |
| Add missing DB indexes (`connections(app_id, status)`, `applications(status)`) | Faster per-app queries | Small |
| Batch Claude calls — group flagged logs from multiple apps into fewer requests | Reduce API call count by 5–10x | Medium |
| Skip apps already in-flight from a previous tick (per-app dedup) | Prevent backlog compounding | Small |
| Verify ONNX classifier thread safety or add mutex | Prevent concurrent inference bugs | Small |

### Medium-Term (supports ~100–500 apps)

Introduce work distribution without a full architectural rewrite.

| Change | Impact | Effort |
|--------|--------|--------|
| **Claim-based work queue** — replace `ListActiveApplications` with `SELECT ... FOR UPDATE SKIP LOCKED` on a work table. Each backend instance claims N apps per tick. | Horizontal scaling, no double-processing | Medium |
| **Adaptive tick interval** — if a tick exceeds its interval, dynamically adjust or skip low-priority apps | Prevent backlog | Small |
| **Tiered escalation** — only call Claude for high-confidence flags (e.g., confidence > 0.8); lower-confidence flags logged but not escalated | Reduce API calls by 50–80% | Small |
| **Per-org/per-app rate limiting** — ensure one noisy app can't monopolize the pipeline | Tenant fairness | Medium |

### Long-Term (500+ apps)

Full distributed architecture.

| Change | Impact | Effort |
|--------|--------|--------|
| **Distributed task queue** (e.g., Redis/NATS-based) — monitoring jobs dispatched to a pool of worker processes | True horizontal scaling | Large |
| **Event-driven processing** — replace polling with log ingestion triggering classification (push vs pull) | Lower latency, no wasted ticks on empty apps | Large |
| **Dedicated classifier service** — run ONNX inference as a separate service, called via gRPC | Independent scaling of classification vs assessment | Large |
| **Assessment batching service** — aggregate flagged logs across apps and time windows, batch into optimized Claude calls | Cost reduction at scale | Medium |

---

## Recommendation

The current architecture comfortably supports **up to ~20–30 apps**. To reach the hundreds-of-customers scale:

1. **Immediately**: add the missing indexes and verify ONNX thread safety (Phase 9 deferred items)
2. **Next milestone**: implement claim-based work distribution (`FOR UPDATE SKIP LOCKED`) — this is the highest-impact change with moderate effort, enabling multiple backend instances without double-processing
3. **Before launch**: introduce tiered escalation and Claude call batching to control API cost
4. **Post-traction**: move to event-driven architecture if polling latency becomes unacceptable

---

## Open Questions

- What is our target scale for launch? 50 apps? 500?
- What Anthropic API tier/rate limits do we expect to operate under?
- Should monitoring latency be a guaranteed SLA (e.g., "logs assessed within 60 seconds")?
- Is the single-database model sufficient, or do we need read replicas for the monitoring queries?
