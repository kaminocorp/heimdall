package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIngestWebhookLogs(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	// Create a webhook_logs connection (auto-generates a webhook_token in config).
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"name": "webhook-ingest",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code)

	var conn map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))

	connID, err := uuid.Parse(conn["id"].(string))
	require.NoError(t, err)

	// Extract the webhook_token from the connection config.
	cfgRaw, ok := conn["config"].(map[string]interface{})
	require.True(t, ok, "config should be an object")
	token, ok := cfgRaw["webhook_token"].(string)
	require.True(t, ok, "config should contain webhook_token")
	require.NotEmpty(t, token)

	// The webhook query requires status='active', so activate the connection.
	_, err = env.Pool.Exec(ctx,
		"UPDATE connections SET status = 'active' WHERE id = $1", connID)
	require.NoError(t, err)

	// POST a log entry via the webhook endpoint with the bearer token.
	payload := map[string]interface{}{
		"source_type": "webhook",
		"severity":    "info",
		"payload":     map[string]string{"message": "test webhook log"},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	env.Router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&result))
	assert.NotEmpty(t, result["id"])
}

func TestIngestWebhookLogs_InvalidToken(t *testing.T) {
	env := testSetup(t)

	payload := map[string]interface{}{
		"source_type": "webhook",
		"severity":    "info",
		"payload":     map[string]string{"message": "bad token"},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid-token-12345")

	recorder := httptest.NewRecorder()
	env.Router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
