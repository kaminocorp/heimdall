# GitHub App Integration — Completion Notes

Reference: [Implementation Plan](../executing/github-app-integration.md) · [Changelog](../changelog.md)

---

## Summary

Implemented the full GitHub App integration across all 7 phases defined in the plan. Heimdall can now authenticate as a GitHub App, let users install it on their GitHub organizations, manage which repos are accessible, and use a `search_codebase` agent tool to search/read code during both interactive chat and monitoring investigations.

---

## Phase 1: GitHub App Client & Token Management

### New Files

| File | Purpose |
|------|---------|
| `backend/internal/github/client.go` | Core GitHub App client: RSA private key parsing, JWT generation (RS256, 10min expiry), installation token fetching with in-memory cache (`sync.Map` keyed by installation ID, 5-min refresh margin), and authenticated API request helper |

### Modified Files

| File | Change |
|------|--------|
| `backend/internal/config/config.go` | Added `GitHubAppID`, `GitHubPrivateKey`, `GitHubClientID`, `GitHubAppSlug`, `GitHubWebhookSecret` (all optional env vars) |

### Implementation Details

- **JWT generation**: Standard GitHub App JWT — `iss` = App ID, `iat` = now-60s, `exp` = now+10min, signed RS256 with the app's private key. Uses `github.com/golang-jwt/jwt/v5` (already in go.mod).
- **Private key loading**: Supports both raw PEM string in env var and file path (detects by `-----BEGIN` prefix).
- **Nil-safe pattern**: If `GitHubAppID` is empty, `NewClient()` returns `(nil, nil)`. All consumers check for nil before use (matches the existing `notifier` pattern).
- **Installation token cache**: In-memory `sync.RWMutex`-protected map keyed by installation ID. Tokens refreshed 5 minutes before expiry. Each call to `InstallationTokenFor()` checks cache first, then fetches from `POST /app/installations/{id}/access_tokens`.
- **HTTP client**: Dedicated 10-second timeout client for all GitHub API calls.
- **Context-aware API requests**: `APIRequest()` accepts `context.Context` for caller-controlled cancellation and timeout propagation.

---

## Phase 2: Installation Flow (Backend)

### New Files

| File | Purpose |
|------|---------|
| `backend/internal/api/handlers/github.go` | `InstallGitHub`, `GitHubCallback`, `ListGitHubRepos`, `UpdateGitHubRepos`, `TestGitHubConnection` handler methods |

### Modified Files

| File | Change |
|------|--------|
| `backend/internal/api/handlers/server.go` | Added `GitHub *github.Client` field to `Server` struct; updated `NewServer` signature to accept `*github.Client` |
| `backend/internal/api/router.go` | Updated `NewRouter` signature to accept `*github.Client`; registered `GET /api/github/install` and `GET /api/github/callback` in protected group |
| `backend/cmd/heimdall/main.go` | Conditional GitHub client initialization; passed to `agent.New()` and `api.NewRouter()` |
| `backend/internal/api/handlers/connections.go` | `TestConnection` now handles `type="github"` — verifies installation token validity |

### Endpoints

| Method | Path | Handler | Auth |
|--------|------|---------|------|
| `GET` | `/api/github/install?app_id={appId}` | `InstallGitHub` | JWT |
| `GET` | `/api/github/callback?installation_id={id}&state={state}` | `GitHubCallback` | State JWT |

### How the Install Flow Works

