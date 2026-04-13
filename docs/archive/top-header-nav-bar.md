# Top Header Navigation Bar & Sidebar Slim-Down

## Status: Complete

## Summary

Replaced the sidebar-first layout with a header-first layout inspired by
Supabase's dashboard. A full-width top header bar now handles identity context
(org, app, profile), while the sidebar is slimmed down to pure page navigation.
The "HEIMDALL" branding is restyled to match the public homepage's spaced-out
monospace treatment, and the green status dot + "Status: Active" label are
removed.

---

## Motivation

The original sidebar packed five concerns into one vertical column: branding,
status indicator, app selection, page navigation, and user identity. This
created a dense, top-heavy sidebar where the most-used function (page
navigation) was pushed below two chrome sections. The Supabase/Vercel/Linear
pattern separates **hierarchy context** (top bar: which org, which app, who am
I) from **page navigation** (sidebar: what page am I on), which is a better
fit for a multi-app SaaS product.

Specific requests addressed:
1. **HEIMDALL text** — spell it out with letter-spacing to match the public
   homepage (`H E I M D A L L` with `tracking-[0.25em]`).
2. **Remove green dot and "Status: Active"** — unnecessary chrome; monitoring
   status is visible on the Dashboard page itself.
3. **Add Supabase-style top header** — breadcrumb trail: Logo → Org → App,
   with profile dropdown on the right.

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/common/AppHeader.vue` | **New** | Top nav bar component (260 lines) |
| `frontend/src/components/common/AppSidebar.vue` | Edit | Stripped to navigation-only (95 lines, down from 203) |
| `frontend/src/layouts/DefaultLayout.vue` | Edit | Header-first layout structure |

---

## AppHeader.vue — New Component

### Structure

```
┌─────────────────────────────────────────────────────────────────┐
│ [eye icon] H E I M D A L L  /  Org Name ▾  /  App Name ▾   [P]│
└─────────────────────────────────────────────────────────────────┘
```

**Height:** 48px (`h-12`), fixed via `flex-shrink-0`.

### Left section: Logo + breadcrumbs

- **Heimdall icon** — the favicon SVG (radial eye pattern) inlined so it can
  use `currentColor` with `text-accent`. Avoids `<img>` + CSS filter hacks.
- **"H E I M D A L L"** — `font-mono text-xs font-semibold uppercase
  tracking-[0.25em]`, matching the public `PublicNav.vue` style. Hidden on
  small screens (`hidden sm:inline`) to save horizontal space.
- **Org breadcrumb** — building icon + org name + chevron. Dropdown shows the
  current org (currently single-org per user, but the dropdown is structurally
  ready for multi-org). Clicking elsewhere closes it.
- **App breadcrumb** — database icon + current app name + chevron. Dropdown
  lists all apps with the active one highlighted in accent, plus a
  "+ New application" action at the bottom (separated by a border). Selecting
  an app calls `app.selectApp()` and closes the dropdown. "+ New application"
  opens the `AppWizard` modal.

### Right section: Profile

- **Avatar circle** — 28px rounded circle showing the first letter of the
  user's email, styled with `border-border bg-bg-surface`.
- **Profile dropdown** — email display, "Settings" link (navigates to settings
  route), "Sign out" button (calls `auth.logout()` + `app.reset()`).

### Interaction patterns

- **Mutual exclusion** — opening any dropdown closes the other two. Prevents
  overlapping panels.
- **Outside-click dismiss** — a document-level click listener checks each
  dropdown's ref container. This is the same pattern used by `BaseSelect.vue`.
- **Transitions** — `opacity + translate-y` enter/leave matching the existing
  `BaseSelect` dropdown animation.

---

## AppSidebar.vue — Slim-Down

### Removed

| Section | Lines removed | Moved to |
|---------|--------------|----------|
| Brand header (green dot, "Heimdall", "Status: Active") | 10 | Removed entirely; branding now in AppHeader |
| App selector (`BaseSelect` + label) | 10 | AppHeader app breadcrumb |
| User footer (Settings link, org name, email, Sign Out) | 20 | AppHeader profile dropdown |
| `AppWizard` mount point | 2 | AppHeader |
| Related imports (`useAuthStore`, `useAppStore`, `BaseSelect`, `AppWizard`, `useRouter`) | 8 | No longer needed |
| `handleSelectApp`, `handleLogout`, `closeWizard`, `appOptions` | 25 | Logic moved to AppHeader |

### Retained

- Navigation sections (Overview, Infrastructure, Agent, Intelligence) — unchanged.
- Active indicator bar with phosphor glow — unchanged.
- Section labels in accent green — unchanged.
- Mobile support (`mobile` prop, `close` emit) — unchanged.

### Dimensional change

- Width: `w-60` (240px) → `w-56` (224px). The sidebar no longer needs space
  for the app selector dropdown, so 16px narrower feels tighter without
  cramping the nav items.

---

## DefaultLayout.vue — Layout Restructure

### Before

```
┌──────────────────────────────────────┐
│ Sidebar (full height) │   Content    │
│  - Brand                             │
│  - App selector                      │
│  - Navigation        │   (scrolls    │
│  - User footer       │    whole      │
│                      │    page)      │
└──────────────────────────────────────┘
```

`min-h-screen flex` — the entire page scrolls as one unit.

### After

```
┌──────────────────────────────────────┐
│           AppHeader (48px)           │
├──────────┬───────────────────────────┤
│ Sidebar  │                           │
│ (nav     │   Content                 │
│  only)   │   (overflow-y-auto)       │
│          │                           │
└──────────┴───────────────────────────┘
```

`h-screen flex-col overflow-hidden` → header is `flex-shrink-0`, the row
below is `flex-1 min-h-0`. Content area scrolls independently via
`overflow-y-auto`. This is important because:

1. **Agent Chat** uses `h-full flex-col` and needs a fixed-height parent to
   avoid infinite expansion.
2. **macOS scroll bounce** is prevented — the viewport is locked, only the
   content pane scrolls.
3. **Header stays pinned** without `position: fixed` or z-index gymnastics.

### Mobile adjustments

- Hamburger button repositioned from `top-4` to `top-14` so it sits below the
  header bar instead of overlapping it.
- Mobile sidebar overlay unchanged — it still teleports to `<body>` with
  slide-in animation.

---

## Design decisions

### Why inline the favicon SVG?

The favicon (`/favicon.svg`) uses a black `<rect>` background with white
`<polygon>` shapes. Using it as an `<img>` in the header would require CSS
filter hacks (`invert`, `brightness`) to make it work on the dark UI — fragile
and imprecise. Inlining the SVG `<g>` group (without the black background
rect) and using `fill="currentColor"` with `text-accent` gives exact colour
control and keeps the icon consistent with the design token system.

### Why not remove the sidebar entirely?

The Supabase reference has both a top nav and a left sidebar. With 10
navigation items across 4 sections, a top-only nav would either need a mega
menu or would crowd the header. The sidebar-for-navigation pattern scales
cleanly as new pages are added.

### Why `h-screen overflow-hidden` instead of `min-h-screen`?

The previous `min-h-screen` layout let the entire page scroll, which meant
the sidebar scrolled off-screen on long pages. The new fixed-viewport approach
keeps header and sidebar always visible, with only the content area scrolling.
This is the standard pattern for dashboard apps (Supabase, Vercel, GitHub).

### Why keep the org dropdown with only one org?

The dropdown structure is already built and costs nothing at runtime. When
multi-org support is added later, the UI is ready — just populate the list.
Currently it shows the single org as a highlighted (non-interactive) row,
which still serves as a visual confirmation of context.
