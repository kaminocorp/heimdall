package agent

import (
	"testing"
)

// TestCatalogueConsistency verifies that every model entry in both catalogues
// has all required fields populated and that IDs are unique across the combined
// list. This is a standard (non-tagged) test that runs in CI on every commit.
func TestCatalogueConsistency(t *testing.T) {
	all := make([]ModelOption, 0, len(AnthropicModels)+len(OpenRouterModels))
	all = append(all, AnthropicModels...)
	all = append(all, OpenRouterModels...)

	seen := make(map[string]bool, len(all))

	for _, m := range all {
		t.Run(m.ID, func(t *testing.T) {
			// Unique ID
			if seen[m.ID] {
				t.Errorf("duplicate model ID: %s", m.ID)
			}
			seen[m.ID] = true

			// Required string fields
			if m.ID == "" {
				t.Error("ID is empty")
			}
			if m.Name == "" {
				t.Error("Name is empty")
			}
			if m.Provider == "" {
				t.Error("Provider is empty")
			}
			if m.Vendor == "" {
				t.Error("Vendor is empty")
			}
			if m.Tier == "" {
				t.Error("Tier is empty")
			}
			if m.Description == "" {
				t.Error("Description is empty")
			}

			// Provider must be one of the known values
			if m.Provider != "anthropic" && m.Provider != "openrouter" {
				t.Errorf("Provider %q is not anthropic or openrouter", m.Provider)
			}

			// Tier must be one of the known values
			validTiers := map[string]bool{
				TierFlagship: true, TierBalanced: true,
				TierEconomy: true, TierSpecialist: true,
			}
			if !validTiers[m.Tier] {
				t.Errorf("Tier %q is not a valid tier", m.Tier)
			}

			// Numeric fields must be positive
			if m.ContextLength <= 0 {
				t.Errorf("ContextLength must be positive, got %d", m.ContextLength)
			}
			if m.Pricing.Prompt <= 0 {
				t.Errorf("Pricing.Prompt must be positive, got %f", m.Pricing.Prompt)
			}
			if m.Pricing.Completion <= 0 {
				t.Errorf("Pricing.Completion must be positive, got %f", m.Pricing.Completion)
			}

			// Strengths must have at least one tag
			if len(m.Strengths) == 0 {
				t.Error("Strengths must have at least one tag")
			}

			// Description should be concise
			if len(m.Description) > 150 {
				t.Errorf("Description is %d chars (max 150): %s", len(m.Description), m.Description)
			}

			// Anthropic-direct models should have provider "anthropic"
			// OpenRouter models should have provider "openrouter"
			for _, am := range AnthropicModels {
				if am.ID == m.ID && m.Provider != "anthropic" {
					t.Errorf("Anthropic model %s should have provider anthropic, got %s", m.ID, m.Provider)
				}
			}
		})
	}
}

// TestCatalogueSize verifies the catalogue stays within the expected range.
func TestCatalogueSize(t *testing.T) {
	total := len(AnthropicModels) + len(OpenRouterModels)
	if total < 15 {
		t.Errorf("catalogue has only %d models — expected at least 15", total)
	}
	if total > 35 {
		t.Errorf("catalogue has %d models — expected at most 35 (are we over-curating?)", total)
	}
	t.Logf("catalogue size: %d (anthropic-direct: %d, openrouter: %d)",
		total, len(AnthropicModels), len(OpenRouterModels))
}

// TestResolveModel verifies the lookup helper used by server-side validation.
func TestResolveModel(t *testing.T) {
	// Known Anthropic-direct model
	m, ok := ResolveModel("claude-sonnet-4-6")
	if !ok {
		t.Fatal("expected claude-sonnet-4-6 to resolve")
	}
	if m.Provider != "anthropic" {
		t.Errorf("expected provider anthropic, got %s", m.Provider)
	}

	// Known OpenRouter model
	m, ok = ResolveModel("openai/gpt-5.4")
	if !ok {
		t.Fatal("expected openai/gpt-5.4 to resolve")
	}
	if m.Provider != "openrouter" {
		t.Errorf("expected provider openrouter, got %s", m.Provider)
	}

	// Unknown model
	_, ok = ResolveModel("anthropic/claude-sonnet-99")
	if ok {
		t.Error("expected unknown model to not resolve")
	}

	// Empty string
	_, ok = ResolveModel("")
	if ok {
		t.Error("expected empty string to not resolve")
	}
}

// TestDefaultModelInCatalogue verifies that DefaultModelID points to an entry
// that actually exists in the catalogue.
func TestDefaultModelInCatalogue(t *testing.T) {
	all := make([]ModelOption, 0, len(AnthropicModels)+len(OpenRouterModels))
	all = append(all, AnthropicModels...)
	all = append(all, OpenRouterModels...)

	found := false
	for _, m := range all {
		if m.ID == DefaultModelID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("DefaultModelID %q not found in catalogue", DefaultModelID)
	}
}
