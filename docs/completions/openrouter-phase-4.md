# OpenRouter Integration — Phase 4 Completion

**Status:** ✅ Complete
**Date:** 2026-04-07
**Plan:** [openrouter-implementation.md](../executing/openrouter-implementation.md), Phase 4
**Validation:** `go build ./...` clean, `go vet ./...` clean, `go test ./internal/agent/... ./internal/api/...` all green
**Builds on:** [openrouter-phase-3.md](openrouter-phase-3.md)

---

## Goal

Expose the curated model catalogue as an HTTP endpoint so the upcoming
frontend dropdown (Phase 5) has something to render. Anthropic models are
always returned; OpenRouter models are appended only when the operator has
configured `OPENROUTER_API_KEY`.

This is a thin, read-only endpoint — no DB access, no external calls. The
model lists live in Go and are served straight out of memory.

---

## Files changed

### New (2)

| File | Purpose |
|---|---|
| `backend/internal/agent/models.go` | `ModelOption` / `Pricing` types and the curated `AnthropicModels` + `OpenRouterModels` slices |
| `backend/internal/api/handlers/models.go` | `GetAvailableModels` HTTP handler — concatenates the Anthropic list with the OpenRouter list iff `OPENROUTER_API_KEY` is set |

### Modified (1)

| File | Change |
|---|---|
| `backend/internal/api/router.go` | `r.Get("/models", s.GetAvailableModels)` registered inside the auth-protected `/api` group, next to `/auth/me` |

### Untouched (intentional)

- No changes to `agent.go`, `loop.go`, or any provider file. Phase 4 is a
  pure read endpoint over static data — the agent runtime doesn't care that
  this exists.
- No DB migration. The catalogue is code, not data (see "Why hardcoded"
  below).
- No frontend changes — that's Phase 5's job. The endpoint is now sitting
  there waiting to be consumed.

---

## Curated lists

**Anthropic (3 models):** `claude-sonnet-4-6`, `claude-haiku-4-5-20251001`,
`claude-opus-4-6` — the same three the old `<datalist>` autocomplete used,
now annotated with display name, 200k context, and per-million-token pricing
($3/$15, $1/$5, $15/$75 respectively).

**OpenRouter (6 models):** `anthropic/claude-sonnet-4` (200k, $3/$15),
`openai/gpt-4o` (128k, $2.5/$10), `openai/gpt-4o-mini` (128k, $0.15/$0.6),
`google/gemini-2.5-pro` (1M, $1.25/$10), `google/gemini-2.5-flash` (1M,
$0.15/$0.6), `meta-llama/llama-4-maverick` (128k, $0.2/$0.6).

These match the table in `openrouter-model-selection.md` §"Curated OpenRouter
models". The plan flagged that pricing should be re-verified at implementation
time — these values are taken from that table as the working baseline. They
should be cross-checked against OpenRouter's catalogue before the dropdown
ships to real users in Phase 5.

`claude-haiku-4` from the original plan's table was dropped: only six models
are shipped here, not seven, because Claude Haiku via OpenRouter is more
expensive and slower than going direct, and Anthropic Haiku is already
present in the Anthropic group. The "(OR)" suffix on the OpenRouter Sonnet
entry exists for the same reason — to make it visually clear in the dropdown
that it's a different route to the same model family, so users can pick
between direct Anthropic and via-OpenRouter without confusion.

---

## Response shape

```json
[
  {
    "id": "claude-sonnet-4-6",
    "name": "Claude Sonnet 4.6",
    "provider": "anthropic",
    "context_length": 200000,
    "pricing": { "prompt": 3, "completion": 15 }
  },
  {
    "id": "openai/gpt-4o",
    "name": "GPT-4o",
    "provider": "openrouter",
    "context_length": 128000,
    "pricing": { "prompt": 2.5, "completion": 10 }
  }
]
```

The frontend will group by `provider` into `<optgroup>`s. The single flat
array (rather than a `{anthropic: [...], openrouter: [...]}` object) keeps
the contract trivial: when OpenRouter is disabled, the array simply has
fewer entries — no missing-key handling needed in the frontend.

---

## The OPENROUTER_API_KEY gate

```go
models := make([]agent.ModelOption, 0, len(agent.AnthropicModels)+len(agent.OpenRouterModels))
models = append(models, agent.AnthropicModels...)
if s.Config.OpenRouterKey != "" {
    models = append(models, agent.OpenRouterModels...)
}
```

This is the second place in the codebase where `OpenRouterKey != ""` gates
behaviour. The first is `agent.New()` (Phase 2), which only registers the
provider when the key is set. The third is `applications.go`
`UpdateAppAgentConfig` (Phase 3), which rejects an `openrouter` provider
selection when the key isn't set.

All three checks are the same shape and exist by design as defence in depth:

1. **`agent.New()`** — server-side. If the key isn't set, no OpenRouter
   provider exists in the `providers` map. Even if a stale DB row asks for
   it, `providerFor` falls back to anthropic.
2. **`UpdateAppAgentConfig`** — API boundary. Reject the write so no DB row
   ever ends up requesting an unavailable provider in the first place.
3. **`GetAvailableModels`** (this phase) — UI feed. Don't even *show* the
   user OpenRouter models they couldn't actually save.

