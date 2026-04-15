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
	mongoAPIBase         = "https://cloud.mongodb.com/api/atlas/v2"
	mongoDefaultInterval = 120
	mongoMinInterval     = 60
	mongoPollLimit       = 500
)

// MongoDBConfig represents the fields stored in a MongoDB Atlas connection's config JSONB.
type MongoDBConfig struct {
	PublicKey        string `json:"public_key"`
	PrivateKey       string `json:"private_key"`
	GroupID          string `json:"group_id"` // Atlas project ID
	ClusterName      string `json:"cluster_name"`
	PollIntervalSecs int    `json:"poll_interval_secs"`
}

// MongoDB polls the MongoDB Atlas Admin API for audit and access logs.
type MongoDB struct {
	config       MongoDBConfig
	connectionID uuid.UUID
	userID       uuid.UUID
	appID        uuid.UUID
	httpClient   *http.Client
	apiBase      string

	mu       sync.Mutex
	cursor   time.Time
	hostname string // cached cluster hostname, resolved on first successful poll
}

func NewMongoDB(configJSON json.RawMessage, connectionID, userID, appID uuid.UUID) (*MongoDB, error) {
	var cfg MongoDBConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, fmt.Errorf("mongodb: invalid config: %w", err)
	}
	if cfg.PublicKey == "" {
		return nil, fmt.Errorf("mongodb: public_key is required")
	}
	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("mongodb: private_key is required")
	}
	if cfg.GroupID == "" {
		return nil, fmt.Errorf("mongodb: group_id (project ID) is required")
	}
	if cfg.ClusterName == "" {
		return nil, fmt.Errorf("mongodb: cluster_name is required")
	}
	if cfg.PollIntervalSecs < mongoMinInterval {
		cfg.PollIntervalSecs = mongoDefaultInterval
	}

	return &MongoDB{
		config:       cfg,
		connectionID: connectionID,
		userID:       userID,
		appID:        appID,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		apiBase:      mongoAPIBase,
		cursor:       time.Now().Add(-10 * time.Minute),
	}, nil
}

func (m *MongoDB) ParsedConfig() MongoDBConfig { return m.config }

func (m *MongoDB) Connect(ctx context.Context) error {
	path := fmt.Sprintf("/groups/%s", url.PathEscape(m.config.GroupID))
	_, err := m.apiRequest(ctx, path)
	return err
}

func (m *MongoDB) Health(ctx context.Context) error { return m.Connect(ctx) }

func (m *MongoDB) Close() error {
	m.httpClient.CloseIdleConnections()
	return nil
}

