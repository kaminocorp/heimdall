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
type Dispatcher struct {
	queries *db.Queries
	config  *config.Config
}

// NewDispatcher creates a new notification dispatcher.
func NewDispatcher(queries *db.Queries, cfg *config.Config) *Dispatcher {
	return &Dispatcher{queries: queries, config: cfg}
}

// Notify is called by the monitoring loop after emitting an agent_log entry.
// Fire-and-forget: errors are logged, never returned.
func (d *Dispatcher) Notify(ctx context.Context, appID uuid.UUID, agentLogID uuid.UUID, appName, severity, summary, assessment string) {
	prefs, err := d.queries.GetNotificationPreferences(ctx, appID)
	if err != nil || !prefs.Enabled {
		return
	}

	if severityRank(severity) < severityRank(prefs.SeverityThreshold) {
		return
	}

	lastNotif, err := d.queries.GetLastNotificationForApp(ctx, appID)
	if err == nil && time.Since(lastNotif.CreatedAt) < time.Duration(prefs.CooldownMinutes)*time.Minute {
		slog.Debug("notification suppressed by cooldown", "app_id", appID, "last_sent", lastNotif.CreatedAt)
		return
	}

	channels, err := d.queries.ListEnabledChannelsByApp(ctx, appID)
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

	for _, ch := range channels {
		d.dispatchToChannel(ctx, ch, appID, agentLogID, payload)
	}
}

func (d *Dispatcher) dispatchToChannel(ctx context.Context, ch db.NotificationChannel, appID, agentLogID uuid.UUID, p Payload) {
	logEntry, err := d.queries.InsertNotificationLog(ctx, db.InsertNotificationLogParams{
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
		d.markFailed(ctx, logEntry.ID, err.Error())
		return
	}

	if err := sender.Send(ctx, p); err != nil {
		slog.Warn("notification: send failed, retrying once", "channel", ch.Name, "err", err)
		if retryErr := sender.Send(ctx, p); retryErr != nil {
			slog.Warn("notification: retry failed", "channel", ch.Name, "err", retryErr)
			d.markFailed(ctx, logEntry.ID, retryErr.Error())
			metrics.NotificationsTotal.WithLabelValues("failed", ch.Type).Inc()
			return
		}
	}

	d.markSent(ctx, logEntry.ID)
	metrics.NotificationsTotal.WithLabelValues("sent", ch.Type).Inc()
}

func (d *Dispatcher) markSent(ctx context.Context, id uuid.UUID) {
	if err := d.queries.UpdateNotificationLogStatus(ctx, db.UpdateNotificationLogStatusParams{
		ID:     id,
		Status: "sent",
	}); err != nil {
		slog.Warn("notification: failed to mark as sent", "err", err, "id", id)
	}
}

func (d *Dispatcher) markFailed(ctx context.Context, id uuid.UUID, errMsg string) {
	if err := d.queries.UpdateNotificationLogStatus(ctx, db.UpdateNotificationLogStatusParams{
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
