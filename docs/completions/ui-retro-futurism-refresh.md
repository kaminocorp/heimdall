# UI Refresh — Technical Retro-Futurism

## Status: Complete

## Summary

Evolved Heimdall's visual identity from muted techno-brutalist to **technical retro-futurism** — higher contrast, richer colour signatures, phosphor-green glow effects, and atmospheric textures evoking mission-control interfaces. No layout, font, or structural changes. The same dark-first, monospace, miltech bones with substantially more visual authority.

---

## Phase 1 — Design Token Update

**File:** `frontend/src/assets/styles/main.css` (`:root` block)

### Backgrounds — deeper tier separation
| Token | Before | After | Why |
|-------|--------|-------|-----|
| `--bg-primary` | `#070808` | `#06080a` | Cool blue-black shift (CRT housing tone) |
| `--bg-surface` | `rgba(11,13,12,0.5)` | `rgba(12,16,14,0.55)` | More visible card separation |
| `--bg-surface-hover` | `rgba(11,13,12,0.7)` | `rgba(14,20,17,0.75)` | Clearer hover feedback |
| `--bg-elevated` | `#0a0c0b` | `#0c100e` | Distinct elevation layer |

### Accent — feldgrau → phosphor-feldgrau (~12% → ~25% saturation)
| Token | Before | After |
|-------|--------|-------|
| `--accent` | `#4d5d53` | `#4a7a5c` |
| `--accent-hover` | `#5a6e62` | `#5a9068` |
| `--accent-bright` | `#6e8578` | `#6aad7a` |
| `--accent-subtle` | `rgba(77,93,83,0.1)` | `rgba(74,122,92,0.12)` |
| `--accent-border` | `rgba(77,93,83,0.3)` | `rgba(74,122,92,0.25)` |
| `--accent-glow` | *(new)* | `rgba(74,122,92,0.15)` |

### Action — bolder CTA
| Token | Before | After |
|-------|--------|-------|
| `--action` | `#3b8a5a` | `#38a85c` |
| `--action-hover` | `#449e66` | `#42be68` |

### Text — crisper contrast, phosphor undertone
| Token | Before | After |
|-------|--------|-------|
| `--text-primary` | `#e6eae8` | `#e8ede9` |
| `--text-secondary` | `#8a938e` | `#8a9e92` |
| `--text-muted` | `#576058` | `#4e6556` |

### Borders — visible control-panel chrome (14% → 18%)
| Token | Before | After |
|-------|--------|-------|
| `--border` | `rgba(77,93,83,0.14)` | `rgba(74,122,92,0.18)` |
| `--border-hover` | `rgba(77,93,83,0.25)` | `rgba(74,122,92,0.32)` |

### Status — richer mission-control warning lights
| Token | Before | After |
|-------|--------|-------|
| `--status-ok` | `#4d5d53` | `#4a7a5c` |
| `--status-warn` | `#c4a84a` | `#d4a832` |
| `--status-critical` | `#c45a4a` | `#d44a3a` |
| `--status-info` | `#4a8aae` | `#4a92c4` |

### New tokens
- `--accent-glow` — phosphor halo for box-shadows
- `--status-ok-glow`, `--status-warn-glow`, `--status-critical-glow`, `--status-info-glow` — status-specific glow halos
- All registered in `@theme` block for Tailwind utility access

---

## Phase 2 — Glow Effects & Animation Enhancements

**File:** `frontend/src/assets/styles/main.css`

### New CSS classes
| Class | Effect |
|-------|--------|
| `.glow-pulse` | 3-second breathing glow cycle using `--accent-glow` (replaces `animate-pulse` on status dots) |
| `.glow-pulse-warn` | Same cycle, amber (`--status-warn-glow`) |
| `.glow-pulse-critical` | Same cycle, red (`--status-critical-glow`) |
| `.brand-glow` | `text-shadow: 0 0 20px var(--accent-glow)` — CRT character emission |

### Upgraded effects
| Element | Before | After |
|---------|--------|-------|
| `.glow-active` | `box-shadow: 0 0 15px rgba(77,93,83,0.06)` | `box-shadow: 0 0 20px var(--accent-glow), 0 0 4px var(--accent-glow)` |
| `@keyframes fadeIn` | Simple opacity+translateY | Added `border-color: var(--accent) → var(--border)` transition (cards flash accent on entry, then settle) |
| `@keyframes reveal` | Simple clip-path wipe | Added `text-shadow` phosphor flash at 0-70%, fades to none (CRT character write effect) |
| `:focus-visible` | `outline: 2px solid` | Added `box-shadow: 0 0 8px var(--accent-glow)` |
| Scrollbar `:active` | *(none)* | `background: rgba(74,122,92,0.5)` + glow shadow |

### Design rationale
- 3-second glow cycle (vs 2s `animate-pulse`) because monitoring systems breathe slowly
- Phosphor flash on reveal simulates CRT character write — text "flares" green then settles
- Border flash on card entry creates "instruments coming online" effect

---

## Phase 3 — Atmospheric Textures

**File:** `frontend/src/assets/styles/main.css`

