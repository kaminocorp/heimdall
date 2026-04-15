package agent

import "testing"

func TestIsReadOnlySQL(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want bool
	}{
		// Allowed queries.
		{"simple select", "SELECT * FROM users", true},
		{"select with where", "SELECT id FROM users WHERE name = 'Alice'", true},
		{"explain", "EXPLAIN SELECT 1", true},
		{"explain analyze", "EXPLAIN ANALYZE SELECT 1", true},
		{"simple CTE", "WITH x AS (SELECT 1) SELECT * FROM x", true},
		{"CTE with keyword in string literal", "WITH x AS (SELECT 'DELETE ME' AS val) SELECT * FROM x", true},
		{"leading line comment", "-- comment\nSELECT 1", true},
		{"leading block comment", "/* block */  SELECT 1", true},
		{"select with semicolon in string", "SELECT 'a;b' FROM t", true},

		// Rejected queries.
		{"empty string", "", false},
		{"insert", "INSERT INTO users VALUES (1)", false},
		{"update", "UPDATE users SET name = 'x'", false},
		{"delete", "DELETE FROM users", false},
		{"drop table", "DROP TABLE users", false},
		{"CTE with delete", "WITH x AS (DELETE FROM users RETURNING *) SELECT * FROM x", false},
		{"CTE with insert", "WITH x AS (INSERT INTO t VALUES (1) RETURNING *) SELECT * FROM x", false},
		{"CTE with update", "WITH x AS (UPDATE t SET a=1 RETURNING *) SELECT * FROM x", false},
		{"multi-statement select then drop", "SELECT 1; DROP TABLE users", false},
		{"multi-statement with spaces", "SELECT 1 ; DELETE FROM t", false},
		{"dollar-quoted semicolon bypass", "SELECT $$;$$ FROM t; DROP TABLE users", false},
		{"dollar-quoted write keyword", "SELECT $$ DELETE FROM users $$", true},
		{"dollar-tagged write keyword", "SELECT $tag$ INSERT INTO t $tag$", true},
		{"dollar-quoted semicolon in select", "SELECT $$hello;world$$", true},
		{"escaped single quote in literal", "SELECT 'it''s fine' FROM t", true},
		{"escaped quote with semicolon inside", "SELECT 'foo''bar;baz' FROM t", true},
		{"escaped quote with write keyword inside", "SELECT 'it''s a DELETE FROM users' FROM t", true},
		{"dollar-quoted bypass attempt", "SELECT $$;$$ || (DELETE FROM users)", false},
		{"unclosed dollar-quote", "SELECT $$ never closed", true},
		{"unclosed block comment", "/* never closed", false},
		{"comment-only", "-- just a comment", false},
		{"create table", "CREATE TABLE t (id int)", false},
		{"alter table", "ALTER TABLE t ADD COLUMN x int", false},
		{"truncate", "TRUNCATE users", false},
		{"grant", "GRANT ALL ON users TO public", false},

		// Dangerous function blocklist (H1).
		{"set_config bypass", "SELECT set_config('default_transaction_read_only', 'off', false)", false},
		{"pg_read_file", "SELECT pg_read_file('/etc/passwd')", false},
		{"pg_write_file", "SELECT pg_write_file('/tmp/evil', 'data')", false},
		{"lo_import", "SELECT lo_import('/etc/passwd')", false},
		{"lo_export", "SELECT lo_export(12345, '/tmp/evil')", false},
		{"set_config in CTE", "WITH x AS (SELECT set_config('a','b',true)) SELECT * FROM x", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isReadOnlySQL(tt.sql)
			if got != tt.want {
				t.Errorf("isReadOnlySQL(%q) = %v, want %v", tt.sql, got, tt.want)
			}
		})
	}
}
