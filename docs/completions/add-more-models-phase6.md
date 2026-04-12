# Add More Models — Phase 6 Completion Notes

**Scope:** Validation, changelog, and docs cleanup (see `docs/executing/add-more-models.md` Phase 6).

**Branch:** `master`

---

## What shipped

### 1. Changelog entry

**File:** `docs/changelog.md`

Added `0.33.0 — Expanded Model Catalogue & Picker (2026-04-12)` as the latest release. Covers all six phases: catalogue enrichment, server-side validation, ModelPicker component, AgentConfigPage wiring, refresh tooling.

### 2. Plan doc moved to completions

**File:** `docs/completions/add-more-models.md` (copied from `docs/executing/add-more-models.md`)

The original plan doc is now in the completions directory alongside the per-phase notes.

### 3. Full validation pass

All gates green:

```
Backend:
  go vet ./...          → clean
  go build ./...        → clean
  go test ./...         → all PASS

Frontend:
  npx vue-tsc -b --noEmit  → clean
  npx vite build            → ✓ built in 1.09s
  npx vitest run            → 50/50 (7 files, 0 failures)
```

### 4. `formProvider` legacy fallback verified

The plan doc flagged a specific check: `formProvider` must still derive correctly for legacy config rows whose model ID is not in the new catalogue. Verified by reading the code — the computed falls back to `config.value?.provider ?? 'anthropic'`, which still works. No code change needed.

---

## Manual testing remaining

The plan doc calls for a manual end-to-end test: start dev server, open Agent Config, cycle through one model from each tier via the new picker, save, reload, confirm round-trip. This should be done in a browser before shipping to production.

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `docs/changelog.md` | Edit | +0.33.0 entry |
| `docs/completions/add-more-models.md` | New (copy) | Plan doc moved from executing/ |
| `docs/completions/add-more-models-phase6.md` | New | This file |

---

## Summary of all phases

| Phase | Scope | Key deliverables |
|-------|-------|-----------------|
| 1 | Catalogue curation | 22-model catalogue, enriched `ModelOption`, archive JSON, vetting test |
| 2 | Backend validation | `ResolveModel` helper, server-side model + mismatch validation, 4 handler tests |
| 3 | Frontend ModelPicker | `ModelPicker.vue`, `ModelCard.vue`, 15 vitest tests |
| 4 | AgentConfigPage wiring | `<select>` → `<ModelPicker>`, enriched display row, deleted obsolete code |
| 5 | Refresh tooling | `cmd/refresh-models` CLI (--diff, --suggest) |
| 6 | Validation & docs | Changelog, plan doc move, full validation pass |
