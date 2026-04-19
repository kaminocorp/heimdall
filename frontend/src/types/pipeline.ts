// Mirrors backend pipelineEventJSON in backend/internal/api/handlers/pipeline.go.
// Optional fields use `omitempty` on the wire — they stay undefined here when
// the stage doesn't emit them (e.g. severity is only set on ingestion).

export type PipelineStage = 'ingestion' | 'classified' | 'gate' | 'assessment'

export interface PipelineEvent {
  id: string
  log_id: string
  app_id: string
  stage: PipelineStage
  occurred_at: string
  source_type?: string
  severity?: string
  type?: string
  category?: string
  confidence?: number
  summary?: string
  escalated?: boolean
  rule_hit?: string
  assessment_id?: string
}

export interface PipelineStats {
  ingestion_count: number
  classified_count: number
  flagged_count: number
  safe_count: number
  assessment_count: number
  avg_confidence: number
  window_seconds: number
}

export interface PipelineBootstrap {
  stats: PipelineStats
  recent_events: PipelineEvent[]
  cursor: string
}

export interface PipelineJourney {
  log_id: string
  events: PipelineEvent[]
}

// One row per log for the Time Machine picker. Mirrors backend
// pipelineLogSummaryJSON — the GROUP BY collapses a log's ≤4 stage events
// into a single summary with fields sourced from the matching stage row.
export interface PipelineLogSummary {
  log_id: string
  first_seen_at: string
  last_seen_at: string
  stage_count: number
  source_type?: string
  severity?: string
  type?: string
  category?: string
  confidence?: number
  summary?: string
  escalated?: boolean
  rule_hit?: string
  assessment_id?: string
}

export interface PipelineLogsResponse {
  logs: PipelineLogSummary[]
  since: string
  until: string
  limit: number
  offset: number
  truncated: boolean
}

export type ConnectionState =
  | 'idle'
  | 'bootstrapping'
  | 'streaming'
  | 'reconnecting'
  | 'resyncing'
  | 'error'
