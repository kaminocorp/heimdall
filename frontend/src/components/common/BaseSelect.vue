<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'

export interface SelectOption {
  value: string | number
  label: string
}

const props = withDefaults(defineProps<{
  modelValue: string | number
  options: SelectOption[]
  disabled?: boolean
  size?: 'sm' | 'default'
}>(), {
  disabled: false,
  size: 'default',
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
}>()

const open = ref(false)
const root = ref<HTMLElement | null>(null)
const listRef = ref<HTMLElement | null>(null)
const focusedIndex = ref(-1)

const selectedLabel = computed(() => {
  const match = props.options.find(o => o.value === props.modelValue)
  return match?.label ?? ''
})

function toggle() {
  if (props.disabled) return
  open.value = !open.value
  if (open.value) {
    focusedIndex.value = props.options.findIndex(o => o.value === props.modelValue)
  }
}

function select(option: SelectOption) {
  emit('update:modelValue', option.value)
  open.value = false
}

function onClickOutside(e: MouseEvent) {
  if (root.value && !root.value.contains(e.target as Node)) {
    open.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
  if (!open.value) {
    if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
      e.preventDefault()
      toggle()
    }
    return
  }

  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      focusedIndex.value = Math.min(focusedIndex.value + 1, props.options.length - 1)
      scrollToFocused()
      break
    case 'ArrowUp':
      e.preventDefault()
      focusedIndex.value = Math.max(focusedIndex.value - 1, 0)
      scrollToFocused()
      break
    case 'Enter':
    case ' ':
      e.preventDefault()
      if (focusedIndex.value >= 0) {
        select(props.options[focusedIndex.value])
      }
      break
    case 'Escape':
      e.preventDefault()
      open.value = false
      break
  }
}

function scrollToFocused() {
  const list = listRef.value
  if (!list) return
  const item = list.children[focusedIndex.value] as HTMLElement | undefined
  item?.scrollIntoView({ block: 'nearest' })
}

onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))

watch(open, (isOpen) => {
  if (!isOpen) focusedIndex.value = -1
})
</script>

<template>
  <div ref="root" class="relative" :class="disabled ? 'opacity-50 cursor-not-allowed' : ''">
    <!-- Trigger -->
    <button
      type="button"
      @click="toggle"
      @keydown="onKeydown"
      :disabled="disabled"
      class="w-full flex items-center justify-between gap-2 bg-bg-elevated/80 border rounded font-mono text-text-primary text-left transition-colors cursor-pointer disabled:cursor-not-allowed"
      :class="[
        open ? 'border-accent/50 ring-1 ring-accent/20' : 'border-border hover:border-border-hover',
        size === 'sm' ? 'px-3 py-1.5 text-xs' : 'px-3 py-2 text-sm',
      ]"
    >
      <span class="truncate">{{ selectedLabel }}</span>
      <svg
        class="w-3 h-3 text-text-muted flex-shrink-0 transition-transform duration-150"
        :class="open ? 'rotate-180' : ''"
        viewBox="0 0 12 12"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
      >
        <path d="M3 4.5L6 7.5L9 4.5" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>

    <!-- Dropdown panel -->
    <Transition
      enter-active-class="transition duration-100 ease-out"
      enter-from-class="opacity-0 -translate-y-1"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-75 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-1"
    >
      <div
        v-if="open"
        ref="listRef"
        class="absolute z-50 mt-1 w-full max-h-48 overflow-y-auto border border-border rounded bg-bg-elevated shadow-lg shadow-black/30"
      >
        <button
          v-for="(option, i) in options"
          :key="option.value"
          type="button"
          @click="select(option)"
          @mouseenter="focusedIndex = i"
          class="w-full text-left font-mono transition-colors cursor-pointer"
          :class="[
            size === 'sm' ? 'px-3 py-1.5 text-xs' : 'px-3 py-2 text-sm',
            option.value === modelValue
              ? 'text-accent-bright bg-accent-subtle'
              : focusedIndex === i
                ? 'text-text-primary bg-bg-surface-hover'
                : 'text-text-secondary hover:text-text-primary hover:bg-bg-surface',
          ]"
        >
          {{ option.label }}
        </button>
      </div>
    </Transition>
  </div>
</template>
