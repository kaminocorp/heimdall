package agent

import "github.com/kaminocorp/lumber/pkg/lumber"

// ShouldEscalate returns true if a classified log event warrants LLM attention.
// Rules are hardcoded per the Phase 8 severity gate table.
//
// Lumber's current taxonomy has six root types: ERROR, REQUEST, DEPLOY, SYSTEM,
// ACCESS, PERFORMANCE. Any log that does not match a known type is returned as
// UNCLASSIFIED by the model and escalated unconditionally via the default branch.
func ShouldEscalate(event lumber.Event) bool {
	switch event.Type {
	case "ERROR":
		return true
	case "PERFORMANCE":
		return true

	case "REQUEST":
		switch event.Category {
		case "server_error", "slow_request":
			return true
		default:
			return false
		}

	case "DEPLOY":
		return true

	case "SYSTEM":
		switch event.Category {
		case "resource_alert", "config_change":
			return true
		default:
			return false
		}

	case "ACCESS":
		switch event.Category {
		case "login_failure", "auth_failure", "permission_change", "api_key_event":
			return true
		default:
			return false
		}

	case "UNCLASSIFIED":
		return true

	default:
		return true
	}
}
