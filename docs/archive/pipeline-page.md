# Pipeline Page

Visualise the full log pipeline in real time — from ingestion through classification to agent assessment — as an interactive, Heimdall-branded flow diagram.

---

## The Pipeline (What We're Visualising)

The pipeline is a fixed, five-stage topology. Every log that enters Heimdall passes through these stages:

```
┌─────────────┐    ┌──────────────┐    ┌────────────────┐    ┌───────────────┐    ┌──────────────┐
│  INGESTION   │───▶│   LUMBER     │───▶│  SEVERITY GATE  │───▶│    AGENT      │───▶│  ACTIVITY    │
│  Sources     │    │  Classifier  │    │  Escalation     │    │  Assessment   │    │  Feed        │
└─────────────┘    └──────────────┘    └────────────────┘    └───────────────┘    └──────────────┘
  Webhooks          Type + Category     Flagged vs Safe       Claude tool-use      agent_log +
  Pollers           Confidence score    Hardcoded rules       loop (max 10 iter)   notifications
  OTLP              Summary text        per Type.Category     Assessment text
```

**Key I/O at each stage:**

| Stage | Input | Output | Interesting metrics |
|-------|-------|--------|---------------------|
| **Ingestion** | Raw HTTP payload (Vercel, Fly, AWS, GCP, native) | Normalised `log_buffer` row (source_type, severity, payload) | Logs/sec per source, format breakdown |
| **Lumber** | Extracted text (max 1000 chars) | `Type.Category` + confidence (0–1) + summary | Classification distribution, avg confidence |
| **Severity Gate** | Classified log | Boolean: escalate or safe | Flagged %, escalation rules hit |
| **Agent** | Batch of ≤50 flagged logs | Assessment text + auto-severity + tool calls | Assessment count, tool calls/assessment, latency |
| **Activity** | Assessment + metadata | `agent_log` row + optional notification | Entries/hour, severity distribution |

---

## Proposal A — Horizontal Flow Diagram (Recommended)

**Concept:** A left-to-right (or top-to-bottom on mobile) flow diagram with five node cards connected by animated particle streams. Each node shows real-time stats. Logs appear as small glowing particles that flow between nodes, splitting/filtering at the Gate stage.

### Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  PIPELINE                                                    [app] │
│                                                                     │
│  ┌─────────┐   ━━━▶   ┌─────────┐   ━━━▶   ┌──────────┐          │
│  │ INGEST  │   ~~~~   │ LUMBER  │   ~~~~   │   GATE   │          │
│  │         │          │         │          │          │          │
│  │ 142/min │          │ 98.2%   │          │ 12% esc  │          │
│  │ 3 src   │          │ confid  │          │ 88% safe │          │
│  └─────────┘          └─────────┘          └──────────┘          │
│                                                │    │              │
│                                           flagged  safe            │
│                                                │    │              │
│                                                ▼    ▼              │
│                                          ┌─────────┐ ┌────────┐   │
│                                          │  AGENT  │ │discard │   │
│                                          │         │ │(dimmed)│   │
│                                          │ 2 assess│ └────────┘   │
│                                          │ /hour   │              │
│                                          └────┬────┘              │
│                                               │                    │
│                                               ▼                    │
│                                          ┌─────────┐              │
│                                          │ACTIVITY │              │
│                                          │         │              │
│                                          │ 48 today│              │
│                                          └─────────┘              │
│                                                                     │
│  ─── LIVE LOG STREAM ────────────────────────────────────────────  │
│  12:04:32  fly/app  ERROR  connection_failure  → FLAGGED → Agent   │
│  12:04:31  vercel   INFO   request.success     → SAFE              │
│  12:04:30  fly/app  WARN   slow_request        → FLAGGED → Agent   │
└─────────────────────────────────────────────────────────────────────┘
```

### Visual Design

- **Node cards**: Dark glass-morphism panels (`bg-neutral-900/80 backdrop-blur`) with a subtle border glow matching the Heimdall green (`#5a9e6a`) when actively processing. Same card aesthetic as ConnectionBubble.
- **Particle streams**: Small glowing dots (2–4px) that travel along cubic bezier paths between nodes. Reuse the FlowLines.vue SVG overlay pattern — same `getBoundingClientRect()` approach, same dash animation system, but with actual particle sprites instead of dashes.
- **Gate split**: The stream visually forks — flagged logs glow amber/red and flow to Agent, safe logs dim to grey and fade out (or flow to a muted "processed" sink). This is the most visually impactful moment.
- **Dormant state**: When no logs are flowing, nodes show historical stats and the particles slow to a gentle drift (like AgentNebula's dormant mode).

### Node Detail (Click-to-Expand)

Clicking a node opens an inline detail panel (slide-down, not modal) showing:

| Node | Detail panel contents |
|------|-----------------------|
| **Ingestion** | Per-source breakdown (Fly: 80/min, Vercel: 62/min), format distribution pie, last 5 raw payloads |
| **Lumber** | Classification distribution bar chart (ERROR 12%, REQUEST 45%, SYSTEM 8%…), confidence histogram, last 5 classified logs with before/after |
| **Gate** | Escalation rules table with hit counts, flagged % over time sparkline |
| **Agent** | Recent assessments list (summary + severity badge), avg tool calls per assessment, avg latency |
| **Activity** | Links to Activity page (filtered), severity distribution over time |

### Implementation

**Backend — New SSE Endpoint:**

```
GET /api/apps/{appId}/pipeline/stream   (SSE, JWT auth)
```

A Server-Sent Events stream that emits pipeline events in real time. The monitor loop already has all the data — we instrument it to publish events to a per-app channel:

```json
{"stage": "ingestion", "ts": "...", "source_type": "flyio/myapp", "severity": "error", "log_id": "..."}
{"stage": "classified", "ts": "...", "log_id": "...", "type": "ERROR", "category": "connection_failure", "confidence": 0.92}
{"stage": "gate", "ts": "...", "log_id": "...", "escalated": true}
{"stage": "assessment", "ts": "...", "app_id": "...", "flagged_count": 3, "severity": "high", "summary": "..."}
```

Implementation approach:
1. Add a `PipelineBus` (in-memory pub/sub per app, similar to how AgentEvent channels work in `loop_stream.go`). A simple `sync.Map[appID][]chan PipelineEvent`.
2. Instrument `IngestWebhookLogs` to publish `ingestion` events after DB write.
3. Instrument `monitorApp` to publish `classified`, `gate`, and `assessment` events during the existing processing loop.
4. New handler `StreamPipeline` reads from the channel and writes SSE frames. Auto-closes on client disconnect.
5. Periodic `stats` event every 5s with aggregate counts (total ingested, classified, flagged, assessed) for initial page load and reconnection.

**Backend — Stats Snapshot Endpoint:**

```
GET /api/apps/{appId}/pipeline/stats
```

Returns point-in-time aggregate stats for all five stages (counts, rates, distributions). Used for initial page render before the SSE stream fills in. This is a simple SQL query against `log_buffer` and `agent_log` with time-windowed counts.

**Frontend:**

- New page: `PipelinePage.vue` at route `/pipeline`
- Components:
  - `PipelineCanvas.vue` — orchestrates layout, positions nodes, manages SSE connection
  - `PipelineNode.vue` — individual stage card with stats, status indicator, click-to-expand
  - `PipelineStream.vue` — SVG overlay with animated particles (extends FlowLines.vue pattern)
  - `PipelineLogTicker.vue` — bottom live-log stream showing individual log journeys
  - `PipelineNodeDetail.vue` — expandable detail panel per node
- SSE connection via `EventSource` wrapped in a composable (`usePipelineStream`)
- Particle animation via `requestAnimationFrame` on an SVG or Canvas overlay

### Complexity: Medium-High
- Backend: ~300 LOC (bus + SSE handler + instrumentation points)
- Frontend: ~800 LOC (5 components + composable + route)
- Biggest risk: particle animation performance with high log volume → mitigate by sampling (show 1 in N particles when rate > threshold)

---

## Proposal B — Vertical Sankey / Funnel View

**Concept:** A vertical funnel that narrows at each stage, showing proportional flow. Think Mixpanel funnel or a Sankey diagram rotated 90°. The width of each stream segment represents volume.

### Layout

```
┌──────────────────────────────────────────────┐
│              ╔═══════════════╗                │
│              ║   INGESTION   ║                │
│              ║   142 logs/m  ║                │
│              ╚═══════╤═══════╝                │
│              ════════╪════════  (full width)  │
│              ╔═══════╧═══════╗                │
│              ║    LUMBER     ║                │
│              ║  7 categories ║                │
│              ╚═══════╤═══════╝                │
│           ┌──────────┴──────────┐             │
│     ╔═════╧═════╗         ╔════╧════╗        │
│     ║  FLAGGED  ║         ║  SAFE   ║        │
│     ║   12%     ║  (wide) ║   88%   ║        │
│     ╚═════╤═════╝         ╚═════════╝        │
│      (narrow)                                 │
│     ╔═════╧═════╗                             │
│     ║   AGENT   ║                             │
│     ║  2/hour   ║                             │
│     ╚═════╤═════╝                             │
│     ╔═════╧═════╗                             │
│     ║ ACTIVITY  ║                             │
│     ╚═══════════╝                             │
└──────────────────────────────────────────────┘
```

### Visual Design

- **Funnel streams**: Gradient-filled SVG paths whose width is proportional to log volume. Full width at ingestion, narrows dramatically at the Gate split. The visual narrowing *is the story* — "Lumber + Gate filtered out 88% of noise so the Agent only sees what matters."
- **Colour coding**: Green for safe/processed, amber for flagged, red pulse for critical assessments.
- **Node cards**: Minimal — just label + key stat, overlaid on the stream. Hover for detail tooltip, click for full detail panel.
- **Stream animation**: Slow gradient scroll within the filled paths (CSS `background-position` animation on a striped gradient).

### Implementation

Same backend as Proposal A (SSE + stats endpoints). Frontend differs:

- `PipelineFunnel.vue` — SVG-based Sankey layout with proportional widths calculated from stats
- Requires a lightweight Sankey layout algorithm (or hand-rolled since topology is fixed)
- Responsive: collapses to a simple numbered list on narrow viewports

### Complexity: Medium
- Backend: Same as Proposal A
- Frontend: ~600 LOC (simpler than particle animation, but Sankey path math is fiddly)
- Strength: The funnel shape immediately communicates Heimdall's value — "we filter the noise"
- Weakness: Less interactive feel, fewer "wow" moments than animated particles

---

## Proposal C — Hybrid: Flow Diagram + Sankey Widths

**Concept:** Combine A and B. Horizontal flow layout (Proposal A's node cards and particle streams) but the connecting streams have Sankey-style proportional widths. The stream from Ingestion → Lumber is fat; the stream from Gate → Agent is thin. Particles still flow along the paths but within width-proportional channels.

This gives you the interactive, n8n-style node layout *and* the at-a-glance funnel narrative. More complex to build but the most visually distinctive.

### Complexity: High
- Frontend: ~1200 LOC
- Risk: visual clutter if not carefully tuned

---

## Recommendation

**Start with Proposal A** (Horizontal Flow Diagram). Reasons:

1. **Closest to the n8n aesthetic** you described — discrete nodes with animated connections.
2. **Reuses existing patterns** — FlowLines.vue (SVG beziers + dash animation) and AgentNebula (canvas animation, dormant state) give us a head start.
3. **Most extensible** — easy to add new nodes later (e.g., a "Notifications" node after Activity, or a "Memory" node for future long-term learning).
4. **Gate split is the hero moment** — particles forking into flagged/safe streams is visually striking and tells Heimdall's story.
5. **The live log ticker at the bottom** gives power users the raw detail they want while the flow diagram gives everyone the big picture.

We can evolve toward Proposal C (adding proportional stream widths) in a later iteration once the base flow is solid.

---

## Questions

1. **Scope: app-scoped or org-wide?** The Connections page is app-scoped. Should Pipeline also be scoped to the current app, or show an aggregated cross-app view? (Recommendation: app-scoped to match existing patterns, with a future "org overview" as a stretch goal.)

2. **Historical vs live-only?** Should the page show only real-time flow, or also a time-range selector to replay historical pipeline stats (e.g., "show me yesterday's flow")? Historical adds backend complexity (time-windowed aggregation queries) but is very useful for debugging.

3. **Mobile / narrow viewport?** The horizontal flow won't fit on mobile. Options: (a) collapse to a vertical list with mini-stats, (b) horizontal scroll, (c) declare it desktop-only. What's your preference?

4. **Particle density control?** At high log volume (100+ logs/sec), rendering every log as a particle will tank performance. Plan is to sample (show 1 in N) and display the true count on the node. Acceptable, or do you want every log represented?

5. **Navigation placement?** New top-level nav item between Connections and Agent Config? Or a sub-view within Connections? (Recommendation: top-level — it's a distinct concept from connection management.)

6. **Node detail depth?** How deep should click-to-expand go? The proposal includes per-source breakdowns, classification distributions, and recent log samples. Is that enough, or do you want full log inspection inline (basically embedding parts of the Activity page)?

7. **Name preference?** "Pipeline", "Log Pipeline", or something else? The route would be `/pipeline`.