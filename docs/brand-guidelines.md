# Heimdall — Brand Guidelines

## Identity

Heimdall is an autonomous AI monitoring agent for production systems. The brand evokes **surveillance**, **precision**, and **quiet authority** — an all-seeing eye that watches without blinking. The aesthetic is **techno-brutalist**: stark, monochromatic, typographic, with a single desaturated accent.

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

**Feldgrau** — a desaturated, military grey-green inspired by German field uniforms. Saturation sits around 12%, producing a steely tone that avoids the organic feel of brighter greens. Everything lives in a narrow band between near-black and muted grey-green.

### Palette

#### Backgrounds

| Token | Hex | Use |
|-------|-----|-----|
| `--bg-primary` | `#070808` | Page background, default canvas |
| `--bg-surface` | `rgba(11, 13, 12, 0.5)` | Cards, panels (translucent) |
| `--bg-surface-hover` | `rgba(11, 13, 12, 0.7)` | Hovered cards/panels |
| `--bg-elevated` | `#0a0c0b` | Modals, dropdowns, raised surfaces |

#### Text

| Token | Hex | Use |
|-------|-----|-----|
| `--text-primary` | `#e6eae8` | Headings, body text |
| `--text-secondary` | `#8a938e` | Supporting text, labels |
| `--text-muted` | `#576058` | Captions, disabled text |

#### Accent (Feldgrau)

| Token | Value | Use |
|-------|-------|-----|
| `--accent` | `#4d5d53` | Primary accent — buttons, links, active states |
| `--accent-hover` | `#5a6e62` | Hover state for accent elements |
| `--accent-bright` | `#6e8578` | Emphasis, highlights, high-contrast accent |
| `--accent-subtle` | `rgba(77, 93, 83, 0.1)` | Tinted backgrounds, selected rows |
| `--accent-border` | `rgba(77, 93, 83, 0.3)` | Accent-tinted borders |

#### Borders

| Token | Value | Use |
|-------|-------|-----|
| `--border` | `rgba(77, 93, 83, 0.06)` | Structural dividers, card edges |
| `--border-hover` | `rgba(77, 93, 83, 0.14)` | Hovered borders |

#### Status

| Token | Hex | Use |
|-------|-----|-----|
| `--status-ok` | `#4d5d53` | Healthy, connected, passing |
| `--status-warn` | `#c4a84a` | Warnings, degraded |
| `--status-critical` | `#c45a4a` | Errors, outages, failures |
| `--status-info` | `#4a8aae` | Informational, neutral alerts |

### External Contexts

When the full token set isn't available (Stripe checkout, email templates, social cards):

| Context | Background | Foreground | Accent |
|---------|-----------|------------|--------|
| Dark background | `#000000` or `#070808` | `#ffffff` or `#e6eae8` | `#6e8578` (bright) |
| Light background | `#ffffff` | `#070808` | `#4d5d53` (base) |

Use `--accent-bright` (`#6e8578`) against pure black for better contrast on interactive elements.

---

## Typography

### Typefaces

| Role | Family | Fallback Stack |
|------|--------|----------------|
| **Monospace** (primary) | JetBrains Mono | Fira Code, SF Mono, monospace |
| **Sans-serif** (body) | Inter | system-ui, -apple-system, sans-serif |

### Hierarchy

- **Wordmark / display:** JetBrains Mono, semibold, uppercase, wide tracking (`0.3em`)
- **Navigation:** JetBrains Mono, 12.5px, uppercase, `tracking-widest`
- **Headings:** Inter or JetBrains Mono, depending on context
- **Body text:** Inter, regular weight, `--text-primary`
- **Labels / captions:** Inter, `--text-secondary` or `--text-muted`

### Conventions

- Navigation and brand elements are **always uppercase**.
- Body copy is sentence case.
- Monospace is used for anything that should feel "terminal" or "system-level" — nav, brand, data labels, status badges.

---

## Voice & Tone

- **Blunt, not friendly.** Statements, not invitations. "Autonomous system surveillance." not "We help you monitor your apps!"
- **Technical, not corporate.** Assume the reader is an engineer. No buzzwords, no fluff.
- **Sparse.** Say it in fewer words. If a heading can be two words, don't make it five.
- **Lowercase in prose.** Avoid Title Case outside of proper nouns and the brand name.

---

## Design Principles

| Principle | Meaning |
|-----------|---------|
| **Dark-first** | Everything starts on near-black. Light backgrounds are exceptions, not defaults. |
| **Monochrome + one accent** | The entire UI lives in greyscale with feldgrau as the sole colour. No secondary colours outside status indicators. |
| **Typography over ornament** | Spaced uppercase, monospace type, and structural whitespace do the visual work. No icons where a label suffices. |
| **Brutalist restraint** | No rounded corners on primary surfaces. No gradients. No shadows except for elevation. Borders are hair-thin and near-invisible. |
| **Density over padding** | Information-dense layouts. Compact spacing. The interface should feel like a command terminal, not a marketing page. |
