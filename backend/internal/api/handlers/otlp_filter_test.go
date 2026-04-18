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

// TestIngestOTLP_ServiceNameFiltering verifies the Phase 4 behaviour:
// service.name is extracted from OTLP resource attributes and used as the
// filter key. Enabled services are persisted, others are dropped.
func TestIngestOTLP_ServiceNameFiltering(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	// Create an OTLP connection. CreateConnection auto-generates a
	// webhook_token for type="otlp".
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]any{
		"app_id": env.AppID,
		"name":   "otlp-test",
		"type":   "otlp",
	})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	var conn map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	connID, err := uuid.Parse(conn["id"].(string))
	require.NoError(t, err)
	token := conn["config"].(map[string]any)["webhook_token"].(string)

	_, err = env.Pool.Exec(ctx,
		"UPDATE connections SET status = 'active' WHERE id = $1", connID)
	require.NoError(t, err)

	// Enable "checkout-service" but NOT "payment-service".
	rr = env.request(t, http.MethodPost, "/api/connections/"+connID.String()+"/sources",
		map[string]string{"source_name": "checkout-service"})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	// Build an OTLP payload with two resourceLogs — one per service.
	otlpPayload := map[string]any{
		"resourceLogs": []map[string]any{
			{
				"resource": map[string]any{
					"attributes": []map[string]any{
						{"key": "service.name", "value": map[string]any{"stringValue": "checkout-service"}},
					},
				},
				"scopeLogs": []map[string]any{
					{
						"scope": map[string]any{"name": "checkout"},
						"logRecords": []map[string]any{
							{"severityNumber": 9, "body": map[string]any{"stringValue": "checkout log 1"}},
						},
					},
				},
			},
			{
				"resource": map[string]any{
					"attributes": []map[string]any{
						{"key": "service.name", "value": map[string]any{"stringValue": "payment-service"}},
					},
				},
				"scopeLogs": []map[string]any{
					{
						"scope": map[string]any{"name": "payment"},
						"logRecords": []map[string]any{
							{"severityNumber": 17, "body": map[string]any{"stringValue": "payment log 1"}},
						},
					},
				},
			},
		},
	}
	body, err := json.Marshal(otlpPayload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	env.Router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var result map[string]any
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&result))
	assert.Equal(t, float64(1), result["accepted"], "only checkout-service passes the filter")
	assert.Equal(t, float64(1), result["filtered"], "payment-service dropped")

	// Discovery recorded both services, enabled or not — so the user can
	// see both in the selector.
	rr = env.request(t, http.MethodGet, "/api/connections/"+connID.String()+"/sources", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	var sources []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&sources))
	require.Len(t, sources, 2)
}

// TestIngestOTLP_DropByDefault confirms that OTLP ingestion now applies the
// same drop-by-default semantic as webhook ingestion after Phase 4.
func TestIngestOTLP_DropByDefault(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	rr := env.request(t, http.MethodPost, "/api/connections", map[string]any{
		"app_id": env.AppID,
		"name":   "otlp-default",
		"type":   "otlp",
	})
	require.Equal(t, http.StatusCreated, rr.Code)
	var conn map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	connID, _ := uuid.Parse(conn["id"].(string))
	token := conn["config"].(map[string]any)["webhook_token"].(string)
	_, _ = env.Pool.Exec(ctx,
		"UPDATE connections SET status = 'active' WHERE id = $1", connID)

	// Don't enable any filter — whatever we send should be dropped.
	otlpPayload := map[string]any{
		"resourceLogs": []map[string]any{
			{
				"resource": map[string]any{
					"attributes": []map[string]any{
						{"key": "service.name", "value": map[string]any{"stringValue": "myservice"}},
					},
				},
				"scopeLogs": []map[string]any{
					{
						"logRecords": []map[string]any{
							{"body": map[string]any{"stringValue": "test"}},
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(otlpPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	env.Router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	var result map[string]any
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&result))
	assert.Equal(t, float64(0), result["accepted"])
	assert.Equal(t, float64(1), result["filtered"])

	// Verify nothing landed in log_buffer.
	var count int
	err := env.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM log_buffer WHERE connection_id = $1", connID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestIngestOTLP_OrgScopedFanOut verifies the Phase 2 fan-out now works for
// OTLP too — the Phase 2 rejection ("org-scoped OTLP not supported") that
// guarded against silent routing failures is gone, because Phase 4's
// source_name plumbing makes routing well-defined.
func TestIngestOTLP_OrgScopedFanOut(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	// Second app in the same org.
	secondAppID := uuid.New()
	_, err := env.Pool.Exec(ctx,
		"INSERT INTO applications (id, org_id, name, status) VALUES ($1, $2, $3, $4)",
		secondAppID, env.OrgID, "Test App 2", "active",
	)
	require.NoError(t, err)
	_, _ = env.Pool.Exec(ctx,
		"INSERT INTO app_agent_config (app_id, model, mode, schedule_interval_secs) VALUES ($1, $2, $3, $4)",
		secondAppID, "claude-sonnet-4-6", "off", 60,
	)

	// Org-scoped OTLP connection (no app_id).
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]any{
		"name": "org-otlp",
		"type": "otlp",
	})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())
	var conn map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	require.Nil(t, conn["app_id"], "confirm org-scoped")
	connID, _ := uuid.Parse(conn["id"].(string))
	token := conn["config"].(map[string]any)["webhook_token"].(string)
	_, _ = env.Pool.Exec(ctx,
		"UPDATE connections SET status = 'active' WHERE id = $1", connID)

	// Both apps enable the same service — fan-out should duplicate.
	for _, appID := range []string{env.AppID, secondAppID.String()} {
		rr := env.request(t, http.MethodPost,
			"/api/connections/"+connID.String()+"/sources?app_id="+appID,
			map[string]string{"source_name": "shared-service"})
		require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())
	}

	otlpPayload := map[string]any{
		"resourceLogs": []map[string]any{
			{
				"resource": map[string]any{
					"attributes": []map[string]any{
						{"key": "service.name", "value": map[string]any{"stringValue": "shared-service"}},
					},
				},
				"scopeLogs": []map[string]any{
					{"logRecords": []map[string]any{{"body": map[string]any{"stringValue": "hello"}}}},
				},
			},
		},
	}
	body, _ := json.Marshal(otlpPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	env.Router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	// One input record fans out into two rows, one per app.
	var total int
	err = env.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM log_buffer WHERE connection_id = $1", connID).Scan(&total)
	require.NoError(t, err)
	assert.Equal(t, 2, total, "org-scoped OTLP fans out across enabled apps")
}
