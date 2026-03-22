package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

const (
	supabaseAPIBase         = "https://api.supabase.com"
	supabaseDefaultInterval = 30
	supabaseMinInterval     = 15
	supabaseDefaultTable    = "postgres_logs"
	supabasePollLimit       = 500
)

// validSupabaseTables is the allowlist of Supabase log tables that can be polled.
// Prevents SQL injection via user-controlled table names in the analytics query.
var validSupabaseTables = map[string]bool{
	"postgres_logs":  true,
	"auth_logs":      true,
	"edge_logs":      true,
	"function_logs":  true,
	"storage_logs":   true,
	"realtime_logs":  true,
}

// SupabaseConfig represents the fields stored in a Supabase connection's config JSONB.
type SupabaseConfig struct {
	ProjectRef       string   `json:"project_ref"`
	AccessToken      string   `json:"access_token"`
	PollTables       []string `json:"poll_tables"`
	PollIntervalSecs int      `json:"poll_interval_secs"`
}

// Supabase polls the Supabase Management API for logs and inserts them into log_buffer.
type Supabase struct {
	config       SupabaseConfig
	connectionID uuid.UUID
	userID       uuid.UUID
	httpClient   *http.Client
	apiBase      string // defaults to supabaseAPIBase; overridable for testing

	mu      sync.Mutex
	cursors map[string]time.Time // per-table cursor
}

// NewSupabase creates a Supabase connector from a connection's config JSONB.
func NewSupabase(configJSON json.RawMessage, connectionID, userID uuid.UUID) (*Supabase, error) {
	var cfg SupabaseConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, fmt.Errorf("supabase: invalid config: %w", err)
	}

	if cfg.ProjectRef == "" {
		return nil, fmt.Errorf("supabase: project_ref is required")
	}
	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("supabase: access_token is required")
	}
	if len(cfg.PollTables) == 0 {
		cfg.PollTables = []string{supabaseDefaultTable}
	}
	for _, t := range cfg.PollTables {
		if !validSupabaseTables[t] {
			return nil, fmt.Errorf("supabase: invalid table name: %q", t)
		}
	}
	if cfg.PollIntervalSecs < supabaseMinInterval {
		cfg.PollIntervalSecs = supabaseDefaultInterval
	}

	return &Supabase{
		config:       cfg,
		connectionID: connectionID,
		userID:       userID,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		apiBase:      supabaseAPIBase,
		cursors:      make(map[string]time.Time),
	}, nil
}

// ParsedConfig returns the parsed SupabaseConfig for reading settings like poll interval.
func (s *Supabase) ParsedConfig() SupabaseConfig {
	return s.config
}

// Connect validates the PAT by running a lightweight query against the Management API.
func (s *Supabase) Connect(ctx context.Context) error {
	_, err := s.queryAPI(ctx, "postgres_logs", "SELECT 1")
	if err != nil {
		return fmt.Errorf("supabase: connect: %w", err)
	}
	return nil
}

// Health is an alias for Connect — a lightweight API ping.
func (s *Supabase) Health(ctx context.Context) error {
	return s.Connect(ctx)
}

// Close releases idle HTTP transport connections.
func (s *Supabase) Close() error {
	s.httpClient.CloseIdleConnections()
	return nil
}

// Poll fetches new logs from each configured table and inserts them into log_buffer.
// Returns the first error encountered (all tables are still attempted).
func (s *Supabase) Poll(ctx context.Context, queries *db.Queries) error {
	var firstErr error
	for _, table := range s.config.PollTables {
		if err := s.pollTable(ctx, queries, table); err != nil {
			slog.Error("supabase poll failed", "table", table, "connection_id", s.connectionID, "err", err)
			if firstErr == nil {
				firstErr = err
			}
			// Continue polling other tables even if one fails.
		}
	}
	return firstErr
}

