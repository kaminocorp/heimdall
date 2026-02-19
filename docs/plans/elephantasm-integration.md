# Elephantasm Integration — Heimdall

## Overview

[Elephantasm](https://www.elephantasm.com/) is a long-term agentic memory (LTAM) framework that gives AI agents persistent, evolving memory. It transforms raw events into layered memory through a structured pipeline:

```
Events → Memories → Lessons → Knowledge → Identity
```

For Heimdall, this means the monitoring agent can remember past incidents, learn system behaviour patterns over time, and build institutional knowledge about the applications it watches.

Reference: [Heimdall Blueprint](./blueprint.md)

---

## Why Elephantasm for Heimdall

A monitoring agent without memory treats every incident as if it's the first time. With Elephantasm:

- **Incident memory** — the agent remembers past incidents, their root causes, and resolutions. When a similar pattern appears, it can reference historical context.
- **System pattern learning** — over time the agent learns what "normal" looks like for each monitored application (typical log patterns, DB load, deploy schedules).
- **Evolving diagnostics** — lessons distilled from past events improve the agent's diagnostic reasoning without retraining or prompt engineering.
- **Audit trail** — Elephantasm's lineage tracing lets users understand _why_ the agent believes something — which events led to which memories led to which conclusions.

---

## Integration Approach

### Constraint

Elephantasm provides Python and TypeScript SDKs but no Go SDK. The Heimdall backend is Go.

### Recommended: Direct API Calls from Go

Elephantasm exposes an API. We write a thin Go client package that wraps these API calls.

```
backend/
└── internal/
    └── memory/
        ├── client.go       # HTTP client for Elephantasm API
        ├── types.go        # Go structs mirroring Elephantasm models
        └── memory.go       # High-level memory operations for the agent
```

**Advantages:**
- No extra services to deploy or manage
- Stays fully within the Go backend
- Simple to implement — it's just HTTP calls + JSON marshalling
- No dependency on Python/Node runtime

**Implementation sketch:**

```go
// memory/client.go

type Client struct {
    baseURL    string
    httpClient *http.Client
    apiKey     string
}

func (c *Client) RecordEvent(ctx context.Context, event Event) error { ... }
func (c *Client) QueryMemories(ctx context.Context, query string) ([]Memory, error) { ... }
func (c *Client) GetLessons(ctx context.Context, topic string) ([]Lesson, error) { ... }
```

```go
// Usage in agent loop

func (a *Agent) handleIncident(ctx context.Context, incident Incident) {
    // Check if we've seen something similar before
    memories, _ := a.memory.QueryMemories(ctx, incident.Summary)

    // Include relevant memories in the agent's context
    context := buildContext(incident, memories)

    // Run agent loop with historical context
    response := a.runLoop(ctx, context)

    // Record this incident as a new event for future reference
    a.memory.RecordEvent(ctx, Event{
        Type:    "incident",
        Content: incident.Summary,
        Result:  response.Diagnosis,
    })
}
```

### Fallback: Sidecar Microservice

If the Elephantasm API turns out to be incomplete or the SDKs do significant client-side processing that can't be replicated with raw API calls, fall back to a sidecar:

```
┌──────────────┐     HTTP/gRPC     ┌─────────────────────┐
│  Go Backend  │ ◄──────────────►  │  memory-sidecar     │
│              │                    │  (Python or Node)   │
└──────────────┘                    │  wraps Elephantasm  │
                                    │  SDK                │
                                    └─────────────────────┘
```

The sidecar would live in the monorepo:

```
heimdall/
└── services/
    └── memory-sidecar/
        ├── main.py (or index.ts)
        ├── requirements.txt (or package.json)
        └── Dockerfile
```

Added to `docker-compose.yml` as a service alongside the Go backend.

**When to use the sidecar approach:**
- The Elephantasm API doesn't cover all SDK functionality
- The SDK does complex client-side memory processing (embedding, chunking, etc.)
- You want to use SDK features the moment they ship without writing Go wrappers

**Downsides:**
- Extra service to deploy and monitor
- Added latency (hop through sidecar)
- Mixed runtime (Python/Node alongside Go)

### Not Recommended: Building a Go SDK

Writing a full Go SDK for Elephantasm is significant effort and maintenance burden. Only consider this if Heimdall becomes deeply coupled to Elephantasm internals, or as an open-source contribution once the integration is proven.

---

## Integration Points in Heimdall

### Where memory gets written (events → Elephantasm)

| Event | When | What gets stored |
|-------|------|------------------|
| Incident detected | Agent identifies an anomaly | Incident summary, affected services, severity |
| Incident resolved | Agent or user confirms resolution | Root cause, resolution steps, time to resolve |
| Agent observation | Agent notes a pattern during monitoring | Pattern description, confidence, supporting evidence |
| User interaction | User asks agent a question | Question topic, answer given, user feedback |

### Where memory gets read (Elephantasm → agent context)

| Situation | What gets queried | How it's used |
|-----------|-------------------|---------------|
| New anomaly detected | Similar past incidents | Agent references historical context in its diagnosis |
| Periodic monitoring | Learned "normal" patterns | Agent compares current state against known baselines |
| User asks a question | Relevant memories + lessons | Agent answers with accumulated knowledge |
| Report generation | All memories for an incident | Report includes historical context and pattern analysis |

### Agent tool integration

Elephantasm access is exposed to the agent as tools in the tool-use loop:

```go
var memoryTools = []Tool{
    {
        Name:        "recall_similar_incidents",
        Description: "Search agent memory for similar past incidents or patterns",
        Parameters: map[string]Param{
            "query": {Type: "string", Description: "Description of what to search for"},
        },
    },
    {
        Name:        "recall_lessons",
        Description: "Retrieve lessons the agent has learned about a topic",
        Parameters: map[string]Param{
            "topic": {Type: "string", Description: "The topic to retrieve lessons for"},
        },
    },
}
```

This lets the agent decide when to consult its memory rather than hardcoding memory lookups into every flow.

---

## Rollout Plan

### Phase 1 — Evaluate (Pre-MVP)
- Review Elephantasm API docs in detail
- Confirm API coverage vs SDK-only features
- Decide: direct API calls vs sidecar
- Spike: write a minimal Go client and test core operations

### Phase 2 — Basic Integration (Post-MVP)
- Implement Go client package (`internal/memory/`)
- Record incidents and resolutions as events
- Query memories during incident investigation
- Add memory tools to the agent's tool set

### Phase 3 — Deep Integration
- Feed all agent observations into Elephantasm continuously
- Use distilled lessons to improve system prompts dynamically
- Surface memory lineage in the Reports UI (why the agent believed X)
- Expose memory management in the Agent Config UI (view/edit/forget)
