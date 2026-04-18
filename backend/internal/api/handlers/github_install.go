package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// TestGitHubConnection tests a GitHub connection by checking that the stored
// installation id is still valid (i.e. GitHub still grants us an installation
// token). Returns a user-facing message rather than an error so the caller
// can render it verbatim in the connection-test modal.
//
// Lives here rather than in a github_repos.go file because that file was
// removed with the github_repos table in migration 036; github_install.go
// already owns GitHub-specific handler code.
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

	if _, err := s.GitHub.InstallationTokenFor(ctx, int64(installationID)); err != nil {
		return false, fmt.Sprintf("GitHub auth failed: %v", err)
	}

	return true, "GitHub connection active"
}

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
		jsonServerError(w, "failed to generate install URL", err)
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
	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	app, err := queries.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
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
		jsonServerError(w, "failed to build config", err)
		return
	}

	// Check for existing GitHub connection with same installation_id for this app.
	existingConns, err := queries.ListConnectionsByApp(r.Context(), appID)
	if err != nil {
		jsonServerError(w, "database error", err)
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

	// resultConnID carries the id through to the redirect URL so the frontend
	// can open the source selector for the *exact* connection that just
	// finished installing, even when the user has multiple GitHub installs.
	// Without this disambiguator the callback page falls back to first-match
	// which guesses wrong under multi-install usage.
	var resultConnID uuid.UUID

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
			jsonServerError(w, "failed to update connection", err)
			return
		}
		resultConnID = *existingConnID
		slog.Info("github callback: updated existing connection",
			"connection_id", resultConnID, "installation_id", installationID, "setup_action", setupAction)
	} else {
		// Create new connection. GitHub is always app-scoped — code search
		// binds to one app at a time, so there's no meaningful org-wide mode
		// to opt into here.
		conn, err := queries.CreateConnection(r.Context(), db.CreateConnectionParams{
			UserID:    userID,
			OrgID:     app.OrgID,
			AppID:     &appID,
			Name:      fmt.Sprintf("GitHub: %s", accountLogin),
			Type:      "github",
			Direction: "two_way",
			Config:    configJSON,
			Status:    "active",
		})
		if err != nil {
			jsonServerError(w, "failed to create connection", err)
			return
		}
		resultConnID = conn.ID
		slog.Info("github callback: created new connection",
			"connection_id", resultConnID, "installation_id", installationID, "setup_action", setupAction)
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	// Redirect browser back to the frontend connections page with the
	// connection id so the post-install source selector opens for the
	// correct install (multi-install disambiguation).
	redirectURL := fmt.Sprintf("%s/connections?github=installed&connection_id=%s",
		s.Config.FrontendURL, resultConnID)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
