# Heimdall — Brand Guidelines

## Identity

Heimdall is an autonomous AI monitoring agent for production systems. The brand evokes **surveillance**, **precision**, and **quiet authority** — an all-seeing eye that watches without blinking. The aesthetic is **technical retro-futurism**: mission-control interfaces, phosphor-green CRT readouts, and aerospace engineering visuals. Dark-first, typographic, with a single accent colour family (phosphor-feldgrau) and atmospheric textures (grid overlays, scan-lines, glow halos) that make the interface feel alive.

The name comes from Norse mythology — Heimdall, the watchman of the gods, who guards the Bifrost bridge and sees all.

---

## Logo

### The Bifrost Starburst

A six-pointed forked starburst with a centre eye dot. Six forked rays at 60° intervals (hexagonal / Bifrost symmetry), each split into two prongs with angled tips. A small solid dot at the centre represents the pupil — the all-seeing eye.

| File | Use |
|------|-----|
| `docs/logos/heimdall-logo.svg` | Master logo — `currentColor` fill, CSS-tintable |
| `docs/logos/heimdall-logo-preview.svg` | Preview on dark background with labels |
| `frontend/public/favicon.svg` | Favicon — white `#ffffff` on black `#000000` |

### Usage Rules

- **Minimum clear space:** Half the logo's width on all sides.
- **Minimum size:** 16 × 16px (favicon scale). The design is optimised for small rendering.
- **On dark backgrounds:** Use white `#ffffff` or off-white `#d0ddd0`.
- **On light backgrounds:** Use near-black `#070808` or pure black `#000000`.
- **Never** rotate, stretch, add drop shadows, apply gradients, or enclose in a shape.

### Wordmark

The brand name is rendered as spaced uppercase letters in JetBrains Mono:

```
H E I M D A L L
```

- Font: JetBrains Mono, semibold (600)
- Letter-spacing: `0.3em`
- Case: Uppercase
- The wordmark stands alone or appears beside the starburst — never overlaid on it.

---

## Colour

### Philosophy

**Phosphor-Feldgrau** — evolved from the original feldgrau (12% saturation) to a richer phosphor-green (~25% saturation) inspired by CRT readouts, mission-control displays, and radar screens. The military grey-green heritage remains, but the accent now carries enough chroma to register as a deliberate colour signal — like a powered-on instrument panel rather than a powered-off one. Backgrounds stay near-black with a cool blue-black undertone.

### Palette

#### Backgrounds

| Token | Hex | Use |
|-------|-----|-----|
| `--bg-primary` | `#06080a` | Page background, default canvas (cool blue-black) |
| `--bg-surface` | `rgba(12, 16, 14, 0.55)` | Cards, panels (translucent) |
| `--bg-surface-hover` | `rgba(14, 20, 17, 0.75)` | Hovered cards/panels |
| `--bg-elevated` | `#0c100e` | Modals, dropdowns, raised surfaces |

#### Text

| Token | Hex | Use |
|-------|-----|-----|
| `--text-primary` | `#e8ede9` | Headings, body text (cool phosphor white) |
| `--text-secondary` | `#8a9e92` | Supporting text, labels (green-cast secondary) |
| `--text-muted` | `#4e6556` | Captions, disabled text |

#### Accent (Phosphor-Feldgrau)

| Token | Value | Use |
|-------|-------|-----|
| `--accent` | `#4a7a5c` | Primary accent — buttons, links, active states |
| `--accent-hover` | `#5a9068` | Hover state for accent elements |
| `--accent-bright` | `#6aad7a` | Emphasis, highlights, high-contrast accent |
| `--accent-subtle` | `rgba(74, 122, 92, 0.12)` | Tinted backgrounds, selected rows |
| `--accent-border` | `rgba(74, 122, 92, 0.25)` | Accent-tinted borders |
| `--accent-glow` | `rgba(74, 122, 92, 0.15)` | Box-shadow / drop-shadow for active elements, phosphor halos |

#### Borders

| Token | Value | Use |
|-------|-------|-----|
| `--border` | `rgba(74, 122, 92, 0.18)` | Structural dividers, card edges (visible control-panel chrome) |
| `--border-hover` | `rgba(74, 122, 92, 0.32)` | Hovered borders |

#### Status

| Token | Hex | Use |
|-------|-----|-----|
| `--status-ok` | `#4a7a5c` | Healthy, connected, passing (aligned with accent) |
| `--status-warn` | `#d4a832` | Warnings, degraded (rich amber caution light) |
| `--status-critical` | `#d44a3a` | Errors, outages, failures (hot red) |
| `--status-info` | `#4a92c4` | Informational, neutral alerts (CRT-cyan) |

Each status colour has a matching glow token (`--status-*-glow`) at 20% opacity for box-shadow halos on active indicators.

### External Contexts

When the full token set isn't available (Stripe checkout, email templates, social cards):

| Context | Background | Foreground | Accent |
|---------|-----------|------------|--------|
| Dark background | `#000000` or `#06080a` | `#ffffff` or `#e8ede9` | `#6aad7a` (bright) |
| Light background | `#ffffff` | `#06080a` | `#4a7a5c` (base) |

Use `--accent-bright` (`#6aad7a`) against pure black for better contrast on interactive elements.

---

## Typography

### Typefaces

| Role | Family | Fallback Stack |
|------|--------|----------------|
| **Monospace** (primary) | JetBrains Mono | Fira Code, SF Mono, monospace |
| **Sans-serif** (body) | Inter | system-ui, -apple-system, sans-serif |

### Hierarchy

