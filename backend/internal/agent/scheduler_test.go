package agent

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// ptrTime returns a pointer to t — helper for sqlc-generated *time.Time fields.
func ptrTime(t time.Time) *time.Time { return &t }

// cronText is a small helper to construct pgtype.Text values for cron_expr in
// test fixtures. Keeps the test tables readable compared to inlining
// `pgtype.Text{String: ..., Valid: true}` on every row.
func cronText(s string) pgtype.Text { return pgtype.Text{String: s, Valid: true} }

func TestShouldFire_NeverRun(t *testing.T) {
	// A schedule with LastRunAt == nil should fire immediately regardless
	// of interval_secs — it has never executed, so "elapsed since last run"
	// is effectively infinite.
	s := db.InvestigationSchedule{
		IntervalSecs: 300,
		LastRunAt:    nil,
	}
	assert.True(t, shouldFire(s, time.Now()))
}

func TestShouldFire_IntervalNotYetElapsed(t *testing.T) {
	// LastRunAt was 30 seconds ago, interval is 60 seconds → not due yet.
	now := time.Now()
	s := db.InvestigationSchedule{
		IntervalSecs: 60,
		LastRunAt:    ptrTime(now.Add(-30 * time.Second)),
	}
	assert.False(t, shouldFire(s, now))
}

func TestShouldFire_IntervalJustElapsed(t *testing.T) {
	// LastRunAt was exactly 60 seconds ago, interval is 60 → due.
	// This tests the boundary condition — >= not >.
	now := time.Now()
	s := db.InvestigationSchedule{
		IntervalSecs: 60,
		LastRunAt:    ptrTime(now.Add(-60 * time.Second)),
	}
	assert.True(t, shouldFire(s, now))
}

func TestShouldFire_IntervalLongPast(t *testing.T) {
	// A schedule that should have fired many times but didn't (e.g. because
	// it was disabled and re-enabled, or the server was down) fires on the
	// next tick — no make-up runs.
	now := time.Now()
	s := db.InvestigationSchedule{
		IntervalSecs: 60,
		LastRunAt:    ptrTime(now.Add(-2 * time.Hour)),
	}
	assert.True(t, shouldFire(s, now))
}

func TestSchedulerTick_DBError(t *testing.T) {
	// stubDBTX returns errors for all DB operations, so ListEnabledSchedules
	// will fail. schedulerTick must log and return cleanly without panicking.
	agent := newToolTestAgent()
	ctx := context.Background()

	require.NotPanics(t, func() {
		agent.schedulerTick(ctx)
	})
}

func TestInvestigationScheduler_CancelExits(t *testing.T) {
	// Verify the loop exits promptly when ctx is cancelled, mirroring the
	// Monitor/Prune shutdown contract. Without this, Stop() would hang on
	// a.wg.Wait() and the whole agent would leak on restart.
	agent := newToolTestAgent()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		agent.InvestigationScheduler(ctx)
		close(done)
	}()

	// Give the goroutine a moment to enter its select loop, then cancel.
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success.
	case <-time.After(2 * time.Second):
		t.Fatal("InvestigationScheduler did not exit within 2 seconds")
	}
}

// --- Cron-path tests (Phase 4) ---------------------------------------------
//
// These exercise shouldFire's cron branch. The goal is to verify three
// behaviors: (1) cron_expr takes precedence over interval_secs when both
// happen to be set; (2) a due cron schedule fires exactly when its computed
// next-run is at or before now; (3) a malformed cron_expr is logged and
// returns false without panicking (no infinite error loop on bad input).

func TestShouldFire_CronDue(t *testing.T) {
	// "*/5 * * * *" fires every 5 minutes. Last run was 6 minutes ago, so the
	// next slot (one 5-minute tick after last_run_at) is in the past.
	now := time.Now()
	s := db.InvestigationSchedule{
		IntervalSecs: 0, // unused when cron_expr is set
		CronExpr:     cronText("*/5 * * * *"),
		LastRunAt:    ptrTime(now.Add(-6 * time.Minute)),
	}
	assert.True(t, shouldFire(s, now))
}

