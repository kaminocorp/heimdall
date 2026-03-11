<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { Connection, CreateConnectionPayload } from '@/types/connection'
import { useAppStore } from '@/stores/app'
import { getGitHubInstallURL } from '@/api/github'

const appStore = useAppStore()

interface FieldDef {
  key: string
  label: string
  type: 'text' | 'password' | 'number' | 'select'
  required: boolean
  placeholder?: string
  default?: string | number
  options?: { value: string; label: string }[]
}

const configFields: Record<string, FieldDef[]> = {
  postgres: [
    { key: 'host', label: 'Host', type: 'text', required: true, placeholder: 'db.example.com' },
    { key: 'port', label: 'Port', type: 'number', required: false, default: 5432, placeholder: '5432' },
    { key: 'database', label: 'Database', type: 'text', required: true, placeholder: 'mydb' },
    { key: 'user', label: 'Username', type: 'text', required: true, placeholder: 'postgres' },
    { key: 'password', label: 'Password', type: 'password', required: true, placeholder: '••••••••' },
    { key: 'ssl_mode', label: 'SSL Mode', type: 'select', required: false, default: 'require', options: [
      { value: 'disable', label: 'Disable' },
      { value: 'require', label: 'Require' },
      { value: 'verify-full', label: 'Verify Full' },
    ]},
  ],
  webhook_logs: [],
  syslog: [
    { key: 'host', label: 'Host', type: 'text', required: true, placeholder: 'syslog.example.com' },
    { key: 'port', label: 'Port', type: 'number', required: false, default: 514, placeholder: '514' },
    { key: 'protocol', label: 'Protocol', type: 'select', required: false, default: 'udp', options: [
      { value: 'udp', label: 'UDP' },
      { value: 'tcp', label: 'TCP' },
    ]},
  ],
  github: [],
}

const githubInstalling = ref(false)
const githubError = ref<string | null>(null)

async function installGitHubApp() {
  if (!appStore.currentAppId) return
  githubInstalling.value = true
  githubError.value = null
  try {
    const { url } = await getGitHubInstallURL(appStore.currentAppId)
    window.location.href = url
  } catch (e: any) {
    githubError.value = e.response?.data?.error ?? 'Failed to start GitHub App installation'
    githubInstalling.value = false
  }
}

const props = defineProps<{
  initialValues?: Connection | null
}>()

const isEditing = computed(() => !!props.initialValues)

const name = ref('')
const type = ref('postgres')
const direction = ref<'one_way' | 'two_way'>('one_way')
const config = ref<Record<string, string | number>>({})

const activeFields = computed(() => configFields[type.value] ?? [])

watch(type, (_newType, oldType) => {
  if (oldType !== undefined) {
    config.value = {}
  }
})

watch(() => props.initialValues, (conn) => {
  if (conn) {
    name.value = conn.name
    type.value = conn.type
    direction.value = conn.direction
    config.value = { ...conn.config } as Record<string, string | number>
  } else {
    name.value = ''
    type.value = 'postgres'
    direction.value = 'one_way'
    config.value = {}
  }
}, { immediate: true })

function getFieldValue(field: FieldDef): string | number {
  return config.value[field.key] ?? field.default ?? ''
}

function setFieldValue(field: FieldDef, value: string) {
  if (field.type === 'number') {
    config.value[field.key] = value === '' ? '' as unknown as number : Number(value)
  } else {
    config.value[field.key] = value
  }
}

const emit = defineEmits<{
  submit: [data: Omit<CreateConnectionPayload, 'app_id'>]
  cancel: []
}>()

function handleSubmit() {
  const builtConfig: Record<string, unknown> = {}
  for (const field of activeFields.value) {
    const val = config.value[field.key] ?? field.default
    if (val !== undefined && val !== '') {
      builtConfig[field.key] = val
    }
  }

  emit('submit', {
    name: name.value,
    type: type.value,
    direction: direction.value,
    config: builtConfig,
  })

  if (!isEditing.value) {
    name.value = ''
    type.value = 'postgres'
    direction.value = 'one_way'
    config.value = {}
  }
}
</script>

<template>
  <form @submit.prevent="handleSubmit" class="space-y-4 border border-border rounded-lg p-5 bg-bg-surface">
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Name</label>
      <input v-model="name" type="text" required placeholder="e.g. Production DB"
        class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors" />
    </div>
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Type</label>
      <select v-model="type" :disabled="isEditing" class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
        <option value="postgres">PostgreSQL</option>
        <option value="webhook_logs">Webhook Logs</option>
        <option value="syslog">Syslog</option>
        <option value="github">GitHub</option>
      </select>
    </div>
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Direction</label>
      <select v-model="direction" class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors">
        <option value="one_way">One-way (ingest only)</option>
        <option value="two_way">Two-way (ingest + query)</option>
      </select>
    </div>
    <!-- Config fields -->
    <template v-if="type === 'github' && !isEditing">
      <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40 space-y-3">
        <p class="font-mono text-xs text-text-secondary">
          Heimdall uses a GitHub App for secure, org-scoped access to your repositories. Click below to install the app on your GitHub organization.
        </p>
        <div v-if="githubError" class="text-sm font-mono text-status-critical">{{ githubError }}</div>
        <button
          type="button"
          :disabled="githubInstalling"
          @click="installGitHubApp"
          class="px-4 py-2 bg-accent text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-accent-hover transition-colors cursor-pointer disabled:opacity-50"
        >
          {{ githubInstalling ? 'Redirecting...' : 'Install GitHub App' }}
        </button>
      </div>
    </template>
    <template v-else-if="type === 'webhook_logs'">
      <div class="border border-border rounded px-4 py-3 bg-bg-elevated/40">
        <p class="font-mono text-xs text-text-secondary">
          A webhook token will be generated automatically when this connection is created.
        </p>
      </div>
    </template>
    <template v-else-if="activeFields.length > 0">
      <div class="border-t border-border pt-4 mt-2 space-y-4">
        <p class="font-mono text-[10px] font-medium uppercase tracking-widest text-text-muted">Configuration</p>
        <div v-for="field in activeFields" :key="field.key">
          <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">{{ field.label }}</label>
          <select v-if="field.type === 'select'" :value="getFieldValue(field)" @change="setFieldValue(field, ($event.target as HTMLSelectElement).value)"
            class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors">
            <option v-for="opt in field.options" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>
          <input v-else :type="field.type === 'number' ? 'text' : field.type" :value="getFieldValue(field)" @input="setFieldValue(field, ($event.target as HTMLInputElement).value)"
            :required="field.required" :placeholder="field.placeholder"
            :inputmode="field.type === 'number' ? 'numeric' : undefined"
            class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors" />
        </div>
      </div>
    </template>

    <div v-if="!(type === 'github' && !isEditing)" class="flex gap-2 pt-2">
      <button type="submit" class="px-4 py-2 bg-accent text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-accent-hover transition-colors cursor-pointer">
        {{ isEditing ? 'Save Changes' : 'Add Connection' }}
      </button>
      <button type="button" @click="emit('cancel')" class="px-4 py-2 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer">
        Cancel
      </button>
    </div>
    <div v-else class="flex gap-2 pt-2">
      <button type="button" @click="emit('cancel')" class="px-4 py-2 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer">
        Cancel
      </button>
    </div>
  </form>
</template>
