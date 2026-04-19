# Agent Architecture

A single-document overview of Heimdall's LLM agent — what it is, how it's wired, which models run where, and how work flows from a log arriving at the edge all the way to an LLM assessment landing in the Activity feed.

Audience: new LLMs opening this repo cold, senior engineers trying to orient themselves before changing anything in `backend/internal/agent/`, and product managers who need a mental model without reading Go.

---

## 1. What "the agent" actually is

In Heimdall, "the agent" is not a single daemon. It's a **struct** (`agent.Agent`) with a handful of goroutines hanging off it, and three **entry points** that all funnel into the same underlying tool-use loop. One struct, three modes.

The struct is constructed once at boot and injected into both the API server (for interactive chat) and its own background goroutines (for monitoring, scheduling, and retention pruning). All three modes share:

- The same **tool registry** (`search_logs`, `query_database`, `search_codebase`)
- The same **provider abstraction** (Anthropic direct, or OpenRouter)
- The same **rate limiter** (30 rpm / burst 5 across monitoring + scheduler)
- The same **Activity feed emission** (`agent_log` rows via `EmitLog`)

What changes across modes is: *who holds the receiver*, *what prompt the loop starts with*, *whether a human is waiting on the other end*, and *whether conversation history persists*.

```
                ┌────────────────────────────────────────┐
                │              agent.Agent               │
                │  (one per process, constructed in      │
                │   main.go, owns background goroutines) │
                └────────────────────────────────────────┘
                              │
         ┌────────────────────┼──────────────────────────┐
         ▼                    ▼                          ▼
   Interactive            Monitoring                 Scheduled
   (user in loop)      (background poll)          (cron / interval)

   ws://ws/chat         15s ticker goroutine       1m ticker goroutine
   RunConversation-     Monitor + monitorApp       InvestigationScheduler
   Stream                + RunMonitoring            + RunScheduledInvestigation
                                                     (reuses RunMonitoring)
```

---

## 2. The three operating modes

Every LLM invocation in Heimdall belongs to exactly one of these.

### 2.1 Interactive mode — a human in the loop

**Trigger:** A user opens the chat page. The browser opens a WebSocket to `/ws/chat?token=<jwt>&app_id=<uuid>&conversation_id=<uuid>` and sends a text message.

**Handler:** `backend/internal/api/handlers/chat.go`
- Validates the JWT, accepts the WebSocket, loads/creates a `conversations` row (JSONB messages column, user-scoped), hydrates prior history, persists the new user message, then calls `s.Agent.RunConversationStream(...)`.
- Streams progress events (`tool_start`, `tool_result`, `message`, `error`) back to the browser as the loop iterates.

**Agent entry point:** `RunConversationStream` (streaming) or `RunConversation` (blocking). Both funnel into `runConversationCore` — see `backend/internal/agent/loop.go:68` and `loop_stream.go:54`. The streaming wrapper only streams *loop progress* (tool transitions + a final full message); the underlying `ChatCompletion` is still a blocking call. Token-by-token streaming is flagged as a follow-up in `loop_stream.go:14`.

**Prompt:** `BuildSystemPrompt(override)` — the general-purpose "You are Heimdall" system prompt, optionally appended with a per-app override from `app_agent_config.system_prompt_override`. Defined in `backend/internal/agent/prompt.go:3`.

**Model selection:** Prefers per-app config (`GetAppAgentConfig`) when `app_id` is provided, else falls back to the global `agent_config` singleton, else `DefaultModelID = "claude-sonnet-4-6"`. Same goes for the provider (`anthropic` / `openrouter`) and the system-prompt override.

**Max iterations:** 10 (`maxIterations` in `loop.go:17`). On overflow, returns an error rather than a partial answer.

**Persistence:** Every user/assistant turn is written to `conversations.messages` (JSONB). Tool calls, tool results, and the final agent response are *also* emitted to `agent_log` with `entry_type` in `{"tool_call", "tool_result", "observation"}` so they show up in the Activity feed.

### 2.2 Monitoring mode — no human, classifier-driven

**Trigger:** A 15-second ticker goroutine spawned by `agent.Start()`. See `monitor.go:30`.

**Loop:** Every tick, `monitorTick` calls `ListActiveApplications`, then for each app (up to 10 in parallel via a semaphore) runs `monitorApp` with a 2-minute per-app timeout:

