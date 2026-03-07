package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
)

// createBareUser creates a user with no org and returns the user ID.
func createBareUser(t *testing.T, env *testEnv) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	userID := uuid.New()
	email := fmt.Sprintf("bare-%s@heimdall.test", userID.String())

	_, err := env.Pool.Exec(ctx,
		`INSERT INTO auth.users (id, email, instance_id, aud, role, encrypted_password, confirmation_token, created_at, updated_at)
		 VALUES ($1, $2, '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', '', '', now(), now())`,
		userID, email,
	)
	require.NoError(t, err)
	_, err = env.Pool.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, email)
	require.NoError(t, err)

	t.Cleanup(func() {
		env.Pool.Exec(context.Background(), "DELETE FROM auth.users WHERE id = $1", userID)
	})
	return userID
}

// envForUser returns a testEnv that authenticates as the given user,
// reusing the pool and server from the parent env.
func envForUser(t *testing.T, parent *testEnv, userID uuid.UUID) *testEnv {
	t.Helper()
	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := middleware.ContextWithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	router.Route("/api", func(r chi.Router) {
		r.Post("/onboard", parent.Server.Onboard)
	})
	return &testEnv{
		Pool:    parent.Pool,
		Queries: parent.Queries,
		Server:  parent.Server,
		Router:  router,
		UserID:  userID.String(),
	}
}

// requestJSON is a convenience alias so envForUser envs can call request().
func (e *testEnv) requestJSON(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
	return e.request(t, method, path, body)
}

func TestGetOrganization(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodGet, "/api/org", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var org map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&org))

	assert.Equal(t, env.OrgID, org["id"])
	assert.Equal(t, "Test Org", org["name"])
	assert.NotEmpty(t, org["slug"])
	assert.NotEmpty(t, org["created_at"])
}

func TestOnboard(t *testing.T) {
	env := testSetup(t)

	// testSetup's user already has an org, so create a fresh user without one.
	userID := createBareUser(t, env)

	// Build a second test env that authenticates as the bare user.
	bareEnv := envForUser(t, env, userID)

	body := map[string]string{
		"org_name": "Onboard Test Org",
		"org_slug": "onboard-test-" + userID.String()[:8],
		"app_name": "My First App",
	}

	rr := bareEnv.request(t, http.MethodPost, "/api/onboard", body)
	require.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	org, ok := resp["organization"].(map[string]interface{})
	require.True(t, ok, "response should contain organization")
	assert.Equal(t, "Onboard Test Org", org["name"])
	assert.Equal(t, "onboard-test-"+userID.String()[:8], org["slug"])

	app, ok := resp["application"].(map[string]interface{})
	require.True(t, ok, "response should contain application")
	assert.Equal(t, "My First App", app["name"])
	assert.Equal(t, "active", app["status"])
	assert.Equal(t, org["id"], app["org_id"])

	// Clean up the new org (cascade deletes app + config).
	env.Pool.Exec(t.Context(), "DELETE FROM organizations WHERE slug = $1", "onboard-test-"+userID.String()[:8])
}

func TestOnboard_AlreadyOnboarded(t *testing.T) {
	env := testSetup(t)

	// testSetup's user already has an org — should get 409.
	body := map[string]string{
		"org_name": "Duplicate Onboard",
		"org_slug": "dup-onboard-" + env.UserID[:8],
	}

	rr := env.request(t, http.MethodPost, "/api/onboard", body)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestOnboard_DuplicateSlug(t *testing.T) {
	env := testSetup(t)

	// Need a bare user (no org) so the idempotency guard doesn't fire first.
	userID := createBareUser(t, env)
	bareEnv := envForUser(t, env, userID)

	// Get the existing slug from testSetup's org.
	rr := env.request(t, http.MethodGet, "/api/org", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var org map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&org))
	existingSlug := org["slug"].(string)

	body := map[string]string{
		"org_name": "Duplicate Org",
		"org_slug": existingSlug,
	}

	rr = bareEnv.request(t, http.MethodPost, "/api/onboard", body)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestOnboard_MissingFields(t *testing.T) {
	env := testSetup(t)

	// Missing org_name.
	rr := env.request(t, http.MethodPost, "/api/onboard", map[string]string{
		"org_slug": "test-slug",
	})
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Missing org_slug.
	rr = env.request(t, http.MethodPost, "/api/onboard", map[string]string{
		"org_name": "Test",
	})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
