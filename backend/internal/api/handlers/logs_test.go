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
	connID := env.createTestConnection(t, "log-source")
	connUUID, err := uuid.Parse(connID)
	require.NoError(t, err)

	userID, err := uuid.Parse(env.UserID)
	require.NoError(t, err)

	// Insert two log entries directly via SQL.
	for i := 0; i < 2; i++ {
		_, err := env.Pool.Exec(ctx,
			`INSERT INTO log_buffer (id, connection_id, source_type, payload, user_id)
			 VALUES ($1, $2, 'webhook', '{"message":"test log"}', $3)`,
			uuid.New(), connUUID, userID,
		)
		require.NoError(t, err)
	}

	rr := env.request(t, http.MethodGet, "/api/logs?source=raw", nil)
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

func TestListLogs_AppScoped(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	connID := env.createTestConnection(t, "app-scoped-source")
	connUUID, err := uuid.Parse(connID)
	require.NoError(t, err)

	userID, err := uuid.Parse(env.UserID)
	require.NoError(t, err)
	appID, err := uuid.Parse(env.AppID)
	require.NoError(t, err)

	// Insert one log WITH app_id and one WITHOUT.
	_, err = env.Pool.Exec(ctx,
		`INSERT INTO log_buffer (id, connection_id, source_type, payload, user_id, app_id)
		 VALUES ($1, $2, 'webhook', '{"message":"scoped"}', $3, $4)`,
		uuid.New(), connUUID, userID, appID,
	)
	require.NoError(t, err)

	_, err = env.Pool.Exec(ctx,
		`INSERT INTO log_buffer (id, connection_id, source_type, payload, user_id, app_id)
		 VALUES ($1, $2, 'webhook', '{"message":"unscoped"}', $3, NULL)`,
		uuid.New(), connUUID, userID,
	)
	require.NoError(t, err)

	// Without app_id filter: both rows returned.
	rr := env.request(t, http.MethodGet, "/api/logs?source=raw", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var allResp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&allResp))
	allTotal := allResp["total"].(float64)
	assert.GreaterOrEqual(t, allTotal, float64(2))

	// With app_id filter: only the scoped row returned.
	rr = env.request(t, http.MethodGet, "/api/logs?source=raw&app_id="+env.AppID, nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var scopedResp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&scopedResp))
	scopedTotal := scopedResp["total"].(float64)
	assert.Equal(t, float64(1), scopedTotal)

	data, ok := scopedResp["data"].([]interface{})
	require.True(t, ok)
	require.Len(t, data, 1)

	entry := data[0].(map[string]interface{})
	assert.Equal(t, "scoped", entry["summary"])
}

func TestListLogs_AppScoped_InvalidAppID(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodGet, "/api/logs?app_id=not-a-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestListLogs_AppScoped_WrongOrg(t *testing.T) {
	env := testSetup(t)

	// Use a random UUID that doesn't belong to this user's org.
	rr := env.request(t, http.MethodGet, "/api/logs?app_id="+uuid.New().String(), nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

