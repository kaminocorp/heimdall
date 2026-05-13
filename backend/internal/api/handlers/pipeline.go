package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// pipelineEventJSON is the stable wire shape for frontend consumption.
// Mirrors log_pipeline_events columns but flattens pgtype.Text / pgtype.Bool
// / pgtype.Float8 into plain Go zero-value semantics the frontend can read
// without special-case null handling. Fields that aren't meaningful for a
// given stage are omitted via `omitempty`.
type pipelineEventJSON struct {
	ID           uuid.UUID  `json:"id"`
	LogID        uuid.UUID  `json:"log_id"`
	AppID        uuid.UUID  `json:"app_id"`
	Stage        string     `json:"stage"`
	OccurredAt   time.Time  `json:"occurred_at"`
	SourceType   string     `json:"source_type,omitempty"`
	Severity     string     `json:"severity,omitempty"`
	Type         string     `json:"type,omitempty"`
	Category     string     `json:"category,omitempty"`
	Confidence   float64    `json:"confidence,omitempty"`
	Summary      string     `json:"summary,omitempty"`
	Escalated    *bool      `json:"escalated,omitempty"`
	RuleHit      string     `json:"rule_hit,omitempty"`
	AssessmentID *uuid.UUID `json:"assessment_id,omitempty"`
}

func rowToJSON(row db.LogPipelineEvent) pipelineEventJSON {
	out := pipelineEventJSON{
		ID:           row.ID,
		LogID:        row.LogID,
		AppID:        row.AppID,
		Stage:        row.Stage,
		OccurredAt:   row.OccurredAt,
		AssessmentID: row.AssessmentID,
	}
	if row.SourceType.Valid {
		out.SourceType = row.SourceType.String
	}
	if row.Severity.Valid {
		out.Severity = row.Severity.String
	}
	if row.Type.Valid {
		out.Type = row.Type.String
	}
	if row.Category.Valid {
		out.Category = row.Category.String
	}
	if row.Confidence.Valid {
		out.Confidence = row.Confidence.Float64
	}
	if row.Summary.Valid {
		out.Summary = row.Summary.String
	}
	if row.Escalated.Valid {
		v := row.Escalated.Bool
		out.Escalated = &v
	}
	if row.RuleHit.Valid {
		out.RuleHit = row.RuleHit.String
	}
	return out
}

func busEventToJSON(evt agent.PipelineEvent) pipelineEventJSON {
	out := pipelineEventJSON{
		ID:           evt.ID,
		LogID:        evt.LogID,
		AppID:        evt.AppID,
		Stage:        evt.Stage,
		OccurredAt:   evt.OccurredAt,
		SourceType:   evt.SourceType,
		Severity:     evt.Severity,
		Type:         evt.Type,
		Category:     evt.Category,
		Confidence:   evt.Confidence,
		Summary:      evt.Summary,
		RuleHit:      evt.RuleHit,
		AssessmentID: evt.AssessmentID,
	}
	if evt.HasGate {
		v := evt.Escalated
		out.Escalated = &v
	}
	return out
}

type pipelineStatsJSON struct {
	IngestionCount  int64   `json:"ingestion_count"`
	ClassifiedCount int64   `json:"classified_count"`
	FlaggedCount    int64   `json:"flagged_count"`
	SafeCount       int64   `json:"safe_count"`
	AssessmentCount int64   `json:"assessment_count"`
	AvgConfidence   float64 `json:"avg_confidence"`
	WindowSeconds   int     `json:"window_seconds"`
}

type bootstrapResponse struct {
	Stats        pipelineStatsJSON   `json:"stats"`
	RecentEvents []pipelineEventJSON `json:"recent_events"`
	Cursor       time.Time           `json:"cursor"`
}

const (
	bootstrapDefaultWindow = time.Hour
	bootstrapMaxWindow     = 24 * time.Hour
	bootstrapDefaultTicker = 50
	bootstrapMaxTicker     = 200
	streamReplayCap        = 500
	streamHeartbeat        = 15 * time.Second

	// Time Machine picker bounds. Max range matches the log_buffer retention
	// ceiling (48h today); querying past that returns empty by construction
	// because ON DELETE CASCADE sheds pipeline events with their parent log.
	logsPickerMaxRange = 48 * time.Hour
	logsPickerDefault  = 24 * time.Hour
	logsPickerMaxLimit = 200
	logsPickerDefLimit = 50
)

