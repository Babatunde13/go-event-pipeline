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
		[]string{"system"}, // system = kafka | eventbridge
	)

	eventDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "event_duration_seconds",
			Help:    "Event processing duration in seconds",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 20), // 0.1s to 2s
		},
		[]string{"system"},
	)
)

func PushMetrics(url string, duration float64, isKafka, isProducer, success bool) {
	var system string
	if isKafka {
		system = "kafka"
	} else {
		system = "eventbridge"
	}
	jobName := "consumers"
	if isProducer {
		jobName = "producers"
	}
	eventDuration.WithLabelValues(system).Observe(duration)
	if success {
		totalEvents.WithLabelValues(system).Inc()
	}
	err := push.New(url, jobName).
		Collector(totalEvents).
		Collector(eventDuration).
		Grouping("system", system).
		Grouping("producer", jobName).
		Push()

	if err != nil {
		log.Printf("Failed to push metrics to Pushgateway: %v", err)
	}
}
