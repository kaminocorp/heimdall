<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import {
  WebGLRenderer,
  Scene,
  PerspectiveCamera,
  Points,
  BufferGeometry,
  Float32BufferAttribute,
  ShaderMaterial,
  AdditiveBlending,
  Color,
} from 'three'

/* ─────────────────────────────────────────────────
   Props
   ───────────────────────────────────────────────── */

const props = withDefaults(defineProps<{
  /** Dormant mode dims the nebula when no connections exist */
  dormant?: boolean
}>(), { dormant: false })

/* ─────────────────────────────────────────────────
   Refs & State
   ───────────────────────────────────────────────── */

const containerRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)

let renderer: WebGLRenderer | null = null
let scene: Scene | null = null
let camera: PerspectiveCamera | null = null
let animationId: number | null = null
let resizeObserver: ResizeObserver | null = null

// Layer refs for cleanup
let primaryCloud: Points | null = null
let wispTendrils: Points | null = null
let coreMotes: Points | null = null

/* ─────────────────────────────────────────────────
   Ashima 3D Simplex Noise — GLSL
   Shared constant inlined into each ShaderMaterial.
   ───────────────────────────────────────────────── */

const SIMPLEX_NOISE_GLSL = /* glsl */ `
vec3 mod289(vec3 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
vec4 mod289(vec4 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
vec4 permute(vec4 x) { return mod289(((x * 34.0) + 10.0) * x); }
vec4 taylorInvSqrt(vec4 r) { return 1.79284291400159 - 0.85373472095314 * r; }

float snoise(vec3 v) {
  const vec2 C = vec2(1.0/6.0, 1.0/3.0);
  const vec4 D = vec4(0.0, 0.5, 1.0, 2.0);

  vec3 i  = floor(v + dot(v, C.yyy));
  vec3 x0 = v - i + dot(i, C.xxx);

  vec3 g = step(x0.yzx, x0.xyz);
  vec3 l = 1.0 - g;
  vec3 i1 = min(g.xyz, l.zxy);
  vec3 i2 = max(g.xyz, l.zxy);

  vec3 x1 = x0 - i1 + C.xxx;
  vec3 x2 = x0 - i2 + C.yyy;
  vec3 x3 = x0 - D.yyy;

  i = mod289(i);
  vec4 p = permute(permute(permute(
    i.z + vec4(0.0, i1.z, i2.z, 1.0))
  + i.y + vec4(0.0, i1.y, i2.y, 1.0))
  + i.x + vec4(0.0, i1.x, i2.x, 1.0));

  float n_ = 0.142857142857;
  vec3 ns = n_ * D.wyz - D.xzx;

  vec4 j = p - 49.0 * floor(p * ns.z * ns.z);

  vec4 x_ = floor(j * ns.z);
  vec4 y_ = floor(j - 7.0 * x_);

  vec4 x = x_ * ns.x + ns.yyyy;
  vec4 y = y_ * ns.x + ns.yyyy;
  vec4 h = 1.0 - abs(x) - abs(y);

  vec4 b0 = vec4(x.xy, y.xy);
  vec4 b1 = vec4(x.zw, y.zw);

  vec4 s0 = floor(b0) * 2.0 + 1.0;
  vec4 s1 = floor(b1) * 2.0 + 1.0;
  vec4 sh = -step(h, vec4(0.0));

  vec4 a0 = b0.xzyw + s0.xzyw * sh.xxyy;
  vec4 a1 = b1.xzyw + s1.xzyw * sh.zzww;

  vec3 p0 = vec3(a0.xy, h.x);
  vec3 p1 = vec3(a0.zw, h.y);
  vec3 p2 = vec3(a1.xy, h.z);
  vec3 p3 = vec3(a1.zw, h.w);

  vec4 norm = taylorInvSqrt(vec4(dot(p0,p0), dot(p1,p1), dot(p2,p2), dot(p3,p3)));
  p0 *= norm.x;
  p1 *= norm.y;
  p2 *= norm.z;
  p3 *= norm.w;

  vec4 m = max(0.6 - vec4(dot(x0,x0), dot(x1,x1), dot(x2,x2), dot(x3,x3)), 0.0);
  m = m * m;
  return 42.0 * dot(m*m, vec4(dot(p0,x0), dot(p1,x1), dot(p2,x2), dot(p3,x3)));
}
`

