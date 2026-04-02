<script setup lang="ts">
const problems = [
  {
    icon: 'M15.536 8.464a5 5 0 010 7.072M12 12h.01M18.364 5.636a9 9 0 010 12.728M5.636 18.364a9 9 0 010-12.728M8.464 15.536a5 5 0 010-7.072',
    title: 'Alerts are dumb',
    text: 'Pattern-match alerts fire on every spike with no understanding of context. Real incidents drown in noise.',
  },
  {
    icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z',
    title: 'Investigation is manual',
    text: 'When something looks wrong, a developer has to context-switch, cross-reference dashboards, and piece together a picture from scratch.',
  },
  {
    icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
    title: 'Knowledge is lost',
    text: "The dev who debugged last week's incident carries the context in their head. At 3am, the on-call engineer starts from zero.",
  },
]

const steps = [
  { number: '1', title: 'Connect your infrastructure', text: 'Link your databases, log sources, and codebases. Takes under 5 minutes.' },
  { number: '2', title: 'Heimdall watches', text: 'The agent monitors continuously, learning what normal looks like for your system.' },
  { number: '3', title: 'Get actionable reports', text: 'When anomalies appear, Heimdall investigates autonomously and surfaces structured diagnoses.' },
]

const features = [
  {
    icon: 'M15 12a3 3 0 11-6 0 3 3 0 016 0zM2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z',
    title: 'Always watching',
    text: 'Continuous log and database monitoring. Knows what normal looks like. Recognises when something deviates.',
  },
  {
    icon: 'M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z',
    title: 'Investigates, not just alerts',
    text: 'When anomalies appear, the agent queries your database, searches logs, and builds a diagnosis — before paging you.',
  },
  {
    icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2',
    title: 'Remembers every incident',
    text: 'Long-term memory accumulates past incidents, patterns, and failure modes. Each investigation makes the next one better.',
  },
  {
    icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z',
    title: 'Ask anything, anytime',
    text: '"What happened with payments last night?" Chat with the agent to get instant answers backed by real data.',
  },
  {
    icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
    title: 'Incident reports, not log dumps',
    text: 'Timeline, investigation steps, historical context, and diagnostic assessment — all in one structured document.',
  },
  {
    icon: 'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1',
    title: 'Connect in minutes',
    text: 'PostgreSQL, webhooks, syslog, GitHub. More connectors coming. No complex setup.',
  },
]
</script>

<template>
  <main class="pt-8 flex-1">
      <!-- Section 1: The Problem -->
      <section class="py-20 px-6">
        <div class="max-w-6xl mx-auto">
          <p class="text-xs font-mono text-accent tracking-widest uppercase mb-3">The Problem</p>
          <h2 class="font-mono text-3xl sm:text-4xl font-bold tracking-tight mb-4">
            Modern monitoring is broken
          </h2>
          <p class="text-text-secondary max-w-2xl mb-12">
            Too much noise, not enough insight. Your team deserves better than raw log dumps.
          </p>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div v-for="p in problems" :key="p.title" class="p-6 rounded border border-border bg-bg-surface">
              <div class="w-8 h-8 rounded bg-accent-subtle border border-accent-border flex items-center justify-center mb-4">
                <svg class="w-4 h-4 text-accent" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" :d="p.icon"/></svg>
              </div>
              <h3 class="font-mono font-semibold text-sm mb-2">{{ p.title }}</h3>
              <p class="text-sm text-text-secondary leading-relaxed">{{ p.text }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- Section 2: How It Works -->
      <section class="py-20 px-6 border-t border-border">
        <div class="max-w-6xl mx-auto">
          <p class="text-xs font-mono text-accent tracking-widest uppercase mb-3">How It Works</p>
          <h2 class="font-mono text-3xl sm:text-4xl font-bold tracking-tight mb-12">
            Three steps to autonomous monitoring
          </h2>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div v-for="s in steps" :key="s.number">
              <div class="w-10 h-10 rounded-full border-2 border-accent-border flex items-center justify-center mb-4">
                <span class="font-mono font-bold text-accent text-sm">{{ s.number }}</span>
              </div>
              <h3 class="font-mono font-semibold text-sm mb-2">{{ s.title }}</h3>
              <p class="text-sm text-text-secondary leading-relaxed">{{ s.text }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- Section 3: Core Features -->
      <section class="py-20 px-6 border-t border-border">
        <div class="max-w-6xl mx-auto">
          <p class="text-xs font-mono text-accent tracking-widest uppercase mb-3">Core Features</p>
          <h2 class="font-mono text-3xl sm:text-4xl font-bold tracking-tight mb-12">
            Everything you need to stop babysitting logs
          </h2>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div v-for="f in features" :key="f.title" class="p-6 rounded border border-border bg-bg-surface hover:border-border-hover transition-colors">
              <div class="w-8 h-8 rounded bg-accent-subtle border border-accent-border flex items-center justify-center mb-4">
                <svg class="w-4 h-4 text-accent" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" :d="f.icon"/></svg>
              </div>
              <h3 class="font-mono font-semibold text-sm mb-2">{{ f.title }}</h3>
              <p class="text-sm text-text-secondary leading-relaxed">{{ f.text }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- Section 4: CTA Banner -->
      <section class="py-20 px-6 border-t border-border">
        <div class="max-w-3xl mx-auto text-center">
          <h2 class="font-mono text-3xl sm:text-4xl font-bold tracking-tight mb-6">
            Stop babysitting your logs.
          </h2>
          <RouterLink
            to="/login"
            class="inline-block px-8 py-3 text-sm font-mono font-medium bg-accent hover:bg-accent-hover text-bg-primary rounded transition-colors"
          >
            Start Monitoring
          </RouterLink>
          <p class="mt-4 text-xs font-mono text-text-muted">
            14-day free trial &middot; Unlimited connections &middot; Cancel anytime
          </p>
        </div>
      </section>
  </main>
</template>
