<script setup lang="ts">
const tiers = [
  {
    name: 'Starter',
    price: '$0',
    period: '/mo',
    description: 'For individuals getting started with autonomous monitoring.',
    features: ['2 connections', '7-day log retention', 'Agent chat', 'Incident reports'],
    cta: 'Get Started',
    ctaStyle: 'outline' as const,
    highlight: false,
  },
  {
    name: 'Pro',
    price: '$49',
    period: '/mo',
    description: 'For teams running production systems that need always-on monitoring.',
    features: ['Unlimited connections', '30-day log retention', 'Continuous monitoring mode', 'Institutional memory', 'Slack & Discord alerts'],
    cta: 'Start Free Trial',
    ctaStyle: 'filled' as const,
    highlight: true,
    badge: 'Popular',
  },
  {
    name: 'Enterprise',
    price: 'Custom',
    period: '',
    description: 'For organisations with complex infrastructure and compliance needs.',
    features: ['Everything in Pro', '90-day log retention', 'SSO & audit logs', 'Dedicated support', 'Custom integrations'],
    cta: 'Contact Us',
    ctaStyle: 'outline' as const,
    highlight: false,
  },
]
</script>

<template>
  <main class="pt-8 flex-1">
      <section class="py-20 px-6">
        <div class="max-w-6xl mx-auto">
          <p class="text-xs font-mono text-accent tracking-widest uppercase mb-3">Pricing</p>
          <h2 class="font-mono text-3xl sm:text-4xl font-bold tracking-tight mb-4">
            Simple, transparent pricing
          </h2>
          <p class="text-text-secondary max-w-2xl mb-12">
            Start free, scale when you're ready.
          </p>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-5xl">
            <div
              v-for="tier in tiers"
              :key="tier.name"
              class="p-6 rounded flex flex-col relative"
              :class="tier.highlight ? 'border-2 border-accent/40 bg-bg-surface' : 'border border-border bg-bg-surface'"
            >
              <!-- Badge -->
              <div v-if="tier.badge" class="absolute -top-3 left-6">
                <span class="px-2.5 py-0.5 text-xs font-mono font-medium bg-accent text-bg-primary rounded tracking-wider uppercase">
                  {{ tier.badge }}
                </span>
              </div>

              <p class="text-xs font-mono text-text-muted tracking-widest uppercase mb-1">{{ tier.name }}</p>
              <div class="flex items-baseline gap-1 mb-4">
                <span class="font-mono text-3xl font-bold">{{ tier.price }}</span>
                <span v-if="tier.period" class="text-sm text-text-muted font-mono">{{ tier.period }}</span>
              </div>
              <p class="text-sm text-text-secondary mb-6">{{ tier.description }}</p>

              <ul class="space-y-2 mb-8 flex-1">
                <li v-for="feature in tier.features" :key="feature" class="text-sm text-text-secondary flex items-start gap-2">
                  <svg class="w-4 h-4 text-accent mt-0.5 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/></svg>
                  {{ feature }}
                </li>
              </ul>

              <RouterLink
                v-if="tier.name !== 'Enterprise'"
                to="/login"
                class="block text-center px-4 py-2.5 text-sm font-mono font-medium rounded transition-colors"
                :class="tier.ctaStyle === 'filled'
                  ? 'bg-accent hover:bg-accent-hover text-bg-primary'
                  : 'border border-accent-border text-text-secondary hover:text-text-primary hover:border-accent/40'"
              >
                {{ tier.cta }}
              </RouterLink>
              <a
                v-else
                href="mailto:hello@crimsonsun.dev"
                class="block text-center px-4 py-2.5 text-sm font-mono font-medium border border-accent-border text-text-secondary hover:text-text-primary hover:border-accent/40 rounded transition-colors"
              >
                {{ tier.cta }}
              </a>
            </div>
          </div>
        </div>
      </section>
  </main>
</template>
