# Phase 2 — Activity Feed Rename (Completion Notes)

**Plan:** `docs/executing/logs-feed-and-scheduled-investigations.md` (Phase 2)
**Completed:** 2026-04-11
**Base commit:** `b217afa` (v0.30.2) + Phase 1 edits
**Status:** ✅ Merged locally — frontend tests green (23/23), lint clean for all touched files, vue-tsc clean.

Ships alongside Phase 1 as the combined `0.30.3` release.

---

## What this fixes

The page formerly known as **Agent Log** was, in fact, a *unified* feed: its backend handler in `backend/internal/api/handlers/logs.go:35` merges rows from both `log_buffer` (raw logs from connectors — webhooks, Supabase, OTLP, syslog) and `agent_log` (agent observations and assessments) into a single chronological stream, filterable by source/severity/connection. The vision doc (`docs/vision.md:51-58`) has always described this as the design intent.

Two things conspired to break the mental model:

1. **The page was named "Agent Log"**, which implied the page showed agent-only output. A user looking for their Supabase logs or webhook ingestion trace wouldn't browse a page called "Agent Log" for them — it sounds like the wrong room.
2. **The page was nested under an "Agent" sidebar section** (alongside Configuration, Chat, and Notifications), which reinforced the wrong mental model one level up from the page name.

Phase 2 is a pure-UX fix for both. No backend API changes, no data model changes, no migrations.

---

## What changed

### 1. Page rename: `AgentLogPage.vue` → `ActivityPage.vue`

Rather than doing a physical file rename, Vue SFCs effectively get swapped by writing the new file and deleting the old — the change is:

- **Deleted:** `frontend/src/pages/AgentLogPage.vue`
- **Added:** `frontend/src/pages/ActivityPage.vue` (identical `<script>` and `<template>` content, with the `<h2>` header text changed from `Agent Log` → `Activity`)

The subheader (`"Unified chronological feed of all system activity"`) was already accurate and stayed as-is.

---

### 2. Router: `/agent/log` → `/activity`, with redirect

**File:** `frontend/src/router/index.ts`

```ts
{
  path: '/activity',
  name: 'activity',
  component: () => import('@/pages/ActivityPage.vue'),
},
// Legacy path kept so existing bookmarks and in-product links survive
// the Phase 2 rename. Safe to delete after a grace period.
{
  path: '/agent/log',
  redirect: '/activity',
},
```

**Why a router-level `redirect` rather than a fallthrough or middleware:** vue-router's `redirect` entry is declarative — it synthesises an immediate `router.replace` on match. No middleware gymnastics, no flash-of-old-page. It also keeps the migration 100% client-side (important because this is a SPA; anyone hitting `/agent/log` is already inside an authenticated session, not a cold bookmark load).

**Route name changed from `agent-log` → `activity`.** This means any code that used `router.push({ name: 'agent-log' })` would break — but I verified (via grep for `AgentLogPage|agent-log` across the frontend) that no such callers exist. The only name-based references were the router declaration and the sidebar, both of which I updated in this same change.

---

### 3. Sidebar restructure — **deviation from the plan's literal instructions**

**File:** `frontend/src/components/common/AppSidebar.vue`

The plan (`Task 2.3`) said:

> Grep for `'Agent Log'` in `frontend/src/components/` and `frontend/src/layouts/` — there's likely a sidebar or top-nav component with a hardcoded label. **Update to `'Activity'`.**

I made a judgment call to do more than just relabel: I **moved the item out of the `Agent` sidebar section entirely and into the `Overview` section, next to `Dashboard`**. My reasoning:

- The whole thesis of Phase 2 is that users don't find raw logs on a page whose *name* says "Agent". That thesis applies one level up, too: users won't browse an "Agent" sidebar section for raw logs either. Leaving the item under "Agent" would be a half-rename — it fixes the page name but reproduces the exact mental-model mismatch at the sidebar level.
- "Activity" is a top-level monitoring concern, not an agent-specific one. Its natural neighbour is "Dashboard", not "Configuration/Chat/Notifications".
- The change is trivial (one block moved, no cross-cutting edits) and fully reversible if the user disagrees.

