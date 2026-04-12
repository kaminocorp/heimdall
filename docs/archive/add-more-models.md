# Add More Models — Expand OpenRouter Catalogue & Rethink Model Selection

## Original brief

> We're using OpenRouter and the user can select models in Agent Configuration.
>
> Let's add more models than the current ones (incl. the most popular Qwen, GLM models etc.)
>
> Research/check which are the top 20 models on OpenRouter and let's pull those into the list. Perhaps let's also include some guidance for cost and what each model excels at into the UI (someplace sensible — the dropdown is a bit squeezed for that in its current form/shape, so perhaps we need to rethink the UI to make model selection more natural and easy to navigate).
>
> If all clear what I'm trying to achieve, think about how we can implement this enhancement/improvement, and outline your detailed proposal(s)/implementation plan with clearly defined implementation phases/tasks/steps in this md file.

---

## Current state (as of 2026-04-11)

- **Backend catalogue:** `backend/internal/agent/models.go` hard-codes 3 Anthropic models + 6 OpenRouter models (Claude Sonnet 4 via OR, GPT-4o, GPT-4o Mini, Gemini 2.5 Pro, Gemini 2.5 Flash, Llama 4 Maverick). All are deliberately curated, not fetched live — the file-level comment on `OpenRouterModels` (lines 51–58) explains why: OpenRouter exposes hundreds of models, many silently fail multi-turn tool use, and a bad model silently breaks the agent loop.
- **API surface:** `GET /api/models` (`backend/internal/api/handlers/models.go`) returns the concatenated list, gating OpenRouter entries behind `OPENROUTER_API_KEY`. The response shape is `ModelOption { id, name, provider, context_length, pricing }`.
- **Frontend UI:** `frontend/src/pages/AgentConfigPage.vue:152-174` renders a single native `<select>` with two `<optgroup>`s. Each `<option>` crams `name — context · $in/$out` into its label (`formatModelLabel`, line 46). There is no description, no "best-for" guidance, no grouping beyond provider, and no search. The helper text (line 172) just says "Context window · prompt $/M · completion $/M" — a legend for the squashed labels.
- **Provider is derived, not stored twice.** `formProvider` is a `computed` off the selected `formModel` via the catalogue (lines 35–39). Any richer picker can therefore change freely without touching the DB schema or the `UpdateAppAgentConfig` wire format — only the `model` ID needs to end up correct.
- **No server-side model validation.** `UpdateAppAgentConfig` (`backend/internal/api/handlers/applications.go:249-250`) only defaults empty `model` to `DefaultModelID`. Typos like `anthropic/claude-sonnet-5` save cleanly and fail at first invocation. Expanding the catalogue is the natural moment to close that hole.

## Goals

1. **Broaden the catalogue to ~20–25 models** covering the vendors the market is actually using in April 2026 — Anthropic, OpenAI, Google, DeepSeek, Qwen, GLM (z-ai), MiniMax, xAI, Moonshot, Mistral, Meta, Xiaomi — without abandoning the "vetted for tool use" principle.
2. **Give users meaningful selection guidance in the UI** — what each model excels at, price, context window, and a tier hint (flagship / balanced / economy / specialist) — without forcing them to read an external leaderboard.
3. **Make the picker navigable** — search, grouping, filtering — so scaling from 9 → 25 doesn't turn the dropdown into a wall of text.
4. **Validate at the API boundary** so adding models to the catalogue is the only way users can select them, and typos are rejected early.
5. **Make future catalogue updates cheap** — a small tooling + process change so bumping prices or adding a new model doesn't require re-reading this plan.

## Non-goals

- No live fetch of `GET https://openrouter.ai/api/v1/models` at request time. Curation stays manual; tooling just *helps* the curation.
- No per-user / per-org custom catalogues. Same list for everyone on a given deployment.
- No rewriting `provider_openrouter.go` — the wire shape already handles anything OpenRouter accepts. The only reason to touch it would be vendor-specific quirks uncovered during tool-use vetting (see Phase 1).
- No changes to the `provider` column in `app_agent_config`. It's still derived from the selected model at save time by the frontend, and passed through as-is.

---

## Research — top OpenRouter models (April 2026)

