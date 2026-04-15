package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// ListGitHubRepos returns repos accessible to a GitHub connection's installation,
// merged with the enabled/disabled state from the github_repos table.
func (s *Server) ListGitHubRepos(w http.ResponseWriter, r *http.Request) {
	if s.GitHub == nil {
		jsonError(w, "GitHub App not configured", http.StatusServiceUnavailable)
		return
	}

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

	if conn.Type != "github" {
		jsonError(w, "not a GitHub connection", http.StatusBadRequest)
		return
	}

	// Parse installation_id from config.
	var cfg map[string]any
	if err := json.Unmarshal(conn.Config, &cfg); err != nil {
		jsonServerError(w, "invalid connection config", err)
		return
	}
	installationID, ok := cfg["installation_id"].(float64)
	if !ok {
		jsonError(w, "missing installation_id in config", http.StatusInternalServerError)
		return
	}

	// Get installation token.
	token, err := s.GitHub.InstallationTokenFor(r.Context(), int64(installationID))
	if err != nil {
		slog.Error("github: failed to get installation token", "err", err)
		jsonError(w, "failed to authenticate with GitHub", http.StatusBadGateway)
		return
	}

	// Fetch repos from GitHub (paginate).
	type ghRepo struct {
		RepoID        int64  `json:"repo_id"`
		RepoFullName  string `json:"repo_full_name"`
		DefaultBranch string `json:"default_branch"`
		Enabled       bool   `json:"enabled"`
	}

	var allRepos []ghRepo
	page := 1
	const maxPages = 50 // Safety bound: 50 pages × 100 repos = 5,000 repos max.
	for page <= maxPages {
		url := fmt.Sprintf("https://api.github.com/installation/repositories?per_page=100&page=%d", page)
		resp, err := s.GitHub.APIRequest(r.Context(), "GET", url, token.Token, nil)
		if err != nil {
			slog.Error("github: failed to list repos", "err", err)
			jsonError(w, "failed to fetch repos from GitHub", http.StatusBadGateway)
			return
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
		resp.Body.Close()
		if err != nil {
			slog.Error("github: failed to read response body", "err", err)
			jsonError(w, "failed to read GitHub response", http.StatusBadGateway)
			return
		}

		if resp.StatusCode != http.StatusOK {
			slog.Error("github: list repos failed", "status", resp.StatusCode, "body", string(body))
			jsonError(w, "GitHub API error", http.StatusBadGateway)
			return
		}

		var result struct {
			Repositories []struct {
				ID            int64  `json:"id"`
				FullName      string `json:"full_name"`
				DefaultBranch string `json:"default_branch"`
			} `json:"repositories"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			jsonServerError(w, "failed to parse GitHub response", err)
			return
		}

		if len(result.Repositories) == 0 {
			break
		}

		for _, repo := range result.Repositories {
			allRepos = append(allRepos, ghRepo{
				RepoID:        repo.ID,
				RepoFullName:  repo.FullName,
				DefaultBranch: repo.DefaultBranch,
				Enabled:       false,
			})
		}

		if len(result.Repositories) < 100 {
			break
		}
		page++
	}

	// Merge with DB state.
	dbRepos, err := queries.ListGitHubReposByConnection(r.Context(), connID)
	if err != nil {
		jsonServerError(w, "failed to load repo state", err)
		return
	}
	enabledMap := make(map[int64]bool, len(dbRepos))
	for _, r := range dbRepos {
		enabledMap[r.RepoID] = r.Enabled
	}
	for i := range allRepos {
		if enabled, ok := enabledMap[allRepos[i].RepoID]; ok {
			allRepos[i].Enabled = enabled
		}
	}

	if allRepos == nil {
		allRepos = []ghRepo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allRepos)
}

// UpdateGitHubRepos updates which repos are enabled for a GitHub connection.
func (s *Server) UpdateGitHubRepos(w http.ResponseWriter, r *http.Request) {
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

	if conn.Type != "github" {
		jsonError(w, "not a GitHub connection", http.StatusBadRequest)
		return
	}

	var repos []struct {
		RepoID        int64  `json:"repo_id"`
		RepoFullName  string `json:"repo_full_name"`
		DefaultBranch string `json:"default_branch"`
		Enabled       bool   `json:"enabled"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&repos); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(repos) > 100 {
		jsonError(w, "maximum 100 repositories per update", http.StatusBadRequest)
		return
	}

	for _, repo := range repos {
		_, err := queries.UpsertGitHubRepo(r.Context(), db.UpsertGitHubRepoParams{
			ConnectionID:  connID,
			RepoFullName:  repo.RepoFullName,
			RepoID:        repo.RepoID,
			DefaultBranch: repo.DefaultBranch,
			Enabled:       repo.Enabled,
		})
		if err != nil {
			jsonServerError(w, "failed to update repos", err)
			return
		}
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestGitHubConnection tests a GitHub connection by checking the installation token.
func (s *Server) TestGitHubConnection(ctx context.Context, connConfig json.RawMessage) (bool, string) {
	if s.GitHub == nil {
		return false, "GitHub App not configured"
	}

	var cfg map[string]any
	if err := json.Unmarshal(connConfig, &cfg); err != nil {
		return false, "invalid connection config"
	}

	installationID, ok := cfg["installation_id"].(float64)
	if !ok {
		return false, "missing installation_id in config"
	}

	_, err := s.GitHub.InstallationTokenFor(ctx, int64(installationID))
	if err != nil {
		return false, fmt.Sprintf("GitHub auth failed: %v", err)
	}

	return true, "GitHub connection active"
}
