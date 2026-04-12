# refresh-models

Compares the curated Heimdall model catalogue (`internal/agent/models.go`) against the live OpenRouter API. Run this before editing `models.go` to catch price drifts, context length changes, and model removals.

## Usage

```bash
# Default: diff curated vs live
go run ./cmd/refresh-models

# Explicit diff mode
go run ./cmd/refresh-models --diff

# Suggest uncurated models by vendor
go run ./cmd/refresh-models --suggest

# Both at once
go run ./cmd/refresh-models --diff --suggest

# Pass API key inline (otherwise reads OPENROUTER_API_KEY env var)
go run ./cmd/refresh-models --key sk-or-...
```

## When to run

- Before adding or removing models from the catalogue
- After seeing reports of pricing changes on OpenRouter
- Quarterly, as a hygiene check

## Output

**`--diff`** compares every OpenRouter entry in `models.go` against live data:

```
  ✓ OK       anthropic/claude-sonnet-4.6
  △ DRIFT    openai/gpt-5.4-mini               prompt: $0.75 → $0.60
  ✗ MISSING  deepseek/deepseek-v3.2             not found on OpenRouter
```

**`--suggest`** lists the top 3 uncurated models per vendor, sorted by context length:

```
  google (12 uncurated):
    google/gemma-3-27b                            ctx=131k  $0.10/$0.20
    ...
```

## Not for production

This tool is not wired into CI, not run at request time, and not scheduled. It is a developer-operated aid for catalogue curation.
