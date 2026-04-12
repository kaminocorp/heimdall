# OpenRouter Multi-Provider — Code Assessment (v0.29.0)

**Scope:** Phases 1–8 of the OpenRouter integration (`docs/completions/openrouter-phase-*.md`).
**Bar:** 8.5/10 production readiness.
**Verdict:** **9/10 — ship it.** No blocking issues. A small number of low-severity nits are listed at the bottom; none need to land before the merge.

---

## Files reviewed

| File | LOC | Status |
|---|---|---|
| `backend/internal/agent/provider.go` | 95 | Clean |
| `backend/internal/agent/provider_anthropic.go` | 147 | Clean |
| `backend/internal/agent/provider_openrouter.go` | 350 | Clean |
| `backend/internal/agent/provider_openrouter_test.go` | 230 | Good coverage |
| `backend/internal/agent/agent.go` | 96 | Clean |
| `backend/internal/agent/loop.go` | 360 | Clean, under the 500-LOC refactor line |
| `backend/internal/agent/tools.go` | 91 | Clean |
| `backend/internal/agent/models.go` | 102 | Clean |
| `backend/internal/api/handlers/applications.go` | 303 | Clean |
| `backend/internal/api/handlers/models.go` | 23 | Clean |
| `backend/internal/config/config.go` (OpenRouter bits) | — | Clean |
| `backend/migrations/022_agent_config_provider.{up,down}.sql` | — | Correct, symmetric |
| `frontend/src/pages/AgentConfigPage.vue` | 270 | Clean |
| `frontend/src/{api,types}/models.ts` | tiny | Matches backend contract |

**No files flagged for refactor.** The biggest candidate (`loop.go` at 360) is well within the 500-LOC guideline and has clear sections (blocking loop, monitoring loop, helpers) that make it readable as one file.

---

## What the review confirmed is correct

### 1. Anthropic ↔ OpenRouter translation is faithful

Cross-checked `buildOpenRouterRequest` / `translateOpenRouterResponse` against the OpenAI chat completions spec on each of the five divergence points called out in the Phase 2 completion doc:

| Divergence | Anthropic | OpenRouter translator | Correct? |
|---|---|---|---|
| **System prompt** | Top-level `system` field | First `role:"system"` message | Yes — `buildOpenRouterRequest` L184–189 |
| **Tool schema** | Inner properties bag | Wrapped in full JSON Schema `{type:"object", properties, required}` | Yes — L261–279 |
| **Tool calls (response)** | `tool_use` content blocks | Flattened `tool_calls[]` → `[]ContentBlock` | Yes — L317–332 |
| **Tool results (request)** | One `user` message w/ N `tool_result` blocks | N separate `role:"tool"` messages | Yes — L194–208 |
| **Stop reason** | `end_turn` / `tool_use` | `tool_calls` → `StopReasonToolUse`; everything else → `EndTurn` | Yes — L301–305; matches the Anthropic provider's own fallthrough at `provider_anthropic.go:123–128` |

Two subtleties are handled correctly:

- **`function.arguments` is a JSON-encoded string**, not an object. The builder stringifies `json.RawMessage` at L233; the translator round-trips the string back to `json.RawMessage` at L320 so `loop.go:154`'s `json.Unmarshal(tu.Input, ...)` works unchanged. Empty-string edge case falls back to `"{}"` on both sides (L235, L322).
- **`is_error` has no OpenAI equivalent.** The `"ERROR: "` content prefix (L198–202) is the pragmatic workaround and is consistent with the Phase 2 completion doc.

### 2. `providerFor` fallback is defence-in-depth

Three layers gate the "user picks a provider that isn't enabled" case (`agent.go:64–71`):

1. **Frontend** — dropdown hides OpenRouter `<optgroup>` when `openrouterModels.length === 0` (`AgentConfigPage.vue:163`). User never sees an option the backend won't accept.
2. **Handler** — `UpdateAppAgentConfig` rejects `openrouter` with a 400 when `OPENROUTER_API_KEY` is unset (`applications.go:210–213`), so a crafted request hits a visible error.
3. **Agent loop** — `providerFor` silently falls back to `defaultProviderName="anthropic"` if a stale DB row somehow points at an unregistered provider. This is the right failure mode for a background monitoring goroutine where a 500 would poison the loop.

The empty-string default in the DB migration (`'anthropic'`) also means pre-0.29 rows read correctly without a backfill — verified by the handler default at `applications.go:203–205`.

