# Fix 012 — ESLint TypeScript parser

**Issue:** #12 from scaffolding-review.md
**Date:** 2026-02-19

## Problem

`frontend/eslint.config.js` applied rules to `*.ts` and `*.vue` files but had no TypeScript parser configured. ESLint would fail on type annotations, treating TS syntax as invalid JS.

## Fix

- Installed `typescript-eslint` as a dev dependency.
- Replaced the manual file-matching block with `tseslint.configs.recommended` (handles `*.ts` files with the correct parser).
- Added a `*.vue` override that sets `tseslint.parser` as the `parserOptions.parser`, so `vue-eslint-parser` (bundled by `eslint-plugin-vue`) delegates `<script lang="ts">` blocks to the TS parser.

## Verification

- `npx eslint src/` passes clean.
- `vue-tsc --noEmit` still passes.
