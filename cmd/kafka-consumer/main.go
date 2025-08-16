package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"github.com/Babatunde13/event-pipeline/internal/config"
	"github.com/Babatunde13/event-pipeline/internal/database"
	"github.com/Babatunde13/event-pipeline/internal/event"
	"github.com/Babatunde13/event-pipeline/internal/telemetry"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type kafkaEventMap struct {
	Records map[string][]events.KafkaRecord `json:"records"`
}
type kafkaEventArray struct {
	Records []events.KafkaRecord `json:"records"`
}

var ddb database.Database

func init() {
	config.Load("event-pipeline-secret")
	log.Println("Configuration loaded successfully")
	ddb = database.NewDynamo(config.Cfg.AwsConfig)
	log.Printf("DynamoDB client initialized")
}

func processBatch(ctx context.Context, batch []events.KafkaRecord) {
	for _, record := range batch {
		msg, err := base64.StdEncoding.DecodeString(record.Value)
		if err != nil {
			log.Printf("failed to decode message: %v", err)
			continue
		}
		if len(msg) == 0 {
			log.Println("Empty message, skipping")
			continue
		}

		log.Printf("Message: topic=%s partition=%d offset=%d key=%q",
			record.Topic, record.Partition, record.Offset, string(record.Key))

		var e event.Event
		if err := json.Unmarshal(msg, &e); err != nil {
			log.Printf("Invalid event data: %v", err)
			continue
		}

		start := time.Now()
		err = e.Save(ctx, ddb, event.SourceKafka)
		telemetry.PushMetrics(
			config.Cfg.PrometheusPushGatewayUrl,
			float64(time.Since(start).Milliseconds()),
			true, false, err == nil,
		)

		if err != nil {
			log.Printf("Save failed: %v", err)
		} else {
			log.Printf("Event processed: %s - %s", e.EventType, e.EventID)
		}
	}
}

func handler(ctx context.Context, payload []byte) {
	log.Println("Kafka consumer initialized with topic:", config.Cfg.KafkaTopic)

	// Try map shape first: {"records": {"topic-0": [ {...}, ... ]}}
	var m kafkaEventMap
	if err := json.Unmarshal(payload, &m); err == nil && len(m.Records) > 0 {
		for partKey, batch := range m.Records {
			log.Printf("Processing partition key: %s with %d records", partKey, len(batch))
			processBatch(ctx, batch)
		}
		return
	}

	// Try array shape: {"records": [ {...}, ... ]}
	var a kafkaEventArray
	if err := json.Unmarshal(payload, &a); err == nil && len(a.Records) > 0 {
		log.Printf("Processing array of %d records", len(a.Records))
		processBatch(ctx, a.Records)
		return
	}
}

func main() {
	lambda.Start(handler)
	log.Println("Lambda function started")
}
