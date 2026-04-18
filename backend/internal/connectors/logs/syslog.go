package logs

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

const (
	syslogDefaultPort     = 6514
	syslogMaxMessageSize  = 64 * 1024 // 64 KB per message
	syslogShutdownTimeout = 5 * time.Second
	syslogMaxConnections  = 500              // max concurrent TCP connections per listener
	syslogReadTimeout     = 5 * time.Minute  // idle read deadline per connection
	syslogRefreshInterval = 60 * time.Second // how often to refresh the enabled-hostname cache
	syslogMaxDiscovered   = 10000            // per-listener cap on the in-memory discovered-hostnames dedupe set
	syslogInitialRetries  = 3                // initial allow-list load retries before surfacing via Health()
	syslogStaleThreshold  = 3 * syslogRefreshInterval
)

// SyslogConfig represents the fields stored in a syslog connection's config JSONB.
type SyslogConfig struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"` // "tcp" (plaintext) or "tls" (default)
	TLSCert  string `json:"tls_cert"` // PEM-encoded certificate (for TLS mode)
	TLSKey   string `json:"tls_key"`  // PEM-encoded private key (for TLS mode)
}

// Syslog implements a TCP/TLS syslog listener that accepts RFC 5424 and RFC 3164
// messages and inserts them into log_buffer.
//
// Phase 4: each incoming message is filtered by hostname via the generic
// source-filter system. The listener caches the set of enabled hostnames
// in memory (refreshed every minute via a background goroutine), so the
// hot path is an O(1) map lookup rather than a DB query per message.
// Drop-by-default applies — hosts must be explicitly enabled in the source
// selector before their messages are persisted.
type Syslog struct {
	config       SyslogConfig
	connectionID uuid.UUID
	userID       uuid.UUID
	appID        uuid.UUID
	queries      *db.Queries

	listener net.Listener
	mu       sync.Mutex
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	connSem  chan struct{} // semaphore to cap concurrent connections

	// filterMu guards enabled, discovered, and lastRefreshAt — all read on
	// every message or health check. A plain Mutex is fine at the per-message
	// rate; RWMutex would save a few nanoseconds but complicate the refresh
	// path.
	filterMu      sync.Mutex
	enabled       map[string]struct{} // hostnames with enabled=true in app_source_filters
	discovered    map[string]struct{} // hostnames already upserted in this listener's lifetime (per-process dedupe for connection_sources writes)
	lastRefreshAt time.Time           // timestamp of the last successful refreshEnabledSet; zero value when never refreshed

	// initialRefreshDone flips to true once the allow-list has been
	// successfully loaded at least once. Until then, Health() reports the
	// listener as degraded — the accept loop is running but every message
	// is drop-by-default because the cache is empty, which is operationally
	// indistinguishable from a broken connection.
	initialRefreshDone atomic.Bool
}

// NewSyslog creates a Syslog connector from a connection's config JSONB.
// The queries handle is used for inserting log entries as they arrive.
func NewSyslog(configJSON json.RawMessage, connectionID, userID, appID uuid.UUID, queries *db.Queries) (*Syslog, error) {
	var cfg SyslogConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, fmt.Errorf("syslog: invalid config: %w", err)
	}

	if cfg.Port == 0 {
		cfg.Port = syslogDefaultPort
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("syslog: port must be between 1 and 65535")
	}
	if cfg.Protocol == "" {
		cfg.Protocol = "tls"
	}
	if cfg.Protocol != "tcp" && cfg.Protocol != "tls" {
		return nil, fmt.Errorf("syslog: protocol must be 'tcp' or 'tls'")
	}

	return &Syslog{
		config:       cfg,
		connectionID: connectionID,
		userID:       userID,
		appID:        appID,
		queries:      queries,
		connSem:      make(chan struct{}, syslogMaxConnections),
		enabled:      make(map[string]struct{}),
		discovered:   make(map[string]struct{}),
	}, nil
}

// ParsedConfig returns the parsed SyslogConfig.
func (s *Syslog) ParsedConfig() SyslogConfig {
	return s.config
}

// Connect validates the config and starts the TCP/TLS listener.
func (s *Syslog) Connect(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Port)

	var listener net.Listener
	var err error

	if s.config.Protocol == "tls" {
		if s.config.TLSCert == "" || s.config.TLSKey == "" {
			return fmt.Errorf("syslog: tls_cert and tls_key are required for TLS mode")
		}
		cert, err := tls.X509KeyPair([]byte(s.config.TLSCert), []byte(s.config.TLSKey))
		if err != nil {
			return fmt.Errorf("syslog: invalid TLS certificate: %w", err)
		}
		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
		listener, err = tls.Listen("tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("syslog: tls listen on %s: %w", addr, err)
		}
	} else {
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("syslog: tcp listen on %s: %w", addr, err)
		}
	}

	s.mu.Lock()
	s.listener = listener
	s.mu.Unlock()

	slog.Info("syslog listener started",
		"connection_id", s.connectionID,
		"protocol", s.config.Protocol,
		"port", s.config.Port,
	)
	return nil
}

