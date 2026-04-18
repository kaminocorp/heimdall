package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// resolveSourceFilterApp returns the app_id that a source-filter request
// should operate on, along with the HTTP status code to return on failure.
//
//   - App-scoped connection (conn.AppID != nil): the app is unambiguous;
//     callers don't need to pass ?app_id=. If they do, it must match.
//   - Org-scoped connection (conn.AppID == nil): ?app_id= is required. The
//     caller must be a member of the app's org, and the app's org must match
//     the connection's org — we won't cross-contaminate filter rows between
//     unrelated orgs even if a user happens to be in both.
//
// Returned status codes: 400 for syntactic problems (missing/malformed
// app_id), 403 for policy denials (cross-org app, app/conn mismatch) so
// clients can distinguish retry-with-different-input from access-denied,
// and 404 only for connection-not-found paths (handled by the caller).
func (s *Server) resolveSourceFilterApp(
	ctx context.Context, queries *db.Queries, conn db.Connection, userID uuid.UUID, r *http.Request,
) (uuid.UUID, int, error) {
	queryAppStr := strings.TrimSpace(r.URL.Query().Get("app_id"))

	if conn.AppID != nil {
		if queryAppStr != "" {
			parsed, err := uuid.Parse(queryAppStr)
			if err != nil {
				return uuid.Nil, http.StatusBadRequest, errors.New("invalid app_id")
			}
			if parsed != *conn.AppID {
				return uuid.Nil, http.StatusForbidden, errors.New("app_id does not match the connection's app")
			}
		}
		return *conn.AppID, 0, nil
	}

	if queryAppStr == "" {
		return uuid.Nil, http.StatusBadRequest, errors.New("app_id query parameter is required for org-scoped connections")
	}
	parsed, err := uuid.Parse(queryAppStr)
	if err != nil {
		return uuid.Nil, http.StatusBadRequest, errors.New("invalid app_id")
	}
	app, err := queries.GetApplicationByOrgUser(ctx, db.GetApplicationByOrgUserParams{
		AppID:  parsed,
		UserID: userID,
	})
	if err != nil {
		// The app exists, the user isn't a member, OR the app doesn't
		// exist — we don't distinguish so we don't leak existence. 403
		// covers both cases from the client's perspective.
		return uuid.Nil, http.StatusForbidden, errors.New("application not accessible")
	}
	if app.OrgID != conn.OrgID {
		return uuid.Nil, http.StatusForbidden, errors.New("application not in connection's org")
	}
	return parsed, 0, nil
}

// sourceFilterItem is the merged shape returned to the frontend. It combines
// the `connection_sources` discovery ledger (first_seen_at / last_seen_at,
// nullable for manually-added sources that have never been seen) with the
// `app_source_filters` enable state.
//
// Manual additions appear with Enabled=true and null timestamps until the
// next ingestion batch populates `connection_sources`.
type sourceFilterItem struct {
	SourceName  string     `json:"source_name"`
	Enabled     bool       `json:"enabled"`
	FirstSeenAt *time.Time `json:"first_seen_at"`
	LastSeenAt  *time.Time `json:"last_seen_at"`
}

