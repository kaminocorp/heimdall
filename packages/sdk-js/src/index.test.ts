import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { Heimdall } from './index'

// Mock fetch globally.
const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

function okResponse() {
  return Promise.resolve(new Response(JSON.stringify({ ok: true }), { status: 201 }))
}

function errorResponse(status: number) {
  return Promise.resolve(new Response('error', { status }))
}

describe('Heimdall SDK', () => {
  let sdk: Heimdall

  beforeEach(() => {
    vi.useFakeTimers()
    mockFetch.mockReset()
    mockFetch.mockImplementation(okResponse)
    sdk = new Heimdall({
      endpoint: 'https://heimdall.example.com',
      token: 'test-token',
      batchSize: 3,
      flushInterval: 1000,
      maxRetries: 2,
      retryDelay: 100,
    })
  })

  afterEach(async () => {
    await sdk.shutdown()
    vi.useRealTimers()
  })

  it('buffers entries until batchSize is reached', async () => {
    sdk.info('test.event', { key: 'v1' })
    sdk.info('test.event', { key: 'v2' })
    expect(mockFetch).not.toHaveBeenCalled()
    expect(sdk.pending).toBe(2)

    // Third entry triggers flush.
    sdk.info('test.event', { key: 'v3' })
    // flush() is async — let microtasks run.
    await vi.runAllTimersAsync()
    expect(mockFetch).toHaveBeenCalledTimes(1)
    expect(sdk.pending).toBe(0)
  })

  it('flushes on timer when batchSize is not reached', async () => {
    sdk.info('test.event', { key: 'v1' })
    expect(mockFetch).not.toHaveBeenCalled()

    await vi.advanceTimersByTimeAsync(1100)
    expect(mockFetch).toHaveBeenCalledTimes(1)
  })

  it('sends correct payload format', async () => {
    sdk.error('payment.failed', { orderId: '456' })
    await sdk.flush()

    expect(mockFetch).toHaveBeenCalledWith(
      'https://heimdall.example.com/api/webhooks/logs',
      expect.objectContaining({
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer test-token',
        },
        body: JSON.stringify([
          { source_type: 'payment.failed', severity: 'error', payload: { orderId: '456' } },
        ]),
      }),
    )
  })

  it('strips trailing slash from endpoint', async () => {
    const s = new Heimdall({
      endpoint: 'https://heimdall.example.com/',
      token: 'tok',
    })
    s.info('test', {})
    await s.flush()
    expect(mockFetch).toHaveBeenCalledWith(
      'https://heimdall.example.com/api/webhooks/logs',
      expect.anything(),
    )
  })

  it('provides severity shorthands', async () => {
    // Use a large batchSize so all 5 entries stay in one batch.
    const s = new Heimdall({
      endpoint: 'https://heimdall.example.com',
      token: 'test-token',
      batchSize: 100,
    })
    s.debug('d', {})
    s.info('i', {})
    s.warn('w', {})
    s.error('e', {})
    s.critical('c', {})
    await s.flush()

    const body = JSON.parse(mockFetch.mock.calls[0][1].body)
    expect(body.map((e: any) => e.severity)).toEqual([
      'debug', 'info', 'warning', 'error', 'critical',
    ])
  })

  it('calls onError for 4xx responses without retrying', async () => {
    mockFetch.mockImplementation(() => errorResponse(401))
    const onError = vi.fn()
    const s = new Heimdall({
      endpoint: 'https://heimdall.example.com',
      token: 'bad-token',
      onError,
    })
    s.info('test', {})
    await s.flush()

    expect(mockFetch).toHaveBeenCalledTimes(1) // no retries
    expect(onError).toHaveBeenCalledTimes(1)
    expect(onError.mock.calls[0][0].message).toContain('401')
  })

  it('retries on 5xx and calls onError after exhaustion', async () => {
    mockFetch.mockImplementation(() => errorResponse(503))
    const onError = vi.fn()
    const s = new Heimdall({
      endpoint: 'https://heimdall.example.com',
      token: 'tok',
      maxRetries: 2,
      retryDelay: 10,
      onError,
    })
    s.info('test', {})

    // Use real timers for retry delays.
    vi.useRealTimers()
    await s.flush()

    expect(mockFetch).toHaveBeenCalledTimes(3) // 1 initial + 2 retries
    expect(onError).toHaveBeenCalledTimes(1)
  })

  it('retries on network error', async () => {
    mockFetch.mockRejectedValue(new Error('fetch failed'))
    const onError = vi.fn()
    const s = new Heimdall({
      endpoint: 'https://heimdall.example.com',
      token: 'tok',
      maxRetries: 1,
      retryDelay: 10,
      onError,
    })
    s.info('test', {})

    vi.useRealTimers()
    await s.flush()

    expect(mockFetch).toHaveBeenCalledTimes(2) // 1 + 1 retry
    expect(onError).toHaveBeenCalledTimes(1)
    expect(onError.mock.calls[0][0].message).toBe('fetch failed')
  })

  it('flush() is a no-op when buffer is empty', async () => {
    await sdk.flush()
    expect(mockFetch).not.toHaveBeenCalled()
  })

  it('shutdown flushes remaining entries', async () => {
    sdk.info('final', {})
    await sdk.shutdown()
    expect(mockFetch).toHaveBeenCalledTimes(1)
    expect(sdk.pending).toBe(0)
  })
})
