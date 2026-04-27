# Claude Agent SDK Abstractions Research for Heimdall

**Status:** research complete, implementation planning only.
**Date:** 2026-04-26.
**Scope:** identify abstractions and concepts from Anthropic's Claude Agent SDK that are worth borrowing into Heimdall's investigation runtime, without coupling Heimdall to Anthropic's SDK or to a coding-agent-specific execution model.

---

## 1. Executive summary

Heimdall should **not** adopt the Claude Agent SDK as its runtime.

It **should** study and borrow a number of its abstractions:

1. A first-class **agent loop model** with typed messages, typed results, explicit stop reasons, and a cost / budget envelope.
2. A formal **permission and policy layer** that is distinct from tool registration.
3. A formal **hook / interceptor lifecycle** around sessions, prompts, tool calls, and agent completion.
4. A richer **session model** with resume, fork, and external session storage.
5. A more deliberate **tool system** that separates tool availability, permissioning, and execution semantics.
6. A first-class **subagent abstraction** for isolated investigations.
7. A first-class **observability model** for LLM calls, tool calls, and hook decisions.
8. A more deliberate **context-loading model** for static instructions, reusable skills, and tool discovery.

The part to avoid copying is just as important:

1. Do **not** tie Heimdall's runtime to Anthropic's CLI process model.
2. Do **not** inherit Claude Code's filesystem-global configuration model into a multi-tenant backend.
3. Do **not** import coding-agent assumptions wholesale. Heimdall is a diagnostics agent, not a local-dev assistant.
4. Do **not** anchor the runtime on Anthropic-only capabilities when Heimdall already supports Anthropic direct and OpenRouter and should stay provider-neutral.

As of **April 26, 2026**, there is no official Anthropic **Go Agent SDK**. Anthropic has an official Go API SDK, but the Agent SDK itself is documented for TypeScript and Python.

---

## 2. What the Agent SDK actually is

One framing mistake to avoid: the Claude Agent SDK is **not** just "the Anthropic API plus tool calling."

It is closer to a **full local agent harness** with:

1. A built-in tool-execution loop.
2. Built-in tools for filesystem, shell, web, orchestration, and user interaction.
3. A permission engine.
4. Hook points around execution.
5. Session persistence and session branching.
6. Subagents.
7. Filesystem-loaded instructions, skills, commands, agents, and hooks.
8. Observability and cost reporting.
9. Deployment guidance for sandboxing and proxying.

One particularly important implementation detail: Anthropic documents that the Agent SDK runs the **Claude Code CLI as a child process** and communicates with it over a local pipe. That is useful as a clue to how much machinery the SDK bundles, but it is **not** a good design target for Heimdall.

For Heimdall, the right move is to copy the **shape** of the runtime, not the **packaging**.

---

## 3. Current Heimdall position

Today Heimdall already has the beginnings of the right architecture:

1. A provider-neutral interface in [backend/internal/agent/provider.go](../../backend/internal/agent/provider.go).
2. Native loop ownership in [backend/internal/agent/loop.go](../../backend/internal/agent/loop.go).
3. A local tool registry and dispatcher in [backend/internal/agent/tools.go](../../backend/internal/agent/tools.go).
4. Two major operating paths:
   - interactive chat via [backend/internal/api/handlers/chat.go](../../backend/internal/api/handlers/chat.go)
   - monitoring / scheduled investigation via [backend/internal/agent/monitor.go](../../backend/internal/agent/monitor.go) and [backend/internal/agent/scheduler.go](../../backend/internal/agent/scheduler.go)

That means Heimdall is **not starting from zero**. The opportunity is to evolve the existing agent package into a more explicit runtime instead of replacing it.

---

## 4. The abstractions worth borrowing

### 4.1 Agent loop as an explicit runtime primitive

### What the SDK does

The SDK treats the loop itself as a first-class thing:

1. prompt enters
2. model responds
3. tool calls are executed
4. tool results are fed back
5. loop repeats until final response
6. final result object closes the run

The docs also expose typed lifecycle messages such as system, assistant, user/tool-result, stream events, and final result.

### Why this matters for Heimdall

Heimdall already owns its loop, but the loop is still closer to "internal implementation" than "runtime contract."

Borrow:

1. A formal `RunResult` type with:
   - final text
   - stop reason
   - model usage
   - tool usage summary
   - budget / limit hit metadata
   - session ID
