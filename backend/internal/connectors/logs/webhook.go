package logs

import "context"

// Webhook implements log ingestion via HTTP webhook.
// Applications POST log entries to Heimdall's ingestion endpoint.
type Webhook struct {
	// TODO: add config (endpoint path, auth token)
}

func NewWebhook() *Webhook {
	return &Webhook{}
}

func (w *Webhook) Connect(ctx context.Context) error {
	return nil
}

func (w *Webhook) Health(ctx context.Context) error {
	return nil
}

func (w *Webhook) Stream(ctx context.Context, out chan<- []byte) error {
	// TODO: buffer incoming webhook payloads and send to channel
	return nil
}

func (w *Webhook) Close() error {
	return nil
}
