package notifications

import (
	"context"
	"encoding/json"
	"fmt"
)

// DiscordChannel sends notifications via Discord webhook.
type DiscordChannel struct {
	webhookURL string
}

type discordConfig struct {
	WebhookURL string `json:"webhook_url"`
}

func newDiscordChannel(configBytes json.RawMessage) (*DiscordChannel, error) {
	var cfg discordConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return nil, fmt.Errorf("invalid discord config: %w", err)
	}
	if cfg.WebhookURL == "" {
		return nil, fmt.Errorf("discord webhook_url is required")
	}
	return &DiscordChannel{webhookURL: cfg.WebhookURL}, nil
}

func (d *DiscordChannel) Type() string { return "discord" }

func (d *DiscordChannel) Send(ctx context.Context, p Payload) error {
	description := truncate(p.Summary, 2000)
	if p.Assessment != "" {
		description += "\n\n```\n" + truncate(p.Assessment, 1800) + "\n```"
	}

	payload := map[string]any{
		"embeds": []map[string]any{
			{
				"title":       fmt.Sprintf("%s [%s] %s", SeverityEmoji(p.Severity), p.Severity, p.AppName),
				"description": description,
				"color":       SeverityColor(p.Severity),
				"footer": map[string]any{
					"text": fmt.Sprintf("Heimdall · %s", p.Timestamp),
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: marshal payload: %w", err)
	}

	return postWebhook(ctx, d.webhookURL, body)
}
