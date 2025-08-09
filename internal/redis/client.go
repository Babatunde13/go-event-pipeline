package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

var ctx = context.Background()

type Client struct {
	client *redis.Client
}

func New(addr string) *Client {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return &Client{client: client}
}

func (c *Client) Set(key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Get(key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Close() error {
	return c.client.Close()
}
