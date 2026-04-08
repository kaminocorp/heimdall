package agent

// ModelOption is the public shape returned by GET /api/models. The frontend
// uses it to populate the model-selection dropdown on the Agent Config page.
//
// Pricing is in USD per million tokens. ContextLength is the model's full
// context window in tokens.
type ModelOption struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Provider      string  `json:"provider"`
	ContextLength int     `json:"context_length"`
	Pricing       Pricing `json:"pricing"`
}

// Pricing tracks per-million-token costs for input and output.
type Pricing struct {
	Prompt     float64 `json:"prompt"`
	Completion float64 `json:"completion"`
}

// AnthropicModels is the curated list of models reachable via the direct
// Anthropic provider. These IDs are passed verbatim to the Anthropic SDK.
//
// Source of truth for prices/context: Anthropic's public pricing page at the
// time of curation. Update when models are added or pricing changes.
var AnthropicModels = []ModelOption{
	{
		ID:            "claude-sonnet-4-6",
		Name:          "Claude Sonnet 4.6",
		Provider:      "anthropic",
		ContextLength: 200_000,
		Pricing:       Pricing{Prompt: 3, Completion: 15},
	},
	{
		ID:            "claude-haiku-4-5-20251001",
		Name:          "Claude Haiku 4.5",
		Provider:      "anthropic",
		ContextLength: 200_000,
		Pricing:       Pricing{Prompt: 1, Completion: 5},
	},
	{
		ID:            "claude-opus-4-6",
		Name:          "Claude Opus 4.6",
		Provider:      "anthropic",
		ContextLength: 200_000,
		Pricing:       Pricing{Prompt: 15, Completion: 75},
	},
}

// OpenRouterModels is the curated list of OpenRouter models verified to handle
// multi-turn tool use reliably. Adding a model here is a deliberate code
// review — not a runtime catalogue fetch — because OpenRouter exposes hundreds
// of models and many of them silently fail tool-use round-trips, which would
// break the agent loop in surprising ways.
//
// IDs are OpenRouter's canonical "<vendor>/<model>" form and are passed
// verbatim in the chat completions request body.
var OpenRouterModels = []ModelOption{
	{
		ID:            "anthropic/claude-sonnet-4",
		Name:          "Claude Sonnet 4 (OR)",
		Provider:      "openrouter",
		ContextLength: 200_000,
		Pricing:       Pricing{Prompt: 3, Completion: 15},
	},
	{
		ID:            "openai/gpt-4o",
		Name:          "GPT-4o",
		Provider:      "openrouter",
		ContextLength: 128_000,
		Pricing:       Pricing{Prompt: 2.5, Completion: 10},
	},
	{
		ID:            "openai/gpt-4o-mini",
		Name:          "GPT-4o Mini",
		Provider:      "openrouter",
		ContextLength: 128_000,
		Pricing:       Pricing{Prompt: 0.15, Completion: 0.6},
	},
	{
		ID:            "google/gemini-2.5-pro",
		Name:          "Gemini 2.5 Pro",
		Provider:      "openrouter",
		ContextLength: 1_000_000,
		Pricing:       Pricing{Prompt: 1.25, Completion: 10},
	},
	{
		ID:            "google/gemini-2.5-flash",
		Name:          "Gemini 2.5 Flash",
		Provider:      "openrouter",
		ContextLength: 1_000_000,
		Pricing:       Pricing{Prompt: 0.15, Completion: 0.6},
	},
	{
		ID:            "meta-llama/llama-4-maverick",
		Name:          "Llama 4 Maverick",
		Provider:      "openrouter",
		ContextLength: 128_000,
		Pricing:       Pricing{Prompt: 0.2, Completion: 0.6},
	},
}
