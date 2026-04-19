# Heimdall: Market Intelligence & Go-To-Market Strategy
## AI-Native Infra Monitoring Agent — Competitive Landscape, Pricing, ICP, and Product Blueprint

***

## Executive Summary

Heimdall enters one of the fastest-growing infrastructure software categories at an unusually good moment. The broader observability and log management market is valued between $3–28B depending on scope definition, with consistent 12–23% CAGR forecasts across all segments through 2034. The incumbents (Datadog, Dynatrace, Splunk) charge enterprise prices for enterprise complexity — largely inaccessible to the 10–200 engineer teams who need 24/7 reliability most acutely. A new tier of AI-native SRE agents (Resolve.ai, Cleric, Traversal) has emerged and is attracting major VC capital, but these too target enterprise customers with custom pricing. The gap Heimdall exploits: an AI-native, developer-first, root-cause-to-findings pipeline that integrates log feeds, databases, and codebases simultaneously — at a price point and onboarding experience that appeals to the Vercel-generation of cloud-native developers.[1][2][3]

***

## Market Size & Growth

The market Heimdall plays in spans several overlapping categories. Taking the most conservative definition:

- **Log Management alone**: $3.27B in 2024, projected $10.08B by 2034 at a CAGR of 11.92%[1]
- **Observability Platforms** (broader): $2.1B in 2025, growing to $13.9B by 2034 at 23.3% CAGR[2]
- **Full Observability Tools and Platforms** (widest scope, including APM + AIOps): estimated $28.5B in 2025, projecting $172.1B by 2035 at 19.7% CAGR[3]
- **Data Observability** (a sub-segment aligned to Heimdall's DB-connected approach): $3.51B in 2026, reaching $6.03B by 2031[4]

The combined log management + APM + AI SRE serviceable market is a realistic $8–15B addressable globally today, growing well into double digits annually. Telemetry volume is climbing 35% each year, driven by microservices, serverless, and AI workloads.[4]

**Cloud deployment is now dominant**, commanding 68% of log management deployments, and OpenTelemetry standardization is reducing vendor lock-in concerns, making it easier for new entrants to offer competing solutions on open infrastructure.[5][6]

### Regional Breakdown

| Region | Market Share (2024–2025) | CAGR | Key Notes |
|--------|--------------------------|------|-----------|
| North America | ~38.8% of log management[5] | ~12% | Largest market; US enterprise dominates |
| Asia-Pacific | 34% of log management growth[7] | 13.2–19.6%[8][9] | Fastest-growing region overall |
| Europe | ~$504M in 2025 (observability platforms)[2] | 22.1% | GDPR compliance a key driver |
| Singapore/ANZ | Part of APAC cluster with strong cloud adoption[10] | — | High relevance for Heimdall's home market |

India's observability market is growing at 13.9% CAGR; Japan at 12.5%. China commands ~1/3 of APAC revenue. Singapore, Australia, and South Korea are flagged as having "strong cloud adoption" in analyst reports.[8][10]

***

## Competitive Landscape

The competitive field breaks into four distinct tiers with very different positioning, pricing philosophies, and ICPs.

### Tier 1: Legacy Enterprise Behemoths

These are the dominant revenue holders — complex platforms with broad feature surface areas, agent-based instrumentation, and enterprise sales motions.

**Datadog** is the category leader with $3.43B in FY2025 revenue (+28% YoY), ~32,700 customers, and 84% of customers using 2+ products. Its pricing is notoriously complex: infrastructure from $15/host/month, APM from ~$31/host/month, and log management at $0.10/GB ingested *plus* $1.70/million events indexed (double-dipping). Mid-sized companies routinely spend $50K–$150K per year; enterprise deployments exceed $1M. Datadog launched its "Bits AI" trio of SRE, Dev, and Security Analyst agents at DASH 2025, but these are add-ons to an already expensive platform stack.[11][12][13][14][15]

**Dynatrace** prices on a per-host-unit model at $0.08/hour/8GiB host for full-stack monitoring, with logs at $0.20/GiB. Its "Davis AI" engine provides automated root-cause analysis through causal topology mapping (Smartscape), correlating all symptoms to a single root cause alert. Davis is the closest existing AI-RCA feature to what Heimdall does natively — but it's embedded in a $100K–$400K/year platform and requires Dynatrace's OneAgent for data collection. Dynatrace was named a Gartner Magic Quadrant Leader for the 15th consecutive time.[16][17][18][19][20][21]

**Splunk (Cisco)** prices at $150–225/GB/day for the cloud platform. At 10 GB/day, that's $55K–$82K/year just for ingest. Enterprise-only in practice. New Relic offers a more accessible model at $49/user/month for core users and $0.30–$0.60/GB/month for data ingestion, with typical annual costs of $80K–$300K for mid-market teams.[17][22][23]

### Tier 2: Modern Mid-Market Challengers

These tools have modernized pricing but still largely position around observability *data infrastructure*, not autonomous *investigation*.

**SigNoz** is the most aggressive on pricing: Teams plan starts at $49/month (down from $199), with $0.30/GB for logs and traces, $0.10/million samples for metrics, and no per-host or per-user pricing. A startup program at $19/month exists. Built on OpenTelemetry. Strong for dev teams but no autonomous AI investigation layer.[24][25][26]

**Coralogix** uses per-GB consumption: logs at $0.42/GB, traces at $0.16/GB, metrics at $0.05/GB, with unlimited users and hosts included. Their "TCO Optimizer" allows tiered data routing to control costs. AI tokens at $1.50/million. 100 GB/day ≈ ~$20K/year. Growing feature set but primarily a data pipeline platform.[27][28][29]

**Better Stack (Logtail)** bundles uptime monitoring, log management, incident management, and status pages: Team plan at $24/month, Business at $84/month. Claims 30x cheaper than Datadog. Newer AI SRE features available at ~$29/responder/month. Strong DX, but AI investigation is not its core competency.[30][31][32]

**Grafana (Loki)** is effectively free as open source but requires significant self-managed infrastructure. Cloud pricing at ~$0.50/GB ingestion. Most teams use it for dashboards and visualization, not autonomous investigation.[33]

**middleware.io** offers $0.30/GB for metrics/logs/traces with 100 GB free and an "OpsAI" layer for error solving. Consumer-grade pricing, very early stage on AI.[34]

**Logz.io** (Open 360 platform) launched "Open 360 AI" specifically designed for AI agents and humans to co-investigate incidents — the closest product philosophy to Heimdall among Tier 2. Starts at $38/month.[35][36]

### Tier 3: Pure-Play AI SRE Agents (Closest Architectural Cousins)

This is the newest and fastest-moving tier — standalone AI agents that investigate production incidents autonomously. They are Heimdall's closest conceptual competitors but currently target enterprise only.

| Tool | Key Claim | Pricing | Target | Funding |
|------|-----------|---------|--------|---------|
| **Resolve.ai** | 80% autonomous incident resolution; multi-agent parallel hypothesis testing[37][38] | Custom enterprise[39] | Fortune 500 | $250M Series A, $1B valuation (Dec 2025)[40] |
| **Traversal** | Causal ML, 300M logs/incident, 90%+ accuracy; DigitalOcean 36K hrs/year saved[38] | Enterprise (AWS Marketplace)[41] | Fortune 100, AmEx, Pepsi | Sequoia + Kleiner Perkins[41] |
| **Cleric** | Self-learning AI SRE; learns from every incident; Slack-native findings delivery[42] | Custom | Series-A and above | Undisclosed |
| **incident.io** | Netflix/Etsy trusted; full incident lifecycle + AI SRE; auto fix PRs[37][38] | $15–25/user/month + $10–20 on-call[37] | Series A-C SaaS | — |
| **Datadog Bits AI** | Native integration, trained on 2,000+ customer environments[43] | Add-on to Datadog | Existing Datadog customers | Built-in |
| **IncidentFox (YC W26)** | Apache 2.0, zero-config; 85–95% alert noise reduction; Docker in 5 min[44] | OSS / SaaS (early) | Dev teams, SMB | YC W26 |

**Critical gap Heimdall fills**: Resolve.ai, Traversal, and Cleric all require you to *already have* an observability stack (Datadog/Grafana) and layer on top. Heimdall is the stack. It collects, triages, investigates, and reports — with the unique addition of direct codebase and database integration for root cause analysis, which none of the above offer out of the box.

### Pricing Overview

| Platform | Starting Price | Typical Mid-Market Annual Cost | AI Investigation |
|----------|---------------|-------------------------------|-----------------|
| Datadog | $15/host/month[11] | $50K–$150K[12] | Bits AI add-on |
| Dynatrace | $0.08/hour/host[18] | $100K–$400K[17] | Davis AI (included) |
| Splunk | $150–225/GB/day[23] | $200K–$800K+[17] | Limited |
| New Relic | $49/user/month[22] | $80K–$300K[17] | AIOps module |
| SigNoz | $49/month base[25] | $5K–$30K | None |
| Coralogix | $0.42/GB logs[27] | $15K–$50K | Early stage |
| Better Stack | $24/month[30] | $300–$12K[45] | Basic AI SRE |
| Resolve.ai | Custom enterprise[39] | $100K+ est. | Core product |
| incident.io | $31–45/user/month[32] | $20K–$100K | Full AI SRE |
| **Heimdall (target)** | **~$29–99/project/month** | **$500–$10K** | **Core product** |

***

## Industry Verticals Most Likely to Buy

Observability adoption and willingness-to-pay are highest where downtime has a direct, immediate revenue or compliance impact.

**1. SaaS / Tech Startups (Primary)**
The highest-density potential customer base. Series A-C companies have production systems but no dedicated SRE team. Incidents cost revenue and reputation, on-call fatigue is endemic. They're already using Vercel, AWS, and Supabase. This is Heimdall's natural home market.[46]

**2. FinTech / Payments (High-Value)**
Uptime is regulatory compliance. A payment processor down for 5 minutes is catastrophic. BFSI held the largest application-specific share of log management in 2025, with SIEM integration at 39.45% of revenue. Singapore's fintech ecosystem (MAS-regulated, fast-growing) is an ideal early beachhead.[47]

**3. E-commerce / Marketplace Platforms**
Revenue is directly correlated to uptime. A database anomaly during peak traffic needs to be found in minutes, not hours. Particularly high-value in APAC where e-commerce is growing rapidly.[8]

**4. AI-Native Companies (Emerging High-Priority)**
Datadog itself cited 14 of the top 20 AI-native firms as customers. SigNoz specifically noted a surge in AI-native companies signing up. AI companies need to monitor model serving infrastructure, database performance for vector stores, and agentic pipeline logs — exactly what Heimdall covers. These companies also have the highest cultural affinity for AI-native tooling.[14][24]

**5. HealthTech / MedTech (Compliance-Driven)**
HIPAA compliance requires audit trails. Database anomaly detection is critical. Slower sales cycle but very sticky once adopted.

***

## Ideal Customer Profile (ICP)

### ICP A — First 10 Teams: "The Founding Engineer"

These are the people who will actually sign up via PLG:

- **Role**: CTO/founding engineer, or head of backend/infra at a seed-to-Series A startup
- **Company size**: 3–30 engineers, no dedicated SRE or DevOps hire yet
- **Stack**: Vercel or Railway/Render/Fly.io + AWS/Supabase + GitHub; building with Next.js, FastAPI, Node
- **Pain**: Post-incident manual log spelunking taking 2–6 hours; Datadog too expensive and complex; no one wants to be on-call at 3am
- **Budget authority**: They are the buyer. Budget is $50–$500/month
- **Location**: SF, NYC, Berlin, Singapore, London; active on Twitter/X, GitHub, Hacker News
- **Discovery path**: Dev Twitter, Product Hunt, GitHub trending, Hacker News "Show HN"
- **Trigger event**: Just had a bad incident, or just got scared of their Datadog bill

**Why Heimdall wins with them**: Zero-config setup (GitHub app + Vercel integration + DB connection string), immediate value on first incident, pricing that fits a startup budget, reports they can hand to an AI coding agent (Cursor/Windsurf) to implement the fix.

### ICP B — First 100 Teams: "The Pragmatic Engineering Manager"

- **Role**: VP Engineering or Engineering Manager at a Series A–B company (30–150 engineers)
- **Company size**: 30–200 employees, engineering team with some specialization but no SRE team
- **Stack**: Multi-cloud (AWS + GCP), PostgreSQL/Aurora/MongoDB, GitHub/GitLab
- **Pain**: Spending $8K–$30K/year on Datadog but still getting paged for incidents that take hours to resolve; teams lack observability culture; post-mortems are painful
- **Budget authority**: Engineering manager or VP; $1K–$5K/month is justifiable with ROI argument
- **Decision criteria**: Ease of onboarding, time-to-value, quality of RCA output, Slack integration, price vs Datadog
- **Location**: North America, Western Europe, Singapore/Australia
- **Trigger event**: Failed SOC 2 audit prep, major customer-facing outage, scaling pains post-Series B

**Why Heimdall wins with them**: Clear ROI story (replace $X Datadog spend + save Y engineer-hours/month), better findings reports than manual post-mortems, fits their GitHub-native workflows.

### ICP Anti-Patterns to Avoid Early

- **Large enterprises with security reviews**: Sales cycle 6–18 months, requires SOC 2, SSO, custom contracts — too slow for early stage
- **Infrastructure-heavy teams (bare metal, on-prem)**: Integration complexity; they need on-prem deployment options Heimdall doesn't have yet
- **Pure DevSecOps/SIEM buyers**: They want Splunk-style compliance tooling, not AI RCA

***

## Go-To-Market: Getting to 10 and 100 Teams

### Phase 1: First 10 Teams (0–3 months)

**Channel**: Founder-led outreach + "Show HN" + dev Twitter. Target founding engineers who have publicly complained about Datadog pricing or shared incident post-mortems.

**Entry wedge**: Free tier with full AI investigation on up to 1 project. No credit card. GitHub app install + Vercel integration in under 5 minutes. When Heimdall catches its first real incident and generates a findings report, that moment is the activation event.[48][49]

**Key metric**: Time to first "Heimdall saved me" moment. Target: under 2 weeks from signup.

**Community**: Dev Twitter, Hacker News, Singapore tech community (SGInnovate, Antler network), relevant Discord servers (Supabase, Vercel, Prisma).

### Phase 2: First 100 Teams (3–12 months)

**Channel**: PLG expansion from founding engineer to their team; outbound to Series A companies in Singapore, US, Germany (Ebb & Flow's geographic footprint).

**Expansion trigger**: When a solo user shares a findings report with their team or hands it to a coding agent for a fix, the product sells itself horizontally. Add team collaboration features and Slack notifications to create virality within organizations.[50][51]

**Integration partnerships**: Vercel, Supabase, Railway, Render as distribution channels. A Supabase integration listing gets Heimdall in front of tens of thousands of Supabase users who already have the exact stack Heimdall targets.

**Pricing unlock**: Introduce a Team plan ($99–$299/month) with multi-project support, Slack alerts, and custom report formats as the upgrade path from a free individual plan.

***

## Product Specification: The "10x Better, 10x Cheaper" Blueprint

To achieve undisputed market leadership in the SMB/startup segment, Heimdall needs to deliver across six capability pillars:

### 1. Integrations (Breadth = Moat)

**Immediate priorities (Day 1)**:
- Vercel (log drain), AWS CloudWatch, Railway, Render, Fly.io
- PostgreSQL (read-only role), MySQL, MongoDB, Redis
- GitHub App (repo access), GitLab

**Near-term (3–6 months)**:
- Supabase (native integration via pg_net), PlanetScale, Neon
- GCP Stackdriver, Azure Monitor
- Docker/container log ingestion
- Sentry (error events), PagerDuty (incident events)

**Medium-term (6–18 months)**:
- Full OpenTelemetry (OTel) receiver — this enables any OTel-instrumented stack to send data to Heimdall[6]
- Kubernetes pod logs, Helm chart for self-hosted deployment
- Linear/Jira for auto-creating incident tickets from findings

### 2. AI Triage Engine (Signal/Noise Separation)

The upstream filter before the investigation agent. Incumbent tools flood users with alerts — Dynatrace's Davis differentiates specifically by collapsing thousands of symptoms to one root-cause alert. Heimdall needs similar intelligence.[21]

**Requirements**:
- Pattern recognition across log streams: ignore flapping deploys, known-benign errors, health check noise
- Anomaly detection: detect deviations from baseline behavior (latency spikes, error rate changes, unusual DB query patterns)
- Cross-source correlation: connect a spike in Vercel function errors with a PostgreSQL slow query log event with a recent GitHub commit
- Configurable sensitivity per environment (production vs staging)

**Benchmark target**: 85–95% alert noise reduction from raw logs (IncidentFox claims this); Cleric early adopters report 20–30% engineering capacity recovered[44][42]

### 3. Investigation Agent (Core Differentiator)

This is Heimdall's most defensible capability. The investigation chain that no other SMB-priced tool offers:

```
Triggered Alert
    → Pull relevant log window (5 min before/after)
    → Query DB for slow queries, locks, anomalous writes around that timestamp
    → Search GitHub commit history for relevant recent changes
    → Correlate all signals
    → Generate structured findings report
```

**Findings report format** must include:
- **TL;DR**: 2-sentence plain English summary
- **Timeline**: chronological chain of events
- **Root Cause Hypothesis**: ranked list with confidence scores
- **Evidence**: specific log lines, DB queries, commit SHAs with excerpts
- **Suggested Fix**: concrete recommendations
- **Machine-readable section**: structured JSON for AI coding agents (Cursor, Windsurf) to consume via MCP

Traversal processes 300M logs per incident with causal ML. Heimdall doesn't need to match that scale on day one, but the *quality* of the reasoning and the *depth* of the findings (especially the code + DB correlation) must be demonstrably superior to anything in the sub-$500/month tier.[38]

### 4. User Experience ("Vercel of Logging")

The "Vercel of Logging" framing is apt and should drive every UX decision:[49][52]

| Vercel Principle | Heimdall Application |
|-----------------|---------------------|
| Git-push to deploy | Connect repo + log source in one OAuth flow |
| Tab status reflects deployment state[49] | Dashboard favicon/tab title shows active incident count |
| Optimistic UI eliminates perceived latency[49] | Findings start streaming in real-time, not batch |
| Empty states are instructions, not illustrations[49] | Empty state shows exact integration steps with copy-pasteable commands |
| CLI-first, dashboard optional[49] | Heimdall CLI for local log forwarding; API for programmatic access |
| Dark-mode first[49] | Dark mode default |
| Information-dense without being cluttered | Single-pane incident view: timeline + evidence + report |

**Onboarding target**: Time-to-first-integration under 5 minutes (IncidentFox's benchmark: Docker in 5 min, production K8s in 30 min). Vercel-style auto-detection of framework and log format.[44]

### 5. Pricing Architecture

The pricing anti-patterns of incumbents are well-documented and hated:[53][12][11]
- **Avoid**: Per-host, per-GB, per-user, per-custom-metric, high-watermark billing
- **Avoid**: Charging separately for ingestion *and* indexing (Datadog's double-dip)[11]
- **Target**: Per-project flat rate with generous included data volume

**Proposed pricing structure**:

| Plan | Price | What's Included |
|------|-------|-----------------|
| **Free** | $0 | 1 project, 14-day history, 5 AI investigations/month |
| **Starter** | $29/month | 3 projects, 90-day history, unlimited AI investigations |
| **Team** | $99/month | 10 projects, 12-month history, Slack + PagerDuty, findings API |
| **Growth** | $299/month | Unlimited projects, custom retention, SSO, priority support |
| **Enterprise** | Custom | On-prem, SOC 2, dedicated SLAs, custom integrations |

**Key pricing principles**:
- No per-seat pricing until Enterprise tier — teams should be able to add unlimited viewers
- AI investigation must be included, not an add-on (unlike Datadog's Bits AI surcharge)
- Data volume limits should be generous enough that 90% of startups never hit them
- Better Stack benchmarks show $24/month can undercut Datadog by 30x; Heimdall at $29 with AI investigation included is a compelling alternative[31]

### 6. MCP / Agent Integration (Forward-Looking Moat)

The most defensible long-term capability. Observe is already shipping an MCP server to let coding tools access observability data. Heimdall should ship this on day one.[54]

**Heimdall MCP server should expose**:
- `get_recent_incidents()` — fetch last N incidents with findings
- `get_findings_report(incident_id)` — full structured report
- `query_logs(service, time_range, filter)` — raw log search
- `get_db_anomalies(time_range)` — flagged DB events

This makes Heimdall the natural data source for AI coding agents (Cursor, Windsurf, Claude Code, etc.) when they need to understand what went wrong in production. No competitor at Heimdall's price point offers this.

***

## Competitive Differentiation Matrix

| Capability | Datadog | Dynatrace | SigNoz | Better Stack | Resolve.ai | **Heimdall** |
|-----------|---------|-----------|--------|--------------|------------|**-------**|
| Log ingestion | ✅ | ✅ | ✅ | ✅ | ❌ (needs source) | ✅ |
| DB integration (RO) | ❌ | ❌ | ❌ | ❌ | Partial | ✅ **Unique** |
| Codebase/GitHub integration | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| Autonomous 24/7 investigation | Bits AI (addon) | Davis AI | ❌ | Basic | ✅ | ✅ |
| Findings reports for AI agents | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ **Unique** |
| MCP server | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ **Unique** |
| SMB pricing (<$100/month) | ❌ | ❌ | ✅ ($49) | ✅ ($24) | ❌ | ✅ |
| 5-min onboarding | ❌ | ❌ | Partial | ✅ | ❌ | Target ✅ |
| No per-host/seat pricing | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ |

***

## Risks and Mitigations

**1. Enterprise AI SRE tools move downstream**
Resolve.ai ($1B valuation), Traversal, and incident.io could release SMB tiers. *Mitigation*: Heimdall's tightest moat is the code+DB integration forming a complete investigation chain. Incumbents can't add this without fundamental platform changes. Move fast to lock in integrations and PLG distribution.

**2. Datadog Bits AI improves rapidly**
Datadog is investing heavily in AI. *Mitigation*: Datadog's pricing complexity is structural — they cannot simplify it without destroying revenue. Their AI features will remain expensive add-ons. Heimdall's price advantage compounds over time.[15]

**3. OpenTelemetry commoditizes log collection**
OTel becoming the standard means collection becomes free. *Mitigation*: This is actually positive for Heimdall — it means native OTel support gives instant compatibility with all OTel-instrumented stacks. The value shifts from data collection (commodity) to investigation intelligence (Heimdall's core).[6]

**4. LLM costs eroding margins**
AI investigation at scale gets expensive if not optimized. *Mitigation*: Intelligent pre-filtering (Tier 1 triage before investigation agent), caching of findings patterns, use of distilled/smaller models for triage and larger models only for full investigations. Coralogix charges $1.50/million tokens separately; Heimdall should bake this into plan economics.[27]

**5. IncidentFox and other YC startups**
YC W26 is already funding IncidentFox with a similar thesis. *Mitigation*: Heimdall's DB + codebase integration creates a more complete product. Distribution via Vercel/Supabase integration marketplaces provides reach that a raw OSS project lacks. Ship faster.[44]

***

## Key Trends Heimdall Should Ride

1. **Observability as a delivery requirement** — in 2026, microservices, distributed workloads, and AI-related services make observability mandatory for any serious engineering team[55]
2. **AI coding agents need production context** — Cursor/Windsurf users need to understand what's failing in production; Heimdall's MCP integration plugs directly into this workflow
3. **Pricing backlash against Datadog** — forums, Hacker News, and developer Twitter are full of "$50K/year Datadog bill" horror stories; this is a genuine market pull moment
4. **OpenTelemetry standardization** — 97% of global companies now operate connected cloud estates; OTel as the common schema means Heimdall's integrations get easier over time[4]
5. **APAC cloud-native growth** — APAC observability tools market growing at 13–20% CAGR; Singapore as a hub gives Heimdall early access to a high-growth, underserved region[9][8]

***

## Conclusion

Heimdall's product thesis — AI-native, 24/7, investigates logs + database + code together, produces findings reports for humans and AI agents — has no direct equivalent below the $100K/year enterprise tier. The "Vercel of Logging" brand position is accurate and actionable: clear, developer-first, opinionated defaults, beautiful UX, and pricing that doesn't require a finance approval. 

The first 10 teams come from founding engineers in the Vercel/Supabase/Railway ecosystem who are already Heimdall's users by proxy — they're just using inferior alternatives or doing it manually. The first 100 come from Series A–B SaaS companies in fintech, AI, and e-commerce who are paying Datadog for less than they'll get from Heimdall. The 10x better / 10x cheaper claim is achievable: code + database investigation is genuinely unique, and $29/month vs $15,000–$50,000/year is not an incremental price advantage, it's a category-level disruption.