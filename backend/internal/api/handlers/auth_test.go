package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMe(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodGet, "/api/auth/me", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var user map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&user))

	// The test user's email should match the pattern from testSetup.
	email, ok := user["email"].(string)
	require.True(t, ok, "response should contain email")
	assert.Contains(t, email, "@heimdall.test")
	assert.Equal(t, env.UserID, user["id"])
}
