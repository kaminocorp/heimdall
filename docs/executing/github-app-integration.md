# GitHub App Integration — Implementation Plan

Reference: [Vision](../vision.md) · [Incomplete Features](./incomplete-features.md) (item 3) · [Changelog](../changelog.md)

## Context

Heimdall's vision includes codebase search during investigations — "Agent queries a GitHub repo to understand relevant code." Currently, a GitHub connector stub exists (`backend/internal/connectors/codebase/github.go`) with no implementation, and the frontend has placeholder PAT config fields (owner, repo, token).

This plan replaces the PAT approach with a **GitHub App** — the standard, secure way for SaaS products to access user repositories. Users install the GitHub App on their GitHub org, granting Heimdall read access to selected repos. The agent can then search and read code during monitoring and interactive investigations.

### Why GitHub App over PATs?

- **PATs** are user-scoped — if the user leaves the org, access breaks. Manual rotation required.
- **OAuth Apps** authenticate as the user, inheriting all their permissions — overly broad.
- **GitHub Apps** install on an org with fine-grained permissions (Contents:read, Metadata:read only), generate short-lived installation tokens (1hr), and the org admin controls which repos are accessible. Revoking is as simple as uninstalling the app.

---

## Phase 1: GitHub App Client & Token Management

**Goal**: Backend infrastructure for GitHub App JWT generation and installation token caching.

### New files

| File | Purpose |
|------|---------|
| `backend/internal/github/client.go` | Core client: App ID, parsed RSA private key, JWT generation (RS256, 10min expiry), installation token fetching with in-memory cache |
| `backend/internal/github/client_test.go` | Unit tests: JWT claims/signing, token cache expiry, mock HTTP server for installation tokens |

### Modified files

| File | Change |
|------|--------|
| `backend/internal/config/config.go` | Add `GitHubAppID`, `GitHubPrivateKey`, `GitHubClientID`, `GitHubWebhookSecret` (all optional env vars) |
| `backend/.env.example` | Document `GITHUB_APP_ID`, `GITHUB_PRIVATE_KEY`, `GITHUB_CLIENT_ID`, `GITHUB_WEBHOOK_SECRET` |

### Implementation details

- **JWT generation**: Standard GitHub App JWT — `iss` = App ID, `iat` = now-60s, `exp` = now+10min, signed with RS256 using the private key. Use `github.com/golang-jwt/jwt/v5` (already in go.mod).
- **Installation tokens**: `POST /app/installations/{id}/access_tokens` with App JWT in Authorization header. Response includes `token` and `expires_at`. Cache in `sync.Map` keyed by installation ID, refresh 5 minutes before expiry.
- **Private key loading**: Support both raw PEM string in env var and file path (detect by checking for `-----BEGIN` prefix vs file path).
- **Nil-safe**: If `GitHubAppID` is empty, `NewClient()` returns nil. All consumers check for nil before use (same pattern as `notifier`).

### Verification

```bash
cd backend && go test ./internal/github/...
```

---

## Phase 2: Installation Flow (Backend)

**Goal**: Endpoints that redirect users to GitHub to install the app and handle the callback.

### New files

| File | Purpose |
|------|---------|
| `backend/internal/api/handlers/github.go` | `InstallGitHub` and `GitHubCallback` handler methods on `*Server` |

### Modified files

| File | Change |
|------|--------|
| `backend/internal/api/handlers/server.go` | Add `GitHub *github.Client` field to `Server` struct; update `NewServer` signature |
| `backend/internal/api/router.go` | Register `GET /api/github/install` and `GET /api/github/callback` in the protected group |
| `backend/cmd/heimdall/main.go` | Create `github.Client` if configured, pass to `NewServer` |

### Endpoints

**`GET /api/github/install?app_id={appId}`** (protected)
- Generates a signed state JWT containing `{ user_id, app_id, nonce, exp }` using the GitHub App private key
- Returns `{ url: "https://github.com/apps/{slug}/installations/new?state={state}" }`

