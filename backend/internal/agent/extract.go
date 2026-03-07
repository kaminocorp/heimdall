package agent

import (
	"encoding/json"
	"strings"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// messageFields is the priority-ordered list of JSON fields to check
// for the primary log message text.
var messageFields = []string{"message", "msg", "error", "text", "log", "body"}

// ExtractText pulls classifiable text from a LogBuffer entry's JSON payload.
// Strategy:
//  1. Try known message fields in priority order
//  2. If a "level"/"severity" field exists, prepend it for classification context
//  3. Fall back to the full JSON string (Lumber handles raw JSON fine)
func ExtractText(entry db.LogBuffer) string {
	var obj map[string]any
	if err := json.Unmarshal(entry.Payload, &obj); err != nil {
		return strings.TrimSpace(string(entry.Payload))
	}

	var message string
	for _, field := range messageFields {
		if v, ok := obj[field]; ok {
			if s, ok := v.(string); ok && s != "" {
				message = s
				break
			}
		}
	}

	var prefix string
	for _, field := range []string{"level", "severity", "error_severity"} {
		if v, ok := obj[field]; ok {
			if s, ok := v.(string); ok && s != "" {
				prefix = strings.ToUpper(s)
				break
			}
		}
	}

	if message != "" {
		if prefix != "" {
			return prefix + ": " + message
		}
		return message
	}

	return strings.TrimSpace(string(entry.Payload))
}

// ExtractTexts extracts classifiable text from a batch of log entries.
// Returns texts in the same order as the input slice.
func ExtractTexts(logs []db.LogBuffer) []string {
	texts := make([]string, len(logs))
	for i, log := range logs {
		texts[i] = ExtractText(log)
	}
	return texts
}
