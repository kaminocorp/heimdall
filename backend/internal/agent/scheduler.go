package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/robfig/cron/v3"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

const maxConcurrentSchedules = 5

// cronParser is the shared 5-field cron parser (Minute Hour Dom Month Dow).
// We construct it once at package init rather than per-call because the parser
// is stateless and reusable — hot path efficiency aside, it also means a
// malformed cron_expr is caught with the same error shape every time.
//
// Five-field (no seconds) matches what users expect from `crontab` and what
// `humanizeCron` on the frontend renders. The constraint is strict enough to
// reject ambiguous input like "* * * * * *" (six fields) with a clear error.
var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// ParseCronExpression is a small wrapper around cronParser.Parse exposed for
// the HTTP handler, which wants to validate cron_expr at create/update time
// so invalid expressions never reach the database. The scheduler reuses the
// same parser internally via shouldFire.
func ParseCronExpression(expr string) (cron.Schedule, error) {
	return cronParser.Parse(expr)
}

// Scheduler tuning.
//
//   - schedulerTickInterval — how often the scheduler wakes up to check for
//     due schedules. 1 minute matches the MVP's minimum allowed interval_secs
//     (60s), so a schedule can't fire more frequently than the tick anyway.
//   - schedulerRunTimeout — hard cap per scheduled run. The 5-minute number
//     mirrors monitor.go's approach to bounding LLM tool-use loops and gives
//     max_iterations * ~30s per iteration of breathing room.
const (
	schedulerTickInterval = 1 * time.Minute
	schedulerRunTimeout   = 5 * time.Minute
)

// InvestigationScheduler runs the scheduled-investigations loop.
//
// Unlike Monitor, which is reactive (polls for new logs and escalates
// classifier-flagged entries), the scheduler is proactive: every tick it
// lists enabled schedules, checks which ones are due based on last_run_at +
// interval_secs, and fires those by calling into RunMonitoring with the
// schedule's stored prompt.
//
// Reusing RunMonitoring is the key architectural move here — it already
// handles provider resolution, tool registry, rate limiting, and
// agent_log emission. Introducing a parallel loop body would duplicate all
// of that and force every future RunMonitoring improvement to be ported.
func (a *Agent) InvestigationScheduler(ctx context.Context) {
	slog.Info("investigation scheduler started", "tick", schedulerTickInterval)
	defer slog.Info("investigation scheduler stopped")

	// Create the semaphore once and pass it to every tick, matching the Monitor
	// pattern. A per-tick semaphore would allow concurrent ticks to exceed the
	// concurrency cap if schedules take longer than one tick interval.
	sem := make(chan struct{}, maxConcurrentSchedules)
	ticker := time.NewTicker(schedulerTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.schedulerTick(ctx, sem)
		case <-ctx.Done():
			return
		}
	}
}

// schedulerTick performs a single scheduling pass: list enabled schedules,
// fire any whose interval has elapsed. Schedules run concurrently with a
// semaphore cap (maxConcurrentSchedules) to prevent a hung LLM call from
// blocking all other schedules.
func (a *Agent) schedulerTick(ctx context.Context, sem chan struct{}) {
	schedules, err := a.queries.ListEnabledSchedules(ctx)
	if err != nil {
		slog.Error("scheduler: list enabled schedules failed", "err", err)
		return
	}

	var wg sync.WaitGroup
	defer wg.Wait() // always join, even on early return from ctx cancellation

	now := time.Now()
	for _, s := range schedules {
		fire, cronErr := shouldFire(s, now)
		if cronErr != "" {
			// Emit to agent_log so the user sees broken cron expressions in the Activity feed.
			// Resolve app → org → user for log attribution (best-effort).
			if app, err := a.queries.GetApplication(ctx, s.AppID); err == nil {
				if userID, err := a.resolveOrgUser(ctx, app.OrgID); err == nil {
					a.EmitLog(ctx, userID, &s.AppID, nil, "schedule_error", cronErr,
						map[string]any{"schedule_id": s.ID, "cron_expr": s.CronExpr.String})
				}
			}
		}
		if !fire {
			continue
		}

		wg.Add(1)
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			wg.Done()
			return
		}

		go func(s db.InvestigationSchedule) {
			defer wg.Done()
			defer func() { <-sem }()
			defer func() {
				if r := recover(); r != nil {
					slog.Error("scheduler: panic in per-schedule goroutine", "schedule_id", s.ID, "app_id", s.AppID, "panic", r)
				}
			}()
			a.RunScheduledInvestigation(ctx, s)
		}(s)
	}
}

