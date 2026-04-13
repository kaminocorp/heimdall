<script setup lang="ts">
import { computed } from 'vue'
import { flows } from './flows'
import PlatformCard from './PlatformCard.vue'

const emit = defineEmits<{
  select: [flowId: string]
}>()

const platformLogs = computed(() => flows.filter(f => f.section === 'platform_log'))
const directProtocols = computed(() => flows.filter(f => f.section === 'direct_protocol'))
const agentTools = computed(() => flows.filter(f => f.section === 'agent_tool'))
const outbound = computed(() => flows.filter(f => f.section === 'outbound'))
</script>

<template>
  <div class="space-y-8">

    <!-- Section 1: Platform log sources -->
    <section>
      <div class="mb-4">
        <h3 class="font-mono text-xs font-bold uppercase tracking-widest text-text-primary">
          Where do your logs come from?
        </h3>
        <p class="font-sans text-xs text-text-muted mt-1">
          Pick your platform — we'll handle the wiring.
        </p>
      </div>
      <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
        <PlatformCard
          v-for="flow in platformLogs"
          :key="flow.id"
          :flow="flow"
          @select="emit('select', $event)"
        />
      </div>
    </section>

    <!-- "or" divider -->
    <div class="flex items-center gap-4">
      <div class="flex-1 h-px bg-border" />
      <span class="font-mono text-[10px] uppercase tracking-widest text-text-muted">or</span>
      <div class="flex-1 h-px bg-border" />
    </div>

    <!-- Section 2: Direct protocols -->
    <section>
      <div class="mb-4">
        <h3 class="font-mono text-xs font-bold uppercase tracking-widest text-text-primary">
          Connect directly
        </h3>
        <p class="font-sans text-xs text-text-muted mt-1">
          Already know the protocol? Skip the platform.
        </p>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <PlatformCard
          v-for="flow in directProtocols"
          :key="flow.id"
          :flow="flow"
          @select="emit('select', $event)"
        />
      </div>
    </section>

    <!-- Strong divider — different purpose below -->
    <div class="border-t border-border pt-2">
      <div class="flex items-center gap-2 mb-4">
        <svg class="w-3.5 h-3.5 text-text-muted shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
        </svg>
        <div>
          <h3 class="font-mono text-xs font-bold uppercase tracking-widest text-text-primary">
            Agent investigation tools
          </h3>
          <p class="font-sans text-xs text-text-muted mt-0.5">
            These don't send logs — they let the agent query your systems during investigations and chat.
          </p>
        </div>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <PlatformCard
          v-for="flow in agentTools"
          :key="flow.id"
          :flow="flow"
          @select="emit('select', $event)"
        />
      </div>
    </div>

    <!-- Strong divider — outbound channels -->
    <div class="border-t border-border pt-2">
      <div class="flex items-center gap-2 mb-4">
        <svg class="w-3.5 h-3.5 text-text-muted shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M22 2 11 13" />
          <path d="m22 2-7 20-4-9-9-4 20-7z" />
        </svg>
        <div>
          <h3 class="font-mono text-xs font-bold uppercase tracking-widest text-text-primary">
            Outbound channels
          </h3>
          <p class="font-sans text-xs text-text-muted mt-0.5">
            Where the agent delivers alerts, reports, and tickets when it finds something.
          </p>
        </div>
      </div>
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <PlatformCard
          v-for="flow in outbound"
          :key="flow.id"
          :flow="flow"
          @select="emit('select', $event)"
        />
      </div>
    </div>

  </div>
</template>
