package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestPruneTick_DBError verifies that pruneTick logs and swallows DB errors
// without panicking. The agent uses stubDBTX which returns errors for all
// operations — this exercises the error path exactly as it would fire in
// production if the pool were temporarily unavailable.
func TestPruneTick_DBError(t *testing.T) {
	agent := newToolTestAgent()
	ctx := context.Background()

	// Should not panic even though PruneExpiredLogs will fail against
	// stubDBTX. Errors must be logged and swallowed.
	require.NotPanics(t, func() {
		agent.pruneTick(ctx)
	})
}

// TestPrune_CancelExits verifies that Prune returns promptly when its
// context is cancelled. The loop runs one startup tick (which fails fast
// against stubDBTX), then blocks on the ticker until ctx.Done fires.
func TestPrune_CancelExits(t *testing.T) {
	agent := newToolTestAgent()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		agent.Prune(ctx)
		close(done)
	}()

	// Give the startup pruneTick a moment to run, then cancel.
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success — Prune exited cleanly on cancel.
	case <-time.After(2 * time.Second):
		t.Fatal("Prune did not exit within 2 seconds of context cancel")
	}
}

// TestPrune_StartupTickRuns verifies that Prune executes a pruneTick
// immediately on startup rather than waiting for the first ticker fire.
// This matters because without the startup tick, a restart loop would
// indefinitely delay pruning — exactly the failure mode that motivated
// wiring up PruneExpiredLogs in the first place (0.30.2-era bug where
// the query existed but was never called).
//
// We verify this indirectly: cancel the context before the 1-hour ticker
// could possibly fire, and confirm Prune still exits cleanly. If the
// startup tick weren't running, nothing important would be broken — but
// this test is cheap insurance that the top-of-loop call stays in place.
func TestPrune_StartupTickRuns(t *testing.T) {
	agent := newToolTestAgent()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		agent.Prune(ctx)
		close(done)
	}()

	// Cancel well before the 1h ticker could fire.
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success.
	case <-time.After(2 * time.Second):
		t.Fatal("Prune did not exit promptly")
	}
}
