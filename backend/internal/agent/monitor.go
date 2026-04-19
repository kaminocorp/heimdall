package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/db"
	"github.com/hejijunhao/heimdall/backend/internal/metrics"
)

const (
	monitorTickInterval = 15 * time.Second
	maxConcurrentApps   = 10
	logBatchLimit       = 200
	maxFlaggedForLLM    = 50
	maxPayloadChars     = 2000
	monitorAppTimeout   = 2 * time.Minute
)

// Monitor runs the monitoring loop, polling active applications
// and processing their logs through the classification pipeline.
func (a *Agent) Monitor(ctx context.Context) {
	slog.Info("monitoring goroutine started")
	defer slog.Info("monitoring goroutine stopped")

	sem := make(chan struct{}, maxConcurrentApps)
	ticker := time.NewTicker(monitorTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.monitorTick(ctx, sem)
		case <-ctx.Done():
			return
		}
	}
}

// monitorTick runs a single monitoring cycle across all active applications.
func (a *Agent) monitorTick(ctx context.Context, sem chan struct{}) {
	tickStart := time.Now()
	defer func() { metrics.MonitorTickDuration.Observe(time.Since(tickStart).Seconds()) }()

	apps, err := a.queries.ListActiveApplications(ctx)
	if err != nil {
		slog.Error("monitor: failed to list active applications", "err", err)
		return
	}

	if len(apps) == 0 {
		return
	}

	var wg sync.WaitGroup
	defer wg.Wait() // always join, even on early return from ctx cancellation
	for _, app := range apps {
		if !a.shouldMonitor(ctx, app) {
			continue
		}
		wg.Add(1)
		// Acquire semaphore with shutdown awareness so a full semaphore
		// doesn't prevent the monitor goroutine from responding to ctx cancellation.
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			wg.Done()
			return
		}
		go func(app db.ListActiveApplicationsRow) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore
			defer func() {
				if r := recover(); r != nil {
					slog.Error("monitor: panic in per-app goroutine", "app_id", app.ID, "panic", r)
				}
			}()
			appCtx, cancel := context.WithTimeout(ctx, monitorAppTimeout)
			defer cancel()
			a.monitorApp(appCtx, app)
		}(app)
	}
}

// shouldMonitor checks whether enough time has elapsed since the last
// monitoring cycle for this application.
func (a *Agent) shouldMonitor(ctx context.Context, app db.ListActiveApplicationsRow) bool {
	// Continuous mode always monitors on every tick.
	if app.Mode == "continuous" {
		return true
	}

	// Periodic mode: check if the schedule interval has elapsed.
	state, err := a.queries.GetMonitoringState(ctx, app.ID)
	if err != nil {
		// No state yet — first run, should monitor.
		return true
	}

	elapsed := time.Since(state.LastMonitoredAt)
	return elapsed >= time.Duration(app.ScheduleIntervalSecs)*time.Second
}