// parseWindow reads `?window=` as a Go duration (e.g. "1h", "15m"). Falls
// back to the default when missing or malformed; caps at bootstrapMaxWindow
// so clients can't DOS the stats query with a 30-day window.
func parseWindow(raw string) time.Duration {
	if raw == "" {
		return bootstrapDefaultWindow
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return bootstrapDefaultWindow
	}
	if d > bootstrapMaxWindow {
		return bootstrapMaxWindow
	}
	return d
}

func parseTickerLimit(raw string) int32 {
	if raw == "" {
		return bootstrapDefaultTicker
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return bootstrapDefaultTicker
	}
	if n > bootstrapMaxTicker {
		return bootstrapMaxTicker
	}
	return int32(n)
}

// PipelineBootstrap handles GET /api/apps/{appId}/pipeline/bootstrap.
// Returns stats + recent events + cursor in one roundtrip so the frontend
// can render the funnel and ticker synchronously on first paint.
//
// Phase 3.3: a 5-second in-memory cache keyed by (appID, window, limit)
// elides the DB round-trip for hot reloads. The cache is populated after
// authorisation so the auth check still runs on every request — we never
// return cached data to an unauthorised caller. Tenant isolation is
// preserved because appID is part of the cache key and only a caller
// who passed authorizeApp for that exact appID can reach the cache-read
// branch.
func (s *Server) PipelineBootstrap(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	window := parseWindow(r.URL.Query().Get("window"))
	tickerLimit := parseTickerLimit(r.URL.Query().Get("tickerLimit"))

	cacheKey := bootstrapCacheKey{
		AppID:       app.ID,
		Window:      window,
		TickerLimit: tickerLimit,
	}
	if s.pipelineBootstrap != nil {
		if payload := s.pipelineBootstrap.Get(cacheKey); payload != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			_, _ = w.Write(payload)
			return
		}
	}

	since := time.Now().Add(-window)
	userID, _ := middleware.UserIDFromContext(r.Context())

	// Both reads share one UserQueries scope so they see the same RLS
	// context and snapshot. The bootstrap endpoint is a single
	// dashboard-paint round-trip, so per-call transaction overhead is
	// negligible relative to the LLM-classification work it surfaces.
	var stats db.PipelineStatsByAppRow
	var recent []db.LogPipelineEvent
	err := s.Pools.WithUserQueries(r.Context(), userID, func(q *db.Queries) error {
		var e error
		stats, e = q.PipelineStatsByApp(r.Context(), db.PipelineStatsByAppParams{
			AppID:      app.ID,
			OccurredAt: since,
		})
		if e != nil {
			return e
		}
		recent, e = q.GetRecentPipelineEventsByApp(r.Context(), db.GetRecentPipelineEventsByAppParams{
			AppID: app.ID,
			Limit: tickerLimit,
		})
		return e
	})
	if err != nil {
		jsonServerError(w, "failed to load pipeline bootstrap", err)
		return
	}

	events := make([]pipelineEventJSON, len(recent))
	for i, row := range recent {
		events[i] = rowToJSON(row)
	}

	// Cursor is the newest row's occurred_at (events are newest-first), or
	// time.Now() when the window is empty. The SSE endpoint's `?since=`
	// replay is exclusive (WHERE occurred_at > cursor) so using the newest
	// row's timestamp means we don't re-send what the bootstrap already
	// delivered.
	cursor := time.Now().UTC()
	if len(recent) > 0 {
		cursor = recent[0].OccurredAt
	}

	payload, err := json.Marshal(bootstrapResponse{
		Stats: pipelineStatsJSON{
			IngestionCount:  stats.IngestionCount,
			ClassifiedCount: stats.ClassifiedCount,
			FlaggedCount:    stats.FlaggedCount,
			SafeCount:       stats.SafeCount,
			AssessmentCount: stats.AssessmentCount,
			AvgConfidence:   stats.AvgConfidence,
			WindowSeconds:   int(window.Seconds()),
		},
		RecentEvents: events,
		Cursor:       cursor,
	})
	if err != nil {
		jsonServerError(w, "failed to marshal bootstrap response", err)
		return
	}

	if s.pipelineBootstrap != nil {
		s.pipelineBootstrap.Put(cacheKey, payload)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	_, _ = w.Write(payload)
}

