<script setup lang="ts">
defineProps<{
  variant?: 'inline' | 'full'
}>()
</script>

<template>
  <!-- Inline footer (landing page) -->
  <footer v-if="variant === 'inline'" class="border-t border-border">
    <div class="max-w-6xl mx-auto px-6 py-4 flex items-center justify-between text-xs font-mono text-text-muted">
      <span>&copy; 2026 Kamino Corp</span>
      <div class="flex items-center gap-4">
        <a href="https://github.com" target="_blank" rel="noopener noreferrer" class="hover:text-text-secondary transition-colors">GitHub</a>
        <a href="https://x.com/kaminocorpo" target="_blank" rel="noopener noreferrer" class="hover:text-text-secondary transition-colors">X</a>
      </div>
    </div>
  </footer>

  <!-- Full footer (features, pricing) -->
  <footer v-else class="border-t border-border bg-bg-elevated/50">
    <div class="max-w-6xl mx-auto px-6 py-8">
      <div class="flex flex-col sm:flex-row items-start justify-between gap-6">
        <div class="flex flex-col gap-3">
          <div class="flex items-center gap-3">
            <svg class="w-6 h-6" viewBox="0 0 200 200" fill="none" xmlns="http://www.w3.org/2000/svg">
              <g transform="translate(100, 100)">
                <polygon points="0,-80 69.3,-40 69.3,40 0,80 -69.3,40 -69.3,-40" fill="none" stroke="#5a6b5a" stroke-width="2.8" stroke-linejoin="miter"/>
                <polygon points="0,-70 60.6,-35 60.6,35 0,70 -60.6,35 -60.6,-35" fill="none" stroke="#5a6b5a" stroke-width="0.8" stroke-linejoin="miter"/>
                <g fill="#5a6b5a">
                  <path d="M -2.4,-21 L -5.6,-66 L 5.6,-66 L 2.4,-21 Z" transform="rotate(0)"/>
                  <path d="M -2.4,-21 L -5.6,-66 L 5.6,-66 L 2.4,-21 Z" transform="rotate(60)"/>
                  <path d="M -2.4,-21 L -5.6,-66 L 5.6,-66 L 2.4,-21 Z" transform="rotate(120)"/>
                  <path d="M -2.4,-21 L -5.6,-66 L 5.6,-66 L 2.4,-21 Z" transform="rotate(180)"/>
                  <path d="M -2.4,-21 L -5.6,-66 L 5.6,-66 L 2.4,-21 Z" transform="rotate(240)"/>
                  <path d="M -2.4,-21 L -5.6,-66 L 5.6,-66 L 2.4,-21 Z" transform="rotate(300)"/>
                </g>
                <g fill="#5a6b5a" opacity="0.7">
                  <path d="M -1.6,-21 L -3.6,-46 L 3.6,-46 L 1.6,-21 Z" transform="rotate(30)"/>
                  <path d="M -1.6,-21 L -3.6,-46 L 3.6,-46 L 1.6,-21 Z" transform="rotate(90)"/>
                  <path d="M -1.6,-21 L -3.6,-46 L 3.6,-46 L 1.6,-21 Z" transform="rotate(150)"/>
                  <path d="M -1.6,-21 L -3.6,-46 L 3.6,-46 L 1.6,-21 Z" transform="rotate(210)"/>
                  <path d="M -1.6,-21 L -3.6,-46 L 3.6,-46 L 1.6,-21 Z" transform="rotate(270)"/>
                  <path d="M -1.6,-21 L -3.6,-46 L 3.6,-46 L 1.6,-21 Z" transform="rotate(330)"/>
                </g>
                <circle r="18" fill="none" stroke="#5a6b5a" stroke-width="2"/>
                <circle r="12" fill="none" stroke="#5a6b5a" stroke-width="1.4"/>
                <circle r="5.5" fill="#5a6b5a"/>
              </g>
            </svg>
            <span class="font-mono font-semibold text-xs tracking-wide text-text-secondary">HEIMDALL</span>
          </div>
          <p class="text-xs text-text-muted font-mono">&copy; 2026 Kamino Corp</p>
        </div>
        <div class="flex gap-10 text-xs font-mono">
          <div class="flex flex-col gap-2">
            <span class="text-text-secondary font-medium">Product</span>
            <RouterLink to="/features" class="text-text-muted hover:text-text-secondary transition-colors">Features</RouterLink>
            <RouterLink to="/pricing" class="text-text-muted hover:text-text-secondary transition-colors">Pricing</RouterLink>
            <RouterLink to="/login" class="text-text-muted hover:text-text-secondary transition-colors">Login</RouterLink>
          </div>
          <div class="flex flex-col gap-2">
            <span class="text-text-secondary font-medium">Links</span>
            <a href="https://github.com" target="_blank" rel="noopener noreferrer" class="text-text-muted hover:text-text-secondary transition-colors">GitHub</a>
            <a href="https://x.com/kaminocorpo" target="_blank" rel="noopener noreferrer" class="text-text-muted hover:text-text-secondary transition-colors">X</a>
          </div>
        </div>
      </div>
    </div>
  </footer>
</template>
