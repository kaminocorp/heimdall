package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// chatMessage matches the frontend ChatMessage interface and the JSONB storage format.
type chatMessage struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// incomingMessage is what the frontend sends over WebSocket.
type incomingMessage struct {
	Content string `json:"content"`
}

func (s *Server) HandleChat(w http.ResponseWriter, r *http.Request) {
	// Authenticate via query param — browser WebSocket API can't set custom headers.
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, `{"error":"missing token query parameter"}`, http.StatusUnauthorized)
		return
	}

	userID, err := middleware.ValidateJWT(tokenStr, s.JWKS)
	if err != nil {
		http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
		return
	}

	// Accept WebSocket connection after auth succeeds.
	// Use the same origin allowlist as the CORS middleware to prevent
	// cross-origin WebSocket connections from untrusted pages.
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: wsOriginPatterns(s.Config.FrontendURL),
	})
	if err != nil {
		slog.Error("websocket accept error", "err", err)
		return
	}
	defer conn.CloseNow()

	// Parse optional app_id for app-scoped tools (e.g. search_codebase).
	// Validate that the app belongs to the caller's org to prevent cross-tenant access.
	var appID uuid.UUID
	if appIDStr := r.URL.Query().Get("app_id"); appIDStr != "" {
		parsed, err := uuid.Parse(appIDStr)
		if err == nil {
			if _, authErr := s.Queries.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
				AppID:  parsed,
				UserID: userID,
			}); authErr != nil {
				writeWSError(r.Context(), conn, "app not found or access denied")
				return
			}
			appID = parsed
		}
	}

	ctx := r.Context()

	// Load or create conversation. Extracted into a helper so defer done()
	// handles transaction cleanup on all return paths.
	convID, storedMessages, setupErr := s.setupConversation(ctx, userID, r.URL.Query().Get("conversation_id"))
	if setupErr != "" {
		writeWSError(ctx, conn, setupErr)
		return
	}

	// Send system message with conversation ID.
	if err := wsjson.Write(ctx, conn, map[string]any{
		"type":            "system",
		"conversation_id": convID.String(),
	}); err != nil {
		slog.Error("websocket write error", "err", err)
		return
	}

	slog.Info("websocket chat connected", "user_id", userID, "conversation_id", convID)

	// Message read loop.
	for {
		var msg incomingMessage
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			slog.Info("websocket read closed", "err", err, "user_id", userID)
			return
		}

		if msg.Content == "" {
			continue
		}

		// Build user message and append to stored messages.
		userMsg := chatMessage{
			ID:        uuid.New().String(),
			Role:      "user",
			Content:   msg.Content,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		storedMessages = append(storedMessages, userMsg)
		s.persistMessages(ctx, convID, userID, storedMessages)

		// Set conversation title from first user message.
		if len(storedMessages) == 1 {
			title := userMsg.Content
			titleRunes := []rune(title)
			if len(titleRunes) > 50 {
				title = string(titleRunes[:50]) + "..."
			}
			if q, c, d, err := s.UserQueries(ctx, userID); err == nil {
				if err := q.UpdateConversationTitleByUser(ctx, db.UpdateConversationTitleByUserParams{
					ID:     convID,
					Title:  pgtype.Text{String: title, Valid: true},
					UserID: userID,
				}); err != nil {
					slog.Error("failed to set conversation title", "err", err, "conversation_id", convID)
				} else if err := c(); err != nil {
					slog.Error("failed to commit conversation title", "err", err, "conversation_id", convID)
				}
				d()
			}
		}

		// Signal thinking state.
		if err := wsjson.Write(ctx, conn, map[string]string{
			"type":    "status",
			"content": "thinking",
		}); err != nil {
			slog.Error("websocket write error", "err", err)
			return
		}

		// Convert stored messages to agent history (exclude the current message — it goes as input).
		history := toAgentHistory(storedMessages[:len(storedMessages)-1])

		// Run agent with conversation history. RunConversationStream emits
		// tool_start / tool_result events as the loop iterates, then either a
		// final "message" event with the full assistant text or an "error"
		// event. The underlying LLM call is still blocking — only the loop's
		// progress is streamed (see loop_stream.go).
		events := s.Agent.RunConversationStream(ctx, userID, appID, &convID, history, msg.Content)

		var fullResponse string
		var streamErr string
		for ev := range events {
			switch ev.Type {
			case "tool_start":
				if err := wsjson.Write(ctx, conn, map[string]string{
					"type": "tool_start",
					"tool": ev.Tool,
				}); err != nil {
					slog.Error("websocket write error", "err", err)
					return
				}
			case "tool_result":
				if err := wsjson.Write(ctx, conn, map[string]any{
					"type":     "tool_result",
					"tool":     ev.Tool,
					"is_error": ev.IsError,
				}); err != nil {
					slog.Error("websocket write error", "err", err)
					return
				}
			case "message":
				fullResponse = ev.Content
			case "error":
				streamErr = ev.Content
			}
		}

		if streamErr != "" {
			slog.Error("agent error", "err", streamErr, "user_id", userID, "conversation_id", convID)
			if writeErr := wsjson.Write(ctx, conn, map[string]string{
				"type":    "error",
				"content": "Agent error: " + streamErr,
			}); writeErr != nil {
				slog.Error("websocket write error", "err", writeErr)
				return
			}
			continue
		}

		// Build agent message and append to stored messages.
		agentMsg := chatMessage{
			ID:        uuid.New().String(),
			Role:      "assistant",
			Content:   fullResponse,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		storedMessages = append(storedMessages, agentMsg)
		s.persistMessages(ctx, convID, userID, storedMessages)

		// Send agent response to client (role "agent" for frontend compatibility).
		if err := wsjson.Write(ctx, conn, chatMessage{
			ID:        agentMsg.ID,
			Role:      "agent",
			Content:   agentMsg.Content,
			Timestamp: agentMsg.Timestamp,
		}); err != nil {
			slog.Error("websocket write error", "err", err)
			return
		}
	}
}

