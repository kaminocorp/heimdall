package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListConnections(t *testing.T) {
	env := testSetup(t)

	// Create two connections.
	for _, name := range []string{"conn-a", "conn-b"} {
		rr := env.request(t, http.MethodPost, "/api/connections", map[string]string{
			"name": name,
			"type": "webhook_logs",
		})
		require.Equal(t, http.StatusCreated, rr.Code, "create %s", name)
	}

	// List and verify both are returned.
	rr := env.request(t, http.MethodGet, "/api/connections", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var conns []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conns))
	assert.Len(t, conns, 2)

	names := map[string]bool{}
	for _, c := range conns {
		names[c["name"].(string)] = true
	}
	assert.True(t, names["conn-a"])
	assert.True(t, names["conn-b"])
}

func TestCreateConnection(t *testing.T) {
	env := testSetup(t)

	body := map[string]interface{}{
		"name":   "my-postgres",
		"type":   "postgres",
		"config": map[string]string{"host": "localhost"},
	}

	rr := env.request(t, http.MethodPost, "/api/connections", body)
	require.Equal(t, http.StatusCreated, rr.Code)

	var conn map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	assert.Equal(t, "my-postgres", conn["name"])
	assert.Equal(t, "postgres", conn["type"])
	assert.Equal(t, "inactive", conn["status"])
	assert.NotEmpty(t, conn["id"])
	assert.NotEmpty(t, conn["created_at"])
}

func TestCreateConnection_InvalidType(t *testing.T) {
	env := testSetup(t)

	// Missing name.
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"name": "",
		"type": "postgres",
	})
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Missing type.
	rr = env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"name": "valid-name",
		"type": "",
	})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateConnection(t *testing.T) {
	env := testSetup(t)

	// Create a connection.
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"name": "original-name",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code)

	var created map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	connID := created["id"].(string)

	// Update the connection name.
	rr = env.request(t, http.MethodPut, "/api/connections/"+connID, map[string]string{
		"name": "updated-name",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusOK, rr.Code)

	var updated map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&updated))
	assert.Equal(t, "updated-name", updated["name"])
	assert.Equal(t, connID, updated["id"])
}

func TestDeleteConnection(t *testing.T) {
	env := testSetup(t)

	// Create a connection.
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"name": "to-delete",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code)

	var created map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	connID := created["id"].(string)

	// Delete the connection.
	rr = env.request(t, http.MethodDelete, "/api/connections/"+connID, nil)
	require.Equal(t, http.StatusNoContent, rr.Code)

	// List should return empty.
	rr = env.request(t, http.MethodGet, "/api/connections", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var conns []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conns))
	assert.Empty(t, conns)
}
