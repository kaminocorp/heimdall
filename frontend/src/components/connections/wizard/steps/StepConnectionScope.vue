<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import type { WizardState } from '../flows'

// Scope step: pick whether this connection lives inside one app or spans
// every app in the org. The choice is persisted on state.config under
// `__scope` (a wizard-only key, stripped before sending to the backend —
// see ConnectionWizard.createConnection).
//
// Neither option is marked "recommended"; both are first-class. Default is
// "app" because it matches Phase 1 and is the safer, tighter-scoped choice.

const props = defineProps<{ modelValue: WizardState }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

const appStore = useAppStore()

const currentAppName = computed(() => appStore.currentApp?.name ?? 'this app')
const currentOrgName = computed(() => appStore.organization?.name ?? 'your organisation')

// Initialise from whatever's already in config (so Back preserves selection).
const scope = ref<'app' | 'org'>(
  (props.modelValue.config.__scope as 'app' | 'org' | undefined) ?? 'app',
)

function select(choice: 'app' | 'org') {
  scope.value = choice
  emit('update:modelValue', {
    ...props.modelValue,
    config: { ...props.modelValue.config, __scope: choice },
  })
}

onMounted(() => {
  // Seed the initial choice into config so Continue works immediately.
  select(scope.value)
  emit('valid', true)
})
</script>

<template>
  <div class="space-y-4">
    <p class="font-mono text-xs text-text-secondary leading-relaxed">
      Where should this connection be available? Both options are fully
      supported — pick whichever matches how your log source is configured.
    </p>

    <button
      type="button"
      class="w-full border rounded-lg px-4 py-3 text-left transition-colors cursor-pointer"
      :class="scope === 'app'
        ? 'border-accent bg-accent/10'
        : 'border-border hover:border-border-hover bg-bg-elevated/40'"
      @click="select('app')"
    >
      <div class="flex items-start gap-3">
        <span
          class="w-4 h-4 mt-1 rounded-full border flex items-center justify-center shrink-0"
          :class="scope === 'app' ? 'border-accent' : 'border-border'"
        >
          <span v-if="scope === 'app'" class="w-2 h-2 rounded-full bg-accent" />
        </span>
        <div class="min-w-0">
          <p class="font-mono text-sm text-text-primary font-medium">
            This app only — {{ currentAppName }}
          </p>
          <p class="font-mono text-[11px] text-text-muted mt-1">
            Logs stay scoped to this application. Use this when your log
            source only serves one of your apps.
          </p>
        </div>
      </div>
    </button>

    <button
      type="button"
      class="w-full border rounded-lg px-4 py-3 text-left transition-colors cursor-pointer"
      :class="scope === 'org'
        ? 'border-accent bg-accent/10'
        : 'border-border hover:border-border-hover bg-bg-elevated/40'"
      @click="select('org')"
    >
      <div class="flex items-start gap-3">
        <span
          class="w-4 h-4 mt-1 rounded-full border flex items-center justify-center shrink-0"
          :class="scope === 'org' ? 'border-accent' : 'border-border'"
        >
          <span v-if="scope === 'org'" class="w-2 h-2 rounded-full bg-accent" />
        </span>
        <div class="min-w-0">
          <p class="font-mono text-sm text-text-primary font-medium">
            Entire organisation — {{ currentOrgName }}
          </p>
          <p class="font-mono text-[11px] text-text-muted mt-1">
            Available to every app in the org. Each app then picks which
            sources (e.g. individual Fly apps) to accept.
          </p>
        </div>
      </div>
    </button>
  </div>
</template>
