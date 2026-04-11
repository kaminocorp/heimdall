# OpenRouter Integration & Streaming — Implementation Plan

Detailed, phase-by-phase implementation guide. Each phase is a mergeable unit with a clear validation gate. References use `file:line` notation against the current codebase.

**Decisions baked in:**
- Provider abstraction (Anthropic direct + OpenRouter)
- Curated model list (<10 models, hardcoded)
- Platform-level `OPENROUTER_API_KEY` (no BYOK yet)
- Streaming for interactive chat only (monitoring stays blocking)
- No usage tracking yet

---

## Phase 1 — Provider Abstraction

**Goal:** Decouple the agent loop from `anthropic-sdk-go` types. Zero behaviour change — pure refactor.

### Task 1.1 — Define the Provider interface and domain types

**New file:** `backend/internal/agent/provider.go`

```go
package agent

import "context"

// StopReason indicates why the model stopped generating.
type StopReason int

const (
    StopReasonEndTurn StopReason = iota
    StopReasonToolUse
)

// ChatParams holds everything needed for a model invocation.
type ChatParams struct {
    Model        string
    SystemPrompt string
    Messages     []ChatMessage
    Tools        []ToolDef
    MaxTokens    int
}

// ChatMessage is a provider-agnostic message in the conversation.
type ChatMessage struct {
    Role         string           // "user", "assistant"
    TextContent  string           // plain text (for user messages and simple assistant responses)
    Blocks       []ContentBlock   // rich content (text + tool_use blocks from assistant)
    ToolResults  []ToolResult     // tool results (for user messages returning tool output)
}

// ContentBlock is a single block in an assistant response.
type ContentBlock struct {
    Type  string // "text" or "tool_use"
    Text  string // populated when Type == "text"
    ToolCall *ToolCall // populated when Type == "tool_use"
}

// ToolCall represents a model's request to invoke a tool.
type ToolCall struct {
    ID    string
    Name  string
    Input json.RawMessage
}

// ToolResult is the outcome of executing a tool.
type ToolResult struct {
    ToolCallID string
    Content    string
    IsError    bool
}

// ToolDef is a provider-agnostic tool definition.
type ToolDef struct {
    Name        string
    Description string
    Parameters  map[string]any // JSON Schema "properties" object
    Required    []string
}

// ChatResponse is the full response from a non-streaming invocation.
type ChatResponse struct {
    StopReason StopReason
    Blocks     []ContentBlock
    Usage      Usage
}

// Usage tracks token consumption.
type Usage struct {
    InputTokens  int
    OutputTokens int
}

// StreamEvent is a single event from a streaming invocation.
type StreamEvent struct {
    Type     string // "text_delta", "tool_call_start", "tool_call_delta", "done"
    Delta    string // text content for "text_delta"
    ToolCall *ToolCall // populated on "tool_call_start"
    Response *ChatResponse // populated on "done" — full accumulated response
}

// Provider is the interface that both Anthropic and OpenRouter implement.
type Provider interface {
    // ChatCompletion performs a blocking invocation. Used by monitoring mode.
    ChatCompletion(ctx context.Context, params ChatParams) (*ChatResponse, error)

    // ChatCompletionStream performs a streaming invocation. Used by interactive chat.
    // The returned channel is closed when the response is complete.
    // The final event has Type "done" and includes the full accumulated ChatResponse.
    ChatCompletionStream(ctx context.Context, params ChatParams) (<-chan StreamEvent, error)
}
```

**Why these types:** The `ContentBlock` / `ToolCall` / `ToolResult` structure mirrors what both Anthropic and OpenAI use, just with different wire formats. Each provider maps to/from its native SDK types.

### Task 1.2 — Implement AnthropicProvider

**New file:** `backend/internal/agent/provider_anthropic.go`

Wraps the existing `anthropic.Client`. Two methods:

**`ChatCompletion`** — translates `ChatParams` → `anthropic.MessageNewParams`, calls `Messages.New()`, translates `*anthropic.Message` → `*ChatResponse`. This is a mechanical extraction of what currently happens inline in `loop.go:75-86`.

