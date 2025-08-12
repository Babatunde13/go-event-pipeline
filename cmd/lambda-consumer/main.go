package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/Babatunde13/event-pipeline/internal/config"
	"github.com/Babatunde13/event-pipeline/internal/event"
	"github.com/Babatunde13/event-pipeline/internal/redis"
	"github.com/Babatunde13/event-pipeline/internal/telemetry"
)

func init() {
	log.Println("Initializing Lambda function handler...")
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	config.Load("event-pipeline-secret")
	log.Printf("Received SQS event with %d records", len(sqsEvent.Records))
	redisClient := redis.New(config.Cfg.RedisAddress)
	log.Printf("Redis client initialized with address: %s", config.Cfg.RedisAddress)
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
		err = redisClient.Set(fmt.Sprintf("%s-kafka", e.EventID), string(jsonStr), 5*time.Minute)
		telemetry.PushMetrics(config.Cfg.PrometheusPushGatewayUrl, time.Since(start).Seconds(), false, false, err == nil)

		if err != nil {
			log.Printf("failed to store event in redis: %v", err)
		} else {
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
