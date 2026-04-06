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

func (c *LumberClassifier) Classify(logs []db.LogBuffer) ([]ClassifiedLog, int) {
	if len(logs) == 0 {
		return nil, 0
	}

	texts := ExtractTexts(logs)

	events, err := c.engine.ClassifyBatch(texts)
	if err != nil {
		slog.Error("lumber classification failed, escalating all logs", "err", err, "count", len(logs))
		flagged := make([]ClassifiedLog, len(logs))
		for i, log := range logs {
			flagged[i] = ClassifiedLog{
				Log:      log,
				Type:     "UNCLASSIFIED",
				Severity: "warning",
				Summary:  texts[i],
			}
		}
		return flagged, 0
	}

	var flagged []ClassifiedLog
	safeCount := 0

	for i, event := range events {
		// We own the confidence threshold here rather than delegating to Lumber's
		// WithConfidenceThreshold option, so the policy stays in one place.
		if event.Confidence < 0.5 || ShouldEscalate(event) {
			flagged = append(flagged, ClassifiedLog{
				Log:        logs[i],
				Type:       event.Type,
				Category:   event.Category,
				Severity:   event.Severity,
				Confidence: event.Confidence,
				Summary:    event.Summary,
			})
		} else {
			safeCount++
		}
	}

	return flagged, safeCount
}

func (c *LumberClassifier) Close() error {
	return c.engine.Close()
}