**`GET /api/github/callback?installation_id={id}&setup_action={action}&state={state}`** (protected)
- Validates state JWT (expiry, signature)
- Calls `GET /app/installations/{id}` with App JWT to verify the installation exists and get `account.login` / `account.type`
- Authorises the `app_id` from state belongs to the user's org via `authorizeApp`
- Creates a connection record: `type="github"`, `direction="two_way"`, `config={"installation_id": N, "account_login": "...", "account_type": "Organization|User"}`, `status="active"`
- If a connection with the same `installation_id` already exists for this app, updates it (upsert)
- Redirects browser to `/connections?github=installed`

### Edge cases

- **Re-install / modify permissions**: GitHub triggers callback with `setup_action=update` → update existing connection config
- **Multiple GitHub orgs**: A user could install the app on multiple GitHub orgs. Each creates a separate connection. The agent searches across all connected repos.

### Verification

```bash
cd backend && go test ./internal/api/handlers/ -run TestGitHub
# Manual: start server, visit GET /api/github/install?app_id=X, verify redirect URL
```

---

## Phase 3: Repository Management

**Goal**: Let users see and select which repos from their GitHub installation Heimdall should use for code search.

### Database migration

**`backend/migrations/020_github_repos.up.sql`**:
```sql
CREATE TABLE github_repos (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id   UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    repo_full_name  TEXT NOT NULL,          -- e.g. "my-org/my-app"
    repo_id         BIGINT NOT NULL,        -- GitHub's numeric repo ID
    default_branch  TEXT NOT NULL DEFAULT 'main',
    enabled         BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_github_repos_conn_repo ON github_repos(connection_id, repo_id);
CREATE INDEX idx_github_repos_connection_id ON github_repos(connection_id);
```

**`backend/migrations/020_github_repos.down.sql`**: `DROP TABLE github_repos;`

### New sqlc queries

**`backend/internal/db/queries/github_repos.sql`**:

| Query | Type | Purpose |
|-------|------|---------|
| `ListGitHubReposByConnection` | `:many` | SELECT by connection_id, ordered by repo_full_name |
| `UpsertGitHubRepo` | `:one` | INSERT ... ON CONFLICT(connection_id, repo_id) DO UPDATE |
| `DeleteGitHubRepo` | `:exec` | DELETE by id |
| `ListEnabledGitHubReposByApp` | `:many` | JOIN connections WHERE app_id=$1 AND enabled=true |

After writing the queries, run `make sqlc-generate` to regenerate Go code.

### API endpoints (added to `github.go`)

**`GET /api/connections/{id}/github/repos`**
- Fetches repos accessible to this installation from GitHub API (`GET /installation/repositories` using the installation token)
- Merges with `github_repos` table to show which are enabled/disabled
- Paginates through all GitHub pages (max 100/page)
- Returns combined list: `[{ repo_id, repo_full_name, default_branch, enabled }]`

**`PUT /api/connections/{id}/github/repos`**
- Accepts `[{ repo_id, repo_full_name, default_branch, enabled }]`
- Upserts each entry into `github_repos` table

### Router changes

Add to `router.go`:
```go
r.Get("/connections/{id}/github/repos", s.ListGitHubRepos)
r.Put("/connections/{id}/github/repos", s.UpdateGitHubRepos)
```

### Connection test update

Update `TestConnection` in `connections.go`: for `type="github"`, use the installation token to call GitHub's `GET /rate_limit` endpoint as a health check (verifies token validity and API access).

### Verification

```bash
make sqlc-generate
cd backend && go test ./internal/api/handlers/ -run TestGitHubRepo
# Manual: create GitHub connection, call GET /api/connections/{id}/github/repos, verify repo list
```

---

## Phase 4: GitHub Connector (QueryConnector)

**Goal**: Implement the `QueryConnector` interface for code search and file reading.

