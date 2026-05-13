package db_test

import (
	"context"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// rlsPolicyAllowlist enumerates the per-table verb coverage the catalog
// drift check expects. Keys are public-schema table names with
// `rowsecurity = true`; values are the set of policy `cmd`s that must be
// present (mapped from pg_policies.cmd; "ALL" means a single FOR ALL
// policy). Empty value = "RLS-enabled by design but no policy" — only
// schema_migrations matches that shape today.
//
// The allowlist commits the *Phase-7 target state*. Tables that Phase 1
// §5.2 surfaced as RLS-enabled-without-policies (agent_config,
// webhook_idempotency, log_pipeline_events) are listed here with the
// verb sets Migration B will produce. The catalog test will fail until
// Migration B ships — that's by design; Phase 7's gating accept-check
// is "this test goes green."
//
// Adding a new RLS-enabled table: update this allowlist in the same PR
// as the migration. The pairing is the diff reviewers look for.
var rlsPolicyAllowlist = map[string][]string{
	// Tenant-scoped tables — full CRUD policies (exposed as "ALL").
	// users: users_self (FOR ALL) for the caller's own row + users_org_visible
	// (FOR SELECT, added in migration 040) for shared-org sibling reads.
	"users":                    {"ALL", "SELECT"},
	"organizations":            {"ALL"},
	"org_members":              {"ALL"},
	"applications":             {"ALL"},
	"connections":              {"ALL"},
	"connection_sources":       {"ALL"},
	"app_source_filters":       {"ALL"},
	"log_buffer":               {"ALL"},
	"conversations":            {"ALL"},
	"investigations":           {"ALL"},
	"agent_log":                {"ALL"},
	"app_agent_config":         {"ALL"},
	"monitoring_state":         {"ALL"},
	"investigation_schedules":  {"ALL"},
	"notification_channels":    {"ALL"},
	"notification_preferences": {"ALL"},
	"notification_log":         {"ALL"},

	// Phase 1 §5.2 surfaced as RLS-enabled-without-policies. Migration 040
	// (Phase 7) shipped these policies; the allowlist now matches the
	// catalog state on a Migration-040-applied DB.
	"agent_config":        {"SELECT"}, // 040+P9: agent_config_global FOR SELECT USING (true) — singleton, reads only; writes go through DIRECT_URL.
	"webhook_idempotency": {"ALL"},    // 040: scope by connection_id → connections_org_member transitively.
	"log_pipeline_events": {"ALL"},    // 040: scope by app_id → applications_user_policy transitively.

	// schema_migrations is RLS-enabled (defence-in-depth against
	// PostgREST) but is touched only by `make migrate-up` via DIRECT_URL
	// (postgres). No runtime role reads it, no policy is needed. Empty
	// verb list encodes "rowsecurity=true, no policies, by design."
	"schema_migrations": {},
}

// cronWriteAllowlist is the §5.6a Test 5 source-of-truth for which
// tables `cron_user` may write to (and which verbs). Anything in
// information_schema.role_table_grants beyond this set is a regression.
//
// Matches Migration 039's narrow-write grants exactly:
//   - monitoring_state INSERT/UPDATE  (cursor advancement, parent plan §4.2)
//   - log_buffer       DELETE         (Phase-1 audit-discovered widening for the pruner)
var cronWriteAllowlist = map[string]map[string]bool{
	"monitoring_state": {"INSERT": true, "UPDATE": true},
	"log_buffer":       {"DELETE": true},
}

// TestCatalogDrift_ForceCoverage (§5.6a Test 1) asserts every public
// table with `rowsecurity = true` also has `forcerowsecurity = true`.
//
// Expected to be skipped until Phase 7's Migration B ships. Gate via
// HEIMDALL_RLS_FORCE_LIVE=1 so CI stays green between Phase 5 and
// Phase 7. Phase 7 sets the env var in CI as part of its rollout, and
// at that point this test becomes the load-bearing acceptance check.
func TestCatalogDrift_ForceCoverage(t *testing.T) {
	env := requireRLSEnv(t)
	ctx := context.Background()

	if !envFlagOn("HEIMDALL_RLS_FORCE_LIVE") {
		t.Skip("HEIMDALL_RLS_FORCE_LIVE not set; FORCE coverage check skipped until Phase 7 (Migration B) ships")
	}

	rows, err := env.pool.Query(ctx,
		`SELECT tablename
		 FROM pg_tables
		 WHERE schemaname = 'public'
		   AND rowsecurity = true
		   AND forcerowsecurity = false
		 ORDER BY tablename`)
	require.NoError(t, err)
	defer rows.Close()

	var missing []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		missing = append(missing, name)
	}
	require.NoError(t, rows.Err())

	require.Empty(t, missing,
		"tables with rowsecurity=true but forcerowsecurity=false (Phase 7 must FORCE these): %v", missing)
}