2. A formal `RunEvent` stream:
   - `session_init`
   - `assistant_turn`
   - `tool_requested`
   - `tool_started`
   - `tool_finished`
   - `tool_denied`
   - `hook_fired`
   - `compact_boundary`
   - `run_completed`
   - `run_failed`
3. Explicit stop reasons:
   - end_turn
   - max_turns
   - max_budget
   - permission_denied
   - provider_error
   - user_input_required
   - policy_blocked

### Heimdall recommendation

Turn the current loop into a reusable `runtime.Run()` primitive and make all entrypoints use it.

---

### 4.2 Separate tool availability from tool permission

### What the SDK does

The SDK has a crucial distinction:

1. `tools` controls what tools are even exposed to the model.
2. `allowedTools` / `disallowedTools` / permission mode control whether a requested tool execution is approved.

This is a high-signal abstraction. It is easy to miss and very worth copying.

### Why this matters for Heimdall

Today Heimdall's registry is mostly "if tool exists, model may call it."

That is too flat for production diagnostics.

Heimdall should separate:

1. **availability**
   - what tools are present in the prompt for this run
2. **discoverability**
   - what tools the model is told about up front versus discovered on demand
3. **permission**
   - what tools are auto-approved, denied, or require policy evaluation
4. **capability scope**
   - what connection IDs, schemas, repos, time ranges, and output sizes each tool may touch

### Heimdall recommendation

Introduce a `ToolCatalog` plus `PolicyEngine`, rather than treating `ToolRegistry()` as both.

---

### 4.3 Permission engine with an explicit evaluation order

### What the SDK does

The official permission order is:

1. hooks
2. deny rules
3. permission mode
4. allow rules
5. approval callback

That sequencing is excellent because it is predictable and explainable.

### Why this matters for Heimdall

Heimdall is currently read-mostly, but several tools are still sensitive:

1. `query_database`
2. future cloud / metrics / incident APIs
3. future write-capable operational tools, if they ever exist
4. any cross-tenant or org-level data access

Even for a diagnostics-only product, a policy layer is still required because "read-only" does not mean "low-risk."

### Heimdall recommendation

Add a policy engine with:

1. static deny rules
2. static allow rules
3. runtime policy callbacks
4. mode presets

Proposed modes:

1. `observe_only`
   - no external tools, model can summarize only the provided input
2. `safe_readonly`
   - logs, bounded metrics, bounded DB reads, code search
3. `guided_investigation`
   - same as `safe_readonly` plus approval workflow for expensive tools
4. `fully_autonomous_readonly`
   - auto-approved read-only tools within strict bounds
5. `plan_only`
   - no tool execution, diagnostics plan only

Do **not** copy `bypassPermissions` as a normal runtime mode. In Heimdall that should only exist as a test or internal-debug escape hatch.

---

### 4.4 Hooks and interceptors

### What the SDK does

The hook system is one of the richest parts of the SDK:

1. session lifecycle hooks
2. user prompt hooks
3. pre-tool and post-tool hooks
4. permission hooks
5. subagent hooks
6. stop hooks
7. compaction hooks

Some hooks can inject context, some can deny execution, and some are observability-only.

### Why this matters for Heimdall

Heimdall needs deterministic policy and context injection points far more than it needs ad hoc prompt tricks.

Strong Heimdall use cases:

1. `SessionStart`
   - inject incident metadata, app metadata, connection state, current deploy marker, last similar reports
2. `PreToolUse`
   - enforce row limits, time windows, SQL read-only guardrails, repo scoping
3. `PostToolUse`
   - capture summaries, redact sensitive outputs, emit trace data
4. `Stop`
   - validate the final report schema, severity, and required evidence before the run closes
5. `SubagentStart` / `SubagentStop`
   - tag delegated investigations and merge their conclusions safely
6. `PreCompact`
   - archive raw context or preserve critical incident identifiers before summarization

### Heimdall recommendation

Define a small but explicit hook surface in Go:

1. `OnSessionStart`
2. `OnRunStart`
3. `PreToolUse`
4. `PostToolUse`
5. `OnProviderRequest`
6. `OnProviderResponse`
7. `PreFinalize`
8. `OnRunEnd`

This should be implemented as deterministic code hooks, not shell hooks.

---

### 4.5 Sessions, resume, fork, and external session storage

### What the SDK does

The SDK treats session continuity as first-class:

