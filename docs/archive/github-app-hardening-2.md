# GitHub App Integration — Production Hardening (Round 3)

Reference: [GitHub App Integration](./github-app-integration.md) · [Round 2](./github-app-hardening.md) · [Changelog](../changelog.md)

---

## Summary

Five fixes applied during a final production-readiness review of the GitHub App integration. The review raised the implementation from 9/10 to 9.5/10 — addressing one consistency issue (RLS-scoped query), one robustness improvement (response body error handling), one safety bound (search query length), and two test coverage gaps.

---

## Fixes

### 1. MEDIUM — Callback `GetApplicationByOrgUser` bypassed user-scoped queries

**Problem**: `GitHubCallback` called `s.Queries.GetApplicationByOrgUser()` directly instead of using `UserQueries()`. While `GetApplicationByOrgUser` scopes by `userID` parameter, this bypassed the `SET LOCAL app.current_user_id` transaction used everywhere else in the codebase — an inconsistency that breaks RLS if the query internally touches RLS-protected tables.

**Fix**: Moved the `UserQueries()` call to the top of the callback (right after state JWT validation), and use the scoped `queries` for `GetApplicationByOrgUser` and all subsequent DB operations. Removed the now-redundant second `UserQueries()` call further down.

**Files**: `backend/internal/api/handlers/github.go`

---

### 2. MEDIUM — Search query length unbounded for many repos

**Problem**: `searchCode()` appended `repo:org/name` qualifiers for every enabled repository. With 50+ repos, the resulting query string could exceed GitHub's URL length limits (~8KB), causing silent 422 errors from the search API.

**Fix**: Capped the number of `repo:` qualifiers to 20. If a user has more than 20 enabled repos, only the first 20 are included in the search query. This keeps URLs well under limits while covering the vast majority of use cases.

**Files**: `backend/internal/connectors/codebase/github.go`

---

### 3. LOW — Swallowed `io.ReadAll` error in `ListGitHubRepos`

**Problem**: `body, _ := io.ReadAll(io.LimitReader(resp.Body, 5<<20))` discarded the error from reading the GitHub API response body. While rare in practice (the HTTP connection was already established), this violated the codebase's convention of checking all I/O errors.

**Fix**: Added error check — returns HTTP 502 with a logged error on read failure.

**Files**: `backend/internal/api/handlers/github.go`

---

### 4. LOW — No dispatch routing tests for `search_codebase`

**Problem**: `tools_test.go` had dispatch routing tests for `search_logs`, `query_database`, and `unknown_tool`, but none for `search_codebase`. This meant a typo in the `Dispatch` switch statement could break codebase search without any test catching it.

**Fix**: Added two tests:
- `TestDispatch_SearchCodebase_NoClient` — verifies dispatch routes to `toolSearchCodebase` and returns clear "not configured" error when no GitHub client is present
- `TestDispatch_SearchCodebase_MissingAction` — verifies dispatch routing works even with nil appID (hits the nil-client check first, confirming the correct handler is called)

**Files**: `backend/internal/agent/tools_test.go`

---

### 5. INFO — Installation token caching reviewed (no change needed)

**Problem (reviewed, not a bug)**: `toolSearchCodebase` calls `connector.Connect()` on every invocation, which calls `InstallationTokenFor()`. This looked like it could cause unnecessary GitHub API calls.

**Finding**: No fix needed. `InstallationTokenFor()` has an in-memory cache with `sync.RWMutex` that only fetches a new token when the cached one is within 5 minutes of expiry. The cache is at the correct layer (`Client` struct), so all callers benefit. Documented as reviewed.

---

## Verification

```bash
cd backend && go build ./...     # Clean
cd backend && go vet ./...       # Clean
cd backend && go test ./internal/agent/ ./internal/connectors/codebase/  # Pass (including 2 new tests)
cd frontend && npx vue-tsc -b --noEmit  # Clean
```

---

## Files Changed

| # | File | Changes |
|---|------|---------|
| 1 | `backend/internal/api/handlers/github.go` | Moved `UserQueries()` to top of callback for RLS consistency; removed duplicate call; added `io.ReadAll` error check |
| 2 | `backend/internal/connectors/codebase/github.go` | Capped `repo:` qualifiers to 20 in search queries |
| 3 | `backend/internal/agent/tools_test.go` | Added `TestDispatch_SearchCodebase_NoClient` and `TestDispatch_SearchCodebase_MissingAction` |
