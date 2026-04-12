// refresh-models compares the curated Heimdall model catalogue against the
// live OpenRouter API. It is a human-operated tool run before editing
// models.go — not wired into CI or executed at request time.
//
// Usage:
//
//	go run ./cmd/refresh-models              # default: --diff
//	go run ./cmd/refresh-models --diff       # compare curated vs live (prices, context)
//	go run ./cmd/refresh-models --suggest    # list top models by vendor not yet curated
//
// Requires OPENROUTER_API_KEY in the environment (or pass via --key flag).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
)

const openRouterModelsURL = "https://openrouter.ai/api/v1/models"

// orModel is the relevant subset of an OpenRouter /api/v1/models entry.
type orModel struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	ContextLength int       `json:"context_length"`
	Pricing       orPricing `json:"pricing"`
}

type orPricing struct {
	Prompt     string `json:"prompt"`     // per-token, string-encoded float
	Completion string `json:"completion"` // per-token, string-encoded float
}

type orResponse struct {
	Data []orModel `json:"data"`
}

func main() {
	diffMode := flag.Bool("diff", false, "Compare curated catalogue against live OpenRouter data (default)")
	suggestMode := flag.Bool("suggest", false, "List top models by vendor not yet in the catalogue")
	apiKey := flag.String("key", "", "OpenRouter API key (defaults to OPENROUTER_API_KEY env var)")
	flag.Parse()

	// Default to diff mode if neither flag is set.
	if !*diffMode && !*suggestMode {
		*diffMode = true
	}

	key := *apiKey
	if key == "" {
		key = os.Getenv("OPENROUTER_API_KEY")
	}
	if key == "" {
		fmt.Fprintln(os.Stderr, "error: OPENROUTER_API_KEY not set (use --key or env var)")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Fetching %s ...\n", openRouterModelsURL)
	live, err := fetchModels(key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Fetched %d models from OpenRouter.\n\n", len(live))

	// Index live models by ID for O(1) lookup.
	liveByID := make(map[string]orModel, len(live))
	for _, m := range live {
		liveByID[m.ID] = m
	}

	if *diffMode {
		runDiff(liveByID)
	}
	if *suggestMode {
		runSuggest(live, liveByID)
	}
}

func fetchModels(apiKey string) ([]orModel, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodGet, openRouterModelsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}

	var result orResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return result.Data, nil
}

// perMillionTokens converts a per-token string price to per-million-tokens float.
func perMillionTokens(perToken string) float64 {
	var f float64
	fmt.Sscanf(perToken, "%f", &f)
	return f * 1_000_000
}

// round2 rounds to 2 decimal places for display.
func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

// runDiff compares every OpenRouter model in the curated catalogue against live data.
func runDiff(liveByID map[string]orModel) {
	fmt.Println("=== CATALOGUE DIFF (OpenRouter models) ===")
	fmt.Println()

	drifts := 0
	missing := 0

	for _, curated := range agent.OpenRouterModels {
		live, found := liveByID[curated.ID]
		if !found {
			fmt.Printf("  ✗ MISSING  %s — not found on OpenRouter (renamed or removed?)\n", curated.ID)
			missing++
			continue
		}

		livePrompt := round2(perMillionTokens(live.Pricing.Prompt))
		liveCompletion := round2(perMillionTokens(live.Pricing.Completion))
		liveCtx := live.ContextLength

		var issues []string

		if livePrompt != curated.Pricing.Prompt {
			issues = append(issues, fmt.Sprintf("prompt: $%.2f → $%.2f", curated.Pricing.Prompt, livePrompt))
		}
		if liveCompletion != curated.Pricing.Completion {
			issues = append(issues, fmt.Sprintf("completion: $%.2f → $%.2f", curated.Pricing.Completion, liveCompletion))
		}
		if liveCtx != curated.ContextLength {
			issues = append(issues, fmt.Sprintf("context: %d → %d", curated.ContextLength, liveCtx))
		}

		if len(issues) > 0 {
			fmt.Printf("  △ DRIFT    %-40s  %s\n", curated.ID, strings.Join(issues, " | "))
			drifts++
		} else {
			fmt.Printf("  ✓ OK       %s\n", curated.ID)
		}
	}

	fmt.Println()
	fmt.Printf("Summary: %d OK, %d drifted, %d missing (of %d OpenRouter entries)\n",
		len(agent.OpenRouterModels)-drifts-missing, drifts, missing, len(agent.OpenRouterModels))
	fmt.Println()
}

// runSuggest lists models from each vendor that are not in the curated catalogue,
// sorted by context length (a rough proxy for capability/popularity).
func runSuggest(live []orModel, liveByID map[string]orModel) {
	fmt.Println("=== SUGGESTIONS (top uncurated models by vendor) ===")
	fmt.Println()

	// Build set of curated IDs.
	curated := make(map[string]bool, len(agent.AnthropicModels)+len(agent.OpenRouterModels))
	for _, m := range agent.AnthropicModels {
		curated[m.ID] = true
	}
	for _, m := range agent.OpenRouterModels {
		curated[m.ID] = true
	}

	// Group uncurated models by vendor (the part before "/").
	byVendor := make(map[string][]orModel)
	for _, m := range live {
		if curated[m.ID] {
			continue
		}
		parts := strings.SplitN(m.ID, "/", 2)
		if len(parts) < 2 {
			continue
		}
		vendor := parts[0]
		byVendor[vendor] = append(byVendor[vendor], m)
	}

	// Sort vendors alphabetically.
	vendors := make([]string, 0, len(byVendor))
	for v := range byVendor {
		vendors = append(vendors, v)
	}
	sort.Strings(vendors)

	for _, vendor := range vendors {
		models := byVendor[vendor]
		// Sort by context length descending within each vendor.
		sort.Slice(models, func(i, j int) bool {
			return models[i].ContextLength > models[j].ContextLength
		})
		// Show top 3 per vendor.
		limit := 3
		if len(models) < limit {
			limit = len(models)
		}

		fmt.Printf("  %s (%d uncurated):\n", vendor, len(models))
		for _, m := range models[:limit] {
			prompt := round2(perMillionTokens(m.Pricing.Prompt))
			completion := round2(perMillionTokens(m.Pricing.Completion))
			ctx := m.ContextLength
			fmt.Printf("    %-45s  ctx=%dk  $%.2f/$%.2f\n",
				m.ID, ctx/1000, prompt, completion)
		}
		fmt.Println()
	}
}