This list is assembled from OpenRouter's `GET /api/v1/models` endpoint (sampled April 2026), the public rankings page, and third-party usage summaries for April 2026. **It is intentionally a proposal, not a final catalogue** — every entry must be re-verified against live OpenRouter data at implementation time, both because prices drift and because some IDs rename (e.g. `claude-sonnet-4` → `claude-sonnet-4.6`).

### Shortlist by tier

Legend: `$in / $out` are USD per million tokens. `ctx` is the model's advertised context window. Prices below are rounded to 2 decimals where appropriate and reflect spot-checked April 2026 values — **the Phase 1 archive JSON is the authoritative source**, not this table.

#### Flagship (quality-first, premium price)

| Vendor | Proposed ID | Name | ctx | $in | $out | Strengths |
|---|---|---|---|---|---|---|
| Anthropic (direct) | `claude-opus-4-6` | Claude Opus 4.6 | 200k | 15 | 75 | SWE-bench leader, deep reasoning, deliberate diagnosis |
| Anthropic (OR) | `anthropic/claude-opus-4.6` | Claude Opus 4.6 (OR) | 1M | 5 | 25 | Same model via OR — cheaper but subject to OR markup changes |
| OpenAI (OR) | `openai/gpt-5.4` | GPT-5.4 | 1.05M | 2.5 | 15 | Balanced frontier model, strong general reasoning |
| OpenAI (OR) | `openai/gpt-5.4-pro` | GPT-5.4 Pro | 1.05M | 30 | 180 | Hardest cases; pricey — use sparingly |
| Google (OR) | `google/gemini-3.1-pro-preview` | Gemini 3.1 Pro | 1M | 2 | 12 | Huge context, strong agentic tool use |
| xAI (OR) | `x-ai/grok-4.20` | Grok 4.20 | 2M | 2 | 6 | Very large context, competitive coding/reasoning |

#### Balanced (default recommendations)

| Vendor | Proposed ID | Name | ctx | $in | $out | Strengths |
|---|---|---|---|---|---|---|
| Anthropic (direct) | `claude-sonnet-4-6` | Claude Sonnet 4.6 | 200k | 3 | 15 | **Default.** Near-Opus quality at Sonnet price, strong tool use |
| Anthropic (OR) | `anthropic/claude-sonnet-4.6` | Claude Sonnet 4.6 (OR) | 1M | 3 | 15 | Same model via OR — same price, 1M context |
| OpenAI (OR) | `openai/gpt-5.4-mini` | GPT-5.4 Mini | 400k | 0.75 | 4.5 | Cheap OpenAI reasoning, good day-to-day |
| Google (OR) | `google/gemini-3-flash-preview` | Gemini 3 Flash | 1M | 0.5 | 3 | Fast, cheap, near-Pro agentic capability |
| DeepSeek (OR) | `deepseek/deepseek-v3.2` | DeepSeek V3.2 | 164k | ~0.3 | ~1.2 | ~90% of GPT-5.4 at fraction of cost; thinking+tools integrated |
| Qwen (OR) | `qwen/qwen3.6-plus` | Qwen 3.6 Plus | 1M | 0.33 | 1.95 | 1M context, chain-of-thought, cheap Alibaba flagship |
| GLM (OR) | `z-ai/glm-4.7` | GLM 4.7 | 203k | 0.39 | 1.75 | Z.ai flagship, well-rounded, tool use is solid |
| MiniMax (OR) | `minimax/minimax-m2.7` | MiniMax M2.7 | 205k | 0.30 | 1.20 | Multimodal-leaning; competitive coding |

> Prices above marked `~` are approximate; DeepSeek in particular fluctuates heavily and must be re-checked at import time.

#### Economy (cheap, high-throughput — monitoring-loop candidates)

