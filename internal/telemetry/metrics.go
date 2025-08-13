package telemetry

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	totalEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_events",
			Help: "Total number of processed events",
		},
		[]string{"system", "role"}, // system = kafka | eventbridge, role = producer | consumer
	)

	eventDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "event_duration_seconds",
			Help:    "Event processing duration in seconds",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 20), // 0.1s to 2s
		},
		[]string{"system", "role"},
	)
)

func PushMetrics(url string, duration float64, isKafka, isProducer, success bool) {
	system := "eventbridge"
	if isKafka {
		system = "kafka"
	}
	role := "consumers"
	if isProducer {
		role = "producers"
	}
	eventDuration.WithLabelValues(system, role).Observe(duration)
	if success {
		totalEvents.WithLabelValues(system, role).Inc()
	}
	err := push.New(url, "event_pipeline").
		Collector(totalEvents).
		Collector(eventDuration).
		Push()

	if err != nil {
		log.Printf("Failed to push metrics to Pushgateway: %v", err)
	}
}