Translation mapping:
| Provider type | Anthropic SDK type |
|---|---|
| `ChatMessage{Role: "user", TextContent: "..."}` | `anthropic.NewUserMessage(anthropic.NewTextBlock(...))` |
| `ChatMessage{Role: "assistant", Blocks: [...]}` | `anthropic.NewAssistantMessage(...)` with `TextBlock` / `ToolUseBlockParam` |
| `ChatMessage{Role: "user", ToolResults: [...]}` | `anthropic.NewUserMessage(anthropic.NewToolResultBlock(...), ...)` |
| `ToolDef` | `anthropic.ToolUnionParam{OfTool: &anthropic.ToolParam{...}}` |
| `anthropic.StopReasonEndTurn` | `StopReasonEndTurn` |
| `anthropic.StopReasonToolUse` | `StopReasonToolUse` |

**`ChatCompletionStream`** — uses `Messages.NewStreaming()` from the Anthropic SDK. Spawns a goroutine that reads from the stream and writes `StreamEvent`s to the returned channel. Accumulates the full response and sends a final `"done"` event.

### Task 1.3 — Convert tool definitions to provider-agnostic format

**Modified file:** `backend/internal/agent/tools.go`

Currently `ToolRegistry()` returns `[]anthropic.ToolUnionParam` (line 12). Change to:

```go
// ToolRegistry returns provider-agnostic tool definitions.
func ToolRegistry() []ToolDef {
    return []ToolDef{
        {
            Name:        "search_logs",
            Description: "Search recent logs...",
            Parameters: map[string]any{
                "query":    map[string]any{"type": "string", "description": "..."},
                "severity": map[string]any{"type": "string", "description": "..."},
                "limit":    map[string]any{"type": "integer", "description": "..."},
            },
            Required: []string{"query"},
        },
        // ... search_codebase, query_database
    }
}
```

The `AnthropicProvider` converts these to `anthropic.ToolUnionParam` internally. `OpenRouterProvider` will convert to OpenAI function-calling format.

### Task 1.4 — Refactor the agent loop to use Provider

**Modified file:** `backend/internal/agent/loop.go`

**`RunConversation` (line 29)** — currently builds `anthropic.MessageParam` slices and calls `a.client.Messages.New()` directly. Refactor to:

1. Resolve provider via `a.providerFor(providerName)` (new method on Agent, see Task 1.5)
2. Build `[]ChatMessage` instead of `[]anthropic.MessageParam`
3. Call `provider.ChatCompletion(ctx, params)` instead of `a.client.Messages.New(ctx, ...)`
4. Process `*ChatResponse` instead of `*anthropic.Message`
5. Tool result handling: build `ChatMessage{Role: "user", ToolResults: [...]}` instead of `anthropic.NewUserMessage(toolResults...)`

The **loop structure, tool dispatch, EmitLog calls, and iteration limit** stay exactly the same. Only the types flowing through the loop change.

**`RunMonitoring` (line 185)** — same refactor but simpler (no history).

> ⚠️ **Streaming boundary — read this before touching the monitoring path.**
> `RunMonitoring` MUST call `provider.ChatCompletion()` (blocking) and MUST NOT call `provider.ChatCompletionStream()`. Streaming is an interactive-chat-only feature. Reasons:
> 1. **No human is watching.** Monitoring runs on a 15-second background tick with no UI consumer. Streaming would add SSE parsing and partial-state bookkeeping for zero perceived benefit.
> 2. **The 0.27.0 rate limiter assumes atomic calls.** `a.limiter.Wait(ctx)` wraps each `RunMonitoring` invocation as a single unit. Streaming muddies that contract (does a stream count as one token? what about mid-stream errors and retries?) and would force a rework of the cost-bound guarantee.
> 3. **Symmetry is not a goal here.** The asymmetry lives in the *callers* (`loop.go` chat path streams; `monitor.go` blocks), not in the `Provider` interface — both providers still implement both methods. Do not "unify" the two paths for tidiness.
>
> If you ever want a live tail of monitoring activity in the UI, that is a *separate* WebSocket pushing `agent_log` rows as they are written — not LLM token streaming. Keep those two ideas distinct.

**`extractText` (line 315)** — adapts to `ChatResponse.Blocks` instead of `anthropic.Message.Content`. Trivial.

### Task 1.5 — Update Agent struct and initialization

