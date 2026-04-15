package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

type conversationSummary struct {
	ID        uuid.UUID `json:"id"`
	Title     *string   `json:"title"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

func (s *Server) ListConversations(w http.ResponseWriter, r *http.Request) {
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

	limit := int32(50)
	offset := int32(0)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = int32(n)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}

	conversations, err := queries.ListConversationsByUser(r.Context(), db.ListConversationsByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		jsonServerError(w, "failed to list conversations", err)
		return
	}

	// Return summaries without full message payload.
	summaries := make([]conversationSummary, len(conversations))
	for i, c := range conversations {
		var title *string
		if c.Title.Valid {
			title = &c.Title.String
		}
		summaries[i] = conversationSummary{
			ID:        c.ID,
			Title:     title,
			CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

func (s *Server) GetConversation(w http.ResponseWriter, r *http.Request) {
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

	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid conversation id", http.StatusBadRequest)
		return
	}

	conv, err := queries.GetConversationByUser(r.Context(), db.GetConversationByUserParams{
		ID:     convID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "conversation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conv)
}
