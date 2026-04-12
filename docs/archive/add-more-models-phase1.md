# Add More Models — Phase 1 Completion Notes

**Scope:** Catalogue curation, struct enrichment, consistency tests, and tool-use vetting test harness (see `docs/executing/add-more-models.md` Phase 1).

**Branch:** `master`
**Validation:** `go vet ./...` clean, `go build ./...` clean, `go test ./...` green (all standard tests pass, toolvet tests excluded via build tag as designed).

---

## What shipped

### 1. Enriched `ModelOption` struct

**File:** `backend/internal/agent/models.go`

Four new fields added to `ModelOption`:

```go
Vendor      string   `json:"vendor"`
Tier        string   `json:"tier"`
Strengths   []string `json:"strengths"`
Description string   `json:"description"`
```

**Why `Vendor` is distinct from `Provider`:** `Provider` is *how* we reach the model (`"anthropic"` direct SDK vs `"openrouter"` HTTP proxy), while `Vendor` is *who made it*. A model like `anthropic/claude-sonnet-4.6` routes through OpenRouter (`provider: "openrouter"`) but is vendored by Anthropic (`vendor: "anthropic"`). The Phase 3 picker groups by vendor, not provider — this separation makes that possible without ambiguity.

**Why constants for vendors and tiers:** `VendorAnthropic`, `TierFlagship`, etc. are exported `const` blocks. They prevent typo-drift across the catalogue and give future consumers (e.g. the `ResolveModel` helper in Phase 2) compile-time guarantees. The alternative — bare strings — would only fail at test time via `TestCatalogueConsistency`.

The change is fully additive — `GET /api/models` now returns richer JSON but every existing consumer ignores the new fields. No frontend changes needed until Phase 3.

### 2. Expanded catalogue: 3 Anthropic-direct + 19 OpenRouter = 22 models

**File:** `backend/internal/agent/models.go`

The old catalogue had 9 models (3 Anthropic + 6 OpenRouter). The new one has 22, covering 10 vendors across 4 tiers:

| Tier | Count | Models |
|------|-------|--------|
| Flagship | 5+1 | Claude Opus 4.6 (direct + OR), GPT-5.4, GPT-5.4 Pro, Gemini 3.1 Pro, Grok 4.20 |
| Balanced | 6+1 | Claude Sonnet 4.6 (direct + OR), GPT-5.4 Mini, Gemini 3 Flash, Qwen 3.6 Plus, GLM 4.7, MiniMax M2.7 |
| Economy | 4+1 | Claude Haiku 4.5 (direct), GPT-5.4 Nano, Gemini 3.1 Flash Lite, GLM 4.7 Flash, Qwen 3.5 Flash, MiMo V2 Flash |
| Specialist | 3 | Qwen 3 Coder, Kimi K2.5, Devstral |

**Dropped from the plan doc's proposed list (and why):**

- **`deepseek/deepseek-v3.2`** and **`deepseek/deepseek-chat-v3.1`** — not found on OpenRouter under standalone IDs as of 2026-04-12. Only third-party derivative/fine-tuned variants exist. Can be added later if DeepSeek re-publishes them.
- **`meta-llama/llama-4-maverick`** (was in old catalogue) — not in the top-tier research, historically patchy tool-use compliance. Dropped rather than carried forward.

**Old OpenRouter models that were replaced:**

| Old ID | Replacement | Reason |
|--------|-------------|--------|
| `anthropic/claude-sonnet-4` | `anthropic/claude-sonnet-4.6` | Model generation bump |
| `openai/gpt-4o` | `openai/gpt-5.4` | Next-gen successor |
| `openai/gpt-4o-mini` | `openai/gpt-5.4-mini` | Next-gen successor |
| `google/gemini-2.5-pro` | `google/gemini-3.1-pro-preview` | Next-gen successor |
| `google/gemini-2.5-flash` | `google/gemini-3-flash-preview` | Next-gen successor |
| `meta-llama/llama-4-maverick` | *(dropped)* | Not competitive |

**Legacy config impact:** Existing `app_agent_config` rows that reference old model IDs (e.g. `anthropic/claude-sonnet-4`) will continue to work on *read* — the provider receives whatever string is stored and OpenRouter may still route it. They will fail server-side validation in Phase 2 on the next *save*, prompting the user to pick a current model. This is the expected graceful-degradation path documented in the plan.

### 3. Pricing source of truth

**File:** `docs/archive/openrouter-models-2026-04-12.json`

A frozen snapshot of the OpenRouter API response for every selected model ID. Contains both per-token and per-million-token pricing for audit reproducibility. All prices in `models.go` are derived from this file — if a price looks wrong, check the archive first.

