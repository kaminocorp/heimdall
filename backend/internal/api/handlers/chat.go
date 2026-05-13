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
	// Authorisation goes through UserQueries so the GetApplicationByOrgUser
	// read sees app.current_user_id and the post-Phase-7 RLS policy
	// evaluates correctly.
	var appID uuid.UUID
	if appIDStr := r.URL.Query().Get("app_id"); appIDStr != "" {
		parsed, err := uuid.Parse(appIDStr)
		if err == nil {
			authzErr := s.Pools.WithUserQueries(r.Context(), userID, func(q *db.Queries) error {
				_, e := q.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
					AppID:  parsed,
					UserID: userID,
				})
				return e
			})
			if authzErr != nil {
				writeWSError(r.Context(), conn, "app not found or access denied")
				return
			}
			appID = parsed
		}
	}

	ctx := r.Context()

	// Load or create conversation. setupConversation owns its own short
	// UserQueries — it runs once per WebSocket connection, not per
	// message, so it doesn't share scope with the per-message loop.
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

	// Per-message loop. Each iteration runs the entire user→agent→persist
	// cycle inside one UserQueriesForLoop transaction (parent plan §5.3b
	// Option A). The transaction's idle_in_txn timeout is disabled so a
	// slow LLM round-trip can't abort it mid-cycle. The agent goroutine
	// and this loop share the *db.Queries handle but never use it
	// concurrently: the handler writes pre-agent (title, pre-persist),
	// drains the event channel synchronously while the agent goroutine
	// owns q exclusively, then writes post-agent (post-persist, commit)
	// after the channel closes.
	for {
		var msg incomingMessage
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			slog.Info("websocket read closed", "err", err, "user_id", userID)
			return
		}

		if msg.Content == "" {
			continue
		}

		s.handleChatMessage(ctx, conn, userID, appID, convID, &storedMessages, msg.Content)
	}
}

