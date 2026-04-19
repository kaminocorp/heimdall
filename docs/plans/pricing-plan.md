# Heimdall Pricing Blueprint

**Pricing philosophy: a cheap floor, a value-aligned meter, and a BYOK ceiling.**

Heimdall charges a low base price per Monitored Application and a per-investigation overage. Idle monitoring is essentially free — the agent only costs you money when it actually does work on your behalf. At the top of the ladder, Enterprise customers bring their own LLM key and their own database, and pay a flat fee for Heimdall as a managed harness.

The model is built around three principles:

1. **Price the unit of value, not the substrate.** Customers experience value in *investigations completed* — not in GBs ingested or hosts monitored. So that's the meter.
2. **Idle is free, work is paid.** A quiet app pays the floor. A noisy app pays for the work the agent does. This makes the bill self-justifying: the customers paying the most are by definition the ones getting the most value.
3. **Massively cheaper than incumbents at every tier.** Heimdall lands at 1–20% of comparable Datadog, Dynatrace, or Resolve.ai bills across every realistic customer profile. The "10x better, 10x cheaper" claim must hold from the indie hobbyist to the regulated enterprise.

---

## Pricing at a Glance

| Feature                          |        Hobby |                Pro |               Business |             Enterprise |
| -------------------------------- | -----------: | -----------------: | ---------------------: | ---------------------: |
| **Base price**                   |       **$0** | **$29 / app / mo** | **$149 / mo (5 apps)** |  **$1,999 / mo flat**  |
| **LLM provider**                 |   Heimdall   |    Heimdall        |    Heimdall            |   **BYOK** (your key)  |
| **Investigations included**      |  3 (capped)  |       10 / mo      |     100 / mo           |       Unlimited        |
| **Investigation overage**        |    — (cap)   |        $2 each     |       $2 each          |    n/a (you pay LLM)   |
| **Log volume included**          |  1 GB (cap)  |    25 GB / app     |    250 GB total        |       Unlimited        |
| **Log overage**                  |    — (cap)   |     $0.50 / GB     |    $0.50 / GB          |   n/a (you own storage)|
| **Historical retention**         |    48 hours  |       30 days      |       30 days          |  Custom (your storage) |
| **Applications**                 |       1      |          1         |         5              |       Unlimited        |
| **Storage location**             |   Heimdall   |    Heimdall        |    Heimdall            |   S3 / GCS / Postgres  |
| **Default model**                |    Sonnet    |     Sonnet         |     Opus               |   Customer's choice    |
| **Investigation targets**        |  Unlimited   |    Unlimited       |    Unlimited           |       Unlimited        |
| **Ingestion sources**            |  Unlimited   |    Unlimited       |    Unlimited           |       Unlimited        |
| **Deploy correlation**           |       ✓      |          ✓         |          ✓             |          ✓             |
| **Auto-investigation reports**   |       ✓      |          ✓         |          ✓             |          ✓             |
| **MCP server**                   |       ✓      |          ✓         |          ✓             |          ✓             |
| **Trajan ticket creation**       | With Trajan  |    With Trajan     |    With Trajan         |    With Trajan         |
| **SSO**                          |       —      |          —         |          ✓             |          ✓             |
| **SOC 2 / DPA**                  |       —      |          —         |          —             |          ✓             |
| **Support**                      |  Community   |        Email       |     Priority email     |  Dedicated Slack + SLA |
| **Contract**                     |       —      |  Monthly / Annual  |    Monthly / Annual    |     Annual only        |
| **Annual discount (base only)**  |       —      |         20%        |        20%             |   Negotiated           |

---

## Each Tier in One Sentence

The model is intentionally simple enough to recite from memory:

* **Hobby** — Free, with 3 investigations and 1 GB of logs to try Heimdall on us.
* **Pro** — $29/month per app, with 10 investigations and 25 GB of logs included; overage at $2 per investigation and $0.50 per GB.
* **Business** — $149/month for 5 apps, with 100 investigations and 250 GB of logs included at the same overage rates, plus SSO.
* **Enterprise** — $1,999/month flat — bring your own Anthropic key and your own database for unlimited everything.

Pro and Business share the same overage rates ($2 / investigation, $0.50 / GB) and the same retention (30 days). The only things that change between them are scale (apps, included quota) and SSO. This is deliberate: customers don't have to do "but at what overage rate?" math when comparing tiers, and the upgrade decision becomes pure arithmetic on volume.

---

## Typical Customer Bills

Base prices reflect the entry point. Real bills depend on how active each application is.

