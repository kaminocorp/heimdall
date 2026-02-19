# Fix 011 — Remove duplicate useAuth composable

**Issue:** #11 from scaffolding-review.md
**Date:** 2026-02-19

## Problem

`frontend/src/composables/useAuth.ts` duplicated the exact same token/login/logout logic already provided by the Pinia store (`frontend/src/stores/auth.ts`). The composable used a module-scope `ref` (poor man's store) and was never imported anywhere — dead code.

## Fix

Deleted `frontend/src/composables/useAuth.ts`. The Pinia store `useAuthStore` is the single source of truth for auth state.

## Verification

- `vue-tsc --noEmit` passes clean.
- Grep confirms no remaining imports of `useAuth` (only `useAuthStore` is used).