### 3. Rate-limiter contract preserved

`RunMonitoring` at `loop.go:226` still takes a blocking path regardless of provider. The comment at L221–225 explicitly pins the non-streaming constraint to the 0.27.0 rate limiter, and the code matches — each `provider.ChatCompletion` call is one atomic rate-limiter unit, whether it routes through Anthropic or OpenRouter. The two providers even share the same `ChatParams{MaxTokens: defaultMaxTokens}` call site, so there's no drift risk between them.

### 4. Security posture

- **API keys** — `OPENROUTER_API_KEY` is loaded via `getEnv` (`config.go:36`), never logged. `agent.go:48` logs only `"openrouter provider enabled"`, not the key. `Authorization: Bearer` header is set per-request and never stored outside the struct.
- **Response-body DoS guard** — `openRouterMaxBodyBytes = 10 MiB` with `io.LimitReader` at `provider_openrouter.go:147`. The 2 GB Fly VM with Lumber loaded is the stated constraint; 10 MiB is ~100× the real payload size. Correct trade-off.
- **Timeout** — 2 minute `http.Client.Timeout` at L44. Long enough for slow Gemini responses, short enough to fail loudly on upstream stalls.
- **No silent fallback on provider error** — a 401/402/429 from OpenRouter surfaces as a loud error to the agent loop, which is the Phase 2 completion doc's stated policy. Correct: the alternative ("why is my GPT-4o app responding like Claude") would be a debugging nightmare.

### 5. Test coverage is appropriately scoped

`provider_openrouter_test.go` has four tests that pin the risky parts at the field level:

- **`TestOpenRouter_RequestShape`** — system-prompt-as-first-message, tool schema wrapping, required headers, usage decoding.
- **`TestOpenRouter_ToolCallResponse`** — `finish_reason:"tool_calls"` → `StopReasonToolUse`, arguments-string round-trip.
- **`TestOpenRouter_AssistantToolUseRoundTrip`** — the full multi-turn asymmetry: assistant with `Blocks` → single assistant message with `tool_calls`, then tool result → `role:"tool"` with matching `tool_call_id`. **This is the single most important test in the suite** because it proves the agent loop's conversation history replays correctly on OpenRouter.
- **`TestOpenRouter_ErrorMapping`** — 401/402/429/500 all produce readable errors.

The existing Anthropic path is covered by `loop_test.go` (unchanged, running through `NewAnthropicProviderWithClient`). The Phase 1 refactor preserved this test verbatim, which is a strong correctness signal — the abstraction change was pure.

### 6. Frontend/backend contract alignment

- `ModelOption` Go struct (`models.go:8`) and TypeScript interface (`types/models.ts:1`) match field-for-field (`id`, `name`, `provider`, `context_length`, `pricing.prompt`, `pricing.completion`). The `provider` discriminator is typed as a union of the same two literals on both sides.
- `GET /api/models` returns a flat list; the Vue page filters it by `provider` into two `<optgroup>`s (`AgentConfigPage.vue:29–30`). No split endpoint needed.
- `onModelSelect` auto-syncs `formProvider` from the selected model's group membership (L42–45), eliminating the invalid `provider:"anthropic" + model:"openai/gpt-4o"` combo client-side. The backend doesn't currently cross-validate model↔provider, but the client-side auto-sync plus the provider enum check is adequate.
- `AppAgentConfig.provider` on the frontend is `provider ?? 'anthropic'` (L75, L249) — graceful handling of legacy rows that predate the field.

---

## Low-severity findings (non-blocking)

These do not affect the 8.5/10 bar. Listed for awareness; none need to ship with the PR.

### L1 — `Content: string(textParts)` with `omitempty` drops the field when empty

**Location:** `provider_openrouter.go:72` (`Content string \`json:"content,omitempty"\``) combined with L247–251.

**Observation:** When an assistant message has `tool_calls` but no text block, `Content` is `""` and `omitempty` drops it from the JSON body entirely. OpenAI's API accepts this (and so does OpenRouter's passthrough), but the canonical spec is `content: null` when `tool_calls` is present. Every production provider I'm aware of tolerates omission, but if you ever hit a strict downstream that rejects it, the fix is one line: drop `,omitempty` from `Content` and set it to an explicit empty string (which is also spec-legal).

**Recommendation:** Leave as-is. If a specific downstream provider breaks, flip the omitempty then.

