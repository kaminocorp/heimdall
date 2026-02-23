package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func (a *Agent) toolSearchLogs(ctx context.Context, userID uuid.UUID, input map[string]any) (string, error) {
	// Parse limit (default 20, max 200).
	limit := int32(20)
	if v, ok := input["limit"].(float64); ok {
		limit = int32(v)
	}
	if limit > 200 {
		limit = 200
	}
	if limit < 1 {
		limit = 20
	}

	// Parse optional severity filter.
	severity, _ := input["severity"].(string)

	var logs []db.LogBuffer
	var err error

	if severity != "" {
		logs, err = a.queries.ListLogsByUserAndSeverity(ctx, db.ListLogsByUserAndSeverityParams{
			UserID:   userID,
			Severity: pgtype.Text{String: severity, Valid: true},
			Limit:    limit,
			Offset:   0,
		})
	} else {
		logs, err = a.queries.ListLogsByUser(ctx, db.ListLogsByUserParams{
			UserID: userID,
			Limit:  limit,
			Offset: 0,
		})
	}
	if err != nil {
		return "", fmt.Errorf("search_logs: %w", err)
	}

	if len(logs) == 0 {
		return `{"results": [], "count": 0, "message": "No log entries found"}`, nil
	}

	type logEntry struct {
		ID           string          `json:"id"`
		ConnectionID string          `json:"connection_id"`
		Severity     string          `json:"severity"`
		Payload      json.RawMessage `json:"payload"`
		IngestedAt   string          `json:"ingested_at"`
	}

	entries := make([]logEntry, len(logs))
	for i, l := range logs {
		sev := ""
		if l.Severity.Valid {
			sev = l.Severity.String
		}
		entries[i] = logEntry{
			ID:           l.ID.String(),
			ConnectionID: l.ConnectionID.String(),
			Severity:     sev,
			Payload:      l.Payload,
			IngestedAt:   l.IngestedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	result, err := json.Marshal(map[string]any{
		"results": entries,
		"count":   len(entries),
	})
	if err != nil {
		return "", fmt.Errorf("search_logs: marshal: %w", err)
	}
	return string(result), nil
}