| Customer profile                                  | Plan            |        Monthly bill | Notes                                              |
| ------------------------------------------------- | --------------- | ------------------: | -------------------------------------------------- |
| Weekend project / Hobby                           | Hobby           |              **$0** | 3 investigations, then upgrade prompt              |
| Indie startup, ~10 investigations / mo            | Pro             |             **$29** | Stays inside included quota                        |
| Active SaaS, ~30 investigations / mo              | Pro             |             **$69** | $29 base + 20 × $2 overage                         |
| Production SaaS, ~80 investigations / mo          | Pro             |            **$169** | $29 base + 70 × $2 overage                         |
| Multi-app team, ~300 investigations / mo (5 apps) | Business        |            **$549** | $149 base + 200 × $2 overage                       |
| Heavy production, ~600 investigations / mo        | Business        |          **$1,149** | Approaching the Enterprise crossover               |
| Regulated enterprise, unlimited everything        | Enterprise BYOK |  **$1,999** + LLM   | Customer pays Anthropic directly                   |

Because Pro and Business share overage rates, the upgrade decision is pure arithmetic on **included quota**: a Pro customer at ~70 investigations / month is paying $169, and Business at $149 already includes more than that — so 70+ investigations is the natural Pro → Business signal. Above ~925 investigations / month, Business and Enterprise cross over (and customers above that threshold typically also want SOC 2 / data residency, which only Enterprise provides).

---

## Competitive Check

Heimdall sits inside the **10–20% of competitor cost** band across every realistic customer profile, with 5x cheaper at the high end and 50–100x cheaper at the low end.

| Customer profile                          | Heimdall    | Datadog mid-market[^dd] | incident.io / Resolve.ai[^ai] |
| ----------------------------------------- | ----------: | ----------------------: | ----------------------------: |
| Indie startup                             |        $29  |       $4,000 – $12,000  |             $1,600 – $8,000   |
| Active SaaS                               |        $69  |       $4,000 – $12,000  |             $1,600 – $8,000   |
| Production SaaS (single app)              |       $169  |       $4,000 – $12,000  |             $1,600 – $8,000   |
| Multi-app team (5 apps)                   |       $549  |       $8,000 – $20,000  |             $3,000 – $15,000  |
| Regulated enterprise (BYOK + BYODB)       |     $1,999  |     $8,000 – $33,000+   |             $8,000 – $12,000+ |
| **Heimdall as % of competitor**           |             |          **0.5 – 7%**   |                  **2 – 25%**  |

[^dd]: Per-month equivalent of Datadog mid-market annual contracts ($50K–$400K range), per the market-research deck.
[^ai]: AI-SRE peer category — incident.io published pricing ($20K–$100K/year) and Resolve.ai estimated enterprise pricing ($100K+/year).