// shouldFire decides whether a schedule is due. A schedule fires when:
//  1. It has never run before (last_run_at is nil), OR
//  2. cron_expr is set and its next scheduled time after last_run_at has
//     already passed (Phase 4 cron path), OR
//  3. cron_expr is unset and the time since last_run_at is at least
//     interval_secs seconds (legacy Phase 3 integer-interval path).
//
// Pure function — takes current time as a parameter so tests can pin it.
// The cron path tolerates a malformed cron_expr by logging and returning
// false: we'd rather have a stuck schedule the user can fix via the UI than
// an infinite error loop firing the same broken expression every tick.
// shouldFire returns (true, "") when the schedule is due. If a cron expression
// is malformed, it returns (false, errorMessage) so the caller can surface
// the error in the Activity feed.
func shouldFire(s db.InvestigationSchedule, now time.Time) (bool, string) {
	if s.LastRunAt == nil {
		return true, ""
	}

	// Prefer cron_expr when present. The handler guarantees that at create/
	// update time exactly one of cron_expr or interval_secs is "active", but
	// the shouldFire contract is broader — it handles any stored row. The
	// rule here is simply "cron wins if set", which maps cleanly to how the
	// UI will present the two modes.
	if s.CronExpr.Valid && s.CronExpr.String != "" {
		sched, err := cronParser.Parse(s.CronExpr.String)
		if err != nil {
			slog.Error("scheduler: invalid cron expression",
				"schedule_id", s.ID,
				"cron", s.CronExpr.String,
				"err", err)
			return false, fmt.Sprintf("Schedule %s has invalid cron expression %q: %v", s.ID, s.CronExpr.String, err)
		}
		next := sched.Next(*s.LastRunAt)
		return !next.After(now), ""
	}

	// Legacy interval_secs path — unchanged from Phase 3.
	elapsed := now.Sub(*s.LastRunAt)
	return elapsed >= time.Duration(s.IntervalSecs)*time.Second, ""
}

