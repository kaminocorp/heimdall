# OpenRouter Integration & Model Selection Dropdown

## Context

The Agent Configuration page currently offers model selection via a text input with a `<datalist>` autocomplete of three hardcoded Claude models (`claude-sonnet-4-6`, `claude-haiku-4-5-20251001`, `claude-opus-4-6`). The backend is tightly coupled to the Anthropic Go SDK — a single `ANTHROPIC_API_KEY` initialises one client at startup, and the model string is passed directly to `anthropic.MessageNewParams`.

This plan introduces **OpenRouter** as a model provider, giving users access to hundreds of models (Claude, GPT, Gemini, Llama, Mistral, etc.) through a single integration, and replaces the text input with a proper dropdown.

---

## Architecture Decision: Provider Abstraction

We introduce a **provider abstraction layer** rather than replacing Anthropic outright. Reasons:

1. **Direct Anthropic stays available.** Lower latency, no middleman markup, and guaranteed feature parity (extended thinking, prompt caching, etc.). Power users or self-hosters who only want Claude should not be forced through OpenRouter.
2. **OpenRouter for breadth.** Users who want GPT-4o, Gemini, or open-source models get them without us integrating each provider individually.
3. **Tool use compatibility.** The agent loop's tool-use format (tool definitions, tool results, multi-turn) must work identically regardless of provider. Anthropic and OpenAI use different tool schemas — the abstraction normalises this.

### Provider Interface

```go
// internal/agent/provider.go

type Provider interface {
    // ChatCompletion sends a message and returns the model's response.
    // Tools, system prompt, and conversation history are passed in.
    ChatCompletion(ctx context.Context, params ChatParams) (*ChatResponse, error)
}

type ChatParams struct {
    Model        string
    SystemPrompt string
    Messages     []Message
    Tools        []Tool
    MaxTokens    int
}

type ChatResponse struct {
    StopReason StopReason          // EndTurn | ToolUse
    TextBlocks []TextBlock
    ToolCalls  []ToolCall
    Usage      Usage
}
```

Two implementations:

| Provider | SDK / Client | Base URL | Auth |
|----------|-------------|----------|------|
| `AnthropicProvider` | `anthropic-sdk-go` (existing) | `https://api.anthropic.com` | `ANTHROPIC_API_KEY` |
| `OpenRouterProvider` | `openai-go` or raw HTTP | `https://openrouter.ai/api/v1` | `OPENROUTER_API_KEY` |

### Provider Resolution

Each `app_agent_config` row stores both a `provider` and a `model`. At runtime:

```
provider = "openrouter", model = "anthropic/claude-sonnet-4" → OpenRouterProvider
provider = "anthropic",  model = "claude-sonnet-4-6"         → AnthropicProvider
provider = "" (legacy)                                        → AnthropicProvider (backwards compat)
```

---

## Implementation Plan

### Phase 1 — Provider Abstraction (backend)

Decouple the agent loop from the Anthropic SDK without changing any external behaviour.

| # | File | Change |
|---|------|--------|
| 1 | `internal/agent/provider.go` | **New.** Define `Provider` interface, `ChatParams`, `ChatResponse`, `Message`, `ToolCall`, `StopReason` types |
| 2 | `internal/agent/provider_anthropic.go` | **New.** `AnthropicProvider` struct wrapping the existing `anthropic.Client`. Implement `ChatCompletion` by translating `ChatParams` → `anthropic.MessageNewParams` and response back to `ChatResponse` |
| 3 | `internal/agent/agent.go` | Replace `client *anthropic.Client` field with `providers map[string]Provider`. Initialise `AnthropicProvider` from `ANTHROPIC_API_KEY` at startup. Add `providerFor(providerName string) Provider` resolver method |
| 4 | `internal/agent/loop.go` | Replace direct `a.client.Messages.New()` calls with `provider.ChatCompletion()`. Extract model + provider from app config. The tool-use iteration loop stays the same — just operates on `ChatResponse` instead of `anthropic.Message` |
| 5 | `internal/agent/monitor.go` | Same refactor as loop.go for the `RunMonitoring` call path |
| 6 | `internal/agent/tools.go` | Adapt tool definitions to the provider-agnostic `Tool` type. Each provider implementation maps to its native format |

