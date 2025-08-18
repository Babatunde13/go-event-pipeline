package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/Babatunde13/event-pipeline/internal/config"
	"github.com/Babatunde13/event-pipeline/internal/database"
	"github.com/Babatunde13/event-pipeline/internal/event"
	"github.com/Babatunde13/event-pipeline/internal/telemetry"
)

var ddb database.Database

func init() {
	config.Load("event-pipeline-secret")
	ddb = database.NewDynamo(config.Cfg.AwsConfig)
}

func handler(ctx context.Context, ebEvent events.EventBridgeEvent) error {
	log.Printf("Received EventBridge event: source=%s type=%s id=%s", ebEvent.Source, ebEvent.DetailType, ebEvent.ID)

	var e event.Event
	if err := json.Unmarshal(ebEvent.Detail, &e); err != nil {
		log.Printf("failed to parse event detail: %v", err)
		return nil // continue to next
	}

	log.Printf("[EventBridge] received event: %s - %s", e.EventType, e.EventID)

	startMs := int64(e.Timestamp) // timestamp in ms in utc
	var start time.Time
	if startMs == 0 {
		start = time.Now()
	} else {
		start = time.Unix(0, startMs*int64(time.Millisecond))
	}
	err := e.Save(ctx, ddb, event.SourceEventBridge)
	telemetry.PushMetrics(
		config.Cfg.PrometheusPushGatewayUrl,
		float64(time.Since(start).Milliseconds()),
		false, err == nil,
	)

	if err != nil {
		log.Printf("failed to store event in DynamoDB: %v", err)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
