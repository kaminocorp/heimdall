import { ref, onUnmounted } from 'vue'
import type { ConnectionState, PipelineEvent } from '@/types/pipeline'
import { fetchPipelineBootstrap, type BootstrapOpts } from '@/api/pipeline'
import { usePipelineStore } from '@/stores/pipeline'
import { useAuthStore } from '@/stores/auth'

const RECONNECT_BASE_MS = 1000
const RECONNECT_MAX_MS = 30000

export interface PipelineStreamOptions extends BootstrapOpts {
  appId: string
}

export function usePipelineStream(opts: PipelineStreamOptions) {
  const store = usePipelineStore()
  const auth = useAuthStore()

  const connectionState = ref<ConnectionState>('idle')
  const lastError = ref<string | null>(null)

  let cursor = ''
  let source: EventSource | null = null
  let reconnectAttempts = 0
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let destroyed = false

  function buildStreamUrl(): string {
    const u = new URL(`/api/apps/${opts.appId}/pipeline/stream`, location.origin)
    if (cursor) u.searchParams.set('since', cursor)
    if (auth.token) u.searchParams.set('token', auth.token)
    return u.toString()
  }

  function handleEvent(name: 'replay' | 'live', raw: string) {
    try {
      const evt = JSON.parse(raw) as PipelineEvent
      const inserted = store.insertEvent(evt)
      // Advance cursor on every accepted event so reconnects don't replay
      // anything we've already seen. Events arrive monotonic from the server
      // (replay sorted ASC, live by occurrence), so a simple max is enough.
      if (inserted && evt.occurred_at > cursor) {
        cursor = evt.occurred_at
      }
      // Suppress unused-var linter on `name` — kept for future debug logging.
      void name
    } catch (e) {
      console.warn('pipeline: failed to parse', name, 'frame:', e)
    }
  }

  async function handleResync() {
    // Server hit replay cap — our buffer is too stale to be trusted. Drop
    // everything and re-bootstrap from scratch.
    connectionState.value = 'resyncing'
    store.reset()
    cursor = ''
    closeSource()
    await bootstrap()
    openStream()
  }

  function openStream() {
    if (destroyed) return
    closeSource()
    source = new EventSource(buildStreamUrl())

    source.addEventListener('replay', (e) => handleEvent('replay', (e as MessageEvent).data))
    source.addEventListener('live', (e) => {
      if (connectionState.value !== 'streaming') {
        connectionState.value = 'streaming'
        reconnectAttempts = 0
      }
      handleEvent('live', (e as MessageEvent).data)
    })
    source.addEventListener('resync', () => {
      void handleResync()
    })

    source.onopen = () => {
      // EventSource fires onopen once handshake completes; we still wait for
      // the first frame (replay or live) to flip to 'streaming' so the UI
      // doesn't claim live until data is actually flowing.
      reconnectAttempts = 0
      lastError.value = null
    }

    source.onerror = () => {
      // EventSource auto-reconnects, but its built-in retry doesn't refresh
      // the URL — meaning a token rotation mid-stream would loop on 401.
      // Tear down and reconnect manually so a fresh token is picked up.
      connectionState.value = 'reconnecting'
      scheduleReconnect()
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer || destroyed) return
    closeSource()
    const delay = Math.min(
      RECONNECT_BASE_MS * Math.pow(2, reconnectAttempts),
      RECONNECT_MAX_MS,
    )
    reconnectAttempts++
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      openStream()
    }, delay)
  }

  function closeSource() {
    if (source) {
      source.close()
      source = null
    }
  }

  async function bootstrap() {
    connectionState.value = 'bootstrapping'
    try {
      const data = await fetchPipelineBootstrap(opts.appId, {
        window: opts.window,
        tickerLimit: opts.tickerLimit,
      })
      store.setStats(data.stats)
      store.mergeEvents(data.recent_events)
      cursor = data.cursor
      lastError.value = null
    } catch (e) {
      connectionState.value = 'error'
      lastError.value = e instanceof Error ? e.message : 'failed to load pipeline'
      throw e
    }
  }

  async function start() {
    destroyed = false
    store.startRates()
    await bootstrap()
    openStream()
  }

  function destroy() {
    destroyed = true
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    closeSource()
    store.stopRates()
    connectionState.value = 'idle'
  }

  function reconnect() {
    reconnectAttempts = 0
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    openStream()
  }

  onUnmounted(() => {
    destroy()
  })

  return {
    connectionState,
    lastError,
    start,
    destroy,
    reconnect,
  }
}
