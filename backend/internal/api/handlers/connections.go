package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/connectors"
	"github.com/hejijunhao/heimdall/backend/internal/connectors/logs"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func (s *Server) ListConnections(w http.ResponseWriter, r *http.Request) {
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

	queries, _, done, err := s.UserQueries(r.Context(), userID)
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

// createConnectionRequest accepts either an app_id (app-scoped, the Phase 1
// behaviour) or omits it to create an org-scoped connection that's visible to
// every app in the caller's currently active org. The org itself is resolved
// server-side via resolveOrgForUser (X-Org-ID header or user's primary org).
type createConnectionRequest struct {
	AppID     string          `json:"app_id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Direction string          `json:"direction"`
	Config    json.RawMessage `json:"config"`
}

// appIDOrZero returns the AppID from a Connection, or uuid.Nil for org-scoped
// connections. Used when passing the id into connector APIs that always
// expect a concrete UUID — which in practice only run for pull-style connection
// types (postgres, supabase, syslog, flyio polling, etc.) that are exclusively
// app-scoped today.
func appIDOrZero(p *uuid.UUID) uuid.UUID {
	if p == nil {
		return uuid.Nil
	}
	return *p
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

	// Resolve scope. Two valid modes:
	//   - app_id provided → app-scoped connection (Phase 1 behaviour)
	//   - app_id omitted  → org-scoped connection (Phase 2 addition),
	//     shared across every app in the caller's active org
	//
	// Org is always resolved via X-Org-ID header (or primary org fallback)
	// so the server doesn't trust client-supplied org_id. We derive org_id
	// from the application itself when app-scoped, and from
	// resolveOrgForUser when org-scoped.
	var appID *uuid.UUID
	var orgID uuid.UUID
	if req.AppID != "" {
		parsed, err := uuid.Parse(req.AppID)
		if err != nil {
			jsonError(w, "invalid app_id", http.StatusBadRequest)
			return
		}
		appID = &parsed
	} else {
		// Org-scoping only makes sense for connection types that can receive
		// logs from many sources — a single Postgres instance or GitHub org
		// always binds to one Heimdall app. The wizard hides the option for
		// non-eligible types; the backend enforces it defensively.
		if !supportsOrgScope(req.Type) {
			jsonError(w, "this connection type must be app-scoped — include app_id", http.StatusBadRequest)
			return
		}
		org, _, err := s.resolveOrgForUser(r, userID)
		if err != nil {
			jsonError(w, "organization not found", http.StatusNotFound)
			return
		}
		orgID = org.ID
	}

	direction := req.Direction
	if direction == "" {
		direction = "one_way"
	}
	config := req.Config
	if config == nil {
		config = json.RawMessage(`{}`)
	}

	// Validate connector config eagerly — before DB insert — to prevent
	// creating orphaned connections that can't start polling.
	var err error
	config, err = s.validateConnectorConfig(req.Type, config)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Auto-generate a webhook token for webhook_logs and OTLP connections.
	if req.Type == "webhook_logs" || req.Type == "otlp" {
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

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	// App-scoped: verify app belongs to the user's org inside the tx (no
	// TOCTOU gap). The app's org becomes the connection's org.
	if appID != nil {
		app, err := queries.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
			AppID:  *appID,
			UserID: userID,
		})
		if err != nil {
			jsonError(w, "application not found", http.StatusNotFound)
			return
		}
		orgID = app.OrgID
	}

	conn, err := queries.CreateConnection(r.Context(), db.CreateConnectionParams{
		UserID:    userID,
		OrgID:     orgID,
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

	// Start polling goroutines for poll-based connectors. Org-scoped
	// connections (appID == nil) are webhook-only in practice — the pull
	// connectors below require a concrete app context, so StartPoller / the
	// syslog branch simply no-op for them via appIDOrZero == uuid.Nil.
	if err := connectors.StartPoller(s.Poller, req.Type, config, conn.ID, userID, appIDOrZero(conn.AppID)); err != nil {
		slog.Error("poller init failed", "type", req.Type, "connection_id", conn.ID, "err", err)
	}

	// Start listener for syslog connections.
	if req.Type == "syslog" {
		sl, err := logs.NewSyslog(config, conn.ID, userID, appIDOrZero(conn.AppID), s.Pools)
		if err != nil {
			slog.Error("syslog listener init failed", "connection_id", conn.ID, "err", err)
			conn.Status = "error"
			if err := queries.UpdateConnectionStatus(r.Context(), db.UpdateConnectionStatusParams{ID: conn.ID, Status: "error"}); err != nil {
				slog.Error("failed to persist listener error status", "connection_id", conn.ID, "err", err)
			}
		} else if err := s.Listener.Start(r.Context(), sl, conn.ID); err != nil {
			slog.Error("syslog listener start failed", "connection_id", conn.ID, "err", err)
			conn.Status = "error"
			if err := queries.UpdateConnectionStatus(r.Context(), db.UpdateConnectionStatusParams{ID: conn.ID, Status: "error"}); err != nil {
				slog.Error("failed to persist listener error status", "connection_id", conn.ID, "err", err)
			}
		}
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conn)
}

// supportsOrgScope returns true for connection types where it makes sense to
// have one connection feed multiple apps — i.e. the stream carries logs from
// more than one source that the user subsequently partitions per-app via
// source filters. Tight-scoped agent tools (postgres, github, etc.) always
// bind to a single app.
func supportsOrgScope(connType string) bool {
	switch connType {
	case "webhook_logs", "otlp":
		return true
	default:
		return false
	}
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

	// Validate connector config eagerly — before DB update.
	config, err = s.validateConnectorConfig(req.Type, config)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	// Fetch existing connection inside the transaction to avoid TOCTOU gap.
	existing, err := queries.GetConnectionByUser(r.Context(), db.GetConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}

	status := req.Status
	if status == "" {
		// Preserve the existing status when the client omits the field,
		// so a rename-only update doesn't silently deactivate a connection.
		status = existing.Status
	}

	// Preserve webhook token for webhook_logs/otlp connections.
	// If the update payload omits the token, copy it from the existing connection
	// so that active integrations aren't broken by a config update.
	if req.Type == "webhook_logs" || req.Type == "otlp" {
		var cfgMap map[string]interface{}
		if err := json.Unmarshal(config, &cfgMap); err != nil {
			jsonError(w, "invalid config JSON", http.StatusBadRequest)
			return
		}
		if cfgMap == nil {
			cfgMap = make(map[string]interface{})
		}
		if _, ok := cfgMap["webhook_token"]; !ok {
			var oldCfg map[string]interface{}
			if err := json.Unmarshal(existing.Config, &oldCfg); err == nil {
				if tok, ok := oldCfg["webhook_token"]; ok {
					cfgMap["webhook_token"] = tok
				}
			}
			var marshalErr error
			config, marshalErr = json.Marshal(cfgMap)
			if marshalErr != nil {
				jsonServerError(w, "failed to encode config", marshalErr)
				return
			}
		}
	}

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

	// Stop existing poller/listener unconditionally, then only restart
	// if the connection is not paused.
	s.Poller.Stop(connID)
	s.Listener.Stop(connID)

	if status != "paused" {
		if err := connectors.StartPoller(s.Poller, req.Type, config, conn.ID, userID, appIDOrZero(conn.AppID)); err != nil {
			slog.Error("poller init failed", "type", req.Type, "connection_id", conn.ID, "err", err)
		}

		if req.Type == "syslog" {
			sl, err := logs.NewSyslog(config, conn.ID, userID, appIDOrZero(conn.AppID), s.Pools)
			if err != nil {
				slog.Error("syslog listener init failed", "connection_id", conn.ID, "err", err)
				conn.Status = "error"
				if err := queries.UpdateConnectionStatus(r.Context(), db.UpdateConnectionStatusParams{ID: conn.ID, Status: "error"}); err != nil {
					slog.Error("failed to persist listener error status", "connection_id", conn.ID, "err", err)
				}
			} else if err := s.Listener.Start(r.Context(), sl, conn.ID); err != nil {
				slog.Error("syslog listener start failed", "connection_id", conn.ID, "err", err)
				conn.Status = "error"
				if err := queries.UpdateConnectionStatus(r.Context(), db.UpdateConnectionStatusParams{ID: conn.ID, Status: "error"}); err != nil {
					slog.Error("failed to persist listener error status", "connection_id", conn.ID, "err", err)
				}
			}
		}
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conn)
}

func (s *Server) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
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

	// Stop any active poller or listener before deleting.
	s.Poller.Stop(connID)
	s.Listener.Stop(connID)

	err = queries.DeleteConnectionByUser(r.Context(), db.DeleteConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonServerError(w, "failed to delete connection", err)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