// monitorApp processes a single application's logs through the
// classification pipeline and escalates flagged logs to the LLM.
func (a *Agent) monitorApp(ctx context.Context, app db.ListActiveApplicationsRow) {
	// Resolve a user ID for this org (for agent_log attribution).
	userID, err := a.resolveOrgUser(ctx, app.OrgID)
	if err != nil {
		slog.Error("monitor: failed to resolve org user", "err", err, "app_id", app.ID, "org_id", app.OrgID)
		return
	}

	// Get cursor position.
	state, err := a.queries.GetMonitoringState(ctx, app.ID)
	if err != nil {
		// First run: initialize cursor to now so we don't process historical logs.
		if err := a.queries.ResetMonitoringCursor(ctx, app.ID); err != nil {
			slog.Error("monitor: failed to initialize cursor", "err", err, "app_id", app.ID)
			return
		}
		slog.Info("monitor: initialized cursor for app", "app_id", app.ID, "app_name", app.Name)
		return
	}

	// Fetch new logs since cursor.
	logs, err := a.queries.ListLogsSinceForApp(ctx, db.ListLogsSinceForAppParams{
		AppID:      app.ID,
		IngestedAt: state.LastMonitoredAt,
		Limit:      logBatchLimit,
	})
	if err != nil {
		slog.Error("monitor: failed to fetch logs", "err", err, "app_id", app.ID)
		return
	}

	if len(logs) == 0 {
		return
	}

	// Classify logs through the pipeline. The classifier returns one
	// ClassifiedLog per input (escalated and safe alike) so the Pipeline
	// page can render per-log stage events for every log — not just the
	// flagged subset. FilterFlagged pulls out the LLM-bound slice.
	results := a.classifier.Classify(logs)
	flagged, safeCount := FilterFlagged(results)
	metrics.LogsClassifiedTotal.WithLabelValues("safe").Add(float64(safeCount))
	metrics.LogsClassifiedTotal.WithLabelValues("flagged").Add(float64(len(flagged)))

	// Emit Pipeline-page Classified + Gate events for every log. Fire-
	// and-forget: any write failure is logged inside pipeline_writer and
	// never bubbles out here, so a pipeline-events table outage can't
	// stall monitoring.
	if pw := a.pipeline; pw != nil {
		for _, r := range results {
			pw.WriteClassified(ctx, ClassifiedInput{
				LogID:      r.Log.ID,
				AppID:      r.Log.AppID,
				Type:       r.Type,
				Category:   r.Category,
				Severity:   r.Severity,
				Confidence: r.Confidence,
				Summary:    r.Summary,
			})
			pw.WriteGate(ctx, GateInput{
				LogID:     r.Log.ID,
				AppID:     r.Log.AppID,
				Escalated: r.Escalated,
				RuleID:    r.RuleID,
			})
		}
	}

	// If flagged logs exist, escalate to LLM.
	if len(flagged) > 0 {
		escalated := flagged
		if len(escalated) > maxFlaggedForLLM {
			dropped := len(flagged) - maxFlaggedForLLM
			slog.Warn("monitor: capping flagged logs for LLM", "total", len(flagged), "cap", maxFlaggedForLLM, "dropped", dropped, "app_id", app.ID)
			escalated = escalated[:maxFlaggedForLLM]

			a.EmitLog(ctx, userID, &app.ID, nil, "monitoring",
				fmt.Sprintf("%d flagged logs exceeded the per-cycle cap (%d) and were not assessed by the LLM", dropped, maxFlaggedForLLM),
				map[string]any{"app_id": app.ID, "total_flagged": len(flagged), "cap": maxFlaggedForLLM, "dropped": dropped},
			)
		}

		appConfig, err := a.queries.GetAppAgentConfig(ctx, app.ID)
		if err != nil {
			slog.Error("monitor: failed to load app agent config", "err", err, "app_id", app.ID)
			// Continue with default config.
			appConfig = db.AppAgentConfig{
				AppID: app.ID,
				Model: "claude-sonnet-4-6",
			}
		}

		// Acquire rate limiter token before calling the LLM.
		// This bounds Claude API cost if the classifier falls back to passthrough.
		if err := a.limiter.Wait(ctx); err != nil {
			slog.Warn("monitor: LLM rate limit wait interrupted", "app_id", app.ID, "err", err)
			return
		}

		input := formatFlaggedLogs(app, escalated)
		assessment, severity, providerFailed := a.RunMonitoring(ctx, userID, appConfig, input)

		// If the LLM provider failed, do not advance the cursor so these
		// logs are reprocessed on the next tick.
		if providerFailed {
			slog.Warn("monitor: LLM provider failed, cursor not advanced", "app_id", app.ID)
			return
		}

		summary := assessment
		if utf8.RuneCountInString(summary) > 200 {
			summary = string([]rune(summary)[:200]) + "..."
		}
		logEntryID := a.EmitLogWithSeverity(ctx, userID, &app.ID, nil, "monitoring", summary,
			map[string]any{
				"app_id":         app.ID,
				"app_name":       app.Name,
				"flagged":        len(flagged),
				"assessment":     assessment,
				"auto_severity":  severity,
			},
			severity,
		)

		// Emit one Pipeline-page Assessment event per flagged log that
		// actually reached the LLM (the post-cap slice, not the full
		// flagged set). Carrying the agent_log.id lets the Time Machine
		// replay view link each log's particle back to the assessment
		// that evaluated it. Skipped when EmitLog returned uuid.Nil (the
		// agent_log insert itself failed) — a dangling assessment_id
		// would be misleading.
		if pw := a.pipeline; pw != nil && logEntryID != uuid.Nil {
			for _, cl := range escalated {
				pw.WriteAssessment(ctx, AssessmentInput{
					LogID:        cl.Log.ID,
					AppID:        cl.Log.AppID,
					AssessmentID: logEntryID,
				})
			}
		}

		// Dispatch notification (fire-and-forget).
		// Use context.WithoutCancel so the notification isn't killed when monitorApp returns,
		// plus a 30s timeout to prevent permanent goroutine leaks if the target hangs.
		if a.notifier != nil && logEntryID != uuid.Nil {
			go func() {
				notifyCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
				defer cancel()
				a.notifier.Notify(notifyCtx, app.ID, logEntryID, app.Name, severity, summary, assessment)
			}()
		}
	}

	// Advance cursor to the last processed log's timestamp.
	lastLog := logs[len(logs)-1]
	if err := a.queries.UpsertMonitoringState(ctx, db.UpsertMonitoringStateParams{
		AppID:           app.ID,
		LastMonitoredAt: lastLog.IngestedAt,
	}); err != nil {
		slog.Error("monitor: failed to advance cursor", "err", err, "app_id", app.ID)
	}
}

// resolveOrgUser finds the first user in the organization for log attribution.
func (a *Agent) resolveOrgUser(ctx context.Context, orgID uuid.UUID) (uuid.UUID, error) {
	return a.queries.GetFirstUserInOrg(ctx, orgID)
}

// formatFlaggedLogs builds a text summary of flagged logs for the LLM.
func formatFlaggedLogs(app db.ListActiveApplicationsRow, flagged []ClassifiedLog) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Application: %s\nFlagged logs (%d):\n\n", app.Name, len(flagged))

	for i, cl := range flagged {
		fmt.Fprintf(&b, "--- Log %d ---\n", i+1)
		fmt.Fprintf(&b, "Classification: %s.%s (confidence: %.2f, severity: %s)\n", cl.Type, cl.Category, cl.Confidence, cl.Severity)
		fmt.Fprintf(&b, "Summary: %s\n", cl.Summary)
		fmt.Fprintf(&b, "Timestamp: %s\n", cl.Log.IngestedAt.Format(time.RFC3339))
		payload := string(cl.Log.Payload)
		if utf8.RuneCountInString(payload) > maxPayloadChars {
			payload = string([]rune(payload)[:maxPayloadChars]) + "... [truncated]"
		}
		fmt.Fprintf(&b, "Raw payload: %s\n\n", payload)
	}

	return b.String()
}
