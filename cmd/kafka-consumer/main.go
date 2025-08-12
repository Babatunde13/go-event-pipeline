package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Babatunde13/event-pipeline/internal/config"
	"github.com/Babatunde13/event-pipeline/internal/event"
	"github.com/Babatunde13/event-pipeline/internal/kafka"
	"github.com/Babatunde13/event-pipeline/internal/redis"
	"github.com/Babatunde13/event-pipeline/internal/telemetry"
)

func main() {
	config.Load("kafka-consumer-secret")
	rds := redis.New(config.Cfg.RedisAddress)
	consumer := kafka.NewConsumer(config.Cfg.KafkaBrokers, config.Cfg.KafkaTopic, "event-consumer-group")
	defer consumer.Close()

	ctx := context.Background()
	log.Println("Kafka consumer started...")

	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		var e event.Event
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			log.Printf("invalid event data: %v", err)
			continue
		}

		jsonStr, _ := json.Marshal(e)
		start := time.Now()
		err = rds.Set(e.EventID, string(jsonStr), 5*time.Minute)
		telemetry.PushMetrics(config.Cfg.PrometheusPushGatewayUrl, time.Since(start).Seconds(), true, false, err == nil)

		if err != nil {
			log.Printf("failed to store in redis: %v", err)
		} else {
			log.Printf("event processed: %s - %s", e.EventType, e.EventID)
		}
	}
}
