# Phase 8 — Frontend Redesign: Foundation

Phase 1 of the techno-brutalist frontend redesign. Installs self-hosted fonts, establishes the full design token system as CSS custom properties, registers tokens with Tailwind v4's `@theme` directive, and sets the dark base.

Spec: [`docs/executing/redesign-fe.md`](../executing/redesign-fe.md) → Phase 1

---

## What Changed

### Dependencies Added

| Package | Version | Purpose |
|---------|---------|---------|
| `@fontsource/jetbrains-mono` | ^5.x | Self-hosted JetBrains Mono (display/UI font) — weights 400, 500, 700 |
| `@fontsource/inter` | ^5.x | Self-hosted Inter (body/readable font) — weights 400, 500, 600 |

No external font CDN calls. Fonts ship with the bundle.

### `src/main.ts`

Added font CSS imports (6 total — 3 weights per font) before the main stylesheet import. This ensures font-face declarations are available when the CSS theme references them.

### `src/assets/styles/main.css`

Rebuilt from the single `@import "tailwindcss"` line into the full theme foundation:

**CSS Custom Properties (`:root`)** — 25 design tokens covering:
- Backgrounds: `--bg-primary` (`#080c08`), `--bg-surface`, `--bg-surface-hover`, `--bg-elevated`
- Text: `--text-primary` (`#e8ede8`), `--text-secondary`, `--text-muted`
- Accent: `--accent` (`#5a9e6a`), `--accent-hover`, `--accent-bright`, `--accent-subtle`, `--accent-border`
- Borders: `--border`, `--border-hover`
- Status: `--status-ok`, `--status-warn`, `--status-critical`, `--status-info`
- Typography: `--font-mono`, `--font-sans`

**Tailwind v4 `@theme` block** — registers all tokens as Tailwind colors and font families. This enables utility classes like `bg-bg-primary`, `text-accent`, `border-border`, `font-mono`, `font-sans` throughout all Vue templates.

**Base styles:**
- `html`: dark background, primary text color, Inter font family
- `body`: min-height viewport, antialiased font rendering
- Custom scrollbar: 6px wide, green-tinted thumb matching the accent palette
- Selection highlight: green-tinted background
- Focus ring: green accent outline for accessibility
- `prefers-reduced-motion`: disables all animations/transitions for users who opt out

### `index.html`

Added inline `style="background-color:#080c08"` to `<html>` tag. This prevents the flash of white background before CSS loads — the browser paints the dark surface immediately.

### `src/App.vue`

Updated the loading state (shown while auth initializes) to use the new token classes: `bg-bg-primary text-text-muted font-mono text-sm uppercase tracking-wider`. Text changed from "Loading…" to "Initializing…" to match the military-tech tone.

---

## Design Decisions

1. **CSS custom properties as the source of truth, `@theme` as the bridge.** Tokens defined once in `:root`, then referenced by Tailwind's `@theme`. This means tokens are usable both in Tailwind utilities (`bg-accent`) and raw CSS (`background: var(--accent)`).

2. **Only necessary font weights.** JetBrains Mono: 400 (body), 500 (medium emphasis), 700 (bold headings). Inter: 400 (body), 500 (medium), 600 (semibold). This keeps the font bundle lean.

3. **Scrollbar and selection styling.** Both use the accent green at low opacity to maintain the immersive green-black atmosphere without being distracting.

4. **Reduced motion support from day one.** The `prefers-reduced-motion` media query blanket-disables all animations — later phases can add effects knowing this safety net exists.

---

## Verification

- `vue-tsc -b --noEmit`: passes (no type errors)
- `vite build`: succeeds, font files included in dist output
- All 25 design tokens available as Tailwind utility classes
