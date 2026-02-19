# Fix 023 — ChatWindow uses index as v-for key

**Issue:** #23 from scaffolding-review.md
**Date:** 2026-02-20

## Problem

`ChatWindow.vue` used array index as the `v-for` key (`(msg, i) in messages :key="i"`). Index keys cause rendering bugs when the message list changes (e.g. messages inserted, reordered, or removed).

## Fix

- Added `id: string` field to the `ChatMessage` type.
- Updated `useAgent.ts` to assign `crypto.randomUUID()` when creating messages (uses server-provided `id` if available, falls back to client-generated).
- Changed `ChatWindow.vue` to use `:key="msg.id"`.

## Verification

- `vue-tsc --noEmit` passes clean.
