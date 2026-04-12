import { mount } from '@vue/test-utils'
import { describe, it, expect, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import ModelPicker from '../ModelPicker.vue'
import type { ModelOption } from '@/types/models'

const mockModels: ModelOption[] = [
  {
    id: 'claude-sonnet-4-6',
    name: 'Claude Sonnet 4.6',
    provider: 'anthropic',
    context_length: 200_000,
    pricing: { prompt: 3, completion: 15 },
    vendor: 'anthropic',
    tier: 'balanced',
    strengths: ['reasoning', 'coding', 'tool-use'],
    description: 'Near-Opus quality at Sonnet price',
  },
  {
    id: 'claude-opus-4-6',
    name: 'Claude Opus 4.6',
    provider: 'anthropic',
    context_length: 200_000,
    pricing: { prompt: 15, completion: 75 },
    vendor: 'anthropic',
    tier: 'flagship',
    strengths: ['reasoning', 'coding', 'diagnosis'],
    description: 'SWE-bench leader, deep reasoning',
  },
  {
    id: 'openai/gpt-5.4',
    name: 'GPT-5.4',
    provider: 'openrouter',
    context_length: 1_050_000,
    pricing: { prompt: 2.5, completion: 15 },
    vendor: 'openai',
    tier: 'flagship',
    strengths: ['reasoning', 'general', 'long-context'],
    description: 'Balanced frontier model',
  },
  {
    id: 'openai/gpt-5.4-nano',
    name: 'GPT-5.4 Nano',
    provider: 'openrouter',
    context_length: 400_000,
    pricing: { prompt: 0.20, completion: 1.25 },
    vendor: 'openai',
    tier: 'economy',
    strengths: ['fast', 'cheap', 'reasoning'],
    description: 'Ultra-cheap tiny reasoning',
  },
  {
    id: 'qwen/qwen3-coder-next',
    name: 'Qwen 3 Coder',
    provider: 'openrouter',
    context_length: 262_144,
    pricing: { prompt: 0.12, completion: 0.75 },
    vendor: 'qwen',
    tier: 'specialist',
    strengths: ['coding'],
    description: 'Purpose-built coding model',
  },
]

function mountPicker(modelValue = 'claude-sonnet-4-6') {
  return mount(ModelPicker, {
    props: {
      modelValue,
      models: mockModels,
    },
  })
}

describe('ModelPicker', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders the selected model name in the trigger', () => {
    const wrapper = mountPicker()
    expect(wrapper.text()).toContain('Claude Sonnet 4.6')
    expect(wrapper.text()).toContain('balanced')
  })

  it('shows placeholder when no model selected', () => {
    const wrapper = mountPicker('nonexistent-id')
    expect(wrapper.text()).toContain('Select a model...')
  })

  it('opens the picker panel on trigger click', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')
    expect(wrapper.find('[role="listbox"]').exists()).toBe(true)
  })

  it('renders all models when no search or filter', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')

    const options = wrapper.findAll('[role="option"]')
    expect(options).toHaveLength(mockModels.length)
  })

  it('filters by search substring across name', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')

    const searchInput = wrapper.find('input[type="text"]')
    await searchInput.setValue('Opus')

    const options = wrapper.findAll('[role="option"]')
    expect(options).toHaveLength(1)
    expect(wrapper.text()).toContain('Claude Opus 4.6')
  })

  it('filters by search substring across vendor', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')

    const searchInput = wrapper.find('input[type="text"]')
    await searchInput.setValue('qwen')

    const options = wrapper.findAll('[role="option"]')
    expect(options).toHaveLength(1)
    expect(wrapper.text()).toContain('Qwen 3 Coder')
  })

  it('filters by search substring across strength tags', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')

    const searchInput = wrapper.find('input[type="text"]')
    await searchInput.setValue('diagnosis')

    const options = wrapper.findAll('[role="option"]')
    expect(options).toHaveLength(1)
    expect(wrapper.text()).toContain('Claude Opus 4.6')
  })

  it('filters by tier', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')

    // Click the "Flagship" tier chip
    const tierButtons = wrapper.findAll('.px-2.py-0\\.5')
    const flagshipBtn = tierButtons.find(b => b.text() === 'Flagship')
    expect(flagshipBtn).toBeTruthy()
    await flagshipBtn!.trigger('click')

    const options = wrapper.findAll('[role="option"]')
    expect(options).toHaveLength(2) // Opus + GPT-5.4
  })

  it('emits update:modelValue on card click', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')

    const options = wrapper.findAll('[role="option"]')
    // Click the second option (Claude Opus)
    await options[1].trigger('click')

    const emitted = wrapper.emitted('update:modelValue')
    expect(emitted).toHaveLength(1)
    expect(emitted![0]).toEqual(['claude-opus-4-6'])
  })

  it('closes panel after selection', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')
    expect(wrapper.find('[role="listbox"]').exists()).toBe(true)

    const options = wrapper.findAll('[role="option"]')
    await options[0].trigger('click')

    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
  })

  it('shows empty state when filters exclude everything', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')

    const searchInput = wrapper.find('input[type="text"]')
    await searchInput.setValue('zzz-nonexistent-model')

    expect(wrapper.findAll('[role="option"]')).toHaveLength(0)
    expect(wrapper.text()).toContain('No models match')
  })

  it('keyboard: arrow down then Enter selects the highlighted model', async () => {
    const wrapper = mountPicker()

    // Open via ArrowDown
    await wrapper.find('button').trigger('keydown', { key: 'ArrowDown' })
    expect(wrapper.find('[role="listbox"]').exists()).toBe(true)

    // Arrow down to the first model
    await wrapper.trigger('keydown', { key: 'ArrowDown' })
    // Arrow down to the second model
    await wrapper.trigger('keydown', { key: 'ArrowDown' })
    // Select it
    await wrapper.trigger('keydown', { key: 'Enter' })

    const emitted = wrapper.emitted('update:modelValue')
    expect(emitted).toHaveLength(1)
    expect(emitted![0]).toEqual(['claude-opus-4-6'])
  })

  it('keyboard: Escape closes the panel', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')
    expect(wrapper.find('[role="listbox"]').exists()).toBe(true)

    await wrapper.trigger('keydown', { key: 'Escape' })
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
  })

  it('groups models by vendor', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')

    const html = wrapper.html()
    // Vendor headers should appear
    expect(html).toContain('anthropic')
    expect(html).toContain('openai')
    expect(html).toContain('qwen')
  })

  it('renders long model names without overflow', () => {
    const longNameModel: ModelOption = {
      id: 'very-long-model-id-that-might-cause-overflow-issues-in-narrow-containers',
      name: 'A Very Long Model Name That Should Be Truncated By The Trigger Button Without Overflow',
      provider: 'openrouter',
      context_length: 200_000,
      pricing: { prompt: 1, completion: 5 },
      vendor: 'test',
      tier: 'balanced',
      strengths: ['test'],
      description: 'Test model for overflow behavior',
    }

    const wrapper = mount(ModelPicker, {
      props: {
        modelValue: longNameModel.id,
        models: [longNameModel],
      },
    })

    // The trigger button should have overflow: hidden via the truncate class
    const trigger = wrapper.find('button')
    expect(trigger.exists()).toBe(true)
    // Verify the trigger contains the truncate class on the name span
    const nameSpan = trigger.find('.truncate')
    expect(nameSpan.exists()).toBe(true)
  })
})
