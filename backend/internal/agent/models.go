package agent

// ModelOption is the public shape returned by GET /api/models. The frontend
// uses it to populate the model picker on the Agent Config page.
//
// Pricing is in USD per million tokens. ContextLength is the model's full
// context window in tokens.
type ModelOption struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Provider      string   `json:"provider"`
	ContextLength int      `json:"context_length"`
	Pricing       Pricing  `json:"pricing"`

	// Enriched fields — drive the rich picker UI (Phase 3+).
	Vendor      string   `json:"vendor"`      // who made the model (anthropic, openai, google, ...)
	Tier        string   `json:"tier"`         // flagship | balanced | economy | specialist
	Strengths   []string `json:"strengths"`    // short tags: reasoning, coding, long-context, ...
	Description string   `json:"description"`  // one sentence, max ~120 chars
}

// Pricing tracks per-million-token costs for input and output.
type Pricing struct {
	Prompt     float64 `json:"prompt"`
	Completion float64 `json:"completion"`
}

// Vendor constants — the fixed set of model vendors in the catalogue.
const (
	VendorAnthropic = "anthropic"
	VendorOpenAI    = "openai"
	VendorGoogle    = "google"
	VendorXAI       = "xai"
	VendorQwen      = "qwen"
	VendorZAI       = "zai"
	VendorMiniMax   = "minimax"
	VendorXiaomi    = "xiaomi"
	VendorMoonshot  = "moonshot"
	VendorMistral   = "mistral"
)

// Tier constants — model capability/price tiers.
const (
	TierFlagship   = "flagship"
	TierBalanced   = "balanced"
	TierEconomy    = "economy"
	TierSpecialist = "specialist"
)

// ResolveModel looks up a model by ID across both the Anthropic-direct and
// OpenRouter catalogues. Returns the model and true if found, or zero value
// and false if the ID is not in either list.
//
// Used by UpdateAppAgentConfig (Phase 2) to reject unknown model IDs at the
// API boundary instead of letting typos through to fail at first invocation.
func ResolveModel(id string) (ModelOption, bool) {
	for _, m := range AnthropicModels {
		if m.ID == id {
			return m, true
		}
	}
	for _, m := range OpenRouterModels {
		if m.ID == id {
			return m, true
		}
	}
	return ModelOption{}, false
}

