package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// jsonErrorWithCode is like jsonError but attaches a stable machine-readable
// `code` field alongside the human-readable message. Use this for guarded
// failures the frontend needs to distinguish from generic errors — e.g. the
// last-app delete guard, where the UI renders a specific helper text rather
// than a generic toast.
func jsonErrorWithCode(w http.ResponseWriter, message, code string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
		"code":  code,
	})
}

// isValidEmail performs a basic format check using Go's net/mail parser.
func isValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	// Reject display names like "Name <email@x.com>" — require bare address.
	return addr.Address == email && strings.Contains(email, ".")
}

// jsonServerError logs the internal error for debugging, then sends a generic message to the client.
func jsonServerError(w http.ResponseWriter, message string, err error) {
	slog.Error(message, "error", err)
	jsonError(w, message, http.StatusInternalServerError)
}

// emitActivity is a fire-and-forget activity-feed write that opens its
// own short UserQueries scope. Replaces the pre-Phase-2 pattern of
// `s.Agent.EmitLog(...)` calls in handlers, which used the agent's
// shared (cron-pool) queries handle and bypassed RLS via the postgres
// owner. Every per-tenant write now runs under app.current_user_id ==
// userID so post-Phase-7 RLS policies evaluate correctly.
//
// pools may be nil in struct-literal test environments that don't wire
// it; we treat that the same as Agent==nil pre-refactor (silently skip).
func emitActivity(ctx context.Context, pools *db.Pools, userID uuid.UUID, entryType, summary string, detail map[string]any) {
	if pools == nil {
		return
	}
	if err := pools.WithUserQueries(ctx, userID, func(q *db.Queries) error {
		agent.EmitLog(ctx, q, userID, nil, nil, entryType, summary, detail)
		return nil
	}); err != nil {
		slog.Warn("emitActivity: UserQueries failed", "err", err, "entry_type", entryType, "user_id", userID)
	}
}
