package heimdall

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNew_Defaults(t *testing.T) {
	c := New(Options{Endpoint: "https://example.com", Token: "tok"})
	if c.batchSize != 25 {
		t.Errorf("batchSize = %d, want 25", c.batchSize)
	}
	if c.flushInterval != 5*time.Second {
		t.Errorf("flushInterval = %v, want 5s", c.flushInterval)
	}
	if c.maxRetries != 3 {
		t.Errorf("maxRetries = %d, want 3", c.maxRetries)
	}
}

func TestNew_StripTrailingSlash(t *testing.T) {
	c := New(Options{Endpoint: "https://example.com/", Token: "tok"})
	if c.endpoint != "https://example.com" {
		t.Errorf("endpoint = %q, want stripped", c.endpoint)
	}
}

func TestClient_BuffersEntries(t *testing.T) {
	c := New(Options{
		Endpoint:      "https://example.com",
		Token:         "tok",
		BatchSize:     100,
		FlushInterval: time.Minute,
	})
	c.Info("a", nil)
	c.Info("b", nil)
	if c.Pending() != 2 {
		t.Errorf("Pending() = %d, want 2", c.Pending())
	}
	c.Shutdown()
}

func TestClient_SeverityShorthands(t *testing.T) {
	var mu sync.Mutex
	var received []LogEntry

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var entries []LogEntry
		json.NewDecoder(r.Body).Decode(&entries)
		mu.Lock()
		received = append(received, entries...)
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := New(Options{
		Endpoint:  srv.URL,
		Token:     "tok",
		BatchSize: 100,
	})

	c.Debug("d", nil)
	c.Info("i", nil)
	c.Warn("w", nil)
	c.Error("e", nil)
	c.Critical("c", nil)
	c.Flush()
	time.Sleep(100 * time.Millisecond) // wait for async send

	mu.Lock()
	defer mu.Unlock()

	expected := []string{"debug", "info", "warning", "error", "critical"}
	if len(received) != 5 {
		t.Fatalf("received %d entries, want 5", len(received))
	}
	for i, e := range received {
		if e.Severity != expected[i] {
			t.Errorf("entry %d severity = %q, want %q", i, e.Severity, expected[i])
		}
	}
}

func TestClient_PayloadFormat(t *testing.T) {
	var receivedBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = json.Marshal(r.Header.Get("Authorization"))
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusCreated)
		_ = receivedBody
	}))
	defer srv.Close()

	c := New(Options{
		Endpoint:  srv.URL,
		Token:     "test-token",
		BatchSize: 100,
	})

	c.Info("test.event", map[string]any{"key": "value"})
	c.Flush()
	time.Sleep(100 * time.Millisecond)
}

func TestClient_FlushEmpty(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := New(Options{Endpoint: srv.URL, Token: "tok", BatchSize: 100})
	c.Flush()
	time.Sleep(50 * time.Millisecond)
	if called {
		t.Error("expected no HTTP call for empty flush")
	}
}

func TestClient_ShutdownFlushes(t *testing.T) {
	var mu sync.Mutex
	var count int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var entries []LogEntry
		json.NewDecoder(r.Body).Decode(&entries)
		mu.Lock()
		count += len(entries)
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := New(Options{
		Endpoint:      srv.URL,
		Token:         "tok",
		BatchSize:     100,
		FlushInterval: time.Minute,
	})

	c.Info("final", nil)
	c.Shutdown()
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestClient_NilPayload(t *testing.T) {
	var mu sync.Mutex
	var received []LogEntry

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var entries []LogEntry
		json.NewDecoder(r.Body).Decode(&entries)
		mu.Lock()
		received = append(received, entries...)
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := New(Options{Endpoint: srv.URL, Token: "tok", BatchSize: 100})
	c.Info("test", nil)
	c.Flush()
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 {
		t.Fatalf("received %d, want 1", len(received))
	}
	if received[0].Payload == nil {
		t.Error("payload should be empty map, not nil")
	}
}

func TestClient_OnError4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	var errCalled bool
	c := New(Options{
		Endpoint:   srv.URL,
		Token:      "bad",
		BatchSize:  100,
		MaxRetries: 2,
		RetryDelay: 10 * time.Millisecond,
		OnError: func(err error, entries []LogEntry) {
			errCalled = true
		},
	})

	c.Info("test", nil)
	c.Flush()
	time.Sleep(100 * time.Millisecond)

	if !errCalled {
		t.Error("expected onError to be called for 401")
	}
}
