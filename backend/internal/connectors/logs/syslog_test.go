package logs

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"encoding/json"

	"github.com/google/uuid"
)

func TestNewSyslog_Defaults(t *testing.T) {
	cfg := `{"port": 0}`
	sl, err := NewSyslog(json.RawMessage(cfg), uuid.New(), uuid.New(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Equal(t, syslogDefaultPort, sl.config.Port)
	assert.Equal(t, "tls", sl.config.Protocol)
}

func TestNewSyslog_InvalidPort(t *testing.T) {
	cfg := `{"port": 99999}`
	_, err := NewSyslog(json.RawMessage(cfg), uuid.New(), uuid.New(), uuid.New(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port must be between 1 and 65535")
}

func TestNewSyslog_InvalidProtocol(t *testing.T) {
	cfg := `{"port": 6514, "protocol": "udp"}`
	_, err := NewSyslog(json.RawMessage(cfg), uuid.New(), uuid.New(), uuid.New(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "protocol must be 'tcp' or 'tls'")
}

func TestNewSyslog_ValidTCP(t *testing.T) {
	cfg := `{"port": 1514, "protocol": "tcp"}`
	sl, err := NewSyslog(json.RawMessage(cfg), uuid.New(), uuid.New(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1514, sl.config.Port)
	assert.Equal(t, "tcp", sl.config.Protocol)
}

func TestParseSyslogMessage_RFC5424(t *testing.T) {
	line := `<165>1 2026-04-05T12:00:00Z myhost myapp 1234 ID47 This is a test`
	msg := parseSyslogMessage(line)
	assert.Equal(t, 20, msg.Facility) // 165 / 8 = 20
	assert.Equal(t, 5, msg.Severity)  // 165 % 8 = 5 (Notice)
	assert.Equal(t, "2026-04-05T12:00:00Z", msg.Timestamp)
	assert.Equal(t, "myhost", msg.Hostname)
	assert.Equal(t, "myapp", msg.AppName)
	assert.Equal(t, "1234", msg.ProcID)
	assert.Equal(t, "ID47", msg.MsgID)
	assert.Equal(t, "This is a test", msg.Message)
	assert.Equal(t, line, msg.Raw)
}

func TestParseSyslogMessage_RFC5424_NilValues(t *testing.T) {
	line := `<134>1 2026-04-05T12:00:00Z - - - - Some message`
	msg := parseSyslogMessage(line)
	assert.Equal(t, "", msg.Hostname)
	assert.Equal(t, "", msg.AppName)
	assert.Equal(t, "", msg.ProcID)
	assert.Equal(t, "", msg.MsgID)
	assert.Equal(t, "Some message", msg.Message)
}

func TestParseSyslogMessage_RFC3164(t *testing.T) {
	line := `<13>Apr  5 12:00:00 myhost myapp[1234]: Something happened`
	msg := parseSyslogMessage(line)
	assert.Equal(t, 1, msg.Facility)  // 13 / 8 = 1
	assert.Equal(t, 5, msg.Severity)  // 13 % 8 = 5 (Notice)
	assert.Equal(t, "Apr  5 12:00:00", msg.Timestamp)
	assert.Equal(t, "myhost", msg.Hostname)
	assert.Contains(t, msg.Message, "myapp[1234]: Something happened")
}

func TestParseSyslogMessage_Fallback(t *testing.T) {
	line := `Just a plain log line with no syslog framing`
	msg := parseSyslogMessage(line)
	assert.Equal(t, 6, msg.Severity) // default Informational
	assert.Equal(t, line, msg.Message)
	assert.Equal(t, line, msg.Raw)
}

func TestMapSyslogSeverity(t *testing.T) {
	tests := []struct {
		sev      int
		expected string
	}{
		{0, "critical"},
		{1, "critical"},
		{2, "critical"},
		{3, "error"},
		{4, "warning"},
		{5, "info"},
		{6, "info"},
		{7, "debug"},
	}

	for _, tt := range tests {
		result := mapSyslogSeverity(tt.sev)
		assert.True(t, result.Valid)
		assert.Equal(t, tt.expected, result.String, "severity %d", tt.sev)
	}
}

// TestSyslogHealth_StalenessTransitions verifies that Health() reports the
// three distinct states the connector can be in after start-up:
//   1. listener present but allow-list never loaded → degraded (initialRefreshDone=false)
//   2. allow-list loaded recently → healthy
//   3. allow-list loaded but not refreshed within syslogStaleThreshold → degraded (stale)
//
// Together these exercise the Tier 2.1 hardening: a DB outage that begins
// AFTER successful startup must eventually surface via Health(), not stay
// masked behind an old initialRefreshDone=true.
func TestSyslogHealth_StalenessTransitions(t *testing.T) {
	// A dummy listener on an ephemeral port — we never accept from it, just
	// need Health()'s "listener != nil" branch to pass.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer lis.Close()

	s := &Syslog{
		connectionID: uuid.New(),
		userID:       uuid.New(),
		appID:        uuid.New(),
		connSem:      make(chan struct{}, 1),
		enabled:      map[string]struct{}{},
		discovered:   map[string]struct{}{},
		listener:     lis,
	}

	// Phase 1: never-refreshed — Health must fail with the "not loaded" message.
	err = s.Health(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "has not loaded yet")

	// Phase 2: simulate a successful refresh just now.
	s.filterMu.Lock()
	s.lastRefreshAt = time.Now()
	s.filterMu.Unlock()
	s.initialRefreshDone.Store(true)
	require.NoError(t, s.Health(context.Background()), "fresh refresh should report healthy")

	// Phase 3: age the last refresh past the stale threshold — Health must
	// degrade even though initialRefreshDone is still true. This is the
	// regression guard for the "silent staleness" bug that motivated Tier 2.1.
	s.filterMu.Lock()
	s.lastRefreshAt = time.Now().Add(-(syslogStaleThreshold + time.Second))
	s.filterMu.Unlock()
	err = s.Health(context.Background())
	require.Error(t, err, "stale cache should report degraded")
	assert.Contains(t, err.Error(), "stale")

	// Phase 4: a new successful refresh brings it back — confirms the check
	// is time-based, not a one-way latch.
	s.filterMu.Lock()
	s.lastRefreshAt = time.Now()
	s.filterMu.Unlock()
	require.NoError(t, s.Health(context.Background()))
}

// TestSyslogHealth_NoListener covers the simplest failure mode: Close()
// sets listener to nil, Health() reports not-running regardless of the
// allow-list state.
func TestSyslogHealth_NoListener(t *testing.T) {
	s := &Syslog{
		connectionID: uuid.New(),
		enabled:      map[string]struct{}{},
		discovered:   map[string]struct{}{},
	}
	s.initialRefreshDone.Store(true)
	s.filterMu.Lock()
	s.lastRefreshAt = time.Now()
	s.filterMu.Unlock()

	err := s.Health(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}