| Vendor | Proposed ID | Name | ctx | $in | $out | Strengths |
|---|---|---|---|---|---|---|
| Anthropic (direct) | `claude-haiku-4-5-20251001` | Claude Haiku 4.5 | 200k | 1 | 5 | Fast, cheap, Anthropic-native tool use |
| OpenAI (OR) | `openai/gpt-5.4-nano` | GPT-5.4 Nano | 400k | 0.20 | 1.25 | Ultra-cheap tiny reasoning |
| Google (OR) | `google/gemini-3.1-flash-lite-preview` | Gemini 3.1 Flash Lite | 1M | 0.25 | 1.50 | Cheapest Google option with 1M ctx |
| GLM (OR) | `z-ai/glm-4.7-flash` | GLM 4.7 Flash | 203k | 0.06 | 0.40 | Among cheapest tool-capable models |
| Qwen (OR) | `qwen/qwen3.5-flash-02-23` | Qwen 3.5 Flash | 1M | 0.07 | 0.26 | Dirt cheap, huge context, Chinese-origin fallback |
| Xiaomi (OR) | `xiaomi/mimo-v2-flash` | MiMo V2 Flash | 262k | 0.09 | 0.29 | High usage, cheap general-purpose |

#### Specialist (coding/agentic niches)

| Vendor | Proposed ID | Name | ctx | $in | $out | Strengths |
|---|---|---|---|---|---|---|
| Qwen (OR) | `qwen/qwen3-coder-next` | Qwen 3 Coder | 262k | 0.12 | 0.75 | Purpose-built coding model, cheap |
| Moonshot (OR) | `moonshotai/kimi-k2.5` | Kimi K2.5 | 262k | 0.38 | 1.72 | Dominates coding-focused workloads per leaderboards |
| Mistral (OR) | `mistralai/devstral-2512` | Devstral | 262k | 0.40 | 2.00 | Coding-specialised, European alt |
| DeepSeek (OR) | `deepseek/deepseek-chat-v3.1` | DeepSeek Chat v3.1 | 164k | ~0.27 | ~1.10 | General-purpose sibling of V3.2 |

**Size of final catalogue: ~25 models (3 Anthropic-direct + 22 OpenRouter).** Deliberately larger than "just 20" so we can drop 3–5 after tool-use vetting (Phase 1) without falling below the user's target.

### What we're *not* adding (and why)

- **Free tiers (`:free` suffix).** OpenRouter's free endpoints are rate-limited, have unpredictable latency, and are a poor fit for a monitoring loop that runs unattended. Economy-tier paid models are a better floor.
- **Embedding / audio / vision-only models.** Heimdall's agent loop is chat + tool-use; these would appear in the dropdown but break on first invocation.
- **`claude-opus-4.6-fast`**, **`grok-4.20-multi-agent`**, and other premium SKU variants. Same capability, higher price, no improvement for our use case.
- **Models without tool-use support.** Filtered out during Phase 1 vetting.

---

## Architectural decisions

### 1. Enrich `ModelOption`, don't replace it

Add four new fields. Every existing consumer (backend + frontend) keeps working because the fields are additive.

```go
// backend/internal/agent/models.go
type ModelOption struct {
    ID            string   `json:"id"`
    Name          string   `json:"name"`
    Provider      string   `json:"provider"`
    ContextLength int      `json:"context_length"`
    Pricing       Pricing  `json:"pricing"`

    // New fields — drive the rich picker UI.
    Vendor        string   `json:"vendor"`        // "anthropic" | "openai" | "google" | "deepseek" | "qwen" | "zai" | "minimax" | "xai" | "moonshot" | "mistral" | "meta" | "xiaomi"
    Tier          string   `json:"tier"`          // "flagship" | "balanced" | "economy" | "specialist"
    Strengths     []string `json:"strengths"`     // short tags, e.g. ["reasoning", "coding", "long-context"]
    Description   string   `json:"description"`   // one sentence, max ~120 chars
}
```

`Vendor` is distinct from `Provider`: **`provider`** is *how* we reach the model (`anthropic` direct vs. `openrouter`), while **`vendor`** is *who* made it. A model like `anthropic/claude-sonnet-4.6` routes via `openrouter` but is vendored by `anthropic` — that distinction matters for UI grouping.

The frontend `ModelOption` type in `frontend/src/types/models.ts` mirrors the same addition.

### 2. UI rethink: combobox picker, not `<select>`

Replace the native `<select>` in `AgentConfigPage.vue` with a **custom combobox component** (`ModelPicker.vue`). Specifics:

