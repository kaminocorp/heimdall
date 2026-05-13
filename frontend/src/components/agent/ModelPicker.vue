<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import type { ModelOption } from '@/types/models'
import ModelCard from './ModelCard.vue'

const props = defineProps<{
  modelValue: string
  models: ModelOption[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const open = ref(false)
const search = ref('')
const tierFilter = ref<string>('all')
const focusedIndex = ref(-1)

const root = ref<HTMLElement | null>(null)
const listRef = ref<HTMLElement | null>(null)
const searchRef = ref<HTMLInputElement | null>(null)

const tiers = [
  { value: 'all', label: 'All' },
  { value: 'flagship', label: 'Flagship' },
  { value: 'balanced', label: 'Balanced' },
  { value: 'economy', label: 'Economy' },
  { value: 'specialist', label: 'Specialist' },
]

// ── Filtering ──────────────────────────────────────────────────────────

const filteredModels = computed(() => {
  let result = props.models

  // Tier filter
  if (tierFilter.value !== 'all') {
    result = result.filter(m => m.tier === tierFilter.value)
  }

  // Search filter — matches across name, id, vendor, strengths
  if (search.value.trim()) {
    const q = search.value.toLowerCase().trim()
    result = result.filter(m =>
      m.name.toLowerCase().includes(q) ||
      m.id.toLowerCase().includes(q) ||
      m.vendor.toLowerCase().includes(q) ||
      m.strengths.some(s => s.toLowerCase().includes(q))
    )
  }

  return result
})

// Group filtered models by vendor for display
const groupedModels = computed(() => {
  const groups: { vendor: string; models: ModelOption[] }[] = []
  const vendorMap = new Map<string, ModelOption[]>()

  for (const m of filteredModels.value) {
    const existing = vendorMap.get(m.vendor)
    if (existing) {
      existing.push(m)
    } else {
      const arr = [m]
      vendorMap.set(m.vendor, arr)
      groups.push({ vendor: m.vendor, models: arr })
    }
  }

  return groups
})

// Flat list for keyboard navigation (preserves render order)
const flatList = computed(() => filteredModels.value)

// ── Selected model info for trigger ────────────────────────────────────

const selectedModel = computed(() =>
  props.models.find(m => m.id === props.modelValue)
)

// ── Open / close ───────────────────────────────────────────────────────

function openPicker() {
  open.value = true
  search.value = ''
  tierFilter.value = 'all'
  focusedIndex.value = -1
  nextTick(() => searchRef.value?.focus())
}

function closePicker() {
  open.value = false
  focusedIndex.value = -1
}

function toggle() {
  if (open.value) closePicker()
  else openPicker()
}

function selectModel(id: string) {
  emit('update:modelValue', id)
  closePicker()
}

// ── Click outside ──────────────────────────────────────────────────────

function onClickOutside(e: MouseEvent) {
  if (root.value && !root.value.contains(e.target as Node)) {
    closePicker()
  }
}

onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))

// ── Keyboard ───────────────────────────────────────────────────────────

function onKeydown(e: KeyboardEvent) {
  if (!open.value) {
    if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
      e.preventDefault()
      openPicker()
    }
    return
  }

  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      focusedIndex.value = Math.min(focusedIndex.value + 1, flatList.value.length - 1)
      scrollToFocused()
      break
    case 'ArrowUp':
      e.preventDefault()
      focusedIndex.value = Math.max(focusedIndex.value - 1, 0)
      scrollToFocused()
      break
    case 'Enter':
      e.preventDefault()
      if (focusedIndex.value >= 0 && focusedIndex.value < flatList.value.length) {
        selectModel(flatList.value[focusedIndex.value].id)
      }
      break
    case 'Escape':
      e.preventDefault()
      closePicker()
      break
  }
}

function onSearchKeydown(e: KeyboardEvent) {
  // Let arrow keys and Enter/Escape bubble to the root handler
  if (['ArrowDown', 'ArrowUp', 'Enter', 'Escape'].includes(e.key)) {
    return
  }
  // All other keys are for typing in the search — stop propagation
  e.stopPropagation()
}

