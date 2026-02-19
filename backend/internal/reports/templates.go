package reports

// ReportTemplate defines the structure of an incident report.
type ReportTemplate struct {
	Sections []Section
}

type Section struct {
	Title string
	Key   string
}

// DefaultTemplate returns the standard incident report structure.
func DefaultTemplate() ReportTemplate {
	return ReportTemplate{
		Sections: []Section{
			{Title: "Summary", Key: "summary"},
			{Title: "Timeline", Key: "timeline"},
			{Title: "Investigation", Key: "investigation"},
			{Title: "Findings", Key: "findings"},
			{Title: "Historical Context", Key: "historical_context"},
			{Title: "Diagnostic Assessment", Key: "assessment"},
		},
	}
}
