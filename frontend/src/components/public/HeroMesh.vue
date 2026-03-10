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

  // ── Grid — ultra-dense for fabric feel ──────
  const COLS = 420
  const ROWS = 256
  const SPACING = 3.2

  // ── Camera ────────────────────────────────────
  const PERSPECTIVE = 680
  const CAMERA_Y = 140
  const CAMERA_Z = 380
  const ROT_X = -0.78
  const ROT_Y = 0.22

  const cosRX = Math.cos(ROT_X), sinRX = Math.sin(ROT_X)
  const cosRY = Math.cos(ROT_Y), sinRY = Math.sin(ROT_Y)

  const gridW = (COLS - 1) * SPACING
  const gridH = (ROWS - 1) * SPACING
  const halfW = gridW / 2
  const halfH = gridH / 2

  // ── Anomaly peaks ─────────────────────────────
  const peaks = [
    { cx: 0.30, cz: 0.35, sigma: 0.07, amp: 110, freqX: 0.13, freqZ: 0.09, pulseFreq: 0.35, pulsePhase: 0 },
    { cx: 0.62, cz: 0.55, sigma: 0.06, amp: 130, freqX: -0.08, freqZ: 0.11, pulseFreq: 0.28, pulsePhase: 2.1 },
    { cx: 0.78, cz: 0.25, sigma: 0.055, amp: 95, freqX: 0.10, freqZ: -0.07, pulseFreq: 0.40, pulsePhase: 4.0 },
    { cx: 0.45, cz: 0.70, sigma: 0.065, amp: 100, freqX: -0.06, freqZ: 0.08, pulseFreq: 0.32, pulsePhase: 1.4 },
    { cx: 0.15, cz: 0.60, sigma: 0.05, amp: 80, freqX: 0.09, freqZ: 0.12, pulseFreq: 0.45, pulsePhase: 3.2 },
  ]

  // ── Scan sweep ────────────────────────────────
  const SCAN_SPEED = 0.08
  const SCAN_WIDTH = 0.07
  const SCAN_FADE = 0.12

  let w = 0, h = 0, dpr = 1

  function resize() {
    dpr = Math.min(window.devicePixelRatio, 2)
    const rect = el!.getBoundingClientRect()
    w = rect.width
    h = rect.height
    el!.width = w * dpr
    el!.height = h * dpr
  }

  resize()
  window.addEventListener('resize', resize)

  function gaussian(dx: number, dz: number, sigma: number): number {
    return Math.exp(-(dx * dx + dz * dz) / (2 * sigma * sigma))
  }

  function getHeight(nx: number, nz: number, t: number): number {
    let y = Math.sin(nx * 7.5 - t * 1.4) * 12
      + Math.sin(nz * 6.0 + t * 0.4) * 8
      + Math.sin((nx * 3.0 + nz * 2.0) - t * 0.7) * 6
      + Math.cos(nx * 10.0 - t * 0.9) * Math.sin(nz * 8.0 + t * 0.5) * 5

    for (let i = 0; i < peaks.length; i++) {
      const p = peaks[i]
      const pcx = ((p.cx + Math.sin(t * p.freqX + p.pulsePhase) * 0.08) % 1 + 1) % 1
      const pcz = ((p.cz + Math.sin(t * p.freqZ + p.pulsePhase) * 0.06) % 1 + 1) % 1
      const pulse = 0.3 + 0.7 * Math.max(0, Math.sin(t * p.pulseFreq + p.pulsePhase))
      y += gaussian(nx - pcx, nz - pcz, p.sigma) * p.amp * pulse
    }

    return y
  }

  // ── Render direct to ImageData for speed ──────
  function render(time: number) {
    if (!isActive) return

    const pw = el!.width   // pixel width (CSS * dpr)
    const ph = el!.height
    const imgData = ctx!.createImageData(pw, ph)
    const data = imgData.data

    const t = prefersReducedMotion ? 0 : time / 1000
    const cxScreen = w * dpr / 2
    const cyScreen = h * dpr / 2
    const scanPos = (t * SCAN_SPEED) % 1.0
    const perspDpr = PERSPECTIVE * dpr

    for (let row = 0; row < ROWS; row++) {
      const nz = row / (ROWS - 1)
      const lz = row * SPACING - halfH

      for (let col = 0; col < COLS; col++) {
        const nx = col / (COLS - 1)
        const lx = col * SPACING - halfW
        const ly = getHeight(nx, nz, t)

        // Rotate Y then X
        const rx = lx * cosRY - lz * sinRY
        const rz1 = lx * sinRY + lz * cosRY
        const ry = ly * cosRX - rz1 * sinRX
        const rz = ly * sinRX + rz1 * cosRX

        const wz = rz + CAMERA_Z
        if (wz <= 20) continue

        // Project
        const scale = perspDpr / wz
        const sx = (cxScreen + rx * scale) | 0
        const sy = (cyScreen + (ry + CAMERA_Y) * scale) | 0

        if (sx < 0 || sx >= pw || sy < 0 || sy >= ph) continue

        // Depth alpha
        const depthAlpha = Math.max(0.15, 1.0 - (wz - 100) / 700)

        // Scan glow
        let distToScan = Math.abs(nx - scanPos)
        if (distToScan > 0.5) distToScan = 1.0 - distToScan
        const inScanBand = Math.max(0, 1.0 - distToScan / SCAN_WIDTH)
        const inScanTrail = Math.max(0, 1.0 - distToScan / SCAN_FADE)
        const heightFactor = Math.max(0, Math.min(1, ly / 80))
        const scanGlow = Math.max(inScanBand * 0.6, inScanTrail * heightFactor * 0.9)

        const alpha = Math.min(1, depthAlpha * (0.6 + heightFactor * 0.4)) * 255

        // Write pixel
        const idx = (sy * pw + sx) * 4
        if (scanGlow > 0.1) {
          // Green — blend toward accent color
          const g = scanGlow
          data[idx]     = (220 * (1 - g) + 90 * g) | 0   // R
          data[idx + 1] = (225 * (1 - g) + 158 * g) | 0  // G
          data[idx + 2] = (220 * (1 - g) + 106 * g) | 0  // B
          data[idx + 3] = Math.min(255, alpha * (1 + scanGlow * 0.5)) | 0

          // Glow halo: write surrounding pixels for anomaly peaks
          if (scanGlow > 0.3 && ly > 40) {
            const glowR = ((1 + scanGlow) * dpr) | 0
            const ga = (alpha * scanGlow * 0.25) | 0
            for (let dy = -glowR; dy <= glowR; dy++) {
              for (let dx = -glowR; dx <= glowR; dx++) {
                if (dx === 0 && dy === 0) continue
                const gx = sx + dx, gy = sy + dy
                if (gx < 0 || gx >= pw || gy < 0 || gy >= ph) continue
                const gi = (gy * pw + gx) * 4
                // Additive blend
                data[gi]     = Math.min(255, data[gi] + 30)
                data[gi + 1] = Math.min(255, data[gi + 1] + 55)
                data[gi + 2] = Math.min(255, data[gi + 2] + 35)
                data[gi + 3] = Math.min(255, data[gi + 3] + ga)
              }
            }
          }
        } else {
          data[idx]     = 220
          data[idx + 1] = 225
          data[idx + 2] = 220
          data[idx + 3] = alpha | 0
        }
      }
    }

    ctx!.putImageData(imgData, 0, 0)
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
