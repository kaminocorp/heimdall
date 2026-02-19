package database

import "context"

// Postgres implements the Connector and QueryConnector interfaces
// for PostgreSQL databases being monitored by Heimdall.
type Postgres struct {
	// TODO: add pgx pool, config
}

func New() *Postgres {
	return &Postgres{}
}

func (p *Postgres) Connect(ctx context.Context) error {
	// TODO: establish connection pool to monitored database
	return nil
}

func (p *Postgres) Health(ctx context.Context) error {
	// TODO: ping database
	return nil
}

func (p *Postgres) Query(ctx context.Context, query string) (any, error) {
	// TODO: execute read-only query
	return nil, nil
}

func (p *Postgres) Close() error {
	return nil
}
