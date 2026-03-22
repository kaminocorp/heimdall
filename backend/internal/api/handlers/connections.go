package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/connectors/database"
	"github.com/hejijunhao/heimdall/backend/internal/connectors/logs"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func (s *Server) ListConnections(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	queries, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	connections, err := queries.ListConnectionsByUser(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "failed to list connections", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}

func (s *Server) GetConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	queries, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	conn, err := queries.GetConnectionByUser(r.Context(), db.GetConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conn)
}

type createConnectionRequest struct {
	AppID     string          `json:"app_id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Direction string          `json:"direction"`
	Config    json.RawMessage `json:"config"`
}

func (s *Server) CreateConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	var req createConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Type == "" || req.AppID == "" {
		jsonError(w, "app_id, name and type are required", http.StatusBadRequest)
		return
	}
	if !isValidConnectionType(req.Type) {
		jsonError(w, "invalid connection type", http.StatusBadRequest)
		return
	}
	if req.Direction != "" && !isValidDirection(req.Direction) {
		jsonError(w, "invalid direction", http.StatusBadRequest)
		return
	}

	appID, err := uuid.Parse(req.AppID)
	if err != nil {
		jsonError(w, "invalid app_id", http.StatusBadRequest)
		return
	}

	// Verify the app belongs to the authenticated user's org.
	_, err = s.Queries.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
		AppID:  appID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "application not found", http.StatusNotFound)
		return
	}

	direction := req.Direction
	if direction == "" {
		direction = "one_way"
	}
	config := req.Config
	if config == nil {
		config = json.RawMessage(`{}`)
	}

	// Validate Supabase config eagerly — before DB insert — to prevent
	// creating orphaned connections that can't start polling.
	if req.Type == "supabase" {
		if _, err := logs.NewSupabase(config, uuid.Nil, uuid.Nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid Supabase config: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Auto-generate a webhook token for webhook_logs connections.
	if req.Type == "webhook_logs" {
		var cfgMap map[string]interface{}
		if err := json.Unmarshal(config, &cfgMap); err != nil {
			jsonError(w, "invalid config JSON", http.StatusBadRequest)
			return
		}
		if cfgMap == nil {
			cfgMap = make(map[string]interface{})
		}
		if _, ok := cfgMap["webhook_token"]; !ok {
			b := make([]byte, 32)
			if _, err := rand.Read(b); err != nil {
				jsonServerError(w, "failed to generate webhook token", err)
				return
			}
			cfgMap["webhook_token"] = hex.EncodeToString(b)
			var err error
			config, err = json.Marshal(cfgMap)
			if err != nil {
				jsonServerError(w, "failed to encode config", err)
				return
			}
		}
	}

	queries, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	conn, err := queries.CreateConnection(r.Context(), db.CreateConnectionParams{
		UserID:    userID,
		AppID:     appID,
		Name:      req.Name,
		Type:      req.Type,
		Direction: direction,
		Config:    config,
		Status:    "inactive",
	})
	if err != nil {
		jsonServerError(w, "failed to create connection", err)
		return
	}

	// Start polling goroutine for Supabase connections.
	// Config was already validated above, so NewSupabase should not fail here.
	if req.Type == "supabase" {
		sb, err := logs.NewSupabase(config, conn.ID, userID)
		if err != nil {
			slog.Error("supabase poller init failed (unexpected)", "connection_id", conn.ID, "err", err)
		} else {
			cfg := sb.ParsedConfig()
			interval := time.Duration(cfg.PollIntervalSecs) * time.Second
			s.Poller.Start(sb, conn.ID, interval)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conn)
}

type updateConnectionRequest struct {
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Direction string          `json:"direction"`
	Config    json.RawMessage `json:"config"`
	Status    string          `json:"status"`
}

func (s *Server) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	var req updateConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Type == "" {
		jsonError(w, "name and type are required", http.StatusBadRequest)
		return
	}
	if !isValidConnectionType(req.Type) {
		jsonError(w, "invalid connection type", http.StatusBadRequest)
		return
	}
	if req.Direction != "" && !isValidDirection(req.Direction) {
		jsonError(w, "invalid direction", http.StatusBadRequest)
		return
	}
	if req.Status != "" && !isValidConnectionStatus(req.Status) {
		jsonError(w, "invalid status", http.StatusBadRequest)
		return
	}

	direction := req.Direction
	if direction == "" {
		direction = "one_way"
	}
	config := req.Config
	if config == nil {
		config = json.RawMessage(`{}`)
	}
	status := req.Status
	if status == "" {
		status = "inactive"
	}

	queries, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	conn, err := queries.UpdateConnection(r.Context(), db.UpdateConnectionParams{
		ID:        connID,
		Name:      req.Name,
		Type:      req.Type,
		Direction: direction,
		Config:    config,
		Status:    status,
		UserID:    userID,
	})
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}

	// Restart poller if this is a Supabase connection so it picks up config changes.
	// Stop first (handles the case where type changed away from supabase too).
	s.Poller.Stop(connID)
	if req.Type == "supabase" {
		sb, err := logs.NewSupabase(config, conn.ID, userID)
		if err != nil {
			slog.Error("supabase poller init failed", "connection_id", conn.ID, "err", err)
		} else {
			cfg := sb.ParsedConfig()
			interval := time.Duration(cfg.PollIntervalSecs) * time.Second
			s.Poller.Start(sb, conn.ID, interval)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conn)
}

func (s *Server) TestConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	queries, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	conn, err := queries.GetConnectionByUser(r.Context(), db.GetConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}

	type testResult struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	var result testResult

	switch conn.Type {
	case "postgres":
		pg, err := database.New(conn.Config)
		if err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Failed to initialize database connector: %v", err)}
			break
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := pg.Connect(ctx); err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Failed to connect to database: %v", err)}
		} else {
			defer pg.Close(ctx)
			result = testResult{Success: true, Message: "Connection established"}
		}
	case "supabase":
		sb, err := logs.NewSupabase(conn.Config, conn.ID, userID)
		if err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Invalid config: %v", err)}
			break
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if err := sb.Connect(ctx); err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Failed to connect to Supabase API: %v", err)}
		} else {
			defer sb.Close()
			result = testResult{Success: true, Message: "Connected to Supabase Management API"}
		}
	case "github":
		if s.GitHub != nil {
			success, msg := s.TestGitHubConnection(r.Context(), conn.Config)
			result = testResult{Success: success, Message: msg}
		} else {
			result = testResult{Success: false, Message: "GitHub App not configured on server"}
		}
	default:
		// webhook_logs, syslog — no remote target to test, auto-pass.
		result = testResult{Success: true, Message: "Connection established"}
	}

	newStatus := "active"
	if !result.Success {
		newStatus = "error"
	}
	if err := queries.UpdateConnectionStatus(r.Context(), db.UpdateConnectionStatusParams{
		ID:     connID,
		Status: newStatus,
	}); err != nil {
		slog.Error("failed to update connection status", "connection_id", connID, "err", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// --- Input validation helpers ---

var validConnectionTypes = map[string]bool{
	"webhook_logs": true,
	"postgres":     true,
	"supabase":     true,
	"github":       true,
	"syslog":       true,
}

func isValidConnectionType(t string) bool  { return validConnectionTypes[t] }
func isValidDirection(d string) bool        { return d == "one_way" || d == "two_way" }
func isValidConnectionStatus(s string) bool { return s == "active" || s == "inactive" || s == "error" }

func (s *Server) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	queries, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	// Stop any active poller before deleting.
	s.Poller.Stop(connID)

	err = queries.DeleteConnectionByUser(r.Context(), db.DeleteConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonServerError(w, "failed to delete connection", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