### New CSS classes
| Class | Effect | Usage |
|-------|--------|-------|
| `.surface-grid` | 24×24px coordinate grid via `linear-gradient` borders | Activity feed, Dashboard recent activity |
| `.surface-scanlines` | `::after` overlay with 2% opacity horizontal banding (4px pitch) | Chat window |
| `.chrome-brackets` | `::before`/`::after` corner bracket decorations (12×12px, 60% opacity) | Dashboard overview cards |

### Design rationale
- Grid evokes graph paper / radar coordinate systems — barely perceptible at normal distance
- Scan-lines are 2% opacity — the absolute minimum to register as texture without noise
- Corner brackets are the visual grammar of HUD overlays / targeting reticles — they frame panels as "viewports"
- All effects use `pointer-events: none` to avoid interfering with content interaction
- `.surface-scanlines` uses `border-radius: inherit` to respect parent container rounding

---

## Phase 4 — Component Application

### AppSidebar.vue
| Element | Change |
|---------|--------|
| Brand dot | `animate-pulse` → `glow-pulse` (slower, glowing breathing) |
| Wordmark "HEIMDALL" | Added `.brand-glow` (phosphor text-shadow) |
| Section labels ("OVERVIEW", "INFRASTRUCTURE", etc.) | `text-text-muted` → `text-accent` (phosphor-green section headers) |
| Active nav indicator bar | `bg-accent` → `bg-accent-bright` + inline `box-shadow: 0 0 6px var(--accent-glow)` |

### StatusBadge.vue
| Element | Change |
|---------|--------|
| Badge container | Added conditional `box-shadow` using matching `--status-*-glow` token per state |
| Status dot | Added `glow-pulse` (active), `glow-pulse-critical` (error), `glow-pulse-warn` (warning) |

### LogEntry.vue
| Element | Change |
|---------|--------|
| Container | `transition-colors` → `transition-all` (needed for shadow transitions) |
| Agent entry hover | `hover:bg-accent-subtle/80` → `hover:bg-accent-subtle hover:shadow-[0_0_12px_var(--accent-glow)]` |

### LogFeed.vue
| Element | Change |
|---------|--------|
| Entry list container | Added `surface-grid rounded-lg p-3 -mx-3` (grid texture on activity list) |

### DashboardPage.vue
| Element | Change |
|---------|--------|
| Four overview cards | Added `chrome-brackets` class (HUD corner decorations) |
| Recent Activity container | Added `surface-grid` class (coordinate grid texture) |
| Monitoring continuous mode dot | `animate-pulse` → `glow-pulse` |

### AgentChatPage.vue
| Element | Change |
|---------|--------|
| Connection status dot (open) | `bg-status-ok` → `bg-status-ok glow-pulse` |
| Connection status dot (connecting) | `animate-pulse` → `glow-pulse-warn` |
| Connection status dot (closed) | `bg-status-critical` → `bg-status-critical glow-pulse-critical` |
| Tool-progress indicator dot | `animate-pulse` → `glow-pulse` |

### ChatWindow.vue
| Element | Change |
|---------|--------|
| Chat container | Added `surface-scanlines` class (CRT scan-line overlay) |
| "Processing" text | Added `brand-glow` class (phosphor text-shadow during thinking) |

---

## Phase 5 — Brand Guidelines Update

**File:** `docs/brand-guidelines.md`

### Key changes
- Identity description: "techno-brutalist" → "technical retro-futurism" with mission-control and CRT references
- Colour philosophy: "Feldgrau" → "Phosphor-Feldgrau" with explanation of saturation evolution (12% → 25%)
- All colour tables updated with new hex values
- External contexts updated with new bright/base accent values
- Added `--accent-glow` to the accent table
- Added status glow note under status table
- Design principles table: replaced "Brutalist restraint" with "Information emits light" and "Atmospheric depth"
- Focus state documentation updated to include glow halo

---

## Files Changed

| File | Kind | Lines changed |
|------|------|---------------|
| `frontend/src/assets/styles/main.css` | Edit | ~120 lines (tokens, effects, textures) |
| `frontend/src/components/common/AppSidebar.vue` | Edit | 4 lines (brand glow, section labels, nav indicator) |
| `frontend/src/components/common/StatusBadge.vue` | Edit | 6 lines (status glow shadows, pulse animations) |
| `frontend/src/components/log/LogEntry.vue` | Edit | 2 lines (transition-all, hover glow) |
| `frontend/src/components/log/LogFeed.vue` | Edit | 1 line (surface-grid on container) |
| `frontend/src/pages/DashboardPage.vue` | Edit | 6 lines (chrome-brackets, surface-grid, glow-pulse) |
| `frontend/src/pages/AgentChatPage.vue` | Edit | 4 lines (status glow-pulse variants) |
| `frontend/src/components/agent/ChatWindow.vue` | Edit | 2 lines (surface-scanlines, brand-glow) |
| `docs/brand-guidelines.md` | Edit | ~40 lines (palette, principles, descriptions) |

## Verification

- `npm run build` — passes cleanly (built in 1.14s)
- `npm run test` — all 50 tests pass (891ms)
- No layout, spacing, or structural changes — only colour/glow/texture
