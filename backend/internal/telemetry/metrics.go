package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	IngestReceivedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ingest_received_total",
			Help: "Total count of MQTT events received by the ingestion gateway",
		},
		[]string{"line_id", "event_type"},
	)

	IngestDuplicateTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ingest_duplicate_total",
			Help: "Total count of duplicate events discarded idempotently",
		},
	)

	IngestInvalidTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ingest_invalid_total",
			Help: "Total count of malformed or invalid events sent to dead letter",
		},
	)

	IngestLatencySeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ingest_latency_seconds",
			Help:    "Latency from simulation event occurred_at to database persistence",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		},
	)

	EventToDashboardSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "event_to_dashboard_seconds",
			Help:    "End-to-end latency from machine event publication to WebSocket dispatch",
			Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0},
		},
	)

	WSClients = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ws_clients",
			Help: "Current number of active WebSocket dashboard subscribers",
		},
	)

	IncidentOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "incident_open",
			Help: "Current count of active downtime incidents",
		},
	)
)