1. continue the latest session
2. resume by ID
3. fork a session into a new branch
4. persist sessions to external stores
5. expose transcript readers and session metadata helpers

This is one of the most valuable abstractions for Heimdall.

### Why this matters for Heimdall

Heimdall investigations are naturally branchable:

1. "continue from yesterday's incident thread"
2. "fork this investigation into hypothesis A and hypothesis B"
3. "resume the run that hit the turn budget"
4. "replay a report with new evidence"

This is better than today's more linear chat-plus-monitoring model.

### Heimdall recommendation

Implement:

1. `SessionStore`
   - `Create`
   - `AppendEvent`
   - `Load`
   - `Resume`
   - `Fork`
   - `List`
   - `Tag`
2. session IDs as first-class API objects
3. a distinction between:
   - session history
   - report artifacts
   - file / code / tool evidence snapshots
4. org-app-user scoping at the storage layer, not just at handler entrypoints

For Heimdall, the external store should be DB-backed from day one, not local-disk-backed.

---

### 4.6 Subagents as deliberate context isolation

### What the SDK does

The SDK has explicit subagents:

1. their own prompt
2. their own tool set
3. fresh context
4. optional model override
5. parent receives only the summary, not the full working transcript in-band

This is a real architectural idea, not just a UI trick.

### Why this matters for Heimdall

Heimdall has obvious subagent candidates:

1. log-pattern explorer
2. SQL-diagnostics agent
3. codebase causality agent
4. incident report drafter
5. false-positive verifier

This gives Heimdall a way to preserve main-run context while pushing deep investigations into isolated branches.

### Heimdall recommendation

Borrow the subagent abstraction, but start with tight rules:

1. no recursive explosion
2. bounded fan-out
3. bounded model budget per subagent
4. explicit write set for each subagent's artifacts
5. explicit merge contract back into parent

Heimdall does **not** need general-purpose "spawn endless teammates." It needs disciplined delegated investigations.

---

### 4.7 Tool annotations and execution semantics

### What the SDK does

The SDK documents whether tools are read-only and uses that to decide whether parallel execution is safe.

It also normalizes tool result semantics:

1. return result blocks
2. mark `isError: true` for recoverable tool failures
3. let the model adapt rather than hard-crash the loop

### Why this matters for Heimdall

Heimdall already follows the `isError` pattern conceptually, which is correct.

What is missing is formal tool metadata:

1. read-only versus side-effecting
2. expensive versus cheap
3. cacheable versus non-cacheable
4. tenant-scoped versus app-scoped versus org-scoped
5. streaming versus buffered

### Heimdall recommendation

Add tool metadata:

| Field | Why it matters |
|---|---|
| `ReadOnly` | permits concurrency and safer autonomous execution |
| `CostClass` | budget enforcement and scheduling |
| `LatencyClass` | timeout selection |
| `MaxPayloadBytes` | context protection |
| `Scope` | auth and tenancy enforcement |
| `CacheTTL` | repeated investigation efficiency |
| `RequiresApproval` | policy shortcut |

This is a low-regret improvement.

---

### 4.8 Tool search and on-demand loading

### What the SDK does

Anthropic explicitly calls out that:

1. tool definitions consume significant context
2. tool selection quality degrades when too many tools are loaded
3. tool search can load only a relevant subset on demand

This is especially important once tool count grows beyond a few dozen.

### Why this matters for Heimdall

Heimdall is still early, but this becomes relevant the moment it has:

1. multiple connector families
2. multiple vendor-specific tools
3. numerous investigation helpers
4. many org-level integrations

Today Heimdall is small enough to ignore this.
Future Heimdall probably is not.

### Heimdall recommendation

Do **not** build tool search immediately.

Do build the tool catalog in a way that allows it later:

1. short summaries per tool
2. full schema loaded lazily
3. tags / categories
4. capability indexing
5. source connector metadata

This is a "design now, implement later" abstraction.

---

### 4.9 Context-loading model: static instructions, skills, commands, plugins

### What the SDK does

The SDK has a layered context-loading model:

1. static instructions via `CLAUDE.md`
2. reusable skills loaded on demand
3. slash commands
4. filesystem agents
5. plugins packaging these concepts
6. `settingSources` to control where context comes from

### Why this matters for Heimdall

The exact Claude Code filesystem model is the wrong fit for Heimdall, but the abstraction is useful.