1. Resolve a user ID for `agent_log` attribution (`GetFirstUserInOrg` — first user in the app's org).
2. Load the app's monitoring cursor (`monitoring_state.last_monitored_at`). First run initializes the cursor to "now" and skips — we never retroactively process historical logs.
3. Fetch up to 200 new logs since the cursor (`ListLogsSinceForApp`).
4. Run them through the **classifier pipeline** (see §5). This is where most traffic gets dropped before ever touching an LLM.
5. If any logs are flagged: cap at 50, acquire a rate-limit token, call `RunMonitoring(ctx, userID, appConfig, formattedLogs)`.
6. Emit the assessment to `agent_log` with `entry_type = "monitoring"` and the parsed severity, then fire-and-forget a notification dispatch (Slack / email / webhook, depending on app config).
7. Advance the cursor to the last processed log's `ingested_at`.

**Periodic vs continuous:** Apps have a `mode` column. `continuous` runs every tick (every 15s); `periodic` respects `schedule_interval_secs` and only runs when `elapsed >= interval`. See `shouldMonitor` in `monitor.go:83`.

**Prompt:** `BuildMonitoringPrompt(override)` — a much stricter prompt than interactive. It's structured around a tiny three-line output format (`Assessment:` / `Severity:` / `Action:`) so the backend can regex the severity out of the response (`parseSeverityFromResponse` in `loop.go:339`) and route alerts accordingly. Defined in `prompt.go:23`.

**Sessionless:** No conversation history. Each monitoring invocation is a fresh message list: one `user` message containing the formatted flagged-log batch, and the loop runs until `StopReasonEndTurn` or the iteration cap.

**Why the monitoring path stays blocking (non-streaming):** Flagged in `loop.go:227`. The rate limiter treats each `RunMonitoring` call as a single atomic unit that counts against the 30-rpm budget. Streaming would muddy that contract — a partially streamed response that gets cancelled mid-way would still have consumed a token.

### 2.3 Scheduled mode — proactive probing of pull-only sources

**Trigger:** A 1-minute ticker goroutine, also spawned by `agent.Start()`. See `scheduler.go:59`.

**Why it exists:** Monitoring mode is **reactive** — it fires when new logs arrive via a push-style connector. But `QueryConnector`s like Postgres and GitHub are *pull-only*: you have to run a query to see state. The monitoring loop has no logs to flag there, so nothing would ever trigger investigation of them. Scheduled mode closes that gap. Example schedule: "every hour, check `pg_stat_statements` and summarise the worst offenders".

**Loop:** Every tick, `schedulerTick` calls `ListEnabledSchedules` and fires any whose `shouldFire` returns true. Runs serially (not in parallel) because the 1-minute tick interval has plenty of headroom and the shared rate limiter would serialize them anyway.

**Firing decision (`shouldFire` in `scheduler.go:109`):**
1. Never-run → fire.
2. `cron_expr` set → parse with `robfig/cron/v3`, fire if `next(last_run_at) <= now`. Malformed expressions log and return `false` (we'd rather have a stuck schedule the user can fix via the UI than an infinite error loop).
3. Else → legacy Phase 3 path: fire if `now - last_run_at >= interval_secs`.

**Entry point:** `RunScheduledInvestigation(ctx, schedule)` in `scheduler.go:146`. The big architectural move here is that it **reuses `RunMonitoring`** as its LLM call. The schedule's stored prompt gets passed verbatim as the `flaggedLogs` parameter (the name is misnamed for this path — `RunMonitoring` doesn't care, it just feeds the string to the LLM as the first user message). Without this reuse, every future improvement to `RunMonitoring` — provider logic, rate limiting, tool dispatch, `agent_log` emission — would have to be ported to a parallel loop body.

**Emission:** Writes to `agent_log` with a distinct `entry_type = "scheduled_investigation"` so the UI can render scheduled rows differently from classifier-driven monitoring rows. Updates `schedules.last_run_at` / `last_status` / `last_summary` / `last_error` via `MarkScheduleRun`.

**Manual run-now:** `POST /api/apps/{appId}/schedules/{id}/run-now` calls `RunScheduledInvestigation` synchronously from the HTTP handler, bypassing the tick loop. The frontend uses per-row `runningId` state to disable the button while it's in flight.

### 2.4 The pruner (bonus goroutine)

Not an agent mode, but it's spawned alongside them by `agent.Start()`. `Prune` in `pruner.go:21` calls `PruneExpiredLogs` once on startup and then every hour. The retention window (48h) lives in the SQL query itself, not in Go, so the DB is the single source of truth. **Historical note:** This was dead code for ~20 releases — see changelog 0.30.3. `log_buffer` had been growing unbounded in prod until it was finally wired up.

---

## 3. The shared loop body

Interactive and monitoring mode each have their own public entry point (`RunConversation` / `RunMonitoring`), but the *structure* of the tool-use loop is identical. Understanding one means you understand the other.

```
┌─ preflight ────────────────────────────────────────────────────┐
│  Resolve model, system prompt, provider from app/global config │
│  Build neutral []ChatMessage from history + new input          │
│  Load ToolRegistry()                                           │
└────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─ iteration (up to maxIterations = 10) ─────────────────────────┐
│                                                                │
│   provider.ChatCompletion(ctx, ChatParams{...})                │
│                              │                                 │
│                              ▼                                 │
│    ┌────────────────┐    ┌──────────────┐                      │
│    │ StopReason =   │    │ StopReason = │                      │
│    │   EndTurn      │    │   ToolUse    │                      │
│    └───────┬────────┘    └──────┬───────┘                      │
│            │                    │                              │
│            ▼                    ▼                              │
│  extract text, emit     For each tool_use block:               │
│  observation to          - emit tool_call to agent_log         │
│  agent_log, return       - emit "tool_start" stream event      │
│                          - a.Dispatch(userID, appID, name)     │
│                          - on error: ToolResult{IsError:true}  │
│                          - on success: ToolResult{Content:...} │
│                          - emit tool_result to agent_log       │
│                          - emit "tool_result" stream event     │
│                         Append assistant + tool_results to     │
│                         messages, loop                         │
└────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                   "exceeded max iterations"
```

**Why two public entry points instead of one?** Monitoring is sessionless and needs the strict `Assessment / Severity / Action` prompt, plus a severity-extraction pass on the final text. Interactive mode needs conversation hydration, `agent_log` emissions keyed by `conversationID`, and optional streaming of progress events. Collapsing them would require parameter soup at the callsite. Keeping them split means each reads cleanly; the duplication is mostly cosmetic.

**Key file:** `backend/internal/agent/loop.go`. `runConversationCore` (line 68) is the interactive loop body with optional streaming via an events channel. `RunMonitoring` (line 231) is the monitoring loop body. They look ~80% the same on purpose.

---

## 4. The provider abstraction

Everything above talks to models through a **single interface**:

```go
type Provider interface {
    ChatCompletion(ctx context.Context, params ChatParams) (*ChatResponse, error)
}
```

Defined in `backend/internal/agent/provider.go:92`. The agent loop never imports any LLM SDK directly — it only speaks `ChatParams`, `ChatMessage`, `ContentBlock`, `ToolDef`, and `ChatResponse`. These are provider-neutral types.

Two implementations ship today:

### 4.1 AnthropicProvider (`provider_anthropic.go`)

- Wraps the official `github.com/anthropics/anthropic-sdk-go`.
- Translates `ChatParams → anthropic.MessageNewParams` and the response blocks back to neutral `ContentBlock`s.
- The only file in the agent package that imports the Anthropic SDK.
- Always registered at boot — `ANTHROPIC_API_KEY` is a required config value.
- Model IDs: `claude-sonnet-4-6` (default), `claude-haiku-4-5-20251001`, `claude-opus-4-6`. See `models.go:27`.

### 4.2 OpenRouterProvider (`provider_openrouter.go`)

- Talks to OpenRouter's OpenAI-compatible `/chat/completions` endpoint via **raw `net/http`** — no second SDK dependency.
- Translates both directions, including the shape differences: system prompt becomes the first `role:"system"` message, tool results become `role:"tool"` messages, tool-call arguments are JSON-encoded strings (not objects), `finish_reason = "tool_calls"` maps to `StopReasonToolUse`.
- **Opt-in:** only registered when `OPENROUTER_API_KEY` is set. Without the key, the model dropdown hides OpenRouter models and any legacy app config requesting `"openrouter"` silently falls back to `anthropic` via `providerFor`'s default branch.
- Response body is bounded to 10 MiB (`openRouterMaxBodyBytes`) to prevent OOM from a broken/hostile upstream.
- Curated model list in `models.go:59`. Adding a model is a deliberate code change, not a runtime catalogue fetch — OpenRouter exposes hundreds of models and many of them silently fail tool-use round-trips.

### 4.3 Provider routing (`providerFor`)

One short function, `agent.go:64`:

```
appConfig.Provider ──► providers[name] ──► found? use it : fall back to anthropic
```

- Empty string → default.
- Unknown name → default (defensive — lets a legacy DB row referencing `"openrouter"` keep working even if the key gets unset).
- The default is always `anthropic`.

### 4.4 Who picks which model

| Mode | Source of model/provider |
|---|---|
| Interactive (with `app_id`) | `app_agent_config` for that app |
| Interactive (no `app_id`) | Global `agent_config` singleton |
| Monitoring | `app_agent_config` for the app being monitored |
| Scheduled | `app_agent_config` for the app the schedule belongs to |

In every case, a missing value falls through to `DefaultModelID = "claude-sonnet-4-6"` and provider `"anthropic"`. The default is exported from `loop.go:23` specifically so `handlers/applications.go` and `handlers/agent_config.go` can reference the same constant and not drift on the next Sonnet release.

`GET /api/models` (`handlers/models.go`) returns the curated list, merging Anthropic + OpenRouter based on whether the key is configured. The frontend renders this in the Agent Config dropdown.

---

## 5. The classifier pipeline (monitoring-only)

Monitoring mode runs a deterministic **pre-filter** before ever touching an LLM. This is the most important cost-control lever in the product: a typical app produces tens of thousands of logs per tick, almost all of which are routine `200 OK`s. We'd go bankrupt sending every log to Claude.

```
raw logs  ─► Classifier.Classify(logs)  ─► (flagged []ClassifiedLog, safeCount int)
                                                     │
                                                     ▼
                                           len(flagged) > 0 ?
                                                     │
                                              yes ──┴── no
                                               │         │
                                               ▼         ▼
                                       rate limiter   advance cursor, done
                                               │
                                               ▼
                                         RunMonitoring(...)
```

### 5.1 The Classifier interface

```go
type Classifier interface {
    Classify(logs []db.LogBuffer) (flagged []ClassifiedLog, safeCount int)
    Close() error
}
```

Defined in `classifier.go:9`. Two implementations:

- **`LumberClassifier`** (`classifier_lumber.go`) — wraps the Lumber ONNX model (`github.com/kaminocorp/lumber`). Classifies each log into a root type (`ERROR`, `REQUEST`, `DEPLOY`, `SYSTEM`, `ACCESS`, `PERFORMANCE`, `UNCLASSIFIED`) + leaf category + severity + confidence score. If `confidence < 0.5` OR `ShouldEscalate(event) == true`, the log is flagged. On model failure, *every* log gets escalated (safe default — we never silently drop).
- **`PassthroughClassifier`** (`classifier.go:28`) — flags everything unconditionally. Used when classification is disabled or when Lumber can't load. Every log gets escalated → every log gets the 30-rpm rate limiter applied upstream (this is why that rate limit exists at all).

### 5.2 The severity gate (`severity_gate.go`)

`ShouldEscalate(event)` is a hardcoded switch table over Lumber's taxonomy. Condensed:

| Root type | Escalation rule |
|---|---|
| `ERROR` | always escalate |
| `PERFORMANCE` | always escalate |
| `DEPLOY` | always escalate |
| `REQUEST` | only `server_error` and `slow_request` |
| `SYSTEM` | only `resource_alert` and `config_change` |
| `ACCESS` | only `login_failure`, `auth_failure`, `permission_change`, `api_key_event` |
| `UNCLASSIFIED` | always escalate (safe default) |

This table is the single source of truth for "what's worth paging Claude about", and it's intentionally *not* in the database or in config — changing escalation policy should be a code review, not a runtime tweak.

### 5.3 Which classifier is chosen (`main.go`)

Controlled by `CLASSIFIER_MODE` env var. Selection happens once at boot:

| Mode | Behaviour |
|---|---|
| `on` | Must load Lumber. Fails boot if the model is unavailable. |
| `off` | Skips Lumber entirely, uses `PassthroughClassifier`. Useful for debugging or air-gapped dev. |
| `fallback` (default) | Tries Lumber, falls back to passthrough if it can't load. |

---

## 6. The tool registry

Every LLM invocation (all three modes) gets the same set of tools. Defined in `backend/internal/agent/tools.go:12`:

| Tool | Implementation | Purpose |
|---|---|---|
| `search_logs` | `tools_logs.go` | Searches the user-scoped `log_buffer` table. Supports `query` (LIKE), `severity`, `limit`. |
| `query_database` | `tools_db.go` | Runs read-only SQL against a user's connected Postgres (`default_transaction_read_only=on`). Requires a `connection_id` UUID. |
| `search_codebase` | `tools_codebase.go` | Proxies through the GitHub connector. Actions: `search_code`, `read_file`, `list_tree`. Requires an `app_id` (tools are app-scoped) and at least one enabled GitHub repo for that app. |

Dispatch is a single switch in `tools.go:80`:

```go
func (a *Agent) Dispatch(ctx, userID, appID, name, input) (string, error) {
    switch name {
    case "search_logs":     return a.toolSearchLogs(...)
    case "query_database":  return a.toolQueryDatabase(...)
    case "search_codebase": return a.toolSearchCodebase(...)
    default:                return "", fmt.Errorf("unknown tool: %s", name)
    }
}
```

**Memory tools (`tools_memory.go`) are deliberately unimplemented** — the file is one comment line long. `recall_similar_incidents` and `recall_lessons` are flagged for a post-MVP integration with Elephantasm (the long-term memory system referenced in the vision doc).

**Tool error contract:** Errors from `Dispatch` are returned to the LLM as `ToolResult{IsError: true, Content: "Tool error: <msg>"}` — never thrown. This is load-bearing: throwing would bail out of the entire loop, whereas returning an error result lets the model recover, retry with different parameters, or acknowledge failure and still produce a final assessment.

---

## 7. Concurrency and resource bounds

The agent touches production LLM APIs under variable load. Several safety nets stack up:

| Bound | Where | Why |
|---|---|---|
| `maxIterations = 10` | `loop.go:17` | Caps any single tool-use loop — runaway agents can't blow through the rate limit on one request. |
| `maxConcurrentApps = 10` | `monitor.go:22` | Semaphore in `monitorTick` — can't spawn unbounded goroutines if 1000 apps all need monitoring. |
| `monitorAppTimeout = 2 * time.Minute` | `monitor.go:25` | Per-app ctx timeout — a wedged LLM call can't pin a goroutine forever. |
| `schedulerRunTimeout = 5 * time.Minute` | `scheduler.go:44` | Same idea for scheduled runs, longer because they run serially. |
| `monitorLLMRate = 30 rpm, burst 5` | `agent.go:20` | Shared between monitoring and scheduler — a dozen noisy schedules can't starve the monitor loop. Makes the worst case (passthrough classifier + every log escalated) financially survivable. |
| `logBatchLimit = 200` | `monitor.go:22` | Caps how many logs get pulled per tick per app. |
| `maxFlaggedForLLM = 50` | `monitor.go:23` | Caps how many *flagged* logs get sent to the LLM — if 200 logs are all flagged, we send the first 50 and log a warning. |
| `maxPayloadChars = 2000` | `monitor.go:24` | Truncates the raw payload of each log before it hits the prompt. |
| `openRouterMaxBodyBytes = 10 MiB` | `provider_openrouter.go:37` | Response body cap — bounds memory on hostile upstreams. |
| `streamBufferSize = 16` | `loop_stream.go:40` | Channel buffer between the interactive loop producer and the WebSocket consumer. |

The rate limiter is the most important of these. It's shared between modes specifically so that scheduler growth can't starve monitoring of its LLM budget — both compete for the same 30-rpm pool.

---

## 8. Lifecycle: boot → run → shutdown

### Boot (`backend/cmd/heimdall/main.go`)

```
1. config.Load() + Validate()
2. Build classifier based on CLASSIFIER_MODE
3. pgxpool.New → db.New(pool) → queries
4. notifications.NewDispatcher(queries, cfg) → notifier
5. github.NewClient(...) → ghClient  (nil-safe — optional)
6. agent.New(queries, cfg, classifier, notifier, ghClient) → ag
7. ag.Start(ctx)     ← spawns 3 goroutines
8. api.NewRouter(...) with Agent injected
9. http.Server.ListenAndServe()
```

`agent.New` (`agent.go:37`) constructs the provider map: `anthropic` is always registered, `openrouter` is conditionally added if its key is set. The rate limiter is built here once and reused by all modes.

### Run (`agent.Start` in `agent.go:77`)

Spawns three goroutines, tracked by a `sync.WaitGroup`:

1. `Monitor(ctx)` — the 15s monitoring loop.
2. `Prune(ctx)` — the 1h log retention pruner.
3. `InvestigationScheduler(ctx)` — the 1m scheduler loop.

Idempotent-ish: if called while already running, it logs a warning, stops the previous instance, and restarts. The cancel function is stored on the Agent struct so `Stop` can cancel all three at once.

### Shutdown (`agent.Stop` in `agent.go:106`)

1. Call the stored `cancel()` — every goroutine watches `ctx.Done()`.
2. `wg.Wait()` — blocks until all three exit.
3. Log "agent stopped".

The Lumber classifier has its own `Close()` to release ONNX runtime memory; that's called separately from `main.go` during graceful shutdown.

---

## 9. The Activity feed (`agent_log`)

Every LLM invocation produces rows in the `agent_log` table. This is how the Activity page in the frontend reconstructs what the agent has been doing.

| `entry_type` | Emitted by | Example summary |
|---|---|---|
| `observation` | Interactive mode (final response) | Truncated to 200 runes |
| `tool_call` | All modes | `"Called search_logs"` with `{tool, input}` in detail |
| `tool_result` | All modes | `"search_logs returned results"` with `{tool, result_preview}` |
| `monitoring` | Monitoring mode (final assessment) | `{app_id, app_name, flagged, assessment, auto_severity}` — carries parsed severity |
| `scheduled_investigation` | Scheduled mode (final assessment) | `{schedule_id, schedule_name, app_id, app_name, assessment, auto_severity}` |

All writes go through `EmitLog` / `EmitLogWithSeverity` (`emit.go`). **Fire-and-forget by design** — a failed insert is logged and swallowed, never propagated. Logging failures must never degrade agent work. This is called out explicitly in `emit.go:15`.

The `conversation_id` field is optional: interactive mode fills it (so tool calls and observations attach to the chat thread); monitoring and scheduled modes leave it null (no conversation exists).

---

## 10. File-by-file guide to `backend/internal/agent/`

The agent package is flat by design — 25 files, mostly paired with tests. If you're lost, start here.

**Core struct and lifecycle**
- `agent.go` — `Agent` struct, `New`, `Start`, `Stop`, `providerFor`, rate limiter wiring. Start here.

**Loop bodies**
- `loop.go` — `RunConversation`, `RunLoop`, `runConversationCore`, `RunMonitoring`, `parseSeverityFromResponse`, `extractText`. The heart of the agent.
- `loop_stream.go` — `RunConversationStream` + `AgentEvent`. Thin streaming wrapper over `runConversationCore`.
- `loop_test.go` — Unit tests via a mocked Anthropic client, including the OpenRouter-routing smoke test.

**Background goroutines**
- `monitor.go` — The 15s monitoring loop. `Monitor`, `monitorTick`, `monitorApp`, `shouldMonitor`, `formatFlaggedLogs`.
- `scheduler.go` — The 1m scheduler loop. `InvestigationScheduler`, `schedulerTick`, `shouldFire`, `RunScheduledInvestigation`, `markRunSuccess`, `markRunError`, `ParseCronExpression`.
- `pruner.go` — The 1h log-retention pruner. Unrelated to LLM work but lives here because it shares the agent's lifecycle.

**Provider abstraction**
- `provider.go` — Neutral `Provider` interface + `ChatParams`, `ChatMessage`, `ContentBlock`, `ToolCall`, `ToolResult`, `ToolDef`, `ChatResponse`, `Usage`, `StopReason`.
- `provider_anthropic.go` — `AnthropicProvider`, the only file importing the Anthropic SDK.
- `provider_openrouter.go` — `OpenRouterProvider` over raw HTTP. ~350 lines including wire types.
- `models.go` — `AnthropicModels` + `OpenRouterModels` curated catalogues, surfaced via `GET /api/models`.

**Prompts**
- `prompt.go` — `systemPrompt` (interactive) and `monitoringSystemPrompt` (monitoring). Both support per-app overrides.

**Tools**
- `tools.go` — `ToolRegistry` + `Dispatch`.
- `tools_logs.go` — `search_logs` implementation.
- `tools_db.go` — `query_database` implementation (read-only Postgres connector).
- `tools_codebase.go` — `search_codebase` implementation (GitHub connector, app-scoped).
- `tools_memory.go` — Intentionally empty placeholder for post-MVP Elephantasm memory tools.

**Classifier**
- `classifier.go` — `Classifier` interface + `ClassifiedLog` + `PassthroughClassifier`.
- `classifier_lumber.go` — `LumberClassifier` (ONNX via `kaminocorp/lumber`).
- `severity_gate.go` — `ShouldEscalate` hardcoded policy table.

**Emission + plumbing**
- `emit.go` — `EmitLog`, `EmitLogWithSeverity`, fire-and-forget.
- `extract.go` — `ExtractText` / `ExtractTexts` helpers used by the classifier to pull text out of JSONB log payloads.
- `message.go` — `Message` domain type for conversation storage.

**Tests** (`*_test.go`) — paired with the file they cover; integration tests skip cleanly without `DATABASE_URL`.

---

## 11. Mental model cheat sheet

If you only remember one diagram, make it this one:

```
                         ┌───────────────┐
                         │   Postgres    │
                         │  (log_buffer, │
                         │  conversations│
                         │  agent_log,   │
                         │  schedules)   │
                         └───────┬───────┘
                                 │
                                 ▼
                   ┌─────────────────────────┐
                   │       agent.Agent       │
                   │  ┌───────────────────┐  │
                   │  │ providers map     │  │
                   │  │  "anthropic"      │  │
                   │  │  "openrouter"(?)  │  │
                   │  └─────────┬─────────┘  │
                   │            │            │
                   │            ▼            │
                   │  ┌───────────────────┐  │
                   │  │  tool registry:   │  │
                   │  │  search_logs      │  │
                   │  │  query_database   │  │
                   │  │  search_codebase  │  │
                   │  └───────────────────┘  │
                   └──┬────────┬────────┬────┘
                      │        │        │
        ┌─────────────┘        │        └──────────────┐
        ▼                      ▼                       ▼
 ┌────────────┐        ┌──────────────┐         ┌─────────────┐
 │ Interactive│        │ Monitoring   │         │  Scheduled  │
 │            │        │              │         │             │
 │ ws/chat ── │        │ 15s ticker   │         │ 1m ticker   │
 │ RunConv-   │        │ Lumber class.│         │ cron check  │
 │ Stream     │        │ ShouldEscal. │         │ shouldFire  │
 │            │        │ RunMonitoring│         │ reuses      │
 │ persists   │        │ (blocking)   │         │ RunMonitor- │
 │ history    │        │              │         │ ing         │
 │            │        │ rate limit:──┼─────────┼──► shared   │
 │            │        │              │         │             │
 └─────┬──────┘        └──────┬───────┘         └──────┬──────┘
       │                      │                        │
       └──────────────────────┴────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │     agent_log    │
                    │  (Activity feed) │
                    └──────────────────┘
```

---

## 12. Where to go next

- **To understand the *data* side** (where logs come from, how they reach `log_buffer`): `docs/blueprints/connections.md` and `docs/blueprints/database-connection-blueprint.md`.
- **To understand the Lumber model specifically**: `docs/blueprints/lumber-integration.md`.
- **To understand HTTP routing and middleware**: `docs/blueprints/backend-blueprint.md`.
- **To read the product-level "why"**: `docs/vision.md`. The three agent modes map 1:1 to the three sections of the "The Agent" part of the vision doc.
- **To see the most recent architectural changes**: `docs/changelog.md` — scheduled mode (0.31.0), OpenRouter (0.29.0), and Lumber hardening (0.27.0) are the three releases that shaped the current agent package the most.