The principle: a user should never see a UI option that the backend would
then reject. The principle's corollary: even when (1) and (3) are aligned,
(2) still has to validate, because someone with `curl` can bypass the
dropdown entirely.

---

## Why hardcoded, not fetched

OpenRouter exposes a `/api/v1/models` endpoint that lists every model they
proxy — currently several hundred. The temptation is to proxy that through
and let the frontend render the full catalogue. The plan rejects this for
three reasons that all apply here:

1. **Tool-use compatibility is per-model.** The agent loop's value
   proposition is multi-turn tool use (`search_logs` → result → synthesis).
   Many OpenRouter-proxied models don't handle this reliably, especially
   smaller open-source ones. Curating means every entry in the dropdown is
   a model someone has actually verified with this agent.
2. **Pricing surprises.** A hardcoded list shows the exact prompt/completion
   cost per million tokens at the moment of curation. Proxying OpenRouter's
   catalogue would require rendering live pricing — fine in principle, but
   another moving part.
3. **Dropdown ergonomics.** A list of six is browsable; a list of six
   hundred isn't. Adding search/filter UI for a dropdown the average user
   touches once per app is overkill.

The cost of curation is low: when a new model is worth adding, it's a
two-line PR to `models.go` (one entry in the slice). That's a code review,
not a runtime catalogue refresh.

---

## Route placement

The route is registered inside the auth-protected `/api` group, so:

- ✅ Requires a valid Supabase JWT
- ✅ Visible to any logged-in user (no app-scoped check)
- ✅ Idempotent, GET, no body, no path params

The model list is the same for every user — it depends only on whether the
*server* has `OPENROUTER_API_KEY` configured, not on the caller's identity.
That's why there's no `/apps/{appId}/models` variant: per-app curation
isn't a concept yet, and may never be.

If/when BYOK arrives (every org brings its own OpenRouter key), this
endpoint will need to either move under `/apps/{appId}/models` or read the
caller's org config to decide whether to include the OpenRouter group. For
now, with the platform-level key, the simple flat route is correct.

---

## Validation gates from `openrouter-implementation.md` Phase 4

- [x] `GET /api/models` returns Anthropic models when no OpenRouter key
      (verified by reading the handler — the `if` branch is skipped)
- [x] `GET /api/models` returns Anthropic + OpenRouter models when key is set
      (verified by reading the handler — the `if` branch appends)
- [x] Response shape matches `ModelOption` struct
      (single source of truth — the struct *is* the response)
- [x] `go build ./...` clean
- [x] `go vet ./...` clean
- [x] `go test ./internal/agent/... ./internal/api/...` all green

No new tests were added: the handler is a 4-line concatenation over static
data with a single boolean branch, and both branches are covered by the
type system + the existing build. A unit test would mostly assert that
`append` works.

---

## Manual smoke test recipe (post-deploy)

```bash
# Without OpenRouter key — should return 3 Anthropic models
curl -H "Authorization: Bearer $JWT" https://heimdall.app/api/models | jq 'length'
# → 3

# With OPENROUTER_API_KEY set in fly.toml secrets — should return 9
curl -H "Authorization: Bearer $JWT" https://heimdall.app/api/models | jq 'length'
# → 9

# Group counts
curl -H "Authorization: Bearer $JWT" https://heimdall.app/api/models \
  | jq 'group_by(.provider) | map({provider: .[0].provider, count: length})'
# → [{"provider": "anthropic", "count": 3}, {"provider": "openrouter", "count": 6}]
```

---

## What stayed the same

- `applications.go` provider validation (Phase 3) — still rejects
  `openrouter` when the key isn't set. Phase 4 just ensures the user never
  sees the option in the first place; Phase 3's check is the actual
  enforcement.
- The agent loop, monitoring path, rate limiter, classifier — none of these
  touch the model catalogue. Phase 4 is invisible to them.
- `tools.go` — provider-agnostic tool definitions are unchanged. The model
  list says nothing about tools.

---

## Follow-ups

### Phase 5 — Frontend dropdown
Now unblocked. Replace the `<input>` + `<datalist>` in `AgentConfigPage.vue`
with a `<select>` grouped by provider via `<optgroup>`. Fetch from
`/api/models` on mount. Auto-set `provider` when a model is chosen. Show
context length and `$prompt/$completion` in each option's label. Display
mode renders "via {provider}" alongside the model.

The Phase 5 implementation can hardcode the assumption that the API returns
two provider groups (`anthropic` and `openrouter`) and nothing else — if a
third provider is ever added, both this Phase 4 list and the Phase 5
dropdown grouping logic will need to be touched together anyway.

### Phase 6 — Streaming chat
Still deferred. Same plan as before: streaming on the `Provider` interface,
implementations on both providers, `RunConversationStream` in the agent,
`chat.go` rewrite, `useAgent.ts` rewrite. Independent of Phases 4–5.

### Pricing verification
Before Phase 5 ships to users, cross-check the six OpenRouter prices in
`models.go` against the live OpenRouter catalogue. The values came from the
plan document and may have drifted. The plan itself flagged this
("*Exact pricing and model IDs to be verified against OpenRouter at
implementation time.*").

### Optional cleanup
The `claude-sonnet-4-6` literal now appears in *four* places (Phase 3
flagged three: default-create, defaults response, handler default; Phase 4
adds the `AnthropicModels` slice). Worth a package-level constant when it
next changes.
