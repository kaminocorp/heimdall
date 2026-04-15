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

	// Verify the app belongs to the authenticated user's org inside the
	// transaction to avoid a TOCTOU gap between the auth check and the write.
	_, err = queries.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
		AppID:  appID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "application not found", http.StatusNotFound)
		return
	}

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

	// Start polling goroutines for poll-based connectors.
	if err := connectors.StartPoller(s.Poller, req.Type, config, conn.ID, userID, conn.AppID); err != nil {
		slog.Error("poller init failed", "type", req.Type, "connection_id", conn.ID, "err", err)
	}

	// Start listener for syslog connections.
	if req.Type == "syslog" {
		sl, err := logs.NewSyslog(config, conn.ID, userID, conn.AppID, s.Queries)
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

	// Fetch existing connection to preserve defaults for omitted fields.
	existing, err := s.Queries.GetConnectionByUser(r.Context(), db.GetConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
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
		// Preserve the existing status when the client omits the field,
		// so a rename-only update doesn't silently deactivate a connection.
		status = existing.Status
	}

	// Validate connector config eagerly — before DB update.
	config, err = s.validateConnectorConfig(req.Type, config)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
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
			// Use the already-fetched existing connection to preserve the token.
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

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
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

	// Restart poller/listener if config changed.
	s.Poller.Stop(connID)
	s.Listener.Stop(connID)

	if err := connectors.StartPoller(s.Poller, req.Type, config, conn.ID, userID, conn.AppID); err != nil {
		slog.Error("poller init failed", "type", req.Type, "connection_id", conn.ID, "err", err)
	}

	if req.Type == "syslog" {
		sl, err := logs.NewSyslog(config, conn.ID, userID, conn.AppID, s.Queries)
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
