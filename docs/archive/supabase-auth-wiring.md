# Supabase Auth Wiring

Wire real Supabase authentication into the frontend, replacing the placeholder login with actual email/password auth via `@supabase/supabase-js`.

Reference: [Post-MVP Roadmap — Phase 7](./post-mvp-roadmap.md) · [Vision](../vision.md)

---

## Current State

- `LoginPage.vue` calls `auth.login('placeholder-token')` — no real auth
- `auth.ts` store holds a raw token string in `localStorage`
- `client.ts` attaches the token as `Authorization: Bearer <token>` on all API requests
- `router/index.ts` redirects unauthenticated users to `/login`
- Backend verifies JWTs via JWKS endpoint (`middleware/auth.go`) — this is already wired
- Backend has `GET /api/auth/me` that returns the user record for the authenticated JWT
- No Supabase JS dependency exists in the frontend

## What's Missing

1. **Supabase JS client** — no `@supabase/supabase-js` dependency
2. **Environment variables** — no `VITE_SUPABASE_URL` or `VITE_SUPABASE_ANON_KEY` (the publishable key)
3. **Real login flow** — no call to `supabase.auth.signInWithPassword()`
4. **Token lifecycle** — no token refresh, no session recovery on page reload, no proper logout
5. **Signup flow** — no user registration page or flow

---

## Implementation Plan

### Step 1 — Environment & Supabase Client

**Add dependency:**
```bash
cd frontend && npm install @supabase/supabase-js
```

**Add frontend env vars** (create `frontend/.env`):
```
VITE_SUPABASE_URL=https://qghlzoszcgetjrpvhkyw.supabase.co
VITE_SUPABASE_ANON_KEY=sb_publishable_yQDpg0jzm6H3Cs2qQ7RTqA_Pr6-IVAI
```

**Create Supabase client** (`frontend/src/lib/supabase.ts`):
```ts
import { createClient } from '@supabase/supabase-js'

export const supabase = createClient(
  import.meta.env.VITE_SUPABASE_URL,
  import.meta.env.VITE_SUPABASE_ANON_KEY,
)
```

**Files:** new `frontend/.env`, new `frontend/src/lib/supabase.ts`, `frontend/package.json`

---

### Step 2 — Rewrite Auth Store

Replace the simple token store with a Supabase session-aware store.

**`frontend/src/stores/auth.ts`** — rewrite:

- `init()` — call `supabase.auth.getSession()` to recover session on page load, then subscribe to `onAuthStateChange` for token refresh and logout events
- `login(email, password)` — call `supabase.auth.signInWithPassword({ email, password })`, extract the access token
- `signup(email, password)` — call `supabase.auth.signUp({ email, password })`
- `logout()` — call `supabase.auth.signOut()`, clear state
- `token` — computed from the current Supabase session (`session.access_token`), auto-updates on refresh
- `isAuthenticated` — derived from session presence
- `user` — expose `session.user` for display (email, id)

Key detail: Supabase JS handles token refresh automatically via `onAuthStateChange`. The store listens to this event and updates the reactive `token` ref, so the axios interceptor in `client.ts` always has a fresh token.

**Files:** `frontend/src/stores/auth.ts`

---

### Step 3 — Wire Login Page

**`frontend/src/pages/LoginPage.vue`** — rewrite:

- Call `auth.login(email, password)` (the new store method)
- Show loading state while request is in flight
- Display error message on failure (invalid credentials, network error)
- On success, redirect to `/` (dashboard)
- Add a link to signup (or a toggle between login/signup on the same page)

**Files:** `frontend/src/pages/LoginPage.vue`

---

### Step 4 — Init Auth on App Mount

The auth store's `init()` must run before the router guard checks `isAuthenticated`, otherwise a page reload always redirects to `/login` even if a valid Supabase session exists.

**`frontend/src/App.vue`** (or `main.ts`):

- Call `auth.init()` on app mount, before the first route resolves
- Show a loading spinner until session recovery completes (prevents flash of login page)

**`frontend/src/router/index.ts`**:

- Update the `beforeEach` guard to await auth initialization before redirecting

**Files:** `frontend/src/App.vue` or `frontend/src/main.ts`, `frontend/src/router/index.ts`

---

### Step 5 — Logout

- Add a logout button to the app shell / sidebar
- Calls `auth.logout()`, which calls `supabase.auth.signOut()`
- `onAuthStateChange` fires with `SIGNED_OUT`, store clears token
- Router guard redirects to `/login`

**Files:** whichever component renders the sidebar/nav

---

### Step 6 — Update .env.example

```
# Frontend (prefix with VITE_ for Vite exposure)
VITE_SUPABASE_URL=https://your-project.supabase.co
VITE_SUPABASE_ANON_KEY=sb_publishable_...
```

**Files:** `.env.example`

---

## Token Flow Summary

```
User submits email/password
  → supabase.auth.signInWithPassword()
  → Supabase returns session { access_token (JWT), refresh_token }
  → auth store saves session, exposes access_token as `token`
  → axios interceptor attaches `Authorization: Bearer <token>`
  → Heimdall backend validates JWT via JWKS endpoint
  → WebSocket connects with ?token=<access_token>

Token expires (~1 hour)
  → Supabase JS auto-refreshes via refresh_token
  → onAuthStateChange fires TOKEN_REFRESHED
  → auth store updates token ref
  → next API call uses new token automatically
```

---

## Verification

1. `npm run build` — confirm TypeScript compilation
2. Start backend + frontend, navigate to `/login`
3. Sign in with a Supabase user — verify redirect to dashboard
4. Check `GET /api/auth/me` returns the correct user
5. Open Agent Chat — verify WebSocket connects with real JWT
6. Wait for token expiry (or manually expire) — verify auto-refresh
7. Click logout — verify redirect to `/login`, API calls rejected

---

## Not In Scope

- **Signup email confirmation flow** — depends on Supabase email template config
- **Password reset** — separate page, can follow later
- **OAuth providers** (Google, GitHub login) — future enhancement
- **Role-based access** — single-user model for now
