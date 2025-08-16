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
	log.Println("Configuration loaded successfully")
	ddb = database.NewDynamo(config.Cfg.AwsConfig)
	log.Printf("DynamoDB client initialized")
}

func handler(ctx context.Context, data events.KafkaEvent) {
	log.Println("Kafka consumer initialized with topic:", config.Cfg.KafkaTopic)

	for partition, batch := range data.Records {
		log.Printf("Processing partition: %s with %d records", partition, len(batch))
		for _, record := range batch {
			msg, err := base64.StdEncoding.DecodeString(record.Value)
			if err != nil {
				log.Printf("failed to decode message: %v", err)
				continue
			}
			log.Printf("Received message: topic=%s partition=%d offset=%d key=%s", record.Topic, record.Partition, record.Offset, record.Key)
			// decode the message into an Event struct
			log.Printf("Processing message: %s", msg)
			if len(msg) == 0 {
				log.Println("Received empty message, skipping")
				continue
			}
			var e event.Event
			if err = json.Unmarshal([]byte(msg), &e); err != nil {
				log.Printf("invalid event data: %v", err)
				continue
			}

			start := time.Now()
			err = e.Save(ctx, ddb, event.SourceKafka)
			telemetry.PushMetrics(config.Cfg.PrometheusPushGatewayUrl, float64(time.Since(start).Milliseconds()), true, false, err == nil)

			if err != nil {
				log.Printf("failed to save: %v", err)
			} else {
				log.Printf("event processed: %s - %s", e.EventType, e.EventID)
			}
		}
	}
}

func main() {
	lambda.Start(handler)
	log.Println("Lambda function started")
}