**Modified file:** `backend/internal/agent/agent.go`

```go
type Agent struct {
    queries      *db.Queries
    providers    map[string]Provider  // was: client *anthropic.Client
    config       *config.Config
    classifier   Classifier
    notifier     *notifications.Dispatcher
    githubClient *github.Client
    limiter      *rate.Limiter
    cancel       context.CancelFunc
    wg           sync.WaitGroup
}

func New(queries *db.Queries, cfg *config.Config, ...) *Agent {
    providers := make(map[string]Provider)

    // Anthropic provider (always available — ANTHROPIC_API_KEY is required)
    providers["anthropic"] = NewAnthropicProvider(cfg.AnthropicKey)

    return &Agent{
        queries:   queries,
        providers: providers,
        // ...
    }
}

// providerFor resolves a provider by name. Falls back to "anthropic".
func (a *Agent) providerFor(name string) Provider {
    if p, ok := a.providers[name]; ok {
        return p
    }
    return a.providers["anthropic"]
}
```

### Task 1.6 — Update chat.go to use streaming RunConversation

**Modified file:** `backend/internal/api/handlers/chat.go`

This is where streaming hits the WebSocket. Currently (line 176):

```go
response, err := s.Agent.RunConversation(ctx, userID, appID, &convID, history, msg.Content)
```

Change `RunConversation` signature for interactive mode to accept an events channel:

```go
// New signature for interactive (streaming) mode
func (a *Agent) RunConversationStream(ctx context.Context, userID uuid.UUID, appID uuid.UUID,
    conversationID *uuid.UUID, history []Message, input string) (<-chan AgentEvent, error)
```

Where `AgentEvent` is:

```go
type AgentEvent struct {
    Type    string // "chunk", "tool_start", "tool_result", "done", "error"
    Content string // text delta for "chunk", tool name for "tool_start", etc.
    Tool    string // tool name (for tool_start/tool_result)
}
```

**Chat handler changes (chat.go:164-208):**

Replace the current block:

```go
// Send "thinking" status
wsjson.Write(ctx, conn, map[string]string{"type": "status", "content": "thinking"})

// Run agent (blocking)
response, err := s.Agent.RunConversation(...)

// Send complete response
wsjson.Write(ctx, conn, chatMessage{...Content: response...})
```

With:

```go
// Send "thinking" status
wsjson.Write(ctx, conn, map[string]string{"type": "status", "content": "thinking"})

// Run agent (streaming)
events, err := s.Agent.RunConversationStream(ctx, userID, appID, &convID, history, msg.Content)
if err != nil { ... }

var fullContent strings.Builder
for event := range events {
    switch event.Type {
    case "chunk":
        fullContent.WriteString(event.Content)
        wsjson.Write(ctx, conn, map[string]string{
            "type":    "chunk",
            "content": event.Content,
        })
    case "tool_start":
        wsjson.Write(ctx, conn, map[string]string{
            "type": "tool_start",
            "tool": event.Tool,
        })
    case "tool_result":
        wsjson.Write(ctx, conn, map[string]string{
            "type":    "tool_result",
            "tool":    event.Tool,
            "content": event.Content,
        })
    case "error":
        wsjson.Write(ctx, conn, map[string]string{
            "type":    "error",
            "content": event.Content,
        })
    }
}

// Persist full response
agentMsg := chatMessage{
    ID:        uuid.New().String(),
    Role:      "assistant",
    Content:   fullContent.String(),
    Timestamp: time.Now().UTC().Format(time.RFC3339),
}
storedMessages = append(storedMessages, agentMsg)
s.persistMessages(ctx, convID, userID, storedMessages)

// Send "done" signal so frontend knows the message is complete
wsjson.Write(ctx, conn, map[string]string{
    "type": "done",
    "id":   agentMsg.ID,
})
```

**Key detail:** We still persist the **full accumulated response** to the conversations table. Streaming only changes delivery, not storage.

### Validation Gate — Phase 1

- [ ] All existing backend tests pass (`cd backend && go test ./...`)
- [ ] Agent chat works identically via WebSocket (now streaming token-by-token)
- [ ] Agent monitoring works identically (still blocking)
- [ ] No `anthropic-sdk-go` types leak outside `provider_anthropic.go`
- [ ] `tools.go` returns `[]ToolDef`, not `[]anthropic.ToolUnionParam`

