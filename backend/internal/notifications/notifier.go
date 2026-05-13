package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/db"
	"github.com/hejijunhao/heimdall/backend/internal/metrics"
)

// Payload represents a notification to be sent.
type Payload struct {
	AppName    string
	Severity   string
	Summary    string
	Assessment string
	Timestamp  string
}

// Channel sends notifications to a specific destination.
type Channel interface {
	Type() string
	Send(ctx context.Context, payload Payload) error
}

// NewChannel creates a Channel from its type and config JSONB.
func NewChannel(channelType string, configBytes json.RawMessage, cfg *config.Config) (Channel, error) {
	switch channelType {
	case "email":
		return newEmailChannel(configBytes, cfg)
	case "slack":
		return newSlackChannel(configBytes)
	case "discord":
		return newDiscordChannel(configBytes)
	default:
		return nil, fmt.Errorf("unknown channel type: %s", channelType)
	}
}

// Dispatcher checks notification preferences, applies filters, and dispatches to enabled channels.
//
// pools is the canonical DB handle: every Notify call resolves to one
// owning user (passed by the caller — typically monitor.go's monitorApp)
// and the dispatcher opens a UserQueries scope on the App pool to do its
// reads (preferences, channels) and writes (notification_log) under
// app.current_user_id == userID. This matches the post-Phase-6 RLS
// posture where notification_log's policy scopes by applications →
// org_members → user.
type Dispatcher struct {
	pools  *db.Pools
	config *config.Config
}

// NewDispatcher creates a new notification dispatcher.
func NewDispatcher(pools *db.Pools, cfg *config.Config) *Dispatcher {
	return &Dispatcher{pools: pools, config: cfg}
}

// Notify is called by the monitoring loop after emitting an agent_log entry.
// Fire-and-forget: errors are logged, never returned.
//
// userID is the agent_log row's owner — the dispatcher opens its
// UserQueries scope under that user so notification_log writes satisfy
// the table's per-app RLS policy.
func (d *Dispatcher) Notify(ctx context.Context, userID uuid.UUID, appID uuid.UUID, agentLogID uuid.UUID, appName, severity, summary, assessment string) {
	// UserQueriesForLoop (not UserQueries) because dispatchToChannel posts
	// to external HTTP endpoints (Slack, Discord, SMTP) inside this txn.
	// A slow Slack webhook (multi-second tail) plus Supabase's per-role
	// 8s statement_timeout would otherwise abort the next markSent /
	// markFailed update — leaving notification_log rows stuck in
	// `pending` and the metric counters out of sync with reality. The
	// loop variant disables both idle-in-txn and per-statement timeouts
	// for the duration. The trade-off (one App-pool conn pinned for the
	// dispatch window) is bounded: max ~3 channels × ~5s send budget =
	// ~15s typical, ~30s worst case.
	q, commit, done, err := d.pools.UserQueriesForLoop(ctx, userID)
	if err != nil {
		slog.Warn("notification: failed to open UserQueries", "err", err, "app_id", appID, "user_id", userID)
		return
	}
	defer done()

	prefs, err := q.GetNotificationPreferences(ctx, appID)
	if err != nil || !prefs.Enabled {
		return
	}

	if severityRank(severity) < severityRank(prefs.SeverityThreshold) {
		return
	}

	lastNotif, err := q.GetLastNotificationForApp(ctx, appID)
	if err == nil && time.Since(lastNotif.CreatedAt) < time.Duration(prefs.CooldownMinutes)*time.Minute {
		slog.Debug("notification suppressed by cooldown", "app_id", appID, "last_sent", lastNotif.CreatedAt)
		return
	}

	channels, err := q.ListEnabledChannelsByApp(ctx, appID)
	if err != nil || len(channels) == 0 {
		return
	}

	payload := Payload{
		AppName:    appName,
		Severity:   severity,
		Summary:    summary,
		Assessment: assessment,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	// Insert notification_log rows + dispatch each channel inside the
	// UserQueries scope. Send-side errors update the same row's status
	// via markFailed/markSent — staying inside the txn keeps log
	// state consistent with the dispatch outcome.
	for _, ch := range channels {
		d.dispatchToChannel(ctx, q, ch, appID, agentLogID, payload)
	}

	if err := commit(); err != nil {
		slog.Warn("notification: commit failed", "err", err, "app_id", appID)
	}
}

func (d *Dispatcher) dispatchToChannel(ctx context.Context, q *db.Queries, ch db.NotificationChannel, appID, agentLogID uuid.UUID, p Payload) {
	logEntry, err := q.InsertNotificationLog(ctx, db.InsertNotificationLogParams{
		AppID:      appID,
		ChannelID:  ch.ID,
		AgentLogID: &agentLogID,
		Severity:   p.Severity,
		Summary:    p.Summary,
		Status:     "pending",
	})
	if err != nil {
		slog.Warn("notification: failed to insert log entry", "err", err, "channel", ch.Name)
		return
	}

	sender, err := NewChannel(ch.Type, ch.Config, d.config)
	if err != nil {
		slog.Warn("notification: failed to create channel", "err", err, "channel", ch.Name)
		d.markFailed(ctx, q, logEntry.ID, err.Error())
		return
	}

	if err := sender.Send(ctx, p); err != nil {
		slog.Warn("notification: send failed, retrying once", "channel", ch.Name, "err", err)
		if retryErr := sender.Send(ctx, p); retryErr != nil {
			slog.Warn("notification: retry failed", "channel", ch.Name, "err", retryErr)
			d.markFailed(ctx, q, logEntry.ID, retryErr.Error())
			metrics.NotificationsTotal.WithLabelValues("failed", ch.Type).Inc()
			return
		}
	}

	d.markSent(ctx, q, logEntry.ID)
	metrics.NotificationsTotal.WithLabelValues("sent", ch.Type).Inc()
}

func (d *Dispatcher) markSent(ctx context.Context, q *db.Queries, id uuid.UUID) {
	if err := q.UpdateNotificationLogStatus(ctx, db.UpdateNotificationLogStatusParams{
		ID:     id,
		Status: "sent",
	}); err != nil {
		slog.Warn("notification: failed to mark as sent", "err", err, "id", id)
	}
}

func (d *Dispatcher) markFailed(ctx context.Context, q *db.Queries, id uuid.UUID, errMsg string) {
	if err := q.UpdateNotificationLogStatus(ctx, db.UpdateNotificationLogStatusParams{
		ID:           id,
		Status:       "failed",
		ErrorMessage: pgtype.Text{String: errMsg, Valid: true},
	}); err != nil {
		slog.Warn("notification: failed to mark as failed", "err", err, "id", id)
	}
}

func severityRank(s string) int {
	switch s {
	case "critical":
		return 4
	case "error":
		return 3
	case "warning":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}