### Modified files

| File | Change |
|------|--------|
| `backend/internal/connectors/codebase/github.go` | Full rewrite — implement QueryConnector with GitHub API calls |
| `backend/internal/connectors/codebase/github_test.go` | Tests with mock HTTP server |

### Connector struct

```go
type GitHub struct {
    ghClient       *githubpkg.Client  // from internal/github
    installationID int64
    repos          []string           // enabled repo full names (e.g. "org/repo")
    token          string             // cached installation token
}
```

### Constructor

```go
func New(configJSON json.RawMessage, ghClient *githubpkg.Client, repos []string) (*GitHub, error)
```

Parses `installation_id` from config JSON. Follows the pattern of `database.New(configJSON)`.

### Query method

The `query` string is a JSON-encoded action descriptor:

```json
{"action": "search_code", "query": "handlePayment", "repo": "org/repo"}
{"action": "read_file", "path": "src/main.go", "repo": "org/repo", "ref": "main"}
{"action": "list_tree", "path": "src/", "repo": "org/repo", "ref": "main"}
```

### GitHub API calls

| Action | GitHub API | Notes |
|--------|-----------|-------|
| `search_code` | `GET /search/code?q={query}+repo:{repo}` | Cap at 20 results. If no repo specified, search all enabled repos |
| `read_file` | `GET /repos/{owner}/{repo}/contents/{path}?ref={ref}` | Base64 decode. Skip binary files (`encoding: null/none`). Truncate content at 50KB |
| `list_tree` | `GET /repos/{owner}/{repo}/git/trees/{sha}?recursive=1` | Resolve branch → commit → tree SHA. Cap at 5000 entries |

### Safety guardrails

- **Rate limit tracking**: Parse `X-RateLimit-Remaining` after each call. Warn if <100, error if 0.
- **Search rate limit**: GitHub Search API allows 30 req/min. Implement simple sliding window.
- **File size**: Skip files > 1MB. Truncate content > 50KB with `[TRUNCATED — file is {size} bytes]`.
- **Binary detection**: Check `encoding` field from Contents API. Return `[Binary file — skipped]` for non-text files.
- **HTTP timeout**: 10-second timeout on all GitHub API calls.

### Verification

```bash
cd backend && go test ./internal/connectors/codebase/...
```

---

## Phase 5: Agent Tool Integration

**Goal**: Wire `search_codebase` into the agent's tool registry and dispatch so both monitoring and interactive modes can search code.

### Modified files

| File | Change |
|------|--------|
| `backend/internal/agent/tools.go` | Add `search_codebase` to `ToolRegistry()` and `Dispatch()`. Add `appID uuid.UUID` parameter to `Dispatch` signature |
| `backend/internal/agent/tools_codebase.go` | Implement `toolSearchCodebase()` — find GitHub connections for app, create connector, execute query |
| `backend/internal/agent/agent.go` | Add `githubClient *github.Client` field; update `New()` signature |
| `backend/internal/agent/loop.go` | Add `appID` param to `RunConversation` and `RunLoop`. Pass `appID` through `Dispatch` calls in both `RunConversation` and `RunMonitoring` |
| `backend/internal/agent/prompt.go` | Update system prompts to mention `search_codebase` tool availability |
| `backend/internal/api/handlers/chat.go` | Read `app_id` from WebSocket query param `?app_id=`; pass to `RunConversation` |
| `backend/cmd/heimdall/main.go` | Pass GitHub client to `agent.New()` |

### Dispatch signature change

The key architectural change: add `appID` to the Dispatch signature so tools can find app-scoped connections.

```go
// Before
func (a *Agent) Dispatch(ctx context.Context, userID uuid.UUID, name string, input map[string]any) (string, error)

// After
func (a *Agent) Dispatch(ctx context.Context, userID uuid.UUID, appID uuid.UUID, name string, input map[string]any) (string, error)
```

