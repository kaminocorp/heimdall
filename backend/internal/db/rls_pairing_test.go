package db_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// rawDBAccessAllowlist enumerates every production-code file (relative
// to the `backend/` module root) that legitimately uses a raw DB-access
// shape — `pgxpool.New*`, `Pool.Begin/Exec/Query`, `App.Begin`,
// `Cron.Begin` — without going through the `Pools.UserQueries` /
// `Pools.CronQueries` / `Pools.WithUserQueries` chokepoint.
//
// Each entry's value is the load-bearing justification. Reviewers should
// reject any new allowlist entry whose justification reduces to "we
// haven't refactored this yet" — the parent plan §5.6c is explicit that
// the allowlist is the failsafe, not the hammock.
//
// Adding a new entry: pair it with the production change in the same PR,
// and let CI's TestPairing_AllowlistJustifications subtest read the
// justification back out as a contract.
var rawDBAccessAllowlist = map[string]string{
	// The chokepoint itself. Pools.UserQueries / Pools.CronQueries /
	// Pools.WithUserQueries are defined here against `App.Begin` /
	// raw `New(p.Cron)` — the file's whole purpose is to be the one
	// place those primitives are called.
	"internal/db/pools.go": "primary chokepoint; UserQueries / CronQueries / WithUserQueries originate here",

	// Bootstrap. main.go calls pgxpool.NewWithConfig to construct the
	// two physical pools, then runs `SELECT current_user` against each
	// for the Phase 3 production invariant. Neither path touches tenant
	// data; both are pre-Pools-construction by definition.
	"cmd/heimdall/main.go": "pgxpool.NewWithConfig pool construction + assertRoleSplit's SELECT current_user; pre-Pools wiring",

	// Liveness probe. Pools.App.Ping() and Pools.Cron.Ping() are no-data
	// round-trips used by the health endpoint; routing them through
	// UserQueries would require a userID that the health endpoint by
	// definition does not have. Phase 9 split the probe across both pools
	// so a cron-pool outage isn't invisible to the LB.
	"internal/api/handlers/health.go": "Pools.App.Ping() + Pools.Cron.Ping() liveness probes; no tenant data",
}

// rawDBAccessPatterns is the regex set used by the tripwire. Each
// pattern is one shape of "this file is opening a transaction or
// running a statement against a raw pool / tx without going through the
// chokepoint." The set is deliberately conservative — false positives
// just mean a new allowlist entry; false negatives mean the tripwire
// silently lets a raw-access file ship.
var rawDBAccessPatterns = []*regexp.Regexp{
	regexp.MustCompile(`pgxpool\.New(WithConfig)?\b`),
	regexp.MustCompile(`\bPool\.Begin\(`),
	regexp.MustCompile(`\bPool\.Exec\(`),
	regexp.MustCompile(`\bPool\.Query\(`),
	regexp.MustCompile(`\bPool\.QueryRow\(`),
	regexp.MustCompile(`\bPool\.Ping\(`),
	// pools.go's internal entry point — match here too so the tripwire
	// notices files that copy the pattern instead of calling through.
	regexp.MustCompile(`\bApp\.Begin\(`),
	regexp.MustCompile(`\bCron\.Begin\(`),
	// Phase 9: Pools.App.Ping / Pools.Cron.Ping shapes from health.go.
	// `\bPool\.Ping\(` doesn't match `s.Pools.App.Ping(` because the
	// word boundary between `Pools` and `.` blocks it; without these
	// extra patterns the health.go allowlist entry would be stale and
	// future raw `App.Ping()` adoption elsewhere would slip past unflagged.
	regexp.MustCompile(`\bApp\.Ping\(`),
	regexp.MustCompile(`\bCron\.Ping\(`),
}

// chokepointReferences are the symbols that, when present in a file,
// signal "this file uses the chokepoint." Any of them is sufficient to
// rebut a raw-access match.
var chokepointReferences = []string{
	"UserQueries",
	"WithUserQueries",
	"CronQueries",
}

