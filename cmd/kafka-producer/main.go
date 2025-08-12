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

func init() {
	config.Load("event-pipeline-secret")
}

func router() {
	r := gin.Default()
	producer := kafka.NewProducer(config.Cfg.KafkaBroker, config.Cfg.KafkaTopic)
	defer producer.Close()

	err := producer.CreateTopic(context.Background(), config.Cfg.KafkaTopic, 1, 1)
	if err != nil {
		log.Println("topic creation error (might already exist):", err)
	}
	r.POST("/event", func(c *gin.Context) {
		var e event.Event
		if err := c.ShouldBindJSON(&e); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		data, _ := e.ToJSON()
		start := time.Now()
		err := producer.SendMessage(context.Background(), e.EventID, data)
		telemetry.PushMetrics(config.Cfg.PrometheusPushGatewayUrl, time.Since(start).Seconds(), true, true, err == nil)
		if err != nil {
			log.Printf("failed to send event: %v", err)
			c.JSON(500, gin.H{"error": err.Error()})
		} else {
			log.Printf("event sent: %s - %s", e.EventType, e.EventID)
			c.JSON(200, gin.H{"status": "ok"})
		}
	})

	ginLambda = ginadapter.New(r)
	log.Println("Starting lambda server....")
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	router()
	lambda.Start(handler)
}