// PipelineJourney handles GET /api/apps/{appId}/pipeline/logs/{logId}/journey.
// Returns the ordered stage history for a single log — the data backing
// Phase 4's Time Machine modal.
func (s *Server) PipelineJourney(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	logID, err := uuid.Parse(chi.URLParam(r, "logId"))
	if err != nil {
		jsonError(w, "invalid log id", http.StatusBadRequest)
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())

	var rows []db.LogPipelineEvent
	err = s.Pools.WithUserQueries(r.Context(), userID, func(q *db.Queries) error {
		var e error
		rows, e = q.GetPipelineEventsByLog(r.Context(), logID)
		return e
	})
	if err != nil {
		jsonServerError(w, "failed to load log journey", err)
		return
	}

	// Defence-in-depth: authorizeApp verified app ownership, but
	// GetPipelineEventsByLog doesn't filter by app_id. If a caller
	// guessed a log_id belonging to a different app in the same user
	// session, we'd leak its journey. Cross-check app_id on every row
	// before serialising.
	out := make([]pipelineEventJSON, 0, len(rows))
	for _, row := range rows {
		if row.AppID != app.ID {
			jsonError(w, "log not found", http.StatusNotFound)
			return
		}
		out = append(out, rowToJSON(row))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"log_id": logID,
		"events": out,
	})
}

// pipelineLogSummaryJSON is the wire shape for the Time Machine picker. One
// row per log, collapsing the log's up-to-four stage events into a single
// summary the frontend can render as a list item.
type pipelineLogSummaryJSON struct {
	LogID        uuid.UUID  `json:"log_id"`
	FirstSeenAt  time.Time  `json:"first_seen_at"`
	LastSeenAt   time.Time  `json:"last_seen_at"`
	StageCount   int64      `json:"stage_count"`
	SourceType   string     `json:"source_type,omitempty"`
	Severity     string     `json:"severity,omitempty"`
	Type         string     `json:"type,omitempty"`
	Category     string     `json:"category,omitempty"`
	Confidence   float64    `json:"confidence,omitempty"`
	Summary      string     `json:"summary,omitempty"`
	Escalated    *bool      `json:"escalated,omitempty"`
	RuleHit      string     `json:"rule_hit,omitempty"`
	AssessmentID *uuid.UUID `json:"assessment_id,omitempty"`
}

type pipelineLogsResponse struct {
	Logs      []pipelineLogSummaryJSON `json:"logs"`
	Since     time.Time                `json:"since"`
	Until     time.Time                `json:"until"`
	Limit     int                      `json:"limit"`
	Offset    int                      `json:"offset"`
	Truncated bool                     `json:"truncated"`
}

func summaryRowToJSON(row db.ListPipelineLogsByAppRow) pipelineLogSummaryJSON {
	out := pipelineLogSummaryJSON{
		LogID:        row.LogID,
		FirstSeenAt:  row.FirstSeenAt,
		LastSeenAt:   row.LastSeenAt,
		StageCount:   row.StageCount,
		AssessmentID: row.AssessmentID,
	}
	if row.SourceType.Valid {
		out.SourceType = row.SourceType.String
	}
	if row.Severity.Valid {
		out.Severity = row.Severity.String
	}
	if row.Type.Valid {
		out.Type = row.Type.String
	}
	if row.Category.Valid {
		out.Category = row.Category.String
	}
	if row.Confidence.Valid {
		out.Confidence = row.Confidence.Float64
	}
	if row.Summary.Valid {
		out.Summary = row.Summary.String
	}
	if row.Escalated.Valid {
		v := row.Escalated.Bool
		out.Escalated = &v
	}
	if row.RuleHit.Valid {
		out.RuleHit = row.RuleHit.String
	}
	return out
}