**Notable pricing observations vs. the plan doc:**
- Prices matched the plan's proposed values almost exactly. The plan doc's `~` (approximate) markers on DeepSeek were irrelevant since those models weren't available anyway.
- Kimi K2.5 pricing rounded from `$0.3827/M` to `$0.38/M` in the catalogue for readability. The archive preserves the exact per-token figure.

### 4. Catalogue consistency test

**File:** `backend/internal/agent/models_test.go`

Three standard (non-tagged) tests that run in CI on every commit:

| Test | What it guards |
|------|----------------|
| `TestCatalogueConsistency` | Every entry has all required fields populated (ID, Name, Provider, Vendor, Tier, Description, Strengths, ContextLength, Pricing). IDs are unique across the combined list. Provider is `anthropic` or `openrouter`. Tier is one of the four valid values. Description ≤ 150 chars. |
| `TestCatalogueSize` | Combined catalogue is between 15 and 35 entries. Lower bound catches accidental mass deletion; upper bound flags over-curation. |
| `TestDefaultModelInCatalogue` | `DefaultModelID` (`claude-sonnet-4-6`) actually exists in the catalogue. Prevents drift where the default gets renamed/removed but the constant isn't updated. |

**Why table-driven subtests (`t.Run(m.ID, ...)`):** When a model fails consistency, the test name directly identifies which entry is broken — e.g. `TestCatalogueConsistency/qwen/qwen3.6-plus` — instead of a line number in a loop body. This matters when the catalogue has 22 entries.

### 5. Tool-use vetting test

**File:** `backend/internal/agent/tool_use_vet_test.go`

Build-tagged integration test (`//go:build toolvet`) that runs the real agent loop against every catalogue entry with a trivial monitoring task.

**How it works:**
1. Registers a single `search_logs` tool definition
2. Sends a user message: "Investigate any recent errors in the api-gateway service"
3. When the model calls `search_logs`, returns canned log data (3 entries: 2 ERRORs + 1 WARN)
4. Expects the model to produce a final text answer summarising its findings
5. Asserts: ≥1 tool call made, final answer produced within 5 turns

**Design decisions:**

- **Build tag `toolvet`, not env-var guard:** A build tag is stronger than `t.Skip()` on a missing env var — the test binary doesn't even include the test code unless `-tags=toolvet` is passed. This means `go test ./...` in CI is guaranteed to skip it, even if someone accidentally sets `ANTHROPIC_API_KEY` in CI env.
- **Sequential subtests, not parallel:** Each subtest hits a live API. Running 22 models in parallel would trip rate limits on both Anthropic and OpenRouter. Sequential is slower (~3 min total) but reliable.
- **5-turn limit:** Most models should call `search_logs` once and answer on turn 2. The 5-turn budget is generous headroom for models that reason aloud first or make multiple search calls. Any model that doesn't converge in 5 turns is flagged as a vetting failure.
- **Canned data is a `TestToolUseVettingCannedResultIsValidJSON` sanity check** — also build-tagged, validates the JSON is parseable so we don't feed broken data to models and blame them for the failure.

**Run command:**
```bash
cd backend && go test -tags=toolvet -timeout=600s -v ./internal/agent/ -run TestToolUseVetting
```

**Cost:** ~$0.10 per full run across all 22 models (dominated by flagship entries; economy models are negligible).

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `backend/internal/agent/models.go` | Edit | Enriched `ModelOption` struct (4 new fields), vendor/tier constants, expanded catalogue from 9 → 22 entries |
| `backend/internal/agent/models_test.go` | **New** | 3 standard tests: consistency, size, default-model presence |
| `backend/internal/agent/tool_use_vet_test.go` | **New** | Build-tagged (`toolvet`) integration test — vets multi-turn tool use for every catalogue entry |
| `docs/archive/openrouter-models-2026-04-12.json` | **New** | Frozen OpenRouter API snapshot for pricing reproducibility |

No migrations. No DB schema changes. No frontend changes. No `provider_openrouter.go` edits.

---

## Validation results

```
$ go vet ./...                                        → clean
$ go build ./...                                      → clean
$ go test ./...                                       → all PASS (toolvet excluded by build tag)
$ go test -v ./internal/agent/ -run TestCatalogue     → 22/22 entries pass consistency
$ go test -v ./internal/agent/ -run TestDefaultModel  → DefaultModelID found in catalogue
```

The `toolvet` test has not been run against live APIs in this session (requires API keys). It is designed to be run manually before merging, as specified in the plan doc.

---

## What comes next

**Phase 2** will consume this catalogue:
1. Add `agent.ResolveModel(id) (ModelOption, bool)` helper
2. Wire server-side validation into `UpdateAppAgentConfig`
3. Mirror new `ModelOption` fields in `frontend/src/types/models.ts`
