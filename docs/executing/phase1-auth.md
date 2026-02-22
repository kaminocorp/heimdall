# Phase 1 — User & Auth Implementation Plan

Reference: [MVP Roadmap](./mvp-roadmap.md) · [Blueprint](./blueprint.md)

---

## Approach

Leverage **Supabase Auth** (GoTrue) instead of rolling our own auth system. Supabase handles registration, login, password hashing, JWT issuance, and token refresh. Our backend simply **verifies** Supabase-issued JWTs. A database trigger syncs new sign-ups from `auth.users` into a `public.users` table for application-level use.

### Why this approach

- Supabase Auth is battle-tested — handles bcrypt, rate limiting, email confirmation, password reset out of the box.
- We already have a Supabase Postgres instance; Auth is included at no extra cost.
- The backend stays simple: one middleware that validates JWT signatures using the Supabase JWT secret. No token issuance, no password storage.
- The frontend uses `@supabase/supabase-js` — handles login, signup, token refresh, and session persistence automatically.

---

## Tasks

### 1. Migration: `public.users` table + sync trigger

Create migration `006_create_users.up.sql` / `006_create_users.down.sql`.

**Up migration:**

```sql
-- Application users table (synced from Supabase Auth)
CREATE TABLE users (
    id         UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    email      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Trigger: auto-create public.users row on auth.users insert
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.users (id, email)
    VALUES (NEW.id, NEW.email);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW
    EXECUTE FUNCTION public.handle_new_user();
```

**Down migration:**

```sql
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
DROP FUNCTION IF EXISTS public.handle_new_user();
DROP TABLE IF EXISTS users;
```

**Notes:**
- `SECURITY DEFINER` lets the trigger function write to `public.users` even though the insert on `auth.users` runs under a different role.
- The `users` table is intentionally minimal — just `id`, `email`, `created_at`. We can add `display_name`, `avatar_url`, etc. later.
- FK to `auth.users(id)` with `ON DELETE CASCADE` ensures cleanup.

**Run:** Apply via `golang-migrate` CLI against the Supabase instance.

---

### 2. Backend: JWT verification middleware

**New env vars:**

| Var | Required | Description |
|-----|----------|-------------|
| `SUPABASE_JWT_SECRET` | Yes | The JWT secret from Supabase project settings (Settings → API → JWT Secret) |

**Config changes** (`internal/config/config.go`):

- Add `SupabaseJWTSecret string` field to `Config`.
- Load from `SUPABASE_JWT_SECRET` env var.
- Add to `Validate()` — required.

**New dependency:**

- `github.com/golang-jwt/jwt/v5` — standard Go JWT library.

**Middleware changes** (`internal/api/middleware/auth.go`):

Replace the stub with real JWT validation:

1. Extract `Authorization: Bearer <token>` from request header.
2. Parse and validate the JWT using the Supabase JWT secret (HMAC-SHA256).
3. Check standard claims: `exp` (not expired), `aud` (optional — can validate against `authenticated`).
4. Extract the user's UUID from the `sub` claim.
5. Store the user ID in the request context (via `context.WithValue`).
6. If validation fails → return `401 Unauthorized` JSON response.

**Context helper** (new file or in middleware package):

```go
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool)
```

Handlers can call this to get the authenticated user's ID.

---

### 3. Backend: Apply auth middleware to protected routes

**Router changes** (`internal/api/router.go`):

- Create the auth middleware instance (pass JWT secret from config).
- Apply it to all `/api/*` routes **except** public endpoints.
- The only public endpoint for now: none — Supabase Auth handles login/signup directly. The existing `POST /api/auth/login` stub can be removed (or repurposed as a lightweight `/api/auth/me` endpoint that returns the current user from the JWT).

Proposed route structure:

```
/api
  /auth/me          GET  — (protected) return current user info from JWT
  /connections/*     — (protected) existing routes
  /agent/*           — (protected) existing routes
  /logs              — (protected) existing route
  /reports/*         — (protected) existing routes
/ws/chat             — (protected, token via query param or first message)
```

