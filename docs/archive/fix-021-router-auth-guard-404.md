# Fix 021 — Router auth guard and 404 route

**Issue:** #21 from scaffolding-review.md
**Date:** 2026-02-20

## Problem

No `router.beforeEach` navigation guard for auth redirects, and no catch-all route for 404s. Unauthenticated users could navigate to any page, and unknown paths showed a blank page.

## Fix

- Added a `beforeEach` guard that redirects to `/login` when the user is not authenticated (skips the login route itself to avoid infinite redirect).
- Added a catch-all `/:pathMatch(.*)*` route pointing to a new `NotFoundPage.vue`.
- Created `NotFoundPage.vue` — minimal 404 page with a link back to the dashboard.

## Verification

- `vue-tsc --noEmit` passes clean.