Heimdall needs a server-native equivalent:

1. **org instructions**
   - operational conventions
   - naming conventions
   - service ownership notes
2. **app instructions**
   - architecture hints
   - known failure modes
   - important schemas
3. **playbooks / skills**
   - reusable investigation workflows
4. **connector capabilities**
   - documented per-connector affordances and limits

### Heimdall recommendation

Borrow the layering idea, not the local-filesystem implementation.

Proposed hierarchy:

1. org memory
2. app memory
3. run-time injected context
4. reusable investigation skills
5. ephemeral prompt

These should live in DB-backed configuration and versioned records, not `.claude/` directories.

---

### 4.10 Context compaction and compact-boundary events

### What the SDK does

The SDK treats compaction as an explicit runtime behavior:

1. old context is summarized automatically
2. a compact-boundary event is emitted
3. hooks can run before compaction
4. docs explicitly recommend putting persistent rules in static memory rather than in transient prompts

### Why this matters for Heimdall

Long-running investigations will eventually need compaction.

Without an explicit model, compaction becomes invisible and dangerous:

1. evidence disappears without auditability
2. conclusions drift from original facts
3. report quality degrades silently

### Heimdall recommendation

Borrow:

1. compact-boundary events
2. pre-compaction archiving
3. persistent memory separate from transient turns
4. explicit "preserve these facts" compaction instructions

This is more important once sessions and subagents become real.

---

### 4.11 Structured outputs

### What the SDK does

The SDK supports schema-validated structured outputs and retries / re-prompts when the result does not validate.

### Why this matters for Heimdall

This is directly applicable to:

1. incident reports
2. assessment summaries
3. severity objects
4. recommended next actions
5. evidence bundles
6. scheduled investigation outputs

### Heimdall recommendation

Add a structured-result layer for monitoring and report-generation paths before adding it to free-form chat.

The ideal target is:

1. provider-neutral schema
2. validation after each final answer
3. repair pass when validation fails
4. typed persistence

This is probably one of the highest-value near-term borrowings.

---

### 4.12 User-input and approval callbacks

### What the SDK does

The SDK has a formal callback path for:

1. tool approval requests
2. clarifying questions via `AskUserQuestion`

### Why this matters for Heimdall

Heimdall is primarily autonomous, but there are still high-value interactive cases:

1. "which connection should I use?"
2. "should I widen the time range?"
3. "should I run the expensive query?"
4. "which hypothesis do you want me to pursue?"

### Heimdall recommendation

Borrow this concept for interactive chat only.

It is lower priority than sessions, hooks, policy, and structured outputs, but still worth designing into the runtime.

---

### 4.13 File checkpointing as a general "reversible state" idea

### What the SDK does

The SDK can checkpoint file edits and rewind them later, with explicit limitations.

### Why this matters for Heimdall

Heimdall is read-only by product principle, so file checkpointing itself is not directly relevant.

What is relevant is the underlying discipline:

1. if the runtime changes state, it should be reversible
2. if the runtime branches, the branch boundary should be explicit
3. if a result is based on mutable evidence, the evidence snapshot should be capturable

### Heimdall recommendation

Do **not** build checkpointing now.

Do adopt the principle for:

1. evidence snapshotting
2. report regeneration
3. session forks

---

### 4.14 Observability and telemetry

### What the SDK does

The SDK has a strong observability story:

1. OpenTelemetry traces
2. tool-level spans
3. model-request spans
4. hook spans
5. cost and token reporting
6. session-aware trace correlation

### Why this matters for Heimdall

This maps almost perfectly to Heimdall's needs.

The investigation runtime should expose:

1. provider request counts and latency
2. token usage and approximate cost
3. per-tool latency and failure rates
4. per-run budget consumption
5. subagent fan-out
6. compaction counts
7. policy denial counts
8. final stop reasons

### Heimdall recommendation

Build this in from the start of the runtime refactor, not afterward.

If the runtime becomes more sophisticated before traces exist, debugging and production hardening will be harder than necessary.

---

### 4.15 Hosting, isolation, and security boundaries

### What the SDK does

Anthropic's docs emphasize:

1. persistent execution environments
2. container / sandbox isolation
3. least privilege
4. proxy-based credential injection
5. read-only mounts where possible

### Why this matters for Heimdall

Heimdall is a server product that touches:

1. production logs
2. production databases
3. code repositories
4. future external APIs