**Validation gate:** All existing tests pass. Agent chat + monitoring behave identically with `AnthropicProvider`.

### Phase 2 — OpenRouter Provider (backend)

| # | File | Change |
|---|------|--------|
| 1 | `internal/agent/provider_openrouter.go` | **New.** `OpenRouterProvider` implementing `Provider`. Uses OpenAI chat completions format. Translates tools to OpenAI function-calling schema. Sets `HTTP-Referer` and `X-Title` headers per OpenRouter requirements |
| 2 | `internal/config/config.go` | Add `OpenRouterKey string` from `OPENROUTER_API_KEY` env var. **Not required** — only validated if OpenRouter provider is used |
| 3 | `internal/agent/agent.go` | If `OPENROUTER_API_KEY` is set, initialise `OpenRouterProvider` and add to `providers` map |
| 4 | `.env.example` | Add `OPENROUTER_API_KEY=` with comment |

**Tool schema translation detail:**

Anthropic tool format:
```json
{ "name": "search_logs", "description": "...", "input_schema": { "type": "object", ... } }
```

OpenAI/OpenRouter tool format:
```json
{ "type": "function", "function": { "name": "search_logs", "description": "...", "parameters": { "type": "object", ... } } }
```

The `Provider` interface uses its own neutral format; each implementation maps to native.

### Phase 3 — Database Migration

| # | File | Change |
|---|------|--------|
| 1 | `migrations/017_agent_config_provider.up.sql` | `ALTER TABLE app_agent_config ADD COLUMN provider TEXT NOT NULL DEFAULT 'anthropic';` |
| 2 | `migrations/017_agent_config_provider.down.sql` | `ALTER TABLE app_agent_config DROP COLUMN provider;` |
| 3 | `internal/db/queries/app_agent_config.sql` | Add `provider` to SELECT, INSERT, and UPDATE queries |
| 4 | Run `make sqlc-generate` | Regenerate Go types to include `Provider` field |

### Phase 4 — Model List Endpoint (backend)

Expose an endpoint that returns available models for the dropdown.

| # | File | Change |
|---|------|--------|
| 1 | `internal/api/handlers/models.go` | **New.** `GetAvailableModels` handler |
| 2 | `internal/api/router.go` | `GET /api/models` route (auth-protected) |

**Implementation logic:**

Both Anthropic and OpenRouter model lists are **hardcoded curated lists** — no runtime API calls to OpenRouter. This keeps the endpoint fast, deterministic, and avoids a runtime dependency on OpenRouter's availability just to populate a dropdown.

```go
func (s *Server) GetAvailableModels(w http.ResponseWriter, r *http.Request) {
    models := anthropicModels // always available

    // Only include OpenRouter models if the platform key is configured
    if s.config.OpenRouterKey != "" {
        models = append(models, openRouterModels...)
    }

    json.NewEncoder(w).Encode(models)
}
```

**Curated OpenRouter models (initial list, ~7–9 models):**

All selected for reliable multi-turn tool use support:

| Model ID (OpenRouter) | Display Name | Context | Prompt $/M | Completion $/M |
|------------------------|-------------|---------|------------|----------------|
| `anthropic/claude-sonnet-4` | Claude Sonnet 4 | 200k | $3 | $15 |
| `anthropic/claude-haiku-4` | Claude Haiku 4 | 200k | $1 | $5 |
| `openai/gpt-4o` | GPT-4o | 128k | $2.50 | $10 |
| `openai/gpt-4o-mini` | GPT-4o Mini | 128k | $0.15 | $0.60 |
| `google/gemini-2.5-pro` | Gemini 2.5 Pro | 1M | $1.25 | $10 |
| `google/gemini-2.5-flash` | Gemini 2.5 Flash | 1M | $0.15 | $0.60 |
| `meta-llama/llama-4-maverick` | Llama 4 Maverick | 128k | $0.20 | $0.60 |

