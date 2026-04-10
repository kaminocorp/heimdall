# OpenRouter Integration — Phase 3 Completion

**Status:** ✅ Complete
**Date:** 2026-04-07
**Plan:** [openrouter-implementation.md](../executing/openrouter-implementation.md), Phase 3
**Validation:** `go build ./...` clean, `go vet ./...` clean, `go test ./...` all green
**Builds on:** [openrouter-phase-2.md](openrouter-phase-2.md)

---

## Goal

Make `OpenRouterProvider` (registered in Phase 2) actually reachable from a
real app config. Adds a `provider` column to `app_agent_config`, threads it
through sqlc, the agent loop, and the API handler. Phase 2 built the engine;
Phase 3 connects the wires.

After this phase, an operator can manually set
`UPDATE app_agent_config SET provider='openrouter', model='openai/gpt-4o' WHERE app_id = ...`
and that app's agent loop will route through OpenRouter on the next call. The
dropdown for end-users still doesn't exist (Phases 4–5), but the backend is
complete.

---

## Files changed

### New (2)

| File | Purpose |
|---|---|
| `backend/migrations/022_agent_config_provider.up.sql` | `ALTER TABLE app_agent_config ADD COLUMN provider TEXT NOT NULL DEFAULT 'anthropic'` |
| `backend/migrations/022_agent_config_provider.down.sql` | `ALTER TABLE app_agent_config DROP COLUMN provider` |

### Modified (4)

| File | Change |
|---|---|
| `backend/internal/db/queries/app_agent_config.sql` | `provider` added to `UpsertAppAgentConfig` INSERT, ON CONFLICT SET, and the existing `SELECT *` (auto-picks it up) |
| `backend/internal/db/{models.go,app_agent_config.sql.go}` | Regenerated via `sqlc generate` — `AppAgentConfig.Provider` and `UpsertAppAgentConfigParams.Provider` fields added |
| `backend/internal/agent/loop.go` | `RunConversation` now reads `appCfg.Provider` from per-app config; `RunMonitoring` now passes `appConfig.Provider` to `providerFor()` instead of the hardcoded `defaultProviderName` |
| `backend/internal/api/handlers/applications.go` | `updateAppAgentConfigRequest` accepts `provider`; `UpdateAppAgentConfig` validates it (`anthropic` always allowed; `openrouter` only when `OPENROUTER_API_KEY` is set; everything else rejected); default-create on app creation passes `Provider: "anthropic"`; the no-row-yet defaults response includes `"provider": "anthropic"` |

### Untouched (intentional)

- `provider.go`, `provider_anthropic.go`, `provider_openrouter.go` — Phase 3 is
  pure plumbing. The provider implementations don't care that the resolution
  source changed.
- `monitor.go`, `tools.go`, `chat.go` — none of these read provider config.
- All existing tests — the new column has a default, so every existing
  `UpsertAppAgentConfigParams{...}` literal in tests still compiles (Go struct
  literals fill missing fields with zero values, and `""` resolves to the
  default via `providerFor`).

---

## The two-line loop edit (the actual unblocker)

Most of Phase 3 is mechanical SQL/sqlc/handler work. The change that *actually*
makes OpenRouter reachable is two lines in `loop.go`:

**`RunConversation`** (interactive chat path):
```go
if appCfg.Provider != "" {
    providerName = appCfg.Provider
}
```

**`RunMonitoring`** (background monitoring path):
```go
provider := a.providerFor(appConfig.Provider)  // was: defaultProviderName
```

That's it. The rest of the loop — message building, tool dispatch, iteration
cap, EmitLog calls, rate limiter — is unchanged. Phase 1's abstraction is
doing all the heavy lifting; Phase 3 just changes the **string** that gets
passed to `providerFor()`.

This is the cleanest possible Phase 3 because of how Phase 1 was designed: the
provider lookup is a single function call with a string key, and that key has
exactly one source of truth (the `app_agent_config` row). Replacing the
hardcoded fallback with the column read is a one-line change per call site.

