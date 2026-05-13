package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// insertPipelineTestLog creates a log_buffer row owned by the test app +
// user so subsequent pipeline-event inserts have a valid FK target.
// Returns the log UUID.
func insertPipelineTestLog(t *testing.T, env *testEnv, connID uuid.UUID) uuid.UUID {
	t.Helper()
	appUUID := uuid.MustParse(env.AppID)
	userUUID := uuid.MustParse(env.UserID)
	row, err := env.Queries.InsertLogEntry(context.Background(), db.InsertLogEntryParams{
		ConnectionID: connID,
		SourceType:   "test",
		Severity:     pgtype.Text{String: "info", Valid: true},
		Payload:      []byte(`{"msg":"test"}`),
		UserID:       userUUID,
		AppID:        appUUID,
	})
	require.NoError(t, err)
	return row.ID
}

// walkLogThroughPipeline writes one event per stage for a given log using
// the PipelineWriter, so the integration test covers the exact production
// code path rather than raw SQL. Phase 2 of the RLS rollout made the writer
// caller-supplied for queries — every Write* takes a *db.Queries the
// caller has scoped to a UserQueries transaction. The test passes
// env.Queries (raw, RLS-cosmetic) which is the right contract today;
// post-Phase-7 the test would need to wrap each call in
// env.Server.Pools.WithUserQueries.
func walkLogThroughPipeline(t *testing.T, pw *agent.PipelineWriter, logID, appID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	pw.WriteIngestion(ctx, env.Queries, agent.IngestionInput{
		LogID: logID, AppID: appID, SourceType: "test", Severity: "info",
	})
	pw.WriteClassified(ctx, env.Queries, agent.ClassifiedInput{
		LogID: logID, AppID: appID,
		Type: "ERROR", Category: "runtime_exception",
		Severity: "error", Confidence: 0.92,
		Summary: "runtime exception in handler",
	})
	pw.WriteGate(ctx, env.Queries, agent.GateInput{
		LogID: logID, AppID: appID,
		Escalated: true, RuleID: agent.RuleErrorType,
	})
	// Create a fake agent_log row to reference as AssessmentID. Using
	// EmitLog-equivalent raw insert since env.Server.Agent is nil.
	var assessmentID uuid.UUID
	err := env.Pool.QueryRow(context.Background(),
		`INSERT INTO agent_log (user_id, entry_type, summary, app_id)
		 VALUES ($1, 'monitoring', 'test', $2) RETURNING id`,
		uuid.MustParse(env.UserID), appID,
	).Scan(&assessmentID)
	require.NoError(t, err)
	pw.WriteAssessment(ctx, env.Queries, agent.AssessmentInput{
		LogID: logID, AppID: appID, AssessmentID: assessmentID,
	})
}

// env is set as a package-level var by TestPipeline_Integration below so
// walkLogThroughPipeline can reach env.Pool. Tests use a single env per
// test function; this keeps the helper signature readable without an
// extra plumbing argument.
var env *testEnv

