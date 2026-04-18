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
		"app_id": env.AppID,
		"name":   "webhook-ingest",
		"type":   "webhook_logs",
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

	// Drop-by-default requires explicit opt-in — pre-enable the "webhook"
	// source for this app/connection so the entry doesn't get filtered out.
	rr = env.request(t, http.MethodPost, "/api/connections/"+connID.String()+"/sources",
		map[string]string{"source_name": "webhook"})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

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
	assert.Equal(t, float64(1), result["accepted"])
	assert.Equal(t, "native", result["format"])
}

// TestIngestWebhookLogs_DropByDefault verifies that without any enabled
// source filter, every entry in the batch is dropped — the core semantic of
// Phase 1 source filtering.
func TestIngestWebhookLogs_DropByDefault(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	// Create and activate a webhook_logs connection.
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"app_id": env.AppID,
		"name":   "drop-by-default",
		"type":   "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code)

	var conn map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	connID, err := uuid.Parse(conn["id"].(string))
	require.NoError(t, err)
	token := conn["config"].(map[string]interface{})["webhook_token"].(string)

	_, err = env.Pool.Exec(ctx,
		"UPDATE connections SET status = 'active' WHERE id = $1", connID)
	require.NoError(t, err)

	// POST without first enabling any filter.
	payload := map[string]interface{}{
		"source_type": "app",
		"severity":    "info",
		"payload":     map[string]string{"message": "dropped"},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	env.Router.ServeHTTP(recorder, req)

	// The batch is accepted (201) but no entries pass the filter.
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&result))
	assert.Equal(t, float64(0), result["accepted"], "drop-by-default — no filter, no entries stored")
	assert.Equal(t, float64(1), result["filtered"])

	// The source should nonetheless be discovered for the UI to surface.
	rr = env.request(t, http.MethodGet, "/api/connections/"+connID.String()+"/sources", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	var sources []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&sources))
	require.Len(t, sources, 1, "discovery upsert runs even when filtering drops the entry")
	assert.Equal(t, "app", sources[0]["source_name"])
	assert.Equal(t, false, sources[0]["enabled"])
}

// TestIngestWebhookLogs_FlyioSourceName verifies that Fly.io logs are
// filtered by the bare app name (fly.app.name), not the prefixed source
// type. This is the whole point of Phase 1 — one org-wide Fly drain can
// feed multiple Heimdall apps, each enabling different source names.
func TestIngestWebhookLogs_FlyioSourceName(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	rr := env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"app_id": env.AppID,
		"name":   "fly-drain",
		"type":   "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code)

	var conn map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	connID, err := uuid.Parse(conn["id"].(string))
	require.NoError(t, err)
	token := conn["config"].(map[string]interface{})["webhook_token"].(string)

	_, err = env.Pool.Exec(ctx,
		"UPDATE connections SET status = 'active' WHERE id = $1", connID)
	require.NoError(t, err)

	// Enable "trajan" but NOT "other-app". The Fly.io parser will set
	// SourceName = fly.app.name for each entry.
	rr = env.request(t, http.MethodPost, "/api/connections/"+connID.String()+"/sources",
		map[string]string{"source_name": "trajan"})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	// Batch with one enabled source and one disabled source.
	body := []byte(`[
		{"message":"from trajan","host":"a","source_type":"fly_io","fly":{"app":{"name":"trajan"},"machine":{"id":"m1"},"region":"lhr"},"log":{"level":"info"}},
		{"message":"from other","host":"b","source_type":"fly_io","fly":{"app":{"name":"other-app"},"machine":{"id":"m2"},"region":"lhr"},"log":{"level":"info"}}
	]`)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	env.Router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&result))
	assert.Equal(t, float64(1), result["accepted"], "only trajan passes")
	assert.Equal(t, float64(1), result["filtered"], "other-app dropped")

	// Discovery sees both sources regardless of filter state.
	rr = env.request(t, http.MethodGet, "/api/connections/"+connID.String()+"/sources", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	var sources []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&sources))
	require.Len(t, sources, 2)
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
