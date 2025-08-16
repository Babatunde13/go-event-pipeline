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

var ddb database.Database

func init() {
	config.Load("event-pipeline-secret")
	ddb = database.NewDynamo(config.Cfg.AwsConfig)
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

func handler(ctx context.Context, payload events.KafkaEvent) {
	log.Println("Kafka consumer initialized with topic:", config.Cfg.KafkaTopic)

	if len(payload.Records) > 0 {
		for partKey, batch := range payload.Records {
			log.Printf("Processing partition key: %s with %d records", partKey, len(batch))
			processBatch(ctx, batch)
		}
	} else {
		log.Printf("Received empty payload for %s", payload.EventSourceARN)
	}
}

func main() {
	lambda.Start(handler)
	log.Println("Lambda function started")
}
