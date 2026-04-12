# Add More Models — Phase 4 Completion Notes

**Scope:** Wire `ModelPicker` into `AgentConfigPage.vue`, replace the native `<select>`, and enrich the display-mode row (see `docs/executing/add-more-models.md` Phase 4).

**Branch:** `master`
**Validation:** `npx vue-tsc -b --noEmit` clean, `npx vite build` succeeds, `npx vitest run` 50/50 green.

---

## What shipped

### 1. Replaced `<select>` with `<ModelPicker>` in edit mode

**File:** `frontend/src/pages/AgentConfigPage.vue`

The old 15-line `<select>` block (two `<optgroup>`s cramming `name — ctx · $in/$out` into `<option>` labels) is replaced with a single line:

```vue
<ModelPicker v-model="formModel" :models="models" />
```

The `v-model` binding is unchanged — `formModel` is still the source of truth, and `formProvider` still derives from it via the catalogue lookup. The picker handles its own search, filtering, and grouping internally.

### 2. Deleted obsolete code

| Removed | Why |
|---------|-----|
| `formatModelLabel(m)` function | ModelCard renders its own structured layout — no need for a single-line label |
| `formatContext(tokens)` function | ModelCard has its own copy; the page doesn't need it separately |
| `anthropicModels` computed | ModelPicker groups by vendor, not provider — the page no longer groups |
| `openrouterModels` computed | Same |
| `<p>` helper text ("Context window · prompt $/M...") | All that information is now visible in the picker cards |

### 3. Enriched display-mode model row

The read-only display row now shows:
- **Model name** from the catalogue (not the raw ID) — via `selectedModelDisplay.name`
- **Tier badge** — same styling as the picker trigger (accent-bright rounded pill)
- **Provider label** — `direct` / `via OR` (replacing the old `via anthropic` / `via openrouter`)

**Graceful degradation for legacy models:** If the stored model ID doesn't match any catalogue entry, `selectedModelDisplay` is null — the template falls back to `config.model` (the raw ID string) and hides the tier badge. This handles rows with old IDs like `anthropic/claude-sonnet-4` without errors.

### 4. `selectedModelDisplay` computed

```typescript
const selectedModelDisplay = computed(() =>
  models.value.find(m => m.id === config.value?.model) ?? null
)
```

Resolves the stored config model to its full catalogue entry for the display row. Distinct from `formProvider` (which resolves from `formModel` during editing) — this one reads from the persisted `config.value.model` for read-only display.

---

## What was NOT changed

- **`formProvider` computed** — unchanged, still derives provider from `formModel` via the catalogue
- **`BaseSelect` import** — kept, still used for the Monitoring Mode dropdown
- **`saveConfig` / `fetchConfig`** — no changes, the wire format (`model` + `provider`) is identical
- **Backend** — no changes in this phase

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/pages/AgentConfigPage.vue` | Edit | +`ModelPicker` import, +`selectedModelDisplay` computed, replaced `<select>` with `<ModelPicker>`, enriched display row, removed `formatModelLabel`, `formatContext`, `anthropicModels`, `openrouterModels` |

---

## Validation results

```
$ npx vue-tsc -b --noEmit   → clean
$ npx vite build             → ✓ built in 1.06s (AgentConfigPage chunk: 15.59 kB)
$ npx vitest run             → 50/50 (7 files, 0 failures)
```

**Manual verification needed:** The plan doc calls for starting the dev server and cycling through one model from each tier. This should be done before merging — the type/build/test checks verify correctness, but only a browser test confirms the picker opens, filters work, selection persists, and the display row renders correctly after reload.

---

## What comes next

**Phase 5** — catalogue refresh tooling (`backend/cmd/refresh-models/main.go`)
**Phase 6** — validation pass, changelog entry, move the plan doc to completions