func (m *MongoDB) Poll(ctx context.Context, queries *db.Queries) error {
	m.mu.Lock()
	cursor := m.cursor
	m.mu.Unlock()

	// Fetch access logs for the cluster's primary hostname.
	// Atlas returns logs for a specific hostname, which we discover from the
	// cluster API. The hostname is cached after the first successful lookup
	// since it doesn't change at runtime.
	hostname, err := m.cachedHostname(ctx)
	if err != nil {
		return fmt.Errorf("mongodb: get hostname: %w", err)
	}

	startDate := cursor.Format(time.RFC3339)
	endDate := time.Now().Format(time.RFC3339)
	path := fmt.Sprintf("/groups/%s/dbAccessHistory/clusters/%s?start=%s&end=%s&nLogs=%d",
		url.PathEscape(m.config.GroupID),
		url.PathEscape(hostname),
		url.QueryEscape(startDate),
		url.QueryEscape(endDate),
		mongoPollLimit,
	)

	body, err := m.apiRequest(ctx, path)
	if err != nil {
		return fmt.Errorf("mongodb: poll access logs: %w", err)
	}

	var resp struct {
		AccessLogs []struct {
			AuthResult  bool   `json:"authResult"`
			AuthSource  string `json:"authSource"`
			FailureReason string `json:"failureReason"`
			GroupID     string `json:"groupId"`
			Hostname    string `json:"hostname"`
			IPAddress   string `json:"ipAddress"`
			LogLine     string `json:"logLine"`
			Timestamp   string `json:"timestamp"`
			Username    string `json:"username"`
		} `json:"accessLogs"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("mongodb: parse response: %w", err)
	}

	if len(resp.AccessLogs) == 0 {
		return nil
	}

	var maxTS time.Time
	var insertCount int

	for _, entry := range resp.AccessLogs {
		ts, parseErr := time.Parse(time.RFC3339Nano, entry.Timestamp)
		if parseErr != nil {
			slog.Warn("mongodb: unparseable timestamp, using now", "raw", entry.Timestamp, "err", parseErr, "connection_id", m.connectionID)
			ts = time.Now()
		}
		if !ts.After(cursor) {
			continue
		}

		severity := pgtype.Text{String: "info", Valid: true}
		if !entry.AuthResult {
			severity = pgtype.Text{String: "warning", Valid: true}
		}
		if entry.FailureReason != "" {
			severity = pgtype.Text{String: "error", Valid: true}
		}

		payload, err := json.Marshal(map[string]any{
			"timestamp":      entry.Timestamp,
			"auth_result":    entry.AuthResult,
			"auth_source":    entry.AuthSource,
			"failure_reason": entry.FailureReason,
			"hostname":       entry.Hostname,
			"ip_address":     entry.IPAddress,
			"log_line":       entry.LogLine,
			"username":       entry.Username,
		})
		if err != nil {
			slog.Error("mongodb: marshal payload", "err", err, "connection_id", m.connectionID)
			continue
		}

		_, err = queries.InsertLogEntry(ctx, db.InsertLogEntryParams{
			ConnectionID: m.connectionID,
			SourceType:   "mongodb/" + m.config.ClusterName,
			Severity:     severity,
			Payload:      payload,
			UserID:       m.userID,
			AppID:        m.appID,
		})
		if err != nil {
			slog.Error("mongodb: insert log entry", "err", err, "connection_id", m.connectionID)
			continue
		}

		insertCount++
		if ts.After(maxTS) {
			maxTS = ts
		}
	}

	// Advance cursor past the last seen timestamp to avoid re-processing
	// entries with an identical timestamp on the next poll.
	if !maxTS.IsZero() {
		m.mu.Lock()
		m.cursor = maxTS.Add(time.Nanosecond)
		m.mu.Unlock()
	}

	if insertCount > 0 {
		slog.Info("mongodb poll complete", "cluster", m.config.ClusterName, "rows", insertCount, "connection_id", m.connectionID)
	}
	return nil
}

// cachedHostname returns the cluster hostname, resolving and caching it on
// first successful call. Unlike sync.Once, a transient failure does not
// permanently cache the error — the next Poll will retry the lookup.
func (m *MongoDB) cachedHostname(ctx context.Context) (string, error) {
	m.mu.Lock()
	if m.hostname != "" {
		h := m.hostname
		m.mu.Unlock()
		return h, nil
	}
	m.mu.Unlock()

	h, err := m.getClusterHostname(ctx)
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	m.hostname = h
	m.mu.Unlock()
	return h, nil
}

func (m *MongoDB) getClusterHostname(ctx context.Context) (string, error) {
	path := fmt.Sprintf("/groups/%s/clusters/%s",
		url.PathEscape(m.config.GroupID),
		url.PathEscape(m.config.ClusterName),
	)

	body, err := m.apiRequest(ctx, path)
	if err != nil {
		return "", err
	}

	var cluster struct {
		ConnectionStrings struct {
			Standard string `json:"standard"`
		} `json:"connectionStrings"`
		MongoURI string `json:"mongoURI"`
		Name     string `json:"name"`
	}
	if err := json.Unmarshal(body, &cluster); err != nil {
		return "", fmt.Errorf("mongodb: parse cluster: %w", err)
	}

	// Extract hostname from connection string or fall back to cluster name.
	if cluster.ConnectionStrings.Standard != "" {
		u, err := url.Parse(cluster.ConnectionStrings.Standard)
		if err == nil && u.Host != "" {
			return u.Host, nil
		}
	}

	return m.config.ClusterName, nil
}

func (m *MongoDB) apiRequest(ctx context.Context, path string) ([]byte, error) {
	endpoint := m.apiBase + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("mongodb: build request: %w", err)
	}
	// Atlas programmatic API keys use Basic auth (public key as username,
	// private key as password). The v2 API accepts this when the Accept
	// header opts into the versioned contract. Digest auth is a legacy v1.0
	// requirement that does not apply here.
	req.SetBasicAuth(m.config.PublicKey, m.config.PrivateKey)
	req.Header.Set("Accept", "application/vnd.atlas.2023-01-01+json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mongodb: api request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("mongodb: read body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("mongodb: invalid API credentials")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("mongodb: rate limited (429), will retry next interval")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mongodb: api error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
