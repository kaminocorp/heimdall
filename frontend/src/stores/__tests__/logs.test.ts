import { useLogsStore } from '@/stores/logs'
import { useAppStore } from '@/stores/app'
import client from '@/api/client'

const mockResponse = {
  data: {
    data: [
      { id: '1', source: 'raw', timestamp: '2026-01-01T00:00:00Z', severity: 'info', source_type: 'webhook_logs', summary: 'test log' },
      { id: '2', source: 'agent', timestamp: '2026-01-01T00:01:00Z', severity: null, source_type: 'observation', summary: 'agent note' },
    ],
    total: 2,
    limit: 50,
    offset: 0,
  },
}

describe('logs store', () => {
  beforeEach(() => {
    vi.mocked(client.get).mockReset()
  })

  it('fetchLogs populates entries and total', async () => {
    vi.mocked(client.get).mockResolvedValueOnce(mockResponse as any)

    const store = useLogsStore()
    await store.fetchLogs()

    expect(store.entries).toHaveLength(2)
    expect(store.total).toBe(2)
    expect(store.loading).toBe(false)
  })

  it('fetchLogs sets error on failure', async () => {
    vi.mocked(client.get).mockRejectedValueOnce({ response: { data: { error: 'fail' } } })

    const store = useLogsStore()
    await store.fetchLogs()

    expect(store.error).toBe('fail')
  })

  it('nextPage adjusts offset and re-fetches', async () => {
    vi.mocked(client.get).mockResolvedValue(mockResponse as any)

    const store = useLogsStore()
    store.total = 100
    store.offset = 0

    store.nextPage()
    expect(store.offset).toBe(50)
  })

  it('prevPage adjusts offset and re-fetches', async () => {
    vi.mocked(client.get).mockResolvedValue(mockResponse as any)

    const store = useLogsStore()
    store.offset = 50

    store.prevPage()
    expect(store.offset).toBe(0)
  })

  it('setSource resets pagination', () => {
    const store = useLogsStore()
    store.offset = 100

    store.setSource('agent')
    expect(store.source).toBe('agent')
    expect(store.offset).toBe(0)
  })

  it('fetchLogs includes app_id from app store', async () => {
    vi.mocked(client.get).mockResolvedValueOnce(mockResponse as any)

    const appStore = useAppStore()
    appStore.currentAppId = 'test-app-uuid'

    const store = useLogsStore()
    await store.fetchLogs()

    expect(client.get).toHaveBeenCalledWith('/logs', {
      params: expect.objectContaining({ app_id: 'test-app-uuid' }),
    })
  })

  it('fetchLogs omits app_id when no app selected', async () => {
    vi.mocked(client.get).mockResolvedValueOnce(mockResponse as any)

    const appStore = useAppStore()
    appStore.currentAppId = null

    const store = useLogsStore()
    await store.fetchLogs()

    expect(client.get).toHaveBeenCalledWith('/logs', {
      params: expect.not.objectContaining({ app_id: expect.anything() }),
    })
  })
})
