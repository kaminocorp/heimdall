package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/hejijunhao/heimdall/backend/internal/api/handlers"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// testEnv holds shared resources for a test run.
type testEnv struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
	Server  *handlers.Server
	Router  http.Handler
	UserID  string
}

// testSetup creates a full test environment:
//  1. Connects to the database using DATABASE_URL from the environment.
//  2. Creates a test user (in both auth.users and public.users).
//  3. Builds a Server and a chi Router that bypasses JWT validation
//     by injecting the test user ID into every request context.
//  4. Registers t.Cleanup to delete the test user (cascade deletes all data)
//     and close the pool.
func testSetup(t *testing.T) *testEnv {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		t.Skip("SUPABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)

	queries := db.New(pool)

	// Create a test user. The public.users table has a FK to auth.users,
	// so we insert into auth.users first.
	userID := uuid.New()
	email := fmt.Sprintf("test-%s@heimdall.test", userID.String())

	_, err = pool.Exec(ctx,
		`INSERT INTO auth.users (id, email, instance_id, aud, role, encrypted_password, confirmation_token, created_at, updated_at)
		 VALUES ($1, $2, '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', '', '', now(), now())`,
		userID, email,
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, email)
	require.NoError(t, err)

	// Cleanup: cascade-delete the auth.users row (which cascades to public.users
	// and all user-scoped data), then close the pool.
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM auth.users WHERE id = $1", userID)
		pool.Close()
	})

	cfg := &config.Config{
		Port:        "0",
		DatabaseURL: dbURL,
		SupabaseURL: supabaseURL,
	}

	srv := &handlers.Server{
		Config:  cfg,
		Pool:    pool,
		Queries: queries,
		Agent:   nil, // Agent not needed for handler tests
	}

	// Build a test router that mirrors the production routes from router.go
	// but replaces JWT auth with a middleware that injects the test user ID.
	router := chi.NewRouter()

	// Middleware that sets the authenticated user ID on every request.
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := middleware.ContextWithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// Mount all API routes exactly as router.go does (minus auth middleware).
	router.Route("/api", func(r chi.Router) {
		r.Post("/webhooks/logs", srv.IngestWebhookLogs)

		r.Route("/connections", func(r chi.Router) {
			r.Get("/", srv.ListConnections)
			r.Post("/", srv.CreateConnection)
			r.Get("/{id}", srv.GetConnection)
			r.Put("/{id}", srv.UpdateConnection)
			r.Delete("/{id}", srv.DeleteConnection)
			r.Post("/{id}/test", srv.TestConnection)
		})

		r.Route("/agent", func(r chi.Router) {
			r.Get("/config", srv.GetAgentConfig)
			r.Put("/config", srv.UpdateAgentConfig)
			r.Post("/run", srv.RunAgent)
		})

		r.Get("/logs", srv.ListLogs)

		r.Route("/conversations", func(r chi.Router) {
			r.Get("/", srv.ListConversations)
			r.Get("/{id}", srv.GetConversation)
		})

		r.Get("/auth/me", srv.Me)

		r.Get("/stats", srv.GetDashboardStats)

		r.Route("/reports", func(r chi.Router) {
			r.Get("/", srv.ListReports)
			r.Get("/{id}", srv.GetReport)
		})
	})

	return &testEnv{
		Pool:    pool,
		Queries: queries,
		Server:  srv,
		Router:  router,
		UserID:  userID.String(),
	}
}

// request makes an HTTP request against the test router and returns the recorder.
func (e *testEnv) request(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	e.Router.ServeHTTP(rr, req)
	return rr
}
