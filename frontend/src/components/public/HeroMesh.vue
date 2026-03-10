<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const canvas = ref<HTMLCanvasElement>()
let animationId = 0
let isActive = true

const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches

onMounted(() => {
  const el = canvas.value
  if (!el) return
  const ctx = el.getContext('2d')
  if (!ctx) return

  // ── Grid ──────────────────────────────────────
  const COLS = 150
  const ROWS = 90
  const SPACING = 9

  // ── Camera ────────────────────────────────────
  const PERSPECTIVE = 680
  const CAMERA_Y = -160
  const CAMERA_Z = 380
  const ROT_X = -0.78
  const ROT_Y = 0.22

  const cosRX = Math.cos(ROT_X), sinRX = Math.sin(ROT_X)
  const cosRY = Math.cos(ROT_Y), sinRY = Math.sin(ROT_Y)

  const gridW = (COLS - 1) * SPACING
  const gridH = (ROWS - 1) * SPACING
  const halfW = gridW / 2
  const halfH = gridH / 2

  // ── Anomaly peak definitions ──────────────────
  // Each peak drifts slowly across the grid and pulses in amplitude.
  // cx/cz are base positions (0–1 normalized), drift adds slow movement.
  const peaks = [
    { cx: 0.30, cz: 0.35, sigma: 0.07, amp: 110, freqX: 0.13, freqZ: 0.09, pulseFreq: 0.35, pulsePhase: 0 },
    { cx: 0.62, cz: 0.55, sigma: 0.06, amp: 130, freqX: -0.08, freqZ: 0.11, pulseFreq: 0.28, pulsePhase: 2.1 },
    { cx: 0.78, cz: 0.25, sigma: 0.055, amp: 95, freqX: 0.10, freqZ: -0.07, pulseFreq: 0.40, pulsePhase: 4.0 },
    { cx: 0.45, cz: 0.70, sigma: 0.065, amp: 100, freqX: -0.06, freqZ: 0.08, pulseFreq: 0.32, pulsePhase: 1.4 },
    { cx: 0.15, cz: 0.60, sigma: 0.05, amp: 80, freqX: 0.09, freqZ: 0.12, pulseFreq: 0.45, pulsePhase: 3.2 },
  ]

  // ── Scan sweep state ──────────────────────────
  const SCAN_SPEED = 0.08       // normalized units per second
  const SCAN_WIDTH = 0.07       // width of the green band
  const SCAN_FADE = 0.12        // trail fade width behind the sweep

  // ── Projected point storage ───────────────────
  // 2D array [row][col] of projected screen coords + metadata
  const projected: {
    sx: number; sy: number; wz: number; height: number; scanGlow: number
  }[][] = Array.from({ length: ROWS }, () =>
    Array.from({ length: COLS }, () => ({ sx: 0, sy: 0, wz: 0, height: 0, scanGlow: 0 }))
  )

  let w = 0, h = 0

  function resize() {
    const dpr = Math.min(window.devicePixelRatio, 2)
    const rect = el!.getBoundingClientRect()
    w = rect.width
    h = rect.height
    el!.width = w * dpr
    el!.height = h * dpr
    ctx!.setTransform(dpr, 0, 0, dpr, 0, 0)
  }

  resize()
  window.addEventListener('resize', resize)

  // Gaussian function (for anomaly peaks)
  function gaussian(dx: number, dz: number, sigma: number): number {
    return Math.exp(-(dx * dx + dz * dz) / (2 * sigma * sigma))
  }

  function getHeight(nx: number, nz: number, t: number): number {
    // Base: gentle flowing waves (data stream moving right-to-left)
    let y = Math.sin(nx * 7.5 - t * 1.4) * 12
      + Math.sin(nz * 6.0 + t * 0.4) * 8
      + Math.sin((nx * 3.0 + nz * 2.0) - t * 0.7) * 6
      + Math.cos(nx * 10.0 - t * 0.9) * Math.sin(nz * 8.0 + t * 0.5) * 5

    // Anomaly peaks: Gaussian bumps that pulse and drift
    for (let i = 0; i < peaks.length; i++) {
      const p = peaks[i]
      const pcx = ((p.cx + Math.sin(t * p.freqX + p.pulsePhase) * 0.08) % 1 + 1) % 1
      const pcz = ((p.cz + Math.sin(t * p.freqZ + p.pulsePhase) * 0.06) % 1 + 1) % 1
      const pulse = 0.3 + 0.7 * Math.max(0, Math.sin(t * p.pulseFreq + p.pulsePhase))
      y += gaussian(nx - pcx, nz - pcz, p.sigma) * p.amp * pulse
    }

    return y
  }

  function render(time: number) {
    if (!isActive) return

    ctx!.clearRect(0, 0, w, h)

    const t = prefersReducedMotion ? 0 : time / 1000
    const cx = w / 2
    const cy = h / 2

    // Scan position (loops 0→1)
    const scanPos = (t * SCAN_SPEED) % 1.0

    // ── Phase 1: Project all points ─────────────
    let minWz = Infinity, maxWz = -Infinity

    for (let row = 0; row < ROWS; row++) {
      for (let col = 0; col < COLS; col++) {
        const nx = col / (COLS - 1) // normalized 0–1
        const nz = row / (ROWS - 1)

        const lx = col * SPACING - halfW
        const lz = row * SPACING - halfH
        const ly = getHeight(nx, nz, t)

        // Rotate Y then X
        const rx = lx * cosRY - lz * sinRY
        const rz1 = lx * sinRY + lz * cosRY
        const ry = ly * cosRX - rz1 * sinRX
        const rz = ly * sinRX + rz1 * cosRX

        const wz = rz + CAMERA_Z

        // Scan glow: how close is this column to the scan sweep?
        // Use normalized X position for the sweep direction
        let distToScan = Math.abs(nx - scanPos)
        if (distToScan > 0.5) distToScan = 1.0 - distToScan // wrap-around
        const inScanBand = Math.max(0, 1.0 - distToScan / SCAN_WIDTH)
        const inScanTrail = Math.max(0, 1.0 - distToScan / SCAN_FADE)

        // Height factor: taller points glow more (anomalies light up)
        const heightFactor = Math.max(0, Math.min(1, ly / 80))
        const scanGlow = Math.max(inScanBand * 0.6, inScanTrail * heightFactor * 0.9)

        const pt = projected[row][col]
        pt.wz = wz
        pt.height = ly
        pt.scanGlow = scanGlow

        if (wz > 10) {
          const scale = PERSPECTIVE / wz
          pt.sx = cx + rx * scale
          pt.sy = cy + (ry + CAMERA_Y) * scale
          if (wz < minWz) minWz = wz
          if (wz > maxWz) maxWz = wz
        } else {
          pt.sx = -9999
        }
      }
    }

    const wzRange = maxWz - minWz || 1

    // ── Phase 2: Draw wireframe lines ───────────
    // Horizontal lines (along the data-stream direction)
    ctx!.lineWidth = 0.5
    for (let row = 0; row < ROWS; row++) {
      for (let col = 0; col < COLS - 1; col++) {
        const a = projected[row][col]
        const b = projected[row][col + 1]
        if (a.sx < -999 || b.sx < -999) continue

        const avgWz = (a.wz + b.wz) / 2
        const depthAlpha = 1.0 - ((avgWz - minWz) / wzRange) * 0.85
        const avgGlow = (a.scanGlow + b.scanGlow) / 2

        if (avgGlow > 0.05) {
          ctx!.strokeStyle = `rgba(90, 158, 106, ${depthAlpha * avgGlow * 0.4})`
        } else {
          ctx!.strokeStyle = `rgba(200, 210, 200, ${depthAlpha * 0.08})`
        }
        ctx!.beginPath()
        ctx!.moveTo(a.sx, a.sy)
        ctx!.lineTo(b.sx, b.sy)
        ctx!.stroke()
      }
    }

    // Vertical lines (cross-stream structure)
    for (let row = 0; row < ROWS - 1; row++) {
      for (let col = 0; col < COLS; col++) {
        const a = projected[row][col]
        const b = projected[row + 1][col]
        if (a.sx < -999 || b.sx < -999) continue

        const avgWz = (a.wz + b.wz) / 2
        const depthAlpha = 1.0 - ((avgWz - minWz) / wzRange) * 0.85
        const avgGlow = (a.scanGlow + b.scanGlow) / 2

        if (avgGlow > 0.05) {
          ctx!.strokeStyle = `rgba(90, 158, 106, ${depthAlpha * avgGlow * 0.3})`
        } else {
          ctx!.strokeStyle = `rgba(200, 210, 200, ${depthAlpha * 0.04})`
        }
        ctx!.beginPath()
        ctx!.moveTo(a.sx, a.sy)
        ctx!.lineTo(b.sx, b.sy)
        ctx!.stroke()
      }
    }

    // ── Phase 3: Draw dots (on top of lines) ────
    for (let row = 0; row < ROWS; row++) {
      for (let col = 0; col < COLS; col++) {
        const pt = projected[row][col]
        if (pt.sx < -20 || pt.sx > w + 20 || pt.sy < -20 || pt.sy > h + 20) continue
        if (pt.wz <= 10) continue

        const scale = PERSPECTIVE / pt.wz
        const depthNorm = (pt.wz - minWz) / wzRange
        const baseAlpha = 1.0 - depthNorm * 0.75

        // Dot size: base + height boost (anomaly peaks get bigger dots)
        const heightBoost = Math.max(0, pt.height / 100) * 0.8
        const size = (1.2 + heightBoost) * scale * 0.7

        if (size < 0.25) continue

        const glow = pt.scanGlow

        if (glow > 0.1) {
          // Green glow for scan-detected points
          const glowAlpha = baseAlpha * (0.5 + glow * 0.5)
          ctx!.fillStyle = `rgba(90, 158, 106, ${glowAlpha})`

          // Outer glow for strong detections (anomaly peaks in scan band)
          if (glow > 0.4 && pt.height > 40) {
            ctx!.globalAlpha = glow * 0.3
            ctx!.beginPath()
            ctx!.arc(pt.sx, pt.sy, size * 3, 0, Math.PI * 2)
            ctx!.fill()
          }

          ctx!.globalAlpha = 1
          ctx!.beginPath()
          ctx!.arc(pt.sx, pt.sy, size, 0, Math.PI * 2)
          ctx!.fill()
        } else {
          // White dot with depth fade
          ctx!.fillStyle = `rgba(220, 225, 220, ${baseAlpha})`
          ctx!.beginPath()
          ctx!.arc(pt.sx, pt.sy, size, 0, Math.PI * 2)
          ctx!.fill()
        }
      }
    }

    animationId = requestAnimationFrame(render)
  }

  animationId = requestAnimationFrame(render)

  onUnmounted(() => {
    isActive = false
    cancelAnimationFrame(animationId)
    window.removeEventListener('resize', resize)
  })
})
</script>

<template>
  <canvas
    ref="canvas"
    class="absolute inset-0 w-full h-full"
    aria-hidden="true"
  />
</template>
