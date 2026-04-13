# Ingestion Redesign — Phase 7: Polish & Performance

**Status:** Complete  
**Plan:** `docs/executing/ingestion-page-redesign.md`

---

## What was done

Final cleanup pass: deleted orphaned components, split Three.js into a lazy vendor chunk, raised the chunk size warning limit, and verified all tests pass.

## Dead code removal

Deleted 6 orphaned components that were replaced by the new Ingestion view. Grep confirmed zero imports of any of these outside their own internal references, and no test files reference them.

| Deleted file | Was used by |
|-------------|-------------|
| `BlueprintView.vue` | Old ConnectionsPage (blueprint mode) |
| `BlueprintZone.vue` | BlueprintView |
| `BlueprintNode.vue` | BlueprintZone |
| `ViewToggle.vue` | Old ConnectionsPage header |
| `ConnectionList.vue` | Old ConnectionsPage (list mode) |
| `ConnectionCard.vue` | ConnectionList |

The `heimdall_connections_view` localStorage key is no longer read or written anywhere.

## Bundle size optimisation

### Problem

Three.js was inlined into the `ConnectionsPage` chunk, making it 551KB (143KB gzipped). Since the Connections page is lazy-loaded via the router, this only affects navigation to that route — but 143KB gzipped for a single page is heavy.

### Fix

Added `manualChunks` to `vite.config.js` (the active config — `.js` takes precedence over `.ts` when both exist) to split `node_modules/three` into a separate `vendor-three` chunk.

### Result

| Chunk | Before | After | Delta |
|-------|--------|-------|-------|
| `ConnectionsPage` | 551KB (143KB gz) | 62.5KB (21KB gz) | **-88%** |
| `vendor-three` | — | 487KB (122KB gz) | New (lazy) |
| `index` (main) | 366KB (115KB gz) | 366KB (115KB gz) | No change |

Three.js loads lazily only when the user navigates to the Connections page. The main bundle is completely unaffected. The `chunkSizeWarningLimit` was raised to 600KB to suppress the warning for the vendor-three chunk (it's a known, expected large dependency).

### Discovery: dual vite config files

The project has both `vite.config.js` and `vite.config.ts`. Vite picks `.js` over `.ts` when both exist, so the `.ts` changes were silently ignored during initial attempts. The `manualChunks` config was applied to both files.

## `prefers-reduced-motion` coverage

Already implemented across all new components:

| Component | Behaviour with reduced motion |
|-----------|-------------------------------|
| `AgentNebula.vue` | Renders one frame, stops animation loop |
| `FlowLines.vue` | Dash animation frozen at offset 0 |
| `ConnectionBubble.vue` | Mount animation disabled, immediate opacity 1 |

## Test results

All 52 frontend tests pass (7 test files, 0 failures).

## Files changed

| File | Kind | Change |
|------|------|--------|
| `BlueprintView.vue` | **Deleted** | Orphaned |
| `BlueprintZone.vue` | **Deleted** | Orphaned |
| `BlueprintNode.vue` | **Deleted** | Orphaned |
| `ViewToggle.vue` | **Deleted** | Orphaned |
| `ConnectionList.vue` | **Deleted** | Orphaned |
| `ConnectionCard.vue` | **Deleted** | Orphaned |
| `vite.config.js` | Edit | Added `build.rollupOptions.output.manualChunks` + `chunkSizeWarningLimit` |
| `vite.config.ts` | Edit | Synced same build config |

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean, no chunk size warnings |
| `vitest run` | 52/52 tests pass |
| Grep for deleted imports | Zero references remaining |
| `vendor-three` chunk exists | Yes — 487KB, loaded lazily |
| `ConnectionsPage` chunk | 62.5KB (down from 551KB) |
