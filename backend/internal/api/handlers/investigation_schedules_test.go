package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Valid baseline request used across tests. Individual tests copy and mutate
// this to exercise specific validation edges.
func baselineCreateBody() map[string]any {
	return map[string]any{
		"name":          "Check slow queries",
		"prompt":        "Look for slow DB queries in the last hour and summarise the worst offenders.",
		"interval_secs": 300,
		"enabled":       true,
	}
}

func TestListSchedules_Empty(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var schedules []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&schedules))
	assert.Empty(t, schedules)
}

func TestCreateSchedule_Happy(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodPost, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), baselineCreateBody())
	require.Equal(t, http.StatusCreated, rr.Code)

	var created map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	assert.Equal(t, "Check slow queries", created["name"])
	assert.Equal(t, env.AppID, created["app_id"])
	assert.Equal(t, true, created["enabled"])
	assert.NotEmpty(t, created["id"])

	// Verify it shows up in the list.
	rr = env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), nil)
	require.Equal(t, http.StatusOK, rr.Code)
	var schedules []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&schedules))
	assert.Len(t, schedules, 1)
}

func TestCreateSchedule_DefaultsEnabledWhenOmitted(t *testing.T) {
	env := testSetup(t)

	body := baselineCreateBody()
	delete(body, "enabled")

	rr := env.request(t, http.MethodPost, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), body)
	require.Equal(t, http.StatusCreated, rr.Code)

	var created map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	assert.Equal(t, true, created["enabled"], "enabled should default to true when omitted")
}

func TestCreateSchedule_ValidationErrors(t *testing.T) {
	env := testSetup(t)

	cases := []struct {
		name    string
		mutate  func(m map[string]any)
		message string
	}{
		{
			name:    "empty name",
			mutate:  func(m map[string]any) { m["name"] = "" },
			message: "name is required",
		},
		{
			name:    "whitespace-only name",
			mutate:  func(m map[string]any) { m["name"] = "   " },
			message: "name is required",
		},
		{
			name:    "name too long",
			mutate:  func(m map[string]any) { m["name"] = strings.Repeat("x", 200) },
			message: "name exceeds maximum length",
		},
		{
			name:    "empty prompt",
			mutate:  func(m map[string]any) { m["prompt"] = "" },
			message: "prompt is required",
		},
		{
			name:    "prompt too long",
			mutate:  func(m map[string]any) { m["prompt"] = strings.Repeat("x", 6000) },
			message: "prompt exceeds maximum length",
		},
		{
			name:    "interval below minimum",
			mutate:  func(m map[string]any) { m["interval_secs"] = 30 },
			message: "interval_secs must be between 60 and 86400",
		},
		{
			name:    "interval above maximum",
			mutate:  func(m map[string]any) { m["interval_secs"] = 100000 },
			message: "interval_secs must be between 60 and 86400",
		},
		{
			// Phase 4: exactly-one-of check — both fields non-zero is rejected.
			name: "both interval and cron set",
			mutate: func(m map[string]any) {
				m["interval_secs"] = 300
				m["cron_expr"] = "*/5 * * * *"
			},
			message: "provide either interval_secs or cron_expr, not both",
		},
		{
			// Phase 4: exactly-one-of check — neither field is rejected.
			name: "neither interval nor cron",
			mutate: func(m map[string]any) {
				delete(m, "interval_secs")
				delete(m, "cron_expr")
			},
			message: "provide either interval_secs or cron_expr",
		},
		{
			// Phase 4: cron parser rejects garbage before the DB sees it.
			name: "invalid cron expression",
			mutate: func(m map[string]any) {
				delete(m, "interval_secs")
				m["cron_expr"] = "not a cron expression"
			},
			message: "invalid cron_expr",
		},
		{
			// Phase 4: cron expression length sanity cap.
			name: "cron expression too long",
			mutate: func(m map[string]any) {
				delete(m, "interval_secs")
				m["cron_expr"] = strings.Repeat("*", 300)
			},
			message: "cron_expr exceeds maximum length",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := baselineCreateBody()
			tc.mutate(body)

			rr := env.request(t, http.MethodPost, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), body)
			assert.Equal(t, http.StatusBadRequest, rr.Code)

			var errBody map[string]string
			require.NoError(t, json.NewDecoder(rr.Body).Decode(&errBody))
			assert.Contains(t, errBody["error"], tc.message)
		})
	}
}

// TestCreateSchedule_CronMode verifies the Phase 4 cron path end-to-end:
// a client submits a cron expression (with no interval_secs), the handler
// validates and stores it, and the returned JSON reflects the cron_expr
// with interval_secs zeroed out. Guards the "exactly-one-of" storage rule.
func TestCreateSchedule_CronMode(t *testing.T) {
	env := testSetup(t)

	body := map[string]any{
		"name":      "Hourly pg_stat check",
		"prompt":    "Query pg_stat_statements for queries taking >1s in the last hour.",
		"cron_expr": "0 * * * *",
		"enabled":   true,
	}
	rr := env.request(t, http.MethodPost, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), body)
	require.Equal(t, http.StatusCreated, rr.Code)

	var created map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	assert.Equal(t, "Hourly pg_stat check", created["name"])
	assert.Equal(t, "0 * * * *", created["cron_expr"])
	// When cron mode is used, interval_secs is stored as 0 (sentinel: unused).
	assert.Equal(t, float64(0), created["interval_secs"])
}

