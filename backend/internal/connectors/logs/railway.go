package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

const (
	railwayAPIBase         = "https://backboard.railway.app/graphql/v2"
	railwayDefaultInterval = 60
	railwayMinInterval     = 30
	railwayPollLimit       = 200
)

// RailwayConfig represents the fields stored in a Railway connection's config JSONB.
type RailwayConfig struct {
	APIToken         string `json:"api_token"`
	ProjectID        string `json:"project_id"`
	ServiceID        string `json:"service_id"`
	EnvironmentID    string `json:"environment_id"`
	PollIntervalSecs int    `json:"poll_interval_secs"`
}

// Railway polls the Railway GraphQL API for deployment logs.
type Railway struct {
	config       RailwayConfig
	connectionID uuid.UUID
	userID       uuid.UUID
	appID        uuid.UUID
	httpClient   *http.Client
	apiBase      string

	mu     sync.Mutex
	cursor time.Time
}

func NewRailway(configJSON json.RawMessage, connectionID, userID, appID uuid.UUID) (*Railway, error) {
	var cfg RailwayConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, fmt.Errorf("railway: invalid config: %w", err)
	}
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("railway: api_token is required")
	}
	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("railway: project_id is required")
	}
	if cfg.PollIntervalSecs < railwayMinInterval {
		cfg.PollIntervalSecs = railwayDefaultInterval
	}

	return &Railway{
		config:       cfg,
		connectionID: connectionID,
		userID:       userID,
		appID:        appID,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		apiBase:      railwayAPIBase,
		cursor:       time.Now().Add(-5 * time.Minute),
	}, nil
}

func (r *Railway) ParsedConfig() RailwayConfig { return r.config }

func (r *Railway) Connect(ctx context.Context) error {
	// Validate token with a simple project query.
	query := `query($id: String!) { project(id: $id) { id name } }`
	_, err := r.graphQL(ctx, query, map[string]any{"id": r.config.ProjectID})
	return err
}

func (r *Railway) Health(ctx context.Context) error { return r.Connect(ctx) }

func (r *Railway) Close() error {
	r.httpClient.CloseIdleConnections()
	return nil
}

func (r *Railway) Poll(ctx context.Context, queries *db.Queries) error {
	r.mu.Lock()
	cursor := r.cursor
	r.mu.Unlock()

	// Query deployment logs via the Railway GraphQL API.
	query := `query($projectId: String!, $limit: Int!, $since: DateTime) {
		deploymentLogs(projectId: $projectId, limit: $limit, since: $since) {
			timestamp
			message
			severity
			tags
		}
	}`

	variables := map[string]any{
		"projectId": r.config.ProjectID,
		"limit":     railwayPollLimit,
		"since":     cursor.Format(time.RFC3339),
	}

	body, err := r.graphQL(ctx, query, variables)
	if err != nil {
		return fmt.Errorf("railway: poll: %w", err)
	}

	var resp struct {
		Data struct {
			DeploymentLogs []struct {
				Timestamp string            `json:"timestamp"`
				Message   string            `json:"message"`
				Severity  string            `json:"severity"`
				Tags      map[string]string `json:"tags"`
			} `json:"deploymentLogs"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("railway: parse response: %w", err)
	}

	if len(resp.Errors) > 0 {
		return fmt.Errorf("railway: graphql error: %s", resp.Errors[0].Message)
	}

	if len(resp.Data.DeploymentLogs) == 0 {
		return nil
	}

	var maxTS time.Time
	var insertCount int

	for _, entry := range resp.Data.DeploymentLogs {
		ts, parseErr := time.Parse(time.RFC3339Nano, entry.Timestamp)
		if parseErr != nil {
			slog.Warn("railway: unparseable timestamp, using now", "raw", entry.Timestamp, "err", parseErr, "connection_id", r.connectionID)
			ts = time.Now()
		}
		if !ts.After(cursor) {
			continue
		}

		payload, _ := json.Marshal(map[string]any{
			"timestamp": entry.Timestamp,
			"message":   entry.Message,
			"severity":  entry.Severity,
			"tags":      entry.Tags,
		})

		severity := normalizeSev(entry.Severity)

		_, err := queries.InsertLogEntry(ctx, db.InsertLogEntryParams{
			ConnectionID: r.connectionID,
			SourceType:   "railway/" + r.config.ProjectID,
			Severity:     severity,
			Payload:      payload,
			UserID:       r.userID,
			AppID:        pgtype.UUID{Bytes: r.appID, Valid: true},
		})
		if err != nil {
			slog.Error("railway: insert log entry", "err", err, "connection_id", r.connectionID)
			return fmt.Errorf("railway: insert log entry: %w", err)
		}

		insertCount++
		if ts.After(maxTS) {
			maxTS = ts
		}
	}

	// Advance cursor past the last seen timestamp to avoid re-processing
	// entries with an identical timestamp on the next poll.
	if !maxTS.IsZero() {
		r.mu.Lock()
		r.cursor = maxTS.Add(time.Nanosecond)
		r.mu.Unlock()
	}

	if insertCount > 0 {
		slog.Info("railway poll complete", "project", r.config.ProjectID, "rows", insertCount, "connection_id", r.connectionID)
	}
	return nil
}

func (r *Railway) graphQL(ctx context.Context, query string, variables map[string]any) ([]byte, error) {
	payload := map[string]any{"query": query}
	if variables != nil {
		payload["variables"] = variables
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("railway: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.apiBase, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("railway: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.config.APIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("railway: api request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("railway: read body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("railway: invalid API token")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("railway: rate limited (429), will retry next interval")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("railway: api error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
