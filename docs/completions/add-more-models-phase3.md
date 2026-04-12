# Add More Models — Phase 3 Completion Notes

**Scope:** Frontend `ModelPicker` and `ModelCard` components with search, tier filtering, vendor grouping, keyboard navigation, and vitest coverage (see `docs/executing/add-more-models.md` Phase 3).

**Branch:** `master`
**Validation:** `npx vue-tsc -b --noEmit` clean, `npx vitest run` 50/50 green (15 new ModelPicker tests + 35 existing).

---

## What shipped

### 1. `ModelCard.vue` — per-model row rendering

**File:** `frontend/src/components/agent/ModelCard.vue`

A presentational component that renders a single model entry as a three-line card:

- **Line 1:** Model name (truncated) + tier badge (colour-coded) + provider badge (`direct` / `via OR`)
- **Line 2:** One-line description (wraps instead of truncating — information is the point)
- **Line 3:** Context window + pricing + strength tags as inline pills

**Tier badge colours:**
| Tier | Colour scheme |
|------|---------------|
| Flagship | `status-warn` (amber) — premium, eye-catching |
| Balanced | `accent-bright` (feldgrau bright) — the default/recommended choice |
| Economy | `status-ok` (muted green) — safe, affordable |
| Specialist | `status-info` (blue) — distinct purpose |

These use the existing design token system — no new CSS variables. Each tier maps to a `border/text/bg` triplet via the `tierColors` record object.

**Props:** `model: ModelOption`, `selected: boolean`, `focused: boolean`
**Emits:** `select(id: string)`

The component has `role="option"` and `aria-selected` for listbox semantics, matching the plan doc's accessibility requirements.

### 2. `ModelPicker.vue` — combobox picker

**File:** `frontend/src/components/agent/ModelPicker.vue`

Replaces the native `<select>` with a custom combobox. The component is self-contained (no Pinia coupling) and follows `v-model` conventions: `modelValue` prop + `update:modelValue` emit.

**Trigger button:**
- Shows the selected model's name, tier badge, and pricing
- Truncates long names via `text-overflow: ellipsis` (enforced by `.truncate` + `overflow-hidden` + `min-w-0`)
- Width is fixed to the form column (the button is `w-full` within its parent)
- Shows "Select a model..." placeholder when the selected ID doesn't match any catalogue entry

**Picker panel:**
- Anchored below the trigger, `w-full` (never wider than trigger), `max-height: min(70vh, 32rem)` with `overflow-y: auto`
- Three sections:
  1. **Search input** — auto-focused on open. Substring matches across `name`, `id`, `vendor`, and `strengths` array. Case-insensitive.
  2. **Tier filter chips** — `All` / `Flagship` / `Balanced` / `Economy` / `Specialist`. Single-select, defaults to `All`.
  3. **Model list** — grouped by vendor with sticky section headers. Each entry renders via `ModelCard`.

**Grouping implementation:**
```
filteredModels → groupedModels (computed)
```
`groupedModels` is a `{ vendor: string, models: ModelOption[] }[]` built in a single pass with insertion-order preservation via `Map`. The flat filtered list is kept separately for keyboard navigation indexing.

**Keyboard handling:**
- `ArrowDown` / `ArrowUp` navigate the flat list
- `Enter` selects the focused model
- `Escape` closes the panel
- Arrow/Enter/Escape events from the search input bubble up to the root handler; all other keystrokes are stopped so typing doesn't trigger navigation

**Click outside:** Closes the panel. Uses `document.addEventListener('click', ...)` with cleanup in `onBeforeUnmount`, matching the pattern in `BaseSelect.vue`.

**Layout safety (meeting the plan doc's hard requirements):**
- Trigger width fixed by parent — long names truncate, never grow
- Panel width equals trigger (`w-full`) — never wider, never causes horizontal scroll
- Panel height capped at `min(70vh, 32rem)` — never pushes the page taller
- Long card text wraps — no hidden information
- No tooltips or hover-only affordances — everything visible without pointer interaction

### 3. Vitest tests — 15 cases

**File:** `frontend/src/components/agent/__tests__/ModelPicker.test.ts`

| Test | What it guards |
|------|----------------|
| `renders the selected model name in the trigger` | Trigger shows correct model + tier |
| `shows placeholder when no model selected` | Fallback text for unknown IDs |
| `opens the picker panel on trigger click` | Panel visibility toggle |
| `renders all models when no search or filter` | All 5 mock models appear |
| `filters by search substring across name` | "Opus" → 1 result |
| `filters by search substring across vendor` | "qwen" → 1 result |
| `filters by search substring across strength tags` | "diagnosis" → 1 result |
| `filters by tier` | Flagship chip → 2 results |
| `emits update:modelValue on card click` | v-model contract |
| `closes panel after selection` | UX: panel disappears on select |
| `shows empty state when filters exclude everything` | "No models match" message |
| `keyboard: arrow down then Enter selects` | Full keyboard flow |
| `keyboard: Escape closes the panel` | Escape key handler |
| `groups models by vendor` | Vendor section headers render |
| `renders long model names without overflow` | Truncation class present on trigger |

The tests use a 5-model mock catalogue covering all 4 tiers and 3 vendors. No real API calls — the component is pure props-in/events-out.

---

## Design decisions

**Why a new component instead of extending `BaseSelect`:** `BaseSelect` renders single-line text options in a flat list. The new picker has multi-line cards, a search input, tier filter chips, and vendor grouping — fundamentally different interaction. Retrofitting this onto `BaseSelect` would bloat the simple component that the rest of the UI relies on. The plan doc explicitly called this out: "should be a new component rather than hacking multi-line options onto `BaseSelect`."

**Why not a modal:** The plan doc's reasoning stands — a modal forces a focus trap and dims the page, which is overkill for "pick one value from a list." The anchored dropdown preserves surrounding context (the user can still see the other config fields).

**Why `data-index` for scroll-to-focused:** With vendor grouping, `ModelCard` elements are nested inside vendor `<div>`s — `listRef.children[i]` doesn't work because the children are group wrappers, not cards. `data-index` + `querySelector` addresses the right element regardless of nesting depth.

**Why search input stops propagation selectively:** The root `@keydown` handler captures arrow keys and Enter/Escape for navigation. Without `stopPropagation` on the search input, typing "e" would do nothing wrong (no handler matches), but if we later add `/` to focus search, we'd have a conflict. The selective approach is forward-safe: search swallows its own keystrokes, navigation keys bubble.

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/agent/ModelPicker.vue` | **New** | Combobox picker with search, tier filter, vendor grouping, keyboard nav |
| `frontend/src/components/agent/ModelCard.vue` | **New** | Per-model card row (3-line: name+badges, description, pricing+tags) |
| `frontend/src/components/agent/__tests__/ModelPicker.test.ts` | **New** | 15 vitest cases |

No backend changes. No existing file edits.

---

## Validation results

```
$ npx vue-tsc -b --noEmit                             → clean
$ npx vitest run                                      → 50/50 (7 files, 0 failures)
$ npx vitest run src/components/agent/__tests__/       → 15/15 PASS
```

---

## What comes next

**Phase 4** will wire the picker into `AgentConfigPage.vue`:
1. Replace the `<select>` block with `<ModelPicker v-model="formModel" :models="models" />`
2. Delete `formatModelLabel`, the helper text, and the provider-grouped computeds
3. Enrich the display-mode row with richer model info
