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
)
