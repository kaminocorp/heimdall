package memory

import "context"

// Service provides high-level memory operations for the agent.
type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{client: client}
}

func (s *Service) RecallSimilarIncidents(ctx context.Context, description string) ([]Memory, error) {
	return s.client.QueryMemories(ctx, description)
}

func (s *Service) RecallLessons(ctx context.Context, topic string) ([]Lesson, error) {
	return s.client.GetLessons(ctx, topic)
}

func (s *Service) RecordIncident(ctx context.Context, event Event) error {
	return s.client.RecordEvent(ctx, event)
}
