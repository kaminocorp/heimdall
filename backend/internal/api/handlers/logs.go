package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

type paginatedLogsResponse struct {
	Data   []db.LogBuffer `json:"data"`
	Total  int64          `json:"total"`
	Limit  int32          `json:"limit"`
	Offset int32          `json:"offset"`
}

func (s *Server) ListLogs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

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

	severity := r.URL.Query().Get("severity")
	connectionID := r.URL.Query().Get("connection_id")

	var logs []db.LogBuffer
	var err error

	switch {
	case connectionID != "":
		connID, parseErr := uuid.Parse(connectionID)
		if parseErr != nil {
			jsonError(w, "invalid connection_id", http.StatusBadRequest)
			return
		}
		logs, err = s.Queries.ListLogsByUserAndConnection(r.Context(), db.ListLogsByUserAndConnectionParams{
			UserID:       userID,
			ConnectionID: connID,
			Limit:        limit,
			Offset:       offset,
		})
	case severity != "":
		logs, err = s.Queries.ListLogsByUserAndSeverity(r.Context(), db.ListLogsByUserAndSeverityParams{
			UserID:   userID,
			Severity: pgtype.Text{String: severity, Valid: true},
			Limit:    limit,
			Offset:   offset,
		})
	default:
		logs, err = s.Queries.ListLogsByUser(r.Context(), db.ListLogsByUserParams{
			UserID: userID,
			Limit:  limit,
			Offset: offset,
		})
	}

	if err != nil {
		jsonError(w, "failed to list logs", http.StatusInternalServerError)
		return
	}

	total, err := s.Queries.CountLogsByUser(r.Context(), userID)
	if err != nil {
		jsonError(w, "failed to count logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginatedLogsResponse{
		Data:   logs,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}