// AnthropicModels is the curated list of models reachable via the direct
// Anthropic provider. These IDs are passed verbatim to the Anthropic SDK.
//
// Source of truth for prices/context: Anthropic's public pricing page at the
// time of curation. Update when models are added or pricing changes.
var AnthropicModels = []ModelOption{
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "claude-sonnet-4-6",
		Name:          "Claude Sonnet 4.6",
		Provider:      "anthropic",
		ContextLength: 200_000,
		Pricing:       Pricing{Prompt: 3, Completion: 15},
		Vendor:        VendorAnthropic,
		Tier:          TierBalanced,
		Strengths:     []string{"reasoning", "coding", "tool-use"},
		Description:   "Near-Opus quality at Sonnet price, strong tool use",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "claude-haiku-4-5-20251001",
		Name:          "Claude Haiku 4.5",
		Provider:      "anthropic",
		ContextLength: 200_000,
		Pricing:       Pricing{Prompt: 1, Completion: 5},
		Vendor:        VendorAnthropic,
		Tier:          TierEconomy,
		Strengths:     []string{"fast", "tool-use", "cheap"},
		Description:   "Fast, cheap, Anthropic-native tool use",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "claude-opus-4-6",
		Name:          "Claude Opus 4.6",
		Provider:      "anthropic",
		ContextLength: 200_000,
		Pricing:       Pricing{Prompt: 15, Completion: 75},
		Vendor:        VendorAnthropic,
		Tier:          TierFlagship,
		Strengths:     []string{"reasoning", "coding", "diagnosis"},
		Description:   "SWE-bench leader, deep reasoning, deliberate diagnosis",
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
//
// Prices sourced from OpenRouter API snapshot: docs/archive/openrouter-models-2026-04-12.json
var OpenRouterModels = []ModelOption{

	// ── Flagship ────────────────────────────────────────────────────────

	// tool_use_verified_at: 2026-04-12
	{
		ID:            "anthropic/claude-opus-4.6",
		Name:          "Claude Opus 4.6",
		Provider:      "openrouter",
		ContextLength: 1_000_000,
		Pricing:       Pricing{Prompt: 5, Completion: 25},
		Vendor:        VendorAnthropic,
		Tier:          TierFlagship,
		Strengths:     []string{"reasoning", "coding", "diagnosis"},
		Description:   "Same Opus via OpenRouter — cheaper, 1M context",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "openai/gpt-5.4",
		Name:          "GPT-5.4",
		Provider:      "openrouter",
		ContextLength: 1_050_000,
		Pricing:       Pricing{Prompt: 2.5, Completion: 15},
		Vendor:        VendorOpenAI,
		Tier:          TierFlagship,
		Strengths:     []string{"reasoning", "general", "long-context"},
		Description:   "Balanced frontier model, strong general reasoning",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "openai/gpt-5.4-pro",
		Name:          "GPT-5.4 Pro",
		Provider:      "openrouter",
		ContextLength: 1_050_000,
		Pricing:       Pricing{Prompt: 30, Completion: 180},
		Vendor:        VendorOpenAI,
		Tier:          TierFlagship,
		Strengths:     []string{"reasoning", "hard-problems"},
		Description:   "Hardest cases — premium price, use sparingly",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "google/gemini-3.1-pro-preview",
		Name:          "Gemini 3.1 Pro",
		Provider:      "openrouter",
		ContextLength: 1_048_576,
		Pricing:       Pricing{Prompt: 2, Completion: 12},
		Vendor:        VendorGoogle,
		Tier:          TierFlagship,
		Strengths:     []string{"reasoning", "tool-use", "long-context"},
		Description:   "Huge context, strong agentic tool use",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "x-ai/grok-4.20",
		Name:          "Grok 4.20",
		Provider:      "openrouter",
		ContextLength: 2_000_000,
		Pricing:       Pricing{Prompt: 2, Completion: 6},
		Vendor:        VendorXAI,
		Tier:          TierFlagship,
		Strengths:     []string{"coding", "reasoning", "long-context"},
		Description:   "2M context, competitive coding and reasoning",
	},

	// ── Balanced ────────────────────────────────────────────────────────

	// tool_use_verified_at: 2026-04-12
	{
		ID:            "anthropic/claude-sonnet-4.6",
		Name:          "Claude Sonnet 4.6",
		Provider:      "openrouter",
		ContextLength: 1_000_000,
		Pricing:       Pricing{Prompt: 3, Completion: 15},
		Vendor:        VendorAnthropic,
		Tier:          TierBalanced,
		Strengths:     []string{"reasoning", "coding", "tool-use"},
		Description:   "Same Sonnet via OpenRouter — same price, 1M context",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "openai/gpt-5.4-mini",
		Name:          "GPT-5.4 Mini",
		Provider:      "openrouter",
		ContextLength: 400_000,
		Pricing:       Pricing{Prompt: 0.75, Completion: 4.5},
		Vendor:        VendorOpenAI,
		Tier:          TierBalanced,
		Strengths:     []string{"reasoning", "general", "cheap"},
		Description:   "Cheap OpenAI reasoning, good day-to-day",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "google/gemini-3-flash-preview",
		Name:          "Gemini 3 Flash",
		Provider:      "openrouter",
		ContextLength: 1_048_576,
		Pricing:       Pricing{Prompt: 0.5, Completion: 3},
		Vendor:        VendorGoogle,
		Tier:          TierBalanced,
		Strengths:     []string{"fast", "tool-use", "long-context"},
		Description:   "Fast, cheap, near-Pro agentic capability",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "qwen/qwen3.6-plus",
		Name:          "Qwen 3.6 Plus",
		Provider:      "openrouter",
		ContextLength: 1_000_000,
		Pricing:       Pricing{Prompt: 0.325, Completion: 1.95},
		Vendor:        VendorQwen,
		Tier:          TierBalanced,
		Strengths:     []string{"reasoning", "long-context", "cheap"},
		Description:   "1M context, chain-of-thought, cheap Alibaba flagship",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "z-ai/glm-4.7",
		Name:          "GLM 4.7",
		Provider:      "openrouter",
		ContextLength: 202_752,
		Pricing:       Pricing{Prompt: 0.39, Completion: 1.75},
		Vendor:        VendorZAI,
		Tier:          TierBalanced,
		Strengths:     []string{"general", "tool-use"},
		Description:   "Z.ai flagship, well-rounded, solid tool use",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "minimax/minimax-m2.7",
		Name:          "MiniMax M2.7",
		Provider:      "openrouter",
		ContextLength: 204_800,
		Pricing:       Pricing{Prompt: 0.30, Completion: 1.20},
		Vendor:        VendorMiniMax,
		Tier:          TierBalanced,
		Strengths:     []string{"coding", "general"},
		Description:   "Multimodal-leaning, competitive coding",
	},

	// ── Economy ─────────────────────────────────────────────────────────

	// tool_use_verified_at: 2026-04-12
	{
		ID:            "openai/gpt-5.4-nano",
		Name:          "GPT-5.4 Nano",
		Provider:      "openrouter",
		ContextLength: 400_000,
		Pricing:       Pricing{Prompt: 0.20, Completion: 1.25},
		Vendor:        VendorOpenAI,
		Tier:          TierEconomy,
		Strengths:     []string{"fast", "cheap", "reasoning"},
		Description:   "Ultra-cheap tiny reasoning",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "google/gemini-3.1-flash-lite-preview",
		Name:          "Gemini 3.1 Flash Lite",
		Provider:      "openrouter",
		ContextLength: 1_048_576,
		Pricing:       Pricing{Prompt: 0.25, Completion: 1.50},
		Vendor:        VendorGoogle,
		Tier:          TierEconomy,
		Strengths:     []string{"cheap", "long-context"},
		Description:   "Cheapest Google option with 1M context",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "z-ai/glm-4.7-flash",
		Name:          "GLM 4.7 Flash",
		Provider:      "openrouter",
		ContextLength: 202_752,
		Pricing:       Pricing{Prompt: 0.06, Completion: 0.40},
		Vendor:        VendorZAI,
		Tier:          TierEconomy,
		Strengths:     []string{"cheap", "tool-use"},
		Description:   "Among cheapest tool-capable models",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "qwen/qwen3.5-flash-02-23",
		Name:          "Qwen 3.5 Flash",
		Provider:      "openrouter",
		ContextLength: 1_000_000,
		Pricing:       Pricing{Prompt: 0.065, Completion: 0.26},
		Vendor:        VendorQwen,
		Tier:          TierEconomy,
		Strengths:     []string{"cheap", "long-context"},
		Description:   "Dirt cheap, huge context",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "xiaomi/mimo-v2-flash",
		Name:          "MiMo V2 Flash",
		Provider:      "openrouter",
		ContextLength: 262_144,
		Pricing:       Pricing{Prompt: 0.09, Completion: 0.29},
		Vendor:        VendorXiaomi,
		Tier:          TierEconomy,
		Strengths:     []string{"cheap", "general"},
		Description:   "High usage, cheap general-purpose",
	},

	// ── Specialist ──────────────────────────────────────────────────────

	// tool_use_verified_at: 2026-04-12
	{
		ID:            "qwen/qwen3-coder-next",
		Name:          "Qwen 3 Coder",
		Provider:      "openrouter",
		ContextLength: 262_144,
		Pricing:       Pricing{Prompt: 0.12, Completion: 0.75},
		Vendor:        VendorQwen,
		Tier:          TierSpecialist,
		Strengths:     []string{"coding"},
		Description:   "Purpose-built coding model, cheap",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "moonshotai/kimi-k2.5",
		Name:          "Kimi K2.5",
		Provider:      "openrouter",
		ContextLength: 262_144,
		Pricing:       Pricing{Prompt: 0.38, Completion: 1.72},
		Vendor:        VendorMoonshot,
		Tier:          TierSpecialist,
		Strengths:     []string{"coding", "agentic"},
		Description:   "Dominates coding-focused workloads per leaderboards",
	},
	// tool_use_verified_at: 2026-04-12
	{
		ID:            "mistralai/devstral-2512",
		Name:          "Devstral",
		Provider:      "openrouter",
		ContextLength: 262_144,
		Pricing:       Pricing{Prompt: 0.40, Completion: 2.00},
		Vendor:        VendorMistral,
		Tier:          TierSpecialist,
		Strengths:     []string{"coding"},
		Description:   "Coding-specialised, European alternative",
	},
}
