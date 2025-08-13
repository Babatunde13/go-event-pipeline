package telemetry

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	kafkaEventsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "events_processed_total_kafka",
			Help: "Total number of kafka events processed.",
		},
	)

	kafkaProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "event_latency_seconds_kafka",
			Help:    "Duration of event processing in seconds for kafka.",
			Buckets: prometheus.DefBuckets,
		},
	)

	eventBridgeEventsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "events_processed_total_eventbridge",
			Help: "Total number of eventbridge events processed.",
		},
	)

	eventBridgeProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "event_latency_seconds_eventbridge",
			Help:    "Duration of event processing in seconds for eventbridge.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func PushMetrics(url string, duration float64, isKafka, isProducer, success bool) {
	var counter prometheus.Counter
	var durationMetric prometheus.Histogram
	if isKafka {
		counter = kafkaEventsProcessed
		durationMetric = kafkaProcessingDuration
	} else {
		counter = eventBridgeEventsProcessed
		durationMetric = eventBridgeProcessingDuration
	}
	jobName := "consumers"
	if isProducer {
		jobName = "producers"
	}
	durationMetric.Observe(duration)
	if success {
		counter.Inc()
	}
	err := push.New(url, jobName).
		Collector(counter).
		Collector(durationMetric).
		Push()
	if err != nil {
		log.Printf("Failed to push metrics to Pushgateway: %v", err)
	}
}