- **Trigger:** a read-only button styled like the current form inputs, showing the currently-selected model's name + tier badge + $/M prices. Click opens the picker panel.
- **Picker panel:** an anchored dropdown (not a full modal) that expands to the full form width. Contents:
  - **Search input** at the top. Substring match over `name`, `id`, `vendor`, `strengths`. Live-filters the list.
  - **Tier filter chips:** `All` / `Flagship` / `Balanced` / `Economy` / `Specialist` — single-select, defaults to `All`.
  - **List of cards**, grouped by vendor (collapsible section headers). Each card shows:
    - Line 1: **Name** + tier badge (colour-coded) + small provider badge (`direct` or `via OR`)
    - Line 2: one-line description
    - Line 3: `ctx · $in / $out · [strength-tag] [strength-tag]`
  - **Active row** is highlighted with the accent colour.
  - **Keyboard:** arrow keys navigate, Enter selects, Esc closes, `/` focuses search.
- **Empty state:** "No models match — try clearing filters."
- **Loading state:** skeleton row cards while `getAvailableModels()` is in flight.

**Layout safety (hard requirements — must be met, not nice-to-have):**

- The **trigger button** does not grow with the selected model's name. Long names are truncated via `text-overflow: ellipsis` with `overflow: hidden` and `min-width: 0` on the flex child. The button's width is fixed to the form column width regardless of the selection.
- The **panel width** equals the trigger width — never wider. Panel never expands the page horizontally and never triggers horizontal page scrolling. Enforced with `max-width: 100%` and `box-sizing: border-box`.
- The **panel height** is capped at `min(70vh, 32rem)` with internal `overflow-y: auto` — never pushes the page taller and never clips the last row at the viewport edge. If there are more models than fit, the user scrolls the panel, not the page.
- **Long card text** (description, strength tags) wraps to a second visual line instead of ellipsis-truncating — information is the whole point of the redesign, hiding it defeats the exercise. The card's fixed structure (3 logical lines) absorbs the wrap without pushing into the next card.
- **Panel positioning** flips above the trigger when there isn't enough viewport space below (standard popper behavior). If both fail (tiny viewport), the panel becomes a bottom-sheet on narrow screens — but this is a fallback, not the default.
- **No information is hidden inside a tooltip or hover-only affordance.** Everything that matters (price, context, description) is visible without pointer interaction. Tooltips are for optional detail like full strength-tag explanations.
- **Vitest test `renders long model names without overflow`** asserts the trigger button's clientWidth equals its parent's, even when the selected model's name is 80 chars long.

The picker is contained to the Agent Config form — it is *not* a site-wide component library addition. If a second use-case appears later, we promote it.

**Why not a modal?** A modal forces a focus trap and dims the rest of the page, which is overkill for "pick one value from a list." An anchored dropdown preserves context.

**Why not `BaseSelect`?** The existing `frontend/src/components/common/BaseSelect.vue` is a single-line styled `<select>` replacement. The new picker is a fundamentally different interaction (search + multi-line cards), so it should be a new component rather than hacking multi-line options onto `BaseSelect`.

### 3. Server-side model validation

`UpdateAppAgentConfig` should reject any `model` that is not in the concatenated catalogue (with OpenRouter gated by `OPENROUTER_API_KEY` as today). Implementation: a tiny helper `agent.ResolveModel(id string) (ModelOption, bool)` that scans `AnthropicModels` + `OpenRouterModels`. The handler calls it after normalising provider.

Benefit: one source of truth, no silent typo-breakage, tests can assert catalogue consistency.

Risk: users with legacy config rows pointing at removed models would fail validation on *update*, but not on *read* (we still return the row as-is). Acceptable — if they hit Edit, they pick a current model; we don't force-migrate.

### 4. Catalogue refresh tooling (light-touch)

A small Go program at `backend/cmd/refresh-models/main.go` that:

- Hits `GET https://openrouter.ai/api/v1/models`
- For every ID in `OpenRouterModels`, compares `context_length` and `pricing` and prints a diff
- Optionally: a `--suggest` flag that lists the current top-20 by vendor, so when we next curate we don't forget a new release.

It is **not** run at request time and **not** wired into CI. It's a human-operated tool the developer runs before editing `models.go`. Keeps Phase 1 curation honest without encroaching on runtime behaviour.

---

## Implementation phases

### Phase 1 — Catalogue curation & tool-use vetting

Upfront work; produces the final Go catalogue the rest of the phases depend on.