---

## Phase 2 — OpenRouter Provider

**Goal:** Add a second provider that routes requests through OpenRouter's OpenAI-compatible API.

### Task 2.1 — Add OpenRouter config

**Modified file:** `backend/internal/config/config.go`

```go
type Config struct {
    // ... existing fields ...
    OpenRouterKey string  // add at line 27
}
```

In `Load()` (line 29):
```go
OpenRouterKey: getEnv("OPENROUTER_API_KEY", ""),
```

In `Validate()` (line 52): **no change** — OpenRouter key is optional.

**Modified file:** `backend/.env.example`

Add:
```
# OpenRouter — optional, enables multi-model selection
# Get your key at https://openrouter.ai/keys
OPENROUTER_API_KEY=
```

### Task 2.2 — Implement OpenRouterProvider

**New file:** `backend/internal/agent/provider_openrouter.go`

Implements `Provider` using raw HTTP against `https://openrouter.ai/api/v1/chat/completions`.

**Why raw HTTP instead of an OpenAI SDK:** Avoids adding a second large SDK dependency. The API surface we need is small (one endpoint, one streaming endpoint). A focused HTTP client is ~150 lines and has zero transitive dependencies.

**Request translation (`ChatParams` → OpenAI format):**

```json
{
  "model": "anthropic/claude-sonnet-4",
  "max_tokens": 4096,
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."},
    {"role": "assistant", "content": "...", "tool_calls": [...]},
    {"role": "tool", "tool_call_id": "...", "content": "..."}
  ],
  "tools": [
    {"type": "function", "function": {"name": "search_logs", "description": "...", "parameters": {...}}}
  ],
  "stream": false
}
```

**Key translation differences from Anthropic:**

| Concern | Anthropic format | OpenAI/OpenRouter format |
|---------|-----------------|------------------------|
| System prompt | Separate `system` param | First message with `role: "system"` |
| Tool definitions | `input_schema` at top level | Nested under `function.parameters` |
| Tool use in response | `content[].type: "tool_use"` | `tool_calls[]` on the message |
| Tool results | `role: "user"` with `tool_result` blocks | `role: "tool"` with `tool_call_id` |
| Stop reason | `stop_reason: "tool_use"` | `finish_reason: "tool_calls"` |

**Required headers:**

```
Authorization: Bearer $OPENROUTER_API_KEY
HTTP-Referer: https://heimdall.app
X-Title: Heimdall
Content-Type: application/json
```

**`ChatCompletion`** — POST with `"stream": false`, parse JSON response, translate to `*ChatResponse`.

**`ChatCompletionStream`** — POST with `"stream": true`, parse SSE events (`data: {...}`), emit `StreamEvent`s to channel. SSE format is identical to OpenAI's streaming format.

**Error handling:**
- HTTP 429 (rate limit): wrap in a typed error so the agent loop can distinguish
- HTTP 402 (insufficient credits): surface as a user-visible error
- Non-2xx: wrap status + body in error

### Task 2.3 — Register OpenRouter provider at startup

**Modified file:** `backend/internal/agent/agent.go`

In `New()`, after the Anthropic provider:

```go
if cfg.OpenRouterKey != "" {
    providers["openrouter"] = NewOpenRouterProvider(cfg.OpenRouterKey)
    slog.Info("openrouter provider enabled")
}
```

### Validation Gate — Phase 2

- [ ] With `OPENROUTER_API_KEY` set, agent chat works with an OpenRouter model (manually test by setting model in DB to e.g. `openai/gpt-4o` and provider to `openrouter`)
- [ ] Without `OPENROUTER_API_KEY`, everything works as before (Anthropic only)
- [ ] Tool use works end-to-end through OpenRouter (search_logs → tool result → synthesis)
- [ ] Streaming works through OpenRouter WebSocket path

---

## Phase 3 — Database Migration

**Goal:** Add `provider` column to `app_agent_config` so each app can select its provider.

### Task 3.1 — Create migration

**New file:** `backend/migrations/022_agent_config_provider.up.sql`

