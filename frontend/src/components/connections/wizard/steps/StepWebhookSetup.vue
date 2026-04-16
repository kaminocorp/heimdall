<script setup lang="ts">
import { onMounted } from 'vue'
import type { WizardState } from '../flows'
import CopyableField from '@/components/common/CopyableField.vue'

defineProps<{ modelValue: WizardState }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

const payloadExample = `{
  "source": "my-app",
  "level": "error",
  "message": "Connection refused to db-primary",
  "attrs": {
    "host": "web-1",
    "request_id": "req-abc123"
  }
}`

// Webhook connections are auto-configured — always valid.
// The actual token is generated server-side on creation.
onMounted(() => {
  emit('valid', true)
})
</script>

<template>
  <div class="space-y-4">
    <p class="font-mono text-xs text-text-secondary leading-relaxed">
      After creating this connection, Heimdall will generate a unique webhook endpoint and bearer token.
      Configure your application to send logs via HTTP POST.
    </p>

    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-3">
      <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Endpoint Format</p>
      <CopyableField value="POST /api/webhooks/logs" />
    </div>

    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-3">
      <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Payload Format</p>
      <pre class="font-mono text-xs text-text-secondary leading-relaxed overflow-x-auto">{{ payloadExample }}</pre>
    </div>

    <p class="font-mono text-xs text-text-tertiary">
      The webhook token and full URL will be shown after the connection is created.
    </p>
  </div>
</template>
