package connectors

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// mockPollConnector counts how many times Poll is called.
type mockPollConnector struct {
	pollCount atomic.Int32
}

func (m *mockPollConnector) Connect(_ context.Context) error  { return nil }
func (m *mockPollConnector) Health(_ context.Context) error   { return nil }
func (m *mockPollConnector) Close() error                     { return nil }
func (m *mockPollConnector) Poll(_ context.Context, _ *db.Queries) error {
	m.pollCount.Add(1)
	return nil
}

func TestPoller_StartStop(t *testing.T) {
	p := NewPoller(nil)
	mock := &mockPollConnector{}
	connID := uuid.New()

	p.Start(mock, connID, uuid.Nil, 10*time.Millisecond)

	// Wait for at least one poll.
	time.Sleep(50 * time.Millisecond)
	count := mock.pollCount.Load()
	require.Greater(t, count, int32(0), "expected at least one poll")

	p.Stop(connID)

	// Record count after stop, wait, and verify no more polls.
	countAfterStop := mock.pollCount.Load()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, countAfterStop, mock.pollCount.Load(), "expected no more polls after stop")
}

func TestPoller_StopAll(t *testing.T) {
	p := NewPoller(nil)
	mock1 := &mockPollConnector{}
	mock2 := &mockPollConnector{}
	id1 := uuid.New()
	id2 := uuid.New()

	p.Start(mock1, id1, uuid.Nil, 10*time.Millisecond)
	p.Start(mock2, id2, uuid.Nil, 10*time.Millisecond)

	time.Sleep(50 * time.Millisecond)
	require.Greater(t, mock1.pollCount.Load(), int32(0))
	require.Greater(t, mock2.pollCount.Load(), int32(0))

	p.StopAll()

	count1 := mock1.pollCount.Load()
	count2 := mock2.pollCount.Load()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, count1, mock1.pollCount.Load())
	assert.Equal(t, count2, mock2.pollCount.Load())
}

func TestPoller_StartReplacesExisting(t *testing.T) {
	p := NewPoller(nil)
	mock1 := &mockPollConnector{}
	mock2 := &mockPollConnector{}
	connID := uuid.New()

	p.Start(mock1, connID, uuid.Nil, 10*time.Millisecond)
	time.Sleep(30 * time.Millisecond)

	// Replace with a new connector on the same ID.
	p.Start(mock2, connID, uuid.Nil, 10*time.Millisecond)
	time.Sleep(30 * time.Millisecond)

	// mock1 should have stopped; record its count.
	count1 := mock1.pollCount.Load()
	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, count1, mock1.pollCount.Load(), "old connector should have stopped")

	// mock2 should be actively polling.
	assert.Greater(t, mock2.pollCount.Load(), int32(0), "new connector should be polling")

	p.StopAll()
}

func TestPoller_StopNonExistent(t *testing.T) {
	p := NewPoller(nil)
	// Should not panic.
	p.Stop(uuid.New())
}
