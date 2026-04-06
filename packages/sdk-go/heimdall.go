// Package heimdall provides a lightweight Go SDK for sending logs to a Heimdall instance.
//
// Usage:
//
//	monitor := heimdall.New(heimdall.Options{
//	    Endpoint: "https://heimdall.example.com",
//	    Token:    "whk_...",
//	})
//	defer monitor.Shutdown()
//
//	monitor.Info("user.signup", map[string]any{"user_id": "123"})
//	monitor.Error("payment.failed", map[string]any{"order_id": "456"})
package heimdall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Options configures the Heimdall client.
type Options struct {
	Endpoint      string        // Heimdall backend URL (required)
	Token         string        // Webhook bearer token (required)
	BatchSize     int           // Flush after N entries (default: 25)
	FlushInterval time.Duration // Flush after this duration (default: 5s)
	MaxRetries    int           // Retry attempts for 5xx/network errors (default: 3)
	RetryDelay    time.Duration // Base delay for exponential backoff (default: 1s)
	OnError       func(err error, entries []LogEntry) // Called on permanent failure
	HTTPClient    *http.Client  // Custom HTTP client (default: 30s timeout)
}

// LogEntry is a single log entry sent to Heimdall.
type LogEntry struct {
	SourceType string         `json:"source_type"`
	Severity   string         `json:"severity"`
	Payload    map[string]any `json:"payload"`
}

// Client is a Heimdall logging client with automatic batching and retry.
type Client struct {
	endpoint      string
	token         string
	batchSize     int
	flushInterval time.Duration
	maxRetries    int
	retryDelay    time.Duration
	onError       func(error, []LogEntry)
	httpClient    *http.Client

	mu     sync.Mutex
	wg     sync.WaitGroup
	buffer []LogEntry
	timer  *time.Timer
	closed bool
}

// New creates a new Heimdall client.
func New(opts Options) *Client {
	if opts.BatchSize <= 0 {
		opts.BatchSize = 25
	}
	if opts.FlushInterval <= 0 {
		opts.FlushInterval = 5 * time.Second
	}
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 3
	}
	if opts.RetryDelay <= 0 {
		opts.RetryDelay = time.Second
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		endpoint:      strings.TrimRight(opts.Endpoint, "/"),
		token:         opts.Token,
		batchSize:     opts.BatchSize,
		flushInterval: opts.FlushInterval,
		maxRetries:    opts.MaxRetries,
		retryDelay:    opts.RetryDelay,
		onError:       opts.OnError,
		httpClient:    opts.HTTPClient,
	}
}

// Log sends a log entry with the given severity.
func (c *Client) Log(severity, sourceType string, payload map[string]any) {
	if payload == nil {
		payload = map[string]any{}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.buffer = append(c.buffer, LogEntry{
		SourceType: sourceType,
		Severity:   severity,
		Payload:    payload,
	})

	if len(c.buffer) >= c.batchSize {
		c.flushLocked()
	} else {
		c.scheduleFlush()
	}
}

// Debug logs at debug severity.
func (c *Client) Debug(sourceType string, payload map[string]any) {
	c.Log("debug", sourceType, payload)
}

// Info logs at info severity.
func (c *Client) Info(sourceType string, payload map[string]any) {
	c.Log("info", sourceType, payload)
}

// Warn logs at warning severity.
func (c *Client) Warn(sourceType string, payload map[string]any) {
	c.Log("warning", sourceType, payload)
}

// Error logs at error severity.
func (c *Client) Error(sourceType string, payload map[string]any) {
	c.Log("error", sourceType, payload)
}

// Critical logs at critical severity.
func (c *Client) Critical(sourceType string, payload map[string]any) {
	c.Log("critical", sourceType, payload)
}

// Flush sends all buffered entries immediately.
func (c *Client) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.flushLocked()
}

// Shutdown flushes remaining entries, waits for in-flight sends to complete,
// and prevents further logging.
func (c *Client) Shutdown() {
	c.mu.Lock()
	c.closed = true
	c.cancelTimer()
	c.mu.Unlock()

	c.Flush()
	c.wg.Wait()
}

// Pending returns the number of buffered entries.
func (c *Client) Pending() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.buffer)
}

func (c *Client) flushLocked() {
	c.cancelTimer()
	if len(c.buffer) == 0 {
		return
	}

	entries := make([]LogEntry, len(c.buffer))
	copy(entries, c.buffer)
	c.buffer = c.buffer[:0]

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.send(entries)
	}()
}

func (c *Client) scheduleFlush() {
	if c.timer != nil {
		return
	}
	c.timer = time.AfterFunc(c.flushInterval, func() {
		c.mu.Lock()
		c.timer = nil
		c.flushLocked()
		c.mu.Unlock()
	})
}

func (c *Client) cancelTimer() {
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
}

func (c *Client) send(entries []LogEntry) {
	url := c.endpoint + "/api/webhooks/logs"

	body, err := json.Marshal(entries)
	if err != nil {
		if c.onError != nil {
			c.onError(fmt.Errorf("heimdall: marshal: %w", err), entries)
		}
		return
	}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			if c.onError != nil {
				c.onError(fmt.Errorf("heimdall: build request: %w", err), entries)
			}
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < c.maxRetries {
				time.Sleep(c.retryDelay * time.Duration(math.Pow(2, float64(attempt))))
				continue
			}
			if c.onError != nil {
				c.onError(fmt.Errorf("heimdall: %w", err), entries)
			}
			return
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode < 300 {
			return
		}

		// 4xx — permanent failure.
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			if c.onError != nil {
				c.onError(fmt.Errorf("heimdall: %d", resp.StatusCode), entries)
			}
			return
		}

		// 5xx — retryable.
		if attempt < c.maxRetries {
			time.Sleep(c.retryDelay * time.Duration(math.Pow(2, float64(attempt))))
			continue
		}

		if c.onError != nil {
			c.onError(fmt.Errorf("heimdall: %d after %d attempts", resp.StatusCode, c.maxRetries+1), entries)
		}
		return
	}
}
