package agent

func (a *Agent) toolRecallSimilarIncidents(input map[string]any) (string, error) {
	// TODO: query Elephantasm for similar past incidents
	return "recall_similar_incidents tool not yet implemented", nil
}

func (a *Agent) toolRecallLessons(input map[string]any) (string, error) {
	// TODO: query Elephantasm for lessons on a topic
	return "recall_lessons tool not yet implemented", nil
}