```sql
ALTER TABLE app_agent_config
ADD COLUMN provider TEXT NOT NULL DEFAULT 'anthropic';
```

**New file:** `backend/migrations/022_agent_config_provider.down.sql`

```sql
ALTER TABLE app_agent_config
DROP COLUMN provider;
```

### Task 3.2 — Update sqlc queries

**Modified file:** `backend/internal/db/queries/app_agent_config.sql`

```sql
-- name: GetAppAgentConfig :one
SELECT * FROM app_agent_config WHERE app_id = $1;

-- name: UpsertAppAgentConfig :one
INSERT INTO app_agent_config (app_id, model, mode, schedule_interval_secs, system_prompt_override, provider)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (app_id) DO UPDATE
SET model = EXCLUDED.model,
    mode = EXCLUDED.mode,
    schedule_interval_secs = EXCLUDED.schedule_interval_secs,
    system_prompt_override = EXCLUDED.system_prompt_override,
    provider = EXCLUDED.provider,
    updated_at = now()
RETURNING *;
```

### Task 3.3 — Regenerate sqlc and update handlers

Run: `make sqlc-generate`

This will add `Provider string` to the `AppAgentConfig` struct in `models.go`.

**Modified file:** `backend/internal/api/handlers/applications.go`

In `GetAppAgentConfig` (line 130) — add `"provider": "anthropic"` to the defaults response.

In `updateAppAgentConfigRequest` (line 153) — add `Provider string` field.

In `UpdateAppAgentConfig` (line 160):
- Accept `provider` from request body
- Validate: must be `"anthropic"` or `"openrouter"` (or empty → default `"anthropic"`)
- Pass to `UpsertAppAgentConfig`

### Task 3.4 — Wire provider into agent loop

**Modified file:** `backend/internal/agent/loop.go`

In `RunConversation` (line 35-55), after loading `appCfg`:

```go
providerName := "anthropic" // default
if appCfg.Provider != "" {
    providerName = appCfg.Provider
}
provider := a.providerFor(providerName)
```

Same in `RunMonitoring` (line 185-188):

```go
providerName := "anthropic"
if appConfig.Provider != "" {
    providerName = appConfig.Provider
}
provider := a.providerFor(providerName)
```

### Validation Gate — Phase 3

- [ ] `make migrate-up` succeeds
- [ ] `make sqlc-generate` produces clean output
- [ ] Existing configs default to `provider = 'anthropic'`
- [ ] Setting `provider = 'openrouter'` + an OpenRouter model ID routes through OpenRouterProvider

---

## Phase 4 — Model List Endpoint

**Goal:** Expose available models for the frontend dropdown.

### Task 4.1 — Define curated model lists

**New file:** `backend/internal/agent/models.go`

```go
package agent

// ModelOption is returned by the model list endpoint.
type ModelOption struct {
    ID            string  `json:"id"`
    Name          string  `json:"name"`
    Provider      string  `json:"provider"`
    ContextLength int     `json:"context_length"`
    Pricing       Pricing `json:"pricing"`
}

type Pricing struct {
    Prompt     float64 `json:"prompt"`     // $ per million input tokens
    Completion float64 `json:"completion"` // $ per million output tokens
}

// AnthropicModels is the static list of directly-supported Anthropic models.
var AnthropicModels = []ModelOption{
    {ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6", Provider: "anthropic", ContextLength: 200_000, Pricing: Pricing{Prompt: 3, Completion: 15}},
    {ID: "claude-haiku-4-5-20251001", Name: "Claude Haiku 4.5", Provider: "anthropic", ContextLength: 200_000, Pricing: Pricing{Prompt: 1, Completion: 5}},
    {ID: "claude-opus-4-6", Name: "Claude Opus 4.6", Provider: "anthropic", ContextLength: 200_000, Pricing: Pricing{Prompt: 15, Completion: 75}},
}

// OpenRouterModels is the curated list of OpenRouter models verified for multi-turn tool use.
// Update this list as new models are tested and confirmed working.
var OpenRouterModels = []ModelOption{
    {ID: "anthropic/claude-sonnet-4", Name: "Claude Sonnet 4 (OR)", Provider: "openrouter", ContextLength: 200_000, Pricing: Pricing{Prompt: 3, Completion: 15}},
    {ID: "openai/gpt-4o", Name: "GPT-4o", Provider: "openrouter", ContextLength: 128_000, Pricing: Pricing{Prompt: 2.5, Completion: 10}},
    {ID: "openai/gpt-4o-mini", Name: "GPT-4o Mini", Provider: "openrouter", ContextLength: 128_000, Pricing: Pricing{Prompt: 0.15, Completion: 0.6}},
    {ID: "google/gemini-2.5-pro", Name: "Gemini 2.5 Pro", Provider: "openrouter", ContextLength: 1_000_000, Pricing: Pricing{Prompt: 1.25, Completion: 10}},
    {ID: "google/gemini-2.5-flash", Name: "Gemini 2.5 Flash", Provider: "openrouter", ContextLength: 1_000_000, Pricing: Pricing{Prompt: 0.15, Completion: 0.6}},
    {ID: "meta-llama/llama-4-maverick", Name: "Llama 4 Maverick", Provider: "openrouter", ContextLength: 128_000, Pricing: Pricing{Prompt: 0.2, Completion: 0.6}},
}
```

