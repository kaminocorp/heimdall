# Ingestion Redesign — Phase 1: Agent Nebula Component

**Status:** Complete  
**Plan:** `docs/executing/ingestion-page-redesign.md`

---

## What was built

`frontend/src/components/connections/AgentNebula.vue` — a self-contained Vue 3 component that renders a 3-layer volumetric particle nebula using raw Three.js and custom GLSL shaders. It represents the Heimdall agent as a breathing, living entity on the Ingestion page.

## Dependencies added

| Package | Type | Version |
|---------|------|---------|
| `three` | runtime | ^0.175.0 |
| `@types/three` | dev | ^0.175.0 |

No other packages. No TresJS — raw imperative Three.js in a Vue component.

## Architecture

```
AgentNebula.vue
  └─ <canvas> (owned by Three.js WebGLRenderer)
       ├─ PrimaryCloud   — 4,000 pts, Gaussian volume, 3-octave simplex noise
       ├─ WispTendrils   — 1,000 pts, uniform shell, radial drift + tangential noise
       └─ CoreMotes      —   400 pts, tight Gaussian core, high-freq micro-jitter
```

Each layer is a `THREE.Points` object with custom `ShaderMaterial` (vertex + fragment shaders). All three share an inlined Ashima 3D Simplex noise GLSL function. The component creates the renderer/camera/scene in `onMounted`, runs `requestAnimationFrame`, and fully disposes everything in `onBeforeUnmount`.

## Key design decisions

### Raw Three.js over TresJS

TresJS adds a declarative abstraction layer that's unnecessary here — the nebula has no reactive data binding to Vue. The particle positions, colours, and animation are entirely GPU-driven via shader uniforms. A single Vue component with imperative Three.js lifecycle management is simpler, avoids a dependency, and maps directly to the Elephantasm reference architecture.

### Dark base colours for additive blending

The initial implementation used pale base colours `(0.82, 0.90, 0.84)` ported from the Elephantasm's pearlescent palette. At the Elephantasm's full-screen scale, particles spread across thousands of pixels so additive accumulation stays within range. At 340px height, the same Gaussian density compressed into fewer pixels — the center blew out to pure white.

**Fix:** Base colours dropped to `(0.18, 0.30, 0.22)` (dark feldgrau). With additive blending, dozens of overlapping particles now peak at a visible phosphor green rather than saturating to white. Per-particle alpha was also slashed from max 0.65 to max 0.12.

### Reduced particle budget

| Layer | Elephantasm ref | Heimdall |
|-------|----------------|----------|
| Primary Cloud | 14,000 | 4,000 |
| Wisp Tendrils | 2,500 | 1,000 |
| Core Motes | 800 | 400 |
| **Total** | **17,300** | **5,400** |

The viewport is ~340px tall vs full-screen. Density per screen pixel scales with the square of size reduction, so a ~3x size reduction needs more than a 3x particle cut.

### No OrbitControls

The nebula auto-rotates and breathes. User interaction is with the Vue layer (bubbles, modals), not the 3D scene. Removing OrbitControls keeps the component simpler and avoids scroll/drag conflicts with the page.

## Props

| Prop | Type | Default | Purpose |
|------|------|---------|---------|
| `dormant` | `boolean` | `false` | Dims all layers (lower alpha, desaturated colours) when no connections exist |

`dormant` is driven by a GLSL uniform (`uDormant`), so the dimming happens per-fragment with zero CPU cost.

## Colour palette

| Role | RGB | Hex approx | Design token |
|------|-----|------------|--------------|
| Base | `(0.18, 0.30, 0.22)` | #2e4d38 | Dark feldgrau |
| Deep green mood | `(0.20, 0.40, 0.28)` | #335a47 | Near `--accent` |
| Phosphor mood | `(0.28, 0.55, 0.34)` | #478c57 | Near `--accent-bright` |
| Teal mood | `(0.18, 0.38, 0.52)` | #2e6185 | Near `--status-info` |
| Warm amber mood | `(0.45, 0.38, 0.15)` | #736126 | Near `--status-warn` |

## Tuned parameters

| Parameter | Value | Location |
|-----------|-------|----------|
| Primary count | 4,000 | `createPrimaryCloud()` |
| Primary point sizes | 1.5–3.5 | attribute init |
| Primary alpha range | 0.005–0.12 | vertex shader |
| Size scale factor | 150 | all vertex shaders |
| Wisp alpha range | 0.003–0.06 | vertex shader |
| Wisp point sizes | 3.0–3.8 base | vertex shader |
| Core motes alpha | 0.06–0.35 | vertex shader |
| Core motes point size | 1.8 base | vertex shader |
| Coherence range | 0.82–1.18 | animation loop |
| Canvas height | 340px | CSS `.agent-nebula` |
| Camera position | z=4.5, FOV 50 | scene setup |

## Accessibility

- `prefers-reduced-motion: reduce` — renders one frame then stops the animation loop
- No keyboard/focus interaction on the canvas itself

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/package.json` | Edit | Added `three`, `@types/three` |
| `frontend/src/components/connections/AgentNebula.vue` | **New** | The 3D particle component |
| `frontend/src/pages/ConnectionsPage.vue` | Edit | Temporary preview wiring (to be replaced in Phase 6) |

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| Visual at 340px height | Dark atmospheric green nebula, no white blowout |
| Mount/unmount cycle | WebGL context properly disposed |