1. Frontend calls `GET /api/github/install?app_id=X`
2. Backend generates a signed **state JWT** (RS256 with the GitHub App private key) containing `{user_id, app_id, nonce, exp}` — 15-minute expiry
3. Returns `{url: "https://github.com/apps/{slug}/installations/new?state={jwt}"}` (slug from `GITHUB_APP_SLUG` env var, defaults to `heimdall-agent`)
4. Frontend redirects browser to the GitHub install URL
5. User installs the app on their GitHub org, selects repos
6. GitHub redirects to `GET /api/github/callback?installation_id=X&setup_action=install&state={jwt}`
7. Backend validates the state JWT (signature + expiry), verifies the installation via `GET /app/installations/{id}`
8. Creates a connection record: `type="github"`, `direction="two_way"`, `config={"installation_id": N, "account_login": "...", "account_type": "Organization|User"}`, `status="active"`
9. If a connection with the same `installation_id` already exists for this app → updates it (upsert semantics for re-installs)
10. Redirects browser to `/connections?github=installed`

### Edge Cases Handled

- **Re-install / modify permissions**: `setup_action=update` triggers update of existing connection config
- **Duplicate prevention**: Searches existing connections for matching `installation_id` before creating a new one. Errors from `ListConnectionsByApp` are returned to the user (not swallowed) to prevent silent duplicate creation.

---

## Phase 3: Repository Management

### Database Migration

**`backend/migrations/020_github_repos.up.sql`**:
```sql
CREATE TABLE github_repos (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id   UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    repo_full_name  TEXT NOT NULL,
    repo_id         BIGINT NOT NULL,
    default_branch  TEXT NOT NULL DEFAULT 'main',
    enabled         BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_github_repos_conn_repo ON github_repos(connection_id, repo_id);
CREATE INDEX idx_github_repos_connection_id ON github_repos(connection_id);

-- Row Level Security: scope via parent connection's user_id.
ALTER TABLE github_repos ENABLE ROW LEVEL SECURITY;
CREATE POLICY github_repos_owner ON github_repos
  FOR ALL
  USING (connection_id IN (SELECT id FROM connections WHERE user_id = app_current_user_id()));
```

### sqlc Queries

| File | Query | Type | Purpose |
|------|-------|------|---------|
| `github_repos.sql` | `ListGitHubReposByConnection` | `:many` | List all repos for a connection |
| | `UpsertGitHubRepo` | `:one` | Insert or update repo (ON CONFLICT by connection_id + repo_id) |
| | `DeleteGitHubRepo` | `:exec` | Delete a repo record |
| | `ListEnabledGitHubReposByApp` | `:many` | JOIN connections to find enabled repos for an app (used by agent tool) |

### API Endpoints

| Method | Path | Handler |
|--------|------|---------|
| `GET` | `/api/connections/{id}/github/repos` | `ListGitHubRepos` — Fetches repos from GitHub API, merges with DB enabled state |
| `PUT` | `/api/connections/{id}/github/repos` | `UpdateGitHubRepos` — Upserts repo enabled/disabled state |

### Repo Listing Logic

- Paginates through GitHub's `GET /installation/repositories` endpoint (100/page, max 50 pages / 5,000 repos)
- Loads existing `github_repos` rows from DB using user-scoped queries
- Merges: if a repo exists in DB, uses its `enabled` state; otherwise defaults to `false`
- Returns combined list so the frontend shows all accessible repos with toggle state

---

## Phase 4: GitHub Connector (QueryConnector)

### Modified Files

| File | Change |
|------|--------|
| `backend/internal/connectors/codebase/github.go` | Full rewrite — implements `QueryConnector` interface with GitHub API calls |
| `backend/internal/connectors/codebase/github_test.go` | Updated test to match new constructor signature |

### Connector Architecture

```go
type GitHub struct {
    ghClient       *github.Client  // from internal/github
    httpClient     *http.Client    // dedicated client with 10s timeout
    installationID int64
    repos          []string        // enabled repo full names
    token          string          // cached installation token
}
```

### Supported Actions

| Action | GitHub API | Safety Guardrails |
|--------|-----------|-------------------|
| `search_code` | `GET /search/code?q={query}+repo:{repo}` | Capped at 20 results |
| `read_file` | `GET /repos/{owner}/{repo}/contents/{path}?ref={ref}` | Skip files >1MB, truncate at 50KB, detect binary via encoding field |
| `list_tree` | `GET /repos/{owner}/{repo}/git/trees/{ref}?recursive=1` | Capped at 5,000 entries, 10s timeout |