*(Exact pricing and model IDs to be verified against OpenRouter at implementation time.)*

The curated list lives in a single file (`provider_openrouter_models.go`) for easy updates.

**Response shape:**

```json
[
  {
    "id": "claude-sonnet-4-6",
    "name": "Claude Sonnet 4.6",
    "provider": "anthropic",
    "context_length": 200000,
    "pricing": { "prompt": 3.0, "completion": 15.0 }
  },
  {
    "id": "openai/gpt-4o",
    "name": "GPT-4o",
    "provider": "openrouter",
    "context_length": 128000,
    "pricing": { "prompt": 2.5, "completion": 10.0 }
  }
]
```

### Phase 5 — Frontend Dropdown

Replace the text input with a grouped, searchable dropdown.

| # | File | Change |
|---|------|--------|
| 1 | `src/api/models.ts` | **New.** `getAvailableModels()` API function |
| 2 | `src/types/models.ts` | **New.** `ModelOption` interface matching the API response |
| 3 | `src/pages/AgentConfigPage.vue` | Replace `<input>` + `<datalist>` with a `<select>` dropdown grouped by provider. Add provider field to form state. Fetch models on mount. Show pricing/context info in option labels |
| 4 | `src/types/organization.ts` | Add `provider` field to `AppAgentConfig` interface |

**Dropdown UX:**

```
┌─ Model ──────────────────────────────────────┐
│ ▾ Claude Sonnet 4.6                          │
├──────────────────────────────────────────────┤
│ ANTHROPIC (DIRECT)                           │
│   Claude Sonnet 4.6         200k · $3/$15    │
│   Claude Haiku 4.5          200k · $1/$5     │
│   Claude Opus 4.6           200k · $15/$75   │
│ ─────────────────────────────────────────── │
│ OPENROUTER                                   │
│   Claude Sonnet 4           200k · $3/$15    │
│   GPT-4o                    128k · $2.5/$10  │
│   Gemini 2.5 Pro            1M   · $1.25/$10 │
│   Llama 4 Maverick          128k · $0.2/$0.6 │
│   ...                                        │
└──────────────────────────────────────────────┘
```

- Models grouped by provider with `<optgroup>` labels
- Each option shows context window size and per-million-token pricing
- OpenRouter section only appears if the backend returns OpenRouter models (i.e. key is configured)
- When user selects a model, `provider` is set automatically based on the model's group
- Search/filter via a text input above the `<select>` (or use a custom component if the list is long)

### Phase 6 — Config Save Flow Update

| # | File | Change |
|---|------|--------|
| 1 | `src/pages/AgentConfigPage.vue` | Include `provider` in the PUT payload. Display current provider + model in read mode |
| 2 | `internal/api/handlers/applications.go` | Accept `provider` in `UpdateAppAgentConfig`. Validate it's one of `anthropic`, `openrouter`. Default to `anthropic` if empty (backwards compat) |

---

## Data Flow (After Implementation)

```
User selects model in dropdown
        │
        ▼
Frontend sends PUT /api/apps/{id}/agent/config
  { "provider": "openrouter", "model": "anthropic/claude-sonnet-4", "mode": "continuous", ... }
        │
        ▼
Backend persists to app_agent_config table
        │
        ▼
Agent loop loads config for app
        │
        ▼
providerFor("openrouter") → OpenRouterProvider
        │
        ▼
OpenRouterProvider.ChatCompletion() → POST https://openrouter.ai/api/v1/chat/completions
  { "model": "anthropic/claude-sonnet-4", "messages": [...], "tools": [...] }
        │
        ▼
Response translated back to ChatResponse → tool-use loop continues as normal
```

