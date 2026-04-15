import { ref } from 'vue'

export interface Toast {
  id: number
  message: string
  type: 'success' | 'error' | 'info'
  duration: number
}

// Module-level state — shared singleton across all components
const toasts = ref<Toast[]>([])
const timers = new Map<number, ReturnType<typeof setTimeout>>()
let nextId = 0

export function useToast() {
  function show(message: string, type: Toast['type'] = 'info', duration = 4000) {
    const id = nextId++
    toasts.value.push({ id, message, type, duration })

    if (duration > 0) {
      timers.set(id, setTimeout(() => dismiss(id), duration))
    }
  }

  function dismiss(id: number) {
    // Clear any pending auto-dismiss timer so it doesn't fire after removal.
    const timer = timers.get(id)
    if (timer) {
      clearTimeout(timer)
      timers.delete(id)
    }
    const idx = toasts.value.findIndex((t) => t.id === id)
    if (idx !== -1) toasts.value.splice(idx, 1)
  }

  return { toasts, show, dismiss }
}
