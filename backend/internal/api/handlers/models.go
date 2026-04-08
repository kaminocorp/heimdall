package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
)

// GetAvailableModels returns the curated list of models the user can select
// for an app's agent config. Anthropic models are always included.
// OpenRouter models are appended only when OPENROUTER_API_KEY is configured —
// this keeps the dropdown honest about what the backend can actually route.
func (s *Server) GetAvailableModels(w http.ResponseWriter, r *http.Request) {
	models := make([]agent.ModelOption, 0, len(agent.AnthropicModels)+len(agent.OpenRouterModels))
	models = append(models, agent.AnthropicModels...)
	if s.Config.OpenRouterKey != "" {
		models = append(models, agent.OpenRouterModels...)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}
