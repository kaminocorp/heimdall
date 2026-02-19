# Fix 022 — Axios client auth interceptor

**Issue:** #22 from scaffolding-review.md
**Date:** 2026-02-20

## Problem

The axios client never attached the auth token to outgoing requests. API calls would fail authentication even when the user was logged in.

## Fix

Added a request interceptor that reads the token from `useAuthStore()` and sets the `Authorization: Bearer <token>` header on every request.

## Verification

- `vue-tsc --noEmit` passes clean.
