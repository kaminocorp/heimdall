package logs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

const (
	flyAPIBase         = "https://api.machines.dev"
	flyDefaultInterval = 30
	flyMinInterval     = 15
	// flyMaxBodyBytes caps the total bytes read from the NDJSON response to
	// prevent unbounded memory growth if the endpoint returns an unexpectedly
	// large payload.
	flyMaxBodyBytes = 10 << 20 // 10 MiB
)

// FlyioConfig represents the fields stored in a Fly.io connection's config JSONB.
type FlyioConfig struct {
	AppName          string `json:"app_name"`
	APIToken         string `json:"api_token"`
	PollIntervalSecs int    `json:"poll_interval_secs"`
}

// Flyio polls the Fly.io app-level logs endpoint for application logs.
type Flyio struct {
	config       FlyioConfig
	connectionID uuid.UUID
	userID       uuid.UUID
	appID        uuid.UUID
	httpClient   *http.Client
	apiBase      string

	mu     sync.Mutex
	cursor time.Time
}

// NewFlyio creates a Fly.io connector from a connection's config JSONB.
func NewFlyio(configJSON json.RawMessage, connectionID, userID, appID uuid.UUID) (*Flyio, error) {
	var cfg FlyioConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, fmt.Errorf("flyio: invalid config: %w", err)
	}
	if cfg.AppName == "" {
		return nil, fmt.Errorf("flyio: app_name is required")
	}
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("flyio: api_token is required")
	}
	if cfg.PollIntervalSecs < flyMinInterval {
		cfg.PollIntervalSecs = flyDefaultInterval
	}

	return &Flyio{
		config:       cfg,
		connectionID: connectionID,
		userID:       userID,
		appID:        appID,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		apiBase:      flyAPIBase,
		cursor:       time.Now().Add(-5 * time.Minute),
	}, nil
}

func (f *Flyio) ParsedConfig() FlyioConfig { return f.config }

// Connect validates the API token and app name by hitting the app-level logs
// endpoint with a short time window.
func (f *Flyio) Connect(ctx context.Context) error {
	endpoint := f.logsEndpoint(time.Now().Add(-1 * time.Minute))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("flyio: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+f.config.APIToken)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("flyio: api request: %w", err)
	}
	defer resp.Body.Close()
	io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("flyio: invalid API token")
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("flyio: app %q not found", f.config.AppName)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("flyio: api error %d", resp.StatusCode)
	}
	return nil
}

func (f *Flyio) Health(ctx context.Context) error { return f.Connect(ctx) }

func (f *Flyio) Close() error {
	f.httpClient.CloseIdleConnections()
	return nil
}

// logsEndpoint builds the app-level NDJSON logs URL with a start_time cursor.
func (f *Flyio) logsEndpoint(since time.Time) string {
	return fmt.Sprintf("%s/v1/apps/%s/logs?start_time=%s",
		f.apiBase,
		url.PathEscape(f.config.AppName),
		url.QueryEscape(since.UTC().Format(time.RFC3339Nano)),
	)
}

// flyLogEntry represents a single line from the NDJSON app-level logs response.
type flyLogEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level"`
	Instance  string `json:"instance"`
	Region    string `json:"region"`
	Meta      any    `json:"meta"`
}

// Poll fetches recent logs from the Fly.io app-level NDJSON endpoint.
// This replaces the previous N+1 per-machine approach with a single API call
// that returns logs across all machines (including stopped/crashed).
func (f *Flyio) Poll(ctx context.Context, queries *db.Queries) error {
	f.mu.Lock()
	cursor := f.cursor
	f.mu.Unlock()

	endpoint := f.logsEndpoint(cursor)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("flyio: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+f.config.APIToken)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("flyio: api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("flyio: rate limited (429), will retry next interval")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("flyio: api error %d: %s", resp.StatusCode, string(body))
	}

	// Parse NDJSON: one JSON object per line.
	scanner := bufio.NewScanner(io.LimitReader(resp.Body, flyMaxBodyBytes))
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20) // up to 1 MiB per line

	var maxTS time.Time
	var insertCount int

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry flyLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			slog.Warn("flyio: skipping malformed NDJSON line",
				"err", err, "connection_id", f.connectionID)
			continue
		}

		ts, parseErr := time.Parse(time.RFC3339Nano, entry.Timestamp)
		if parseErr != nil {
			slog.Warn("flyio: unparseable timestamp, using now",
				"raw", entry.Timestamp, "err", parseErr, "connection_id", f.connectionID)
			ts = time.Now()
		}
		if !ts.After(cursor) {
			continue
		}

		payload, err := json.Marshal(map[string]any{
			"timestamp": entry.Timestamp,
			"message":   entry.Message,
			"level":     entry.Level,
			"instance":  entry.Instance,
			"region":    entry.Region,
			"meta":      entry.Meta,
		})
		if err != nil {
			slog.Error("flyio: marshal payload", "err", err, "connection_id", f.connectionID)
			continue
		}

		severity := normalizeSev(entry.Level)

		_, err = queries.InsertLogEntry(ctx, db.InsertLogEntryParams{
			ConnectionID: f.connectionID,
			SourceType:   "flyio/" + f.config.AppName,
			Severity:     severity,
			Payload:      payload,
			UserID:       f.userID,
			AppID:        f.appID,
		})
		if err != nil {
			slog.Error("flyio: insert log entry", "err", err, "connection_id", f.connectionID)
			continue
		}

		insertCount++
		if ts.After(maxTS) {
			maxTS = ts
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("flyio: scanner error", "err", err, "connection_id", f.connectionID)
	}

	// Advance cursor past the last seen timestamp to avoid re-processing
	// entries with an identical timestamp on the next poll.
	if !maxTS.IsZero() {
		f.mu.Lock()
		f.cursor = maxTS.Add(time.Nanosecond)
		f.mu.Unlock()
	}

	if insertCount > 0 {
		slog.Info("flyio poll complete",
			"app", f.config.AppName, "rows", insertCount, "connection_id", f.connectionID)
	}
	return nil
}

func normalizeSev(level string) pgtype.Text {
	var s string
	switch level {
	case "error", "err":
		s = "error"
	case "warn", "warning":
		s = "warning"
	case "debug":
		s = "debug"
	default:
		s = "info"
	}
	return pgtype.Text{String: s, Valid: true}
}
