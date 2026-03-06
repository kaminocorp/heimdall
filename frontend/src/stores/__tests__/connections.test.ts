import { useConnectionsStore } from '@/stores/connections'
import client from '@/api/client'
import type { Connection } from '@/types/connection'

const mockConn: Connection = {
  id: '1',
  name: 'Test DB',
  type: 'postgres',
  direction: 'two_way',
  config: {},
  status: 'inactive',
  last_seen: null,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  user_id: 'u1',
}

describe('connections store', () => {
  beforeEach(() => {
    vi.mocked(client.get).mockReset()
    vi.mocked(client.post).mockReset()
    vi.mocked(client.put).mockReset()
    vi.mocked(client.delete).mockReset()
  })

  it('fetchConnections populates array', async () => {
    vi.mocked(client.get).mockResolvedValueOnce({ data: [mockConn] } as any)

    const store = useConnectionsStore()
    await store.fetchConnections()

    expect(store.connections).toHaveLength(1)
    expect(store.connections[0].name).toBe('Test DB')
    expect(store.loading).toBe(false)
  })

  it('fetchConnections sets error on failure', async () => {
    vi.mocked(client.get).mockRejectedValueOnce({ response: { data: { error: 'boom' } } })

    const store = useConnectionsStore()
    await store.fetchConnections()

    expect(store.error).toBe('boom')
    expect(store.connections).toHaveLength(0)
  })

  it('createConnection adds to array', async () => {
    vi.mocked(client.post).mockResolvedValueOnce({ data: mockConn } as any)

    const store = useConnectionsStore()
    const result = await store.createConnection({
      name: 'Test DB',
      type: 'postgres',
      direction: 'two_way',
      config: {},
    })

    expect(result.id).toBe('1')
    expect(store.connections).toHaveLength(1)
  })

  it('deleteConnection removes from array', async () => {
    vi.mocked(client.delete).mockResolvedValueOnce({} as any)

    const store = useConnectionsStore()
    store.connections = [mockConn]
    await store.deleteConnection('1')

    expect(store.connections).toHaveLength(0)
  })

  it('testConnection sets testingId during test', async () => {
    vi.mocked(client.post).mockResolvedValueOnce({
      data: { success: true, message: 'ok' },
    } as any)

    const store = useConnectionsStore()
    store.connections = [{ ...mockConn }]

    const promise = store.testConnection('1')
    expect(store.testingId).toBe('1')

    await promise
    expect(store.testingId).toBeNull()
    expect(store.connections[0].status).toBe('active')
  })
})
