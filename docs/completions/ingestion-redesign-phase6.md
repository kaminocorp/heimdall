# Ingestion Redesign — Phase 6: Compose the New Ingestion View

**Status:** Complete  
**Plan:** `docs/executing/ingestion-page-redesign.md`

---

## What was done

Rewrote `ConnectionsPage.vue` to remove the old Blueprint/List dual-view system and replace it with the unified Ingestion view: connection bubbles → flow lines → agent nebula.

## What was removed from ConnectionsPage

| Removed | Why |
|---------|-----|
| `BlueprintView` import + template block | Replaced by bubble + nebula layout |
| `ConnectionList` import + template block | Replaced by bubble + nebula layout |
| `ViewToggle` import + template usage | No longer two view modes |
| `ConnectorLogo` import | Was only used by the Phase 2 preview gallery |
| `viewMode` ref + localStorage watcher | No longer needed |
| Blueprint/List conditional `<template v-if/else>` | Replaced by single unified view |
| Skeleton loader (grid of card placeholders) | Replaced with bubble-shaped skeleton |

## What replaced it

The main content area is now a single `<div class="relative">` containing:

1. **Connection bubbles** — `flex flex-wrap justify-center gap-4` row of `ConnectionBubble` components, `z-10` above the flow lines
2. **Flow lines** — `FlowLines` SVG overlay (only rendered when connections > 0)
3. **Agent nebula** — `AgentNebula` with `dormant` prop tied to empty connections
4. **Empty state overlay** — centred text + CTA button over the dormant nebula when no connections exist

## Design decisions

### No IngestionView wrapper component

The plan originally called for a separate `IngestionView.vue` orchestrator. In practice, `ConnectionsPage.vue` already manages all the state (store, modals, forms, refs). Adding a wrapper would just proxy every event and ref through an extra layer. The page itself is the orchestrator.

### Empty state over dormant nebula

When no connections exist, the nebula still renders in dormant mode (dimmer, desaturated). An absolutely positioned overlay shows "No connections yet" with a "+ New Connection" CTA button. The breathing dormant nebula beneath communicates "the system is here, waiting to be activated" — more evocative than a blank page.

### Skeleton loader restyled to match bubbles

The old skeleton was a 2-column grid of card placeholders matching `ConnectionCard`. The new skeleton shows 3 bubble-shaped placeholders (48px square + two text lines) matching `ConnectionBubble` proportions.

### Page header simplified

The `<div>` wrapper around ViewToggle + button was replaced with just the button directly. No wrapping div needed for a single element.

## Orphaned components (safe to delete in Phase 7)

Grep confirmed **zero imports** of these components outside their own internal references:

| File | Internal refs only |
|------|--------------------|
| `BlueprintView.vue` | Imports BlueprintZone |
| `BlueprintZone.vue` | Imports BlueprintNode |
| `BlueprintNode.vue` | Leaf component |
| `ViewToggle.vue` | Standalone |
| `ConnectionList.vue` | Imports ConnectionCard |
| `ConnectionCard.vue` | Leaf component |

No test files reference them. The `heimdall_connections_view` localStorage key is also no longer read or written anywhere.

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/pages/ConnectionsPage.vue` | Rewrite | Stripped old views, ViewToggle, viewMode; unified bubble + flow + nebula layout; new skeleton; empty state |

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| Old imports removed | Grep confirms zero references to BlueprintView, ConnectionList, ViewToggle, ConnectionCard outside their own files |
| All existing flows preserved | Wizard, edit form, test modal, GitHub repo selector, error banners unchanged |
| Empty state | Renders correctly with 0 connections (dormant nebula + CTA) |