The runtime security model cannot be an afterthought.

### Heimdall recommendation

Borrow the boundary thinking:

1. tools should hold credentials, not the model
2. network access should be per-tool, not globally open
3. tool runners should be individually constrainable
4. read-only guarantees should be enforced at the connector level and re-checked at runtime

This is more important than any specific SDK feature.

---

## 5. Concepts to avoid copying directly

### 5.1 The CLI child-process architecture

This is an Anthropic packaging choice, not a universal runtime ideal.

For Heimdall, a direct in-process Go runtime remains the correct architecture.

### 5.2 Local filesystem settings as the main source of truth

The SDK's `settingSources` model is useful conceptually, but the actual `.claude/` and home-directory loading behavior is risky in a multi-tenant backend.

Heimdall should use DB-backed, tenant-scoped configuration.

### 5.3 Coding-agent-first built-in tool assumptions

The SDK is optimized for file editing, shell commands, and coding workflows.

Heimdall should not distort itself to match those assumptions. Its core tool families are:

1. logs
2. metrics
3. traces
4. databases
5. codebase search
6. historical memory
7. reports

### 5.4 Vendor-specific capability assumptions

Heimdall should keep the loop and tool framework provider-neutral and keep provider-specific translations at the boundary.

### 5.5 Host-level ambient memory bleed

Anthropic explicitly warns that some settings and memory can still load outside explicit `settingSources`.

That is exactly the kind of ambient coupling Heimdall should avoid.

---

## 6. A borrowing priority matrix for Heimdall

| Priority | Abstraction | Why |
|---|---|---|
| P0 | Typed run events and results | foundational runtime contract |
| P0 | Policy engine and permission order | safety, predictability, future-proofing |
| P0 | Hook lifecycle | context injection, observability, policy |
| P0 | Structured outputs for reports | immediate product value |
| P0 | Runtime telemetry | required for production hardening |
| P1 | Session resume and fork | major UX and investigation value |
| P1 | Tool metadata and execution semantics | concurrency, budgets, safe autonomy |
| P1 | Subagents | strong value once session model exists |
| P1 | Org/app memory layering | improves investigation quality |
| P2 | External session store adapters | needed when runtime topology grows |
| P2 | Context compaction controls | needed for longer investigations |
| P2 | Approval / clarifying-question callbacks | useful for interactive workflows |
| P3 | Tool search | valuable later when tool count grows |
| P3 | Checkpointing analogues | lower priority for read-only product |
| P3 | Plugin packaging | useful only after runtime primitives settle |

---

## 7. Proposed Go-native architecture for Heimdall

This is the most important practical output from the research.

### 7.1 Runtime package split

Proposed packages:

1. `backend/internal/agent/runtime`
2. `backend/internal/agent/policy`
3. `backend/internal/agent/session`
4. `backend/internal/agent/hooks`
5. `backend/internal/agent/tools`
6. `backend/internal/agent/subagents`
7. `backend/internal/agent/trace`

### 7.2 Interface sketch

```go
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
}

type Tool interface {
    Name() string
    Schema() json.RawMessage
    Metadata() ToolMetadata
    Execute(ctx context.Context, call ToolCall, scope ToolScope) (*ToolResult, error)
}

type PolicyEngine interface {
    Evaluate(ctx context.Context, req PolicyRequest) PolicyDecision
}

type HookSet interface {
    OnSessionStart(ctx context.Context, s *Session) error
    PreToolUse(ctx context.Context, call *ToolCall) HookDecision
    PostToolUse(ctx context.Context, call *ToolCall, result *ToolResult) error
    PreFinalize(ctx context.Context, result *RunResult) error
    OnRunEnd(ctx context.Context, result *RunResult) error
}

type SessionStore interface {
    Create(ctx context.Context, meta SessionMeta) (*Session, error)
    AppendEvent(ctx context.Context, sessionID uuid.UUID, ev RunEvent) error
    Load(ctx context.Context, sessionID uuid.UUID) (*Session, error)
    Fork(ctx context.Context, sessionID uuid.UUID, forkMeta ForkMeta) (*Session, error)
}
```

The point is not the exact signatures. The point is to make the runtime shape explicit.

### 7.3 Design principle

The loop should become:

1. provider-neutral
2. tool-neutral
3. policy-aware
4. hook-aware
5. session-aware
6. trace-aware
7. subagent-capable

