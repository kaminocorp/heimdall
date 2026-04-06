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

	// Validate API poller configs eagerly.
	if req.Type == "flyio" {
		if _, err := logs.NewFlyio(config, uuid.Nil, uuid.Nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid Fly.io config: %v", err), http.StatusBadRequest)
			return
		}
	}
	if req.Type == "vercel" {
		if _, err := logs.NewVercel(config, uuid.Nil, uuid.Nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid Vercel config: %v", err), http.StatusBadRequest)
			return
		}
	}
	if req.Type == "railway" {
		if _, err := logs.NewRailway(config, uuid.Nil, uuid.Nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid Railway config: %v", err), http.StatusBadRequest)
			return
		}
	}
	if req.Type == "mongodb" {
		if _, err := logs.NewMongoDB(config, uuid.Nil, uuid.Nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid MongoDB Atlas config: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Validate syslog config eagerly.
	if req.Type == "syslog" {
		if _, err := logs.NewSyslog(config, uuid.Nil, uuid.Nil, nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid syslog config: %v", err), http.StatusBadRequest)
			return
		}
		// Inject server-level TLS cert/key if not provided in config and available globally.
		if s.Config.SyslogTLSCert != "" && s.Config.SyslogTLSKey != "" {
			var cfgMap map[string]interface{}
			if err := json.Unmarshal(config, &cfgMap); err == nil {
				if cfgMap == nil {
					cfgMap = make(map[string]interface{})
				}
				if _, ok := cfgMap["tls_cert"]; !ok {
					cfgMap["tls_cert"] = s.Config.SyslogTLSCert
					cfgMap["tls_key"] = s.Config.SyslogTLSKey
					config, _ = json.Marshal(cfgMap)
				}
			}
		}
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

	// Start polling goroutines for poll-based connectors.
	startPoller(s, req.Type, config, conn.ID, userID)

	// Start listener for syslog connections.
	if req.Type == "syslog" {
		sl, err := logs.NewSyslog(config, conn.ID, userID, s.Queries)
		if err != nil {
			slog.Error("syslog listener init failed", "connection_id", conn.ID, "err", err)
			conn.Status = "error"
			_ = queries.UpdateConnectionStatus(r.Context(), db.UpdateConnectionStatusParams{ID: conn.ID, Status: "error"})
		} else if err := s.Listener.Start(r.Context(), sl, conn.ID); err != nil {
			slog.Error("syslog listener start failed", "connection_id", conn.ID, "err", err)
			conn.Status = "error"
			_ = queries.UpdateConnectionStatus(r.Context(), db.UpdateConnectionStatusParams{ID: conn.ID, Status: "error"})
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

	// --- Config validation (mirrors CreateConnection) ---

	if req.Type == "supabase" {
		if _, err := logs.NewSupabase(config, uuid.Nil, uuid.Nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid Supabase config: %v", err), http.StatusBadRequest)
			return
		}
	}
	if req.Type == "flyio" {
		if _, err := logs.NewFlyio(config, uuid.Nil, uuid.Nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid Fly.io config: %v", err), http.StatusBadRequest)
			return
		}
	}
	if req.Type == "vercel" {
		if _, err := logs.NewVercel(config, uuid.Nil, uuid.Nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid Vercel config: %v", err), http.StatusBadRequest)
			return
		}
	}
	if req.Type == "railway" {
		if _, err := logs.NewRailway(config, uuid.Nil, uuid.Nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid Railway config: %v", err), http.StatusBadRequest)
			return
		}
	}
	if req.Type == "mongodb" {
		if _, err := logs.NewMongoDB(config, uuid.Nil, uuid.Nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid MongoDB Atlas config: %v", err), http.StatusBadRequest)
			return
		}
	}
	if req.Type == "syslog" {
		if _, err := logs.NewSyslog(config, uuid.Nil, uuid.Nil, nil); err != nil {
			jsonError(w, fmt.Sprintf("Invalid syslog config: %v", err), http.StatusBadRequest)
			return
		}
		// Inject server-level TLS cert/key if not provided in config and available globally.
		if s.Config.SyslogTLSCert != "" && s.Config.SyslogTLSKey != "" {
			var cfgMap map[string]interface{}
			if err := json.Unmarshal(config, &cfgMap); err == nil {
				if cfgMap == nil {
					cfgMap = make(map[string]interface{})
				}
				if _, ok := cfgMap["tls_cert"]; !ok {
					cfgMap["tls_cert"] = s.Config.SyslogTLSCert
					cfgMap["tls_key"] = s.Config.SyslogTLSKey
					config, _ = json.Marshal(cfgMap)
				}
			}
		}
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
			// Fetch existing connection to preserve the token.
			existing, err := s.Queries.GetConnectionByUser(r.Context(), db.GetConnectionByUserParams{
				ID:     connID,
				UserID: userID,
			})
			if err != nil {
				jsonError(w, "connection not found", http.StatusNotFound)
				return
			}
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

	// Restart poller/listener if config changed.
	s.Poller.Stop(connID)
	s.Listener.Stop(connID)

	startPoller(s, req.Type, config, conn.ID, userID)

	if req.Type == "syslog" {
		sl, err := logs.NewSyslog(config, conn.ID, userID, s.Queries)
		if err != nil {
			slog.Error("syslog listener init failed", "connection_id", conn.ID, "err", err)
			conn.Status = "error"
			_ = queries.UpdateConnectionStatus(r.Context(), db.UpdateConnectionStatusParams{ID: conn.ID, Status: "error"})
		} else if err := s.Listener.Start(r.Context(), sl, conn.ID); err != nil {
			slog.Error("syslog listener start failed", "connection_id", conn.ID, "err", err)
			conn.Status = "error"
			_ = queries.UpdateConnectionStatus(r.Context(), db.UpdateConnectionStatusParams{ID: conn.ID, Status: "error"})
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
	case "flyio":
		f, err := logs.NewFlyio(conn.Config, conn.ID, userID)
		if err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Invalid config: %v", err)}
			break
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if err := f.Connect(ctx); err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Failed to connect to Fly.io API: %v", err)}
		} else {
			defer f.Close()
			result = testResult{Success: true, Message: "Connected to Fly.io Machines API"}
		}
	case "vercel":
		v, err := logs.NewVercel(conn.Config, conn.ID, userID)
		if err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Invalid config: %v", err)}
			break
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if err := v.Connect(ctx); err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Failed to connect to Vercel API: %v", err)}
		} else {
			defer v.Close()
			result = testResult{Success: true, Message: "Connected to Vercel API"}
		}
	case "railway":
		rl, err := logs.NewRailway(conn.Config, conn.ID, userID)
		if err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Invalid config: %v", err)}
			break
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if err := rl.Connect(ctx); err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Failed to connect to Railway API: %v", err)}
		} else {
			defer rl.Close()
			result = testResult{Success: true, Message: "Connected to Railway GraphQL API"}
		}
	case "mongodb":
		mg, err := logs.NewMongoDB(conn.Config, conn.ID, userID)
		if err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Invalid config: %v", err)}
			break
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if err := mg.Connect(ctx); err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Failed to connect to MongoDB Atlas API: %v", err)}
		} else {
			defer mg.Close()
			result = testResult{Success: true, Message: "Connected to MongoDB Atlas API"}
		}
	case "syslog":
		sl, err := logs.NewSyslog(conn.Config, conn.ID, userID, s.Queries)
		if err != nil {
			slog.Error("connection test failed", "connection_id", connID, "err", err)
			result = testResult{Success: false, Message: fmt.Sprintf("Invalid config: %v", err)}
			break
		}
		cfg := sl.ParsedConfig()
		result = testResult{
			Success: true,
			Message: fmt.Sprintf("Syslog listener configured on port %d (%s)", cfg.Port, cfg.Protocol),
		}
	default:
		// webhook_logs, otlp — no remote target to test, auto-pass.
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
	"otlp":         true,
	"flyio":        true,
	"vercel":       true,
	"railway":      true,
	"mongodb":      true,
}

func isValidConnectionType(t string) bool  { return validConnectionTypes[t] }
func isValidDirection(d string) bool        { return d == "one_way" || d == "two_way" }
func isValidConnectionStatus(s string) bool { return s == "active" || s == "inactive" || s == "error" }

// startPoller initializes and starts a polling goroutine for poll-based connector types.
func startPoller(s *Server, connType string, config json.RawMessage, connID, userID uuid.UUID) {
	switch connType {
	case "supabase":
		sb, err := logs.NewSupabase(config, connID, userID)
		if err != nil {
			slog.Error("supabase poller init failed", "connection_id", connID, "err", err)
			return
		}
		s.Poller.Start(sb, connID, time.Duration(sb.ParsedConfig().PollIntervalSecs)*time.Second)
	case "flyio":
		f, err := logs.NewFlyio(config, connID, userID)
		if err != nil {
			slog.Error("flyio poller init failed", "connection_id", connID, "err", err)
			return
		}
		s.Poller.Start(f, connID, time.Duration(f.ParsedConfig().PollIntervalSecs)*time.Second)
	case "vercel":
		v, err := logs.NewVercel(config, connID, userID)
		if err != nil {
			slog.Error("vercel poller init failed", "connection_id", connID, "err", err)
			return
		}
		s.Poller.Start(v, connID, time.Duration(v.ParsedConfig().PollIntervalSecs)*time.Second)
	case "railway":
		r, err := logs.NewRailway(config, connID, userID)
		if err != nil {
			slog.Error("railway poller init failed", "connection_id", connID, "err", err)
			return
		}
		s.Poller.Start(r, connID, time.Duration(r.ParsedConfig().PollIntervalSecs)*time.Second)
	case "mongodb":
		m, err := logs.NewMongoDB(config, connID, userID)
		if err != nil {
			slog.Error("mongodb poller init failed", "connection_id", connID, "err", err)
			return
		}
		s.Poller.Start(m, connID, time.Duration(m.ParsedConfig().PollIntervalSecs)*time.Second)
	}
}

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

	w.WriteHeader(http.StatusNoContent)
}
