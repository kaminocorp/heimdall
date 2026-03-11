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
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/db"
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
	apps, err := a.queries.ListActiveApplications(ctx)
	if err != nil {
		slog.Error("monitor: failed to list active applications", "err", err)
		return
	}

	if len(apps) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, app := range apps {
		if !a.shouldMonitor(ctx, app) {
			continue
		}
		wg.Add(1)
		sem <- struct{}{} // acquire semaphore
		go func(app db.ListActiveApplicationsRow) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore
			appCtx, cancel := context.WithTimeout(ctx, monitorAppTimeout)
			defer cancel()
			a.monitorApp(appCtx, app)
		}(app)
	}
	wg.Wait()
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

	// Classify logs through the pipeline.
	flagged, safeCount := a.classifier.Classify(logs)

	// Emit heartbeat (always, even if no flagged logs).
	a.EmitLogWithSeverity(ctx, userID, nil, "heartbeat",
		fmt.Sprintf("[%s] Processed %d logs: %d safe, %d flagged", app.Name, len(logs), safeCount, len(flagged)),
		map[string]any{
			"app_id":         app.ID,
			"app_name":       app.Name,
			"logs_processed": len(logs),
			"safe":           safeCount,
			"flagged":        len(flagged),
		},
		"info",
	)

	// If flagged logs exist, escalate to LLM.
	if len(flagged) > 0 {
		escalated := flagged
		if len(escalated) > maxFlaggedForLLM {
			slog.Warn("monitor: capping flagged logs for LLM", "total", len(flagged), "cap", maxFlaggedForLLM, "app_id", app.ID)
			escalated = escalated[:maxFlaggedForLLM]
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

		input := formatFlaggedLogs(app, escalated)
		assessment, severity := a.RunMonitoring(ctx, userID, appConfig, input)

		summary := assessment
		if utf8.RuneCountInString(summary) > 200 {
			summary = string([]rune(summary)[:200]) + "..."
		}
		logEntryID := a.EmitLogWithSeverity(ctx, userID, nil, "monitoring", summary,
			map[string]any{
				"app_id":         app.ID,
				"app_name":       app.Name,
				"flagged":        len(flagged),
				"assessment":     assessment,
				"auto_severity":  severity,
			},
			severity,
		)

		// Dispatch notification (fire-and-forget).
		// Use context.WithoutCancel so the notification isn't killed when monitorApp returns.
		if a.notifier != nil && logEntryID != uuid.Nil {
			go a.notifier.Notify(context.WithoutCancel(ctx), app.ID, logEntryID, app.Name, severity, summary, assessment)
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
	pgOrgID := pgtype.UUID{Bytes: orgID, Valid: true}
	return a.queries.GetFirstUserInOrg(ctx, pgOrgID)
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
