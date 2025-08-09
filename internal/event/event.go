package event

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type EventType string

var (
	ViewProduct EventType = "view_product"
	AddToCart   EventType = "add_to_cart"
	Checkout    EventType = "checkout"

	EventTypes = []EventType{
		ViewProduct,
		AddToCart,
		Checkout,
	}
)

type Event struct {
	EventID   string                 `json:"event_id"`
	EventType EventType              `json:"event_type"`
	UserID    string                 `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
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
