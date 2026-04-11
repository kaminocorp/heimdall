import { useSchedulesStore } from '@/stores/schedules'
import client from '@/api/client'
import type { InvestigationSchedule } from '@/types/schedule'

// Minimal mock-response shape so we can satisfy the axios types the store
// destructures without leaning on `any`. We only care about .data in tests.
type MockAxiosResponse<T> = { data: T }

// Baseline fixture — individual tests mutate the returned value as needed.
function fixture(overrides: Partial<InvestigationSchedule> = {}): InvestigationSchedule {
  return {
    id: 's1',
    app_id: 'a1',
    name: 'Slow query check',
    prompt: 'Find slow queries in pg_stat_statements.',
    interval_secs: 0,
    cron_expr: '*/5 * * * *',
    enabled: true,
    last_run_at: null,
    last_status: null,
    last_error: null,
    last_summary: null,
    created_at: '2026-04-11T12:00:00Z',
    updated_at: '2026-04-11T12:00:00Z',
    ...overrides,
  }
}

describe('schedules store', () => {
  beforeEach(() => {
    vi.mocked(client.get).mockReset()
    vi.mocked(client.post).mockReset()
    vi.mocked(client.patch).mockReset()
    vi.mocked(client.delete).mockReset()
  })

  it('fetchSchedules populates the list', async () => {
    vi.mocked(client.get).mockResolvedValueOnce({ data: [fixture()] } as MockAxiosResponse<InvestigationSchedule[]>)

    const store = useSchedulesStore()
    await store.fetchSchedules('a1')

    expect(store.schedules).toHaveLength(1)
    expect(store.schedules[0].name).toBe('Slow query check')
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('fetchSchedules surfaces backend error messages', async () => {
    vi.mocked(client.get).mockRejectedValueOnce({ response: { data: { error: 'boom' } } })

    const store = useSchedulesStore()
    await store.fetchSchedules('a1')

    expect(store.error).toBe('boom')
    expect(store.schedules).toHaveLength(0)
  })

  it('createSchedule prepends the new row', async () => {
    const created = fixture({ id: 's2', name: 'New one' })
    vi.mocked(client.post).mockResolvedValueOnce({ data: created } as MockAxiosResponse<InvestigationSchedule>)

    const store = useSchedulesStore()
    store.schedules = [fixture({ id: 's1', name: 'Old one' })]

    const result = await store.createSchedule('a1', {
      name: 'New one',
      prompt: 'prompt',
      cron_expr: '*/5 * * * *',
      enabled: true,
    })

    expect(result.id).toBe('s2')
    expect(store.schedules).toHaveLength(2)
    // New row should be at the head — list endpoint returns created_at DESC.
    expect(store.schedules[0].id).toBe('s2')
  })

  it('updateSchedule replaces the row in place', async () => {
    const original = fixture({ name: 'Before' })
    const updated = fixture({ name: 'After' })
    vi.mocked(client.patch).mockResolvedValueOnce({ data: updated } as MockAxiosResponse<InvestigationSchedule>)

    const store = useSchedulesStore()
    store.schedules = [original]

    const result = await store.updateSchedule('a1', 's1', {
      name: 'After',
      prompt: 'prompt',
      cron_expr: '*/5 * * * *',
      enabled: true,
    })

    expect(result.name).toBe('After')
    expect(store.schedules).toHaveLength(1)
    expect(store.schedules[0].name).toBe('After')
  })

  it('deleteSchedule drops the row from the list', async () => {
    vi.mocked(client.delete).mockResolvedValueOnce({ data: undefined } as MockAxiosResponse<undefined>)

    const store = useSchedulesStore()
    store.schedules = [fixture(), fixture({ id: 's2', name: 'Other' })]
    await store.deleteSchedule('a1', 's1')

    expect(store.schedules).toHaveLength(1)
    expect(store.schedules[0].id).toBe('s2')
  })

  it('runNow toggles runningId during the call and updates the row', async () => {
    const seeded = fixture({ last_status: null, last_summary: null })
    const after = fixture({
      last_status: 'success',
      last_summary: 'Found 3 slow queries',
      last_run_at: '2026-04-11T13:00:00Z',
    })
    vi.mocked(client.post).mockResolvedValueOnce({ data: after } as MockAxiosResponse<InvestigationSchedule>)

    const store = useSchedulesStore()
    store.schedules = [seeded]

    const promise = store.runNow('a1', 's1')
    // Mid-flight: runningId reflects the target row. This is the piece the UI
    // uses to disable a card's "Run now" button individually.
    expect(store.runningId).toBe('s1')

    await promise
    expect(store.runningId).toBeNull()
    expect(store.schedules[0].last_status).toBe('success')
    expect(store.schedules[0].last_summary).toBe('Found 3 slow queries')
  })
})
