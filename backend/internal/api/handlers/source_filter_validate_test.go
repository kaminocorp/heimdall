package handlers

import (
	"strings"
	"testing"
)

// TestValidateSourceName covers the shared helper used by POST/PUT/DELETE
// on /api/connections/{id}/sources. The cases here are the exhaustive
// acceptance/rejection surface for source_name, so any regression in one
// endpoint is caught without needing the DB-backed integration harness.
func TestValidateSourceName(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		wantOK    bool
		wantName  string
		wantErrIs string
	}{
		{"plain", "my-app", true, "my-app", ""},
		{"with slash", "vercel/lambda", true, "vercel/lambda", ""},
		{"unicode", "日本語-app", true, "日本語-app", ""},
		{"trim leading whitespace", "  trimmed", true, "trimmed", ""},
		{"trim trailing whitespace", "trimmed  ", true, "trimmed", ""},
		{"trim both", "  trimmed  ", true, "trimmed", ""},
		{"max length (256)", strings.Repeat("a", 256), true, strings.Repeat("a", 256), ""},

		{"empty", "", false, "", "required"},
		{"whitespace only", "   ", false, "", "required"},
		{"tab only", "\t\t", false, "", "required"},
		{"too long (257)", strings.Repeat("a", 257), false, "", "exceeds 256"},
		{"null byte in middle", "foo\x00bar", false, "", "null byte"},
		{"null byte at end", "foo\x00", false, "", "null byte"},
		{"null byte at start", "\x00foo", false, "", "null byte"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateSourceName(tc.in)
			if tc.wantOK {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				if got != tc.wantName {
					t.Fatalf("name = %q, want %q", got, tc.wantName)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got success with name %q", tc.wantErrIs, got)
			}
			if !strings.Contains(err.Error(), tc.wantErrIs) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErrIs)
			}
		})
	}
}
