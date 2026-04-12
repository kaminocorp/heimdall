export interface ModelOption {
  id: string
  name: string
  provider: 'anthropic' | 'openrouter'
  context_length: number
  pricing: {
    prompt: number     // $ per million input tokens
    completion: number // $ per million output tokens
  }

  // Enriched fields — drive the rich picker UI (Phase 3+).
  vendor: string        // who made the model (anthropic, openai, google, ...)
  tier: 'flagship' | 'balanced' | 'economy' | 'specialist'
  strengths: string[]   // short tags: reasoning, coding, long-context, ...
  description: string   // one sentence, max ~120 chars
}
