We need a public website that briefly, simple outlines and describes what Heimdall is / what it does.

For sitemap archiecture/structure and broad outline inspiration, look at Heimdall's "sister products":
- elephantasm.com
- trajancloud.com
- corpovault.com

Outline your proposed public website design/structure/implementation in this md file.

---

# Proposed Website Design

## Shared Pattern (from sister sites)

All three sister products follow a consistent formula:
- Dark theme, minimalist, developer-focused
- Simple top nav: 3–4 links + Login/CTA + GitHub icon
- Hero section: bold one-liner, short paragraph, CTA buttons
- Feature section: 3-column grid of capabilities
- Lightweight footer: © Heimdall, Terms, Privacy, socials, "A Kamino Corporation platform"

Heimdall follows the same skeleton with two key differences: the homepage is a **single-viewport landing page** (no scroll), and detailed features live on a dedicated **/features** page.

---

## Sitemap

```
/                — Landing page (single viewport, no scroll)
/features        — How it works + core feature showcase
/pricing         — Pricing tiers
/login           — Redirect to app login
```

Top nav: **Features** (`/features`) · **Pricing** · **GitHub** · **Start Monitoring** (primary CTA)

---

## Landing Page `/` (single viewport, no scroll)

The entire landing page fits in one screen. Clean, focused, immediate.

**Headline:**

> Watches everything.
> Investigates automatically.
> Reports what matters.

**Subtext:** Heimdall monitors your production systems 24/7, autonomously investigates anomalies, and delivers actionable incident reports — so you know exactly what went wrong and how to fix it.

**CTAs:** `Start Monitoring` (primary) · `See How It Works` (secondary/outline → links to `/features`)

**Sub-note:** "Set up in under 5 minutes. No credit card required."

**Footer strip** (minimal, inline at bottom of viewport):
© 2026 Crimson Sun · Terms · Privacy · GitHub · X

---

## Features Page `/features`

### Section 1 — The Problem

Short 3-column grid showing the pain points Heimdall solves:

| Column | Heading | Copy |
|--------|---------|------|
| 1 | Alerts are dumb | Pattern-match alerts fire on every spike with no understanding of context. Real incidents drown in noise. |
| 2 | Investigation is manual | When something looks wrong, a developer has to context-switch, cross-reference dashboards, and piece together a picture from scratch. |
| 3 | Knowledge is lost | The dev who debugged last week's incident carries the context in their head. At 3am, the on-call engineer starts from zero. |

### Section 2 — How It Works

Visual 3-step sequence:

1. **Connect your infrastructure** — Link your databases, log sources, and codebases. Under 5 minutes.
2. **Heimdall watches** — The agent monitors continuously, learning what normal looks like for your system.
3. **Get actionable reports** — When anomalies appear, Heimdall investigates autonomously and surfaces structured diagnoses.

### Section 3 — Core Features

3-column card grid (2 rows):

| Feature | Heading | Copy |
|---------|---------|------|
| Autonomous monitoring | Always watching | Continuous log and database monitoring. Knows what normal looks like. Recognises when something deviates. |
| Autonomous investigation | Investigates, not just alerts | When anomalies appear, the agent queries your database, searches logs, and builds a diagnosis — before paging you. |
| Institutional memory | Remembers every incident | Long-term memory accumulates past incidents, patterns, and failure modes. Each investigation makes the next one better. |
| On-demand chat | Ask anything, anytime | "What happened with payments last night?" Chat with the agent to get instant answers backed by real data. |
| Structured reports | Incident reports, not log dumps | Timeline, investigation steps, historical context, and diagnostic assessment — all in one structured document. |
| Easy integrations | Connect in minutes | PostgreSQL, webhooks, syslog, GitHub. More connectors coming. No complex setup. |

### Section 4 — CTA Banner

"Stop babysitting your logs."

`Start Monitoring` button.

"14-day free trial · Unlimited connections · Cancel anytime"

### Footer

- © 2026 Crimson Sun
- Links: Terms · Privacy · Changelog
- Socials: GitHub · X (Twitter)

---

## Design Direction

| Aspect | Choice |
|--------|--------|
| **Theme** | Heimdall's existing techno-brutalist dark theme |
| **Background** | Near-black with green undertone (`#060806` primary, `#0a0e0a` elevated) |
| **Accent colour** | Forest sage green `#5a9e6a`, hover `#4a8c5a`, bright `#7ab889` |
| **Borders** | Green-tinted at low opacity (`rgba(90, 158, 106, 0.06)`) |
| **Text** | Warm whites with green undertone (`#e8ede8` primary, `#8a9a8a` secondary, `#5a6b5a` muted) |
| **Typography** | Inter (body) + JetBrains Mono (code/technical elements) |
| **Tone** | Developer-focused, direct, no-nonsense. Confident but not salesy. |
| **Illustrations** | Minimal. Prefer subtle terminal/code visuals or animated data-flow diagrams over stock imagery. |

---

## Implementation

| Aspect | Choice |
|--------|--------|
| **Framework** | Next.js (consistent with Elephantasm) or Astro (lighter, static-first) |
| **Hosting** | Vercel or Fly.io (already used for backend) |
| **Styling** | Tailwind CSS |
| **Domain** | TBD |

---

## Open Questions

1. **Domain** — What domain will Heimdall's public site live on? (e.g. `heimdall.dev`, `getheimdall.com`, `heimdall.crimsonsun.dev`?)
2. **Framework preference** — Next.js (consistency with Elephantasm) vs Astro (lighter, faster for a mostly-static marketing site)?
3. **Pricing page** — Should we include a pricing page now or hold off until pricing is finalised? Could show a "Coming soon" or "Contact us" placeholder.
4. **Demo video** — Trajan has a demo video placeholder. Do we want one for Heimdall on the features page?
5. **Parent branding** — Trajan and CorpoVault credit "Kamino Corp" in the footer. Heimdall should credit "Crimson Sun" — correct?
