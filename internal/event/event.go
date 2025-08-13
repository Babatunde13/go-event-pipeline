package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Babatunde13/event-pipeline/internal/database"
	"github.com/google/uuid"
)

type EventType string
type EventSource string

var (
	ViewProduct EventType = "view_product"
	AddToCart   EventType = "add_to_cart"
	Checkout    EventType = "checkout"

	EventTypes = []EventType{
		ViewProduct,
		AddToCart,
		Checkout,
	}

	SourceKafka       EventSource = "kafka"
	SourceEventBridge EventSource = "eventbridge"
)

type Event struct {
	EventID     string                 `json:"event_id" dynamodbav:"event_id"` // partition key
	EventType   EventType              `json:"event_type" dynamodbav:"event_type"`
	UserID      string                 `json:"user_id" dynamodbav:"user_id"`
	Timestamp   time.Time              `json:"timestamp,omitempty" dynamodbav:"-"`
	Metadata    map[string]interface{} `json:"metadata" dynamodbav:"metadata"`
	Source      EventSource            `dynamodbav:"source,omitempty"`
	dbTimeStamp int                    `dynamodbav:"timestamp,omitempty"` // sort key for DynamoDB
}

func New(eventType EventType, userID string, metadata map[string]interface{}) Event {
	return Event{
		EventID:   uuid.NewString(),
		EventType: eventType,
		UserID:    userID,
		Timestamp: time.Now().UTC(),
		Metadata:  metadata,
	}
}

func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func FromJSON(data []byte) (*Event, error) {
	var e Event
	err := json.Unmarshal(data, &e)
	return &e, err
}

func (e *Event) Save(ctx context.Context, dbClient database.Database, source EventSource) error {
	e.Source = source
	e.dbTimeStamp = int(time.Now().Unix())
	err := dbClient.Save(ctx, "events", e)
	if err != nil {
		return err
	}
	return nil
}
