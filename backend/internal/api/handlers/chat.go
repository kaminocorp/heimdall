package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
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
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		slog.Error("websocket accept error", "err", err)
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()

	// Load or create conversation.
	var convID uuid.UUID
	var storedMessages []chatMessage

	queries, done, err := s.UserQueries(ctx, userID)
	if err != nil {
		writeWSError(ctx, conn, "database error")
		return
	}

	if cidStr := r.URL.Query().Get("conversation_id"); cidStr != "" {
		cid, err := uuid.Parse(cidStr)
		if err != nil {
			done()
			writeWSError(ctx, conn, "invalid conversation_id")
			return
		}
		conv, err := queries.GetConversationByUser(ctx, db.GetConversationByUserParams{
			ID:     cid,
			UserID: userID,
		})
		if err != nil {
			done()
			writeWSError(ctx, conn, "conversation not found")
			return
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
			done()
			slog.Error("failed to create conversation", "err", err)
			writeWSError(ctx, conn, "failed to create conversation")
			return
		}
		convID = conv.ID
		storedMessages = []chatMessage{}
	}
	done()

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
			if len(title) > 50 {
				title = title[:50] + "..."
			}
			if q, d, err := s.UserQueries(ctx, userID); err == nil {
				q.UpdateConversationTitleByUser(ctx, db.UpdateConversationTitleByUserParams{
					ID:     convID,
					Title:  pgtype.Text{String: title, Valid: true},
					UserID: userID,
				})
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

		// Run agent with conversation history.
		response, err := s.Agent.RunConversation(ctx, userID, &convID, history, msg.Content)
		if err != nil {
			slog.Error("agent error", "err", err, "user_id", userID, "conversation_id", convID)
			if writeErr := wsjson.Write(ctx, conn, map[string]string{
				"type":    "error",
				"content": "Agent error: " + err.Error(),
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
			Content:   response,
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
	queries, done, err := s.UserQueries(ctx, userID)
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
	}
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
