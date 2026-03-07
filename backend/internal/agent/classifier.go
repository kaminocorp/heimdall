package agent

import (
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// Classifier classifies a batch of log entries and returns which ones
// should be escalated to the LLM agent.
type Classifier interface {
	// Classify takes a batch of log buffer entries, classifies each one,
	// and returns the flagged (escalation-worthy) logs plus the count of safe logs.
	Classify(logs []db.LogBuffer) (flagged []ClassifiedLog, safeCount int)

	// Close releases resources (ONNX runtime, model memory).
	Close() error
}

// ClassifiedLog pairs a log buffer entry with its Lumber classification metadata.
type ClassifiedLog struct {
	Log        db.LogBuffer
	Type       string  // Lumber root category: ERROR, REQUEST, DEPLOY, etc.
	Category   string  // Lumber leaf label: connection_failure, success, etc.
	Severity   string  // error, warning, info, debug
	Confidence float64 // Cosine similarity score (0.0–1.0)
	Summary    string  // Lumber's compacted summary of the log
}

// PassthroughClassifier escalates all logs without classification.
// Used when classifier mode is "off" or when Lumber is unavailable in "fallback" mode.
type PassthroughClassifier struct{}

func (p *PassthroughClassifier) Classify(logs []db.LogBuffer) ([]ClassifiedLog, int) {
	flagged := make([]ClassifiedLog, len(logs))
	for i, log := range logs {
		flagged[i] = ClassifiedLog{
			Log:      log,
			Type:     "UNCLASSIFIED",
			Severity: "unknown",
			Summary:  ExtractText(log),
		}
	}
	return flagged, 0
}

func (p *PassthroughClassifier) Close() error { return nil }
