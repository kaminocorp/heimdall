package agent

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// Stage identifiers for log_pipeline_events. Kept as typed constants so the
// four call sites (ingestion, classify, gate, assessment) can't drift via
// typos — a future `monitoring_cycle` stage would add one line here.
const (
	StageIngestion  = "ingestion"
	StageClassified = "classified"
	StageGate       = "gate"
	StageAssessment = "assessment"
)

// PipelineWriter persists a pipeline event to log_pipeline_events and then
// fans the event out via the in-memory PipelineBus. Both operations are
// fire-and-forget from the caller's perspective: errors are logged but
// never propagated, because a pipeline-event failure must not fail or
// degrade the underlying work (log ingestion, monitoring assessment).
//
// Order is persist-first-then-publish. If the DB write fails, we log and
// skip the publish. Better to lose a live particle than to render a
// particle for a log whose journey can't be reconstructed from
// log_pipeline_events later.
type PipelineWriter struct {
	queries *db.Queries
	bus     *PipelineBus
}

// NewPipelineWriter constructs a writer. Both dependencies are required —
// a nil bus would silently swallow all live events and a nil queries would
// skip all persistence.
func NewPipelineWriter(queries *db.Queries, bus *PipelineBus) *PipelineWriter {
	return &PipelineWriter{queries: queries, bus: bus}
}

// IngestionInput is the caller-supplied info for a StageIngestion write.
// Mirrors log_buffer columns; the writer handles nullable wrapping.
type IngestionInput struct {
	LogID      uuid.UUID
	AppID      uuid.UUID
	SourceType string
	Severity   string
}

// WriteIngestion records that a log landed in log_buffer for an app. Called
// from the webhook handler inside the ingestion transaction's trailing tail
// — after commit — because the event row references log_buffer.id via an
// FK that only resolves once the log insert has committed.
func (w *PipelineWriter) WriteIngestion(ctx context.Context, in IngestionInput) {
	if w == nil {
		return
	}
	row, err := w.queries.InsertPipelineEvent(ctx, db.InsertPipelineEventParams{
		LogID:      in.LogID,
		AppID:      in.AppID,
		Stage:      StageIngestion,
		SourceType: textOrNull(in.SourceType),
		Severity:   textOrNull(in.Severity),
		Metadata:   emptyMeta(),
	})
	if err != nil {
		slog.Warn("pipeline writer: failed to persist ingestion event",
			"err", err, "log_id", in.LogID, "app_id", in.AppID)
		return
	}
	w.publish(row)
}

// ClassifiedInput captures the Lumber classification result for one log.
type ClassifiedInput struct {
	LogID      uuid.UUID
	AppID      uuid.UUID
	Type       string
	Category   string
	Severity   string
	Confidence float64
	Summary    string
}

// WriteClassified records that Lumber returned a classification for the log.
// Emitted for *every* classified log (flagged and safe alike) so the
// Lumber-stage detail panel can show the full type/category histogram —
// not just the escalated slice.
func (w *PipelineWriter) WriteClassified(ctx context.Context, in ClassifiedInput) {
	if w == nil {
		return
	}
	row, err := w.queries.InsertPipelineEvent(ctx, db.InsertPipelineEventParams{
		LogID:      in.LogID,
		AppID:      in.AppID,
		Stage:      StageClassified,
		Severity:   textOrNull(in.Severity),
		Type:       textOrNull(in.Type),
		Category:   textOrNull(in.Category),
		Confidence: float8OrNull(in.Confidence),
		Summary:    textOrNull(in.Summary),
		Metadata:   emptyMeta(),
	})
	if err != nil {
		slog.Warn("pipeline writer: failed to persist classified event",
			"err", err, "log_id", in.LogID, "app_id", in.AppID)
		return
	}
	w.publish(row)
}

// GateInput captures the severity-gate decision. Escalated==true means the
// log will be forwarded to the LLM; RuleID identifies which branch fired.
type GateInput struct {
	LogID     uuid.UUID
	AppID     uuid.UUID
	Escalated bool
	RuleID    string
}

