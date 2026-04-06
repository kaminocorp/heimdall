package database

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"
)

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
	if cfg.Port == 0 {
		cfg.Port = 5432
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "require"
	}

	// Build connection string with read-only enforcement.
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&default_transaction_read_only=on",
		url.PathEscape(cfg.User),
		url.PathEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		url.PathEscape(cfg.Database),
		cfg.SSLMode,
	)

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

// Query executes a read-only SQL query and returns the results as a slice of maps.
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

	for rows.Next() {
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
