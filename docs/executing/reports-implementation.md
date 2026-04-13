# Reports — Implementation Plan

**Status:** Not started
**Owner:** TBD
**Prereqs:** None — all scaffolding (frontend skeleton, API stubs, types) already exists

---

## Why this exists

Heimdall's monitoring pipeline currently terminates at the Activity feed. When
the agent assesses flagged logs and outputs `Action: investigate`, that action
directive is never parsed — the assessment text is stored in `agent_log` and
a notification fires, but nothing further happens. Reports are listed as a
core platform section in `docs/vision.md`, the frontend has a full UI skeleton
(`ReportsPage`, `ReportList`, `ReportCard`, `ReportDetail`, store, types, API
client), and the backend has stub endpoints — but no report is ever *created*.

This plan works backwards from what a finished report should look like, then
derives the generation pipeline from that structure.

---

## Part 1 — The finished report

### Design priorities

1. **Accuracy** — every claim in the report must trace back to evidence (a log
   entry, a query result, a code snippet). No hallucinated context.
2. **Transparency** — if something is uncertain or the investigation hit a dead
   end, say so explicitly. Never paper over gaps.
3. **Ease of reading** — a human should be able to skim the report in 30
   seconds and understand severity + what happened. Detail is available but
   not forced.
4. **Portability** — the report must be easy to copy/download as Markdown and
   hand to another agent or developer. This means structured text, not opaque
   JSON blobs.

### Report sections

A report is a self-contained document. When rendered in the UI it uses
collapsible sections; when exported it's a flat Markdown file. The sections
below are ordered by reading priority — the most important information first.

```
┌─────────────────────────────────────────────────────────────┐
│  REPORT HEADER                                              │
│  Title (1 line), severity badge, status, timestamps         │
├─────────────────────────────────────────────────────────────┤
│  1. SUMMARY                                                 │
│     2-3 sentence plain-language overview.                    │
│     What happened, what's affected, how severe.             │
│     A non-technical stakeholder should understand this.      │
├─────────────────────────────────────────────────────────────┤
│  2. TRIGGER                                                 │
│     What initiated this report:                             │
│     - Source: monitoring | scheduled | manual                │
│     - The original assessment that triggered investigation   │
│     - Link back to the Activity feed entry (agent_log ID)   │
├─────────────────────────────────────────────────────────────┤
│  3. TIMELINE                                                │
│     Chronological sequence of relevant events.              │
│     Each entry: timestamp + what happened.                  │
│     Built from log evidence, not fabricated.                │
├─────────────────────────────────────────────────────────────┤
│  4. INVESTIGATION                                           │
│     What the agent did to dig deeper:                       │
│     - Which tools it called and why                         │
│     - Key findings from each tool call                      │
│     - Dead ends / tools that returned nothing useful        │
│     This section is the transparency layer.                 │
├─────────────────────────────────────────────────────────────┤
│  5. ROOT CAUSE ANALYSIS                                     │
│     The agent's diagnosis:                                  │
│     - What went wrong and why                               │
│     - Confidence level (confirmed / likely / uncertain)     │
│     - Supporting evidence (references to timeline entries   │
│       or tool results)                                      │
│     If root cause is unknown, say so explicitly.            │
├─────────────────────────────────────────────────────────────┤
│  6. IMPACT                                                  │
│     What was affected:                                      │
│     - Services, endpoints, users, data                      │
│     - Duration of impact (if determinable)                  │
│     - Blast radius (isolated vs. cascading)                 │
├─────────────────────────────────────────────────────────────┤
│  7. RECOMMENDATIONS                                         │
│     Suggested next steps for the human:                     │
│     - Immediate actions (if any)                            │
│     - Longer-term fixes                                     │
│     Clearly labelled as suggestions, not directives.        │
├─────────────────────────────────────────────────────────────┤
│  8. TOOL TRACE  [collapsed by default]                      │
│     Raw tool call / result pairs from the investigation.    │
│     Full fidelity — the reader can verify any claim above   │
│     against the actual data the agent saw.                  │
└─────────────────────────────────────────────────────────────┘
```

### Markdown export format

When downloaded or copied, the report renders as:

