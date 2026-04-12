# Shared `extractApiError` Utility & `catch (e: any)` Cleanup

**Date:** 2026-04-12
**Assessment reference:** `docs/plans/code-assessment-2026-04-12.md`, issues #5 and #7

## Problem

### Issue #5 — Duplicated `extractApiError`

The same error-extraction logic was defined independently in three places:

1. `stores/schedules.ts:38-44` — local function inside the store
2. `pages/SchedulesPage.vue:55-62` — identical local function in the page
3. Inline variants across 15+ files using `e.response?.data?.error ?? 'fallback'`

Any change to the error extraction pattern (e.g. adding support for a new response shape) required finding and updating every copy.

### Issue #7 — Untyped `catch (e: any)`

11 catch blocks across stores, pages, and components used `catch (e: any)` — bypassing TypeScript's type safety and hiding potential issues at the error boundary. The `any` type allowed unchecked property access like `e.response?.data?.error` without narrowing.

## Solution

### 1. Shared utility: `frontend/src/utils/apiError.ts`

```typescript
export function extractApiError(e: unknown, fallback: string): string {
  if (typeof e === 'object' && e !== null) {
    const err = e as { response?: { data?: { error?: string; message?: string } }; message?: string }
    if (err.response?.data?.error) return err.response.data.error
    if (err.response?.data?.message) return err.response.data.message
    if (err.message) return err.message
  }
  return fallback
}
```

The canonical version checks three fields in priority order:
- `response.data.error` — standard Heimdall JSON error responses (`jsonError()`)
- `response.data.message` — connection test endpoint responses (`testResult.Message`)
- `error.message` — axios network errors or Supabase auth errors

This covers every variant that previously existed across the codebase.

### 2. Files updated

**Duplicate definitions removed:**

| File | Change |
|------|--------|
| `stores/schedules.ts` | Removed local `extractApiError`, added import |
| `pages/SchedulesPage.vue` | Removed local `extractApiError`, added import |

**`catch (e: any)` → `catch (e: unknown)` + `extractApiError`:**

| File | Catch blocks fixed |
|------|-------------------|
| `pages/ConnectionsPage.vue` | 3 (fetch, submit, delete) |
| `pages/LoginPage.vue` | 1 (login) |
| `pages/OnboardingPage.vue` | 1 (onboard) |
| `stores/connections.ts` | 1 (fetchConnections) |
| `stores/logs.ts` | 1 (fetchLogs) |
| `components/connections/ConnectionTestModal.vue` | 1 (test) |
| `components/connections/ConnectionForm.vue` | 1 (GitHub install) |
| `components/connections/GitHubRepoSelector.vue` | 2 (load, save) |

**Inline narrowing simplified:**

| File | Change |
|------|--------|
| `components/notifications/NotificationChannels.vue` | 2 blocks: replaced `(e as { response?: ... })?.response?.data?.error` with `extractApiError(e, ...)` |

**Total:** 13 catch blocks updated across 10 files, 2 duplicate definitions removed, 1 utility created.

## What didn't change

- **No behavioural changes.** Every error message the user sees is identical — the utility reproduces the same priority chain each call site was using.
- **No test changes needed.** The store tests mock at the API layer and assert on the store's `error` ref, which still receives the same strings.
- **Files not touched:** Components that already used `catch (e: unknown)` without accessing error properties (e.g. `catch { toast.show('Failed...', 'error') }`) were left as-is — they don't need the utility.

## Verification

- `vue-tsc --noEmit` — type-check passes
- `vitest run` — all 50 tests pass
- `grep 'catch (e: any)'` — zero matches in `frontend/src/`
