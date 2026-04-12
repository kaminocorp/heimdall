package handlers

import (
	"context"
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
			defer pg.Close()
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