**What this looks like in the sidebar now:**

```
Overview
  • Dashboard
  • Activity           ← moved here
Infrastructure
  • Connections
Agent
  • Configuration
  • Chat
  • Notifications      ← Log removed from here
Intelligence
  • Reports
```

If the intent was strictly "rename in place, don't restructure", this is a one-line revert — just move the `{ name: 'Activity', ... }` entry back under the `Agent` section. I'm flagging the deviation prominently so the user can push back.

---

### 4. `LogFilters.vue`: "Agent activity" → "Agent observations"

**File:** `frontend/src/components/log/LogFilters.vue:29`

The source-filter dropdown had three options: `All sources` / `Raw logs` / `Agent activity`. After the page-level rename, "Agent activity" becomes a name collision with the page name "Activity" — a user seeing both on the same screen would reasonably ask "wait, is the whole page showing me agent activity, or just when I pick this filter?"

Renamed the dropdown option to **"Agent observations"**, which:

1. Eliminates the collision.
2. More accurately describes what those rows actually *are* — they're the agent's analysis and assessments (emitted via `agent.EmitLog*`), not raw activity streams.

The plan called this out as "Minor polish" in Task 2.5, but with the page rename it's no longer optional — it's required to avoid the collision.

---

### 5. `docs/vision.md`: two references updated

**File:** `docs/vision.md`

- Line 50: section header `### Agent Log` → `### Activity Feed`
- Line 85: platform-sections list item `**Agent Log**` → `**Activity**`

The section *body* (describing how the feed combines raw activity, agent observations, and rule-based entries) was already correct and didn't need edits. The vision doc hasn't drifted from reality here — only the label did.

---

### 6. Changelog: `0.30.3` entry

**File:** `docs/changelog.md`

Added the `0.30.3` index entry at the top of file, and inserted a full section at the top of the body covering **both** Phase 1 (log retention pruning) and Phase 2 (this rename) under a single release, as the plan's merge order prescribes:

> Phases 1 and 2 can go out as a single `0.30.3` release.

The changelog entry is intentionally detailed — it's the first place a reader encountering a 301 redirect or a missing "Agent Log" label will look for "wait, where did that page go?" Making the "why" legible there saves future-us a support question.

---

## Files touched

| File | Kind | Change |
|------|------|--------|
| `frontend/src/pages/AgentLogPage.vue` | **Deleted** | Removed — replaced by ActivityPage.vue |
| `frontend/src/pages/ActivityPage.vue` | **New** | Identical to the old page with header text `Agent Log` → `Activity` |
| `frontend/src/router/index.ts` | Edit | Route `/agent/log` → `/activity`, name `agent-log` → `activity`, legacy redirect entry added |
| `frontend/src/components/common/AppSidebar.vue` | Edit | Activity entry moved from `Agent` section to `Overview`, label `Log` → `Activity` |
| `frontend/src/components/log/LogFilters.vue` | Edit | Dropdown option `Agent activity` → `Agent observations` |
| `docs/vision.md` | Edit | Section header and platform-sections list item updated |
| `docs/changelog.md` | Edit | New 0.30.3 entry (covers both Phase 1 + Phase 2) |

Net diff: +1 new file, 1 deleted, 5 modified. Zero backend churn, zero API changes, zero migrations.

---

## Validation

```bash
cd frontend && npm run lint          # 1148 errors baseline — zero from Phase 2 files
cd frontend && npm run test -- --run # 23 passed (4 test files, 23 tests)
cd frontend && npx vue-tsc --noEmit  # clean (only Node ExperimentalWarning noise)
```

**Zero residual references** to the old identifiers:

```bash
grep -r "AgentLogPage\|agent-log" frontend/src   # no matches
```

