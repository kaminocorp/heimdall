Idea:

1. Sign-up for free very quickly <10 sec
2. Drag and drop a log
3. Get your identification of issue / whats happening
4. Optional: Link up GH repo to get agent to investigate

The idea is to have a public-facing mini mechanism that anyone can use to immediately see what Heimdall can do / how it works, without having to signup, enter credit card etc. etc.

We'd obviously have to limit this somehow in terms of token blasting i.e. apply some sort of commonsense rate limiting.

Also, the UI would have to be ultra-intuitive, showing even in an empty state what we can do i.e. I/O

A bit like those eg png-to-avif file conversion sites that allow you to drag in smth (in our case log file(s)) and see the output.

Before we go further, dyou get what I'm describing here?

If all clear, outline one or several proposals how we could achieve/design this, in this md file.

---

## Proposals

### Proposal A — "Log Playground" (Zero-Auth, Stateless)

**Concept:** A single public page at `/playground` (or `/try`). No signup whatsoever. The user lands on the page, drops a log file or pastes log text, and gets back a structured analysis within seconds. Nothing is stored — fully stateless.

**How it works:**

```
User drops file / pastes text
        ↓
  Frontend parses into lines, auto-detects format
  (reuse existing webhook parser logic client-side,
   or send raw to backend)
        ↓
  POST /api/public/analyze  (no auth)
        ↓
  Backend: rate-limit by IP → format logs →
  call RunMonitoring() with PassthroughClassifier
  (skip Lumber, send everything to Claude)
        ↓
  Returns: { assessment, severity, summary }
        ↓
  Frontend renders structured output card
```

**UI flow:**

1. **Empty state** — Large drop zone with example ghost text showing "what goes in" (sample log snippet) and "what comes out" (sample assessment card). A "try with sample logs" button pre-fills with a realistic error scenario so users can click once and see the output immediately.
2. **Processing state** — Log lines appear stacked on the left, a pulsing Heimdall "eye" animation in the centre, assessment builds on the right (streamed if possible).
3. **Result state** — Structured card: severity badge, plain-English assessment, key findings as bullet points, suggested actions. A CTA below: "Want Heimdall watching your systems 24/7? Sign up in 30 seconds →".

**Rate limiting:**
- IP-based: 5 analyses per hour, 20 per day per IP.
- Payload cap: 100KB or ~500 log lines per submission.
- Model: Use Haiku or Sonnet (not Opus) to keep cost per analysis under ~$0.01.

