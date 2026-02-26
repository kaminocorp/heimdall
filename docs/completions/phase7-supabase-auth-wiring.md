# Phase 7 — Supabase Auth Wiring

Replaced placeholder login flow with real Supabase email/password authentication via `@supabase/supabase-js`.

Spec: [supabase-auth-wiring.md](../executing/supabase-auth-wiring.md)

---

## What Changed

### New Files

- **`frontend/src/lib/supabase.ts`** — singleton Supabase client, reads `VITE_SUPABASE_URL` and `VITE_SUPABASE_ANON_KEY` from Vite env vars.
- **`frontend/.env`** — frontend environment variables (not committed, gitignored). Contains the project's Supabase URL and publishable anon key.

### Modified Files

#### `frontend/package.json`
- Added `@supabase/supabase-js` runtime dependency.

#### `frontend/src/stores/auth.ts` — full rewrite
- **Before:** Simple token string in `localStorage`, `login(token)` / `logout()`.
- **After:** Supabase session-aware store with:
  - `init()` — calls `supabase.auth.getSession()` to recover session on page load, subscribes to `onAuthStateChange` for token refresh and logout events.
  - `login(email, password)` — calls `supabase.auth.signInWithPassword()`.
  - `signup(email, password)` — calls `supabase.auth.signUp()`.
  - `logout()` — calls `supabase.auth.signOut()`.
  - `token` — computed from `session.access_token`, auto-updates on refresh.
  - `user` — exposes `session.user` for display (email, id).
  - `initialized` — flag for router guard to avoid premature redirects.

#### `frontend/src/pages/LoginPage.vue` — full rewrite
- Calls real `auth.login(email, password)` / `auth.signup(email, password)`.
- Loading state disables submit button, shows "Please wait…".
- Error banner displays Supabase error messages.
- Toggle between sign-in and sign-up modes on the same page.
- Signup with email confirmation shows "Check your email" message.

#### `frontend/src/App.vue` — rewrite
- Calls `auth.init()` on mount, gates the app behind a `ready` flag.
- Shows "Loading…" while session recovery is in progress — prevents flash of login page on reload.

#### `frontend/src/router/index.ts` — guard update
- Skips redirect if auth is not yet initialized (`!auth.initialized`).
- Redirects authenticated users away from `/login` → `/` (dashboard).

#### `frontend/src/components/common/AppSidebar.vue` — logout + user display
- Shows authenticated user's email at the bottom of the sidebar.
- "Sign out" button calls `auth.logout()` then redirects to `/login`.
- Uses `mt-auto` to push the logout section to the bottom of the nav.

#### `.env.example`
- Added `VITE_SUPABASE_URL` and `VITE_SUPABASE_ANON_KEY` placeholder entries under a "Frontend" section.

### Unchanged Files

- **`frontend/src/api/client.ts`** — no changes needed. The axios interceptor reads `auth.token`, which is now a computed that auto-updates from the Supabase session. Token refresh is transparent.
- **`backend/`** — no backend changes. The backend already validates JWTs via JWKS endpoint and `GET /api/auth/me` works with real Supabase tokens.

---

## Token Flow

```
User submits email/password
  → supabase.auth.signInWithPassword()
  → Supabase returns session { access_token (JWT), refresh_token }
  → auth store updates reactive session ref
  → token computed exposes access_token
  → axios interceptor attaches Authorization: Bearer <token>
  → Backend validates JWT via JWKS endpoint
  → WebSocket connects with ?token=<access_token>

Token expires (~1 hour)
  → Supabase JS auto-refreshes via refresh_token
  → onAuthStateChange fires TOKEN_REFRESHED
  → session ref updates, token computed updates
  → Next API call uses new token automatically
```

---

## Verification

1. `npm run build` — TypeScript compilation clean, no errors.
2. Navigate to `/login` — sign-in form renders.
3. Sign in with Supabase user — redirects to dashboard.
4. `GET /api/auth/me` returns correct user.
5. Agent Chat WebSocket connects with real JWT.
6. Token refresh — transparent, no user-visible interruption.
7. Sign out — redirects to `/login`, subsequent API calls rejected.
