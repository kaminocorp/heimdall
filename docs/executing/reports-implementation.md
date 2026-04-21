# Reports — Implementation Plan

**Status:** Not started
**Owner:** TBD
**Prereqs:** Decision needed on how to reconcile with the existing (orphan) `investigations` table — see Part 0.

> **Revision note (2026-04-21):** Earlier drafts of this plan asserted that a
> frontend skeleton (`ReportsPage`, `ReportList`, `ReportCard`, `ReportDetail`,
> store, types, API client) and backend stub endpoints already existed. They do
> not — neither in the current `master` nor in any ancestor reachable from it.
> Phase 4 has been rewritten as a from-scratch build, and Part 0 has been added
> to deal with the `investigations` table (migration 003) that predates this
> plan and partially anticipates it.

---

## Why this exists

Heimdall's monitoring pipeline currently terminates at the Activity feed. When
the agent assesses flagged logs and outputs `Action: investigate`, that action
directive is never parsed — the assessment text is stored in `agent_log` and
a notification fires, but nothing further happens. Reports are listed as a
core platform section in `docs/vision.md`, but no report is ever *created* and
the platform has no user-facing surface for them.

Since the Pipeline Page landed (0.47.0–0.47.3, 2026-04-19), the agent's
assessment step is now a first-class pipeline stage (`log_pipeline_events`
with `stage = 'assessment'`). Reports are the natural next layer: a durable,
structured artifact built *from* an assessment, enriched by an investigation
loop, and handed to a human for resolution.

This plan works backwards from what a finished report should look like, then
derives the generation pipeline from that structure.

---

## Part 0 — Reckoning with the existing `investigations` table

Migration `003_create_investigations.up.sql` (one of the earliest migrations in
the project) introduced an `investigations` table with substantial schema
overlap to the `reports` table this plan proposes:

```
investigations (
  id, trigger_type, trigger_source, summary, severity, status,
  context JSONB, findings JSONB, tool_trace JSONB,
  resolution, started_at, resolved_at
)
```

sqlc generated the full CRUD surface — `CreateInvestigation`, `GetInvestigation`,
`ListOpenInvestigations`, `ListInvestigationsByDateRange`,
`UpdateInvestigationFindings`, `ResolveInvestigation`, `DismissInvestigation`.
**None of these functions are called from anywhere outside the generated file.**
No HTTP handler, no agent code, no tests. It's orphan infrastructure from an
earlier pass at this feature, predating multi-app (migration 014) — which is
why it has no `app_id` and no RLS policies of the modern shape.

Three options, in order of preference:

### Option 1 (recommended) — Extend `investigations` in place

Rename nothing at the database layer. Alter the existing table: add `app_id`,
`user_id` (if not present), `agent_log_id`, `conversation_id`, `title`,
`confidence`, and split the opaque JSONB `findings` into structured columns
(`timeline`, `investigation`, `root_cause`, `impact`, `recommendations`). Add
an RLS policy consistent with other system tables (see migration 030). Delete
the unused sqlc query file and rewrite it to match the new schema.

The UI can still call them "Reports" — the table name is an implementation
detail. This is the cheapest path: no data migration (the table is empty in
practice), no naming cleanup, one forward migration.

**Why it's recommended:** the existing table was designed for exactly this
feature and was abandoned before multi-app. Reviving it costs less than
deleting and re-adding, and avoids leaving two near-identical tables in the
schema.

### Option 2 — Drop `investigations`, create `reports` fresh

Two migrations: `DROP TABLE investigations` and `CREATE TABLE reports (...)`.
Cleaner terminology at the SQL layer, but adds a migration purely to remove
state nothing depends on, and leaves two sqlc query files (one deleted, one
new) in the history for a net nil data change.

### Option 3 — Keep both tables, distinct semantics

`investigations` becomes the agent's internal process log (one row per
investigation run), `reports` is the user-facing artifact (one row per
published report). This sounds cleaner until you write it down: the two would
share >80% of their columns, and the user surface ultimately needs both the
tool trace (investigation) and the synthesis (report). Adds a relational hop
for no real benefit.

### Decision

**Default to Option 1** unless there's a reason not to. The rest of this plan
assumes Option 1 — migration `039_extend_investigations_for_reports.up.sql`
rather than `039_create_reports.up.sql`. If you pick Option 2, swap the
migration name and drop the `DROP TABLE` step in; schema and handler shapes
are identical.