// TestPairing_RawDBAccessIsAllowlisted (§5.6c) walks every production
// `.go` file in `backend/internal/` and `backend/cmd/`, finds files that
// match a raw-DB-access pattern, and asserts each one either references
// the chokepoint API or is in `rawDBAccessAllowlist`.
//
// Files matched: production code only (`_test.go` excluded; sqlc-generated
// `*.sql.go` excluded — the generated code uses `*Queries` methods that
// are themselves a chokepoint of a different kind).
func TestPairing_RawDBAccessIsAllowlisted(t *testing.T) {
	backendRoot := backendRootFromTest(t)

	roots := []string{
		filepath.Join(backendRoot, "internal"),
		filepath.Join(backendRoot, "cmd"),
	}

	var violations []string
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !isProductionGoFile(path) {
				return nil
			}

			data, readErr := os.ReadFile(path)
			require.NoError(t, readErr)
			content := string(data)

			if !matchesAny(content, rawDBAccessPatterns) {
				return nil
			}
			if referencesAny(content, chokepointReferences) {
				return nil
			}

			rel, relErr := filepath.Rel(backendRoot, path)
			require.NoError(t, relErr)
			rel = filepath.ToSlash(rel)

			if _, ok := rawDBAccessAllowlist[rel]; ok {
				return nil
			}

			violations = append(violations, rel)
			return nil
		})
		require.NoError(t, err)
	}

	require.Empty(t, violations,
		"files with raw DB access that neither use the Pools chokepoint nor appear in rawDBAccessAllowlist:\n  %s\n\nFix the file to route through Pools.UserQueries / Pools.CronQueries / Pools.WithUserQueries, or add an allowlist entry with a load-bearing justification.",
		strings.Join(violations, "\n  "))
}

// TestPairing_AllowlistJustifications (§5.6c second sub-test) asserts
// every allowlist entry still names an existing file AND that the file
// still matches a raw-access pattern. Catches stale allowlist drift —
// e.g. a file that's been refactored to use the chokepoint but kept its
// allowlist entry "just in case."
func TestPairing_AllowlistJustifications(t *testing.T) {
	backendRoot := backendRootFromTest(t)

	for rel, justification := range rawDBAccessAllowlist {
		t.Run(rel, func(t *testing.T) {
			require.NotEmpty(t, justification,
				"rawDBAccessAllowlist entry for %q has no justification; remove it or document why raw access is required", rel)

			path := filepath.Join(backendRoot, filepath.FromSlash(rel))
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("rawDBAccessAllowlist references missing file %q: %v", rel, err)
			}
			content := string(data)

			if !matchesAny(content, rawDBAccessPatterns) {
				t.Errorf("rawDBAccessAllowlist entry %q no longer matches any raw-access pattern; remove the allowlist entry", rel)
			}
		})
	}
}

// backendRootFromTest derives the absolute path of `backend/` from the
// runtime location of this test file. The test file lives at
// `backend/internal/db/rls_pairing_test.go`, so `backend/` is two
// directories up. Using runtime.Caller keeps the test independent of
// the CWD `go test` happens to be invoked in.
func backendRootFromTest(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed; cannot derive backend root")
	dbDir := filepath.Dir(thisFile)            // backend/internal/db
	internalDir := filepath.Dir(dbDir)         // backend/internal
	backendRoot := filepath.Dir(internalDir)   // backend
	return backendRoot
}

// isProductionGoFile returns true for `.go` files that are not Go test
// files and not sqlc-generated query bindings. The generated `*.sql.go`
// files routinely match raw-access patterns (they call `db.QueryRow` on
// a `DBTX` interface) but are themselves the chokepoint sqlc layer —
// excluding them isn't a soft pass, it's that they're already gated by
// the `*Queries` interface every caller routes through.
func isProductionGoFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".go") {
		return false
	}
	if strings.HasSuffix(base, "_test.go") {
		return false
	}
	if strings.HasSuffix(base, ".sql.go") {
		return false
	}
	return true
}

func matchesAny(content string, patterns []*regexp.Regexp) bool {
	for _, p := range patterns {
		if p.MatchString(content) {
			return true
		}
	}
	return false
}

func referencesAny(content string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(content, n) {
			return true
		}
	}
	return false
}
