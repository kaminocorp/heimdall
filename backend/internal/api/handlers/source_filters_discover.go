package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// discoverError tags a failure in discoverGitHubRepos with the category of
// problem so the handler can map it to the right HTTP status. Without this,
// every error surfaces as 502 Bad Gateway, which lies to clients about
// misconfigurations (503) and corrupt internal state (500).
type discoverError struct {
	status int
	msg    string
}

func (e *discoverError) Error() string { return e.msg }

func newDiscoverErr(status int, format string, args ...any) *discoverError {
	return &discoverError{status: status, msg: fmt.Sprintf(format, args...)}
}

// discoverResponse is the acknowledgement returned by the discover endpoint.
// Minimal by design — the frontend refreshes its source list via a separate
// GET /sources call after success; `discovered` is just confirmation the
// upstream fetch actually returned something.
type discoverResponse struct {
	Discovered     int    `json:"discovered"`
	ConnectionType string `json:"connection_type"`
}

// DiscoverSources pulls the canonical source list from a connection's
// upstream and upserts every result into `connection_sources`. For webhook /
// OTLP connections, discovery happens automatically as traffic arrives, so
// this endpoint only applies to connector types with an explicit "list what
// we have access to" API — today that's GitHub (repos accessible to the
// installation). Other types return 400.
//
// POST /api/connections/{id}/sources/discover
func (s *Server) DiscoverSources(w http.ResponseWriter, r *http.Request) {
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

	// Defense-in-depth: resolve the app scope even though the current sole
	// discovery caller (GitHub) is always app-scoped and GetConnectionByUser
	// has already gated org membership. When a future non-GitHub connector
	// gets an explicit discovery path, this pre-check ensures org-scoped
	// connections can't write to `connection_sources` for a connection the
	// caller can't also read via the app-scoped filter APIs. The returned
	// appID is intentionally unused today — discovery writes connection-wide
	// rows, not per-app rows.
	if _, status, err := s.resolveSourceFilterApp(r.Context(), queries, conn, userID, r); err != nil {
		jsonError(w, err.Error(), status)
		return
	}

	switch conn.Type {
	case "github":
		n, err := s.discoverGitHubRepos(r.Context(), queries, conn)
		if err != nil {
			// Map the tagged discoverError's status directly onto the
			// response — jsonServerError forces 500, which is wrong for
			// 502/503/429 tags. Log-to-oncall side-band for 5xx so server
			// errors stay visible.
			var de *discoverError
			if errors.As(err, &de) {
				if de.status >= 500 {
					slog.Error("discovery failed", "err", err, "status", de.status)
				}
				jsonError(w, de.msg, de.status)
			} else {
				jsonServerError(w, "discovery failed", err)
			}
			return
		}
		if err := commit(); err != nil {
			jsonServerError(w, "failed to save discovered sources", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(discoverResponse{Discovered: n, ConnectionType: "github"})

	default:
		jsonError(w,
			fmt.Sprintf("discovery not supported for connection type %q — sources appear automatically as traffic arrives", conn.Type),
			http.StatusBadRequest)
	}
}

// discoverGitHubRepos pulls every repo accessible to the connection's
// installation and records each `owner/repo` as a row in connection_sources.
// Returns the count of distinct repos seen.
//
// Pagination: hard cap of 50 pages × 100 repos = 5,000 repos. Matches the
// pre-Phase-3 ListGitHubRepos handler, so behaviour is preserved for any
// organisation that was working under that limit.
func (s *Server) discoverGitHubRepos(ctx context.Context, queries *db.Queries, conn db.Connection) (int, error) {
	if s.GitHub == nil {
		// Server-level configuration issue — the installation isn't the
		// problem, our deployment is. 503 signals "try again later when
		// the operator has fixed this".
		return 0, newDiscoverErr(http.StatusServiceUnavailable, "GitHub App not configured on this server")
	}

	var cfg map[string]any
	if err := json.Unmarshal(conn.Config, &cfg); err != nil {
		// Our stored config is corrupt — bug on our side, not the user's
		// or GitHub's. 500 routes this through server-error logging so
		// oncall sees the underlying cause.
		return 0, newDiscoverErr(http.StatusInternalServerError, "invalid connection config")
	}
	installationID, ok := cfg["installation_id"].(float64)
	if !ok {
		return 0, newDiscoverErr(http.StatusInternalServerError, "connection is missing installation_id")
	}

	token, err := s.GitHub.InstallationTokenFor(ctx, int64(installationID))
	if err != nil {
		slog.Error("discover: failed to get installation token", "err", err)
		return 0, newDiscoverErr(http.StatusBadGateway, "failed to authenticate with GitHub")
	}

	const maxPages = 50
	seen := 0
	page := 1
	for page <= maxPages {
		url := fmt.Sprintf("https://api.github.com/installation/repositories?per_page=100&page=%d", page)
		resp, err := s.GitHub.APIRequest(ctx, "GET", url, token.Token, nil)
		if err != nil {
			slog.Error("discover: list repos failed", "err", err)
			return seen, newDiscoverErr(http.StatusBadGateway, "failed to fetch repositories from GitHub")
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
		resp.Body.Close()
		if err != nil {
			return seen, newDiscoverErr(http.StatusBadGateway, "failed to read GitHub response")
		}
		if resp.StatusCode != http.StatusOK {
			slog.Error("discover: bad status from GitHub", "status", resp.StatusCode, "body", string(body))
			// 429: GitHub is rate-limiting us; surface that so the client
			// can back off. Everything else is "upstream returned non-200",
			// which is 502. (Note: the token-for-installation call above
			// hits a different endpoint, so 401 here means the installation
			// token was revoked between issuance and use — rare.)
			if resp.StatusCode == http.StatusTooManyRequests {
				return seen, newDiscoverErr(http.StatusTooManyRequests, "GitHub rate limit exceeded — try again in a few minutes")
			}
			return seen, newDiscoverErr(http.StatusBadGateway, "GitHub API error (status %d)", resp.StatusCode)
		}

		var result struct {
			Repositories []struct {
				FullName string `json:"full_name"`
			} `json:"repositories"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return seen, newDiscoverErr(http.StatusBadGateway, "failed to parse GitHub response")
		}

		if len(result.Repositories) == 0 {
			break
		}
		for _, repo := range result.Repositories {
			if repo.FullName == "" {
				continue
			}
			if _, err := queries.UpsertConnectionSource(ctx, db.UpsertConnectionSourceParams{
				ConnectionID: conn.ID,
				SourceName:   repo.FullName,
			}); err != nil {
				// Upsert failure is an internal DB problem, not an upstream
				// problem — 500 so it routes through server-error logging.
				return seen, newDiscoverErr(http.StatusInternalServerError, "failed to record source %q: %v", repo.FullName, err)
			}
			seen++
		}

		if len(result.Repositories) < 100 {
			break
		}
		page++
	}

	return seen, nil
}
