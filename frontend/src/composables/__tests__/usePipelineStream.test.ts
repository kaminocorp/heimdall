import { defineComponent, h, nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { vi } from 'vitest'
import client from '@/api/client'
import { usePipelineStream } from '@/composables/usePipelineStream'
import { usePipelineStore } from '@/stores/pipeline'
import type { PipelineBootstrap } from '@/types/pipeline'

// Minimal EventSource stub. Captures handlers so the test can drive frames.
class StubEventSource {
  static instances: StubEventSource[] = []
  url: string
  onopen: (() => void) | null = null
  onerror: (() => void) | null = null
  listeners: Record<string, ((e: { data: string }) => void)[]> = {}
  closed = false

  constructor(url: string) {
    this.url = url
    StubEventSource.instances.push(this)
  }
  addEventListener(name: string, fn: (e: { data: string }) => void) {
    ;(this.listeners[name] ||= []).push(fn)
  }
  emit(name: string, data: unknown) {
    for (const fn of this.listeners[name] ?? []) {
      fn({ data: JSON.stringify(data) })
    }
  }
  close() {
    this.closed = true
  }
}

const TestHarness = defineComponent({
  props: { appId: { type: String, required: true } },
  setup(props) {
    const handle = usePipelineStream({ appId: props.appId })
    // The composable doesn't auto-start; the page calls .start() on mount.
    void handle.start()
    return () => h('div', handle.connectionState.value)
  },
})

describe('usePipelineStream', () => {
  let originalEventSource: typeof globalThis.EventSource
  beforeEach(() => {
    StubEventSource.instances = []
    originalEventSource = globalThis.EventSource
    // Cast through unknown to satisfy the structural EventSource type; the
    // composable only uses .close, .onopen, .onerror, and addEventListener.
    globalThis.EventSource = StubEventSource as unknown as typeof EventSource
    vi.mocked(client.get).mockReset()
  })
  afterEach(() => {
    globalThis.EventSource = originalEventSource
  })

  function bootstrapFixture(): PipelineBootstrap {
    return {
      stats: {
        ingestion_count: 12,
        classified_count: 12,
        flagged_count: 4,
        safe_count: 8,
        assessment_count: 3,
        avg_confidence: 0.6,
        window_seconds: 3600,
      },
      recent_events: [
        {
          id: 'evt-1',
          log_id: 'log-1',
          app_id: 'app-1',
          stage: 'ingestion',
          occurred_at: '2026-04-18T10:00:00Z',
          source_type: 'fly',
          severity: 'info',
        },
      ],
      cursor: '2026-04-18T10:00:00Z',
    }
  }

  it('bootstraps stats and events into the store', async () => {
    vi.mocked(client.get).mockResolvedValueOnce({ data: bootstrapFixture() })

    const wrapper = mount(TestHarness, { props: { appId: 'app-1' } })
    // Wait one microtask so the bootstrap promise settles before we assert.
    await new Promise((r) => setTimeout(r, 0))
    await nextTick()

    const store = usePipelineStore()
    expect(store.stats.ingestion_count).toBe(12)
    expect(store.events).toHaveLength(1)
    expect(StubEventSource.instances).toHaveLength(1)
    // The cursor should appear in the stream URL as ?since= so the SSE
    // catch-up replay knows where to start.
    expect(StubEventSource.instances[0].url).toContain('since=2026-04-18T10%3A00%3A00Z')

    wrapper.unmount()
    expect(StubEventSource.instances[0].closed).toBe(true)
  })

  it('flips to streaming on first live frame and dedupes against bootstrap', async () => {
    vi.mocked(client.get).mockResolvedValueOnce({ data: bootstrapFixture() })

    const wrapper = mount(TestHarness, { props: { appId: 'app-1' } })
    await new Promise((r) => setTimeout(r, 0))
    await nextTick()

    const es = StubEventSource.instances[0]
    // Re-deliver the bootstrap event as a replay frame — store.insertEvent
    // must reject the duplicate.
    es.emit('replay', {
      id: 'evt-1',
      log_id: 'log-1',
      app_id: 'app-1',
      stage: 'ingestion',
      occurred_at: '2026-04-18T10:00:00Z',
    })
    es.emit('live', {
      id: 'evt-2',
      log_id: 'log-2',
      app_id: 'app-1',
      stage: 'classified',
      occurred_at: '2026-04-18T10:00:01Z',
      type: 'request',
      category: 'server_error',
    })
    await nextTick()

    const store = usePipelineStore()
    expect(store.events).toHaveLength(2)
    // Newest first.
    expect(store.events[0].id).toBe('evt-2')
    expect(wrapper.text()).toBe('streaming')

    wrapper.unmount()
  })
})