/* ─────────────────────────────────────────────────
   Layer 1: Primary Cloud
   6,000 particles — Gaussian volume — 3-octave noise
   ───────────────────────────────────────────────── */

function createPrimaryCloud(): Points {
  const COUNT = 4000
  const positions = new Float32Array(COUNT * 3)
  const phases = new Float32Array(COUNT)
  const sizes = new Float32Array(COUNT)

  for (let i = 0; i < COUNT; i++) {
    // Box-Muller Gaussian distribution
    const u = Math.random() || 0.0001
    const v = Math.random()
    const r = 1.3 * Math.sqrt(-2 * Math.log(u)) * 0.42
    const theta = Math.random() * Math.PI * 2
    const phi = Math.acos(2 * v - 1)

    positions[i * 3] = r * Math.sin(phi) * Math.cos(theta)
    positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta)
    positions[i * 3 + 2] = r * Math.cos(phi)

    phases[i] = Math.random() * Math.PI * 2
    sizes[i] = 1.5 + Math.random() * 2.0
  }

  const geometry = new BufferGeometry()
  geometry.setAttribute('position', new Float32BufferAttribute(positions, 3))
  geometry.setAttribute('phase', new Float32BufferAttribute(phases, 1))
  geometry.setAttribute('size', new Float32BufferAttribute(sizes, 1))

  const material = new ShaderMaterial({
    uniforms: {
      uTime: { value: 0 },
      uLowAmp: { value: 0.45 },
      uMidAmp: { value: 0.18 },
      uCoherence: { value: 1.0 },
      uDormant: { value: props.dormant ? 1.0 : 0.0 },
    },
    vertexShader: /* glsl */ `
      ${SIMPLEX_NOISE_GLSL}

      attribute float phase;
      attribute float size;
      uniform float uTime;
      uniform float uLowAmp;
      uniform float uMidAmp;
      uniform float uCoherence;
      uniform float uDormant;

      varying float vAlpha;
      varying vec3 vDisplaced;
      varying float vDepth;

      void main() {
        float t = uTime;
        vec3 bp = position;

        // 3-octave noise displacement
        vec3 low = vec3(
          snoise(bp * 0.8 + vec3(t * 0.12, t * 0.09, t * 0.07)),
          snoise(bp * 0.8 + vec3(t * 0.08 + 50.0, t * 0.11, t * 0.06)),
          snoise(bp * 0.8 + vec3(t * 0.10, t * 0.07 + 80.0, t * 0.13))
        );

        vec3 mid = vec3(
          snoise(bp * 2.2 + vec3(t * 0.25, t * 0.2, t * 0.18)),
          snoise(bp * 2.2 + vec3(t * 0.22 + 30.0, t * 0.28, t * 0.15)),
          snoise(bp * 2.2 + vec3(t * 0.19, t * 0.23 + 60.0, t * 0.27))
        );

        vec3 high = vec3(
          snoise(bp * 6.0 + vec3(t * 0.5)),
          snoise(bp * 6.0 + vec3(t * 0.45 + 20.0)),
          snoise(bp * 6.0 + vec3(t * 0.55 + 40.0))
        );

        vec3 displaced = bp + low * uLowAmp + mid * uMidAmp + high * 0.035;
        displaced *= uCoherence;

        vDisplaced = displaced;
        vDepth = length(displaced);

        // Alpha — aggressive reduction for additive blending at small viewport
        float dist = length(displaced);
        float coreAlpha = smoothstep(1.8, 0.0, dist) * 0.04;
        float edgeAlpha = smoothstep(0.0, 1.4, dist) * 0.025;
        float flicker = sin(t * 0.4 + phase) * 0.015
                      + sin(t * 0.17 + phase * 2.3) * 0.01;
        vAlpha = clamp(coreAlpha + edgeAlpha + flicker, 0.003, 0.07);

        // Dim in dormant mode
        vAlpha *= mix(1.0, 0.35, uDormant);

        vec4 mvPosition = modelViewMatrix * vec4(displaced, 1.0);
        gl_PointSize = size * (150.0 / -mvPosition.z);
        gl_Position = projectionMatrix * mvPosition;
      }
    `,
    fragmentShader: /* glsl */ `
      uniform float uTime;
      uniform float uDormant;
      varying float vAlpha;
      varying vec3 vDisplaced;
      varying float vDepth;

      void main() {
        // Soft circular point
        vec2 center = gl_PointCoord - 0.5;
        float d = length(center);
        if (d > 0.5) discard;
        float soft = exp(-d * d * 8.0);

        // Spatial colour seed
        float spatial = vDisplaced.x * 0.8 + vDisplaced.y * 0.6 + vDepth * 0.15;

        // Heimdall palette — neutral base, green appears as accent not foundation
        vec3 base     = vec3(0.10, 0.11, 0.13);   // neutral dark slate
        vec3 accent   = vec3(0.14, 0.30, 0.20);   // muted green accent
        vec3 phosphor = vec3(0.18, 0.42, 0.26);   // phosphor highlight
        vec3 tealInfo = vec3(0.12, 0.28, 0.42);   // teal info
        vec3 warmAmb  = vec3(0.30, 0.24, 0.12);   // warm amber muted

        // Mood cycling — varied speeds for organic movement
        float moodA = sin(uTime * 0.037 + spatial * 1.5) * 0.5 + 0.5;
        float moodB = sin(uTime * 0.023 + spatial * 1.2 + 1.8) * 0.5 + 0.5;
        float moodC = sin(uTime * 0.053 + spatial * 1.8 + 3.5) * 0.5 + 0.5;
        float moodD = sin(uTime * 0.041 + spatial * 1.0 + 5.1) * 0.5 + 0.5;

        vec3 color = base;
        color = mix(color, accent,   moodA * 0.35);
        color = mix(color, phosphor, moodB * 0.30);
        color = mix(color, tealInfo, moodC * 0.30);
        color = mix(color, warmAmb,  moodD * 0.20);

        // Global brightness pulse — slow sine so the whole nebula breathes
        float brightPulse = 0.85 + 0.15 * sin(uTime * 0.067);
        color *= brightPulse;

        // In dormant mode, desaturate toward neutral grey
        color = mix(color, vec3(0.10, 0.10, 0.11), uDormant * 0.5);

        gl_FragColor = vec4(color * soft, vAlpha * soft);
      }
    `,
    transparent: true,
    blending: AdditiveBlending,
    depthWrite: false,
  })

  return new Points(geometry, material)
}