### Also in scope: recent migrations to be aware of

- **031 `assessment_fixes`** — hardened the assessment data shape. Read before
  touching the assessment path in `monitor.go`.
- **038 `log_pipeline_events`** — the Pipeline Page's backing table. A report
  row should carry a foreign-key reference to the triggering assessment's
  pipeline event (or at minimum the `log_id`) so that the "Copy as Markdown"
  export, the report detail page, and the pipeline replay at
  `/pipeline/logs/:logId` can cross-link. See Part 2 for the specific column.

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

Target schema for the `investigations` table after Option 1's extension
migration. Columns marked **(new)** are added by migration 039; the rest
already exist and are kept as-is.

```
investigations              -- surfaced in UI as "Reports"
├── id              UUID PK                         -- existing
├── app_id          UUID FK → applications(id)    NEW  -- scoped to application
├── user_id         UUID FK → users(id)            -- existing (added by migration 011)
├── agent_log_id    UUID FK → agent_log(id) NULL NEW  -- the triggering assessment
├── pipeline_event_id UUID NULL                   NEW  -- soft-link to log_pipeline_events row
├── conversation_id UUID FK → conversations(id) NULL NEW  -- if investigation used interactive loop
│
├── title           TEXT NOT NULL                 NEW  -- 1-line headline
├── summary         TEXT NOT NULL                   -- existing (2-3 sentence overview)
├── severity        TEXT NOT NULL                   -- existing (info | warning | error | critical)
├── status          TEXT NOT NULL DEFAULT 'open'    -- existing (open | investigating | resolved | dismissed)
├── confidence      TEXT NOT NULL DEFAULT 'uncertain' NEW  -- confirmed | likely | uncertain
│
├── trigger_type    TEXT NOT NULL                   -- existing (monitoring | scheduled | manual)
├── trigger_source  TEXT                            -- existing — kept for backcompat
├── trigger_detail  JSONB                         NEW  -- structured: original assessment, schedule metadata
├── context         JSONB NOT NULL                  -- existing — flagged logs / investigation inputs
├── timeline        JSONB NOT NULL DEFAULT '[]'  NEW  -- [{timestamp, event}]
├── investigation   JSONB NOT NULL DEFAULT '[]'  NEW  -- [{tool, input_summary, finding, raw_result}]
├── root_cause      TEXT                          NEW  -- markdown prose, nullable if unknown
├── impact          TEXT                          NEW  -- markdown prose
├── recommendations TEXT                          NEW  -- markdown prose
├── findings        JSONB                           -- existing — kept; synthesis writes split into new columns
├── tool_trace      JSONB                           -- existing ([{tool, input, result}] full fidelity)
│
├── resolution      TEXT                            -- existing
├── started_at      TIMESTAMPTZ NOT NULL DEFAULT now() -- existing
├── resolved_at     TIMESTAMPTZ                     -- existing
├── created_at      TIMESTAMPTZ NOT NULL DEFAULT now() NEW
└── updated_at      TIMESTAMPTZ NOT NULL DEFAULT now() NEW
```

Design notes:
- **`pipeline_event_id` is a soft reference (no FK).** The pipeline-events
  table has a 48h retention cascade (`log_buffer`-tied), per the 0.47.3
  changelog. A report must outlive that window, so the FK would force cascade
  choices that don't make sense. Store the UUID, resolve lazily in the UI, and
  render a "pipeline replay no longer available" fallback when the row is gone.
- **`context` and `findings` are retained** rather than dropped. New code writes
  the structured columns; the old JSONB columns stay nullable and unused, to
  be removed in a later cleanup migration once no rows reference them.
- **`app_id` is `NOT NULL` for new rows, nullable on the table.** Backfill isn't
  needed (the table is empty), but making the column `NOT NULL` in migration
  039 would require a default that doesn't make semantic sense. Enforce
  non-null at the handler/query level instead, following the pattern of
  migration 029 (`app_id_not_null`) — tighten later once rows exist.
- **RLS policy** for the modernised table follows migration 030
  (`rls_system_tables`). Reports must be scoped via `app_id → applications.org_id`
  rather than the legacy `user_id` path.

---

## Part 2 — Report generation pipeline

Working backwards from the report structure, here's what the generation
pipeline needs to produce each section.

### Trigger: `Action: investigate` in monitor output

