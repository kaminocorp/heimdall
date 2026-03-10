<template>
  <footer class="border-t border-border">
    <div class="w-full px-6 py-3 flex items-center justify-between text-[11px] font-mono text-text-muted tracking-wide">
      <!-- Left: attribution -->
      <span class="opacity-50 hidden sm:inline">A Kamino Corporation product.</span>

      <!-- Center: copyright + links -->
      <div class="flex items-center gap-1.5 uppercase">
        <span>&copy; 2026 Heimdall</span>
        <span class="opacity-30">&middot;</span>
        <RouterLink to="/terms" class="hover:text-text-secondary transition-colors">Terms</RouterLink>
        <span class="opacity-30">&middot;</span>
        <RouterLink to="/privacy" class="hover:text-text-secondary transition-colors">Privacy</RouterLink>
        <span class="opacity-30">&middot;</span>
        <RouterLink to="/features" class="hover:text-text-secondary transition-colors">Platform</RouterLink>
        <span class="opacity-30">&middot;</span>
        <RouterLink to="/pricing" class="hover:text-text-secondary transition-colors">Pricing</RouterLink>
        <span class="opacity-30">&middot;</span>
        <RouterLink to="/security" class="hover:text-text-secondary transition-colors">Security</RouterLink>
        <span class="opacity-30">&middot;</span>
        <a href="https://github.com" target="_blank" rel="noopener noreferrer" class="hover:text-text-secondary transition-colors">GitHub</a>
        <span class="opacity-30">&middot;</span>
        <a href="https://x.com/kaminocorpo" target="_blank" rel="noopener noreferrer" class="hover:text-text-secondary transition-colors"><svg class="w-3 h-3 inline-block" viewBox="0 0 24 24" fill="currentColor"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg></a>
      </div>

      <!-- Right: quote -->
      <span class="opacity-30 italic hidden sm:inline">"Who watches the watchers?"</span>
    </div>
  </footer>
</template>
