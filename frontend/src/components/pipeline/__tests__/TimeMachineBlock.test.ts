import { vi, describe, it, expect, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import client from '@/api/client'
import TimeMachineBlock from '@/components/pipeline/TimeMachineBlock.vue'

describe('TimeMachineBlock', () => {
  const getMock = vi.mocked(client.get)

  beforeEach(() => {
    getMock.mockReset()
  })

  it('clicking Replay triggers the picker fetch and renders results', async () => {
    getMock.mockResolvedValue({
      data: {
        logs: [
          {
            log_id: 'log-1',
            first_seen_at: '2026-04-19T10:00:00Z',
            last_seen_at: '2026-04-19T10:00:01Z',
            stage_count: 4,
            source_type: 'fly',
            severity: 'error',
            type: 'ERROR',
            category: 'runtime_exception',
            escalated: true,
            rule_hit: 'error_type',
          },
        ],
        since: '2026-04-18T10:00:00Z',
        until: '2026-04-19T10:00:00Z',
        limit: 50,
        offset: 0,
        truncated: false,
      },
    })

    const wrapper = mount(TimeMachineBlock, {
      props: { appId: 'app-123' },
    })

    // Trigger via the Replay button.
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(getMock).toHaveBeenCalledWith(
      '/apps/app-123/pipeline/logs',
      expect.objectContaining({
        params: expect.objectContaining({ limit: 50, offset: 0 }),
      }),
    )
    // The picker row shows human-readable summary fields rather than the
    // raw log_id (which would be a 36-char UUID); the row is click-able to
    // inspect the log, covered by the next test.
    expect(wrapper.text()).toContain('1 journey')
    expect(wrapper.text()).toContain('fly')
    expect(wrapper.text()).toContain('ERROR')
    expect(wrapper.text()).toContain('flagged')
  })

  it('clicking a log row emits inspect with the log id', async () => {
    getMock.mockResolvedValue({
      data: {
        logs: [
          {
            log_id: 'log-xyz',
            first_seen_at: '2026-04-19T10:00:00Z',
            last_seen_at: '2026-04-19T10:00:01Z',
            stage_count: 3,
          },
        ],
        since: '',
        until: '',
        limit: 50,
        offset: 0,
        truncated: false,
      },
    })

    const wrapper = mount(TimeMachineBlock, {
      props: { appId: 'app-123' },
    })

    await wrapper.get('button').trigger('click')
    await flushPromises()

    const row = wrapper.get('ul li')
    await row.trigger('click')

    expect(wrapper.emitted('inspect')?.[0]).toEqual(['log-xyz'])
  })

  it('surfaces an empty-state message when the query returns zero journeys', async () => {
    getMock.mockResolvedValue({
      data: { logs: [], since: '', until: '', limit: 50, offset: 0, truncated: false },
    })

    const wrapper = mount(TimeMachineBlock, {
      props: { appId: 'app-123' },
    })
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('No journeys in this window')
  })
})
