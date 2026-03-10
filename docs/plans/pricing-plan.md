# Heimdall Pricing Overview

**Pricing is usage-based with a low base price per Monitored Application.**

A *Monitored Application* represents one deployable production system (e.g. API backend, frontend app, worker service, etc.), including its ingestion sources, investigation targets, deploy signals and repositories.

Base prices cover platform access and modest included allowances. Real cost scales with usage — log volume, agent runtime, and ingestion sources — so customers only pay more as they get more value.

---

## Pricing Table

| Feature                          |         Starter |          Growth |                      Scale |
| -------------------------------- | --------------: | --------------: | -------------------------: |
| **Base price (per app / month)** |             $39 |            $149 |                       $449 |
| **Applications included**        |               1 |               1 |                          3 |
| **Additional application**       |       $29 / mo  |      $119 / mo  |                 $149 / mo  |
| **Included log volume**          |    5 GB / month |  50 GB / month  |  150 GB / month (per app)  |
| **Log overage**                  |      $1.00 / GB |      $0.50 / GB |                 $0.30 / GB |
| **Ingestion sources included**   |               1 |               5 |               15 (per app) |
| **Additional ingestion source**  |   $15 / mo each |   $12 / mo each |              $10 / mo each |
| **Investigation targets**        |       Unlimited |       Unlimited |                  Unlimited |
| **Per-investigation runtime**    |          10 min |          30 min |                     60 min |
| **Included monthly runtime**     |            1 hr |          10 hrs |           30 hrs (per app) |
| **Runtime overage**              |        $15 / hr |        $12 / hr |                  $10 / hr  |
| **Historical retention**         |          7 days |         30 days |                    90 days |
| **Deploy correlation**           |               ✓ |               ✓ |                          ✓ |
| **Auto-investigation reports**   |               ✓ |               ✓ |                          ✓ |
| **Trajan ticket creation**       |     With Trajan |     With Trajan |                With Trajan |
| **Annual discount**              |             20% |             20% |                        20% |

### Typical Monthly Spend

Base prices reflect the entry point, not the expected bill. Real usage for active applications will scale naturally:

| Profile                          | Starter       | Growth         | Scale (per app)  |
| -------------------------------- | ------------: | -------------: | ---------------: |
| **Base**                         |           $39 |           $149 |             $449 |
| **~15 GB logs**                  |     +$10 logs |             — |               — |
| **~60 GB logs**                  |             — |      +$5 logs  |               — |
| **~200 GB logs**                 |             — |             — |       +$15 logs  |
| **+2 ingestion sources**         |           +$30 |             — |               — |
| **+3 ingestion sources**         |             — |          +$36 |               — |
| **~3 hrs runtime**               |  +$30 runtime |             — |               — |
| **~15 hrs runtime**              |             — | +$60 runtime   |               — |
| **~40 hrs runtime**              |             — |             — |  +$100 runtime   |
| **Typical total**                |      **~$109** |      **~$250** |        **~$564** |

---

# Definition: Monitored Application

A **Monitored Application** is:

* One logical production system
* With its ingestion sources (logs streaming in)
* Its investigation targets (repos, databases, services the agent can query on demand)
* Its deploy history
* Its supporting data stores

It is *not* priced by number of databases, servers, or repositories.
Infrastructure complexity is abstracted away from billing logic.

Investigation targets are unlimited — the cost of the agent querying a repo or database during an investigation is captured in agent runtime, not connection count. Users should be encouraged to connect everything; more context makes the agent better.

---

# Definition: Ingestion Source

An **Ingestion Source** is a distinct connected telemetry input that streams data into Heimdall continuously.

Examples of what counts as one ingestion source:

* A CloudWatch log group
* A Kubernetes namespace log stream
* A Datadog service feed
* A Sentry project
* A Postgres log stream
* A Redis log stream
* A Vercel deployment log stream
* A CI/CD pipeline deploy signal

In general:

> If it is a separate external stream Heimdall connects to and processes independently, it counts as one ingestion source.

Multiple containers feeding into a single log aggregator = 1 source.
Separate services each with distinct log groups = multiple sources.

This keeps billing objective and prevents ambiguity.

---

# Definition: Investigation Target

An **Investigation Target** is an on-demand data source the agent queries during investigations. Unlike ingestion sources, these do not stream data in — the agent reaches for them when it needs context.

Examples:

* A GitHub repository (to inspect recent commits or code)
* A Postgres database (to run diagnostic queries)
* A Redis instance (to check cache state)
* An external API (to verify upstream health)

Investigation targets are **unlimited on all tiers**. The cost of using them is captured in agent runtime, not connection count. This avoids creating a perverse incentive to limit the agent's access.

---

# Volume Model

Each Monitored Application includes a tier-dependent log volume:

* Starter: 5 GB / month
* Growth: 50 GB / month
* Scale: 150 GB / month (per app)

Included volumes are intentionally modest — they cover light or early-stage usage. Active production applications will typically exceed included volume, with overages billed per additional GB at a rate that decreases with tier:

* Starter: $1.00 / GB
* Growth: $0.50 / GB
* Scale: $0.30 / GB