---

## File Impact Summary

| Category | New Files | Modified Files |
|----------|-----------|----------------|
| Backend — Agent | 3 (`provider.go`, `provider_anthropic.go`, `provider_openrouter.go`) | 3 (`agent.go`, `loop.go`, `monitor.go`) |
| Backend — API | 1 (`handlers/models.go`) | 2 (`router.go`, `handlers/applications.go`) |
| Backend — Config | 0 | 2 (`config.go`, `.env.example`) |
| Backend — DB | 2 (migration up/down) | 1 (`app_agent_config.sql`) + regenerated sqlc |
| Backend — Tools | 0 | 1 (`tools.go`) |
| Frontend | 2 (`api/models.ts`, `types/models.ts`) | 2 (`AgentConfigPage.vue`, `types/organization.ts`) |
| **Total** | **8** | **11** |

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| OpenRouter tool-use format diverges from OpenAI spec for some models | Tool calls fail silently or return malformed JSON | Curated model list only includes models verified for tool use. Add response validation in `OpenRouterProvider` that catches malformed tool calls and returns a clean error |
| Some OpenRouter models handle multi-turn tool use poorly | Agent loop gets stuck or produces garbage | `max_iterations` guard already exists (10). Curated list limits exposure to known-good models |
| OpenRouter API downtime | Agent monitoring stops for OpenRouter-configured apps | Provider abstraction makes fallback possible. Log clearly when OpenRouter calls fail. Don't fall back silently to Anthropic — that would change the user's model without consent |
| Cost surprise — user selects expensive model without realising | Unexpected API bills via OpenRouter | Show pricing in dropdown options |

---

## Execution Order

Phases are designed to be mergeable independently:

1. **Phase 1** (provider abstraction) is a pure refactor — zero behaviour change, safe to merge first
2. **Phase 2** (OpenRouter provider) can be developed in parallel with Phase 3 (migration)
3. **Phase 4** (model list endpoint) depends on Phase 2
4. **Phase 5 + 6** (frontend) depend on Phase 3 + 4

Critical path: **Phase 1 → Phase 2+3 → Phase 4 → Phase 5+6**

---

## Decisions

1. **Bring your own key?** — **Deferred.** Platform-level `OPENROUTER_API_KEY` for now. BYOK (per-org or per-app) will be introduced later — requires encrypted key storage, rotation, and per-org billing visibility.

2. **Model allowlist vs. full catalogue?** — **Curated list (<10 models).** Hardcode a vetted list of models known to handle multi-turn tool use reliably. No "show all" toggle for now. The list lives in a single Go file (`provider_openrouter_models.go` or similar) for easy updates.

3. **Monitoring mode model constraints?** — **No restrictions.** Trust the user (primarily the developer/operator during testing). Pricing is visible in the dropdown, which is sufficient guard.

4. **Streaming?** — **Not now (Option A).** The entire stack is non-streaming today (blocking Claude API call → full response → single WebSocket message). Adding streaming touches every layer (agent loop, WebSocket handler, frontend composable) and would double the blast radius of this change. The `Provider` interface is designed to accommodate a future `ChatCompletionStream()` method without breaking the existing contract. Streaming is a natural follow-up project.

5. **Usage tracking?** — **Deferred.** Manual checking via OpenRouter dashboard for now. Future iteration can capture `x-openrouter-cost` response headers and surface per-app spend.

---

## Future Work (Out of Scope)

- **BYOK** — per-org OpenRouter API keys with encrypted storage
- **Streaming** — token-by-token delivery for chat responses (Option B or C from streaming analysis)
- **Usage tracking** — per-app cost capture from OpenRouter response headers
- **Model catalogue expansion** — "show all" toggle beyond the curated list
