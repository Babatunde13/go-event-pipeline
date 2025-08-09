package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/Babatunde13/event-pipeline/internal/config"
	"github.com/Babatunde13/event-pipeline/internal/event"
	"github.com/Babatunde13/event-pipeline/internal/redis"
	"github.com/Babatunde13/event-pipeline/internal/telemetry"
)

var redisClient *redis.Client

func init() {
	config.Load("go-event-pipeline-secret")
	redisClient = redis.New(config.Cfg.RedisAddress)
	telemetry.Init()
	go telemetry.StartServer(config.Cfg.PrometheusPort)
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		var e event.Event
		if err := json.Unmarshal([]byte(record.Body), &e); err != nil {
			log.Printf("failed to parse event: %v", err)
			continue
		}

		log.Printf("[SQS] received event: %s - %s", e.EventType, e.EventID)

		jsonStr, err := json.Marshal(e)
		if err != nil {
			log.Printf("failed to serialize event: %v", err)
			continue
		}

		start := time.Now()
		err = redisClient.Set(e.EventID, string(jsonStr), 5*time.Minute)
		telemetry.EventBridgeProcessingDuration.Observe(time.Since(start).Seconds())

		if err != nil {
			log.Printf("failed to store event in redis: %v", err)
		} else {
			telemetry.EventBridgeEventsProcessed.Inc()
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
