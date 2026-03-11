package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SlackChannel sends notifications via Slack incoming webhook.
type SlackChannel struct {
	webhookURL string
}

type slackConfig struct {
	WebhookURL string `json:"webhook_url"`
}

func newSlackChannel(configBytes json.RawMessage) (*SlackChannel, error) {
	var cfg slackConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return nil, fmt.Errorf("invalid slack config: %w", err)
	}
	if cfg.WebhookURL == "" {
		return nil, fmt.Errorf("slack webhook_url is required")
	}
	return &SlackChannel{webhookURL: cfg.WebhookURL}, nil
}

func (s *SlackChannel) Type() string { return "slack" }

func (s *SlackChannel) Send(ctx context.Context, p Payload) error {
	blocks := []map[string]any{
		{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": fmt.Sprintf("%s %s — %s", SeverityEmoji(p.Severity), p.Severity, p.AppName),
			},
		},
		{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": truncate(p.Summary, 2000),
			},
		},
	}

	if p.Assessment != "" {
		blocks = append(blocks, map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": fmt.Sprintf("```%s```", truncate(p.Assessment, 2900)),
			},
		})
	}

	blocks = append(blocks, map[string]any{
		"type": "context",
		"elements": []map[string]any{
			{
				"type": "mrkdwn",
				"text": fmt.Sprintf("Heimdall · %s", p.Timestamp),
			},
		},
	})

	body, err := json.Marshal(map[string]any{"blocks": blocks})
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	return postWebhook(ctx, s.webhookURL, body)
}

func truncate(s string, max int) string {
	if max < 4 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

var webhookClient = &http.Client{Timeout: 10 * time.Second}

func postWebhook(ctx context.Context, url string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := webhookClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, 512))
		if err != nil {
			return fmt.Errorf("webhook returned %d (could not read body)", resp.StatusCode)
		}
		return fmt.Errorf("webhook returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