// WriteGate records the gate decision for a single log.
func (w *PipelineWriter) WriteGate(ctx context.Context, in GateInput) {
	if w == nil {
		return
	}
	row, err := w.queries.InsertPipelineEvent(ctx, db.InsertPipelineEventParams{
		LogID:     in.LogID,
		AppID:     in.AppID,
		Stage:     StageGate,
		Escalated: pgtype.Bool{Bool: in.Escalated, Valid: true},
		RuleHit:   textOrNull(in.RuleID),
		Metadata:  emptyMeta(),
	})
	if err != nil {
		slog.Warn("pipeline writer: failed to persist gate event",
			"err", err, "log_id", in.LogID, "app_id", in.AppID)
		return
	}
	w.publish(row)
}

// AssessmentInput captures the tail of the pipeline: a flagged log landed
// in an assessment batch that produced agent_log row AssessmentID.
type AssessmentInput struct {
	LogID        uuid.UUID
	AppID        uuid.UUID
	AssessmentID uuid.UUID
}

// WriteAssessment records that a flagged log was part of an LLM assessment.
// One call per flagged log per batch — never per-batch aggregate — so the
// Time Machine view can link each log back to the assessment that evaluated
// it.
func (w *PipelineWriter) WriteAssessment(ctx context.Context, in AssessmentInput) {
	if w == nil {
		return
	}
	assessmentID := in.AssessmentID
	row, err := w.queries.InsertPipelineEvent(ctx, db.InsertPipelineEventParams{
		LogID:        in.LogID,
		AppID:        in.AppID,
		Stage:        StageAssessment,
		AssessmentID: &assessmentID,
		Metadata:     emptyMeta(),
	})
	if err != nil {
		slog.Warn("pipeline writer: failed to persist assessment event",
			"err", err, "log_id", in.LogID, "app_id", in.AppID)
		return
	}
	w.publish(row)
}

func (w *PipelineWriter) publish(row db.LogPipelineEvent) {
	if w.bus == nil {
		return
	}
	w.bus.Publish(rowToEvent(row))
}

// rowToEvent converts the sqlc row type into the bus's in-memory shape.
// Kept in one place so the frontend SSE payload schema is derivable from
// a single Go struct.
func rowToEvent(row db.LogPipelineEvent) PipelineEvent {
	evt := PipelineEvent{
		ID:           row.ID,
		LogID:        row.LogID,
		AppID:        row.AppID,
		Stage:        row.Stage,
		OccurredAt:   row.OccurredAt,
		AssessmentID: row.AssessmentID,
	}
	if row.SourceType.Valid {
		evt.SourceType = row.SourceType.String
	}
	if row.Severity.Valid {
		evt.Severity = row.Severity.String
	}
	if row.Type.Valid {
		evt.Type = row.Type.String
	}
	if row.Category.Valid {
		evt.Category = row.Category.String
	}
	if row.Confidence.Valid {
		evt.Confidence = row.Confidence.Float64
	}
	if row.Summary.Valid {
		evt.Summary = row.Summary.String
	}
	if row.Escalated.Valid {
		evt.HasGate = true
		evt.Escalated = row.Escalated.Bool
	}
	if row.RuleHit.Valid {
		evt.RuleHit = row.RuleHit.String
	}
	return evt
}

func textOrNull(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func float8OrNull(f float64) pgtype.Float8 {
	return pgtype.Float8{Float64: f, Valid: true}
}

// emptyMeta returns the NOT-NULL default for metadata. Using `{}` explicitly
// keeps the INSERT deterministic — depending on the server-side DEFAULT
// would be fine too, but the Phase 3 server-side 5s cache key is easier to
// reason about when the payload is byte-for-byte stable per stage.
func emptyMeta() json.RawMessage {
	return json.RawMessage(`{}`)
}
