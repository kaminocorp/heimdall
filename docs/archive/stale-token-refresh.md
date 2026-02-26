# Stale Token on Refresh

## Problem

After a backend redeploy (or if a Supabase JWT expires naturally), hitting browser refresh produces a 400 error. Closing the tab and navigating to heimdallwatch.com fresh works fine.

### Root Cause

`auth.init()` calls `supabase.auth.getSession()`, which returns whatever session is cached in **localStorage** — including the access token. If that token has expired (common during a deploy cycle where 5–10 minutes pass), the app sets `initialized = true` and all downstream consumers immediately fire requests with the stale JWT:

1. **Axios interceptor** (`api/client.ts`) reads `auth.token` and attaches it as `Authorization: Bearer <stale>`
2. **WebSocket** (`useWebSocket.ts`) passes it as a query param: `/ws/chat?token=<stale>`
3. Backend rejects both with 400/401

The `onAuthStateChange` listener *eventually* triggers a background refresh, but by then the initial page-load requests have already failed and the user sees errors.

### Why close + reopen works

When the user closes the tab and opens a new one, Supabase's client constructor runs its full cold-start initialization — which includes an automatic token refresh before returning the session. On a warm refresh, `getSession()` skips this and returns the cached (potentially expired) token directly.

## Fix

One file: `frontend/src/stores/auth.ts`

In `init()`, after `getSession()` finds a cached session, call `refreshSession()` before setting `initialized = true`. This guarantees a fresh token before any API call or WebSocket connection fires.

```ts
async function init() {
  const { data } = await supabase.auth.getSession()

  // If a cached session exists, force a token refresh so we never send
  // a stale JWT to the backend (e.g. after a redeploy or token expiry).
  if (data.session) {
    const { data: refreshed } = await supabase.auth.refreshSession()
    session.value = refreshed.session
    user.value = refreshed.session?.user ?? null
  } else {
    session.value = null
    user.value = null
  }

  supabase.auth.onAuthStateChange((_event, newSession) => {
    session.value = newSession
    user.value = newSession?.user ?? null
  })

  initialized.value = true
}
```

If the refresh itself fails (e.g. refresh token revoked), `refreshed.session` will be `null`, which correctly logs the user out and the router guard redirects to `/login`.

### Why this is safe

- `refreshSession()` is a single HTTP call to Supabase's `/token?grant_type=refresh_token` — adds ~50–100ms to init, only when a session exists.
- The `onAuthStateChange` listener still handles mid-session refreshes (Supabase auto-refreshes ~60s before expiry).
- No other files need to change — the Axios interceptor and WebSocket composable both read `auth.token` reactively, so they automatically pick up the fresh value.

## Files Changed

1 file: `frontend/src/stores/auth.ts`