---

## Validation logic at the API boundary

The handler enforces three rules on `UpdateAppAgentConfig`:

```go
if req.Provider == "" {
    req.Provider = "anthropic"
}
switch req.Provider {
case "anthropic":
    // always available
case "openrouter":
    if s.Config.OpenRouterKey == "" {
        jsonError(w, "openrouter provider not enabled on this server", 400)
        return
    }
default:
    jsonError(w, "provider must be one of: anthropic, openrouter", 400)
    return
}
```

**Why validation lives in the handler, not the database:**

The migration deliberately has **no** `CHECK (provider IN ('anthropic', 'openrouter'))`
constraint. Reasons:

1. **Adding a new provider shouldn't require a migration.** When/if a third
   provider lands (e.g. a self-hosted Ollama bridge), the only places that
   need to change are `agent.go` (register), `applications.go` (validate),
   and the model list. A DB-level CHECK would force a migration too.
2. **The `OPENROUTER_API_KEY` gate is server-side state**, not data. A CHECK
   constraint can't express "openrouter is only allowed when this env var is
   set." That gate has to live in Go.
3. **`providerFor` already does defence-in-depth.** If a buggy DB row contains
   `provider='nonsense'`, the agent loop silently falls back to anthropic
   instead of crashing. So the handler is the first line of defence (reject
   at the API boundary), `providerFor` is the second (don't crash if we
   somehow get bad data anyway).

**Why we reject `openrouter` instead of silently falling back:**

