# Fix #5 — Log severity mismatch `'error'` vs `'critical'`

**Review ref:** scaffolding-review.md, Must Fix #5
**Date:** 2026-02-19

## Problem

The domain model (`types/log.ts`) and constants (`SEVERITY_LEVELS = ['info', 'warning', 'critical']`) define the highest severity as `'critical'`. Two log components used `'error'` instead:

- `LogEntry.vue` — conditional CSS class checked `entry.severity === 'error'`, so critical entries were never highlighted red.
- `LogFilters.vue` — dropdown option had value `'error'`, so filtering by "Error" would match nothing from the backend.

## Fix

Changed `'error'` → `'critical'` in both components to match the domain model.

## Files modified

| File | Line | Change |
|------|------|--------|
| `frontend/src/components/log/LogEntry.vue` | 13 | `'error'` → `'critical'` in `:class` binding |
| `frontend/src/components/log/LogFilters.vue` | 20 | `<option value="error">Error</option>` → `<option value="critical">Critical</option>` |

## Verification

- `vue-tsc --noEmit` — zero errors