// parseLogsRange reads ?since= and ?until= as RFC3339 timestamps, falling
// back to a sane default window (until=now, since=now-24h). The result is
// clamped so the span never exceeds logsPickerMaxRange — at 48h retention
// anything wider is wasted work, and leaving the clamp implicit lets a
// buggy client DOS the picker. Returns (since, until) always with since <= until.
func parseLogsRange(q map[string][]string) (time.Time, time.Time) {
	now := time.Now().UTC()
	until := now
	since := now.Add(-logsPickerDefault)

	if raw, ok := q["until"]; ok && len(raw) > 0 && raw[0] != "" {
		if t, err := time.Parse(time.RFC3339Nano, raw[0]); err == nil {
			until = t
		}
	}
	if raw, ok := q["since"]; ok && len(raw) > 0 && raw[0] != "" {
		if t, err := time.Parse(time.RFC3339Nano, raw[0]); err == nil {
			since = t
		}
	}
	if until.Before(since) {
		since, until = until, since
	}
	if until.Sub(since) > logsPickerMaxRange {
		since = until.Add(-logsPickerMaxRange)
	}
	return since, until
}

func parseLogsLimit(raw string) int32 {
	if raw == "" {
		return logsPickerDefLimit
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return logsPickerDefLimit
	}
	if n > logsPickerMaxLimit {
		return logsPickerMaxLimit
	}
	return int32(n)
}

func parseLogsOffset(raw string) int32 {
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return int32(n)
}

// PipelineLogs handles GET /api/apps/{appId}/pipeline/logs.
//
// Query params:
//
//	since  — RFC3339 lower bound (default: now - 24h)
//	until  — RFC3339 upper bound (default: now)
//	limit  — 1..200 (default 50)
//	offset — >= 0 (default 0)
//
// Returns one summary row per distinct log within the window, newest last-seen
// first. Clamped to a 48h max range — the same retention ceiling that governs
// log_buffer/pipeline events. Used by the Time Machine picker list.
//
// Offset pagination (not cursor) is sufficient here: the picker is virtualised
// at the UI layer, the window is bounded to 48h, and the supporting index
// (idx_lpe_app_occurred) scans the range efficiently regardless of offset.
func (s *Server) PipelineLogs(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	since, until := parseLogsRange(r.URL.Query())
	limit := parseLogsLimit(r.URL.Query().Get("limit"))
	offset := parseLogsOffset(r.URL.Query().Get("offset"))

	userID, _ := middleware.UserIDFromContext(r.Context())
	var rows []db.ListPipelineLogsByAppRow
	err := s.Pools.WithUserQueries(r.Context(), userID, func(q *db.Queries) error {
		var e error
		rows, e = q.ListPipelineLogsByApp(r.Context(), db.ListPipelineLogsByAppParams{
			AppID:        app.ID,
			OccurredAt:   since,
			OccurredAt_2: until,
			Limit:        limit,
			Offset:       offset,
		})
		return e
	})
	if err != nil {
		jsonServerError(w, "failed to list pipeline logs", err)
		return
	}

	out := make([]pipelineLogSummaryJSON, len(rows))
	for i, row := range rows {
		out[i] = summaryRowToJSON(row)
	}

	// Truncated reports whether the result hit the limit — lets the client
	// know to offer a "load more" button without making it probe with an
	// off-by-one request.
	truncated := int32(len(rows)) >= limit

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pipelineLogsResponse{
		Logs:      out,
		Since:     since,
		Until:     until,
		Limit:     int(limit),
		Offset:    int(offset),
		Truncated: truncated,
	})
}

