package agent

import (
	"log/slog"

	"github.com/kaminocorp/lumber/pkg/lumber"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// LumberClassifier wraps the Lumber library to classify log entries.
type LumberClassifier struct {
	engine *lumber.Lumber
}

// NewLumberClassifier creates a new classifier backed by the Lumber ONNX model.
// modelDir must contain model_quantized.onnx, vocab.txt, and 2_Dense/model.safetensors.
func NewLumberClassifier(modelDir string) (*LumberClassifier, error) {
	l, err := lumber.New(
		lumber.WithModelDir(modelDir),
		lumber.WithVerbosity("minimal"),
	)
	if err != nil {
		return nil, err
	}
	return &LumberClassifier{engine: l}, nil
}

// Classify returns one ClassifiedLog per input — both escalated and safe.
// The monitor loop uses FilterFlagged to pull out the LLM-bound subset;
// the Pipeline page uses the full slice to emit per-log Lumber + Gate
// stage events (including for safe logs).
func (c *LumberClassifier) Classify(logs []db.LogBuffer) []ClassifiedLog {
	if len(logs) == 0 {
		return nil
	}

	texts := ExtractTexts(logs)

	events, err := c.engine.ClassifyBatch(texts)
	if err != nil {
		slog.Error("lumber classification failed, escalating all logs", "err", err, "count", len(logs))
		results := make([]ClassifiedLog, len(logs))
		for i, log := range logs {
			results[i] = ClassifiedLog{
				Log:       log,
				Type:      "UNCLASSIFIED",
				Severity:  "warning",
				Summary:   texts[i],
				Escalated: true,
				RuleID:    RuleUnclassified,
			}
		}
		return results
	}

	if len(events) != len(logs) {
		slog.Error("lumber: ClassifyBatch returned mismatched count, escalating all",
			"expected", len(logs), "got", len(events))
		results := make([]ClassifiedLog, len(logs))
		for i, log := range logs {
			results[i] = ClassifiedLog{
				Log:       log,
				Type:      "UNCLASSIFIED",
				Severity:  "warning",
				Summary:   texts[i],
				Escalated: true,
				RuleID:    RuleUnclassified,
			}
		}
		return results
	}

	results := make([]ClassifiedLog, len(logs))
	for i, event := range events {
		escalated, ruleID := ShouldEscalate(event)
		results[i] = ClassifiedLog{
			Log:        logs[i],
			Type:       event.Type,
			Category:   event.Category,
			Severity:   event.Severity,
			Confidence: event.Confidence,
			Summary:    event.Summary,
			Escalated:  escalated,
			RuleID:     ruleID,
		}
	}

	return results
}

func (c *LumberClassifier) Close() error {
	return c.engine.Close()
}
