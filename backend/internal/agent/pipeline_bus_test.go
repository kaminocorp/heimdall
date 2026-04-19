package agent

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPipelineBus_Fanout asserts every subscriber for the app receives the
// event exactly once, and subscribers for a different app see nothing.
func TestPipelineBus_Fanout(t *testing.T) {
	bus := NewPipelineBus()
	appA := uuid.New()
	appB := uuid.New()

	chA1, unsubA1 := bus.Subscribe(appA)
	chA2, unsubA2 := bus.Subscribe(appA)
	chB, unsubB := bus.Subscribe(appB)
	defer unsubA1()
	defer unsubA2()
	defer unsubB()

	evt := PipelineEvent{ID: uuid.New(), AppID: appA, Stage: StageIngestion}
	bus.Publish(evt)

	for i, ch := range []<-chan PipelineEvent{chA1, chA2} {
		select {
		case got := <-ch:
			assert.Equal(t, evt.ID, got.ID, "subscriber %d received wrong event", i)
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d timed out waiting for event", i)
		}
	}

	// B must not see an event destined for A.
	select {
	case got := <-chB:
		t.Fatalf("app-B subscriber received cross-app event: %+v", got)
	case <-time.After(50 * time.Millisecond):
		// expected
	}
}

// TestPipelineBus_DropsWhenFull asserts that a slow subscriber drops
// overflow events without blocking the publisher, and that the drop
// counter reflects the overflow count.
func TestPipelineBus_DropsWhenFull(t *testing.T) {
	bus := NewPipelineBus()
	appID := uuid.New()

	ch, unsub := bus.Subscribe(appID)
	defer unsub()

	// Fill the subscriber's buffer without draining.
	for i := 0; i < pipelineBusBufferSize; i++ {
		bus.Publish(PipelineEvent{ID: uuid.New(), AppID: appID, Stage: StageIngestion})
	}
	assert.Equal(t, uint64(0), bus.Dropped(), "no drops while buffer had space")

	// Next N publishes must each be dropped, never block.
	done := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			bus.Publish(PipelineEvent{ID: uuid.New(), AppID: appID, Stage: StageGate})
		}
		close(done)
	}()
	select {
	case <-done:
		// Publishing never blocked despite full buffer.
	case <-time.After(time.Second):
		t.Fatal("Publish blocked on full subscriber channel")
	}
	assert.Equal(t, uint64(10), bus.Dropped(), "drop counter must match overflow")

	// Buffer still has the original events — drain one to prove the
	// channel survived the overflow intact.
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("subscriber buffer lost its queued events after overflow")
	}
}

// TestPipelineBus_UnsubscribeCleanup asserts that unsubscribing removes the
// channel, closes it, and lets the per-app entry be garbage-collected
// when the last subscriber leaves.
func TestPipelineBus_UnsubscribeCleanup(t *testing.T) {
	bus := NewPipelineBus()
	appID := uuid.New()

	ch, unsub := bus.Subscribe(appID)
	require.Equal(t, 1, bus.SubscriberCount(appID))

	unsub()
	assert.Equal(t, 0, bus.SubscriberCount(appID))

	// Channel must be closed — receive returns zero value with ok=false.
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "channel should be closed after unsubscribe")
	case <-time.After(time.Second):
		t.Fatal("channel not closed after unsubscribe")
	}

	// Publishing to an app with no subscribers is a no-op.
	bus.Publish(PipelineEvent{AppID: appID, Stage: StageGate})
	assert.Equal(t, uint64(0), bus.Dropped())

	// Double-unsubscribe is safe (idempotent).
	unsub()
}

// TestPipelineBus_SubscriberCountsSnapshot asserts SubscriberCounts returns
// a point-in-time snapshot per app — used by the Phase 3.4 leak canary.
func TestPipelineBus_SubscriberCountsSnapshot(t *testing.T) {
	bus := NewPipelineBus()
	appA := uuid.New()
	appB := uuid.New()

	_, unsubA1 := bus.Subscribe(appA)
	_, unsubA2 := bus.Subscribe(appA)
	_, unsubB := bus.Subscribe(appB)

	counts := bus.SubscriberCounts()
	assert.Equal(t, 2, counts[appA])
	assert.Equal(t, 1, counts[appB])

	// Snapshot is owned by the caller — mutating it must not poison
	// the bus's internal state.
	counts[appA] = 999
	again := bus.SubscriberCounts()
	assert.Equal(t, 2, again[appA], "snapshot mutation leaked into bus")

	unsubA1()
	unsubA2()
	unsubB()
	assert.Empty(t, bus.SubscriberCounts(), "empty after all unsubscribe")
}

// TestPipelineBus_ConcurrentSubscribeUnsubscribe exercises the mutex
// against parallel subscribers to surface any races under -race.
func TestPipelineBus_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	bus := NewPipelineBus()
	appID := uuid.New()

	const workers = 16
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				ch, unsub := bus.Subscribe(appID)
				bus.Publish(PipelineEvent{AppID: appID, Stage: StageClassified})
				// drain at most one event so the buffer doesn't fill;
				// dropped events are fine for this test's purposes.
				select {
				case <-ch:
				default:
				}
				unsub()
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, 0, bus.SubscriberCount(appID), "all subscribers cleaned up after test")
}
