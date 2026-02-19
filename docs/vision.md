# Heimdall — Vision

## What is Heimdall?

Heimdall is an autonomous AI monitoring agent for production applications. It replaces passive, noisy alerting — log dumps piped into Slack or Discord — with an intelligent agent that watches your systems 24/7, understands what it's seeing, and tells you what matters.

You deploy your app, connect Heimdall to your infrastructure, and hand it the responsibility of keeping watch. It monitors continuously, investigates anomalies on its own, and surfaces only what's worth your attention.

## The Problem

Modern monitoring is broken in a specific way: it generates too much noise and not enough insight.

- **Log alerts are dumb.** They fire on pattern matches with no understanding of context. A spike in error rates might be a deploy rolling out, a downstream dependency hiccup, or a real incident — but a traditional alert treats them all the same.
- **Investigation is manual.** When something looks wrong, a developer has to context-switch, open multiple dashboards, cross-reference logs with database state, check recent code changes, and piece together a picture. This takes time and expertise.
- **Knowledge is ephemeral.** The developer who debugged last week's incident carries that context in their head. When a similar pattern appears at 3am, the on-call engineer starts from scratch.

## The Vision

Heimdall is a **monitoring and diagnostics** tool — not a problem-solver. Its job is to watch, understand, investigate, and report. It's the all-seeing eye over your production systems.

Specifically, Heimdall should:

1. **Monitor continuously.** Ingest server logs and database activity around the clock. Know what normal looks like, and recognise when something deviates.
2. **Investigate autonomously.** When anomalies appear, the agent doesn't just alert — it digs in. It queries the database, inspects recent log patterns, and checks relevant code to build a diagnosis before surfacing it to the team.
3. **Remember and learn.** Through long-term memory, the agent accumulates institutional knowledge — past incidents, system patterns, known failure modes. Each incident makes it better at the next one.
4. **Be available on-demand.** Developers can chat with the agent at any time to ask questions: "What happened with the payment service last night?", "Show me error trends for the last 6 hours", "Has this query pattern caused issues before?".
5. **Produce actionable reports.** When incidents occur, Heimdall generates structured reports covering what happened, what it found during investigation, relevant historical context, and its diagnostic assessment.

## Core Concepts

### Connections

Connections are integrations between Heimdall and your infrastructure. Each connection links the agent to a data source it can observe or query.

| Connection type | Direction | Example |
|----------------|-----------|---------|
| Server logs | One-way (streaming in) | Application logs feed into Heimdall continuously |
| Database activity | One-way (streaming in) + on-demand queries out | DB events stream in; agent can also run SQL queries when investigating |
| Codebase | On-demand queries out | Agent queries a GitHub repo to understand relevant code during investigation |

Connections can be **one-way** (data flows into Heimdall) or **two-way** (data flows in, and the agent can also query the source on-demand).

### The Agent

The agent is the intelligence at the centre of Heimdall. It is powered by an LLM with access to a defined set of tools (query database, search logs, inspect codebase, recall past incidents). It operates in two modes:

- **Monitoring mode** — the agent runs continuously (or on a configurable schedule), processing incoming data streams and deciding when something warrants investigation or escalation.
- **Interactive mode** — a developer opens a chat with the agent and asks questions. The agent uses its tools and memory to answer in real time.

### Agent Log

The agent log is a curated, chronological master feed. It combines:

- Raw activity from connected sources (log entries, DB events)
- Agent observations and annotations (what the agent noticed, flagged, or investigated)
- Rule-based entries (configurable triggers independent of the agent)

This gives teams a single pane of glass over all system activity, enriched by the agent's analysis.

### Reports

Reports are generated for incidents. When the agent detects and investigates an issue, it produces a structured report covering:

- Timeline of events
- What was investigated and what was found
- Historical context from similar past incidents
- Diagnostic assessment and severity

## Design Principles

### Fast onboarding
Setting up Heimdall should take under 5 minutes. Connect your data sources, configure the agent, and you're live. No complex setup, no steep learning curve.

### Monitoring first
Heimdall watches and diagnoses — it doesn't take action on your systems. It won't restart services, roll back deploys, or modify data. Its value is in understanding and surfacing insight, not in autonomous remediation.

### Simple by default
The platform should feel straightforward. A small number of well-defined sections, clear data flows, and minimal configuration. Complexity is hidden behind sensible defaults, not exposed as options.

## Platform Sections

1. **Connections** — Add, configure, and manage integrations (databases, log sources, codebases). View connection health and data flow status.
2. **Agent Configuration** — Choose the backing model, set the monitoring schedule (always-on vs periodic), and tune agent behaviour.
3. **Agent Chat** — Real-time conversational interface with the agent for on-demand queries and investigation.
4. **Agent Log** — The master chronological feed of all system activity and agent observations.
5. **Reports** — Incident reports generated by the agent, searchable and browsable.
