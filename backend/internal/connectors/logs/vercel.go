package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

const (
	vercelAPIBase         = "https://api.vercel.com"
	vercelDefaultInterval = 60
	vercelMinInterval     = 30
	vercelPollLimit       = 100
)

// VercelConfig represents the fields stored in a Vercel connection's config JSONB.
type VercelConfig struct {
	APIToken         string `json:"api_token"`
	ProjectID        string `json:"project_id"`
	TeamID           string `json:"team_id"`
	PollIntervalSecs int    `json:"poll_interval_secs"`
}

// Vercel polls the Vercel REST API for deployment events and logs.
type Vercel struct {
	config       VercelConfig
	connectionID uuid.UUID
	userID       uuid.UUID
	httpClient   *http.Client
	apiBase      string

	mu     sync.Mutex
	cursor time.Time
}

func NewVercel(configJSON json.RawMessage, connectionID, userID uuid.UUID) (*Vercel, error) {
	var cfg VercelConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, fmt.Errorf("vercel: invalid config: %w", err)
	}
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("vercel: api_token is required")
	}
	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("vercel: project_id is required")
	}
	if cfg.PollIntervalSecs < vercelMinInterval {
		cfg.PollIntervalSecs = vercelDefaultInterval
	}

	return &Vercel{
		config:       cfg,
		connectionID: connectionID,
		userID:       userID,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		apiBase:      vercelAPIBase,
		cursor:       time.Now().Add(-5 * time.Minute),
	}, nil
}

func (v *Vercel) ParsedConfig() VercelConfig { return v.config }

func (v *Vercel) Connect(ctx context.Context) error {
	_, err := v.apiRequest(ctx, fmt.Sprintf("/v9/projects/%s", url.PathEscape(v.config.ProjectID)))
	return err
}

func (v *Vercel) Health(ctx context.Context) error { return v.Connect(ctx) }

func (v *Vercel) Close() error {
	v.httpClient.CloseIdleConnections()
	return nil
}

func (v *Vercel) Poll(ctx context.Context, queries *db.Queries) error {
	v.mu.Lock()
	cursor := v.cursor
	v.mu.Unlock()

	sinceMS := cursor.UnixMilli()
	path := fmt.Sprintf("/v6/deployments?projectId=%s&since=%d&limit=%d&state=READY,ERROR",
		url.QueryEscape(v.config.ProjectID), sinceMS, vercelPollLimit)

	body, err := v.apiRequest(ctx, path)
	if err != nil {
		return fmt.Errorf("vercel: poll deployments: %w", err)
	}

	var resp struct {
		Deployments []struct {
			UID       string `json:"uid"`
			Name      string `json:"name"`
			URL       string `json:"url"`
			State     string `json:"state"`
			CreatedAt int64  `json:"createdAt"`
			Ready     int64  `json:"ready"`
			Source    string `json:"source"` // "cli", "git", "api"
			Meta      map[string]string `json:"meta"`
		} `json:"deployments"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("vercel: parse response: %w", err)
	}

	if len(resp.Deployments) == 0 {
		return nil
	}

	var maxTS time.Time
	var insertCount int

	for _, d := range resp.Deployments {
		ts := time.UnixMilli(d.CreatedAt)
		if !ts.After(cursor) {
			continue
		}

		severity := pgtype.Text{String: "info", Valid: true}
		if strings.EqualFold(d.State, "ERROR") {
			severity = pgtype.Text{String: "error", Valid: true}
		}

		commitMsg := d.Meta["githubCommitMessage"]
		commitSha := d.Meta["githubCommitSha"]

		payload, _ := json.Marshal(map[string]any{
			"deployment_id": d.UID,
			"name":          d.Name,
			"url":           d.URL,
			"state":         d.State,
			"source":        d.Source,
			"created_at":    d.CreatedAt,
			"ready_at":      d.Ready,
			"commit_message": commitMsg,
			"commit_sha":    commitSha,
		})

		_, err := queries.InsertLogEntry(ctx, db.InsertLogEntryParams{
			ConnectionID: v.connectionID,
			SourceType:   "vercel/deployment",
			Severity:     severity,
			Payload:      payload,
			UserID:       v.userID,
		})
		if err != nil {
			slog.Error("vercel: insert log entry", "err", err, "connection_id", v.connectionID)
			return fmt.Errorf("vercel: insert log entry: %w", err)
		}

		insertCount++
		if ts.After(maxTS) {
			maxTS = ts
		}
	}

	// Advance cursor past the last seen timestamp to avoid re-processing
	// entries with an identical timestamp on the next poll.
	if !maxTS.IsZero() {
		v.mu.Lock()
		v.cursor = maxTS.Add(time.Nanosecond)
		v.mu.Unlock()
	}

	if insertCount > 0 {
		slog.Info("vercel poll complete", "project", v.config.ProjectID, "rows", insertCount, "connection_id", v.connectionID)
	}
	return nil
}

func (v *Vercel) apiRequest(ctx context.Context, path string) ([]byte, error) {
	endpoint := v.apiBase + path
	if v.config.TeamID != "" {
		sep := "?"
		if strings.Contains(path, "?") {
			sep = "&"
		}
		endpoint += sep + "teamId=" + url.QueryEscape(v.config.TeamID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("vercel: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+v.config.APIToken)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vercel: api request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("vercel: read body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("vercel: invalid API token or insufficient permissions")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("vercel: rate limited (429), will retry next interval")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vercel: api error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
