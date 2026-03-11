package codebase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	ghpkg "github.com/hejijunhao/heimdall/backend/internal/github"
)

const (
	maxSearchResults = 20
	maxTreeEntries   = 5000
	maxFileSize      = 1 << 20  // 1MB — skip files larger than this
	truncateAt       = 50 * 1024 // 50KB — truncate content beyond this
	apiTimeout       = 10 * time.Second
)

// GitHub implements the QueryConnector interface for code search and file reading.
type GitHub struct {
	ghClient       *ghpkg.Client
	installationID int64
	repos          []string // enabled repo full names (e.g. "org/repo")
	token          string   // cached installation token
}

// New creates a GitHub connector from a connection's config.
func New(configJSON json.RawMessage, ghClient *ghpkg.Client, repos []string) (*GitHub, error) {
	if ghClient == nil {
		return nil, fmt.Errorf("github connector: GitHub App client not configured")
	}

	var cfg struct {
		InstallationID int64 `json:"installation_id"`
	}
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, fmt.Errorf("github connector: invalid config: %w", err)
	}
	if cfg.InstallationID == 0 {
		return nil, fmt.Errorf("github connector: missing installation_id")
	}

	return &GitHub{
		ghClient:       ghClient,
		installationID: cfg.InstallationID,
		repos:          repos,
	}, nil
}

// Connect obtains an installation token.
func (g *GitHub) Connect(ctx context.Context) error {
	token, err := g.ghClient.InstallationTokenFor(ctx, g.installationID)
	if err != nil {
		return fmt.Errorf("github connector: %w", err)
	}
	g.token = token.Token
	return nil
}

// Health checks the rate limit endpoint.
func (g *GitHub) Health(ctx context.Context) error {
	resp, err := g.ghClient.APIRequest(ctx, "GET", "https://api.github.com/rate_limit", g.token, nil)
	if err != nil {
		return fmt.Errorf("github connector: health check: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github connector: health check failed (%d)", resp.StatusCode)
	}
	return nil
}

// Close is a no-op for the GitHub connector.
func (g *GitHub) Close() error {
	return nil
}

// Query executes a code action. The query string is a JSON-encoded action descriptor.
func (g *GitHub) Query(ctx context.Context, query string) (any, error) {
	var action struct {
		Action string `json:"action"`
		Query  string `json:"query"`
		Path   string `json:"path"`
		Repo   string `json:"repo"`
		Ref    string `json:"ref"`
	}
	if err := json.Unmarshal([]byte(query), &action); err != nil {
		return nil, fmt.Errorf("github connector: invalid action: %w", err)
	}

	switch action.Action {
	case "search_code":
		return g.searchCode(ctx, action.Query, action.Repo)
	case "read_file":
		return g.readFile(ctx, action.Path, action.Repo, action.Ref)
	case "list_tree":
		return g.listTree(ctx, action.Path, action.Repo, action.Ref)
	default:
		return nil, fmt.Errorf("github connector: unknown action: %s", action.Action)
	}
}

