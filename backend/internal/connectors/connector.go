package connectors

import "context"

// Connector is the common interface for all external integrations.
type Connector interface {
	Connect(ctx context.Context) error
	Health(ctx context.Context) error
	Close() error
}

// StreamConnector extends Connector with one-way streaming capability.
type StreamConnector interface {
	Connector
	Stream(ctx context.Context, out chan<- []byte) error
}

// QueryConnector extends Connector with on-demand query capability.
type QueryConnector interface {
	Connector
	Query(ctx context.Context, query string) (any, error)
}