// RunScheduledInvestigation is the entire fire path for one schedule:
// look up the app, resolve a user for agent_log attribution, load the agent
// config, rate-limit, invoke RunMonitoring with the stored prompt, emit to
// agent_log, and mark the schedule as run.
//
// Exported receiver-public so the RunScheduleNow HTTP handler can invoke
// this synchronously without going through the tick loop. Accepts the full
// schedule row rather than just an ID so tests can construct stubs without
// round-tripping through the DB.
func (a *Agent) RunScheduledInvestigation(ctx context.Context, s db.InvestigationSchedule) {
	runCtx, cancel := context.WithTimeout(ctx, schedulerRunTimeout)
	defer cancel()

	// Resolve app → OrgID → a user for agent_log attribution. We use the
	// existing GetApplication query (not GetApplicationByID, which the plan
	// assumed might not exist — it already does, just under a different name).
	app, err := a.queries.GetApplication(runCtx, s.AppID)
	if err != nil {
		slog.Error("scheduler: load app failed", "err", err, "schedule_id", s.ID)
		a.markRunError(ctx, s.ID, "load app: "+err.Error())
		return
	}
	userID, err := a.resolveOrgUser(runCtx, app.OrgID)
	if err != nil {
		slog.Error("scheduler: resolve user failed", "err", err, "schedule_id", s.ID)
		a.markRunError(ctx, s.ID, "resolve user: "+err.Error())
		return
	}

	appConfig, err := a.queries.GetAppAgentConfig(runCtx, s.AppID)
	if err != nil {
		// Missing config is survivable — the agent uses the default model
		// and an empty provider (which falls back to anthropic). We log
		// because it usually indicates a CreateApplication race or an app
		// that was created out-of-band.
		slog.Warn("scheduler: app agent config missing, using default", "err", err, "app_id", s.AppID)
		appConfig = db.AppAgentConfig{AppID: s.AppID, Model: DefaultModelID}
	}

	// Share the monitor's rate limiter. Scheduled runs count against the
	// same 30 rpm / burst 5 budget, so a dozen noisy schedules can't
	// starve the monitor loop running next door. Flagged by the plan's
	// open-question #2 — we're answering "yes, share the limiter".
	if err := a.limiter.Wait(runCtx); err != nil {
		slog.Warn("scheduler: rate limit wait interrupted", "schedule_id", s.ID, "err", err)
		return
	}

	// Reuse the monitoring entry point. Note that the "flaggedLogs" parameter
	// is misnamed for this path — we pass the schedule's prompt verbatim.
	// RunMonitoring doesn't care; it just feeds the string to the LLM as the
	// first user message. This is the architectural reuse that justifies not
	// building a parallel loop body.
	assessment, severity, providerFailed := a.RunMonitoring(runCtx, userID, appConfig, s.Prompt)
	if providerFailed {
		slog.Warn("scheduler: LLM provider failed", "schedule_id", s.ID, "app_id", s.AppID)
		// Mark the run as errored so last_run_at advances. Without this, the
		// schedule fires again on the very next tick (every 60s), creating a
		// tight retry loop that burns rate-limiter tokens while the provider is down.
		a.markRunError(ctx, s.ID, "LLM provider failed")
		return
	}

	// Truncate for storage/display. The 200-rune cap matches what monitor.go
	// uses for its own summary field — keeps agent_log rows cheap to scan
	// and avoids blowing up the Activity feed with paragraphs.
	summary := assessment
	if utf8.RuneCountInString(summary) > 200 {
		summary = string([]rune(summary)[:200]) + "..."
	}

	// Emit to agent_log with a distinct entry_type so the UI (Phase 4) can
	// render scheduled-investigation rows differently from classifier-driven
	// monitoring rows. Detail carries the schedule metadata for traceability.
	//
	// We use ctx (parent) rather than runCtx (timed-out) for the emit so that
	// a slow-but-succeeded RunMonitoring can still persist its result even if
	// the run context is on the edge of expiring.
	a.EmitLogWithSeverity(ctx, userID, &s.AppID, nil, "scheduled_investigation", summary,
		map[string]any{
			"schedule_id":   s.ID,
			"schedule_name": s.Name,
			"app_id":        s.AppID,
			"app_name":      app.Name,
			"assessment":    assessment,
			"auto_severity": severity,
		},
		severity,
	)

	a.markRunSuccess(ctx, s.ID, summary)
}

// markRunSuccess updates last_run_at, sets last_status='success', and
// stores the truncated summary. Errors are logged and swallowed — MarkRun
// is best-effort bookkeeping, not load-bearing for correctness (the next
// tick's `shouldFire` only uses last_run_at, which is advanced regardless).
func (a *Agent) markRunSuccess(ctx context.Context, id uuid.UUID, summary string) {
	if err := a.queries.MarkScheduleRun(ctx, db.MarkScheduleRunParams{
		ID:          id,
		LastStatus:  pgtype.Text{String: "success", Valid: true},
		LastError:   pgtype.Text{}, // explicit NULL clears any stale error
		LastSummary: pgtype.Text{String: summary, Valid: true},
	}); err != nil {
		slog.Warn("scheduler: failed to mark run success", "err", err, "schedule_id", id)
	}
}

// markRunError is the error-path counterpart to markRunSuccess. Captures the
// error text in last_error so the UI can surface it to operators without
// digging through server logs.
func (a *Agent) markRunError(ctx context.Context, id uuid.UUID, errMsg string) {
	if err := a.queries.MarkScheduleRun(ctx, db.MarkScheduleRunParams{
		ID:          id,
		LastStatus:  pgtype.Text{String: "error", Valid: true},
		LastError:   pgtype.Text{String: errMsg, Valid: true},
		LastSummary: pgtype.Text{}, // clear any previous success summary
	}); err != nil {
		slog.Warn("scheduler: failed to mark run error", "err", err, "schedule_id", id)
	}
}