The system prompt (`backend/internal/agent/prompt.go:37`) instructs Claude to
emit `Action: <none | monitor | investigate>` as part of every assessment.
The monitor loop in `backend/internal/agent/monitor.go` currently parses
severity (via a helper in `extract.go`) but silently discards the `Action:`
line — there is no `parseActionFromResponse` helper anywhere in the codebase.
The first change is to **parse it** and, when `action == "investigate"`,
spawn an investigation. Look for the assessment-parsing call site by grepping
`parseSeverityFromResponse` rather than a line number — the monitor file has
moved considerably since earlier drafts of this plan.

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

Assumes **Option 1** from Part 0 (extend the existing `investigations`
table). Latest migration in `master` is **038_log_pipeline_events**, so the
next free slot is **039**. If Option 2 is chosen instead, swap 1.1/1.2 for
`drop_investigations` + `create_reports` and adjust the query file name —
everything else is identical.

| Step | File | What |
|------|------|------|
| 1.1 | `backend/migrations/039_extend_investigations_for_reports.up.sql` | Add columns per Part 1 schema (`app_id`, `agent_log_id`, `pipeline_event_id`, `conversation_id`, `title`, `confidence`, `trigger_detail`, `timeline`, `investigation`, `root_cause`, `impact`, `recommendations`, `created_at`, `updated_at`). Add RLS policy in the shape of migration 030. |
| 1.2 | `backend/migrations/039_extend_investigations_for_reports.down.sql` | `ALTER TABLE ... DROP COLUMN` the same set; drop the RLS policy. |
| 1.3 | `backend/internal/db/queries/investigations.sql` | **Rewrite** the existing query file — current queries are orphan and unused. New set: `CreateReport`, `GetReportByID`, `ListReportsByApp` (status filter), `UpdateReportStatus`, `UpdateReportFindings` (synthesis-call outputs), `UpdateReportStatusToInvestigating`. All scoped via `app_id → applications.org_id → organizations.user_id`, matching the pattern in `applications.sql`. |
| 1.4 | Run `make sqlc-generate` | Regenerate Go types. Verify the old `CreateInvestigation` / `ResolveInvestigation` / `DismissInvestigation` functions are removed from `backend/internal/db/investigations.sql.go`. |
| 1.5 | `backend/internal/api/handlers/reports.go` | **New file** — no stubs exist. Handlers: `ListReports` (filtered by `app_id`, optional `status`), `GetReport`, `UpdateReportStatus` (resolve/dismiss with optional resolution note). Use `authorizeApp` helper for per-app authorisation, same as other per-app handlers. |
| 1.6 | `backend/internal/api/server.go` | Mount routes under `/api/apps/{appId}/reports` (list) and `/api/reports/{id}` (get + PATCH status). Match the per-app routing pattern already used for connections, logs, and pipeline. |

No report *creation* endpoint — reports are created by the agent pipeline,
not by humans. Humans can only change status (resolve / dismiss) and
optionally add a resolution note. A manual-trigger POST endpoint is a Phase 5
concern, not Phase 1.

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
- **Soft-links to the triggering pipeline event.** Before the investigation
  goroutine exits, the report row's `pipeline_event_id` column is set to
  the `id` of the `log_pipeline_events` row where `stage = 'assessment'`
  and `log_id` matches the triggering log. No new pipeline stage — see
  Open Question #8 for the reasoning. The Pipeline Page's per-log replay
  at `/pipeline/logs/:logId` can later render a "View report" link by
  joining pipeline events to reports via this column; do *not* add that
  join to the existing `ListPipelineLogsByApp` query in this phase (the
  writer-contract invariant it relies on would get harder to reason about).

### Phase 4 — Frontend build (from scratch)

None of the files in this phase exist today. Earlier drafts of this plan
asserted otherwise; that claim was wrong. Build the whole thing new.

