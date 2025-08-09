package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	KafkaEventsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "events_processed_total_kafka",
			Help: "Total number of kafka events processed.",
		},
	)

	KafkaProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "event_processing_duration_seconds_kafka",
			Help:    "Duration of event processing in seconds for kafka.",
			Buckets: prometheus.DefBuckets,
		},
	)

	EventBridgeEventsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "events_processed_total_eventbridge",
			Help: "Total number of eventbridge events processed.",
		},
	)

	EventBridgeProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "event_processing_duration_seconds_eventbridge",
			Help:    "Duration of event processing in seconds for eventbridge.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func Init() {
	prometheus.MustRegister(KafkaEventsProcessed)
	prometheus.MustRegister(KafkaProcessingDuration)
	prometheus.MustRegister(EventBridgeEventsProcessed)
	prometheus.MustRegister(EventBridgeProcessingDuration)
}

func StartServer(port string) {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+port, nil)
}
