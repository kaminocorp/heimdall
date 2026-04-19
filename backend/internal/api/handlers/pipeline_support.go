package handlers

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/metrics"
)

// Phase 3 hardening helpers for the pipeline page:
//   - sseConnLimiter caps concurrent PipelineStream connections per user so a
//     runaway tab can't pin unbounded backend channels.
//   - bootstrapCache memoises /pipeline/bootstrap responses for a short TTL
//     so repeated page loads don't re-run the stats aggregate back-to-back.

// pipelineSSEPerUserCap bounds the number of concurrent SSE subscriptions
// a single user may hold. Five matches the plan (Phase 3.2) and comfortably
// covers "one tab per app with one or two duplicates" while closing the
// door on runaway subscribe loops. Exceeding returns 429.
const pipelineSSEPerUserCap = 5

// sseConnLimiter is a simple per-user connection counter. Concurrency-safe;
// the counters live in-process only — multi-instance deployments (if we
// ever scale past one backend) would need a shared store, but as of 0.46
// the pod count is one.
type sseConnLimiter struct {
	mu    sync.Mutex
	conns map[uuid.UUID]int
	cap   int
}

func newSSEConnLimiter(perUserCap int) *sseConnLimiter {
	return &sseConnLimiter{
		conns: make(map[uuid.UUID]int),
		cap:   perUserCap,
	}
}

// Acquire tries to reserve a slot for userID. Returns true on success
// (caller MUST defer Release); false when the user is already at the cap.
// A rejected acquire increments the rejection metric so the dashboard can
// show it distinctly from drops / crashes.
func (l *sseConnLimiter) Acquire(userID uuid.UUID) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.conns[userID] >= l.cap {
		metrics.PipelineSSERejectedTotal.Inc()
		return false
	}
	l.conns[userID]++
	metrics.PipelineSSEActiveConnections.Inc()
	return true
}

// Release frees the slot. Safe to double-release? No — paired with Acquire.
// We guard against underflow defensively because a miscount would corrupt
// the cap for that user for the life of the process.
func (l *sseConnLimiter) Release(userID uuid.UUID) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.conns[userID] > 0 {
		l.conns[userID]--
		metrics.PipelineSSEActiveConnections.Dec()
	}
	if l.conns[userID] == 0 {
		delete(l.conns, userID)
	}
}

// Count returns the current held slot count for a user. Test-only hook;
// production code shouldn't branch on this.
func (l *sseConnLimiter) Count(userID uuid.UUID) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.conns[userID]
}

// bootstrapCacheTTL is the per-entry freshness window. Five seconds
// matches the plan (Phase 3.3) — long enough that a user opening the page
// twice in quick succession (React/Vue dev HMR, navigation back-forward)
// hits the cache, short enough that the stats block never feels stale.
const bootstrapCacheTTL = 5 * time.Second

// bootstrapCacheKey tags an entry with every input that changes the
// response. `window` and `tickerLimit` are the two query-string knobs;
// omitting either from the key would cross-pollute responses that share
// the same appID but different query shapes.
type bootstrapCacheKey struct {
	AppID       uuid.UUID
	Window      time.Duration
	TickerLimit int32
}

type bootstrapCacheEntry struct {
	payload   []byte // pre-marshalled JSON so cache hits skip json.Encoder
	expiresAt time.Time
}

// bootstrapCache is a tiny fixed-TTL cache over /pipeline/bootstrap
// responses. Entries age out lazily on Get; a background sweeper isn't
// worth it at this cache's size (one entry per active user-tab in a 5s
// window). Size grows, but every entry older than its TTL is logically
// dead and Get returns nothing — the next Put overwrites.
type bootstrapCache struct {
	mu      sync.Mutex
	entries map[bootstrapCacheKey]bootstrapCacheEntry
}

func newBootstrapCache() *bootstrapCache {
	return &bootstrapCache{entries: make(map[bootstrapCacheKey]bootstrapCacheEntry)}
}

// Get returns a fresh cached payload, or nil when absent / expired.
// Expired entries are deleted in-place so the map can shrink between
// misses.
func (c *bootstrapCache) Get(key bootstrapCacheKey) []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		metrics.PipelineBootstrapCache.WithLabelValues("miss").Inc()
		return nil
	}
	if time.Now().After(entry.expiresAt) {
		delete(c.entries, key)
		metrics.PipelineBootstrapCache.WithLabelValues("miss").Inc()
		return nil
	}
	metrics.PipelineBootstrapCache.WithLabelValues("hit").Inc()
	return entry.payload
}

// Put stores a freshly-marshalled payload under the key. We pre-marshal
// into `[]byte` so cache hits can `w.Write(payload)` directly without
// re-running json encoding on every hot request.
func (c *bootstrapCache) Put(key bootstrapCacheKey, payload []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = bootstrapCacheEntry{
		payload:   payload,
		expiresAt: time.Now().Add(bootstrapCacheTTL),
	}
}
