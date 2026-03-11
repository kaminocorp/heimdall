package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hejijunhao/heimdall/backend/internal/config"
)

// EmailChannel sends notifications via the Resend API.
type EmailChannel struct {
	recipients []string
	apiKey     string
	fromEmail  string
}

type emailConfig struct {
	Recipients []string `json:"recipients"`
}

func newEmailChannel(configBytes json.RawMessage, cfg *config.Config) (*EmailChannel, error) {
	var ec emailConfig
	if err := json.Unmarshal(configBytes, &ec); err != nil {
		return nil, fmt.Errorf("invalid email config: %w", err)
	}
	if len(ec.Recipients) == 0 {
		return nil, fmt.Errorf("email recipients is required")
	}
	return &EmailChannel{
		recipients: ec.Recipients,
		apiKey:     cfg.ResendAPIKey,
		fromEmail:  cfg.NotificationFromEmail,
	}, nil
}

func (e *EmailChannel) Type() string { return "email" }

func (e *EmailChannel) Send(ctx context.Context, p Payload) error {
	if e.apiKey == "" {
		return fmt.Errorf("email: RESEND_API_KEY not configured")
	}
	if e.fromEmail == "" {
		return fmt.Errorf("email: NOTIFICATION_FROM_EMAIL not configured")
	}

	reqBody, err := json.Marshal(map[string]any{
		"from":    e.fromEmail,
		"to":      e.recipients,
		"subject": FormatSubject(p.Severity, p.AppName),
		"html":    FormatEmailHTML(p),
	})
	if err != nil {
		return fmt.Errorf("email: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("email: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := webhookClient.Do(req)
	if err != nil {
		return fmt.Errorf("email: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, 512))
		if err != nil {
			return fmt.Errorf("email: resend returned %d (could not read body)", resp.StatusCode)
		}
		return fmt.Errorf("email: resend returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
