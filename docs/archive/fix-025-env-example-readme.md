# Fix 025 — Add .env.example and README prerequisites

**Issue:** #25 from scaffolding-review.md
**Date:** 2026-02-20

## Problem

`docker-compose.yml` references `env_file: .env` but no `.env.example` documented the required variables. README had no prerequisites section and no database setup instructions.

## Fix

- Created `.env.example` listing all env vars (required and optional) with the docker-compose Postgres defaults pre-filled.
- Updated `README.md` with a prerequisites list (Go, Node, Docker, migrate, sqlc) and step-by-step setup instructions (env, Postgres, migrations, npm install, dev servers).