Log volume is measured in **GB ingested**, not number of logs and not LLM tokens.

This aligns pricing with:

* Storage cost
* Processing cost
* Observability industry standards

---

# Tier Differentiation Logic

Plans differ in four dimensions:

## 1. Agent Runtime

Agent runtime is the primary differentiator and the axis most aligned with cost. Each investigation consumes agent runtime as the LLM reasons, queries tools, and builds its diagnosis.

**Per-investigation cap** — the maximum wall-clock time for a single investigation:

* Starter: 10 minutes — Quick-response investigations. Agent checks recent logs, correlates with the last deploy, produces a report. Handles the majority of straightforward issues.
* Growth: 30 minutes — Multi-source deep investigations. Agent cross-references logs with database state, checks the repo for recent changes, reviews historical patterns within the retention window.
* Scale: 60 minutes — Full investigative sweeps. Broad historical context, multi-source correlation, extended reasoning chains. For complex multi-factor production incidents.

**Included monthly runtime** — agent compute time included in the base price:

* Starter: 1 hour
* Growth: 10 hours
* Scale: 30 hours (per app)

**Runtime overage** — billed per additional hour when the included budget is exceeded:

* Starter: $15 / hr
* Growth: $12 / hr
* Scale: $10 / hr

Runtime overage is the primary revenue scaling mechanism on Starter. An indie dev whose app generates a few investigations per week will stay near the base price. A production app with frequent incidents naturally consumes more runtime and pays proportionally.

If runtime is exhausted and overage billing is not enabled, investigations run at reduced depth (graceful degradation) rather than hard cutoff.

---

## 2. Log Volume

Included volume scales with tier. Starter includes enough for light or early-stage usage; production workloads are expected to exceed it:

* Starter: 5 GB / month
* Growth: 50 GB / month
* Scale: 150 GB / month (per app)

---

## 3. Historical Retention

How long logs and investigation outputs are retained:

* Starter: 7 days
* Growth: 30 days
* Scale: 90 days

Retention directly affects correlation depth and pattern detection. Longer retention enables the agent to recognise recurring failure modes and reference past incidents.

---

## 4. Applications Covered

* Starter and Growth: 1 Monitored Application
* Scale: 3 Monitored Applications

Additional applications can be added at a discount relative to the base price:

* Starter: $29 / mo per additional app (Starter limits apply)
* Growth: $119 / mo per additional app (Growth limits apply)
* Scale: $149 / mo per additional app (Scale limits apply)

Scale is intended for teams running multiple production systems.

---

# Core Capabilities (All Tiers)

All plans include:

* Log ingestion and anomaly detection
* Change / Deploy Correlation
  (Automatically linking runtime anomalies to recent commits or deploys)
* Structured Auto-Investigation Reports
* Investigation targets (repos, databases) — unlimited

---

# Trajan Integration

Automated ticket creation is available on all tiers **via Trajan integration**.

Heimdall does **not** maintain its own ticketing system. When the agent completes an investigation, it produces a structured report available to all users. For teams using Trajan, this report can automatically generate a tracked ticket in Trajan's workflow system.

* Investigation reports: available to everyone, all tiers
* Automated ticket creation: requires Trajan

This positions Trajan as the natural execution layer without making Heimdall feel incomplete without it.

---

# Annual Billing

All tiers offer a **20% discount** on the base price for annual commitment:

| Tier    | Monthly | Annual (per month) | Annual (total) |
| ------- | ------: | -----------------: | -------------: |
| Starter |     $39 |                $31 |           $372 |
| Growth  |    $149 |               $119 |         $1,428 |
| Scale   |    $449 |               $359 |         $4,308 |

Annual discount applies to the base price only. Usage-based overages (logs, runtime, ingestion sources) are billed at the same rates regardless of billing cycle.

---

# Pricing Philosophy

Heimdall uses a **land-and-expand** pricing model. The base price is set low enough that an indie developer or small team can start monitoring without a significant commitment. Revenue scales with usage as the product proves its value.

**Why this matters for a new product:**

* **$39/mo is a credit card swipe. $149/mo is a decision that needs approval.** The entry price determines conversion rate. For vibecoders and indie devs — our early adopters — the bar must be low enough that trying Heimdall is a non-event.
* **Usage-based pricing aligns cost with value.** A customer only pays more when they're ingesting more logs, running more investigations, and getting more value. This is fairer and more defensible than front-loading cost.
* **The total spend for an active user lands in the same range.** A Starter customer with real production traffic will typically spend $100–150/mo. The difference is psychological: they *grew into* that spend rather than committing upfront.

This follows the pattern established by Supabase ($0 → $25 base, scales with usage), Vercel ($0 → $20 base), and PlanetScale — low entry, usage-based expansion, natural upgrade path to higher tiers when limits are hit frequently.

---

# Product Positioning

Heimdall is positioned as:

> An autonomous runtime investigator for production systems.

It is:

* Not priced per seat
* Not token-metered
* Not a basic log viewer
* Not a ticketing tool

Heimdall handles runtime intelligence.
Trajan handles execution and workflow.

Together they form a cohesive system.

---