### Task 4.2 — Add handler and route

**New file:** `backend/internal/api/handlers/models.go`

```go
func (s *Server) GetAvailableModels(w http.ResponseWriter, r *http.Request) {
    models := agent.AnthropicModels
    if s.Config.OpenRouterKey != "" {
        models = append(models, agent.OpenRouterModels...)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(models)
}
```

**Modified file:** `backend/internal/api/router.go`

Add inside the protected route group (after line 84):

```go
r.Get("/models", s.GetAvailableModels)
```

### Validation Gate — Phase 4

- [ ] `GET /api/models` returns Anthropic models when no OpenRouter key
- [ ] `GET /api/models` returns Anthropic + OpenRouter models when key is set
- [ ] Response shape matches `ModelOption` struct

---

## Phase 5 — Frontend: Model Dropdown

**Goal:** Replace the text input with a grouped dropdown showing provider, context length, and pricing.

### Task 5.1 — Add types and API function

**New file:** `frontend/src/types/models.ts`

```typescript
export interface ModelOption {
  id: string
  name: string
  provider: 'anthropic' | 'openrouter'
  context_length: number
  pricing: {
    prompt: number     // $/M input tokens
    completion: number // $/M output tokens
  }
}
```

**New file:** `frontend/src/api/models.ts`

```typescript
import client from './client'
import type { ModelOption } from '@/types/models'

export async function getAvailableModels(): Promise<ModelOption[]> {
  const { data } = await client.get<ModelOption[]>('/models')
  return data
}
```

### Task 5.2 — Update AppAgentConfig type

**Modified file:** `frontend/src/types/organization.ts`

Add `provider` field to `AppAgentConfig`:

```typescript
export interface AppAgentConfig {
  app_id: string
  model: string
  provider: 'anthropic' | 'openrouter'  // new
  mode: 'continuous' | 'periodic' | 'off'
  schedule_interval_secs: number
  system_prompt_override: string | null
  created_at: string
  updated_at: string
}
```

### Task 5.3 — Rewrite AgentConfigPage model selection

**Modified file:** `frontend/src/pages/AgentConfigPage.vue`

**Script changes:**

Remove `knownModels` array (line 24-28).

Add:
```typescript
import { getAvailableModels } from '@/api/models'
import type { ModelOption } from '@/types/models'

const models = ref<ModelOption[]>([])
const formProvider = ref<'anthropic' | 'openrouter'>('anthropic')

// Computed: group models by provider for the dropdown
const anthropicModels = computed(() => models.value.filter(m => m.provider === 'anthropic'))
const openrouterModels = computed(() => models.value.filter(m => m.provider === 'openrouter'))
```

In `fetchConfig()` (line 37), also fetch models:
```typescript
async function fetchConfig() {
  const appId = appStore.currentAppId
  if (!appId) return
  loading.value = true
  try {
    const [cfg, availableModels] = await Promise.all([
      getAppAgentConfig(appId),
      getAvailableModels(),
    ])
    config.value = cfg
    models.value = availableModels
  } catch {
    toast.show('Failed to load agent config', 'error')
  } finally {
    loading.value = false
  }
}
```

