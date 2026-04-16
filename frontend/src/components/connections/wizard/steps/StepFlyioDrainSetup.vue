<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useConnectionsStore } from '@/stores/connections'
import type { WizardState } from '../flows'
import CopyableField from '@/components/common/CopyableField.vue'

const props = defineProps<{ modelValue: WizardState; connectionId?: string | null }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

const store = useConnectionsStore()

const connection = computed(() =>
  store.connections.find(c => c.id === props.connectionId) ?? undefined
)

const webhookUrl = computed(() => `${window.location.origin}/api/webhooks/logs`)

const webhookToken = computed(() => {
  const cfg = connection.value?.config as Record<string, unknown> | undefined
  return (cfg?.webhook_token as string) ?? ''
})

const secretsCmd = computed(() =>
  `fly secrets set HTTP_URL=${webhookUrl.value} HTTP_TOKEN=${webhookToken.value}`
)

// Drain setup is display-only — always valid.
onMounted(() => {
  emit('valid', true)
})
</script>

<template>
  <div class="space-y-5">
    <p class="font-mono text-xs text-text-secondary leading-relaxed">
      Your connection has been created. Now deploy the
      <span class="text-text-primary font-medium">Fly Log Shipper</span>
      to start pushing logs to Heimdall.
      Follow each step below in order.
    </p>

    <!-- Step 1: Launch -->
    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-3">
      <div class="flex items-center gap-2">
        <span class="flex items-center justify-center w-5 h-5 rounded-full bg-accent/15 text-accent font-mono text-[10px] font-bold shrink-0">1</span>
        <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Launch the Log Shipper</p>
      </div>
      <p class="font-mono text-[11px] text-text-tertiary">
        Create a new Fly app using the official Log Shipper image.
        Run this in your terminal from any directory.
      </p>
      <CopyableField value="fly launch --image flyio/log-shipper:latest --no-public-ips" />
      <p class="font-mono text-[11px] text-text-tertiary">
        When prompted, choose a name (e.g. <code class="text-text-secondary">my-app-logs</code>)
        and the same region as your main app.
        Answer <code class="text-text-secondary">no</code> to the deploy prompt — you'll deploy after setting secrets.
      </p>
    </div>

    <!-- Step 2: Set Fly access token -->
    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-3">
      <div class="flex items-center gap-2">
        <span class="flex items-center justify-center w-5 h-5 rounded-full bg-accent/15 text-accent font-mono text-[10px] font-bold shrink-0">2</span>
        <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Set Fly Access Token</p>
      </div>
      <p class="font-mono text-[11px] text-text-tertiary">
        The Log Shipper needs access to Fly's internal NATS log stream.
        This grants read-only access to your org's logs.
      </p>
      <CopyableField value="fly secrets set ACCESS_TOKEN=$(fly auth token)" />
    </div>

    <!-- Step 3: Set Heimdall secrets (real values) -->
    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-3">
      <div class="flex items-center gap-2">
        <span class="flex items-center justify-center w-5 h-5 rounded-full bg-accent/15 text-accent font-mono text-[10px] font-bold shrink-0">3</span>
        <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Point it at Heimdall</p>
      </div>
      <p class="font-mono text-[11px] text-text-tertiary">
        Tell the Log Shipper where to send logs. These values are specific to this connection — copy the command below as-is.
      </p>
      <CopyableField :value="secretsCmd" />
      <div class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 mt-1">
        <span class="font-mono text-[10px] uppercase tracking-wider text-text-muted">URL</span>
        <span class="font-mono text-[11px] text-text-secondary truncate">{{ webhookUrl }}</span>
        <span class="font-mono text-[10px] uppercase tracking-wider text-text-muted">Token</span>
        <span class="font-mono text-[11px] text-text-secondary truncate">{{ webhookToken || 'Generating...' }}</span>
      </div>
    </div>

    <!-- Step 4: Configure source_type -->
    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-3">
      <div class="flex items-center gap-2">
        <span class="flex items-center justify-center w-5 h-5 rounded-full bg-accent/15 text-accent font-mono text-[10px] font-bold shrink-0">4</span>
        <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Configure Vector</p>
      </div>
      <p class="font-mono text-[11px] text-text-tertiary">
        The Log Shipper uses <span class="text-text-secondary">Vector</span> internally to forward logs.
        Heimdall needs each log event to include a
        <code class="text-text-secondary">source_type</code> field starting with
        <code class="text-text-secondary">"fly"</code> so it can auto-detect Fly.io logs.
      </p>
      <p class="font-mono text-[11px] text-text-tertiary">
        <span class="text-text-primary font-medium">If you're using the stock Log Shipper image</span>,
        the <code class="text-text-secondary">fly</code> metadata object is included by default — no extra config needed.
      </p>
      <p class="font-mono text-[11px] text-text-tertiary">
        <span class="text-text-primary font-medium">If you have a custom Vector config</span>,
        add this transform before your HTTP sink to tag events:
      </p>
      <div class="bg-bg-elevated/60 rounded px-3 py-2 border border-border">
        <pre class="font-mono text-xs text-text-secondary leading-relaxed overflow-x-auto">[transforms.tag_for_heimdall]
