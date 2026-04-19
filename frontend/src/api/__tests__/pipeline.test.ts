import { vi, describe, it, expect, beforeEach } from 'vitest'
import client from '@/api/client'
import { fetchPipelineLogs } from '@/api/pipeline'

describe('fetchPipelineLogs', () => {
  const getMock = vi.mocked(client.get)

  beforeEach(() => {
    getMock.mockReset()
  })

  it('calls /apps/{appId}/pipeline/logs with the expected params and passes through defaults', async () => {
    getMock.mockResolvedValue({
      data: { logs: [], since: 'x', until: 'y', limit: 50, offset: 0, truncated: false },
    })
    await fetchPipelineLogs('app-123', { since: '2026-04-19T00:00:00Z', until: '2026-04-19T12:00:00Z' })

    expect(getMock).toHaveBeenCalledTimes(1)
    const [url, opts] = getMock.mock.calls[0]
    expect(url).toBe('/apps/app-123/pipeline/logs')
    expect(opts?.params).toMatchObject({
      since: '2026-04-19T00:00:00Z',
      until: '2026-04-19T12:00:00Z',
      limit: 50,
      offset: 0,
    })
  })

  it('forwards custom limit + offset for paging', async () => {
    getMock.mockResolvedValue({
      data: { logs: [], since: 'x', until: 'y', limit: 100, offset: 50, truncated: true },
    })
    await fetchPipelineLogs('app-123', { limit: 100, offset: 50 })

    const [, opts] = getMock.mock.calls[0]
    expect(opts?.params).toMatchObject({ limit: 100, offset: 50 })
  })
})
