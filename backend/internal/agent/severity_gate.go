package agent

import "github.com/kaminocorp/lumber/pkg/lumber"

// Rule IDs exposed via the Gate stage of log_pipeline_events.rule_hit. Stable
// identifiers per escalation branch so Pipeline-page analytics can "this
// rule fired N times in the last hour" without hard-coding Lumber taxonomy
// strings on the frontend. Chosen per `pipeline-rule-options.md` Option A:
// hardcoded rule-to-ID mapping, no DB table, no expression engine.
//
// `RuleNone` marks non-escalated events so the Gate-stage row always carries
// a reason — empty-string in the column would blur "no rule matched" with
// "we haven't written this column yet."
const (
	RuleErrorType              = "error_type"
	RulePerformanceType        = "performance_type"
	RuleDeployType             = "deploy_type"
	RuleRequestServerError     = "request_server_error"
	RuleRequestSlowRequest     = "request_slow_request"
	RuleSystemResourceAlert    = "system_resource_alert"
	RuleSystemConfigChange     = "system_config_change"
	RuleAccessLoginFailure     = "access_login_failure"
	RuleAccessAuthFailure      = "access_auth_failure"
	RuleAccessPermissionChange = "access_permission_change"
	RuleAccessAPIKeyEvent      = "access_api_key_event"
	RuleUnclassified           = "unclassified"
	RuleUnknownType            = "unknown_type"
	RuleNone                   = ""
)

// ShouldEscalate decides whether a classified log event warrants LLM
// attention, and returns the stable rule ID that made the decision. The
// rule ID is only populated for escalating branches; safe branches return
// `RuleNone` so callers don't need to branch on the bool to know whether
// to record the reason.
//
// Lumber's current taxonomy has six root types: ERROR, REQUEST, DEPLOY,
// SYSTEM, ACCESS, PERFORMANCE. Anything outside that set is escalated via
// the catch-all `unknown_type` rule, and UNCLASSIFIED hits its own ID.
func ShouldEscalate(event lumber.Event) (bool, string) {
	switch event.Type {
	case "ERROR":
		return true, RuleErrorType
	case "PERFORMANCE":
		return true, RulePerformanceType

	case "REQUEST":
		switch event.Category {
		case "server_error":
			return true, RuleRequestServerError
		case "slow_request":
			return true, RuleRequestSlowRequest
		default:
			return false, RuleNone
		}

	case "DEPLOY":
		return true, RuleDeployType

	case "SYSTEM":
		switch event.Category {
		case "resource_alert":
			return true, RuleSystemResourceAlert
		case "config_change":
			return true, RuleSystemConfigChange
		default:
			return false, RuleNone
		}

	case "ACCESS":
		switch event.Category {
		case "login_failure":
			return true, RuleAccessLoginFailure
		case "auth_failure":
			return true, RuleAccessAuthFailure
		case "permission_change":
			return true, RuleAccessPermissionChange
		case "api_key_event":
			return true, RuleAccessAPIKeyEvent
		default:
			return false, RuleNone
		}

	case "UNCLASSIFIED":
		return true, RuleUnclassified

	default:
		return true, RuleUnknownType
	}
}