1. **Fetch live OpenRouter catalogue.** Run `curl https://openrouter.ai/api/v1/models | jq` once and archive the JSON to `docs/archive/openrouter-models-2026-04-11.json` for reproducibility of pricing decisions.
2. **Build the proposed catalogue as a draft** following the tier tables above. Lift real IDs, context lengths, and prices from the archived JSON.
3. **Write the build-tagged tool-use vetting test.** New file `backend/internal/agent/tool_use_vet_test.go` with `//go:build toolvet` at the top. The test iterates every entry in `AnthropicModels` + `OpenRouterModels`, constructs a minimal agent with a fake `search_logs` tool that returns a canned result, and sends a task crafted to require **at least two tool calls** before the model concludes. Each model is a table-driven subtest (`t.Run(model.ID, ...)`) so failures name the broken model directly. Runs only when `-tags=toolvet` is passed — CI skips it by default so it never runs without explicit intent. Requires `ANTHROPIC_API_KEY` + `OPENROUTER_API_KEY` in the environment; skips cleanly if absent. Cost per full run: ~$0.10.
4. **Run the vetting test** once against the proposed catalogue. Drop any failing entries from `models.go` with a one-line PR-body note (`qwen/qwen3.5-flash-02-23 — tool_use returned empty tool_calls on turn 2, dropped`).
5. **Write the catalogue** into `backend/internal/agent/models.go` with the enriched `ModelOption` fields (Vendor, Tier, Strengths, Description). Add a `tool_use_verified_at` date comment above each entry so future curators see when it was last vetted.

**Deliverable:** updated `models.go`, archived models JSON, `tool_use_vet_test.go`, vetting run log in the PR body.
**Validation gate:**
- `go vet ./...` + `go test ./internal/agent/...` clean (standard CI path, skips `toolvet`).
- `go test -tags=toolvet ./internal/agent/...` green against a live API with every catalogue entry passing.
- Every `ModelOption` entry has all required fields populated (asserted by a separate non-tagged `TestCatalogueConsistency`).

### Phase 2 — Backend schema + validation

1. **Extend the `ModelOption` struct** (new fields). Add a `Vendor` constant block for the fixed list.
2. **New helper `agent.ResolveModel(id string) (ModelOption, bool)`** — scans both lists.
3. **`UpdateAppAgentConfig` validation.** After provider normalisation (`applications.go:275`), call `ResolveModel`. If not found, 400 with `"unknown model: <id>"`. Also cross-check that the resolved model's `Provider` matches `req.Provider` (reject mismatch: user sent `{provider: "anthropic", model: "openai/gpt-5.4"}`).
4. **Handler test coverage:**
   - `TestUpdateAppAgentConfig_RejectsUnknownModel`
   - `TestUpdateAppAgentConfig_RejectsProviderModelMismatch`
   - `TestUpdateAppAgentConfig_AcceptsFlagshipModel`
   - `TestUpdateAppAgentConfig_AcceptsEconomyModel`
5. **Frontend types** — update `frontend/src/types/models.ts` to mirror the new fields. Stub their use in `AgentConfigPage.vue` temporarily so the existing dropdown keeps working until Phase 4 lands.

**Validation gate:** `go test ./internal/api/handlers/... -run TestUpdateAppAgentConfig` green. `GET /api/models` returns the enriched shape.

### Phase 3 — Frontend `ModelPicker` component

1. **New file `frontend/src/components/agent/ModelPicker.vue`.** Props: `modelValue` (selected ID), `models` (array). Emits: `update:modelValue`. Internal state: `open`, `search`, `tierFilter`. No Pinia coupling.
2. **`ModelCard.vue`** — one component per entry, keeping `ModelPicker.vue` focused on layout + filtering.
3. **Keyboard handling** — arrow-key navigation, Enter to select, Esc to close, `/` to focus search. Implemented via a single `@keydown` handler on the panel wrapper.
4. **Accessibility:** `role="listbox"` on the panel, `role="option"` on cards, `aria-activedescendant` tracking the keyboard cursor.
5. **Vitest tests** at `frontend/src/components/agent/__tests__/ModelPicker.test.ts`:
   - renders all models when no search/filter
   - filters by search substring across name, vendor, strength tags
   - filters by tier
   - emits `update:modelValue` on card click
   - keyboard navigation: arrow down → Enter selects the highlighted model
   - empty state when filters exclude everything

