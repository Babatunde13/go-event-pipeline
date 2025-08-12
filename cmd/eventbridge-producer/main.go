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

type router struct {
	eb *eventbridge.Client
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

	start := time.Now()
	e.Timestamp = start.UTC() // Ensure timestamp is set to current time
	err := r.eb.PutEvent(context.Background(), config.Cfg.EventBusSource, string(e.EventType), e)
	telemetry.PushMetrics(config.Cfg.PrometheusPushGatewayUrl, time.Since(start).Seconds(), false, true, err == nil)

	if err != nil {
		log.Printf("failed to send event: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
	} else {
		log.Printf("event sent: %s - %s", e.EventType, e.EventID)
		c.JSON(200, gin.H{"status": "ok"})
	}
}

func main() {
	eb := eventbridge.New(*config.Cfg.AwsConfig, config.Cfg.EventBusName)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gin.Recovery())
	api := &router{eb: eb}
	r.POST("/event/eventbridge", api.sendEvent)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "not found"})
	})

	ginLambda = ginadapter.New(r)
	log.Println("Starting lambda server....")
	lambda.Start(handler)
}
