# NotificationsPage Decomposition — Phase 1

**Date:** 2026-04-12
**Assessment reference:** `docs/plans/code-assessment-2026-04-12.md`, issue #4

## Problem

`frontend/src/pages/NotificationsPage.vue` was **541 lines** — the only frontend file exceeding the 500-line threshold. It handled four distinct concerns in a single component:

1. **Preferences** — display + inline edit form with toggle, dropdown, and number input
2. **Channels** — full CRUD (add/edit/delete) with a type-switching form (webhook URL vs email recipients)
3. **Channel testing** — test button with loading state per channel
4. **Notification history** — read-only list with severity indicators, status badges, and relative timestamps

All state (15 `ref`s), all API calls (8 imported functions), and all helper functions (5) were mixed together in a single `<script setup>` block.

## Solution

Decomposed into three sub-components under `frontend/src/components/notifications/`, with the page becoming a thin data-fetching shell.

### File breakdown

| File | Lines | Responsibility |
|------|-------|----------------|
| `NotificationsPage.vue` | 102 | Page header, skeleton loader, data fetching, composition |
| `NotificationPreferences.vue` | 151 | Preferences display/edit form, save logic |
| `NotificationChannels.vue` | 294 | Channel list, add/edit/delete form, test button |
| `NotificationHistory.vue` | 58 | Read-only notification log with severity/status indicators |

### Architecture

**Data flow:** The page fetches all three data sets in parallel via `Promise.all` (preserving the original batched fetch for efficiency), then passes data down as props. Sub-components emit `update:preferences` and `update:channels` events to propagate mutations back up.

**State reset on app switch:** The page holds template refs to `NotificationPreferences` and `NotificationChannels`, calling their exposed `resetEditing()` / `resetForm()` methods when `currentAppId` changes. This preserves the original behaviour of closing open forms when the user switches applications.

**API call ownership:** Each sub-component imports and calls its own mutation APIs (save, delete, test) directly — avoiding prop/emit ping-pong for actions. The page only handles the initial batch fetch.

### Additional fix

Replaced `catch (e: any)` (2 occurrences in the original page) with `catch (e: unknown)` and proper type narrowing in `NotificationChannels.vue`:

```typescript
// Before (in original page)
catch (e: any) {
  const msg = e?.response?.data?.error || 'Failed to save channel'

// After (in NotificationChannels.vue)
catch (e: unknown) {
  const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Failed to save channel'
```

This addresses assessment issue #7 (untyped error catches) for this component.

## What didn't change

- **No visual changes.** The rendered HTML, CSS classes, and user interactions are identical.
- **No API changes.** Same endpoints called with same payloads.
- **No routing changes.** The page component path and route registration are unchanged.
- **Batch fetch preserved.** The page still loads preferences, channels, and history in a single `Promise.all`, not three sequential fetches.

## Verification

- `vue-tsc --noEmit` — type-check passes
- `vitest run` — all 50 tests pass (7 test files)
- No `catch (e: any)` in the new components
