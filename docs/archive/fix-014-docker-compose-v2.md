# Fix 014 — Makefile uses deprecated docker-compose

**Issue:** #14 from scaffolding-review.md
**Date:** 2026-02-19

## Problem

`Makefile` lines 51 and 54 used `docker-compose` (hyphenated), the legacy Python CLI deprecated in favour of `docker compose` (space-separated, Go-based Compose v2). Fails on systems with only Compose v2.

## Fix

Replaced `docker-compose` with `docker compose` in both the `docker-up` and `docker-down` targets.
