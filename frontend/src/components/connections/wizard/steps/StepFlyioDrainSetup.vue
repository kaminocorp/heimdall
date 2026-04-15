<script setup lang="ts">
import { onMounted } from 'vue'
import type { WizardState } from '../flows'
import CopyableField from '@/components/common/CopyableField.vue'

defineProps<{ modelValue: WizardState }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

// Drain setup is display-only — always valid.
// The actual webhook token is generated server-side on connection creation.
onMounted(() => {
  emit('valid', true)
})
</script>

<template>
  <div class="space-y-4">
    <p class="font-mono text-xs text-text-secondary leading-relaxed">
      Heimdall will create a webhook endpoint for your logs.
      Deploy the
      <span class="text-text-primary font-medium">Fly Log Shipper</span>
      in your Fly org to push logs to this endpoint in near-real-time.
    </p>

    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-3">
      <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Setup Commands</p>
      <p class="font-mono text-[11px] text-text-tertiary mb-2">
        Run these in your terminal after the connection is created.
        The webhook URL and token will be shown on the connection detail page.
      </p>
      <CopyableField value="fly launch --image flyio/log-shipper:latest --no-public-ips" />
      <CopyableField value="fly secrets set ACCESS_TOKEN=$(fly auth token)" />
      <CopyableField value="fly secrets set HTTP_URL=<your-heimdall-webhook-url>" />
      <CopyableField value="fly secrets set HTTP_TOKEN=<your-webhook-token>" />
    </div>

    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-2">
      <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">How it works</p>
      <ul class="font-mono text-xs text-text-tertiary leading-relaxed space-y-1">
        <li>The Log Shipper connects to Fly's internal NATS log stream</li>
        <li>Logs are forwarded to Heimdall's webhook endpoint via HTTP</li>
        <li>Sub-second latency with at-least-once delivery</li>
        <li>Runs as a small Fly Machine in your org (~$2/month)</li>
      </ul>
    </div>
  </div>
</template>
