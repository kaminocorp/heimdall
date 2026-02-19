package memory

import (
	"context"
	"net/http"
)

// Client is a thin HTTP client for the Elephantasm API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		apiKey:     apiKey,
	}
}

func (c *Client) RecordEvent(ctx context.Context, event Event) error {
	// TODO: POST event to Elephantasm
	return nil
}

func (c *Client) QueryMemories(ctx context.Context, query string) ([]Memory, error) {
	// TODO: search Elephantasm for relevant memories
	return nil, nil
}

func (c *Client) GetLessons(ctx context.Context, topic string) ([]Lesson, error) {
	// TODO: retrieve lessons from Elephantasm
	return nil, nil
}