// handleChatMessage runs one user→agent round trip inside one
// UserQueriesForLoop transaction. Extracted into its own method so
// `defer done()` cleanly bounds the transaction lifetime to a single
// message cycle — putting the defer inside the for-loop body in
// HandleChat would only fire on function return.
func (s *Server) handleChatMessage(ctx context.Context, conn *websocket.Conn, userID uuid.UUID, appID uuid.UUID, convID uuid.UUID, storedMessagesPtr *[]chatMessage, content string) {
	storedMessages := *storedMessagesPtr

	// Build user message and append to local stored messages.
	userMsg := chatMessage{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   content,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	storedMessages = append(storedMessages, userMsg)

	q, commit, done, err := s.Pools.UserQueriesForLoop(ctx, userID)
	if err != nil {
		slog.Error("chat: failed to open UserQueriesForLoop", "err", err, "conversation_id", convID)
		return
	}
	defer done()

	// Persist the user message before the agent runs so a mid-loop crash
	// still leaves the conversation auditable. The whole txn rolls back
	// if commit isn't called, so this isn't a partial commit.
	if err := persistMessagesTx(ctx, q, convID, userID, storedMessages); err != nil {
		slog.Error("chat: failed to persist user message", "err", err, "conversation_id", convID)
		return
	}

	// Set conversation title from first user message.
	if len(storedMessages) == 1 {
		title := userMsg.Content
		titleRunes := []rune(title)
		if len(titleRunes) > 50 {
			title = string(titleRunes[:50]) + "..."
		}
		if err := q.UpdateConversationTitleByUser(ctx, db.UpdateConversationTitleByUserParams{
			ID:     convID,
			Title:  pgtype.Text{String: title, Valid: true},
			UserID: userID,
		}); err != nil {
			slog.Error("failed to set conversation title", "err", err, "conversation_id", convID)
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

	// Convert stored messages to agent history (exclude the current
	// message — it goes as input).
	history := toAgentHistory(storedMessages[:len(storedMessages)-1])

	// Run the agent inside the same transaction. The agent goroutine
	// has exclusive use of q while the for-range drains its events;
	// we resume using q for post-persist after the channel closes.
	//
	// Concurrency invariant: the producer goroutine in
	// RunConversationStream owns q exclusively while it is running, and
	// pgx.Tx is not safe for concurrent use. If we returned here while
	// the producer were still running, defer done() would race tx.Rollback
	// against the producer's next q.X call. To avoid that, every early
	// return below MUST happen after the producer has finished — which
	// is signalled by the events channel closing (defer close(ch) in
	// loop_stream.go). The wsWriteErr / drainEvents pattern below
	// preserves the early-return semantics without violating the
	// invariant.
	events := s.Agent.RunConversationStream(ctx, q, userID, appID, &convID, history, content)

	var fullResponse string
	var streamErr string
	var wsWriteErr error
	for ev := range events {
		// Once a WS write has failed we can't talk to the client any
		// more, but we must still drain `events` to completion so the
		// producer goroutine returns and stops touching q before
		// done() rolls back the txn.
		if wsWriteErr != nil {
			if ev.Type == "message" {
				fullResponse = ev.Content
			} else if ev.Type == "error" {
				streamErr = ev.Content
			}
			continue
		}
		switch ev.Type {
		case "tool_start":
			if err := wsjson.Write(ctx, conn, map[string]string{
				"type": "tool_start",
				"tool": ev.Tool,
			}); err != nil {
				slog.Error("websocket write error", "err", err)
				wsWriteErr = err
			}
		case "tool_result":
			if err := wsjson.Write(ctx, conn, map[string]any{
				"type":     "tool_result",
				"tool":     ev.Tool,
				"is_error": ev.IsError,
			}); err != nil {
				slog.Error("websocket write error", "err", err)
				wsWriteErr = err
			}
		case "message":
			fullResponse = ev.Content
		case "error":
			streamErr = ev.Content
		}
	}

	// If the WS write failed mid-stream, return now — *after* the
	// producer has finished (events channel closed). done() runs
	// safely because nothing is using q any more.
	if wsWriteErr != nil {
		return
	}

	if streamErr != "" {
		slog.Error("agent error", "err", streamErr, "user_id", userID, "conversation_id", convID)
		// Don't commit — the message persist + agent_log emits should
		// roll back together so a failed turn doesn't leave dangling
		// state.
		if writeErr := wsjson.Write(ctx, conn, map[string]string{
			"type":    "error",
			"content": "Agent error: " + streamErr,
		}); writeErr != nil {
			slog.Error("websocket write error", "err", writeErr)
		}
		return
	}

	// Build agent message and append to stored messages.
	agentMsg := chatMessage{
		ID:        uuid.New().String(),
		Role:      "assistant",
		Content:   fullResponse,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	storedMessages = append(storedMessages, agentMsg)

	if err := persistMessagesTx(ctx, q, convID, userID, storedMessages); err != nil {
		slog.Error("chat: failed to persist agent message", "err", err, "conversation_id", convID)
		return
	}

	if err := commit(); err != nil {
		slog.Error("chat: failed to commit message cycle", "err", err, "conversation_id", convID)
		return
	}

	*storedMessagesPtr = storedMessages

	// Send agent response to client (role "agent" for frontend compatibility).
	if err := wsjson.Write(ctx, conn, chatMessage{
		ID:        agentMsg.ID,
		Role:      "agent",
		Content:   agentMsg.Content,
		Timestamp: agentMsg.Timestamp,
	}); err != nil {
		slog.Error("websocket write error", "err", err)
	}
}

// setupConversation loads an existing conversation or creates a new one
// inside a single transactional scope. Runs once per WebSocket
// connection (pre-message-loop).
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

// persistMessagesTx writes the conversation's JSONB column inside a
// caller-owned transaction. Replaces the pre-Phase-2 persistMessages
// helper that opened its own UserQueries — under Option A the
// per-message cycle is one transaction so the persist must use the
// caller's queries handle.
func persistMessagesTx(ctx context.Context, q *db.Queries, convID, userID uuid.UUID, msgs []chatMessage) error {
	data, err := json.Marshal(msgs)
	if err != nil {
		return err
	}
	return q.UpdateConversationMessagesByUser(ctx, db.UpdateConversationMessagesByUserParams{
		ID:       convID,
		Messages: data,
		UserID:   userID,
	})
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