func TestShouldFire_CronNotYetDue(t *testing.T) {
	// Hourly cron ("0 * * * *") fires at the top of every hour. If last_run
	// was 10 minutes ago and we're currently at :25, the next slot is :00
	// of the next hour — still in the future, don't fire.
	now := time.Date(2026, 4, 11, 14, 25, 0, 0, time.UTC)
	s := db.InvestigationSchedule{
		IntervalSecs: 0,
		CronExpr:     cronText("0 * * * *"),
		LastRunAt:    ptrTime(now.Add(-10 * time.Minute)), // 14:15
	}
	assert.False(t, shouldFire(s, now))
}

func TestShouldFire_CronPrefersCronOverInterval(t *testing.T) {
	// Both fields set with conflicting verdicts:
	//   - interval_secs=60 says "fire if >= 60s elapsed" (30s elapsed → would NOT fire)
	//   - cron "*/1 * * * *" says "fire at every minute boundary" (crossed → WOULD fire)
	// shouldFire must pick the cron branch, so the result is true.
	// This guards the documented precedence rule: cron wins when set.
	now := time.Date(2026, 4, 11, 14, 25, 30, 0, time.UTC)
	s := db.InvestigationSchedule{
		IntervalSecs: 60,
		CronExpr:     cronText("*/1 * * * *"),
		LastRunAt:    ptrTime(now.Add(-35 * time.Second)), // 14:24:55
	}
	assert.True(t, shouldFire(s, now))
}

func TestShouldFire_CronInvalidExpression(t *testing.T) {
	// Malformed cron_expr must NOT panic and must NOT fire (a stuck schedule
	// is preferable to an infinite retry loop). The error is logged inside
	// shouldFire — we just assert the return value and the non-panic.
	now := time.Now()
	s := db.InvestigationSchedule{
		ID:           uuid.New(),
		IntervalSecs: 0,
		CronExpr:     cronText("not a cron expression"),
		LastRunAt:    ptrTime(now.Add(-1 * time.Hour)),
	}
	assert.False(t, shouldFire(s, now))
}

func TestShouldFire_CronNeverRun(t *testing.T) {
	// Never-run schedules fire immediately regardless of cron_expr — the
	// "has last_run_at ever been set?" check short-circuits before we
	// touch the parser at all.
	s := db.InvestigationSchedule{
		IntervalSecs: 0,
		CronExpr:     cronText("0 9 * * *"), // "every day at 09:00"
		LastRunAt:    nil,
	}
	assert.True(t, shouldFire(s, time.Now()))
}

func TestParseCronExpression_Valid(t *testing.T) {
	// Sanity-check the exported wrapper the HTTP handler uses. If this ever
	// stops accepting a plain 5-field expression we want to know at the
	// unit-test level, not after a deploy.
	sched, err := ParseCronExpression("*/5 * * * *")
	require.NoError(t, err)
	require.NotNil(t, sched)

	_, err = ParseCronExpression("0 9 * * *")
	assert.NoError(t, err)

	_, err = ParseCronExpression("garbage")
	assert.Error(t, err)
}

func TestRunScheduledInvestigation_DBErrorMarksRunError(t *testing.T) {
	// Against stubDBTX, GetApplication fails. runScheduledInvestigation
	// should log, call markRunError, and return cleanly without panicking.
	// markRunError itself is best-effort (also swallows errors), so the
	// whole path must survive even when the DB is completely broken.
	agent := newToolTestAgent()
	ctx := context.Background()

	s := db.InvestigationSchedule{
		ID:           uuid.New(),
		AppID:        uuid.New(),
		Name:         "test schedule",
		Prompt:       "test prompt",
		IntervalSecs: 60,
		Enabled:      true,
	}

	require.NotPanics(t, func() {
		agent.RunScheduledInvestigation(ctx, s)
	})
}
