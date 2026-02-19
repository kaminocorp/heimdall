package ws

// MessageType identifies the kind of WebSocket message.
type MessageType string

const (
	MessageTypeChat    MessageType = "chat"
	MessageTypeLog     MessageType = "log"
	MessageTypeSystem  MessageType = "system"
)

// Message is the envelope for all WebSocket communication.
type Message struct {
	Type    MessageType    `json:"type"`
	Payload map[string]any `json:"payload"`
}
