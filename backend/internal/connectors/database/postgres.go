package database

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"

	"github.com/jackc/pgx/v5"
)

// validHost matches a hostname (labels separated by dots), an IPv4 address,
// or a bracketed IPv6 address.  Rejects values containing query-string
// metacharacters (?, &, =) that could inject connection-string parameters.
var validHost = regexp.MustCompile(`^(\[[0-9a-fA-F:]+\]|[a-zA-Z0-9._-]+)$`)

// validSSLModes is the set of PostgreSQL SSL modes that can safely appear in
// a connection URL.  Validating against this set prevents injection of extra
// query parameters (e.g. "require&default_transaction_read_only=off").
var validSSLModes = map[string]bool{
	"disable":     true,
	"require":     true,
	"verify-ca":   true,
	"verify-full": true,
}

// Postgres connects to a user's monitored PostgreSQL database and executes
// read-only queries on behalf of the agent.
type Postgres struct {
	connStr string
	conn    *pgx.Conn
}

// postgresConfig represents the fields stored in a connection's config JSONB.
type postgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

// New creates a Postgres connector from a connection's config JSONB.
func New(configJSON json.RawMessage) (*Postgres, error) {
	var cfg postgresConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, fmt.Errorf("postgres: invalid config: %w", err)
	}

	if cfg.Host == "" || cfg.Database == "" || cfg.User == "" {
		return nil, fmt.Errorf("postgres: host, database, and user are required")
	}
	if !validHost.MatchString(cfg.Host) {
		return nil, fmt.Errorf("postgres: invalid host %q", cfg.Host)
	}
	if cfg.Port == 0 {
		cfg.Port = 5432
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("postgres: port must be between 1 and 65535")
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "require"
	}
	if !validSSLModes[cfg.SSLMode] {
		return nil, fmt.Errorf("postgres: invalid ssl_mode %q (must be one of: disable, require, verify-ca, verify-full)", cfg.SSLMode)
	}

	// Build connection string with read-only enforcement.
	// Use url.URL struct builder so userinfo characters (@, :) are correctly
	// percent-encoded — url.PathEscape does not escape these.
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:     cfg.Database,
		RawQuery: fmt.Sprintf("sslmode=%s&default_transaction_read_only=on", cfg.SSLMode),
	}
	connStr := u.String()

	return &Postgres{connStr: connStr}, nil
}

// Connect establishes a connection to the monitored database.
func (p *Postgres) Connect(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, p.connStr)
	if err != nil {
		return fmt.Errorf("postgres: connect: %w", err)
	}
	p.conn = conn
	return nil
}

// maxQueryRows is the maximum number of rows returned from a single query.
// This prevents LLM-generated queries like "SELECT * FROM large_table"
// from exhausting the agent process's memory.
const maxQueryRows = 1000

// Query executes a read-only SQL query and returns the results as a slice of maps.
// Results are capped at maxQueryRows to prevent unbounded memory consumption.
func (p *Postgres) Query(ctx context.Context, sql string) ([]map[string]any, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("postgres: not connected")
	}

	rows, err := p.conn.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("postgres: query: %w", err)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	var results []map[string]any
	truncated := false

	for rows.Next() {
		if len(results) >= maxQueryRows {
			truncated = true
			break
		}
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("postgres: scan: %w", err)
		}

		row := make(map[string]any, len(fields))
		for i, fd := range fields {
			row[fd.Name] = values[i]
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: rows: %w", err)
	}

	if truncated {
		results = append(results, map[string]any{
			"_truncated": fmt.Sprintf("Results limited to %d rows. Add a LIMIT clause for narrower queries.", maxQueryRows),
		})
	}

	return results, nil
}

// Health checks whether the connection is still alive.
func (p *Postgres) Health(ctx context.Context) error {
	if p.conn == nil {
		return fmt.Errorf("postgres: not connected")
	}
	return p.conn.Ping(ctx)
}

// Close closes the database connection.
// Uses context.Background so teardown is not cancelled by an expired caller context.
func (p *Postgres) Close() error {
	if p.conn != nil {
		return p.conn.Close(context.Background())
	}
	return nil
}