**Validation gate:** `cd frontend && npm run test` green. Component renders in isolation via a dev scratch page.

### Phase 4 — Wire `ModelPicker` into `AgentConfigPage`

1. **Replace the `<select>` block** (`AgentConfigPage.vue:154-174`) with `<ModelPicker v-model="formModel" :models="models" />`.
2. **Delete `formatModelLabel`** (line 46) and the helper text below the select (lines 171–173) — all that information now lives inside the picker cards.
3. **Update the display-mode row** (lines 247–253) to show the richer model info: name, tier badge, vendor-via-provider. Either inline or via a minimal `ModelSummary.vue`.
4. **Delete `anthropicModels` / `openrouterModels` computeds** — irrelevant once the picker handles its own grouping.

**Validation gate:** `cd frontend && npm run build && npx vue-tsc -b --noEmit` clean. Manual test: pick one model from each tier, save, reload, confirm the selection round-trips.

### Phase 5 — Catalogue refresh tool

1. **New `backend/cmd/refresh-models/main.go`** — ~150 LoC. Uses `net/http` + `encoding/json`, no SDK.
2. **Flags:** `--diff` (default) compares curated vs. live, prints a table; `--suggest` prints top models by vendor the curator hasn't yet adopted.
3. **Docs:** short `backend/cmd/refresh-models/README.md` explaining when to run it (before editing `models.go`) and how to interpret the output.

**Validation gate:** `go run ./cmd/refresh-models --diff` against the current catalogue prints a clean or actionable diff.

### Phase 6 — Validation + docs

1. **End-to-end manual test.** Start the backend + frontend, create a new app, cycle through one model from each tier via the new picker, run an interactive chat turn to confirm tool use works. The Phase 1 `toolvet` test already covers the protocol-level tool-use check; this manual pass is specifically a UI + round-trip sanity check (picker opens, selection saves, reload shows the right row, agent chat responds).
2. **Changelog entry** for the release (0.32.0 or 0.31.1 depending on scope framing).
3. **Update `docs/vision.md`** only if the rethought Agent Config section deserves a mention — otherwise leave alone.
4. **Move this doc** to `docs/completions/add-more-models.md` once done.

**Validation gate:** all of the above, plus spot-check that `formProvider` still derives correctly for legacy config rows whose model ID is not in the new catalogue (the computed falls back to `config.value?.provider`, which still works).

---

## Files touched (estimated)

| File | Kind | Phase | Why |
|---|---|---|---|
| `backend/internal/agent/models.go` | Edit | 1, 2 | Enriched struct + ~25-entry catalogue |
| `backend/internal/agent/models_test.go` | **New** | 2 | Catalogue consistency assertions (unique IDs, all fields populated) |
| `backend/internal/agent/tool_use_vet_test.go` | **New** | 1 | Build-tagged (`toolvet`) vetting test — runs real agent loop against every entry |
| `backend/internal/api/handlers/applications.go` | Edit | 2 | Server-side model validation |
| `backend/internal/api/handlers/applications_test.go` | Edit | 2 | +4 validation cases |
| `backend/cmd/refresh-models/main.go` | **New** | 5 | Catalogue drift detector |
| `backend/cmd/refresh-models/README.md` | **New** | 5 | Usage notes |
| `frontend/src/types/models.ts` | Edit | 2 | Mirror new fields |
| `frontend/src/components/agent/ModelPicker.vue` | **New** | 3 | Search + filter + list combobox |
| `frontend/src/components/agent/ModelCard.vue` | **New** | 3 | Per-model row rendering |
| `frontend/src/components/agent/__tests__/ModelPicker.test.ts` | **New** | 3 | Vitest coverage |
| `frontend/src/pages/AgentConfigPage.vue` | Edit | 4 | Swap `<select>` for `<ModelPicker>`, enrich display mode |
| `docs/archive/openrouter-models-2026-04-11.json` | **New** | 1 | Frozen catalogue snapshot for reproducibility |
| `docs/changelog.md` | Edit | 6 | Release entry |
| `docs/completions/add-more-models.md` | **Move** | 6 | This file, post-completion |

No migrations. No DB schema changes. No `provider_openrouter.go` edits unless a tool-use failure forces a workaround.

