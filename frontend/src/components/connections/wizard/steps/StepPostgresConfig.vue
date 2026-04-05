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
const host = ref((initCfg.host as string) ?? '')
const port = ref(String(initCfg.port ?? '5432'))
const database = ref((initCfg.database as string) ?? '')
const user = ref((initCfg.user as string) ?? '')
const password = ref((initCfg.password as string) ?? '')
const sslMode = ref((initCfg.ssl_mode as string) ?? 'require')

// Sync inward when parent resets state (e.g. platform re-selection).
watch(() => props.modelValue.config, (cfg) => {
  const newHost = (cfg.host as string) ?? ''
  const newPort = String(cfg.port ?? '5432')
  const newDb = (cfg.database as string) ?? ''
  const newUser = (cfg.user as string) ?? ''
  const newPw = (cfg.password as string) ?? ''
  const newSsl = (cfg.ssl_mode as string) ?? 'require'
  if (newHost !== host.value) host.value = newHost
  if (newPort !== port.value) port.value = newPort
  if (newDb !== database.value) database.value = newDb
  if (newUser !== user.value) user.value = newUser
  if (newPw !== password.value) password.value = newPw
  if (newSsl !== sslMode.value) sslMode.value = newSsl
}, { deep: true })

function sync() {
  const parsed = parseInt(port.value, 10)
  const portNum = Number.isNaN(parsed) ? 5432 : Math.max(1, Math.min(65535, parsed))
  emit('update:modelValue', {
    ...props.modelValue,
    config: {
      ...props.modelValue.config,
      host: host.value,
      port: portNum,
      database: database.value,
      user: user.value,
      password: password.value,
      ssl_mode: sslMode.value,
    },
  })
  const portValid = !Number.isNaN(parsed) && parsed >= 1 && parsed <= 65535
  emit('valid', host.value.trim().length > 0 && database.value.trim().length > 0 && user.value.trim().length > 0 && portValid)
}

watch([host, port, database, user, password, sslMode], sync)
onMounted(sync)
</script>

<template>
  <div class="space-y-4">
    <div class="grid grid-cols-2 gap-4">
      <div class="col-span-2 sm:col-span-1">
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Host</label>
        <input v-model="host" type="text" placeholder="db.example.com" autofocus
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors" />
      </div>
      <div class="col-span-2 sm:col-span-1">
        <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Port</label>
        <input v-model="port" type="text" inputmode="numeric" placeholder="5432"
          class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors" />
      </div>
    </div>
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Database</label>
      <input v-model="database" type="text" placeholder="mydb"
        class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors" />
    </div>
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Username</label>
      <input v-model="user" type="text" placeholder="postgres"
        class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors" />
    </div>
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Password</label>
      <input v-model="password" type="password" placeholder="••••••••"
        class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent focus:ring-1 focus:ring-accent/30 focus:outline-none transition-colors" />
    </div>
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">SSL Mode</label>
      <BaseSelect
        :modelValue="sslMode"
        @update:modelValue="sslMode = String($event)"
        :options="[
          { value: 'disable', label: 'Disable' },
          { value: 'require', label: 'Require' },
          { value: 'verify-full', label: 'Verify Full' },
        ]"
      />
    </div>
  </div>
</template>