### Safety Guardrails

- **File size**: Files >1MB are skipped with `[Binary or large file — skipped]`
- **Content truncation**: Files >50KB are truncated with `[TRUNCATED — file is N bytes]`
- **Binary detection**: Non-base64 encoding → `[Binary file — skipped]`
- **HTTP timeout**: 10-second timeout on all GitHub API calls
- **Response body limiting**: `io.LimitReader` on all responses (5MB cap)

---

## Phase 5: Agent Tool Integration

### Modified Files

| File | Change |
|------|--------|
| `backend/internal/agent/tools.go` | Added `search_codebase` to `ToolRegistry()` and `Dispatch()`. **Breaking change**: `Dispatch` signature now includes `appID uuid.UUID` |
| `backend/internal/agent/tools_codebase.go` | Full rewrite — implements `toolSearchCodebase()` |
| `backend/internal/agent/agent.go` | Added `githubClient *github.Client` field; updated `New()` to accept `*github.Client` |
| `backend/internal/agent/loop.go` | `RunConversation` now accepts `appID uuid.UUID`; `RunLoop` passes `uuid.Nil`; monitoring passes `appConfig.AppID` |
| `backend/internal/agent/prompt.go` | Updated both system prompts to mention `search_codebase` tool |
| `backend/internal/api/handlers/chat.go` | Reads `app_id` from WebSocket query param; passes to `RunConversation` |
| `backend/internal/agent/tools_test.go` | Updated `Dispatch` calls to include `appID` parameter |

### Dispatch Signature Change

```go
// Before
func (a *Agent) Dispatch(ctx, userID, name, input) (string, error)

// After
func (a *Agent) Dispatch(ctx, userID, appID, name, input) (string, error)
```

This is the key architectural change — it enables app-scoped tool access. The `appID` flows from:
- **Interactive mode**: WebSocket `?app_id=` query param → `HandleChat` → `RunConversation` → `Dispatch`
- **Monitoring mode**: `appConfig.AppID` → `RunMonitoring` → `Dispatch`

### Tool Definition

```json
{
  "name": "search_codebase",
  "description": "Search or read code in connected GitHub repositories...",
  "input_schema": {
    "properties": {
      "action": {"type": "string", "enum": ["search_code", "read_file", "list_tree"]},
      "query": {"type": "string"},
      "path": {"type": "string"},
      "repo": {"type": "string"},
      "ref": {"type": "string"}
    },
    "required": ["action"]
  }
}
```

### toolSearchCodebase Flow

1. Check `githubClient != nil` → clear error if not configured
2. Check `appID != uuid.Nil` → error if no app context
3. `ListEnabledGitHubReposByApp(appID)` → get enabled repos + connection config
4. Create `codebase.GitHub` connector with installation ID and repo list
5. `Connect()` to obtain installation token
6. Marshal input as action JSON → `Query(ctx, actionJSON)`
7. Return JSON result (follows existing error-as-tool-result pattern)

---

## Phase 6: Frontend

### New Files

| File | Purpose |
|------|---------|
| `frontend/src/types/github.ts` | `GitHubRepo` interface |
| `frontend/src/api/github.ts` | API client: `getGitHubInstallURL`, `listGitHubRepos`, `updateGitHubRepos` |
| `frontend/src/components/connections/GitHubRepoSelector.vue` | Repo list with toggle checkboxes, save/cancel buttons |

### Modified Files

