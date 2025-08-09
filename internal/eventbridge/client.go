// eventbridge/client.go
package eventbridge

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"log"
)

type Client struct {
	ebClient *eventbridge.Client
	busName  string
}

func New(cfg aws.Config, busName string) *Client {
	return &Client{
		ebClient: eventbridge.NewFromConfig(cfg),
		busName:  busName,
	}
}

func (c *Client) PutEvent(ctx context.Context, source, detailType string, detail interface{}) error {
	payload, err := json.Marshal(detail)
	if err != nil {
		return err
	}

	_, err = c.ebClient.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{
			{
				Source:       aws.String(source),
				DetailType:   aws.String(detailType),
				Detail:       aws.String(string(payload)),
				EventBusName: aws.String(c.busName),
			},
		},
	})
	if err != nil {
		log.Printf("Failed to publish to EventBridge: %v", err)
	}
	return err
}
