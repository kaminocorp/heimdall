package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// sourceFilterStats summarises what a filter pipeline did with a batch of
// entries. Handler-specific response bodies read from it; the shared pipeline
// has no opinion about HTTP shape.
type sourceFilterStats struct {
	// Accepted counts entries that matched at least one enabled filter and
	// were inserted into log_buffer (before fan-out).
	Accepted int
	// Filtered counts entries dropped because their source name wasn't
	// enabled in any app.
	Filtered int
	// Inserts is the post-fan-out row count — equals Accepted for app-scoped
	// connections, and Accepted × (avg apps per source) for org-scoped.
	Inserts int
	// Sources counts distinct source names seen in the batch — used for
	// logging (visibility into how multi-source a batch was).
	Sources int
}

// insertedRow captures the minimal fields the Pipeline-page ingestion-stage
// emit needs. Collected during fan-out and emitted by the caller after
// commit — an FK-bound pipeline event can only be inserted once its
// referenced log_buffer row has committed.
type insertedRow struct {
	LogID      uuid.UUID
	AppID      uuid.UUID
	SourceType string
	Severity   string
}

// routeSources implements the discover-then-filter step shared by every
// source-filtered ingestion path (webhook, OTLP, syslog). It:
//
//  1. Upserts each distinct source_name in the batch into
//     connection_sources so the UI selector always surfaces newly-arriving
//     sources, even when they're currently disabled. Without this pass,
//     drop-by-default would be user-hostile — people would see "nothing
//     enabled" with no way to know what's available to enable.
//  2. Builds a map from source_name → []app_id describing *where each
//     source's entries should be inserted*. The map encodes both filter
//     state and org-scoping: app-scoped connections yield 0 or 1 targets
//     per source; org-scoped connections yield 0–N targets.
//
// `appID == nil` means the connection is org-scoped — the function uses
// ListAppsEnabledForSource per distinct source name (one query per source,
// not per entry). `appID != nil` means app-scoped — one ListEnabledSourceNames
// call for the whole batch.
//
// Caller is responsible for running inside a transaction (`qtx`) — all
// writes (source upserts and the downstream log inserts) must commit or
// roll back atomically, so we never partially persist a batch.
func routeSources(
	ctx context.Context,
	qtx *db.Queries,
	connID uuid.UUID,
	appID *uuid.UUID,
	entries []webhookLogRequest,
) (routes map[string][]uuid.UUID, seenSources map[string]struct{}, err error) {
	seenSources = make(map[string]struct{}, len(entries))
	for _, e := range entries {
		name := sourceNameOf(e)
		if name == "" {
			continue
		}
		if _, dup := seenSources[name]; dup {
			continue
		}
		seenSources[name] = struct{}{}
		if _, err := qtx.UpsertConnectionSource(ctx, db.UpsertConnectionSourceParams{
			ConnectionID: connID,
			SourceName:   name,
		}); err != nil {
			return nil, nil, fmt.Errorf("record connection source: %w", err)
		}
	}

	routes = make(map[string][]uuid.UUID, len(seenSources))
	if appID != nil {
		enabledNames, err := qtx.ListEnabledSourceNames(ctx, db.ListEnabledSourceNamesParams{
			ConnectionID: connID,
			AppID:        *appID,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("load source filters: %w", err)
		}
		for _, n := range enabledNames {
			routes[n] = []uuid.UUID{*appID}
		}
	} else {
		for name := range seenSources {
			apps, err := qtx.ListAppsEnabledForSource(ctx, db.ListAppsEnabledForSourceParams{
				ConnectionID: connID,
				SourceName:   name,
			})
			if err != nil {
				return nil, nil, fmt.Errorf("resolve source routes: %w", err)
			}
			if len(apps) > 0 {
				routes[name] = apps
			}
		}
	}

	return routes, seenSources, nil
}

// insertFiltered runs the actual log_buffer inserts for entries whose source
// name has at least one target app. Entries without a route fall through to
// `filtered++`. Fan-out inserts happen inside the iteration so that a
// failure halfway rolls back the whole batch via the caller's transaction.
//
// Returns the post-commit emit list (one entry per inserted log_buffer row)
// so the caller can fire Pipeline-page ingestion events after tx.Commit.
// Emitting inside the transaction would race with the FK-bound
// log_pipeline_events insert — the log_buffer row isn't visible to a
// separate connection until commit.
func insertFiltered(
	ctx context.Context,
	qtx *db.Queries,
	connID uuid.UUID,
	userID uuid.UUID,
	entries []webhookLogRequest,
	routes map[string][]uuid.UUID,
	seenSources map[string]struct{},
) (sourceFilterStats, []insertedRow, error) {
	stats := sourceFilterStats{Sources: len(seenSources)}
	var inserted []insertedRow
	for _, e := range entries {
		name := sourceNameOf(e)
		targets := routes[name]
		if len(targets) == 0 {
			stats.Filtered++
			continue
		}

		var severity pgtype.Text
		if e.Severity != "" {
			severity = pgtype.Text{String: e.Severity, Valid: true}
		}

		for _, appID := range targets {
			row, err := qtx.InsertLogEntry(ctx, db.InsertLogEntryParams{
				ConnectionID: connID,
				SourceType:   e.SourceType,
				Severity:     severity,
				Payload:      e.Payload,
				UserID:       userID,
				AppID:        appID,
			})
			if err != nil {
				return stats, nil, fmt.Errorf("insert log entry: %w", err)
			}
			inserted = append(inserted, insertedRow{
				LogID:      row.ID,
				AppID:      appID,
				SourceType: e.SourceType,
				Severity:   e.Severity,
			})
			stats.Inserts++
		}
		stats.Accepted++
	}
	return stats, inserted, nil
}
