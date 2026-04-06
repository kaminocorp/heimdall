<script setup lang="ts">
import { onMounted } from 'vue'
import type { WizardState } from '../flows'
import CopyableField from '@/components/common/CopyableField.vue'

defineProps<{ modelValue: WizardState }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

// OTLP connections are auto-configured — always valid.
// The token is generated server-side on creation.
onMounted(() => {
  emit('valid', true)
})
</script>

<template>
  <div class="space-y-6">
    <p class="font-mono text-xs text-text-secondary leading-relaxed">
      After creating this connection, Heimdall will generate a bearer token.
      Configure your OpenTelemetry SDK or collector to export logs to the endpoint below.
    </p>

    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-3">
      <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Endpoint</p>
      <CopyableField value="POST /api/v1/logs" />
    </div>

    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-3">
      <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Payload Format</p>
      <pre class="font-mono text-xs text-text-secondary leading-relaxed overflow-x-auto">{{ payloadExample }}</pre>
    </div>

    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-2">
      <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Compatible Sources</p>
      <ul class="font-mono text-xs text-text-secondary space-y-1 list-disc list-inside">
        <li><span class="text-text-primary">OpenTelemetry SDKs</span> — JS, Python, Go, Java, Rust</li>
        <li><span class="text-text-primary">OTel Collector</span> — OTLP HTTP exporter</li>
        <li><span class="text-text-primary">Neon</span> — Native OTLP log export</li>
        <li><span class="text-text-primary">Heroku Fir</span> — Telemetry drains (OTLP)</li>
        <li><span class="text-text-primary">Fluent Bit / Vector</span> — OTel output plugins</li>
      </ul>
    </div>

    <p class="font-mono text-xs text-text-tertiary">
      The bearer token will be shown after the connection is created. Set it as the
      <code class="text-text-secondary">Authorization: Bearer &lt;token&gt;</code> header in your exporter config.
    </p>
  </div>
</template>

<script lang="ts">
const payloadExample = `{
  "resourceLogs": [{
    "resource": {
      "attributes": [
        { "key": "service.name",
          "value": { "stringValue": "my-app" } }
      ]
    },
    "scopeLogs": [{
      "logRecords": [{
        "timeUnixNano": "1712345678000000000",
        "severityText": "ERROR",
        "severityNumber": 17,
        "body": { "stringValue": "connection refused" },
        "attributes": []
      }]
    }]
  }]
}`
</script>