```markdown
# [Title]

**Severity:** error | **Status:** open | **App:** my-backend-service
**Opened:** 2026-04-13 14:32 UTC | **Trigger:** monitoring

---

## Summary

[2-3 sentences]

## Trigger

[Original assessment text, agent_log entry reference]

## Timeline

| Time (UTC) | Event |
|------------|-------|
| 14:28:03   | First "unable to acquire lock" error logged |
| 14:28:15   | 3 more lock errors from different request handlers |
| 14:30:01   | Monitoring cycle flagged the pattern |

## Investigation

- **search_logs** for "unable to acquire lock" (last 1h) → 47 matches,
  all originating from the `orders` service
- **query_database** for active locks → found long-running transaction
  (PID 4821, running for 12m, started by `update_value` job)
- **search_codebase** for `update_value` → `src/jobs/update_value.py`
  opens transaction at line 34, no matching commit

## Root Cause

**Confidence:** likely

A background job (`src/jobs/update_value.py`) opens a database transaction
but does not commit or roll back on the error path. When the job encounters
an exception, the transaction remains open indefinitely, holding row-level
locks that block subsequent queries.

## Impact

- **Affected:** all database queries touching rows locked by the orphaned
  transaction
- **Duration:** ongoing since the job last failed (~12 minutes at time of
  detection)
- **Blast radius:** any service issuing writes to the `orders` table

## Recommendations

- **Immediate:** terminate the orphaned transaction (PID 4821) to release
  locks
- **Fix:** add `try/finally` with explicit rollback in
  `src/jobs/update_value.py` to prevent transaction leaks

---

<details>
<summary>Tool trace (3 calls)</summary>

### 1. search_logs
**Input:** `{ "query": "unable to acquire lock", "hours": 1 }`
**Result:** 47 entries, newest first [truncated to first 10]
...

### 2. query_database
**Input:** `{ "query": "SELECT * FROM pg_locks WHERE granted = false" }`
**Result:** [full result]
...

### 3. search_codebase
**Input:** `{ "query": "update_value", "path": "src/jobs/" }`
**Result:** [matched files and snippets]
...

</details>
```

### Data model (revised)

The existing `Report` type in `frontend/src/types/report.ts` is close but
needs adjustment to support the sections above. Proposed schema:

```
reports
├── id              UUID PK
├── app_id          UUID FK → applications(id)     -- scoped to application
├── user_id         UUID FK → users(id)            -- org owner
├── agent_log_id    UUID FK → agent_log(id) NULL   -- the triggering assessment
├── conversation_id UUID FK → conversations(id) NULL -- if investigation used interactive loop
│
├── title           TEXT NOT NULL                   -- 1-line headline
├── summary         TEXT NOT NULL                   -- 2-3 sentence overview
├── severity        TEXT NOT NULL                   -- info | warning | error | critical
├── status          TEXT NOT NULL DEFAULT 'open'    -- open | investigating | resolved | dismissed
├── confidence      TEXT NOT NULL DEFAULT 'uncertain' -- confirmed | likely | uncertain
│
├── trigger_type    TEXT NOT NULL                   -- monitoring | scheduled | manual
├── trigger_detail  JSONB                          -- original assessment, schedule metadata
├── timeline        JSONB NOT NULL DEFAULT '[]'    -- [{timestamp, event}]
├── investigation   JSONB NOT NULL DEFAULT '[]'    -- [{tool, input_summary, finding, raw_result}]
├── root_cause      TEXT                           -- markdown prose, nullable if unknown
├── impact          TEXT                           -- markdown prose
├── recommendations TEXT                           -- markdown prose
├── tool_trace      JSONB                          -- [{tool, input, result}] full fidelity
│
├── resolution      TEXT                           -- how it was resolved (filled on close)
├── started_at      TIMESTAMPTZ NOT NULL DEFAULT now()
├── resolved_at     TIMESTAMPTZ                    -- set when status → resolved
├── created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
└── updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
```

Key differences from the existing `Report` type:
- **`app_id`** added — reports are per-application, like everything else
- **`agent_log_id`** added — links back to the triggering Activity entry
- **`conversation_id`** added — links to the investigation conversation
- **`title`** added — distinct from summary; one-line headline
- **`confidence`** added — explicit uncertainty signal
- **Structured sections** (`timeline`, `investigation`, `root_cause`, `impact`,
  `recommendations`) replace the opaque `findings` blob
- **`trigger_detail`** replaces `trigger_source` — structured JSON, not a string
- **`error` severity** added — the existing type skips it, but the agent uses it

---

## Part 2 — Report generation pipeline

Working backwards from the report structure, here's what the generation
pipeline needs to produce each section.

### Trigger: `Action: investigate` in monitor output

The monitor loop (`monitor.go:168`) currently discards the `Action:` field.
The first change is to **parse it** and, when `action == "investigate"`,
spawn an investigation.

