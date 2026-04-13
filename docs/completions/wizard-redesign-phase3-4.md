# Wizard Redesign — Phases 3 & 4: Grid Layout + Card Update

**Status:** Complete  
**Plan:** `docs/executing/connection-wizard-redesign.md`

---

## Phase 3: PlatformGrid Rewrite

### What changed

`PlatformGrid.vue` was rewritten from a flat 3-category list into a structured 3-section layout with distinct visual hierarchy and purpose-driven headers.

### Old layout

```
LOG SOURCES
  [Card] [Card] [Card] ...
DATABASES
  [Card] [Card]
GENERIC
  [Card]
```

Flat category grouping. No indication of what happens when you add a connection.

### New layout

```
WHERE DO YOUR LOGS COME FROM?
Pick your platform — we'll handle the wiring.
  [Supabase] [Fly.io] [Vercel] [Render]
  [Railway]  [Heroku] [AWS]    [DigitalOcean]

──────────── or ────────────

CONNECT DIRECTLY
Already know the protocol? Skip the platform.
  [Webhook HTTP]  [Syslog TCP/TLS]  [OpenTelemetry OTLP]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 AGENT INVESTIGATION TOOLS
These don't send logs — they let the agent query
your systems during investigations and chat.
  [PostgreSQL]  [GitHub]  [MySQL]
```

### Visual hierarchy

Three distinct separation levels between sections:

1. **Sections 1 → 2**: Light "or" divider — horizontal line with centred "or" text. Both sections are about log ingestion, so the visual break is soft. The "or" communicates these are alternative paths to the same outcome.

2. **Sections 2 → 3**: Strong `border-t` divider + search icon in the header. This is a fundamentally different category — not log sources, but investigation tools. The visual break is intentional and the subtitle "These don't send logs" prevents the most common onboarding confusion.

### Grid density per section

| Section | Columns (mobile) | Columns (sm) | Columns (md) | Why |
|---------|-------------------|--------------|--------------|-----|
| Platforms | 2 | 3 | 4 | Many items, compact cards |
| Protocols | 1 | 3 | 3 | Exactly 3, even spacing |
| Tools | 1 | 3 | 3 | Exactly 3, even spacing |

---

## Phase 4: PlatformCard Update

### What changed

`PlatformCard.vue` was updated to use `ConnectorLogo` (native brand SVGs) instead of 2-letter abbreviation badges, and to display the `subtitle` field instead of the longer `description`.

### Old card

```
┌──────────────────┐
│ [SB]  Supabase   │  ← 2-letter badge + name
│       Database    │  ← Full description
│       logs...    │
└──────────────────┘
```

Horizontal layout, left-aligned, description takes up space.

### New card

```
┌──────────┐
│   [⚡]    │  ← Native brand logo (ConnectorLogo)
│ Supabase │  ← Name, centred
│Log poll..│  ← Subtitle (short), centred
└──────────┘
```

Vertical layout, centre-aligned, compact. Better fit for the denser 4-column grid.

### Key changes

- **Logo**: `<ConnectorLogo :type="flow.id" :size="26" />` replaces the 2-letter `<div>` badge
- **Subtitle**: `flow.subtitle` replaces `flow.description` — shorter, more informative at grid scale
- **Layout**: `flex-col items-center text-center` instead of horizontal `flex items-start`
- **Sizing**: Reduced padding (`px-3 py-4`) for denser grids
- **"Soon" badge**: Smaller (`text-[8px]`, `top-1.5 right-1.5`) to fit compact card

---

## Files changed

| File | Phase | Change |
|------|-------|--------|
| `frontend/src/components/connections/wizard/PlatformGrid.vue` | 3 | Rewritten to 3-section layout with headers, subtitles, dividers |
| `frontend/src/components/connections/wizard/PlatformCard.vue` | 4 | ConnectorLogo + subtitle, vertical centre layout |

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| `vitest run` | 52/52 tests pass |