// TestUpdateSchedule_IntervalToCron verifies that flipping a schedule from
// interval mode to cron mode cleanly clears interval_secs and populates
// cron_expr. This is the realistic "user clicks Advanced toggle in the modal"
// path — if the handler didn't zero out interval_secs, shouldFire would
// still behave correctly (cron wins) but the stored row would be ambiguous.
func TestUpdateSchedule_IntervalToCron(t *testing.T) {
	env := testSetup(t)

	// Create in interval mode.
	rr := env.request(t, http.MethodPost, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), baselineCreateBody())
	require.Equal(t, http.StatusCreated, rr.Code)
	var created map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	scheduleID := created["id"].(string)

	// PATCH to cron mode.
	update := map[string]any{
		"name":      "Renamed cron schedule",
		"prompt":    "Look for slow DB queries in the last hour.",
		"cron_expr": "*/15 * * * *",
		"enabled":   true,
	}
	rr = env.request(t, http.MethodPatch, fmt.Sprintf("/api/apps/%s/schedules/%s", env.AppID, scheduleID), update)
	require.Equal(t, http.StatusOK, rr.Code)

	var updated map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&updated))
	assert.Equal(t, "*/15 * * * *", updated["cron_expr"])
	assert.Equal(t, float64(0), updated["interval_secs"], "interval_secs should be zeroed in cron mode")
}

func TestUpdateSchedule_Happy(t *testing.T) {
	env := testSetup(t)

	// Create first, then update.
	rr := env.request(t, http.MethodPost, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), baselineCreateBody())
	require.Equal(t, http.StatusCreated, rr.Code)
	var created map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	scheduleID := created["id"].(string)

	update := map[string]any{
		"name":          "Renamed schedule",
		"prompt":        "New prompt text.",
		"interval_secs": 600,
		"enabled":       false,
	}
	rr = env.request(t, http.MethodPatch, fmt.Sprintf("/api/apps/%s/schedules/%s", env.AppID, scheduleID), update)
	require.Equal(t, http.StatusOK, rr.Code)

	var updated map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&updated))
	assert.Equal(t, "Renamed schedule", updated["name"])
	assert.Equal(t, "New prompt text.", updated["prompt"])
	assert.Equal(t, float64(600), updated["interval_secs"])
	assert.Equal(t, false, updated["enabled"])
}

func TestUpdateSchedule_NotFound(t *testing.T) {
	env := testSetup(t)

	update := baselineCreateBody()
	rr := env.request(t, http.MethodPatch, fmt.Sprintf("/api/apps/%s/schedules/%s", env.AppID, uuid.New().String()), update)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateSchedule_WrongApp(t *testing.T) {
	env := testSetup(t)

	// Create a schedule in env.AppID.
	rr := env.request(t, http.MethodPost, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), baselineCreateBody())
	require.Equal(t, http.StatusCreated, rr.Code)
	var created map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	scheduleID := created["id"].(string)

	// Create a second app in the same org.
	otherAppID := uuid.New()
	_, err := env.Pool.Exec(t.Context(),
		"INSERT INTO applications (id, org_id, name, status) VALUES ($1, $2, $3, $4)",
		otherAppID, env.OrgID, "Other App", "active",
	)
	require.NoError(t, err)

	// Try to update the first app's schedule via the second app's URL.
	// Ownership check must return 404 (not leak existence via 403).
	update := baselineCreateBody()
	rr = env.request(t, http.MethodPatch, fmt.Sprintf("/api/apps/%s/schedules/%s", otherAppID, scheduleID), update)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestDeleteSchedule_Happy(t *testing.T) {
	env := testSetup(t)

	// Create then delete.
	rr := env.request(t, http.MethodPost, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), baselineCreateBody())
	require.Equal(t, http.StatusCreated, rr.Code)
	var created map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	scheduleID := created["id"].(string)

	rr = env.request(t, http.MethodDelete, fmt.Sprintf("/api/apps/%s/schedules/%s", env.AppID, scheduleID), nil)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify it's gone from the list.
	rr = env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), nil)
	require.Equal(t, http.StatusOK, rr.Code)
	var schedules []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&schedules))
	assert.Empty(t, schedules)
}

func TestDeleteSchedule_NotFound(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodDelete, fmt.Sprintf("/api/apps/%s/schedules/%s", env.AppID, uuid.New().String()), nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCreateSchedule_WrongOrgApp(t *testing.T) {
	env := testSetup(t)

	// Create a completely separate org + app that env's user doesn't own.
	otherOrgID := uuid.New()
	otherOrgSlug := fmt.Sprintf("other-org-%s", otherOrgID.String()[:8])
	_, err := env.Pool.Exec(t.Context(),
		"INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)",
		otherOrgID, "Other Org", otherOrgSlug,
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		env.Pool.Exec(t.Context(), "DELETE FROM organizations WHERE id = $1", otherOrgID)
	})

	otherAppID := uuid.New()
	_, err = env.Pool.Exec(t.Context(),
		"INSERT INTO applications (id, org_id, name, status) VALUES ($1, $2, $3, $4)",
		otherAppID, otherOrgID, "Other App", "active",
	)
	require.NoError(t, err)

	// authorizeApp should reject — user has no access to otherAppID.
	rr := env.request(t, http.MethodPost, fmt.Sprintf("/api/apps/%s/schedules", otherAppID), baselineCreateBody())
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRunScheduleNow_NoAgent(t *testing.T) {
	// testSetup creates a Server with Agent=nil. RunScheduleNow should
	// detect this and return 503 rather than panicking with a nil-pointer
	// dereference when it tries to call s.Agent.RunScheduledInvestigation.
	env := testSetup(t)

	// Create a schedule first.
	rr := env.request(t, http.MethodPost, fmt.Sprintf("/api/apps/%s/schedules", env.AppID), baselineCreateBody())
	require.Equal(t, http.StatusCreated, rr.Code)
	var created map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	scheduleID := created["id"].(string)

	rr = env.request(t, http.MethodPost, fmt.Sprintf("/api/apps/%s/schedules/%s/run", env.AppID, scheduleID), nil)
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}
