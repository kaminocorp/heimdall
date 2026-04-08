export interface ModelOption {
  id: string
  name: string
  provider: 'anthropic' | 'openrouter'
  context_length: number
  pricing: {
    prompt: number     // $ per million input tokens
    completion: number // $ per million output tokens
  }
}