func (g *GitHub) searchCode(ctx context.Context, query, repo string) (any, error) {
	if query == "" {
		return nil, fmt.Errorf("search_code: query is required")
	}

	// Build the GitHub search query string (uses spaces as boolean AND).
	searchQuery := query
	if repo != "" {
		searchQuery = query + " repo:" + repo
	} else if len(g.repos) > 0 {
		// Search across all connected repos, capped at 20 to prevent
		// excessively long query strings (GitHub has URL length limits).
		repos := g.repos
		if len(repos) > 20 {
			repos = repos[:20]
		}
		parts := make([]string, len(repos))
		for i, r := range repos {
			parts[i] = "repo:" + r
		}
		searchQuery = query + " " + strings.Join(parts, " ")
	}

	// Use url.Values for proper query string encoding.
	params := url.Values{}
	params.Set("q", searchQuery)
	params.Set("per_page", fmt.Sprintf("%d", maxSearchResults))
	apiURL := "https://api.github.com/search/code?" + params.Encode()
	body, err := g.apiGet(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("search_code: %w", err)
	}

	var result struct {
		TotalCount int `json:"total_count"`
		Items      []struct {
			Name       string `json:"name"`
			Path       string `json:"path"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			HTMLURL string `json:"html_url"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("search_code: parse response: %w", err)
	}

	type searchResult struct {
		File       string `json:"file"`
		Path       string `json:"path"`
		Repository string `json:"repository"`
	}
	results := make([]searchResult, 0, len(result.Items))
	for _, item := range result.Items {
		results = append(results, searchResult{
			File:       item.Name,
			Path:       item.Path,
			Repository: item.Repository.FullName,
		})
	}

	return map[string]any{
		"total_count": result.TotalCount,
		"results":     results,
	}, nil
}

func (g *GitHub) readFile(ctx context.Context, path, repo, ref string) (any, error) {
	if path == "" {
		return nil, fmt.Errorf("read_file: path is required")
	}
	if repo == "" && len(g.repos) > 0 {
		repo = g.repos[0]
	}
	if repo == "" {
		return nil, fmt.Errorf("read_file: repo is required")
	}

	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("read_file: invalid repo format (expected owner/repo)")
	}

	// GitHub Contents API expects literal slashes — escape individual path segments only.
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		segments[i] = url.PathEscape(seg)
	}
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", url.PathEscape(parts[0]), url.PathEscape(parts[1]), strings.Join(segments, "/"))
	if ref != "" {
		apiURL += "?ref=" + url.QueryEscape(ref)
	}

	body, err := g.apiGet(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("read_file: %w", err)
	}

	var file struct {
		Name     string `json:"name"`
		Path     string `json:"path"`
		Size     int    `json:"size"`
		Encoding string `json:"encoding"`
		Content  string `json:"content"`
		Type     string `json:"type"`
	}
	if err := json.Unmarshal(body, &file); err != nil {
		return nil, fmt.Errorf("read_file: parse response: %w", err)
	}

	if file.Type == "dir" {
		return nil, fmt.Errorf("read_file: path is a directory, use list_tree instead")
	}

	// Skip large files.
	if file.Size > maxFileSize {
		return map[string]any{
			"path": file.Path,
			"size": file.Size,
			"note": fmt.Sprintf("[Binary or large file — %d bytes, skipped]", file.Size),
		}, nil
	}

	// Detect binary files.
	if file.Encoding != "base64" && file.Encoding != "" {
		return map[string]any{
			"path": file.Path,
			"note": "[Binary file — skipped]",
		}, nil
	}

	// Decode content.
	content, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(file.Content, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("read_file: decode content: %w", err)
	}

	contentStr := string(content)
	truncated := false
	if len(content) > truncateAt {
		contentStr = string(content[:truncateAt])
		truncated = true
	}

	result := map[string]any{
		"path":    file.Path,
		"size":    file.Size,
		"content": contentStr,
	}
	if truncated {
		result["note"] = fmt.Sprintf("[TRUNCATED — file is %d bytes]", file.Size)
	}
	return result, nil
}

func (g *GitHub) listTree(ctx context.Context, path, repo, ref string) (any, error) {
	if repo == "" && len(g.repos) > 0 {
		repo = g.repos[0]
	}
	if repo == "" {
		return nil, fmt.Errorf("list_tree: repo is required")
	}

	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("list_tree: invalid repo format (expected owner/repo)")
	}

	if ref == "" {
		ref = "HEAD"
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", url.PathEscape(parts[0]), url.PathEscape(parts[1]), url.PathEscape(ref))
	body, err := g.apiGet(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("list_tree: %w", err)
	}

	var tree struct {
		SHA       string `json:"sha"`
		Truncated bool   `json:"truncated"`
		Tree      []struct {
			Path string `json:"path"`
			Mode string `json:"mode"`
			Type string `json:"type"`
			Size int    `json:"size"`
		} `json:"tree"`
	}
	if err := json.Unmarshal(body, &tree); err != nil {
		return nil, fmt.Errorf("list_tree: parse response: %w", err)
	}

	// Filter by path prefix if specified.
	type entry struct {
		Path string `json:"path"`
		Type string `json:"type"`
		Size int    `json:"size,omitempty"`
	}
	var entries []entry
	for _, item := range tree.Tree {
		if path != "" && !strings.HasPrefix(item.Path, path) {
			continue
		}
		entries = append(entries, entry{
			Path: item.Path,
			Type: item.Type,
			Size: item.Size,
		})
		if len(entries) >= maxTreeEntries {
			break
		}
	}

	if entries == nil {
		entries = []entry{}
	}

	return map[string]any{
		"entries":   entries,
		"count":     len(entries),
		"truncated": tree.Truncated || len(entries) >= maxTreeEntries,
	}, nil
}

// apiGet makes an authenticated GET request to the GitHub API via ghClient.
func (g *GitHub) apiGet(ctx context.Context, apiURL string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	resp, err := g.ghClient.APIRequest(ctx, "GET", apiURL, g.token, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusNotFound:
		return nil, fmt.Errorf("not found")
	case http.StatusForbidden:
		return nil, fmt.Errorf("rate limited or forbidden (%d): %s", resp.StatusCode, string(body))
	default:
		return nil, fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(body))
	}
}