```
monitor loop
  ├── classify logs (Lumber ONNX)
  ├── escalate flagged → Claude assessment
  ├── parse severity   ← already done
  ├── parse action     ← NEW
  ├── emit to agent_log + notify   ← already done
  │
  └── if action == "investigate":
        ├── create report row (status: investigating)
        ├── spawn investigation goroutine
        │     ├── run interactive-style tool-use loop
        │     │   (search_logs, query_database, search_codebase)
        │     ├── accumulate tool_trace
        │     └── final Claude call: synthesise report sections
        ├── update report row with findings
        └── emit agent_log entry (entry_type: "investigation")
```

### Investigation prompt

The investigation is a second, separate Claude call (or tool-use loop) with
a dedicated system prompt. It receives:

- The original assessment as context
- The flagged logs that triggered it
- Access to the same tools (search_logs, query_database, search_codebase)

Its output format is structured to map directly onto report sections:

```
TITLE: <1-line headline>
CONFIDENCE: <confirmed | likely | uncertain>

TIMELINE:
- <ISO timestamp> | <event description>
- ...

ROOT CAUSE:
<1-3 paragraphs, markdown>

IMPACT:
<1-2 paragraphs, markdown>

RECOMMENDATIONS:
- <immediate action>
- <longer-term fix>
```

The `investigation` and `tool_trace` sections are populated automatically
from the tool-use loop's call/result pairs — the agent doesn't write these
itself.

### Report lifecycle

```
         ┌──────────┐
         │   open   │  ← created, investigation hasn't started yet
         └────┬─────┘
              │ investigation goroutine starts
              v
      ┌───────────────┐
      │ investigating  │  ← agent is running tool-use loop
      └───────┬───────┘
              │ investigation completes
              v
         ┌──────────┐
         │   open   │  ← findings written, awaiting human review
         └────┬─────┘
              │ human action
              v
    ┌──────────┐  or  ┌───────────┐
    │ resolved │      │ dismissed │
    └──────────┘      └───────────┘
```

Humans resolve or dismiss reports via the UI. The agent never closes its own
reports — it diagnoses, humans decide.

---

## Part 3 — Implementation phases

### Phase 1 — Database + backend CRUD

| Step | File | What |
|------|------|------|
| 1.1 | `backend/migrations/025_create_reports.up.sql` | Create `reports` table per schema above |
| 1.2 | `backend/migrations/025_create_reports.down.sql` | Drop table |
| 1.3 | `backend/internal/db/queries/reports.sql` | sqlc queries: Insert, GetByID, ListByApp, UpdateStatus, UpdateFindings |
| 1.4 | Run `make sqlc-generate` | Generate Go types |
| 1.5 | `backend/internal/api/handlers/reports.go` | Replace stubs with real handlers: List (filtered by app_id, status), Get, UpdateStatus |
| 1.6 | `backend/internal/api/router.go` | Add PATCH `/reports/{id}/status` route |

No report *creation* endpoint — reports are created by the agent pipeline,
not by humans. Humans can only change status (resolve / dismiss) and
optionally add a resolution note.

### Phase 2 — Action parsing + investigation trigger

| Step | File | What |
|------|------|------|
| 2.1 | `backend/internal/agent/loop.go` | Add `parseActionFromResponse()` — regex for `Action: <value>`, same pattern as severity parser |
| 2.2 | `backend/internal/agent/monitor.go` | Call action parser after assessment. When `investigate`: insert report row, spawn `a.RunInvestigation()` goroutine |
| 2.3 | `backend/internal/agent/loop.go` | New `RunInvestigation()` method — similar to `RunMonitoring()` but with investigation system prompt, tool-trace accumulation, and report-update on completion |
| 2.4 | `backend/internal/agent/prompt.go` | Add `investigationSystemPrompt` — instructs Claude to investigate and produce structured output (title, confidence, timeline, root cause, impact, recommendations) |

### Phase 3 — Investigation loop details

`RunInvestigation()` is structurally similar to `RunMonitoring()` but:

- Uses the investigation prompt, not the monitoring prompt
- Receives the original assessment + flagged logs as initial context
- Records every tool call/result pair in a `[]ToolTraceEntry` slice
- After the tool-use loop completes (end_turn or max iterations), makes a
  final synthesis call asking Claude to produce the structured report sections
- Parses the synthesis output into `title`, `confidence`, `timeline`,
  `root_cause`, `impact`, `recommendations`
- Updates the report row with all findings and sets status back to `open`
- Emits an `agent_log` entry with `entry_type: "investigation"`

### Phase 4 — Frontend wiring

