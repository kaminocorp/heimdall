package agent

import (
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// Classifier classifies a batch of log entries. The returned slice contains
// one ClassifiedLog per input log (both escalated and safe) with the
// Escalated bit set — callers that only care about the flagged subset use
// FilterFlagged to partition. Emitting per-log rows (Pipeline page's
// Lumber + Gate stages) needs the full list; the monitor-loop LLM path
// only needs the flagged subset.
type Classifier interface {
	Classify(logs []db.LogBuffer) []ClassifiedLog

	// Close releases resources (ONNX runtime, model memory).
	Close() error
}

// FilterFlagged partitions a Classify result into the escalated subset and
// the count of safe entries. Kept as a free function rather than a method
// on []ClassifiedLog to avoid locking a slice type — callers that want
// different partitions (e.g. by RuleID) build their own iteration.
func FilterFlagged(results []ClassifiedLog) (flagged []ClassifiedLog, safeCount int) {
	for _, r := range results {
		if r.Escalated {
			flagged = append(flagged, r)
		} else {
			safeCount++
		}
	}
	return flagged, safeCount
}

// ClassifiedLog pairs a log buffer entry with its Lumber classification
// metadata and the severity-gate decision. Escalated flips to true when
// ShouldEscalate fires; RuleID names the branch that fired (or RuleNone
// for safe logs) — used by pipeline_writer to populate
// log_pipeline_events.rule_hit on the Gate-stage row.
type ClassifiedLog struct {
	Log        db.LogBuffer
	Type       string  // Lumber root category: ERROR, REQUEST, DEPLOY, etc.
	Category   string  // Lumber leaf label: connection_failure, success, etc.
	Severity   string  // error, warning, info, debug
	Confidence float64 // Cosine similarity score (0.0–1.0)
	Summary    string  // Lumber's compacted summary of the log
	Escalated  bool    // True if the severity gate flagged this log for the LLM
	RuleID     string  // Escalation rule id, or RuleNone for safe logs
}

// PassthroughClassifier escalates all logs without classification.
// Used when classifier mode is "off" or when Lumber is unavailable in "fallback" mode.
type PassthroughClassifier struct{}

func (p *PassthroughClassifier) Classify(logs []db.LogBuffer) []ClassifiedLog {
	results := make([]ClassifiedLog, len(logs))
	for i, log := range logs {
		results[i] = ClassifiedLog{
			Log:       log,
			Type:      "UNCLASSIFIED",
			Severity:  "unknown",
			Summary:   ExtractText(log),
			Escalated: true,
			RuleID:    RuleUnclassified,
		}
	}
	return results
}

func (p *PassthroughClassifier) Close() error { return nil }