| File | Change |
|------|--------|
| `frontend/src/components/connections/ConnectionForm.vue` | Replaced GitHub PAT config fields (owner/repo/token) with "Install GitHub App" button. For `type="github"`, shows install button instead of form fields + submit |
| `frontend/src/components/connections/ConnectionCard.vue` | Added "Repos" button for GitHub connections (triggers `manage-repos` event) |
| `frontend/src/components/connections/ConnectionList.vue` | Added `manage-repos` event passthrough |
| `frontend/src/pages/ConnectionsPage.vue` | Handles `?github=installed` query param (success banner + auto-opens repo selector); added repo selector integration |
| `frontend/src/composables/useWebSocket.ts` | Added `appId` to `WebSocketOptions`; passes as `?app_id=` query param |
| `frontend/src/composables/useAgent.ts` | Passes `currentAppId` from app store to WebSocket connection |

### User Flow

1. User clicks "+ New Connection" → selects "GitHub" type → enters a name
2. Instead of config fields, sees **"Install GitHub App"** button
3. Button calls `GET /api/github/install?app_id=X` → gets install URL → `window.location.href = url`
4. User installs the app on GitHub, selects repos
5. GitHub redirects to `/api/github/callback` → backend creates connection → redirects to `/connections?github=installed`
6. Frontend detects query param, shows success banner, auto-opens **GitHubRepoSelector**
7. User toggles repos, clicks "Save Selection"

### GitHubRepoSelector Component

- Fetches repos from `GET /api/connections/{id}/github/repos`
- Shows repo list with custom checkbox toggles and branch name badges
- Scrollable list (max 320px) for large installations
- "Save Selection" calls `PUT /api/connections/{id}/github/repos`
- Loading, error, and empty states handled
- Accessible from ConnectionCard via "Repos" button

---

## Phase 7: Hardening

### Implemented Guardrails

| Area | Guardrail |
|------|-----------|
| **Token cache** | In-memory with `sync.RWMutex`, refresh 5min before expiry |
| **HTTP timeouts** | 10-second timeout on all GitHub API calls (dedicated `http.Client` in both client and connector) |
| **Context propagation** | `APIRequest()` accepts `context.Context` — callers can cancel outbound requests |
| **Response limits** | `io.LimitReader` (1MB for tokens, 5MB for API responses) |
| **File safety** | Skip >1MB, truncate >50KB, detect binary via encoding |
| **Search results** | Capped at 20 items per query |
| **Tree listing** | Capped at 5,000 entries |
| **Pagination bound** | Repo listing capped at 50 pages (5,000 repos max) to prevent unbounded loops |
| **State JWT** | Signed RS256, 15-minute expiry, nonce for replay prevention |
| **Nil-safe** | GitHub client is nil when not configured; all callers check before use |
| **Error as tool result** | GitHub errors returned as tool results to Claude (not thrown), following existing pattern |
| **Error propagation** | DB errors in callback handler returned to user (not swallowed) to prevent duplicate connections |
| **Connection test** | GitHub connections test via installation token validity |
| **Upsert semantics** | Re-installs update existing connection rather than creating duplicates |
| **Row Level Security** | `github_repos` table has RLS policy scoped via parent connection's `user_id` |
| **User-scoped queries** | All handler DB operations use user-scoped `queries` (via `UserQueries()` transaction) |
| **Configurable slug** | GitHub App slug is configurable via `GITHUB_APP_SLUG` env var (not hardcoded) |

### Test Updates

