# Fix 024 — Inconsistent page state management for reports

**Issue:** #24 from scaffolding-review.md
**Date:** 2026-02-20

## Problem

`ReportsPage.vue` used local `ref`s for state while all other list pages (`ConnectionsPage`, `AgentLogPage`, etc.) used Pinia stores. Inconsistent pattern.

## Fix

- Created `stores/reports.ts` Pinia store following the same pattern as `stores/connections.ts` (state + loading + fetch action).
- Updated `ReportsPage.vue` to use `useReportsStore()` instead of local refs.

## Verification

- `vue-tsc --noEmit` passes clean.
