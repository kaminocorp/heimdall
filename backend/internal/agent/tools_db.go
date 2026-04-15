package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/connectors/database"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func (a *Agent) toolQueryDatabase(ctx context.Context, userID uuid.UUID, input map[string]any) (string, error) {
	sql, _ := input["sql"].(string)
	if sql == "" {
		return "", fmt.Errorf("query_database: sql parameter is required")
	}

	if !isReadOnlySQL(sql) {
		return "", fmt.Errorf("query_database: only SELECT, EXPLAIN, and WITH...SELECT statements are allowed")
	}

	connIDStr, _ := input["connection_id"].(string)
	if connIDStr == "" {
		return "", fmt.Errorf("query_database: connection_id parameter is required")
	}

	connID, err := uuid.Parse(connIDStr)
	if err != nil {
		return "", fmt.Errorf("query_database: invalid connection_id: %w", err)
	}

	// Fetch the connection — user-scoped to prevent unauthorized access.
	conn, err := a.queries.GetConnectionByUser(ctx, db.GetConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		return "", fmt.Errorf("query_database: connection not found or not owned by user")
	}

	// Validate connection type.
	if conn.Type != "postgres" {
		return "", fmt.Errorf("query_database: connection type %q is not a database", conn.Type)
	}

	// Create connector, connect, query, close.
	// TODO: each invocation opens a fresh connection — no pooling or reuse.
	// In a busy session (10 iterations × multiple queries) this creates
	// connection storms against the user's monitored database. Consider
	// caching connectors per connection ID for the loop duration.
	pg, err := database.New(conn.Config)
	if err != nil {
		return "", fmt.Errorf("query_database: %w", err)
	}

	if err := pg.Connect(ctx); err != nil {
		return "", fmt.Errorf("query_database: %w", err)
	}
	defer pg.Close()

	rows, err := pg.Query(ctx, sql)
	if err != nil {
		return "", fmt.Errorf("query_database: %w", err)
	}

	result, err := json.Marshal(map[string]any{
		"rows":      rows,
		"row_count": len(rows),
	})
	if err != nil {
		return "", fmt.Errorf("query_database: marshal: %w", err)
	}
	return string(result), nil
}

// writeKeywords are SQL keywords that indicate a mutation. Used by
// isReadOnlySQL to reject CTEs containing write operations.
var writeKeywords = []string{
	"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "TRUNCATE", "CREATE", "GRANT", "REVOKE",
}

// isReadOnlySQL validates that the SQL statement is a read-only query.
// Only SELECT, EXPLAIN, and WITH...SELECT (CTE) statements are allowed.
// This prevents LLM-generated write operations (UPDATE, DELETE, DROP, SET, etc.).
//
// Defence-in-depth: the Postgres connector also sets default_transaction_read_only=on,
// so the DB itself rejects writes. This function guards at the application layer so
// future non-Postgres connectors inherit the same protection.
func isReadOnlySQL(sql string) bool {
	// Normalize: trim whitespace and collapse to uppercase for matching.
	normalized := strings.TrimSpace(sql)
	if normalized == "" {
		return false
	}
	upper := strings.ToUpper(normalized)

	// Strip leading comments (-- line comments and /* block comments */).
	for {
		upper = strings.TrimSpace(upper)
		if strings.HasPrefix(upper, "--") {
			if idx := strings.Index(upper, "\n"); idx >= 0 {
				upper = upper[idx+1:]
				continue
			}
			return false // entire statement is a comment
		}
		if strings.HasPrefix(upper, "/*") {
			if idx := strings.Index(upper, "*/"); idx >= 0 {
				upper = upper[idx+2:]
				continue
			}
			return false // unclosed block comment
		}
		break
	}

	upper = strings.TrimSpace(upper)

	// Reject multi-statement queries: strip string literals then check for ';'.
	if containsSemicolon(upper) {
		return false
	}

	// SELECT, EXPLAIN, and WITH (CTE): allowed only if no write keywords
	// appear outside of string literals. The write-keyword scan is applied
	// even to SELECT statements as defence-in-depth — it blocks attempts
	// like "SELECT $$ $$ || (DELETE FROM ...)" where a write keyword sits
	// outside any quoted context.
	if strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "EXPLAIN") || strings.HasPrefix(upper, "WITH") {
		return !containsWriteKeyword(upper)
	}

	return false
}

