package handlers

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestSSEConnLimiter_CapsPerUser verifies the per-user cap blocks the
// (cap+1)-th acquire, and that Release frees a slot for the same user.
func TestSSEConnLimiter_CapsPerUser(t *testing.T) {
	l := newSSEConnLimiter(2)
	user := uuid.New()

	assert.True(t, l.Acquire(user), "1st acquire under cap")
	assert.True(t, l.Acquire(user), "2nd acquire at cap ceiling")
	assert.False(t, l.Acquire(user), "3rd acquire must be rejected")
	assert.Equal(t, 2, l.Count(user))

	l.Release(user)
	assert.Equal(t, 1, l.Count(user))
	assert.True(t, l.Acquire(user), "after Release, a slot opens back up")
}

// TestSSEConnLimiter_PerUserIsolation asserts one user's cap doesn't bleed
// into another's — a runaway tab from user A must not block user B.
func TestSSEConnLimiter_PerUserIsolation(t *testing.T) {
	l := newSSEConnLimiter(1)
	a := uuid.New()
	b := uuid.New()

	assert.True(t, l.Acquire(a))
	assert.False(t, l.Acquire(a), "user A capped")
	assert.True(t, l.Acquire(b), "user B independent")
	assert.Equal(t, 1, l.Count(a))
	assert.Equal(t, 1, l.Count(b))
}

// TestSSEConnLimiter_ReleaseIdempotentUnderflow verifies a stray Release
// with no matching Acquire doesn't corrupt the counter. Protects against
// a miscount permanently locking out a user.
func TestSSEConnLimiter_ReleaseIdempotentUnderflow(t *testing.T) {
	l := newSSEConnLimiter(3)
	user := uuid.New()

	l.Release(user) // no Acquire yet — must not go negative
	assert.Equal(t, 0, l.Count(user))

	assert.True(t, l.Acquire(user))
	assert.Equal(t, 1, l.Count(user))
	l.Release(user)
	l.Release(user) // extra release
	assert.Equal(t, 0, l.Count(user))

	// User still able to acquire after stray releases.
	assert.True(t, l.Acquire(user))
}

// TestBootstrapCache_HitAndMiss verifies Get returns a stored payload
// before TTL and nil after.
func TestBootstrapCache_HitAndMiss(t *testing.T) {
	c := newBootstrapCache()
	key := bootstrapCacheKey{
		AppID:       uuid.New(),
		Window:      time.Hour,
		TickerLimit: 50,
	}
	payload := []byte(`{"stats":{}}`)

	assert.Nil(t, c.Get(key), "miss on empty cache")
	c.Put(key, payload)
	assert.Equal(t, payload, c.Get(key), "hit on fresh entry")
}

// TestBootstrapCache_KeyIsolation asserts entries are segregated by the
// full (appID, window, tickerLimit) tuple — a cross-tenant hit would be
// a deploy-blocker.
func TestBootstrapCache_KeyIsolation(t *testing.T) {
	c := newBootstrapCache()
	appA := uuid.New()
	appB := uuid.New()
	keyA := bootstrapCacheKey{AppID: appA, Window: time.Hour, TickerLimit: 50}
	keyB := bootstrapCacheKey{AppID: appB, Window: time.Hour, TickerLimit: 50}
	keyAWide := bootstrapCacheKey{AppID: appA, Window: 6 * time.Hour, TickerLimit: 50}
	keyABig := bootstrapCacheKey{AppID: appA, Window: time.Hour, TickerLimit: 200}

	c.Put(keyA, []byte(`"A-1h-50"`))
	assert.Nil(t, c.Get(keyB), "different appID — miss")
	assert.Nil(t, c.Get(keyAWide), "different window — miss")
	assert.Nil(t, c.Get(keyABig), "different tickerLimit — miss")
	assert.Equal(t, []byte(`"A-1h-50"`), c.Get(keyA))
}

// TestBootstrapCache_ExpiredEntryEvicted verifies the TTL path: a stale
// entry returns nil AND is removed from the map (so the next Put reuses
// the slot cleanly rather than overwriting).
func TestBootstrapCache_ExpiredEntryEvicted(t *testing.T) {
	c := newBootstrapCache()
	key := bootstrapCacheKey{AppID: uuid.New(), Window: time.Hour, TickerLimit: 50}

	// Insert with an expiry in the past — bypass Put so we can control
	// the timestamp precisely without sleeping the whole TTL window.
	c.mu.Lock()
	c.entries[key] = bootstrapCacheEntry{
		payload:   []byte(`"stale"`),
		expiresAt: time.Now().Add(-time.Second),
	}
	c.mu.Unlock()

	assert.Nil(t, c.Get(key), "expired entry must not be returned")

	c.mu.Lock()
	_, present := c.entries[key]
	c.mu.Unlock()
	assert.False(t, present, "expired entry must be evicted on Get")
}