// TestCatalogDrift_PolicyVerbCoverage (§5.6a Test 2) asserts every
// rowsecurity=true table's policy verb set matches rlsPolicyAllowlist
// exactly. Bidirectional: extra catalog rows AND missing catalog rows
// both fail. That bidirectionality is what makes this a *drift* check
// rather than an existence check.
func TestCatalogDrift_PolicyVerbCoverage(t *testing.T) {
	env := requireRLSEnv(t)
	ctx := context.Background()

	rows, err := env.pool.Query(ctx,
		`SELECT t.tablename,
		        coalesce(array_agg(DISTINCT p.cmd ORDER BY p.cmd) FILTER (WHERE p.cmd IS NOT NULL), '{}')::text[] AS verbs
		 FROM pg_tables t
		 LEFT JOIN pg_policies p
		   ON p.schemaname = t.schemaname AND p.tablename = t.tablename
		 WHERE t.schemaname = 'public' AND t.rowsecurity = true
		 GROUP BY t.tablename
		 ORDER BY t.tablename`)
	require.NoError(t, err)
	defer rows.Close()

	got := map[string][]string{}
	for rows.Next() {
		var name string
		var verbs []string
		require.NoError(t, rows.Scan(&name, &verbs))
		got[name] = verbs
	}
	require.NoError(t, rows.Err())

	for name, verbs := range got {
		expected, ok := rlsPolicyAllowlist[name]
		if !ok {
			t.Errorf("table %q is RLS-enabled in the catalog but missing from rlsPolicyAllowlist; add it (or drop the migration that enabled RLS)", name)
			continue
		}
		require.ElementsMatch(t, normaliseVerbs(expected), normaliseVerbs(verbs),
			"table %q verb-set mismatch: allowlist=%v catalog=%v", name, expected, verbs)
	}
	for name := range rlsPolicyAllowlist {
		if _, ok := got[name]; !ok {
			t.Errorf("rlsPolicyAllowlist entry %q has no rowsecurity=true row in the catalog; the table was dropped or RLS was disabled", name)
		}
	}
}

// TestCatalogDrift_AppUserAttributes (§5.6a Test 3).
func TestCatalogDrift_AppUserAttributes(t *testing.T) {
	env := requireRLSEnv(t)
	ctx := context.Background()
	requireRoleExists(t, ctx, env.pool, "app_user")

	var bypass, super bool
	err := env.pool.QueryRow(ctx,
		"SELECT rolbypassrls, rolsuper FROM pg_roles WHERE rolname = 'app_user'").Scan(&bypass, &super)
	require.NoError(t, err)
	require.False(t, bypass, "app_user must not have BYPASSRLS — that defeats the entire role split")
	require.False(t, super, "app_user must not be SUPERUSER")
}

// TestCatalogDrift_CronUserAttributes (§5.6a Test 4).
func TestCatalogDrift_CronUserAttributes(t *testing.T) {
	env := requireRLSEnv(t)
	ctx := context.Background()
	requireRoleExists(t, ctx, env.pool, "cron_user")

	var super, bypass bool
	err := env.pool.QueryRow(ctx,
		"SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = 'cron_user'").Scan(&super, &bypass)
	require.NoError(t, err)
	require.False(t, super, "cron_user must not be SUPERUSER (BYPASSRLS is the only elevated attribute)")
	require.True(t, bypass, "cron_user must have BYPASSRLS — without it, cross-tenant enumeration breaks")
}

// TestCatalogDrift_CronWriteAllowlist (§5.6a Test 5) asserts
// information_schema.role_table_grants for `cron_user` matches
// cronWriteAllowlist exactly across the write verbs (INSERT/UPDATE/DELETE).
// SELECT is granted blanket and is not enumerated here.
func TestCatalogDrift_CronWriteAllowlist(t *testing.T) {
	env := requireRLSEnv(t)
	ctx := context.Background()
	requireRoleExists(t, ctx, env.pool, "cron_user")

	rows, err := env.pool.Query(ctx,
		`SELECT table_name, privilege_type
		 FROM information_schema.role_table_grants
		 WHERE grantee = 'cron_user'
		   AND table_schema = 'public'
		   AND privilege_type IN ('INSERT', 'UPDATE', 'DELETE')
		 ORDER BY table_name, privilege_type`)
	require.NoError(t, err)
	defer rows.Close()

	got := map[string]map[string]bool{}
	for rows.Next() {
		var table, verb string
		require.NoError(t, rows.Scan(&table, &verb))
		if got[table] == nil {
			got[table] = map[string]bool{}
		}
		got[table][verb] = true
	}
	require.NoError(t, rows.Err())

	for table, verbs := range got {
		allowedVerbs, ok := cronWriteAllowlist[table]
		if !ok {
			t.Errorf("cron_user has unexpected write grants on %q (verbs=%v); add to cronWriteAllowlist with justification, or revoke in a migration", table, sortedKeys(verbs))
			continue
		}
		for v := range verbs {
			if !allowedVerbs[v] {
				t.Errorf("cron_user has unexpected %s grant on %q; allowlist permits only %v", v, table, sortedKeys(allowedVerbs))
			}
		}
	}
	for table, allowedVerbs := range cronWriteAllowlist {
		gotVerbs := got[table]
		for v := range allowedVerbs {
			if !gotVerbs[v] {
				t.Errorf("cronWriteAllowlist claims cron_user has %s on %q but the catalog disagrees; either the migration didn't apply or the allowlist is stale", v, table)
			}
		}
	}
}

// envFlagOn returns true when the named env var is set to a truthy
// value (1, true, yes, on — case-insensitive). Used by tests gated on
// phase-specific rollout flags.
func envFlagOn(name string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// normaliseVerbs returns a sorted, non-nil copy of `in` so
// require.ElementsMatch comparisons are deterministic and the
// "empty by design" entry for schema_migrations matches `[]string{}`
// rather than a nil slice.
func normaliseVerbs(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	sort.Strings(out)
	if len(out) == 0 {
		return []string{}
	}
	return out
}

func sortedKeys[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