// setupConversation loads an existing conversation or creates a new one inside a
// single transactional scope with proper defer-based cleanup.  Returns the
// conversation ID, stored messages, and an error string for the WebSocket client
// (empty on success).
func (s *Server) setupConversation(ctx context.Context, userID uuid.UUID, cidStr string) (uuid.UUID, []chatMessage, string) {
	queries, commit, done, err := s.UserQueries(ctx, userID)
	if err != nil {
		return uuid.Nil, nil, "database error"
	}
	defer done()

	var convID uuid.UUID
	var storedMessages []chatMessage

	if cidStr != "" {
		cid, err := uuid.Parse(cidStr)
		if err != nil {
			return uuid.Nil, nil, "invalid conversation_id"
		}
		conv, err := queries.GetConversationByUser(ctx, db.GetConversationByUserParams{
			ID:     cid,
			UserID: userID,
		})
		if err != nil {
			return uuid.Nil, nil, "conversation not found"
		}
		convID = conv.ID
		if err := json.Unmarshal(conv.Messages, &storedMessages); err != nil {
			storedMessages = []chatMessage{}
		}
	} else {
		conv, err := queries.CreateConversation(ctx, db.CreateConversationParams{
			UserID:   userID,
			Messages: json.RawMessage(`[]`),
		})
		if err != nil {
			slog.Error("failed to create conversation", "err", err)
			return uuid.Nil, nil, "failed to create conversation"
		}
		convID = conv.ID
		storedMessages = []chatMessage{}
	}

	if err := commit(); err != nil {
		slog.Error("failed to commit conversation", "err", err)
		return uuid.Nil, nil, "failed to save conversation"
	}

	return convID, storedMessages, ""
}

// toAgentHistory converts stored chat messages to the agent's Message type.
func toAgentHistory(msgs []chatMessage) []agent.Message {
	history := make([]agent.Message, len(msgs))
	for i, m := range msgs {
		role := m.Role
		// Map "agent" to "assistant" for Claude API compatibility.
		if role == "agent" {
			role = "assistant"
		}
		history[i] = agent.Message{
			ID:        m.ID,
			Role:      role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		}
	}
	return history
}

// persistMessages marshals and saves messages to the conversation's JSONB column.
func (s *Server) persistMessages(ctx context.Context, convID, userID uuid.UUID, msgs []chatMessage) {
	data, err := json.Marshal(msgs)
	if err != nil {
		slog.Error("failed to marshal messages", "err", err)
		return
	}
	queries, commit, done, err := s.UserQueries(ctx, userID)
	if err != nil {
		slog.Error("failed to acquire user queries", "err", err)
		return
	}
	defer done()
	if err := queries.UpdateConversationMessagesByUser(ctx, db.UpdateConversationMessagesByUserParams{
		ID:       convID,
		Messages: data,
		UserID:   userID,
	}); err != nil {
		slog.Error("failed to persist messages", "err", err, "conversation_id", convID)
		return
	}
	if err := commit(); err != nil {
		slog.Error("failed to commit messages", "err", err, "conversation_id", convID)
	}
}

// wsOriginPatterns returns the list of allowed WebSocket origin patterns.
// Reads CORS_ALLOWED_ORIGINS (same env as the CORS middleware), falling back
// to the configured FrontendURL and localhost dev origins.
func wsOriginPatterns(frontendURL string) []string {
	if env := os.Getenv("CORS_ALLOWED_ORIGINS"); env != "" {
		var patterns []string
		for _, o := range strings.Split(env, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				patterns = append(patterns, o)
			}
		}
		return patterns
	}
	// Dev defaults — matches the CORS middleware fallback.
	patterns := []string{
		"http://localhost:5173",
		"http://localhost:4173",
	}
	if frontendURL != "" && frontendURL != "http://localhost:5173" {
		patterns = append(patterns, frontendURL)
	}
	return patterns
}

// writeWSError writes an error message over WebSocket.
func writeWSError(ctx context.Context, conn *websocket.Conn, message string) {
	if err := wsjson.Write(ctx, conn, map[string]string{
		"type":    "error",
		"content": message,
	}); err != nil {
		slog.Error("websocket write error", "err", err)
	}
}