In `startEdit()` (line 50), also set provider:
```typescript
formProvider.value = config.value.provider ?? 'anthropic'
```

In `saveConfig()` (line 63), include provider:
```typescript
config.value = await updateAppAgentConfig(appId, {
  model: formModel.value,
  provider: formProvider.value,
  mode: formMode.value,
  // ...
})
```

**Template changes:**

Replace the `<input>` + `<datalist>` block (lines 128-139) with:

```html
<div>
  <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Model</label>
  <select
    v-model="formModel"
    required
    @change="onModelSelect"
    class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors appearance-none cursor-pointer"
  >
    <optgroup label="Anthropic (Direct)">
      <option v-for="m in anthropicModels" :key="m.id" :value="m.id">
        {{ m.name }} — {{ formatContext(m.context_length) }} · ${{ m.pricing.prompt }}/${{ m.pricing.completion }}
      </option>
    </optgroup>
    <optgroup v-if="openrouterModels.length" label="OpenRouter">
      <option v-for="m in openrouterModels" :key="m.id" :value="m.id">
        {{ m.name }} — {{ formatContext(m.context_length) }} · ${{ m.pricing.prompt }}/${{ m.pricing.completion }}
      </option>
    </optgroup>
  </select>
</div>
```

Add helper functions:

```typescript
function onModelSelect() {
  // Auto-set provider based on selected model's group
  const selected = models.value.find(m => m.id === formModel.value)
  if (selected) {
    formProvider.value = selected.provider
  }
}

function formatContext(tokens: number): string {
  if (tokens >= 1_000_000) return `${tokens / 1_000_000}M`
  return `${tokens / 1_000}k`
}
```

In display mode (lines 212-215), show provider alongside model:

```html
<div class="flex items-baseline justify-between py-2 border-b border-border">
  <span class="font-mono text-xs font-medium uppercase tracking-wider text-text-muted">Model</span>
  <div class="text-right">
    <span class="font-mono text-sm text-text-primary">{{ config.model }}</span>
    <span class="font-mono text-xs text-text-muted ml-2">via {{ config.provider ?? 'anthropic' }}</span>
  </div>
</div>
```

### Validation Gate — Phase 5

- [ ] Dropdown shows Anthropic models grouped under "Anthropic (Direct)"
- [ ] If OpenRouter key is configured, dropdown also shows "OpenRouter" group
- [ ] Selecting a model auto-sets the provider
- [ ] Saving persists both model + provider to backend
- [ ] Display mode shows model + provider
- [ ] Pricing and context length visible in options

---

## Phase 6 — Frontend: Streaming Chat

**Goal:** Update the chat UI to render streamed tokens and show tool-use progress.

### Task 6.1 — Update useAgent composable

**Modified file:** `frontend/src/composables/useAgent.ts`

Add new message types to the `watch(data, ...)` handler (line 23-65):

```typescript
// Streaming chunk — append to the current in-progress message
if (parsed.type === 'chunk') {
  isThinking.value = false
  if (!currentStreamingMsg.value) {
    // First chunk — create a new message entry
    currentStreamingMsg.value = {
      id: crypto.randomUUID(),
      role: 'agent',
      content: parsed.content,
      timestamp: new Date().toISOString(),
    }
    messages.value.push(currentStreamingMsg.value)
  } else {
    // Subsequent chunks — append to existing message
    currentStreamingMsg.value.content += parsed.content
  }
  return
}

// Tool progress
if (parsed.type === 'tool_start') {
  isThinking.value = false
  activeTools.value.push(parsed.tool)
  return
}

if (parsed.type === 'tool_result') {
  activeTools.value = activeTools.value.filter(t => t !== parsed.tool)
  return
}

// Stream complete — finalize the message with the server-assigned ID
if (parsed.type === 'done') {
  if (currentStreamingMsg.value) {
    currentStreamingMsg.value.id = parsed.id
    currentStreamingMsg.value = null
  }
  activeTools.value = []
  return
}
```

New refs to export:

```typescript
const currentStreamingMsg = ref<ChatMessage | null>(null)
const activeTools = ref<string[]>([])

return { messages, status, conversationId, isThinking, error, activeTools, sendMessage, loadMessages }
```