| Step | File | What |
|------|------|------|
| 4.1 | `frontend/src/types/report.ts` | **New** — `Report` interface matching the revised schema: `id`, `appId`, `agentLogId`, `pipelineEventId`, `conversationId`, `title`, `summary`, `severity`, `status`, `confidence`, `triggerType`, `triggerDetail`, `timeline`, `investigation`, `rootCause`, `impact`, `recommendations`, `toolTrace`, `resolution`, `startedAt`, `resolvedAt`, `createdAt`, `updatedAt`. Enum unions for `severity`, `status`, `confidence`, `triggerType`. |
| 4.2 | `frontend/src/api/reports.ts` | **New** — `fetchReportsByApp(appId, status?)`, `fetchReport(id)`, `updateReportStatus(id, status, resolution?)`. Follows the axios-client pattern in `api/logs.ts` and `api/pipeline.ts`. |
| 4.3 | `frontend/src/stores/reports.ts` | **New** Pinia composition-API store: `reports` ref, `currentReport` ref, `fetchReportsByApp`, `fetchReport`, `updateStatus` actions. Follow the `stores/pipeline.ts` shape. |
| 4.4 | `frontend/src/pages/ReportsPage.vue` | **New** — list view scoped to `currentAppId` from the `app` store. Filter chips for status (open / investigating / resolved / dismissed). Watch `appId` to refetch on app switch, following `TimeMachineBlock.vue`'s pattern. |
| 4.5 | `frontend/src/components/reports/ReportDetail.vue` | **New** — full report view with all sections from Part 1, collapsible tool trace via `<details>`, back-link to `/pipeline/logs/:logId` when `pipelineEventId` resolves, "Resolve" / "Dismiss" action buttons. |
| 4.6 | `frontend/src/components/reports/ReportCard.vue` | **New** — list row: title, severity pill, status pill, confidence badge, relative timestamp. |
| 4.7 | `frontend/src/lib/reportMarkdown.ts` | **New** — `formatReportAsMarkdown(report: Report): string` pure utility. Consumed by "Copy as Markdown" and "Download .md" buttons on `ReportDetail.vue`. Tested in isolation — this is the portability spec from Part 4 and deserves its own unit tests. |
| 4.8 | `frontend/src/router/index.ts` (or equivalent) | Register `/reports` and `/reports/:reportId` routes. Follow the `/pipeline/logs/:logId` pattern from 0.47.3 for deep-linking. |
| 4.9 | Sidebar nav | Add a "Reports" entry. Vision §Platform Sections lists Reports as a first-class section, so it belongs in the primary nav alongside Activity. |

### Phase 5 — Manual investigation trigger

| Step | File | What |
|------|------|------|
| 5.1 | `backend/internal/api/handlers/reports.go` | `POST /api/apps/{appId}/reports` — create a report with `trigger_type: manual` and a user-provided prompt. Inserts the report row with `status: investigating`, then spawns `RunInvestigation()` as a goroutine (subject to the investigation semaphore from Open Question #1). Respond with 202 Accepted and the report ID so the UI can redirect to `/reports/:id` immediately. |
| 5.2 | Frontend | "Investigate" button on Activity entries (pre-filled with the activity's assessment context) and a "New Investigation" action on the Reports page (free-form prompt). On success, `router.push({ name: 'report-detail', params: { reportId } })`. |

This allows humans to trigger an investigation on demand, not just wait for
the agent to decide `Action: investigate`. Because the Activity-entry variant
preloads assessment context, and the Reports-page variant takes a free-form
prompt, they should share the same POST endpoint but differ only in the
request body — keep the handler input shape permissive.

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
   logic applies to the scheduled investigation code path in
   `backend/internal/agent/scheduler.go` (`RunScheduledInvestigation`).

8. **Pipeline stage for reports** — Phase 3 proposes writing a
   `stage = 'report'` row into `log_pipeline_events` so the Pipeline Page's
   Time Machine picker can show reports as the terminal step of a log's
   journey. Two sub-questions:
   (a) Is "report" a true pipeline stage or should it live in a separate
       table linked by `log_id`? The existing stages (ingestion, classified,
       gate, assessment) are all bounded, per-log, and synchronous with the
       monitor cycle; a report is unbounded (a multi-turn investigation) and
       asynchronous. Adding it as a stage stretches the stage abstraction.
   (b) If we *do* add it: `log_pipeline_events` inherits the 48h retention
       cascade from `log_buffer` (0.47.2 hardening). Reports must outlive
       that window. Either decouple the retention (plan's §4.4 punted item)
       or accept that the pipeline-event link goes stale after 48h while the
       report itself persists, with the UI rendering a "pipeline replay no
       longer available" fallback. The latter is cheaper and matches the
       soft-link decision already baked into the `pipeline_event_id` column
       design (Part 1, data model).
   Leaning toward the soft-link approach: don't add a `report` stage; instead
   link reports to the *assessment* pipeline event they were spawned from,
   and let the pipeline-events retention stay decoupled from report
   retention. Revisit if the UX demands a terminal stage pill on the funnel.
