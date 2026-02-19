package handlers

import (
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func HandleChat(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		slog.Error("websocket accept error", "err", err)
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()

	for {
		var msg map[string]any
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			slog.Info("websocket read closed", "err", err)
			return
		}

		// TODO: route message to agent engine
		reply := map[string]any{
			"role":    "agent",
			"content": "Agent not yet connected. This is a scaffold placeholder.",
		}
		if err := wsjson.Write(ctx, conn, reply); err != nil {
			slog.Error("websocket write error", "err", err)
			return
		}
	}
}