/* ─────────────────────────────────────────────────
   Layer 2: Wisp Tendrils
   1,000 particles — uniform shell — radial drift
   ───────────────────────────────────────────────── */

function createWispTendrils(): Points {
  const COUNT = 1000
  const positions = new Float32Array(COUNT * 3)
  const phases = new Float32Array(COUNT)

  for (let i = 0; i < COUNT; i++) {
    const r = 0.6 + Math.random() * 0.9
    const theta = Math.random() * Math.PI * 2
    const phi = Math.acos(2 * Math.random() - 1)

    positions[i * 3] = r * Math.sin(phi) * Math.cos(theta)
    positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta)
    positions[i * 3 + 2] = r * Math.cos(phi)

    phases[i] = Math.random() * Math.PI * 2
  }

  const geometry = new BufferGeometry()
  geometry.setAttribute('position', new Float32BufferAttribute(positions, 3))
  geometry.setAttribute('phase', new Float32BufferAttribute(phases, 1))

  const material = new ShaderMaterial({
    uniforms: {
      uTime: { value: 0 },
      uDormant: { value: props.dormant ? 1.0 : 0.0 },
    },
    vertexShader: /* glsl */ `
      ${SIMPLEX_NOISE_GLSL}

      attribute float phase;
      uniform float uTime;
      uniform float uDormant;

      varying float vAlpha;
      varying vec3 vDisplaced;
      varying float vDepth;

      void main() {
        float t = uTime;
        vec3 bp = position;

        // Radial drift + tangential noise
        float drift = snoise(bp * 1.2 + t * 0.08) * 0.55 + 0.3;
        vec3 dir = normalize(bp + vec3(0.001));
        vec3 tangent = vec3(
          snoise(bp * 1.5 + vec3(t * 0.15, 0.0, 0.0)),
          snoise(bp * 1.5 + vec3(0.0, t * 0.12, 0.0)),
          snoise(bp * 1.5 + vec3(0.0, 0.0, t * 0.18))
        );
        vec3 displaced = bp + dir * drift + tangent * 0.35;

        vDisplaced = displaced;
        vDepth = length(displaced);

        // Very faint wisps
        float dist = length(displaced);
        float fadeOut = smoothstep(2.5, 0.6, dist);
        float flicker = sin(t * 0.3 + phase) * 0.015 + sin(t * 0.13 + phase * 1.7) * 0.01;
        vAlpha = clamp(fadeOut * 0.04 + flicker, 0.003, 0.06);
        vAlpha *= mix(1.0, 0.25, uDormant);

        vec4 mvPosition = modelViewMatrix * vec4(displaced, 1.0);
        gl_PointSize = (3.0 + phase * 0.8) * (150.0 / -mvPosition.z);
        gl_Position = projectionMatrix * mvPosition;
      }
    `,
    fragmentShader: /* glsl */ `
      uniform float uTime;
      uniform float uDormant;
      varying float vAlpha;
      varying vec3 vDisplaced;
      varying float vDepth;

      void main() {
        vec2 center = gl_PointCoord - 0.5;
        float d = length(center);
        if (d > 0.5) discard;
        float soft = exp(-d * d * 6.0);

        float spatial = vDisplaced.x * 0.8 + vDisplaced.y * 0.6 + vDepth * 0.15;

        vec3 base     = vec3(0.08, 0.09, 0.11);
        vec3 accent   = vec3(0.12, 0.28, 0.18);
        vec3 phosphor = vec3(0.18, 0.42, 0.26);
        vec3 tealInfo = vec3(0.12, 0.28, 0.42);

        // 3 moods for wisps — organic cycling
        float moodA = sin(uTime * 0.037 + spatial * 1.5 + 2.0) * 0.5 + 0.5;
        float moodB = sin(uTime * 0.023 + spatial * 1.2 + 3.8) * 0.5 + 0.5;
        float moodC = sin(uTime * 0.053 + spatial * 1.8 + 5.5) * 0.5 + 0.5;

        vec3 color = base;
        color = mix(color, accent,   moodA * 0.35);
        color = mix(color, phosphor, moodB * 0.30);
        color = mix(color, tealInfo, moodC * 0.35);

        color = mix(color, vec3(0.07, 0.07, 0.08), uDormant * 0.5);

        gl_FragColor = vec4(color * soft, vAlpha * soft);
      }
    `,
    transparent: true,
    blending: AdditiveBlending,
    depthWrite: false,
  })

  return new Points(geometry, material)
}