---

### 4. Backend: `/api/auth/me` endpoint

Replace the existing `Login` handler stub with a `Me` handler:

- Reads user ID from request context (set by auth middleware).
- Queries `public.users` for the user's email (or returns the JWT claims directly).
- Returns `{ "id": "...", "email": "..." }`.

This gives the frontend a way to verify the session is valid and fetch user info.

**sqlc query** (new file `internal/db/queries/users.sql`):

```sql
-- name: GetUser :one
SELECT id, email, created_at FROM users WHERE id = $1;
```

Run `sqlc generate` after adding this.

---

### 5. Frontend: Supabase client setup

**New dependency:**

```bash
npm install @supabase/supabase-js
```

**New env vars** (Vite — prefixed with `VITE_`):

| Var | Description |
|-----|-------------|
| `VITE_SUPABASE_URL` | Supabase project URL (e.g. `https://xyz.supabase.co`) |
| `VITE_SUPABASE_ANON_KEY` | Supabase anon/public key (safe for frontend) |

**New file** (`src/lib/supabase.ts`):

```typescript
import { createClient } from '@supabase/supabase-js'

export const supabase = createClient(
  import.meta.env.VITE_SUPABASE_URL,
  import.meta.env.VITE_SUPABASE_ANON_KEY
)
```

---

### 6. Frontend: Rewrite auth store

Rewrite `src/stores/auth.ts` to use Supabase Auth:

- `login(email, password)` → calls `supabase.auth.signInWithPassword()`.
- `register(email, password)` → calls `supabase.auth.signUp()`.
- `logout()` → calls `supabase.auth.signOut()`.
- `session` reactive ref — populated from `supabase.auth.getSession()` on init.
- `isAuthenticated` computed — checks session exists and is not expired.
- `token` computed — returns `session.access_token` for the axios interceptor.
- Listen to `supabase.auth.onAuthStateChange()` to keep session reactive.
- Remove manual `localStorage` management — Supabase client handles persistence.

---

### 7. Frontend: Update LoginPage

Rewrite `src/pages/LoginPage.vue`:

- Call `auth.login(email, password)` (which now calls Supabase).
- Show error messages on failure (wrong password, user not found, etc.).
- Add a "Create account" link/toggle to a registration form.
- On success, redirect to `/`.

**Optional:** Add a `RegisterPage.vue` or toggle registration inline on the login page.

---

### 8. Frontend: Update axios interceptor

The existing interceptor in `src/api/client.ts` already attaches `Authorization: Bearer <token>`. Just ensure it reads from the updated auth store's `token` computed (which now returns the Supabase access token).

No changes needed if the store exposes `.token` with the same interface.

---

### 9. Update `.env.example`

Add the new env vars to `.env.example`:

```
SUPABASE_JWT_SECRET=
```

And for the frontend (or a separate `frontend/.env.example`):

```
VITE_SUPABASE_URL=
VITE_SUPABASE_ANON_KEY=
```

---

## Execution order

```
Task 1: Migration (006_create_users)
    └──▶ Task 2: Backend JWT middleware
              └──▶ Task 3: Apply middleware to routes
              └──▶ Task 4: /api/auth/me endpoint + sqlc query
Task 5: Frontend Supabase client
    └──▶ Task 6: Rewrite auth store
              └──▶ Task 7: Update LoginPage
              └──▶ Task 8: Update axios interceptor
Task 9: Update .env.example (can run anytime)
```

Tasks 1-4 (backend) and Tasks 5-8 (frontend) are independent tracks and can be worked in parallel. Task 9 is standalone.

---

## Out of scope (for now)

- Email confirmation flow (Supabase supports it, but we'll disable it initially for dev speed)
- Password reset
- OAuth providers (Google, GitHub, etc.)
- Role-based access control
- Multi-user / teams
- Refresh token rotation (Supabase handles this automatically)
- WebSocket auth (defer to Phase 5 — Agent Chat)
