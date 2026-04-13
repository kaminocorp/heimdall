<script setup lang="ts">
import { computed } from 'vue'
import type { Connection } from '@/types/connection'
import ConnectorLogo from '@/components/icons/ConnectorLogo.vue'

const props = defineProps<{
  connection: Connection
  testing?: boolean
  /** Stagger index for mount animation delay */
  index?: number
}>()

defineEmits<{
  click: [connection: Connection]
}>()

const typeLabels: Record<string, string> = {
  postgres: 'PostgreSQL',
  supabase: 'Supabase',
  webhook_logs: 'Webhook',
  syslog: 'Syslog',
  github: 'GitHub',
  otlp: 'OpenTelemetry',
  datadog: 'Datadog',
  mysql: 'MySQL',
}

const displayType = computed(() => typeLabels[props.connection.type] ?? props.connection.type)

const statusColor = computed(() => {
  switch (props.connection.status) {
    case 'active': return 'bg-status-ok'
    case 'error': return 'bg-status-critical'
    default: return 'bg-text-muted'
  }
})

const statusGlow = computed(() => {
  switch (props.connection.status) {
    case 'active': return 'box-shadow: 0 0 6px var(--status-ok-glow)'
    case 'error': return 'box-shadow: 0 0 6px var(--status-critical-glow)'
    default: return ''
  }
})
</script>

<template>
  <button
    class="connection-bubble group"
    :class="{ 'bubble-active': connection.status === 'active', 'bubble-testing': testing }"
    :style="{ '--bubble-index': index ?? 0 }"
    @click="$emit('click', connection)"
  >
    <!-- Logo -->
    <div class="bubble-logo">
      <ConnectorLogo
        :type="connection.type"
        :size="28"
        class="transition-colors duration-200"
        :class="connection.status === 'active' ? 'text-accent' : 'text-text-muted'"
      />
    </div>

    <!-- Info -->
    <div class="bubble-info">
      <span class="bubble-name">{{ connection.name }}</span>
      <span class="bubble-type">{{ displayType }}</span>
    </div>

    <!-- Status dot -->
    <span
      class="bubble-status"
      :class="[statusColor, { 'animate-pulse': connection.status === 'active' || testing }]"
      :style="statusGlow"
    />

    <!-- Testing overlay -->
    <span v-if="testing" class="bubble-testing-label">Testing</span>
  </button>
</template>

<style scoped>
.connection-bubble {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 1.25rem;
  min-width: 120px;
  max-width: 160px;
  border: 1px solid var(--border);
  border-radius: 0.75rem;
  background: var(--bg-surface);
  cursor: pointer;
  transition: all 0.2s ease;

  /* Staggered fade-in on mount */
  opacity: 0;
  transform: translateY(8px);
  animation: bubble-enter 0.35s ease-out forwards;
  animation-delay: calc(var(--bubble-index, 0) * 80ms);
}

.connection-bubble:hover {
  border-color: var(--border-hover);
  background: var(--bg-surface-hover);
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
}

.connection-bubble:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.bubble-active {
  border-color: var(--accent-border);
  box-shadow: 0 0 12px var(--accent-glow);
}

.bubble-active:hover {
  box-shadow: 0 0 18px var(--accent-glow), 0 4px 16px rgba(0, 0, 0, 0.3);
}

.bubble-testing {
  border-color: var(--accent-border);
  animation: bubble-enter 0.35s ease-out forwards, bubble-pulse 1.5s ease-in-out infinite;
  animation-delay: calc(var(--bubble-index, 0) * 80ms), 0s;
}

.bubble-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 0.5rem;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  transition: border-color 0.2s ease;
}

.connection-bubble:hover .bubble-logo {
  border-color: var(--border-hover);
}

.bubble-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.125rem;
  min-width: 0;
  width: 100%;
}

.bubble-name {
  font-family: var(--font-mono);
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.bubble-type {
  font-family: var(--font-mono);
  font-size: 0.6rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-muted);
}

.bubble-status {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.bubble-testing-label {
  position: absolute;
  top: 0.375rem;
  right: 0.375rem;
  font-family: var(--font-mono);
  font-size: 0.5rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--accent);
  animation: bubble-pulse 1.5s ease-in-out infinite;
}

@keyframes bubble-enter {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes bubble-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

@media (prefers-reduced-motion: reduce) {
  .connection-bubble {
    animation: none;
    opacity: 1;
    transform: none;
  }
  .bubble-testing {
    animation: none;
    opacity: 1;
  }
}
</style>
