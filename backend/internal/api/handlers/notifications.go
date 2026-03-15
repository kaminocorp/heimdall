package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/db"
	"github.com/hejijunhao/heimdall/backend/internal/notifications"
)

// --- Notification Preferences ---

func (s *Server) GetNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	prefs, err := s.Queries.GetNotificationPreferences(r.Context(), app.ID)
	if err != nil {
		// No row yet — return defaults.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"app_id":             app.ID,
			"enabled":            false,
			"severity_threshold": "warning",
			"cooldown_minutes":   15,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

type updatePreferencesRequest struct {
	Enabled           bool   `json:"enabled"`
	SeverityThreshold string `json:"severity_threshold"`
	CooldownMinutes   int32  `json:"cooldown_minutes"`
}

func (s *Server) UpdateNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	var req updatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	validThresholds := map[string]bool{"info": true, "warning": true, "error": true, "critical": true}
	if !validThresholds[req.SeverityThreshold] {
		jsonError(w, "severity_threshold must be one of: info, warning, error, critical", http.StatusBadRequest)
		return
	}

	if req.CooldownMinutes < 1 || req.CooldownMinutes > 1440 {
		jsonError(w, "cooldown_minutes must be between 1 and 1440", http.StatusBadRequest)
		return
	}

	prefs, err := s.Queries.UpsertNotificationPreferences(r.Context(), db.UpsertNotificationPreferencesParams{
		AppID:             app.ID,
		Enabled:           req.Enabled,
		SeverityThreshold: req.SeverityThreshold,
		CooldownMinutes:   req.CooldownMinutes,
	})
	if err != nil {
		jsonServerError(w, "failed to update notification preferences", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

// --- Notification Channels ---

func (s *Server) ListNotificationChannels(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	channels, err := s.Queries.ListNotificationChannelsByApp(r.Context(), app.ID)
	if err != nil {
		jsonServerError(w, "failed to list notification channels", err)
		return
	}
	if channels == nil {
		channels = []db.NotificationChannel{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channels)
}

type createChannelRequest struct {
	Type    string          `json:"type"`
	Name    string          `json:"name"`
	Config  json.RawMessage `json:"config"`
	Enabled *bool           `json:"enabled"`
}

func (s *Server) CreateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	var req createChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateChannelType(req.Type); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}
	if err := validateChannelConfig(req.Type, req.Config); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	channel, err := s.Queries.CreateNotificationChannel(r.Context(), db.CreateNotificationChannelParams{
		AppID:   app.ID,
		Type:    req.Type,
		Name:    req.Name,
		Config:  req.Config,
		Enabled: enabled,
	})
	if err != nil {
		jsonServerError(w, "failed to create notification channel", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(channel)
}

type updateChannelRequest struct {
	Name    string          `json:"name"`
	Config  json.RawMessage `json:"config"`
	Enabled bool            `json:"enabled"`
}

func (s *Server) UpdateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	channelID, err := uuid.Parse(chi.URLParam(r, "channelId"))
	if err != nil {
		jsonError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	// Verify channel belongs to this app.
	existing, err := s.Queries.GetNotificationChannel(r.Context(), channelID)
	if err != nil {
		jsonError(w, "notification channel not found", http.StatusNotFound)
		return
	}
	if existing.AppID != app.ID {
		jsonError(w, "notification channel not found", http.StatusNotFound)
		return
	}

	var req updateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}
	if err := validateChannelConfig(existing.Type, req.Config); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	channel, err := s.Queries.UpdateNotificationChannel(r.Context(), db.UpdateNotificationChannelParams{
		ID:      channelID,
		Name:    req.Name,
		Config:  req.Config,
		Enabled: req.Enabled,
	})
	if err != nil {
		jsonServerError(w, "failed to update notification channel", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channel)
}

func (s *Server) DeleteNotificationChannel(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	channelID, err := uuid.Parse(chi.URLParam(r, "channelId"))
	if err != nil {
		jsonError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	existing, err := s.Queries.GetNotificationChannel(r.Context(), channelID)
	if err != nil {
		jsonError(w, "notification channel not found", http.StatusNotFound)
		return
	}
	if existing.AppID != app.ID {
		jsonError(w, "notification channel not found", http.StatusNotFound)
		return
	}

	if err := s.Queries.DeleteNotificationChannel(r.Context(), channelID); err != nil {
		jsonServerError(w, "failed to delete notification channel", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) TestNotificationChannel(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	channelID, err := uuid.Parse(chi.URLParam(r, "channelId"))
	if err != nil {
		jsonError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	ch, err := s.Queries.GetNotificationChannel(r.Context(), channelID)
	if err != nil {
		jsonError(w, "notification channel not found", http.StatusNotFound)
		return
	}
	if ch.AppID != app.ID {
		jsonError(w, "notification channel not found", http.StatusNotFound)
		return
	}

	sender, err := notifications.NewChannel(ch.Type, ch.Config, s.Config)
	if err != nil {
		jsonError(w, "failed to initialize channel: "+err.Error(), http.StatusBadRequest)
		return
	}

	payload := notifications.Payload{
		AppName:    app.Name,
		Severity:   "info",
		Summary:    "Test notification from Heimdall",
		Assessment: "This is a test notification to verify your channel configuration is working correctly.",
		Timestamp:  "now",
	}

	if err := sender.Send(r.Context(), payload); err != nil {
		jsonError(w, "test notification failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

// --- Notification History ---

func (s *Server) ListNotificationHistory(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	limit := int32(20)
	offset := int32(0)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = int32(n)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}

	history, err := s.Queries.ListNotificationLogByApp(r.Context(), db.ListNotificationLogByAppParams{
		AppID:  app.ID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		jsonServerError(w, "failed to list notification history", err)
		return
	}
	if history == nil {
		history = []db.ListNotificationLogByAppRow{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// --- Validation helpers ---

func validateChannelType(t string) error {
	switch t {
	case "email", "slack", "discord":
		return nil
	default:
		return &validationError{"type must be one of: email, slack, discord"}
	}
}

func validateChannelConfig(channelType string, config json.RawMessage) error {
	if len(config) == 0 {
		return &validationError{"config is required"}
	}

	switch channelType {
	case "email":
		var c struct {
			Recipients []string `json:"recipients"`
		}
		if err := json.Unmarshal(config, &c); err != nil {
			return &validationError{"invalid email config"}
		}
		if len(c.Recipients) == 0 {
			return &validationError{"email config requires at least one recipient"}
		}

	case "slack":
		var c struct {
			WebhookURL string `json:"webhook_url"`
		}
		if err := json.Unmarshal(config, &c); err != nil {
			return &validationError{"invalid slack config"}
		}
		if err := validateWebhookURL(c.WebhookURL, "hooks.slack.com", "/services/"); err != nil {
			return &validationError{"slack webhook_url must be a valid https://hooks.slack.com/services/… URL"}
		}

	case "discord":
		var c struct {
			WebhookURL string `json:"webhook_url"`
		}
		if err := json.Unmarshal(config, &c); err != nil {
			return &validationError{"invalid discord config"}
		}
		if err := validateWebhookURL(c.WebhookURL, "discord.com", "/api/webhooks/"); err != nil {
			return &validationError{"discord webhook_url must be a valid https://discord.com/api/webhooks/… URL"}
		}
	}

	return nil
}

// validateWebhookURL parses the URL and checks scheme, exact host, and path prefix.
func validateWebhookURL(raw, expectedHost, pathPrefix string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host != expectedHost {
		return fmt.Errorf("invalid webhook URL")
	}
	if len(u.Path) <= len(pathPrefix) || u.Path[:len(pathPrefix)] != pathPrefix {
		return fmt.Errorf("invalid webhook URL path")
	}
	return nil
}

type validationError struct {
	msg string
}

func (e *validationError) Error() string { return e.msg }
