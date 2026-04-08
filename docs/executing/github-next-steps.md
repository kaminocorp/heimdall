# GitHub App Integration — Setup Steps

The codebase has the full GitHub App integration implemented (v0.16.0). It's not working because the 4 required environment variables are missing from Fly.io. This guide walks through creating the GitHub App and wiring it up.

**Deployed backend:** `heimdall-backend.fly.dev`
**Callback URL used by code:** `GET /api/github/callback`

---

## Step 1 — Create the GitHub App

1. Go to **GitHub → Settings → Developer settings → GitHub Apps → New GitHub App**
2. Fill in the form:

| Field | Value |
|-------|-------|
| **GitHub App name** | `heimdall-agent` (must match `GITHUB_APP_SLUG`, which defaults to this) |
| **Homepage URL** | `https://heimdall-backend.fly.dev` (or your frontend domain if separate) |
| **Callback URL** | `https://heimdall-backend.fly.dev/api/github/callback` |
| **Setup URL** | _(leave blank)_ |
| **Webhook** | **Inactive** (webhook handler not yet implemented — `GITHUB_WEBHOOK_SECRET` is defined in config but unused) |
| **Permissions → Repository** | `Contents: Read-only`, `Metadata: Read-only` (needed for `search_code`, `read_file`, `list_tree`) |
| **Where can this app be installed?** | `Only on this account` (change to `Any account` later if needed) |

3. Click **Create GitHub App**

---

## Step 2 — Collect the credentials

After creating the app, you'll be on its settings page. Collect three values:

### 2a. App ID
- Shown at the top of the **General** page as **App ID** (a numeric value like `123456`)
- This becomes `GITHUB_APP_ID`

### 2b. Client ID
- Shown on the same page as **Client ID** (a string like `Iv1.abc123...`)
- This becomes `GITHUB_CLIENT_ID`

### 2c. Private Key
- Scroll down to **Private keys** → click **Generate a private key**
- A `.pem` file downloads
- Open it and copy the full contents (including `-----BEGIN RSA PRIVATE KEY-----` and `-----END RSA PRIVATE KEY-----`)
- This becomes `GITHUB_PRIVATE_KEY`

---

## Step 3 — Set secrets on Fly.io

Run each command, substituting your actual values:

```bash
# App ID (numeric)
fly secrets set GITHUB_APP_ID="123456" --app heimdall-backend

# Client ID
fly secrets set GITHUB_CLIENT_ID="Iv1.abc123def456" --app heimdall-backend

# Private key — use quotes to preserve newlines
fly secrets set GITHUB_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
(full PEM contents here)
...
-----END RSA PRIVATE KEY-----" --app heimdall-backend
```

> **Note:** `GITHUB_APP_SLUG` defaults to `heimdall-agent` in config. Only set it if you chose a different app name in Step 1.

---

## Step 4 — Verify deployment

The secrets deploy automatically. Confirm the app restarted:

```bash
fly status --app heimdall-backend
```

Check logs to verify the GitHub client initialised (it logs nothing on success, but will log an error if the PEM is malformed):

```bash
fly logs --app heimdall-backend | head -30
```

---

## Step 5 — Test the install flow

1. Open Heimdall in the browser
2. Go to **Connections** → add a new connection of type **GitHub**
3. Click **Install GitHub App** — this hits `GET /api/github/install`, which generates a state JWT and redirects you to `https://github.com/apps/heimdall-agent/installations/new`
4. On GitHub, select the repositories you want Heimdall to access → **Install**
5. GitHub redirects back to `/api/github/callback` with `installation_id` and `state` params
6. The callback creates a `github` connection in the DB and redirects to `/connections?github=installed`
7. You should see a success banner and the new GitHub connection card

---

## Step 6 — Enable repositories

1. On the new GitHub connection card, click **Repos**
2. The repo selector fetches all repos accessible to the installation via GitHub API
3. Toggle on the repos you want the agent to search during investigations
4. Click **Save**

---

## Step 7 — Verify agent tool access

Open a chat with the agent and ask something like:

> "Search the codebase for how authentication middleware works"

The agent should use the `search_codebase` tool, which dispatches to the GitHub codebase connector (`search_code` / `read_file` / `list_tree` actions).

---

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| "GitHub App not configured" (503) | `GITHUB_APP_ID` not set or empty | Verify with `fly secrets list --app heimdall-backend` |
| "failed to verify GitHub installation" (502) | Private key doesn't match the app | Re-generate the PEM in GitHub App settings and re-set the secret |
| Callback redirects but no connection appears | State JWT expired (>15 min) or app_id mismatch | Try the install flow again; check Fly logs for the specific error |
| Repos endpoint returns empty list | Installation has no repo access | Go to GitHub → Settings → Applications → Configure the app → grant repo access |
| Agent can't use `search_codebase` | No repos enabled in the repo selector | Open the GitHub connection → Repos → toggle repos on |

---

## Current secrets on Fly (as of 2026-04-06)

```
ANTHROPIC_API_KEY  ✅ Set
DATABASE_URL       ✅ Set
SUPABASE_URL       ✅ Set
GITHUB_APP_ID      ❌ Missing
GITHUB_CLIENT_ID   ❌ Missing
GITHUB_PRIVATE_KEY ❌ Missing
```

After completing Steps 1–3, all six should be set.