- `RunConversation` gains an `appID uuid.UUID` parameter, passed from the chat handler via `?app_id=` WebSocket query param
- `RunMonitoring` already has `appConfig.AppID` — passes it to Dispatch
- `search_logs` and `query_database` ignore the `appID` param (backward compatible)

### Tool definition

```go
{OfTool: &anthropic.ToolParam{
    Name:        "search_codebase",
    Description: anthropic.String("Search or read code in connected GitHub repositories. Use to understand application code, find error sources, or investigate configuration."),
    InputSchema: anthropic.ToolInputSchemaParam{
        Properties: map[string]any{
            "action": map[string]any{
                "type": "string",
                "enum": []string{"search_code", "read_file", "list_tree"},
                "description": "The operation to perform",
            },
            "query": map[string]any{
                "type":        "string",
                "description": "Search query (required for search_code)",
            },
            "path": map[string]any{
                "type":        "string",
                "description": "File or directory path (required for read_file/list_tree)",
            },
            "repo": map[string]any{
                "type":        "string",
                "description": "Repository in owner/repo format. If omitted, searches all connected repos.",
            },
            "ref": map[string]any{
                "type":        "string",
                "description": "Git ref (branch/tag/SHA). Defaults to the repo's default branch.",
            },
        },
        Required: []string{"action"},
    },
}}
```

### toolSearchCodebase implementation

1. Check `a.githubClient != nil` — return clear error if GitHub not configured
2. Query `ListEnabledGitHubReposByApp(ctx, appID)` — get enabled repos and their connection info
3. Get `installation_id` from connection config
4. Create `codebase.GitHub` connector with installation ID, GitHub client, and repo list
5. Call `Connect(ctx)` to obtain installation token
6. Marshal action from input params → call `Query(ctx, actionJSON)`
7. Return result as string (JSON-encoded search results or file content)
8. Follow existing error-as-tool-result pattern from `tools_db.go`

### System prompt update

Update both `systemPrompt` and `monitoringSystemPrompt` in `prompt.go`:
```
You have access to tools that let you search logs, query connected databases, and search/read code in connected GitHub repositories.
```

### Verification

```bash
cd backend && go test ./internal/agent/ -run TestSearchCodebase
cd backend && go vet ./...
```

---

## Phase 6: Frontend

**Goal**: Replace PAT config fields with GitHub App install flow and add a repo selector component.

### New files

| File | Purpose |
|------|---------|
| `frontend/src/api/github.ts` | API client: `getGitHubInstallURL(appId)`, `listGitHubRepos(connectionId)`, `updateGitHubRepos(connectionId, repos)` |
| `frontend/src/components/connections/GitHubRepoSelector.vue` | Repo list with toggle switches, "Save" button |

### Modified files

| File | Change |
|------|--------|
| `frontend/src/components/connections/ConnectionForm.vue` | Replace `github` config fields (owner/repo/token) with "Install GitHub App" button. When `type === 'github'`, show install button instead of config form fields |
| `frontend/src/pages/ConnectionsPage.vue` | Handle `?github=installed` query param: show success toast, open repo selector for new connection. Add "Manage Repos" action to GitHub connection cards |
| `frontend/src/types/connection.ts` | Add `GitHubRepo` interface |
| `frontend/src/composables/useWebSocket.ts` (or `useAgent.ts`) | Pass `app_id` query param in WebSocket URL alongside `token` |

### User flow

1. User clicks "+ New Connection" → selects "GitHub" type → enters a name
2. Instead of config fields, sees **"Install GitHub App"** button
3. Button calls `GET /api/github/install?app_id=X` → gets install URL → `window.location.href = url` (full-page redirect to GitHub)
4. User installs the app on GitHub, selects which repos to grant access to
5. GitHub redirects to `/api/github/callback?installation_id=X&state=Y`
6. Backend creates the connection, redirects to `/connections?github=installed`
7. Frontend detects query param, shows success toast, opens **GitHubRepoSelector** for the new connection
8. User toggles which repos Heimdall can search, clicks "Save"

