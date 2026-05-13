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

// TestIngestWebhookLogs_OrphanTokenReturns401 pins the Phase-10 H2 TOCTOU
// re-check at webhooks.go:ingestEntries. Webhook token resolution runs on
// the cron pool (BYPASSRLS), so a still-valid bearer token whose owner has
// been removed from the connection's org would otherwise reach the
// transactional inserts and surface as opaque 500s from a WITH CHECK
// violation. The re-read through GetConnectionByUser inside the user-scoped
// txn is what converts that into a clean 401.
//
// A regression that drops the re-check would either return 500 (the
// pre-Phase-10 shape) or successfully ingest under the cron BYPASSRLS path
// — both are detectable failures.
func TestIngestWebhookLogs_OrphanTokenReturns401(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	// Create a webhook_logs connection. Activate it and enable the source
	// filter so a healthy request would normally succeed.
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"app_id": env.AppID,
		"name":   "toctou-webhook",
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

	rr = env.request(t, http.MethodPost, "/api/connections/"+connID.String()+"/sources",
		map[string]string{"source_name": "webhook"})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	// Sanity check: a healthy POST works.
	payload := map[string]interface{}{
		"source_type": "webhook",
		"severity":    "info",
		"payload":     map[string]string{"message": "before orphan"},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, "sanity: healthy POST should succeed")

	// Now sever the owner's org membership. Token resolution still
	// finds the connection (cron pool, no RLS, no membership join), but
	// GetConnectionByUser inside the user-scoped txn no longer matches
	// because it joins through org_members.
	_, err = env.Pool.Exec(ctx,
		"DELETE FROM org_members WHERE user_id = $1 AND org_id = $2",
		env.UserID, env.OrgID)
	require.NoError(t, err)

	// Replay the same payload. The TOCTOU re-check must now reject with
	// 401 — not 500, not 201.
	req = httptest.NewRequest(http.MethodPost, "/api/webhooks/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code,
		"orphan-token POST must return 401 (TOCTOU re-check); got body=%s", rec.Body.String())
}

// TestIngestOTLPLogs_OrphanTokenReturns401 mirrors the webhook test for the
// OTLP HTTP endpoint. The two ingestion paths share the cron-pool token
// resolution + user-scoped re-check pattern (see otlp.go:139), so any
// regression would land in both at once if the helper was extracted, or in
// only one if the logic stays duplicated — either case is caught by having
// both tests.
func TestIngestOTLPLogs_OrphanTokenReturns401(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	// OTLP connections use the same webhook_token convention.
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"app_id": env.AppID,
		"name":   "toctou-otlp",
		"type":   "otlp",
	})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	var conn map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	connID, err := uuid.Parse(conn["id"].(string))
	require.NoError(t, err)
	token := conn["config"].(map[string]interface{})["webhook_token"].(string)

	_, err = env.Pool.Exec(ctx,
		"UPDATE connections SET status = 'active' WHERE id = $1", connID)
	require.NoError(t, err)

	// Sever the owner's org membership.
	_, err = env.Pool.Exec(ctx,
		"DELETE FROM org_members WHERE user_id = $1 AND org_id = $2",
		env.UserID, env.OrgID)
	require.NoError(t, err)

	// Minimal valid OTLP payload — one resource, one scope, one record.
	otlp := []byte(`{
		"resourceLogs":[{
			"resource":{"attributes":[{"key":"service.name","value":{"stringValue":"svc"}}]},
			"scopeLogs":[{
				"scope":{"name":"test"},
				"logRecords":[{"timeUnixNano":"0","severityText":"INFO","severityNumber":9,"body":{"stringValue":"hello"}}]
			}]
		}]
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewReader(otlp))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code,
		"orphan-token OTLP POST must return 401 (TOCTOU re-check); got body=%s", rec.Body.String())
}
