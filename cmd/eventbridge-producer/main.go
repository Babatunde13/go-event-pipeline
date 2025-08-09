package main

import (
	"context"
	"log"
	"time"

	"github.com/Babatunde13/event-pipeline/internal/config"
	"github.com/Babatunde13/event-pipeline/internal/event"
	"github.com/Babatunde13/event-pipeline/internal/eventbridge"
	"github.com/Babatunde13/event-pipeline/internal/telemetry"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func init() {
	config.Load("go-event-pipeline-secret")
	telemetry.Init()
	go telemetry.StartServer(config.Cfg.PrometheusPort)
}

func router() {
	eb := eventbridge.New(*config.Cfg.AwsConfig, config.Cfg.EventBusName)
	r := gin.Default()
	r.POST("/event", func(c *gin.Context) {
		var e event.Event
		if err := c.ShouldBindJSON(&e); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		start := time.Now()
		err := eb.PutEvent(context.Background(), "ecommerce.analytics", string(e.EventType), e)
		telemetry.EventBridgeProcessingDuration.Observe(time.Since(start).Seconds())

		if err != nil {
			log.Printf("failed to send event: %v", err)
			c.JSON(500, gin.H{"error": err.Error()})
		} else {
			telemetry.EventBridgeEventsProcessed.Inc()
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
