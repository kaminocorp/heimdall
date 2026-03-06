import { useAgentStore } from '@/stores/agent'
import client from '@/api/client'

const mockConfig = {
  id: '1',
  model: 'claude-sonnet-4-5-20250929',
  mode: 'continuous',
  schedule: null,
  system_prompt_override: null,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
}

describe('agent store', () => {
  beforeEach(() => {
    vi.mocked(client.get).mockReset()
    vi.mocked(client.put).mockReset()
  })

  it('fetchConfig stores config', async () => {
    vi.mocked(client.get).mockResolvedValueOnce({ data: mockConfig } as any)

    const store = useAgentStore()
    await store.fetchConfig()

    expect(store.config).toEqual(mockConfig)
    expect(store.loading).toBe(false)
  })

  it('updateConfig sends PUT and updates local state', async () => {
    const updated = { ...mockConfig, mode: 'scheduled', schedule: '*/5 * * * *' }
    vi.mocked(client.put).mockResolvedValueOnce({ data: updated } as any)

    const store = useAgentStore()
    store.config = mockConfig as any
    await store.updateConfig({ mode: 'scheduled', schedule: '*/5 * * * *' })

    expect(store.config?.mode).toBe('scheduled')
    expect(store.config?.schedule).toBe('*/5 * * * *')
    expect(vi.mocked(client.put)).toHaveBeenCalledWith('/agent/config', {
      mode: 'scheduled',
      schedule: '*/5 * * * *',
    })
  })
})
