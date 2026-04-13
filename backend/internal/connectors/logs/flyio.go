package logs

import (
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
	flyPollLimit       = 200
)

// FlyioConfig represents the fields stored in a Fly.io connection's config JSONB.
type FlyioConfig struct {
	AppName          string `json:"app_name"`
	APIToken         string `json:"api_token"`
	PollIntervalSecs int    `json:"poll_interval_secs"`
}

// Flyio polls the Fly.io Machines API for application logs.
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

func (f *Flyio) Connect(ctx context.Context) error {
	// Validate the token by listing machines for the app.
	endpoint := fmt.Sprintf("%s/v1/apps/%s/machines", f.apiBase, url.PathEscape(f.config.AppName))
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

// Poll fetches recent logs from the Fly.io Machines API Nats-backed log endpoint.
func (f *Flyio) Poll(ctx context.Context, queries *db.Queries) error {
	f.mu.Lock()
	cursor := f.cursor
	f.mu.Unlock()

	endpoint := fmt.Sprintf("%s/v1/apps/%s/machines", f.apiBase, url.PathEscape(f.config.AppName))
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

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return fmt.Errorf("flyio: read body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("flyio: rate limited (429), will retry next interval")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("flyio: api error %d: %s", resp.StatusCode, string(body))
	}

	// Parse machine list to get IDs, then fetch logs per machine.
	var machines []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		State string `json:"state"`
	}
	if err := json.Unmarshal(body, &machines); err != nil {
		return fmt.Errorf("flyio: parse machines: %w", err)
	}

	var maxTS time.Time
	var insertCount int

	for _, machine := range machines {
		if machine.State != "started" && machine.State != "running" {
			continue
		}
		count, ts, err := f.pollMachineLogs(ctx, queries, machine.ID, machine.Name, cursor)
		if err != nil {
			slog.Error("flyio: poll machine logs", "machine", machine.ID, "err", err)
			continue
		}
		insertCount += count
		if ts.After(maxTS) {
			maxTS = ts
		}
	}

	if !maxTS.IsZero() {
		f.mu.Lock()
		f.cursor = maxTS
		f.mu.Unlock()
	}

	if insertCount > 0 {
		slog.Info("flyio poll complete", "app", f.config.AppName, "rows", insertCount, "connection_id", f.connectionID)
	}
	return nil
}

func (f *Flyio) pollMachineLogs(ctx context.Context, queries *db.Queries, machineID, machineName string, since time.Time) (int, time.Time, error) {
	endpoint := fmt.Sprintf("%s/v1/apps/%s/machines/%s/logs?limit=%d",
		f.apiBase, url.PathEscape(f.config.AppName), url.PathEscape(machineID), flyPollLimit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, time.Time{}, err
	}
	req.Header.Set("Authorization", "Bearer "+f.config.APIToken)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return 0, time.Time{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return 0, time.Time{}, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return 0, time.Time{}, fmt.Errorf("flyio: rate limited (429), back off")
	}
	if resp.StatusCode != http.StatusOK {
		return 0, time.Time{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	var logs []struct {
		Timestamp string `json:"timestamp"`
		Message   string `json:"message"`
		Level     string `json:"level"`
	}
	if err := json.Unmarshal(body, &logs); err != nil {
		return 0, time.Time{}, err
	}

	var count int
	var maxTS time.Time

	for _, entry := range logs {
		ts, parseErr := time.Parse(time.RFC3339Nano, entry.Timestamp)
		if parseErr != nil {
			slog.Warn("flyio: unparseable timestamp, using now", "raw", entry.Timestamp, "err", parseErr, "connection_id", f.connectionID)
			ts = time.Now()
		}
		if !ts.After(since) {
			continue
		}

		payload, _ := json.Marshal(map[string]any{
			"timestamp": entry.Timestamp,
			"message":   entry.Message,
			"level":     entry.Level,
			"machine":   machineName,
		})

		severity := normalizeSev(entry.Level)

		_, err := queries.InsertLogEntry(ctx, db.InsertLogEntryParams{
			ConnectionID: f.connectionID,
			SourceType:   "flyio/" + f.config.AppName,
			Severity:     severity,
			Payload:      payload,
			UserID:       f.userID,
			AppID:        pgtype.UUID{Bytes: f.appID, Valid: true},
		})
		if err != nil {
			slog.Error("flyio: insert log entry", "err", err, "connection_id", f.connectionID)
			return count, maxTS, fmt.Errorf("flyio: insert log entry: %w", err)
		}

		count++
		if ts.After(maxTS) {
			maxTS = ts
		}
	}

	// Advance cursor past the last seen timestamp to avoid re-processing
	// entries with an identical timestamp on the next poll.
	if !maxTS.IsZero() {
		maxTS = maxTS.Add(time.Nanosecond)
	}

	return count, maxTS, nil
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
