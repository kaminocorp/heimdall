package logs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"encoding/json"

	"github.com/google/uuid"
)

func TestNewSyslog_Defaults(t *testing.T) {
	cfg := `{"port": 0}`
	sl, err := NewSyslog(json.RawMessage(cfg), uuid.New(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Equal(t, syslogDefaultPort, sl.config.Port)
	assert.Equal(t, "tls", sl.config.Protocol)
}

func TestNewSyslog_InvalidPort(t *testing.T) {
	cfg := `{"port": 99999}`
	_, err := NewSyslog(json.RawMessage(cfg), uuid.New(), uuid.New(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port must be between 1 and 65535")
}

func TestNewSyslog_InvalidProtocol(t *testing.T) {
	cfg := `{"port": 6514, "protocol": "udp"}`
	_, err := NewSyslog(json.RawMessage(cfg), uuid.New(), uuid.New(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "protocol must be 'tcp' or 'tls'")
}

func TestNewSyslog_ValidTCP(t *testing.T) {
	cfg := `{"port": 1514, "protocol": "tcp"}`
	sl, err := NewSyslog(json.RawMessage(cfg), uuid.New(), uuid.New(), nil)
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