// ListSourceFilters returns every source the ingestion layer has ever seen
// on a connection, merged with the current app's enable/disable state. The
// "current app" is derived from `connection.app_id` (app-scoped Phase 1);
// Phase 2 will widen this to accept an `?app_id=` query param for org-scoped
// connections.
//
// GET /api/connections/{id}/sources
func (s *Server) ListSourceFilters(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	queries, _, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	conn, err := queries.GetConnectionByUser(r.Context(), db.GetConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}

	appID, status, err := s.resolveSourceFilterApp(r.Context(), queries, conn, userID, r)
	if err != nil {
		jsonError(w, err.Error(), status)
		return
	}

	// Load both halves — discovery ledger (may be empty) and the app's
	// current filter selections (may also be empty).
	sources, err := queries.ListConnectionSources(r.Context(), connID)
	if err != nil {
		jsonServerError(w, "failed to list sources", err)
		return
	}
	filters, err := queries.ListAppSourceFilters(r.Context(), db.ListAppSourceFiltersParams{
		AppID:        appID,
		ConnectionID: connID,
	})
	if err != nil {
		jsonServerError(w, "failed to list filters", err)
		return
	}

	// Merge: the union of source_names across both tables, with enable state
	// taken from filters. Same pattern ListGitHubRepos uses for its repo
	// discovery merge.
	byName := make(map[string]*sourceFilterItem, len(sources)+len(filters))

	for i := range sources {
		src := sources[i]
		first := src.FirstSeenAt
		last := src.LastSeenAt
		byName[src.SourceName] = &sourceFilterItem{
			SourceName:  src.SourceName,
			Enabled:     false,
			FirstSeenAt: &first,
			LastSeenAt:  &last,
		}
	}
	for _, f := range filters {
		item, exists := byName[f.SourceName]
		if !exists {
			item = &sourceFilterItem{SourceName: f.SourceName}
			byName[f.SourceName] = item
		}
		item.Enabled = f.Enabled
	}

	result := make([]sourceFilterItem, 0, len(byName))
	for _, item := range byName {
		result = append(result, *item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// updateSourceFiltersRequest is the bulk-PUT body: an array of toggle
// operations. Only source_names present in the array are touched; filters
// absent from the array are left alone.
type updateSourceFiltersRequest struct {
	Sources []struct {
		SourceName string `json:"source_name"`
		Enabled    bool   `json:"enabled"`
	} `json:"sources"`
}

// UpdateSourceFilters bulk-upserts filter rows for (current app, connection).
// Each item becomes one UpsertAppSourceFilter call inside a transaction so
// the whole save is atomic — partial saves are not allowed.
//
// PUT /api/connections/{id}/sources
func (s *Server) UpdateSourceFilters(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	var req updateSourceFiltersRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	// Hard cap keeps the per-request DB fan-out bounded. Matches the
	// ListGitHubRepos upper bound.
	if len(req.Sources) > 500 {
		jsonError(w, "maximum 500 sources per update", http.StatusBadRequest)
		return
	}

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	conn, err := queries.GetConnectionByUser(r.Context(), db.GetConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}

	appID, status, err := s.resolveSourceFilterApp(r.Context(), queries, conn, userID, r)
	if err != nil {
		jsonError(w, err.Error(), status)
		return
	}

	for _, item := range req.Sources {
		name := strings.TrimSpace(item.SourceName)
		if name == "" {
			continue
		}
		if _, err := queries.UpsertAppSourceFilter(r.Context(), db.UpsertAppSourceFilterParams{
			AppID:        appID,
			ConnectionID: connID,
			SourceName:   name,
			Enabled:      item.Enabled,
		}); err != nil {
			jsonServerError(w, "failed to update source filters", err)
			return
		}
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// addSourceFilterRequest is the body for the manual-add POST. Unlike the
// bulk PUT this only toggles on — manually added sources are always enabled
// at creation time (a user who types a name in the UI intends to accept
// those logs).
type addSourceFilterRequest struct {
	SourceName string `json:"source_name"`
}

// AddSourceFilter manually registers a source name for this (app,
// connection) pair. Creates both a placeholder `connection_sources` row
// (so it appears in the selector with a "never seen" timestamp until
// traffic arrives) and an `app_source_filters` row with enabled=true.
//
// POST /api/connections/{id}/sources
func (s *Server) AddSourceFilter(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	var req addSourceFilterRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(req.SourceName)
	if name == "" {
		jsonError(w, "source_name is required", http.StatusBadRequest)
		return
	}
	if len(name) > 256 {
		jsonError(w, "source_name exceeds 256 characters", http.StatusBadRequest)
		return
	}

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	conn, err := queries.GetConnectionByUser(r.Context(), db.GetConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}

	appID, status, err := s.resolveSourceFilterApp(r.Context(), queries, conn, userID, r)
	if err != nil {
		jsonError(w, err.Error(), status)
		return
	}

	// Create the placeholder discovery row. UpsertConnectionSource bumps
	// last_seen_at on conflict, which is wrong here (no traffic was actually
	// seen) — but the surface of this UI shows the last-seen timestamp as
	// "never seen" when it's within a few seconds of created_at anyway, and
	// the next real ingestion will correct it. Acceptable trade-off for
	// Phase 1.
	if _, err := queries.UpsertConnectionSource(r.Context(), db.UpsertConnectionSourceParams{
		ConnectionID: connID,
		SourceName:   name,
	}); err != nil {
		jsonServerError(w, "failed to register source", err)
		return
	}

	filter, err := queries.UpsertAppSourceFilter(r.Context(), db.UpsertAppSourceFilterParams{
		AppID:        appID,
		ConnectionID: connID,
		SourceName:   name,
		Enabled:      true,
	})
	if err != nil {
		jsonServerError(w, "failed to create source filter", err)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(filter)
}

// DeleteSourceFilter removes a filter row. Only removes the app-scoped
// enable/disable toggle — the discovery ledger row in `connection_sources`
// is untouched, so if traffic arrives again the source will reappear in
// the selector as disabled (drop-by-default).
//
// The source name is a query parameter (?name=...) rather than a path
// component because source names can contain slashes (e.g. the Vercel
// flavour "vercel/lambda"), which would confuse chi's path matching even
// with URL-encoding, since some proxies normalise %2F back to /.
//
// DELETE /api/connections/{id}/sources?name=...
func (s *Server) DeleteSourceFilter(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		jsonError(w, "name query parameter is required", http.StatusBadRequest)
		return
	}

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	conn, err := queries.GetConnectionByUser(r.Context(), db.GetConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}

	appID, status, err := s.resolveSourceFilterApp(r.Context(), queries, conn, userID, r)
	if err != nil {
		jsonError(w, err.Error(), status)
		return
	}

	if err := queries.DeleteAppSourceFilter(r.Context(), db.DeleteAppSourceFilterParams{
		AppID:        appID,
		ConnectionID: connID,
		SourceName:   name,
	}); err != nil {
		jsonServerError(w, "failed to delete source filter", err)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