- `backend/internal/connectors/codebase/github_test.go` — Updated to match new constructor (nil client returns error)
- `backend/internal/agent/tools_test.go` — Updated `Dispatch` calls to include `appID` parameter

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GITHUB_APP_ID` | No | — | GitHub App ID (numeric) |
| `GITHUB_PRIVATE_KEY` | No | — | RSA private key (PEM string or file path) |
| `GITHUB_CLIENT_ID` | No | — | GitHub App client ID |
| `GITHUB_APP_SLUG` | No | `heimdall-agent` | GitHub App URL slug (used in install redirect URL) |
| `GITHUB_WEBHOOK_SECRET` | No | — | Webhook secret (reserved for future use) |

All are optional — if `GITHUB_APP_ID` is empty, the GitHub integration is disabled entirely.

---

## Post-Review Hardening (Round 1)

Seven fixes applied during first code review (assessed 7.5 → 8.5/10):

| # | Severity | Fix | Files |
|---|----------|-----|-------|
| 1 | CRITICAL | Added RLS policy on `github_repos` — scoped via parent connection's `user_id` | `020_github_repos.up.sql` |
| 2 | HIGH | Made GitHub App slug configurable via `GITHUB_APP_SLUG` env var (was hardcoded `heimdall-agent`) | `config.go`, `handlers/github.go` |
| 3 | HIGH | `ListConnectionsByApp` error now returns to user instead of silently continuing (prevented duplicate connection creation) | `handlers/github.go` |
| 4 | HIGH | Switched `ListGitHubReposByConnection` and `UpsertGitHubRepo` to use user-scoped `queries` (was bypassing RLS transaction via `s.Queries`) | `handlers/github.go` |
| 5 | MEDIUM | Replaced `http.DefaultClient` with dedicated `http.Client{Timeout: 10s}` on connector struct | `codebase/github.go` |
| 6 | MEDIUM | `APIRequest()` now accepts `context.Context` — enables caller-controlled cancellation; all callers updated | `github/client.go`, callers |
| 7 | MEDIUM | Added pagination safety bound (50 pages / 5,000 repos max) to `ListGitHubRepos` | `handlers/github.go` |

## Post-Review Hardening (Round 2)

Eight fixes applied during production-readiness review (assessed 8.5 → 9/10):

| # | Severity | Fix | Files |
|---|----------|-----|-------|
| 1 | HIGH | Moved `GET /api/github/callback` out of JWT-protected route group — browser redirect from GitHub has no session JWT; callback self-authenticates via state JWT | `router.go` |
| 2 | MEDIUM | Fixed search query double-encoding — replaced manual `url.QueryEscape` + string concatenation with `url.Values` for proper query string construction | `codebase/github.go` |
| 3 | MEDIUM | Added `context.Context` to `InstallationTokenFor()` and `GetInstallation()` — enables caller-controlled cancellation; updated all callers | `github/client.go`, `handlers/github.go`, `handlers/connections.go`, `codebase/github.go` |
| 4 | MEDIUM | Documented single-installation limitation in `toolSearchCodebase` — repos from multiple GitHub installations only use the first installation's token | `agent/tools_codebase.go` |
| 5 | MEDIUM | `ListGitHubReposByConnection` DB error now returns 500 instead of silently continuing with empty enabled state | `handlers/github.go` |
| 6 | LOW | Added `io.LimitReader(r.Body, 1MB)` to `UpdateGitHubRepos` request body parsing | `handlers/github.go` |
| 7 | LOW | Unified connector HTTP path — removed standalone `httpClient` from connector struct, all API calls now route through `ghClient.APIRequest()` for consistent timeout/header behavior | `codebase/github.go` |
| 8 | LOW | Fixed `url.PathEscape` on nested file paths — was escaping slashes, breaking `src/main.go` → `src%2Fmain.go`. Now escapes individual path segments | `codebase/github.go` |

---

## Verification

```bash
# Backend
cd backend && go build ./...     # Compiles clean
cd backend && go vet ./...       # No issues
cd backend && go test ./...      # Run all tests

# Frontend
cd frontend && npx vue-tsc -b --noEmit  # Type-checks clean
cd frontend && npx eslint src/           # Only pre-existing lint warnings

# End-to-end
# 1. Set GITHUB_APP_ID, GITHUB_PRIVATE_KEY, GITHUB_CLIENT_ID env vars
# 2. make dev → confirm "GitHub App configured" in logs
# 3. Connections → + New Connection → GitHub → Install GitHub App
# 4. Install on test org → verify redirect to /connections?github=installed
# 5. Select repos → Save → Agent Chat → "What does main.go do?"
```
