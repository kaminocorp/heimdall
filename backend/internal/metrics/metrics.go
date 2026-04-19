package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// MonitorTickDuration tracks how long each full monitoring tick takes (across all apps).
	MonitorTickDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "heimdall_monitor_tick_duration_seconds",
		Help:    "Duration of a single monitoring tick across all active applications.",
		Buckets: prometheus.DefBuckets,
	})

	// LogsClassifiedTotal counts log entries processed by the classifier.
	// Label "result" is either "safe" or "flagged".
	LogsClassifiedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "heimdall_logs_classified_total",
		Help: "Total log entries processed by the classifier, by result.",
	}, []string{"result"})

	// NotificationsTotal counts notification dispatch attempts.
	// Label "status" is "sent" or "failed"; "channel_type" is "email", "slack", or "discord".
	NotificationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "heimdall_notifications_total",
		Help: "Total notification dispatch attempts, by status and channel type.",
	}, []string{"status", "channel_type"})

	// PipelineBusDroppedEvents counts PipelineBus events that were dropped
	// because a subscriber's channel was full. Non-zero values indicate at
	// least one SSE consumer is stalled relative to the producer (the
	// monitor-tick + ingestion emit rate). Counter, not gauge — we care
	// about the total across a rollout window, not the instantaneous
	// snapshot.
	PipelineBusDroppedEvents = promauto.NewCounter(prometheus.CounterOpts{
		Name: "heimdall_pipeline_bus_dropped_events_total",
		Help: "Total pipeline-bus events dropped due to full subscriber channels.",
	})

	// PipelineSSEActiveConnections is the live count of PipelineStream SSE
	// connections currently held open, labelled by user (not emitted per
	// user to avoid cardinality — aggregated across all users). Gauge:
	// increments on connect, decrements on disconnect.
	PipelineSSEActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "heimdall_pipeline_sse_active_connections",
		Help: "Current number of open Pipeline SSE connections across all users.",
	})

	// PipelineSSERejectedTotal counts SSE connection attempts that were
	// rejected for exceeding the per-user cap. A non-trivial rise here is
	// either a runaway tab (user opening the Pipeline page many times) or
	// an attacker probing the stream endpoint.
	PipelineSSERejectedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "heimdall_pipeline_sse_rejected_total",
		Help: "Total Pipeline SSE connection attempts rejected by the per-user cap.",
	})

	// PipelineBootstrapCache tracks /pipeline/bootstrap cache outcomes.
	// Label "result" is "hit" or "miss". A low hit rate would mean the TTL
	// is tuned too short (or traffic is spread across too many cache keys
	// — e.g. users picking custom windows); a high hit rate validates the
	// caching decision.
	PipelineBootstrapCache = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "heimdall_pipeline_bootstrap_cache_total",
		Help: "Pipeline /bootstrap cache outcomes, by result.",
	}, []string{"result"})
)