// Health checks whether the listener is still open, has loaded its hostname
// allow-list at least once, and has refreshed it within the staleness
// threshold. Without the staleness check, a DB outage that begins *after*
// startup would leave the cache frozen indefinitely while Health() still
// reported OK — operators would discover the problem from silent data gaps
// rather than a paging signal. Staleness threshold is 3× the refresh
// interval, i.e. two missed ticks.
func (s *Syslog) Health(ctx context.Context) error {
	s.mu.Lock()
	l := s.listener
	s.mu.Unlock()
	if l == nil {
		return fmt.Errorf("syslog: listener not running")
	}
	if !s.initialRefreshDone.Load() {
		return fmt.Errorf("syslog: hostname allow-list has not loaded yet — incoming messages are being dropped")
	}
	s.filterMu.Lock()
	age := time.Since(s.lastRefreshAt)
	s.filterMu.Unlock()
	if age > syslogStaleThreshold {
		return fmt.Errorf("syslog: hostname allow-list is stale (last refresh %s ago); filters may not reflect recent changes", age.Round(time.Second))
	}
	return nil
}

// Listen starts accepting connections and processing syslog messages.
// It blocks until the context is cancelled or an unrecoverable error occurs.
// Each TCP connection is handled in its own goroutine.
func (s *Syslog) Listen(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.cancel = cancel
	s.mu.Unlock()

	defer cancel()

	// Kick off the enabled-hostname refresh goroutine. A one-shot load
	// happens synchronously here so the very first accepted connection
	// sees the correct filter state; then the background loop keeps it
	// fresh without blocking accepts.
	//
	// If the initial load fails transiently (DB hiccup during startup),
	// retry a few times with backoff. If it still fails, we proceed anyway —
	// Health() will report the degraded state, and the background ticker
	// will keep trying. Better than crashing the listener during a database
	// blip, but visible enough that oncall will notice.
	for attempt := 1; attempt <= syslogInitialRetries; attempt++ {
		if err := s.refreshEnabledSet(ctx); err == nil {
			break
		} else if attempt == syslogInitialRetries {
			slog.Warn("syslog: initial allow-list load failed after retries; listener will drop messages until next refresh succeeds",
				"err", err, "connection_id", s.connectionID, "attempts", attempt)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(attempt) * time.Second):
		}
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(syslogRefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = s.refreshEnabledSet(ctx)
			}
		}
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Check if we were shut down intentionally.
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			// Temporary error — log and continue.
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				slog.Warn("syslog: accept timeout", "connection_id", s.connectionID)
				continue
			}
			return fmt.Errorf("syslog: accept: %w", err)
		}

		// Enforce max concurrent connections via semaphore.
		select {
		case s.connSem <- struct{}{}:
		default:
			slog.Warn("syslog: max connections reached, rejecting", "remote", conn.RemoteAddr(), "connection_id", s.connectionID)
			conn.Close()
			continue
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer func() { <-s.connSem }()
			s.handleConnection(ctx, conn)
		}()
	}
}

// Stream satisfies the StreamConnector interface. The syslog connector is
// listener-based (TCP/TLS), so it delegates to Listen() which writes directly
// to the database. The out channel parameter is intentionally unused — syslog
// messages are inserted as they arrive rather than being piped through a channel.
func (s *Syslog) Stream(ctx context.Context, _ chan<- []byte) error {
	return s.Listen(ctx)
}

// Close shuts down the listener and waits for in-flight connections to drain.
func (s *Syslog) Close() error {
	s.mu.Lock()
	cancel := s.cancel
	listener := s.listener
	s.listener = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if listener != nil {
		listener.Close()
	}

	// Wait for in-flight connections to finish, with a timeout.
	// Use a channel that is closed inline so the goroutine is joined
	// in all code paths (avoiding a goroutine leak on timeout).
	waitDone := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(waitDone)
	}()

	timer := time.NewTimer(syslogShutdownTimeout)
	defer timer.Stop()

	select {
	case <-waitDone:
	case <-timer.C:
		slog.Warn("syslog: shutdown timeout, some connections may not have drained",
			"connection_id", s.connectionID)
		// The wg.Wait goroutine will finish naturally once connections
		// hit their read deadline (syslogReadTimeout). We accept the
		// brief leak rather than blocking the caller indefinitely.
	}

	slog.Info("syslog listener stopped", "connection_id", s.connectionID)
	return nil
}

