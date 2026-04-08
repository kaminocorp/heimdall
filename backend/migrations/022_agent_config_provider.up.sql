-- Phase 3 of OpenRouter integration: per-app provider selection.
-- Existing rows default to 'anthropic' so behaviour is unchanged after migrate-up.
-- Valid values today: 'anthropic', 'openrouter'. Validation lives in the API
-- handler (applications.go UpdateAppAgentConfig) — no DB-level CHECK so we can
-- add new providers without a migration.

ALTER TABLE app_agent_config
ADD COLUMN provider TEXT NOT NULL DEFAULT 'anthropic';
