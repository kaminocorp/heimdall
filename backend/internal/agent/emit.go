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
// Fire-and-forget: errors are logged but never propagated so agent work is
// not degraded by logging failures.
//
// Caller-supplied queries: q must be a UserQueries-bound *db.Queries for
// the user the entry attributes to (`userID` parameter and the
// transaction's app.current_user_id GUC must agree). Pre-Phase-2 this
// used the agent's shared queries handle and bypassed RLS via the
// postgres-owner pool; under post-Phase-6 app_user the insert needs the
// owner's GUC to satisfy agent_log's RLS policy.
func EmitLog(ctx context.Context, q *db.Queries, userID uuid.UUID, appID *uuid.UUID, conversationID *uuid.UUID, entryType, summary string, detail map[string]any) uuid.UUID {
	return emitLog(ctx, q, userID, appID, conversationID, entryType, summary, detail, "")
}

// EmitLogWithSeverity writes an entry to the agent_log table with an explicit severity.
func EmitLogWithSeverity(ctx context.Context, q *db.Queries, userID uuid.UUID, appID *uuid.UUID, conversationID *uuid.UUID, entryType, summary string, detail map[string]any, severity string) uuid.UUID {
	return emitLog(ctx, q, userID, appID, conversationID, entryType, summary, detail, severity)
}

func emitLog(ctx context.Context, q *db.Queries, userID uuid.UUID, appID *uuid.UUID, conversationID *uuid.UUID, entryType, summary string, detail map[string]any, severity string) uuid.UUID {
	if q == nil {
		slog.Warn("agent emit: nil queries handle, skipping", "entry_type", entryType)
		return uuid.Nil
	}

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

	row, err := q.InsertAgentLog(ctx, db.InsertAgentLogParams{
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
