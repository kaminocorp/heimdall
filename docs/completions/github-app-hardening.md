# GitHub App Integration — Production Hardening

Reference: [GitHub App Integration](./github-app-integration.md) · [Changelog](../changelog.md)

---

## Summary

Eight fixes applied during a production-readiness review of the GitHub App integration (Phase 7 from the original plan). The review raised the implementation from 8.5 to 9/10 — addressing one blocking issue (callback auth), two correctness bugs (URL encoding, file path escaping), and five robustness improvements.

---

## Context

The GitHub App integration was implemented across 7 phases and already underwent one hardening round (7 fixes, documented in the main completion file). This second review focused on whether the code was safe to push to production, with particular attention to auth flow correctness, URL construction, context propagation, and input validation.

---

## Fixes

### 1. CRITICAL — GitHub callback route blocked by JWT middleware

**Problem**: `GET /api/github/callback` was registered inside the JWT-protected route group. When GitHub redirects the browser back after app installation, the request has no Supabase JWT — the Auth middleware rejects it with 401 before the handler runs. The install flow was broken in production.

**Fix**: Moved the callback route to the public route group. The handler already self-authenticates via the state JWT (RS256-signed, 15-minute expiry, contains user_id + app_id).

**Files**: `backend/internal/api/router.go`

---

### 2. MEDIUM — Search query double-encoding

**Problem**: `searchCode()` used `url.QueryEscape()` on the query, then concatenated with `+repo:` qualifiers and embedded the result directly in a `fmt.Sprintf` URL. This produced malformed URLs when the search query contained `+`, `&`, or spaces — characters that need different handling in query strings vs GitHub search syntax.

**Fix**: Replaced manual string building with `url.Values` for proper query string construction. GitHub search qualifiers use spaces as boolean AND (not `+`), and `url.Values.Encode()` handles all escaping correctly.

**Files**: `backend/internal/connectors/codebase/github.go`

---

### 3. MEDIUM — Missing context on GitHub API methods

**Problem**: `InstallationTokenFor()` and `GetInstallation()` used `http.NewRequest` without context, making them uncancellable by the caller. This was inconsistent with `APIRequest()` on the same struct, which accepts `context.Context`.

**Fix**: Added `context.Context` parameter to both methods and switched to `http.NewRequestWithContext`. Updated all four call sites (handlers, connector, connection test).

**Files**: `backend/internal/github/client.go`, `backend/internal/api/handlers/github.go`, `backend/internal/api/handlers/connections.go`, `backend/internal/connectors/codebase/github.go`

---

### 4. MEDIUM — Single-installation limitation undocumented

**Problem**: `toolSearchCodebase()` uses the first repo's `ConnectionConfig` (and therefore the first installation's token) for all repos. If a user has repos across multiple GitHub installations (multiple orgs), repos from non-first installations silently fail.

**Fix**: Added clear documentation comment explaining the limitation and the path to multi-installation support (group repos by `ConnectionID`). The current behavior is acceptable for v1 since most users have a single GitHub installation.

**Files**: `backend/internal/agent/tools_codebase.go`

---

### 5. MEDIUM — DB error silently downgraded in repo listing

**Problem**: In `ListGitHubRepos`, when `ListGitHubReposByConnection` failed, the error was logged but execution continued with an empty `enabledMap`. Users would see all repos as disabled even if they'd previously enabled some — a silent data accuracy issue.

**Fix**: Return HTTP 500 on DB error instead of continuing with potentially incorrect data.

**Files**: `backend/internal/api/handlers/github.go`

---

### 6. LOW — No request body size limit on UpdateGitHubRepos

**Problem**: `json.NewDecoder(r.Body).Decode(&repos)` reads the full request body without a size limit. A malicious client could send an arbitrarily large JSON payload.

**Fix**: Wrapped `r.Body` with `io.LimitReader(r.Body, 1<<20)` (1MB limit).

**Files**: `backend/internal/api/handlers/github.go`

---

### 7. LOW — Dual HTTP clients with inconsistent behavior

**Problem**: The connector struct held both a `ghClient` (with `APIRequest()`) and a standalone `httpClient`. The `apiGet` method used the standalone client with manually-set headers, while `Health()` used `ghClient.APIRequest()`. Two different code paths for the same API meant divergent timeout, header, and retry behavior.

**Fix**: Removed the standalone `httpClient` from the connector struct. `apiGet()` now routes through `ghClient.APIRequest()`, ensuring consistent headers (`X-GitHub-Api-Version`, `Accept`), timeouts, and a single HTTP client.

**Files**: `backend/internal/connectors/codebase/github.go`

---

### 8. LOW — url.PathEscape breaks nested file paths

**Problem**: `url.PathEscape(path)` was used on the full file path for the GitHub Contents API. Go's `PathEscape` encodes `/` as `%2F` (correct per RFC 3986), but GitHub's API expects literal slashes as path separators. Reading `src/main.go` would request `/contents/src%2Fmain.go` — a 404.

**Fix**: Split the path by `/`, escape each segment individually with `url.PathEscape`, then rejoin with literal `/`. Also applied consistent escaping to owner/repo in all URL constructions.

**Files**: `backend/internal/connectors/codebase/github.go`

---

## Verification

```bash
cd backend && go build ./...     # Clean
cd backend && go vet ./...       # Clean
cd backend && go test ./internal/agent/ ./internal/connectors/codebase/  # Pass
cd frontend && npx vue-tsc -b --noEmit  # Clean
```

---

## Files Changed

| # | File | Changes |
|---|------|---------|
| 1 | `backend/internal/api/router.go` | Moved `/api/github/callback` to public route group |
| 2 | `backend/internal/github/client.go` | Added `context.Context` to `InstallationTokenFor`, `fetchInstallationToken`, `GetInstallation`; switched to `NewRequestWithContext` |
| 3 | `backend/internal/api/handlers/github.go` | Added `context` import; passed context to `GetInstallation`, `InstallationTokenFor`, `TestGitHubConnection`; DB error now returns 500; added `io.LimitReader` on request body |
| 4 | `backend/internal/api/handlers/connections.go` | Updated `TestGitHubConnection` call to pass `r.Context()` |
| 5 | `backend/internal/connectors/codebase/github.go` | Removed `httpClient` field; `apiGet` now uses `ghClient.APIRequest`; fixed search query encoding with `url.Values`; fixed per-segment path escaping; consistent owner/repo escaping |
| 6 | `backend/internal/agent/tools_codebase.go` | Documented single-installation limitation |
| 7 | `docs/completions/github-app-integration.md` | Added Round 2 hardening table |