func TestPipeline_Integration(t *testing.T) {
	env = testSetup(t)
	connIDStr := env.createTestConnection(t, "pipeline-conn")
	connID := uuid.MustParse(connIDStr)

	bus := agent.NewPipelineBus()
	pw := agent.NewPipelineWriter(bus)

	t.Run("four-stage walk persists and fans out", func(t *testing.T) {
		appUUID := uuid.MustParse(env.AppID)
		sub, unsub := bus.Subscribe(appUUID)
		defer unsub()

		logID := insertPipelineTestLog(t, env, connID)
		walkLogThroughPipeline(t, pw, logID, appUUID)

		// Four published events should reach the subscriber (one per
		// stage). Drain with a generous timeout so CI wobble doesn't
		// flake.
		seen := map[string]bool{}
		deadline := time.After(3 * time.Second)
		for len(seen) < 4 {
			select {
			case evt := <-sub:
				assert.Equal(t, logID, evt.LogID)
				seen[evt.Stage] = true
			case <-deadline:
				t.Fatalf("timed out; saw stages: %v", seen)
			}
		}
		assert.True(t, seen[agent.StageIngestion])
		assert.True(t, seen[agent.StageClassified])
		assert.True(t, seen[agent.StageGate])
		assert.True(t, seen[agent.StageAssessment])

		// DB journey reflects all four, in order.
		rows, err := env.Queries.GetPipelineEventsByLog(context.Background(), logID)
		require.NoError(t, err)
		require.Len(t, rows, 4)
		assert.Equal(t, agent.StageIngestion, rows[0].Stage)
		assert.Equal(t, agent.StageAssessment, rows[3].Stage)
		// Gate row carries the rule id.
		for _, r := range rows {
			if r.Stage == agent.StageGate {
				assert.True(t, r.RuleHit.Valid)
				assert.Equal(t, agent.RuleErrorType, r.RuleHit.String)
			}
		}
	})

	t.Run("cascade delete sheds pipeline events with log_buffer", func(t *testing.T) {
		appUUID := uuid.MustParse(env.AppID)
		logID := insertPipelineTestLog(t, env, connID)
		walkLogThroughPipeline(t, pw, logID, appUUID)

		rows, err := env.Queries.GetPipelineEventsByLog(context.Background(), logID)
		require.NoError(t, err)
		require.Len(t, rows, 4, "precondition: 4 stage rows inserted")

		// Delete the parent log row — the ON DELETE CASCADE on log_id
		// should sweep every associated pipeline event.
		_, err = env.Pool.Exec(context.Background(), "DELETE FROM log_buffer WHERE id = $1", logID)
		require.NoError(t, err)

		rows, err = env.Queries.GetPipelineEventsByLog(context.Background(), logID)
		require.NoError(t, err)
		assert.Len(t, rows, 0, "cascade delete must sweep all stage rows")
	})

	t.Run("bootstrap returns stats and recent events", func(t *testing.T) {
		appUUID := uuid.MustParse(env.AppID)
		// Fresh log; walk through pipeline so stats are non-zero.
		logID := insertPipelineTestLog(t, env, connID)
		walkLogThroughPipeline(t, pw, logID, appUUID)

		rr := env.request(t, http.MethodGet, "/api/apps/"+env.AppID+"/pipeline/bootstrap?window=1h&tickerLimit=10", nil)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

		var body struct {
			Stats struct {
				IngestionCount  int64   `json:"ingestion_count"`
				ClassifiedCount int64   `json:"classified_count"`
				FlaggedCount    int64   `json:"flagged_count"`
				AssessmentCount int64   `json:"assessment_count"`
				AvgConfidence   float64 `json:"avg_confidence"`
				WindowSeconds   int     `json:"window_seconds"`
			} `json:"stats"`
			RecentEvents []map[string]any `json:"recent_events"`
			Cursor       time.Time        `json:"cursor"`
		}
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
		assert.GreaterOrEqual(t, body.Stats.IngestionCount, int64(1))
		assert.GreaterOrEqual(t, body.Stats.ClassifiedCount, int64(1))
		assert.GreaterOrEqual(t, body.Stats.FlaggedCount, int64(1))
		assert.GreaterOrEqual(t, body.Stats.AssessmentCount, int64(1))
		assert.Equal(t, 3600, body.Stats.WindowSeconds)
		assert.NotEmpty(t, body.RecentEvents)
		assert.False(t, body.Cursor.IsZero())
	})

	t.Run("journey endpoint returns ordered stages for a log", func(t *testing.T) {
		appUUID := uuid.MustParse(env.AppID)
		logID := insertPipelineTestLog(t, env, connID)
		walkLogThroughPipeline(t, pw, logID, appUUID)

		rr := env.request(t, http.MethodGet, "/api/apps/"+env.AppID+"/pipeline/logs/"+logID.String()+"/journey", nil)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

		var body struct {
			LogID  uuid.UUID        `json:"log_id"`
			Events []map[string]any `json:"events"`
		}
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
		assert.Equal(t, logID, body.LogID)
		require.Len(t, body.Events, 4)
		assert.Equal(t, agent.StageIngestion, body.Events[0]["stage"])
		assert.Equal(t, agent.StageAssessment, body.Events[3]["stage"])
	})

	t.Run("bootstrap cache serves hit on second call", func(t *testing.T) {
		appUUID := uuid.MustParse(env.AppID)
		logID := insertPipelineTestLog(t, env, connID)
		walkLogThroughPipeline(t, pw, logID, appUUID)

		// The test Server is built as a struct literal in testhelpers
		// (not via NewServer), so the bootstrap cache wasn't wired. Wire
		// it lazily here — production main.go always wires it. This
		// also verifies that a raw struct-literal Server can opt into
		// the cache without other changes.
		env.Server.SetPipelineBootstrapCacheForTest()

		// First call: MISS. Second call: HIT with identical body.
		rr1 := env.request(t, http.MethodGet, "/api/apps/"+env.AppID+"/pipeline/bootstrap?window=1h&tickerLimit=10", nil)
		require.Equal(t, http.StatusOK, rr1.Code)
		assert.Equal(t, "MISS", rr1.Header().Get("X-Cache"))
		body1 := rr1.Body.Bytes()

		rr2 := env.request(t, http.MethodGet, "/api/apps/"+env.AppID+"/pipeline/bootstrap?window=1h&tickerLimit=10", nil)
		require.Equal(t, http.StatusOK, rr2.Code)
		assert.Equal(t, "HIT", rr2.Header().Get("X-Cache"))
		assert.Equal(t, body1, rr2.Body.Bytes(), "cache hit must return the exact payload")

		// Different tickerLimit is a different cache key → MISS.
		rr3 := env.request(t, http.MethodGet, "/api/apps/"+env.AppID+"/pipeline/bootstrap?window=1h&tickerLimit=5", nil)
		require.Equal(t, http.StatusOK, rr3.Code)
		assert.Equal(t, "MISS", rr3.Header().Get("X-Cache"))
	})

	t.Run("RLS: non-owner roles see zero pipeline rows", func(t *testing.T) {
		// Phase 3.5: log_pipeline_events follows the system-table pattern
		// from migrations 030 / 033 / 038 — RLS enabled with no
		// policies, so non-owner roles (anon, authenticated) see zero
		// rows by construction. Guard against a future migration that
		// accidentally disables RLS or adds a permissive policy.
		appUUID := uuid.MustParse(env.AppID)
		logID := insertPipelineTestLog(t, env, connID)
		walkLogThroughPipeline(t, pw, logID, appUUID)

		// Verify RLS is enabled on the table.
		var rlsEnabled bool
		require.NoError(t, env.Pool.QueryRow(context.Background(),
			`SELECT relrowsecurity FROM pg_class WHERE relname = 'log_pipeline_events'`).Scan(&rlsEnabled))
		assert.True(t, rlsEnabled, "RLS must be enabled on log_pipeline_events")

		// Verify there are no policies granting read access — matches
		// migrations 030 and 033's "enabled, no policies" posture.
		var policyCount int
		require.NoError(t, env.Pool.QueryRow(context.Background(),
			`SELECT count(*) FROM pg_policies WHERE tablename = 'log_pipeline_events'`).Scan(&policyCount))
		assert.Equal(t, 0, policyCount, "no policies should grant read access")

		// Simulate a non-owner connection by temporarily assuming the
		// `authenticated` role within a transaction and asserting the
		// table returns zero rows under that role. SET LOCAL ROLE
		// respects RLS; the owner pool bypasses it, so this isolates
		// the RLS behaviour without a whole new pool.
		tx, err := env.Pool.Begin(context.Background())
		require.NoError(t, err)
		defer tx.Rollback(context.Background())

		_, err = tx.Exec(context.Background(), "SET LOCAL ROLE authenticated")
		require.NoError(t, err)

		var visible int
		require.NoError(t, tx.QueryRow(context.Background(),
			`SELECT count(*) FROM log_pipeline_events WHERE log_id = $1`, logID).Scan(&visible))
		assert.Equal(t, 0, visible, "authenticated role must see zero rows under RLS")
	})

	t.Run("logs picker dedupes per log and respects window", func(t *testing.T) {
		// Phase 4.1: the picker collapses a log's ≤4 stage events into a
		// single summary row. Walk three logs through the full pipeline,
		// then hit /pipeline/logs and assert one row per log with the
		// expected summary fields populated and rows sorted newest-first.
		appUUID := uuid.MustParse(env.AppID)

		// Sentinel — walk through the pipeline BEFORE the assertion
		// window. Used to prove the `?since=` lower bound filters it out.
		oldLogID := insertPipelineTestLog(t, env, connID)
		walkLogThroughPipeline(t, pw, oldLogID, appUUID)
		_, err := env.Pool.Exec(context.Background(),
			`UPDATE log_pipeline_events SET occurred_at = now() - interval '10 hours' WHERE log_id = $1`, oldLogID)
		require.NoError(t, err)

		// Three fresh logs in the assertion window.
		freshIDs := []uuid.UUID{}
		for i := 0; i < 3; i++ {
			id := insertPipelineTestLog(t, env, connID)
			walkLogThroughPipeline(t, pw, id, appUUID)
			freshIDs = append(freshIDs, id)
		}

		since := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339Nano)
		until := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339Nano)
		rr := env.request(t, http.MethodGet,
			"/api/apps/"+env.AppID+"/pipeline/logs?since="+since+"&until="+until+"&limit=50", nil)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

		var body struct {
			Logs []struct {
				LogID       uuid.UUID  `json:"log_id"`
				StageCount  int64      `json:"stage_count"`
				SourceType  string     `json:"source_type"`
				Type        string     `json:"type"`
				Escalated   *bool      `json:"escalated"`
				RuleHit     string     `json:"rule_hit"`
				AssessmentID *uuid.UUID `json:"assessment_id"`
			} `json:"logs"`
			Truncated bool `json:"truncated"`
		}
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))

		// One row per fresh log; the out-of-window log filtered out.
		freshSet := map[uuid.UUID]bool{}
		for _, id := range freshIDs {
			freshSet[id] = true
		}
		seen := 0
		for _, row := range body.Logs {
			if freshSet[row.LogID] {
				seen++
				assert.Equal(t, int64(4), row.StageCount, "each log walked through all four stages")
				assert.Equal(t, "test", row.SourceType)
				assert.Equal(t, "ERROR", row.Type)
				require.NotNil(t, row.Escalated)
				assert.True(t, *row.Escalated)
				assert.Equal(t, agent.RuleErrorType, row.RuleHit)
				assert.NotNil(t, row.AssessmentID)
			}
			assert.NotEqual(t, oldLogID, row.LogID, "out-of-window log must be filtered")
		}
		assert.Equal(t, 3, seen, "all three fresh logs should appear exactly once")
	})

	t.Run("logs picker refuses cross-org app", func(t *testing.T) {
		// A user hitting /pipeline/logs with someone else's appId must
		// 404 via authorizeApp, same posture as /pipeline/bootstrap. The
		// test re-uses the foreign app provisioned in the cross-app
		// journey subtest — that subtest also runs before this one in
		// source order, so `foreign-app` already exists in the fixture.
		foreignApp := uuid.New()
		foreignOrg := uuid.New()
		_, err := env.Pool.Exec(context.Background(),
			"INSERT INTO organizations (id, name, slug) VALUES ($1, 'foreign-picker', 'foreign-picker-"+foreignOrg.String()[:8]+"')", foreignOrg)
		require.NoError(t, err)
		t.Cleanup(func() {
			env.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", foreignOrg)
		})
		_, err = env.Pool.Exec(context.Background(),
			"INSERT INTO applications (id, org_id, name, status) VALUES ($1, $2, 'foreign-picker', 'active')",
			foreignApp, foreignOrg,
		)
		require.NoError(t, err)

		rr := env.request(t, http.MethodGet, "/api/apps/"+foreignApp.String()+"/pipeline/logs?limit=10", nil)
		assert.Equal(t, http.StatusNotFound, rr.Code, "cross-org app must 404, not leak empty 200")
	})

	t.Run("journey refuses cross-app log", func(t *testing.T) {
		// Create a second app in a different org to simulate a
		// neighbouring tenant. A request for this app's log through the
		// test user's app route must 404, not leak journey rows.
		foreignOrg := uuid.New()
		foreignApp := uuid.New()
		foreignUser := uuid.New()
		foreignEmail := "foreign-" + foreignUser.String() + "@heimdall.test"
		_, err := env.Pool.Exec(context.Background(),
			`INSERT INTO auth.users (id, email, instance_id, aud, role, encrypted_password, confirmation_token, created_at, updated_at)
			 VALUES ($1, $2, '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', '', '', now(), now())`,
			foreignUser, foreignEmail,
		)
		require.NoError(t, err)
		t.Cleanup(func() {
			env.Pool.Exec(context.Background(), "DELETE FROM auth.users WHERE id = $1", foreignUser)
		})
		_, err = env.Pool.Exec(context.Background(), "INSERT INTO users (id, email) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING", foreignUser, foreignEmail)
		require.NoError(t, err)
		_, err = env.Pool.Exec(context.Background(), "INSERT INTO organizations (id, name, slug) VALUES ($1, 'foreign', 'foreign-"+foreignOrg.String()[:8]+"')", foreignOrg)
		require.NoError(t, err)
		t.Cleanup(func() {
			env.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", foreignOrg)
		})
		_, err = env.Pool.Exec(context.Background(), "INSERT INTO org_members (user_id, org_id, role) VALUES ($1, $2, 'owner')", foreignUser, foreignOrg)
		require.NoError(t, err)
		_, err = env.Pool.Exec(context.Background(),
			"INSERT INTO applications (id, org_id, name, status) VALUES ($1, $2, 'foreign-app', 'active')",
			foreignApp, foreignOrg,
		)
		require.NoError(t, err)

		// Insert a foreign-app log + walk it through the pipeline.
		foreignConnID := uuid.New()
		_, err = env.Pool.Exec(context.Background(),
			`INSERT INTO connections (id, user_id, org_id, app_id, name, type, status, config)
			 VALUES ($1, $2, $3, $4, 'foreign', 'webhook_logs', 'active', '{}'::jsonb)`,
			foreignConnID, foreignUser, foreignOrg, foreignApp,
		)
		require.NoError(t, err)
		foreignLogRow, err := env.Queries.InsertLogEntry(context.Background(), db.InsertLogEntryParams{
			ConnectionID: foreignConnID,
			SourceType:   "test",
			Payload:      []byte(`{}`),
			UserID:       foreignUser,
			AppID:        foreignApp,
		})
		require.NoError(t, err)
		walkLogThroughPipeline(t, pw, foreignLogRow.ID, foreignApp)

		// Test user hitting their own app's route with the foreign log
		// must 404 — app_id mismatch caught by the defence-in-depth
		// check in PipelineJourney.
		rr := env.request(t, http.MethodGet, "/api/apps/"+env.AppID+"/pipeline/logs/"+foreignLogRow.ID.String()+"/journey", nil)
		assert.Equal(t, http.StatusNotFound, rr.Code, "cross-app journey lookup must 404")
	})
}