The key structural advantage: **Heimdall does not double-dip on ingestion + indexing** (Datadog's primary pricing complaint), does not charge per-host (Dynatrace), does not charge per-seat (incident.io), and does not bill for investigation runtime as a separate line item. One base + one investigation meter + one log-volume safety meter — that's the entire bill.

---

## What Counts as an Investigation

This is the central billing unit, so the definition needs to be precise. An **investigation** is a Claude reasoning loop with tool access (search_logs, query_database, repo inspection) that produces a structured assessment.

**Counted:**

- Lumber-flagged log → Claude assessment (the monitoring loop's primary mode)
- Scheduled cron-based investigation runs (you opt in, you pay)
- Manually triggered "investigate this" actions from the UI

**Not counted:**

- Interactive chat sessions (user-pulled, conversational — capped only by reasonable use)
- Failed or aborted investigations (model errored, timed out, infra fault on our side)
- Lumber classification itself (runs locally on our ONNX model, no LLM cost)
- Log ingestion, parsing, retention reads (covered by the log-volume meter)

The chat exemption is deliberate. Half the product's value is the on-demand interrogation surface; making chats meter would push customers off it, which is the opposite of what we want.

The "failed investigation" exemption is a trust signal — customers should never pay for our infra problems.

---

## Definitions

### Monitored Application

A **Monitored Application** is one logical production system — an API backend, a frontend app, a worker service. It includes its ingestion sources, investigation targets, deploy signals, and supporting data stores. It is *not* priced by number of databases, servers, or repositories. Infrastructure complexity is abstracted away from billing logic.

Investigation targets are unlimited on every tier. Connect everything — more context makes the agent better.

### Ingestion Source

An **Ingestion Source** is a distinct connected telemetry input that streams data into Heimdall continuously: a CloudWatch log group, a Vercel deployment log stream, a Postgres log feed, a Sentry project, a CI/CD deploy signal.

Ingestion sources are **unlimited on all tiers**. (The previous pricing model billed per source; that line item has been removed because it created perverse incentives — customers were under-connecting their infra to save $15/mo, which made the agent worse at its job.)

The cost of running an ingestion source is captured in the log-volume meter, which is the right place for it: a connector that streams 200 GB/month costs us more than one that streams 200 MB/month, regardless of how the customer logically draws the boundary.

### Investigation Target

An **Investigation Target** is an on-demand data source the agent queries during investigations: a GitHub repository, a Postgres database, a Redis instance, an external API. Unlike ingestion sources, these do not stream data in.

Investigation targets are **unlimited on all tiers**. The cost of using them is captured per-investigation, not per-target.

---

## Tier Differentiation Logic

Plans differ along five dimensions. Each dimension addresses a specific customer maturity level.

### 1. Investigation Quota

The primary cost-aligned meter. Each investigation burns LLM tokens (roughly $0.30–$0.80 on Sonnet, $1.50–$3.00 on Opus with extended thinking). The included quota grows with tier; the overage rate is held flat across paid tiers for simplicity.

| Tier       | Included          | Overage      |
| ---------- | ----------------- | -----------: |
| Hobby      | 3 (hard cap)      |   — (cap)    |
| Pro        | 10 / mo           |    $2 each   |
| Business   | 100 / mo          |    $2 each   |
| Enterprise | Unlimited (BYOK)  |        n/a   |

The flat $2 overage rate is set to maintain healthy margins on typical Sonnet investigations (~75%) while staying defensible on heavier Opus runs (~25%). Customers with sustained heavy usage upgrade to Business (or Enterprise) for the larger included quota, not for a discount on overage.

If a customer hits their included quota and overage billing is not enabled, investigations are paused (with banner notification) rather than cut off mid-execution. Customers can flip the overage toggle from the dashboard at any time.

### 2. Log Volume

A safety meter, not the primary cost driver. Included volumes scale with tier; the overage rate is held flat across paid tiers.

| Tier       | Included          | Overage     |
| ---------- | ----------------- | ----------: |
| Hobby      | 1 GB (hard cap)   |   — (cap)   |
| Pro        | 25 GB / app       | $0.50 / GB  |
| Business   | 250 GB total      | $0.50 / GB  |
| Enterprise | Unlimited         |        n/a  |

Log volume is measured in **GB ingested**, not number of logs and not LLM tokens. This aligns with industry billing standards and with the actual underlying storage and processing cost.

### 3. Historical Retention

How long logs and investigation outputs are retained on Heimdall's infrastructure.

* Hobby: 48 hours
* Pro: 30 days
* Business: 30 days
* Enterprise: Custom (logs drain to your S3 / GCS / Postgres; you own retention)

Retention is held at 30 days for both paid tiers — long enough for meaningful correlation and historical pattern detection on production workloads, short enough to keep the model simple. Customers needing longer retention move to Enterprise, where they own the storage and set their own policy.

### 4. Applications Covered

* Hobby: 1 application
* Pro: 1 application (additional apps at $29 / mo each, Pro limits apply)
* Business: 5 applications included (additional apps at $29 / mo each, Pro-tier limits per added app)
* Enterprise: Unlimited applications

Customers running multiple production systems are the natural Business and Enterprise audience.

### 5. BYOK & BYODB (Enterprise only)

Enterprise unlocks two architectural changes that are unavailable at lower tiers:

**BYOK (Bring Your Own Key):** the customer supplies their own Anthropic API key. Every investigation Heimdall runs on their behalf bills directly to their Anthropic account, not to Heimdall's. This removes the per-investigation meter entirely and unlocks unlimited investigations, because the variable cost has moved off Heimdall's P&L.

**BYODB (Bring Your Own Database):** logs and investigation outputs persist to the customer's own Postgres database (and optionally drain to S3 / GCS for cold storage). Heimdall's process still runs the agent and processes the logs, but **nothing is persisted on Heimdall's infrastructure** — the customer owns the data plane end-to-end.

Together, BYOK + BYODB give Enterprise customers a clean answer for security review: "logs and AI outputs never live in a third-party database." This is positioned as **data residency**, not zero-trust isolation (logs still transit Heimdall's process); see `enterprise-archi.md` for the architectural detail.

The Enterprise base price ($1,999 / mo) covers the harness — ingestion, classification, agent orchestration, dashboards, MCP, support, SOC 2 — and is held flat regardless of LLM or storage cost, because BYOK/BYODB is what the customer values, not what reduces our cost.

---

## Hobby Tier — Hard Caps, Not Overage

Hobby is a true demo tier. When a Hobby user hits 3 investigations or 1 GB of logs, the wall says "upgrade to Pro" — there is no overage option, no surprise bill, no card charge for a $0 plan.

This is deliberate:

* No support tickets about $4 invoices on free accounts.
* No "I thought it was free" friction.
* Crisp psychological boundary: the wall is the conversion event.

The 3-investigation cap costs Heimdall roughly $3 of unrecoverable LLM spend per signup — treat that as the customer-acquisition cost line, not as a free product. At even a 5% Hobby → Pro conversion rate, payback is in month one.

---

## Annual Billing

All paid tiers offer a **20% discount on the base price** for annual commitment. Enterprise pricing is negotiated; the published $1,999 / mo is the monthly-equivalent floor.

| Tier       | Monthly      | Annual (per month) | Annual (total) |
| ---------- | -----------: | -----------------: | -------------: |
| Pro        |        $29   |              $23   |          $279  |
| Business   |       $149   |             $119   |        $1,428  |
| Enterprise |     $1,999   |   Negotiated       |   Negotiated   |

The annual discount applies to the base price only. Usage-based overages (investigations, log volume) bill at the same per-unit rate regardless of billing cycle — this prevents customers from gaming the discount by front-loading commitment then driving heavy usage.

---

## Pricing Philosophy

Heimdall uses a **land-and-expand** model with one twist that distinguishes it from Supabase / Vercel-style usage pricing: the meter is investigations, not infrastructure consumption.

**Why this matters:**

* **$29 / mo is a credit-card swipe. $149 / mo is a manager conversation.** The Pro entry price determines conversion rate. For founding engineers, vibe-coders, and indie devs — Heimdall's first 10 customers — the bar must be low enough that trying it is a non-event.
* **The investigation meter aligns cost with value.** Idle apps pay nothing extra. Active apps pay for the work the agent does on their behalf. This is fairer and more defensible than front-loading cost on log volume or seat counts, which is what every incumbent does.
* **Customers who outgrow Pro don't feel trapped.** A Pro customer hitting 80 investigations / month sees Business at $149 with a much larger included quota and recognises the upgrade as math, not coercion. The product sells the upgrade; sales doesn't have to.
* **Each tier fits in one sentence.** The model is simple enough to recite from memory, simple enough to compare against any competitor's pricing page in 30 seconds, and simple enough that a customer can predict next month's bill without a calculator. This is a real product feature — every observability incumbent has lost deals on pricing complexity alone.
* **BYOK at the top is a feature, not a discount.** Enterprises want BYOK independent of price — for compliance, data residency, and direct cost control with their existing Anthropic enterprise contract. Heimdall holds the harness fee flat at $1,999 because that's what the customer is buying; the inference cost is theirs to bear, and they prefer it that way.

This follows the pattern established by Vercel ($0 → $20 base, scales with edge / function usage) and Supabase ($0 → $25 base, scales with DB / storage), with the meter changed to match Heimdall's actual cost driver: AI investigation runtime, not infrastructure consumption.

---

## Trajan Integration

Automated ticket creation is available on all tiers **via Trajan integration**.

Heimdall does not maintain its own ticketing system. When the agent completes an investigation, it produces a structured report available to all users. For teams using Trajan, this report can automatically generate a tracked ticket in Trajan's workflow system.

* Investigation reports: available to everyone, all tiers
* Automated ticket creation: requires Trajan

This positions Trajan as the natural execution layer without making Heimdall feel incomplete without it. Heimdall handles runtime intelligence; Trajan handles execution and workflow. Together they form a cohesive system.

---

## Product Positioning

Heimdall is positioned as:

> **An autonomous runtime investigator for production systems.**

It is:

* **Not priced per seat** — teams add unlimited viewers at every tier
* **Not priced per host** — infrastructure complexity does not affect the bill
* **Not token-metered** — customers pay per investigation, not per LLM token (except at Enterprise, where they pay Anthropic directly)
* **Not a basic log viewer** — the value is in autonomous investigation, not log search
* **Not a ticketing tool** — Trajan integrates for that

The cheapest paid tier is $29 / mo; the most expensive published tier is $1,999 / mo. Across that 70x range, the pricing model has the same shape: a base for "Heimdall is watching" plus a meter for "Heimdall is investigating." Easy to explain, easy to compare, easy to forecast.

---

## Open Questions for Future Iteration

1. **BYOK on Business as a paid add-on?** A motivated mid-market customer might want BYOK without going to Enterprise. Possible: $99 / mo Business add-on that drops the per-investigation overage to $0. Leaks Enterprise demand downward but generates incremental revenue. Not in v1; revisit if mid-market demand surfaces in sales calls.
2. **Investigation quota separation.** Should "auto-investigations triggered by Lumber" and "scheduled cron investigations" share a single quota, or have separate meters? Single quota is simpler; split quotas let customers schedule aggressively without burning their reactive budget. Default to single; revisit on customer feedback.
3. **Fair-use cap on chats.** Chats are uncounted, but at extreme volumes they're real LLM cost. Worth quietly publishing a fair-use line ("~500 chat messages / day per user, contact us above") so the model isn't open-ended on the upside.
4. **Enterprise pricing ceiling.** The published $1,999 is a floor. Negotiated Enterprise contracts for orgs with 50+ apps, custom integrations, on-prem, or dedicated infra can land much higher and remain inside the 10–20% of competitor cost band. Keep "custom pricing for orgs > 50 apps" as the standard escape hatch in sales conversations.
