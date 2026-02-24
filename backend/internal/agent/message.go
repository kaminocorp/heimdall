package agent

// Message is a domain type representing a single chat message
// stored in the conversations JSONB column.
type Message struct {
	ID        string `json:"id"`
	Role      string `json:"role"`      // "user" or "assistant"
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}