---

## Risks & mitigations

| Risk | Mitigation |
|---|---|
| **Tool-use silently breaks** on a newly-added model in production. | Phase 1 smoke test is the gate; anything that fails is dropped before merge. Add a `tool_use_verified_at` comment in `models.go` per entry so the next curator sees the vetting date. |
| **Pricing drifts** and the UI shows stale numbers. | Phase 5 refresh tool runs on demand; schedule a quarterly reminder (or wire into `/loop`) to run it. |
| **Model IDs get renamed** on OpenRouter (e.g. `claude-sonnet-4` → `claude-sonnet-4.6`). | Server validation rejects unknown IDs immediately, so the first affected request surfaces a loud 400 instead of a silent 500 at agent call time. Diff tool catches renames before users hit them. |
| **Picker UI performance** with ~25 models + live search. | 25 rows is trivially fast. No virtual scrolling needed until we cross ~200. |
| **Vendors with poor tool-use compliance** (historically some Llama 3.x and Qwen variants). | Vetting filters them. A few economy picks (MiMo Flash, Qwen 3.5 Flash) may not make the cut — that's fine, the catalogue stays honest. |
| **Legacy config rows** pointing at the old 6-entry OR list (`anthropic/claude-sonnet-4`, `google/gemini-2.5-pro`, etc.) | Keep them in the new catalogue as aliases where the model still exists, or accept that they'll 400 on next *save*. Read path is unaffected — the display mode shows whatever string is stored. |

---

## Decisions (resolved 2026-04-11)

All open questions answered — these are the rules the implementation follows.

1. **Catalogue size: ~25 is fine.** Proceed with 3 Anthropic-direct + ~22 OR. Drop entries that fail Phase 1 vetting rather than padding to a round number.

2. **Tier labels: Flagship / Balanced / Economy / Specialist.** Kept the original proposal — reads cleanly and encodes both price tier and intent (Specialist ≠ Economy even when cheap).

3. **Anchored dropdown — confirmed.** Explicit layout safety requirements captured in Phase 3 below (no page stretching, no clipped content, no hidden overflow). These are acceptance criteria, not nice-to-haves.

4. **Keep both Anthropic-direct and via-OR rows** with clear labels. Direct entries render as e.g. `Claude Sonnet 4.6 · direct` and OR entries render as `Claude Sonnet 4.6 · via OpenRouter`. The "(OR)" suffix from the draft is replaced by a small provider badge on the card so the label text stays short.

5. **Strict server-side validation.** Unknown model IDs are rejected at the API boundary with 400 regardless of provider. If OpenRouter ships a model we haven't curated, users wait for the next catalogue bump.

6. **Manual refresh CLI only.** No scheduled investigation, no Slack pings, no CI wiring. Developer runs `go run ./cmd/refresh-models --diff` before editing `models.go`.

7. **Automated tool-use vetting test — included.** A build-tagged integration test (`go test -tags=toolvet ./internal/agent/...`) runs the real agent loop against every catalogue entry with a trivial two-tool-call task. Cost is ~$0.10/run, it's opt-in (build tag keeps it out of CI), and it converts Phase 1's manual smoke test into a reproducible gate future curators can re-run. Worth it.

8. **Default stays `claude-sonnet-4-6` (Anthropic direct).** No change to `agent.DefaultModelID`.

## Sources

Used during research:

- [LLM Rankings | OpenRouter](https://openrouter.ai/rankings)
- [OpenRouter Rankings April 2026: Top AI Models by Data — DigitalApplied](https://www.digitalapplied.com/blog/openrouter-rankings-april-2026-top-ai-models-data)
- [Top AI Models on OpenRouter (March 2026) — TeamDay.ai](https://www.teamday.ai/blog/top-ai-models-openrouter-2026)
- [Best AI Models for Coding | OpenRouter](https://openrouter.ai/collections/programming)
- [Best AI Models April 2026: Ranked by Benchmarks — BuildFastWithAI](https://www.buildfastwithai.com/blogs/best-ai-models-april-2026)
- [State of AI 2025: 100T Token LLM Usage Study — OpenRouter](https://openrouter.ai/state-of-ai)
- `GET https://openrouter.ai/api/v1/models` — live catalogue (sampled 2026-04-11)