function scrollToFocused() {
  const list = listRef.value
  if (!list) return
  // Find the card element by data attribute
  const item = list.querySelector(`[data-index="${focusedIndex.value}"]`) as HTMLElement | undefined
  item?.scrollIntoView({ block: 'nearest' })
}

// Reset focus when filters change
watch([search, tierFilter], () => {
  focusedIndex.value = -1
})
</script>

<template>
  <div ref="root" class="relative" @keydown="onKeydown">
    <!-- Trigger button -->
    <button
      type="button"
      @click="toggle"
      class="w-full flex items-center justify-between gap-2 bg-bg-elevated/80 border rounded font-mono text-left transition-colors cursor-pointer overflow-hidden"
      :class="open ? 'border-accent/50 ring-1 ring-accent/20' : 'border-border hover:border-border-hover'"
    >
      <div v-if="selectedModel" class="flex items-center gap-2 px-3 py-2 min-w-0 overflow-hidden">
        <span class="text-sm text-text-primary truncate">{{ selectedModel.name }}</span>
        <span
          class="flex-shrink-0 font-mono text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded-full border border-accent-bright/30 text-accent-bright bg-accent-bright/10"
        >{{ selectedModel.tier }}</span>
        <span class="flex-shrink-0 text-[11px] text-text-muted">
          ${{ selectedModel.pricing.prompt }}/${{ selectedModel.pricing.completion }}
        </span>
      </div>
      <div v-else class="px-3 py-2 text-sm text-text-muted">
        Select a model...
      </div>
      <svg
        class="w-3 h-3 text-text-muted flex-shrink-0 mr-3 transition-transform duration-150"
        :class="open ? 'rotate-180' : ''"
        viewBox="0 0 12 12"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
      >
        <path d="M3 4.5L6 7.5L9 4.5" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>

    <!-- Picker panel -->
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
        class="absolute z-50 mt-1 w-full border border-border rounded bg-bg-elevated shadow-lg shadow-black/30 flex flex-col"
        style="max-height: min(70vh, 32rem); box-sizing: border-box;"
      >
        <!-- Search -->
        <div class="px-3 pt-3 pb-2 border-b border-border flex-shrink-0">
          <input
            ref="searchRef"
            v-model="search"
            @keydown="onSearchKeydown"
            type="text"
            placeholder="Search models..."
            class="w-full bg-bg-primary border border-border rounded px-2.5 py-1.5 text-text-primary font-mono text-xs placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors"
          />
        </div>

        <!-- Tier filter chips -->
        <div class="px-3 py-2 border-b border-border flex gap-1.5 flex-wrap flex-shrink-0">
          <button
            v-for="tier in tiers"
            :key="tier.value"
            type="button"
            @click="tierFilter = tier.value"
            class="px-2 py-0.5 rounded font-mono text-[11px] uppercase tracking-wider transition-colors cursor-pointer"
            :class="tierFilter === tier.value
              ? 'bg-accent-subtle text-text-primary border border-accent-bright/30'
              : 'text-text-muted hover:text-text-secondary border border-transparent'
            "
          >
            {{ tier.label }}
          </button>
        </div>

        <!-- Model list -->
        <div ref="listRef" class="overflow-y-auto flex-1" role="listbox">
          <template v-if="groupedModels.length">
            <div v-for="group in groupedModels" :key="group.vendor">
              <!-- Vendor header -->
              <div class="px-3 py-1.5 font-mono text-[10px] uppercase tracking-widest text-text-muted bg-bg-primary/50 sticky top-0">
                {{ group.vendor }}
              </div>
              <!-- Model cards -->
              <ModelCard
                v-for="model in group.models"
                :key="model.id"
                :model="model"
                :selected="model.id === modelValue"
                :focused="flatList.indexOf(model) === focusedIndex"
                :data-index="flatList.indexOf(model)"
                @select="selectModel"
                @mouseenter="focusedIndex = flatList.indexOf(model)"
              />
            </div>
          </template>

          <!-- Empty state -->
          <div v-else class="px-3 py-6 text-center">
            <p class="font-mono text-xs text-text-muted">No models match — try clearing filters.</p>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>