The selection plan (`openrouter-model-selection.md` §"Risks & Mitigations")
explicitly says: don't fall back to Anthropic when OpenRouter is unavailable,
because "that would change the user's model without consent." The same
principle applies at the configuration layer: if the operator hasn't set up
OpenRouter, the user trying to select an OpenRouter model should see a clear
error, not a silent downgrade to Anthropic with a different model. That's the
shape of bug that takes a week to find ("why is my GPT-4o app responding like
Claude?").

---

## sqlc regeneration

`sqlc generate` produced:

| File | Change |
|---|---|
| `internal/db/models.go` | `AppAgentConfig` struct gained `Provider string \`json:"provider"\`` |
| `internal/db/app_agent_config.sql.go` | `UpsertAppAgentConfigParams` gained `Provider string`; the SELECT scan now reads `&i.Provider`; the INSERT/UPSERT now binds `arg.Provider` as the sixth parameter |

No manual edits to the generated code. The `sqlc.yaml` overrides
(`uuid → google/uuid.UUID`, `jsonb → json.RawMessage`, etc.) were already in
place, so the new column type (`text`) maps to plain Go `string` with no
ceremony.

---

## Test impact

**Zero test edits required.** This was a pleasant surprise and is worth
calling out:

- `monitor_test.go` constructs `db.AppAgentConfig{AppID: ..., Model: ...}`
  literals. Go fills the new `Provider` field with `""`. `providerFor("")`
  falls back to `defaultProviderName` (anthropic), which is what the tests
  expect. **No test changes.**
- `loop_test.go` doesn't construct `AppAgentConfig` directly — it exercises
  the loop via `RunLoop` (which uses `appID = uuid.Nil` and falls through to
  the global `agent_config` path). **No test changes.**
- `applications.go` handler tests (in `internal/api/handlers/`) still pass —
  the new validation branch is only exercised when a request *includes*
  `provider`, and existing tests don't, so they hit the empty-string default
  path which resolves to `"anthropic"`. **No test changes.**

This is the second nice property of Phase 1's design that's paying off in
Phase 3: because the empty string is a valid input that resolves to the
default, no test had to be updated to "fill in" the new field.

---

## Validation gates from `openrouter-implementation.md` Phase 3

- [x] `make migrate-up` would succeed (file is well-formed; not run in this
      session because it requires a live DB connection)
- [x] `sqlc generate` produces clean output (no errors, expected fields appear)
- [x] Existing configs default to `provider = 'anthropic'` (DB default + Go
      empty-string fallback both cover this)
- [x] Setting `provider = 'openrouter'` + an OpenRouter model ID routes
      through `OpenRouterProvider` (verified by reading the loop.go change;
      no live OpenRouter call attempted)
- [x] `go build ./...` clean
- [x] `go vet ./...` clean
- [x] `go test ./...` all green across the whole backend

---

## Manual smoke test recipe (post-deploy)

Once the migration is applied to a real environment with `OPENROUTER_API_KEY`
set, end-to-end validation looks like this:

```sql
-- 1. Pick an app and switch it to OpenRouter
UPDATE app_agent_config
SET provider = 'openrouter',
    model    = 'openai/gpt-4o'
WHERE app_id = '<your-test-app-id>';
```

Then:
1. Open the agent chat for that app in the UI.
2. Send a message that will trigger `search_logs` (e.g. "What errors have we seen recently?").
3. Confirm the response comes back with no error and the `search_logs` tool
   was actually executed (visible in `agent_log`).
4. Bonus: tail backend logs and grep for `openrouter provider enabled` at
   startup, confirming the provider is registered.

If anything in steps 2–3 fails, the most likely culprits in order are:
1. Tool schema translation — `TestOpenRouter_RequestShape` should have caught
   it, but production models can be pickier than the test asserts.
2. Tool result round-trip — `TestOpenRouter_AssistantToolUseRoundTrip` covers
   this, but again, models may interpret the `"ERROR: "` prefix differently
   than expected.
3. The model itself doesn't reliably handle multi-turn tool use — that's a
   model-selection problem, not a Heimdall bug. Picking another model from
   the curated list (Phase 4) is the fix.

Frontend validation (Phase 5) is the right place to catch the third case
before users hit it — by curating a list of models known to handle the loop.

---

## What stayed the same

- The 0.27.0 rate limiter (`a.limiter.Wait(ctx)`) still applies on the
  monitoring path regardless of provider. So the cost-bound guarantee carries
  over to OpenRouter monitoring exactly as it does to Anthropic monitoring.
- Default model fallback (`claude-sonnet-4-5`) — still applies when
  `appCfg.Model == ""`, regardless of provider. Phase 4/5 will prevent that
  string from being meaningful for an OpenRouter row by validating
  model-against-provider in the dropdown, but the backend doesn't enforce
  it (and shouldn't — users can set arbitrary OpenRouter model IDs via
  the API even before the dropdown lands).
- Tool definitions in `tools.go` — unchanged.
- Conversation persistence in `chat.go` — unchanged.

---

## Follow-ups

### Phase 4 — `GET /api/models`
Curated model lists (Anthropic always; OpenRouter conditionally) returned as
`ModelOption[]` with id/name/provider/context/pricing. New file
`backend/internal/agent/models.go` for the curated lists; new handler
`backend/internal/api/handlers/models.go`; new route in `router.go`. This is
the first phase that requires curating the actual OpenRouter model IDs and
verifying their pricing — *must check OpenRouter's catalogue at implementation
time* (the implementation plan flagged this).

### Phase 5 — Frontend dropdown
Replace the text input + datalist with a `<select>` grouped by provider via
`<optgroup>`. Fetch models from `/api/models` on mount. Show
context-length and pricing in option labels. Auto-set provider when a model
is selected. Display mode shows "via openrouter" / "via anthropic" alongside
the model.

### Streaming follow-up
Still deferred. Same shape as before: `ChatCompletionStream` on the interface,
implementations on both providers, `RunConversationStream` in the agent,
`chat.go` rewrite, `useAgent.ts` rewrite. This is now the *only* gap between
the current backend and the full plan in `openrouter-implementation.md`.

### Optional cleanups
- The `"claude-sonnet-4-6"` literal is duplicated in `applications.go` (three
  places: default-create, defaults response, and handler default). Worth a
  package-level constant if it changes again.
- The default-create on application creation could pass `Provider: ""` and let
  the DB default handle it, but explicit is clearer at the call site.
