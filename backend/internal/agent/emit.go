package agent

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// EmitLog writes an entry to the agent_log table.
// Fire-and-forget: errors are logged but never propagated so agent work is not degraded by logging failures.
func (a *Agent) EmitLog(ctx context.Context, userID uuid.UUID, appID *uuid.UUID, conversationID *uuid.UUID, entryType, summary string, detail map[string]any) uuid.UUID {
	return a.emitLog(ctx, userID, appID, conversationID, entryType, summary, detail, "")
}

// EmitLogWithSeverity writes an entry to the agent_log table with an explicit severity.
func (a *Agent) EmitLogWithSeverity(ctx context.Context, userID uuid.UUID, appID *uuid.UUID, conversationID *uuid.UUID, entryType, summary string, detail map[string]any, severity string) uuid.UUID {
	return a.emitLog(ctx, userID, appID, conversationID, entryType, summary, detail, severity)
}

func (a *Agent) emitLog(ctx context.Context, userID uuid.UUID, appID *uuid.UUID, conversationID *uuid.UUID, entryType, summary string, detail map[string]any, severity string) uuid.UUID {
	var detailBytes []byte
	if detail != nil {
		var err error
		detailBytes, err = json.Marshal(detail)
		if err != nil {
			slog.Warn("agent emit: failed to marshal detail, writing log without detail", "err", err)
			detailBytes = nil
		}
	}

	var sev pgtype.Text
	if severity != "" {
		sev = pgtype.Text{String: severity, Valid: true}
	}

	var appUUID uuid.UUID
	if appID != nil {
		appUUID = *appID
	}

	row, err := a.queries.InsertAgentLog(ctx, db.InsertAgentLogParams{
		UserID:         userID,
		EntryType:      entryType,
		Summary:        summary,
		Detail:         detailBytes,
		Severity:       sev,
		ConversationID: conversationID,
		AppID:          appUUID,
	})
	if err != nil {
		slog.Warn("agent emit: failed to insert agent log", "err", err, "entry_type", entryType)
		return uuid.Nil
	}
	return row.ID
}