// containsSemicolon checks for ';' outside of string literals (single-quoted
// and PostgreSQL dollar-quoted).
func containsSemicolon(upper string) bool {
	stripped := stripAllLiterals(upper)
	return strings.Contains(stripped, ";")
}

// containsWriteKeyword scans for SQL write keywords as whole words.
// It strips all string literals (single-quoted and dollar-quoted) before
// scanning so that keywords inside data values don't trigger false positives.
func containsWriteKeyword(upper string) bool {
	stripped := stripAllLiterals(upper)
	for _, kw := range writeKeywords {
		idx := 0
		for {
			pos := strings.Index(stripped[idx:], kw)
			if pos < 0 {
				break
			}
			pos += idx
			// Check word boundary: character before must not be alphanumeric/underscore.
			if pos > 0 && isWordChar(stripped[pos-1]) {
				idx = pos + len(kw)
				continue
			}
			// Character after must not be alphanumeric/underscore.
			end := pos + len(kw)
			if end < len(stripped) && isWordChar(stripped[end]) {
				idx = end
				continue
			}
			return true
		}
	}
	return false
}

// stripAllLiterals replaces the contents of both single-quoted and
// PostgreSQL dollar-quoted string literals with spaces, preserving
// string length so index positions remain stable. Dollar-quoting uses
// the form $tag$...$tag$ where tag can be empty ($$...$$) or any
// identifier ($foo$...$foo$).
func stripAllLiterals(s string) string {
	buf := []byte(s)
	i := 0
	for i < len(buf) {
		// Dollar-quoted strings: $tag$...$tag$
		if buf[i] == '$' {
			tag := extractDollarTag(buf, i)
			if tag != "" {
				// Skip past the opening tag.
				start := i + len(tag)
				// Find the closing tag.
				end := strings.Index(string(buf[start:]), tag)
				if end >= 0 {
					// Blank out everything between opening and closing tag (inclusive of tags).
					for j := i; j < start+end+len(tag); j++ {
						buf[j] = ' '
					}
					i = start + end + len(tag)
					continue
				}
				// Unclosed dollar-quote: blank to end.
				for j := i; j < len(buf); j++ {
					buf[j] = ' '
				}
				break
			}
		}
		// Single-quoted strings. PostgreSQL represents a literal single quote
		// inside a string as '' (two adjacent quotes). We must consume the
		// pair and keep blanking rather than treating the second ' as the
		// closing delimiter.
		if buf[i] == '\'' {
			i++ // skip opening quote
			for i < len(buf) {
				if buf[i] == '\'' {
					// Check for escaped quote (''): if the next char is also ',
					// blank both and continue inside the literal.
					if i+1 < len(buf) && buf[i+1] == '\'' {
						buf[i] = ' '
						buf[i+1] = ' '
						i += 2
						continue
					}
					i++ // skip closing quote
					break
				}
				buf[i] = ' '
				i++
			}
			continue
		}
		i++
	}
	return string(buf)
}

// extractDollarTag checks if buf[pos] starts a dollar-quote tag ($tag$)
// and returns the full tag string including both $ delimiters. Returns ""
// if no valid tag is found.
func extractDollarTag(buf []byte, pos int) string {
	if pos >= len(buf) || buf[pos] != '$' {
		return ""
	}
	// Scan for the closing $ of the tag.
	for end := pos + 1; end < len(buf); end++ {
		if buf[end] == '$' {
			return string(buf[pos : end+1])
		}
		// Tag characters must be identifier-like (letters, digits, underscore).
		// The first character after $ cannot be a digit (per PostgreSQL spec).
		c := buf[end]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_' || (c >= '0' && c <= '9' && end > pos+1)) {
			return ""
		}
	}
	return ""
}

func isWordChar(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '_'
}
