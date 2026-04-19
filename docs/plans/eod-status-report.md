# EOD Status Report — Feature Proposals

## Original Brief

Given that we receive and record/briefly store all incoming logs, we could design a new feature/function:

- An EOD cron job that reviews all of that day's logs (even if no negative events occured that triggered an investigation) that uses an LLM to summarise into an EOD status report
- This report then gets sent to X team members as defined/configured on Heimdall

Get what I mean?

If you're clear about this idea, outline a few proposals for what this could look like / how it might work/be implemented into our application, in this md file.

---

## Framing

The idea sits at the intersection of three primitives Heimdall already ships:

1. **Scheduled Investigations** (0.31.0) — cron-driven Claude prompts that already write to the Activity feed and can escalate to Reports.
2. **Notifications & Escalation** (0.15.0) — the existing fan-out rail for "tell humans about this."
3. **Pipeline / log_buffer + log_pipeline_events** — 48h-retention raw + per-log journey data, suitable for daily windowing.

Two design tensions shape the options below:

- **Reuse vs. first-class.** Do we treat the digest as a specialised scheduled investigation, or as its own entity (config + runs + delivery)?
- **Raw vs. aggregated input to the LLM.** A noisy app can emit 50k–500k log lines/day. The LLM never sees raw logs at that scale — it consumes pre-computed aggregates and reads raw samples only on demand. This is the single biggest cost/scope lever.

---

## Proposal A — Lightest: "Digest" template on top of Scheduled Investigations

Treat the EOD report as a specialised scheduled investigation prompt, distinguished by a `kind = 'digest'` flag and a recipients list.

**What lands:**
- Add `kind` (`'investigation' | 'digest'`) and `recipient_user_ids uuid[]` columns to `scheduled_investigations`.
- A "Daily summary at 18:00 in {timezone}" template prompt: `"Summarise today's log activity for {app}. Counts by severity, top 5 error sources, anything anomalous vs. yesterday."`
- The existing scheduler picks it up and fires it through the existing agent loop. The agent uses `search_logs` / `query_database` exactly as it does today — no new tools.
- Output rendered as the existing investigation result page, plus an email/Slack notification to recipients linking back.

**Why it's tempting:** ~200 LOC. No new tables, no new background worker, no new UI surface beyond a recipients picker on the existing scheduled-investigations form.

**Why it's wrong long-term:**
- Investigations are *prompted* explorations; digests are *structured* outputs. Conflating them creates UX drift the moment the digest needs sections, KPIs, or comparisons to prior periods.
- The agent's tool-loop is built for "find the answer" with bounded iteration. It's a clumsy fit for "render this fixed structure deterministically."
- Cost: the agent will issue several `search_logs` calls per run that are predictable and could be precomputed once instead of relearned daily.

Use this only as a **prototype** to validate that recipients actually read and act on a daily summary before committing to Proposal B.

---

## Proposal B — Recommended: First-class Daily Digest feature

A dedicated subsystem with its own data model, scheduler, aggregation pipeline, LLM step, and delivery — composed from existing parts where they fit, isolated where they don't.

### Data model

Two new tables, app-scoped to match the rest of the platform:

```sql
-- migrations/0XX_digest_configs.up.sql
CREATE TABLE digest_configs (
  id              uuid PRIMARY KEY,
  app_id          uuid NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
  name            text NOT NULL,                 -- e.g. "EOD summary"
  schedule_cron   text NOT NULL,                 -- "0 18 * * *"
  timezone        text NOT NULL DEFAULT 'UTC',   -- IANA tz
  window_hours    int  NOT NULL DEFAULT 24
                  CHECK (window_hours BETWEEN 1 AND 48),
  sections        jsonb NOT NULL DEFAULT '[]',   -- ["summary","anomalies","top_errors","by_source","agent_findings"]
  recipient_user_ids uuid[] NOT NULL DEFAULT '{}',
  paused          bool NOT NULL DEFAULT false,
  last_run_at     timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE digest_runs (
  id               uuid PRIMARY KEY,
  digest_config_id uuid NOT NULL REFERENCES digest_configs(id) ON DELETE CASCADE,
  period_start     timestamptz NOT NULL,
  period_end       timestamptz NOT NULL,
  aggregates       jsonb NOT NULL,    -- the structured input given to the LLM
  llm_summary_md   text,              -- the rendered digest body (markdown)
  status           text NOT NULL,     -- 'pending' | 'sent' | 'failed'
  error            text,
  sent_at          timestamptz,
  created_at       timestamptz NOT NULL DEFAULT now()
);
```