**Backend delta:**
- One new handler: `AnalyzePublicLogs` — no auth middleware, accepts raw text or JSON array.
- Reuses `formatFlaggedLogs()` to structure input for Claude.
- Calls `RunMonitoring()` with a hardcoded `AppAgentConfig` (Sonnet, no tools — pure assessment, no search_logs/query_database since there's no user data to search).
- No database writes — nothing persisted.

**Pros:**
- Zero friction — literally zero signup. Maximum top-of-funnel.
- Simple to build — one endpoint, one page, no new data model.
- No abuse surface beyond API cost (no data stored, no accounts created).

**Cons:**
- No tool use (no DB to search, no repo to inspect) — limits what the agent can show off.
- Stateless means no "come back and see your history" hook.
- Rate limiting by IP is imperfect (VPNs, shared networks).

---

### Proposal B — "Instant Account" (Ephemeral Session, Light Auth)

**Concept:** One-click anonymous account creation. The user clicks "Try Heimdall", gets an ephemeral session (stored in a cookie/localStorage, no email required), and enters a stripped-down version of the real product. They can drop logs, see the analysis, and even chat with the agent about what it found. The session expires after 24 hours.

**How it works:**

```
User clicks "Try it free"
        ↓
  POST /api/public/session  →  returns ephemeral JWT
  (creates a shadow user + org + app in DB,
   marked as `ephemeral = true`)
        ↓
  User lands in a simplified dashboard:
  drop zone + results panel + mini chat
        ↓
  Dropped logs are ingested via the normal webhook
  pipeline (scoped to the ephemeral app)
        ↓
  Agent runs classification + assessment
  (same as real monitoring flow)
        ↓
  User can open chat and ask follow-up questions
  ("what caused the spike at 14:32?")
        ↓
  CTA: "Keep your workspace — sign up with email →"
  (converts ephemeral user to permanent)
```

**What the user gets:**
- Real log ingestion (logs stored temporarily).
- Real classification via Lumber + Claude assessment.
- Real agent chat with `search_logs` tool working against their uploaded logs.
- A taste of the actual product, not a simulation.

**Rate limiting:**
- Session-level: 3 analyses, 10 chat messages per session.
- Ephemeral data TTL: 24 hours, then a cron job purges the user/org/app/logs.
- Captcha or Turnstile on session creation to block bots.

**Backend delta:**
- New handler: `CreateEphemeralSession` — creates user/org/app with `ephemeral` flag, returns a short-lived JWT.
- Ephemeral users route through the normal pipeline but with stricter limits (enforced in middleware or agent config).
- Cleanup job: delete all ephemeral users + cascaded data older than 24h.
- Conversion endpoint: `POST /api/public/convert` — attaches email/password to an ephemeral user, clears the ephemeral flag.

**Pros:**
- Showcases the real product — chat, tool use, classification. Much more impressive demo.
- "Convert to real account" is a powerful retention hook (sunk-cost effect — they already have data in the system).
- Reuses almost the entire existing stack with no special-case code paths.

**Cons:**
- More complex: ephemeral user lifecycle, cleanup jobs, conversion flow.
- Stores data (even temporarily) — introduces abuse surface for data exfiltration or storage spam.
- Requires some form of bot protection on session creation.

---

### Proposal C — Hybrid (Recommended)

**Concept:** Combine A and B into a two-stage funnel. Stage 1 is the zero-auth playground (Proposal A) — instant gratification, no friction. If the user wants more (chat, history, deeper investigation), they click through to Stage 2, which creates an ephemeral session (Proposal B).

**Flow:**

```
Landing page → "Drop your logs" (Proposal A, zero auth)
        ↓
  User sees assessment result
        ↓
  CTA: "Want to investigate further? Ask the agent →"
        ↓
  One-click ephemeral session created (Proposal B)
        ↓
  User enters mini-dashboard with their logs
  already loaded, can chat with agent
        ↓
  CTA: "Keep your workspace — sign up →"
```

**Why this is the strongest option:**
- Stage 1 captures the broadest top-of-funnel (zero friction, just curiosity).
- Stage 2 captures engaged users who want to go deeper (self-selected high-intent).
- Each stage can be built independently — ship Stage 1 first, add Stage 2 later.
- The conversion funnel is natural: try → explore → commit.

**Build order:**
1. Ship Proposal A (playground page + one endpoint). ~1-2 days of work.
2. Ship Proposal B additions (ephemeral sessions, chat, conversion). Adds ~2-3 days.
3. Connect the two stages with the handoff CTA.

---

### UI Sketch (Applies to all proposals)

```
┌──────────────────────────────────────────────────────────┐
│  HEIMDALL                              [Sign In]         │
├──────────────────────────────────────────────────────────┤
│                                                          │
│     See what your logs are telling you.                  │
│                                                          │
│  ┌────────────────────────┐  ┌────────────────────────┐  │
│  │                        │  │                        │  │
│  │   Drop your log file   │  │   ▸ Assessment         │  │
│  │   here, or paste text  │  │     (appears here)     │  │
│  │                        │  │                        │  │
│  │   ┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈   │  │   ▸ Severity: ██      │  │
│  │   Supports: plaintext, │  │                        │  │
│  │   JSON, NDJSON, Fly.io │  │   ▸ Key findings       │  │
│  │   Vector, Vercel,      │  │     • ...              │  │
│  │   AWS Firehose, GCP    │  │     • ...              │  │
│  │   Pub/Sub              │  │                        │  │
│  │                        │  │   ▸ Suggested actions   │  │
│  │  [Try with sample logs]│  │     • ...              │  │
│  │                        │  │                        │  │
│  └────────────────────────┘  └────────────────────────┘  │
│                                                          │
│  "Want 24/7 monitoring? Sign up in 30 seconds →"        │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

The empty state is the key UX moment — the right panel should show a *faded example* of a real assessment so users understand the output format before they even drop a file. The "Try with sample logs" button lets someone see a live result with one click.

---

## Questions

1. **Scope of analysis** — Should the playground analysis be pure text assessment (Claude reads the logs and writes a report), or should we also run Lumber classification and show the structured categories (ERROR/PERFORMANCE/REQUEST etc.) as part of the output? Showing Lumber output would make it more "product-like" but adds ONNX model dependency to the public path.

2. **Streaming** — Should the assessment stream in token-by-token (like ChatGPT), or arrive as a complete card? Streaming feels more alive but adds WebSocket complexity to the public page.

3. **File formats** — Beyond structured JSON logs, should we accept plain-text log files (e.g., a raw nginx access log, syslog dump)? This would need a simple line-splitter that wraps each line as a `{ "payload": { "message": "..." } }` entry. Broadens the input surface significantly.

4. **GitHub integration** — You mentioned "optional: link up GH repo". Should this be part of the initial playground (Proposal A), or reserved for the ephemeral session (Proposal B)? GitHub OAuth in a zero-auth context gets architecturally awkward.

5. **Branding** — Should this live on the main Heimdall domain (`heimdall.app/playground`) or a separate micro-site (`try.heimdall.app`)? Separate domain gives more design freedom but splits SEO.

6. **Sample logs** — Should we curate 2–3 sample log sets that showcase different scenarios (e.g., "Payment service errors", "Memory leak pattern", "Deploy rollout")? This lets users see the range without having their own logs handy.