// handleConnection reads syslog messages from a single TCP connection.
func (s *Syslog) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	remote := conn.RemoteAddr().String()
	slog.Debug("syslog: new connection", "remote", remote, "connection_id", s.connectionID)

	// Set initial read deadline to evict idle connections.
	conn.SetReadDeadline(time.Now().Add(syslogReadTimeout))

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, syslogMaxMessageSize), syslogMaxMessageSize)

	for scanner.Scan() {
		// Reset deadline after each successful read.
		conn.SetReadDeadline(time.Now().Add(syslogReadTimeout))
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		msg := parseSyslogMessage(line)

		// Hostname is the source_name for syslog, matching the generic
		// filter system's vocabulary. Falls back to "unknown" so a message
		// without a hostname still gets recorded under *some* source —
		// users who want to ingest those opt in by enabling "unknown".
		hostname := msg.Hostname
		if hostname == "" {
			hostname = "unknown"
		}

		// Discovery: record the hostname so the source selector can surface
		// it. Deduped per-listener-lifetime — we don't hammer the DB for
		// every packet from the same host.
		s.discoverHostname(ctx, hostname)

		// Filter: drop messages whose hostname isn't in the enabled set.
		// Matches webhook ingestion's drop-by-default semantic.
		if !s.isEnabled(hostname) {
			continue
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			slog.Error("syslog: marshal payload", "err", err, "connection_id", s.connectionID)
			continue
		}

		severity := mapSyslogSeverity(msg.Severity)

		_, err = s.queries.InsertLogEntry(ctx, db.InsertLogEntryParams{
			ConnectionID: s.connectionID,
			SourceType:   "syslog",
			Severity:     severity,
			Payload:      payload,
			UserID:       s.userID,
			AppID:        s.appID,
		})
		if err != nil {
			slog.Error("syslog: insert log entry", "err", err, "connection_id", s.connectionID)
		}
	}

	if err := scanner.Err(); err != nil {
		// Don't log errors caused by shutdown.
		select {
		case <-ctx.Done():
		default:
			slog.Error("syslog: read error", "err", err, "remote", remote, "connection_id", s.connectionID)
		}
	}
}