| Element | Font | Size | Style |
|---------|------|------|-------|
| Wordmark / display | JetBrains Mono | — | Semibold, uppercase, `tracking-wider` (`0.3em`) |
| Page headings | JetBrains Mono | `text-2xl` (24px) | Bold, uppercase, `tracking-wider` |
| Section headings | JetBrains Mono | `text-xs` (12px) | Medium, uppercase, `tracking-widest` |
| Navigation items | JetBrains Mono | `text-sm` (14px) | Regular |
| Body text | Inter | `text-sm` (14px) | Regular, `--text-primary` |
| Labels / captions | JetBrains Mono | `text-xs` (12px) | Medium, uppercase, `tracking-wider` |
| Helper / hint text | JetBrains Mono | `text-xs` (12px) | Regular, `--text-muted` |

### Type Scale

| Token | Size | Use |
|-------|------|-----|
| `text-xs` | 12px (0.75rem) | **Minimum size.** Labels, captions, badges, section headers, helper text |
| `text-sm` | 14px (0.875rem) | Body text, nav items, buttons, input values |
| `text-base` | 16px (1rem) | Large body, subsection headings |
| `text-lg` | 18px (1.125rem) | Section titles |
| `text-2xl` | 24px (1.5rem) | Page titles |
| `text-3xl`+ | 30px+ | Hero titles, public pages only |

**12px floor rule:** No text in the application may be set below `text-xs` (12px). This applies to all labels, badges, metadata, timestamps, and helper text. Sub-12px text (`text-[10px]`, `text-[11px]`) is explicitly prohibited.

### Conventions

- Navigation and brand elements are **always uppercase**.
- Body copy is sentence case.
- Monospace (JetBrains Mono) is used for anything that should feel "terminal" or "system-level" — nav, brand, data labels, status badges, form labels, buttons.
- All uppercase text must use `tracking-wider` or `tracking-widest` for legibility.

---

## Voice & Tone

- **Blunt, not friendly.** Statements, not invitations. "Autonomous system surveillance." not "We help you monitor your apps!"
- **Technical, not corporate.** Assume the reader is an engineer. No buzzwords, no fluff.
- **Sparse.** Say it in fewer words. If a heading can be two words, don't make it five.
- **Lowercase in prose.** Avoid Title Case outside of proper nouns and the brand name.

---

## Sizing & Spacing

### Base Unit

4px (Tailwind's `1` = 0.25rem = 4px). Most spacing values are multiples of 4 or 8.

### Component Sizing

| Element | Specification |
|---------|---------------|
| **Primary buttons** | `px-5 py-2.5`, `text-sm`, `font-semibold uppercase tracking-wider`, `rounded` |
| **Secondary buttons** | `px-5 py-2.5`, `text-sm`, `uppercase tracking-wider`, `border border-accent-border`, `rounded` |
| **Dialog/wizard buttons** | `px-5 py-2`, `text-xs`, `font-medium uppercase tracking-wider`, `rounded` |
| **Text inputs** | `px-3 py-2`, `text-sm`, `rounded`, `bg-bg-elevated/80`, `border border-border` |
| **Cards/panels** | `p-6`, `rounded-lg`, `border border-border`, `bg-bg-surface` |
| **Modals** | Header `px-6 py-4`, body `px-6 py-6`, footer `px-6 py-3` |
| **Modal max-width** | `max-w-xl` (576px) standard, `max-w-lg` (512px) compact |

### Spacing Scale

```
2    = 0.5rem    (8px)     Standard gap, inline element spacing
3    = 0.75rem   (12px)    Component internal spacing
4    = 1rem      (16px)    Standard padding, form field gaps
5    = 1.25rem   (20px)    Comfortable padding
6    = 1.5rem    (24px)    Card padding, form section gaps, modal body
8    = 2rem      (32px)    Page horizontal padding (mobile)
10   = 2.5rem    (40px)    Dashboard section spacing
12   = 3rem      (48px)    Page vertical padding (desktop)
```

### Layout

| Element | Value |
|---------|-------|
| **Sidebar width** | `w-60` (240px) |
| **Sidebar brand header** | `px-6 py-6` |
| **Sidebar nav items** | `px-3 py-2`, `gap-2.5` |
| **Page content padding** | `px-6 py-8` (mobile), `px-8 py-12` (desktop `lg:`) |
| **Page header** | `pb-6 mb-8 border-b border-border` |
| **Section gaps** | `space-y-6` between form fields, `space-y-8` between major sections |
| **Card grid gaps** | `gap-5` |

### Input Focus States

All text inputs, selects, and textareas use the same focus treatment:

```
focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none
```

The border shifts to full accent colour on focus, with a 30% opacity ring for reinforcement. The global `:focus-visible` outline is `2px solid var(--accent)` with `2px` offset and an `8px` phosphor glow halo (`box-shadow: 0 0 8px var(--accent-glow)`).

---

## Design Principles

| Principle | Meaning |
|-----------|---------|
| **Dark-first** | Everything starts on near-black. Light backgrounds are exceptions, not defaults. |
| **Monochrome + one accent** | The entire UI lives in greyscale with phosphor-feldgrau as the sole colour family. No secondary colours outside status indicators. |
| **Typography over ornament** | Spaced uppercase, monospace type, and structural whitespace do the visual work. No icons where a label suffices. |
| **Information emits light** | Active data and status indicators glow (phosphor halos, glow-pulse animations). Structural chrome stays muted. The interface reads like a powered-on control panel — instruments bright, housing dark. |
| **Atmospheric depth** | Subtle textures (grid overlays, scan-lines, corner brackets) create mission-control atmosphere through accumulation, not spectacle. Each effect is individually barely perceptible; together they create immersion. |
| **Confident spacing** | Components have enough internal padding and external margin to feel substantial. The interface should feel like a well-built instrument panel — precise and weighty, not cramped or fragile. Minimum font size is 12px; minimum card padding is 24px. |