The `window_hours BETWEEN 1 AND 48` check is the load-bearing line — it makes the 48h `log_buffer` retention an explicit invariant, not an implicit one. Weekly digests are out of scope until retention changes (see §Open Questions).

RLS on both tables, app-scoped via the existing `org_id → applications.id` chain — same pattern as `monitoring_state`.

### Scheduler

A new `internal/agent/digest_scheduler.go` modelled on `monitor.go`:

- 60s tick (digests don't need the monitor's 15s cadence).
- For each unpaused `digest_configs` row, compute the next fire time from `schedule_cron + timezone`. If `now() >= next_fire`, enqueue a run.
- Bounded concurrency (semaphore = 4) to prevent a flood of large apps from saturating the LLM tier.
- Crash-safe: a `digest_runs` row is inserted with `status = 'pending'` *before* the LLM call, so a restart can reconcile orphans rather than silently re-firing.

### Aggregation step (the part that protects token budget)

Before any LLM call, run pure-SQL aggregations over the window. Roughly:

```sql
-- counts by severity
SELECT severity, COUNT(*)
FROM log_buffer WHERE app_id = $1 AND occurred_at BETWEEN $2 AND $3
GROUP BY severity;

-- top error sources
SELECT source, COUNT(*) AS n
FROM log_buffer WHERE app_id = $1 AND occurred_at BETWEEN $2 AND $3 AND severity IN ('error','fatal')
GROUP BY source ORDER BY n DESC LIMIT 10;

-- pipeline outcomes (uses log_pipeline_events)
SELECT
  COUNT(*) FILTER (WHERE stage = 'gate' AND escalated)         AS escalated,
  COUNT(*) FILTER (WHERE stage = 'classified' AND classification = 'flagged') AS flagged,
  COUNT(*) FILTER (WHERE stage = 'assessment')                 AS assessed
FROM log_pipeline_events WHERE app_id = $1 AND occurred_at BETWEEN $2 AND $3;

-- agent findings (already in agent_log / reports)
SELECT id, severity, summary, created_at
FROM reports WHERE app_id = $1 AND created_at BETWEEN $2 AND $3;
```

Plus message-cluster sampling: bucket error messages by a normalised template (regex-stripped IDs/numbers) and pick the top N representative samples. This is the "what changed today" signal without sending 100k lines.

The `aggregates` JSONB stored on `digest_runs` is the *complete* input to the LLM — reproducible, replayable, auditable.

### LLM step

A single non-tool-use Claude call (no `search_logs`/`query_database` loop — those are for investigation, not summarisation). Prompt template:

```
You are summarising 24h of production activity for {app_name}.
Input is structured aggregates, not raw logs. Do not hallucinate counts.

<aggregates>{json}</aggregates>

Produce a markdown digest with sections: {sections}.
For "anomalies", compare today's counts to the prior 24h (provided in aggregates.prior_period).
Be brief. If nothing notable, say so — do not pad.
```

Prompt caching on the system prompt + the static section template, since the per-day variable input is small.

Model choice: Haiku 4.5 by default — this is a structured-summarisation task, not investigation, and Haiku is plenty for it at a tenth the cost. User-overridable per `digest_configs` row if they want Opus for the daily exec summary.

### Delivery

Extend the existing `internal/notifications` package with a `DigestChannel`:

- **Email** (primary): rendered markdown → HTML, sent via the same SMTP/Resend integration that 0.15.0 already uses.
- **Slack** (secondary): summary as a Slack block-kit message with a "View full digest" link to the persisted `digest_runs` row in the UI.
- Recipients resolved from `recipient_user_ids` against `users.email`; fail loudly (and write to `digest_runs.error`) if a recipient lacks a deliverable address.

### Frontend

Two surfaces:

