import { mount } from '@vue/test-utils'
import { describe, it, expect, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import PipelineFunnel from '../PipelineFunnel.vue'
import { usePipelineStore } from '@/stores/pipeline'

describe('PipelineFunnel', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders the five stage labels', () => {
    // The CSS `text-transform: uppercase` only applies in a real browser; the
    // JSDOM/happy-dom test env returns the source-text casing.
    const wrapper = mount(PipelineFunnel)
    const labels = wrapper.findAll('text').map((t) => t.text())
    expect(labels).toEqual(
      expect.arrayContaining(['Ingestion', 'Lumber', 'Gate', 'Agent', 'Activity']),
    )
  })

  it('renders three stream paths (trunk, flagged, safe) plus the inflow beam', () => {
    const wrapper = mount(PipelineFunnel)
    // Each Sankey stream is a single <path>; the inflow beam adds a fourth.
    const paths = wrapper.findAll('path')
    expect(paths.length).toBeGreaterThanOrEqual(4)
  })

  it('post-gate path widths are derived from store flagged/safe ratios', async () => {
    const store = usePipelineStore()
    store.setStats({
      ingestion_count: 100,
      classified_count: 100,
      flagged_count: 80,
      safe_count: 20,
      assessment_count: 50,
      avg_confidence: 0.7,
      window_seconds: 3600,
    })
    const wrapper = mount(PipelineFunnel)
    await wrapper.vm.$nextTick()
    // Heavy-flagged ratio should produce a noticeably-larger flagged stream
    // than safe stream. We can't easily assert pixel widths through SVG path
    // strings, so check the exposed geometry refs the page binds against.
    const vm = wrapper.vm as unknown as { flaggedHalf: number; safeHalf: number }
    expect(vm.flaggedHalf).toBeGreaterThan(vm.safeHalf)
  })
})
