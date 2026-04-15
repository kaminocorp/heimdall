package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// Schedule request/validation limits. Kept as named constants so the
// handler reads cleanly and the scheduler tests can reference them without
// duplicating magic numbers.
const (
	// minIntervalSecs matches the scheduler tick — a schedule can't usefully
	// fire more frequently than the loop checks it anyway. Only applies to
	// the legacy interval_secs path; cron schedules have their own cadence.
	minIntervalSecs = 60
	// maxIntervalSecs caps the legacy interval path at one day. Anything
	// coarser than daily should use cron_expr instead.
	maxIntervalSecs = 86400
	// maxScheduleNameLen keeps the UI list rendering predictable.
	maxScheduleNameLen = 100
	// maxSchedulePromptLen bounds per-run LLM input cost. 5000 chars is
	// deliberately generous — tighten later if abuse shows up.
	maxSchedulePromptLen = 5000
	// maxCronExprLen is a sanity cap. A well-formed cron expression is
	// tens of bytes; anything longer is either pathological or an attempt
	// to fill the DB with padding.
	maxCronExprLen = 200
)

// scheduleRequest is the JSON body shape for Create and Update. A client
// must populate exactly one of IntervalSecs (>0) or CronExpr (non-empty);
// sending both or neither is a 400.
//
// enabled is a pointer so we can distinguish "not provided" from "explicitly
// false". Create treats nil as true (default enabled); Update treats nil as
// "keep current value".
type scheduleRequest struct {
	Name         string `json:"name"`
	Prompt       string `json:"prompt"`
	IntervalSecs int32  `json:"interval_secs,omitempty"`
	CronExpr     string `json:"cron_expr,omitempty"`
	Enabled      *bool  `json:"enabled"`
}

// validateSchedule runs the shared validation pipeline for Create/Update.
// Writes an HTTP 400 and returns false on failure; caller should return
// immediately in that case. Mutates req in place (trim whitespace on string
// fields and the cron expression).
//
// The "exactly-one-of" rule for interval_secs/cron_expr is enforced here:
// sending both or neither is a 400. This keeps the shouldFire contract
// (cron wins if set) from being an implicit priority rule the user has to
// learn — the API refuses ambiguous input up front.
func validateSchedule(w http.ResponseWriter, req *scheduleRequest) bool {
	req.Name = strings.TrimSpace(req.Name)
	req.Prompt = strings.TrimSpace(req.Prompt)
	req.CronExpr = strings.TrimSpace(req.CronExpr)

	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return false
	}
	if len(req.Name) > maxScheduleNameLen {
		jsonError(w, "name exceeds maximum length", http.StatusBadRequest)
		return false
	}
	if req.Prompt == "" {
		jsonError(w, "prompt is required", http.StatusBadRequest)
		return false
	}
	if len(req.Prompt) > maxSchedulePromptLen {
		jsonError(w, "prompt exceeds maximum length", http.StatusBadRequest)
		return false
	}

	// Exactly-one-of check. IntervalSecs == 0 is treated as "not set" so
	// the int32 zero value matches JSON field omission; the same logic
	// applies to an empty CronExpr after trimming.
	hasInterval := req.IntervalSecs > 0
	hasCron := req.CronExpr != ""
	if hasInterval && hasCron {
		jsonError(w, "provide either interval_secs or cron_expr, not both", http.StatusBadRequest)
		return false
	}
	if !hasInterval && !hasCron {
		jsonError(w, "provide either interval_secs or cron_expr", http.StatusBadRequest)
		return false
	}

	if hasInterval {
		if req.IntervalSecs < minIntervalSecs || req.IntervalSecs > maxIntervalSecs {
			jsonError(w, "interval_secs must be between 60 and 86400", http.StatusBadRequest)
			return false
		}
	}
	if hasCron {
		if len(req.CronExpr) > maxCronExprLen {
			jsonError(w, "cron_expr exceeds maximum length", http.StatusBadRequest)
			return false
		}
		// Parse at the API boundary so the DB never sees an invalid expression.
		// The scheduler also handles parse failures defensively (see
		// scheduler.go:shouldFire) but catching them here gives the user a
		// crisp 400 instead of a silent stuck schedule.
		if _, err := agent.ParseCronExpression(req.CronExpr); err != nil {
			jsonError(w, "invalid cron_expr: "+err.Error(), http.StatusBadRequest)
			return false
		}
	}
	return true
}

// toDBScheduleParams packs a validated scheduleRequest into the shape sqlc
// wants. When CronExpr is set we zero out IntervalSecs (and vice versa) so
// the DB always reflects exactly one mode — even though shouldFire is
// defensively designed to tolerate both, we don't want the storage layer
// to carry ambiguous state.
func toDBScheduleParams(req *scheduleRequest) (intervalSecs int32, cronExpr pgtype.Text) {
	if req.CronExpr != "" {
		return 0, pgtype.Text{String: req.CronExpr, Valid: true}
	}
	return req.IntervalSecs, pgtype.Text{} // NULL
}

