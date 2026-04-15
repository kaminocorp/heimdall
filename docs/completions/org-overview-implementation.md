# Organisation Overview & Context Switching — Implementation Plan

**Status:** Not started
**Owner:** TBD
**Prereqs:** None — all core scaffolding (auth, org, app CRUD, sidebar, header) exists

---

## Why this exists

Heimdall currently operates in a single implicit context: the user is always
"inside" an application. The organisation exists in the data model but is
invisible in the UX — it's a read-only card on the Settings page. There is no
way to:

- View all your applications at a glance (Supabase's "Projects" page)
- Manage organisation-level concerns (team members, billing, general settings)
- Switch between organisations (multi-org support)
- Create or join a new organisation

Clicking the org name in the header should "zoom out" from the per-app world
into an **org-level context** with its own sidebar navigation and overview page —
exactly the pattern used by Supabase, Vercel, and Fly.io.

---

## Design goals

1. **Two navigation contexts.** App-level pages (Dashboard, Connections, Agent,
   etc.) keep their current sidebar. Org-level pages (Projects, Team, Billing,
   Settings) get a distinct sidebar. The top header breadcrumb indicates which
   context you're in.
2. **Org overview as the hub.** The org home page shows all applications as
   cards (grid + list toggle), with health/status at a glance and a prominent
   "New application" action — the Supabase Projects page pattern.
3. **Multi-org foundation.** The data model moves from 1:1 (`users.org_id`) to
   many-to-many (`org_members` join table with roles). Even if V1 only surfaces
   a single org, the schema supports switching from day one.
4. **Non-destructive migration.** Existing `users.org_id` data is migrated into
   `org_members` rows, then the column is dropped. Zero data loss, rollback-safe
   with down migration.
5. **Incremental delivery.** Each phase is independently shippable and testable.

---

## Data model changes

### Current

```
users.org_id → organizations.id   (1:1, nullable FK)
```

### Target

```
┌─────────────┐       ┌──────────────────┐       ┌─────────────────┐
│    users     │──────▶│   org_members    │◀──────│  organizations  │
│              │       │                  │       │                 │
│ id           │       │ user_id (FK)     │       │ id              │
│ email        │       │ org_id  (FK)     │       │ name            │
│ created_at   │       │ role (enum)      │       │ slug            │
│              │       │ created_at       │       │ created_at      │
└──────────────┘       └──────────────────┘       └─────────────────┘
                             ▲
                             │ roles: 'owner' | 'admin' | 'member'
```

The `org_members` table replaces `users.org_id` and adds role-based access. A
user can belong to multiple orgs. The "current org" is tracked client-side
(localStorage + store), not in the database.

---

## Route structure

### Current (flat, app-context only)

```
/dashboard
/connections
/agent/config
/agent/chat
/schedules
/activity
/reports
/notifications
/settings          ← mixed: profile + org + app management
```

### Target (two contexts)

```
# Org-level routes (org sidebar)
/org                        ← Org overview: application cards grid
/org/team                   ← Team members: invite, roles, remove
/org/settings               ← Org settings: name, slug, danger zone
/org/billing                ← Billing (placeholder/coming soon)

# App-level routes (app sidebar — unchanged)
/dashboard
/connections
/agent/config
/agent/chat
/schedules
/activity
/reports
/notifications

# Account-level (no sidebar, or minimal)
/account/profile            ← Extracted from current /settings
/login
/onboarding
```

The existing `/settings` page is decomposed: org settings move to `/org/settings`,
profile moves to `/account/profile`, and app management becomes the org overview
(`/org`).

---

## Sidebar definitions

### Org sidebar (`/org/*` routes)

```
┌────────────────────────┐
│  ● Projects            │  ← /org (the app cards grid)
│  ● Team                │  ← /org/team
│  ● Integrations        │  ← /org/integrations (future: org-wide integrations)
│  ● Usage               │  ← /org/usage (future: usage metrics)
│  ● Billing             │  ← /org/billing (placeholder)
│  ● Organisation Settings│ ← /org/settings
└────────────────────────┘
```

### App sidebar (current routes — unchanged)

```
┌────────────────────────┐
│ Overview               │
│  ● Dashboard           │
│  ● Activity            │
│ Infrastructure         │
│  ● Connections         │
│ Agent                  │
│  ● Configuration       │
│  ● Schedules           │
│ Intelligence           │
│  ● Reports             │
│  ● Notifications       │
└────────────────────────┘
```

The sidebar component becomes context-aware: it reads the current route prefix
and renders the appropriate navigation set.

---

## Header breadcrumb behaviour

### When in org context (`/org/*`)

```
[Heimdall logo] / Crimson Sun Technologies
                  ^^^^^^^^^^^^^^^^^^^^^^^^^^^
                  Org dropdown (switch orgs)
```

No app selector shown — you're at the org level.

### When in app context (`/dashboard`, `/connections`, etc.)

```
[Heimdall logo] / Crimson Sun Technologies / heimdall-prod
                  ^^^^^^^^^^^^^^^^^^^^^^^^^^^  ^^^^^^^^^^^^^
                  Clickable → navigates to /org   App dropdown (switch apps)
```

Clicking the org name navigates to `/org` (the overview). This is the primary
entry point to the org context.

---

## Implementation phases

---

### Phase 1 — Database: `org_members` join table

**Goal:** Replace `users.org_id` with a proper many-to-many membership table
with roles, migrate existing data, and update all backend queries.

#### Tasks

1. **Create migration `017_org_members.up.sql`**
   - Create `org_member_role` enum: `'owner'`, `'admin'`, `'member'`
   - Create `org_members` table: `user_id`, `org_id`, `role`, `created_at`
   - Unique constraint on `(user_id, org_id)`
   - Migrate existing data: `INSERT INTO org_members SELECT id, org_id, 'owner' FROM users WHERE org_id IS NOT NULL`
   - Drop `users.org_id` column
   - Down migration reverses all steps

2. **Update sqlc queries**
   - `organizations.sql`: Replace `GetOrganizationByUser` (joins on `users.org_id`) with query that joins via `org_members`
   - Add `ListOrganizationsByUser` query (returns all orgs for a user)
   - Add `GetOrgMembership` query (user_id + org_id → role)
   - Add `ListOrgMembers` query (org_id → list of members with email + role)
   - Add `CreateOrgMember`, `UpdateOrgMemberRole`, `RemoveOrgMember` queries
   - `users.sql`: Remove `SetUserOrg` query, add `GetUserPrimaryOrg` (first org by `created_at`)
   - `applications.sql`: Update `GetApplicationByOrgUser` — now joins via `org_members`

3. **Run `make sqlc-generate`** — regenerate Go types

4. **Update backend handlers**
   - `onboard` handler: Insert into `org_members` instead of setting `users.org_id`
   - Auth middleware: Resolve user's org(s) via `org_members` join
   - Any handler using `user.OrgID` must now query `org_members`

5. **Update RLS policies** if any reference `users.org_id` directly

6. **Write migration tests** — verify data migrates correctly, verify rollback

#### Verification
- `make migrate-up` succeeds
- `make migrate-down` succeeds (rollback)
- All existing backend tests pass
- Manual: existing user can still log in and see their org/apps

---

### Phase 2 — Backend: Org-level API endpoints

**Goal:** Expose org membership and management via new API routes.

#### Tasks

1. **New handler file: `internal/api/handlers/org_members.go`**
   - `GET /api/org/members` — list members of user's current org
   - `POST /api/org/members/invite` — invite user by email (creates pending membership)
   - `PUT /api/org/members/{userId}/role` — update member role (owner/admin only)
   - `DELETE /api/org/members/{userId}` — remove member (owner/admin only, can't remove self if sole owner)

2. **Update existing org handler**
   - `GET /api/org` — include `role` field in response (from `org_members`)
   - `PUT /api/org` — update org name/slug (owner/admin only)
   - `GET /api/orgs` — **new**: list all orgs the user belongs to (for multi-org switcher)

3. **Role-based authorization helper**
   - `requireOrgRole(minRole)` middleware: checks `org_members.role >= minRole`
   - Role hierarchy: `owner > admin > member`

4. **Wire routes in `server.go`**
   ```
   r.Route("/api/org", func(r chi.Router) {
       r.Use(authMiddleware)
       r.Get("/", getOrganization)
       r.Put("/", requireOrgRole("admin"), updateOrganization)
       r.Get("/members", listOrgMembers)
       r.Post("/members/invite", requireOrgRole("admin"), inviteMember)
       r.Put("/members/{userId}/role", requireOrgRole("owner"), updateMemberRole)
       r.Delete("/members/{userId}", requireOrgRole("admin"), removeMember)
   })
   r.Get("/api/orgs", authMiddleware, listUserOrganizations)
   ```

5. **Write handler tests** for each endpoint

#### Verification
- All new endpoints return correct data
- Role checks enforce correctly (member can't invite, admin can't change roles, etc.)
- Existing endpoints unaffected
- `make test` passes

---

### Phase 3 — Frontend: Routing & layout context switching

**Goal:** Introduce the two-context routing model. Org routes get an org sidebar,
app routes keep the current sidebar. The header breadcrumb becomes the context
switch mechanism.

#### Tasks

1. **Create org-level pages (empty shells initially)**
   - `pages/org/OrgOverviewPage.vue` — placeholder "Projects" page
   - `pages/org/OrgTeamPage.vue` — placeholder "Team" page
   - `pages/org/OrgSettingsPage.vue` — placeholder "Settings" page
   - `pages/org/OrgBillingPage.vue` — placeholder "Billing" page

2. **Create `OrgSidebar.vue`** component
   - Same visual structure as `AppSidebar.vue` (reuse styling)
   - Renders org-level nav items: Projects, Team, Billing, Organisation Settings
   - Active state detection via `route.name` prefix

3. **Update `DefaultLayout.vue`** — context-aware sidebar
   - Compute `isOrgContext` from `route.path.startsWith('/org')`
   - Render `OrgSidebar` when in org context, `AppSidebar` when in app context
   - Both sidebars share the same slot/position in the layout grid

4. **Register routes in `router/index.ts`**
   ```typescript
   // Org-level routes
   { path: '/org', name: 'org-overview', component: OrgOverviewPage },
   { path: '/org/team', name: 'org-team', component: OrgTeamPage },
   { path: '/org/settings', name: 'org-settings', component: OrgSettingsPage },
   { path: '/org/billing', name: 'org-billing', component: OrgBillingPage },
   ```

5. **Update `AppHeader.vue`** breadcrumb behaviour
   - Org name becomes a `<router-link to="/org">` (always clickable)
   - When on `/org/*` routes: hide the app dropdown
   - When on app routes: show both org link + app dropdown (current behaviour)

6. **Update route guard** in `router/index.ts`
   - Org routes should not require `currentAppId` to be set
   - App routes should still redirect to onboarding if no org exists

7. **Redirect `/settings` → split targets**
   - `/settings` → redirect to `/org/settings` (or show a deprecation notice)
   - Profile section → `/account/profile` (or keep inline for now)

#### Verification
- Clicking org name in header navigates to `/org`
- `/org` shows org sidebar, `/dashboard` shows app sidebar
- Route guards work correctly for both contexts
- No regressions on existing app-level navigation
- `vue-tsc --noEmit` clean, `vite build` clean

---

### Phase 4 — Org overview page (Projects grid)

**Goal:** Build the main org landing page — a card grid of all applications with
status, health indicators, and quick actions. This is the Supabase "Projects"
page equivalent.

#### Tasks

1. **API client additions** (`api/applications.ts`)
   - Ensure `listApplicationsWithCounts()` returns all data needed for cards
   - Add `getAppHealthSummary(appId)` if not already available (connection
     status, recent error count, monitoring state)

2. **Build `OrgOverviewPage.vue`**
   - **Header row:** "Projects" title + search input + sort dropdown (by name / status / date) + view toggle (grid/list) + "+ New project" button
   - **Card grid:** Each app rendered as a card showing:
     - App name + status badge (active/paused/archived)
     - Connection count + schedule count
     - Last activity timestamp
     - Monitoring mode indicator (continuous/periodic/off)
     - Three-dot menu: Open, Pause, Settings, Delete
   - **Empty state:** Friendly message + prominent "Create your first project" CTA
   - **Search:** Client-side filter by app name
   - **Grid/List toggle:** Grid = cards, List = compact table rows (persist preference to localStorage)

3. **App card component: `components/org/AppCard.vue`**
   - Reusable card matching Heimdall's design language
   - Click anywhere on card → `router.push('/dashboard')` + `appStore.selectApp(appId)`
   - Status-dependent styling (paused = muted, archived = greyed out)

4. **App list row component: `components/org/AppListRow.vue`**
   - Table row variant for list view
   - Same data, compact layout

5. **Wire "New application" to existing `AppWizard` component**
   - On create success: select the new app and navigate to `/dashboard`

6. **Wire card click navigation**
   - Selecting an app from this page should: set `appStore.currentAppId` → navigate to `/dashboard`
   - This is the "zoom in" action (org → app context)

#### Verification
- `/org` shows all apps as cards with correct data
- Search filters correctly
- Grid/List toggle works and persists
- Clicking a card enters app context (sidebar switches, breadcrumb updates)
- "New application" flow works end-to-end
- Responsive: cards reflow on mobile

---

### Phase 5 — Team management page

**Goal:** Build the Team page where org owners/admins can view members, invite
new ones, and manage roles.

#### Tasks

1. **API client additions** (`api/organizations.ts`)
   - `listOrgMembers()` → `GET /api/org/members`
   - `inviteOrgMember(email, role)` → `POST /api/org/members/invite`
   - `updateMemberRole(userId, role)` → `PUT /api/org/members/{userId}/role`
   - `removeOrgMember(userId)` → `DELETE /api/org/members/{userId}`

2. **Build `OrgTeamPage.vue`**
   - **Member list table:**
     - Avatar placeholder (initials) + email + role badge + joined date
     - Role dropdown (owner can change roles of others)
     - Remove button (with confirmation modal)
     - Current user row highlighted, can't remove self if sole owner
   - **Invite section:**
     - Email input + role selector (admin/member) + "Send invite" button
     - Pending invites shown separately with "Revoke" action
   - **Role descriptions:** Brief tooltip/help text explaining what each role can do

3. **Invite flow (backend)**
   - For V1: "invite" creates an `org_members` row immediately if the email
     matches an existing user. If no user exists, store a `pending_invites`
     record and show a "pending" badge. When that user signs up and logs in,
     auto-associate them.
   - Stretch: send invite email via Supabase edge function or webhook

4. **Permission gating in UI**
   - Members see the list but can't invite/remove/change roles
   - Admins can invite and remove members
   - Owners can do everything including changing roles

#### Verification
- Member list shows all org members with correct roles
- Owner can invite, change roles, remove members
- Admin can invite and remove but not change roles
- Member sees read-only list
- Self-removal prevented for sole owner
- `make test` passes for all new backend endpoints

---

### Phase 6 — Organisation settings page

**Goal:** Build a proper org settings page with editable fields and a danger zone.

#### Tasks

1. **Build `OrgSettingsPage.vue`**
   - **General section:**
     - Editable org name (text input + save button)
     - Editable org slug (text input + save, with uniqueness validation)
     - Org ID (read-only, copyable)
     - Created date
   - **Danger zone** (red-bordered section at bottom):
     - "Delete organisation" — requires typing org name to confirm
     - Warning: "This will permanently delete all applications, connections,
       logs, and agent data. This action cannot be undone."
   - Save triggers `PUT /api/org` (existing endpoint, enhanced with validation)

2. **Backend: org deletion endpoint**
   - `DELETE /api/org` — owner only
   - Cascade delete: org → applications → connections, logs, agent configs, etc.
   - Requires confirmation header or body field (`confirm: org-slug`)

3. **Update `appStore`** — if org is deleted, clear state and redirect to onboarding

#### Verification
- Org name/slug updates persist and reflect in header immediately
- Slug uniqueness enforced
- Delete flow requires exact name match
- Successful deletion redirects to onboarding
- Only owners see danger zone

---

### Phase 7 — Multi-org switching

**Goal:** Enable users who belong to multiple orgs to switch between them via the
header dropdown.

#### Tasks

1. **Update `appStore`** (or create dedicated `orgStore`)
   - New state: `organizations: Organization[]` (all orgs user belongs to)
   - New state: `currentOrgId` (persisted to localStorage as `heimdall_current_org`)
   - `init()` fetches all orgs via `GET /api/orgs`, restores last-selected from localStorage
   - `selectOrg(orgId)` — switches org, refetches applications for that org, clears current app

2. **Update `AppHeader.vue`** org dropdown
   - List all orgs the user belongs to
   - Clicking a different org: `appStore.selectOrg(orgId)` → navigate to `/org`
   - Show role badge next to each org
   - "+ New organisation" option at bottom → opens org creation flow

3. **Org creation flow**
   - Simple modal: org name + slug → `POST /api/orgs` (new endpoint)
   - On success: auto-select new org, navigate to `/org`

4. **Join org flow** (stretch)
   - Accept invite link → auto-join org → navigate to `/org`

5. **Update all data-fetching pages**
   - Ensure pages refetch when `currentOrgId` changes (not just `currentAppId`)
   - Connections, logs, etc. are already app-scoped, so switching org → switching
     app handles most cases

#### Verification
- Users in multiple orgs see all orgs in dropdown
- Switching org clears app context and navigates to `/org`
- New org creation works end-to-end
- App-level pages refetch correctly after org switch

---

### Phase 8 — Polish & migration from old settings

**Goal:** Clean up the transition, remove the old `/settings` page, and ensure
all edges are smooth.

#### Tasks

1. **Remove old `SettingsPage.vue`**
   - Profile → `/account/profile` or keep as a lightweight page
   - Organisation section → now at `/org/settings`
   - Applications section → now the org overview at `/org`
   - Redirect `/settings` → `/org/settings` for bookmarks

2. **Update header profile dropdown**
   - "Settings" link → `/org/settings` (or split into "Org Settings" + "Profile")

3. **Sidebar polish**
   - Smooth transition animation when switching between org/app sidebar
   - Consistent active-state styling across both sidebars

4. **Breadcrumb refinement**
   - Ensure breadcrumb hierarchy always makes sense
   - On org pages: `Heimdall / Crimson Sun Technologies / [page name]`
   - On app pages: `Heimdall / Crimson Sun Technologies / heimdall-prod`

5. **Mobile navigation**
   - Org sidebar works in mobile overlay drawer
   - Context switching works on mobile

6. **Empty states**
   - No orgs (new user) → onboarding flow (already exists)
   - No apps in org → prominent CTA on org overview
   - No team members (just you) → invite prompt on team page

7. **Update frontend tests**
   - Sidebar context switching tests
   - Route guard tests for org vs app context
   - Header breadcrumb tests

#### Verification
- Full regression: all existing features work
- All new features work end-to-end
- `vue-tsc --noEmit` clean
- `vite build` clean
- `vitest run` all pass
- `go test ./...` all pass
- Mobile navigation works

---

## Risk & considerations

| Risk | Mitigation |
|------|------------|
| **Migration breaks existing users** | Down migration restores `users.org_id`; data copied back from `org_members` |
| **RLS policies depend on `users.org_id`** | Audit all policies in Phase 1; rewrite to join via `org_members` |
| **Implicit `currentAppId` breaks on org pages** | Route guard allows org pages without `currentAppId`; app pages still require it |
| **Performance: N+1 on org overview** | Use `listApplicationsWithCounts` (single query with joins) |
| **Invite without email service** | V1 uses "instant add" for existing users; email invites are a stretch goal |

---

## Out of scope (future work)

- **Billing integration** — page exists as placeholder, actual Stripe/payment integration is separate
- **Usage metrics** — org-level usage dashboard (API calls, log volume, etc.)
- **Org-wide integrations** — shared connections across all apps in an org
- **RBAC on app level** — per-app roles beyond org membership
- **Audit log** — who changed what at the org level
- **SSO / SAML** — enterprise auth for org membership
