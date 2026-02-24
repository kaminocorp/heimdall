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
func (a *Agent) EmitLog(ctx context.Context, userID uuid.UUID, conversationID *uuid.UUID, entryType, summary string, detail map[string]any) {
	var detailBytes []byte
	if detail != nil {
		var err error
		detailBytes, err = json.Marshal(detail)
		if err != nil {
			slog.Warn("agent emit: failed to marshal detail", "err", err)
			return
		}
	}

	var convID pgtype.UUID
	if conversationID != nil {
		convID = pgtype.UUID{Bytes: *conversationID, Valid: true}
	}

	_, err := a.queries.InsertAgentLog(ctx, db.InsertAgentLogParams{
		UserID:         userID,
		EntryType:      entryType,
		Summary:        summary,
		Detail:         detailBytes,
		ConversationID: convID,
	})
	if err != nil {
		slog.Warn("agent emit: failed to insert agent log", "err", err, "entry_type", entryType)
	}
}