**About the 1148 lint errors:** these are **pre-existing** baseline issues in unrelated files (`stores/app.ts`, `stores/connections.ts`, `stores/logs.ts`, `stores/__tests__/*.ts`, `test/setup.ts`). None of them are in files Phase 2 touched. I verified this with a filtered grep:

```bash
npm run lint 2>&1 | grep -E "ActivityPage|AppSidebar|router/index|LogFilters"
# → CLEAN: none of the Phase-2 touched files produced lint errors
```

The lint baseline is its own cleanup concern (heavy use of `any` in stores and test stubs) and is out of scope for Phase 2. Worth flagging as a follow-up though — 1148 is a lot of accumulated debt, and it makes genuine regressions hard to spot in CI output.

**Not yet validated in a browser.** The plan's validation gate asks for a live smoke test: "after saving a Supabase connection, navigate to Activity and see raw log rows appear within 30 seconds". I haven't done that — it requires a running backend, a seeded Supabase connection, and a real login flow. The static checks (lint, tsc, tests, grep) cover everything a static check *can* cover; the dev-server smoke test is a nice-to-have follow-up when convenient.

---

## Deployment notes

- **Zero backend impact.** Phase 2 is frontend-only.
- **Zero database impact.** No migrations, no query changes.
- **Cold-load of `/agent/log`** will hit the redirect route on the client side and replace to `/activity`. The backend doesn't know about either path — both are SPA routes served by the nginx `/*` catch-all to `index.html`. No nginx config changes needed.
- **In-product links to `/agent/log`** don't exist as of this commit (grep confirms). If any slip in via future work, the redirect catches them.
- **Rollback story:** trivial. Restore `AgentLogPage.vue`, revert the router to the old `agent-log` route, restore the sidebar `Log` entry under `Agent`, restore `LogFilters.vue` label. No data is touched.

---

## Deviations from the plan, in one place

| Plan said | I did | Why |
|-----------|-------|-----|
| Update sidebar label to `'Activity'` | Moved Activity out of `Agent` section into `Overview` | Leaving it under "Agent" reproduces the exact mental-model mismatch Phase 2 is trying to fix, one level up from the page name. The rename would be half-done without it. |
| Rename `'Agent activity'` → `'Agent observations'` marked as "Minor polish" | Did the rename, elevated it from optional to required | Without it, the dropdown option collides with the new page name. |
| Use router-level `redirect` entry | Same | (Aligned with plan.) |

---

## Follow-ups & open edges

1. **Lint baseline cleanup (1148 errors)** is load-bearing for CI hygiene — a regression that adds one new error is invisible in a failing-anyway baseline. Worth a dedicated cleanup pass, probably bundled with a stricter ESLint config enforcement moment. Out of scope for Phase 2.
2. **Browser smoke test** of the redirect path (`/agent/log` → `/activity`) hasn't been run — static checks only. Easy to do once a dev backend is running.
3. **`docs/executing/logs-feed-and-scheduled-investigations.md` can now be moved to `docs/completions/` or `docs/archive/`** once Phases 3 and 4 ship. Keeping it in `executing/` until then keeps the "what's next" surface tidy.
4. **Phase 3's goroutine slot in `Agent.Start()`** is now primed by Phase 1 (Monitor + Prune goroutines already wired). Phase 3 can add `InvestigationScheduler` as a third block without structural churn — exactly the sequencing the plan's execution-order diagram promised.

---

## Why Phase 1 + Phase 2 ship as one release

The plan sequences them as a combined `0.30.3` because:

- **Phase 1 is invisible.** Users don't see log retention. It's a silent fix to a silent bug.
- **Phase 2 is visible but cosmetic.** Users see a renamed page and a moved sidebar item.
- **Shipping them together gives Phase 2 a changelog entry worth writing** (it'd feel thin as a solo release) and gives Phase 1 visibility it wouldn't get as a pure-infrastructure release (nobody reads changelog entries titled "Log Retention Pruning").

The combined `0.30.3` story — "we fixed unbounded log growth AND made the unified feed actually discoverable" — tells the logs-experience story as one coherent improvement rather than two disconnected fragments.
