package reports

// Generator builds incident reports from agent analysis.
type Generator struct {
	// TODO: add dependencies (db queries)
}

func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a report for a resolved investigation.
func (g *Generator) Generate(investigationID string) (map[string]any, error) {
	// TODO: fetch investigation, format into structured report
	return nil, nil
}