// syslogMessage represents a parsed syslog message.
type syslogMessage struct {
	Facility  int    `json:"facility,omitempty"`
	Severity  int    `json:"severity"` // 0=Emergency .. 7=Debug
	Timestamp string `json:"timestamp,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
	AppName   string `json:"app_name,omitempty"`
	ProcID    string `json:"proc_id,omitempty"`
	MsgID     string `json:"msg_id,omitempty"`
	Message   string `json:"message"`
	Raw       string `json:"raw"`
}

// RFC 5424 pattern: <PRI>VERSION SP TIMESTAMP SP HOSTNAME SP APP-NAME SP PROCID SP MSGID SP STRUCTURED-DATA SP MSG
// RFC 3164 pattern: <PRI>TIMESTAMP SP HOSTNAME SP MSG
// The STRUCTURED-DATA field is either "-" (nil) or one or more "[...]" blocks.
// We capture it as group 7 and the message as group 8.
// Group 7 captures the optional STRUCTURED-DATA field (either "-" or "[...]" blocks).
// Group 8 captures the message. If STRUCTURED-DATA is absent (non-compliant but common),
// group 7 will be empty and the message lands in group 8 via the fallthrough.
var rfc5424Re = regexp.MustCompile(`^<(\d{1,3})>\d+ (\S+) (\S+) (\S+) (\S+) (\S+) (?:(-|(?:\[.*?\])+) )?(.*)$`)
var rfc3164Re = regexp.MustCompile(`^<(\d{1,3})>(\w{3}\s+\d{1,2}\s\d{2}:\d{2}:\d{2}) (\S+) (.*)$`)

// parseSyslogMessage attempts RFC 5424 first, then RFC 3164, then treats as raw.
func parseSyslogMessage(line string) syslogMessage {
	// Try RFC 5424.
	if m := rfc5424Re.FindStringSubmatch(line); m != nil {
		pri, _ := strconv.Atoi(m[1])
		return syslogMessage{
			Facility:  pri / 8,
			Severity:  pri % 8,
			Timestamp: m[2],
			Hostname:  nilDash(m[3]),
			AppName:   nilDash(m[4]),
			ProcID:    nilDash(m[5]),
			MsgID:     nilDash(m[6]),
			Message:   strings.TrimSpace(m[8]),
			Raw:       line,
		}
	}

	// Try RFC 3164.
	if m := rfc3164Re.FindStringSubmatch(line); m != nil {
		pri, _ := strconv.Atoi(m[1])
		return syslogMessage{
			Facility:  pri / 8,
			Severity:  pri % 8,
			Timestamp: strings.TrimSpace(m[2]),
			Hostname:  m[3],
			Message:   strings.TrimSpace(m[4]),
			Raw:       line,
		}
	}

	// Fallback: treat as unstructured message.
	return syslogMessage{
		Severity: 6, // Informational
		Message:  line,
		Raw:      line,
	}
}

// nilDash returns "" for the syslog NILVALUE "-".
func nilDash(s string) string {
	if s == "-" {
		return ""
	}
	return s
}

// mapSyslogSeverity converts a numeric syslog severity (0-7) to a Heimdall severity string.
func mapSyslogSeverity(sev int) pgtype.Text {
	var s string
	switch {
	case sev <= 2: // Emergency, Alert, Critical
		s = "critical"
	case sev <= 4: // Error, Warning
		if sev == 3 {
			s = "error"
		} else {
			s = "warning"
		}
	case sev <= 5: // Notice
		s = "info"
	case sev <= 6: // Informational
		s = "info"
	default: // Debug
		s = "debug"
	}
	return pgtype.Text{String: s, Valid: true}
}

// refreshEnabledSet reloads the hostname allow-list from app_source_filters.
// Called on listener start and every syslogRefreshInterval thereafter, so
// toggling a host in the selector takes effect within ~a minute without
// needing a per-message DB query. Failures are logged and leave the existing
// cache intact — preferable to suddenly flipping every host off because of
// a transient DB hiccup.
func (s *Syslog) refreshEnabledSet(ctx context.Context) error {
	rows, err := s.queries.ListEnabledSourceNames(ctx, db.ListEnabledSourceNamesParams{
		ConnectionID: s.connectionID,
		AppID:        s.appID,
	})
	if err != nil {
		slog.Warn("syslog: refresh enabled set failed, keeping previous cache",
			"err", err, "connection_id", s.connectionID)
		return err
	}

	next := make(map[string]struct{}, len(rows))
	for _, h := range rows {
		next[h] = struct{}{}
	}

	s.filterMu.Lock()
	s.enabled = next
	s.lastRefreshAt = time.Now()
	s.filterMu.Unlock()
	s.initialRefreshDone.Store(true)
	return nil
}

// isEnabled returns true if the given hostname is currently in the allow
// list. Per-message hot path; keeps the lock scope tight.
func (s *Syslog) isEnabled(hostname string) bool {
	s.filterMu.Lock()
	_, ok := s.enabled[hostname]
	s.filterMu.Unlock()
	return ok
}

// discoverHostname upserts a connection_sources row for a newly-seen
// hostname. Deduped via the in-memory `discovered` map — once a hostname
// has been successfully recorded, subsequent packets from the same host
// skip the DB hop.
//
// Ordering note (regression fix for the rollback race): the upsert runs
// BEFORE marking `discovered`, so nothing is ever added to the map until
// the DB write succeeds. If an earlier version had done "mark, upsert,
// roll back on error", a concurrent second packet could have observed
// the marker mid-flight, short-circuited, and — if the first upsert then
// rolled back — left the host permanently absent from the selector.
// Under this ordering, concurrent first-arrivals from the same host may
// each execute the upsert (bounded waste: ON CONFLICT DO UPDATE just
// bumps last_seen_at), but the DB state is always correct. In steady
// state, only one message per host per listener lifetime hits the DB.
func (s *Syslog) discoverHostname(ctx context.Context, hostname string) {
	s.filterMu.Lock()
	_, seen := s.discovered[hostname]
	s.filterMu.Unlock()
	if seen {
		return
	}

	if _, err := s.queries.UpsertConnectionSource(ctx, db.UpsertConnectionSourceParams{
		ConnectionID: s.connectionID,
		SourceName:   hostname,
	}); err != nil {
		// Don't mark discovered — leaves the door open for the next packet
		// from this host to retry. Logged at warn so sustained failures
		// surface in ops dashboards.
		slog.Warn("syslog: discovery upsert failed",
			"err", err, "hostname", hostname, "connection_id", s.connectionID)
		return
	}

	// Upsert succeeded — promote to the dedupe map. Bounded cache: at the
	// cap, wipe the whole map and let the new entry repopulate it. A
	// re-upsert after clearing is cheap (ON CONFLICT DO UPDATE bumps
	// last_seen_at), so the trade-off is bounded memory for slightly more
	// DB writes under hostname sprawl — an attacker spraying forged
	// hostnames can't exhaust the listener's RAM.
	s.filterMu.Lock()
	if len(s.discovered) >= syslogMaxDiscovered {
		slog.Warn("syslog: discovered-hostname cache at cap, clearing",
			"cap", syslogMaxDiscovered, "connection_id", s.connectionID)
		s.discovered = make(map[string]struct{})
	}
	s.discovered[hostname] = struct{}{}
	s.filterMu.Unlock()
}
