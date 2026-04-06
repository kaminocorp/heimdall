<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import type { WizardState } from '../flows'
import BaseSelect from '@/components/common/BaseSelect.vue'

const props = defineProps<{ modelValue: WizardState }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

const initCfg = props.modelValue.config
const port = ref(String(initCfg.port ?? '6514'))
const protocol = ref((initCfg.protocol as string) ?? 'tcp')

watch(() => props.modelValue.config, (cfg) => {
  const newPort = String(cfg.port ?? '6514')
  const newProtocol = (cfg.protocol as string) ?? 'tcp'
  if (newPort !== port.value) port.value = newPort
  if (newProtocol !== protocol.value) protocol.value = newProtocol
}, { deep: true })

function sync() {
  const parsed = parseInt(port.value, 10)
  const portNum = Number.isNaN(parsed) ? 6514 : Math.max(1, Math.min(65535, parsed))
  emit('update:modelValue', {
    ...props.modelValue,
    config: {
      ...props.modelValue.config,
      port: portNum,
      protocol: protocol.value,
    },
  })
  const portValid = !Number.isNaN(parsed) && parsed >= 1 && parsed <= 65535
  emit('valid', portValid)
}

watch([port, protocol], sync)
onMounted(sync)
</script>

<template>
  <div class="space-y-6">
    <p class="font-mono text-xs text-text-secondary leading-relaxed">
      Heimdall will listen for syslog messages over TCP on the configured port.
      Platforms like Render, Heroku, and DigitalOcean can forward logs via syslog.
    </p>

    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Port</label>
      <input
        v-model="port"
        type="text"
        inputmode="numeric"
        placeholder="6514"
        autofocus
        class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors"
      />
      <p class="mt-1 font-mono text-xs text-text-muted">TCP port to listen on (default: 6514 for syslog-TLS)</p>
    </div>

    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Protocol</label>
      <BaseSelect
        :modelValue="protocol"
        @update:modelValue="protocol = String($event)"
        :options="[
          { value: 'tcp', label: 'TCP (plaintext)' },
          { value: 'tls', label: 'TLS (encrypted)' },
        ]"
      />
      <p class="mt-1 font-mono text-xs text-text-muted">TLS requires server-level certificate configuration</p>
    </div>

    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-2">
      <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Platform Setup</p>
      <p class="font-mono text-xs text-text-secondary leading-relaxed">
        After creating this connection, configure your platform to forward logs to Heimdall's syslog endpoint.
      </p>
      <ul class="font-mono text-xs text-text-secondary space-y-1 list-disc list-inside">
        <li><span class="text-text-primary">Render:</span> Dashboard → Account Settings → Log Streams</li>
        <li><span class="text-text-primary">Heroku:</span> <code class="text-accent">heroku drains:add syslog+tls://host:port</code></li>
        <li><span class="text-text-primary">Linux:</span> Configure rsyslog/syslog-ng to forward via TCP</li>
      </ul>
    </div>
  </div>
</template>
