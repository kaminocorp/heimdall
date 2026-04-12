# Add More Models — Phase 2 Completion Notes

**Scope:** Backend schema enrichment, server-side model validation, and frontend type mirroring (see `docs/executing/add-more-models.md` Phase 2).

**Branch:** `master`
**Validation:** `go vet ./...` clean, `go build ./...` clean, `go test ./...` green, `npx vue-tsc -b --noEmit` clean.

---

## What shipped

### 1. `agent.ResolveModel` helper

**File:** `backend/internal/agent/models.go`

```go
func ResolveModel(id string) (ModelOption, bool)
```

A simple linear scan across `AnthropicModels` + `OpenRouterModels`. Returns the full `ModelOption` and `true` if the ID exists in either catalogue, or zero value and `false` otherwise.

**Why linear scan instead of a map:** The catalogue has 22 entries. A `map[string]ModelOption` would save ~20 ns per lookup on a path that runs once per config-save request. The map would need initialisation (either `init()` or `sync.Once`), adds a second source of truth to keep in sync with the slice, and gains nothing measurable. If the catalogue grows past ~100 entries, revisit — but the plan doc caps it at ~35.

**Test coverage:** `TestResolveModel` in `backend/internal/agent/models_test.go` covers four cases:
- Known Anthropic-direct model → resolves with correct provider
- Known OpenRouter model → resolves with correct provider
- Unknown model → returns false
- Empty string → returns false

### 2. Server-side model validation in `UpdateAppAgentConfig`

**File:** `backend/internal/api/handlers/applications.go`

Two new validation checks added after provider normalisation and before the `UpsertAppAgentConfig` call:

1. **Unknown model rejection:** `agent.ResolveModel(req.Model)` — if the model ID is not in the catalogue, returns `400` with `"unknown model: <id>"`. This closes the hole identified in the plan doc where typos like `anthropic/claude-sonnet-5` would save cleanly and fail silently at first agent invocation.

2. **Provider/model mismatch rejection:** `resolved.Provider != req.Provider` — catches requests like `{provider: "anthropic", model: "openai/gpt-5.4"}`. Returns `400` with `"model <id> belongs to provider <x>, not <y>"`. This prevents a class of silent misconfiguration where the frontend sends an inconsistent pair and the backend dutifully stores it.

**Ordering of checks:** The validation runs *after* the provider gate (which checks `OPENROUTER_API_KEY` availability) and *after* the empty-model default (`req.Model = DefaultModelID`). This means:
- An empty model always resolves to `claude-sonnet-4-6` (which exists) → no false rejection
- An OpenRouter model with no key configured → hits the "openrouter not enabled" error *before* reaching model validation, which is the more specific and actionable error

**Legacy config rows:** Existing rows with old model IDs (e.g. `anthropic/claude-sonnet-4` from the pre-Phase-1 catalogue) are unaffected on *read*. The validation only fires on *write* (`PUT /api/apps/{appId}/agent/config`). If a user opens the Agent Config page with a legacy model, the display mode shows whatever string is stored. When they click Edit and Save, the new validation catches the stale ID and they must pick a current model. This is the expected graceful-degradation path.

### 3. Handler tests — 4 new cases

**File:** `backend/internal/api/handlers/applications_test.go`

| Test | What it guards |
|------|----------------|
| `TestUpdateAppAgentConfig_RejectsUnknownModel` | Sends `anthropic/claude-sonnet-99` → 400 with "unknown model" |
| `TestUpdateAppAgentConfig_RejectsProviderModelMismatch` | Sends `openai/gpt-5.4` with `provider: "anthropic"` → 400 with "belongs to provider" |
| `TestUpdateAppAgentConfig_AcceptsFlagshipModel` | `claude-opus-4-6` with `provider: "anthropic"` → 200, persists correctly |
| `TestUpdateAppAgentConfig_AcceptsEconomyModel` | `claude-haiku-4-5-20251001` with `provider: "anthropic"` → 200, persists correctly |

**Why the mismatch test uses an OpenRouter model with provider "anthropic":** The test environment doesn't set `OPENROUTER_API_KEY`, so sending `provider: "openrouter"` would hit the "not enabled" gate before reaching model validation. By sending an OpenRouter-catalogue model (`openai/gpt-5.4`) with `provider: "anthropic"` (which passes the provider gate), the request reaches `ResolveModel` and the cross-check catches the mismatch. This tests the exact code path we care about.

**All tests are integration tests** that require `DATABASE_URL` and `SUPABASE_URL` — they skip cleanly without those env vars, matching the existing test pattern.

### 4. Frontend types — mirrored enrichment

**File:** `frontend/src/types/models.ts`

Four new fields added to the `ModelOption` interface:

```typescript
vendor: string
tier: 'flagship' | 'balanced' | 'economy' | 'specialist'
strengths: string[]
description: string
```

The `tier` field uses a union type matching the Go constants, so TypeScript catches invalid tier values at compile time. `vendor` stays `string` (not a union) because the set of vendors may grow without frontend changes — the picker (Phase 3) will group dynamically, not via a hardcoded switch.

**No functional changes to `AgentConfigPage.vue`** — the existing `<select>` dropdown and `formatModelLabel` helper continue to work. They simply ignore the new fields until Phase 4 wires in the `ModelPicker` component.

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `backend/internal/agent/models.go` | Edit | +`ResolveModel` helper |
| `backend/internal/agent/models_test.go` | Edit | +`TestResolveModel` (4 cases) |
| `backend/internal/api/handlers/applications.go` | Edit | +model validation + provider/model cross-check in `UpdateAppAgentConfig` |
| `backend/internal/api/handlers/applications_test.go` | Edit | +4 handler tests |
| `frontend/src/types/models.ts` | Edit | +4 enriched fields mirroring Go struct |

No migrations. No DB schema changes. No new files.

---

## Validation results

```
$ go vet ./...                                        → clean
$ go build ./...                                      → clean
$ go test ./...                                       → all PASS
$ go test -v ./internal/agent/ -run TestResolveModel  → PASS
$ npx vue-tsc -b --noEmit                             → clean (no type errors)
```

---

## What comes next

**Phase 3** will build the `ModelPicker` component:
1. New `frontend/src/components/agent/ModelPicker.vue` — combobox with search, tier filtering, vendor grouping
2. New `frontend/src/components/agent/ModelCard.vue` — per-model row rendering
3. Vitest tests for filtering, selection, and keyboard navigation
