package agent

import (
	"testing"

	"github.com/kaminocorp/lumber/pkg/lumber"
	"github.com/stretchr/testify/assert"
)

func event(typ, category string) lumber.Event {
	return lumber.Event{Type: typ, Category: category, Confidence: 0.9}
}

func TestShouldEscalate(t *testing.T) {
	escalated := []struct {
		name     string
		typ      string
		category string
	}{
		// ERROR — all escalate
		{"error/connection_failure", "ERROR", "connection_failure"},
		{"error/runtime_exception", "ERROR", "runtime_exception"},
		{"error/timeout", "ERROR", "timeout"},

		// PERFORMANCE — all escalate
		{"performance/latency_spike", "PERFORMANCE", "latency_spike"},
		{"performance/db_slow_query", "PERFORMANCE", "db_slow_query"},

		// REQUEST — selective
		{"request/server_error", "REQUEST", "server_error"},
		{"request/slow_request", "REQUEST", "slow_request"},

		// DEPLOY — all escalate
		{"deploy/build_started", "DEPLOY", "build_started"},
		{"deploy/deploy_failed", "DEPLOY", "deploy_failed"},
		{"deploy/rollback", "DEPLOY", "rollback"},

		// SYSTEM — selective
		{"system/resource_alert", "SYSTEM", "resource_alert"},
		{"system/config_change", "SYSTEM", "config_change"},

		// ACCESS — selective
		{"access/login_failure", "ACCESS", "login_failure"},
		{"access/auth_failure", "ACCESS", "auth_failure"},
		{"access/permission_change", "ACCESS", "permission_change"},
		{"access/api_key_event", "ACCESS", "api_key_event"},

		// UNCLASSIFIED — always
		{"unclassified", "UNCLASSIFIED", ""},

		// Unknown types (incl. DATA and SCHEDULED, which are not in Lumber's taxonomy)
		// fall through to the default branch and escalate unconditionally.
		{"unknown_type", "BANANA", "whatever"},
		{"data/migration", "DATA", "migration"},
		{"data/query_executed", "DATA", "query_executed"},
		{"data/replication", "DATA", "replication"},
		{"scheduled/cron_failed", "SCHEDULED", "cron_failed"},
		{"scheduled/cron_started", "SCHEDULED", "cron_started"},
		{"scheduled/cron_completed", "SCHEDULED", "cron_completed"},
	}

	for _, tt := range escalated {
		t.Run("escalate/"+tt.name, func(t *testing.T) {
			ok, rule := ShouldEscalate(event(tt.typ, tt.category))
			assert.True(t, ok)
			assert.NotEmpty(t, rule, "escalating branches must return a rule id")
		})
	}

	safe := []struct {
		name     string
		typ      string
		category string
	}{
		// REQUEST — safe
		{"request/success", "REQUEST", "success"},
		{"request/redirect", "REQUEST", "redirect"},
		{"request/client_error", "REQUEST", "client_error"},

		// SYSTEM — safe
		{"system/health_check", "SYSTEM", "health_check"},
		{"system/process_lifecycle", "SYSTEM", "process_lifecycle"},
		{"system/scaling_event", "SYSTEM", "scaling_event"},

		// ACCESS — safe
		{"access/login_success", "ACCESS", "login_success"},
		{"access/session_expired", "ACCESS", "session_expired"},

	}

	for _, tt := range safe {
		t.Run("safe/"+tt.name, func(t *testing.T) {
			ok, rule := ShouldEscalate(event(tt.typ, tt.category))
			assert.False(t, ok)
			assert.Equal(t, RuleNone, rule, "safe branches must return RuleNone")
		})
	}
}