1. **Digests tab** (per app, sidebar): list of `digest_configs` with paused/active toggle, "Run now" button, and recent run history.
2. **Digest detail page**: rendered markdown digest, the structured aggregates that produced it (collapsible), recipient list with delivery status, "Re-send" button.

The configuration form is intentionally small — schedule, timezone, recipients, sections checklist, model dropdown. No prompt editor in v1 (that's the slippery slope back to Proposal A).

### Scope estimate

~600–900 LOC across backend + frontend, two migrations, ~150 LOC of integration tests. Roughly the size of the 0.31.0 Scheduled Investigations release, with most of the effort going into the aggregation queries + the recipients/delivery polish.

---

## Proposal C — Heaviest: Generic "Reports & Digests" subsystem

Promote the digest concept into a fully generic reporting layer:

- Report templates (daily, weekly, monthly, custom).
- Multi-app and org-level rollups ("digest for all of `acme-corp`'s apps").
- Subscription model: users opt themselves in/out, with per-recipient delivery preferences.
- Comparative trends over arbitrary windows (requires snapshotting aggregates beyond the 48h `log_buffer` window into a new `daily_aggregates` table — a real new piece of infrastructure).

**Why not now:** Every additional axis (template engine, multi-app rollup, snapshot aggregates, subscription UX) is independently justified only after the v1 daily digest is shown to be valuable. Building this first is premature abstraction; building Proposal B first and *then* generalising is cheap.

The right path is B → observe usage → graduate to C if and when users ask for weekly digests, multi-app rollups, or executive summaries.

---

## Recommendation

**Ship Proposal B.** It's the smallest design that doesn't paint us into a corner:

- Reuses the existing scheduler shape (`monitor.go` is a known-good template), the existing notifications rail, the existing app-scoping and RLS patterns.
- Treats the LLM as a *summariser of aggregates*, not an investigator — which is what makes the cost story sane at any log volume.
- Stores the aggregates separately from the rendered summary, so we can rerun or reformat without re-paying the ingestion cost.
- Leaves a clean evolution path to Proposal C (weekly digests, multi-app rollups, comparative trends) without needing to redesign the data model.

If you want a fast validation step before committing the full ~700 LOC, we can prototype Proposal A behind a feature flag in a few hundred lines, run it for a week against your own org, and use what we learn to refine B's section list before building it.

---

## Open Questions

1. **Scope: app-level or org-level digests?** All current resources (connections, agent config, monitoring state) are per-app. App-level is the consistent choice, but org-level rollups are a natural ask. Pick one for v1.
2. **Recipients: org members only, or arbitrary email addresses?** Org-members-only is simpler (we already have their addresses); arbitrary emails opens up "send to my non-Heimdall manager" but adds an email-verification surface.
3. **Delivery channels for v1: email only, or email + Slack?** Slack is a common destination but requires per-org workspace OAuth — non-trivial.
4. **Expected daily log volume per app?** Drives whether the aggregation step needs streaming computation or fits comfortably in a single SQL pass. Order-of-magnitude estimate is enough.
5. **Should digests live as a new entity (this proposal) or be folded into the existing `reports` model with a `kind = 'digest'` discriminator?** Reuse is cleaner if you intend digests to surface in the Reports browse view; separation is cleaner if Digests deserve their own navigation.
6. **Multiple schedules per app?** e.g. an EOD digest + a Monday-morning weekly. v1 trivially supports this since `digest_configs` is a 1:N relation; just need the UX to make it discoverable.
7. **Retention beyond 48h.** Daily digests fit the current `log_buffer` window. Weekly digests do not — they require either (a) extended `log_buffer` retention, or (b) a `daily_aggregates` snapshot table that the weekly digest reads from. Is weekly a v1 requirement or a "later" one?
8. **Default-on or default-off for new apps?** A new app with no recipients configured shouldn't auto-generate digests, but having a sensible "Daily summary at 18:00 UTC" template pre-seeded (paused, recipient list empty) makes the feature discoverable.
9. **Localisation of the digest body** — generate in English only, or follow a per-org `locale` setting? (No `locale` column exists today; this would be the first feature to need one.)
10. **Audit/compliance:** should `digest_runs.aggregates` retention match log retention, or should digests be kept indefinitely as a historical record of "what was true on day X"? The latter is more useful but creates a new long-lived data class.