### GitHubRepoSelector component

- Fetches `GET /api/connections/{id}/github/repos`
- Shows repo list with checkboxes/toggles, grouped by org
- "Save" calls `PUT /api/connections/{id}/github/repos`
- Loading and error states
- Accessible from ConnectionCard for existing GitHub connections via "Manage Repos" button

### WebSocket app_id

The chat handler needs `app_id` to pass to `RunConversation`. Add `?app_id={currentAppId}` to the WebSocket URL. The `useWebSocket` composable already constructs the URL with `?token=` — append `&app_id=`.

### Verification

```bash
cd frontend && npm run lint
make build-frontend   # vue-tsc type check
# Manual: full flow — install GitHub App → select repos → chat with agent about code
```

---

## Phase 7: Hardening & Tests

**Goal**: Production readiness — error handling, rate limiting, comprehensive tests.

### Error handling

| GitHub Response | Handling |
|----------------|----------|
| 401 Unauthorized | Token expired mid-request → retry once with a fresh installation token |
| 403 Rate Limit | Return clear error with reset timestamp to agent |
| 404 Not Found | Return as tool result (not error) so Claude can adapt — "File not found" |
| 422 Unprocessable | Invalid search query → return as tool error |

### Rate limiting

- Track `X-RateLimit-Remaining` and `X-RateLimit-Reset` headers on every GitHub API call
- **REST API**: 5000 req/hr per installation — warn when low, error when exhausted
- **Search API**: 30 req/min — implement simple sliding window tracker

### Large repo safety

- Tree API: cap at 5000 entries, 10-second timeout
- File content: truncate at 50KB, skip files > 1MB
- Search results: cap at 20 items per query
- Binary files: detect via `encoding` field, return `[Binary file — skipped]`

### Test matrix

| File | Tests |
|------|-------|
| `backend/internal/github/client_test.go` | JWT generation (claims, signing, expiry), token cache (hit, miss, refresh), mock HTTP for installation tokens |
| `backend/internal/connectors/codebase/github_test.go` | Mock HTTP server: search_code, read_file, list_tree, binary detection, content truncation, rate limit error |
| `backend/internal/agent/tools_codebase_test.go` | Tool dispatch with mocked connector, user/app scoping, nil GitHub client handling |
| `backend/internal/api/handlers/github_test.go` | Install URL generation, callback processing (happy path, invalid state, duplicate install), repo listing |

### Verification

```bash
cd backend && go test ./...
cd frontend && npm run test
make lint
make build-frontend
make build-backend
```

---

## Dependency Graph

```
Phase 1 (client + tokens)
   ↓
Phase 2 (install flow) ───── depends on Phase 1
   ↓
Phase 3 (repo management) ── depends on Phase 2
   ↓
Phase 4 (connector) ──────── depends on Phase 1 + 3
   ↓
Phase 5 (agent tool) ─────── depends on Phase 4
   ↓
Phase 6 (frontend) ────────── depends on Phase 2 + 3 (can start during Phase 4/5)
   ↓
Phase 7 (hardening) ───────── depends on all above
```

## End-to-End Verification

1. Set `GITHUB_APP_ID`, `GITHUB_PRIVATE_KEY`, `GITHUB_CLIENT_ID` env vars (create a test GitHub App first)
2. Run `make dev` — confirm startup logs show "GitHub App configured"
3. Go to Connections → Add GitHub → click "Install GitHub App"
4. Install on a test GitHub org with a test repo
5. Verify redirect back to `/connections` with new GitHub connection (status: active)
6. Open repo selector → verify repos appear → toggle and save
7. Open Agent Chat → ask "What does the main function do in [repo]?"
8. Verify the agent calls `search_codebase` tool and returns code content
9. Check agent log for `tool_call` and `tool_result` entries for `search_codebase`