### Task 6.2 — Update chat UI for streaming

The chat page component (find via `AgentChatPage.vue` or similar) needs minor updates:

1. **Tool progress indicator:** When `activeTools` is non-empty, show a small status bar below the thinking indicator:

```html
<div v-if="activeTools.length" class="flex items-center gap-2 text-text-muted font-mono text-xs px-4 py-2">
  <span class="animate-pulse">&#9679;</span>
  <span>{{ activeTools.map(t => t.replace('_', ' ')).join(', ') }}...</span>
</div>
```

2. **Remove "thinking" once chunks arrive:** Already handled — `isThinking` is set to `false` on first `chunk` event.

3. **Message rendering stays the same.** The existing markdown renderer already works with a growing string — Vue's reactivity re-renders as `content` grows.

### Validation Gate — Phase 6

- [ ] Chat shows tokens appearing incrementally (not all at once)
- [ ] Tool use shows "search logs..." / "query database..." progress indicator
- [ ] Final message is complete and matches what's persisted in the database
- [ ] Resuming a conversation (loading history) still works
- [ ] Error messages still display correctly

---

## Full Dependency Graph

```
Phase 1 (provider abstraction)
    │
    ├──→ Phase 2 (OpenRouter provider)
    │        │
    │        └──→ Phase 4 (model list endpoint)
    │                  │
    │                  └──→ Phase 5 (frontend dropdown)
    │
    ├──→ Phase 3 (DB migration)
    │        │
    │        └──→ Phase 5 (frontend dropdown)
    │
    └──→ Phase 6 (frontend streaming)
         (can start as soon as Phase 1 streaming is working)
```

**Critical path:** Phase 1 → Phase 2 → Phase 4 → Phase 5

**Parallel work:**
- Phase 3 can run in parallel with Phase 2
- Phase 6 can start as soon as Phase 1 is complete (streaming via Anthropic works before OpenRouter exists)

---

## Files Summary

### New Files (9)

| File | Phase | Purpose |
|------|-------|---------|
| `backend/internal/agent/provider.go` | 1 | Interface + domain types |
| `backend/internal/agent/provider_anthropic.go` | 1 | Anthropic SDK wrapper |
| `backend/internal/agent/provider_openrouter.go` | 2 | OpenRouter HTTP client |
| `backend/internal/agent/models.go` | 4 | Curated model lists |
| `backend/internal/api/handlers/models.go` | 4 | GET /api/models handler |
| `backend/migrations/022_agent_config_provider.up.sql` | 3 | Add provider column |
| `backend/migrations/022_agent_config_provider.down.sql` | 3 | Rollback |
| `frontend/src/types/models.ts` | 5 | ModelOption type |
| `frontend/src/api/models.ts` | 5 | API client function |

### Modified Files (11)

| File | Phase | Change |
|------|-------|--------|
| `backend/internal/agent/agent.go` | 1, 2 | `providers` map replaces `client` field |
| `backend/internal/agent/loop.go` | 1, 3 | Use Provider interface + provider resolution |
| `backend/internal/agent/monitor.go` | 1, 3 | Use Provider (non-streaming) + provider resolution |
| `backend/internal/agent/tools.go` | 1 | Return `[]ToolDef` instead of SDK types |
| `backend/internal/config/config.go` | 2 | Add `OpenRouterKey` |
| `backend/internal/api/router.go` | 4 | Add `/api/models` route |
| `backend/internal/api/handlers/applications.go` | 3 | Accept/return `provider` field |
| `backend/internal/db/queries/app_agent_config.sql` | 3 | Add `provider` to queries |
| `backend/.env.example` | 2 | Add `OPENROUTER_API_KEY` |
| `frontend/src/types/organization.ts` | 5 | Add `provider` to `AppAgentConfig` |
| `frontend/src/pages/AgentConfigPage.vue` | 5 | Dropdown replaces text input |
| `frontend/src/composables/useAgent.ts` | 6 | Streaming chunk + tool progress handling |

### Regenerated Files

| File | Phase | Trigger |
|------|-------|---------|
| `backend/internal/db/models.go` | 3 | `make sqlc-generate` |
| `backend/internal/db/app_agent_config.sql.go` | 3 | `make sqlc-generate` |
