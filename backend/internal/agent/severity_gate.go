package agent

import "github.com/kaminocorp/lumber/pkg/lumber"

// ShouldEscalate returns true if a classified log event warrants LLM attention.
// Rules are hardcoded per the Phase 8 severity gate table.
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

	case "DATA":
		switch event.Category {
		case "migration":
			return true
		default:
			return false
		}

	case "SCHEDULED":
		switch event.Category {
		case "cron_failed":
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