type = "remap"
inputs = ["your_source"]
source = '.source_type = "fly_app"'</pre>
      </div>
    </div>

    <!-- Step 5: Deploy -->
    <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-3">
      <div class="flex items-center gap-2">
        <span class="flex items-center justify-center w-5 h-5 rounded-full bg-accent/15 text-accent font-mono text-[10px] font-bold shrink-0">5</span>
        <p class="font-mono text-xs font-medium uppercase tracking-widest text-text-muted">Deploy</p>
      </div>
      <p class="font-mono text-[11px] text-text-tertiary">
        Deploy the Log Shipper. Logs should start appearing in Heimdall within seconds.
      </p>
      <CopyableField value="fly deploy" />
    </div>

    <!-- Reference: What Heimdall extracts -->
    <details class="border border-border rounded bg-bg-elevated/40">
      <summary class="px-4 py-3 font-mono text-xs font-medium uppercase tracking-widest text-text-muted cursor-pointer hover:text-text-secondary transition-colors select-none">
        Reference — Expected Payload Format
      </summary>
      <div class="px-4 pb-3 space-y-2">
        <p class="font-mono text-[11px] text-text-tertiary">
          Heimdall auto-extracts these fields from each log event:
        </p>
        <div class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1">
          <code class="font-mono text-[11px] text-text-secondary">fly.app.name</code>
          <span class="font-mono text-[11px] text-text-tertiary">App name (used for source labelling)</span>
          <code class="font-mono text-[11px] text-text-secondary">fly.machine.id</code>
          <span class="font-mono text-[11px] text-text-tertiary">Machine ID</span>
          <code class="font-mono text-[11px] text-text-secondary">fly.region</code>
          <span class="font-mono text-[11px] text-text-tertiary">Deployment region</span>
          <code class="font-mono text-[11px] text-text-secondary">log.level</code>
          <span class="font-mono text-[11px] text-text-tertiary">Severity (info, warning, error, etc.)</span>
          <code class="font-mono text-[11px] text-text-secondary">message</code>
          <span class="font-mono text-[11px] text-text-tertiary">Log message body</span>
          <code class="font-mono text-[11px] text-text-secondary">timestamp</code>
          <span class="font-mono text-[11px] text-text-tertiary">Event timestamp</span>
        </div>
      </div>
    </details>

    <!-- Cost note -->
    <p class="font-mono text-[11px] text-text-muted leading-relaxed">
      The Log Shipper runs as a small Fly Machine (~$2/month).
      Sub-second latency with at-least-once delivery.
    </p>
  </div>
</template>
