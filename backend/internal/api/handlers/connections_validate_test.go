package handlers

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestValidateSourceNamePathConfig exercises the write-time validator that
// catches typos in `source_name_path` before they turn into silent
// drop-by-default at ingestion time. The ingestion extractor (webhooks.go
// extractStringByPath) tolerates all of these gracefully, so without a
// write-time check the operator's only signal that something is wrong is
// "none of my entries are making it through."
func TestValidateSourceNamePathConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{"empty config", ``, ""},
		{"null config", `null`, ""},
		{"no path field", `{"webhook_token": "abc"}`, ""},
		{"path null", `{"source_name_path": null}`, ""},
		{"path absent from object", `{}`, ""},
		{"empty string path", `{"source_name_path": ""}`, ""},
		{"whitespace-only path", `{"source_name_path": "   "}`, ""},

		{"simple path", `{"source_name_path": "log_group"}`, ""},
		{"nested path", `{"source_name_path": "meta.source"}`, ""},
		{"deeply nested", `{"source_name_path": "a.b.c.d.e"}`, ""},

		{"leading dot", `{"source_name_path": ".log_group"}`, "must not start or end with a dot"},
		{"trailing dot", `{"source_name_path": "log_group."}`, "must not start or end with a dot"},
		{"consecutive dots", `{"source_name_path": "a..b"}`, "empty segments"},
		{"triple dots", `{"source_name_path": "a...b"}`, "empty segments"},
		{"single dot", `{"source_name_path": "."}`, "must not start or end with a dot"},
		{"leading whitespace", `{"source_name_path": " log_group"}`, "leading or trailing whitespace"},
		{"trailing whitespace", `{"source_name_path": "log_group "}`, "leading or trailing whitespace"},
		{"space in segment", `{"source_name_path": "log group"}`, "must not contain whitespace"},
		{"tab in segment", `{"source_name_path": "a\tb"}`, "must not contain whitespace"},

		// The helper is deliberately tolerant of unrelated JSON shape
		// problems — those are caught by the parent connector-specific
		// validator (NewXyz(...) rejects unknown fields / wrong types).
		{"malformed outer JSON", `{not-json`, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSourceNamePathConfig(json.RawMessage(tc.config))
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}
