import { usePipelineStore } from '@/stores/pipeline'
import type { PipelineEvent, PipelineStats } from '@/types/pipeline'

function makeEvent(overrides: Partial<PipelineEvent> = {}): PipelineEvent {
  return {
    id: overrides.id ?? crypto.randomUUID(),
    log_id: overrides.log_id ?? 'log-' + Math.random(),
    app_id: 'app-1',
    stage: overrides.stage ?? 'ingestion',
    occurred_at: overrides.occurred_at ?? new Date().toISOString(),
    ...overrides,
  }
}

describe('pipeline store', () => {
  it('insertEvent dedupes by id', () => {
    const store = usePipelineStore()
    const evt = makeEvent({ id: 'a' })
    expect(store.insertEvent(evt)).toBe(true)
    expect(store.insertEvent(evt)).toBe(false)
    expect(store.events).toHaveLength(1)
  })

  it('insertEvent prepends newest', () => {
    const store = usePipelineStore()
    const a = makeEvent({ id: 'a', occurred_at: '2026-04-18T10:00:00Z' })
    const b = makeEvent({ id: 'b', occurred_at: '2026-04-18T10:00:01Z' })
    store.insertEvent(a)
    store.insertEvent(b)
    // Newest at index 0 — the ticker reads top-down.
    expect(store.events[0].id).toBe('b')
    expect(store.events[1].id).toBe('a')
  })

  it('mergeEvents preserves newest-first order from a newest-first batch', () => {
    const store = usePipelineStore()
    // Bootstrap delivers newest-first; the store should end up with the same order.
    const batch: PipelineEvent[] = [
      makeEvent({ id: 'newest', occurred_at: '2026-04-18T12:00:02Z' }),
      makeEvent({ id: 'middle', occurred_at: '2026-04-18T12:00:01Z' }),
      makeEvent({ id: 'oldest', occurred_at: '2026-04-18T12:00:00Z' }),
    ]
    store.mergeEvents(batch)
    expect(store.events.map((e) => e.id)).toEqual(['newest', 'middle', 'oldest'])
  })

  it('flaggedRatio defaults to 0.5 when no gate decisions exist', () => {
    const store = usePipelineStore()
    expect(store.flaggedRatio).toBeCloseTo(0.5)
    expect(store.safeRatio).toBeCloseTo(0.5)
  })

  it('flaggedRatio reflects stats counts', () => {
    const store = usePipelineStore()
    const stats: PipelineStats = {
      ingestion_count: 100,
      classified_count: 100,
      flagged_count: 30,
      safe_count: 70,
      assessment_count: 25,
      avg_confidence: 0.7,
      window_seconds: 3600,
    }
    store.setStats(stats)
    expect(store.flaggedRatio).toBeCloseTo(0.3)
    expect(store.safeRatio).toBeCloseTo(0.7)
    expect(store.assessmentRatio).toBeCloseTo(0.25)
  })

  it('reset clears events, stats, and the dedupe set', () => {
    const store = usePipelineStore()
    store.insertEvent(makeEvent({ id: 'a' }))
    store.setStats({
      ingestion_count: 5,
      classified_count: 5,
      flagged_count: 1,
      safe_count: 4,
      assessment_count: 1,
      avg_confidence: 0,
      window_seconds: 3600,
    })
    store.reset()
    expect(store.events).toHaveLength(0)
    expect(store.stats.ingestion_count).toBe(0)
    // Dedupe set must clear too — re-inserting the same id should succeed.
    expect(store.insertEvent(makeEvent({ id: 'a' }))).toBe(true)
  })

  it('ring buffer evicts oldest beyond cap', () => {
    const store = usePipelineStore()
    // Cap is 500 — insert 510 and check we kept 500 with the most recent ones.
    for (let i = 0; i < 510; i++) {
      store.insertEvent(
        makeEvent({ id: 'evt-' + i, occurred_at: new Date(2026, 3, 18, 0, 0, i).toISOString() }),
      )
    }
    expect(store.events).toHaveLength(500)
    // Newest (last inserted) is at index 0.
    expect(store.events[0].id).toBe('evt-509')
    // Oldest 10 should have been evicted.
    expect(store.events[store.events.length - 1].id).toBe('evt-10')
  })
})
