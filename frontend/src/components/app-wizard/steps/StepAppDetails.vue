<script setup lang="ts">
import { ref, watch } from 'vue'

// Step 1 — collect the application name (required) and an optional
// description. The description is collected here so the UX feels complete,
// even though the backend schema doesn't persist a description field yet;
// if/when the backend grows that column, the wizard is already passing the
// value through. For now, we simply ignore it on submit.
const props = defineProps<{
  modelValue: { name: string; description: string }
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: { name: string; description: string }]
  valid: [isValid: boolean]
  // Emitted when the user hits Enter in the name input with a valid name.
  // The wizard wires this to `advanceFromDetails()`, matching the Continue
  // button's behaviour — keyboard-first users should never need the mouse
  // to advance a simple one-input form.
  submit: []
}>()

const name = ref(props.modelValue.name)
const description = ref(props.modelValue.description)

function push() {
  emit('update:modelValue', {
    name: name.value,
    description: description.value,
  })
  emit('valid', name.value.trim().length > 0)
}

function onNameEnter() {
  // Guard against empty submissions — we only fire `submit` when the
  // input is actually valid, so the wizard shell's receiver doesn't
  // need to re-check validity itself.
  if (name.value.trim().length > 0) {
    emit('submit')
  }
}

// Initial validity check so the Continue button reflects the current state
// the instant the step mounts (e.g. when the user Backs into this step with
// a pre-filled name).
push()

watch([name, description], push)
</script>

<template>
  <div>
    <p class="font-mono text-xs text-text-secondary mb-4">
      Give your new application a name. You can add connectors next, or skip and wire them up later.
    </p>

    <label class="block mb-4">
      <span class="block font-mono text-xs uppercase tracking-wider text-text-muted mb-1.5">
        Name <span class="text-status-critical">*</span>
      </span>
      <input
        v-model="name"
        type="text"
        placeholder="e.g. billing-service"
        :disabled="disabled"
        maxlength="100"
        @keydown.enter.prevent="onNameEnter"
        class="w-full px-3 py-2 bg-bg-elevated/80 border border-border rounded font-mono text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent/50 focus:ring-1 focus:ring-accent/20 transition-colors"
      />
    </label>

    <label class="block">
      <span class="block font-mono text-xs uppercase tracking-wider text-text-muted mb-1.5">
        Description <span class="text-text-muted">(optional)</span>
      </span>
      <textarea
        v-model="description"
        placeholder="What does this application do?"
        rows="3"
        :disabled="disabled"
        maxlength="500"
        class="w-full px-3 py-2 bg-bg-elevated/80 border border-border rounded font-mono text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent/50 focus:ring-1 focus:ring-accent/20 transition-colors resize-none"
      />
    </label>
  </div>
</template>