// PipelineStream handles GET /api/apps/{appId}/pipeline/stream as a Server-
// Sent Events endpoint with catch-up replay.
//
// Auth: token comes via `?token=` rather than an Authorization header
// because EventSource can't set headers — same idiom HandleChat uses.
//
// Sequence per connection:
//  1. Validate token, authorise app.
//  2. Drain events with occurred_at > ?since= (up to streamReplayCap) as
//     `event: replay` frames. Closes the hydration-to-stream gap.
//  3. If the replay hit the cap, emit one `event: resync` frame and skip
//     live subscribe — the client will refetch /bootstrap and reconnect.
//  4. Otherwise subscribe to PipelineBus and forward each event as
//     `event: live`. Heartbeat every 15s.
//  5. Exit on client disconnect or context cancel.
func (s *Server) PipelineStream(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, `{"error":"missing token query parameter"}`, http.StatusUnauthorized)
		return
	}
	userID, err := middleware.ValidateJWT(tokenStr, s.JWKS)
	if err != nil {
		http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "appId"))
	if err != nil {
		http.Error(w, `{"error":"invalid app id"}`, http.StatusBadRequest)
		return
	}
	if err := s.Pools.WithUserQueries(r.Context(), userID, func(q *db.Queries) error {
		_, e := q.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
			AppID:  appID,
			UserID: userID,
		})
		return e
	}); err != nil {
		http.Error(w, `{"error":"application not found"}`, http.StatusNotFound)
		return
	}

	// Phase 3.2: per-user connection cap. Acquire after authz so an
	// unauthorised caller can't influence the counter, and release on
	// handler exit (which on SSE means client disconnect). The limiter
	// is absent in tests that build Server{} as a struct literal — nil
	// path is an unbounded fallback.
	if s.pipelineSSELimiter != nil {
		if !s.pipelineSSELimiter.Acquire(userID) {
			http.Error(w, `{"error":"too many concurrent pipeline streams for this user"}`, http.StatusTooManyRequests)
			return
		}
		defer s.pipelineSSELimiter.Release(userID)
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":"streaming unsupported"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ctx := r.Context()

	// 2. Replay catch-up. Only runs when the client provides ?since= —
	//    a fresh connection with no cursor goes straight to live.
	needResync := false
	if raw := r.URL.Query().Get("since"); raw != "" {
		since, err := time.Parse(time.RFC3339Nano, raw)
		if err == nil {
			var rows []db.LogPipelineEvent
			err := s.Pools.WithUserQueries(ctx, userID, func(q *db.Queries) error {
				var e error
				rows, e = q.GetPipelineEventsSince(ctx, db.GetPipelineEventsSinceParams{
					AppID:      appID,
					OccurredAt: since,
					Limit:      streamReplayCap,
				})
				return e
			})
			if err != nil {
				slog.Warn("pipeline stream: replay query failed", "err", err, "app_id", appID)
			} else {
				for _, row := range rows {
					if err := writeSSEEvent(w, flusher, "replay", rowToJSON(row)); err != nil {
						return
					}
				}
				// If we hit the cap, the client has been disconnected
				// long enough that "since" is effectively unbounded —
				// tell them to drop the buffer and refetch /bootstrap.
				if len(rows) >= streamReplayCap {
					needResync = true
				}
			}
		}
	}

	if needResync {
		_ = writeSSEEvent(w, flusher, "resync", map[string]string{"reason": "replay_cap_exceeded"})
		// Don't subscribe — client will reconnect with a fresh cursor
		// after calling /bootstrap again. Returning cleanly lets the
		// browser's EventSource interpret this as end-of-stream.
		return
	}

	// 3. Live subscription. PipelineBus may be nil in tests or during
	//    startup misconfiguration — in that case hold the connection
	//    open on heartbeat only so the client still sees "streaming".
	var liveCh <-chan agent.PipelineEvent
	var unsub func()
	if s.Agent != nil {
		if bus := s.Agent.PipelineBus(); bus != nil {
			liveCh, unsub = bus.Subscribe(appID)
			defer unsub()
		}
	}

	heartbeat := time.NewTicker(streamHeartbeat)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case evt, ok := <-liveCh:
			if !ok {
				return
			}
			if err := writeSSEEvent(w, flusher, "live", busEventToJSON(evt)); err != nil {
				return
			}
		}
	}
}

// writeSSEEvent formats and flushes a single SSE frame. Returns the
// underlying write error so the caller can bail on a dead connection
// rather than burning CPU on a broken pipe.
func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, name string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("pipeline stream: marshal failed", "event", name, "err", err)
		return nil
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", name, data); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}
