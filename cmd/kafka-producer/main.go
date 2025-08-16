package main

import (
	"context"
	"log"
	"time"

	"github.com/Babatunde13/event-pipeline/internal/config"
	"github.com/Babatunde13/event-pipeline/internal/event"
	"github.com/Babatunde13/event-pipeline/internal/kafka"
	"github.com/Babatunde13/event-pipeline/internal/telemetry"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

type router struct {
	producer *kafka.Producer
}

func init() {
	config.Load("event-pipeline-secret")
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func (r *router) sendEvent(c *gin.Context) {
	var e event.Event
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Received event: %s - %s", e.EventType, e.EventID)
	start := time.Now()
	e.Timestamp = int(start.UTC().UnixMilli()) // Ensure timestamp is set to current time
	data, _ := e.ToJSON()
	err := r.producer.SendMessage(context.Background(), e.EventID, data)
	telemetry.PushMetrics(config.Cfg.PrometheusPushGatewayUrl, time.Since(start).Seconds(), true, true, err == nil)
	if err != nil {
		log.Printf("failed to send event: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
	} else {
		log.Printf("event sent: %s - %s", e.EventType, e.EventID)
		c.JSON(200, gin.H{"status": "ok"})
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gin.Recovery())
	producer, err := kafka.NewProducer()
	if err != nil {
		log.Fatalf("failed to create Kafka producer: %v", err)
	}
	log.Println("Kafka producer initialized with brokers:", config.Cfg.Brokers)

	api := &router{producer: producer}
	r.POST("/event/kafka", api.sendEvent)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "not found"})
	})

	ginLambda = ginadapter.New(r)
	log.Println("Starting lambda server....")
	lambda.Start(handler)
}