/* ─────────────────────────────────────────────────
   Layer 3: Core Motes
   400 particles — tight Gaussian — micro-jitter
   ───────────────────────────────────────────────── */

function createCoreMotes(): Points {
  const COUNT = 400
  const positions = new Float32Array(COUNT * 3)
  const phases = new Float32Array(COUNT)

  for (let i = 0; i < COUNT; i++) {
    const u = Math.random() || 0.0001
    const v = Math.random()
    const r = 0.28 * Math.sqrt(-2 * Math.log(u)) * 0.35
    const theta = Math.random() * Math.PI * 2
    const phi = Math.acos(2 * v - 1)

    positions[i * 3] = r * Math.sin(phi) * Math.cos(theta)
    positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta)
    positions[i * 3 + 2] = r * Math.cos(phi)

    phases[i] = Math.random() * Math.PI * 2
  }

  const geometry = new BufferGeometry()
  geometry.setAttribute('position', new Float32BufferAttribute(positions, 3))
  geometry.setAttribute('phase', new Float32BufferAttribute(phases, 1))

  const material = new ShaderMaterial({
    uniforms: {
      uTime: { value: 0 },
      uDormant: { value: props.dormant ? 1.0 : 0.0 },
    },
    vertexShader: /* glsl */ `
      ${SIMPLEX_NOISE_GLSL}

      attribute float phase;
      uniform float uTime;
      uniform float uDormant;

      varying float vAlpha;

      void main() {
        float t = uTime;
        vec3 bp = position;

        // High-frequency micro-jitter
        vec3 jitter = vec3(
          snoise(bp * 8.0 + vec3(t * 0.8)),
          snoise(bp * 8.0 + vec3(t * 0.7 + 10.0)),
          snoise(bp * 8.0 + vec3(t * 0.9 + 20.0))
        );
        vec3 displaced = bp + jitter * 0.04;

        float dist = length(displaced);
        float core = smoothstep(0.35, 0.0, dist) * 0.12;
        float flicker = sin(t * 0.6 + phase) * 0.03 + sin(t * 0.23 + phase * 1.9) * 0.02;
        vAlpha = clamp(core + 0.04 + flicker, 0.02, 0.18);
        vAlpha *= mix(1.0, 0.4, uDormant);

        vec4 mvPosition = modelViewMatrix * vec4(displaced, 1.0);
        gl_PointSize = 1.8 * (150.0 / -mvPosition.z);
        gl_Position = projectionMatrix * mvPosition;
      }
    `,
    fragmentShader: /* glsl */ `
      uniform float uTime;
      uniform float uDormant;
      varying float vAlpha;

      void main() {
        vec2 center = gl_PointCoord - 0.5;
        float d = length(center);
        if (d > 0.5) discard;
        float soft = exp(-d * d * 10.0);

        // Candlelight/moonlight cycling — subtle organic shift
        float warmCool = sin(uTime * 0.067) * 0.5 + 0.5;
        float tealShift = sin(uTime * 0.041 + 2.0) * 0.5 + 0.5;
        vec3 candle   = vec3(0.18, 0.32, 0.22);  // warm phosphor
        vec3 moonlit  = vec3(0.14, 0.24, 0.30);  // cool slate
        vec3 tealTint = vec3(0.14, 0.26, 0.38);  // teal accent
        vec3 color = mix(candle, moonlit, warmCool);
        color = mix(color, tealTint, tealShift * 0.3);

        color = mix(color, vec3(0.08, 0.08, 0.09), uDormant * 0.5);

        gl_FragColor = vec4(color * soft, vAlpha * soft);
      }
    `,
    transparent: true,
    blending: AdditiveBlending,
    depthWrite: false,
  })

  return new Points(geometry, material)
}

