# Public Website — v1 Implementation

## Summary

Implemented the public marketing website as Vue routes inside the existing `frontend/` app (not a separate Astro project). Three public pages — landing (`/`), features (`/features`), pricing (`/pricing`) — served without authentication alongside the existing app. Dashboard moved from `/` to `/dashboard`. Uses the Concept B "Bifrost Hexagram" logo in the public nav.

## Decisions Made

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Framework** | Vue routes in existing `frontend/` (not separate Astro project) | Single Vercel deployment. No second app to manage. Marketing pages are static content — Vue handles this fine. |
| **Routing** | `meta: { public: true }` on marketing routes | Auth guard updated to skip public routes. Clean, extensible pattern. |
| **Dashboard path** | Moved from `/` to `/dashboard` | Frees `/` for the public landing page. Sidebar, login redirect, and auth guard all updated. |
| **Logo** | Concept B — Bifrost Hexagram (inline SVG) | Scaled down and embedded directly in PublicNav. No external asset loading. |
| **Pricing** | Included with placeholder tiers | Starter $0 / Pro $49 / Enterprise Custom. Adjustable when finalised. |
| **Design tokens** | Reused existing `main.css` tokens | No new CSS needed — the app's techno-brutalist theme already defines all the colours, fonts, and utilities. |

## Pages

### `/` — Landing (single viewport)
- Full-screen hero with `h-screen` flex layout, no scroll
- Three-line headline with accent-coloured middle line ("Investigates automatically.")
- Two CTAs: "Start Monitoring" (filled) + "See How It Works" (outline → `/features`)
- Inline footer pinned to bottom

### `/features` — Features
- **Section 1 — The Problem**: 3-column grid (alerts are dumb, investigation is manual, knowledge is lost)
- **Section 2 — How It Works**: 3-step numbered sequence (connect, watch, report)
- **Section 3 — Core Features**: 2×3 card grid with hover effects
- **Section 4 — CTA Banner**: "Stop babysitting your logs" + CTA

### `/pricing` — Pricing
- 3-column pricing cards (Starter / Pro / Enterprise)
- Pro tier highlighted with accent border + "Popular" badge
- Feature checklists with green check icons
- Enterprise links to `mailto:hello@crimsonsun.dev`

### `/dashboard` — Dashboard (moved from `/`)
- Existing dashboard, now at `/dashboard`
- Sidebar link updated

## Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/router/index.ts` | Added 3 public routes (`/`, `/features`, `/pricing`), moved dashboard to `/dashboard`, updated auth guard to respect `meta.public` |
| 2 | `frontend/src/layouts/DefaultLayout.vue` | Layout bypass extended from login-only to `route.meta?.public` |
| 3 | `frontend/src/pages/LoginPage.vue` | Post-login redirect changed from `/` to `/dashboard` |
| 4 | `frontend/src/components/common/AppSidebar.vue` | Dashboard link updated from `/` to `/dashboard` |

## Files Created

| # | File | Purpose |
|---|------|---------|
| 1 | `frontend/src/components/public/PublicNav.vue` | Marketing nav — Bifrost Hexagram logo, desktop links, mobile hamburger, active route highlighting |
| 2 | `frontend/src/components/public/PublicFooter.vue` | Footer with `inline` (landing) and `full` (features/pricing) variants |
| 3 | `frontend/src/pages/public/LandingPage.vue` | Single-viewport landing page |
| 4 | `frontend/src/pages/public/FeaturesPage.vue` | Features page with 4 sections |
| 5 | `frontend/src/pages/public/PricingPage.vue` | Pricing page with 3 tiers |

## Files Deleted

| # | File | Reason |
|---|------|--------|
| 1 | `website/` (entire directory) | Replaced by Vue routes in `frontend/`. One deploy instead of two. |

## Open Items

1. **GitHub link** — Nav links to `https://github.com` — needs real repo URL
2. **Pricing finalisation** — Tier names, prices, and feature lists are placeholders
3. **Terms & Privacy pages** — Links removed from footer for now (no pages exist yet)