// ListSchedules — GET /api/apps/{appId}/schedules
func (s *Server) ListSchedules(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	schedules, err := s.Queries.ListSchedulesByApp(r.Context(), app.ID)
	if err != nil {
		jsonServerError(w, "failed to list schedules", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedules)
}

// CreateSchedule — POST /api/apps/{appId}/schedules
func (s *Server) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	var req scheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if !validateSchedule(w, &req) {
		return
	}
	// Default to enabled=true on create when the client omits the field.
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "failed to begin transaction", err)
		return
	}
	defer done()

	intervalSecs, cronExpr := toDBScheduleParams(&req)
	schedule, err := queries.CreateSchedule(r.Context(), db.CreateScheduleParams{
		AppID:        app.ID,
		Name:         req.Name,
		Prompt:       req.Prompt,
		IntervalSecs: intervalSecs,
		CronExpr:     cronExpr,
		Enabled:      enabled,
	})
	if err != nil {
		jsonServerError(w, "failed to create schedule", err)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schedule)
}

// UpdateSchedule — PATCH /api/apps/{appId}/schedules/{id}
func (s *Server) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	scheduleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid schedule id", http.StatusBadRequest)
		return
	}

	var req scheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if !validateSchedule(w, &req) {
		return
	}

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "failed to begin transaction", err)
		return
	}
	defer done()

	// Verify schedule belongs to this app inside the transaction to avoid TOCTOU.
	existing, err := queries.GetSchedule(r.Context(), scheduleID)
	if err != nil || existing.AppID != app.ID {
		jsonError(w, "schedule not found", http.StatusNotFound)
		return
	}

	// On update, nil enabled means "keep current value".
	enabled := existing.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	intervalSecs, cronExpr := toDBScheduleParams(&req)
	updated, err := queries.UpdateSchedule(r.Context(), db.UpdateScheduleParams{
		ID:           scheduleID,
		Name:         req.Name,
		Prompt:       req.Prompt,
		IntervalSecs: intervalSecs,
		CronExpr:     cronExpr,
		Enabled:      enabled,
		AppID:        app.ID,
	})
	if err != nil {
		jsonServerError(w, "failed to update schedule", err)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// DeleteSchedule — DELETE /api/apps/{appId}/schedules/{id}
func (s *Server) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	scheduleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid schedule id", http.StatusBadRequest)
		return
	}

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "failed to begin transaction", err)
		return
	}
	defer done()

	// Verify schedule belongs to this app inside the transaction to avoid TOCTOU.
	existing, err := queries.GetSchedule(r.Context(), scheduleID)
	if err != nil || existing.AppID != app.ID {
		jsonError(w, "schedule not found", http.StatusNotFound)
		return
	}

	if err := queries.DeleteSchedule(r.Context(), db.DeleteScheduleParams{
		ID:    scheduleID,
		AppID: app.ID,
	}); err != nil {
		jsonServerError(w, "failed to delete schedule", err)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RunScheduleNow — POST /api/apps/{appId}/schedules/{id}/run
//
// Fires a schedule synchronously, bypassing the 1-minute tick. Useful for
// "test this schedule" in the UI and for ad-hoc investigation kicks.
//
// Shares the monitor rate limiter with automatic runs (see scheduler.go)
// so a burst of manual-click spam can't starve the monitor loop. The UI
// should disable the button for a few seconds after a click as a second
// layer of protection.
func (s *Server) RunScheduleNow(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	if s.Agent == nil {
		// This only happens in handler integration tests where the test env
		// wires a nil agent. In production the agent is always present.
		jsonError(w, "agent not available in this environment", http.StatusServiceUnavailable)
		return
	}

	scheduleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid schedule id", http.StatusBadRequest)
		return
	}

	schedule, err := s.Queries.GetSchedule(r.Context(), scheduleID)
	if err != nil {
		jsonError(w, "schedule not found", http.StatusNotFound)
		return
	}
	if schedule.AppID != app.ID {
		jsonError(w, "schedule not found", http.StatusNotFound)
		return
	}

	// Block on the run so the caller gets the fresh last_status/last_summary
	// in the response. The 5-minute scheduler timeout applies inside
	// runScheduledInvestigation, so worst-case latency is bounded.
	s.Agent.RunScheduledInvestigation(r.Context(), schedule)

	// Re-read to return the updated fields.
	updated, err := s.Queries.GetSchedule(r.Context(), scheduleID)
	if err != nil {
		jsonServerError(w, "failed to reload schedule after run", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
