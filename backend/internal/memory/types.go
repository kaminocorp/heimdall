package memory

import "time"

type Event struct {
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Result    string    `json:"result,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type Memory struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Score     float64   `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

type Lesson struct {
	ID      string `json:"id"`
	Topic   string `json:"topic"`
	Content string `json:"content"`
}