| Step | File | What |
|------|------|------|
| 4.1 | `frontend/src/types/report.ts` | Update `Report` interface to match new schema |
| 4.2 | `frontend/src/api/reports.ts` | Add `updateReportStatus(id, status, resolution?)` |
| 4.3 | `frontend/src/stores/reports.ts` | Add `fetchReportsByApp(appId)`, status update action |
| 4.4 | `frontend/src/pages/ReportsPage.vue` | Filter by current app, wire to app-switch |
| 4.5 | `frontend/src/components/reports/ReportDetail.vue` | Full report view with all sections, collapsible tool trace |
| 4.6 | `frontend/src/components/reports/ReportCard.vue` | Show title, severity, status, confidence, timestamp |
| 4.7 | New: report Markdown export | "Copy as Markdown" / "Download .md" button on ReportDetail |

### Phase 5 — Manual investigation trigger

| Step | File | What |
|------|------|------|
| 5.1 | `backend/internal/api/handlers/reports.go` | `POST /api/apps/{appId}/reports` — create a report with `trigger_type: manual` and a user-provided prompt |
| 5.2 | Frontend | "Investigate" button on Activity entries and/or a "New Investigation" action on the Reports page |

This allows humans to trigger an investigation on demand, not just wait for
the agent to decide `Action: investigate`.

---

## Part 4 — Markdown export spec

The "Copy as Markdown" feature is central to the portability goal. The export
function takes a `Report` object and produces the Markdown format shown in
Part 1. Key details:

- **Tool trace** wrapped in `<details>` for collapsibility in GitHub/rendered
  Markdown
- **No Heimdall-specific markup** — output is vanilla Markdown that renders
  anywhere
- **Timestamps in UTC** — unambiguous across timezones
- **Copyable via `navigator.clipboard.writeText()`** with a toast confirmation
- **Downloadable** as `heimdall-report-{id-short}-{date}.md`

This is a pure frontend concern — a `formatReportAsMarkdown(report: Report): string`
utility function, called by both the copy and download buttons.

---

## Scoping notes

- **Phase 1-2** are the minimum viable feature: assessments that say
  "investigate" actually trigger an investigation, and the results land in the
  Reports page.
- **Phase 3** is the quality layer: structured sections, tool trace, confidence
  levels.
- **Phase 4** is the UX layer: proper rendering, export.
- **Phase 5** is a convenience feature: manual triggers.

Phases 1-3 can ship together as a single release. Phase 4 and 5 are
independent follow-ups.

---

## Open questions

1. **Investigation concurrency** — the monitor loop already has a semaphore
   (10 concurrent apps). Should investigations share that semaphore, have
   their own, or run unbounded? Investigations are heavier (multi-turn
   tool-use loop) so they'll consume more API quota. Proposal: separate
   semaphore with a limit of 3 concurrent investigations.

2. **Investigation budget** — `RunMonitoring` uses the same `maxIterations`
   (10) as interactive chat. Investigations may need more depth. Should we
   allow a higher iteration cap (e.g. 15-20) for investigations, or keep it
   at 10 and let the synthesis call handle the rest?

3. **Duplicate investigations** — if the same issue triggers `Action:
   investigate` on consecutive monitor cycles, we'd spawn duplicate reports.
   Options: (a) deduplicate by checking for open reports with the same
   `app_id` + similar assessment within a time window, (b) let duplicates
   exist and let humans merge/dismiss, (c) skip investigation if an open
   report already exists for the app. Leaning toward (c) as simplest.

4. **Report retention** — should reports auto-archive or auto-delete after
   some period? Or are they permanent records? The Activity feed has log
   retention (`0.30.3`), but reports feel more like audit artifacts that
   should persist.

5. **Severity escalation** — the initial assessment assigns a severity. The
   investigation may reveal the situation is worse (or better) than initially
   thought. Should the report's severity be independently set by the
   investigation, or anchored to the original assessment? Proposal: the
   investigation sets its own severity with a note if it differs from the
   trigger.

6. **Notification on report completion** — should a notification fire when an
   investigation finishes and the report is written? Likely yes, using the
   same notifier pipeline, but it should be clearly distinguished from the
   initial assessment notification (to avoid "double-alerting" fatigue).

7. **`trigger_type: scheduled`** — scheduled investigations
   (`entry_type: "scheduled_investigation"`) currently write to `agent_log`
   the same way monitoring does. Should they also be able to produce reports
   via the same `Action: investigate` mechanism, or are scheduled
   investigations always self-contained? If yes, the same action-parsing
   logic applies to the scheduled investigation code path.