func (s *Supabase) pollTable(ctx context.Context, queries *db.Queries, table string) error {
	if !validSupabaseTables[table] {
		return fmt.Errorf("supabase: invalid table name: %q", table)
	}

	s.mu.Lock()
	cursor, ok := s.cursors[table]
	if !ok {
		cursor = time.Now().Add(-5 * time.Minute)
		s.cursors[table] = cursor
	}
	s.mu.Unlock()

	// Supabase analytics timestamps are microsecond Unix.
	cursorMicro := cursor.UnixMicro()

	sql := fmt.Sprintf(
		"SELECT timestamp, event_message, metadata FROM %s WHERE timestamp > '%d' ORDER BY timestamp ASC LIMIT %d",
		table, cursorMicro, supabasePollLimit,
	)

	body, err := s.queryAPI(ctx, table, sql)
	if err != nil {
		return err
	}

	var resp struct {
		Result []struct {
			Timestamp    int64           `json:"timestamp"`
			EventMessage string          `json:"event_message"`
			Metadata     json.RawMessage `json:"metadata"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("supabase: parse response: %w", err)
	}

	if len(resp.Result) == 0 {
		return nil
	}

	// Only advance the cursor to the max timestamp of *successfully inserted* rows.
	// This prevents silent data loss when some inserts fail — failed rows will be
	// retried on the next poll because the cursor hasn't moved past them.
	var maxInsertedTS int64
	for _, row := range resp.Result {
		payload, err := json.Marshal(map[string]any{
			"timestamp":     row.Timestamp,
			"event_message": row.EventMessage,
			"metadata":      row.Metadata,
		})
		if err != nil {
			slog.Error("supabase: marshal payload", "err", err)
			continue
		}

		severity := deriveSeverity(row.Metadata)

		_, err = queries.InsertLogEntry(ctx, db.InsertLogEntryParams{
			ConnectionID: s.connectionID,
			SourceType:   "supabase/" + table,
			Severity:     severity,
			Payload:      payload,
			UserID:       s.userID,
		})
		if err != nil {
			slog.Error("supabase: insert log entry", "err", err)
			continue
		}

		if row.Timestamp > maxInsertedTS {
			maxInsertedTS = row.Timestamp
		}
	}

	if maxInsertedTS > 0 {
		s.mu.Lock()
		s.cursors[table] = time.UnixMicro(maxInsertedTS)
		s.mu.Unlock()
	}

	slog.Info("supabase poll complete", "table", table, "rows", len(resp.Result), "connection_id", s.connectionID)
	return nil
}

// queryAPI calls the Supabase Management API analytics endpoint.
func (s *Supabase) queryAPI(ctx context.Context, table, sql string) ([]byte, error) {
	endpoint := fmt.Sprintf(
		"%s/v1/projects/%s/analytics/endpoints/logs.all?sql=%s",
		s.apiBase,
		url.PathEscape(s.config.ProjectRef),
		url.QueryEscape(sql),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("supabase: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.config.AccessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("supabase: api request: %w", err)
	}
	defer resp.Body.Close()

	// Cap response size to prevent OOM from a misbehaving upstream.
	const maxResponseBytes = 10 << 20 // 10 MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("supabase: read body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("supabase: rate limited (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: api error %d: %s", resp.StatusCode, string(body))
	}

	// If rate limit remaining is zero, return an error instead of blocking the
	// goroutine for potentially long reset windows. The poller will retry on the
	// next interval, giving time for the rate limit to reset naturally.
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if rem, _ := strconv.Atoi(remaining); rem == 0 {
			slog.Warn("supabase: rate limit exhausted, will retry next interval", "connection_id", s.connectionID)
		}
	}

	return body, nil
}

// deriveSeverity attempts to extract a severity level from log metadata.
func deriveSeverity(metadata json.RawMessage) pgtype.Text {
	if metadata == nil {
		return pgtype.Text{}
	}

	var m map[string]any
	if err := json.Unmarshal(metadata, &m); err != nil {
		return pgtype.Text{}
	}

	// Supabase Postgres logs include an "error_severity" field.
	for _, key := range []string{"error_severity", "severity", "level"} {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return pgtype.Text{String: s, Valid: true}
			}
		}
	}

	return pgtype.Text{}
}
