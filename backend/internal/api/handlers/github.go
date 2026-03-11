package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// --- GitHub App Installation Flow ---

// InstallGitHub returns a URL that redirects the user to install the GitHub App.
func (s *Server) InstallGitHub(w http.ResponseWriter, r *http.Request) {
	if s.GitHub == nil {
		jsonError(w, "GitHub App not configured", http.StatusServiceUnavailable)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	appIDStr := r.URL.Query().Get("app_id")
	if appIDStr == "" {
		jsonError(w, "app_id query parameter required", http.StatusBadRequest)
		return
	}

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		jsonError(w, "invalid app_id", http.StatusBadRequest)
		return
	}

	// Verify app belongs to user's org.
	_, err = s.Queries.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
		AppID:  appID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "application not found", http.StatusNotFound)
		return
	}

	// Generate state JWT containing user_id and app_id.
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"app_id":  appID.String(),
		"nonce":   uuid.New().String(),
		"iat":     now.Unix(),
		"exp":     now.Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	stateStr, err := token.SignedString(s.GitHub.PrivateKey())
	if err != nil {
		slog.Error("github: failed to sign state JWT", "err", err)
		jsonError(w, "failed to generate install URL", http.StatusInternalServerError)
		return
	}

	installURL := fmt.Sprintf("https://github.com/apps/%s/installations/new?state=%s", s.Config.GitHubAppSlug, stateStr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url": installURL,
	})
}

// GitHubCallback handles the redirect from GitHub after app installation.
func (s *Server) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	if s.GitHub == nil {
		jsonError(w, "GitHub App not configured", http.StatusServiceUnavailable)
		return
	}

	installationIDStr := r.URL.Query().Get("installation_id")
	stateStr := r.URL.Query().Get("state")
	setupAction := r.URL.Query().Get("setup_action")

	if installationIDStr == "" || stateStr == "" {
		jsonError(w, "missing installation_id or state", http.StatusBadRequest)
		return
	}

	installationID, err := strconv.ParseInt(installationIDStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid installation_id", http.StatusBadRequest)
		return
	}

	// Validate state JWT.
	token, err := jwt.Parse(stateStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return &s.GitHub.PrivateKey().PublicKey, nil
	})
	if err != nil {
		slog.Warn("github callback: invalid state JWT", "err", err)
		jsonError(w, "invalid or expired state", http.StatusBadRequest)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		jsonError(w, "invalid state claims", http.StatusBadRequest)
		return
	}

	userIDStr, _ := claims["user_id"].(string)
	appIDStr, _ := claims["app_id"].(string)

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		jsonError(w, "invalid user_id in state", http.StatusBadRequest)
		return
	}

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		jsonError(w, "invalid app_id in state", http.StatusBadRequest)
		return
	}

	// Verify app belongs to user's org (user-scoped query).
	queries, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	defer done()

	_, err = queries.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
		AppID:  appID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "application not found", http.StatusNotFound)
		return
	}

	// Fetch installation details from GitHub to get account info.
	installation, err := s.GitHub.GetInstallation(r.Context(), installationID)
	if err != nil {
		slog.Error("github callback: failed to get installation", "err", err, "installation_id", installationID)
		jsonError(w, "failed to verify GitHub installation", http.StatusBadGateway)
		return
	}

	accountLogin := ""
	accountType := ""
	if account, ok := installation["account"].(map[string]any); ok {
		accountLogin, _ = account["login"].(string)
		accountType, _ = account["type"].(string)
	}

	// Build connection config.
	configJSON, err := json.Marshal(map[string]any{
		"installation_id": installationID,
		"account_login":   accountLogin,
		"account_type":    accountType,
	})
	if err != nil {
		jsonError(w, "failed to build config", http.StatusInternalServerError)
		return
	}

	// Check for existing GitHub connection with same installation_id for this app.
	existingConns, err := queries.ListConnectionsByApp(r.Context(), appID)
	if err != nil {
		slog.Error("github callback: failed to list connections", "err", err)
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}

	var existingConnID *uuid.UUID
	for _, c := range existingConns {
		if c.Type != "github" {
			continue
		}
		var cfg map[string]any
		if err := json.Unmarshal(c.Config, &cfg); err != nil {
			continue
		}
		if iid, ok := cfg["installation_id"].(float64); ok && int64(iid) == installationID {
			existingConnID = &c.ID
			break
		}
	}

	if existingConnID != nil {
		// Update existing connection (re-install / modify permissions).
		_, err = queries.UpdateConnection(r.Context(), db.UpdateConnectionParams{
			ID:        *existingConnID,
			Name:      fmt.Sprintf("GitHub: %s", accountLogin),
			Type:      "github",
			Direction: "two_way",
			Config:    configJSON,
			Status:    "active",
			UserID:    userID,
		})
		if err != nil {
			slog.Error("github callback: failed to update connection", "err", err)
			jsonError(w, "failed to update connection", http.StatusInternalServerError)
			return
		}
		slog.Info("github callback: updated existing connection",
			"connection_id", existingConnID, "installation_id", installationID, "setup_action", setupAction)
	} else {
		// Create new connection.
		conn, err := queries.CreateConnection(r.Context(), db.CreateConnectionParams{
			UserID:    userID,
			AppID:     appID,
			Name:      fmt.Sprintf("GitHub: %s", accountLogin),
			Type:      "github",
			Direction: "two_way",
			Config:    configJSON,
			Status:    "active",
		})
		if err != nil {
			slog.Error("github callback: failed to create connection", "err", err)
			jsonError(w, "failed to create connection", http.StatusInternalServerError)
			return
		}
		slog.Info("github callback: created new connection",
			"connection_id", conn.ID, "installation_id", installationID, "setup_action", setupAction)
	}

	// Redirect browser back to the connections page.
	http.Redirect(w, r, "/connections?github=installed", http.StatusFound)
}

// --- Repository Management ---

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

	queries, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
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
		jsonError(w, "invalid connection config", http.StatusInternalServerError)
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
			jsonError(w, "failed to parse GitHub response", http.StatusInternalServerError)
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
		slog.Error("github: failed to list db repos", "err", err)
		jsonError(w, "failed to load repo state", http.StatusInternalServerError)
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

	queries, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
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

	for _, repo := range repos {
		_, err := queries.UpsertGitHubRepo(r.Context(), db.UpsertGitHubRepoParams{
			ConnectionID:  connID,
			RepoFullName:  repo.RepoFullName,
			RepoID:        repo.RepoID,
			DefaultBranch: repo.DefaultBranch,
			Enabled:       repo.Enabled,
		})
		if err != nil {
			slog.Error("github: failed to upsert repo", "err", err, "repo", repo.RepoFullName)
			jsonError(w, "failed to update repos", http.StatusInternalServerError)
			return
		}
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