### L2 — `MaxTokens` also uses `omitempty`

**Location:** `provider_openrouter.go:65` (`MaxTokens int \`json:"max_tokens,omitempty"\``).

**Observation:** If a future caller passes `MaxTokens: 0`, the field is omitted and OpenRouter applies its own (model-dependent) default. The Anthropic provider, by contrast, substitutes 4096 (`provider_anthropic.go:40–42`). This is a silent behavioural divergence between providers — today it's unreachable because `loop.go` always passes `defaultMaxTokens = 4096`, but a future refactor that forgets this would cause Anthropic and OpenRouter to behave differently for the same input.

**Recommendation:** Optional — either mirror the Anthropic provider's `if MaxTokens == 0 { MaxTokens = 4096 }` guard inside `buildOpenRouterRequest`, or drop the `omitempty`. Both are one-line fixes. Safe to defer.

### L3 — Default model ID duplicated across three call sites

**Location:** `"claude-sonnet-4-6"` appears as `defaultModelID` (`loop.go:18`), the `CreateApplication` default (`applications.go:80`), and the `GetAppAgentConfig` fallback (`applications.go:147`), plus `req.Model` default at L177.

**Observation:** Four copies of the same string. When Sonnet 4.7 ships and you update `defaultModelID`, the other three sites will silently drift.

**Recommendation:** Export `agent.DefaultModelID` (rename the existing `defaultModelID` const to be capitalised) and reference it from `applications.go`. ~4 line change, prevents drift. Safe to do post-merge.

### L4 — No integration test exercising the provider switch through `RunConversation`

**Observation:** The provider abstraction is covered by the unit tests in `provider_openrouter_test.go`, and the agent loop itself is covered by `loop_test.go` on the Anthropic path. There is no test that calls `RunConversation` with an `AppAgentConfig{Provider: "openrouter"}` and verifies the loop routes through the OpenRouter provider end-to-end. The risk is low because the routing is literally two lines (`loop.go:79–82`, `loop.go:95`), but a single test would lock the contract.

**Recommendation:** Add a `TestRunConversation_RoutesToOpenRouter` that stands up an `httptest.Server`, registers an `OpenRouterProvider` on a fake agent, and asserts the server received the call. ~40 lines. Nice-to-have, not required for the merge.

### L5 — `parseSeverityFromResponse` is a string scan, not provider-aware

**Location:** `loop.go:334–349`.

**Observation:** Heuristic keyword scan on the model's prose output. Models that aren't prompted to emit `severity: critical` (i.e. most OpenRouter models without the Heimdall system prompt's exact phrasing) may fall through to the `"critical" in text` fallback, which matches on casual mentions of the word. This isn't new in 0.29.0 — it was written before the OpenRouter work — but the OpenRouter launch is the first time the agent will routinely run on models that haven't been specifically tuned to Heimdall's prompt conventions.

**Recommendation:** Monitor in production. If non-Anthropic models produce noisy severity ratings, move severity extraction into a structured `severity` tool call rather than prose parsing. Tracked as future work, not blocking.

### L6 — Frontend `formProvider` is mostly ornamental

**Location:** `AgentConfigPage.vue:22, 42–45, 75, 93`.

**Observation:** The `<select>` options are driven by `anthropicModels` / `openrouterModels` computeds, not by `formProvider`. `formProvider` is used only in the save payload and in the display-mode `via {provider}` line. `onModelSelect` syncs it from the selected model's group, which works, but the double-source-of-truth (model ID ↔ derived provider) is a tiny amount of avoidable state.

**Recommendation:** Derive `formProvider` as a `computed` from `formModel` (looking it up in `models.value`) and drop the ref. ~5 line change. Cosmetic; not blocking.

---

## Summary

The OpenRouter integration is **production-ready**. The Phase 1 abstraction paid off exactly as the completion docs promised — the Phase 3 wire-up was a two-line change because the loop speaks neutral types end-to-end. The translation layer is the riskiest part of the change and is correct on all five Anthropic↔OpenAI divergences, pinned by field-level tests. Security, timeouts, body-size guards, and the provider-not-enabled gate are all in place at the three right layers (frontend, handler, loop fallback).

**Findings L1–L6 are all low-severity and non-blocking.** The most valuable follow-up is L3 (dedupe `defaultModelID`) because it's a real maintenance hazard; the rest can wait for a rainy Friday.

No refactors recommended. No files exceed the 500-LOC threshold.

**Ship it.**
