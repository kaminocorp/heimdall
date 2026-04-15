package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// unifiedLogEntry is the response shape for the combined agent log feed.
type unifiedLogEntry struct {
	ID           string          `json:"id"`
	Source       string          `json:"source"`        // "raw" or "agent"
	Timestamp    string          `json:"timestamp"`     // RFC3339
	Severity     *string         `json:"severity"`      // null when absent
	SourceType   string          `json:"source_type"`   // "webhook" for raw; "tool_call"/"tool_result"/"observation" for agent
	ConnectionID *string         `json:"connection_id"` // null for agent entries
	Summary      string          `json:"summary"`
	Detail       json.RawMessage `json:"detail"` // payload for raw; detail for agent
}

type paginatedLogsResponse struct {
	Data   []unifiedLogEntry `json:"data"`
	Total  int64             `json:"total"`
	Limit  int32             `json:"limit"`
	Offset int32             `json:"offset"`
}

func (s *Server) ListLogs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	queries, _, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	// Parse pagination params.
	limit := int32(50)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	if limit > 200 {
		limit = 200
	}

	offset := int32(0)
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	// Cap offset to prevent O(offset) memory usage on source=all merges.
	const maxOffset int32 = 10000
	if offset > maxOffset {
		offset = maxOffset
	}

	severity := r.URL.Query().Get("severity")
	connectionID := r.URL.Query().Get("connection_id")
	source := r.URL.Query().Get("source")
	if source == "" {
		source = "all"
	}

	// Parse optional app_id for per-app filtering.
	var appID uuid.UUID
	hasAppID := false
	if v := r.URL.Query().Get("app_id"); v != "" {
		parsed, parseErr := uuid.Parse(v)
		if parseErr != nil {
			jsonError(w, "invalid app_id", http.StatusBadRequest)
			return
		}
		// Validate ownership: the app must belong to the user's org.
		if _, authErr := s.Queries.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
			AppID:  parsed,
			UserID: userID,
		}); authErr != nil {
			jsonError(w, "application not found", http.StatusNotFound)
			return
		}
		appID = parsed
		hasAppID = true
	}

	var unified []unifiedLogEntry
	var total int64

	// When filtering by connection_id, only raw logs make sense (agent entries have no connection).
	if connectionID != "" {
		source = "raw"
	}

	fetchRaw := source == "all" || source == "raw"
	fetchAgent := source == "all" || source == "agent"

	// When fetching from both sources ("all"), we need to merge results correctly.
	// We fetch (offset + limit) rows from each source so we have enough to merge,
	// then sort the combined set by timestamp and apply offset/limit to the merged result.
	fetchLimit := limit
	fetchOffset := offset
	if source == "all" {
		fetchLimit = offset + limit // fetch enough rows to cover the merged window
		fetchOffset = 0             // start from the beginning; we paginate the merged set
	}

	// Fetch raw logs.
	if fetchRaw {
		var rawLogs []db.LogBuffer
		var err error

		switch {
		case connectionID != "":
			connID, parseErr := uuid.Parse(connectionID)
			if parseErr != nil {
				jsonError(w, "invalid connection_id", http.StatusBadRequest)
				return
			}
			rawLogs, err = queries.ListLogsByUserAndConnection(r.Context(), db.ListLogsByUserAndConnectionParams{
				UserID:       userID,
				ConnectionID: connID,
				Limit:        fetchLimit,
				Offset:       fetchOffset,
			})
		case hasAppID && severity != "":
			rawLogs, err = queries.ListLogsByAppAndSeverity(r.Context(), db.ListLogsByAppAndSeverityParams{
				UserID:   userID,
				AppID:    appID,
				Severity: pgtype.Text{String: severity, Valid: true},
				Limit:    fetchLimit,
				Offset:   fetchOffset,
			})
		case hasAppID:
			rawLogs, err = queries.ListLogsByApp(r.Context(), db.ListLogsByAppParams{
				UserID: userID,
				AppID:  appID,
				Limit:  fetchLimit,
				Offset: fetchOffset,
			})
		case severity != "":
			rawLogs, err = queries.ListLogsByUserAndSeverity(r.Context(), db.ListLogsByUserAndSeverityParams{
				UserID:   userID,
				Severity: pgtype.Text{String: severity, Valid: true},
				Limit:    fetchLimit,
				Offset:   fetchOffset,
			})
		default:
			rawLogs, err = queries.ListLogsByUser(r.Context(), db.ListLogsByUserParams{
				UserID: userID,
				Limit:  fetchLimit,
				Offset: fetchOffset,
			})
		}
		if err != nil {
			jsonServerError(w, "failed to list logs", err)
			return
		}

		// Use a filtered count that matches the active query so pagination totals are accurate.
		var rawCount int64
		switch {
		case connectionID != "":
			connID, _ := uuid.Parse(connectionID) // already validated above
			rawCount, err = queries.CountLogsByUserAndConnection(r.Context(), db.CountLogsByUserAndConnectionParams{
				UserID:       userID,
				ConnectionID: connID,
			})
		case hasAppID && severity != "":
			rawCount, err = queries.CountLogsByAppAndSeverity(r.Context(), db.CountLogsByAppAndSeverityParams{
				UserID:   userID,
				AppID:    appID,
				Severity: pgtype.Text{String: severity, Valid: true},
			})
		case hasAppID:
			rawCount, err = queries.CountLogsByApp(r.Context(), db.CountLogsByAppParams{
				UserID: userID,
				AppID:  appID,
			})
		case severity != "":
			rawCount, err = queries.CountLogsByUserAndSeverity(r.Context(), db.CountLogsByUserAndSeverityParams{
				UserID:   userID,
				Severity: pgtype.Text{String: severity, Valid: true},
			})
		default:
			rawCount, err = queries.CountLogsByUser(r.Context(), userID)
		}
		if err != nil {
			jsonServerError(w, "failed to count logs", err)
			return
		}
		total += rawCount

		for _, l := range rawLogs {
			unified = append(unified, rawToUnified(l))
		}
	}

	// Fetch agent log entries.
	if fetchAgent {
		var agentLogs []db.AgentLog
		var agentCount int64
		var err error

		if hasAppID {
			agentLogs, err = queries.ListAgentLogByApp(r.Context(), db.ListAgentLogByAppParams{
				UserID: userID,
				AppID:  appID,
				Limit:  fetchLimit,
				Offset: fetchOffset,
			})
			if err != nil {
				jsonServerError(w, "failed to list agent logs", err)
				return
			}
			agentCount, err = queries.CountAgentLogByApp(r.Context(), db.CountAgentLogByAppParams{
				UserID: userID,
				AppID:  appID,
			})
		} else {
			agentLogs, err = queries.ListAgentLogByUser(r.Context(), db.ListAgentLogByUserParams{
				UserID: userID,
				Limit:  fetchLimit,
				Offset: fetchOffset,
			})
			if err != nil {
				jsonServerError(w, "failed to list agent logs", err)
				return
			}
			agentCount, err = queries.CountAgentLogByUser(r.Context(), userID)
		}
		if err != nil {
			jsonServerError(w, "failed to count agent logs", err)
			return
		}
		total += agentCount

		for _, l := range agentLogs {
			unified = append(unified, agentToUnified(l))
		}
	}

	// When fetching from both sources, sort the merged set by timestamp
	// descending, then apply the original offset and limit. The reported
	// total is the sum of both source counts (the true dataset size). Deep
	// pages may return fewer than `limit` items if the merge window
	// (offset+limit per source) is exhausted — an acceptable approximation
	// that avoids an expensive UNION count query.
	if source == "all" && len(unified) > 0 {
		sort.Slice(unified, func(i, j int) bool {
			return unified[i].Timestamp > unified[j].Timestamp
		})
		// Apply offset.
		if int32(len(unified)) > offset {
			unified = unified[offset:]
		} else {
			unified = unified[:0]
		}
		// Apply limit.
		if int32(len(unified)) > limit {
			unified = unified[:limit]
		}
	}

	// Ensure Data is always an empty array in JSON, never null.
	if unified == nil {
		unified = []unifiedLogEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginatedLogsResponse{
		Data:   unified,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// rawToUnified converts a log_buffer row to the unified response shape.
func rawToUnified(l db.LogBuffer) unifiedLogEntry {
	connID := l.ConnectionID.String()

	var sev *string
	if l.Severity.Valid {
		sev = &l.Severity.String
	}

	// Extract summary from payload.message if available.
	summary := ""
	var parsed map[string]any
	if err := json.Unmarshal(l.Payload, &parsed); err == nil {
		if msg, ok := parsed["message"]; ok {
			summary, _ = msg.(string)
		}
	}
	if summary == "" {
		summary = string(l.Payload)
		if len(summary) > 200 {
			summary = summary[:200] + "..."
		}
	}

	return unifiedLogEntry{
		ID:           l.ID.String(),
		Source:       "raw",
		Timestamp:    l.IngestedAt.Format("2006-01-02T15:04:05Z07:00"),
		Severity:     sev,
		SourceType:   l.SourceType,
		ConnectionID: &connID,
		Summary:      summary,
		Detail:       l.Payload,
	}
}

// agentToUnified converts an agent_log row to the unified response shape.
func agentToUnified(l db.AgentLog) unifiedLogEntry {
	var sev *string
	if l.Severity.Valid {
		sev = &l.Severity.String
	}

	var detail json.RawMessage
	if l.Detail != nil {
		detail = l.Detail
	}

	return unifiedLogEntry{
		ID:           l.ID.String(),
		Source:       "agent",
		Timestamp:    l.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Severity:     sev,
		SourceType:   l.EntryType,
		ConnectionID: nil,
		Summary:      l.Summary,
		Detail:       detail,
	}
}