/* ─────────────────────────────────────────────────
   Scene Setup & Animation Loop
   ───────────────────────────────────────────────── */

function handleResize() {
  if (!containerRef.value || !renderer || !camera) return
  const { clientWidth: w, clientHeight: h } = containerRef.value
  renderer.setSize(w, h)
  camera.aspect = w / h
  camera.updateProjectionMatrix()
}

// Respect prefers-reduced-motion
const prefersReducedMotion = typeof window !== 'undefined'
  ? window.matchMedia('(prefers-reduced-motion: reduce)').matches
  : false

onMounted(() => {
  if (!canvasRef.value || !containerRef.value) return

  const container = containerRef.value
  const canvas = canvasRef.value

  // Renderer
  renderer = new WebGLRenderer({
    canvas,
    antialias: true,
    alpha: true,
    powerPreference: 'high-performance',
  })
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  renderer.setSize(container.clientWidth, container.clientHeight)
  renderer.setClearColor(0x000000, 0)

  // Camera
  camera = new PerspectiveCamera(50, container.clientWidth / container.clientHeight, 0.1, 100)
  camera.position.set(0, 0, 4.5)

  // Scene
  scene = new Scene()

  // Create layers
  primaryCloud = createPrimaryCloud()
  wispTendrils = createWispTendrils()
  coreMotes = createCoreMotes()

  scene.add(primaryCloud)
  scene.add(wispTendrils)
  scene.add(coreMotes)

  // Animation loop
  const startTime = performance.now()

  function animate() {
    animationId = requestAnimationFrame(animate)

    if (!renderer || !scene || !camera) return

    const elapsed = (performance.now() - startTime) / 1000

    if (prefersReducedMotion) {
      // Render once then stop
      renderer.render(scene, camera)
      if (animationId !== null) {
        cancelAnimationFrame(animationId)
        animationId = null
      }
      return
    }

    // Update dormant uniform reactively
    const dormantVal = props.dormant ? 1.0 : 0.0

    // Primary Cloud — modulate amplitude uniforms
    if (primaryCloud) {
      const mat = primaryCloud.material as ShaderMaterial
      const lowAmp = 0.45 + 0.22 * Math.sin(elapsed * 0.031) + 0.12 * Math.sin(elapsed * 0.053 + 1.2)
      const midAmp = 0.18 + 0.10 * Math.sin(elapsed * 0.047 + 0.7) + 0.06 * Math.sin(elapsed * 0.073)
      const coherence = 1.0 + 0.12 * Math.sin(elapsed * 0.019) + 0.06 * Math.sin(elapsed * 0.041 + 2.0)

      mat.uniforms.uTime.value = elapsed
      mat.uniforms.uLowAmp.value = lowAmp
      mat.uniforms.uMidAmp.value = midAmp
      mat.uniforms.uCoherence.value = coherence
      mat.uniforms.uDormant.value = dormantVal

      primaryCloud.rotation.y = elapsed * 0.06
      primaryCloud.rotation.x = elapsed * 0.02
    }

    // Wisp Tendrils — slower rotation for parallax
    if (wispTendrils) {
      const mat = wispTendrils.material as ShaderMaterial
      mat.uniforms.uTime.value = elapsed
      mat.uniforms.uDormant.value = dormantVal

      wispTendrils.rotation.y = elapsed * 0.04
      wispTendrils.rotation.x = elapsed * 0.015
    }

    // Core Motes — fastest rotation
    if (coreMotes) {
      const mat = coreMotes.material as ShaderMaterial
      mat.uniforms.uTime.value = elapsed
      mat.uniforms.uDormant.value = dormantVal

      coreMotes.rotation.y = elapsed * 0.07
      coreMotes.rotation.x = elapsed * 0.025
    }

    renderer.render(scene, camera)
  }

  animate()

  // Resize handling
  resizeObserver = new ResizeObserver(handleResize)
  resizeObserver.observe(container)
})

/* ─────────────────────────────────────────────────
   Cleanup
   ───────────────────────────────────────────────── */

onBeforeUnmount(() => {
  // Cancel animation
  if (animationId !== null) {
    cancelAnimationFrame(animationId)
    animationId = null
  }

  // Disconnect resize observer
  resizeObserver?.disconnect()
  resizeObserver = null

  // Dispose Three.js resources
  const layers = [primaryCloud, wispTendrils, coreMotes]
  for (const layer of layers) {
    if (!layer) continue
    layer.geometry.dispose()
    ;(layer.material as ShaderMaterial).dispose()
  }
  primaryCloud = null
  wispTendrils = null
  coreMotes = null

  // Dispose renderer (releases WebGL context)
  renderer?.dispose()
  renderer = null

  scene = null
  camera = null
})
</script>

<template>
  <div ref="containerRef" class="agent-nebula">
    <canvas ref="canvasRef" class="block w-full h-full" />
  </div>
</template>

<style scoped>
.agent-nebula {
  position: relative;
  width: 100%;
  height: 560px;
}

/* prefers-reduced-motion: render once, no animation */
@media (prefers-reduced-motion: reduce) {
  .agent-nebula canvas {
    animation: none !important;
  }
}
</style>
