package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListLogs_Empty(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodGet, "/api/logs", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	// data should be an empty array or null, total should be 0.
	total, ok := resp["total"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(0), total)
}

func TestListLogs_WithEntries(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	// First create a connection to use as the FK reference.
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"name": "log-source",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code)

	var conn map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	connID, err := uuid.Parse(conn["id"].(string))
	require.NoError(t, err)

	userID, err := uuid.Parse(env.UserID)
	require.NoError(t, err)

	// Insert two log entries directly via SQL.
	for i := 0; i < 2; i++ {
		_, err := env.Pool.Exec(ctx,
			`INSERT INTO log_buffer (id, connection_id, source_type, payload, user_id)
			 VALUES ($1, $2, 'webhook', '{"message":"test log"}', $3)`,
			uuid.New(), connID, userID,
		)
		require.NoError(t, err)
	}

	rr = env.request(t, http.MethodGet, "/api/logs?source=raw", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	total, ok := resp["total"].(float64)
	require.True(t, ok)
	assert.GreaterOrEqual(t, total, float64(2))

	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(data), 2)
}

func TestGetDashboardStats(t *testing.T) {
	env := testSetup(t)

	// Create a connection so connection_count >= 1.
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"name": "stats-conn",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code)

	rr = env.request(t, http.MethodGet, "/api/stats", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var stats map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&stats))

	// Verify expected keys exist.
	_, hasLogCount := stats["log_count_24h"]
	assert.True(t, hasLogCount, "response should contain log_count_24h")

	connCount, hasConnCount := stats["connection_count"]
	assert.True(t, hasConnCount, "response should contain connection_count")
	assert.GreaterOrEqual(t, connCount, float64(1))

	_, hasActive := stats["active_connections"]
	assert.True(t, hasActive, "response should contain active_connections")
}