That is the correct long-term shape.

---

## 8. Concrete Heimdall use-cases unlocked by these abstractions

1. Resume a failed investigation after a provider outage without losing context.
2. Fork a report into "deployment regression" and "downstream outage" hypotheses.
3. Send code search to a codebase subagent and SQL analysis to a DB subagent in parallel.
4. Enforce org-specific investigation policies before any expensive DB query runs.
5. Validate that every incident report contains severity, evidence, and confidence before persistence.
6. Trace the exact tool chain behind every report in OpenTelemetry.
7. Load app-specific runbooks and known failure modes as durable memory, not ad hoc prompt text.
8. Introduce new connectors without exploding prompt size or provider coupling.

---

## 9. Recommended implementation sequence

### Phase 1

Refactor the current loop into an explicit runtime contract:

1. typed run result
2. typed run events
3. stop reasons
4. model usage envelope

### Phase 2

Introduce policy and hook layers:

1. pre-tool policy evaluation
2. pre-finalize validation
3. run/session hooks

### Phase 3

Introduce structured outputs for monitoring and reports.

### Phase 4

Introduce session resume and fork on DB-backed storage.

### Phase 5

Introduce subagents for bounded delegated investigations.

### Phase 6

Introduce richer memory layering and optional tool discovery.

This sequencing keeps the complexity slope reasonable.

---

## 10. Final recommendation

The Claude Agent SDK is worth studying because it has already productized several abstractions that matter for real agents:

1. loop lifecycle
2. permissions
3. hooks
4. sessions
5. subagents
6. telemetry
7. context management

But Heimdall should borrow those **as architecture**, not **as dependency**.

The target should be:

1. **Go-native**
2. **provider-neutral**
3. **diagnostics-first**
4. **multi-tenant-safe**
5. **observable**
6. **strictly read-oriented by default**

That gives Heimdall the upside of the SDK's design without inheriting its Claude-specific runtime assumptions.

---

## Sources

Research completed on **2026-04-26** using primary Anthropic documentation and official Anthropic repositories.

1. Agent SDK overview: <https://code.claude.com/docs/en/agent-sdk/overview>
2. How the agent loop works: <https://code.claude.com/docs/en/agent-sdk/agent-loop>
3. Configure permissions: <https://code.claude.com/docs/en/agent-sdk/permissions>
4. Hooks reference: <https://code.claude.com/docs/en/hooks>
5. Work with sessions: <https://code.claude.com/docs/en/agent-sdk/sessions>
6. Persist sessions to external storage: <https://code.claude.com/docs/en/agent-sdk/session-storage>
7. Subagents in the SDK: <https://code.claude.com/docs/en/agent-sdk/subagents>
8. Give Claude custom tools: <https://code.claude.com/docs/en/agent-sdk/custom-tools>
9. Connect to external tools with MCP: <https://platform.claude.com/docs/en/agent-sdk/mcp>
10. Scale to many tools with tool search: <https://code.claude.com/docs/en/agent-sdk/tool-search>
11. Use Claude Code features in the SDK: <https://code.claude.com/docs/en/agent-sdk/claude-code-features>
12. Agent Skills in the SDK: <https://code.claude.com/docs/en/agent-sdk/skills>
13. Slash Commands in the SDK: <https://code.claude.com/docs/en/agent-sdk/slash-commands>
14. Plugins in the SDK: <https://code.claude.com/docs/en/agent-sdk/plugins>
15. Get structured output from agents: <https://code.claude.com/docs/en/agent-sdk/structured-outputs>
16. Handle approvals and user input: <https://code.claude.com/docs/en/agent-sdk/user-input>
17. Rewind file changes with checkpointing: <https://code.claude.com/docs/en/agent-sdk/file-checkpointing>
18. Observability with OpenTelemetry: <https://code.claude.com/docs/en/agent-sdk/observability>
19. Hosting the Agent SDK: <https://code.claude.com/docs/en/agent-sdk/hosting>
20. Securely deploying AI agents: <https://code.claude.com/docs/en/agent-sdk/secure-deployment>
21. Official TypeScript SDK repo: <https://github.com/anthropics/claude-agent-sdk-typescript>
22. Official Python SDK repo: <https://github.com/anthropics/claude-agent-sdk-python>
23. Official Go API SDK repo: <https://github.com/anthropics/anthropic-sdk-go>
