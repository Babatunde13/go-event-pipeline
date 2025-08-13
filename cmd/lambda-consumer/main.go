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
	log.Println("Initializing Lambda function handler...")
	config.Load("event-pipeline-secret")
	ddb = database.NewDynamo(config.Cfg.AwsConfig)
	log.Printf("DynamoDB client initialized")
}

func handler(ctx context.Context, ebEvent events.EventBridgeEvent) error {
	log.Printf("Received EventBridge event: source=%s type=%s id=%s", ebEvent.Source, ebEvent.DetailType, ebEvent.ID)

	var e event.Event
	if err := json.Unmarshal(ebEvent.Detail, &e); err != nil {
		log.Printf("failed to parse event detail: %v", err)
		return nil // continue to next
	}

	log.Printf("[EventBridge] received event: %s - %s", e.EventType, e.EventID)

	start := time.Now()
	err := e.Save(ctx, ddb, event.SourceEventBridge)
	telemetry.PushMetrics(config.Cfg.PrometheusPushGatewayUrl, time.Since(start).Seconds(), false, false, err == nil)

	if err != nil {
		log.Printf("failed to store event in DynamoDB: %v", err)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
