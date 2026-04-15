import {
  WebGLRenderer,
  Scene,
  PerspectiveCamera,
  Points,
  BufferGeometry,
  Float32BufferAttribute,
  ShaderMaterial,
  AdditiveBlending,
} from 'three'
import { SIMPLEX_NOISE_GLSL } from './agentNebulaShaders'

// ── Layer factories ──────────────────────────────

function createPrimaryCloud(dormant: boolean): Points {
  const COUNT = 4000
  const positions = new Float32Array(COUNT * 3)
  const phases = new Float32Array(COUNT)
  const sizes = new Float32Array(COUNT)

  for (let i = 0; i < COUNT; i++) {
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
      uDormant: { value: dormant ? 1.0 : 0.0 },
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
        float dist = length(displaced);
        float coreAlpha = smoothstep(1.8, 0.0, dist) * 0.04;
        float edgeAlpha = smoothstep(0.0, 1.4, dist) * 0.025;
        float flicker = sin(t * 0.4 + phase) * 0.015 + sin(t * 0.17 + phase * 2.3) * 0.01;
        vAlpha = clamp(coreAlpha + edgeAlpha + flicker, 0.003, 0.07);
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
        vec2 center = gl_PointCoord - 0.5;
        float d = length(center);
        if (d > 0.5) discard;
        float soft = exp(-d * d * 8.0);
        float spatial = vDisplaced.x * 0.8 + vDisplaced.y * 0.6 + vDepth * 0.15;
        vec3 base     = vec3(0.10, 0.11, 0.13);
        vec3 accent   = vec3(0.14, 0.30, 0.20);
        vec3 phosphor = vec3(0.18, 0.42, 0.26);
        vec3 tealInfo = vec3(0.12, 0.28, 0.42);
        vec3 warmAmb  = vec3(0.30, 0.24, 0.12);
        float moodA = sin(uTime * 0.037 + spatial * 1.5) * 0.5 + 0.5;
        float moodB = sin(uTime * 0.023 + spatial * 1.2 + 1.8) * 0.5 + 0.5;
        float moodC = sin(uTime * 0.053 + spatial * 1.8 + 3.5) * 0.5 + 0.5;
        float moodD = sin(uTime * 0.041 + spatial * 1.0 + 5.1) * 0.5 + 0.5;
        vec3 color = base;
        color = mix(color, accent,   moodA * 0.35);
        color = mix(color, phosphor, moodB * 0.30);
        color = mix(color, tealInfo, moodC * 0.30);
        color = mix(color, warmAmb,  moodD * 0.20);
        float brightPulse = 0.85 + 0.15 * sin(uTime * 0.067);
        color *= brightPulse;
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

function createWispTendrils(dormant: boolean): Points {
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
      uDormant: { value: dormant ? 1.0 : 0.0 },
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

function createCoreMotes(dormant: boolean): Points {
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
      uDormant: { value: dormant ? 1.0 : 0.0 },
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
        float warmCool = sin(uTime * 0.067) * 0.5 + 0.5;
        float tealShift = sin(uTime * 0.041 + 2.0) * 0.5 + 0.5;
        vec3 candle   = vec3(0.18, 0.32, 0.22);
        vec3 moonlit  = vec3(0.14, 0.24, 0.30);
        vec3 tealTint = vec3(0.14, 0.26, 0.38);
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

// ── Composable ───────────────────────────────────

export interface NebulaRefs {
  container: HTMLElement
  canvas: HTMLCanvasElement
}

export interface NebulaInstance {
  dispose: () => void
}

export function initNebula(refs: NebulaRefs, dormant: boolean): NebulaInstance {
  const { container, canvas } = refs

  const renderer = new WebGLRenderer({
    canvas,
    antialias: true,
    alpha: true,
    powerPreference: 'high-performance',
  })
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  renderer.setSize(container.clientWidth, container.clientHeight)
  renderer.setClearColor(0x000000, 0)

  const camera = new PerspectiveCamera(50, container.clientWidth / container.clientHeight, 0.1, 100)
  camera.position.set(0, 0, 4.5)

  const scene = new Scene()

  const primaryCloud = createPrimaryCloud(dormant)
  const wispTendrils = createWispTendrils(dormant)
  const coreMotes = createCoreMotes(dormant)

  scene.add(primaryCloud)
  scene.add(wispTendrils)
  scene.add(coreMotes)

  // Resize handling
  function handleResize() {
    const { clientWidth: w, clientHeight: h } = container
    renderer.setSize(w, h)
    camera.aspect = w / h
    camera.updateProjectionMatrix()
  }

  const resizeObserver = new ResizeObserver(handleResize)
  resizeObserver.observe(container)

  // Respect prefers-reduced-motion
  const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches

  // Animation loop
  const startTime = performance.now()
  let animationId: number | null = null

  function animate() {
    animationId = requestAnimationFrame(animate)

    const elapsed = (performance.now() - startTime) / 1000

    if (prefersReducedMotion) {
      renderer.render(scene, camera)
      if (animationId !== null) {
        cancelAnimationFrame(animationId)
        animationId = null
      }
      return
    }

    const dormantVal = dormant ? 1.0 : 0.0

    // Primary Cloud
    const pcMat = primaryCloud.material as ShaderMaterial
    pcMat.uniforms.uTime.value = elapsed
    pcMat.uniforms.uLowAmp.value = 0.45 + 0.22 * Math.sin(elapsed * 0.031) + 0.12 * Math.sin(elapsed * 0.053 + 1.2)
    pcMat.uniforms.uMidAmp.value = 0.18 + 0.10 * Math.sin(elapsed * 0.047 + 0.7) + 0.06 * Math.sin(elapsed * 0.073)
    pcMat.uniforms.uCoherence.value = 1.0 + 0.12 * Math.sin(elapsed * 0.019) + 0.06 * Math.sin(elapsed * 0.041 + 2.0)
    pcMat.uniforms.uDormant.value = dormantVal
    primaryCloud.rotation.y = elapsed * 0.06
    primaryCloud.rotation.x = elapsed * 0.02

    // Wisp Tendrils
    const wtMat = wispTendrils.material as ShaderMaterial
    wtMat.uniforms.uTime.value = elapsed
    wtMat.uniforms.uDormant.value = dormantVal
    wispTendrils.rotation.y = elapsed * 0.04
    wispTendrils.rotation.x = elapsed * 0.015

    // Core Motes
    const cmMat = coreMotes.material as ShaderMaterial
    cmMat.uniforms.uTime.value = elapsed
    cmMat.uniforms.uDormant.value = dormantVal
    coreMotes.rotation.y = elapsed * 0.07
    coreMotes.rotation.x = elapsed * 0.025

    renderer.render(scene, camera)
  }

  animate()

  // Dispose function
  function dispose() {
    if (animationId !== null) {
      cancelAnimationFrame(animationId)
      animationId = null
    }
    resizeObserver.disconnect()

    const layers = [primaryCloud, wispTendrils, coreMotes]
    for (const layer of layers) {
      layer.geometry.dispose()
      ;(layer.material as ShaderMaterial).dispose()
    }

    renderer.dispose()
  }

  return { dispose }
}
